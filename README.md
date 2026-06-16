# MT-Canvus-Tools

SDKs, examples, and tools for building on the [Canvus](https://www.multitaction.com/canvus) collaborative infinite-canvas platform. Three languages, one monorepo, 147 REST endpoints covered.

## What's here

| Language | SDK | Examples | Tools |
|---|---|---|---|
| **Go** | [`go/sdk/`](go/sdk/) | [`go/examples/`](go/examples/) — 8 core + 3 project examples | [`go/cli/`](go/cli/), [`go/tools/`](go/tools/) |
| **Python** | [`python/sdk/`](python/sdk/) | [`python/examples/`](python/examples/) — 8 core examples | [`python/tools/`](python/tools/) |
| **TypeScript** | [`typescript/sdk/`](typescript/sdk/) | [`typescript/examples/`](typescript/examples/) — 8 core + WebUI | — |

## Prerequisites

| Language | Runtime | Package manager | Min version |
|---|---|---|---|
| Go | `go` | built-in modules + `go.work` | 1.24 |
| Python | `python` | `uv` | 3.11 |
| TypeScript | `node` | `pnpm` | 20 |

A running Canvus server is required for all examples and live tests. Set two environment variables before running anything:

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key
```

## Quick start

```bash
git clone https://github.com/jaypaulb/MT-Canvus-Tools.git
cd MT-Canvus-Tools
```

**Go:**
```bash
cd go
go build ./...
cd examples/core/01-auth-and-list && go run .
```

**Python:**
```bash
cd python
uv sync
uv run python examples/core/01-auth-and-list/main.py
```

**TypeScript:**
```bash
cd typescript
pnpm install && pnpm build
node examples/core/01-auth-and-list/dist/index.js
```

## Examples

Each language ships the same eight core examples — work through them in order, as each builds on the last. The links below point at the Go version; the Python and TypeScript ports live under [`python/examples/core/`](python/examples/core/) and [`typescript/examples/core/`](typescript/examples/core/).

| # | Example | What it does |
|---|---|---|
| 01 | [`01-auth-and-list`](go/examples/core/01-auth-and-list/) | Authenticate with an API key and list canvases as a table. |
| 02 | [`02-auth-flows`](go/examples/core/02-auth-flows/) | Three auth modes: API key, email + password login, and access-token CRUD. |
| 03 | [`03-widget-crud`](go/examples/core/03-widget-crud/) | Create, patch, delete, and verify a sticky note. |
| 04 | [`04-file-upload`](go/examples/core/04-file-upload/) | Upload a PNG as an image widget and reposition it. |
| 05 | [`05-streaming`](go/examples/core/05-streaming/) | Subscribe to a canvas's notes endpoint and print NDJSON frames. |
| 06 | [`06-llm-integration`](go/examples/core/06-llm-integration/) | Two-binary Canvus + Ollama loop: question notes → LLM → answer notes. |
| 07 | [`07-webhooks-notifications`](go/examples/core/07-webhooks-notifications/) | Synthesise outbound webhooks from the subscribe stream, with retries. |
| 08 | [`08-cross-canvas-clone`](go/examples/core/08-cross-canvas-clone/) | Clone a widget from one canvas to another via the SDK clone helper. |

### Project examples

Larger, end-to-end apps built on the SDKs:

- [`go/examples/projects/ai-personas`](go/examples/projects/ai-personas/) — Turn a Canvus canvas into an interactive AI-persona experience.
- [`go/examples/projects/llm-canvas-companion`](go/examples/projects/llm-canvas-companion/) — A live canvas watcher that runs LLM, OCR, and PDF-précis pipelines in response to widget events.
- [`go/examples/projects/note-mapper`](go/examples/projects/note-mapper/) — Photograph physical Post-it notes and recreate them as digital notes in a canvas anchor zone.
- [`typescript/examples/webui`](typescript/examples/webui/) — A Hono/TypeScript admin/operator UI built on `@mt-canvus-tools/sdk`.

## Documentation

| Doc | Contents |
|---|---|
| [Getting started — Go](docs/getting-started/go.md) | SDK install, first request, example runner |
| [Getting started — Python](docs/getting-started/python.md) | uv setup, first request, example runner |
| [Getting started — TypeScript](docs/getting-started/typescript.md) | pnpm setup, first request, example runner |
| [API reference](docs/api-reference/README.md) | 147 endpoints, schemas, changelog |
| [Conventions — Go](docs/conventions/go.md) | Build, lint, log, error, test defaults |
| [Conventions — Python](docs/conventions/python.md) | Same for Python |
| [Conventions — TypeScript](docs/conventions/typescript.md) | Same for TypeScript |
| [Contributing](CONTRIBUTING.md) | PR workflow, test requirements |

## SDK coverage

All three SDKs cover **147 / 147** Canvus REST endpoints (three endpoints are deliberately omitted — see [API changelog §1–2](docs/api-reference/changelog.md)).

| SDK | Streaming helpers | Extras subpackage | Toolchain |
|---|---|---|---|
| Go | 44 typed Subscribe helpers | geometry, filters, zones, batch\_widgets, search | `go build` / `go vet` / `go test` clean |
| Python | 49 typed AsyncIterators | 10 modules + `__init__` | ruff clean, pytest pass, mypy --strict 0 errors |
| TypeScript | 27 typed async iterators | 9 modules + index | typecheck / build clean, vitest pass |
