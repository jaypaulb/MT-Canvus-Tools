// Phase 4b §4.3 #20: color utilities.
//
// Mirrors Go's `color.go`. All Canvus colors are 8-character uppercase hex
// strings in RRGGBBAA format (alpha last; 00 transparent, FF opaque).

const RGBA_RE = /^[0-9A-F]{8}$/;
const RGB_RE = /^[0-9A-F]{6}$/;

/** Throws if `color` is not a valid uppercase RRGGBBAA string. */
export function validateColor(color: string): void {
  if (color.length !== 8) {
    throw new Error(`color must be exactly 8 characters (RRGGBBAA), got ${color.length.toString()}`);
  }
  if (!RGBA_RE.test(color)) {
    throw new Error(`color must be uppercase hex RRGGBBAA format, got ${JSON.stringify(color)}`);
  }
}

/**
 * Normalise an input colour to Canvus uppercase RRGGBBAA.
 *
 * Accepts: `RRGGBBAA`, `RRGGBB` (alpha defaults to `FF`), or either with
 * a leading `#`. Throws on anything else.
 */
export function normalizeColor(input: string): string {
  let color = input.startsWith("#") ? input.slice(1) : input;
  color = color.toUpperCase();
  if (color.length === 6) {
    if (!RGB_RE.test(color)) {
      throw new Error(`invalid 6-character color format: ${JSON.stringify(input)}`);
    }
    return `${color}FF`;
  }
  validateColor(color);
  return color;
}

/** Convert RRGGBBAA → `[r, g, b, a]` (each 0–255). */
export function colorToRgba(color: string): readonly [number, number, number, number] {
  validateColor(color);
  return [
    Number.parseInt(color.slice(0, 2), 16),
    Number.parseInt(color.slice(2, 4), 16),
    Number.parseInt(color.slice(4, 6), 16),
    Number.parseInt(color.slice(6, 8), 16),
  ];
}

/** Compose RRGGBBAA from 0–255 components. */
export function rgbaToColor(r: number, g: number, b: number, a: number): string {
  return [r, g, b, a]
    .map((n) => clamp(n).toString(16).padStart(2, "0").toUpperCase())
    .join("");
}

/** Convert RRGGBBAA → `#RRGGBB` (alpha dropped). */
export function colorToRgb(color: string): string {
  validateColor(color);
  return `#${color.slice(0, 6)}`;
}

/** Return a copy of `color` with its alpha channel replaced. */
export function colorWithAlpha(color: string, alpha: number): string {
  validateColor(color);
  return color.slice(0, 6) + clamp(alpha).toString(16).padStart(2, "0").toUpperCase();
}

function clamp(n: number): number {
  if (!Number.isFinite(n)) return 0;
  return Math.max(0, Math.min(255, Math.round(n)));
}

/** Common opaque colors for convenience. */
export const ColorBlack = "000000FF";
export const ColorWhite = "FFFFFFFF";
export const ColorRed = "FF0000FF";
export const ColorGreen = "00FF00FF";
export const ColorBlue = "0000FFFF";
export const ColorYellow = "FFFF00FF";
export const ColorCyan = "00FFFFFF";
export const ColorMagenta = "FF00FFFF";
export const ColorGray = "808080FF";
export const ColorLightGray = "D3D3D3FF";
export const ColorDarkGray = "404040FF";
export const ColorTransparent = "00000000";
