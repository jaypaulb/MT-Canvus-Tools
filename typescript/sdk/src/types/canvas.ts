import type { IsoDateTime, Uuid } from "./common.js";

/** Link-sharing permission applied to anonymous viewers. */
export type LinkPermission = "view" | "edit" | "none";

/** Per-user / per-group permission level. */
export type PermissionLevel = "view" | "edit" | "own" | "none" | "owner";

/**
 * A single permission override scoped to either a user or a group.
 *
 * Live-server verification: the server returns user/group permissions
 * as objects with integer `id`, a `permission` string, and a boolean
 * `inherited` flag. Field names are underscored.
 */
export interface CanvasUserPermission {
  readonly id: number;
  readonly permission: PermissionLevel;
  readonly inherited?: boolean;
}

export interface CanvasGroupPermission {
  readonly id: number;
  readonly permission: PermissionLevel;
  readonly inherited?: boolean;
}

/**
 * A canvas document as returned by the API.
 *
 * Live-server verification (2026-05-18): every field uses UNDERSCORED
 * names. `id`, `name`, `mode`, `state`, `access`, `asset_size`,
 * `folder_id`, `preview_hash`, `in_trash`, `modified_at`, `created_at`.
 */
export interface Canvas {
  readonly id: Uuid;
  readonly name: string;
  readonly mode: "demo" | "normal";
  readonly state: string;
  readonly access?: "rw" | "edit" | "view";
  readonly asset_size?: number;
  readonly folder_id?: Uuid;
  readonly preview_hash?: string;
  readonly in_trash?: boolean;
  readonly modified_at?: IsoDateTime;
  readonly created_at?: IsoDateTime;
  readonly description?: string;
  readonly link_permission?: LinkPermission;
  readonly owner_id?: string;
}

/** Request body for `POST /api/v1/canvases`. */
export interface CreateCanvasRequest {
  readonly name: string;
  readonly folder_id?: Uuid;
}

/** Request body for `PATCH /api/v1/canvases/{id}`. */
export interface UpdateCanvasRequest {
  readonly name?: string;
  readonly mode?: "demo" | "normal";
}

/** Request body for `POST /api/v1/canvases/{id}/move`. */
export interface MoveCanvasRequest {
  readonly folder_id: Uuid;
  readonly conflicts?: string;
}

/** Request body for `POST /api/v1/canvases/{id}/copy`. */
export interface CopyCanvasRequest {
  readonly folder_id: Uuid;
  readonly conflicts?: string;
}

/** Canvas background haze. */
export interface CanvasBackgroundHaze {
  readonly color1: string;
  readonly color2: string;
  readonly speed: number;
  readonly scale: number;
}

/** Canvas background grid overlay. */
export interface CanvasBackgroundGrid {
  readonly visible: boolean;
  readonly color: string;
}

/** Canvas background image asset. */
export interface CanvasBackgroundImage {
  readonly hash: string;
  readonly fit: "fill" | "contain" | "cover";
}

/** Canvas background configuration. */
export interface CanvasBackground {
  readonly type: "color" | "image" | "haze" | "none";
  readonly background_color?: string;
  readonly image?: CanvasBackgroundImage;
  readonly grid?: CanvasBackgroundGrid;
  readonly haze?: CanvasBackgroundHaze;
}

/** Request body for `PATCH /api/v1/canvases/{id}/background`. */
export type UpdateCanvasBackgroundRequest = Partial<CanvasBackground>;

/**
 * Canvas colour-preset configuration.
 *
 * Live-server verification: the four preset groups are returned as a
 * flat object with UNDERSCORED keys: `annotation`, `connector`,
 * `note_background`, `note_text`.
 */
export interface CanvasColorPresets {
  readonly annotation: readonly string[];
  readonly connector: readonly string[];
  readonly note_background: readonly string[];
  readonly note_text: readonly string[];
}

/** Request body for `PATCH /api/v1/canvases/{id}/color-presets`. */
export type UpdateCanvasColorPresetsRequest = Partial<CanvasColorPresets>;

/** Canvas permission configuration. */
export interface CanvasPermissions {
  readonly editors_can_share: boolean;
  readonly users: readonly CanvasUserPermission[];
  readonly groups: readonly CanvasGroupPermission[];
  readonly link_permission: LinkPermission;
}

/** Request body for `POST /api/v1/canvases/{id}/permissions`. */
export type SetCanvasPermissionsRequest = Partial<CanvasPermissions>;
