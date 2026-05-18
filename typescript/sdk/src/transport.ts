import { APIError, AuthError, NetworkError, NotFoundError } from "./errors.js";
import type { Config } from "./config.js";

/**
 * HTTP methods used by the SDK.
 */
export type HttpMethod = "GET" | "POST" | "PATCH" | "PUT" | "DELETE";

/** Per-request options that can be passed to {@link Transport.request}. */
export interface RequestOptions {
  /** Query-string parameters. Values are stringified; `undefined` is skipped. */
  readonly query?: Readonly<Record<string, string | number | boolean | undefined>>;
  /** Additional headers (merged on top of defaults). */
  readonly headers?: Readonly<Record<string, string>>;
  /** Per-request abort signal. */
  readonly signal?: AbortSignal;
  /** Override the default `Accept` header (e.g. `application/octet-stream`). */
  readonly accept?: string;
  /** Skip JSON parsing of the response body. */
  readonly raw?: boolean;
  /** Per-request timeout (ms). Defaults to `config.timeoutMs`. */
  readonly timeoutMs?: number;
}

/**
 * Body shapes accepted by the transport.
 *
 * `undefined` sends no body. A plain object is JSON-encoded. `FormData`,
 * `Buffer`, and `Uint8Array` are sent as-is with no content-type override.
 */
export type RequestBody = undefined | unknown | FormData | Buffer | Uint8Array;

/**
 * Low-level HTTP transport used by every resource client.
 *
 * Wraps Node 20's built-in `fetch` with:
 * - automatic `Private-Token` header injection
 * - URL composition with query-string serialisation
 * - error-mapping to the {@link CanvusError} hierarchy
 * - per-request timeout via `AbortController`
 */
export class Transport {
  /**
   * @param config - Frozen runtime config produced by `loadConfig` or
   *   `buildConfig`. Held by reference (not copied) so a future mutating
   *   helper could swap credentials atomically.
   */
  constructor(public readonly config: Config) {}

  /**
   * Issue an HTTP request and parse the response as JSON.
   *
   * For binary responses (download endpoints), pass `raw: true` and the
   * raw `Response` is returned so callers can read `.arrayBuffer()` or
   * `.body`.
   *
   * @throws {AuthError} On HTTP 401/403.
   * @throws {NotFoundError} On HTTP 404.
   * @throws {APIError} On other non-2xx responses.
   * @throws {NetworkError} On transport-level failures (DNS, TLS, timeout).
   */
  async request<T>(
    method: HttpMethod,
    path: string,
    body?: RequestBody,
    opts: RequestOptions = {},
  ): Promise<T> {
    const response = await this.rawRequest(method, path, body, opts);
    if (opts.raw === true) {
      // SAFETY: callers using `raw: true` must specify T = Response.
      return response as unknown as T;
    }
    // 204 No Content
    if (response.status === 204) return undefined as T;
    const contentType = response.headers.get("content-type") ?? "";
    if (contentType.includes("application/json")) {
      return (await response.json()) as T;
    }
    // Treat as text fallback; downstream callers can post-process.
    return (await response.text()) as unknown as T;
  }

  /**
   * Issue an HTTP request and return the raw {@link Response} without
   * consuming the body. Used by binary download endpoints and the
   * streaming subscribe client.
   */
  async rawRequest(
    method: HttpMethod,
    path: string,
    body?: RequestBody,
    opts: RequestOptions = {},
  ): Promise<Response> {
    const url = this.buildUrl(path, opts.query);
    const headers = this.buildHeaders(body, opts);
    const { signal, cancel } = this.combineSignals(
      opts.signal,
      opts.timeoutMs ?? this.config.timeoutMs,
    );

    const serialised = this.serialiseBody(body);
    let response: Response;
    try {
      response = await fetch(url, {
        method,
        headers,
        signal,
        ...(serialised !== undefined && { body: serialised }),
      });
    } catch (err) {
      const aborted = signal.aborted;
      const message = aborted
        ? `${method} ${url.pathname} aborted or timed out after ${(opts.timeoutMs ?? this.config.timeoutMs).toString()}ms`
        : `${method} ${url.pathname} failed`;
      throw new NetworkError(message, { cause: err });
    } finally {
      cancel();
    }

    if (!response.ok) {
      await this.throwForStatus(response, method, url);
    }
    return response;
  }

  /** Build an absolute URL from the configured base + a relative path. */
  buildUrl(
    path: string,
    query?: Readonly<Record<string, string | number | boolean | undefined>>,
  ): URL {
    // Strip leading "/" so relative resolution works correctly.
    const relative = path.startsWith("/") ? path.slice(1) : path;
    const url = new URL(relative, this.config.apiBaseUrl);
    if (query) {
      for (const [k, v] of Object.entries(query)) {
        if (v === undefined) continue;
        url.searchParams.set(k, String(v));
      }
    }
    return url;
  }

  private buildHeaders(body: RequestBody, opts: RequestOptions): Headers {
    const headers = new Headers();
    headers.set("accept", opts.accept ?? "application/json");
    if (this.config.apiKey) {
      headers.set("private-token", this.config.apiKey);
    }
    const isPlainObject =
      body !== undefined &&
      typeof body === "object" &&
      body !== null &&
      !(body instanceof FormData) &&
      !(globalThis.Buffer !== undefined && body instanceof globalThis.Buffer) &&
      !(body instanceof Uint8Array);
    if (isPlainObject) {
      headers.set("content-type", "application/json");
    }
    if (opts.headers) {
      for (const [k, v] of Object.entries(opts.headers)) {
        headers.set(k, v);
      }
    }
    return headers;
  }

  private serialiseBody(body: RequestBody): FormData | Uint8Array | string | undefined {
    if (body === undefined) return undefined;
    if (body instanceof FormData) return body;
    if (globalThis.Buffer !== undefined && body instanceof globalThis.Buffer) {
      return body as unknown as Uint8Array;
    }
    if (body instanceof Uint8Array) return body;
    return JSON.stringify(body);
  }

  private combineSignals(
    external: AbortSignal | undefined,
    timeoutMs: number,
  ): { signal: AbortSignal; cancel: () => void } {
    const ctrl = new AbortController();
    const onAbort = (): void => {
      ctrl.abort(external?.reason);
    };
    if (external) {
      if (external.aborted) ctrl.abort(external.reason);
      else external.addEventListener("abort", onAbort, { once: true });
    }
    const timer = setTimeout(() => {
      ctrl.abort(new Error(`request timed out after ${timeoutMs.toString()}ms`));
    }, timeoutMs);
    return {
      signal: ctrl.signal,
      cancel: (): void => {
        clearTimeout(timer);
        external?.removeEventListener("abort", onAbort);
      },
    };
  }

  private async throwForStatus(response: Response, method: HttpMethod, url: URL): Promise<never> {
    const text = await response.text();
    const parsed = safeJson(text);
    const message = `${method} ${url.pathname} returned ${response.status.toString()}`;
    if (response.status === 401) {
      throw new AuthError("expired", `${message}: unauthorized`, {
        cause: new APIError(401, parsed, message),
      });
    }
    if (response.status === 403) {
      throw new AuthError("forbidden", `${message}: forbidden`, {
        cause: new APIError(403, parsed, message),
      });
    }
    if (response.status === 404) {
      throw new NotFoundError(parsed, message);
    }
    throw new APIError(response.status, parsed, message);
  }
}

/** Best-effort JSON parse; returns the raw text on failure. */
export function safeJson(text: string): unknown {
  if (text === "") return undefined;
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}
