/**
 * Zone bounding-box + spatial helpers — direct port of the legacy
 * server.js helpers (`getZoneBoundingBox`, `widgetIsInZone`,
 * `transformWidgetLocationAndScale`).
 *
 * Semantics preserved verbatim: legacy `widgetIsInZone` checks only the
 * widget's `location.x/y` POINT (not its full bounding box) — a widget
 * that "sticks out" of a zone but whose top-left corner is inside it
 * still counts as in-zone. The `extras/widgetOperations.WidgetZoneManager`
 * helper uses full-bbox containment which would change behaviour, so we
 * keep the legacy POINT semantics here. Document this divergence so
 * future readers don't "improve" it.
 */

import type { Session } from "@mt-canvus-tools/sdk";

export interface ZoneBoundingBox {
  readonly x: number;
  readonly y: number;
  readonly width: number;
  readonly height: number;
  readonly scale: number;
}

/**
 * Fetch a zone (= anchor widget) and read its bounding box.
 *
 * @throws Error when the anchor's `location` or `size` is missing.
 */
export async function getZoneBoundingBox(
  session: Session,
  canvasId: string,
  zoneId: string,
): Promise<ZoneBoundingBox> {
  const anchor = await session.widgets.anchors.get(canvasId, zoneId);
  const loc = (anchor as unknown as { location?: { x: number; y: number } }).location;
  const size = (anchor as unknown as { size?: { width: number; height: number } }).size;
  if (!loc || !size) {
    throw new Error(`Invalid or missing anchor data for zone ID: ${zoneId}`);
  }
  return {
    x: loc.x,
    y: loc.y,
    width: size.width,
    height: size.height,
    scale: (anchor as unknown as { scale?: number }).scale ?? 1,
  };
}

/**
 * Legacy point-containment check: a widget is "in" the zone if its
 * `location.{x,y}` lies within the zone, inset by 2px on each side.
 * The 2px inset is intentional — copied from the legacy server.js so
 * widgets sitting exactly on a zone border aren't double-counted.
 */
export function widgetIsInZone(
  widget: { readonly location?: { x: number; y: number } | undefined; readonly id?: string },
  zoneBB: ZoneBoundingBox,
): boolean {
  const loc = widget.location;
  if (!loc) return false;
  const withinX = loc.x >= zoneBB.x + 2 && loc.x <= zoneBB.x + zoneBB.width - 2;
  const withinY = loc.y >= zoneBB.y + 2 && loc.y <= zoneBB.y + zoneBB.height - 2;
  return withinX && withinY;
}

/**
 * Translate + scale a widget so it lands in the target zone at the
 * same relative position it had in the source zone. Mutates `widget`
 * (intentional — legacy semantics; callers clone first if needed).
 */
export function transformWidgetLocationAndScale<
  W extends { location?: { x: number; y: number }; scale?: number },
>(widget: W, sourceBB: ZoneBoundingBox, targetBB: ZoneBoundingBox): W {
  if (!widget.location) return widget;
  const scaleFactor = targetBB.width / sourceBB.width;
  const deltaX = widget.location.x - sourceBB.x;
  const deltaY = widget.location.y - sourceBB.y;
  widget.location = {
    x: targetBB.x + deltaX * scaleFactor,
    y: targetBB.y + deltaY * scaleFactor,
  };
  const oldScale = widget.scale ?? 1;
  widget.scale = oldScale * scaleFactor;
  return widget;
}

/**
 * Per-widget-type PATCH dispatcher. The legacy server.js used the
 * generic `/widgets/{id}` route in some places and per-type routes in
 * others; the SDK now exposes per-type sub-resources, which is cleaner.
 */
type WidgetTypeKey = "note" | "image" | "video" | "pdf" | "browser" | "anchor" | "connector" | "table" | "video_input" | "ip_video" | "rdp_connection";

const WIDGET_TYPE_NORMALISE: Record<string, WidgetTypeKey> = {
  note: "note",
  image: "image",
  video: "video",
  pdf: "pdf",
  browser: "browser",
  anchor: "anchor",
  connector: "connector",
  table: "table",
  videoinput: "video_input",
  video_input: "video_input",
  "video-input": "video_input",
  ipvideo: "ip_video",
  ip_video: "ip_video",
  "ip-video": "ip_video",
  rdpconnection: "rdp_connection",
  rdp_connection: "rdp_connection",
  "rdp-connection": "rdp_connection",
};

/** Best-effort widget-type → patch helper. Returns a callable. */
export function patchHelperFor(
  session: Session,
  widget: { widget_type?: string },
): (canvasId: string, widgetId: string, body: Record<string, unknown>) => Promise<unknown> {
  const raw = (widget.widget_type ?? "note").toLowerCase();
  const kind = WIDGET_TYPE_NORMALISE[raw] ?? "note";
  // Use updateAny via session.widgets to defer dispatch; this avoids
  // bespoke per-type if-chains and benefits from the SDK's grid_size
  // filtering for tables (changelog §4).
  return (canvasId, widgetId, body) =>
    session.widgets.updateAny(canvasId, widgetId, {
      ...body,
      widget_type: kind === "video_input" ? "VideoInput" :
        kind === "ip_video" ? "IpVideo" :
        kind === "rdp_connection" ? "RdpConnection" :
        // capitalise
        kind.charAt(0).toUpperCase() + kind.slice(1),
    });
}
