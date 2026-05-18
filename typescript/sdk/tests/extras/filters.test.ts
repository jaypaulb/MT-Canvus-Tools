import { describe, expect, it } from "vitest";
import {
  combineFilters,
  createFilter,
  Filter,
  newSpatialCondition,
  newTextFilter,
  newWidgetTypeFilter,
  newWildcardCondition,
} from "../../src/extras/filters.js";

describe("Filter", () => {
  it("equals / not_equals / in operate as expected", () => {
    const f = createFilter().addCondition("widget_type", "equals", "Note");
    expect(f.matches({ widget_type: "Note" })).toBe(true);
    expect(f.matches({ widget_type: "Image" })).toBe(false);

    const fIn = newWidgetTypeFilter(["Note", "Image"]);
    expect(fIn.matches({ widget_type: "Note" })).toBe(true);
    expect(fIn.matches({ widget_type: "Pdf" })).toBe(false);
  });

  it("text contains / wildcard match", () => {
    const f = newTextFilter("hello", ["text"]);
    expect(f.matches({ text: "say hello world" })).toBe(true);
    expect(f.matches({ text: "no greet" })).toBe(false);

    const wf = newWildcardCondition("title", "draft-*");
    expect(wf.matches({ title: "draft-1" })).toBe(true);
    expect(wf.matches({ title: "final" })).toBe(false);
  });

  it("spatial condition (intersects)", () => {
    const f = newSpatialCondition({ x: 0, y: 0, width: 100, height: 100 }, "intersects");
    expect(
      f.matches({ location: { x: 10, y: 10 }, size: { width: 5, height: 5 } }),
    ).toBe(true);
    expect(
      f.matches({ location: { x: 500, y: 500 }, size: { width: 5, height: 5 } }),
    ).toBe(false);
  });

  it("combineFilters AND-merges conditions", () => {
    const a = newWidgetTypeFilter("Note");
    const b = newWildcardCondition("text", "*hello*");
    const c = combineFilters(a, b);
    expect(c.matches({ widget_type: "Note", text: "hello there" })).toBe(true);
    expect(c.matches({ widget_type: "Note", text: "bye" })).toBe(false);
  });

  it("toJSON / fromJSON round-trips", () => {
    const f = newWidgetTypeFilter("Note");
    const round = Filter.fromJSON(f.toJSON());
    expect(round.matches({ widget_type: "Note" })).toBe(true);
  });

  it("dot-notation nested field lookup", () => {
    const f = createFilter().addCondition("location.x", "greater_than", 10);
    expect(f.matches({ location: { x: 20 } })).toBe(true);
    expect(f.matches({ location: { x: 5 } })).toBe(false);
  });
});
