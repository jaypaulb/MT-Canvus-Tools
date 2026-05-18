/**
 * Macro-route tests. We exercise the in-zone filtering, the
 * per-type PATCH dispatch (via `updateAny`), and the delete-record
 * persistence path.
 */

import { describe, expect, it, vi } from "vitest";
import { buildApp } from "../src/index.js";
import { buildFakeDeps } from "./helpers.js";

interface FakeWidget {
  id: string;
  widget_type: string;
  location?: { x: number; y: number };
  scale?: number;
  size?: { width: number; height: number };
  background_color?: string;
  title?: string;
}

function bbAnchor(): FakeWidget & { anchor_index: number } {
  return {
    id: "src",
    widget_type: "Anchor",
    anchor_index: 0,
    location: { x: 0, y: 0 },
    size: { width: 1000, height: 1000 },
  };
}

function dstAnchor(): FakeWidget & { anchor_index: number } {
  return {
    id: "dst",
    widget_type: "Anchor",
    anchor_index: 1,
    location: { x: 2000, y: 0 },
    size: { width: 1000, height: 1000 },
  };
}

describe("macros routes", () => {
  it("/api/macros/move PATCHes every in-zone widget via updateAny", async () => {
    const inZone: FakeWidget[] = [
      { id: "w1", widget_type: "Note", location: { x: 100, y: 100 }, scale: 1 },
      { id: "w2", widget_type: "Image", location: { x: 200, y: 200 }, scale: 1 },
      { id: "w3", widget_type: "Note", location: { x: 9999, y: 9999 } }, // out of zone
    ];
    const updateAny = vi.fn().mockResolvedValue({ id: "ok" });
    const anchorGet = vi
      .fn()
      .mockImplementation((_canvas: string, id: string) =>
        Promise.resolve(id === "src" ? bbAnchor() : dstAnchor()),
      );
    const session = {
      widgets: {
        list: vi.fn().mockResolvedValue(inZone),
        anchors: { get: anchorGet },
        updateAny,
      },
    };
    const deps = buildFakeDeps({ session });
    const app = buildApp(deps);
    const res = await app.request("/api/macros/move", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ sourceZoneId: "src", targetZoneId: "dst" }),
    });
    expect(res.status).toBe(200);
    expect(updateAny).toHaveBeenCalledTimes(2); // w1 + w2 only
    const body = (await res.json()) as { message: string };
    expect(body.message).toBe("2 widgets moved.");
  });

  it("/api/macros/delete records the deletion and calls deleteAny", async () => {
    const inZone: FakeWidget[] = [
      { id: "w1", widget_type: "Note", location: { x: 100, y: 100 } },
    ];
    const deleteAny = vi.fn().mockResolvedValue(undefined);
    const anchorGet = vi.fn().mockResolvedValue(bbAnchor());
    const session = {
      widgets: {
        list: vi.fn().mockResolvedValue(inZone),
        anchors: { get: anchorGet },
        deleteAny,
      },
    };
    const deps = buildFakeDeps({ session });
    const app = buildApp(deps);
    const res = await app.request("/api/macros/delete", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ zoneId: "src" }),
    });
    expect(res.status).toBe(200);
    expect(deleteAny).toHaveBeenCalledWith("canvas-id-1", "w1", "Note");
    const records = await deps.variables.deletedRecordStore.read();
    expect(records).toHaveLength(1);
    expect(records[0]?.widgets).toHaveLength(1);
  });

  it("/api/macros/pin-all toggles `pinned` on every operable widget", async () => {
    const inZone: FakeWidget[] = [
      { id: "w1", widget_type: "Note", location: { x: 100, y: 100 } },
      { id: "w2", widget_type: "Connector", location: { x: 100, y: 100 } }, // filtered out
    ];
    const updateAny = vi.fn().mockResolvedValue({ id: "ok" });
    const anchorGet = vi.fn().mockResolvedValue(bbAnchor());
    const session = {
      widgets: {
        list: vi.fn().mockResolvedValue(inZone),
        anchors: { get: anchorGet },
        updateAny,
      },
    };
    const deps = buildFakeDeps({ session });
    const app = buildApp(deps);
    const res = await app.request("/api/macros/pin-all", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ zoneId: "src" }),
    });
    expect(res.status).toBe(200);
    expect(updateAny).toHaveBeenCalledTimes(1);
    expect(updateAny.mock.calls[0]?.[2]).toMatchObject({ pinned: true, widget_type: "Note" });
  });
});
