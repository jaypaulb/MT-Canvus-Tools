/**
 * Deleted-record audit routes.
 *
 * Lists the cached records that the `/api/macros/delete` handler
 * appends to `state/macros-deleted-records.json`. Used by the admin
 * "undelete" UI.
 */

import { Hono } from "hono";
import { zValidator } from "@hono/zod-validator";
import { z } from "zod";
import type { AppEnv } from "../app-env.js";

export function auditRoutes(): Hono<AppEnv> {
  const app = new Hono<AppEnv>();

  app.get("/api/macros/deleted-records", async (c) => {
    const records = await c.var.deletedRecordStore.read();
    const minimal = records.map((r) => ({ recordId: r.recordId, timestamp: r.timestamp }));
    return c.json({ success: true, records: minimal });
  });

  app.get(
    "/api/macros/deleted-details",
    zValidator("query", z.object({ recordId: z.string().min(1) })),
    async (c) => {
      const { recordId } = c.req.valid("query");
      const records = await c.var.deletedRecordStore.read();
      const record = records.find((r) => r.recordId === recordId);
      if (!record) {
        return c.json({ success: false, error: "No such recordId found." }, 404);
      }
      const types: Record<string, number> = {};
      for (const w of record.widgets) {
        const t = (
          (w as Record<string, unknown>).widget_type as string | undefined
        )?.toLowerCase() ?? "unknown";
        types[t] = (types[t] ?? 0) + 1;
      }
      return c.json({ success: true, count: record.widgets.length, types });
    },
  );

  return app;
}
