import type { Transport } from "../transport.js";
import { streamNdjson, type StreamOptions } from "../streaming.js";
import { AuthError } from "../errors.js";
import type { Uuid } from "../types/common.js";
import type { User } from "../types/user.js";
import type {
  CopyFolderRequest,
  CreateFolderRequest,
  Folder,
  FolderPermissions,
  MoveFolderRequest,
  SetFolderPermissionsRequest,
  UpdateFolderRequest,
} from "../types/folder.js";

/**
 * Canvas-folder management — 12 endpoints under `/api/v1/canvas-folders/*`.
 *
 * The server accepts both POST and PATCH methods for `/move` and `/copy`;
 * the SDK exposes both for compatibility.
 */
export class FoldersResource {
  constructor(private readonly transport: Transport) {}

  /** `GET /api/v1/canvas-folders`. */
  async list(): Promise<readonly Folder[]> {
    return this.transport.request<readonly Folder[]>("GET", "canvas-folders");
  }

  /** Subscribe to the folder tree. */
  subscribe(opts?: StreamOptions): AsyncGenerator<Folder, void, void> {
    return streamNdjson<Folder>(this.transport, "canvas-folders", opts);
  }

  /** `GET /api/v1/canvas-folders/{id}`. */
  async get(folderId: Uuid): Promise<Folder> {
    return this.transport.request<Folder>("GET", `canvas-folders/${folderId}`);
  }

  /** Subscribe to a single folder. */
  subscribeOne(folderId: Uuid, opts?: StreamOptions): AsyncGenerator<Folder, void, void> {
    return streamNdjson<Folder>(this.transport, `canvas-folders/${folderId}`, opts);
  }

  /** `POST /api/v1/canvas-folders`. */
  async create(body: CreateFolderRequest): Promise<Folder> {
    return this.transport.request<Folder>("POST", "canvas-folders", body);
  }

  /** `PATCH /api/v1/canvas-folders/{id}` — rename. */
  async update(folderId: Uuid, body: UpdateFolderRequest): Promise<Folder> {
    return this.transport.request<Folder>("PATCH", `canvas-folders/${folderId}`, body);
  }

  /** `DELETE /api/v1/canvas-folders/{id}` — delete an empty folder. */
  async delete(folderId: Uuid): Promise<void> {
    await this.transport.request<void>("DELETE", `canvas-folders/${folderId}`);
  }

  /** `DELETE /api/v1/canvas-folders/{id}/children` — purge contents. */
  async deleteChildren(folderId: Uuid): Promise<void> {
    await this.transport.request<void>("DELETE", `canvas-folders/${folderId}/children`);
  }

  /** `POST /api/v1/canvas-folders/{id}/move`. */
  async move(folderId: Uuid, body: MoveFolderRequest): Promise<Folder> {
    return this.transport.request<Folder>("POST", `canvas-folders/${folderId}/move`, body);
  }

  /**
   * Phase 4b §4.3 #4: move a folder into the current user's trash folder.
   *
   * Mirrors Go's `TrashFolder` (`folders.go:135`). Throws {@link AuthError}
   * if the calling credential doesn't resolve to a real user.
   */
  async trash(folderId: Uuid): Promise<Folder> {
    const user = await this.transport.request<User>("GET", "users/current");
    if (user.id === 0) {
      throw new AuthError(
        "missing-token",
        "folders.trash: cannot resolve current user — login required",
      );
    }
    return this.move(folderId, { folder_id: `trash.${user.id.toString()}` });
  }

  /** `PATCH /api/v1/canvas-folders/{id}/move` — alternative to POST /move. */
  async movePatch(folderId: Uuid, body: MoveFolderRequest): Promise<Folder> {
    return this.transport.request<Folder>("PATCH", `canvas-folders/${folderId}/move`, body);
  }

  /** `POST /api/v1/canvas-folders/{id}/copy`. */
  async copy(folderId: Uuid, body: CopyFolderRequest): Promise<Folder> {
    return this.transport.request<Folder>("POST", `canvas-folders/${folderId}/copy`, body);
  }

  /** `PATCH /api/v1/canvas-folders/{id}/copy` — alternative to POST /copy. */
  async copyPatch(folderId: Uuid, body: CopyFolderRequest): Promise<Folder> {
    return this.transport.request<Folder>("PATCH", `canvas-folders/${folderId}/copy`, body);
  }

  /** `GET /api/v1/canvas-folders/{id}/permissions`. */
  async getPermissions(folderId: Uuid): Promise<FolderPermissions> {
    return this.transport.request<FolderPermissions>(
      "GET",
      `canvas-folders/${folderId}/permissions`,
    );
  }

  /** Subscribe to a folder's permissions object. */
  subscribePermissions(
    folderId: Uuid,
    opts?: StreamOptions,
  ): AsyncGenerator<FolderPermissions, void, void> {
    return streamNdjson<FolderPermissions>(
      this.transport,
      `canvas-folders/${folderId}/permissions`,
      opts,
    );
  }

  /** `POST /api/v1/canvas-folders/{id}/permissions`. */
  async setPermissions(
    folderId: Uuid,
    body: SetFolderPermissionsRequest,
  ): Promise<FolderPermissions> {
    return this.transport.request<FolderPermissions>(
      "POST",
      `canvas-folders/${folderId}/permissions`,
      body,
    );
  }
}
