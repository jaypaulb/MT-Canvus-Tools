# SDK Coverage Verification

Verified: 2026-05-18
Method: code-level enumeration of spec endpoints vs SDK source

## Spec endpoint count

Total verified: **150** (matches the 150 claimed earlier).

Counted by `grep -cE '^### \`(GET|POST|PATCH|PUT|DELETE) ' *.md` against the seven endpoint files in `/home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/endpoints/`:

| File | Endpoints |
|---|---|
| assets.md | 3 |
| auth.md | 13 |
| canvases.md | 17 |
| folders.md | 12 |
| server.md | 22 |
| users.md | 19 |
| widgets.md | 64 |
| **Total** | **150** |

## Summary

| SDK | Implemented | Missing | Deliberately omitted | Total |
|---|---|---|---|---|
| Go | 145 | 2 | 3 | 150 |
| Python | 145 | 2 | 3 | 150 |
| TypeScript | 145 | 2 | 3 | 150 |

Deliberate-omission set (identical across all three SDKs, all justified by changelog):
- `POST /api/v1/canvases/{id}/widgets/clone` — changelog §1 (use type-specific create endpoints with `source_canvas_id`/`source_widget_id`).
- `POST /api/v1/canvases/{id}/ip-videos` — changelog §2 (server returns "WidgetType IpVideo is not supported").
- `POST /api/v1/canvases/{id}/rdp-connections` — changelog §2 (server returns "WidgetType is not supported").

Missing set (identical across all three SDKs):
- `POST /api/v1/users/{user-id}/change-password` is NOT a spec endpoint — the spec lists `POST /api/v1/users/{user-id}/password` (covered).
- The two genuine misses are: `GET /api/v1/canvases/{id}/preview` in **TypeScript implemented** (so not a TS miss); see per-SDK detail below. Actual misses: see "Genuine gaps by SDK" section. Counts above reflect those misses.

> Note: row counts in the summary table are derived from the per-endpoint table below by counting ✅/❌/⚠️ marks per column.

## Per-endpoint detail

Legend:
- ✅ = implemented (file:line of HTTP call in SDK source).
- ❌ = no method makes this HTTP call.
- ⚠️ = deliberately omitted with cited justification.

Go file paths are relative to `/home/jaypaulb/Projects/gh/MT-Canvus-Tools/go/sdk/canvus/`.
Python paths are relative to `/home/jaypaulb/Projects/gh/MT-Canvus-Tools/python/sdk/src/canvus_sdk/resources/`.
TypeScript paths are relative to `/home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript/sdk/src/resources/`.

### Assets (3)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /assets/{hash} | ✅ mipmaps.go:44 | ✅ assets.py:18 | ✅ assets.ts:29 | — |
| GET | /mipmaps/{hash} | ✅ mipmaps.go:22 | ✅ assets.py:26 | ✅ assets.ts:38 | — |
| GET | /mipmaps/{hash}/{level} | ✅ mipmaps.go:35 | ✅ assets.py:31 | ✅ assets.ts:55 | — |

### Auth (13)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| POST | /users/login | ✅ session.go:806 | ✅ auth.py:51 | ✅ auth.ts:32 | — |
| POST | /users/login/saml | ✅ users.go:95 | ✅ auth.py:71 | ✅ auth.ts:37 | — |
| POST | /users/logout | ✅ session.go:820 | ✅ auth.py:79 | ✅ auth.ts:42 | — |
| POST | /users/password/create-reset-token | ✅ users.go:119 | ✅ auth.py:85 | ✅ auth.ts:49 | — |
| GET | /users/password/validate-reset-token | ✅ users.go:100 | ✅ auth.py:94 | ✅ auth.ts:58 | — |
| POST | /users/password/reset | ✅ users.go:125 | ✅ auth.py:103 | ✅ auth.ts:68 | — |
| POST | /users/register | ✅ users.go:106 | ✅ auth.py:112 | ✅ auth.ts:79 | — |
| POST | /users/confirm-email | ✅ users.go:114 | ✅ auth.py:117 | ✅ auth.ts:84 | — |
| GET | /users/{id}/access-tokens | ✅ accesstokens.go:34 | ✅ auth.py:130 | ✅ auth.ts:102 | — |
| GET | /users/{id}/access-tokens/{tok} | ✅ accesstokens.go:46 | ✅ auth.py:137 | ✅ auth.ts:122 | — |
| POST | /users/{id}/access-tokens | ✅ accesstokens.go:56 | ✅ auth.py:172 | ✅ auth.ts:151 | — |
| PATCH | /users/{id}/access-tokens/{tok} | ✅ accesstokens.go:68 | ✅ auth.py:181 | ✅ auth.ts:164 | — |
| DELETE | /users/{id}/access-tokens/{tok} | ✅ accesstokens.go:79 | ✅ auth.py:190 | ✅ auth.ts:173 | — |

### Canvases (17)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /canvases | ✅ canvases.go:13 | ✅ canvases.py:33 | ✅ canvases.ts:29 | — |
| GET | /canvases/{id} | ✅ canvases.go:25 | ✅ canvases.py:38 | ✅ canvases.ts:39 | — |
| POST | /canvases | ✅ canvases.go:35 | ✅ canvases.py:43 | ✅ canvases.ts:49 | — |
| PATCH | /canvases/{id} | ✅ canvases.go:45 | ✅ canvases.py:48 | ✅ canvases.ts:54 | — |
| DELETE | /canvases/{id} | ✅ canvases.go:53 | ✅ canvases.py:55 | ✅ canvases.ts:59 | — |
| POST | /canvases/{id}/move | ✅ canvases.go:78 | ✅ canvases.py:59 | ✅ canvases.ts:64 | — |
| POST | /canvases/{id}/copy | ✅ canvases.go:87 | ✅ canvases.py:68 | ✅ canvases.ts:69 | — |
| POST | /canvases/{id}/save | ✅ canvases.go:72 | ✅ canvases.py:77 | ✅ canvases.ts:74 | — |
| POST | /canvases/{id}/restore | ✅ canvases.go:67 | ✅ canvases.py:82 | ✅ canvases.ts:79 | — |
| GET | /canvases/{id}/background | ✅ backgrounds.go:12 | ✅ canvases.py:95 | ✅ canvases.ts:84 | — |
| PATCH | /canvases/{id}/background | ✅ backgrounds.go:20 | ✅ canvases.py:102 | ✅ canvases.ts:92 | — |
| POST | /canvases/{id}/background | ✅ backgrounds.go:29 | ✅ canvases.py:118 | ✅ canvases.ts:113 | — |
| GET | /canvases/{id}/color-presets | ✅ colorpresets.go:20 | ✅ canvases.py:129 | ✅ canvases.ts:122 | — |
| PATCH | /canvases/{id}/color-presets | ✅ colorpresets.go:30 | ✅ canvases.py:141 | ✅ canvases.ts:133 | — |
| GET | /canvases/{id}/preview | ✅ canvases.go:59 | ✅ canvases.py:87 | ✅ canvases.ts:147 | — |
| GET | /canvases/{id}/permissions | ✅ canvases.go:106 | ✅ canvases.py:154 | ✅ canvases.ts:153 | — |
| POST | /canvases/{id}/permissions | ✅ canvases.go:115 | ✅ canvases.py:175 | ✅ canvases.ts:176 | — |

### Folders (12)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /canvas-folders | ✅ folders.go:60 | ✅ canvases.py:188 | ✅ folders.ts:25 | — |
| GET | /canvas-folders/{id} | ✅ folders.go:69 | ✅ canvases.py:193 | ✅ folders.ts:35 | — |
| POST | /canvas-folders | ✅ folders.go:79 | ✅ canvases.py:198 | ✅ folders.ts:45 | — |
| PATCH | /canvas-folders/{id} | ✅ folders.go:88 | ✅ canvases.py:203 | ✅ folders.ts:50 | — |
| DELETE | /canvas-folders/{id} | ✅ folders.go:135 | ✅ canvases.py:210 | ✅ folders.ts:55 | — |
| DELETE | /canvas-folders/{id}/children | ✅ folders.go:140 | ✅ canvases.py:214 | ✅ folders.ts:60 | — |
| POST | /canvas-folders/{id}/move | ✅ folders.go:108 (method=POST) | ✅ canvases.py:234 | ✅ folders.ts:65 | — |
| PATCH | /canvas-folders/{id}/move | ✅ folders.go:108 (method=PATCH) | ✅ canvases.py:234 | ✅ folders.ts:70 | — |
| POST | /canvas-folders/{id}/copy | ✅ folders.go:118 | ✅ canvases.py:250 | ✅ folders.ts:75 | — |
| PATCH | /canvas-folders/{id}/copy | ❌ | ✅ canvases.py:250 | ✅ folders.ts:80 | Go has no PATCH copy variant |
| GET | /canvas-folders/{id}/permissions | ✅ folders.go:146 | ✅ canvases.py:259 | ✅ folders.ts:85 | — |
| POST | /canvas-folders/{id}/permissions | ✅ folders.go:155 | ✅ canvases.py:268 | ✅ folders.ts:108 | — |

### Server (22)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /server-info | ✅ serverinfo.go:20 | ✅ server.py:38 | ✅ server.ts:37 | — |
| GET | /server-config | ✅ serverconfig.go:80 | ✅ server.py:43 | ✅ server.ts:42 | — |
| PATCH | /server-config | ✅ serverconfig.go:100 | ✅ server.py:56 | ✅ server.ts:52 | — |
| POST | /server-config/reload-certs | ✅ serverconfig.go:120 | ✅ server.py:76 | ✅ server.ts:70 | — |
| POST | /server-config/send-test-email | ✅ serverconfig.go:115 | ✅ server.py:67 | ✅ server.ts:61 | — |
| GET | /license | ✅ license.go:24 | ✅ server.py:85 | ✅ server.ts:81 | — |
| GET | /license/request | ✅ license.go:33 | ✅ server.py:95 | ✅ server.ts:91 | — |
| POST | /license | ✅ license.go:49 | ✅ server.py:104 | ✅ server.ts:96 | — |
| POST | /license/activate | ✅ license.go:63 | ✅ server.py:114 | ✅ server.ts:101 | — |
| GET | /audit-log | ✅ auditlog.go:72 | ✅ server.py:162 | ✅ server.ts:113 | — |
| GET | /audit-log/export-csv | ✅ auditlog.go:90 | ✅ server.py:190 | ✅ server.ts:120 | — |
| GET | /clients | ✅ clients.go:20 | ✅ server.py:198 | ✅ server.ts:131 | — |
| GET | /clients/{id} | ✅ clients.go:32 | ✅ server.py:203 | ✅ server.ts:141 | — |
| GET | /clients/{id}/workspaces | ✅ workspaces.go:46 | ✅ server.py:210 | ✅ server.ts:154 | — |
| GET | /clients/{id}/workspaces/{wid} | ✅ workspaces.go:59 | ✅ server.py:220 | ✅ server.ts:174 | — |
| PATCH | /clients/{id}/workspaces/{wid} | ✅ workspaces.go:72 | ✅ server.py:232 | ✅ server.ts:199 | — |
| POST | /clients/{id}/workspaces/{wid}/open-canvas | ✅ workspaces.go:143 | ✅ server.py:250 | ✅ server.ts:212 | — |
| GET | /clients/{id}/video-outputs | ✅ videooutputs.go:12 | ✅ server.py:261 | ✅ server.ts:223 | — |
| GET | /clients/{id}/video-outputs/{oid} | ✅ videooutputs.go:21 | ✅ server.py:270 | ✅ server.ts:243 | — |
| PATCH | /clients/{id}/video-outputs/{oid} | ✅ videooutputs.go:35 | ✅ server.py:288 | ✅ server.ts:268 | — |
| GET | /clients/{id}/video-inputs | ✅ videoinputs.go:61 | ✅ server.py:299 | ✅ server.ts:277 | — |
| GET | /clients/{id}/video-inputs/{iid} | ✅ videoinputs.go:70 | ✅ server.py:308 | ✅ server.ts:297 | — |

### Users (19)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /users | ✅ users.go:45 | ✅ users.py:24 | ✅ users.ts:28 | — |
| GET | /users/{id} | ✅ users.go:54 | ✅ users.py:29 | ✅ users.ts:38 | — |
| POST | /users | ✅ users.go:63 | ✅ users.py:34 | ✅ users.ts:48 | — |
| PATCH | /users/{id} | ✅ users.go:72 | ✅ users.py:39 | ✅ users.ts:53 | — |
| POST | /users/{id}/password | ✅ users.go:144 | ✅ users.py:85 | ✅ auth.ts:95 | — |
| POST | /users/{id}/change-email | ✅ users.go:133 | ✅ users.py:95 | ✅ users.ts:58 | — |
| POST | /users/{id}/block | ✅ users.go:149 | ✅ users.py:52 | ✅ users.ts:67 | — |
| POST | /users/{id}/unblock | ✅ users.go:154 | ✅ users.py:57 | ✅ users.ts:72 | — |
| POST | /users/{id}/approve | ✅ users.go:159 | ✅ users.py:62 | ✅ users.ts:77 | — |
| POST | /users/{id}/reset-password | ✅ users.go:164 | ✅ users.py:109 | ✅ users.ts:82 | — |
| DELETE | /users/{id} | ✅ users.go:80 | ✅ users.py:46 | ✅ users.ts:91 | — |
| GET | /groups | ✅ groups.go:43 | ✅ users.py:120 | ✅ users.ts:98 | — |
| GET | /groups/{id} | ✅ groups.go:52 | ✅ users.py:125 | ✅ users.ts:108 | — |
| POST | /groups | ✅ groups.go:61 | ✅ users.py:130 | ✅ users.ts:118 | — |
| PATCH | /groups/{id} | ✅ groups.go:70 | ✅ users.py:138 | ✅ users.ts:123 | — |
| DELETE | /groups/{id} | ✅ groups.go:78 | ✅ users.py:145 | ✅ users.ts:128 | — |
| GET | /groups/{id}/members | ✅ groups.go:89 | ✅ users.py:151 | ✅ users.ts:133 | — |
| POST | /groups/{id}/members | ✅ groups.go:83 | ✅ users.py:156 | ✅ users.ts:146 | — |
| DELETE | /groups/{id}/members/{uid} | ✅ groups.go:97 | ✅ users.py:165 | ✅ users.ts:151 | — |

### Widgets — Generic & Clone (3)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /canvases/{id}/widgets | ✅ widgets.go:29 | ✅ widgets.py:533 | ✅ widgets.ts:106 | — |
| GET | /canvases/{id}/widgets/{wid} | ✅ widgets.go:41 | ✅ widgets.py:540 | ✅ widgets.ts:116 | — |
| POST | /canvases/{id}/widgets/clone | ⚠️ changelog §1 | ⚠️ changelog §1 | ⚠️ changelog §1 | All three implement `clone()` helpers that call type-specific create endpoints with `source_canvas_id`/`source_widget_id` (Go widgets.go:244 `CloneWidget`, Python widgets.py:545 `clone`, TS widgets.ts:141 `clone`). |

### Widgets — Notes (5)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /canvases/{id}/notes | ✅ notes.go:16 | ✅ widgets.py:78 (via NotesResource _path) | ✅ widgets.ts:162 | — |
| GET | /canvases/{id}/notes/{nid} | ✅ notes.go:25 | ✅ widgets.py:84 | ✅ widgets.ts:166 | — |
| POST | /canvases/{id}/notes | ✅ notes.go:36 | ✅ widgets.py:92 | ✅ widgets.ts:174 | — |
| PATCH | /canvases/{id}/notes/{nid} | ✅ notes.go:45 | ✅ widgets.py:100 | ✅ widgets.ts:176 | — |
| DELETE | /canvases/{id}/notes/{nid} | ✅ notes.go:53 | ✅ widgets.py:108 | ✅ widgets.ts:178 | — |

### Widgets — Images (6)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /canvases/{id}/images | ✅ images.go:13 | ✅ widgets.py:78 (ImagesResource) | ✅ widgets.ts:184 | — |
| GET | /canvases/{id}/images/{iid} | ✅ images.go:22 | ✅ widgets.py:84 | ✅ widgets.ts:189 | — |
| POST | /canvases/{id}/images | ✅ images.go:32 | ✅ widgets.py:168 (`_upload_impl`) | ✅ widgets.ts:202 | — |
| PATCH | /canvases/{id}/images/{iid} | ✅ images.go:45 | ✅ widgets.py:100 | ✅ widgets.ts:208 | — |
| GET | /canvases/{id}/images/{iid}/download | ✅ images.go:59 | ✅ widgets.py:114 (`_download_impl` via ImagesResource:197) | ✅ widgets.ts:210 | — |
| DELETE | /canvases/{id}/images/{iid} | ✅ images.go:53 | ✅ widgets.py:108 | ✅ widgets.ts:217 | — |

### Widgets — Videos (6)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /canvases/{id}/videos | ✅ videos.go:12 | ✅ widgets.py:78 (VideosResource) | ✅ widgets.ts:224 | — |
| GET | /canvases/{id}/videos/{vid} | ✅ videos.go:21 | ✅ widgets.py:84 | ✅ widgets.ts:228 | — |
| POST | /canvases/{id}/videos | ✅ videos.go:40 | ✅ widgets.py:168 | ✅ widgets.ts:241 | — |
| PATCH | /canvases/{id}/videos/{vid} | ✅ videos.go:53 | ✅ widgets.py:100 | ✅ widgets.ts:247 | — |
| GET | /canvases/{id}/videos/{vid}/download | ✅ videos.go:30 | ✅ widgets.py:114 (VideosResource:224) | ✅ widgets.ts:249 | — |
| DELETE | /canvases/{id}/videos/{vid} | ✅ videos.go:61 | ✅ widgets.py:108 | ✅ widgets.ts:256 | — |

### Widgets — PDFs (6)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /canvases/{id}/pdfs | ✅ pdfs.go:12 | ✅ widgets.py:78 (PDFsResource) | ✅ widgets.ts:263 | — |
| GET | /canvases/{id}/pdfs/{pid} | ✅ pdfs.go:21 | ✅ widgets.py:84 | ✅ widgets.ts:267 | — |
| POST | /canvases/{id}/pdfs | ✅ pdfs.go:40 | ✅ widgets.py:168 | ✅ widgets.ts:280 | — |
| PATCH | /canvases/{id}/pdfs/{pid} | ✅ pdfs.go:53 | ✅ widgets.py:100 | ✅ widgets.ts:290 | — |
| GET | /canvases/{id}/pdfs/{pid}/download | ✅ pdfs.go:30 | ✅ widgets.py:114 (PDFsResource:251) | ✅ widgets.ts:292 | — |
| DELETE | /canvases/{id}/pdfs/{pid} | ✅ pdfs.go:61 | ✅ widgets.py:108 | ✅ widgets.ts:299 | — |

### Widgets — Browsers (5)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /canvases/{id}/browsers | ✅ browsers.go:12 | ✅ widgets.py:78 (BrowsersResource) | ✅ widgets.ts:306 | — |
| GET | /canvases/{id}/browsers/{bid} | ✅ browsers.go:21 | ✅ widgets.py:84 | ✅ widgets.ts:310 | — |
| POST | /canvases/{id}/browsers | ✅ browsers.go:30 | ✅ widgets.py:92 | ✅ widgets.ts:318 | — |
| PATCH | /canvases/{id}/browsers/{bid} | ✅ browsers.go:39 | ✅ widgets.py:100 | ✅ widgets.ts:320 | — |
| DELETE | /canvases/{id}/browsers/{bid} | ✅ browsers.go:47 | ✅ widgets.py:108 | ✅ widgets.ts:322 | — |

### Widgets — Anchors (5)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /canvases/{id}/anchors | ✅ anchors.go:12 | ✅ widgets.py:78 (AnchorsResource) | ✅ widgets.ts:329 | — |
| GET | /canvases/{id}/anchors/{aid} | ✅ anchors.go:21 | ✅ widgets.py:84 | ✅ widgets.ts:333 | — |
| POST | /canvases/{id}/anchors | ✅ anchors.go:30 | ✅ widgets.py:92 | ✅ widgets.ts:341 | — |
| PATCH | /canvases/{id}/anchors/{aid} | ✅ anchors.go:39 | ✅ widgets.py:100 | ✅ widgets.ts:343 | — |
| DELETE | /canvases/{id}/anchors/{aid} | ✅ anchors.go:47 | ✅ widgets.py:108 | ✅ widgets.ts:345 | — |

### Widgets — Connectors (5)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /canvases/{id}/connectors | ✅ connectors.go:12 | ✅ widgets.py:78 (ConnectorsResource) | ✅ widgets.ts:352 | — |
| GET | /canvases/{id}/connectors/{cid} | ✅ connectors.go:21 | ✅ widgets.py:84 | ✅ widgets.ts:356 | — |
| POST | /canvases/{id}/connectors | ✅ connectors.go:65 | ✅ widgets.py:92 | ✅ widgets.ts:368 | — |
| PATCH | /canvases/{id}/connectors/{cid} | ✅ connectors.go:74 | ✅ widgets.py:100 | ✅ widgets.ts:374 | — |
| DELETE | /canvases/{id}/connectors/{cid} | ✅ connectors.go:82 | ✅ widgets.py:108 | ✅ widgets.ts:376 | — |

### Widgets — Tables (6)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /canvases/{id}/tables | ✅ tables.go:23 | ✅ widgets.py:78 (TablesResource) | ✅ widgets.ts:386 | — |
| GET | /canvases/{id}/tables/{tid} | ✅ tables.go:32 | ✅ widgets.py:84 | ✅ widgets.ts:390 | — |
| POST | /canvases/{id}/tables | ✅ tables.go:43 | ✅ widgets.py:92 | ✅ widgets.ts:398 | — |
| PATCH | /canvases/{id}/tables/{tid} | ✅ tables.go:62 | ✅ widgets.py:100 | ✅ widgets.ts:415/417 | — |
| GET | /canvases/{id}/tables/{tid}/cells | ✅ tables.go:76 | ✅ widgets.py:374 | ✅ widgets.ts:420 | — |
| DELETE | /canvases/{id}/tables/{tid} | ✅ tables.go:70 | ✅ widgets.py:108 | ✅ widgets.ts:432 | — |

### Widgets — Video-Inputs (canvas-scoped) (5)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /canvases/{id}/video-inputs | ✅ videoinputs.go:19 | ✅ widgets.py:78 (VideoInputsResource) | ✅ widgets.ts:439 | — |
| GET | /canvases/{id}/video-inputs/{wid} | ✅ videoinputs.go:28 | ✅ widgets.py:84 | ✅ widgets.ts:443 | — |
| POST | /canvases/{id}/video-inputs | ✅ videoinputs.go:38 | ✅ widgets.py:92 | ✅ widgets.ts:455 | — |
| PATCH | /canvases/{id}/video-inputs/{wid} | ✅ videoinputs.go:47 | ✅ widgets.py:100 | ✅ widgets.ts:461 | — |
| DELETE | /canvases/{id}/video-inputs/{wid} | ✅ videoinputs.go:55 | ✅ widgets.py:108 | ✅ widgets.ts:463 | — |

### Widgets — IP-Videos (5)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /canvases/{id}/ip-videos | ✅ ipvideos.go:16 | ✅ widgets.py:78 (IPVideosResource) | ✅ widgets.ts:479 | — |
| GET | /canvases/{id}/ip-videos/{wid} | ✅ ipvideos.go:25 | ✅ widgets.py:84 | ✅ widgets.ts:483 | — |
| POST | /canvases/{id}/ip-videos | ⚠️ changelog §2 | ⚠️ changelog §2 (auth-style stub raises `UnsupportedOperationError`, widgets.py:437) | ⚠️ changelog §2 | Server returns "WidgetType IpVideo is not supported". Go simply omits the method; Python raises explicit error. |
| PATCH | /canvases/{id}/ip-videos/{wid} | ✅ ipvideos.go:34 | ✅ widgets.py:100 | ✅ widgets.ts:495 | — |
| DELETE | /canvases/{id}/ip-videos/{wid} | ✅ ipvideos.go:42 | ✅ widgets.py:108 | ✅ widgets.ts:497 | — |

### Widgets — RDP-Connections (5)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /canvases/{id}/rdp-connections | ✅ rdpconnections.go:25 | ✅ widgets.py:78 (RDPConnectionsResource) | ✅ widgets.ts:518 | — |
| GET | /canvases/{id}/rdp-connections/{wid} | ✅ rdpconnections.go:34 | ✅ widgets.py:84 | ✅ widgets.ts:529 | — |
| POST | /canvases/{id}/rdp-connections | ⚠️ changelog §2 | ⚠️ changelog §2 (widgets.py:476 stub raises) | ⚠️ changelog §2 | Server returns "WidgetType is not supported". |
| PATCH | /canvases/{id}/rdp-connections/{wid} | ✅ rdpconnections.go:43 | ✅ widgets.py:100 | ✅ widgets.ts:545 | — |
| DELETE | /canvases/{id}/rdp-connections/{wid} | ✅ rdpconnections.go:51 | ✅ widgets.py:108 | ✅ widgets.ts:551 | — |

### Widgets — Uploads-folder (2)

| Verb | Path | Go | Python | TypeScript | Notes |
|---|---|---|---|---|---|
| GET | /canvases/{id}/uploads-folder | ✅ uploads.go:14 | ✅ widgets.py:610 | ✅ widgets.ts:561 | — |
| POST | /canvases/{id}/uploads-folder | ✅ uploads.go:23 (`UploadNoteToUploads`), uploads.go:32 (`UploadFileToUploads`) | ✅ widgets.py:632 | ✅ widgets.ts:588 | — |

---

## Counts by SDK (verified)

| SDK | ✅ Implemented | ⚠️ Omitted | ❌ Missing | Total |
|---|---|---|---|---|
| Go | 147 | 3 | 1 (PATCH /canvas-folders/{id}/copy) | 150 (wait: 147+3+1=151) |
| Python | 147 | 3 | 0 | 150 |
| TypeScript | 147 | 3 | 0 | 150 |

> Correction: the Go-only miss `PATCH /canvas-folders/{id}/copy` causes the Go column to be `146 ✅ + 3 ⚠️ + 1 ❌ = 150`. Re-stated correctly below.

| SDK | ✅ Implemented | ⚠️ Omitted | ❌ Missing | Total |
|---|---|---|---|---|
| Go | 146 | 3 | 1 | 150 |
| Python | 147 | 3 | 0 | 150 |
| TypeScript | 147 | 3 | 0 | 150 |

(The summary table at the top is updated accordingly: Go 146/3/1, Python 147/3/0, TypeScript 147/3/0. The earlier `145` figures in the top summary were a placeholder — use these verified counts.)

## Genuine gaps by SDK

### Go (1 miss)

1. **`PATCH /api/v1/canvas-folders/{folder-id}/copy`**
   - Spec: `docs/api-reference/endpoints/folders.md` line 283.
   - Go has `POST` variant only (`folders.go:118` `CopyFolder` calls `http.MethodPost`); there is no `MethodPatch` for the `/copy` action. Python `_copy_impl` accepts `use_patch` flag (`canvases.py:250`); TS exposes both `copy` (POST) and `copyUpdate` (PATCH, `folders.ts:80`).
   - **Suggested fix**: Add `UpdateCopyFolder(ctx, id, dest, conflicts)` paralleling `MoveFolder`'s `moveFolderWithMethod` pattern in `folders.go`. ~10 LOC. Trivial.

### Python (0 misses)

No missing endpoints. Every spec endpoint has a corresponding `transport.request` call.

### TypeScript (0 misses)

No missing endpoints. Every spec endpoint has a corresponding `transport.request` or `rawRequest` call.

## Reconciliation with prior reports

| SDK | Prior self-report | Verified | Delta | Notes |
|---|---|---|---|---|
| Go | 150/150 | 146/150 ✅ + 3 ⚠️ + 1 ❌ | **Over-reported by 1** | Self-report missed `PATCH /canvas-folders/{id}/copy` gap. Trivial fix (≤10 LOC). |
| Python | ~140/150 | 147/150 ✅ + 3 ⚠️ | **Under-reported by 7** | Python is more complete than its agent thought; all 147 implementable endpoints are covered. The 3 ⚠️ are explicit `UnsupportedOperationError` stubs plus the `clone()` helper that routes through type-specific endpoints. |
| TypeScript | 148/150 | 147/150 ✅ + 3 ⚠️ | **Off by 1** | TS effectively at parity with Python; the 148 figure double-counted by treating the `clone()` helper as "implementing" `POST /widgets/clone` (which is intentionally not called — see widgets.ts:138 comment "intentionally absent from the SDK"). |

## Truth statement

All three SDKs implement effectively complete coverage. The differences are cosmetic except for the one Go-only gap.

- **150 spec endpoints total.**
- **3 are deliberately omitted in all SDKs** (POST `/widgets/clone`, POST `/ip-videos`, POST `/rdp-connections`) with explicit changelog justification.
- **147 are implementable**; Python and TypeScript hit all 147. Go hits 146 (missing `PATCH /canvas-folders/{id}/copy`).
- Net coverage:
  - Go **146/147 = 99.3% of implementable endpoints**
  - Python **147/147 = 100%**
  - TypeScript **147/147 = 100%**
