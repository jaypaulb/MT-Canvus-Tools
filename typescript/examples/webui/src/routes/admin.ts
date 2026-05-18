/**
 * Admin routes — port of legacy `/validateAdmin`,
 * `/admin/env-variables`, `/admin/update-env`,
 * `/admin/createTargets`, `/admin/deleteTargets`,
 * `/admin/listUsers`, `/admin/deleteUsers`.
 *
 * Auth model preserved: Bearer-token header matching `WEBUI_PWD`. This
 * is demo-grade; production deployments should put the WebUI behind a
 * real IdP.
 *
 * The env-vars surface is intentionally read-only against
 * `process.env` (we removed the legacy on-disk `.env` rewriting — too
 * fragile and surprising in container deployments). Persistent
 * canvas switching happens via the `/api/canvas/switch` route in
 * `pages.ts` (mutates `MutableConfig`).
 */

import { Hono } from "hono";
import { zValidator } from "@hono/zod-validator";
import { z } from "zod";
import { createMiddleware } from "hono/factory";
import type { AppEnv } from "../app-env.js";
import { getTeamBaseColor, type TeamNumber } from "../lib/colors.js";

const EXCLUDED_KEY_FRAGMENTS = ["api_key", "apikey", "allow_self_signed_certs"];

function adminAuth() {
  return createMiddleware<AppEnv>(async (c, next) => {
    const expected = c.var.config.WEBUI_PWD;
    if (!expected) {
      return c.json(
        { success: false, message: "Server configuration error: WEBUI_PWD not set." },
        500,
      );
    }
    const header = c.req.header("authorization");
    if (!header || !header.startsWith("Bearer ")) {
      return c.json(
        { success: false, message: "Missing or malformed authorization header." },
        401,
      );
    }
    const token = header.slice("Bearer ".length);
    if (token !== expected) {
      return c.json({ success: false, message: "Invalid authorization token." }, 401);
    }
    await next();
  });
}

const updateEnvBody = z.record(z.string(), z.string());

export function adminRoutes(): Hono<AppEnv> {
  const app = new Hono<AppEnv>();

  app.use("*", adminAuth());

  app.post("/validateAdmin", (c) => c.json({ success: true, message: "Authentication successful" }));

  app.get("/admin/env-variables", (c) => {
    const safe: Record<string, string | undefined> = {};
    for (const [key, value] of Object.entries(c.var.config)) {
      const lower = key.toLowerCase();
      if (EXCLUDED_KEY_FRAGMENTS.some((p) => lower.includes(p))) continue;
      if (typeof value === "string" || value === undefined) safe[key] = value;
      else if (typeof value === "number" || typeof value === "boolean") safe[key] = String(value);
    }
    // Also surface the live canvas id (it can be swapped via /api/canvas/switch).
    safe.CANVUS_CANVAS_ID = c.var.mutableConfig.canvasId;
    if (c.var.mutableConfig.canvasName !== undefined) {
      safe.CANVUS_CANVAS_NAME = c.var.mutableConfig.canvasName;
    }
    return c.json(safe);
  });

  /**
   * Update env vars at runtime — limited to switching canvas ID and
   * its display name. Other vars require a restart by design (env
   * mutation is a container anti-pattern).
   */
  app.post("/admin/update-env", zValidator("json", updateEnvBody), async (c) => {
    const updates = c.req.valid("json");
    const restricted = Object.keys(updates).filter((k) =>
      EXCLUDED_KEY_FRAGMENTS.some((p) => k.toLowerCase().includes(p)),
    );
    if (restricted.length > 0) {
      return c.json(
        {
          success: false,
          error: `Cannot modify restricted variables: ${restricted.join(", ")}`,
        },
        403,
      );
    }
    const cfg = c.var.mutableConfig as { canvasId: string; canvasName?: string };
    const acceptedKeys = ["CANVUS_CANVAS_ID", "CANVAS_ID", "CANVUS_CANVAS_NAME", "CANVAS_NAME"];
    const accepted: Record<string, string> = {};
    for (const [k, v] of Object.entries(updates)) {
      if (!acceptedKeys.includes(k)) continue;
      accepted[k] = v;
    }
    if (Object.keys(accepted).length === 0) {
      return c.json(
        {
          success: false,
          error: "No accepted variables provided. Only CANVUS_CANVAS_ID / CANVUS_CANVAS_NAME may be updated at runtime.",
        },
        400,
      );
    }

    const newId = accepted.CANVUS_CANVAS_ID ?? accepted.CANVAS_ID;
    if (newId) {
      try {
        const canvas = await c.var.session.canvases.get(newId);
        cfg.canvasId = newId;
        cfg.canvasName = canvas.name;
        return c.json({
          success: true,
          message: "Canvas updated.",
          updatedVars: { CANVUS_CANVAS_ID: newId, CANVUS_CANVAS_NAME: canvas.name },
        });
      } catch (err) {
        return c.json(
          { success: false, error: `Canvas ID does not exist: ${(err as Error).message}` },
          400,
        );
      }
    }
    const newName = accepted.CANVUS_CANVAS_NAME ?? accepted.CANVAS_NAME;
    if (newName) cfg.canvasName = newName;
    return c.json({ success: true, message: "Canvas name updated.", updatedVars: accepted });
  });

  app.post("/admin/createTargets", async (c) => {
    const session = c.var.session;
    const canvasId = c.var.mutableConfig.canvasId;
    try {
      const zones = await session.widgets.anchors.list(canvasId);
      if (zones.length < 8) {
        return c.json({ success: false, message: "Not enough zones available" }, 400);
      }
      const created: number[] = [];
      for (let team = 1 as TeamNumber; team <= 7; team++) {
        const zone = zones.find(
          (z) => (z as unknown as { anchor_index?: number }).anchor_index === team,
        );
        const loc = (zone as unknown as { location?: { x: number; y: number } } | undefined)
          ?.location;
        if (!loc) continue;
        try {
          await session.widgets.notes.create(canvasId, {
            auto_text_color: true,
            background_color: getTeamBaseColor(team as TeamNumber),
            depth: 100,
            location: { x: loc.x + 30, y: loc.y + 30 },
            pinned: false,
            scale: 1,
            size: { width: 300, height: 300 },
            state: "normal",
            text: `Team ${team.toString()}`,
            title: `Team_${team.toString()}_Target`,
          } as Parameters<typeof session.widgets.notes.create>[1]);
          created.push(team);
        } catch (err) {
          c.var.logger.warn(
            { team, err: (err as Error).message },
            "admin/createTargets: target failed",
          );
        }
      }
      if (created.length === 0) {
        return c.json({ success: false, message: "Failed to create any team targets." });
      }
      return c.json({ success: true, message: `Created Team Targets ${created.join(", ")}` });
    } catch (err) {
      c.var.logger.error({ err: (err as Error).message }, "admin/createTargets failed");
      return c.json({ success: false, message: "Failed to create target notes" }, 500);
    }
  });

  app.post("/admin/deleteTargets", async (c) => {
    const session = c.var.session;
    const canvasId = c.var.mutableConfig.canvasId;
    try {
      const widgets = await session.widgets.list(canvasId);
      const targets = widgets.filter((w) => {
        const title = (w as unknown as { title?: string }).title;
        return w.widget_type === "Note" && title && /^Team_\d+_Target$/.test(title);
      });
      let deleted = 0;
      for (const t of targets) {
        try {
          await session.widgets.notes.delete(canvasId, t.id);
          deleted++;
        } catch (err) {
          c.var.logger.warn({ id: t.id, err: (err as Error).message }, "delete target failed");
        }
      }
      return c.json({ success: true, message: `Deleted ${deleted.toString()} target notes` });
    } catch (err) {
      c.var.logger.error({ err: (err as Error).message }, "admin/deleteTargets failed");
      return c.json({ success: false, message: "Failed to delete target notes" }, 500);
    }
  });

  app.get("/admin/listUsers", async (c) => {
    const usersObj = await c.var.userStore.read();
    const users: Array<{ team: number; name: string; color: string }> = [];
    for (const [team, members] of Object.entries(usersObj)) {
      for (const [name, color] of Object.entries(members)) {
        users.push({ team: Number.parseInt(team, 10), name, color });
      }
    }
    return c.json({ success: true, users });
  });

  app.post("/admin/deleteUsers", async (c) => {
    await c.var.userStore.write({});
    return c.json({ success: true, message: "All users deleted successfully" });
  });

  return app;
}
