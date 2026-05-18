# Canvus Go SDK — Core Examples

Eight canonical examples that demonstrate the Canvus Go SDK end-to-end.
They are deliberately small and standalone — each is a separate Go module
in the workspace so you can copy one out as the seed for your own tool
without dragging the rest along.

Work through them in order. Each one builds on patterns from the previous
ones.

## Index

| # | Slug | Purpose | Key env vars | Complexity |
| --- | --- | --- | --- | --- |
| 01 | [`01-auth-and-list`](01-auth-and-list/) | Authenticate with API key, list canvases as a tab-aligned table. | `CANVUS_API_URL`, `CANVUS_API_KEY` | * |
| 02 | [`02-auth-flows`](02-auth-flows/) | Three auth modes: API key, email+password login, access-token CRUD. | + `CANVUS_EMAIL`, `CANVUS_PASSWORD` | ** |
| 03 | [`03-widget-crud`](03-widget-crud/) | Create / patch / delete / verify a sticky note. | + `CANVUS_CANVAS_ID` | ** |
| 04 | [`04-file-upload`](04-file-upload/) | Upload a PNG as an image widget, reposition, optionally clean up. | + `CANVUS_IMAGE_PATH`, `CANVUS_KEEP_WIDGET` | *** |
| 05 | [`05-streaming`](05-streaming/) | Subscribe to a canvas's notes endpoint, print NDJSON frames. | + `STREAM_DURATION_SECONDS` | *** |
| 06 | [`06-llm-integration`](06-llm-integration/) | Two-binary Canvus+Ollama integration: question notes → LLM → answer notes. | + `OLLAMA_URL`, `OLLAMA_MODEL` | ***** |
| 07 | [`07-webhooks-notifications`](07-webhooks-notifications/) | Synthesise outbound webhooks from the subscribe stream, with retry. | + `WEBHOOK_URL` | **** |
| 08 | [`08-cross-canvas-clone`](08-cross-canvas-clone/) | Clone a widget from one canvas to another via the SDK clone helper. | + `CANVUS_DEST_CANVAS_ID`, `CANVUS_SOURCE_WIDGET_ID`, `CANVUS_CLEANUP` | ** |

Complexity legend:
- *  — single-call smoke test
- ** — multi-step, single concern
- *** — multipart bodies or long-lived I/O
- **** — production-ish patterns (retries, idempotency)
- ***** — multiple binaries, external service integration

## Running the examples

Each example is its own module. From the example's own directory:

```bash
cp .env.example .env
$EDITOR .env
set -a; source .env; set +a
go run .
```

From the Go workspace root (`go/`):

```bash
go run ./examples/core/01-auth-and-list
go run ./examples/core/06-llm-integration/cmd/responder
go run ./examples/core/06-llm-integration/cmd/watcher
```

### Env-loading pattern

All examples read environment variables with `os.Getenv` only. No dotenv
dependency — idiomatic Go is plain `$ENV`. Four equivalent ways to load
your `.env`:

1. **Shell `source`** — quickest for one-off runs.
   ```bash
   set -a; source .env; set +a
   ```
2. **direnv** — automatic on `cd`. Drop the same lines into `.envrc`:
   ```bash
   cd 01-auth-and-list
   echo 'dotenv' > .envrc
   direnv allow
   ```
3. **systemd / Docker / nomad** — production: ship env vars via your
   orchestration platform's secret mechanism. Never bake them into images.
4. **`env -S` for one-shot runs** — useful in CI:
   ```bash
   env CANVUS_API_URL=... CANVUS_API_KEY=... go run .
   ```

### Where to get .env values

- `CANVUS_API_URL` — your Canvus server's URL including `/api/v1/`. The
  dev server is `https://dev-mtcs.multitaction.com/api/v1/`.
- `CANVUS_API_KEY` — mint from the Canvus UI under
  **Settings → Access Tokens**. Don't reuse a personal token for
  unattended services; create a dedicated one and give it a meaningful
  name.
- `CANVUS_CANVAS_ID` — run example 01 to discover canvas IDs; copy from
  the leftmost column. The Canvus UI also surfaces canvas IDs in URL bars
  and share dialogs.
- `CANVUS_EMAIL` / `CANVUS_PASSWORD` — your normal Canvus credentials.
  Used only by example 02 to demonstrate the password-login auth path.
- `WEBHOOK_URL` — for testing, grab a one-shot URL from
  [webhook.site](https://webhook.site/) and watch deliveries arrive in
  your browser.
- `OLLAMA_URL` / `OLLAMA_MODEL` — defaults are
  `http://localhost:11434` and `llama3.2`. Install Ollama from
  <https://ollama.com>, then `ollama pull llama3.2`.

## Conventions

Every example follows the same conventions, defined in
[`docs/conventions/go.md`](../../../docs/conventions/go.md):

- `log/slog` only — no third-party loggers. `LOG_FORMAT=json` for prod,
  `text` (default) for local dev.
- `main()` is a one-liner that calls `run()`. All real logic lives in
  `run` and downstream functions so errors bubble up via `return err`,
  never `log.Fatal`.
- Every I/O function takes `context.Context` as its first argument.
- All errors are wrapped with `fmt.Errorf("operation: %w", err)` so
  `errors.Is` and `errors.As` work all the way up.
- Sentinel checks (`errors.Is(err, canvus.ErrNotFound)`) over string
  matching.
- Polymorphic widget bodies use `any`, not `interface{}` (Go 1.18+ idiom).

## File layout

```
core/
├── README.md                          # this file
├── 01-auth-and-list/
│   ├── main.go
│   ├── go.mod
│   ├── .env.example
│   └── README.md
├── 02-auth-flows/...
├── 03-widget-crud/...
├── 04-file-upload/...
├── 05-streaming/...
├── 06-llm-integration/
│   ├── cmd/
│   │   ├── responder/main.go
│   │   └── watcher/main.go
│   ├── go.mod
│   ├── .env.example
│   └── README.md
├── 07-webhooks-notifications/...
└── 08-cross-canvas-clone/...
```

Each example is registered in the workspace `go/go.work` so the workspace
build (`cd go && go build ./...`) catches breakage everywhere at once.
