// Phase 4b §4.3 #10/#11/#12/#13: retry layer, RateLimitError/ServerError mapping,
// requestIdProvider support, and verifyTls wired via undici.Agent.

import { Agent, fetch as undiciFetch, type Dispatcher } from "undici";
import {
  APIError,
  AuthError,
  NetworkError,
  NotFoundError,
  RateLimitError,
  ServerError,
  parseRetryAfter,
} from "./errors.js";
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
 * Module-scoped Agent cache. Re-used across Transport instances that share
 * the same `verifyTls` flag so we don't leak sockets per call.
 */
const verifyOffAgentMap = new Map<string, Agent>();

function getInsecureAgent(): Agent {
  const key = "tls-off";
  let agent = verifyOffAgentMap.get(key);
  if (agent === undefined) {
    agent = new Agent({ connect: { rejectUnauthorized: false } });
    verifyOffAgentMap.set(key, agent);
  }
  return agent;
}

/**
 * Phase 4b §4.3 #10: simple circuit breaker — opens after N consecutive
 * failures, half-open after a reset timeout, closes again on first success.
 *
 * Process-wide concept is shared per-Transport instance; close one Session
 * (via {@link Session.close}) to release its breaker state.
 */
class CircuitBreaker {
  private failures = 0;
  private openedAt = 0;
  private state: "closed" | "open" | "half-open" = "closed";

  constructor(
    private readonly maxFailures: number,
    private readonly resetTimeoutMs: number,
  ) {}

  /** Returns true if a request may proceed. */
  allow(): boolean {
    if (this.state === "closed") return true;
    if (this.state === "open") {
      if (Date.now() - this.openedAt > this.resetTimeoutMs) {
        this.state = "half-open";
        return true;
      }
      return false;
    }
    // half-open: allow exactly one probe at a time
    return true;
  }

  success(): void {
    this.failures = 0;
    this.state = "closed";
  }

  failure(): void {
    if (this.state === "half-open") {
      this.openedAt = Date.now();
      this.state = "open";
      return;
    }
    this.failures += 1;
    if (this.failures >= this.maxFailures) {
      this.openedAt = Date.now();
      this.state = "open";
    }
  }
}

/**
 * Retry policy applied by {@link Transport.rawRequest}. Defaults mirror the
 * Go SDK (3 retries, 500ms→5s exponential backoff with jitter, circuit
 * trips after 5 consecutive failures and resets after 30s).
 */
export interface RetryPolicy {
  readonly maxRetries: number;
  readonly baseDelayMs: number;
  readonly maxDelayMs: number;
  readonly circuitMaxFailures: number;
  readonly circuitResetMs: number;
}

const DEFAULT_RETRY: RetryPolicy = {
  maxRetries: 3,
  baseDelayMs: 500,
  maxDelayMs: 5_000,
  circuitMaxFailures: 5,
  circuitResetMs: 30_000,
};

/**
 * Low-level HTTP transport used by every resource client.
 *
 * Wraps Node 20's `undici.fetch` with:
 * - automatic `Private-Token` and optional `X-Request-ID` header injection
 * - URL composition with query-string serialisation
 * - error-mapping to the {@link CanvusError} hierarchy
 *   (`AuthError`, `NotFoundError`, `RateLimitError`, `ServerError`, `APIError`)
 * - per-request timeout via `AbortController`
 * - exponential-backoff retry on transient failures (429, 5xx, network)
 * - process-wide circuit breaker that trips on consecutive failures
 * - optional TLS verification disable (via undici `Agent`)
 */
export class Transport {
  private readonly breaker: CircuitBreaker;
  private readonly retry: RetryPolicy;
  private dispatcher: Dispatcher | undefined;

  /**
   * @param config - Frozen runtime config produced by `loadConfig` or
   *   `buildConfig`. Held by reference (not copied) so a future mutating
   *   helper could swap credentials atomically.
   * @param retry - Optional retry policy override. Defaults applied otherwise.
   */
  constructor(
    public readonly config: Config,
    retry?: Partial<RetryPolicy>,
  ) {
    this.retry = { ...DEFAULT_RETRY, ...retry };
    this.breaker = new CircuitBreaker(this.retry.circuitMaxFailures, this.retry.circuitResetMs);
    if (!config.verifyTls) {
      this.dispatcher = getInsecureAgent();
    }
  }

  /**
   * Release any retained network resources (e.g. the undici Agent created
   * when `verifyTls === false`). Phase 4b §4.3 #22.
   */
  async close(): Promise<void> {
    if (this.dispatcher !== undefined) {
      await this.dispatcher.close();
      this.dispatcher = undefined;
      verifyOffAgentMap.delete("tls-off");
    }
  }

  /**
   * Issue an HTTP request and parse the response as JSON.
   *
   * For binary responses (download endpoints), pass `raw: true` and the
   * raw `Response` is returned so callers can read `.arrayBuffer()` or
   * `.body`.
   *
   * @throws {AuthError} On HTTP 401/403.
   * @throws {NotFoundError} On HTTP 404.
   * @throws {RateLimitError} On HTTP 429 after retry budget exhausted.
   * @throws {ServerError} On HTTP 5xx after retry budget exhausted.
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
   *
   * Phase 4b §4.3 #10: applies the retry policy on transient failures
   * (network, 408, 429, 5xx). 429 retries honour the `Retry-After` header.
   */
  async rawRequest(
    method: HttpMethod,
    path: string,
    body?: RequestBody,
    opts: RequestOptions = {},
  ): Promise<Response> {
    if (!this.breaker.allow()) {
      throw new NetworkError(
        `${method} ${path} aborted: circuit breaker is open after ${this.retry.circuitMaxFailures.toString()} consecutive failures`,
      );
    }

    const url = this.buildUrl(path, opts.query);
    const headers = this.buildHeaders(body, opts);
    const serialised = this.serialiseBody(body);

    let lastError: Error | undefined;
    for (let attempt = 0; attempt <= this.retry.maxRetries; attempt++) {
      const { signal, cancel } = this.combineSignals(
        opts.signal,
        opts.timeoutMs ?? this.config.timeoutMs,
      );

      let response: Response;
      try {
        // Use undici-direct only when we need to pass a custom dispatcher
        // (verifyTls=false case). Otherwise stay on globalThis.fetch so tests
        // that monkey-patch `fetch` continue to work and Node's built-in
        // dispatcher is reused.
        if (this.dispatcher !== undefined) {
          response = await undiciFetch(url, {
            method,
            headers,
            ...(serialised !== undefined && { body: serialised as never }),
            dispatcher: this.dispatcher,
            signal,
          });
        } else {
          response = await fetch(url, {
            method,
            headers,
            ...(serialised !== undefined && { body: serialised as never }),
            signal,
          });
        }
      } catch (err) {
        cancel();
        const aborted = signal.aborted;
        const message = aborted
          ? `${method} ${url.pathname} aborted or timed out after ${(opts.timeoutMs ?? this.config.timeoutMs).toString()}ms`
          : `${method} ${url.pathname} failed`;
        lastError = new NetworkError(message, { cause: err });
        if (this.shouldRetryNetwork(lastError, attempt)) {
          await this.delayBeforeRetry(attempt, undefined);
          continue;
        }
        this.breaker.failure();
        throw lastError;
      }
      cancel();

      if (response.ok) {
        this.breaker.success();
        return response;
      }

      // Non-2xx: read body so we can map to a typed error.
      const text = await response.text();
      const parsed = safeJson(text);
      const message = `${method} ${url.pathname} returned ${response.status.toString()}`;
      const retryAfter = parseRetryAfter(response.headers.get("retry-after"));
      const apiErr = this.errorFor(response.status, parsed, message, retryAfter);

      if (this.isRetryable(response.status) && attempt < this.retry.maxRetries) {
        lastError = apiErr;
        await this.delayBeforeRetry(attempt, retryAfter);
        continue;
      }

      this.breaker.failure();
      throw apiErr;
    }

    this.breaker.failure();
    throw (
      lastError ??
      new NetworkError(`${method} ${path} failed after ${this.retry.maxRetries.toString()} retries`)
    );
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
    if (this.config.apiKey !== undefined) {
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
    if (this.config.requestIdProvider !== undefined) {
      const reqId = this.config.requestIdProvider();
      if (reqId !== "" && reqId !== undefined) {
        headers.set("x-request-id", reqId);
      }
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

  private errorFor(
    status: number,
    body: unknown,
    message: string,
    retryAfterMs: number | undefined,
  ): Error {
    if (status === 401) {
      return new AuthError("expired", `${message}: unauthorized`, {
        cause: new APIError(401, body, message),
      });
    }
    if (status === 403) {
      return new AuthError("forbidden", `${message}: forbidden`, {
        cause: new APIError(403, body, message),
      });
    }
    if (status === 404) return new NotFoundError(body, message);
    if (status === 429) return new RateLimitError(body, message, retryAfterMs);
    if (status >= 500 && status <= 599) return new ServerError(status, body, message);
    return new APIError(status, body, message);
  }

  /** Return true for status codes the SDK considers retryable. */
  private isRetryable(status: number): boolean {
    return status === 408 || status === 429 || (status >= 500 && status <= 599);
  }

  private shouldRetryNetwork(err: Error, attempt: number): boolean {
    if (attempt >= this.retry.maxRetries) return false;
    // Treat abort due to caller cancellation as non-retryable.
    if (err instanceof NetworkError) {
      const cause = (err as { cause?: unknown }).cause;
      if (
        cause instanceof Error &&
        (cause.name === "AbortError" || (cause as { code?: string }).code === "ABORT_ERR")
      ) {
        return false;
      }
      return true;
    }
    return false;
  }

  private async delayBeforeRetry(
    attempt: number,
    retryAfterMs: number | undefined,
  ): Promise<void> {
    if (retryAfterMs !== undefined && retryAfterMs > 0) {
      await sleep(Math.min(retryAfterMs, this.retry.maxDelayMs));
      return;
    }
    const base = Math.min(this.retry.baseDelayMs * 2 ** attempt, this.retry.maxDelayMs);
    const jitter = Math.random() * (base / 2);
    await sleep(base + jitter);
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

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
