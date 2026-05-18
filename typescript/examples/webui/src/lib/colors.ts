/**
 * Team-color generation — direct port of the legacy server.js HSL
 * helpers. 7 teams (Red/Orange/Yellow/Green/Blue/Indigo/Violet);
 * each user gets a ±10% lightness variation of their team's base.
 *
 * Format: `#RRGGBBAA` (8-char hex including alpha). Matches what the
 * legacy server produced and what the Canvus API accepts for note
 * `background_color`.
 */

export type TeamNumber = 1 | 2 | 3 | 4 | 5 | 6 | 7;

const TEAM_BASE_COLORS: Record<TeamNumber, string> = {
  1: "#FF0000FF", // Red
  2: "#FF7F00FF", // Orange
  3: "#FFFF00FF", // Yellow
  4: "#00FF00FF", // Green
  5: "#0000FFFF", // Blue
  6: "#4B0082FF", // Indigo
  7: "#8B00FFFF", // Violet
};

export function getTeamBaseColor(team: TeamNumber): string {
  return TEAM_BASE_COLORS[team];
}

export function isValidTeam(n: number): n is TeamNumber {
  return Number.isInteger(n) && n >= 1 && n <= 7;
}

interface Hsl {
  readonly h: number;
  readonly s: number;
  readonly l: number;
}

export function hexToHsl(hex: string): Hsl {
  const h6 = hex.replace(/^#/, "").slice(0, 6);
  const r = parseInt(h6.substring(0, 2), 16) / 255;
  const g = parseInt(h6.substring(2, 4), 16) / 255;
  const b = parseInt(h6.substring(4, 6), 16) / 255;

  const max = Math.max(r, g, b);
  const min = Math.min(r, g, b);
  let h = 0;
  let s = 0;
  const l = (max + min) / 2;

  if (max !== min) {
    const d = max - min;
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
    if (max === r) h = (g - b) / d + (g < b ? 6 : 0);
    else if (max === g) h = (b - r) / d + 2;
    else h = (r - g) / d + 4;
    h /= 6;
  }
  return { h: h * 360, s: s * 100, l: l * 100 };
}

export function hslToHex(h: number, s: number, l: number): string {
  const hh = h / 360;
  const ss = s / 100;
  const ll = l / 100;
  let r: number;
  let g: number;
  let b: number;

  if (ss === 0) {
    r = g = b = ll;
  } else {
    const hue2rgb = (p: number, q: number, t: number): number => {
      if (t < 0) t += 1;
      if (t > 1) t -= 1;
      if (t < 1 / 6) return p + (q - p) * 6 * t;
      if (t < 1 / 2) return q;
      if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6;
      return p;
    };
    const q = ll < 0.5 ? ll * (1 + ss) : ll + ss - ll * ss;
    const p = 2 * ll - q;
    r = hue2rgb(p, q, hh + 1 / 3);
    g = hue2rgb(p, q, hh);
    b = hue2rgb(p, q, hh - 1 / 3);
  }

  const toHex = (x: number): string => {
    const v = Math.round(x * 255).toString(16);
    return v.length === 1 ? `0${v}` : v;
  };

  return `#${toHex(r)}${toHex(g)}${toHex(b)}FF`;
}

/** Generate a ±10% lightness variation of `baseColor`. */
export function generateColorVariation(baseColor: string): string {
  const { h, s, l } = hexToHsl(baseColor);
  let newL = l + (Math.random() * 20 - 10);
  newL = Math.min(100, Math.max(0, newL));
  return hslToHex(h, s, newL);
}
