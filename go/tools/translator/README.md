# CanvusTranslator

A lightweight web service that translates all note widgets on a Canvus canvas into a user-selected language using the Google Gemini API.

The user clicks a flag icon in the web UI to pick a language; the service
fetches all notes via the Canvus API, translates each one in parallel using
Gemini, and patches them back in place. The canvas is updated live without
requiring a page reload.

---

## What it does

1. Serves a single-page web UI (`web/static/`) that displays a flag and a
   language picker (Mandarin, Spanish, English, Hindi, Arabic, Klingon).
2. On language selection, POSTs `{"language": "<name>"}` to `/translate`.
3. The server calls `session.ListNotes(ctx, canvasID)` to get all note
   widgets, then fans out up to 20 concurrent Gemini translation calls.
4. Each translated note is patched back via `session.UpdateNote(ctx, canvasID,
   noteID, {"text": ...})`.
5. The translated text is prefixed with the language name on its own line
   (e.g. `Spanish\nHola mundo`) so the note label is visible at a glance.

---

## Build

From the monorepo Go workspace root (`go/`):

```sh
go build ./tools/translator/...
```

Or to produce a named binary:

```sh
go build -o canvus-translator ./tools/translator/cmd/server/
```

---

## Run

The server must be started from the `go/tools/translator/` directory (or any
directory containing a `web/static/` tree) so the static file handler finds
the frontend assets.

```sh
cd go/tools/translator
export CANVUS_API_URL=https://your-server.example.com/api/v1/
export CANVUS_API_KEY=your-private-token
export CANVUS_CANVAS_ID=your-canvas-id
export GEMINI_API_KEY=your-gemini-api-key
./canvus-translator
# Listening on :8080
```

Open `http://localhost:8080` in a browser, click the flag, pick a language.

---

## Configuration

All settings are supplied via environment variables. There is no config file.

| Variable | Required | Default | Description |
|---|---|---|---|
| `CANVUS_API_URL` | yes | — | Base URL of the Canvus API, e.g. `https://host/api/v1/` |
| `CANVUS_API_KEY` | yes | — | Canvus Private-Token (API key) |
| `CANVUS_CANVAS_ID` | yes | — | ID of the canvas whose notes will be translated |
| `GEMINI_API_KEY` | yes | — | Google Gemini API key |
| `PORT` | no | `8080` | HTTP port the server listens on |
| `LOG_LEVEL` | no | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | no | text | `json` for structured output; unset for human-readable |

### Migration from the old repo

The original `CanvusTranslator` repo used different env var names:

| Old variable | New variable | Notes |
|---|---|---|
| `CANVUS_SERVER` | `CANVUS_API_URL` | Phase 4a env-var unification |
| `CANVAS_ID` | `CANVUS_CANVAS_ID` | Consistent `CANVUS_` prefix |
| `CANVUS_API_KEY` | `CANVUS_API_KEY` | Unchanged |
| `GEMINI_API_KEY` | `GEMINI_API_KEY` | Unchanged |
| `PORT` | `PORT` | Unchanged |

---

## Tests

```sh
# Unit tests only (no live server required):
go test ./tools/translator/...

# From the workspace root:
GOMAXPROCS=2 go test -count=1 ./tools/translator/...
```

There are no integration tests in this tool — the translation logic depends
on live Gemini and Canvus servers and is exercised by running the service.

---

## What's different from the original repo

| Area | Original | This port |
|---|---|---|
| SDK | Hand-rolled `internal/canvusapi/canvusapi.go` (492 LOC, 2 methods used) | `go/sdk/canvus` — `ListNotes` + `UpdateNote` |
| Gemini client | Deprecated `github.com/google/generative-ai-go/genai` | `google.golang.org/genai` v1.34.0 (unified Gemini/Vertex SDK) |
| Config | `internal/config.go` + `godotenv` | plain `os.Getenv`, no dotenv dependency |
| Architecture | Two-layer wrapper (`canvus_service.go` → `canvusapi.go`) | Single `internal/canvus.go` — direct SDK calls |
| Logging | `log.Printf` throughout | `log/slog` per monorepo convention |
| HTTP server | `http.DefaultServeMux` with no timeouts | Explicit `http.Server` with `ReadTimeout`/`WriteTimeout` |
| Env vars | `CANVUS_SERVER`, `CANVAS_ID` | `CANVUS_API_URL`, `CANVUS_CANVAS_ID` (Phase 4a names) |

The 492-line `canvusapi.go` is gone — only two of its methods were ever called.
The `internal/canvus_service.go` wrapper layer is also gone — the `CanvusSession`
in `internal/canvus.go` calls the SDK directly.

---

## Known limitations

- **Single canvas, single tenant.** The service translates exactly one canvas,
  configured at startup. There is no multi-canvas or multi-tenant routing.
- **One-shot translation.** Clicking a language replaces all note text. There
  is no undo; previous content is overwritten.
- **Gemini rate limits.** The default concurrency cap is 20 goroutines. On
  very large canvases, Gemini Flash rate limits may cause individual note
  failures (logged; other notes continue).
- **Static language list.** The UI hard-codes six languages. Extending the
  list requires editing `web/static/index.html` and `web/static/script.js`.
- **No TLS termination.** The server listens on plain HTTP. Put it behind a
  reverse proxy (nginx, Caddy) for production use.
- **Source language assumed English.** The Gemini prompt always asks to
  translate from English. If the canvas contains notes in other languages,
  the translation quality degrades.
