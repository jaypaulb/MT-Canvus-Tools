import { describe, expect, it } from "vitest";
import {
  contains,
  distanceBetweenWidgets,
  findWidgetsContainingPoint,
  findWidgetsInArea,
  getCanvasBounds,
  getIntersection,
  getUnion,
  getWidgetIntersection,
  getWidgetUnion,
  intersects,
  touches,
  widgetBoundingBox,
  widgetContains,
  widgetsIntersect,
  widgetsTouch,
} from "../../src/extras/geometry.js";

const R = (x: number, y: number, w: number, h: number): { x: number; y: number; width: number; height: number } => ({
  x,
  y,
  width: w,
  height: h,
});

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

describe("geometry primitives", () => {
  it("contains returns true for nested rectangles", () => {
    expect(contains(R(0, 0, 100, 100), R(10, 10, 50, 50))).toBe(true);
    expect(contains(R(0, 0, 100, 100), R(50, 50, 100, 100))).toBe(false);
  });

  it("touches and intersects discriminate edge contact vs overlap", () => {
    const a = R(0, 0, 10, 10);
    const b = R(10, 0, 10, 10);
    expect(touches(a, b)).toBe(true);
    expect(intersects(a, b)).toBe(false);
    const c = R(5, 0, 10, 10);
    expect(intersects(a, c)).toBe(true);
  });

  it("getIntersection returns the overlap rect", () => {
    const isect = getIntersection(R(0, 0, 10, 10), R(5, 5, 10, 10));
    expect(isect).toEqual(R(5, 5, 5, 5));
    expect(getIntersection(R(0, 0, 5, 5), R(10, 10, 5, 5))).toBeUndefined();
  });

  it("getUnion grows to enclose both", () => {
    expect(getUnion(R(0, 0, 5, 5), R(10, 10, 5, 5))).toEqual(R(0, 0, 15, 15));
  });
});

describe("widget geometry", () => {
  it("widgetBoundingBox unpacks location + size", () => {
    expect(widgetBoundingBox(W("a", 1, 2, 3, 4))).toEqual(R(1, 2, 3, 4));
  });

  it("widgetContains / widgetsTouch / widgetsIntersect", () => {
    const outer = W("o", 0, 0, 100, 100);
    const inner = W("i", 10, 10, 20, 20);
    expect(widgetContains(outer, inner)).toBe(true);
    expect(widgetsTouch(outer, inner)).toBe(true);
    expect(widgetsIntersect(outer, inner)).toBe(true);
  });

  it("distanceBetweenWidgets returns 0 when overlapping", () => {
    expect(distanceBetweenWidgets(W("a", 0, 0, 10, 10), W("b", 5, 5, 10, 10))).toBe(0);
  });

  it("distanceBetweenWidgets returns axis gap when separated on one axis", () => {
    expect(distanceBetweenWidgets(W("a", 0, 0, 10, 10), W("b", 30, 0, 10, 10))).toBe(20);
    expect(distanceBetweenWidgets(W("a", 0, 0, 10, 10), W("b", 0, 50, 10, 10))).toBe(40);
  });

  it("distanceBetweenWidgets returns euclidean hypot when separated on both axes (parity with Go + Python)", () => {
    // 3-4-5 triangle: horizontal gap 30, vertical gap 40, expected 50.
    expect(distanceBetweenWidgets(W("a", 0, 0, 10, 10), W("b", 40, 50, 10, 10))).toBe(50);
  });

  it("findWidgetsInArea returns intersecting widgets", () => {
    const a = W("a", 0, 0, 5, 5);
    const b = W("b", 100, 100, 5, 5);
    expect(findWidgetsInArea([a, b], R(0, 0, 50, 50))).toEqual([a]);
  });

  it("findWidgetsContainingPoint", () => {
    const a = W("a", 0, 0, 10, 10);
    expect(findWidgetsContainingPoint([a], { x: 5, y: 5 })).toEqual([a]);
    expect(findWidgetsContainingPoint([a], { x: 50, y: 50 })).toEqual([]);
  });

  it("getCanvasBounds returns union of all widgets", () => {
    expect(getCanvasBounds([W("a", 0, 0, 5, 5), W("b", 10, 10, 5, 5)])).toEqual(R(0, 0, 15, 15));
    expect(getCanvasBounds([])).toBeUndefined();
  });

  it("getWidgetIntersection and getWidgetUnion behave like rect counterparts", () => {
    const a = W("a", 0, 0, 10, 10);
    const b = W("b", 5, 5, 10, 10);
    expect(getWidgetIntersection(a, b)).toEqual(R(5, 5, 5, 5));
    expect(getWidgetUnion(a, b)).toEqual(R(0, 0, 15, 15));
  });
});
