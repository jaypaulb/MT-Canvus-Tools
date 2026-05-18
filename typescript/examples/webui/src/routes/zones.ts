/**
 * Zone CRUD routes — port of legacy `/get-zones`, `/create-zones`,
 * `/delete-zones`, `/create-team-targets`.
 *
 * Zones are simply Canvus `Anchor` widgets — the WebUI treats them as
 * first-class because the macros need them. Sub-zones get a name suffix
 * `(Script Made)` so the delete-zones helper can find them safely.
 */

import { Hono } from "hono";
import { zValidator } from "@hono/zod-validator";
import { z } from "zod";
import type { Anchor, Session, Widget } from "@mt-canvus-tools/sdk";
import type { AppEnv } from "../app-env.js";
import { getTeamBaseColor, type TeamNumber } from "../lib/colors.js";

const createZonesBody = z.object({
  gridSize: z.number().int().refine((n) => [1, 3, 4, 5].includes(n), {
    message: "gridSize must be 1, 3, 4, or 5",
  }),
  gridPattern: z.enum(["Z", "Snake", "Spiral"]).default("Z"),
  subZoneId: z.string().optional(),
  subZoneArray: z
    .string()
    .regex(/^\d+x\d+$/)
    .optional(),
});

// ---------- Pattern generators (ported verbatim from legacy) ----------

function zOrder(size: number): Array<{ row: number; col: number }> {
  const out: Array<{ row: number; col: number }> = [];
  for (let r = 0; r < size; r++) for (let c = 0; c < size; c++) out.push({ row: r, col: c });
  return out;
}

function snakeOrder(size: number): Array<{ row: number; col: number }> {
  const out: Array<{ row: number; col: number }> = [];
  for (let r = 0; r < size; r++) {
    const cols = Array.from({ length: size }, (_, i) => i);
    if (r % 2 !== 0) cols.reverse();
    for (const c of cols) out.push({ row: r, col: c });
  }
  return out;
}

function spiralOrder(size: number): Array<{ row: number; col: number }> {
  const out: Array<{ row: number; col: number }> = [];
  let x = Math.floor(size / 2);
  let y = Math.floor(size / 2);
  out.push({ row: y, col: x });
  let step = 1;
  while (out.length < size * size) {
    for (let i = 0; i < step; i++) {
      x += 1;
      if (x >= 0 && x < size && y >= 0 && y < size) out.push({ row: y, col: x });
    }
    for (let i = 0; i < step; i++) {
      y += 1;
      if (x >= 0 && x < size && y >= 0 && y < size) out.push({ row: y, col: x });
    }
    step++;
    for (let i = 0; i < step; i++) {
      x -= 1;
      if (x >= 0 && x < size && y >= 0 && y < size) out.push({ row: y, col: x });
    }
    for (let i = 0; i < step; i++) {
      y -= 1;
      if (x >= 0 && x < size && y >= 0 && y < size) out.push({ row: y, col: x });
    }
    step++;
  }
  return out;
}

/** Resolve a pattern-name → ordered grid coordinates. */
function orderFor(pattern: "Z" | "Snake" | "Spiral", size: number): Array<{ row: number; col: number }> {
  switch (pattern) {
    case "Snake":
      return snakeOrder(size);
    case "Spiral":
      return spiralOrder(size);
    default:
      return zOrder(size);
  }
}

// ---------- Helper: discover the SharedCanvas widget ----------

interface SharedCanvas {
  readonly widget_type: string;
  readonly size?: { width: number; height: number };
}

function findSharedCanvas(widgets: readonly Widget[]): SharedCanvas | undefined {
  for (const w of widgets) {
    const wt = (w as unknown as { widget_type?: string }).widget_type;
    if (wt === "SharedCanvas") return w as unknown as SharedCanvas;
  }
  return undefined;
}

async function createMainZones(
  session: Session,
  canvasId: string,
  gridSize: number,
  pattern: "Z" | "Snake" | "Spiral",
): Promise<{ created: number; failed: number }> {
  const widgets = await session.widgets.list(canvasId);
  const sharedCanvas = findSharedCanvas(widgets);
  if (!sharedCanvas?.size) {
    throw new Error("SharedCanvas widget not found or has no size.");
  }
  const { width: canvasW, height: canvasH } = sharedCanvas.size;
  const zoneW = canvasW / gridSize;
  const zoneH = canvasH / gridSize;
  const coords = orderFor(pattern, gridSize).map((c) => ({
    x: c.col * zoneW,
    y: c.row * zoneH,
  }));

  let created = 0;
  let failed = 0;
  let n = 1;
  for (const coord of coords) {
    const anchor_name = `${gridSize.toString()}x${gridSize.toString()} Zone ${n.toString()} (Script Made)`;
    try {
      await session.widgets.anchors.create(canvasId, {
        anchor_name,
        location: { x: coord.x, y: coord.y },
        size: { width: zoneW, height: zoneH },
        pinned: true,
        scale: 1,
        depth: 0,
      } as Parameters<typeof session.widgets.anchors.create>[1]);
      created++;
    } catch {
      failed++;
    }
    n++;
  }
  return { created, failed };
}

async function createSubZones(
  session: Session,
  canvasId: string,
  parentZoneId: string,
  cols: number,
  rows: number,
): Promise<{ created: number; failed: number }> {
  const parent = await session.widgets.anchors.get(canvasId, parentZoneId);
  const parentAny = parent as unknown as {
    location?: { x: number; y: number };
    size?: { width: number; height: number };
    anchor_name?: string;
  };
  if (!parentAny.location || !parentAny.size) {
    throw new Error(`Parent zone ${parentZoneId} missing location or size`);
  }
  const baseName = parentAny.anchor_name ?? "";
  // Try to extract a numeric chunk like " 12 " from the parent name; fall back to "X".
  const numMatch = baseName.match(/\s(\d+)\s/);
  const baseNum = numMatch ? numMatch[1] : "X";

  const zoneW = parentAny.size.width / cols;
  const zoneH = parentAny.size.height / rows;
  let created = 0;
  let failed = 0;
  for (let r = 0; r < rows; r++) {
    for (let cc = 0; cc < cols; cc++) {
      const idx = r * cols + cc + 1;
      const anchor_name = `SubZone ${baseNum}.${idx.toString()} (Script Made)`;
      try {
        await session.widgets.anchors.create(canvasId, {
          anchor_name,
          location: {
            x: parentAny.location.x + cc * zoneW,
            y: parentAny.location.y + r * zoneH,
          },
          size: { width: zoneW, height: zoneH },
          pinned: true,
          scale: 1,
          depth: 1,
        } as Parameters<typeof session.widgets.anchors.create>[1]);
        created++;
      } catch {
        failed++;
      }
    }
  }
  return { created, failed };
}

export function zonesRoutes(): Hono<AppEnv> {
  const app = new Hono<AppEnv>();

  app.get("/get-zones", async (c) => {
    try {
      const zones = await c.var.session.widgets.anchors.list(c.var.mutableConfig.canvasId);
      return c.json({ success: true, zones });
    } catch (err) {
      c.var.logger.error({ err: (err as Error).message }, "get-zones failed");
      return c.json({ success: false, error: (err as Error).message }, 500);
    }
  });

  app.post("/create-zones", zValidator("json", createZonesBody), async (c) => {
    const { gridSize, gridPattern, subZoneId, subZoneArray } = c.req.valid("json");
    const session = c.var.session;
    const canvasId = c.var.mutableConfig.canvasId;
    try {
      if (subZoneId && subZoneArray) {
        const parts = subZoneArray.split("x").map((s) => Number.parseInt(s, 10));
        const cols = parts[0];
        const rows = parts[1];
        if (cols === undefined || rows === undefined || cols <= 0 || rows <= 0) {
          return c.json({ success: false, error: "Invalid SubZone Array format." }, 400);
        }
        const result = await createSubZones(session, canvasId, subZoneId, cols, rows);
        return c.json({
          success: true,
          message: `${result.created.toString()} subzones created, ${result.failed.toString()} failed.`,
        });
      }
      const result = await createMainZones(session, canvasId, gridSize, gridPattern);
      return c.json({
        success: true,
        message: `${result.created.toString()} zones created, ${result.failed.toString()} failed.`,
      });
    } catch (err) {
      c.var.logger.error({ err: (err as Error).message }, "create-zones failed");
      return c.json({ success: false, error: (err as Error).message }, 500);
    }
  });

  app.delete("/delete-zones", async (c) => {
    const session = c.var.session;
    const canvasId = c.var.mutableConfig.canvasId;
    try {
      const anchors = await session.widgets.anchors.list(canvasId);
      const scripted: Anchor[] = [];
      for (const a of anchors) {
        const name = (a as unknown as { anchor_name?: string }).anchor_name;
        if (name && name.endsWith("(Script Made)")) scripted.push(a);
      }
      if (scripted.length === 0) {
        return c.json({ success: true, message: "No script-created zones to delete." });
      }
      let deleted = 0;
      let failed = 0;
      for (const a of scripted) {
        try {
          await session.widgets.anchors.delete(canvasId, a.id);
          deleted++;
        } catch {
          failed++;
        }
      }
      return c.json({
        success: true,
        message: `${deleted.toString()} zones deleted, ${failed.toString()} failed.`,
      });
    } catch (err) {
      c.var.logger.error({ err: (err as Error).message }, "delete-zones failed");
      return c.json({ success: false, error: (err as Error).message }, 500);
    }
  });

  // ---------- Team-target creation ----------

  app.post("/create-team-targets", async (c) => {
    const session = c.var.session;
    const canvasId = c.var.mutableConfig.canvasId;
    try {
      const zones = await session.widgets.anchors.list(canvasId);
      const widgets = await session.widgets.list(canvasId);
      const existing = new Set<number>();
      for (const w of widgets) {
        const title = (w as unknown as { title?: string }).title;
        if (w.widget_type === "Note" && title) {
          const m = title.match(/^Team_(\d+)_Target$/);
          if (m?.[1]) existing.add(Number.parseInt(m[1], 10));
        }
      }
      const teams = [1, 2, 3, 4, 5, 6, 7] as TeamNumber[];
      const toCreate = teams.filter((t) => !existing.has(t));
      if (toCreate.length === 0) {
        return c.json({
          success: true,
          message: "All team targets already exist. No new targets created.",
        });
      }
      const created: number[] = [];
      for (const team of toCreate) {
        const zone = zones.find(
          (z) => (z as unknown as { anchor_index?: number }).anchor_index === team,
        );
        const loc = (zone as unknown as { location?: { x: number; y: number } } | undefined)
          ?.location;
        if (!loc) continue;
        await session.widgets.notes.create(canvasId, {
          auto_text_color: true,
          background_color: getTeamBaseColor(team),
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
      }
      if (created.length === 0) {
        return c.json({ success: false, message: "Failed to create any team targets." });
      }
      return c.json({
        success: true,
        message: `Create Team Targets ${created.join(", ")}`,
      });
    } catch (err) {
      c.var.logger.error({ err: (err as Error).message }, "create-team-targets failed");
      return c.json({ success: false, error: (err as Error).message }, 500);
    }
  });

  return app;
}
