/**
 * RCU admin-surface tests. The RCU endpoints are this WebUI's own
 * admin surface (NOT Canvus-server endpoints) — confirmed with Jaypaul
 * 2026-05-18. We verify config round-trip + auth gate.
 *
 * Each test constructs its own `RcuStore` via `buildFakeDeps`, so the
 * suite is parallelisable and has no shared mutable state.
 */

import { describe, expect, it, vi } from "vitest";
import { buildApp } from "../src/index.js";
import { buildFakeDeps } from "./helpers.js";

const AUTH = { authorization: "Bearer admin-pwd" } as const;

describe("rcu routes", () => {
  it("requires bearer token", async () => {
    const deps = buildFakeDeps({ session: {} });
    const app = buildApp(deps);
    const res = await app.request("/api/v1/canvases/cx/rcu/config");
    expect(res.status).toBe(401);
  });

  it("round-trips config without echoing the password", async () => {
    const deps = buildFakeDeps({ session: {} });
    const app = buildApp(deps);
    const put = await app.request("/api/v1/canvases/cx/rcu/config", {
      method: "POST",
      headers: { ...AUTH, "content-type": "application/json" },
      body: JSON.stringify({
        enabled: true,
        endpoint: "mqtt://broker:1883",
        topic: "rcu/test",
        qos: 2,
        username: "alice",
        password: "secret",
      }),
    });
    expect(put.status).toBe(200);
    const body = (await put.json()) as { config: Record<string, unknown> };
    expect(body.config.password).toBeUndefined();
    expect(body.config.username).toBe("alice");

    const get = await app.request("/api/v1/canvases/cx/rcu/config", { headers: AUTH });
    const got = (await get.json()) as { config: Record<string, unknown> };
    expect(got.config.endpoint).toBe("mqtt://broker:1883");
  });

  it("test endpoint surfaces the underlying SDK probe", async () => {
    const session = { server: { info: vi.fn().mockResolvedValue({ ok: true }) } };
    const deps = buildFakeDeps({ session });
    const app = buildApp(deps);
    const res = await app.request("/api/v1/canvases/cx/rcu/test", {
      method: "POST",
      headers: AUTH,
    });
    expect(res.status).toBe(200);
    expect(session.server.info).toHaveBeenCalled();
  });

  it("test endpoint reports failure when SDK probe throws", async () => {
    const session = { server: { info: vi.fn().mockRejectedValue(new Error("down")) } };
    const deps = buildFakeDeps({ session });
    const app = buildApp(deps);
    const res = await app.request("/api/v1/canvases/cx/rcu/test", {
      method: "POST",
      headers: AUTH,
    });
    expect(res.status).toBe(502);
  });
});
