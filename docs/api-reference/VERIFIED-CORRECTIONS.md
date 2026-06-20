# Verified Corrections

This file catalogs the deltas between the as-extracted API spec (`endpoints/*.md`, `authentication.md`, `widget-types.md`, `streaming.md`) and the actual behaviour of the live Canvus server.

**Verification context:**
- Server: `dev-mtcs.multitaction.com`
- API version: `v1.2`
- Server-id: `19c6f03c-a3fc-4fa0-86a4-eacf4de92aea`
- `canvus-server` repo @ `4ed2d94fd822c40b87cf174f9861cd09175e9035`
- Verified: 2026-05-18

**Upstream tracking:** [`canvus-server#96`](https://gitlab.multitaction.com/swrd/conan/canvus/canvus-server/-/work_items/96) catalogs the documentation errors for the docs-of-record (wiki + `mt-restapi-client`). When that issue is resolved and the docs are re-extracted, this file should shrink or disappear.

The SDK source code in `go/sdk/`, `python/sdk/`, `typescript/sdk/` reflects the **verified truth**, not the original spec text.

---

## 1. `POST /api/v1/license` — path and field name

**Back-applied to:** `endpoints/server.md` ✅ (2026-05-18)

**Spec (`endpoints/server.md`):** `POST /api/v1/license/install` with body `{"key": "..."}` (or `{"license-data": "..."}` in some places).

**Verified truth:** `POST /api/v1/license` with body `{"license": "..."}`.

```
$ curl -X POST .../license/install -d '{"key":"PROBE"}'
{"msg":"Unknown action install"}

$ curl -X POST .../license -d '{"key":"PROBE"}'
{"msg":"license parameter is missing"}

$ curl -X POST .../license -d '{"license":"PROBE"}'   # would attempt install
```

---

## 2. `GET /api/v1/license` — response shape

**Back-applied to:** `endpoints/server.md` ✅ (2026-05-18)

**Spec:**
```json
{ "status": "valid", "clients": 5, "max-clients": 10, "valid": true,
  "expiry-date": "...", "seat-model": "fixed_seats", "activation-required": false }
```

**Verified:**
```json
{ "edition": "", "has_expired": false, "is_valid": true,
  "max_clients": -1, "seat_model": "fixed_seats", "type": "lifetime" }
```

Differences:
- Keys use **underscores**, not hyphens.
- `status` does not exist — replaced by `is_valid` + `has_expired`.
- `clients` (current count) does not appear in the response.
- `edition` and `type` are returned but were undocumented.
- `expiry-date` and `activation-required` may appear on certain license types; not present on lifetime licenses.

---

## 3. `GET /api/v1/server-config` — response shape

**Back-applied to:** `endpoints/server.md` ✅ (2026-05-18)

**Spec (`endpoints/server.md`):** "JSON array of configuration elements" with `{setting-key, setting-value, setting-type}` triples.

**Verified:** a deeply nested object:
```json
{
  "access": "rw",
  "authentication": {
    "domain_allow_list": ["*"],
    "password": { "enabled": true, "min_length": 8, ... },
    "saml": { "acs_url": "...", "enabled": true, ... }
  },
  "email": { "smtp_host": "", "smtp_port": 25, ... },
  "external_url": "https://dev-mtcs.multitaction.com",
  "server_name": "ISE/InfoComm"
}
```

Keys use underscores. The flat element-array form documented in the spec does not appear to be served by v1.2.

The `PATCH /api/v1/server-config` write API still uses the `{settings: [{setting-key, setting-value}, ...]}` element-array shape per the spec — only the GET response shape diverges.

---

## 4. `POST /api/v1/users/login` — strict field validation

**Back-applied to:** `endpoints/auth.md` ✅ (2026-05-18)

**Spec:** body is `{"email": "...", "password": "...", "remember": true}`.

**Verified:** the documented body works. **But** the server rejects requests containing **unknown** fields:

```
$ curl -X POST .../users/login -d '{"email":"x@y","password":"wrong"}'
{"msg":"Invalid username or password."}            # auth attempted as expected

$ curl -X POST .../users/login -d '{"email":"x@y","username":"x@y","password":"wrong"}'
{"msg":"Login request must have either email and password or token"}
```

SDKs that defensively double-key `email` and `username` will fail. Send only the documented fields.

**Also back-applied:** The login response `user` object field names. The spec showed `user-id` (UUID), `full-name`, `is-admin`, `is-blocked`. The verified wire shape uses `id` (integer), `name`, `admin`, `blocked` — all underscored.

---

## 5. `/api/v1/users/me` — does not exist

The wiki API reference describes `/users/me` as a convenience alias for the authenticated user. v1.2 returns:

```
$ curl -H "Private-Token: ..." .../users/me
{"msg":"User ID me is not an integer"}
```

`/users/{id}` is integer-only and `/me` is not handled.

---

## 6. User IDs are integers, not UUIDs

**Spec (e.g. login response):** `"user-id": "uuid"`.

**Verified:** integer IDs. Audit log returns `"author_id": 1000`, `"target_id": "1000"` (target_id is a JSON string but always-numeric).

SDK type definitions should use `int` / `number`, not `string` / `UUID`.

---

## 7. Field-naming convention is hybrid

The server uses a **mix** of hyphens and underscores in JSON keys, contrary to single-convention docs in places:

| Convention | Used for | Examples |
|---|---|---|
| `under_score` (majority) | most widget data, user fields, server config, audit log, license | `widget_type`, `parent_id`, `background_color`, `text_color`, `author_id`, `is_valid`, `max_clients`, `external_url` |
| `hyphen-ated` (specific) | IP Video + RDP widget-id fields, server-config write shape, some user fields | `host-id` (IPVideo & RDP), `connection-name` / `content-id` (RDP), `setting-key` / `setting-value` / `setting-type` (server-config PATCH), `is-admin` / `full-name` (some user contexts) |

SDKs must respect this hybrid. Type definitions for IP Video and RDP Connection use bracketed hyphenated keys in TypeScript and `Field(alias="host-id")` in Python.

---

## 8. `GET /api/v1/audit-log` returns a flat array

Some doc extracts describe the response as `{events: [...], total-count: N, page: 1, per-page: 50}`.

**Verified:** the server returns a flat JSON array of event objects:
```json
[
  {"action": "Login failed", "author_id": null, "created_at": "...", "details": "...",
   "id": 10533, "ip_address": "...", "target_id": null, "target_type": "user"},
  ...
]
```

`details` is a JSON-encoded string (not a nested object). No envelope, no pagination metadata.

---

## 9. Color presets path

**Spec:** `/api/v1/canvases/{id}/color-presets` (hyphen). ✅ **Verified correct.**

The path `colorpresets` (no hyphen, used by the pre-migration Go SDK) does NOT work:
```
$ curl .../canvases/.../colorpresets
{"msg":"Unknown object type colorpresets"}
```

---

## 10. Canvas background — nested shape, and the color type is `solid_color`

**Verified:** 2026-06-20 against `dev-mtcs.multitaction.com` (v1.2).

Some doc extracts describe a flat, hyphenated body: `{background-type, background-color, image-fit, grid-visible, grid-size}`. The live shape is **nested with underscore keys**:

```json
GET /api/v1/canvases/{id}/background
{
  "type": "solid_color",
  "background_color": "#25242dff",
  "grid":  {"color": "#525160ff", "visible": true},
  "haze":  {"color1": "#000000ff", "color2": "#165ad0ff", "scale": 8, "speed": 0.5},
  "image": {"fit": "fit", "hash": ""},
  "state": "normal",
  "widget_type": "CanvasBackground"
}
```

- `type` is one of `solid_color` | `haze` | `image` (not `color`). **`background_color` is an 8-digit RGBA hex** (`#rrggbbaa`).
- To set a colour background: `PATCH .../background {"type": "solid_color", "background_color": "#112233ff"}` → 200, persists.
- Sending `type: "color"` is rejected:

```
$ curl -X PATCH .../background -d '{"type":"color","background_color":"#112233ff"}'
{"msg":"Unsupported type: color"}
```

- `grid` / `haze` / `image` are nested objects, not the flat `grid-visible` / `image-fit` keys.

---

## 11. Table widget — `grid_size` is a nested `{columns, rows}` object

**Verified:** 2026-06-20 against `dev-mtcs` (v1.2).

Doc extracts (and changelog §4/§5) describe `grid_size` ambiguously and list `column_widths`/`row_heights`. The live table serializes `grid_size` as a **nested object** and omits the width/height arrays:

```json
GET /api/v1/canvases/{id}/tables/{id}
{
  "title": "",
  "grid_size": {"columns": 3, "rows": 3},
  "size": {"width": 900, "height": 600},
  "location": {"x": …, "y": …},
  "widget_type": "Table"
}
```

- Create: `POST .../tables {"title": "...", "grid_size": {"columns": C, "rows": R}}` → 200. Do **not** send a scalar `grid_size`, nor `column_widths`/`row_heights` (server does not serialize them — confirms changelog §4).

---

## 12. Client video-inputs — response shape

**Verified:** 2026-06-20 against `dev-mtcs` (v1.2).

```json
GET /api/v1/clients/{id}/video-inputs
[
  {
    "name": "video=Video (… Capture …):audio=Audio (… Capture …)",
    "resolution": {"width": 1920, "height": 1080},
    "source": "video=@device_pnp_\\\\?\\pci#ven_…#…\\video:audio=@device_cm_…"
  }
]
```

- Keys: `name`, `resolution` (nested `{width, height}`), `source`. The `source` string is the device identifier set as a video-**output**'s `source` to route that input to that output (`PATCH /clients/{id}/video-outputs/{idx} {"source": "<input.source>"}`).

---

## 13. Cross-canvas widget clone — confirmed working (validates changelog §1)

**Verified:** 2026-06-20 against `dev-mtcs` (v1.2).

`POST /api/v1/canvases/{destId}/{type}` with body `{source_canvas_id, source_widget_id, location?}` clones the source widget into the destination canvas:

```
POST /canvases/{dest}/notes {"source_canvas_id": <src>, "source_widget_id": <wid>, "location": {"x":123,"y":456}}
→ 200; new widget has a regenerated id, the source's text/fields carried over, and `location` honored exactly.
```

Confirms changelog §1 against the live server. The deprecated `POST /canvases/{id}/widgets/clone` (501) is not needed. Note: writes to a canvas that is currently **open on a live client** can block/hang on the server side; this is per-canvas, not a server-wide write outage — fresh/unopened canvases write instantly.

---

## Not yet verified (deferred)

- `PATCH /api/v1/server-config` write shape (read shape diverged; write shape may also).
- Streaming wire format under load (NDJSON confirmed, but keepalive frame exact byte sequence not yet checked).
- File upload multipart shapes.
- Group membership endpoints.
- Folder hierarchy edge cases.

These can be verified opportunistically in Phase 4 example apps.
