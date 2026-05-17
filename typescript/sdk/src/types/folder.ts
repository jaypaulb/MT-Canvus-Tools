import type { IsoDateTime, Uuid } from "./common.js";
import type { PermissionLevel, PermissionOverride } from "./canvas.js";

/** A child reference within a folder. */
export interface FolderChild {
  readonly type: "folder" | "canvas";
  readonly id: Uuid;
  readonly name: string;
}

/** Canvas folder as returned by the API. */
export interface Folder {
  readonly "folder-id": Uuid;
  readonly "folder-name": string;
  readonly "parent-folder-id"?: Uuid;
  readonly owner?: string;
  readonly created?: IsoDateTime;
  readonly modified?: IsoDateTime;
  readonly "permission-overrides"?: readonly PermissionOverride[];
  readonly children?: readonly FolderChild[];
}

/** Folder permission configuration. */
export interface FolderPermissions {
  readonly "folder-id"?: Uuid;
  readonly "default-permission": PermissionLevel;
  readonly "permission-overrides": readonly PermissionOverride[];
}

/** Create-folder request body. */
export interface CreateFolderRequest {
  readonly "folder-name": string;
  readonly "parent-folder-id"?: Uuid;
}

/** Update-folder request body. */
export interface UpdateFolderRequest {
  readonly "folder-name"?: string;
}

/** Move-folder request body. */
export interface MoveFolderRequest {
  readonly "parent-folder-id": Uuid;
}

/** Copy-folder request body. */
export interface CopyFolderRequest {
  readonly "folder-name": string;
  readonly "parent-folder-id"?: Uuid;
}

/** Set-folder-permissions request body. */
export type SetFolderPermissionsRequest = Partial<
  Pick<FolderPermissions, "default-permission" | "permission-overrides">
>;
