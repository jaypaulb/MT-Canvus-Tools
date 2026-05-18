/**
 * RCU (admin remote-control unit) endpoints.
 *
 * Per Jaypaul's 2026-05-18 clarification: these `/api/v1/canvases/{id}/rcu/*`
 * paths are NOT canvus-server endpoints. They are this WebUI's own
 * admin surface, provided so that consumers like the PowerToys port
 * (Round 3) can keep their client-side RCU calls pointed at an embedded
 * WebUI instance.
 *
 * The legacy CanvusWebUI shipped an RCU.html page that turned out to
 * mean "Remote Content Upload" (a user-facing upload form), NOT admin
 * RCU. So the admin surface below is greenfield and intentionally
 * minimal: in-memory config, last-test result, and a probe that pings
 * the configured Canvus server. The shape matches PowerToys'
 * `rcu_handler.go` expectations.
 *
 * Config is admin-scoped (`/api/v1/canvases/{id}/rcu/*` routes share
 * the same Bearer-token auth as the other `/admin` endpoints).
 */

import { Hono } from "hono";
import { zValidator } from "@hono/zod-validator";
import { createMiddleware } from "hono/factory";
import { z } from "zod";
import type { AppEnv } from "../app-env.js";

export interface RcuConfig {
  enabled: boolean;
  endpoint: string;
  topic: string;
  qos: 0 | 1 | 2;
  username?: string;
  // password intentionally omitted from GET responses
}

export interface RcuRuntime {
  config: RcuConfig;
  lastTestAt?: string;
  lastTestOk?: boolean;
  lastTestError?: string;
}

/**
 * In-memory RCU store. Demo-grade — persists for the lifetime of the
 * holding object only. Constructed by `buildApp(...)` and injected via
 * AppDeps so each test gets an isolated store and parallel test runs
 * don't trample each other. (No module-scoped mutable state.)
 */
export class RcuStore {
  private state: RcuRuntime = {
    config: {
      enabled: false,
      endpoint: "",
      topic: "canvus/rcu",
      qos: 1,
    },
  };

  get(): RcuRuntime {
    return this.state;
  }

  setConfig(c: RcuConfig): void {
    this.state.config = c;
  }

  recordTest(at: string, ok: boolean, error?: string): void {
    this.state.lastTestAt = at;
    this.state.lastTestOk = ok;
    if (error !== undefined) {
      this.state.lastTestError = error;
    } else {
      delete this.state.lastTestError;
    }
  }
}

const configBody = z.object({
  enabled: z.boolean(),
  endpoint: z.string(),
  topic: z.string().default("canvus/rcu"),
  qos: z.union([z.literal(0), z.literal(1), z.literal(2)]),
  username: z.string().optional(),
  password: z.string().optional(),
});

function rcuAuth() {
  return createMiddleware<AppEnv>(async (c, next) => {
    const expected = c.var.config.WEBUI_PWD;
    if (!expected) {
      return c.json({ success: false, message: "WEBUI_PWD not configured." }, 500);
    }
    const header = c.req.header("authorization") ?? "";
    if (!header.startsWith("Bearer ")) {
      return c.json({ success: false, message: "Missing bearer token." }, 401);
    }
    if (header.slice("Bearer ".length) !== expected) {
      return c.json({ success: false, message: "Invalid bearer token." }, 401);
    }
    await next();
  });
}

export function rcuRoutes(): Hono<AppEnv> {
  const app = new Hono<AppEnv>();
  app.use("*", rcuAuth());

  // Note: the path param is accepted but not used — the legacy
  // PowerToys client constructs paths as
  // `/api/v1/canvases/{id}/rcu/...`, so we honour that template even
  // though we route by mutableConfig.canvasId internally.
  app.get("/api/v1/canvases/:canvasId/rcu/config", (c) => {
    return c.json({ success: true, config: c.var.rcuStore.get().config });
  });

  app.post(
    "/api/v1/canvases/:canvasId/rcu/config",
    zValidator("json", configBody),
    (c) => {
      const body = c.req.valid("json");
      c.var.rcuStore.setConfig({
        enabled: body.enabled,
        endpoint: body.endpoint,
        topic: body.topic,
        qos: body.qos,
        ...(body.username !== undefined && { username: body.username }),
      });
      return c.json({ success: true, config: c.var.rcuStore.get().config });
    },
  );

  app.get("/api/v1/canvases/:canvasId/rcu/status", (c) => {
    const rt = c.var.rcuStore.get();
    return c.json({
      success: true,
      enabled: rt.config.enabled,
      ...(rt.lastTestAt !== undefined && { lastTestAt: rt.lastTestAt }),
      ...(rt.lastTestOk !== undefined && { lastTestOk: rt.lastTestOk }),
      ...(rt.lastTestError !== undefined && { lastTestError: rt.lastTestError }),
    });
  });

  /** Probe the configured endpoint — currently just a Canvus reachability check. */
  app.post("/api/v1/canvases/:canvasId/rcu/test", async (c) => {
    const at = new Date().toISOString();
    try {
      await c.var.session.server.info();
      c.var.rcuStore.recordTest(at, true);
      return c.json({ success: true, message: "Canvus server reachable." });
    } catch (err) {
      const msg = (err as Error).message;
      c.var.rcuStore.recordTest(at, false, msg);
      return c.json({ success: false, error: msg }, 502);
    }
  });

  return app;
}
