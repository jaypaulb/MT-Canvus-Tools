/**
 * Admin-route tests. Mainly verify the Bearer-token gate and the
 * env-vars filter (no API keys leaked).
 */

import { describe, expect, it, vi } from "vitest";
import { buildApp } from "../src/index.js";
import { buildFakeDeps } from "./helpers.js";

describe("admin routes", () => {
  it("rejects requests without bearer token", async () => {
    const deps = buildFakeDeps({ session: {} });
    const app = buildApp(deps);
    const res = await app.request("/validateAdmin", { method: "POST" });
    expect(res.status).toBe(401);
  });

  it("accepts requests with the configured WEBUI_PWD bearer token", async () => {
    const deps = buildFakeDeps({ session: {} });
    const app = buildApp(deps);
    const res = await app.request("/validateAdmin", {
      method: "POST",
      headers: { authorization: "Bearer admin-pwd" },
    });
    expect(res.status).toBe(200);
  });

  it("/admin/env-variables omits API keys", async () => {
    const deps = buildFakeDeps({ session: {} });
    const app = buildApp(deps);
    const res = await app.request("/admin/env-variables", {
      headers: { authorization: "Bearer admin-pwd" },
    });
    expect(res.status).toBe(200);
    const body = (await res.json()) as Record<string, string>;
    expect(body.CANVUS_API_KEY).toBeUndefined();
    expect(body.CANVUS_CANVAS_ID).toBe("canvas-id-1");
  });

  it("/admin/update-env rejects restricted keys", async () => {
    const deps = buildFakeDeps({ session: {} });
    const app = buildApp(deps);
    const res = await app.request("/admin/update-env", {
      method: "POST",
      headers: { authorization: "Bearer admin-pwd", "content-type": "application/json" },
      body: JSON.stringify({ CANVUS_API_KEY: "new" }),
    });
    expect(res.status).toBe(403);
  });

  it("/admin/update-env verifies + applies a new canvas id", async () => {
    const session = {
      widgets: {},
      canvases: { get: vi.fn().mockResolvedValue({ id: "new-canvas", name: "New" }) },
    };
    const deps = buildFakeDeps({ session });
    const app = buildApp(deps);
    const res = await app.request("/admin/update-env", {
      method: "POST",
      headers: { authorization: "Bearer admin-pwd", "content-type": "application/json" },
      body: JSON.stringify({ CANVUS_CANVAS_ID: "new-canvas" }),
    });
    expect(res.status).toBe(200);
    expect(deps.mutableConfig.canvasId).toBe("new-canvas");
    expect(deps.mutableConfig.canvasName).toBe("New");
  });
});
