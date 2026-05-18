# canvus-sdk migration notes

This document records every non-obvious decision taken while migrating
`CanvusPythonAPI` → `canvus_sdk` and applying the Phase 3 Python work items
and changelog fixes from `docs/api-reference/coverage-matrix.md`.

Written in one pass by the Phase 4b agent; intentionally terse.

## 1. Architecture decisions

### 1.1 Async surface is canonical; sync is a wrapper
- Hand-written `async def` methods for every endpoint.
- A `SyncClient` proxy (exposed as `Client.sync`) runs each coroutine on a dedicated worker-thread event loop (`_DedicatedLoop`).
- Rationale: per `docs/conventions/python.md` §Async, `asyncio.run()` is forbidden because it breaks inside Jupyter / FastAPI / any already-running loop. A worker-loop pattern is the only correct option.
- Trade-off: `SyncClient.close()` must be called explicitly when the sync surface was used, or the worker thread leaks until process exit. Documented in `__init__.py` docstring.

### 1.2 Pydantic v2 models (full), not TypedDict
- The legacy SDK already used Pydantic; the spec is rich enough (canvases, widgets, workspaces, audit log envelopes) that runtime validation pays for itself.
- `populate_by_name=True` and `extra="allow"` are set on the shared `CanvusModel` base so forward-compatible API additions never break deserialisation.
- Hyphenated wire keys are exposed via `alias=` rather than mangling Python attribute names — see `AuditLogPage.total_count` aliased to `total-count` and `RDPConnection.host_id` aliased to `host-id`.

### 1.3 `httpx` only; `aiohttp` removed
- `aiohttp` is a hard dependency only because the legacy SDK used it. `httpx` provides both `Client` (sync) and `AsyncClient` (async) with one timeout, retry, and auth model — exactly what we want for a single unified transport.
- The transport class `canvus_sdk._http.Transport` owns one long-lived `AsyncClient` per `Client`.

### 1.4 `structlog` only; no `print()` in library code
- The legacy SDK printed extensively (request URLs, headers, response bodies — including the redacted token). All of that is now `logger.debug` at most, and the API key is **never** logged.
- `configure_logging()` is opt-in. Callers that don't call it get whatever the structlog default produces (which is reasonable but uncoloured).

### 1.5 Single `CANVUS_` env prefix
- Legacy SDK used a mix of bare names (`API_URL`, etc.). `pydantic-settings` is configured with `env_prefix="CANVUS_"`, matching `docs/conventions/python.md` §Configuration.

### 1.6 Exception rename
- `CanvusAPIError` → `CanvusError` (base).
- New `APIError` subclass for HTTP-level errors specifically. `AuthError`, `NotFoundError`, `RateLimitError`, `ServerError` all derive from `APIError`.
- `ValidationError`, `TransportError`, `UnsupportedOperationError` derive directly from `CanvusError`.
- Rationale: the legacy hierarchy conflated "the server said no" with "the local payload was invalid" with "the socket died". The new split lets callers handle each meaningfully.

### 1.7 Resource grouping replaces god-class
- The legacy `CanvusClient` was 2,956 lines with ~150 methods on one class. The new `Client` is a thin attribute container; resources (`canvases`, `widgets`, `users`, ...) own their own methods. The widgets surface in particular nests further (`client.widgets.notes`, `client.widgets.tables`, ...).
- File sizes target the convention's "300 lines is a red flag" threshold.

## 2. API drift remediations applied (per coverage matrix Python items)

| # | Status | Notes |
|---|---|---|
| 1 | Done | `client.widgets.tables` — full CRUD + `list_cells`. Model omits `column_widths`/`row_heights` per changelog §4. `update()` emits `UserWarning` if caller passes `grid_size` per changelog §5. |
| 2 | Done | `client.widgets.clone(dest_canvas_id, source_canvas_id, source_widget_id, widget_type, location=None)` — hits the standard create endpoints per changelog §1. Supports notes/images/videos/pdfs/browsers/anchors/tables. |
| 3 | Done | `client.widgets.rdp_connections` — list / get / update / delete. `create()` raises `UnsupportedOperationError`. Model uses `host_id` etc. with hyphenated aliases pending changelog §3 verification. |
| 4 | Done | `client.widgets.ip_videos` — list / get / update / delete. `create()` raises `UnsupportedOperationError`. |
| 5 | Done | `client.widgets.video_inputs.get()` and `.update()` added. |
| 6 | Done | `client.widgets.list_uploads_folder()`. |
| 7 | Done | `client.users.change_email()`. |
| 8 | Done | `client.users.force_reset_password()` — admin force-reset, distinct from auth-flow self-reset. |
| 9 | Done | `client.groups.update()` (PATCH). |
| 10 | Done | `client.server.reload_certs()`. |
| 11 | Done | `client.server.activate_license()` — online activation, empty body. |
| 12 | Done | `client.server.open_canvas_in_workspace(client_id, workspace_id, canvas_id)`. |
| 13 | Done | `client.server.get_client_video_output()` and `.get_client_video_input()`. |
| 14 | Done | Legacy broken `update_video_output` removed; `client.server.set_video_output_source()` correctly targets `/clients/{cid}/video-outputs/{oid}`. |
| 15 | Done | `client.server.send_test_email(recipient_email)` — body shape `{recipient-email}`. |
| 16 | Done | `client.server.install_offline_license(license_data)` — body uses `license-data` key. |
| 17 | Done | `client.server.request_offline_activation()` — `?key=` query param removed. |
| 18 | Done | `client.server.get_audit_log(...)` — spec keys (`start-time`, `end-time`, `user-id`, `action`, `filter`, `page`, `per-page`) and `AuditLogPage` envelope. Falls back to flat-list response when server returns the legacy shape. |
| 19 | Done | `client.auth.create_token(user_id, *, name, expires=None, scopes=None)` — replaces legacy `description` kwarg. Returns `AccessTokenWithSecret` (with `plain_token`). |
| 20 | Done | `client.auth.login_saml(*, in_response_to, response_xml, remember=False)`. |
| 21 | Done | All `user_id` / `workspace_id` / `token_id` parameters are typed `str` across `UsersResource`, `AuthResource`, `ServerResource`. No `int` IDs anywhere. |
| 22 | Done | `Table`, `RDPConnection`, `IPVideo` modelled in `models/widgets.py`. |
| 23 | Done | No code path constructs or documents `/widgets/clone`; the README and `WidgetsResource.clone` docstring direct callers to the standard-create-endpoint pattern. |

## 3. Endpoints intentionally NOT migrated from CanvusPythonAPI

The legacy `canvus_api` package shipped several auxiliary modules. The migration focuses on the documented REST surface and leaves these for a future Phase:

- `geometry.py` — pure-Python widget bounding-box math. Useful but out-of-scope for the SDK surface; can ship as an `examples/` helper or a separate `canvus-geometry` package.
- `search.py`, `filters.py` — cross-canvas search and client-side filtering. Useful sugar but not in the API contract.
- `export.py` — widget batch import/export. Workflow helper; belongs in a separate `tools/` package.
- `widget_operations.py` — spatial-grouping / batch operations. Same rationale.
- The circular-parenting guard in `CanvusClient._check_circular_parenting` — useful but expensive (round-trips for every widget) and out of scope for a thin transport. Can be reimplemented as a helper.

These omissions are deliberate, not oversights. If callers need them, lifting the existing files into a `canvus-sdk-extras` workspace member is the cleanest path.

## 4. Unresolved / deferred (call out for Jaypaul's review)

### 4.1 RDP field naming (changelog §3) — still unverified
- The `RDPConnection` model uses snake_case Python attributes with hyphenated aliases (`host-id`, `content-id`, `connection-name`, `host-site`).
- If a future live-server check shows the API actually returns underscored variants, the aliases need to be flipped — but no Python attribute renames will be necessary.

### 4.2 `ServerConfig` wire shape
- The spec describes a flat `[]ConfigElement{key, value, type}` array; the legacy SDK and most observed servers return a nested dict.
- `ServerResource.get_config()` accepts both: if it sees a list it flattens it into the dict-shaped model. Update may emit warnings in unusual cases — call out if hit in integration testing.

### 4.3 Login body field (`username` vs `email`)
- Spec says `email`. Legacy server build accepts `username`. `AuthResource.login()` accepts both kwargs and sends whichever is supplied. Documented; not opinionated.

### 4.4 Streaming subscriptions
- `WidgetsResource.subscribe()` implements the documented `?subscribe=true` newline-delimited-JSON contract. Spec text (`docs/api-reference/streaming.md`) was scanned but not consulted in full. If the actual wire format is SSE rather than NDJSON, swap `stream_lines` for an SSE parser (`httpx-sse`).

### 4.5 Coverage of "missing" tests
- Tests cover client initialisation, URL normalisation, error classification, and a happy-path canvases.list smoke test via `respx`. Integration test scaffold is present but skipped without `CANVUS_API_KEY`. The 70% coverage target from the conventions will not be hit by these tests alone — Phase 4c needs to fill them out.

### 4.6 No retry-after honouring inside async sleep
- The current retry loop uses geometric backoff regardless of any `Retry-After` header. `RateLimitError.retry_after` is exposed to callers, but the transport doesn't sleep that long on its own. Low-priority improvement.

### 4.7 `Client.sync` thread leak surface
- Every call to `client.sync` lazily creates a dedicated thread. Calling `client.aclose()` shuts it down, but callers that abandon a Client without `aclose()` will leak the thread until process exit. Acceptable for scripts; flag if used in a long-lived service.

## 5. Post-verification fixes (2026-05-18)

Live-server verification against Canvus v1.2 revealed wire-format mismatches. The following surgical fixes were applied:

### 5.1 IP Video and RDP Connection field naming — HYBRID (hyphens for specific fields)
**Models affected:** `IPVideo`, `RDPConnection` in `models/widgets.py`

- **IPVideo**: Added `host_id` field with `alias="host-id"` (per live JSON). Other fields stay underscored.
- **RDPConnection**: Already had correct aliases (`host-id`, `connection-name`, `content-id`). Confirmed in use.
- **Verification:** Live GET on IP Video and RDP Connection endpoints return hyphenated keys only for these fields; all others use underscores.

### 5.2 Install license — wrong path and body field
**Method affected:** `ServerResource.install_offline_license()` in `resources/server.py`

- **Old:** `POST /license/install` with body `{"license-data": ...}`
- **New:** `POST /license` with body `{"license": ...}`
- **Verification:** Live `POST /api/v1/license/install` returns `{"msg": "Unknown action install"}`. Correct endpoint is `POST /api/v1/license` with `{"license": "..."}`.

### 5.3 License GET response — cleaned to live shape
**Model affected:** `LicenseInfo` in `models/server.py`

- **Removed:** `license_key`, `status`, `expiry_date`, `features`, `max_users`, `max_canvases` (not in live response).
- **Kept:** `edition`, `has_expired`, `is_valid`, `max_clients`, `type`.
- **Added:** `seat_model` (present in live response).
- **Verification:** Live `GET /api/v1/license` returns exactly `{edition, has_expired, is_valid, max_clients, seat_model, type}`.

### 5.4 Audit log entry — remodelled to live shape
**Model affected:** `AuditLogEntry` in `models/audit.py`

- **Old fields:** `id` (str), `timestamp`, `user_id` (str), `user_email`, `action`, `resource_type`, `resource_id`, `details` (dict), `ip_address`, `user_agent`.
- **New fields:** `id` (int), `action`, `author_id` (int|null), `created_at`, `details` (str — JSON-encoded), `ip_address`, `target_id` (str|null), `target_type`.
- **Verification:** Live `GET /api/v1/audit-log` returns flat array with exactly these 8 fields per entry; `details` is a string containing nested JSON, not a dict.

### 5.5 User ID type — reverted to int
**Parameters affected:** `get_audit_log(user_id)` and `export_audit_log_csv(user_id)` in `resources/server.py`

- **Old:** `user_id: str | None`
- **New:** `user_id: int | None`
- **Rationale:** Coverage matrix item #21 incorrectly specified UUID strings. The server uses integer user IDs (`author_id: 1000`). Reverted per live verification.

## 6. File / line summary

Approximate line counts (after final pass):

- `errors.py` ~90
- `_http.py` ~280
- `config.py` ~65
- `logging_config.py` ~60
- `client.py` ~225
- `models/*` ~520
- `resources/*` ~1,300
- `__init__.py` ~100
- `tests/*` ~250

Total: ~2,900 lines of Python in the SDK package + tests. Conservative versus the 3,000–4,500 target; the legacy SDK's 5,400-line monolith collapsed primarily because (a) per-method `print()` spam is gone, (b) duplicated CRUD bodies are factored into `_TypedSubResource` helpers, and (c) the search/filter/export modules were intentionally excluded (see §3).
