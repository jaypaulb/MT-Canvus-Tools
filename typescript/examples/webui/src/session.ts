/**
 * Single shared Canvus SDK session for the WebUI process.
 *
 * The legacy server.js built one `axios` instance at boot and used it
 * for every Canvus call. The TS rewrite does the same with a single
 * `@mt-canvus-tools/sdk` Session: cheap to construct, shared across
 * route handlers, and disposed on shutdown.
 *
 * If we ever introduce per-user API keys (Phase 4d candidate), this
 * single-session model becomes a `getSessionFor(req)` map keyed by the
 * caller's bearer token. For now: one process, one Canvus key,
 * one session.
 */

import { createSession, type Session } from "@mt-canvus-tools/sdk";
import type { Config } from "./config.js";

export function buildCanvusSession(config: Config): Session {
  return createSession({
    baseUrl: config.CANVUS_API_URL,
    apiKey: config.CANVUS_API_KEY,
    verifyTls: !config.ALLOW_SELF_SIGNED_CERTS,
  });
}
