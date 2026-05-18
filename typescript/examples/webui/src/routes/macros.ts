/**
 * Macro routes — port of the legacy `/api/macros/*` zone-based bulk
 * operations.
 *
 * Endpoints:
 *   - POST /api/macros/move             — translate widgets src → dst zone
 *   - POST /api/macros/copy             — duplicate widgets src → dst zone
 *   - POST /api/macros/delete           — soft-delete widgets in a zone
 *                                          (records logged to JsonStore)
 *   - POST /api/macros/auto-grid        — fit widgets into N×N grid
 *   - POST /api/macros/group-color      — cluster by background_color
 *   - POST /api/macros/group-title      — sort/stack by title
 *   - POST /api/macros/pin-all          — pin every widget in zone
 *   - POST /api/macros/unpin-all        — unpin every widget in zone
 *   - POST /api/macros/export           — return widgets JSON for a zone
 *   - POST /api/macros/import           — recreate widgets from upload
 *
 * Behaviour is preserved verbatim from the legacy server.js, including
 * the legacy POINT-containment semantics of `widgetIsInZone` (see
 * `lib/zone-helpers.ts`). Where the legacy server used naive per-type
 * URL dispatch, this port uses `session.widgets.updateAny` /
 * `createAny`, which handle the per-type routing centrally (including
 * the `grid_size`-stripping fix for Table per changelog §4).
 */

import { Hono } from "hono";
import { zValidator } from "@hono/zod-validator";
import { z } from "zod";
import { setTimeout as sleep } from "node:timers/promises";
import type { Widget } from "@mt-canvus-tools/sdk";
import type { AppEnv } from "../app-env.js";
import {
  getZoneBoundingBox,
  patchHelperFor,
  transformWidgetLocationAndScale,
  widgetIsInZone,
} from "../lib/zone-helpers.js";
import { trimDeletedRecords } from "../lib/store.js";

const moveCopyBody = z.object({
  sourceZoneId: z.string().min(1),
  targetZoneId: z.string().min(1),
});
const zoneBody = z.object({ zoneId: z.string().min(1) });
const groupColorBody = z.object({
  zoneId: z.string().min(1),
  tolerance: z.union([z.number(), z.string()]).transform((v) =>
    typeof v === "string" ? Number.parseInt(v, 10) : Math.floor(v),
  ),
});

/** Filter widgets that legacy macros operate on (exclude anchors / connectors). */
function operableWidgetsInZone(
  widgets: readonly Widget[],
  zoneId: string,
  bb: { x: number; y: number; width: number; height: number; scale: number },
): Widget[] {
  const out: Widget[] = [];
  for (const w of widgets) {
    if (w.id === zoneId) continue;
    const wt = (w.widget_type ?? "").toLowerCase();
    if (wt === "connector" || wt === "anchor") continue;
    if (!widgetIsInZone(w, bb)) continue;
    out.push(w);
  }
  return out;
}

function colorDistance(c1: string, c2: string): number {
  const a = c1.slice(1);
  const b = c2.slice(1);
  const r1 = Number.parseInt(a.substring(0, 2), 16) || 0;
  const g1 = Number.parseInt(a.substring(2, 4), 16) || 0;
  const b1 = Number.parseInt(a.substring(4, 6), 16) || 0;
  const r2 = Number.parseInt(b.substring(0, 2), 16) || 0;
  const g2 = Number.parseInt(b.substring(2, 4), 16) || 0;
  const b2c = Number.parseInt(b.substring(4, 6), 16) || 0;
  return Math.sqrt((r1 - r2) ** 2 + (g1 - g2) ** 2 + (b1 - b2c) ** 2);
}

export function macrosRoutes(): Hono<AppEnv> {
  const app = new Hono<AppEnv>();

  // ---------- move ----------
  app.post("/api/macros/move", zValidator("json", moveCopyBody), async (c) => {
    const { sourceZoneId, targetZoneId } = c.req.valid("json");
    const session = c.var.session;
    const canvasId = c.var.mutableConfig.canvasId;
    try {
      const [sourceBB, targetBB, allWidgets] = await Promise.all([
        getZoneBoundingBox(session, canvasId, sourceZoneId),
        getZoneBoundingBox(session, canvasId, targetZoneId),
        session.widgets.list(canvasId),
      ]);
      const toMove = operableWidgetsInZone(allWidgets, sourceZoneId, sourceBB);
      let moved = 0;
      for (const w of toMove) {
        const cloned = JSON.parse(JSON.stringify(w)) as Widget;
        transformWidgetLocationAndScale(cloned as { location?: { x: number; y: number }; scale?: number }, sourceBB, targetBB);
        const patch = patchHelperFor(session, w);
        const body = {
          location: (cloned as unknown as { location?: { x: number; y: number } }).location,
          scale: (cloned as unknown as { scale?: number }).scale,
        };
        await patch(canvasId, w.id, body as Record<string, unknown>);
        moved++;
      }
      return c.json({ success: true, message: `${moved.toString()} widgets moved.` });
    } catch (err) {
      c.var.logger.error({ err: (err as Error).message }, "macros/move failed");
      return c.json({ success: false, error: (err as Error).message }, 500);
    }
  });

  // ---------- copy (BFS over parent_id + connector endpoints) ----------
  app.post("/api/macros/copy", zValidator("json", moveCopyBody), async (c) => {
    const { sourceZoneId, targetZoneId } = c.req.valid("json");
    const session = c.var.session;
    const canvasId = c.var.mutableConfig.canvasId;
    try {
      const [sourceBB, targetBB, allWidgets] = await Promise.all([
        getZoneBoundingBox(session, canvasId, sourceZoneId),
        getZoneBoundingBox(session, canvasId, targetZoneId),
        session.widgets.list(canvasId),
      ]);

      // Phase 1: build the set of normal widgets in the zone.
      const inZoneIds = new Set<string>();
      for (const w of allWidgets) {
        const wt = (w.widget_type ?? "").toLowerCase();
        if (wt === "anchor") continue;
        if (wt !== "connector" && widgetIsInZone(w, sourceBB)) inZoneIds.add(w.id);
      }

      // Phase 2: build inSource = normal widgets + connectors whose both endpoints are in zone.
      const inSource: Widget[] = [];
      for (const w of allWidgets) {
        const wt = (w.widget_type ?? "").toLowerCase();
        if (wt === "anchor") continue;
        if (wt === "connector") {
          const c2 = w as unknown as { src?: { id?: string }; dst?: { id?: string } };
          const s = c2.src?.id;
          const d = c2.dst?.id;
          if (s && d && inZoneIds.has(s) && inZoneIds.has(d)) inSource.push(w);
        } else if (inZoneIds.has(w.id)) {
          inSource.push(w);
        }
      }

      const allMap = new Map<string, Widget>();
      inSource.forEach((w) => allMap.set(w.id, w));
      const copiedMap = new Map<string, string>(); // old → new id

      function canCopy(w: Widget): boolean {
        const wt = (w.widget_type ?? "").toLowerCase();
        if (wt === "connector") {
          const c2 = w as unknown as { src?: { id?: string }; dst?: { id?: string } };
          const s = c2.src?.id;
          const d = c2.dst?.id;
          if (!s || !d) return false;
          return copiedMap.has(s) && copiedMap.has(d);
        }
        const parent = (w as unknown as { parent_id?: string }).parent_id;
        if (!parent) return true;
        if (!allMap.has(parent)) return true;
        return copiedMap.has(parent);
      }

      let remaining = inSource.length;
      let safetyCounter = 0;
      while (remaining > 0 && safetyCounter < 1000) {
        for (const w of inSource) {
          if (copiedMap.has(w.id) || !canCopy(w)) continue;
          const cloned = JSON.parse(JSON.stringify(w)) as Record<string, unknown>;
          delete cloned.id;
          const wt = (w.widget_type ?? "").toLowerCase();
          if (wt !== "connector") {
            if (cloned.location) {
              transformWidgetLocationAndScale(
                cloned as { location?: { x: number; y: number }; scale?: number },
                sourceBB,
                targetBB,
              );
            }
            const parentId = (w as unknown as { parent_id?: string }).parent_id;
            if (parentId && allMap.has(parentId)) {
              const newParent = copiedMap.get(parentId);
              if (newParent) cloned.parent_id = newParent;
              else delete cloned.parent_id;
            }
            if (cloned.auto_text_color === true) delete cloned.text_color;
          } else {
            const c2 = cloned as { src?: { id?: string }; dst?: { id?: string } };
            const oldSrc = (w as unknown as { src?: { id?: string } }).src?.id;
            const oldDst = (w as unknown as { dst?: { id?: string } }).dst?.id;
            if (oldSrc && oldDst && copiedMap.has(oldSrc) && copiedMap.has(oldDst)) {
              if (c2.src) c2.src.id = copiedMap.get(oldSrc) ?? "";
              if (c2.dst) c2.dst.id = copiedMap.get(oldDst) ?? "";
            }
          }

          await sleep(200); // rate-limit (matches legacy)
          try {
            const created = await session.widgets.createAny(canvasId, cloned);
            copiedMap.set(w.id, created.id);
          } catch (err) {
            c.var.logger.warn({ id: w.id, err: (err as Error).message }, "copy: POST failed");
          }
        }
        const newRemaining = inSource.length - copiedMap.size;
        if (newRemaining === remaining) break; // no progress → bail
        remaining = newRemaining;
        safetyCounter++;
      }
      return c.json({
        success: true,
        message: `${copiedMap.size.toString()} widgets copied (connectors only if both endpoints in zone).`,
      });
    } catch (err) {
      c.var.logger.error({ err: (err as Error).message }, "macros/copy failed");
      return c.json({ success: false, error: (err as Error).message }, 500);
    }
  });

  // ---------- delete (logs to deletedRecordStore) ----------
  app.post("/api/macros/delete", zValidator("json", zoneBody), async (c) => {
    const { zoneId } = c.req.valid("json");
    const session = c.var.session;
    const canvasId = c.var.mutableConfig.canvasId;
    try {
      const [bb, allWidgets] = await Promise.all([
        getZoneBoundingBox(session, canvasId, zoneId),
        session.widgets.list(canvasId),
      ]);
      const toDelete = operableWidgetsInZone(allWidgets, zoneId, bb);
      if (toDelete.length === 0) {
        return c.json({ success: true, message: "No widgets found in the selected zone to delete." });
      }

      const recordId = `rec-${Date.now().toString()}`;
      const record = {
        recordId,
        timestamp: new Date().toISOString(),
        zoneId,
        widgets: toDelete.map((w) => ({ ...(w as unknown as Record<string, unknown>) })),
      };

      const existing = await c.var.deletedRecordStore.read();
      existing.push(record);
      await c.var.deletedRecordStore.write(trimDeletedRecords(existing));

      // Actually delete on the server (legacy logged + actually deleted)
      let deleted = 0;
      for (const w of toDelete) {
        try {
          await session.widgets.deleteAny(canvasId, w.id, w.widget_type);
          deleted++;
        } catch (err) {
          c.var.logger.warn({ id: w.id, err: (err as Error).message }, "delete: failed");
        }
      }
      return c.json({
        success: true,
        message: `${deleted.toString()} widgets deleted from zone ${zoneId} and recorded with ID ${recordId}.`,
      });
    } catch (err) {
      c.var.logger.error({ err: (err as Error).message }, "macros/delete failed");
      return c.json({ success: false, error: (err as Error).message }, 500);
    }
  });

  // ---------- auto-grid ----------
  app.post("/api/macros/auto-grid", zValidator("json", zoneBody), async (c) => {
    const { zoneId } = c.req.valid("json");
    const session = c.var.session;
    const canvasId = c.var.mutableConfig.canvasId;
    try {
      const [bb, allWidgets] = await Promise.all([
        getZoneBoundingBox(session, canvasId, zoneId),
        session.widgets.list(canvasId),
      ]);
      const inZone = operableWidgetsInZone(allWidgets, zoneId, bb);
      if (inZone.length === 0) {
        return c.json({ success: true, message: "No widgets found to auto-grid." });
      }
      const N = inZone.length;
      const gridSize = Math.ceil(Math.sqrt(N));
      const cellW = bb.width / gridSize;
      const cellH = bb.height / gridSize;
      let updated = 0;
      let index = 0;
      for (const w of inZone) {
        const size = (w as unknown as { size?: { width: number; height: number } }).size;
        const origH = size?.height ?? 300;
        let newScale = 1;
        const desiredH = cellH - 50;
        if (desiredH > 0) newScale = desiredH / origH;
        if (size?.width) {
          const desiredW = cellW - 50;
          newScale = Math.min(newScale, desiredW / size.width);
        }
        if (newScale < 0) newScale = 0.1;
        const r = Math.floor(index / gridSize);
        const cc = index % gridSize;
        index++;
        const targetX = bb.x + cc * cellW + 25;
        const targetY = bb.y + r * cellH + 25;
        await patchHelperFor(session, w)(canvasId, w.id, {
          location: { x: targetX, y: targetY },
          scale: newScale,
        });
        updated++;
      }
      return c.json({
        success: true,
        message: `Auto-Grid placed ${updated.toString()} widgets in a ~${gridSize.toString()}x${gridSize.toString()} grid.`,
      });
    } catch (err) {
      c.var.logger.error({ err: (err as Error).message }, "macros/auto-grid failed");
      return c.json({ success: false, error: (err as Error).message }, 500);
    }
  });

  // ---------- group-color ----------
  app.post("/api/macros/group-color", zValidator("json", groupColorBody), async (c) => {
    const { zoneId, tolerance } = c.req.valid("json");
    const session = c.var.session;
    const canvasId = c.var.mutableConfig.canvasId;
    try {
      const [bb, allWidgets] = await Promise.all([
        getZoneBoundingBox(session, canvasId, zoneId),
        session.widgets.list(canvasId),
      ]);
      const inZone = operableWidgetsInZone(allWidgets, zoneId, bb);
      if (inZone.length === 0) {
        return c.json({ success: true, message: "No widgets found to group by color." });
      }
      const threshold = 255 * (tolerance / 100);
      type Cluster = { rep: string; widgets: Widget[] };
      const clusters: Cluster[] = [];
      for (const w of inZone) {
        const color = (w as unknown as { background_color?: string }).background_color ?? "#FFFFFF00";
        let placed = false;
        for (const cl of clusters) {
          if (colorDistance(color, cl.rep) <= threshold) {
            cl.widgets.push(w);
            placed = true;
            break;
          }
        }
        if (!placed) clusters.push({ rep: color, widgets: [w] });
      }
      let clusterX = bb.x + 100;
      const startY = bb.y + 100;
      let updated = 0;
      for (const cl of clusters) {
        let clusterY = startY;
        for (const w of cl.widgets) {
          await patchHelperFor(session, w)(canvasId, w.id, {
            location: { x: clusterX, y: clusterY },
          });
          updated++;
          clusterY += 25;
        }
        clusterX += 100;
      }
      return c.json({
        success: true,
        message: `${updated.toString()} widgets grouped by color into ${clusters.length.toString()} cluster(s).`,
      });
    } catch (err) {
      c.var.logger.error({ err: (err as Error).message }, "macros/group-color failed");
      return c.json({ success: false, error: (err as Error).message }, 500);
    }
  });

  // ---------- group-title ----------
  app.post("/api/macros/group-title", zValidator("json", zoneBody), async (c) => {
    const { zoneId } = c.req.valid("json");
    const session = c.var.session;
    const canvasId = c.var.mutableConfig.canvasId;
    try {
      const [bb, allWidgets] = await Promise.all([
        getZoneBoundingBox(session, canvasId, zoneId),
        session.widgets.list(canvasId),
      ]);
      const inZone = operableWidgetsInZone(allWidgets, zoneId, bb);
      if (inZone.length === 0) {
        return c.json({ success: true, message: "No widgets found to group by title." });
      }
      inZone.sort((a, b) => {
        const at = ((a as unknown as { title?: string }).title ?? "").toLowerCase();
        const bt = ((b as unknown as { title?: string }).title ?? "").toLowerCase();
        return at < bt ? -1 : at > bt ? 1 : 0;
      });
      const x = bb.x + 100;
      let y = bb.y + 100;
      let count = 0;
      for (const w of inZone) {
        await patchHelperFor(session, w)(canvasId, w.id, { location: { x, y } });
        y += 60;
        count++;
      }
      return c.json({
        success: true,
        message: `${count.toString()} widgets grouped by title (alphabetical).`,
      });
    } catch (err) {
      c.var.logger.error({ err: (err as Error).message }, "macros/group-title failed");
      return c.json({ success: false, error: (err as Error).message }, 500);
    }
  });

  // ---------- pin-all / unpin-all ----------
  for (const op of [
    { path: "/api/macros/pin-all", pinned: true, verb: "pinned" },
    { path: "/api/macros/unpin-all", pinned: false, verb: "unpinned" },
  ] as const) {
    app.post(op.path, zValidator("json", zoneBody), async (c) => {
      const { zoneId } = c.req.valid("json");
      const session = c.var.session;
      const canvasId = c.var.mutableConfig.canvasId;
      try {
        const [bb, allWidgets] = await Promise.all([
          getZoneBoundingBox(session, canvasId, zoneId),
          session.widgets.list(canvasId),
        ]);
        const inZone = operableWidgetsInZone(allWidgets, zoneId, bb);
        if (inZone.length === 0) {
          return c.json({
            success: true,
            message: `No widgets found in zone to ${op.verb}.`,
          });
        }
        let count = 0;
        for (const w of inZone) {
          try {
            await patchHelperFor(session, w)(canvasId, w.id, { pinned: op.pinned });
            count++;
          } catch (err) {
            c.var.logger.warn({ id: w.id, err: (err as Error).message }, `${op.verb}: failed`);
          }
        }
        return c.json({
          success: true,
          message: `${count.toString()} widgets ${op.verb} in the zone.`,
        });
      } catch (err) {
        c.var.logger.error({ err: (err as Error).message }, `${op.path} failed`);
        return c.json({ success: false, error: (err as Error).message }, 500);
      }
    });
  }

  // ---------- export ----------
  app.post("/api/macros/export", zValidator("json", zoneBody), async (c) => {
    const { zoneId } = c.req.valid("json");
    const session = c.var.session;
    const canvasId = c.var.mutableConfig.canvasId;
    try {
      const [bb, allWidgets] = await Promise.all([
        getZoneBoundingBox(session, canvasId, zoneId),
        session.widgets.list(canvasId),
      ]);
      const inZone: Widget[] = [];
      for (const w of allWidgets) {
        if (widgetIsInZone(w, bb)) inZone.push(w);
      }
      return c.json({
        success: true,
        timestamp: new Date().toISOString(),
        zoneId,
        widgets: inZone,
      });
    } catch (err) {
      c.var.logger.error({ err: (err as Error).message }, "macros/export failed");
      return c.json({ success: false, error: (err as Error).message }, 500);
    }
  });

  // ---------- import ----------
  app.post("/api/macros/import", async (c) => {
    const session = c.var.session;
    const canvasId = c.var.mutableConfig.canvasId;
    try {
      const form = await c.req.parseBody();
      const file = form["importFile"];
      if (!(file instanceof File)) {
        return c.json({ success: false, error: "No file uploaded for import." }, 400);
      }
      const text = await file.text();
      const parsed = JSON.parse(text) as { widgets?: Array<Record<string, unknown>> };
      if (!parsed.widgets || !Array.isArray(parsed.widgets)) {
        return c.json({ success: false, error: "Uploaded file missing 'widgets' array." }, 400);
      }
      const widgets = parsed.widgets;
      const allMap = new Map<string, Record<string, unknown>>();
      widgets.forEach((w) => {
        if (typeof w.id === "string") allMap.set(w.id, w);
      });
      const importedMap = new Map<string, string>();

      function canImport(w: Record<string, unknown>): boolean {
        const parent = typeof w.parent_id === "string" ? w.parent_id : undefined;
        if (!parent) return true;
        if (!allMap.has(parent)) return true;
        return importedMap.has(parent);
      }

      let remaining = widgets.length;
      let safety = 0;
      let imported = 0;
      while (remaining > 0 && safety < 1000) {
        for (const w of widgets) {
          const oldId = typeof w.id === "string" ? w.id : undefined;
          if (!oldId || importedMap.has(oldId) || !canImport(w)) continue;
          const cloned: Record<string, unknown> = { ...w };
          delete cloned.id;
          const parent = typeof w.parent_id === "string" ? w.parent_id : undefined;
          if (parent && allMap.has(parent)) {
            const newParent = importedMap.get(parent);
            if (newParent) cloned.parent_id = newParent;
            else delete cloned.parent_id;
          }
          try {
            const created = await session.widgets.createAny(canvasId, cloned);
            importedMap.set(oldId, created.id);
            imported++;
          } catch (err) {
            c.var.logger.warn({ oldId, err: (err as Error).message }, "import failed");
          }
        }
        const newRemaining = widgets.length - importedMap.size;
        if (newRemaining === remaining) break;
        remaining = newRemaining;
        safety++;
      }
      return c.json({ success: true, message: `${imported.toString()} widgets imported.` });
    } catch (err) {
      c.var.logger.error({ err: (err as Error).message }, "macros/import failed");
      return c.json({ success: false, error: (err as Error).message }, 500);
    }
  });

  return app;
}
