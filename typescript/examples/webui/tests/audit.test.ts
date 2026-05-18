import { describe, expect, it } from "vitest";
import { buildApp } from "../src/index.js";
import { buildFakeDeps } from "./helpers.js";

describe("audit routes", () => {
  it("/api/macros/deleted-records returns minimal record summaries", async () => {
    const deps = buildFakeDeps({
      session: {},
      records: [
        {
          recordId: "rec-1",
          timestamp: "2026-05-18T00:00:00Z",
          zoneId: "z1",
          widgets: [{ widget_type: "Note", id: "n1" }],
        },
      ],
    });
    const app = buildApp(deps);
    const res = await app.request("/api/macros/deleted-records");
    expect(res.status).toBe(200);
    const body = (await res.json()) as {
      success: boolean;
      records: Array<{ recordId: string; timestamp: string }>;
    };
    expect(body.records).toEqual([
      { recordId: "rec-1", timestamp: "2026-05-18T00:00:00Z" },
    ]);
  });

  it("/api/macros/deleted-details returns type breakdown", async () => {
    const deps = buildFakeDeps({
      session: {},
      records: [
        {
          recordId: "rec-1",
          timestamp: "2026-05-18T00:00:00Z",
          zoneId: "z1",
          widgets: [
            { widget_type: "Note" },
            { widget_type: "Note" },
            { widget_type: "Pdf" },
          ],
        },
      ],
    });
    const app = buildApp(deps);
    const res = await app.request("/api/macros/deleted-details?recordId=rec-1");
    expect(res.status).toBe(200);
    const body = (await res.json()) as { count: number; types: Record<string, number> };
    expect(body.count).toBe(3);
    expect(body.types).toEqual({ note: 2, pdf: 1 });
  });

  it("/api/macros/deleted-details 404s on missing recordId", async () => {
    const deps = buildFakeDeps({ session: {} });
    const app = buildApp(deps);
    const res = await app.request("/api/macros/deleted-details?recordId=nope");
    expect(res.status).toBe(404);
  });
});
