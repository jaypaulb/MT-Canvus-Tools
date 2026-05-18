# llm-canvas-companion

_Refreshed in Phase 4c from [https://github.com/jaypaulb/CanvusAPI-LLMDemo](https://github.com/jaypaulb/CanvusAPI-LLMDemo) (now archived)._

A live Canvus canvas watcher that dispatches LLM, OCR, and PDF-precis pipelines in response to widget events. When a user places a trigger note (e.g. `{{summarise the canvas}}`), the companion picks it up, calls the configured LLM, and writes the response back as a new note adjacent to the trigger.

This is a port of `CanvusAPI-LLMDemo` into the MT-Canvus-Tools monorepo. The application has been renamed `llm-canvas-companion` throughout (module path, binary name, README headings). All vendored Canvus API code has been replaced by the workspace SDK (`go/sdk/canvus`), and the manual NDJSON polling loop has been replaced by `session.SubscribeWidgets` with a snapshot-drain guard.

## Features

- **Note triggers** — any note whose text matches `{{...}}` launches a ChatJSON call and places the response.
- **AI image generation** — if the LLM response type is `image`, generates and uploads an image via OpenAI or Azure OpenAI.
- **Snapshot OCR** — images titled `Snapshot at …` are OCR'd via Google Cloud Vision and the transcript is placed as a note.
- **PDF précis** — an image icon titled `AI_Icon_PDF_Precis` triggers chunk-wise summarisation of the parent PDF widget.
- **Canvas précis** — an image icon titled `AI_Icon_Canvus_Precis` generates a markdown summary of every widget on the canvas.
- **Snapshot drain** — the first `SNAPSHOT_DRAIN_SECONDS` (default 5 s) of subscription events are silently discarded to avoid re-processing widgets that existed before startup.

## Build

```bash
# From the go/ workspace root:
go build ./examples/projects/llm-canvas-companion/...

# Produces binary:
go build -o llm-canvas-companion ./examples/projects/llm-canvas-companion/cmd/llm-canvas-companion/
```

## Run

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key
export CANVAS_ID=your-canvas-uuid
export OPENAI_API_KEY=sk-...
export OPENAI_NOTE_MODEL=gpt-4o          # optional, default gpt-4o
export OPENAI_IMAGE_MODEL=dall-e-3       # optional, default dall-e-3

./llm-canvas-companion
```

## Configuration (environment variables)

| Variable | Required | Default | Description |
|---|---|---|---|
| `CANVUS_API_URL` | yes | — | Base URL of the Canvus API, e.g. `https://host/api/v1` |
| `CANVUS_API_KEY` | yes | — | Canvus private-token API key |
| `CANVAS_ID` | yes | — | UUID of the canvas to watch |
| `OPENAI_API_KEY` | no¹ | — | OpenAI API key (`OPENAI_KEY` accepted as legacy alias) |
| `BASE_LLM_URL` | no | OpenAI default | Override base URL for text LLM (local or Azure) |
| `IMAGE_LLM_URL` | no | `BASE_LLM_URL` | Override base URL for image LLM |
| `OPENAI_NOTE_MODEL` | no | `gpt-4o` | Model for text completions |
| `OPENAI_IMAGE_MODEL` | no | `dall-e-3` | Model for image generation |
| `GOOGLE_VISION_API_KEY` | no | — | Google Cloud Vision key (required for OCR triggers) |
| `PDF_CHUNK_SIZE_TOKENS` | no | `3000` | Max characters per PDF chunk sent to LLM |
| `SNAPSHOT_DRAIN_SECONDS` | no | `5` | Seconds to drain pre-existing widget events on startup |

¹ Required when using note triggers or PDF/canvas précis.

## Tests

Unit tests (no live server required):

```bash
go test -count=1 ./examples/projects/llm-canvas-companion/...
```

Integration tests (require a live Canvus server):

```bash
CANVUS_API_URL=https://... CANVUS_API_KEY=... CANVAS_ID=... \
  go test -tags=integration -count=1 ./examples/projects/llm-canvas-companion/tests/...
```

## Package layout

```
cmd/llm-canvas-companion/   main binary
internal/
  config/                   env-var loading and validation
  ai/                       OpenAI / Azure / local LLM client
  ocr/                      Google Cloud Vision OCR
  handlers/
    common.go               shared helpers (notes, retry, sizing)
    note.go                 {{...}} trigger pipeline
    snapshot.go             OCR pipeline
    pdf.go                  PDF précis pipeline
    canvas.go               canvas précis pipeline
    image.go                AI image generation and upload
    export.go               test-export shims for internal functions
  monitor/                  SubscribeWidgets dispatch loop
tests/                      external test package
testdata/                   fixture JSON files
```

## Known limits

- Local LLM endpoints do not support image generation; the companion logs a warning and skips.
- PDF extraction requires a local file path accessible to the process; remote-only PDFs are not yet supported.
- Google Vision OCR requires a billable GCP project with the Vision API enabled.
