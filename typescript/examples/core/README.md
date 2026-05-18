# Canvus TypeScript Examples — Core

Eight canonical examples for the `@mt-canvus-tools/sdk`. Each is a
self-contained pnpm workspace package that consumes the SDK as a
`workspace:*` dependency, so edits to the SDK source propagate
immediately.

## Index

| # | Slug                        | Purpose                                                                              | Key env vars                                                                                  | Complexity |
| - | --------------------------- | ------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------- | ---------- |
| 1 | `01-auth-and-list`          | Smoke test: API-key auth, list canvases, print a summary table.                      | `CANVUS_API_URL`, `CANVUS_API_KEY`                                                           | *          |
| 2 | `02-auth-flows`             | Three auth flows: API key, login, access-token CRUD.                                 | `+ CANVUS_EMAIL`, `CANVUS_PASSWORD`                                                           | **         |
| 3 | `03-widget-crud`            | Full create / update / delete / verify lifecycle on a sticky note.                   | `+ CANVUS_CANVAS_ID`                                                                          | **         |
| 4 | `04-file-upload`            | Multipart PNG upload + position patch + cleanup.                                     | `+ CANVUS_CANVAS_ID`, `CANVUS_IMAGE_PATH?`, `CANVUS_KEEP_WIDGET?`                             | **         |
| 5 | `05-streaming`              | NDJSON subscribe with clean shutdown on duration or SIGINT.                          | `+ CANVUS_CANVAS_ID`, `STREAM_DURATION_SECONDS?`                                              | ***        |
| 6 | `06-llm-integration`        | Watcher posts Ollama answers to `?`-prefixed notes; responder lists Ollama models.   | `+ CANVUS_CANVAS_ID`, `OLLAMA_URL?`, `OLLAMA_MODEL?`                                          | ****       |
| 7 | `07-webhooks-notifications` | Bridge Canvus widget events to an outbound HTTP webhook with backoff + retry.        | `+ CANVUS_CANVAS_ID`, `WEBHOOK_URL`                                                           | ****       |
| 8 | `08-cross-canvas-clone`     | Clone a widget across canvases via the SDK's `widgets.clone({...})` helper.          | `+ CANVUS_CANVAS_ID`, `CANVUS_SOURCE_WIDGET_ID`, `CANVUS_DEST_CANVAS_ID`, `CANVUS_CLEANUP?`   | ***        |

## Running the examples

Each example is its own package. From the workspace root
(`typescript/`):

```bash
# One-time
pnpm install

# Run any example in dev mode (tsx, no build step)
pnpm --filter @mt-canvus-tools/example-auth-and-list dev

# Or build then run with native Node
pnpm --filter @mt-canvus-tools/example-auth-and-list build
node --env-file=examples/core/01-auth-and-list/.env \
  examples/core/01-auth-and-list/dist/index.js
```

From an individual example directory:

```bash
cd examples/core/01-auth-and-list
cp .env.example .env   # then edit
pnpm dev               # tsx
# or
pnpm build && node --env-file=.env dist/index.js
```

## Env loading

All examples follow the same pattern:

1. Env vars are read from `process.env` (which Node 20+ populates from
   `--env-file=.env` if you pass it, or from your shell).
2. A `zod` schema validates the required vars at startup. Missing or
   malformed vars exit with status 2 and a structured error.
3. A frozen, typed env object is passed into the SDK via
   `createSession({ baseUrl, apiKey, ... })`.

For local dev, [`direnv`](https://direnv.net) is the easiest way to
auto-load `.env` files when you `cd` into an example directory. Create
a `.envrc` that does `dotenv .env`.

## Shared conventions

- **Logging:** [`pino`](https://getpino.io). JSON to stdout in production;
  set `LOG_FORMAT=pretty` for human-readable colour output via
  `pino-pretty`. Set `LOG_LEVEL=debug` for verbose output.
- **Errors:** Every example catches and inspects SDK errors via
  `isCanvusError(err)` / `err.kind`. Non-SDK errors are logged at
  `error` and bubble up.
- **Canvas coordinates:** Pixels, never normalised. `location.x`,
  `location.y`, `size.width`, `size.height` are all raw pixel values.
- **Wire shape:** Field names match the JSON exactly (kebab-case for
  most fields, with documented hybrid hyphen/underscore conventions for
  IP Video, RDP, and a handful of others — see
  `docs/api-reference/VERIFIED-CORRECTIONS.md` §7).
- **Cleanup:** Examples that create resources delete them by default
  unless an explicit env flag (e.g. `CANVUS_KEEP_WIDGET=1`,
  `CANVUS_CLEANUP=1`) tells them otherwise.

## Verified corrections respected

Every example honours the live-server findings in
`docs/api-reference/VERIFIED-CORRECTIONS.md`, in particular:

- **§4** — login sends `email` only, never `username`.
- **§6** — `User.user-id` is an integer.
- **§7** — hybrid field naming for IP Video / RDP.
- **§8** — audit log returns a flat array.
- **Changelog §1** — cross-canvas clone uses the standard create
  endpoint with `source_canvas_id` + `source_widget_id`, not the
  legacy 501 `/widgets/clone` endpoint.
- **Changelog §2** — no POST helpers for `ip-video` or
  `rdp-connection`.

## Toolchain notes

- **Node:** 20 LTS or newer. Examples rely on native `fetch`,
  `--env-file=.env`, and `AbortSignal.timeout`.
- **Runner:** `tsx` for dev, `tsc` for build. No `tsup` — that's
  reserved for the SDK's dual ESM/CJS distribution build.
- **Module format:** ESM only (`"type": "module"` everywhere).
