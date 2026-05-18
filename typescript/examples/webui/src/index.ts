/**
 * @mt-canvus-tools/webui-example — Hono/TS rewrite of the legacy
 * CanvusWebUI Express demo. See README.md for context.
 *
 * Public surface re-exports below let downstream tests + future PowerToys
 * ports pull strongly-typed handles to the route modules without
 * spinning up an HTTP server.
 */

import { Hono } from "hono";
import { serve, type ServerType } from "@hono/node-server";
import { serveStatic } from "@hono/node-server/serve-static";
import { existsSync } from "node:fs";
import { readFileSync } from "node:fs";
import { createSecureServer } from "node:http2";
import { fileURLToPath } from "node:url";
import path from "node:path";
import { loadConfig, type Config, type MutableConfig } from "./config.js";
import { createLogger } from "./logging.js";
import { buildCanvusSession } from "./session.js";
import {
  buildUserStore,
  buildDeletedRecordStore,
  type JsonStore,
  type UsersFile,
  type DeletedRecord,
} from "./lib/store.js";
import type { AppEnv, AppVariables } from "./app-env.js";

import { usersRoutes } from "./routes/users.js";
import { uploadsRoutes } from "./routes/uploads.js";
import { macrosRoutes } from "./routes/macros.js";
import { zonesRoutes } from "./routes/zones.js";
import { pagesRoutes } from "./routes/pages.js";
import { rcuRoutes, RcuStore } from "./routes/rcu.js";
import { auditRoutes } from "./routes/audit.js";
import { adminRoutes } from "./routes/admin.js";

export type { AppEnv, AppVariables };
export { loadConfig };
export type { Config, MutableConfig };
export type { JsonStore, UsersFile, DeletedRecord };

/** Dependencies the route registrar needs. Built in `start()` for prod, or by tests for mocking. */
export interface AppDeps {
  readonly config: Config;
  readonly mutableConfig: MutableConfig;
  readonly variables: Omit<AppVariables, "config" | "mutableConfig">;
}

/**
 * Build a fully-wired Hono app. The returned app has all routes
 * mounted and a request-scoped middleware that injects the AppDeps
 * into `c.var`. Tests can hand-construct deps and call `app.request()`
 * without ever opening a port.
 */
export function buildApp(deps: AppDeps): Hono<AppEnv> {
  const app = new Hono<AppEnv>();

  // Inject deps onto every request.
  app.use("*", async (c, next) => {
    c.set("session", deps.variables.session);
    c.set("logger", deps.variables.logger);
    c.set("config", deps.config);
    c.set("mutableConfig", deps.mutableConfig);
    c.set("userStore", deps.variables.userStore);
    c.set("deletedRecordStore", deps.variables.deletedRecordStore);
    c.set("rcuStore", deps.variables.rcuStore);
    await next();
  });

  // Route modules — order doesn't matter beyond mount-prefix, since
  // each module owns its own path namespace.
  app.route("/", usersRoutes());
  app.route("/", uploadsRoutes());
  app.route("/", zonesRoutes());
  app.route("/", pagesRoutes());
  app.route("/", macrosRoutes());
  app.route("/", auditRoutes());
  app.route("/", rcuRoutes());
  app.route("/", adminRoutes());

  return app;
}

/** Resolve the static asset dir relative to this file. */
function resolveStaticDir(): string {
  const here = path.dirname(fileURLToPath(import.meta.url));
  // src/ → ../public (dev), dist/ → ../public (prod), both correct
  return path.resolve(here, "../public");
}

/**
 * Boot the WebUI: load config, build dependencies, wire the app,
 * mount static assets, start the server.
 *
 * @returns The active `ServerType` so callers (or tests) can `.close()`.
 */
export async function start(env: NodeJS.ProcessEnv = process.env): Promise<ServerType> {
  const config = loadConfig(env);
  const logger = createLogger({
    level: config.LOG_LEVEL,
    format: config.LOG_FORMAT,
    component: "webui",
  });

  const mutableConfig: MutableConfig = {
    canvasId: config.CANVUS_CANVAS_ID,
    ...(config.CANVUS_CANVAS_NAME !== undefined && { canvasName: config.CANVUS_CANVAS_NAME }),
  };

  const session = buildCanvusSession(config);
  const userStore = buildUserStore(config.STATE_DIR);
  const deletedRecordStore = buildDeletedRecordStore(config.STATE_DIR);
  const rcuStore = new RcuStore();

  const app = buildApp({
    config,
    mutableConfig,
    variables: {
      session,
      logger,
      userStore,
      deletedRecordStore,
      rcuStore,
    },
  });

  // Static frontend assets.
  const staticDir = resolveStaticDir();
  if (existsSync(staticDir)) {
    app.use("/*", serveStatic({ root: path.relative(process.cwd(), staticDir) || "./" }));
  } else {
    logger.warn({ staticDir }, "static asset dir not present; only API routes will respond");
  }

  // HTTPS if both cert + key are configured, else HTTP.
  const hasSsl = Boolean(config.SSL_CERT_PATH && config.SSL_KEY_PATH);
  const server = serve(
    {
      fetch: app.fetch,
      port: config.PORT,
      hostname: config.HOSTNAME,
      ...(hasSsl && {
        // @hono/node-server supports passing a pre-built createSecureServer instance.
        createServer: createSecureServer,
        serverOptions: {
          cert: readFileSync(config.SSL_CERT_PATH ?? ""),
          key: readFileSync(config.SSL_KEY_PATH ?? ""),
          allowHTTP1: true,
        },
      }),
    },
    (info) => {
      logger.info(
        { port: info.port, address: info.address, ssl: hasSsl },
        "webui listening",
      );
    },
  );

  const shutdown = async (signal: string): Promise<void> => {
    logger.info({ signal }, "shutting down");
    await new Promise<void>((resolve) => {
      server.close(() => resolve());
    });
    await session.close();
    process.exit(0);
  };
  process.on("SIGINT", () => void shutdown("SIGINT"));
  process.on("SIGTERM", () => void shutdown("SIGTERM"));

  return server;
}

// Module-execution entry point. Skipped when imported by tests.
const invokedDirectly = process.argv[1] === fileURLToPath(import.meta.url);
if (invokedDirectly) {
  start().catch((err: unknown) => {
    // eslint-disable-next-line no-console
    console.error("[webui] failed to start:", err);
    process.exit(1);
  });
}
