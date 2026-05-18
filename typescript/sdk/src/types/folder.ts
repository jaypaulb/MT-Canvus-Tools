import type { IsoDateTime, Uuid } from "./common.js";
import type { PermissionLevel } from "./canvas.js";

/** A child reference within a folder. */
export interface FolderChild {
  readonly type: "folder" | "canvas";
  readonly id: Uuid;
  readonly name: string;
}

/**
 * Canvas folder as returned by the API.
 *
 * Live-server verification: all fields UNDERSCORED. `folder_id` here is
 * the parent folder reference (matching the Go SDK field name).
 */
export interface Folder {
  readonly id: Uuid;
  readonly name: string;
  readonly folder_id?: Uuid;
  readonly access?: string;
  readonly in_trash?: boolean;
  readonly state?: string;
  readonly created_at?: IsoDateTime;
  readonly modified_at?: IsoDateTime;
  readonly children?: readonly FolderChild[];
}

/**
 * A single subject-permission override on a folder.
 *
 * Mirrors the canvas-permission shape: integer `id`, permission level,
 * and an `inherited` flag.
 */
export interface FolderUserPermission {
  readonly id: number;
  readonly permission: PermissionLevel;
  readonly inherited?: boolean;
}

export interface FolderGroupPermission {
  readonly id: number;
  readonly permission: PermissionLevel;
  readonly inherited?: boolean;
}

/** Folder permission configuration. */
export interface FolderPermissions {
  readonly editors_can_share: boolean;
  readonly users: readonly FolderUserPermission[];
  readonly groups: readonly FolderGroupPermission[];
}

/** Create-folder request body. */
export interface CreateFolderRequest {
  readonly name: string;
  readonly folder_id?: Uuid;
}

/** Update-folder request body (rename). */
export interface UpdateFolderRequest {
  readonly name?: string;
}

/** Move-folder request body. */
export interface MoveFolderRequest {
  readonly folder_id: Uuid;
  readonly conflicts?: string;
}

/** Copy-folder request body. */
export interface CopyFolderRequest {
  readonly folder_id: Uuid;
  readonly conflicts?: string;
}

/** Set-folder-permissions request body. */
export type SetFolderPermissionsRequest = Partial<FolderPermissions>;
