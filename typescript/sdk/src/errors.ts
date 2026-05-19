/**
 * Error hierarchy for the Canvus SDK.
 *
 * Every error thrown from SDK code is an instance of {@link CanvusError}.
 * Consumers can `instanceof`-narrow to a subclass, or switch on the
 * `kind` discriminator field for exhaustive handling.
 */

/**
 * Discriminator for the kind of {@link CanvusError}.
 *
 * Phase 4b §4.3 #11 / Senior-dev decision: added `rate-limit`, `server`, and
 * `unsupported` kinds so callers can branch on retryable failures vs.
 * structural API errors without inspecting `status` codes.
 */
export type CanvusErrorKind =
  | "api"
  | "validation"
  | "auth"
  | "network"
  | "not-found"
  | "rate-limit"
  | "server"
  | "unsupported";

/**
 * Base class for every error thrown by the Canvus SDK.
 *
 * Use `err instanceof CanvusError` to detect SDK-originated errors and
 * branch on `err.kind` for exhaustive handling.
 */
export class CanvusError extends Error {
  public readonly kind: CanvusErrorKind;

  constructor(kind: CanvusErrorKind, message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "CanvusError";
    this.kind = kind;
  }
}

/**
 * A single field-level issue produced by validation.
 */
export interface ValidationIssue {
  readonly path: readonly (string | number)[];
  readonly message: string;
}

/**
 * Thrown when an HTTP request returns a non-2xx status.
 *
 * Carries the HTTP status code and the (best-effort parsed) response body.
 */
export class APIError extends CanvusError {
  constructor(
    public readonly status: number,
    public readonly body: unknown,
    message: string,
    options?: ErrorOptions,
  ) {
    super("api", message, options);
    this.name = "APIError";
  }
}

/**
 * Specialised {@link APIError} for HTTP 404 responses.
 *
 * Surfaced separately because "not found" is frequently a normal
 * control-flow case rather than a hard failure.
 */
export class NotFoundError extends APIError {
  constructor(body: unknown, message: string, options?: ErrorOptions) {
    super(404, body, message, options);
    this.name = "NotFoundError";
    // re-tag the discriminator
    (this as { kind: CanvusErrorKind }).kind = "not-found";
  }
}

/**
 * Thrown when input fails runtime validation (e.g. a `zod` schema).
 */
export class ValidationError extends CanvusError {
  constructor(
    public readonly issues: readonly ValidationIssue[],
    message: string,
    options?: ErrorOptions,
  ) {
    super("validation", message, options);
    this.name = "ValidationError";
  }
}

/**
 * Thrown for authentication failures (missing/expired/forbidden token).
 */
export class AuthError extends CanvusError {
  constructor(
    public readonly reason: "missing-token" | "expired" | "forbidden",
    message: string,
    options?: ErrorOptions,
  ) {
    super("auth", message, options);
    this.name = "AuthError";
  }
}

/**
 * Thrown when the underlying network transport fails (DNS, TLS, timeout).
 */
export class NetworkError extends CanvusError {
  constructor(message: string, options?: ErrorOptions) {
    super("network", message, options);
    this.name = "NetworkError";
  }
}

/**
 * Phase 4b §4.3 #11: Specialised {@link APIError} for HTTP 429 Too Many Requests.
 *
 * Parses the optional `Retry-After` header (seconds or HTTP-date) and exposes
 * it as a number-of-milliseconds value on {@link retryAfterMs}.
 */
export class RateLimitError extends APIError {
  /**
   * Number of milliseconds the caller is asked to wait before retrying.
   * `undefined` when the server omitted the `Retry-After` header.
   */
  public readonly retryAfterMs: number | undefined;

  constructor(
    body: unknown,
    message: string,
    retryAfterMs: number | undefined,
    options?: ErrorOptions,
  ) {
    super(429, body, message, options);
    this.name = "RateLimitError";
    this.retryAfterMs = retryAfterMs;
    (this as { kind: CanvusErrorKind }).kind = "rate-limit";
  }
}

/**
 * Phase 4b §4.3 #11: Specialised {@link APIError} for HTTP 5xx responses.
 *
 * Carries the underlying status (500–599); useful for callers that want to
 * implement custom retry policy beyond the SDK's built-in retry layer.
 */
export class ServerError extends APIError {
  constructor(status: number, body: unknown, message: string, options?: ErrorOptions) {
    super(status, body, message, options);
    this.name = "ServerError";
    (this as { kind: CanvusErrorKind }).kind = "server";
  }
}

/**
 * Phase 4b §4.3 #1 / §5.4: Thrown when the caller attempts an operation the
 * server cannot perform.
 *
 * Currently used to reject `widgets.createAny()` calls for `IpVideo` and
 * `RdpConnection` widget types — the server's createElement whitelist does
 * not include them.
 */
export class UnsupportedOperationError extends CanvusError {
  constructor(
    public readonly operation: string,
    message: string,
    options?: ErrorOptions,
  ) {
    super("unsupported", message, options);
    this.name = "UnsupportedOperationError";
  }
}

/**
 * Type guard for {@link CanvusError}.
 */
export function isCanvusError(err: unknown): err is CanvusError {
  return err instanceof CanvusError;
}

/**
 * Parse a `Retry-After` header value into milliseconds.
 *
 * Accepts either an integer number of seconds (per RFC 7231) or an HTTP-date.
 * Returns `undefined` for unparseable or absent values.
 */
export function parseRetryAfter(headerValue: string | null | undefined): number | undefined {
  if (headerValue === null || headerValue === undefined || headerValue === "") {
    return undefined;
  }
  // Integer-seconds form.
  const seconds = Number(headerValue);
  if (Number.isFinite(seconds) && seconds >= 0) {
    return Math.round(seconds * 1000);
  }
  // HTTP-date form.
  const epoch = Date.parse(headerValue);
  if (Number.isFinite(epoch)) {
    const delta = epoch - Date.now();
    return delta > 0 ? delta : 0;
  }
  return undefined;
}
