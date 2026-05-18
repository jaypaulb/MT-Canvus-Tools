import type { Transport } from "../transport.js";
import { streamNdjson, type StreamOptions } from "../streaming.js";
import { logger } from "../logging.js";
import type { Uuid } from "../types/common.js";
import type {
  Anchor,
  Browser,
  CreateAnchorRequest,
  CreateBrowserRequest,
  CreateConnectorRequest,
  CreateNoteRequest,
  CreateTableRequest,
  CreateVideoInputRequest,
  Connector,
  Image,
  IpVideo,
  Note,
  Pdf,
  PdfMetadata,
  RdpConnection,
  Table,
  TableCell,
  UpdateAnchorRequest,
  UpdateBrowserRequest,
  UpdateConnectorRequest,
  UpdateImageRequest,
  UpdateIpVideoRequest,
  UpdateNoteRequest,
  UpdatePdfRequest,
  UpdateRdpConnectionRequest,
  UpdateTableRequest,
  UpdateVideoInputRequest,
  UpdateVideoRequest,
  UploadsFolderItem,
  UploadsFolderUploadJson,
  Video,
  VideoInput,
  Widget,
} from "../types/widget.js";

/** Optional metadata that accompanies a multipart asset upload. */
export interface UploadMetadata {
  readonly title?: string;
  readonly location?: { readonly x: number; readonly y: number };
  readonly size?: { readonly width: number; readonly height: number };
  readonly depth?: number;
  readonly scale?: number;
  readonly pinned?: boolean;
  readonly [key: string]: unknown;
}

/**
 * Cross-canvas clone arguments (changelog §1).
 *
 * `widgetType` values match the server's `widget_type` discriminator
 * (capitalised: `Note`, `Image`, `Video`, `Pdf`, `Browser`, `Anchor`,
 * `Table`). The clone helper internally maps these to the lowercase
 * plural URL segment.
 */
export interface CloneWidgetArgs {
  readonly destCanvasId: Uuid;
  readonly sourceCanvasId: Uuid;
  readonly sourceWidgetId: Uuid;
  readonly widgetType: "Note" | "Image" | "Video" | "Pdf" | "Browser" | "Anchor" | "Table";
  readonly location?: { readonly x: number; readonly y: number };
}

/**
 * The endpoint segment used in URLs for a given widget type.
 *
 * Note that the Canvus API uses plural lower-kebab-case for the URL
 * segments (`notes`, `pdfs`, `video-inputs`, `ip-videos`,
 * `rdp-connections`).
 */
const SEGMENT: Record<CloneWidgetArgs["widgetType"], string> = {
  Note: "notes",
  Image: "images",
  Video: "videos",
  Pdf: "pdfs",
  Browser: "browsers",
  Anchor: "anchors",
  Table: "tables",
};

/**
 * Construct a multipart body for asset uploads (image/video/pdf/background).
 */
function buildUploadForm(
  file: Blob | Buffer,
  filename: string,
  meta?: UploadMetadata,
): FormData {
  const form = new FormData();
  const blob = file instanceof Blob ? file : new Blob([file as unknown as ArrayBuffer]);
  form.append("data", blob, filename);
  if (meta && Object.keys(meta).length > 0) {
    form.append("json", JSON.stringify(meta));
  }
  return form;
}

/**
 * Aggregator for every widget-related endpoint on a canvas.
 *
 * Exposes namespaced sub-resources (e.g. `widgets.notes.list(canvasId)`)
 * to keep IDE autocompletion meaningful at 64 endpoints.
 */
export class WidgetsResource {
  constructor(private readonly transport: Transport) {}

  /** `GET /api/v1/canvases/{cid}/widgets` — list every widget on the canvas. */
  async list(canvasId: Uuid): Promise<readonly Widget[]> {
    return this.transport.request<readonly Widget[]>("GET", `canvases/${canvasId}/widgets`);
  }

  /** Subscribe to every widget on a canvas. */
  subscribe(canvasId: Uuid, opts?: StreamOptions): AsyncGenerator<Widget, void, void> {
    return streamNdjson<Widget>(this.transport, `canvases/${canvasId}/widgets`, opts);
  }

  /** `GET /api/v1/canvases/{cid}/widgets/{wid}` — fetch any widget by ID. */
  async get(canvasId: Uuid, widgetId: Uuid): Promise<Widget> {
    return this.transport.request<Widget>("GET", `canvases/${canvasId}/widgets/${widgetId}`);
  }

  /** Subscribe to a single (typed) widget by ID. */
  subscribeOne(
    canvasId: Uuid,
    widgetId: Uuid,
    opts?: StreamOptions,
  ): AsyncGenerator<Widget, void, void> {
    return streamNdjson<Widget>(
      this.transport,
      `canvases/${canvasId}/widgets/${widgetId}`,
      opts,
    );
  }

  /**
   * Clone a widget into another canvas using the standard create endpoint
   * with `source_canvas_id` / `source_widget_id` in the body
   * (per changelog §1 — these two clone-source fields are underscored,
   * matching the rest of the verified wire shape).
   *
   * This is the canonical way to copy widgets across canvases; the
   * historical `/widgets/clone` endpoint returns 501 Not Implemented and
   * is intentionally absent from the SDK.
   */
  async clone<T = Widget>(args: CloneWidgetArgs): Promise<T> {
    const segment = SEGMENT[args.widgetType];
    // Note: the changelog spec uses underscored keys for clone source
    // fields (`source_canvas_id`, `source_widget_id`) — verified against
    // doc-updates Section 1, lines 9-55. Most other API fields use hyphens;
    // these two are the deliberate exception.
    const body: Record<string, unknown> = {
      source_canvas_id: args.sourceCanvasId,
      source_widget_id: args.sourceWidgetId,
    };
    if (args.location) body.location = args.location;
    return this.transport.request<T>("POST", `canvases/${args.destCanvasId}/${segment}`, body);
  }

  // -----------------------------------------------------------------------
  // Sub-resources
  // -----------------------------------------------------------------------

  /** Note CRUD. */
  readonly notes = {
    list: (canvasId: Uuid): Promise<readonly Note[]> =>
      this.transport.request("GET", `canvases/${canvasId}/notes`),
    subscribe: (canvasId: Uuid, opts?: StreamOptions): AsyncGenerator<Note, void, void> =>
      streamNdjson<Note>(this.transport, `canvases/${canvasId}/notes`, opts),
    get: (canvasId: Uuid, noteId: Uuid): Promise<Note> =>
      this.transport.request("GET", `canvases/${canvasId}/notes/${noteId}`),
    subscribeOne: (
      canvasId: Uuid,
      noteId: Uuid,
      opts?: StreamOptions,
    ): AsyncGenerator<Note, void, void> =>
      streamNdjson<Note>(this.transport, `canvases/${canvasId}/notes/${noteId}`, opts),
    create: (canvasId: Uuid, body: CreateNoteRequest): Promise<Note> =>
      this.transport.request("POST", `canvases/${canvasId}/notes`, body),
    update: (canvasId: Uuid, noteId: Uuid, body: UpdateNoteRequest): Promise<Note> =>
      this.transport.request("PATCH", `canvases/${canvasId}/notes/${noteId}`, body),
    delete: async (canvasId: Uuid, noteId: Uuid): Promise<void> => {
      await this.transport.request<void>("DELETE", `canvases/${canvasId}/notes/${noteId}`);
    },
  };

  /** Image CRUD + multipart upload + download. */
  readonly images = {
    list: (canvasId: Uuid): Promise<readonly Image[]> =>
      this.transport.request("GET", `canvases/${canvasId}/images`),
    subscribe: (canvasId: Uuid, opts?: StreamOptions): AsyncGenerator<Image, void, void> =>
      streamNdjson<Image>(this.transport, `canvases/${canvasId}/images`, opts),
    get: (canvasId: Uuid, imageId: Uuid): Promise<Image> =>
      this.transport.request("GET", `canvases/${canvasId}/images/${imageId}`),
    subscribeOne: (
      canvasId: Uuid,
      imageId: Uuid,
      opts?: StreamOptions,
    ): AsyncGenerator<Image, void, void> =>
      streamNdjson<Image>(this.transport, `canvases/${canvasId}/images/${imageId}`, opts),
    upload: (
      canvasId: Uuid,
      file: Blob | Buffer,
      filename: string,
      meta?: UploadMetadata,
    ): Promise<Image> =>
      this.transport.request(
        "POST",
        `canvases/${canvasId}/images`,
        buildUploadForm(file, filename, meta),
      ),
    update: (canvasId: Uuid, imageId: Uuid, body: UpdateImageRequest): Promise<Image> =>
      this.transport.request("PATCH", `canvases/${canvasId}/images/${imageId}`, body),
    download: async (canvasId: Uuid, imageId: Uuid): Promise<Blob> => {
      const res = await this.transport.rawRequest(
        "GET",
        `canvases/${canvasId}/images/${imageId}/download`,
      );
      return res.blob();
    },
    delete: async (canvasId: Uuid, imageId: Uuid): Promise<void> => {
      await this.transport.request<void>("DELETE", `canvases/${canvasId}/images/${imageId}`);
    },
  };

  /** Video CRUD + multipart upload + download. */
  readonly videos = {
    list: (canvasId: Uuid): Promise<readonly Video[]> =>
      this.transport.request("GET", `canvases/${canvasId}/videos`),
    subscribe: (canvasId: Uuid, opts?: StreamOptions): AsyncGenerator<Video, void, void> =>
      streamNdjson<Video>(this.transport, `canvases/${canvasId}/videos`, opts),
    get: (canvasId: Uuid, videoId: Uuid): Promise<Video> =>
      this.transport.request("GET", `canvases/${canvasId}/videos/${videoId}`),
    subscribeOne: (
      canvasId: Uuid,
      videoId: Uuid,
      opts?: StreamOptions,
    ): AsyncGenerator<Video, void, void> =>
      streamNdjson<Video>(this.transport, `canvases/${canvasId}/videos/${videoId}`, opts),
    upload: (
      canvasId: Uuid,
      file: Blob | Buffer,
      filename: string,
      meta?: UploadMetadata,
    ): Promise<Video> =>
      this.transport.request(
        "POST",
        `canvases/${canvasId}/videos`,
        buildUploadForm(file, filename, meta),
      ),
    update: (canvasId: Uuid, videoId: Uuid, body: UpdateVideoRequest): Promise<Video> =>
      this.transport.request("PATCH", `canvases/${canvasId}/videos/${videoId}`, body),
    download: async (canvasId: Uuid, videoId: Uuid): Promise<Blob> => {
      const res = await this.transport.rawRequest(
        "GET",
        `canvases/${canvasId}/videos/${videoId}/download`,
      );
      return res.blob();
    },
    delete: async (canvasId: Uuid, videoId: Uuid): Promise<void> => {
      await this.transport.request<void>("DELETE", `canvases/${canvasId}/videos/${videoId}`);
    },
  };

  /** PDF CRUD + multipart upload + download. */
  readonly pdfs = {
    list: (canvasId: Uuid): Promise<readonly Pdf[]> =>
      this.transport.request("GET", `canvases/${canvasId}/pdfs`),
    subscribe: (canvasId: Uuid, opts?: StreamOptions): AsyncGenerator<Pdf, void, void> =>
      streamNdjson<Pdf>(this.transport, `canvases/${canvasId}/pdfs`, opts),
    get: (canvasId: Uuid, pdfId: Uuid): Promise<Pdf> =>
      this.transport.request("GET", `canvases/${canvasId}/pdfs/${pdfId}`),
    subscribeOne: (
      canvasId: Uuid,
      pdfId: Uuid,
      opts?: StreamOptions,
    ): AsyncGenerator<Pdf, void, void> =>
      streamNdjson<Pdf>(this.transport, `canvases/${canvasId}/pdfs/${pdfId}`, opts),
    upload: (
      canvasId: Uuid,
      file: Blob | Buffer,
      filename: string,
      meta?: UploadMetadata,
    ): Promise<Pdf> =>
      this.transport.request(
        "POST",
        `canvases/${canvasId}/pdfs`,
        buildUploadForm(file, filename, meta),
      ),
    update: (
      canvasId: Uuid,
      pdfId: Uuid,
      body: UpdatePdfRequest | PdfMetadata,
    ): Promise<Pdf> =>
      this.transport.request("PATCH", `canvases/${canvasId}/pdfs/${pdfId}`, body),
    download: async (canvasId: Uuid, pdfId: Uuid): Promise<Blob> => {
      const res = await this.transport.rawRequest(
        "GET",
        `canvases/${canvasId}/pdfs/${pdfId}/download`,
      );
      return res.blob();
    },
    delete: async (canvasId: Uuid, pdfId: Uuid): Promise<void> => {
      await this.transport.request<void>("DELETE", `canvases/${canvasId}/pdfs/${pdfId}`);
    },
  };

  /** Browser CRUD. */
  readonly browsers = {
    list: (canvasId: Uuid): Promise<readonly Browser[]> =>
      this.transport.request("GET", `canvases/${canvasId}/browsers`),
    subscribe: (canvasId: Uuid, opts?: StreamOptions): AsyncGenerator<Browser, void, void> =>
      streamNdjson<Browser>(this.transport, `canvases/${canvasId}/browsers`, opts),
    get: (canvasId: Uuid, browserId: Uuid): Promise<Browser> =>
      this.transport.request("GET", `canvases/${canvasId}/browsers/${browserId}`),
    subscribeOne: (
      canvasId: Uuid,
      browserId: Uuid,
      opts?: StreamOptions,
    ): AsyncGenerator<Browser, void, void> =>
      streamNdjson<Browser>(this.transport, `canvases/${canvasId}/browsers/${browserId}`, opts),
    create: (canvasId: Uuid, body: CreateBrowserRequest): Promise<Browser> =>
      this.transport.request("POST", `canvases/${canvasId}/browsers`, body),
    update: (canvasId: Uuid, browserId: Uuid, body: UpdateBrowserRequest): Promise<Browser> =>
      this.transport.request("PATCH", `canvases/${canvasId}/browsers/${browserId}`, body),
    delete: async (canvasId: Uuid, browserId: Uuid): Promise<void> => {
      await this.transport.request<void>("DELETE", `canvases/${canvasId}/browsers/${browserId}`);
    },
  };

  /** Anchor CRUD. */
  readonly anchors = {
    list: (canvasId: Uuid): Promise<readonly Anchor[]> =>
      this.transport.request("GET", `canvases/${canvasId}/anchors`),
    subscribe: (canvasId: Uuid, opts?: StreamOptions): AsyncGenerator<Anchor, void, void> =>
      streamNdjson<Anchor>(this.transport, `canvases/${canvasId}/anchors`, opts),
    get: (canvasId: Uuid, anchorId: Uuid): Promise<Anchor> =>
      this.transport.request("GET", `canvases/${canvasId}/anchors/${anchorId}`),
    subscribeOne: (
      canvasId: Uuid,
      anchorId: Uuid,
      opts?: StreamOptions,
    ): AsyncGenerator<Anchor, void, void> =>
      streamNdjson<Anchor>(this.transport, `canvases/${canvasId}/anchors/${anchorId}`, opts),
    create: (canvasId: Uuid, body: CreateAnchorRequest): Promise<Anchor> =>
      this.transport.request("POST", `canvases/${canvasId}/anchors`, body),
    update: (canvasId: Uuid, anchorId: Uuid, body: UpdateAnchorRequest): Promise<Anchor> =>
      this.transport.request("PATCH", `canvases/${canvasId}/anchors/${anchorId}`, body),
    delete: async (canvasId: Uuid, anchorId: Uuid): Promise<void> => {
      await this.transport.request<void>("DELETE", `canvases/${canvasId}/anchors/${anchorId}`);
    },
  };

  /** Connector CRUD. */
  readonly connectors = {
    list: (canvasId: Uuid): Promise<readonly Connector[]> =>
      this.transport.request("GET", `canvases/${canvasId}/connectors`),
    subscribe: (canvasId: Uuid, opts?: StreamOptions): AsyncGenerator<Connector, void, void> =>
      streamNdjson<Connector>(this.transport, `canvases/${canvasId}/connectors`, opts),
    get: (canvasId: Uuid, connectorId: Uuid): Promise<Connector> =>
      this.transport.request("GET", `canvases/${canvasId}/connectors/${connectorId}`),
    subscribeOne: (
      canvasId: Uuid,
      connectorId: Uuid,
      opts?: StreamOptions,
    ): AsyncGenerator<Connector, void, void> =>
      streamNdjson<Connector>(
        this.transport,
        `canvases/${canvasId}/connectors/${connectorId}`,
        opts,
      ),
    create: (canvasId: Uuid, body: CreateConnectorRequest): Promise<Connector> =>
      this.transport.request("POST", `canvases/${canvasId}/connectors`, body),
    update: (
      canvasId: Uuid,
      connectorId: Uuid,
      body: UpdateConnectorRequest,
    ): Promise<Connector> =>
      this.transport.request("PATCH", `canvases/${canvasId}/connectors/${connectorId}`, body),
    delete: async (canvasId: Uuid, connectorId: Uuid): Promise<void> => {
      await this.transport.request<void>(
        "DELETE",
        `canvases/${canvasId}/connectors/${connectorId}`,
      );
    },
  };

  /** Table CRUD + cell read. */
  readonly tables = {
    list: (canvasId: Uuid): Promise<readonly Table[]> =>
      this.transport.request("GET", `canvases/${canvasId}/tables`),
    subscribe: (canvasId: Uuid, opts?: StreamOptions): AsyncGenerator<Table, void, void> =>
      streamNdjson<Table>(this.transport, `canvases/${canvasId}/tables`, opts),
    get: (canvasId: Uuid, tableId: Uuid): Promise<Table> =>
      this.transport.request("GET", `canvases/${canvasId}/tables/${tableId}`),
    subscribeOne: (
      canvasId: Uuid,
      tableId: Uuid,
      opts?: StreamOptions,
    ): AsyncGenerator<Table, void, void> =>
      streamNdjson<Table>(this.transport, `canvases/${canvasId}/tables/${tableId}`, opts),
    create: (canvasId: Uuid, body: CreateTableRequest): Promise<Table> =>
      this.transport.request("POST", `canvases/${canvasId}/tables`, body),
    /**
     * Update a table.
     *
     * Per changelog §4, the server silently drops `grid_size` from PATCH
     * bodies. The SDK filters it out and emits a `logger.warn` so callers
     * notice the no-op.
     */
    update: (canvasId: Uuid, tableId: Uuid, body: UpdateTableRequest): Promise<Table> => {
      const filtered = body as Record<string, unknown>;
      if ("grid_size" in filtered) {
        logger.warn(
          { canvasId, tableId },
          "ignoring grid_size in table PATCH (server silently drops it)",
        );
        const { ["grid_size"]: _omit, ...rest } = filtered;
        void _omit;
        return this.transport.request("PATCH", `canvases/${canvasId}/tables/${tableId}`, rest);
      }
      return this.transport.request("PATCH", `canvases/${canvasId}/tables/${tableId}`, body);
    },
    cells: (canvasId: Uuid, tableId: Uuid): Promise<readonly TableCell[]> =>
      this.transport.request("GET", `canvases/${canvasId}/tables/${tableId}/cells`),
    subscribeCells: (
      canvasId: Uuid,
      tableId: Uuid,
      opts?: StreamOptions,
    ): AsyncGenerator<TableCell, void, void> =>
      streamNdjson<TableCell>(
        this.transport,
        `canvases/${canvasId}/tables/${tableId}/cells`,
        opts,
      ),
    delete: async (canvasId: Uuid, tableId: Uuid): Promise<void> => {
      await this.transport.request<void>("DELETE", `canvases/${canvasId}/tables/${tableId}`);
    },
  };

  /** Canvas-scoped video-input widget CRUD. */
  readonly videoInputs = {
    list: (canvasId: Uuid): Promise<readonly VideoInput[]> =>
      this.transport.request("GET", `canvases/${canvasId}/video-inputs`),
    subscribe: (canvasId: Uuid, opts?: StreamOptions): AsyncGenerator<VideoInput, void, void> =>
      streamNdjson<VideoInput>(this.transport, `canvases/${canvasId}/video-inputs`, opts),
    get: (canvasId: Uuid, widgetId: Uuid): Promise<VideoInput> =>
      this.transport.request("GET", `canvases/${canvasId}/video-inputs/${widgetId}`),
    subscribeOne: (
      canvasId: Uuid,
      widgetId: Uuid,
      opts?: StreamOptions,
    ): AsyncGenerator<VideoInput, void, void> =>
      streamNdjson<VideoInput>(
        this.transport,
        `canvases/${canvasId}/video-inputs/${widgetId}`,
        opts,
      ),
    create: (canvasId: Uuid, body: CreateVideoInputRequest): Promise<VideoInput> =>
      this.transport.request("POST", `canvases/${canvasId}/video-inputs`, body),
    update: (
      canvasId: Uuid,
      widgetId: Uuid,
      body: UpdateVideoInputRequest,
    ): Promise<VideoInput> =>
      this.transport.request("PATCH", `canvases/${canvasId}/video-inputs/${widgetId}`, body),
    delete: async (canvasId: Uuid, widgetId: Uuid): Promise<void> => {
      await this.transport.request<void>(
        "DELETE",
        `canvases/${canvasId}/video-inputs/${widgetId}`,
      );
    },
  };

  /**
   * IP video widgets — read/update/delete only.
   *
   * Per changelog §2, `POST /api/v1/canvases/{cid}/ip-videos` is rejected
   * by the server with "WidgetType IpVideo is not supported". The SDK
   * intentionally omits a create method.
   */
  readonly ipVideos = {
    list: (canvasId: Uuid): Promise<readonly IpVideo[]> =>
      this.transport.request("GET", `canvases/${canvasId}/ip-videos`),
    subscribe: (canvasId: Uuid, opts?: StreamOptions): AsyncGenerator<IpVideo, void, void> =>
      streamNdjson<IpVideo>(this.transport, `canvases/${canvasId}/ip-videos`, opts),
    get: (canvasId: Uuid, widgetId: Uuid): Promise<IpVideo> =>
      this.transport.request("GET", `canvases/${canvasId}/ip-videos/${widgetId}`),
    subscribeOne: (
      canvasId: Uuid,
      widgetId: Uuid,
      opts?: StreamOptions,
    ): AsyncGenerator<IpVideo, void, void> =>
      streamNdjson<IpVideo>(
        this.transport,
        `canvases/${canvasId}/ip-videos/${widgetId}`,
        opts,
      ),
    update: (canvasId: Uuid, widgetId: Uuid, body: UpdateIpVideoRequest): Promise<IpVideo> =>
      this.transport.request("PATCH", `canvases/${canvasId}/ip-videos/${widgetId}`, body),
    delete: async (canvasId: Uuid, widgetId: Uuid): Promise<void> => {
      await this.transport.request<void>(
        "DELETE",
        `canvases/${canvasId}/ip-videos/${widgetId}`,
      );
    },
  };

  /**
   * RDP-connection widgets — read/update/delete only.
   *
   * Per changelog §2, `POST /api/v1/canvases/{cid}/rdp-connections` is
   * rejected by the server with "WidgetType is not supported". The SDK
   * intentionally omits a create method.
   *
   * Per live-server verification (2026-05-18): the C++ serialiser emits
   * HYPHENATED keys for `host-id`, `connection-name`, and `content-id`
   * (other widget fields like `parent_id`, `widget_type` remain
   * underscored). The `RdpConnection` type reflects this hybrid shape.
   */
  readonly rdpConnections = {
    list: (canvasId: Uuid): Promise<readonly RdpConnection[]> =>
      this.transport.request("GET", `canvases/${canvasId}/rdp-connections`),
    subscribe: (
      canvasId: Uuid,
      opts?: StreamOptions,
    ): AsyncGenerator<RdpConnection, void, void> =>
      streamNdjson<RdpConnection>(
        this.transport,
        `canvases/${canvasId}/rdp-connections`,
        opts,
      ),
    get: (canvasId: Uuid, widgetId: Uuid): Promise<RdpConnection> =>
      this.transport.request("GET", `canvases/${canvasId}/rdp-connections/${widgetId}`),
    subscribeOne: (
      canvasId: Uuid,
      widgetId: Uuid,
      opts?: StreamOptions,
    ): AsyncGenerator<RdpConnection, void, void> =>
      streamNdjson<RdpConnection>(
        this.transport,
        `canvases/${canvasId}/rdp-connections/${widgetId}`,
        opts,
      ),
    update: (
      canvasId: Uuid,
      widgetId: Uuid,
      body: UpdateRdpConnectionRequest,
    ): Promise<RdpConnection> =>
      this.transport.request(
        "PATCH",
        `canvases/${canvasId}/rdp-connections/${widgetId}`,
        body,
      ),
    delete: async (canvasId: Uuid, widgetId: Uuid): Promise<void> => {
      await this.transport.request<void>(
        "DELETE",
        `canvases/${canvasId}/rdp-connections/${widgetId}`,
      );
    },
  };

  /** Uploads-folder staging area. */
  readonly uploadsFolder = {
    list: (canvasId: Uuid): Promise<readonly UploadsFolderItem[]> =>
      this.transport.request("GET", `canvases/${canvasId}/uploads-folder`),
    subscribe: (
      canvasId: Uuid,
      opts?: StreamOptions,
    ): AsyncGenerator<UploadsFolderItem, void, void> =>
      streamNdjson<UploadsFolderItem>(
        this.transport,
        `canvases/${canvasId}/uploads-folder`,
        opts,
      ),
    /**
     * Upload to the staging folder.
     *
     * Pass `file: undefined` for `upload_type: "Note"` (no binary part).
     */
    upload: (
      canvasId: Uuid,
      file: Blob | Buffer | undefined,
      filename: string | undefined,
      meta: UploadsFolderUploadJson,
    ): Promise<UploadsFolderItem> => {
      const form = new FormData();
      if (file !== undefined) {
        const blob = file instanceof Blob ? file : new Blob([file as unknown as ArrayBuffer]);
        form.append("data", blob, filename ?? "upload");
      }
      form.append("json", JSON.stringify(meta));
      return this.transport.request(
        "POST",
        `canvases/${canvasId}/uploads-folder`,
        form,
      );
    },
  };
}
