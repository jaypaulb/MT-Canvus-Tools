import type { IsoDateTime, Size, Uuid } from "./common.js";

/** Link-sharing permission applied to anonymous viewers. */
export type LinkPermission = "view" | "edit" | "none";

/** Per-user / per-group permission level. */
export type PermissionLevel = "view" | "edit" | "own" | "none";

/** A single permission override scoped to either a user or a group. */
export interface PermissionOverride {
  readonly "user-id"?: Uuid;
  readonly "group-id"?: Uuid;
  readonly permission: PermissionLevel;
}

/** A canvas document as returned by the API. */
export interface Canvas {
  readonly "canvas-id": Uuid;
  readonly "canvas-name": string;
  readonly owner: string;
  readonly "demo-canvas": boolean;
  readonly "has-main-password": boolean;
  readonly "has-access-password": boolean;
  readonly size: Size;
  readonly created: IsoDateTime;
  readonly modified: IsoDateTime;
  readonly "link-permission": LinkPermission;
  readonly "permission-overrides": readonly PermissionOverride[];
  readonly "parent-folder-id"?: Uuid;
}

/** Request body for `POST /api/v1/canvases`. */
export interface CreateCanvasRequest {
  readonly "canvas-name": string;
  readonly "new-width"?: number;
  readonly "new-height"?: number;
  readonly "initial-main-password"?: string;
  readonly "initial-access-password"?: string;
}

/** Request body for `PATCH /api/v1/canvases/{id}`. */
export interface UpdateCanvasRequest {
  readonly "canvas-name"?: string;
}

/** Request body for `POST /api/v1/canvases/{id}/move`. */
export interface MoveCanvasRequest {
  readonly "parent-folder-id": Uuid;
}

/** Request body for `POST /api/v1/canvases/{id}/copy`. */
export interface CopyCanvasRequest {
  readonly "canvas-name": string;
  readonly "parent-folder-id"?: Uuid;
}

/** Canvas background configuration. */
export interface CanvasBackground {
  readonly "background-type": "color" | "image" | "none";
  readonly "background-color"?: string;
  readonly "background-image"?: string;
  readonly "image-fit"?: "fill" | "contain" | "cover";
  readonly "grid-visible"?: boolean;
  readonly "grid-size"?: number;
  readonly "haze-visible"?: boolean;
  readonly "haze-color"?: string;
  readonly "haze-opacity"?: number;
}

/** Request body for `PATCH /api/v1/canvases/{id}/background`. */
export type UpdateCanvasBackgroundRequest = Partial<CanvasBackground>;

/** Canvas colour-preset configuration. */
export interface CanvasColorPresets {
  readonly "annotation-colors": readonly string[];
  readonly "note-background-colors": readonly string[];
  readonly "note-text-colors": readonly string[];
  readonly "connector-colors"?: readonly string[];
}

/** Request body for `PATCH /api/v1/canvases/{id}/color-presets`. */
export type UpdateCanvasColorPresetsRequest = Partial<CanvasColorPresets>;

/** Canvas permission configuration. */
export interface CanvasPermissions {
  readonly "link-permission": LinkPermission;
  readonly "permission-overrides": readonly PermissionOverride[];
}

/** Request body for `POST /api/v1/canvases/{id}/permissions`. */
export type SetCanvasPermissionsRequest = Partial<CanvasPermissions>;
