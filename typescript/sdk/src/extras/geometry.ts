// Phase 4b §4.3 #14: geometry port.
//
// Spatial utilities for working with widget positions, sizes, and
// rectangle relationships. Ported from
// CanvusPythonAPI/canvus_api/geometry.py:14-419.
//
// All coordinates are absolute canvas pixel values (per CLAUDE.md "Canvus
// API coordinates are in PIXELS, not normalized 0-1").

import type { Location, Size } from "../types/common.js";
import type { Connector, Widget } from "../types/widget.js";

/** A 2D point. */
export interface Point {
  readonly x: number;
  readonly y: number;
}

/** A 2D axis-aligned rectangle. */
export interface Rectangle {
  readonly x: number;
  readonly y: number;
  readonly width: number;
  readonly height: number;
}

/** A widget-shaped object that exposes the position/size needed for geometry. */
export interface HasBounds {
  readonly id?: string;
  readonly widget_type?: string;
  readonly location?: Location;
  readonly size?: Size;
}

/** The left edge of a rectangle. */
export function left(r: Rectangle): number {
  return r.x;
}
/** The right edge of a rectangle. */
export function right(r: Rectangle): number {
  return r.x + r.width;
}
/** The top edge of a rectangle. */
export function top(r: Rectangle): number {
  return r.y;
}
/** The bottom edge of a rectangle. */
export function bottom(r: Rectangle): number {
  return r.y + r.height;
}
/** The geometric centre of a rectangle. */
export function center(r: Rectangle): Point {
  return { x: r.x + r.width / 2, y: r.y + r.height / 2 };
}

/** Returns true if `outer` fully encloses `inner`. */
export function contains(outer: Rectangle, inner: Rectangle): boolean {
  return (
    left(outer) <= left(inner) &&
    right(outer) >= right(inner) &&
    top(outer) <= top(inner) &&
    bottom(outer) >= bottom(inner)
  );
}

/** Returns true if `a` and `b` touch (edge-shared) or overlap. */
export function touches(a: Rectangle, b: Rectangle): boolean {
  return !(right(a) < left(b) || left(a) > right(b) || bottom(a) < top(b) || top(a) > bottom(b));
}

/** Returns true if `a` and `b` overlap by a non-zero area. */
export function intersects(a: Rectangle, b: Rectangle): boolean {
  return !(
    right(a) <= left(b) ||
    left(a) >= right(b) ||
    bottom(a) <= top(b) ||
    top(a) >= bottom(b)
  );
}

/** Returns the overlap rectangle of `a` and `b`, or `undefined`. */
export function getIntersection(a: Rectangle, b: Rectangle): Rectangle | undefined {
  if (!intersects(a, b)) return undefined;
  const l = Math.max(left(a), left(b));
  const t = Math.max(top(a), top(b));
  const r = Math.min(right(a), right(b));
  const bt = Math.min(bottom(a), bottom(b));
  return { x: l, y: t, width: r - l, height: bt - t };
}

/** Returns the union rectangle containing both `a` and `b`. */
export function getUnion(a: Rectangle, b: Rectangle): Rectangle {
  const l = Math.min(left(a), left(b));
  const t = Math.min(top(a), top(b));
  const r = Math.max(right(a), right(b));
  const bt = Math.max(bottom(a), bottom(b));
  return { x: l, y: t, width: r - l, height: bt - t };
}

/**
 * Compute a widget's bounding rectangle.
 *
 * For `Connector` widgets (which have no `location`/`size` but do have
 * `src`/`dst` endpoints), the bounding box is derived from the endpoint
 * relative locations plus a 10-px padding (matches the Python port).
 *
 * @throws {Error} If the widget shape can't be interpreted.
 */
export function widgetBoundingBox(widget: HasBounds | Widget | Connector): Rectangle {
  // Connector special case.
  if ((widget as { widget_type?: string }).widget_type === "Connector") {
    const c = widget as Connector;
    const sx = c.src?.rel_location?.x ?? 0;
    const sy = c.src?.rel_location?.y ?? 0;
    const dx = c.dst?.rel_location?.x ?? 0;
    const dy = c.dst?.rel_location?.y ?? 0;
    const padding = 10;
    const minX = Math.min(sx, dx);
    const maxX = Math.max(sx, dx);
    const minY = Math.min(sy, dy);
    const maxY = Math.max(sy, dy);
    return {
      x: minX - padding,
      y: minY - padding,
      width: maxX - minX + 2 * padding,
      height: maxY - minY + 2 * padding,
    };
  }
  const h = widget as HasBounds;
  if (h.location === undefined || h.size === undefined) {
    throw new Error("widgetBoundingBox: widget must expose location and size");
  }
  return {
    x: h.location.x,
    y: h.location.y,
    width: h.size.width,
    height: h.size.height,
  };
}

/** Returns true if `outer` contains `inner` (widget overload). */
export function widgetContains(outer: HasBounds | Widget, inner: HasBounds | Widget): boolean {
  return contains(widgetBoundingBox(outer), widgetBoundingBox(inner));
}

/** Returns true if two widgets touch (edge-shared) or overlap. */
export function widgetsTouch(a: HasBounds | Widget, b: HasBounds | Widget): boolean {
  return touches(widgetBoundingBox(a), widgetBoundingBox(b));
}

/** Returns true if two widgets overlap by a non-zero area. */
export function widgetsIntersect(a: HasBounds | Widget, b: HasBounds | Widget): boolean {
  return intersects(widgetBoundingBox(a), widgetBoundingBox(b));
}

/** Returns the overlap rectangle of two widgets, or `undefined`. */
export function getWidgetIntersection(
  a: HasBounds | Widget,
  b: HasBounds | Widget,
): Rectangle | undefined {
  return getIntersection(widgetBoundingBox(a), widgetBoundingBox(b));
}

/** Returns the union rectangle of two widgets. */
export function getWidgetUnion(a: HasBounds | Widget, b: HasBounds | Widget): Rectangle {
  return getUnion(widgetBoundingBox(a), widgetBoundingBox(b));
}

/**
 * Cartesian gap distance between two widget bounding rectangles.
 *
 * Returns 0 when the rectangles overlap. When the rectangles are separated
 * on both axes, returns the euclidean distance between the nearest corners
 * (`sqrt(horizGap**2 + vertGap**2)`). When separated on a single axis only,
 * returns that axis's gap directly. Symmetric and matches the Go + Python
 * SDK implementations (`go/sdk/canvus/extras/geometry.go` and
 * `python/sdk/src/canvus_sdk/extras/geometry.py`).
 *
 * Note: this deliberately differs from the legacy `CanvusPythonAPI` helper,
 * which returned `min(horizGap, vertGap)` in the both-positive case — the
 * legacy answer was a single-axis projection, not a cartesian distance.
 */
export function distanceBetweenWidgets(
  a: HasBounds | Widget,
  b: HasBounds | Widget,
): number {
  const ra = widgetBoundingBox(a);
  const rb = widgetBoundingBox(b);
  if (intersects(ra, rb)) return 0;
  const horiz = Math.max(0, Math.max(left(ra) - right(rb), left(rb) - right(ra)));
  const vert = Math.max(0, Math.max(top(ra) - bottom(rb), top(rb) - bottom(ra)));
  if (horiz > 0 && vert > 0) return Math.hypot(horiz, vert);
  return Math.max(horiz, vert);
}

/** Return widgets whose bounding boxes intersect the given area. */
export function findWidgetsInArea<W extends HasBounds | Widget>(
  widgets: readonly W[],
  area: Rectangle,
): readonly W[] {
  return widgets.filter((w) => intersects(widgetBoundingBox(w), area));
}

/** Return widgets whose bounding boxes contain the given point. */
export function findWidgetsContainingPoint<W extends HasBounds | Widget>(
  widgets: readonly W[],
  point: Point,
): readonly W[] {
  return widgets.filter((w) => {
    const r = widgetBoundingBox(w);
    return point.x >= left(r) && point.x <= right(r) && point.y >= top(r) && point.y <= bottom(r);
  });
}

/**
 * The smallest rectangle containing every supplied widget.
 *
 * Returns `undefined` for an empty list. Skips widgets whose bounding box
 * cannot be computed (e.g. malformed payloads).
 */
export function getCanvasBounds(widgets: readonly (HasBounds | Widget)[]): Rectangle | undefined {
  if (widgets.length === 0) return undefined;
  let bounds: Rectangle | undefined;
  for (const w of widgets) {
    let rect: Rectangle;
    try {
      rect = widgetBoundingBox(w);
    } catch {
      continue;
    }
    bounds = bounds === undefined ? rect : getUnion(bounds, rect);
  }
  return bounds;
}
