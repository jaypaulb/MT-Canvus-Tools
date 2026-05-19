# MT-Canvus-Tools

SDKs, examples, and tools for building on the [Canvus](https://www.multitaction.com/canvus) collaborative infinite-canvas platform. Three languages, one monorepo, 147 REST endpoints covered.

## What's here

| Language | SDK | Examples | Tools |
|---|---|---|---|
| **Go** | [`go/sdk/`](go/sdk/) | [`go/examples/`](go/examples/) — 8 core + 4 project examples | [`go/cli/`](go/cli/), [`go/tools/`](go/tools/) |
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

## Status

Phase 5 complete — CanvusPowerToys ported, Python `FoldersResource.subscribe_permissions` added, TypeScript lint floor cleared. See [CONSOLIDATION-STATUS.md](CONSOLIDATION-STATUS.md) for the full phase log.
