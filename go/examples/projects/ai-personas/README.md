# ai-personas

A long-running companion that turns a Canvus canvas into an interactive
focus-group studio: it watches for trigger notes, asks Google Gemini to invent
4 customer personas from a Business Model Canvas, then drives a persona-driven
Q&A loop where every persona answers the user's question, reacts to the
others, and lays the responses out as a 3x3 grid of notes connected to the
original question.

A small web dashboard (also discoverable by QR code on the canvas) lets remote
attendees submit questions without touching the Canvus UI.

## What it does

- **Persona generation.** Place a note titled `Create_Personas` on the canvas
  alongside the 9 Business Model Canvas notes (KEY PARTNERS, KEY ACTIVITIES,
  VALUE PROPOSITIONS, …). Gemini reads the BMC, returns 4 personas, and the
  app draws them inside an anchor named `Personas`, complete with DALL-E
  generated headshots.
- **Q&A workflow.** Place a white-background note titled `New_AI_Question`.
  The app waits for you to type a question ending in `?`, fans the question
  out to all 4 personas in parallel, generates "what do you think now you've
  heard the others?" meta-answers, and lays everything out in a labelled grid
  around the question note with connectors.
- **Follow-ups.** Draw a connector from any persona-answer note to a new
  question note. The app spots the connector event, recognises the source as
  a persona answer, and creates a single follow-up answer in that persona's
  voice opposite the source note.
- **Remote submission.** Open the URL on the printed QR code (or
  `http://<host>:<port>/`) to submit a question from a phone. The note lands
  in the canvas's `Remote` anchor zone, ready to be tagged `New_AI_Question`
  by an operator.
- **/health endpoint.** Returns JSON suitable for liveness/readiness probes.

## Quick start

### Prerequisites

- A Canvus server with API access and a canvas prepared with:
  - The 9 Business Model Canvas notes (titles must match `RequiredNoteTitles`)
  - An anchor named `Personas` where persona notes will be drawn
  - An anchor named `Remote` (only required for the QR/web submit flow)
- A Google Gemini API key (https://aistudio.google.com)
- An OpenAI API key (DALL-E persona portraits)

### Run

```bash
cd go/examples/projects/ai-personas

export CANVUS_API_URL=https://dev-mtcs.multitaction.com/api/v1/
export CANVUS_API_KEY=...
export CANVAS_ID=...
export GEMINI_API_KEY=...
export OPENAI_API_KEY=sk-...

go run ./cmd/ai-personas
```

Then drop a `Create_Personas` note on the canvas and watch the personas
appear, followed by `New_AI_Question` notes for the live Q&A.

### Build the binary

```bash
# build into a temp dir to keep the working tree clean
go build -o /tmp/ai-personas ./cmd/ai-personas
/tmp/ai-personas
```

## Environment variables

| Name | Required | Default | Notes |
|---|---|---|---|
| `CANVUS_API_URL` | yes | — | Canvus base URL, e.g. `https://server/api/v1/` |
| `CANVUS_API_KEY` | yes | — | Private-Token API key |
| `CANVAS_ID` | yes | — | Canvas to monitor |
| `GEMINI_API_KEY` | yes | — | Google Gemini key (https://aistudio.google.com) |
| `OPENAI_API_KEY` | yes | — | OpenAI key (DALL-E persona portraits) |
| `GEMINI_MODEL_PERSONAS` | no | `gemini-2.5-flash` | Persona generation model |
| `GEMINI_MODEL_CHAT` | no | `gemini-2.5-flash` | Q&A model. `gemini-2.5-pro` is slower but smarter. |
| `LLM_TEMP` | no | `0.7` | Gemini sampling temperature (0.0-2.0) |
| `CHAT_TOKEN_LIMIT` | no | `256` | Max characters before re-asking for a succinct version |
| `QUESTION_TIMEOUT` | no | `5m` | How long to wait for a question to be typed (seconds or duration) |
| `SNAPSHOT_DRAIN_SECONDS` | no | `3` | Suppress pre-existing widget triggers on startup |
| `SHUTDOWN_TIMEOUT` | no | `30s` | Deadline for in-flight handlers on SIGINT/SIGTERM |
| `PORT` / `WEB_PORT` | no | `8080` | HTTP port for the QR/question dashboard |
| `PUBLIC_WEB_URL` | no | auto-detected | URL encoded into the QR code (otherwise built from FQDN) |
| `LOG_LEVEL` | no | `info` | `debug` / `info` / `warn` / `error` |
| `LOG_FORMAT` | no | `text` | `text` or `json` |

## What changed from the source repo

The original lives at `gh/AI-personas`. The port:

- Drops the hand-rolled `canvusapi/canvusapi.go` (~680 LOC) in favour of the
  new Go SDK at `go/sdk/canvus`.
- Replaces the manual SSE polling loop in `internal/canvus/events.go` with
  `session.SubscribeWidgets`, plus the snapshot-drain dedup pattern used by
  `llm-canvas-companion`. The Q&A workflow's persona/question waiter is now
  the only polling consumer.
- Decomposes the 1,286-LOC `internal/gemini/aiquestion.go` god-organism into
  focused files:
  - `internal/qa/workflow.go` — orchestrator
  - `internal/qa/helper.go` — helper-note lifecycle
  - `internal/qa/wait.go` — question-text wait + timeout helper
  - `internal/qa/answer.go` — parallel answer/meta-answer generation
  - `internal/qa/grid.go` — answer/meta grid creation, connectors, anchor
- Replaces `log.Printf` calls and the `internal/logutil` package with
  structured `log/slog` per the conventions doc §5.
- Replaces `internal/timing` (a custom DEBUG-gated timer) with `slog` at info
  level for the few timings that mattered, and discards the rest as noise.
- Strips the old SDK-style `map[string]interface{}` plumbing where the new
  SDK returns typed Notes/Anchors/Connectors.
- Drops the global `noteMonitors sync.Map` / `globalEventMonitor` /
  `globalPersonaWorkflow` / `globalQuestionWorkflow` singletons in favour of
  values constructed once in `main` and passed through.
- Drops the `dross/` planning material, the committed binaries
  (`ai-personas`, `ai-personas-linux`, `ai-personas-test`), and
  `qr_remote.png` (the QR is regenerated and uploaded directly to the canvas;
  there is no temp file on disk).
- `Showmax/go-fqdn` and `skip2/go-qrcode` are retained for the dashboard.

## Architecture

```
cmd/ai-personas/main.go         -- entrypoint: config, session, wiring, shutdown
internal/config                 -- env-var loading + defaults
internal/atom                   -- pure helpers (retry math, persona format/parse)
internal/canvasx                -- Canvus-side helpers (multipart, layout, colors, connectors)
internal/ai                     -- Gemini (google.golang.org/genai) + OpenAI DALL-E
internal/business               -- Business Model Canvas extraction + missing-notes helper
internal/persona                -- Persona creation workflow + persona-id store
internal/qa                     -- Q&A workflow (orchestrator + helper/wait/answer/grid)
internal/web                    -- HTTP dashboard, QR uploader, /health
internal/monitor                -- SubscribeWidgets dispatcher with snapshot-drain dedup
```

## Tests

Unit tests are colocated in `*_test.go` files inside each package
(per `docs/conventions/go.md §8`). The suite covers:

- env parsing + defaults
- BMC assembly (present, missing, anchor-missing, case-insensitive titles)
- persona format/parse round-trip and `AgeString` / `GoalsString` JSON
  unmarshalling quirks
- backoff math (no jitter, with jitter, never negative) and Retry-After
  parsing
- grid offsets and persona-column layout
- multipart image-upload body shape
- helper tracker semantics + question extraction
- the QR/question free-segment finder
- the white-background trigger predicate
- store defensive-copy semantics

Run them with:

```bash
go test ./...
```

Integration tests against a live Canvus server are not yet included; add
them under a `//go:build integration` package when needed.

## Known limits

- Persona count is fixed at 4; the layout (anchor columns, color palette,
  cardinal grid) assumes it. To change, update `persona.PersonaCount` and
  the offsets in `canvasx`.
- The startup validator does not exercise Gemini (only Canvus and OpenAI).
  A bad `GEMINI_API_KEY` surfaces on the first persona-generation request.
- The Q&A waiter still polls `GetNote` every 500 ms while waiting for the
  user to type a question. `SubscribeNote` is available in the SDK and would
  be a cleaner upgrade (Phase 4d).
- Persona-id state lives in-memory in `persona.Store`. Restarting the app
  loses the qnote → persona-ids mapping, but the workflow recovers by
  re-scanning the canvas for `Persona N:` notes.
- All ai-personas state is in-memory: there is no persistent task queue,
  so a crash mid-workflow leaves orphaned helper notes on the canvas.
