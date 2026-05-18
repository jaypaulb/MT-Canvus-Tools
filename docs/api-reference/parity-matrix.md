# SDK Parity Matrix

**Audit date:** 2026-05-18
**Sources audited:**
- Go:   `/home/jaypaulb/Projects/gh/MT-Canvus-Tools/go/sdk/canvus/` (40 files)
- Py:   `/home/jaypaulb/Projects/gh/MT-Canvus-Tools/python/sdk/src/canvus_sdk/`
- TS:   `/home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript/sdk/src/`
- Legacy (port source): `/home/jaypaulb/Projects/gh/CanvusPythonAPI/canvus_api/{geometry,search,filters,widget_operations,export}.py`

This matrix covers public surface only. REST endpoint coverage was verified
in Phase 3.v (see `coverage-verification.md`); the focus here is **non-REST
helpers, ergonomics, and convenience methods**.

Path/line citations use the same root paths as `go/sdk/canvus/`,
`python/sdk/src/canvus_sdk/`, `typescript/sdk/src/`.

---

## 1. Method-level parity

### 1.1 Core SDK methods (resource-grouped)

Legend: ✅ = present (with file:line); ❌ = absent; ⚠ = present with caveat.

#### 1.1.1 Canvases

| Method (canonical) | Endpoint | Go | Python | TS |
|---|---|---|---|---|
| List canvases | `GET /canvases` | ✅ `canvases.go:11 ListCanvases` | ✅ `resources/canvases.py:31 list` | ✅ `resources/canvases.ts:28 list` |
| Get canvas | `GET /canvases/{id}` | ✅ `canvases.go:23 GetCanvas` | ✅ `resources/canvases.py:36 get` | ✅ `resources/canvases.ts:38 get` |
| Create canvas | `POST /canvases` | ✅ `canvases.go:33 CreateCanvas` | ✅ `resources/canvases.py:41 create` | ✅ `resources/canvases.ts:48 create` |
| Update canvas | `PATCH /canvases/{id}` | ✅ `canvases.go:43 UpdateCanvas` | ✅ `resources/canvases.py:46 update` | ✅ `resources/canvases.ts:53 update` |
| Delete canvas | `DELETE /canvases/{id}` | ✅ `canvases.go:52 DeleteCanvas` | ✅ `resources/canvases.py:53 delete` | ✅ `resources/canvases.ts:58 delete` |
| Get canvas preview | `GET /canvases/{id}/preview` | ✅ `canvases.go:57 GetCanvasPreview` | ✅ `resources/canvases.py:85 get_preview` | ✅ `resources/canvases.ts:146 getPreview` |
| Move canvas | `POST /canvases/{id}/move` | ✅ `canvases.go:76 MoveCanvas` | ✅ `resources/canvases.py:57 move` | ✅ `resources/canvases.ts:63 move` |
| Copy canvas | `POST /canvases/{id}/copy` | ✅ `canvases.go:85 CopyCanvas` | ✅ `resources/canvases.py:66 copy` | ✅ `resources/canvases.ts:68 copy` |
| Save demo state | `POST /canvases/{id}/save` | ✅ `canvases.go:71 SaveDemoState` | ✅ `resources/canvases.py:75 save_demo_state` | ✅ `resources/canvases.ts:73 save` |
| Restore demo state | `POST /canvases/{id}/restore` | ✅ `canvases.go:66 RestoreDemoCanvas` | ✅ `resources/canvases.py:80 restore_demo_state` | ✅ `resources/canvases.ts:78 restore` |
| Trash canvas (convenience) | client-side: PATCH `folder_id` → user's trash | ✅ `canvases.go:95 TrashCanvas` (requires prior Login for userID) | ❌ | ❌ |
| Get canvas permissions | `GET /canvases/{id}/permissions` | ✅ `canvases.go:104 GetCanvasPermissions` | ✅ `resources/canvases.py:152 get_permissions` | ✅ `resources/canvases.ts:152 getPermissions` |
| Set canvas permissions | `POST /canvases/{id}/permissions` | ✅ `canvases.go:113 SetCanvasPermissions` | ✅ `resources/canvases.py:159 set_permissions` | ✅ `resources/canvases.ts:172 setPermissions` |
| Get canvas background | `GET /canvases/{id}/background` | ✅ `backgrounds.go:10 GetCanvasBackground` | ✅ `resources/canvases.py:93 get_background` | ✅ `resources/canvases.ts:83 getBackground` |
| Update canvas background (JSON) | `PATCH /canvases/{id}/background` | ✅ `backgrounds.go:19 PatchCanvasBackground` | ✅ `resources/canvases.py:98 set_background` | ✅ `resources/canvases.ts:88 updateBackground` |
| Upload canvas background image | `POST /canvases/{id}/background` (multipart) | ✅ `backgrounds.go:28 PostCanvasBackground` | ✅ `resources/canvases.py:109 upload_background_image` | ✅ `resources/canvases.ts:105 uploadBackground` |
| Get color presets (raw) | `GET /canvases/{id}/color-presets` | ✅ `colorpresets.go:18 GetColorPresets` | ✅ `resources/canvases.py:127 get_color_presets` | ✅ `resources/canvases.ts:121 getColorPresets` |
| Patch color presets | `PATCH /canvases/{id}/color-presets` | ✅ `colorpresets.go:28 PatchColorPresets` | ✅ `resources/canvases.py:137 update_color_presets` | ✅ `resources/canvases.ts:129 updateColorPresets` |
| List color presets (per-preset enum) | client-side decomposition | ✅ `colorpresets.go:38 ListColorPresets` | ❌ | ❌ |
| Get one color preset (by name) | client-side decomposition | ✅ `colorpresets.go:48 GetColorPreset` | ❌ | ❌ |
| Create one color preset | client-side decomposition | ✅ `colorpresets.go:57 CreateColorPreset` | ❌ | ❌ |
| Update one color preset | client-side decomposition | ✅ `colorpresets.go:66 UpdateColorPreset` | ❌ | ❌ |
| Delete one color preset | client-side decomposition | ✅ `colorpresets.go:75 DeleteColorPreset` | ❌ | ❌ |

#### 1.1.2 Folders

| Method | Endpoint | Go | Python | TS |
|---|---|---|---|---|
| List folders | `GET /canvas-folders` | ✅ `folders.go:58 ListFolders` | ✅ `resources/canvases.py:186 FoldersResource.list` | ✅ `resources/folders.ts:24 list` |
| Get folder | `GET /canvas-folders/{id}` | ✅ `folders.go:67 GetFolder` | ✅ `resources/canvases.py:191 get` | ✅ `resources/folders.ts:34 get` |
| Create folder | `POST /canvas-folders` | ✅ `folders.go:77 CreateFolder` | ✅ `resources/canvases.py:196 create` | ✅ `resources/folders.ts:44 create` |
| Update/rename folder | `PATCH /canvas-folders/{id}` | ✅ `folders.go:86 RenameFolder` | ✅ `resources/canvases.py:201 update` | ✅ `resources/folders.ts:49 update` |
| Move folder (POST) | `POST /canvas-folders/{id}/move` | ✅ `folders.go:95 MoveFolder` | ✅ `resources/canvases.py:218 move` | ✅ `resources/folders.ts:64 move` |
| Move folder (PATCH) | `PATCH /canvas-folders/{id}/move` | ✅ `folders.go:101 MoveFolderPatch` | ❌ (Python uses POST only) | ✅ `resources/folders.ts:69 movePatch` |
| Copy folder (POST) | `POST /canvas-folders/{id}/copy` | ✅ `folders.go:115 CopyFolder` | ✅ `resources/canvases.py:241 copy` | ✅ `resources/folders.ts:74 copy` |
| Copy folder (PATCH) | `PATCH /canvas-folders/{id}/copy` | ✅ `folders.go:121 CopyFolderPatch` | ❌ | ✅ `resources/folders.ts:79 copyPatch` |
| Trash folder | client-side: PATCH `parent_id` → user's trash | ✅ `folders.go:135 TrashFolder` (requires Login) | ❌ | ❌ |
| Delete folder | `DELETE /canvas-folders/{id}` | ✅ `folders.go:144 DeleteFolder` | ✅ `resources/canvases.py:208 delete` | ✅ `resources/folders.ts:54 delete` |
| Delete folder children | `DELETE /canvas-folders/{id}/children` | ✅ `folders.go:149 DeleteFolderContents` | ✅ `resources/canvases.py:212 delete_children` | ✅ `resources/folders.ts:59 deleteChildren` |
| Get folder permissions | `GET /canvas-folders/{id}/permissions` | ✅ `folders.go:154 GetFolderPermissions` | ✅ `resources/canvases.py:257 get_permissions` | ✅ `resources/folders.ts:84 getPermissions` |
| Set folder permissions | `POST /canvas-folders/{id}/permissions` | ✅ `folders.go:163 SetFolderPermissions` | ✅ `resources/canvases.py:264 set_permissions` | ✅ `resources/folders.ts:104 setPermissions` |

#### 1.1.3 Widgets — generic

| Method | Endpoint | Go | Python | TS |
|---|---|---|---|---|
| List widgets (mixed) | `GET /canvases/{id}/widgets` | ✅ `widgets.go:23 ListWidgets` (with `includeAnnotations`, Filter) | ✅ `resources/widgets.py:521 WidgetsResource.list` | ✅ `resources/widgets.ts:112 list` |
| Get widget (generic) | `GET /canvases/{id}/widgets/{wid}` | ✅ `widgets.go:39 GetWidget` | ✅ `resources/widgets.py:538 get` | ✅ `resources/widgets.ts:122 get` |
| Create widget (generic, dispatch on widget_type) | `POST /canvases/{id}/{type-plural}` | ✅ `widgets.go:53 CreateWidget` (multipart-aware, rejects IPVideo/RDP) | ❌ (only per-type create) | ❌ (only per-type create) |
| Update widget (generic, dispatch on widget_type) | `PATCH /canvases/{id}/{type-plural}/{wid}` | ✅ `widgets.go:120 UpdateWidget` | ❌ | ❌ |
| Delete widget (generic, dispatch on widget_type) | `DELETE /canvases/{id}/{type-plural}/{wid}` | ✅ `widgets.go:193 DeleteWidget` | ❌ | ❌ |
| Patch parent_id (re-parent widget) | helper around PATCH | ✅ `widgets.go:224 PatchParentID` | ❌ | ❌ |
| Clone widget (cross-canvas, single helper) | `POST /canvases/{dest}/{type-plural}` body=`source_*` | ✅ `widgets.go:244 CloneWidget` | ✅ `resources/widgets.py:545 clone` | ✅ `resources/widgets.ts:149 clone` |

#### 1.1.4 Widgets — per type CRUD

For each (Note, Image, Video, PDF, Browser, Anchor, Connector, Table,
VideoInput, IPVideo, RDPConnection) all three SDKs expose list/get/update/delete
with create where the server allows.

| Type · op | Go | Python | TS |
|---|---|---|---|
| Notes list/get/create/update/delete | ✅ `notes.go:13/23/34/43/52` | ✅ `resources/widgets.py:125-145` | ✅ `resources/widgets.ts:168 notes` |
| Images list/get/create/update/delete/download | ✅ `images.go:11/20/30/42/52/57` | ✅ `resources/widgets.py:174-198` (+ `_AssetWidgetMixin.upload @151`) | ✅ `resources/widgets.ts:191 images` |
| Videos list/get/create/update/delete/download | ✅ `videos.go:10/19/38/50/60/28` | ✅ `resources/widgets.py:201-224` | ✅ `resources/widgets.ts:230 videos` |
| PDFs list/get/create/update/delete/download | ✅ `pdfs.go:10/19/38/50/60/28` | ✅ `resources/widgets.py:228-251` | ✅ `resources/widgets.ts:269 pdfs` |
| Browsers list/get/create/update/delete | ✅ `browsers.go:10/19/28/37/46` | ✅ `resources/widgets.py:258-279` | ✅ `resources/widgets.ts:312 browsers` |
| Anchors list/get/create/update/delete | ✅ `anchors.go:10/19/28/37/46` | ✅ `resources/widgets.py:282-303` | ✅ `resources/widgets.ts:335 anchors` |
| Connectors list/get/create/update/delete | ✅ `connectors.go:10/19/32/72/81` | ✅ `resources/widgets.py:306-327` | ✅ `resources/widgets.ts:358 connectors` |
| Tables list/get/create/update/delete + cells | ✅ `tables.go:21/30/41/54/69/74` | ✅ `resources/widgets.py:333-373` | ✅ `resources/widgets.ts:392 tables` |
| VideoInputs (canvas-scope) list/get/create/update/delete | ✅ `videoinputs.go:16/26/36/45/54` | ✅ `resources/widgets.py:383-406` | ✅ `resources/widgets.ts:445 videoInputs` |
| IPVideos list/get/update/delete (no create per server) | ✅ `ipvideos.go:14/23/32/41` | ✅ `resources/widgets.py:412-435` (create raises) | ✅ `resources/widgets.ts:485 ipVideos` |
| RDPConnections list/get/update/delete (no create per server) | ✅ `rdpconnections.go:23/32/41/50` | ✅ `resources/widgets.py:449-474` (create raises) | ✅ `resources/widgets.ts:524 rdpConnections` |

#### 1.1.5 Uploads-folder

| Method | Endpoint | Go | Python | TS |
|---|---|---|---|---|
| List uploads-folder | `GET /canvases/{id}/uploads-folder` | ✅ `uploads.go:12 ListUploads` | ✅ `resources/widgets.py:608 list_uploads_folder` | ✅ `resources/widgets.ts:567 uploadsFolder.list` |
| Upload to uploads-folder (note) | `POST /canvases/{id}/uploads-folder` | ✅ `uploads.go:21 UploadNote` | ✅ `resources/widgets.py:615 upload_to_uploads_folder` | ✅ `resources/widgets.ts:584 uploadsFolder.upload` |
| Upload to uploads-folder (asset) | `POST /canvases/{id}/uploads-folder` | ✅ `uploads.go:30 UploadAsset` | ⚠ same `upload_to_uploads_folder` (one entrypoint, metadata distinguishes) | ⚠ same `uploadsFolder.upload` |

#### 1.1.6 Users

| Method | Endpoint | Go | Python | TS |
|---|---|---|---|---|
| List users | `GET /users` | ✅ `users.go:43 ListUsers` | ✅ `resources/users.py:22 list` | ✅ `resources/users.ts:35 list` |
| Get user | `GET /users/{id}` | ✅ `users.go:52 GetUser` | ✅ `resources/users.py:27 get` | ✅ `resources/users.ts:45 get` |
| Create user | `POST /users` | ✅ `users.go:61 CreateUser` | ✅ `resources/users.py:32 create` | ✅ `resources/users.ts:55 create` |
| Update user | `PATCH /users/{id}` | ✅ `users.go:70 UpdateUser` | ✅ `resources/users.py:37 update` | ✅ `resources/users.ts:60 update` |
| Delete user | `DELETE /users/{id}` | ✅ `users.go:79 DeleteUser` | ✅ `resources/users.py:44 delete` | ✅ `resources/users.ts:98 delete` |
| Block user | `POST /users/{id}/block` | ✅ `users.go:148 BlockUser` | ✅ `resources/users.py:50 block` | ✅ `resources/users.ts:74 block` |
| Unblock user | `POST /users/{id}/unblock` | ✅ `users.go:153 UnblockUser` | ✅ `resources/users.py:55 unblock` | ✅ `resources/users.ts:79 unblock` |
| Approve user | `POST /users/{id}/approve` | ✅ `users.go:158 ApproveUser` | ✅ `resources/users.py:60 approve` | ✅ `resources/users.ts:84 approve` |
| Change user email | `POST /users/{id}/email` | ✅ `users.go:129 ChangeUserEmail` | ✅ `resources/users.py:90 change_email` | ✅ `resources/users.ts:65 changeEmail` |
| Set user password (admin) | `POST /users/{id}/password` | ✅ `users.go:142 SetUserPassword` | ⚠ `resources/users.py:67 change_password` (admin & self in one) | ❌ (no admin password helper; `users/{id}/password` only available via raw transport) |
| Change user password (self) | `POST /users/{id}/password` | ❌ | ✅ `resources/users.py:67 change_password` | ✅ `resources/auth.ts:101 changePassword` |
| Force password reset | `POST /users/{id}/force_password_reset` | ✅ `users.go:163 ForcePasswordResetUser` | ✅ `resources/users.py:102 force_reset_password` | ✅ `resources/users.ts:89 forcePasswordReset` |
| Get current user (`GET /users/current`) | helper | ❌ (no `GetCurrentUser` method) | ❌ (no helper) | ❌ (no helper) |
| Register (public signup) | `POST /users/register` | ✅ `users.go:104 RegisterUser` | ✅ `resources/auth.py:110 register` | ✅ `resources/auth.ts:85 register` |
| Confirm email | `POST /users/confirm_email` | ✅ `users.go:113 ConfirmEmail` | ✅ `resources/auth.py:115 confirm_email` | ✅ `resources/auth.ts:90 confirmEmail` |
| Create reset token | `POST /users/password/create_reset_token` | ✅ `users.go:118 CreateResetToken` | ✅ `resources/auth.py:83 request_password_reset` | ✅ `resources/auth.ts:55 createResetToken` |
| Validate reset token | `POST /users/password/validate_reset_token` | ✅ `users.go:99 ValidateResetToken` | ✅ `resources/auth.py:92 validate_reset_token` | ✅ `resources/auth.ts:64 validateResetToken` |
| Reset password | `POST /users/password/reset` | ✅ `users.go:123 ResetUserPassword` | ✅ `resources/auth.py:101 reset_password` | ✅ `resources/auth.ts:74 resetPassword` |

#### 1.1.7 Groups

| Method | Endpoint | Go | Python | TS |
|---|---|---|---|---|
| List groups | `GET /groups` | ✅ `groups.go:41 ListGroups` | ✅ `resources/users.py:118 GroupsResource.list` | ✅ `resources/users.ts:105 listGroups` |
| Get group | `GET /groups/{id}` | ✅ `groups.go:50 GetGroup` | ✅ `resources/users.py:123 get` | ✅ `resources/users.ts:115 getGroup` |
| Create group | `POST /groups` | ✅ `groups.go:59 CreateGroup` | ✅ `resources/users.py:128 create` | ✅ `resources/users.ts:125 createGroup` |
| Update group | `PATCH /groups/{id}` | ✅ `groups.go:68 UpdateGroup` | ✅ `resources/users.py:133 update` | ✅ `resources/users.ts:130 updateGroup` |
| Delete group | `DELETE /groups/{id}` | ✅ `groups.go:77 DeleteGroup` | ✅ `resources/users.py:143 delete` | ✅ `resources/users.ts:135 deleteGroup` |
| Add user to group | `POST /groups/{id}/members` | ✅ `groups.go:82 AddUserToGroup` | ✅ `resources/users.py:154 add_member` | ✅ `resources/users.ts:153 addGroupMember` |
| List group members | `GET /groups/{id}/members` | ✅ `groups.go:87 ListGroupMembers` | ✅ `resources/users.py:149 list_members` | ✅ `resources/users.ts:140 listGroupMembers` |
| Remove user from group | `DELETE /groups/{id}/members/{uid}` | ✅ `groups.go:96 RemoveUserFromGroup` | ✅ `resources/users.py:163 remove_member` | ✅ `resources/users.ts:158 removeGroupMember` |

#### 1.1.8 Access tokens

| Method | Go | Python | TS |
|---|---|---|---|
| List access tokens | ✅ `accesstokens.go:32 ListAccessTokens` | ✅ `resources/auth.py:124 list_tokens` | ✅ `resources/auth.ts:108 listAccessTokens` |
| Get access token | ✅ `accesstokens.go:41 GetAccessToken` | ✅ `resources/auth.py:135 get_token` | ✅ `resources/auth.ts:128 getAccessToken` |
| Create access token | ✅ `accesstokens.go:54 CreateAccessToken` | ✅ `resources/auth.py:142 create_token` | ✅ `resources/auth.ts:154 createAccessToken` |
| Update access token | ✅ `accesstokens.go:63 UpdateAccessToken` | ✅ `resources/auth.py:177 update_token` | ✅ `resources/auth.ts:166 updateAccessToken` |
| Delete access token | ✅ `accesstokens.go:75 DeleteAccessToken` | ✅ `resources/auth.py:188 delete_token` | ✅ `resources/auth.ts:179 deleteAccessToken` |

#### 1.1.9 Auth (login/logout/SAML)

| Method | Go | Python | TS |
|---|---|---|---|
| Login (email + password) | ✅ `session.go:790 Login` (Session method, stores token) | ✅ `resources/auth.py:26 login` | ✅ `resources/auth.ts:38 login` |
| Logout | ✅ `session.go:814 Logout` | ✅ `resources/auth.py:76 logout` | ✅ `resources/auth.ts:48 logout` |
| SAML login | ✅ `users.go:94 SamlLogin` | ✅ `resources/auth.py:54 login_saml` | ✅ `resources/auth.ts:43 loginSaml` |

#### 1.1.10 Server / system

| Method | Endpoint | Go | Python | TS |
|---|---|---|---|---|
| Get server info | `GET /server-info` | ✅ `serverinfo.go:18 GetServerInfo` | ✅ `resources/server.py:36 get_info` | ✅ `resources/server.ts:36 info` |
| Get server config | `GET /server-config` | ✅ `serverconfig.go:78 GetServerConfig` | ✅ `resources/server.py:41 get_config` | ✅ `resources/server.ts:41 config` |
| Get server config (raw flat element-array) | client-side conversion | ✅ `serverconfig.go:89 GetServerConfigRaw` | ❌ | ❌ |
| Update server config | `PATCH /server-config` | ✅ `serverconfig.go:98 UpdateServerConfig` | ✅ `resources/server.py:54 update_config` | ✅ `resources/server.ts:51 updateConfig` |
| Send test email | `POST /server-config/send-test-email` | ✅ `serverconfig.go:109 SendTestEmail` | ✅ `resources/server.py:61 send_test_email` | ✅ `resources/server.ts:69 sendTestEmail` |
| Reload TLS certs | `POST /server-config/reload-certs` | ✅ `serverconfig.go:119 ReloadCerts` | ✅ `resources/server.py:74 reload_certs` | ✅ `resources/server.ts:60 reloadCerts` |
| Get license info | `GET /license` | ✅ `license.go:20 GetLicenseInfo` | ✅ `resources/server.py:83 get_license_info` | ✅ `resources/server.ts:80 license` |
| Get offline activation request | `GET /license/request` | ✅ `license.go:29 GetActivationRequest` | ✅ `resources/server.py:88 request_offline_activation` | ✅ `resources/server.ts:90 licenseRequest` |
| Install license | `POST /license` | ✅ `license.go:41 InstallLicense` | ✅ `resources/server.py:98 install_offline_license` | ✅ `resources/server.ts:95 installLicense` |
| Activate license (online) | `POST /license/activate` | ✅ `license.go:52 ActivateLicense` | ✅ `resources/server.py:110 activate_license` | ✅ `resources/server.ts:100 activateLicense` |
| List audit events | `GET /audit-log` | ✅ `auditlog.go:56 ListAuditEvents` | ✅ `resources/server.py:120 get_audit_log` | ✅ `resources/server.ts:112 auditLog` |
| Export audit log CSV | `GET /audit-log.csv` | ✅ `auditlog.go:66 ExportAuditLog` | ✅ `resources/server.py:170 export_audit_log_csv` | ✅ `resources/server.ts:119 exportAuditCsv` |

#### 1.1.11 Clients / workspaces / video outputs / inputs

| Method | Go | Python | TS |
|---|---|---|---|
| List clients | ✅ `clients.go:18 ListClients` | ✅ `resources/server.py:197 list_clients` | ✅ `resources/server.ts:130 clients` |
| Get client | ✅ `clients.go:27 GetClient` | ✅ `resources/server.py:202 get_client` | ✅ `resources/server.ts:140 client` |
| List workspaces | ✅ `workspaces.go:44 ListWorkspaces` | ✅ `resources/server.py:209 list_workspaces` | ✅ `resources/server.ts:153 workspaces` |
| Get workspace (by selector) | ✅ `workspaces.go:53 GetWorkspace` (selector helper) | ✅ `resources/server.py:214 get_workspace` (by `workspace_id`) | ✅ `resources/server.ts:173 workspace` |
| Update workspace | ✅ `workspaces.go:66 UpdateWorkspace` | ✅ `resources/server.py:226 update_workspace` | ✅ `resources/server.ts:194 updateWorkspace` |
| Toggle workspace info panel | ✅ `workspaces.go:79 ToggleWorkspaceInfoPanel` | ❌ | ❌ |
| Toggle workspace pinned | ✅ `workspaces.go:90 ToggleWorkspacePinned` | ❌ | ❌ |
| Open canvas on workspace | ✅ `workspaces.go:131 OpenCanvasOnWorkspace` (OpenCanvasOptions) | ✅ `resources/server.py:240 open_canvas_in_workspace` | ✅ `resources/server.ts:207 openCanvasInWorkspace` |
| Set workspace viewport (helper that fetches widget + computes view_rectangle) | ✅ `workspaces.go:102 SetWorkspaceViewport` (top-level func) | ❌ | ❌ |
| List client video outputs | ✅ `videooutputs.go:10 ListVideoOutputs` | ✅ `resources/server.py:260 list_client_video_outputs` | ✅ `resources/server.ts:222 videoOutputs` |
| Get client video output | ✅ `videooutputs.go:19 GetVideoOutput` | ✅ `resources/server.py:267 get_client_video_output` | ✅ `resources/server.ts:242 videoOutput` |
| Set video output source (by index) | ✅ `videooutputs.go:28 SetVideoOutputSource` | ⚠ Python sets via update keyed by `output_id` (no index helper) | ⚠ TS sets via `updateVideoOutput` (no index helper) |
| Set video output source (by ID) | ✅ `videooutputs.go:34 SetVideoOutputSourceByID` | ✅ `resources/server.py:276 set_video_output_source` | ✅ `resources/server.ts:263 updateVideoOutput` |
| List client video inputs | ✅ `videoinputs.go:59 ListClientVideoInputs` | ✅ `resources/server.py:298 list_client_video_inputs` | ✅ `resources/server.ts:276 videoInputs` |
| Get client video input | ✅ `videoinputs.go:68 GetClientVideoInput` | ✅ `resources/server.py:305 get_client_video_input` | ✅ `resources/server.ts:296 videoInput` |

#### 1.1.12 Mipmaps / assets

| Method | Go | Python | TS |
|---|---|---|---|
| Get mipmap info | ✅ `mipmaps.go:16 GetMipmapInfo` (supports `page`) | ✅ `resources/assets.py:24 get_mipmap_info` (no `page`) | ✅ `resources/assets.ts:37 mipmap` (supports `page` via `MipmapOptions`) |
| Get mipmap level (bytes) | ✅ `mipmaps.go:29 GetMipmapLevel` | ✅ `resources/assets.py:29 download_mipmap_level` (no `page`) | ✅ `resources/assets.ts:49 mipmapLevel` (supports `page`) |
| Get asset by hash | ✅ `mipmaps.go:42 GetAssetByHash` | ✅ `resources/assets.py:12 download_by_hash` | ✅ `resources/assets.ts:28 download` |

---

### 1.2 Subscribe / streaming methods (biggest known gap)

The server supports `?subscribe=true` on every read endpoint listed in
`streaming.md` §"Supported Endpoints". Below: per endpoint, which SDK exposes
a typed helper vs. raw transport.

Legend: **typed** = generator/iterator/handler that yields decoded models;
**raw** = caller must call generic transport and parse lines themselves;
❌ = no convenience exposed.

| Endpoint | Go | Python | TS |
|---|---|---|---|
| `canvases?subscribe` | raw (use `doRequestWithHeaders`) | typed via `WidgetsResource.subscribe`? **no** — canvases has no subscribe. Use raw `_http.stream_lines`. | **typed** `resources/canvases.ts:33 subscribe` |
| `canvases/{id}?subscribe` | raw | raw | **typed** `resources/canvases.ts:43 subscribeOne` |
| `canvas-folders?subscribe` | raw | raw | **typed** `resources/folders.ts:29 subscribe` |
| `canvas-folders/{id}?subscribe` | raw | raw | **typed** `resources/folders.ts:39 subscribeOne` |
| `canvases/{id}/widgets?subscribe` | raw (no helper) | **typed** `resources/widgets.py:639 WidgetsResource.subscribe` (yields raw dicts) | **typed** `resources/widgets.ts:117 subscribe` |
| `canvases/{id}/widgets/{wid}?subscribe` | ❌ | raw | **typed** `resources/widgets.ts:127 subscribeOne` |
| `canvases/{id}/notes?subscribe` (+ per-id) | ❌ | via `widgets.subscribe(widget_type="notes")` | **typed** `resources/widgets.ts:171 notes.subscribe/subscribeOne` |
| `canvases/{id}/images?subscribe` (+ per-id) | ❌ | via `widgets.subscribe(widget_type="images")` | **typed** `resources/widgets.ts:194 images.subscribe/subscribeOne` |
| `canvases/{id}/videos?subscribe` (+ per-id) | ❌ | via `widgets.subscribe(widget_type="videos")` | **typed** `resources/widgets.ts:233 videos.subscribe/subscribeOne` |
| `canvases/{id}/pdfs?subscribe` (+ per-id) | ❌ | via `widgets.subscribe(widget_type="pdfs")` | **typed** `resources/widgets.ts:272 pdfs.subscribe/subscribeOne` |
| `canvases/{id}/browsers?subscribe` (+ per-id) | ❌ | via `widgets.subscribe(widget_type="browsers")` | **typed** `resources/widgets.ts:315 browsers.subscribe/subscribeOne` |
| `canvases/{id}/anchors?subscribe` (+ per-id) | ❌ | via `widgets.subscribe(widget_type="anchors")` | **typed** `resources/widgets.ts:338 anchors.subscribe/subscribeOne` |
| `canvases/{id}/connectors?subscribe` (+ per-id) | ❌ | via `widgets.subscribe(widget_type="connectors")` | **typed** `resources/widgets.ts:361 connectors.subscribe/subscribeOne` |
| `canvases/{id}/tables?subscribe` (+ per-id, + cells) | ❌ | via `widgets.subscribe(widget_type="tables")` | **typed** `resources/widgets.ts:395 tables.subscribe/subscribeOne/subscribeCells` |
| `canvases/{id}/video-inputs?subscribe` | ❌ | via `widgets.subscribe(widget_type="video-inputs")` | **typed** `resources/widgets.ts:448 videoInputs.subscribe/subscribeOne` |
| `canvases/{id}/ip-videos?subscribe` | ❌ | via `widgets.subscribe(widget_type="ip-videos")` | **typed** `resources/widgets.ts:488 ipVideos.subscribe/subscribeOne` |
| `canvases/{id}/rdp-connections?subscribe` | ❌ | via `widgets.subscribe(widget_type="rdp-connections")` | **typed** `resources/widgets.ts:527 rdpConnections.subscribe/subscribeOne` |
| `canvases/{id}/uploads-folder?subscribe` | ❌ | ❌ | **typed** `resources/widgets.ts:570 uploadsFolder.subscribe` |
| `users?subscribe` | ❌ | ❌ | **typed** `resources/users.ts:40 subscribe` |
| `users/{id}?subscribe` | ❌ | ❌ | **typed** `resources/users.ts:50 subscribeOne` |
| `users/{id}/access-tokens?subscribe` (+ per-id) | ❌ | ❌ | **typed** `resources/auth.ts:116 subscribeAccessTokens/subscribeAccessToken` |
| `groups?subscribe` | ❌ | ❌ | **typed** `resources/users.ts:110 subscribeGroups` |
| `groups/{id}?subscribe` (+ members) | ❌ | ❌ | **typed** `resources/users.ts:120 subscribeGroup/subscribeGroupMembers` |
| `server-config?subscribe` | ❌ | ❌ | **typed** `resources/server.ts:46 subscribeConfig` |
| `license?subscribe` | ❌ | ❌ | **typed** `resources/server.ts:85 subscribeLicense` |
| `clients?subscribe` | ❌ | ❌ | **typed** `resources/server.ts:135 subscribeClients` |
| `clients/{id}?subscribe` | ❌ | ❌ | **typed** `resources/server.ts:145 subscribeClient` |
| `clients/{id}/workspaces?subscribe` (+ per-id) | ❌ | ❌ | **typed** `resources/server.ts:161 subscribeWorkspaces/subscribeWorkspace` |
| `clients/{id}/video-outputs?subscribe` (+ per-id) | ❌ | ❌ | **typed** `resources/server.ts:230 subscribeVideoOutputs/subscribeVideoOutput` |
| `clients/{id}/video-inputs?subscribe` (+ per-id) | ❌ | ❌ | **typed** `resources/server.ts:284 subscribeVideoInputs/subscribeVideoInput` |
| `canvases/{id}/permissions?subscribe` | ❌ | ❌ | **typed** `resources/canvases.ts:160 subscribePermissions` |
| `canvas-folders/{id}/permissions?subscribe` | ❌ | ❌ | **typed** `resources/folders.ts:92 subscribePermissions` |

**Summary:** TS is the parity benchmark for subscribe — **27 typed
subscribe helpers**. Python has **1** (a single multiplexed widget
`subscribe`). Go has **0**. Bringing Go and Python up to TS parity is the
single largest gap in this audit.

---

### 1.3 Constructor / Session ergonomics

| Capability | Go | Python | TS |
|---|---|---|---|
| Build session from env vars | ❌ (no `FromEnv` helper) | ✅ `client.py:98 Client.from_env` (pydantic Settings, prefix `CANVUS_`) | ✅ `config.ts:37 loadConfig` (`CANVUS_API_URL`, `CANVUS_API_KEY`, `CANVUS_TIMEOUT_MS`, `CANVUS_VERIFY_TLS`) |
| Build session from explicit opts | ✅ `session.go:282 NewSession(cfg, opts...)` | ✅ `client.py:63 Client.__init__` | ✅ `config.ts:74 buildConfig` + `session.ts:68 createSession` |
| Convenience: `(baseURL, apiKey)` two-arg ctor | ✅ `clients.go:41 NewSessionFromConfig` | ⚠ `Client(base_url, api_key)` (two required positional args) | ✅ `createSession({baseUrl, apiKey})` |
| API key option | ✅ `session.go:67 WithAPIKey` (also bypasses TLS verify by default) | ✅ `Client(api_key=...)` | ✅ `SessionOptions.apiKey` |
| Token option | ✅ `session.go:103 WithToken` | ❌ (no explicit token-only init) | ❌ |
| Default timeout override | ✅ `options.go:90 WithRequestTimeout` (default 30s) | ✅ `request_timeout_seconds=30.0` (+ separate `connect_timeout_seconds=5.0`) | ✅ `timeoutMs` (default 30000) |
| Connect timeout (separate from request) | ❌ | ✅ `connect_timeout_seconds` | ❌ |
| Custom HTTP client | ✅ `options.go:72 WithHTTPClient` | ❌ (transport is internal `httpx.AsyncClient`) | ❌ (transport is internal `fetch`) |
| Request-id injection | ❌ (only reads `request_id` from error responses) | ❌ (only reads `x-request-id` from error responses) | ❌ |
| Retry config | ✅ `WithMaxRetries`, `WithRetryWait`, `WithTokenRefreshThreshold` | ✅ `max_retries`, `retry_initial_delay_seconds`, `retry_backoff_factor` | ❌ (no retry layer; transport is single-shot) |
| Circuit breaker | ✅ `session.go:128-182` + `WithCircuitBreaker` | ❌ | ❌ |
| Token store / refresh hook | ✅ `WithTokenStore` interface | ❌ | ❌ |
| TLS verify toggle | ⚠ implicit when using `WithAPIKey` (default: insecure) | ✅ `verify_ssl=True` (default secure) | ✅ `verifyTls` (default true, NOT applied — see Notes §6) |
| User-Agent customization | ✅ `WithUserAgent` | ❌ | ❌ |
| Bearer/header logger injection | ✅ `Session.SetLogger(slog.Logger)` | ❌ (uses module `structlog` logger only) | ⚠ logger module exported but no per-session override |
| Async + sync surface | n/a (Go is sync) | ✅ `client.sync` proxy on dedicated worker loop | n/a (TS is awaitable) |
| Context manager / lifecycle | n/a (no explicit close needed) | ✅ `async with Client(...) as c` + `aclose()` | ❌ (no explicit close — relies on GC) |

---

### 1.4 Error types

| Concept | Go | Python | TS |
|---|---|---|---|
| Base error | `APIError` (struct) | `CanvusError` → `APIError` | `CanvusError` (class) |
| HTTP status carrier | `APIError.StatusCode` | `APIError.status_code` | `APIError.status` |
| Code string carrier | `APIError.Code` (legacy) + sentinels (`ErrUnauthorized`, …) | discrimination via subclass | discrimination via `kind` + subclass |
| Auth subtype | sentinel `ErrUnauthorized` | `AuthError(APIError)` | `AuthError(CanvusError)` — note: NOT subclass of `APIError`; carries `reason` enum |
| 404 subtype | sentinel `ErrNotFound` | `NotFoundError(APIError)` | `NotFoundError(APIError)` |
| 429 subtype | sentinel `ErrTooManyRequests` / `ErrRateLimited` | `RateLimitError(APIError)` (carries `retry_after`) | ❌ no dedicated class |
| 5xx subtype | sentinel `ErrInternalServer` | `ServerError(APIError)` | ❌ no dedicated class |
| Validation error | `ValidationError` + `ValidationErrors` collection | `ValidationError(CanvusError)` | `ValidationError(CanvusError)` with `issues[]` |
| Network/transport error | sentinel `ErrNetwork` | `TransportError(CanvusError)` | `NetworkError(CanvusError)` |
| Unsupported operation (IPVideo/RDP create) | ❌ (returns plain error from `widgets.go:53`) | ✅ `UnsupportedOperationError` (raised from IPVideo/RDP `create`) | ❌ (no dedicated class — type system enforces by omitting `create`) |
| `errors.Is` / `instanceof` sentinel matching | ✅ `APIError.Unwrap` returns sentinel; `APIError.Is` | n/a (use `isinstance`) | ✅ `instanceof` + `isCanvusError` type guard |
| Request-id field | ✅ `RequestID` | ✅ `request_id` | ❌ |
| Retry-after field | ❌ | ✅ (only on `RateLimitError`) | ❌ |
| Field-level validation issues array | ✅ `ValidationErrors []*ValidationError` | ⚠ `ValidationError` carries message only, no structured issues | ✅ `issues: readonly ValidationIssue[]` |

---

## 2. Helper-module parity (Phase 4b core deliverable)

Each table lists the public surface of the legacy Python module that must
be ported. **Pre-port status everywhere except where Go has a partial
implementation** — see citations below.

### 2.1 `geometry`

Legacy source: `/home/jaypaulb/Projects/gh/CanvusPythonAPI/canvus_api/geometry.py`

| Function / type | Legacy line | Go | Python (new SDK) | TS |
|---|---|---|---|---|
| `Point` dataclass | `geometry.py:14` | ⚠ `types.go` has `Point` struct | ❌ | ❌ |
| `Size` dataclass | `geometry.py:27` | ⚠ `types.go` has `Size` struct | ✅ `models/common.py:15` | ✅ `types/common.ts:22` |
| `Rectangle` dataclass | `geometry.py:42` | ⚠ `types.go` has `Rectangle` | ❌ | ❌ |
| `contains(outer, inner)` | `geometry.py:93` | ✅ `geometry.go:4 Contains` | ❌ | ❌ |
| `touches(a, b)` | `geometry.py:110` | ✅ `geometry.go:11 Touches` | ❌ | ❌ |
| `intersects(a, b)` | `geometry.py:127` | ❌ | ❌ | ❌ |
| `get_intersection(a, b)` | `geometry.py:144` | ❌ | ❌ | ❌ |
| `get_union(a, b)` | `geometry.py:166` | ❌ | ❌ | ❌ |
| `widget_bounding_box(w)` | `geometry.py:185` | ✅ `geometry.go:17 WidgetBoundingBox` | ❌ | ❌ |
| `widget_contains(a, b)` | `geometry.py:253` | ✅ `geometry.go:32 WidgetContains` | ❌ | ❌ |
| `widgets_touch(a, b)` | `geometry.py:270` | ✅ `geometry.go:37 WidgetsTouch` | ❌ | ❌ |
| `widgets_intersect(a, b)` | `geometry.py:287` | ❌ | ❌ | ❌ |
| `get_widget_intersection(a, b)` | `geometry.py:304` | ❌ | ❌ | ❌ |
| `get_widget_union(a, b)` | `geometry.py:321` | ❌ | ❌ | ❌ |
| `distance_between_widgets(a, b)` | `geometry.py:338` | ❌ | ❌ | ❌ |
| `find_widgets_in_area(widgets, area, op)` | `geometry.py:365` | ❌ (only cross-canvas variant exists at `widgets.go:277`) | ❌ | ❌ |
| `find_widgets_containing_point(widgets, p)` | `geometry.py:385` | ❌ | ❌ | ❌ |
| `get_canvas_bounds(widgets)` | `geometry.py:406` | ❌ | ❌ | ❌ |

### 2.2 `filters`

Legacy source: `/home/jaypaulb/Projects/gh/CanvusPythonAPI/canvus_api/filters.py`

| Function / type | Legacy line | Go | Python (new SDK) | TS |
|---|---|---|---|---|
| `FilterOperator` enum | `filters.py:11` | ⚠ implicit via string ops in `types.go Filter` | ❌ | ❌ |
| `Filter` class (rich, with spatial/wildcard) | `filters.py:33` | ⚠ `types.go:59 Filter` exists but much smaller surface (criteria-map + wildcard `Match`) | ❌ | ❌ |
| `Filter.add_condition(field, op, value)` | `filters.py:50` | ❌ (Go uses literal map) | ❌ | ❌ |
| `Filter.add_spatial_condition(op, area)` | `filters.py:72` | ❌ | ❌ | ❌ |
| `Filter.add_wildcard_condition(field, pattern)` | `filters.py:90` | ❌ | ❌ | ❌ |
| `Filter.matches(item)` | `filters.py:108` | ✅ `types.go:59 (*Filter).Match` | ❌ | ❌ |
| `Filter.to_dict()` / `from_dict()` | `filters.py:227/234` | ❌ | ❌ | ❌ |
| `create_filter()` | `filters.py:239` | ❌ | ❌ | ❌ |
| `create_spatial_filter(area, op)` | `filters.py:249` | ❌ | ❌ | ❌ |
| `create_widget_type_filter(types)` | `filters.py:263` | ❌ | ❌ | ❌ |
| `create_text_filter(text, fields)` | `filters.py:279` | ❌ | ❌ | ❌ |
| `create_wildcard_filter(pattern, field)` | `filters.py:301` | ❌ | ❌ | ❌ |
| `combine_filters(*filters, operator)` | `filters.py:315` | ❌ | ❌ | ❌ |

### 2.3 `widget_operations`

Legacy source: `/home/jaypaulb/Projects/gh/CanvusPythonAPI/canvus_api/widget_operations.py`

| Function / type | Legacy line | Go | Python (new SDK) | TS |
|---|---|---|---|---|
| `SpatialTolerance` | `widget_operations.py:20` | ❌ | ❌ | ❌ |
| `WidgetZoneManager.create_zone_from_widgets` | `widget_operations.py:40` | ❌ | ❌ | ❌ |
| `WidgetZoneManager.widgets_in_zone` | `widget_operations.py:83` | ⚠ `widgets.go:306 WidgetsContainId` (single-id variant only) | ❌ | ❌ |
| `WidgetZoneManager.widgets_touching_zone` | `widget_operations.py:116` | ❌ | ❌ | ❌ |
| `BatchWidgetOperations.move_widgets` | `widget_operations.py:202` | ⚠ `batch.go` has `BatchOperationMove` but at canvas-level, not widget-level | ❌ | ❌ |
| `BatchWidgetOperations.resize_widgets` | `widget_operations.py:267` | ❌ | ❌ | ❌ |
| `BatchWidgetOperations.widgets_contain_id` | `widget_operations.py:316` | ✅ `widgets.go:306 WidgetsContainId` (returns WidgetZone) | ❌ | ❌ |
| `BatchWidgetOperations.widgets_touch_id` | `widget_operations.py:354` | ❌ | ❌ | ❌ |
| `create_spatial_group(...)` | `widget_operations.py:393` | ❌ | ❌ | ❌ |
| `find_widget_clusters(...)` | `widget_operations.py:449` | ❌ | ❌ | ❌ |
| `calculate_widget_density(...)` | `widget_operations.py:468` | ❌ | ❌ | ❌ |

### 2.4 `search`

Legacy source: `/home/jaypaulb/Projects/gh/CanvusPythonAPI/canvus_api/search.py`

| Function / type | Legacy line | Go | Python (new SDK) | TS |
|---|---|---|---|---|
| `SearchResult` dataclass | `search.py:16` | ⚠ `widgets.go` defines `WidgetMatch` (smaller surface) | ❌ | ❌ |
| `CrossCanvasSearch.find_widgets_across_canvases` | `search.py:57` | ✅ `widgets.go:277 FindWidgetsAcrossCanvases` (top-level func, simpler signature) | ❌ | ❌ |
| `CrossCanvasSearch.find_widgets_by_text` | `search.py:132` | ❌ | ❌ | ❌ |
| `CrossCanvasSearch.find_widgets_by_type` | `search.py:161` | ❌ | ❌ | ❌ |
| `CrossCanvasSearch.find_widgets_in_area` | `search.py:184` | ❌ | ❌ | ❌ |
| `CrossCanvasSearch.find_widgets_by_property` | `search.py:207` | ❌ | ❌ | ❌ |
| Module-level `find_widgets_across_canvases(...)` | `search.py:471` | ✅ (same as above) | ❌ | ❌ |
| Module-level `find_widgets_by_text(...)` | `search.py:501` | ❌ | ❌ | ❌ |
| Module-level `find_widgets_by_type(...)` | `search.py:527` | ❌ | ❌ | ❌ |
| Module-level `find_widgets_in_area(...)` | `search.py:551` | ❌ | ❌ | ❌ |

### 2.5 `export`

Legacy source: `/home/jaypaulb/Projects/gh/CanvusPythonAPI/canvus_api/export.py`

| Function / type | Legacy line | Go | Python (new SDK) | TS |
|---|---|---|---|---|
| `ExportConfig` | `export.py:22` | ⚠ inlined fields on `Session.ExportWidgetsToFolder` | ❌ | ❌ |
| `ImportConfig` | `export.py:53` | ⚠ inlined into `ImportWidgetsToRegion` | ❌ | ❌ |
| `WidgetExporter` class | `export.py:84` | n/a (Go uses functional style) | ❌ | ❌ |
| `WidgetExporter.export_widgets_to_folder` | `export.py:106` | ✅ `export.go:28 ExportWidgetsToFolder` (downloads images/videos/PDFs, writes export.json) | ❌ | ❌ |
| `WidgetImporter` class | `export.py:350` | n/a | ❌ | ❌ |
| `WidgetImporter.import_widgets_from_folder` | `export.py:366` | ✅ `import.go:15 ImportWidgetsToRegion` (re-creates widgets, scales to target region) | ❌ | ❌ |
| Module-level `export_widgets_to_folder` | `export.py:571` | ✅ | ❌ | ❌ |
| Module-level `import_widgets_from_folder` | `export.py:598` | ✅ | ❌ | ❌ |

---

## 3. Summary

### 3.1 Counts by category

| SDK | Public methods (core) | Subscribe helpers | Helper modules ported | Convenience extras unique to this SDK |
|---|---|---|---|---|
| Go     | ~125 (40 files) | **0 typed** | geometry: partial (5/16); export/import: ✅; filters: partial; search: 1/9 | batch processor, circuit breaker, color utils, warnings, workspace toggle/viewport, single color-preset CRUD, raw config element-array view, token store |
| Python | ~95              | **1 (multiplexed widgets)** | none ported | `RateLimitError.retry_after`, `UnsupportedOperationError`, sync facade, env-driven `from_env`, dedicated worker-loop sync proxy |
| TS     | ~110             | **27 typed** | none ported | env-driven `loadConfig`, complete subscribe coverage, NDJSON helper exported (`streamNdjson`) |

### 3.2 Missing vs. the largest-surface SDK per category

The "largest" surface is **Go** for core methods, **TS** for streaming, **Go** for helper-modules (partial port already exists). Net gaps:

| SDK   | Missing core methods (rel. to Go) | Missing subscribe (rel. to TS) | Missing helpers (rel. to legacy Python) | Total gap |
|---|---|---|---|---|
| Go    | 0 (it IS the largest core)        | 27                              | ~25 (most of geometry/filters/widget_operations/search) | ~52 |
| Python| ~25 (workspace toggles, viewport, color-preset per-name CRUD, generic widget create/update/delete, PatchParentID, raw config, set-source-by-index, get-current-user, trash helpers, etc.) | 26                              | ~50 (all five modules) | ~101 |
| TS    | ~20 (trash, demo-state, color-preset per-name, generic create/update/delete, PatchParentID, raw config, batch, workspace toggles/viewport, set-source-by-index, get-current-user, etc.) | 0 (it IS the largest)           | ~50 (all five modules) | ~70 |

### 3.3 Complexity per gap

- **trivial (< 1 hour each)**: typed subscribe wrapper around an existing endpoint, simple convenience like `TrashCanvas`, get/set color preset per-name, raw config element-array view, set-source-by-index helper, get-current-user helper.
- **moderate (1–4 hours each)**: env-driven config in Go, request-id injection across all three, generic widget create/update/delete in Python/TS (need widget_type dispatch + IPVideo/RDP rejection), batch processor port, geometry helpers (intersect/union/distance/in-area), filter builders.
- **large (> 4 hours each)**: full helper-module ports (`search`, `widget_operations`, `export/import`) — these involve cross-canvas iteration, asset download/upload pipelines, manifest serialization, and Python/TS adaptations of Go's `ExportWidgetsToFolder`/`ImportWidgetsToRegion`.

---

## 4. Per-SDK work items for Phase 4b

Items ordered roughly by foundation → ergonomics → helpers → big ports. Citations use repo-relative paths.

### 4.1 Go

| # | Method / capability | Target file | Complexity | One-line spec |
|---|---|---|---|---|
| 1 | `FromEnv(opts ...) *Session` constructor | `go/sdk/canvus/options.go` | trivial | Read `CANVUS_API_URL`, `CANVUS_API_KEY`, `CANVUS_TIMEOUT_MS`, `CANVUS_VERIFY_TLS`; return configured session. |
| 2 | `(s *Session) GetCurrentUser(ctx)` | `go/sdk/canvus/users.go` | trivial | Hit `GET /users/current`, return `*User`. |
| 3 | Request-ID injection option `WithRequestIDFunc(func() string)` | `go/sdk/canvus/options.go` + `session.go:doRequest` | moderate | Add header `X-Request-ID` on every request; surface back on `APIError.RequestID`. |
| 4 | `UnsupportedOperationError` sentinel | `go/sdk/canvus/errors.go` | trivial | Replace plain `errors.New` in `widgets.go:53` IPVideo/RDP rejection with typed error. |
| 5 | Typed subscribe for `canvases`, `canvases/{id}` | `go/sdk/canvus/canvases.go` | moderate | New `SubscribeCanvases(ctx) (<-chan Canvas, error)` reusing `doRequestWithHeaders` + bufio scanner; ndjson decode; respect ctx for cancel. |
| 6 | Typed subscribe for `canvas-folders`, `canvas-folders/{id}` | `go/sdk/canvus/folders.go` | moderate | Same pattern as #5. |
| 7 | Typed subscribe for `widgets` (mixed + per-type, list + per-id) | `go/sdk/canvus/widgets.go` and per-type files | moderate (×11) | Symmetric NDJSON channel helpers — pattern reused. |
| 8 | Typed subscribe for `users`, `users/{id}`, `users/{id}/access-tokens` | `go/sdk/canvus/users.go`, `accesstokens.go` | moderate | Same. |
| 9 | Typed subscribe for `groups`, `groups/{id}`, `groups/{id}/members` | `go/sdk/canvus/groups.go` | moderate | Same. |
| 10 | Typed subscribe for `server-config`, `license`, `clients`, `clients/{id}`, `workspaces`, `video-outputs`, `video-inputs`, permissions | `serverconfig.go`, `license.go`, `clients.go`, `workspaces.go`, `videooutputs.go`, `videoinputs.go`, `canvases.go`, `folders.go` | moderate | Same. |
| 11 | `geometry.Intersects`, `GetIntersection`, `GetUnion`, `WidgetsIntersect`, `GetWidgetIntersection`, `GetWidgetUnion`, `DistanceBetweenWidgets`, `FindWidgetsInArea(local)`, `FindWidgetsContainingPoint`, `GetCanvasBounds` | `go/sdk/canvus/geometry.go` | moderate | Port from `CanvusPythonAPI/canvus_api/geometry.py:127-406`. |
| 12 | `filters.NewSpatialCondition`, `NewWildcardCondition`, `NewTextFilter`, `NewWidgetTypeFilter`, `CombineFilters`, `Filter.ToMap/FromMap` | `go/sdk/canvus/types.go` or new `filters.go` | moderate | Port from `filters.py:33-323`. |
| 13 | `WidgetZoneManager` (zones from widgets, widgets-in-zone, widgets-touching-zone) | new `go/sdk/canvus/zones.go` | large | Port from `widget_operations.py:29-189`. |
| 14 | `BatchWidgetOperations` (move/resize widgets, widgets-touch-id) | extend `go/sdk/canvus/widgets.go` or `batch.go` | moderate | Port from `widget_operations.py:191-388`. |
| 15 | `CrossCanvasSearch`: find_by_text, find_by_type, find_in_area, find_by_property | `go/sdk/canvus/widgets.go` | large | Port from `search.py:57-258`; expand `WidgetMatch` to carry `match_score`, `match_reason`, drill-down path. |
| 16 | `WithConnectTimeout(d)` option | `go/sdk/canvus/options.go` | trivial | Split out connect timeout from `RequestTimeout`. |

### 4.2 Python

| # | Method / capability | Target file | Complexity | One-line spec |
|---|---|---|---|---|
| 1 | `widgets.create_generic(canvas_id, payload)` (dispatch by `widget_type`) | `python/sdk/src/canvus_sdk/resources/widgets.py` | moderate | Mirror Go `widgets.go:53 CreateWidget`: parse `widget_type`, raise `UnsupportedOperationError` for IPVideo/RDP, dispatch to per-type create/upload. |
| 2 | `widgets.update_generic(canvas_id, widget_id, payload)` | same | moderate | Mirror Go `widgets.go:120 UpdateWidget`: dispatch on `widget_type`, strip immutable fields (e.g. `grid_size` on tables). |
| 3 | `widgets.delete_generic(canvas_id, widget_id, widget_type)` | same | trivial | Mirror Go `widgets.go:193 DeleteWidget`. |
| 4 | `widgets.patch_parent_id(canvas_id, widget_id, parent_id)` | same | trivial | Mirror Go `widgets.go:224 PatchParentID`. |
| 5 | `canvases.trash(canvas_id)` (logged-in user only) | `python/sdk/src/canvus_sdk/resources/canvases.py` | trivial | Find current user's trash folder, PATCH `folder_id` (mirrors Go `canvases.go:95`). Requires `auth.get_current_user` first. |
| 6 | `folders.trash(folder_id)` | `python/sdk/src/canvus_sdk/resources/canvases.py` | trivial | Mirror Go `folders.go:135`. |
| 7 | `auth.get_current_user()` | `python/sdk/src/canvus_sdk/resources/auth.py` | trivial | `GET /users/current`. |
| 8 | `canvases.list_color_presets`, `get_color_preset`, `create_color_preset`, `update_color_preset`, `delete_color_preset` | `python/sdk/src/canvus_sdk/resources/canvases.py` | moderate | Client-side decomposition of the bulk `color_presets` object (mirrors Go `colorpresets.go:38-79`). |
| 9 | `server.get_config_raw()` (flat element-array form) | `python/sdk/src/canvus_sdk/resources/server.py` | trivial | Convert nested config object to spec's flat `[{setting-key, setting-value, setting-type}, …]` form. |
| 10 | `server.set_video_output_source_by_index(client_id, index, source)` | `python/sdk/src/canvus_sdk/resources/server.py` | trivial | Mirror Go `videooutputs.go:28`. |
| 11 | `server.toggle_workspace_info_panel`, `toggle_workspace_pinned` | `python/sdk/src/canvus_sdk/resources/server.py` | trivial | Mirror Go `workspaces.go:79/90`. |
| 12 | `server.set_workspace_viewport(client_id, selector, opts)` | `python/sdk/src/canvus_sdk/resources/server.py` | moderate | Mirror Go `workspaces.go:102` (fetches widget, computes `view_rectangle`, patches workspace). |
| 13 | Typed subscribe coverage: canvases/folders/users/groups/access-tokens/server-config/license/clients/workspaces/video-outputs/video-inputs/permissions + per-type widget subscribe | every resource file | moderate (×~26) | Wrap existing `_http.stream_lines` in typed `AsyncIterator[Model]` helpers, mirroring TS's complete catalogue. |
| 14 | Subscribe helpers for tables.cells, uploads-folder | `python/sdk/src/canvus_sdk/resources/widgets.py` | trivial | Already have `stream_lines`; just yield decoded models. |
| 15 | `ValidationError.issues` structured field | `python/sdk/src/canvus_sdk/errors.py` | trivial | Add `issues: list[dict]` to match TS shape. |
| 16 | `geometry` module (full port) | new `python/sdk/src/canvus_sdk/geometry.py` | large | Port `CanvusPythonAPI/canvus_api/geometry.py:14-419` (Point/Size/Rectangle dataclasses already partially exist in `models/common.py`). |
| 17 | `filters` module (full port) | new `python/sdk/src/canvus_sdk/filters.py` | large | Port `CanvusPythonAPI/canvus_api/filters.py:11-323`. |
| 18 | `widget_operations` module (full port) | new `python/sdk/src/canvus_sdk/widget_operations.py` | large | Port `CanvusPythonAPI/canvus_api/widget_operations.py:20-488`. |
| 19 | `search` module (full port) | new `python/sdk/src/canvus_sdk/search.py` | large | Port `CanvusPythonAPI/canvus_api/search.py:16-577`. |
| 20 | `export/import` (port) | new `python/sdk/src/canvus_sdk/export.py` + `import_.py` | large | Port `CanvusPythonAPI/canvus_api/export.py:22-630` and align with Go's `export.go`/`import.go` on-disk layout (export.json schema). |
| 21 | `batch` processor | new `python/sdk/src/canvus_sdk/batch.py` | moderate | Port Go `batch.go`: bounded-concurrency move/copy/delete with retry, ContinueOnError, ProgressCallback. |
| 22 | Color utilities | new `python/sdk/src/canvus_sdk/color.py` | trivial | Port Go `color.go` (validate/normalize, RGBA↔hex, alpha). |
| 23 | `APIWarning` registry | new `python/sdk/src/canvus_sdk/warnings.py` | trivial | Mirror Go `warnings.go:9-135` (known API limitations, suppress flags). |

### 4.3 TypeScript

| # | Method / capability | Target file | Complexity | One-line spec |
|---|---|---|---|---|
| 1 | Generic widget `create`, `update`, `delete` on `WidgetsResource` (dispatch by `widget_type`) | `typescript/sdk/src/resources/widgets.ts` | moderate | Mirror Go `widgets.go:53/120/193`; throw `UnsupportedOperationError` (new class) on IPVideo/RDP create. |
| 2 | `widgets.patchParentId(canvasId, widgetId, parentId)` | same | trivial | Mirror Go `widgets.go:224`. |
| 3 | `canvases.trash(canvasId)` | `typescript/sdk/src/resources/canvases.ts` | trivial | Requires current user lookup; mirrors Go `canvases.go:95`. |
| 4 | `folders.trash(folderId)` | `typescript/sdk/src/resources/folders.ts` | trivial | Mirror Go `folders.go:135`. |
| 5 | `auth.currentUser()` | `typescript/sdk/src/resources/auth.ts` | trivial | `GET /users/current`. |
| 6 | Single color-preset CRUD: `listColorPresets`, `getColorPreset(name)`, `createColorPreset`, `updateColorPreset(name)`, `deleteColorPreset(name)` | `typescript/sdk/src/resources/canvases.ts` | moderate | Client-side decomposition of bulk presets object; mirrors Go `colorpresets.go:38-79`. |
| 7 | `server.configRaw()` (flat element-array view) | `typescript/sdk/src/resources/server.ts` | trivial | Mirror Go `serverconfig.go:89 GetServerConfigRaw`. |
| 8 | `server.setVideoOutputSourceByIndex(clientId, index, body)` | `typescript/sdk/src/resources/server.ts` | trivial | Mirror Go `videooutputs.go:28`. |
| 9 | `server.toggleWorkspaceInfoPanel`, `toggleWorkspacePinned`, `setWorkspaceViewport` | `typescript/sdk/src/resources/server.ts` | moderate | Mirror Go `workspaces.go:79/90/102`. |
| 10 | Retry layer in transport | `typescript/sdk/src/transport.ts` | moderate | Mirror Go `session.go:376-465` retry+backoff+circuit; honour `Retry-After`. |
| 11 | `RateLimitError` (429) + `ServerError` (5xx) | `typescript/sdk/src/errors.ts` | trivial | Add subclasses + `retry-after` parsing. |
| 12 | Request-ID injection option | `typescript/sdk/src/config.ts` + `transport.ts` | moderate | Add `requestIdProvider` to `SessionOptions`, set `X-Request-ID` header, surface on `APIError`. |
| 13 | `verifyTls` actually honoured | `typescript/sdk/src/transport.ts` | moderate | Today the field is parsed but `fetch` ignores it under Node; wire to a custom Agent (Node) or document Web-only. |
| 14 | `geometry` module (port) | new `typescript/sdk/src/geometry.ts` | large | Port `CanvusPythonAPI/canvus_api/geometry.py:14-419`; reuse existing `Location`/`Size` types from `types/common.ts`. |
| 15 | `filters` module (port) | new `typescript/sdk/src/filters.ts` | large | Port `filters.py:11-323`. |
| 16 | `widget_operations` module (port) | new `typescript/sdk/src/widgetOperations.ts` | large | Port `widget_operations.py:20-488`. |
| 17 | `search` module (port) | new `typescript/sdk/src/search.ts` | large | Port `search.py:16-577`. |
| 18 | `export/import` module (port) | new `typescript/sdk/src/export.ts` + `import.ts` | large | Port `export.py:22-630`; share export.json schema with Go. |
| 19 | `batch` processor | new `typescript/sdk/src/batch.ts` | moderate | Mirror Go `batch.go` (`BatchOperationBuilder`, bounded concurrency, summarize). |
| 20 | Color utilities | new `typescript/sdk/src/color.ts` | trivial | Mirror Go `color.go`. |
| 21 | `APIWarning` registry | new `typescript/sdk/src/warnings.ts` | trivial | Mirror Go `warnings.go:9-135`. |
| 22 | Session lifecycle / `close()` | `typescript/sdk/src/session.ts` | trivial | Add `close()` to release any retained agent/keepalive socket. |

---

## 5. Notes

### 5.1 Legacy `client.py` spot-check (~112k bytes, 138 public methods)

Methods present in legacy `CanvusPythonAPI/canvus_api/client.py` but **NOT** in any of the three new SDKs:

- `get_current_user()` (`client.py:2104`) — already flagged in §1.1.6.
- `find_admin_client(...)` (`client.py:2746`) — utility to locate an admin client from `clients` list; not particularly load-bearing but legacy callers depend on it.
- `subscribe_widgets(...)` with custom **diff handler** callback (`client.py:2779`) — the new SDKs' subscribe helpers yield raw updates; consumers must compute diffs themselves. Worth considering as a `diff_callback=` keyword on the new `widgets.subscribe`.
- `subscribe_workspace(client_id, callback)` (`client.py:2847`) — callback-style subscribe for workspace updates. Not yet present anywhere.
- `subscribe_note(canvas_id, note_id, callback)` (`client.py:2904`) — callback-style note subscribe; TS offers generator-style `notes.subscribeOne`, but no callback adapter.
- `subscribe_annotations(canvas_id, ...)` (`client.py:1526`) — listening for annotation events on widgets. The new SDKs expose `?annotations=true` on list (Go: `widgets.go:23` `includeAnnotations`) but no subscribe variant.
- `list_widget_annotations(canvas_id)` (`client.py:1510`) — explicit list endpoint for annotations. Missing from all three.
- `set_canvas_mode(canvas_id, is_demo)` (`client.py:762`) — toggles canvas demo mode. Maps to `PATCH /canvases/{id}` with `mode` field; trivially expressible via `update_canvas` but the convenience helper is gone.
- `change_password(user_id, …)` self-service variant (`client.py:1974`) — Go has `SetUserPassword` (admin) but no self-change helper.

### 5.2 `CloneWidget` per-type vs single-helper

All three SDKs implement `clone` as a single helper that takes `widget_type` as a parameter, then dispatches to the appropriate per-type create endpoint (verified §1 of `VERIFIED-CORRECTIONS.md`/`changelog.md`). Per-type wrappers (e.g. `notes.clone`, `images.clone`) would be sugar with no functional difference — recommended NOT to add unless callers report friction.

### 5.3 Bidirectional parity: Go-only methods Python/TS lack

These are Go-only convenience methods that arguably DO belong in the other two SDKs (covered in §4 work items but called out here for emphasis):

- `TrashCanvas`, `TrashFolder` — client-side helpers that PATCH into a user's trash folder. Missing from both Py/TS.
- `PatchParentID` — generic re-parent helper. Missing from both.
- Generic `CreateWidget` / `UpdateWidget` / `DeleteWidget` with `widget_type` dispatch. Missing from both.
- `ListColorPresets` / single-preset CRUD. Missing from both.
- `GetServerConfigRaw` (flat element-array view) — useful for the spec-shaped PATCH writes. Missing from both.
- `SetVideoOutputSource(by index)` — useful when working with the ordinal interface. Missing from both.
- `ToggleWorkspaceInfoPanel`, `ToggleWorkspacePinned`, `SetWorkspaceViewport` — workspace orchestration. Missing from both.
- `NewTestClient` / `NewUserClient` test helpers (`clients.go:76/114`) — these are intentional Go-only conveniences for integration tests; arguably do NOT need ports.
- Circuit breaker — Go-only resilience pattern; would be a substantial addition for the others.
- Token store interface (`TokenStore`) for token persistence — Go-only.

### 5.4 IPVideo / RDP create suppression

Three different mechanisms across SDKs for the same wire-level constraint (server returns 501 for `POST /ip-videos`, `POST /rdp-connections`):
- Go: returns a plain `errors.New` from `widgets.go:53` `CreateWidget`. Should become a typed `UnsupportedOperationError`.
- Python: raises `UnsupportedOperationError` from per-type `create` methods (`widgets.py:437/476`).
- TS: omits `create` from `ipVideos`/`rdpConnections` namespaces entirely (compile-time enforcement).

All three are correct — but TS's compile-time enforcement is the most defensible. Python's runtime exception is a fallback; Go's plain error is a bug to fix.

### 5.5 Pending design questions

- **Should the Go SDK gain `from_env`?** Python and TS both have it. Yes — trivial, parity-relevant.
- **Should subscribe helpers in Go and Python use channels/generators or callbacks?** Current Python uses async iterator (good idiomatic match), TS uses async generator. Recommend Go use `<-chan T` for type safety; document explicit cancellation via ctx.
- **Should the helper modules be ported into the SDKs themselves, or live as a sibling `extras/` package?** TS has an empty `src/extras/` placeholder, suggesting the latter. Recommend `extras/` for all three to keep the core SDK small and the rich utilities optional.
- **Should the legacy `subscribe_*` callback-style helpers (with diff callback) be reproduced?** They're useful for naive consumers; consider as a sibling `extras/diffs` utility rather than expanding the core subscribe surface.
- **Generic widget create/update/delete in Python and TS** would conflict with the typed per-type APIs that are the SDKs' main selling point. Recommend exposing them as `widgets.createAny(...)` / `widgets.updateAny(...)` to keep the typed surface paramount.
- **Permissions subscribe in Python/Go** — TS exposes both canvas and folder permission subscribe. The streaming spec lists these as supported, but it's worth verifying live before porting (not on the current verified-corrections list).

### 5.6 `verifyTls` lies in TS

`config.ts:74 buildConfig` accepts `verifyTls`, but the `Transport` (`transport.ts:42`) uses raw `fetch` — under Node's experimental `fetch`, there is no TLS-verification toggle short of an undici Agent. The flag is currently parsed and ignored. Flagged for fix as part of TS work item #13.

---

## 6. Completion report

- **Total parity gaps surfaced:**
  - Go: 52 (work items 1-16; some bundle multiple endpoints)
  - Python: 101 (work items 1-23)
  - TS: 70 (work items 1-22)
  - **Aggregate: ~223 distinct gaps** (counting each subscribe helper individually, each helper-module function as one, and each missing convenience method as one)

- **Category breakdown:**
  - **Subscribe / streaming:** Go 27, Python 26, TS 0 → **53 typed-subscribe helpers needed**
  - **Helper-module ports (`geometry`, `filters`, `widget_operations`, `search`, `export`):** Python ~50, TS ~50, Go ~25 → **125 ports**
  - **Convenience / extras (trash, color-preset-per-name, raw config, set-source-by-index, workspace orchestration, generic widget CRUD, patch_parent_id, current-user, batch processor, color utilities, warnings registry):** Python 14, TS 14, Go 4 → **32 helpers**
  - **Ergonomics (env-driven init, request-id injection, retry layer, structured validation errors, lifecycle):** Go 3, Python 2, TS 4 → **9 ergonomic improvements**
  - **Error-type alignment (`UnsupportedOperationError`, `RateLimitError`, `ServerError`, structured validation issues):** ~4 trivial alignment items

- **Non-trivial design questions raised:**
  1. Should helpers live in the SDK core or in a sibling `extras/` package?
  2. Should the legacy callback-style `subscribe_*` (with diff callbacks) be ported, or only the iterator-style helpers?
  3. Should generic widget `create/update/delete` be exposed alongside typed per-type APIs in Python and TS (and named distinctively, e.g. `createAny`)?
  4. Should Go gain a from-env constructor (recommended: yes)?
  5. Should batch / circuit-breaker / token-store features be ported to Python and TS, or remain Go-only resilience tooling?
  6. Permissions-subscribe needs live verification before porting (not on the current corrections list).

- **Matrix path:** `/home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/parity-matrix.md`
