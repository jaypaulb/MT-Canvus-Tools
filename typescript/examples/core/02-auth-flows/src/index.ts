/**
 * Example 02 — auth-flows.
 *
 * Demonstrates three ways to authenticate against the Canvus API:
 *   1. API-key (Private-Token header) — the recommended flow for
 *      automation.
 *   2. Email + password login (POST /users/login) — exchanges credentials
 *      for a session token.
 *   3. Access-token CRUD — create, list, and revoke a programmatic API
 *      token tied to the logged-in user.
 *
 * Each path logs success/failure and runs a trivial authenticated call.
 * A partial auth setup still demonstrates the other modes; the program
 * proceeds even if one path fails.
 *
 * Per VERIFIED-CORRECTIONS.md §4, the login endpoint uses `email` ONLY.
 * Including a `username` field causes the server to reject the request.
 */

import {
  createSession,
  isCanvusError,
  type Session,
} from "@mt-canvus-tools/sdk";
import pino from "pino";
import { z } from "zod";

const isPretty = process.env["LOG_FORMAT"] === "pretty";

const logger = pino({
  level: process.env["LOG_LEVEL"] ?? "info",
  base: { component: "example-auth-flows" },
  ...(isPretty
    ? {
        transport: {
          target: "pino-pretty",
          options: { colorize: true, translateTime: "SYS:HH:MM:ss.l" },
        },
      }
    : {}),
});

const envSchema = z.object({
  CANVUS_API_URL: z.string().url(),
  CANVUS_API_KEY: z.string().min(1),
  CANVUS_EMAIL: z.string().email().optional(),
  CANVUS_PASSWORD: z.string().min(1).optional(),
});

type Env = z.infer<typeof envSchema>;

function loadEnv(): Env {
  const parsed = envSchema.safeParse(process.env);
  if (!parsed.success) {
    logger.error(
      { fieldErrors: parsed.error.flatten().fieldErrors },
      "missing or invalid env",
    );
    process.exit(2);
  }
  return parsed.data;
}

/**
 * Mode 1: API key in the `Private-Token` header.
 *
 * The simplest and most common flow. The SDK injects the header
 * automatically when `apiKey` is set on the session config.
 */
async function runApiKey(env: Env): Promise<void> {
  logger.info({ mode: "api-key" }, "trying API-key auth");
  const session = createSession({
    baseUrl: env.CANVUS_API_URL,
    apiKey: env.CANVUS_API_KEY,
  });
  const canvases = await session.canvases.list();
  logger.info(
    { mode: "api-key", canvasCount: canvases.length },
    "API-key auth succeeded",
  );
}

/**
 * Mode 2: email + password login.
 *
 * `POST /users/login` exchanges credentials for a session token, which
 * the SDK then sends as `Private-Token` on subsequent requests.
 */
/**
 * Result of a successful login: the authenticated session plus the
 * logged-in user's id (needed for access-token CRUD).
 *
 * Note: `User.id` is an INTEGER on the live server (per
 * VERIFIED-CORRECTIONS.md §6); we keep it as a number for type fidelity.
 */
interface LoggedIn {
  readonly session: Session;
  readonly userId: number;
}

async function runLogin(env: Env): Promise<LoggedIn | undefined> {
  if (!env.CANVUS_EMAIL || !env.CANVUS_PASSWORD) {
    logger.warn(
      { mode: "login" },
      "skipping login flow (CANVUS_EMAIL and/or CANVUS_PASSWORD not set)",
    );
    return undefined;
  }

  logger.info({ mode: "login", email: env.CANVUS_EMAIL }, "trying login auth");

  // Step 1: unauthenticated session to call /users/login.
  const anonymous = createSession({ baseUrl: env.CANVUS_API_URL });
  const login = await anonymous.auth.login({
    // VERIFIED-CORRECTIONS.md §4: send `email` ONLY.
    email: env.CANVUS_EMAIL,
    password: env.CANVUS_PASSWORD,
  });

  // Step 2: build a new session using the returned session token.
  const session = createSession({
    baseUrl: env.CANVUS_API_URL,
    apiKey: login.token,
  });
  const userId = login.user.id;
  const me = await session.users.get(userId);
  logger.info(
    {
      mode: "login",
      userId: me.id,
      email: me.email,
      isAdmin: me.admin,
    },
    "login auth succeeded",
  );
  return { session, userId };
}

/**
 * Mode 3: programmatic access-token lifecycle.
 *
 * Mints a new access token via the logged-in session, demonstrates it,
 * then revokes it. The minted token is the only time the secret is
 * returned by the server.
 */
async function runTokenLifecycle(loggedIn: LoggedIn | undefined): Promise<void> {
  if (!loggedIn) {
    logger.warn(
      { mode: "token" },
      "skipping access-token lifecycle (no logged-in session available)",
    );
    return;
  }

  logger.info({ mode: "token", userId: loggedIn.userId }, "trying access-token CRUD");

  const { session, userId } = loggedIn;
  const tokenName = `auth-flows-demo-${Date.now().toString()}`;

  // Create.
  const created = await session.auth.createAccessToken(userId, {
    description: tokenName,
  });
  logger.info(
    { mode: "token", tokenId: created.id, description: created.description },
    "minted access token",
  );

  // List.
  const tokens = await session.auth.listAccessTokens(userId);
  logger.info(
    { mode: "token", visibleTokenCount: tokens.length },
    "listed access tokens",
  );

  // Revoke (cleanup).
  await session.auth.deleteAccessToken(userId, created.id);
  logger.info(
    { mode: "token", tokenId: created.id },
    "revoked access token",
  );
}

async function run(): Promise<void> {
  const env = loadEnv();

  // Run each mode independently so a failure in one does not block the
  // others. Each helper logs its own success/failure.
  try {
    await runApiKey(env);
  } catch (err: unknown) {
    logFailure("api-key", err);
  }

  let loggedIn: LoggedIn | undefined;
  try {
    loggedIn = await runLogin(env);
  } catch (err: unknown) {
    logFailure("login", err);
  }

  try {
    await runTokenLifecycle(loggedIn);
  } catch (err: unknown) {
    logFailure("token", err);
  }
}

function logFailure(mode: string, err: unknown): void {
  if (isCanvusError(err)) {
    logger.error({ mode, kind: err.kind, err: err.message }, "auth mode failed");
  } else {
    logger.error({ mode, err }, "auth mode failed (non-SDK error)");
  }
}

run().catch((err: unknown) => {
  logger.error({ err }, "auth-flows crashed");
  process.exit(1);
});
