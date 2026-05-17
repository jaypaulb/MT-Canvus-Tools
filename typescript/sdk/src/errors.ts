/**
 * Error hierarchy for the Canvus SDK.
 *
 * Every error thrown from SDK code is an instance of {@link CanvusError}.
 * Consumers can `instanceof`-narrow to a subclass, or switch on the
 * `kind` discriminator field for exhaustive handling.
 */

/**
 * Discriminator for the kind of {@link CanvusError}.
 */
export type CanvusErrorKind = "api" | "validation" | "auth" | "network" | "not-found";

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
  readonly path: ReadonlyArray<string | number>;
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
 * Type guard for {@link CanvusError}.
 */
export function isCanvusError(err: unknown): err is CanvusError {
  return err instanceof CanvusError;
}
