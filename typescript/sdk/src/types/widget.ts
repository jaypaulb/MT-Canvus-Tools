import type { AssetHash, HexColor, Location, Size, Uuid } from "./common.js";

/**
 * Discriminator for every widget type known to the Canvus API.
 *
 * Values are the literal strings the server emits in `widget_type`.
 * Verified against dev-mtcs.multitaction.com (v1.2):
 * `Note`, `Image`, `Video`, `Pdf`, `Browser`, `Anchor`, `Connector`,
 * `Table`, `VideoInput`, `IpVideo`, `RdpConnection`.
 */
export type WidgetKind =
  | "Note"
  | "Image"
  | "Video"
  | "Pdf"
  | "Browser"
  | "Anchor"
  | "Connector"
  | "Table"
  | "VideoInput"
  | "IpVideo"
  | "RdpConnection";

/**
 * Fields common to every widget.
 *
 * Live-server verification (2026-05-18) established that the server emits
 * UNDERSCORED keys for the canonical widget fields: `id`, `widget_type`,
 * `parent_id`, `pinned`, etc. The previous hyphenated forms
 * (`widget-id`, `widget-type`, `is-pinned`) do not exist on the wire.
 */
export interface BaseWidget {
  readonly id: Uuid;
  readonly widget_type: WidgetKind;
  readonly parent_id: Uuid;
  readonly location: Location;
  readonly depth: number;
  readonly size: Size;
  readonly scale: number;
  readonly pinned: boolean;
  readonly state: string;
}

/** Fields that any widget can accept on create/update. */
export interface WidgetMutableBase {
  readonly location?: Location;
  readonly depth?: number;
  readonly size?: Size;
  readonly scale?: number;
  readonly pinned?: boolean;
  readonly parent_id?: Uuid;
}

/** Source-widget reference used by the cross-canvas clone endpoint.
 * The clone source fields use underscores (`source_canvas_id`) per the
 * Phase 4 changelog §1 source note. */
export interface CloneSource {
  readonly source_canvas_id: Uuid;
  readonly source_widget_id: Uuid;
  readonly location?: Location;
}

// ---------- Note ----------

export interface Note extends BaseWidget {
  readonly widget_type: "Note";
  readonly text: string;
  readonly title?: string;
  readonly background_color: HexColor;
  readonly text_color: HexColor;
  readonly auto_text_color: boolean;
}

export interface CreateNoteRequest extends WidgetMutableBase {
  readonly text?: string;
  readonly title?: string;
  readonly background_color?: HexColor;
  readonly text_color?: HexColor;
  readonly auto_text_color?: boolean;
}

export type UpdateNoteRequest = CreateNoteRequest;

// ---------- Image ----------

export interface Image extends BaseWidget {
  readonly widget_type: "Image";
  readonly hash: AssetHash;
  readonly original_filename: string;
  readonly title?: string;
  readonly mime_type?: string;
  readonly file_size?: number;
}

export interface ImageMetadata extends WidgetMutableBase {
  readonly title?: string;
}

export type UpdateImageRequest = ImageMetadata;

// ---------- Video ----------

export type VideoPlaybackState = "playing" | "paused" | "stopped" | "PLAYING" | "PAUSED" | "STOPPED";

export interface Video extends BaseWidget {
  readonly widget_type: "Video";
  readonly hash: AssetHash;
  readonly original_filename: string;
  readonly title?: string;
  readonly mime_type?: string;
  readonly file_size?: number;
  readonly playback_position?: number;
  readonly playback_state?: VideoPlaybackState;
  readonly muted?: boolean;
  readonly duration?: string;
}

export interface VideoMetadata extends WidgetMutableBase {
  readonly title?: string;
  readonly playback_position?: number;
  readonly playback_state?: VideoPlaybackState;
  readonly muted?: boolean;
}

export type UpdateVideoRequest = VideoMetadata;

// ---------- PDF ----------

export interface Pdf extends BaseWidget {
  readonly widget_type: "Pdf";
  readonly hash: AssetHash;
  readonly original_filename: string;
  readonly title?: string;
  readonly mime_type?: string;
  readonly file_size?: number;
  readonly index: number;
  readonly page_count?: number;
}

export interface PdfMetadata extends WidgetMutableBase {
  readonly title?: string;
  readonly index?: number;
}

export type UpdatePdfRequest = PdfMetadata;

// ---------- Browser ----------

export interface Browser extends BaseWidget {
  readonly widget_type: "Browser";
  readonly url: string;
  readonly title?: string;
  readonly transparent_mode: boolean;
  readonly main_frame_scroll_offset?: Location;
}

export interface CreateBrowserRequest extends WidgetMutableBase {
  readonly url: string;
  readonly title?: string;
  readonly transparent_mode?: boolean;
}

export interface UpdateBrowserRequest extends WidgetMutableBase {
  readonly url?: string;
  readonly title?: string;
  readonly transparent_mode?: boolean;
  readonly main_frame_scroll_offset?: Location;
}

// ---------- Anchor ----------

export interface Anchor extends BaseWidget {
  readonly widget_type: "Anchor";
  readonly anchor_index: number;
  readonly anchor_name: string;
}

export interface CreateAnchorRequest extends WidgetMutableBase {
  readonly anchor_name?: string;
  readonly anchor_index?: number;
}

export type UpdateAnchorRequest = CreateAnchorRequest;

// ---------- Connector ----------

export type ConnectorType = "line" | "curve" | "arrow";
export type ConnectorTip = "none" | "arrow" | "circle";

/** One endpoint of a connector. */
export interface ConnectorEndpoint {
  readonly id: Uuid;
  readonly rel_location?: Location;
  readonly auto_location?: boolean;
  readonly tip?: ConnectorTip;
}

/**
 * Connector widget.
 *
 * Connectors don't have a `parent_id` in the same sense as other widgets;
 * they reference two endpoints (`src` / `dst`). The wire shape uses
 * underscored field names like other widgets.
 */
export interface Connector {
  readonly id: Uuid;
  readonly widget_type: "Connector";
  readonly state?: string;
  readonly depth: number;
  readonly src?: ConnectorEndpoint;
  readonly dst?: ConnectorEndpoint;
  readonly line_color: HexColor;
  readonly line_width: number;
  readonly type: ConnectorType;
}

export interface CreateConnectorRequest {
  readonly src: ConnectorEndpoint | Uuid;
  readonly dst: ConnectorEndpoint | Uuid;
  readonly type?: ConnectorType;
  readonly line_color?: HexColor;
  readonly line_width?: number;
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
 * Per changelog §3, the server does not currently serialise `column_widths`
 * or `row_heights`. They were previously documented as read-only response
 * fields but are intentionally omitted from this type.
 */
export interface Table extends BaseWidget {
  readonly widget_type: "Table";
  readonly title?: string;
  readonly grid_size?: GridSize;
}

export interface CreateTableRequest extends WidgetMutableBase {
  readonly title?: string;
  readonly grid_size: GridSize;
}

/**
 * Per changelog §4, including `grid_size` in a PATCH request is silently
 * ignored by the server. The SDK filters it out before sending and emits
 * a logger.warn.
 */
export interface UpdateTableRequest extends WidgetMutableBase {
  readonly title?: string;
}

/** A single cell in a {@link Table}. */
export interface TableCell {
  readonly column: number;
  readonly row: number;
  readonly text: string;
  readonly background_color?: HexColor;
  readonly text_color?: HexColor;
}

// ---------- Video Input (canvas-scoped) ----------

/**
 * Canvas-scoped video-input widget.
 *
 * Live-server verification: `host-id` is HYPHENATED on the wire (it's
 * an external client/host reference); other widget fields use underscores.
 */
export interface VideoInput extends BaseWidget {
  readonly widget_type: "VideoInput";
  readonly "host-id": Uuid;
  readonly source: string;
  readonly name?: string;
  readonly resolution?: string;
}

export interface CreateVideoInputRequest extends WidgetMutableBase {
  readonly source: string;
  readonly "host-id"?: Uuid;
  readonly name?: string;
  readonly resolution?: string;
}

export type UpdateVideoInputRequest = Partial<CreateVideoInputRequest>;

// ---------- IP Video (read/update/delete only — see changelog §2) ----------

/**
 * IP Video stream widget.
 *
 * Live-server verification (2026-05-18): the wire emits `host-id`
 * (HYPHENATED) for the host identifier; other widget fields use
 * underscores (`parent_id`, `widget_type`, etc.). Per changelog §2, this
 * widget type cannot be created via the API — only read/update/delete.
 */
export interface IpVideo extends BaseWidget {
  readonly widget_type: "IpVideo";
  readonly "host-id": Uuid;
  readonly source: string;
  readonly name?: string;
  readonly title?: string;
  readonly resolution?: string;
}

export interface UpdateIpVideoRequest extends WidgetMutableBase {
  readonly source?: string;
  readonly name?: string;
  readonly title?: string;
  readonly resolution?: string;
}

// ---------- RDP Connection (read/update/delete only — see changelog §2) ----------

/**
 * RDP connection widget.
 *
 * Live-server verification (2026-05-18): the C++ serialiser emits HYPHENATED
 * keys for `host-id`, `connection-name`, and `content-id`. Other widget
 * fields (`parent_id`, `widget_type`, etc.) remain UNDERSCORED. This type
 * follows the verified hybrid convention.
 */
export interface RdpConnection extends BaseWidget {
  readonly widget_type: "RdpConnection";
  readonly "host-id": Uuid;
  readonly "connection-name": string;
  readonly "content-id": string;
  readonly title?: string;
}

export interface UpdateRdpConnectionRequest extends WidgetMutableBase {
  readonly "connection-name"?: string;
  readonly "host-site"?: string;
  readonly title?: string;
}

// ---------- Discriminated union ----------

/**
 * Discriminated union of every widget type. Narrow with `widget_type`:
 *
 * @example
 * ```ts
 * if (w.widget_type === "Note") {
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
  readonly upload_type?: "Note" | "Image" | "Video" | "PDF";
  readonly title?: string;
  readonly location?: Location;
  readonly [key: string]: unknown;
}

export interface UploadsFolderUploadJson {
  readonly upload_type?: "Note" | "Image" | "Video" | "PDF";
  readonly title?: string;
  readonly location?: Location;
}
