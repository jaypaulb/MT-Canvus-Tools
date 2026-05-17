# SDK Coverage Matrix

Audited: 2026-05-17
Spec freeze: see [SOURCE.md](SOURCE.md)

> **Status taxonomy:** `implemented` / `missing` / `outdated` / `deprecated`. See `_coverage-*` intermediates (deleted post-merge) for full per-SDK detail; this file is the unified view.

## Summary

| Resource group | Total endpoints | Go (impl/miss/old/dep) | Python (impl/miss/old/dep) |
|---|---|---|---|
| Canvases | 17 | 16/0/1/0 | 14/0/1/0 |
| Widgets | 64 | 39/19/6/0 | 49/16/2/1 |
| Auth | 13 | 13/0/0/0 | 9/0/0/0 |
| Users | 19 | 19/0/0/0 | 9/2/0/0 |
| Folders | 12 | 11/1/0/0 | 10/0/0/0 |
| Assets | 3 | 3/0/0/0 | 3/0/0/0 |
| Server | 22 | 21/0/1/0 | 17/5/0/0 |
| **TOTAL** | **150** | **122/20/8/0** | **111/23/3/1** |

---

## Per-endpoint detail

### Canvases

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/canvases` | implemented | `canvases.go:ListCanvases` | implemented | `CanvusClient.list_canvases` | Go adds client-side `Filter` arg not in spec (benign); Python supports `params` + optional `Filter`. |
| `GET /api/v1/canvases/{id}` | implemented | `canvases.go:GetCanvas` | implemented | `CanvusClient.get_canvas` | — |
| `POST /api/v1/canvases` | implemented | `canvases.go:CreateCanvas` | implemented | `CanvusClient.create_canvas` | — |
| `PATCH /api/v1/canvases/{id}` | implemented | `canvases.go:UpdateCanvas` | implemented | `CanvusClient.update_canvas` | — |
| `DELETE /api/v1/canvases/{id}` | implemented | `canvases.go:DeleteCanvas` | implemented | `CanvusClient.delete_canvas` | — |
| `POST /api/v1/canvases/{id}/move` | implemented | `canvases.go:MoveCanvas` | implemented | `CanvusClient.move_canvas` | Go wraps via `TrashCanvas`; Python takes `folder_id` arg. |
| `POST /api/v1/canvases/{id}/copy` | implemented | `canvases.go:CopyCanvas` | implemented | `CanvusClient.copy_canvas` | — |
| `POST /api/v1/canvases/{id}/save` | implemented | `canvases.go:SaveDemoState` | implemented | `CanvusClient.save_demo_state` | Python also exposes via `set_canvas_mode(is_demo=True)`. |
| `POST /api/v1/canvases/{id}/restore` | implemented | `canvases.go:RestoreDemoCanvas` | implemented | `CanvusClient.restore_demo_state` | — |
| `GET /api/v1/canvases/{id}/background` | implemented | `backgrounds.go:GetCanvasBackground` | implemented | `CanvusClient.get_canvas_background` | — |
| `PATCH /api/v1/canvases/{id}/background` | implemented | `backgrounds.go:PatchCanvasBackground` | implemented | `CanvusClient.set_canvas_background` | — |
| `POST /api/v1/canvases/{id}/background` | implemented | `backgrounds.go:PostCanvasBackground` | implemented | `CanvusClient.set_canvas_background_image` | Python notes multipart upload. |
| `GET /api/v1/canvases/{id}/color-presets` | outdated | `colorpresets.go:GetColorPresets` / `ListColorPresets` | implemented | `CanvusClient.get_color_presets` | Go: **Path mismatch:** SDK hits `canvases/{id}/colorpresets` (no hyphen); spec is `color-presets`. Go also exposes per-name CRUD (`GetColorPreset`, etc.) not in spec. Python: correctly implements with hyphens. |
| `PATCH /api/v1/canvases/{id}/color-presets` | outdated | `colorpresets.go:PatchColorPresets` | implemented | `CanvusClient.update_color_presets` | Go: same hyphen/no-hyphen path drift as GET. Python correctly uses hyphens. |
| `GET /api/v1/canvases/{id}/preview` | implemented | `canvases.go:GetCanvasPreview` | implemented | `CanvusClient.get_canvas_preview` | Both return raw bytes. |
| `GET /api/v1/canvases/{id}/permissions` | implemented | `canvases.go:GetCanvasPermissions` | implemented | `CanvusClient.get_canvas_permissions` | — |
| `POST /api/v1/canvases/{id}/permissions` | implemented | `canvases.go:SetCanvasPermissions` | outdated | `CanvusClient.set_canvas_permissions` | Python signature accepts plain `payload` only; spec body shape `{link-permission, permission-overrides[]}` is not validated/typed. |

### Widgets — Generic

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/canvases/{id}/widgets` | implemented | `widgets.go:ListWidgets` | implemented | `CanvusClient.list_widgets` | Go supports `annotations=1`; Python supports `filter_obj`. Both documented to support `subscribe`. |
| `GET /api/v1/canvases/{id}/widgets/{widget-id}` | implemented | `widgets.go:GetWidget` | implemented | `CanvusClient.get_widget` | — |
| `POST /api/v1/canvases/{id}/widgets/clone` | outdated | none / `widgets.go:MoveWidget`,`CopyWidget` | deprecated | — (no SDK method) | Go: Spec marks "experimental (501)"; changelog §1 says cloning via standard create with `source_canvas_id`/`source_widget_id`. SDK has `MoveWidget`/`CopyWidget` hitting non-spec path. Python: No SDK method; spec deprecated; changelog directs use of standard create endpoints. Both need `CloneWidget` helper. |

### Widgets — Notes

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/canvases/{id}/notes` | implemented | `notes.go:ListNotes` | implemented | `CanvusClient.list_notes` | — |
| `GET /api/v1/canvases/{id}/notes/{note-id}` | implemented | `notes.go:GetNote` | implemented | `CanvusClient.get_note` | — |
| `POST /api/v1/canvases/{id}/notes` | implemented | `notes.go:CreateNote` | implemented | `CanvusClient.create_note` | Go and Python both do not yet accept `source_canvas_id`/`source_widget_id` for cross-canvas clone. |
| `PATCH /api/v1/canvases/{id}/notes/{note-id}` | implemented | `notes.go:UpdateNote` | implemented | `CanvusClient.update_note` | — |
| `DELETE /api/v1/canvases/{id}/notes/{note-id}` | implemented | `notes.go:DeleteNote` | implemented | `CanvusClient.delete_note` | — |

### Widgets — Images

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/canvases/{id}/images` | implemented | `images.go:ListImages` | implemented | `CanvusClient.list_images` | — |
| `GET /api/v1/canvases/{id}/images/{image-id}` | implemented | `images.go:GetImage` | implemented | `CanvusClient.get_image` | — |
| `POST /api/v1/canvases/{id}/images` | implemented | `images.go:CreateImage` | implemented | `CanvusClient.create_image` | Go and Python both multipart; neither supports clone params yet. |
| `PATCH /api/v1/canvases/{id}/images/{image-id}` | implemented | `images.go:UpdateImage` | implemented | `CanvusClient.update_image` | Go emits `WarningImageAspectRatioNotPreserved`. |
| `GET /api/v1/canvases/{id}/images/{image-id}/download` | implemented | `images.go:DownloadImage` | implemented | `CanvusClient.download_image` | Both return bytes. |
| `DELETE /api/v1/canvases/{id}/images/{image-id}` | implemented | `images.go:DeleteImage` | implemented | `CanvusClient.delete_image` | — |

### Widgets — Videos

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/canvases/{id}/videos` | implemented | `videos.go:ListVideos` | implemented | `CanvusClient.list_videos` | — |
| `GET /api/v1/canvases/{id}/videos/{video-id}` | implemented | `videos.go:GetVideo` | implemented | `CanvusClient.get_video` | — |
| `POST /api/v1/canvases/{id}/videos` | implemented | `videos.go:CreateVideo` | implemented | `CanvusClient.create_video` | Go and Python both multipart; clone params not supported. |
| `PATCH /api/v1/canvases/{id}/videos/{video-id}` | implemented | `videos.go:UpdateVideo` | implemented | `CanvusClient.update_video` | Go emits `WarningVideoAspectRatioNotPreserved`. |
| `GET /api/v1/canvases/{id}/videos/{video-id}/download` | implemented | `videos.go:DownloadVideo` | implemented | `CanvusClient.download_video` | Both return bytes. |
| `DELETE /api/v1/canvases/{id}/videos/{video-id}` | implemented | `videos.go:DeleteVideo` | implemented | `CanvusClient.delete_video` | — |

### Widgets — PDFs

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/canvases/{id}/pdfs` | implemented | `pdfs.go:ListPDFs` | implemented | `CanvusClient.list_pdfs` | — |
| `GET /api/v1/canvases/{id}/pdfs/{pdf-id}` | implemented | `pdfs.go:GetPDF` | implemented | `CanvusClient.get_pdf` | — |
| `POST /api/v1/canvases/{id}/pdfs` | implemented | `pdfs.go:CreatePDF` | implemented | `CanvusClient.create_pdf` | Go and Python both multipart; clone params not supported. |
| `PATCH /api/v1/canvases/{id}/pdfs/{pdf-id}` | implemented | `pdfs.go:UpdatePDF` | implemented | `CanvusClient.update_pdf` | Go emits `WarningPDFSizeBug`. |
| `GET /api/v1/canvases/{id}/pdfs/{pdf-id}/download` | implemented | `pdfs.go:DownloadPDF` | implemented | `CanvusClient.download_pdf` | Both return bytes. |
| `DELETE /api/v1/canvases/{id}/pdfs/{pdf-id}` | implemented | `pdfs.go:DeletePDF` | implemented | `CanvusClient.delete_pdf` | — |

### Widgets — Browsers

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/canvases/{id}/browsers` | implemented | `browsers.go:ListBrowsers` | implemented | `CanvusClient.list_browsers` | — |
| `GET /api/v1/canvases/{id}/browsers/{browser-id}` | implemented | `browsers.go:GetBrowser` | implemented | `CanvusClient.get_browser` | — |
| `POST /api/v1/canvases/{id}/browsers` | implemented | `browsers.go:CreateBrowser` | implemented | `CanvusClient.create_browser` | Both do not support clone params yet. |
| `PATCH /api/v1/canvases/{id}/browsers/{browser-id}` | implemented | `browsers.go:UpdateBrowser` | implemented | `CanvusClient.update_browser` | — |
| `DELETE /api/v1/canvases/{id}/browsers/{browser-id}` | implemented | `browsers.go:DeleteBrowser` | implemented | `CanvusClient.delete_browser` | — |

### Widgets — Anchors

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/canvases/{id}/anchors` | implemented | `anchors.go:ListAnchors` | implemented | `CanvusClient.list_anchors` | — |
| `GET /api/v1/canvases/{id}/anchors/{anchor-id}` | implemented | `anchors.go:GetAnchor` | implemented | `CanvusClient.get_anchor` | — |
| `POST /api/v1/canvases/{id}/anchors` | implemented | `anchors.go:CreateAnchor` | implemented | `CanvusClient.create_anchor` | Both do not support clone params. Go has widget-creation-on-the-fly behaviour for `src`/`dst`. |
| `PATCH /api/v1/canvases/{id}/anchors/{anchor-id}` | implemented | `anchors.go:UpdateAnchor` | implemented | `CanvusClient.update_anchor` | Python has circular-parenting guard on `parent_id`. |
| `DELETE /api/v1/canvases/{id}/anchors/{anchor-id}` | implemented | `anchors.go:DeleteAnchor` | implemented | `CanvusClient.delete_anchor` | — |

### Widgets — Connectors

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/canvases/{id}/connectors` | implemented | `connectors.go:ListConnectors` | implemented | `CanvusClient.list_connectors` | — |
| `GET /api/v1/canvases/{id}/connectors/{connector-id}` | implemented | `connectors.go:GetConnector` | implemented | `CanvusClient.get_connector` | — |
| `POST /api/v1/canvases/{id}/connectors` | implemented | `connectors.go:CreateConnector` | implemented | `CanvusClient.create_connector` | Go adds widget-creation-on-the-fly behaviour for `src`/`dst`. |
| `PATCH /api/v1/canvases/{id}/connectors/{connector-id}` | implemented | `connectors.go:UpdateConnector` | implemented | `CanvusClient.update_connector` | — |
| `DELETE /api/v1/canvases/{id}/connectors/{connector-id}` | implemented | `connectors.go:DeleteConnector` | implemented | `CanvusClient.delete_connector` | — |

### Widgets — Tables

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/canvases/{id}/tables` | missing | — | missing | — | Python audit grouped table endpoints differently; counted as "missing" for both. Neither SDK has `tables.go`/`list_tables`. |
| `GET /api/v1/canvases/{id}/tables/{table-id}` | missing | — | missing | — | — |
| `POST /api/v1/canvases/{id}/tables` | missing | — | missing | — | Per changelog §4 and §5: omit `column_widths`/`row_heights`; document that `grid_size` is set at create and silently ignored in PATCH. |
| `PATCH /api/v1/canvases/{id}/tables/{table-id}` | missing | — | missing | — | Python audit warns: document silent-ignore of `grid_size`. |
| `GET /api/v1/canvases/{id}/tables/{table-id}/cells` | missing | — | missing | — | — |
| `DELETE /api/v1/canvases/{id}/tables/{table-id}` | missing | — | missing | — | — |

### Widgets — Video Inputs (canvas-scoped)

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/canvases/{id}/video-inputs` | implemented | `videoinputs.go:ListVideoInputs` | implemented | `CanvusClient.list_canvas_video_inputs` | Go emits `WarningVideoInputTitleNotExposed`. |
| `GET /api/v1/canvases/{id}/video-inputs/{widget-id}` | implemented | `videoinputs.go:GetVideoInput` | missing | — | Go: implemented. Python: no single-widget GET. |
| `POST /api/v1/canvases/{id}/video-inputs` | implemented | `videoinputs.go:CreateVideoInput` | implemented | `CanvusClient.create_video_input` | — |
| `PATCH /api/v1/canvases/{id}/video-inputs/{widget-id}` | implemented | `videoinputs.go:UpdateVideoInput` | missing | — | Go: implemented. Python: no update method. |
| `DELETE /api/v1/canvases/{id}/video-inputs/{widget-id}` | implemented | `videoinputs.go:DeleteVideoInput` | implemented | `CanvusClient.delete_video_input` | — |

### Widgets — IP Videos

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/canvases/{id}/ip-videos` | missing | — | missing | — | No `ipvideos.go` file (Go); no SDK method (Python). |
| `GET /api/v1/canvases/{id}/ip-videos/{widget-id}` | missing | — | missing | — | — |
| `POST /api/v1/canvases/{id}/ip-videos` | outdated | — | missing (do NOT add) | — | Per changelog §2: POST is **not supported** server-side. Neither SDK should expose a Create method. Go counts "outdated" since spec row still lists it; Python counts "missing" since it was never implemented. |
| `PATCH /api/v1/canvases/{id}/ip-videos/{widget-id}` | missing | — | missing | — | — |
| `DELETE /api/v1/canvases/{id}/ip-videos/{widget-id}` | missing | — | missing | — | — |

### Widgets — RDP Connections

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/canvases/{id}/rdp-connections` | missing | — | missing | — | No `rdpconnections.go` file (Go); no SDK method (Python). Field-naming verification pending per changelog §3 (hyphen vs underscore). |
| `GET /api/v1/canvases/{id}/rdp-connections/{widget-id}` | missing | — | missing | — | — |
| `POST /api/v1/canvases/{id}/rdp-connections` | outdated | — | missing (do NOT add) | — | Per changelog §2: POST is **not supported** server-side. Go counts "outdated"; Python never implemented, counts "missing". Neither should expose Create. |
| `PATCH /api/v1/canvases/{id}/rdp-connections/{widget-id}` | missing | — | missing | — | — |
| `DELETE /api/v1/canvases/{id}/rdp-connections/{widget-id}` | missing | — | missing | — | — |

### Widgets — Uploads Folder

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/canvases/{id}/uploads-folder` | missing | — | missing | — | Go: `uploads.go` only has POST helpers; no GET listing. Python: no SDK list method. |
| `POST /api/v1/canvases/{id}/uploads-folder` | implemented | `uploads.go:UploadNote`, `uploads.go:UploadAsset` | implemented | `CanvusClient.upload_note`, `CanvusClient.upload_file` | Both POST to same endpoint with two specialized wrappers. |

### Auth

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `POST /api/v1/users/login` | implemented | `session.go:Login` | implemented | `CanvusClient.login` | Go uses `username` key; spec doc says `email` — possible doc/behaviour drift. Python follows standard. |
| `POST /api/v1/users/login/saml` | implemented | `users.go:SamlLogin` | implemented | `CanvusClient.login_saml` | Python: no request-body args; spec expects `inResponseTo`, `responseXml`, `remember` — signature drift. Phase 3 item. |
| `POST /api/v1/users/logout` | implemented | `session.go:Logout` | implemented | `CanvusClient.logout` | Python supports optional token override. |
| `POST /api/v1/users/password/create-reset-token` | implemented | `users.go:CreateResetToken` | implemented | `CanvusClient.request_password_reset` | — |
| `GET /api/v1/users/password/validate-reset-token` | implemented | `users.go:ValidateResetToken` | implemented | `CanvusClient.validate_reset_token` | — |
| `POST /api/v1/users/password/reset` | implemented | `users.go:ResetUserPassword` | implemented | `CanvusClient.reset_password` | — |
| `POST /api/v1/users/register` | implemented | `users.go:RegisterUser` | implemented | `CanvusClient.register_user` | — |
| `POST /api/v1/users/confirm-email` | implemented | `users.go:ConfirmEmail` | implemented | `CanvusClient.confirm_email` | — |
| `GET /api/v1/users/{user-id}/access-tokens` | implemented | `accesstokens.go:ListAccessTokens` | implemented | `CanvusClient.list_tokens` | Python: uses `user_id: int`; spec uses UUID string — type drift (Phase 3 item). |
| `GET /api/v1/users/{user-id}/access-tokens/{token-id}` | implemented | `accesstokens.go:GetAccessToken` | implemented | `CanvusClient.get_token` | Same type drift as above. |
| `POST /api/v1/users/{user-id}/access-tokens` | implemented | `accesstokens.go:CreateAccessToken` | implemented | `CanvusClient.create_token` | Python takes only `description` arg; spec body includes `name`, `expires`, `scopes` — signature drift (Phase 3 item). |
| `PATCH /api/v1/users/{user-id}/access-tokens/{token-id}` | implemented | `accesstokens.go:UpdateAccessToken` | implemented | `CanvusClient.update_token` | — |
| `DELETE /api/v1/users/{user-id}/access-tokens/{token-id}` | implemented | `accesstokens.go:DeleteAccessToken` | implemented | `CanvusClient.delete_token` | — |

### Users

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/users` | implemented | `users.go:ListUsers` | implemented | `CanvusClient.list_users` | — |
| `GET /api/v1/users/{user-id}` | implemented | `users.go:GetUser` | implemented | `CanvusClient.get_user` | Go uses `int64` ID; Python uses `int` — both have type drift vs spec UUID. |
| `POST /api/v1/users` | implemented | `users.go:CreateUser` | implemented | `CanvusClient.create_user` | — |
| `PATCH /api/v1/users/{user-id}` | implemented | `users.go:UpdateUser` | implemented | `CanvusClient.update_user` | — |
| `POST /api/v1/users/{user-id}/password` | implemented | `users.go:SetUserPassword` | implemented | `CanvusClient.change_password` | Go: admin-path only (`password` body field); doesn't expose `old-password` for self-change. Python: implements standard change. |
| `POST /api/v1/users/{user-id}/change-email` | implemented | `users.go:ChangeUserEmail` | missing | — | Go implemented; Python missing. |
| `POST /api/v1/users/{user-id}/block` | implemented | `users.go:BlockUser` | implemented | `CanvusClient.block_user` | — |
| `POST /api/v1/users/{user-id}/unblock` | implemented | `users.go:UnblockUser` | implemented | `CanvusClient.unblock_user` | — |
| `POST /api/v1/users/{user-id}/approve` | implemented | `users.go:ApproveUser` | implemented | `CanvusClient.approve_user` | — |
| `POST /api/v1/users/{user-id}/reset-password` | implemented | `users.go:ForcePasswordResetUser` | missing | — | Go: admin-force-reset. Python missing; only password-flow self-reset exists. |
| `DELETE /api/v1/users/{user-id}` | implemented | `users.go:DeleteUser` | implemented | `CanvusClient.delete_user` | — |

### Users — Groups

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/groups` | implemented | `groups.go:ListGroups` | implemented | `CanvusClient.list_groups` | — |
| `GET /api/v1/groups/{group-id}` | implemented | `groups.go:GetGroup` | implemented | `CanvusClient.get_group` | — |
| `POST /api/v1/groups` | implemented | `groups.go:CreateGroup` | implemented | `CanvusClient.create_group` | — |
| `PATCH /api/v1/groups/{group-id}` | implemented | `groups.go:UpdateGroup` | missing | — | Go implemented; Python missing (Phase 3 item). |
| `DELETE /api/v1/groups/{group-id}` | implemented | `groups.go:DeleteGroup` | implemented | `CanvusClient.delete_group` | — |
| `GET /api/v1/groups/{group-id}/members` | implemented | `groups.go:ListGroupMembers` | implemented | `CanvusClient.list_group_members` | — |
| `POST /api/v1/groups/{group-id}/members` | implemented | `groups.go:AddUserToGroup` | implemented | `CanvusClient.add_user_to_group` | — |
| `DELETE /api/v1/groups/{group-id}/members/{user-id}` | implemented | `groups.go:RemoveUserFromGroup` | implemented | `CanvusClient.remove_user_from_group` | — |

### Folders

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/canvas-folders` | implemented | `folders.go:ListFolders` | implemented | `CanvusClient.list_folders` | — |
| `GET /api/v1/canvas-folders/{folder-id}` | implemented | `folders.go:GetFolder` | implemented | `CanvusClient.get_folder` | — |
| `POST /api/v1/canvas-folders` | implemented | `folders.go:CreateFolder` | implemented | `CanvusClient.create_folder` | — |
| `PATCH /api/v1/canvas-folders/{folder-id}` | implemented | `folders.go:RenameFolder` | implemented | `CanvusClient.update_folder` | — |
| `DELETE /api/v1/canvas-folders/{folder-id}` | implemented | `folders.go:DeleteFolder` | implemented | `CanvusClient.delete_folder` | — |
| `DELETE /api/v1/canvas-folders/{folder-id}/children` | implemented | `folders.go:DeleteFolderContents` | implemented | `CanvusClient.delete_folder_children` | — |
| `POST /api/v1/canvas-folders/{folder-id}/move` | implemented | `folders.go:MoveFolder` | implemented | `CanvusClient.move_folder` | Go wraps via `TrashFolder`. Both support POST. |
| `PATCH /api/v1/canvas-folders/{folder-id}/move` | missing | — | implemented (same method) | `CanvusClient.move_folder` | Go only exposes POST variant. Python supports POST (spec lists both as acceptable). |
| `POST /api/v1/canvas-folders/{folder-id}/copy` | implemented | `folders.go:CopyFolder` | implemented | `CanvusClient.copy_folder` | — |
| `PATCH /api/v1/canvas-folders/{folder-id}/copy` | implemented | `folders.go:CopyFolder` (POST-only) | implemented (same method) | `CanvusClient.copy_folder` | Go and Python both call POST; spec lists both variants — semantically equivalent. |
| `GET /api/v1/canvas-folders/{folder-id}/permissions` | implemented | `folders.go:GetFolderPermissions` | implemented | `CanvusClient.get_folder_permissions` | — |
| `POST /api/v1/canvas-folders/{folder-id}/permissions` | implemented | `folders.go:SetFolderPermissions` | implemented | `CanvusClient.set_folder_permissions` | — |

### Assets

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/assets/{hash}` | implemented | `mipmaps.go:GetAssetByHash` | implemented | `CanvusClient.get_asset_file` | Both use `canvas-id` header per spec. Go path-builds with explicit `api/v1/` prefix; verify no double-up with BaseURL. |
| `GET /api/v1/mipmaps/{hash}` | implemented | `mipmaps.go:GetMipmapInfo` | implemented | `CanvusClient.get_mipmap_info` | — |
| `GET /api/v1/mipmaps/{hash}/{level}` | implemented | `mipmaps.go:GetMipmapLevel` | implemented | `CanvusClient.get_mipmap_level_image` | — |

### Server — Server Info

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/server-info` | implemented | `serverinfo.go:GetServerInfo` | implemented | `CanvusClient.get_server_info` | — |

### Server — Configuration

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/server-config` | implemented | `serverconfig.go:GetServerConfig` | implemented | `CanvusClient.get_server_config` | Go: type models as nested struct; spec shows flat array. **Possible response-shape drift** — verify. Python: similarly decodes/returns nested config. |
| `PATCH /api/v1/server-config` | implemented | `serverconfig.go:UpdateServerConfig` | implemented | `CanvusClient.update_server_config` | Same shape drift as GET. |
| `POST /api/v1/server-config/reload-certs` | implemented | `serverconfig.go:ReloadCerts` | missing | — | Go implemented; Python missing (Phase 3 item). |
| `POST /api/v1/server-config/send-test-email` | implemented | `serverconfig.go:SendTestEmail` | implemented | `CanvusClient.send_test_email` | Go sends no body; spec example body is `{ "recipient-email": "..." }`. Python takes no recipient arg but spec mandates `recipient-email` — outdated (Phase 3 item). |

### Server — License

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/license` | implemented | `license.go:GetLicenseInfo` | implemented | `CanvusClient.get_license_info` | — |
| `GET /api/v1/license/request` | implemented | `license.go:GetActivationRequest` | implemented | `CanvusClient.request_offline_activation` | Go returns `request` field as string. Python adds `?key=` query param not in spec — outdated (Phase 3 item). |
| `POST /api/v1/license` | implemented | `license.go:InstallLicense` | implemented | `CanvusClient.install_offline_license` | Go body uses `{"key": ...}`; spec expects `{"license-data": ...}`. Python sends `{license: ...}`, spec expects `{license-data: ...}` — both have field-name drift. |
| `POST /api/v1/license/activate` | implemented | `license.go:ActivateLicense` | missing | — | Go: body uses `{"key": ...}`; spec says empty body. Python missing (Phase 3 item). |

### Server — Audit Log

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/audit-log` | outdated | `auditlog.go:ListAuditEvents` | implemented | `CanvusClient.get_audit_log` | Go: only supports `per_page`; spec supports `page`, `filter`, `start-time`, `end-time`, `user-id`, `action`. Response shape differs — likely drift. Python: filter keys differ (`created_after`/`created_before` vs spec `start-time`/`end-time`, `author_id` vs `user-id`) — outdated (Phase 3 item). |
| `GET /api/v1/audit-log/export-csv` | implemented | `auditlog.go:ExportAuditLog` | implemented | `CanvusClient.export_audit_log_csv` | Both return bytes. Same missing-filters limitation as JSON endpoint. |

### Server — Clients

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/clients` | implemented | `clients.go:ListClients` | implemented | `CanvusClient.list_clients` | Go also exposes `CreateClient`, `UpdateClient`, `DeleteClient` **not in spec** — these are unofficial endpoints flagged for removal. Python follows spec. |
| `GET /api/v1/clients/{client-id}` | implemented | `clients.go:GetClient` | implemented | `CanvusClient.get_client` | — |

### Server — Workspaces

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/clients/{client-id}/workspaces` | implemented | `workspaces.go:ListWorkspaces` | implemented | `CanvusClient.list_workspaces` / `get_client_workspaces` | Python exposes two methods; latter is thin alias. |
| `GET /api/v1/clients/{client-id}/workspaces/{workspace-id}` | implemented | `workspaces.go:GetWorkspace` | implemented | `CanvusClient.get_workspace` | Go treats as integer index; Python uses `workspace_index: int`. Spec says UUID — type drift (Phase 3 item). |
| `PATCH /api/v1/clients/{client-id}/workspaces/{workspace-id}` | implemented | `workspaces.go:UpdateWorkspace` | implemented | `CanvusClient.update_workspace` | — |
| `POST /api/v1/clients/{client-id}/workspaces/{workspace-id}/open-canvas` | implemented | `workspaces.go:OpenCanvasOnWorkspace` | missing | — | Go implemented; Python missing (Phase 3 item). |

### Server — Video Outputs (client-scoped)

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/clients/{client-id}/video-outputs` | implemented | `videooutputs.go:ListVideoOutputs` | implemented | `CanvusClient.list_client_video_outputs` | — |
| `GET /api/v1/clients/{client-id}/video-outputs/{output-id}` | implemented | `videooutputs.go:GetVideoOutput` | missing | — | Go implemented. Python missing (Phase 3 item). |
| `PATCH /api/v1/clients/{client-id}/video-outputs/{output-id}` | implemented | `videooutputs.go:SetVideoOutputSource` | implemented | `CanvusClient.set_video_output_source` | Go also exposes `UpdateVideoOutput` aimed at non-spec path. Python also has `update_video_output` with incorrect path (Phase 3 item). |

### Server — Video Inputs (client-scoped)

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|
| `GET /api/v1/clients/{client-id}/video-inputs` | implemented | `videoinputs.go:ListClientVideoInputs` | implemented | `CanvusClient.list_client_video_inputs` | — |
| `GET /api/v1/clients/{client-id}/video-inputs/{input-id}` | implemented | `videoinputs.go:GetClientVideoInput` | missing | — | Go implemented. Python missing (Phase 3 item). |

---

## Phase 3 work items

### Go SDK (12 items)

Ordered by impact / unblocking value. Source: `/home/jaypaulb/Projects/gh/Canvus-Go-API`.

1. **Add Table widget support** (LARGE)
   - **Files:** create `/home/jaypaulb/Projects/gh/Canvus-Go-API/canvus/tables.go` plus types in `types.go`; add `tables_test.go`.
   - **What:** Implement `ListTables`, `GetTable`, `CreateTable`, `UpdateTable`, `ListTableCells`, `DeleteTable`. Add dispatch case to `widgets.go:CreateWidget`/`UpdateWidget`/`DeleteWidget`.
   - **Why:** All 6 spec rows for `/canvases/{id}/tables*` are missing. Per changelog §4, omit `column_widths`/`row_heights`. Per changelog §5, document `grid_size` is immutable post-create and silently ignored on PATCH (optionally filter it out client-side with a warning).
   - **Complexity:** large.

2. **Add `CloneWidget` helper using standard create endpoints** (MODERATE)
   - **Files:** `/home/jaypaulb/Projects/gh/Canvus-Go-API/canvus/widgets.go`; possibly extend each type-specific create method.
   - **What:** New method `CloneWidget(ctx, destCanvasID, sourceCanvasID, sourceWidgetID, widgetType string, optLocation *Location) (*Widget, error)`. Internally POST to the matching `canvases/{destCanvasID}/{widgetType}` with body `{source_canvas_id, source_widget_id, location?}`. Also expose `source_canvas_id`/`source_widget_id` as optional fields on each type-specific create (notes, images, videos, pdfs, browsers, anchors, tables).
   - **Why:** Changelog §1. The existing `MoveWidget`/`CopyWidget` methods hit a non-spec path (`widgets/{id}/move|copy`) and don't satisfy the documented clone semantics; they should be removed once the new helper lands.
   - **Complexity:** moderate.

3. **Add IP Video and RDP Connection widget support — read/patch/delete only** (MODERATE)
   - **Files:** create `ipvideos.go` and `rdpconnections.go`, plus types in `types.go`.
   - **What:** Implement `List*`, `Get*`, `Update*`, `Delete*` for both widget types. **Do NOT add Create methods.** If a caller invokes `CreateWidget` with `widget_type: "ip_video"` or `"rdp_connection"`, return an error: "IP Video and RDP Connection widgets can only be created from the Canvus desktop client."
   - **Why:** Changelog §2. Field-naming for RDP (hyphen vs underscore) is unresolved per changelog §3 — verify against live API before finalizing struct JSON tags.
   - **Complexity:** moderate.

4. **Remove non-spec `MoveWidget` / `CopyWidget` and non-spec client CRUD** (TRIVIAL)
   - **Files:** `widgets.go` (remove `MoveWidget`, `CopyWidget`, `PinWidget`, `UnpinWidget` — these all hit `widgets/{id}/...` paths not present in the spec); `clients.go` (remove `CreateClient`, `UpdateClient`, `DeleteClient`).
   - **What:** Delete the methods; replace with `CloneWidget` (item 2) where applicable. Update callers/tests.
   - **Why:** These endpoints are not in the spec and risk runtime errors against current servers. Chesterton's-fence note: investigate first — they may be older endpoints that were deprecated and not in the current spec. Verify before removing.
   - **Complexity:** trivial (verification first; deletion is mechanical).

5. **Fix color-presets path drift** (TRIVIAL)
   - **Files:** `colorpresets.go`.
   - **What:** Change path from `canvases/%s/colorpresets` to `canvases/%s/color-presets` (3 occurrences). Verify against live server first — it's possible the server accepts both, but the spec is canonical. Also evaluate whether the per-name CRUD methods (`GetColorPreset`/`CreateColorPreset`/`UpdateColorPreset`/`DeleteColorPreset` with `/{name}` suffix) correspond to a real endpoint — none are listed in the spec.
   - **Why:** Spec §canvases.md lines 379, 409 show `color-presets`. SDK currently uses `colorpresets`.
   - **Complexity:** trivial.

6. **Fix Audit Log query params and response shape** (MODERATE)
   - **Files:** `auditlog.go`.
   - **What:** Extend `AuditLogOptions` to include `Page`, `Filter`, `StartTime`, `EndTime`, `UserID`, `Action`. Change `ListAuditEvents` to return `(events []AuditEvent, total int, page int, perPage int, err error)` or a wrapper struct, matching the documented `{events, total-count, page, per-page}` envelope. Mirror filters on `ExportAuditLog`.
   - **Why:** Spec server.md §Audit Log lists 6 filter fields and a pagination envelope; SDK only sends `per_page` and decodes a flat array.
   - **Complexity:** moderate (signature change is a breaking API change for SDK consumers).

7. **Add GET handler for uploads-folder listing** (TRIVIAL)
   - **Files:** `uploads.go`.
   - **What:** Add `ListUploads(ctx, canvasID string) ([]UploadItem, error)` that GETs `canvases/{id}/uploads-folder`. Define `UploadItem` in `types.go`.
   - **Why:** Spec widgets.md line 1248 lists `GET /uploads-folder`; SDK only POSTs.
   - **Complexity:** trivial.

8. **Add `PATCH /canvas-folders/{id}/move` variant** (TRIVIAL)
   - **Files:** `folders.go`.
   - **What:** Add `MoveFolderPatch` mirroring `MoveFolder` but using PATCH; or change `MoveFolder` signature to accept `method string`. Low priority — POST variant works.
   - **Why:** Spec folders.md line 217 lists PATCH variant.
   - **Complexity:** trivial.

9. **Verify and align password/login/license body field names** (MODERATE)
   - **Files:** `session.go:Login`, `users.go:SetUserPassword`, `license.go:InstallLicense`, `license.go:ActivateLicense`, `serverconfig.go:SendTestEmail`.
   - **What:** Verify each against the live server, then either update the spec docs (if SDK is right) or update the SDK (if the spec is right). Specific cases:
     - Login uses `username` key; spec doc shows `email`.
     - SetUserPassword body uses `{"password": ...}`; spec shows `{"old-password", "new-password"}` for self-change.
     - InstallLicense body uses `{"key": ...}`; spec shows `{"license-data": "..."}`.
     - ActivateLicense body uses `{"key": ...}`; spec shows empty body.
     - SendTestEmail sends no body; spec shows `{"recipient-email": "..."}`.
   - **Why:** These are response/request shape drifts. Some may be doc-extraction errors (the spec was synthesized from mt-restapi-client source, which may use different examples than the runtime).
   - **Complexity:** moderate (mostly verification; mechanical fixes after).

10. **Verify server-config response shape** (MODERATE)
    - **Files:** `serverconfig.go`.
    - **What:** Either rewrite `ServerConfig`/`AuthenticationConfig`/etc. as a flat `[]ConfigElement{Key, Value, Type}` matching the spec doc, or keep the nested struct if the live server actually returns nested JSON.
    - **Why:** Spec server.md §Server Configuration shows flat element array; SDK decodes into nested objects. Verification needed.
    - **Complexity:** moderate.

11. **Verify workspace ID type** (TRIVIAL)
    - **Files:** `workspaces.go`, related types.
    - **What:** Spec server.md §Client Workspaces shows `workspace-id: uuid`. SDK uses integer `index` via `WorkspaceSelector`. Confirm whether the live API uses integer index (current SDK behaviour) or UUID (per spec doc). If UUID, refactor to accept either via selector.
    - **Why:** Possible spec error or SDK error.
    - **Complexity:** trivial (verification + at most a string/int swap).

12. **Investigate non-spec `canvases/{id}/video-outputs/{id}` PATCH** (TRIVIAL)
    - **Files:** `videooutputs.go:UpdateVideoOutput`.
    - **What:** Verify whether `PATCH /canvases/{id}/video-outputs/{id}` exists on the live server. If not, delete the method.
    - **Why:** Spec only documents the client-scoped video-outputs endpoints.
    - **Complexity:** trivial.

### Python SDK (23 items)

Ordered by impact / unblocking value. Source: `/home/jaypaulb/Projects/gh/CanvusPythonAPI`.

1. **Add full Table widget support** (LARGE)
   - **What:** Add the following methods on `CanvusClient`: `list_tables`, `get_table`, `create_table`, `update_table`, `delete_table`, `list_table_cells`.
   - **Why:** Six spec endpoints under `endpoints/widgets.md#tables` have zero coverage.
   - **Type model:** Add a `Table` model in `canvus_api/models.py` covering only the fields the server actually returns: `title`, `grid_size`. Per changelog §3, do **not** expose `column_widths` or `row_heights` (server does not serialize them).
   - **PATCH behavior note:** Per changelog §5, `grid_size` is silently ignored on PATCH. Add docstring warning in `update_table` and a runtime warning (warnings.warn) when payload includes `grid_size`.
   - **Complexity:** large.

2. **Add `CloneWidget` helper per changelog §1** (MODERATE)
   - **File:** `canvus_api/client.py` (likely add to `canvus_api/widget_operations.py` to keep `client.py` size in check).
   - **What:** Add `CanvusClient.clone_widget(dest_canvas_id, source_canvas_id, source_widget_id, widget_type, location=None)`. Internally calls `POST /canvases/{dest_canvas_id}/{widget_type_path}` with `{source_canvas_id, source_widget_id, location?}` body.
   - **Why:** Changelog §1 — cross-canvas cloning is implemented in the server via standard create endpoints, but the SDK exposes no convenient wrapper. The dead `/widgets/clone` endpoint must not be exposed.
   - **Acceptance:** Supports widget types `notes`, `images`, `videos`, `pdfs`, `browsers`, `anchors`, `tables`. For asset-based types, verify source asset is available before erroring out.
   - **Complexity:** moderate.

3. **Add full RDP Connections widget support — GET/PATCH/DELETE only** (MODERATE)
   - **What:** Add `list_rdp_connections`, `get_rdp_connection`, `update_rdp_connection`, `delete_rdp_connection`. **Do NOT** add `create_rdp_connection` per changelog §2.
   - **Field naming:** Per changelog §3 (unresolved), verify hyphens-vs-underscores on `host-id` / `content-id` / `connection-name` / `host-site` against live API before defining the model. Add field naming reference in docstrings.
   - **Why:** Five spec endpoints with zero coverage; one is a "do not implement".
   - **Complexity:** moderate.

4. **Add full IP Video widget support — GET/PATCH/DELETE only** (MODERATE)
   - **What:** Add `list_ip_videos`, `get_ip_video`, `update_ip_video`, `delete_ip_video`. **Do NOT** add `create_ip_video` per changelog §2.
   - **Why:** Four spec endpoints with zero coverage; one is a "do not implement".
   - **Complexity:** moderate.

5. **Add single-item video-input GET + PATCH on canvas** (TRIVIAL)
   - **What:** Add `get_canvas_video_input(canvas_id, widget_id)` (GET `/canvases/{id}/video-inputs/{wid}`) and `update_video_input(canvas_id, widget_id, payload)` (PATCH).
   - **Why:** Spec lists 5 video-input endpoints under canvas widgets; SDK only has list/create/delete.
   - **Complexity:** trivial.

6. **Add `list_uploads_folder`** (TRIVIAL)
   - **What:** Add `CanvusClient.list_uploads_folder(canvas_id)` calling `GET /canvases/{id}/uploads-folder`. POST is already supported (split between `upload_note` and `upload_file`).
   - **Why:** Spec defines GET; SDK only implements POST.
   - **Complexity:** trivial.

7. **Add `change_email` user method** (TRIVIAL)
   - **What:** Add `CanvusClient.change_email(user_id, new_email)` calling `POST /users/{uid}/change-email` with `{new-email}` body.
   - **Why:** Missing endpoint in Users group.
   - **Complexity:** trivial.

8. **Add admin `force_reset_password` user method** (TRIVIAL)
   - **What:** Add `CanvusClient.force_reset_password(user_id)` calling `POST /users/{uid}/reset-password` with empty body.
   - **Why:** Admin-force-reset endpoint missing; do not conflate with self-reset password flow.
   - **Complexity:** trivial.

9. **Add `update_group`** (TRIVIAL)
   - **What:** Add `CanvusClient.update_group(group_id, payload)` calling `PATCH /groups/{gid}`.
   - **Why:** Spec lists PATCH on groups; no SDK method exists.
   - **Complexity:** trivial.

10. **Add `reload_certs` server method** (TRIVIAL)
    - **What:** Add `CanvusClient.reload_certs()` calling `POST /server-config/reload-certs` with empty body.
    - **Why:** Spec defines endpoint; SDK lacks it.
    - **Complexity:** trivial.

11. **Add `activate_license` (online activation) method** (TRIVIAL)
    - **What:** Add `CanvusClient.activate_license()` calling `POST /license/activate` with empty body.
    - **Why:** Online activation endpoint missing.
    - **Complexity:** trivial.

12. **Add `open_canvas_in_workspace` method** (TRIVIAL)
    - **What:** Add `CanvusClient.open_canvas_in_workspace(client_id, workspace_id, canvas_id)` calling `POST /clients/{cid}/workspaces/{wid}/open-canvas` with `{canvas-id}` body.
    - **Why:** Missing endpoint; needed to drive multi-display setups.
    - **Complexity:** trivial.

13. **Add single-item client video-output/video-input GET** (TRIVIAL)
    - **What:** Add `get_client_video_output(client_id, output_id)` and `get_client_video_input(client_id, input_id)`.
    - **Why:** Spec defines per-item GETs; SDK only has list.
    - **Complexity:** trivial.

14. **Fix `update_video_output` URL path** (TRIVIAL)
    - **What:** `CanvusClient.update_video_output` currently issues PATCH against `canvases/{canvas_id}/video-outputs/{output_id}` — that path does not exist in the spec. Either remove the method (use `set_video_output_source` for `clients/{cid}/video-outputs/{oid}`) or repurpose it to call the correct path.
    - **Why:** Method targets a non-existent endpoint. Will fail with 404 in production. Spec only defines `PATCH /clients/{cid}/video-outputs/{oid}`.
    - **Complexity:** trivial.

15. **Fix `send_test_email` to accept recipient** (TRIVIAL)
    - **What:** Change signature to `send_test_email(recipient_email: str)` and include `{recipient-email}` in body. Currently sends empty body.
    - **Why:** Spec mandates `recipient-email` field; current call will likely 400.
    - **Complexity:** trivial.

16. **Fix `install_offline_license` body field name** (TRIVIAL)
    - **What:** Change body from `{license: license_data}` to `{"license-data": license_data}`.
    - **Why:** Spec requires `license-data` (hyphenated) field name.
    - **Complexity:** trivial.

17. **Fix `request_offline_activation` query param** (TRIVIAL)
    - **What:** Remove the `?key=` query param. Spec defines no query parameters on `GET /license/request`.
    - **Why:** Unsupported parameter; may cause server to reject or be silently ignored.
    - **Complexity:** trivial.

18. **Realign `get_audit_log` filter keys with spec** (TRIVIAL)
    - **What:** Translate SDK filter dict keys to spec keys when building the request:
      - `created_after` → `start-time`
      - `created_before` → `end-time`
      - `author_id` → `user-id`
      - `target_type`, `target_id` → no spec equivalent; remove or map to `filter`/`action`.
      - Add spec params: `page`, `per-page`, `filter`, `action`.
    - **Why:** Spec uses hyphenated time keys; SDK uses snake_case keys with different names.
    - **Complexity:** trivial.

19. **Realign `create_token` signature with spec** (MODERATE)
    - **What:** Replace `create_token(user_id, description)` with `create_token(user_id, name, expires=None, scopes=None)`. Build body `{name, expires?, scopes?}`.
    - **Why:** Spec body shape is `{name, expires, scopes}`. The current `description` param does not match any spec field.
    - **Complexity:** moderate.

20. **Realign `login_saml` signature with spec** (TRIVIAL)
    - **What:** Add required args `in_response_to`, `response_xml`, optional `remember`. Build body matching spec.
    - **Why:** Current `login_saml(self)` sends no body — server will 400.
    - **Complexity:** trivial.

21. **Standardize user_id type on `str` (UUID) across all user/token methods** (MODERATE)
    - **What:** Change `user_id: int` to `user_id: str` on `list_tokens`, `get_token`, `create_token`, `update_token`, `delete_token`, `get_user`, `delete_user`, `approve_user`, `block_user`, `unblock_user`, `update_user`. Also `workspace_index: int` → `workspace_id: str` on `get_workspace`, `update_workspace`, and subscribe methods.
    - **Why:** Spec consistently uses UUID strings for `user-id` and `workspace-id`. The SDK's `int` type forces incorrect coercion (probably works against current server only because of duck-typed path interpolation).
    - **Complexity:** moderate (touches many methods + tests).

22. **Add typed `Table`, `RDPConnection`, `IPVideo` models** (MODERATE)
    - **File:** `canvus_api/models.py`
    - **What:** Add Pydantic/dataclass models for the three widget types being added (items 1, 3, 4). Update widget polymorphism in `list_widgets` to dispatch on these `widget-type` values.
    - **Why:** Without typed models, items 1/3/4 leak `Dict[str, Any]` everywhere.
    - **Complexity:** moderate.

23. **Strip `widgets/clone` from any docs/examples** (TRIVIAL)
    - **File:** `canvus_api/client.py`, `README.md`, `EXAMPLES.md`, `LLM_DEV_GUIDE.md`.
    - **What:** Search for any reference to `/widgets/clone` and replace with the new `clone_widget()` helper from item 2. Add note that endpoint is deprecated.
    - **Why:** Changelog §1 explicitly directs removal.
    - **Complexity:** trivial.

### TypeScript SDK

All endpoints from the spec must be implemented from scratch. The Phase 3.3 TypeScript SDK agent will use the full endpoint list (see `endpoints/` directory) as its work item. Estimated scope: 150 endpoints across 7 resource groups, ~6,000 lines of typed client code targeting the conventions in `docs/conventions/typescript.md`.

---

## Merge Coverage Notes

- **Endpoint ordering:** Per-endpoint tables follow Go SDK ordering (150 endpoints authoritative; Python audit grouped some widget subtypes, reducing its count to 136).
- **Status divergence:** Go and Python SDKs have divergent coverage on specific endpoints:
  - Canvas permissions POST: Go `implemented`, Python `outdated` (signature drift).
  - Canvases color-presets GET/PATCH: Go `outdated` (path drift), Python `implemented` (correct path).
  - Video-input GET (canvas): Go `implemented`, Python `missing`.
  - Video-input PATCH (canvas): Go `implemented`, Python `missing`.
  - Change-email (users): Go `implemented`, Python `missing`.
  - Force-reset-password (users): Go `implemented`, Python `missing`.
  - Update-group: Go `implemented`, Python `missing`.
  - Folders PATCH /move: Go `missing` (POST only), Python `implemented` (POST accepted).
  - Server config reload-certs: Go `implemented`, Python `missing`.
  - Server license activate (online): Go `implemented`, Python `missing`.
  - Audit log: Go `outdated` (query params/shape), Python `implemented` (filter key drift).
  - Open canvas in workspace: Go `implemented`, Python `missing`.
  - Video-output GET (client): Go `implemented`, Python `missing`.
  - Video-input GET (client): Go `implemented`, Python `missing`.
- **Grouped widget endpoints (Python):** Python audit grouped canvas-scoped video-input and other widget subtypes differently than Go's per-type enumeration. Where Python omitted individual endpoints (e.g., single-widget GET), rows are marked `missing` to match Go's canonical breakdown.
- **Non-spec endpoints:** Go SDK exposes non-spec methods (`CreateClient`, `UpdateClient`, `DeleteClient`, non-spec `canvases/{id}/video-outputs/{id}` PATCH). Phase 3 work items flag these for removal.
- **Clone widget:** Both SDKs must replace non-spec MoveWidget/CopyWidget paths and missing CloneWidget helpers with standard create endpoints + `source_canvas_id`/`source_widget_id` (changelog §1).
