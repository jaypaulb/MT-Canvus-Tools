import { z } from "zod";
import { ValidationError } from "./errors.js";

/**
 * Zod schema for `CANVUS_*` (and standard log) environment variables.
 *
 * `apiBaseUrl` must end in a slash so that `new URL(path, base)` works
 * correctly for relative paths like `canvases`.
 */
const ConfigSchema = z.object({
  apiBaseUrl: z
    .string()
    .url()
    .transform((u) => (u.endsWith("/") ? u : `${u}/`)),
  apiKey: z.string().min(1).optional(),
  timeoutMs: z.coerce.number().int().positive().default(30_000),
  verifyTls: z
    .enum(["true", "false"])
    .default("true")
    .transform((v) => v === "true"),
  subscribeBuffer: z.coerce.number().int().min(1).default(4),
});

/**
 * Frozen, typed runtime configuration produced by {@link loadConfig}.
 *
 * Phase 4b §4.3 #12: `requestIdProvider`, if set, is invoked once per request
 * to generate the `X-Request-ID` header value. Returning `""` or `undefined`
 * suppresses the header for that call.
 */
export type Config = z.infer<typeof ConfigSchema> & {
  readonly requestIdProvider?: () => string | undefined;
};

/**
 * Parse and freeze the SDK runtime configuration from environment variables.
 *
 * @param env - Environment object; defaults to `process.env`. Pass an
 *   explicit object in tests to avoid leaking the ambient environment.
 * @returns Frozen, typed config object.
 * @throws {ValidationError} When the supplied environment does not
 *   satisfy the schema (e.g. invalid URL, non-positive timeout).
 */
export function loadConfig(env: NodeJS.ProcessEnv = process.env): Config {
  const parsed = ConfigSchema.safeParse({
    apiBaseUrl: env.CANVUS_API_URL,
    apiKey: env.CANVUS_API_KEY,
    timeoutMs: env.CANVUS_TIMEOUT_MS,
    verifyTls: env.CANVUS_VERIFY_TLS,
    subscribeBuffer: env.CANVUS_SUBSCRIBE_BUFFER,
  });

  if (!parsed.success) {
    throw new ValidationError(
      parsed.error.issues.map((i) => ({ path: i.path, message: i.message })),
      "invalid CANVUS_* configuration",
    );
  }
  return Object.freeze(parsed.data);
}

/**
 * Options accepted by `createSession` and ad-hoc {@link buildConfig} calls.
 */
export interface SessionOptions {
  /** Base API URL, e.g. `https://canvus.example.com/api/v1/`. */
  readonly baseUrl: string;
  /** Long-lived API key (sent in `Private-Token` header). */
  readonly apiKey?: string;
  /** Per-request timeout in milliseconds. Defaults to 30_000. */
  readonly timeoutMs?: number;
  /** Verify TLS certificates. Defaults to true. */
  readonly verifyTls?: boolean;
  /**
   * Phase 4b §4.3 #12: optional callback that returns a per-request
   * identifier injected as the `X-Request-ID` header. Useful for
   * correlating SDK calls with server-side logs and traces.
   */
  readonly requestIdProvider?: () => string | undefined;
  /**
   * Channel capacity for buffered subscribe helpers. High-throughput consumers
   * (live dashboards, ai-personas) can raise this value to absorb bursts
   * without applying backpressure to the HTTP response body. Must be >= 1.
   * Defaults to 4. Phase 4d Round B.
   *
   * Note: the SDK's subscribe primitives are async generators (pull-based).
   * This value is exposed on {@link Config.subscribeBuffer} so that wrappers
   * (e.g. `BufferedSubscriber`) can read it. Callers who write their own queue
   * wrapper SHOULD read from `session.config.subscribeBuffer`.
   */
  readonly subscribeBuffer?: number;
}

/**
 * Build a {@link Config} from explicit {@link SessionOptions} rather than
 * from environment variables. Used internally by `createSession`.
 *
 * @throws {ValidationError} When `baseUrl` is not a valid URL.
 */
export function buildConfig(opts: SessionOptions): Config {
  const parsed = ConfigSchema.safeParse({
    apiBaseUrl: opts.baseUrl,
    apiKey: opts.apiKey,
    timeoutMs: opts.timeoutMs?.toString(),
    verifyTls: opts.verifyTls === undefined ? undefined : opts.verifyTls ? "true" : "false",
    subscribeBuffer: opts.subscribeBuffer,
  });

  if (!parsed.success) {
    throw new ValidationError(
      parsed.error.issues.map((i) => ({ path: i.path, message: i.message })),
      "invalid SessionOptions",
    );
  }
  const merged: Config = {
    ...parsed.data,
    ...(opts.requestIdProvider !== undefined && { requestIdProvider: opts.requestIdProvider }),
  };
  return Object.freeze(merged);
}
