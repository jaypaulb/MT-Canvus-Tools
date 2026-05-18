import { request as undiciRequest } from "undici";
import { APIError, NetworkError } from "./errors.js";
import type { Transport } from "./transport.js";

/** Options accepted by the streaming helpers. */
export interface StreamOptions {
  /** Caller-provided abort signal. Aborting tears down the connection. */
  readonly signal?: AbortSignal;
  /** Extra query parameters (e.g. resource filters). */
  readonly query?: Readonly<Record<string, string | number | boolean | undefined>>;
}

/**
 * Stream newline-delimited JSON from a GET endpoint as an async iterable.
 *
 * Implementation notes:
 * - Uses `undici.request` directly rather than the native `fetch` to get
 *   true backpressure on `response.body` and reliable connection reuse.
 * - Skips empty lines (the Canvus server emits them every 15 s as
 *   keepalive pings).
 * - When the server returns an array as the first payload (initial state
 *   of a list endpoint), it is yielded as a single chunk; subsequent
 *   updates arrive as standalone objects. Callers can post-flatten if
 *   they prefer a uniform per-item stream.
 *
 * @example
 * ```ts
 * const ctrl = new AbortController();
 * for await (const event of streamNdjson<Canvas>(transport, "canvases", {
 *   signal: ctrl.signal,
 * })) {
 *   logger.info({ id: event.id }, "canvas event");
 * }
 * ```
 *
 * @throws {APIError} If the server responds with a non-2xx status.
 * @throws {NetworkError} On transport-level failure.
 */
export async function* streamNdjson<T>(
  transport: Transport,
  path: string,
  opts: StreamOptions = {},
): AsyncGenerator<T, void, void> {
  const url = transport.buildUrl(path, { ...opts.query, subscribe: "true" });
  const headers: Record<string, string> = {
    accept: "application/json",
  };
  if (transport.config.apiKey) headers["private-token"] = transport.config.apiKey;

  let response: Awaited<ReturnType<typeof undiciRequest>>;
  try {
    response = await undiciRequest(url.toString(), {
      method: "GET",
      headers,
      ...(opts.signal ? { signal: opts.signal } : {}),
    });
  } catch (err) {
    throw new NetworkError(`stream ${path} failed to connect`, { cause: err });
  }

  const { statusCode, body } = response;
  if (statusCode >= 400) {
    const text = await body.text();
    throw new APIError(statusCode, safeParse(text), `stream ${path} returned ${statusCode.toString()}`);
  }

  let buffer = "";
  try {
    for await (const chunk of body) {
      // chunk is a Buffer in Node; toString() decodes via utf-8 by default.
      buffer += (chunk as Buffer).toString("utf8");
      let newlineIdx = buffer.indexOf("\n");
      while (newlineIdx !== -1) {
        const line = buffer.slice(0, newlineIdx).trim();
        buffer = buffer.slice(newlineIdx + 1);
        if (line !== "") {
          yield JSON.parse(line) as T;
        }
        newlineIdx = buffer.indexOf("\n");
      }
    }
    // Flush any remaining buffered line at EOF.
    const tail = buffer.trim();
    if (tail !== "") {
      yield JSON.parse(tail) as T;
    }
  } catch (err) {
    if ((opts.signal?.aborted ?? false) && (err as Error).name === "AbortError") {
      return;
    }
    throw new NetworkError(`stream ${path} interrupted`, { cause: err });
  }
}

/** Best-effort JSON parse used for streaming-error bodies. */
function safeParse(text: string): unknown {
  if (text === "") return undefined;
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}
