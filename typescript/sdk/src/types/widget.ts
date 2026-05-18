import type { AssetHash, HexColor, Location, Size, Uuid } from "./common.js";

/**
 * Discriminator for every widget type known to the Canvus API.
 *
 * Note: `ip-video` and `rdp-connection` are intentionally included so that
 * GET/PATCH/DELETE remain typed, but POST helpers for these two types are
 * not exposed — see {@link IpVideo} / {@link RdpConnection} and the
 * changelog §2 entry.
 */
export type WidgetKind =
  | "note"
  | "image"
  | "video"
  | "pdf"
  | "browser"
  | "anchor"
  | "connector"
  | "table"
  | "video-input"
  | "ip-video"
  | "rdp-connection";

/** Fields common to every widget. */
export interface BaseWidget {
  readonly "widget-id": Uuid;
  readonly "widget-type": WidgetKind;
  readonly location: Location;
  readonly depth: number;
  readonly size: Size;
  readonly scale: number;
  readonly "is-pinned": boolean;
}

/** Fields that any widget can accept on create/update. */
export interface WidgetMutableBase {
  readonly location?: Location;
  readonly depth?: number;
  readonly size?: Size;
  readonly scale?: number;
  readonly "is-pinned"?: boolean;
}

/** Source-widget reference used by the cross-canvas clone endpoint.
 * The clone source fields use underscores (`source_canvas_id`) per the
 * Phase 4 changelog §1 source note — most other API fields use hyphens. */
export interface CloneSource {
  readonly source_canvas_id: Uuid;
  readonly source_widget_id: Uuid;
  readonly location?: Location;
}

// ---------- Note ----------

export interface Note extends BaseWidget {
  readonly "widget-type": "note";
  readonly "note-id": Uuid;
  readonly text: string;
  readonly title?: string;
  readonly "background-color": HexColor;
  readonly "text-color": HexColor;
  readonly "auto-text-color": boolean;
}

export interface CreateNoteRequest extends WidgetMutableBase {
  readonly text?: string;
  readonly title?: string;
  readonly "background-color"?: HexColor;
  readonly "text-color"?: HexColor;
  readonly "auto-text-color"?: boolean;
}

export type UpdateNoteRequest = CreateNoteRequest;

// ---------- Image ----------

export interface Image extends BaseWidget {
  readonly "widget-type": "image";
  readonly "image-id": Uuid;
  readonly "original-filename": string;
  readonly title?: string;
  readonly "asset-hash": AssetHash;
  readonly "asset-hash-private"?: AssetHash;
  readonly "mime-type": string;
  readonly "file-size": number;
}

export interface ImageMetadata extends WidgetMutableBase {
  readonly title?: string;
}

export type UpdateImageRequest = ImageMetadata;

// ---------- Video ----------

export type VideoPlaybackState = "playing" | "paused" | "stopped";

export interface Video extends BaseWidget {
  readonly "widget-type": "video";
  readonly "video-id": Uuid;
  readonly "original-filename": string;
  readonly title?: string;
  readonly "asset-hash": AssetHash;
  readonly "mime-type": string;
  readonly "file-size": number;
  readonly "seek-position"?: number;
  readonly "seek-timestamp"?: string;
  readonly "playback-state": VideoPlaybackState;
  readonly muted: boolean;
  readonly duration?: string;
}

export interface VideoMetadata extends WidgetMutableBase {
  readonly title?: string;
  readonly "seek-position"?: number;
  readonly "seek-timestamp"?: string;
  readonly "playback-state"?: VideoPlaybackState;
  readonly muted?: boolean;
}

export type UpdateVideoRequest = VideoMetadata;

// ---------- PDF ----------

export interface Pdf extends BaseWidget {
  readonly "widget-type": "pdf";
  readonly "pdf-id": Uuid;
  readonly "original-filename": string;
  readonly title?: string;
  readonly "asset-hash": AssetHash;
  readonly "mime-type": string;
  readonly "file-size": number;
  readonly index: number;
  readonly "page-count"?: number;
}

export interface PdfMetadata extends WidgetMutableBase {
  readonly title?: string;
  readonly index?: number;
}

export type UpdatePdfRequest = PdfMetadata;

// ---------- Browser ----------

export interface Browser extends BaseWidget {
  readonly "widget-type": "browser";
  readonly "browser-id": Uuid;
  readonly source: string;
  readonly title?: string;
  readonly "transparent-mode": boolean;
  readonly "main-frame-scroll-offset": Location;
}

export interface CreateBrowserRequest extends WidgetMutableBase {
  readonly source: string;
  readonly title?: string;
  readonly "transparent-mode"?: boolean;
}

export interface UpdateBrowserRequest extends WidgetMutableBase {
  readonly source?: string;
  readonly title?: string;
  readonly "transparent-mode"?: boolean;
  readonly "main-frame-scroll-offset"?: Location;
}

// ---------- Anchor ----------

export interface Anchor extends BaseWidget {
  readonly "widget-type": "anchor";
  readonly "anchor-id": Uuid;
  readonly "anchor-name": string;
}

export interface CreateAnchorRequest extends WidgetMutableBase {
  readonly "anchor-name"?: string;
}

export type UpdateAnchorRequest = CreateAnchorRequest;

// ---------- Connector ----------

export type ConnectorType = "line" | "curve" | "arrow";
export type ConnectorTip = "none" | "arrow" | "circle";

export interface Connector {
  readonly "widget-type": "connector";
  readonly "connector-id": Uuid;
  readonly "connector-type": ConnectorType;
  readonly src: Uuid;
  readonly "src-rel-location": Location;
  readonly "src-auto-location": boolean;
  readonly "src-tip": ConnectorTip;
  readonly dst: Uuid;
  readonly "dst-rel-location": Location;
  readonly "dst-auto-location": boolean;
  readonly "dst-tip": ConnectorTip;
  readonly "line-color": HexColor;
  readonly "line-width": number;
  readonly depth: number;
}

export interface CreateConnectorRequest {
  readonly src: Uuid;
  readonly dst: Uuid;
  readonly "connector-type"?: ConnectorType;
  readonly "src-rel-location"?: Location;
  readonly "src-auto-location"?: boolean;
  readonly "src-tip"?: ConnectorTip;
  readonly "dst-rel-location"?: Location;
  readonly "dst-auto-location"?: boolean;
  readonly "dst-tip"?: ConnectorTip;
  readonly "line-color"?: HexColor;
  readonly "line-width"?: number;
  readonly depth?: number;
}

export type UpdateConnectorRequest = Partial<Omit<CreateConnectorRequest, "src" | "dst">>;

// ---------- Table ----------

export interface GridSize {
  readonly columns: number;
  readonly rows: number;
}

/**
 * Table widget.
 *
 * Per changelog §3, the server does not currently serialise `column-widths`
 * or `row-heights`. They were previously documented as read-only response
 * fields but are intentionally omitted from this type.
 */
export interface Table extends BaseWidget {
  readonly "widget-type": "table";
  readonly "table-id": Uuid;
  readonly title?: string;
  readonly "grid-size": GridSize;
}

export interface CreateTableRequest extends WidgetMutableBase {
  readonly title?: string;
  readonly "grid-size": GridSize;
}

/**
 * Per changelog §4, including `grid-size` in a PATCH request is silently
 * ignored by the server. The SDK filters it out before sending and emits
 * a logger.warn.
 */
export interface UpdateTableRequest extends WidgetMutableBase {
  readonly title?: string;
}

/** A single cell in a {@link Table}. */
export interface TableCell {
  readonly "cell-id": Uuid;
  readonly index: readonly [number, number];
  readonly content: string;
}

// ---------- Video Input (canvas-scoped) ----------

export interface VideoInput extends BaseWidget {
  readonly "widget-type": "video-input";
  readonly source: string;
  readonly name: string;
  readonly resolution: string;
}

export interface CreateVideoInputRequest extends WidgetMutableBase {
  readonly source: string;
  readonly name?: string;
  readonly resolution?: string;
}

export type UpdateVideoInputRequest = Partial<CreateVideoInputRequest>;

// ---------- IP Video (read/update/delete only — see changelog §2) ----------

export interface IpVideo extends BaseWidget {
  readonly "widget-type": "ip-video";
  readonly "host-id": string;
  readonly parent_id: string;
  readonly source: string;
  readonly name?: string;
  readonly resolution?: string;
}

export interface UpdateIpVideoRequest extends WidgetMutableBase {
  readonly source?: string;
  readonly name?: string;
  readonly resolution?: string;
}

// ---------- RDP Connection (read/update/delete only — see changelog §2) ----------

/**
 * RDP connection widget.
 *
 * Live-server verification (2026-05-18): the C++ serialiser emits hyphenated keys
 * for `host-id`, `connection-name`, and `content-id`. Other widget fields
 * like `parent_id` remain underscored. This type follows the verified hybrid
 * convention.
 */
export interface RdpConnection extends BaseWidget {
  readonly "widget-type": "rdp-connection";
  readonly "host-id": string;
  readonly "connection-name": string;
  readonly "content-id": string;
  readonly parent_id: string;
  readonly title?: string;
}

export interface UpdateRdpConnectionRequest extends WidgetMutableBase {
  readonly "connection-name"?: string;
  readonly "host-site"?: string;
  readonly title?: string;
}

// ---------- Discriminated union ----------

/**
 * Discriminated union of every widget type. Narrow with `widget-type`:
 *
 * @example
 * ```ts
 * if (w["widget-type"] === "note") {
 *   console.log(w.text);
 * }
 * ```
 */
export type Widget =
  | Note
  | Image
  | Video
  | Pdf
  | Browser
  | Anchor
  | Connector
  | Table
  | VideoInput
  | IpVideo
  | RdpConnection;

/** Generic uploads-folder item (shape varies; treat as opaque JSON). */
export interface UploadsFolderItem {
  readonly id?: Uuid;
  readonly "upload-type"?: "Note" | "Image" | "Video" | "PDF";
  readonly title?: string;
  readonly location?: Location;
  readonly [key: string]: unknown;
}

export interface UploadsFolderUploadJson {
  readonly upload_type?: "Note" | "Image" | "Video" | "PDF";
  readonly title?: string;
  readonly location?: Location;
}
