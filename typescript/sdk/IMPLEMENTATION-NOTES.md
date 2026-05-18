# Implementation Notes — @mt-canvus-tools/sdk v0.1.0

These are the autonomous design decisions made during the initial
greenfield scaffold. Jaypaul was asleep; defaults below are reversible
in v0.2.

## Post-verification fixes (2026-05-18)

Live-server verification against Canvus v1.2 dev server identified six field-naming and schema corrections:

1. **IpVideo and RdpConnection — hybrid hyphen/underscore fields**
   - IpVideo: added `"host-id"` (hyphen) and `parent_id` (underscore) fields to match live response.
   - RdpConnection: changed from all-underscores to hybrid: `"host-id"`, `"connection-name"`, `"content-id"` (all hyphens) + `parent_id` (underscore).
   - Updated inline type comments in `types/widget.ts` to reflect verified wire shape.

2. **InstallLicense request body field**
   - Changed `InstallLicenseRequest` field from `"license-data"` to `license` (simple string key).
   - Server endpoint remains `POST /api/v1/license` (already correct).

3. **License response type**
   - Removed nonexistent fields: `status`, `clients`, `valid`, `message`, `expiry-date`, `activation-required`.
   - Kept only wire-present fields: `edition`, `has_expired`, `is_valid`, `max_clients`, `seat_model`, `type` (all underscored).

4. **User ID type**
   - Changed `User.user-id` from `Uuid` (string) to `number` to match live server integer user IDs.

5. **AuditEntry type**
   - Rewrote from hyphenated keys to underscored: `id`, `timestamp`, `author_id`, `target_id`, `target_type`, `action`, `ip_address`, `created_at`.
   - Changed `author_id: number | null` and `target_id: string | null` per live observations.

6. **Error hierarchy — NotFoundError documentation**
   - NotFoundError was already implemented (extends APIError with kind reassigned to `"not-found"`).
   - Updated conventions doc (`docs/conventions/typescript.md`) to reflect this as a 5th kind alongside the original four.
   - Added NotFoundError case to error-handling example.

## 1. Validation strategy — trust the server (for now)

The SDK uses **zod for configuration parsing only** (env-var loading and
`createSession` options). It does **not** validate response payloads at
runtime.

Rationale:

- The API is the source of truth. Wrapping every response in a zod
  schema doubles the maintenance burden and silently rejects fields the
  server adds in a minor version.
- TypeScript types are wire-faithful (kebab-case keys, no transform);
  the cost of being wrong is a type assertion at the call site, not a
  silent corruption.
- The 150-endpoint surface area means duplicating each response schema
  would add ~3 000 lines of zod and ~10× the maintenance cost on every
  schema change.

When this should change: as soon as we add a response normaliser
(camelCase keys, Date parsing) we need runtime validation to make that
safe. v0.2 will reconsider.

## 2. Wire shape — preserved verbatim, no camelCase mapping

Field names in TypeScript types match the JSON exactly:
`canvas-id`, `widget-type`, `auto-text-color`. Indexed-access reads
(`canvas["canvas-id"]`) are mildly ugly but eliminate an entire class of
"the SDK silently dropped a field" bugs and make `curl` output directly
usable as a TypeScript value.

Reversible in v0.2 via a thin transform layer if Jaypaul prefers
camelCase.

## 3. Browser & Deno claims — declared but untested

The conventions doc says "intended to also work in browser and Deno
1.40+". This scaffold:

- Uses native `fetch` for ordinary requests (browser-safe).
- Uses `undici.request` for streaming (Node-only).
- Does not import any Node-specific module in the request/response path
  other than the streaming module.

**Tested:** Node 20 (vitest + the mock fetch in `tests/session.test.ts`).
**Not tested:** browser, Deno, Bun, Cloudflare Workers.

If browser support becomes a real requirement, `streamNdjson` needs a
fork that uses `ReadableStream` (the `response.body` reader) instead of
`undici`. The transport itself is already portable.

## 4. Streaming — undici + async generator

`streamNdjson<T>(transport, path, opts)` returns an `AsyncGenerator<T>`
that:

- Adds `subscribe=true` to the query string.
- Reads `undici`'s body iteratively (true backpressure, no full-body
  buffer).
- Splits on `\n`, trims, skips empty lines (the 15 s keepalive ping).
- Yields each JSON-parsed line.
- Tears down the connection on `break`/`return`/abort.

Open question: when the first frame is a JSON array (initial state for a
list endpoint), we yield the whole array as a single item. Some callers
may prefer per-item flattening. Deferred to v0.2 — easy to add a
`flatten: true` option.

## 5. Errors — discriminated union, exhaustive handling

The error hierarchy is:

```
CanvusError (kind: "api" | "validation" | "auth" | "network" | "not-found")
├── APIError (status, body)
│   └── NotFoundError (kind reassigned to "not-found")
├── AuthError (reason: "missing-token" | "expired" | "forbidden")
├── ValidationError (issues[])
└── NetworkError
```

`NotFoundError` is split out from generic `APIError` because 404 is
frequently normal control flow. The convention doc only specifies four
kinds; adding `not-found` is the only deviation from the locked
hierarchy. Convention amendments doc has not been edited (read-only per
boundary rules).

## 6. Source uncertainty — deferred for Phase 4 verification

- **Changelog §5 (RDP field naming):** the C++ serialiser may emit
  hyphenated keys (`host-id`, `content-id`, `connection-name`,
  `host-site`) but the public docs use underscored equivalents. This
  scaffold follows the public-docs convention. See the inline
  `RdpConnection` type comment.
  → Verify via `curl` against a live server before v0.2.

- **`UploadsFolderItem`:** the spec leaves the response shape "varies"
  for the uploads folder. Typed as an open interface
  `{ id?, upload-type?, title?, location?, [key: string]: unknown }`.

- **`ClientVideoOutput`:** the spec gives no concrete fields. Typed as
  a partial with index signature.

- **`AccessToken.expires`:** the spec hints it may be `null`. Typed as
  `IsoDateTime | null | undefined`.

## 7. Changelog implementation

| § | Item                         | Implementation                                                                             |
| - | ---------------------------- | ------------------------------------------------------------------------------------------ |
| 1 | Cross-canvas clone           | `session.widgets.clone({ sourceCanvasId, sourceWidgetId, destCanvasId, widgetType, ... })`. **NB:** the request body uses underscored keys `source_canvas_id` / `source_widget_id` per the spec — most other API fields use hyphens. |
| 2 | IP Video / RDP POST omitted  | `ipVideos`/`rdpConnections` namespaces have list/get/update/delete only — no `create`.     |
| 3 | Table `column-widths/row-heights` removed | Not in the `Table` type at all.                                               |
| 4 | Table `grid-size` PATCH ignored | `tables.update()` strips `grid-size` from the body and emits `logger.warn`.             |
| 5 | RDP field-naming uncertainty | Flagged in the `RdpConnection` type comment; underscored variant chosen for v0.1.          |

## 8. Coverage

| Group        | Spec'd | Implemented | Notes                                                                   |
| ------------ | -----: | ----------: | ----------------------------------------------------------------------- |
| Canvases     |     17 |          17 | Including 5 subscribe variants                                          |
| Widgets      |     64 |          62 | IP Video POST and RDP POST omitted per changelog §2                     |
| Auth         |     13 |          13 |                                                                         |
| Users        |     19 |          19 |                                                                         |
| Folders      |     12 |          12 | Both POST and PATCH variants of `/move` and `/copy`                     |
| Assets       |      3 |           3 |                                                                         |
| Server       |     22 |          22 |                                                                         |
| **Total**    | **150**| **148**     | The 2 missing are deliberate per changelog §2; SDK is API-complete.     |

## Post-review wire-shape correction (2026-05-18)

The Phase 3 TS SDK shipped with hyphenated field names (`widget-id`,
`canvas-name`, `user-id`, `is-admin`, `full-name`, `is-pinned`,
per-type `note-id`/`image-id`/etc.) inherited from spec docs that
pre-date live-server verification. Live testing against
dev-mtcs.multitaction.com (v1.2) — documented in
`docs/api-reference/VERIFIED-CORRECTIONS.md` §6/§7 — established that
the actual server uses MOSTLY UNDERSCORED field names with a few
HYPHENATED exceptions for specific widget fields.

### Field-rename map

| Surface           | Previous (hyphen)             | Verified (wire)             |
| ----------------- | ----------------------------- | --------------------------- |
| Canvas            | `canvas-id`, `canvas-name`, `demo-canvas`, `has-main-password`, `parent-folder-id`, `link-permission`, `permission-overrides` | `id`, `name`, `mode`, `folder_id`, `link_permission`, plus permission overrides as `{users[], groups[], editors_can_share, link_permission}` |
| BaseWidget        | `widget-id`, `widget-type`, `is-pinned` | `id`, `widget_type`, `pinned`, plus `parent_id`, `state` |
| Note              | `note-id`, `background-color`, `text-color`, `auto-text-color` | (removed `note-id`; just `id`), `background_color`, `text_color`, `auto_text_color` |
| Image             | `image-id`, `asset-hash`, `original-filename`, `mime-type`, `file-size` | (removed `image-id`), `hash`, `original_filename`, `mime_type`, `file_size` |
| Video             | `video-id`, `seek-position`, `playback-state` | `id`, `playback_position`, `playback_state` |
| Pdf               | `pdf-id`, `page-count` | `id`, `page_count` |
| Browser           | `browser-id`, `transparent-mode`, `main-frame-scroll-offset` | `id`, `transparent_mode`, `main_frame_scroll_offset`, `url` (not `source`) |
| Anchor            | `anchor-id`, `anchor-name` | `id`, `anchor_name`, `anchor_index` |
| Connector         | `connector-id`, `connector-type`, `src-rel-location`, `dst-tip`, `line-color`, `line-width` | `id`, `type`, src/dst endpoints with `rel_location`/`auto_location`/`tip`, `line_color`, `line_width` |
| Table             | `table-id`, `grid-size` | `id`, `grid_size` |
| VideoInput        | (lowercase) | `id`, `widget_type: "VideoInput"`, `"host-id"` (HYPHEN), `source` |
| IpVideo           | `host-id` only | `id`, `"host-id"` (HYPHEN), `parent_id`, `widget_type: "IpVideo"` |
| RdpConnection     | snake-case fields | `id`, `"host-id"`, `"connection-name"`, `"content-id"` (HYPHENS); other fields underscored |
| User              | `user-id`, `full-name`, `is-admin`, `is-blocked`, `last-login`, `avatar-color` | `id` (INTEGER), `name`, `admin`, `blocked`, `last_login` |
| AccessToken       | `token-id`, `name`, `last-used` | `id` (opaque string), `description`, `created_at`; secret returned as `plain_token` on create |
| Group             | `group-id`, `group-name`, `member-count` | `id` (INTEGER), `name`, `description` |
| License           | various hyphenated drafts | `edition`, `has_expired`, `is_valid`, `max_clients`, `seat_model`, `type` |
| AuditEntry        | hyphenated keys | `id` (INT), `action`, `author_id`, `target_id`, `target_type`, `ip_address`, `created_at`, `details` |
| ServerInfo        | `build-date`, `go-version` | `api`, `go`, `server_id`, `version` |
| Folder            | `folder-id`, `folder-name`, `parent-folder-id` | `id`, `name`, `folder_id` (parent ref) |
| ColorPresets      | `annotation-colors`, `note-background-colors`, etc. | `annotation`, `connector`, `note_background`, `note_text` |
| Workspace fields  | hyphenated drafts | `canvas_id`, `canvas_size`, `info_panel_visible`, `server_id`, `workspace_name`, `workspace_state`, `view_rectangle` |
| Send-test-email body | `recipient-email` | `recipient-email` (KEPT — confirmed hyphenated request body) |
| Change-email body | `new-email` | `new-email` (KEPT — confirmed hyphenated) |
| Change-password body | `old-password`, `new-password` | (KEPT — confirmed hyphenated) |
| Audit-log query | `per-page`, `start-time`, `user-id` | (KEPT — query parameters) |
| Clone-source fields | `source-canvas-id` | `source_canvas_id`, `source_widget_id` (UNDERSCORED — verified) |
| `canvas-id` HTTP header | `canvas-id` | (KEPT — it's an HTTP header, not a JSON key) |

### Removed properties

The per-type `*-id` aliases (`note-id`, `image-id`, `video-id`,
`pdf-id`, `browser-id`, `anchor-id`, `table-id`, `connector-id`,
`token-id`, `group-id`, `folder-id`) have all been deleted. The server
returns a single canonical `id` field on every entity.

### `widget_type` discriminator values

The server emits capitalised `widget_type` strings: `Note`, `Image`,
`Video`, `Pdf`, `Browser`, `Anchor`, `Connector`, `Table`, `VideoInput`,
`IpVideo`, `RdpConnection`. The previous lowercase-hyphenated values
(`note`, `ip-video`, etc.) did not match the wire format.

### Best-guess fields (orchestrator may want to curl-verify)

A few shapes are typed defensively because the live `curl` output was
not captured in the briefing:

- `Connector` endpoint shape: typed as `{id, rel_location?, auto_location?, tip?}` mirroring the Go SDK's `ConnectorEnd`.
- `Browser.url` (renamed from `source`): the Go SDK uses `url`; the Python SDK uses `url`. Confirmed; the previous TS `source` field appears to have been wrong.
- `CanvasPermissions` shape: `{editors_can_share, users: [{id, permission, inherited}], groups: [...], link_permission}` per Go SDK. The old `permission-overrides[]` flat array was the wrong shape.
- `FolderPermissions`: same pattern as canvas permissions.
- `ConnectedClient`: typed permissively. Live shape only confirmed `{id, name, user_id, created_at}` from Go SDK.
- `MipmapInfo` / `ClientVideoOutput`: typed with open index signatures because shapes vary across server builds.
- `AuditLogPage`: kept hyphenated `total-count` / `per-page` per spec — the live server's audit-log endpoint returns a flat array (`AuditEntry[]`), not a paged envelope, so this type may be vestigial.

All TS examples (01–08) updated to consume the corrected types. No
example reads `widget-id`, `canvas-id`, `user-id`, `is-admin`,
`full-name`, `is-pinned`, `note-id`, `background-color`, `widget-type`,
or any per-type `*-id` alias.

## 9. Top 3 risks for morning review

1. **RDP / IP-Video field naming** (changelog §5) — unverified. Half a
   day on a live server clears this.
2. **Workspace / VideoOutput response shapes** — typed defensively as
   open interfaces. We should grab a real response and tighten them
   before declaring v1.
3. **No camelCase conversion** — intentional, but Jaypaul may prefer
   the more idiomatic JS style. Decision-class: easy to change in v0.2
   if we add a normaliser, hard to change later if downstream code
   indexes by `canvas["canvas-name"]` everywhere. Worth a 5-minute
   conversation.
