// Phase 4b §4.3 #15: filters port.
//
// Advanced client-side filtering of widgets and canvases. Ported from
// CanvusPythonAPI/canvus_api/filters.py:11-323.

import { contains, intersects, type Rectangle } from "./geometry.js";

/** Operators understood by the {@link Filter} engine. */
export type FilterOperator =
  | "equals"
  | "not_equals"
  | "contains"
  | "not_contains"
  | "starts_with"
  | "ends_with"
  | "greater_than"
  | "less_than"
  | "greater_equal"
  | "less_equal"
  | "in"
  | "not_in"
  | "exists"
  | "not_exists"
  | "spatial_intersects"
  | "spatial_contains"
  | "spatial_within"
  | "wildcard_match";

/**
 * A single filter condition.
 *
 * `field` may use dot notation (e.g. `location.x`) to address nested
 * properties on the item being matched.
 */
export interface FilterCondition {
  readonly field: string;
  readonly operator: FilterOperator;
  readonly value: unknown;
}

/**
 * A composable filter expression evaluated against arbitrary record-shaped
 * items (widgets, canvases). Built by chaining `add*` methods.
 *
 * Conditions are combined with AND semantics; OR semantics can be
 * approximated by running the same item through multiple filters and
 * combining the booleans externally.
 */
export class Filter {
  /** Materialised list of conditions. Public for serialisation; do not mutate. */
  public readonly conditions: FilterCondition[];

  constructor(conditions: readonly FilterCondition[] = []) {
    this.conditions = [...conditions];
  }

  /** Append a generic comparison. Returns `this` for chaining. */
  addCondition(field: string, operator: FilterOperator, value: unknown): this {
    this.conditions.push({ field, operator, value });
    return this;
  }

  /**
   * Append a spatial condition.
   *
   * `op` is one of `intersects`, `contains`, `within` — internally these
   * become `spatial_intersects`/`spatial_contains`/`spatial_within`.
   */
  addSpatialCondition(op: "intersects" | "contains" | "within", area: Rectangle): this {
    const operator: FilterOperator =
      op === "intersects"
        ? "spatial_intersects"
        : op === "contains"
          ? "spatial_contains"
          : "spatial_within";
    this.conditions.push({ field: "spatial", operator, value: area });
    return this;
  }

  /**
   * Append a wildcard match (`*` for any sequence, `?` for one char).
   * Internally a case-insensitive RegExp.
   */
  addWildcardCondition(field: string, pattern: string): this {
    this.conditions.push({ field, operator: "wildcard_match", value: pattern });
    return this;
  }

  /** True if every condition matches the item. */
  matches(item: Record<string, unknown>): boolean {
    return this.conditions.every((c) => matchesCondition(item, c));
  }

  /** Serialise to a plain object for transport / storage. */
  toJSON(): { conditions: FilterCondition[] } {
    return { conditions: this.conditions };
  }

  /** Reconstruct a filter from {@link toJSON} output. */
  static fromJSON(data: { conditions?: readonly FilterCondition[] }): Filter {
    return new Filter(data.conditions ?? []);
  }
}

/** Create an empty filter. */
export function createFilter(): Filter {
  return new Filter();
}

/** Build a filter that selects items inside / overlapping / fully within `area`. */
export function newSpatialCondition(
  area: Rectangle,
  op: "intersects" | "contains" | "within" = "intersects",
): Filter {
  return new Filter().addSpatialCondition(op, area);
}

/** Build a filter that matches items whose `field` matches the wildcard `pattern`. */
export function newWildcardCondition(field: string, pattern: string): Filter {
  return new Filter().addWildcardCondition(field, pattern);
}

/**
 * Build a filter that finds `text` anywhere in any of `fields`
 * (defaults to `title`, `text`, `description`).
 *
 * All conditions are AND-combined, mirroring the Python port. To get OR
 * semantics across fields, evaluate each as a separate filter.
 */
export function newTextFilter(text: string, fields: readonly string[] = ["title", "text", "description"]): Filter {
  const f = new Filter();
  for (const field of fields) {
    f.addCondition(field, "contains", text);
  }
  return f;
}

/** Build a filter that selects widgets whose `widget_type` is one of the given values. */
export function newWidgetTypeFilter(types: string | readonly string[]): Filter {
  const arr = typeof types === "string" ? [types] : [...types];
  return new Filter().addCondition("widget_type", "in", arr);
}

/**
 * Combine multiple filters into one whose conditions are the concatenation
 * (AND semantics). For OR, run each filter independently and merge results.
 */
export function combineFilters(...filters: readonly Filter[]): Filter {
  const out = new Filter();
  for (const f of filters) {
    for (const c of f.conditions) {
      out.conditions.push(c);
    }
  }
  return out;
}

// ---------------------------------------------------------------------------
// Internal — condition matching
// ---------------------------------------------------------------------------

function matchesCondition(item: Record<string, unknown>, condition: FilterCondition): boolean {
  const { field, operator, value } = condition;
  if (field === "spatial" && operator.startsWith("spatial_")) {
    return matchesSpatial(item, operator, value as Rectangle);
  }
  const fieldValue = getNestedValue(item, field);
  switch (operator) {
    case "equals":
      return fieldValue === value;
    case "not_equals":
      return fieldValue !== value;
    case "contains":
      return typeof fieldValue === "string" && typeof value === "string"
        ? fieldValue.includes(value)
        : Array.isArray(fieldValue) && (fieldValue as unknown[]).includes(value);
    case "not_contains":
      if (typeof fieldValue === "string" && typeof value === "string") {
        return !fieldValue.includes(value);
      }
      if (Array.isArray(fieldValue)) {
        return !(fieldValue as unknown[]).includes(value);
      }
      return true;
    case "starts_with":
      return typeof fieldValue === "string" && typeof value === "string"
        ? fieldValue.startsWith(value)
        : false;
    case "ends_with":
      return typeof fieldValue === "string" && typeof value === "string"
        ? fieldValue.endsWith(value)
        : false;
    case "greater_than":
      return fieldValue !== null && fieldValue !== undefined && (fieldValue as number) > (value as number);
    case "less_than":
      return fieldValue !== null && fieldValue !== undefined && (fieldValue as number) < (value as number);
    case "greater_equal":
      return fieldValue !== null && fieldValue !== undefined && (fieldValue as number) >= (value as number);
    case "less_equal":
      return fieldValue !== null && fieldValue !== undefined && (fieldValue as number) <= (value as number);
    case "in":
      return Array.isArray(value) && (value as unknown[]).includes(fieldValue);
    case "not_in":
      return Array.isArray(value) ? !(value as unknown[]).includes(fieldValue) : true;
    case "exists":
      return fieldValue !== undefined && fieldValue !== null;
    case "not_exists":
      return fieldValue === undefined || fieldValue === null;
    case "wildcard_match":
      return matchesWildcard(fieldValue, value as string);
    case "spatial_intersects":
    case "spatial_contains":
    case "spatial_within":
      return matchesSpatial(item, operator, value as Rectangle);
  }
}

function getNestedValue(item: Record<string, unknown>, field: string): unknown {
  if (!field.includes(".")) return item[field];
  const parts = field.split(".");
  let current: unknown = item;
  for (const p of parts) {
    if (current !== null && typeof current === "object" && p in (current as Record<string, unknown>)) {
      current = (current as Record<string, unknown>)[p];
    } else {
      return undefined;
    }
  }
  return current;
}

function matchesSpatial(
  item: Record<string, unknown>,
  operator: FilterOperator,
  area: Rectangle,
): boolean {
  const loc = item.location as { x?: number; y?: number } | undefined;
  const sz = item.size as { width?: number; height?: number } | undefined;
  if (!loc || !sz || loc.x === undefined || loc.y === undefined || sz.width === undefined || sz.height === undefined) {
    return false;
  }
  const rect: Rectangle = { x: loc.x, y: loc.y, width: sz.width, height: sz.height };
  if (operator === "spatial_intersects") return intersects(rect, area);
  if (operator === "spatial_contains") return intersects(area, rect);
  if (operator === "spatial_within") return contains(area, rect);
  return false;
}

function matchesWildcard(fieldValue: unknown, pattern: string): boolean {
  if (fieldValue === null || fieldValue === undefined) return false;
  let str: string;
  if (typeof fieldValue === "string") {
    str = fieldValue;
  } else if (
    typeof fieldValue === "number" ||
    typeof fieldValue === "boolean" ||
    typeof fieldValue === "bigint"
  ) {
    str = fieldValue.toString();
  } else {
    try {
      str = JSON.stringify(fieldValue);
    } catch {
      return false;
    }
  }
  const escaped = pattern
    .replace(/[.+^${}()|[\]\\]/g, "\\$&")
    .replace(/\*/g, ".*")
    .replace(/\?/g, ".");
  try {
    return new RegExp(`^${escaped}$`, "i").test(str);
  } catch {
    return false;
  }
}
