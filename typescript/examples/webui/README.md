# webui — admin/operator UI built on `@mt-canvus-tools/sdk`

A Hono/TypeScript rewrite of the legacy [CanvusWebUI](https://github.com/jaypaulb/CanvusWebUI)
Express demo. The application is preserved feature-for-feature; the
underlying server is rewritten to use the new TypeScript SDK and a
modern HTTP stack.

## Purpose

This example demonstrates a non-trivial workload on top of the TS SDK:

- multi-user identity + per-team color assignment
- file upload (notes, images, videos, PDFs) anchored to "team targets"
- zone-based macros (move / copy / delete / auto-grid /
  group-by-color / group-by-title / pin-all / unpin-all /
  export / import)
- page-sequence + workspace tracking with **typed SDK subscribe**
  instead of the legacy 2-second polling loop
- deleted-record audit + undelete buffer
- admin panel for env-var introspection and runtime canvas switching

## Build & run

```bash
# from typescript/
pnpm install
pnpm --filter '@mt-canvus-tools/webui-example' run typecheck
pnpm --filter '@mt-canvus-tools/webui-example' run build
pnpm --filter '@mt-canvus-tools/webui-example' run test

# dev (watch)
pnpm --filter '@mt-canvus-tools/webui-example' run dev

# prod
pnpm --filter '@mt-canvus-tools/webui-example' run start
```

The legacy Docker Compose files (`docker-compose.yml`, `Dockerfile`)
from the source repo are not ported — the example targets local
operation. If a container deployment is needed, build a slim
Node 20+ image with `dist/` and `public/` and expose `PORT`.

## Environment variables

| Variable | Required | Default | Notes |
| --- | --- | --- | --- |
| `CANVUS_API_URL` | yes | — | Canvus REST base URL (e.g. `https://canvus.example.com/api/v1/`). Falls back to `CANVUS_SERVER`. |
| `CANVUS_API_KEY` | yes | — | Private-Token for the Canvus server. |
| `CANVUS_CANVAS_ID` | yes | — | Default canvas ID. Falls back to `CANVAS_ID`. |
| `CANVUS_CANVAS_NAME` | no | — | Display name for `/get-canvas-info`. Falls back to `CANVAS_NAME`. |
| `WEBUI_PWD` | recommended | — | Bearer-token shared secret for admin endpoints. If unset, admin routes return 500. |
| `PORT` | no | `3000` | TCP port. |
| `HOSTNAME` | no | `0.0.0.0` | Bind address. |
| `ALLOW_SELF_SIGNED_CERTS` | no | `false` | Set to `true` to skip TLS verification when calling Canvus. Demo-only. |
| `SSL_CERT_PATH` / `SSL_KEY_PATH` | no | — | If both are set, the server starts in HTTPS mode. |
| `COOKIE_SECRET` | no | `webui-demo-cookie-secret` | Reserved for future cookie-session use. |
| `LOG_LEVEL` | no | `info` | Pino level. |
| `LOG_FORMAT` | no | `pretty` | `pretty` (dev) or `json` (container). |
| `UPLOAD_DIR` | no | `uploads` | Reserved — uploads stream straight to Canvus without local disk in this rewrite. |
| `STATE_DIR` | no | `state` | Where the demo JSON stores live (`users.json`, `macros-deleted-records.json`). |

## Migration notes — Express/JS → Hono/TS

| Concern | Legacy (Express + JS) | This port (Hono + TS) |
| --- | --- | --- |
| HTTP framework | Express 4 + `body-parser` + `multer` + `express-validator` | Hono 4 + `@hono/node-server` + `@hono/zod-validator` |
| Canvus calls | hand-rolled `axios` client with `Private-Token` header | `@mt-canvus-tools/sdk` `createSession`, type-safe per-resource methods |
| Workspace updates | 2-second polling loop in `/api/clients/:id/workspace/subscribe` | typed `session.server.subscribeWorkspace(clientId, "0")` SSE relay |
| Widget PATCH dispatch | per-type if-chain in `getWidgetPatchURL` | `session.widgets.updateAny(widget_type, ...)` (centralised) |
| Body validation | `express-validator` (error array) | `zod` + `@hono/zod-validator` (typed schemas) |
| Persistence | flat JSON files in cwd (`users.json`, `macros-deleted-records.json`) | same files under `${STATE_DIR}/` (demo-grade — not for production) |
| Env mutation | rewrites `../.env` on disk via `/admin/update-env` | runtime-only mutable canvas swap; env file is immutable from the app |
| Server.js LOC | 3,797 (single file) | ~1,800 split across 8 route modules + 4 lib/helper modules |

### RCU endpoints clarification

The legacy CanvusWebUI shipped an `RCU.html` page that turned out to
mean "Remote Content Upload" — a user-facing upload form. There were
no `/api/v1/canvases/{id}/rcu/*` endpoints in the legacy server. The
legacy upload form has been dropped from `public/` in this rewrite —
use `/upload-item` (POST, multipart) with team selection for the same
user-facing flow.

The PowerToys Go port (Round 3) has client-side code that *does* call
`/api/v1/canvases/{id}/rcu/{config,status,test}`. Per Jaypaul's
2026-05-18 clarification these are NOT canvus-server endpoints; they
are this WebUI's own admin surface, kept here so PowerToys' embedded
client can call into it. The implementation is in `src/routes/rcu.ts`,
gated by the same Bearer token as `/admin/*`, and currently:

- `GET /api/v1/canvases/:id/rcu/config` — returns the in-memory config
  (no password echoed)
- `POST /api/v1/canvases/:id/rcu/config` — stores config, accepts
  `enabled/endpoint/topic/qos/username/password`
- `GET /api/v1/canvases/:id/rcu/status` — reports enabled-flag + last
  test result
- `POST /api/v1/canvases/:id/rcu/test` — probes the configured Canvus
  server via `session.server.info()`

The `:id` path parameter is honoured by template but ignored at runtime
— the WebUI operates on `mutableConfig.canvasId`. This matches what
PowerToys constructs.

## Per-feature walkthrough

### User identity (`src/routes/users.ts`)
`POST /identify-user` assigns each user a HSL color variation of their
team's base hex (`getTeamBaseColor`). The mapping persists to
`${STATE_DIR}/users.json`. Admin can wipe via `/admin/deleteUsers`.

### Uploads (`src/routes/uploads.ts`)
`POST /create-note` creates a sticky note via
`session.widgets.notes.create` near the user's `Team_N_Target` widget
(±300 px random offset). `POST /upload-item` accepts multipart bodies
and dispatches to `session.widgets.{images,videos,pdfs}.upload` based
on extension.

### Zones (`src/routes/zones.ts`)
Zones are Canvus `Anchor` widgets. `POST /create-zones` builds N×N
grids in Z, Snake, or Spiral patterns, or sub-zones inside a parent.
Script-created zones are tagged with the `(Script Made)` suffix so
`DELETE /delete-zones` can find them safely.

### Macros (`src/routes/macros.ts`)
Zone-scoped bulk operations. Filtering uses the legacy POINT-containment
semantics (a widget is "in" the zone if its top-left corner is — not
its bounding box). See `src/lib/zone-helpers.ts` for the documented
divergence from the SDK extras' `WidgetZoneManager`.

### Pages / workspace tracking (`src/routes/pages.ts`)
`GET /api/clients/:id/workspace/subscribe` exposes a SSE stream of
canvas-change events. Internally it forwards the SDK's typed
`subscribeWorkspace` async iterator — no polling.

### RCU (`src/routes/rcu.ts`)
See "RCU endpoints clarification" above.

### Deleted-record audit (`src/routes/audit.ts`)
`GET /api/macros/deleted-records` lists soft-delete records produced
by `/api/macros/delete`. `GET /api/macros/deleted-details?recordId=X`
returns a per-widget-type count for the record.

### Admin panel (`src/routes/admin.ts`)
Bearer-token gated. Surfaces env vars (with API keys filtered),
verifies and applies runtime canvas swaps, manages team targets, and
lists/wipes user color assignments.

## What was dropped

- **`PM2` ecosystem config**: not part of a TS-SDK example; deploy with
  whatever process manager your stack already uses.
- **On-disk `.env` rewriting from `/admin/update-env`**: container
  anti-pattern; runtime swaps now apply only to canvas ID/name.
- **`webui-test-uploads/` fixtures**: not relocated. If we add upload
  integration tests later, fixtures will live under `tests/fixtures/`.
- **The duplicate `/api/macros/{move,copy,delete}` route definitions**
  in legacy `server.js`: collapsed into one definition each.
- **Theme editing**: the legacy `routes/theme.js` rewrote
  `public/css/styles.css` from POST bodies. CSS-rewriting from a web
  request is a wide attack surface for an example; dropped. Add it
  back via a static-asset editor if needed.
