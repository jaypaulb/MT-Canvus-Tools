import { describe, expect, it } from "vitest";
import {
  BatchWidgetOperations,
  calculateWidgetDensity,
  createSpatialGroup,
  findWidgetClusters,
  WidgetZoneManager,
} from "../../src/extras/widgetOperations.js";

const W = (id: string, x: number, y: number, w: number, h: number): {
  id: string;
  widget_type: string;
  location: { x: number; y: number };
  size: { width: number; height: number };
} => ({
  id,
  widget_type: "Note",
  location: { x, y },
  size: { width: w, height: h },
});

describe("WidgetZoneManager", () => {
  it("creates a padded zone enclosing widgets", () => {
    const mgr = new WidgetZoneManager();
    const z = mgr.createZoneFromWidgets([W("a", 0, 0, 10, 10), W("b", 20, 0, 10, 10)], "z");
    expect(z.location.x).toBe(-20);
    expect(z.size.width).toBe(70);
  });

  it("widgetsInZone returns only fully-contained widgets", () => {
    const mgr = new WidgetZoneManager();
    const a = W("a", 5, 5, 5, 5);
    const b = W("b", 100, 100, 5, 5);
    const zone = {
      id: "z",
      name: "test",
      location: { x: 0, y: 0 },
      size: { width: 50, height: 50 },
    };
    expect(mgr.widgetsInZone([a, b], zone)).toEqual([a]);
  });

  it("throws when called with no widgets", () => {
    const mgr = new WidgetZoneManager();
    expect(() => mgr.createZoneFromWidgets([], "z")).toThrow();
  });
});

describe("BatchWidgetOperations", () => {
  it("moveWidgets generates location-shifted payloads", () => {
    const batch = new BatchWidgetOperations();
    const ops = batch.moveWidgets([W("a", 0, 0, 10, 10)], 5, 7);
    expect(ops).toHaveLength(1);
    expect(ops[0]?.operation).toBe("move");
    expect(ops[0]?.payload).toEqual({ location: { x: 5, y: 7 } });
  });

  it("resizeWidgets scales size", () => {
    const batch = new BatchWidgetOperations();
    const ops = batch.resizeWidgets([W("a", 0, 0, 10, 10)], 2);
    expect(ops[0]?.payload).toEqual({ size: { width: 20, height: 20 } });
  });

  it("widgetsContainId and widgetsTouchId return matches", () => {
    const batch = new BatchWidgetOperations();
    const outer = W("o", 0, 0, 100, 100);
    const inner = W("i", 10, 10, 5, 5);
    expect(batch.widgetsContainId([outer, inner], "i")).toEqual([outer]);
    expect(batch.widgetsTouchId([outer, inner], "i")).toEqual([outer]);
  });
});

describe("spatial grouping", () => {
  it("createSpatialGroup buckets nearby widgets", () => {
    const groups = createSpatialGroup(
      [W("a", 0, 0, 1, 1), W("b", 5, 5, 1, 1), W("c", 100, 100, 1, 1)],
      10,
    );
    expect(groups).toHaveLength(2);
  });

  it("findWidgetClusters filters by min size", () => {
    const clusters = findWidgetClusters(
      [W("a", 0, 0, 1, 1), W("b", 5, 5, 1, 1), W("c", 100, 100, 1, 1)],
      2,
      20,
    );
    expect(clusters).toHaveLength(1);
    expect(clusters[0]).toHaveLength(2);
  });

  it("calculateWidgetDensity counts intersecting widgets per area", () => {
    const d = calculateWidgetDensity(
      [W("a", 0, 0, 1, 1), W("b", 50, 50, 1, 1)],
      { x: 0, y: 0, width: 100, height: 100 },
    );
    expect(d).toBe(2 / 10000);
  });
});
