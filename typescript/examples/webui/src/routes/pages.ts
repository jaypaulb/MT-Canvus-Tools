/**
 * Page-sequence + client/workspace routes — port of legacy
 * `/api/clients`, `/api/clients/:id/workspace`,
 * `/api/clients/:id/workspace/subscribe` (SSE), `/api/canvas/switch`,
 * `/verifycanvas/:id`, `/get-canvas-info`.
 *
 * The SSE subscribe stream is rewritten to use the typed
 * `session.server.subscribeWorkspace(clientId, "0")` helper instead of
 * the legacy 2-second polling loop — a Phase 4b upgrade noted in the
 * audit (Cross-item observation #2).
 */

import { Hono } from "hono";
import { zValidator } from "@hono/zod-validator";
import { z } from "zod";
import { streamSSE } from "hono/streaming";
import type { AppEnv } from "../app-env.js";

const switchBody = z.object({
  canvas_id: z.string().min(1),
  canvas_name: z.string().optional(),
});

export function pagesRoutes(): Hono<AppEnv> {
  const app = new Hono<AppEnv>();

  app.get("/get-canvas-info", (c) => {
    const name = c.var.mutableConfig.canvasName;
    if (!name) {
      return c.json({ error: "Canvas name not found." }, 404);
    }
    return c.json({ canvas_name: name });
  });

  app.get("/verifycanvas/:canvasId", async (c) => {
    const canvasId = c.req.param("canvasId");
    try {
      const canvas = await c.var.session.canvases.get(canvasId);
      return c.json({ exists: true, id: canvas.id, name: canvas.name });
    } catch (err) {
      return c.json({ exists: false, error: (err as Error).message }, 200);
    }
  });

  app.get("/api/clients", async (c) => {
    try {
      const clients = await c.var.session.server.clients();
      return c.json({ success: true, clients });
    } catch (err) {
      return c.json({ success: false, error: (err as Error).message }, 500);
    }
  });

  app.get("/api/clients/:clientId/workspace", async (c) => {
    const clientId = c.req.param("clientId");
    try {
      // Workspace 0 is the canonical "primary" workspace per legacy convention.
      const workspace = await c.var.session.server.workspace(clientId, "0");
      return c.json({ success: true, workspace });
    } catch (err) {
      return c.json({ success: false, error: (err as Error).message }, 500);
    }
  });

  // SSE subscription powered by the SDK's typed Subscribe helper (no polling).
  app.get("/api/clients/:clientId/workspace/subscribe", (c) => {
    const clientId = c.req.param("clientId");
    return streamSSE(c, async (stream) => {
      await stream.writeSSE({
        data: JSON.stringify({ type: "connected", clientId }),
      });
      let lastCanvasId: string | undefined;
      try {
        for await (const workspace of c.var.session.server.subscribeWorkspace(clientId, "0")) {
          const currentCanvasId = (workspace as unknown as { canvas_id?: string }).canvas_id;
          if (currentCanvasId && currentCanvasId !== lastCanvasId) {
            lastCanvasId = currentCanvasId;
            let canvasName = "Unknown";
            try {
              const canvas = await c.var.session.canvases.get(currentCanvasId);
              canvasName = canvas.name;
            } catch {
              // ignore — return Unknown
            }
            await stream.writeSSE({
              data: JSON.stringify({
                type: "canvas_change",
                canvas_id: currentCanvasId,
                canvas_name: canvasName,
              }),
            });
          }
        }
      } catch (err) {
        await stream.writeSSE({
          data: JSON.stringify({
            type: "error",
            message: (err as Error).message,
          }),
        });
      }
    });
  });

  app.post("/api/canvas/switch", zValidator("json", switchBody), (c) => {
    const body = c.req.valid("json");
    // MutableConfig is intentionally a non-readonly object so the
    // server can swap canvases without restarting.
    const cfg = c.var.mutableConfig as { canvasId: string; canvasName?: string };
    cfg.canvasId = body.canvas_id;
    if (body.canvas_name !== undefined) cfg.canvasName = body.canvas_name;
    return c.json({
      success: true,
      canvas_id: body.canvas_id,
      ...(body.canvas_name !== undefined && { canvas_name: body.canvas_name }),
    });
  });

  return app;
}
