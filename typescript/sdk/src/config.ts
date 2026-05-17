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
});

/**
 * Frozen, typed runtime configuration produced by {@link loadConfig}.
 */
export type Config = z.infer<typeof ConfigSchema>;

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
    apiBaseUrl: env.CANVUS_API_BASE_URL,
    apiKey: env.CANVUS_API_KEY,
    timeoutMs: env.CANVUS_TIMEOUT_MS,
    verifyTls: env.CANVUS_VERIFY_TLS,
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
  });

  if (!parsed.success) {
    throw new ValidationError(
      parsed.error.issues.map((i) => ({ path: i.path, message: i.message })),
      "invalid SessionOptions",
    );
  }
  return Object.freeze(parsed.data);
}
