// Phase 4b §4.3 #17: search port.
//
// Cross-canvas widget search. Ported from
// CanvusPythonAPI/canvus_api/search.py:16-577.

import type { Session } from "../session.js";
import type { Canvas } from "../types/canvas.js";
import type { Uuid } from "../types/common.js";
import type { Widget } from "../types/widget.js";
import { intersects, widgetBoundingBox, type Rectangle } from "./geometry.js";

/** Result of a cross-canvas search. */
export interface SearchResult {
  readonly canvasId: Uuid;
  readonly canvasName: string;
  readonly widgetId: Uuid;
  readonly widgetType: string;
  readonly widget: Widget;
  /** Match score in `[0, 1]`. 1 = exact, 0 = no match. */
  readonly matchScore: number;
  /** Human-readable reason for the match. */
  readonly matchReason: string;
  /** `<canvasId>:<widgetId>` for drill-down navigation. */
  readonly drillDownPath: string;
}

/** Search options accepted by {@link CrossCanvasSearch.findWidgetsAcrossCanvases}. */
export interface CrossCanvasSearchOptions {
  readonly canvasIds?: readonly Uuid[];
  readonly widgetTypes?: readonly string[];
  readonly spatialFilter?: Rectangle;
  readonly maxResults?: number;
  readonly includeDeleted?: boolean;
}

/**
 * Cross-canvas widget search engine.
 *
 * Wraps a {@link Session} to list canvases + widgets and apply filter
 * criteria client-side. Query can be a plain string (treated as wildcard
 * text search) or a record of `field -> expected value`.
 */
export class CrossCanvasSearch {
  constructor(private readonly session: Session) {}

  /**
   * Search every selected canvas for widgets matching `query`.
   *
   * @param query - Either a record of `field -> expected value` (exact or
   *   wildcard string), or a free-text string that's wrapped as
   *   `{ text: "*<query>*" }`.
   * @param options - Optional filters and limits.
   */
  async findWidgetsAcrossCanvases(
    query: string | Record<string, unknown>,
    options: CrossCanvasSearchOptions = {},
  ): Promise<readonly SearchResult[]> {
    const {
      canvasIds,
      widgetTypes,
      spatialFilter,
      maxResults = 100,
      includeDeleted = false,
    } = options;
    const criteria = parseQuery(query);
    const canvases = await this.collectCanvases(canvasIds);
    const results: SearchResult[] = [];
    for (const canvas of canvases) {
      let widgets: readonly Widget[];
      try {
        widgets = await this.session.widgets.list(canvas.id);
      } catch {
        continue;
      }
      const filtered = this.applyFilters(widgets, criteria, widgetTypes, spatialFilter, includeDeleted);
      for (const w of filtered) {
        results.push({
          canvasId: canvas.id,
          canvasName: canvas.name,
          widgetId: w.id,
          widgetType: w.widget_type,
          widget: w,
          matchScore: matchScore(w, criteria),
          matchReason: matchReason(w, criteria),
          drillDownPath: `${canvas.id}:${w.id}`,
        });
        if (results.length >= maxResults) break;
      }
      if (results.length >= maxResults) break;
    }
    results.sort((a, b) => b.matchScore - a.matchScore);
    return results.slice(0, maxResults);
  }

  /** Convenience: search widget `text`/`title`/`description` (case-insensitive). */
  async findWidgetsByText(
    text: string,
    options: Omit<CrossCanvasSearchOptions, "widgetTypes" | "spatialFilter"> & {
      readonly caseSensitive?: boolean;
    } = {},
  ): Promise<readonly SearchResult[]> {
    const { caseSensitive = false, ...rest } = options;
    const query = caseSensitive ? { text } : { text: `*${text}*` };
    return this.findWidgetsAcrossCanvases(query, rest);
  }

  /** Convenience: search by exact `widget_type`. */
  async findWidgetsByType(
    widgetType: string,
    options: Omit<CrossCanvasSearchOptions, "widgetTypes" | "spatialFilter"> = {},
  ): Promise<readonly SearchResult[]> {
    return this.findWidgetsAcrossCanvases({ widget_type: widgetType }, options);
  }

  /** Convenience: search widgets in a spatial area. */
  async findWidgetsInArea(
    area: Rectangle,
    options: Omit<CrossCanvasSearchOptions, "spatialFilter"> = {},
  ): Promise<readonly SearchResult[]> {
    return this.findWidgetsAcrossCanvases({}, { ...options, spatialFilter: area });
  }

  /** Convenience: search by a (possibly nested) property value. */
  async findWidgetsByProperty(
    propertyPath: string,
    value: unknown,
    options: Omit<CrossCanvasSearchOptions, "widgetTypes" | "spatialFilter"> = {},
  ): Promise<readonly SearchResult[]> {
    return this.findWidgetsAcrossCanvases({ [propertyPath]: value }, options);
  }

  private async collectCanvases(canvasIds?: readonly Uuid[]): Promise<readonly Canvas[]> {
    if (canvasIds === undefined || canvasIds.length === 0) {
      return this.session.canvases.list();
    }
    const out: Canvas[] = [];
    for (const id of canvasIds) {
      try {
        out.push(await this.session.canvases.get(id));
      } catch {
        /* skip */
      }
    }
    return out;
  }

  private applyFilters(
    widgets: readonly Widget[],
    criteria: Record<string, unknown>,
    widgetTypes: readonly string[] | undefined,
    spatialFilter: Rectangle | undefined,
    includeDeleted: boolean,
  ): readonly Widget[] {
    const typeSet = widgetTypes ? new Set(widgetTypes.map((t) => t.toLowerCase())) : undefined;
    return widgets.filter((w) => {
      if (!includeDeleted && (w as { state?: string }).state === "deleted") return false;
      if (typeSet && !typeSet.has(w.widget_type.toLowerCase())) return false;
      if (spatialFilter) {
        try {
          if (!intersects(widgetBoundingBox(w), spatialFilter)) return false;
        } catch {
          return false;
        }
      }
      if (Object.keys(criteria).length > 0 && !matchesCriteria(w, criteria)) return false;
      return true;
    });
  }
}

// ---------------------------------------------------------------------------
// Internal — query parsing and matching
// ---------------------------------------------------------------------------

function parseQuery(query: string | Record<string, unknown>): Record<string, unknown> {
  if (typeof query !== "string") return query;
  // Try JSON, else treat as wildcard text search.
  try {
    const parsed = JSON.parse(query) as unknown;
    if (typeof parsed === "object" && parsed !== null) {
      return parsed as Record<string, unknown>;
    }
  } catch {
    /* fall through */
  }
  return { text: `*${query}*` };
}

function matchesCriteria(widget: Widget, criteria: Record<string, unknown>): boolean {
  const obj = widget as unknown as Record<string, unknown>;
  for (const [key, expected] of Object.entries(criteria)) {
    const actual = key.includes(".") ? getNested(obj, key) : obj[key];
    if (actual === undefined) return false;
    const actualStr = stringifyForCompare(actual);
    if (typeof expected === "string" && expected.includes("*")) {
      if (!wildcardMatch(actualStr, expected)) return false;
      continue;
    }
    const expectedStr = stringifyForCompare(expected);
    if (key === "widget_type") {
      if (actualStr.toLowerCase() !== expectedStr.toLowerCase()) return false;
      continue;
    }
    if (actualStr !== expectedStr) {
      const af = Number(actual);
      const ef = Number(expected);
      if (Number.isFinite(af) && Number.isFinite(ef)) {
        if (af !== ef) return false;
        continue;
      }
      return false;
    }
  }
  return true;
}

/** Safe stringification that avoids `[object Object]` for unknown shapes. */
function stringifyForCompare(value: unknown): string {
  if (value === null || value === undefined) return "";
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "boolean" || typeof value === "bigint") {
    return value.toString();
  }
  try {
    return JSON.stringify(value);
  } catch {
    return Object.prototype.toString.call(value);
  }
}

function matchScore(widget: Widget, criteria: Record<string, unknown>): number {
  if (Object.keys(criteria).length === 0) return 1;
  return matchesCriteria(widget, criteria) ? 1 : 0;
}

function matchReason(widget: Widget, criteria: Record<string, unknown>): string {
  if (Object.keys(criteria).length === 0) return "No filter applied";
  const obj = widget as unknown as Record<string, unknown>;
  for (const [k, v] of Object.entries(criteria)) {
    if (k in obj) {
      const a = stringifyForCompare(obj[k]);
      const b = stringifyForCompare(v);
      if (a === b) return `Exact match on ${k}`;
      if (a.includes(b)) return `Partial match on ${k}`;
    }
  }
  return "Filter criteria matched";
}

function getNested(obj: Record<string, unknown>, path: string): unknown {
  let current: unknown = obj;
  for (const p of path.split(".")) {
    if (current !== null && typeof current === "object" && p in (current as Record<string, unknown>)) {
      current = (current as Record<string, unknown>)[p];
    } else {
      return undefined;
    }
  }
  return current;
}

function wildcardMatch(text: string, pattern: string): boolean {
  const escaped = pattern.replace(/[.+^${}()|[\]\\]/g, "\\$&").replace(/\*/g, ".*").replace(/\?/g, ".");
  try {
    return new RegExp(escaped, "i").test(text);
  } catch {
    return false;
  }
}
