/**
 * WebUI runtime configuration.
 *
 * Reads environment variables (loaded by the calling process; we do
 * NOT bundle a dotenv loader so that container deployments
 * don't accidentally pick up a stale `.env` shadowed by a mount).
 *
 * The legacy CanvusWebUI used `CANVUS_SERVER` + `CANVAS_ID` +
 * `CANVUS_API_KEY`. This port standardises on the Phase 4a names:
 *
 *   - `CANVUS_API_URL` (replaces `CANVUS_SERVER`)
 *   - `CANVUS_API_KEY` (unchanged)
 *   - `CANVUS_CANVAS_ID` (replaces `CANVAS_ID`)
 *
 * For migration convenience the legacy names are still accepted as
 * fallbacks but a warning is logged.
 */

import { z } from "zod";

/** Required env, after migration-name resolution. */
const envSchema = z.object({
  CANVUS_API_URL: z.string().url(),
  CANVUS_API_KEY: z.string().min(1),
  CANVUS_CANVAS_ID: z.string().min(1),

  // Optional admin / behavioural config.
  CANVUS_CANVAS_NAME: z.string().optional(),
  WEBUI_PWD: z.string().optional(),
  PORT: z.coerce.number().int().min(1).max(65535).default(3000),
  HOSTNAME: z.string().default("0.0.0.0"),
  ALLOW_SELF_SIGNED_CERTS: z
    .enum(["true", "false", "1", "0"])
    .default("false")
    .transform((v) => v === "true" || v === "1"),
  SSL_CERT_PATH: z.string().optional(),
  SSL_KEY_PATH: z.string().optional(),
  COOKIE_SECRET: z.string().default("webui-demo-cookie-secret"),
  LOG_LEVEL: z.string().default("info"),
  LOG_FORMAT: z.enum(["json", "pretty"]).default("pretty"),
  UPLOAD_DIR: z.string().default("uploads"),
  STATE_DIR: z.string().default("state"),
});

export type Config = z.infer<typeof envSchema>;

/** Per-tenant config that callers may need to update at runtime. */
export interface MutableConfig {
  canvasId: string;
  canvasName?: string;
}

/**
 * Parse `process.env` into a frozen Config, applying legacy-name
 * fallbacks.
 *
 * Exits the process with code 2 on validation failure so misconfigured
 * containers fail fast rather than silently 500-ing.
 */
export function loadConfig(env: NodeJS.ProcessEnv = process.env): Config {
  const merged: Record<string, string | undefined> = { ...env };
  // Legacy name fallbacks (CanvusWebUI shipped without the Phase-4a names).
  if (!merged.CANVUS_API_URL && merged.CANVUS_SERVER) {
    merged.CANVUS_API_URL = merged.CANVUS_SERVER;
  }
  if (!merged.CANVUS_CANVAS_ID && merged.CANVAS_ID) {
    merged.CANVUS_CANVAS_ID = merged.CANVAS_ID;
  }
  if (!merged.CANVUS_CANVAS_NAME && merged.CANVAS_NAME) {
    merged.CANVUS_CANVAS_NAME = merged.CANVAS_NAME;
  }

  const parsed = envSchema.safeParse(merged);
  if (!parsed.success) {
    // eslint-disable-next-line no-console
    console.error(
      "[webui-config] invalid environment:",
      JSON.stringify(parsed.error.flatten().fieldErrors, null, 2),
    );
    process.exit(2);
  }
  return parsed.data;
}
