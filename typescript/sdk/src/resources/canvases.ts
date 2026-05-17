import type { Transport } from "../transport.js";
import { streamNdjson, type StreamOptions } from "../streaming.js";
import type {
  Canvas,
  CanvasBackground,
  CanvasColorPresets,
  CanvasPermissions,
  CopyCanvasRequest,
  CreateCanvasRequest,
  MoveCanvasRequest,
  SetCanvasPermissionsRequest,
  UpdateCanvasBackgroundRequest,
  UpdateCanvasColorPresetsRequest,
  UpdateCanvasRequest,
} from "../types/canvas.js";
import type { Uuid } from "../types/common.js";

/**
 * Canvas-document management.
 *
 * Implements the 17 endpoints under `/api/v1/canvases/*` defined in
 * `docs/api-reference/endpoints/canvases.md`.
 */
export class CanvasesResource {
  constructor(private readonly transport: Transport) {}

  /** `GET /api/v1/canvases` — list every canvas visible to the caller. */
  async list(): Promise<readonly Canvas[]> {
    return this.transport.request<readonly Canvas[]>("GET", "canvases");
  }

  /** Subscribe to the canvas list as an NDJSON stream. */
  subscribe(opts?: StreamOptions): AsyncGenerator<Canvas, void, void> {
    return streamNdjson<Canvas>(this.transport, "canvases", opts);
  }

  /** `GET /api/v1/canvases/{id}`. */
  async get(canvasId: Uuid): Promise<Canvas> {
    return this.transport.request<Canvas>("GET", `canvases/${canvasId}`);
  }

  /** Subscribe to a single canvas as an NDJSON stream. */
  subscribeOne(canvasId: Uuid, opts?: StreamOptions): AsyncGenerator<Canvas, void, void> {
    return streamNdjson<Canvas>(this.transport, `canvases/${canvasId}`, opts);
  }

  /** `POST /api/v1/canvases`. */
  async create(body: CreateCanvasRequest): Promise<Canvas> {
    return this.transport.request<Canvas>("POST", "canvases", body);
  }

  /** `PATCH /api/v1/canvases/{id}`. */
  async update(canvasId: Uuid, body: UpdateCanvasRequest): Promise<Canvas> {
    return this.transport.request<Canvas>("PATCH", `canvases/${canvasId}`, body);
  }

  /** `DELETE /api/v1/canvases/{id}`. */
  async delete(canvasId: Uuid): Promise<void> {
    await this.transport.request<void>("DELETE", `canvases/${canvasId}`);
  }

  /** `POST /api/v1/canvases/{id}/move`. */
  async move(canvasId: Uuid, body: MoveCanvasRequest): Promise<Canvas> {
    return this.transport.request<Canvas>("POST", `canvases/${canvasId}/move`, body);
  }

  /** `POST /api/v1/canvases/{id}/copy`. */
  async copy(canvasId: Uuid, body: CopyCanvasRequest): Promise<Canvas> {
    return this.transport.request<Canvas>("POST", `canvases/${canvasId}/copy`, body);
  }

  /** `POST /api/v1/canvases/{id}/save` — capture a demo snapshot. */
  async save(canvasId: Uuid): Promise<Canvas> {
    return this.transport.request<Canvas>("POST", `canvases/${canvasId}/save`, {});
  }

  /** `POST /api/v1/canvases/{id}/restore` — restore the demo snapshot. */
  async restore(canvasId: Uuid): Promise<Canvas> {
    return this.transport.request<Canvas>("POST", `canvases/${canvasId}/restore`, {});
  }

  /** `GET /api/v1/canvases/{id}/background`. */
  async getBackground(canvasId: Uuid): Promise<CanvasBackground> {
    return this.transport.request<CanvasBackground>("GET", `canvases/${canvasId}/background`);
  }

  /** `PATCH /api/v1/canvases/{id}/background` — settings-only patch. */
  async updateBackground(
    canvasId: Uuid,
    body: UpdateCanvasBackgroundRequest,
  ): Promise<CanvasBackground> {
    return this.transport.request<CanvasBackground>(
      "PATCH",
      `canvases/${canvasId}/background`,
      body,
    );
  }

  /**
   * `POST /api/v1/canvases/{id}/background` — upload a new background image.
   *
   * Accepts a `Blob` (preferred) or a Node `Buffer`. Filename is optional but
   * recommended so the server preserves the original extension.
   */
  async uploadBackground(
    canvasId: Uuid,
    file: Blob | Buffer,
    filename = "background",
  ): Promise<CanvasBackground> {
    const form = new FormData();
    const blob = file instanceof Blob ? file : new Blob([file as unknown as ArrayBuffer]);
    form.append("data", blob, filename);
    return this.transport.request<CanvasBackground>(
      "POST",
      `canvases/${canvasId}/background`,
      form,
    );
  }

  /** `GET /api/v1/canvases/{id}/color-presets`. */
  async getColorPresets(canvasId: Uuid): Promise<CanvasColorPresets> {
    return this.transport.request<CanvasColorPresets>(
      "GET",
      `canvases/${canvasId}/color-presets`,
    );
  }

  /** `PATCH /api/v1/canvases/{id}/color-presets`. */
  async updateColorPresets(
    canvasId: Uuid,
    body: UpdateCanvasColorPresetsRequest,
  ): Promise<CanvasColorPresets> {
    return this.transport.request<CanvasColorPresets>(
      "PATCH",
      `canvases/${canvasId}/color-presets`,
      body,
    );
  }

  /**
   * `GET /api/v1/canvases/{id}/preview` — download the canvas thumbnail.
   *
   * @returns Binary `Blob`. Callers that prefer a Node `Buffer` can call
   *   `Buffer.from(await blob.arrayBuffer())`.
   */
  async getPreview(canvasId: Uuid): Promise<Blob> {
    const response = await this.transport.rawRequest("GET", `canvases/${canvasId}/preview`);
    return response.blob();
  }

  /** `GET /api/v1/canvases/{id}/permissions`. */
  async getPermissions(canvasId: Uuid): Promise<CanvasPermissions> {
    return this.transport.request<CanvasPermissions>(
      "GET",
      `canvases/${canvasId}/permissions`,
    );
  }

  /** Subscribe to a canvas's permissions object. */
  subscribePermissions(
    canvasId: Uuid,
    opts?: StreamOptions,
  ): AsyncGenerator<CanvasPermissions, void, void> {
    return streamNdjson<CanvasPermissions>(
      this.transport,
      `canvases/${canvasId}/permissions`,
      opts,
    );
  }

  /** `POST /api/v1/canvases/{id}/permissions`. */
  async setPermissions(
    canvasId: Uuid,
    body: SetCanvasPermissionsRequest,
  ): Promise<CanvasPermissions> {
    return this.transport.request<CanvasPermissions>(
      "POST",
      `canvases/${canvasId}/permissions`,
      body,
    );
  }
}
