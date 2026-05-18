import type { Transport } from "../transport.js";
import { streamNdjson, type StreamOptions } from "../streaming.js";
import { AuthError } from "../errors.js";
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
import type { User } from "../types/user.js";

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

  /**
   * Phase 4b §4.3 #3: move a canvas into the current user's trash folder.
   *
   * Implements the same client-side pattern as Go's `TrashCanvas`
   * (`canvases.go:95`): look up the calling user, then PATCH the canvas
   * `folder_id` to the magic `trash.{userId}` folder.
   *
   * Requires an authenticated session — calls `GET /users/current` to
   * resolve the user ID. Throws {@link AuthError} if the server replies
   * with no user (e.g. an unauthenticated API key).
   */
  async trash(canvasId: Uuid): Promise<Canvas> {
    const user = await this.transport.request<User>("GET", "users/current");
    if (user.id === 0) {
      throw new AuthError(
        "missing-token",
        "canvases.trash: cannot resolve current user — login required",
      );
    }
    return this.move(canvasId, { folder_id: `trash.${user.id.toString()}` });
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

  // -------------------------------------------------------------------------
  // Phase 4b §4.3 #6 — single color-preset CRUD (client-side decomposition).
  //
  // The server exposes only the bulk `GET/PATCH /color-presets` object. These
  // helpers split it into per-name views so callers don't have to manage the
  // four parallel arrays. Mirrors Go's `colorpresets.go:38-79`.
  // -------------------------------------------------------------------------

  /**
   * List every (preset-group, color) pair on the canvas.
   *
   * Each entry's `name` encodes the group + index, e.g.
   * `note_background.0`, `connector.3`. Callers reuse the same name on
   * subsequent get/update/delete calls.
   */
  async listColorPresets(canvasId: Uuid): Promise<readonly { name: string; color: string }[]> {
    const all = await this.getColorPresets(canvasId);
    return flattenColorPresets(all);
  }

  /**
   * Look up a single preset by encoded name (e.g. `connector.2`).
   * Returns `undefined` if the name doesn't exist.
   */
  async getColorPreset(
    canvasId: Uuid,
    name: string,
  ): Promise<{ name: string; color: string } | undefined> {
    const list = await this.listColorPresets(canvasId);
    return list.find((p) => p.name === name);
  }

  /**
   * Append a colour to a preset group.
   *
   * `group` must be one of the four preset groups
   * (`annotation`, `connector`, `note_background`, `note_text`).
   */
  async createColorPreset(
    canvasId: Uuid,
    group: keyof CanvasColorPresets,
    color: string,
  ): Promise<CanvasColorPresets> {
    const current = await this.getColorPresets(canvasId);
    const next = { ...current, [group]: [...current[group], color] };
    return this.updateColorPresets(canvasId, { [group]: next[group] });
  }

  /**
   * Replace the colour at the encoded preset name (`group.index`).
   *
   * Throws if the name doesn't decode to a valid group + index.
   */
  async updateColorPreset(
    canvasId: Uuid,
    name: string,
    color: string,
  ): Promise<CanvasColorPresets> {
    const { group, index } = decodePresetName(name);
    const current = await this.getColorPresets(canvasId);
    const arr = [...current[group]];
    if (index < 0 || index >= arr.length) {
      throw new Error(`updateColorPreset: index out of range for ${name}`);
    }
    arr[index] = color;
    return this.updateColorPresets(canvasId, { [group]: arr });
  }

  /**
   * Remove the colour at the encoded preset name (`group.index`).
   */
  async deleteColorPreset(canvasId: Uuid, name: string): Promise<CanvasColorPresets> {
    const { group, index } = decodePresetName(name);
    const current = await this.getColorPresets(canvasId);
    const arr = [...current[group]];
    if (index < 0 || index >= arr.length) {
      throw new Error(`deleteColorPreset: index out of range for ${name}`);
    }
    arr.splice(index, 1);
    return this.updateColorPresets(canvasId, { [group]: arr });
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

/** Encoded preset name → `{ group, index }`. */
function decodePresetName(name: string): {
  group: keyof CanvasColorPresets;
  index: number;
} {
  const dot = name.lastIndexOf(".");
  if (dot <= 0) {
    throw new Error(`invalid preset name ${JSON.stringify(name)}: expected "group.index"`);
  }
  const group = name.slice(0, dot);
  const index = Number.parseInt(name.slice(dot + 1), 10);
  if (Number.isNaN(index) || index < 0) {
    throw new Error(`invalid preset name ${JSON.stringify(name)}: index is not a non-negative integer`);
  }
  if (
    group !== "annotation" &&
    group !== "connector" &&
    group !== "note_background" &&
    group !== "note_text"
  ) {
    throw new Error(
      `invalid preset name ${JSON.stringify(name)}: unknown group ${JSON.stringify(group)}`,
    );
  }
  return { group, index };
}

/** Flatten a CanvasColorPresets object into a per-name list. */
function flattenColorPresets(
  presets: CanvasColorPresets,
): readonly { name: string; color: string }[] {
  const out: { name: string; color: string }[] = [];
  const groups: (keyof CanvasColorPresets)[] = [
    "annotation",
    "connector",
    "note_background",
    "note_text",
  ];
  for (const group of groups) {
    const arr = presets[group];
    for (let i = 0; i < arr.length; i++) {
      const color = arr[i];
      if (color !== undefined) {
        out.push({ name: `${group}.${i.toString()}`, color });
      }
    }
  }
  return out;
}
