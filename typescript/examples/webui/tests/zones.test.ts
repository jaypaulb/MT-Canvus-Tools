/**
 * Zone-route tests with a hand-rolled fake `Session.widgets.anchors`.
 *
 * Focus: the SDK call shapes our route assembles + the
 * filtering/iteration we layer on top, not the SDK's wire shape.
 */

import { describe, expect, it, vi } from "vitest";
import { buildApp } from "../src/index.js";
import { buildFakeDeps } from "./helpers.js";

function makeSession(overrides: Record<string, unknown> = {}): unknown {
  return {
    widgets: {
      list: vi.fn().mockResolvedValue([]),
      anchors: {
        list: vi.fn().mockResolvedValue([]),
        get: vi.fn(),
        create: vi.fn(),
        delete: vi.fn(),
      },
      notes: {
        create: vi.fn(),
        delete: vi.fn(),
      },
      ...overrides,
    },
  };
}

describe("zones routes", () => {
  it("/get-zones forwards anchors.list and returns success envelope", async () => {
    const anchors = [
      { id: "a1", widget_type: "Anchor", anchor_index: 0 },
      { id: "a2", widget_type: "Anchor", anchor_index: 1 },
    ];
    const session = makeSession();
    const anchorsList = vi.fn().mockResolvedValue(anchors);
    (session as { widgets: { anchors: { list: typeof anchorsList } } }).widgets.anchors.list =
      anchorsList;

    const deps = buildFakeDeps({ session });
    const app = buildApp(deps);
    const res = await app.request("/get-zones");
    expect(res.status).toBe(200);
    expect(anchorsList).toHaveBeenCalledWith("canvas-id-1");
    const body = (await res.json()) as { success: boolean; zones: typeof anchors };
    expect(body.zones).toEqual(anchors);
  });

  it("/delete-zones only deletes anchors whose name ends with (Script Made)", async () => {
    const anchors = [
      { id: "a1", anchor_name: "Manual zone" },
      { id: "a2", anchor_name: "3x3 Zone 4 (Script Made)" },
      { id: "a3", anchor_name: "SubZone 1.1 (Script Made)" },
    ];
    const session = makeSession();
    const w = session as {
      widgets: {
        anchors: {
          list: ReturnType<typeof vi.fn>;
          delete: ReturnType<typeof vi.fn>;
        };
      };
    };
    w.widgets.anchors.list = vi.fn().mockResolvedValue(anchors);
    w.widgets.anchors.delete = vi.fn().mockResolvedValue(undefined);

    const deps = buildFakeDeps({ session });
    const app = buildApp(deps);
    const res = await app.request("/delete-zones", { method: "DELETE" });
    expect(res.status).toBe(200);
    expect(w.widgets.anchors.delete).toHaveBeenCalledTimes(2);
    expect(w.widgets.anchors.delete).toHaveBeenCalledWith("canvas-id-1", "a2");
    expect(w.widgets.anchors.delete).toHaveBeenCalledWith("canvas-id-1", "a3");
  });

  it("/create-zones validates gridSize choices via zod", async () => {
    const deps = buildFakeDeps({ session: makeSession() });
    const app = buildApp(deps);
    const res = await app.request("/create-zones", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ gridSize: 2 }),
    });
    expect(res.status).toBe(400);
  });
});
