# API Changelog

> **2026-05-18 verification addendum:** Live-server testing against `dev-mtcs.multitaction.com` (v1.2) discovered additional doc errors beyond what's catalogued here. See **[`VERIFIED-CORRECTIONS.md`](VERIFIED-CORRECTIONS.md)** for the full delta against the as-extracted spec. The upstream docs fix is tracked at [`canvus-server#96`](https://gitlab.multitaction.com/swrd/conan/canvus/canvus-server/-/work_items/96).

> Pending documentation updates extracted from `mt-restapi-client/doc-updates-for-developer-site.md` as of 2026-05-17.
> Each entry describes a change from the current public developer documentation that downstream SDKs must reflect.

## 2026-09-20 — Go SDK request safety (issue #6)

Go actor authentication now uses one selected credential, retains actor identity after 401, and does not restore service authority after explicit logout. Config/client values are copied; raw `HTTPClient` calls no longer inject the configured API key. Default redirects are disabled unless the caller supplies a policy. TLS defaults remain unchanged.

Zero retries is respected; only bodyless GET/HEAD requests can be automatically retried, with cancellable waits. All mutation retries are disabled, including nested BatchProcessor retries; batch retry options are deprecated and ignored. Callers needing three read retries should use `DefaultSessionConfig()`.

Accepted writes no longer require request/response echo equality. Additive Go `AcceptedResponseError` preserves 2xx status and available IDs on read/decode failure, unwrapping the underlying cause without retaining raw content. See the Go SDK README for outcome/reconciliation guidance.

**Explicit parity status:** this change is scoped to Go. Python already supports a zero retry budget and independent clients but does not adopt this Go error type; Python and TypeScript method-aware retry/outcome parity is deferred, and this release makes no claim that their mutation retries are safe. Their existing auth/login behavior is unchanged. Streaming lifecycle, current-user/SAML, geometry and MCP hardening remain separate work. No REST endpoints or existing Go resource signatures changed.

## Updates

### Widget Clone — Cross-Canvas Copy via Standard Create Endpoints

**Type:** changed behaviour (documented feature, not "501 Not Implemented")
**Endpoint(s) affected:** `POST /api/v1/canvases/{destCanvasId}/notes`, `POST /api/v1/canvases/{destCanvasId}/images`, `POST /api/v1/canvases/{destCanvasId}/videos`, `POST /api/v1/canvases/{destCanvasId}/pdfs`, `POST /api/v1/canvases/{destCanvasId}/browsers`, `POST /api/v1/canvases/{destCanvasId}/anchors`, `POST /api/v1/canvases/{destCanvasId}/tables`
**Source location in doc-updates note:** Section 1, lines 9–55

**Summary:** Widget cloning works via standard widget creation endpoints when request body includes `source_canvas_id` and `source_widget_id`; docs incorrectly list a separate 501 Not Implemented endpoint.

**Detail:** Cross-canvas widget cloning is implemented as a feature of the standard `POST /canvases/{canvasId}/{widgetType}` create endpoints. When the request body includes `source_canvas_id` and `source_widget_id`, the server clones the widget from the source canvas into the destination canvas instead of creating a new widget from scratch. The caller must have edit access to the destination canvas and at least view access to the source canvas. For asset-based widgets (Image, Video, PDF), the source asset file must be available and is copied to the destination. The `id`, `parent_id`, `state`, and `widget_type` fields are regenerated; other fields carry over. Optional `location` parameter overrides the source widget's position on the destination canvas.

**Action required for SDKs:**
- Go SDK: Add `CloneWidget()` helper method that accepts `sourceCanvasID`, `sourceWidgetID`, optional `location`, and widget type; internally calls the appropriate standard create endpoint with these parameters in the request body.
- Python SDK: Same pattern as Go.
- TypeScript SDK: Same pattern as Go.
- Remove or deprecate any stub code for `POST /canvases/{id}/widgets/clone` endpoint.

---

### IP Video and RDP Connection — POST Endpoint Not Actually Supported

**Type:** changed behaviour (documented endpoints that don't work)
**Endpoint(s) affected:** `POST /api/v1/canvases/{id}/ip-videos`, `POST /api/v1/canvases/{id}/rdp-connections`
**Source location in doc-updates note:** Section 2, lines 58–87

**Summary:** API docs list POST (create) operations for IP Video and RDP Connection widget types, but the server rejects these requests with "WidgetType is not supported" error.

**Detail:** The `createElement` whitelist in `WidgetTreeModel.cpp` does not include `sIpVideoType` or `sRdpConnectionType`. Attempting POST on these endpoints returns `{"error": "WidgetType IpVideo is not supported"}` or equivalent for RDP. These widget types can only be created from the Canvus desktop client, similar to VideoOutputAnchor.

**Action required for SDKs:**
- Go/Python/TypeScript SDKs: Remove POST/create operations for IP Video and RDP Connection widget types from public APIs. Support only GET (read), PATCH (update), and DELETE operations.
- Update API documentation tables to remove the POST row for `/canvases/{id}/ip-videos` and `/canvases/{id}/rdp-connections`.
- If SDKs expose these as widget types, mark creation as unsupported with clear error messaging: "IP Video and RDP Connection widgets can only be created from the Canvus desktop client."

---

### Table Widget — Missing `column_widths` and `row_heights` in Serialization

**Type:** changed behaviour (documented fields not actually returned)
**Endpoint(s) affected:** `GET /api/v1/canvases/{id}/tables/{widgetId}`, `PATCH /api/v1/canvases/{id}/tables/{widgetId}`
**Source location in doc-updates note:** Section 4, lines 116–131

**Summary:** Documentation lists `column_widths` and `row_heights` as table-specific response fields (read-only), but `serializeTableProperties()` does not serialize these; they are never returned.

**Detail:** Current docs claim table objects include `column_widths` (array, read-only) and `row_heights` (array, read-only). However, the server's serialization only includes `title` and `grid_size`. The fields would logically be computed from cell sizes if implemented.

**Action required for SDKs:**
- Go/Python/TypeScript SDKs: Remove `column_widths` and `row_heights` from table widget type definitions and documentation.
- If SDKs have already exposed these as nullable or optional fields, deprecate them with a notice: "These fields are not currently returned by the API."
- Update response type definitions for Table widgets to only include `title` and `grid_size`.

---

### Field Naming Convention — Hyphens vs Underscores in RDP Connections

**Type:** clarification (needs verification before action)
**Endpoint(s) affected:** `GET /api/v1/canvases/{id}/rdp-connections`, `PATCH /api/v1/canvases/{id}/rdp-connections/{widgetId}`
**Source location in doc-updates note:** Section 3, lines 89–114

**Summary:** C++ serialization code uses hyphens (e.g., `"host-id"`) for certain RDP Connection fields, but API docs use underscores (e.g., `host_id`). Actual API response format unclear.

**Detail:** The C++ code serializes RDP fields with hyphens: `"host-id"`, `"content-id"`, `"connection-name"`, `"host-site"`. Current documentation lists underscores: `host_id`, `content_id`, `connection_name`. The Go frontend layer may or may not translate between conventions; this requires verification against live API responses.

**Action required for SDKs:**
- **First:** Verify live API response format with: `curl -s -H "Private-Token: <token>" https://<server>/api/v1/canvases/<id>/rdp-connections | python3 -m json.tool`
- If API actually returns hyphens: update docs and SDKs to use hyphens (`host-id`, `content-id`, `connection-name`, `host-site`).
- If API actually returns underscores: no SDK change needed; confirm C++ code path used by API server.
- Add explicit field naming reference in SDK documentation to prevent future mismatch.

---

### Table `grid_size` PATCH Behavior — Silent Ignore Not Documented

**Type:** clarification (behavior exists but not documented)
**Endpoint(s) affected:** `PATCH /api/v1/canvases/{id}/tables/{widgetId}`
**Source location in doc-updates note:** Section 5, lines 134–147

**Summary:** PATCH requests that include `grid_size` are silently ignored rather than returning an error. This behavior should be documented or stricter error handling implemented.

**Detail:** The API accepts PATCH requests on table widgets that include `grid_size`, but the field is silently ignored; no error is returned. Current docs state grid size is not modifiable post-creation, but don't clarify the silent-ignore behavior.

**Action required for SDKs:**
- Go/Python/TypeScript SDKs: Update documentation and type hints to clarify: "`grid_size` is set at creation time and cannot be changed. Including `grid_size` in a PATCH request will be silently ignored."
- In validation/helper methods, consider warning or filtering out `grid_size` from PATCH requests with a message: "grid_size cannot be modified after creation and will be ignored."
- Alternatively, if stricter API behavior is preferred, file code bug to return 400 Bad Request when `grid_size` is included in PATCH body.

---

## Summary

**Total updates extracted:** 5

**By type:**
- Changed behaviour: 3 (widget clone, IP Video/RDP POST, table columns/rows)
- Clarification: 2 (field naming, grid_size PATCH)

**Classification notes:**
- Section 1 (Widget Clone): Classified as "changed behaviour" because the feature is implemented but docs are wrong; SDKs should expose a convenient wrapper.
- Section 2 (IP Video/RDP POST): Classified as "changed behaviour" (documentation doesn't match implementation); removal action is required.
- Section 3 (Field Naming): Classified as "clarification" (pending verification); this is a documentation-first concern that may require SDK adjustments once verified.
- Section 4 (Table Fields): Classified as "changed behaviour" (undocumented fields that don't exist); SDK type definitions need cleanup.
- Section 5 (Grid Size PATCH): Classified as "clarification" (behavior exists, just not documented); SDK documentation needs update.
