// Phase 4b §4.3 #16: widget_operations port.
//
// Zone management, batch operations, and spatial grouping. Ported from
// CanvusPythonAPI/canvus_api/widget_operations.py:20-488.

import type { Location, Size, Uuid } from "../types/common.js";
import type { Widget } from "../types/widget.js";
import {
  contains,
  getUnion,
  intersects,
  touches,
  widgetBoundingBox,
  widgetContains,
  widgetsTouch,
  type HasBounds,
  type Rectangle,
} from "./geometry.js";

/** Tunable tolerances used by the zone and batch helpers. */
export interface SpatialTolerance {
  readonly positionTolerance: number;
  readonly sizeTolerance: number;
  readonly overlapTolerance: number;
  readonly distanceTolerance: number;
}

/** Defaults that match the Python port (5, 2, 1, 10). */
export const DEFAULT_SPATIAL_TOLERANCE: SpatialTolerance = Object.freeze({
  positionTolerance: 5.0,
  sizeTolerance: 2.0,
  overlapTolerance: 1.0,
  distanceTolerance: 10.0,
});

/**
 * A named spatial region used by {@link WidgetZoneManager}.
 *
 * This is a client-side concept — the Canvus server has no first-class
 * "zone" entity. The zone exposes the same `location`/`size` shape as a
 * widget so it can participate in geometry helpers transparently.
 */
export interface WidgetZone {
  readonly id: string;
  readonly name: string;
  readonly description?: string;
  readonly location: Location;
  readonly size: Size;
}

/**
 * Spatial grouping helper.
 *
 * Use to build a {@link WidgetZone} from a group of widgets, then query for
 * widgets contained in or touching that zone.
 */
export class WidgetZoneManager {
  constructor(public readonly tolerance: SpatialTolerance = DEFAULT_SPATIAL_TOLERANCE) {}

  /**
   * Build a zone that encloses every supplied widget plus a `padding` margin.
   *
   * @throws {Error} If the widget list is empty.
   */
  createZoneFromWidgets(
    widgets: readonly (HasBounds | Widget)[],
    name: string,
    description?: string,
    padding = 20,
  ): WidgetZone {
    if (widgets.length === 0) throw new Error("createZoneFromWidgets: widgets is empty");
    const bounds = this.computeBounds(widgets);
    const padded: Rectangle = {
      x: bounds.x - padding,
      y: bounds.y - padding,
      width: bounds.width + 2 * padding,
      height: bounds.height + 2 * padding,
    };
    return {
      id: `zone_${widgets.length.toString()}_${simpleHash(name).toString()}`,
      name,
      ...(description !== undefined && { description }),
      location: { x: padded.x, y: padded.y },
      size: { width: padded.width, height: padded.height },
    };
  }

  /** Widgets fully contained inside `zone`. */
  widgetsInZone<W extends HasBounds | Widget>(
    widgets: readonly W[],
    zone: WidgetZone,
  ): readonly W[] {
    const zr: Rectangle = {
      x: zone.location.x,
      y: zone.location.y,
      width: zone.size.width,
      height: zone.size.height,
    };
    const out: W[] = [];
    for (const w of widgets) {
      try {
        if (contains(zr, widgetBoundingBox(w))) out.push(w);
      } catch {
        /* skip widgets we can't bound */
      }
    }
    return out;
  }

  /** Widgets that touch or overlap `zone`. */
  widgetsTouchingZone<W extends HasBounds | Widget>(
    widgets: readonly W[],
    zone: WidgetZone,
  ): readonly W[] {
    const zr: Rectangle = {
      x: zone.location.x,
      y: zone.location.y,
      width: zone.size.width,
      height: zone.size.height,
    };
    const out: W[] = [];
    for (const w of widgets) {
      try {
        if (touches(zr, widgetBoundingBox(w))) out.push(w);
      } catch {
        /* skip */
      }
    }
    return out;
  }

  private computeBounds(widgets: readonly (HasBounds | Widget)[]): Rectangle {
    let bounds: Rectangle | undefined;
    for (const w of widgets) {
      try {
        const r = widgetBoundingBox(w);
        bounds = bounds === undefined ? r : getUnion(bounds, r);
      } catch {
        /* skip */
      }
    }
    if (bounds === undefined) throw new Error("computeBounds: no usable widgets");
    return bounds;
  }
}

/** Operation produced by {@link BatchWidgetOperations}. */
export interface WidgetOperation {
  readonly widgetId: Uuid;
  readonly operation: "move" | "resize";
  readonly payload: Record<string, unknown>;
}

/**
 * Pure functions that build PATCH payloads for bulk widget edits.
 *
 * These do NOT touch the network — call `widgets.updateAny` (or per-type
 * `update`) on the returned operations to apply them.
 */
export class BatchWidgetOperations {
  constructor(public readonly tolerance: SpatialTolerance = DEFAULT_SPATIAL_TOLERANCE) {}

  /** Build move payloads that shift each widget by (`offsetX`, `offsetY`). */
  moveWidgets(
    widgets: readonly (HasBounds | Widget)[],
    offsetX: number,
    offsetY: number,
  ): readonly WidgetOperation[] {
    const ops: WidgetOperation[] = [];
    for (const w of widgets) {
      const ww = w as { id?: Uuid; widget_type?: string; location?: Location };
      if (ww.id === undefined) continue;
      if (ww.widget_type === "Connector") {
        const c = w as unknown as { id: Uuid; src?: { rel_location?: Location }; dst?: { rel_location?: Location } };
        const src = c.src?.rel_location ?? { x: 0, y: 0 };
        const dst = c.dst?.rel_location ?? { x: 0, y: 0 };
        ops.push({
          widgetId: c.id,
          operation: "move",
          payload: {
            src: { rel_location: { x: src.x + offsetX, y: src.y + offsetY } },
            dst: { rel_location: { x: dst.x + offsetX, y: dst.y + offsetY } },
          },
        });
        continue;
      }
      if (ww.location === undefined) continue;
      ops.push({
        widgetId: ww.id,
        operation: "move",
        payload: {
          location: { x: ww.location.x + offsetX, y: ww.location.y + offsetY },
        },
      });
    }
    return ops;
  }

  /** Build resize payloads that scale each widget's `size` (or connector `line_width`). */
  resizeWidgets(
    widgets: readonly (HasBounds | Widget)[],
    scaleFactor: number,
  ): readonly WidgetOperation[] {
    const ops: WidgetOperation[] = [];
    for (const w of widgets) {
      const ww = w as { id?: Uuid; widget_type?: string; size?: Size };
      if (ww.id === undefined) continue;
      if (ww.widget_type === "Connector") {
        const c = w as unknown as { id: Uuid; line_width?: number };
        if (c.line_width === undefined) continue;
        ops.push({
          widgetId: c.id,
          operation: "resize",
          payload: { line_width: c.line_width * scaleFactor },
        });
        continue;
      }
      if (ww.size === undefined) continue;
      ops.push({
        widgetId: ww.id,
        operation: "resize",
        payload: {
          size: { width: ww.size.width * scaleFactor, height: ww.size.height * scaleFactor },
        },
      });
    }
    return ops;
  }

  /** Return widgets that fully enclose the target widget. */
  widgetsContainId<W extends HasBounds | Widget>(
    widgets: readonly W[],
    targetId: Uuid,
  ): readonly W[] {
    const target = widgets.find((w) => (w as { id?: Uuid }).id === targetId);
    if (target === undefined) return [];
    const out: W[] = [];
    for (const w of widgets) {
      if ((w as { id?: Uuid }).id === targetId) continue;
      try {
        if (widgetContains(w, target)) out.push(w);
      } catch {
        /* skip */
      }
    }
    return out;
  }

  /** Return widgets that touch the target widget. */
  widgetsTouchId<W extends HasBounds | Widget>(
    widgets: readonly W[],
    targetId: Uuid,
  ): readonly W[] {
    const target = widgets.find((w) => (w as { id?: Uuid }).id === targetId);
    if (target === undefined) return [];
    const out: W[] = [];
    for (const w of widgets) {
      if ((w as { id?: Uuid }).id === targetId) continue;
      try {
        if (widgetsTouch(w, target)) out.push(w);
      } catch {
        /* skip */
      }
    }
    return out;
  }
}

/**
 * Group widgets by spatial proximity. Two widgets join the same group when
 * their bounding-box top-left corners are within `tolerance` on both axes.
 */
export function createSpatialGroup<W extends HasBounds | Widget>(
  widgets: readonly W[],
  tolerance = 10,
): readonly (readonly W[])[] {
  if (widgets.length === 0) return [];
  const groups: W[][] = [];
  const processed = new Set<Uuid>();

  for (const seed of widgets) {
    const seedId = (seed as { id?: Uuid }).id;
    if (seedId === undefined || processed.has(seedId)) continue;
    const group: W[] = [seed];
    processed.add(seedId);
    let changed = true;
    while (changed) {
      changed = false;
      for (const other of widgets) {
        const otherId = (other as { id?: Uuid }).id;
        if (otherId === undefined || processed.has(otherId)) continue;
        for (const member of group) {
          try {
            const r1 = widgetBoundingBox(member);
            const r2 = widgetBoundingBox(other);
            if (Math.abs(r1.x - r2.x) <= tolerance && Math.abs(r1.y - r2.y) <= tolerance) {
              group.push(other);
              processed.add(otherId);
              changed = true;
              break;
            }
          } catch {
            /* skip */
          }
        }
      }
    }
    groups.push(group);
  }
  return groups;
}

/** Return groups of `>= minClusterSize` widgets produced by {@link createSpatialGroup}. */
export function findWidgetClusters<W extends HasBounds | Widget>(
  widgets: readonly W[],
  minClusterSize = 2,
  tolerance = 20,
): readonly (readonly W[])[] {
  return createSpatialGroup(widgets, tolerance).filter((g) => g.length >= minClusterSize);
}

/** Widgets-per-unit-area inside the given rectangle. */
export function calculateWidgetDensity(
  widgets: readonly (HasBounds | Widget)[],
  area: Rectangle,
): number {
  if (area.width <= 0 || area.height <= 0) return 0;
  const total = area.width * area.height;
  let inArea = 0;
  for (const w of widgets) {
    try {
      if (intersects(area, widgetBoundingBox(w))) inArea++;
    } catch {
      /* skip */
    }
  }
  return inArea / total;
}

/** Cheap, deterministic non-crypto hash for zone-ID generation. */
function simpleHash(s: string): number {
  let h = 0;
  for (let i = 0; i < s.length; i++) {
    h = (h << 5) - h + s.charCodeAt(i);
    h |= 0;
  }
  return Math.abs(h);
}
