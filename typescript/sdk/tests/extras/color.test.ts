import { describe, expect, it } from "vitest";
import {
  colorToRgb,
  colorToRgba,
  ColorWhite,
  ColorTransparent,
  colorWithAlpha,
  normalizeColor,
  rgbaToColor,
  validateColor,
} from "../../src/extras/color.js";

describe("color utilities", () => {
  it("validateColor accepts well-formed RRGGBBAA", () => {
    expect(() => validateColor("FF00AABB")).not.toThrow();
    expect(() => validateColor("nope")).toThrow();
    expect(() => validateColor("ff00aabb")).toThrow();
  });

  it("normalizeColor pads RRGGBB with FF and strips leading #", () => {
    expect(normalizeColor("#ff0000")).toBe("FF0000FF");
    expect(normalizeColor("ABCDEF")).toBe("ABCDEFFF");
    expect(normalizeColor("#ABCDEF12")).toBe("ABCDEF12");
  });

  it("colorToRgba / rgbaToColor round-trips", () => {
    expect(colorToRgba("AABBCCDD")).toEqual([0xaa, 0xbb, 0xcc, 0xdd]);
    expect(rgbaToColor(0xaa, 0xbb, 0xcc, 0xdd)).toBe("AABBCCDD");
  });

  it("colorToRgb drops alpha and prepends #", () => {
    expect(colorToRgb("AABBCCDD")).toBe("#AABBCC");
  });

  it("colorWithAlpha replaces last byte", () => {
    expect(colorWithAlpha("AABBCC11", 0xff)).toBe("AABBCCFF");
  });

  it("constants are well-formed", () => {
    expect(ColorWhite).toBe("FFFFFFFF");
    expect(ColorTransparent).toBe("00000000");
  });
});
