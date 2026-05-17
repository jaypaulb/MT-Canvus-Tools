/**
 * Shared primitive shapes used across the Canvus data model.
 *
 * The Canvus API returns kebab-case JSON keys. These types model the wire
 * shape exactly — no key transformation is performed by the SDK so that
 * what you see in `curl` is what you get in TypeScript.
 */

/** A 2D point in absolute canvas pixel coordinates. */
export interface Location {
  readonly x: number;
  readonly y: number;
}

/** Width and height in pixels. */
export interface Size {
  readonly width: number;
  readonly height: number;
}

/** ISO 8601 timestamp string, e.g. `2025-05-17T14:22:15Z`. */
export type IsoDateTime = string;

/** UUID string. The server validates format; the SDK does not. */
export type Uuid = string;

/** Hex colour string, e.g. `#FFAA00`. */
export type HexColor = string;

/** Hex hash string used to identify uploaded assets. */
export type AssetHash = string;

/**
 * Standard server error envelope.
 *
 * Most Canvus error responses follow this shape; the SDK exposes the raw
 * `body` on {@link APIError} so callers can inspect non-standard shapes.
 */
export interface ServerErrorBody {
  readonly msg?: string;
  readonly error?: string;
}

/** Generic pagination envelope for list responses with totals. */
export interface PagedResult<T> {
  readonly events?: readonly T[];
  readonly items?: readonly T[];
  readonly "total-count": number;
  readonly page: number;
  readonly "per-page": number;
}
