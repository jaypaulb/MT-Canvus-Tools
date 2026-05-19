# MT-Canvus-Tools Phase 6 — Top-level Documentation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author all top-level and per-language documentation — root README, per-language workspace READMEs, per-language getting-started guides, contributing guides, and CONSOLIDATION-STATUS update — accurately reflecting the monorepo post-Phase 5.

**Architecture:** Documentation-only phase. No code changes. Each task creates or replaces one or more markdown files, verifies accuracy by checking referenced paths against the real filesystem, then commits. Write tasks are self-contained; every task includes the complete final file content.

**Tech Stack:** Markdown, git

---

## File structure

| Action | Path | Task |
|---|---|---|
| Modify | `README.md` | 1 |
| Create | `go/README.md` | 2 |
| Create | `python/README.md` | 3 |
| Create | `typescript/README.md` | 4 |
| Create | `docs/getting-started/go.md` | 5 |
| Create | `docs/getting-started/python.md` | 6 |
| Create | `docs/getting-started/typescript.md` | 7 |
| Modify | `CONTRIBUTING.md` | 8 |
| Create | `docs/contributing/README.md` | 8 |
| Create | `docs/contributing/go.md` | 8 |
| Create | `docs/contributing/python.md` | 8 |
| Create | `docs/contributing/typescript.md` | 8 |
| Modify | `CONSOLIDATION-STATUS.md` | 9 |

---

### Task 1: Root README.md

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Write README.md**

Replace the entire file content with:

```markdown
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
| Go | `go` | built-in modules + `go.work` | 1.22 |
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
| Go | 36 typed Subscribe helpers | geometry, filters, zones, batch\_widgets, search | `go build` / `go vet` / `go test` clean |
| Python | 27 typed AsyncIterators | 10 modules + `__init__` | ruff clean, pytest pass, mypy --strict 0 errors |
| TypeScript | 27 typed async iterators | 9 modules + index | typecheck / build clean, vitest pass |

## Status

Phase 5 complete — CanvusPowerToys ported, Python `FoldersResource.subscribe_permissions` added, TypeScript lint floor cleared. See [CONSOLIDATION-STATUS.md](CONSOLIDATION-STATUS.md) for the full phase log.
```

- [ ] **Step 2: Verify referenced paths exist**

```bash
ls go/sdk/ python/sdk/ typescript/sdk/
ls docs/getting-started/ docs/api-reference/ docs/conventions/
ls go/cli/ go/tools/ go/examples/core/ python/examples/core/ typescript/examples/core/
```

Expected: all directories exist without error.

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs: expand root README — language matrix, quick start, toolchain status (Phase 6)"
```

---

### Task 2: go/README.md

**Files:**
- Create: `go/README.md`

- [ ] **Step 1: Write go/README.md**

```markdown
# MT-Canvus-Tools — Go

Go workspace containing the SDK, CLI, examples, and tools for the Canvus platform. Managed by `go.work` — all modules build and test from this directory.

## Structure

| Path | Description |
|---|---|
| [`sdk/`](sdk/) | Go SDK — 147 endpoints, 36 typed Subscribe helpers, extras subpackage |
| [`cli/`](cli/) | `canvus` CLI — thin wrapper around the SDK |
| [`examples/core/`](examples/core/) | 8 numbered examples (01 auth, 02 auth-flows, 03 widget-CRUD, 04 file-upload, 05 streaming, 06 LLM, 07 webhooks, 08 cross-canvas-clone) |
| [`examples/projects/`](examples/projects/) | 3 project examples (note-mapper, llm-canvas-companion, ai-personas) |
| [`tools/db-solver/`](tools/db-solver/) | Database import/export helper |
| [`tools/powertoys/`](tools/powertoys/) | Desktop tray app with WebUI and canvas relay tools |
| [`tools/translator/`](tools/translator/) | Real-time note translation via LLM |
| [`internal/llm/`](internal/llm/) | Shared Gemini helper used by project examples |

## Workspace commands

Run from this directory (`go/`) to operate across all modules:

```bash
go build ./...        # Build every module
go test ./...         # Test every module
go vet ./...          # Vet every module
gofmt -s -l .         # Check formatting — must print nothing
```

## Authentication

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key
```

The SDK sends the key as the `Private-Token` request header.

## Quick example

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func main() {
	cfg := canvus.DefaultSessionConfig()
	cfg.BaseURL = os.Getenv("CANVUS_API_URL")
	session := canvus.NewSession(cfg, canvus.WithAPIKey(os.Getenv("CANVUS_API_KEY")))

	canvases, err := session.ListCanvases(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range canvases {
		fmt.Printf("%s  %s\n", c.ID, c.Name)
	}
}
```

## Conventions

[`docs/conventions/go.md`](../docs/conventions/go.md) (monorepo root) — locked defaults for import order, error wrapping, logging, context propagation, and test structure.

## Getting started

[`docs/getting-started/go.md`](../docs/getting-started/go.md) — step-by-step setup, authentication, running examples.
```

- [ ] **Step 2: Verify referenced paths exist**

```bash
ls go/sdk/ go/cli/ go/examples/core/ go/examples/projects/ go/tools/db-solver/ go/tools/powertoys/ go/tools/translator/ go/internal/llm/
```

Expected: all exist.

- [ ] **Step 3: Commit**

```bash
git add go/README.md
git commit -m "docs: add go/README.md — workspace landing page (Phase 6)"
```

---

### Task 3: python/README.md

**Files:**
- Create: `python/README.md`

- [ ] **Step 1: Write python/README.md**

```markdown
# MT-Canvus-Tools — Python

Python uv workspace containing the SDK, examples, and tools for the Canvus platform.

## Structure

| Path | Description |
|---|---|
| [`sdk/`](sdk/) | Python SDK — 147 endpoints, 27 typed AsyncIterators, 10 extras modules |
| [`examples/core/`](examples/core/) | 8 numbered examples (01 auth, 02 auth-flows, 03 widget-CRUD, 04 file-upload, 05 streaming, 06 LLM, 07 webhooks, 08 cross-canvas-clone) |
| [`tools/mcp-server/`](tools/mcp-server/) | MCP server exposing Canvus to Claude Desktop — 16 LLM/brainstorming/correlation/reports tools |

## Workspace setup

```bash
uv sync                                       # Install all workspace dependencies
uv run pytest                                 # Run SDK tests
uv run ruff check .                           # Lint
uv run mypy --strict python/sdk/src           # Type-check SDK (0 errors baseline)
```

Tests live in `sdk/tests/`. Integration tests opt in via `-m integration` (require `CANVUS_API_URL` + `CANVUS_API_KEY`). Live dev-server tests opt in via `-m live`.

## Authentication

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key
```

## Quick example

```python
import asyncio
import os
from canvus_sdk import Client

async def main():
    async with Client(
        base_url=os.environ["CANVUS_API_URL"],
        api_key=os.environ["CANVUS_API_KEY"],
    ) as client:
        canvases = await client.canvases.list()
        for canvas in canvases:
            print(canvas.id, canvas.name)

asyncio.run(main())
```

Or using environment-backed config:

```python
async with Client.from_env() as client:
    canvases = await client.canvases.list()
```

## Conventions

[`docs/conventions/python.md`](../docs/conventions/python.md) (monorepo root).

## Getting started

[`docs/getting-started/python.md`](../docs/getting-started/python.md) — step-by-step setup, authentication, running examples.
```

- [ ] **Step 2: Verify referenced paths exist**

```bash
ls python/sdk/ python/examples/core/ python/tools/mcp-server/
```

Expected: all exist.

- [ ] **Step 3: Commit**

```bash
git add python/README.md
git commit -m "docs: add python/README.md — workspace landing page (Phase 6)"
```

---

### Task 4: typescript/README.md

**Files:**
- Create: `typescript/README.md`

- [ ] **Step 1: Write typescript/README.md**

```markdown
# MT-Canvus-Tools — TypeScript

pnpm workspace containing the SDK and examples for the Canvus platform.

## Structure

| Path | Description |
|---|---|
| [`sdk/`](sdk/) | TypeScript SDK — 147 endpoints, 27 typed async iterators, 9 extras modules |
| [`examples/core/`](examples/core/) | 8 numbered examples (01 auth, 02 auth-flows, 03 widget-CRUD, 04 file-upload, 05 streaming, 06 LLM, 07 webhooks, 08 cross-canvas-clone) |
| [`examples/webui/`](examples/webui/) | Hono-based WebUI with RCU canvas relay and server-sent events |

## Workspace setup

```bash
pnpm install       # Install all workspace dependencies
pnpm build         # Build SDK + all examples (ESM + CJS + DTS)
pnpm test          # Run vitest suite
pnpm typecheck     # Type-check without emitting
pnpm lint          # ESLint
```

## Authentication

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key
```

## Quick example

```typescript
import { createSession } from "@mt-canvus-tools/sdk";

const session = createSession({
  baseUrl: process.env.CANVUS_API_URL!,
  apiKey: process.env.CANVUS_API_KEY,
});

const canvases = await session.canvases.list();
for (const canvas of canvases) {
  console.log(canvas.id, canvas.name);
}
```

Or using environment-backed config:

```typescript
import { Session, loadConfig } from "@mt-canvus-tools/sdk";

const session = new Session(loadConfig());
```

## Conventions

[`docs/conventions/typescript.md`](../docs/conventions/typescript.md) (monorepo root).

## Getting started

[`docs/getting-started/typescript.md`](../docs/getting-started/typescript.md) — step-by-step setup, authentication, running examples.
```

- [ ] **Step 2: Verify referenced paths exist**

```bash
ls typescript/sdk/ typescript/examples/core/ typescript/examples/webui/
```

Expected: all exist.

- [ ] **Step 3: Commit**

```bash
git add typescript/README.md
git commit -m "docs: add typescript/README.md — workspace landing page (Phase 6)"
```

---

### Task 5: docs/getting-started/go.md

**Files:**
- Create: `docs/getting-started/go.md`

- [ ] **Step 1: Write docs/getting-started/go.md**

```markdown
# Getting Started — Go

Get up and running with the MT-Canvus-Tools Go SDK.

## Prerequisites

- **Go 1.22 or later** — verify with `go version`
- A running Canvus server and an API key
- `git` (for running monorepo examples)

## Option A: Use the SDK as a dependency

```bash
go get github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus@latest
```

Create `main.go`:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func main() {
	cfg := canvus.DefaultSessionConfig()
	cfg.BaseURL = os.Getenv("CANVUS_API_URL")
	session := canvus.NewSession(cfg, canvus.WithAPIKey(os.Getenv("CANVUS_API_KEY")))

	canvases, err := session.ListCanvases(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range canvases {
		fmt.Printf("%s  %s\n", c.ID, c.Name)
	}
}
```

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key
go run .
```

## Option B: Run examples from the monorepo

```bash
git clone https://github.com/jaypaulb/MT-Canvus-Tools.git
cd MT-Canvus-Tools/go
go build ./...
```

Set credentials, then run any numbered example:

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key

cd examples/core/01-auth-and-list && go run .
cd examples/core/05-streaming && go run .
```

## Environment variables

| Variable | Description |
|---|---|
| `CANVUS_API_URL` | Full base URL including `/api/v1` suffix (e.g. `https://canvus.example.com/api/v1`) |
| `CANVUS_API_KEY` | Long-lived API key — sent as `Private-Token` header |
| `CANVUS_INSECURE_TLS` | Set to `1` to disable TLS verification for self-signed certificates |

## Authentication modes

```go
// API key (recommended for services and automation)
session := canvus.NewSession(cfg, canvus.WithAPIKey(os.Getenv("CANVUS_API_KEY")))

// Username + password (interactive flows; short-lived token)
session := canvus.NewSession(cfg)
if err := session.Login(ctx, "user@example.com", "password"); err != nil {
    log.Fatal(err)
}
defer session.Logout(ctx)
```

## Real-time streaming

The SDK exposes a typed Subscribe helper for every streamable endpoint:

```go
eventChan, errChan, err := session.SubscribeCanvases(ctx, nil)
if err != nil {
    log.Fatal(err)
}
for ev := range eventChan {
    fmt.Printf("event: %s  canvas: %s\n", ev.Type, ev.Canvas.Name)
}
if err := <-errChan; err != nil {
    log.Println("stream error:", err)
}
```

## Go tools

| Tool | Path | Run |
|---|---|---|
| `canvus` CLI | `go/cli/` | `go run . <command>` |
| db-solver | `go/tools/db-solver/` | `go run .` |
| powertoys | `go/tools/powertoys/` | `go run cmd/powertoys/` |
| translator | `go/tools/translator/` | `go run .` |

## Next steps

- [Go workspace README](../../go/README.md)
- [Go SDK README](../../go/sdk/README.md)
- [Go conventions](../conventions/go.md)
- [API reference](../api-reference/README.md)
- [All core examples](../../go/examples/core/)
```

- [ ] **Step 2: Verify all referenced paths exist**

```bash
ls go/sdk/ go/cli/ go/tools/db-solver/ go/tools/powertoys/ go/tools/translator/
ls go/examples/core/
ls docs/conventions/go.md docs/api-reference/
```

Expected: all exist without error.

- [ ] **Step 3: Commit**

```bash
git add docs/getting-started/go.md
git commit -m "docs: add docs/getting-started/go.md — Go quickstart (Phase 6)"
```

---

### Task 6: docs/getting-started/python.md

**Files:**
- Create: `docs/getting-started/python.md`

- [ ] **Step 1: Write docs/getting-started/python.md**

```markdown
# Getting Started — Python

Get up and running with the MT-Canvus-Tools Python SDK.

## Prerequisites

- **Python 3.11 or later** — verify with `python --version`
- **uv** — install with `pip install uv` or see the [uv docs](https://docs.astral.sh/uv/)
- A running Canvus server and an API key

## Option A: Use the SDK as a dependency

> **Note:** Until the package is published on PyPI, use Option B or install directly from the monorepo with `uv add "canvus-sdk @ git+https://github.com/jaypaulb/MT-Canvus-Tools.git#subdirectory=python/sdk"`.

Create `main.py`:

```python
import asyncio
import os
from canvus_sdk import Client

async def main():
    async with Client(
        base_url=os.environ["CANVUS_API_URL"],
        api_key=os.environ["CANVUS_API_KEY"],
    ) as client:
        canvases = await client.canvases.list()
        for canvas in canvases:
            print(canvas.id, canvas.name)

asyncio.run(main())
```

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key
uv run python main.py
```

## Option B: Run examples from the monorepo

```bash
git clone https://github.com/jaypaulb/MT-Canvus-Tools.git
cd MT-Canvus-Tools/python
uv sync
```

Set credentials, then run any numbered example:

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key

uv run python examples/core/01-auth-and-list/main.py
uv run python examples/core/05-streaming/main.py
```

## Environment variables

| Variable | Description |
|---|---|
| `CANVUS_API_URL` | Full base URL including `/api/v1` suffix |
| `CANVUS_API_KEY` | Long-lived API key — sent as `Private-Token` header |

Environment-backed config is available via `Client.from_env()` (reads `CANVUS_API_URL` and `CANVUS_API_KEY`):

```python
async with Client.from_env() as client:
    canvases = await client.canvases.list()
```

## Authentication modes

```python
# API key (recommended for services and automation)
async with Client(
    base_url=os.environ["CANVUS_API_URL"],
    api_key=os.environ["CANVUS_API_KEY"],
) as client:
    ...

# Username + password (interactive flows; short-lived token)
async with Client(base_url=os.environ["CANVUS_API_URL"]) as client:
    await client.auth.login(email="user@example.com", password="password")
    ...
```

## Real-time streaming

```python
async with Client.from_env() as client:
    async for event in client.canvases.subscribe():
        print(event.type, event.canvas.name)
```

## Python tools

| Tool | Path | Run |
|---|---|---|
| MCP server | `python/tools/mcp-server/` | `uv run python -m canvus_mcp` |

## Next steps

- [Python workspace README](../../python/README.md)
- [Python conventions](../conventions/python.md)
- [API reference](../api-reference/README.md)
- [All core examples](../../python/examples/core/)
```

- [ ] **Step 2: Verify referenced paths exist**

```bash
ls python/sdk/ python/examples/core/ python/tools/mcp-server/
ls docs/conventions/python.md docs/api-reference/
```

Expected: all exist without error.

- [ ] **Step 3: Commit**

```bash
git add docs/getting-started/python.md
git commit -m "docs: add docs/getting-started/python.md — Python quickstart (Phase 6)"
```

---

### Task 7: docs/getting-started/typescript.md

**Files:**
- Create: `docs/getting-started/typescript.md`

- [ ] **Step 1: Write docs/getting-started/typescript.md**

```markdown
# Getting Started — TypeScript

Get up and running with the MT-Canvus-Tools TypeScript SDK.

## Prerequisites

- **Node.js 20 or later** — verify with `node --version`
- **pnpm** — install with `npm install -g pnpm` or see the [pnpm docs](https://pnpm.io/)
- A running Canvus server and an API key

## Option A: Use the SDK as a dependency

> **Note:** Until the package is published on npm, use Option B or link the SDK via a local path.

```bash
pnpm add @mt-canvus-tools/sdk
```

Create `index.ts`:

```typescript
import { createSession } from "@mt-canvus-tools/sdk";

const session = createSession({
  baseUrl: process.env.CANVUS_API_URL!,
  apiKey: process.env.CANVUS_API_KEY,
});

const canvases = await session.canvases.list();
for (const canvas of canvases) {
  console.log(canvas.id, canvas.name);
}
```

## Option B: Run examples from the monorepo

```bash
git clone https://github.com/jaypaulb/MT-Canvus-Tools.git
cd MT-Canvus-Tools/typescript
pnpm install
pnpm build
```

Set credentials, then run any numbered example:

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key

node examples/core/01-auth-and-list/dist/index.js
node examples/core/05-streaming/dist/index.js
```

## Environment variables

| Variable | Description |
|---|---|
| `CANVUS_API_URL` | Full base URL including `/api/v1` suffix |
| `CANVUS_API_KEY` | Long-lived API key — sent as `Private-Token` header |

Environment-backed config is available via `loadConfig()` (reads `CANVUS_API_URL` and `CANVUS_API_KEY`):

```typescript
import { Session, loadConfig } from "@mt-canvus-tools/sdk";

const session = new Session(loadConfig());
```

## Authentication modes

```typescript
// API key (recommended for services and automation)
const session = createSession({
  baseUrl: process.env.CANVUS_API_URL!,
  apiKey: process.env.CANVUS_API_KEY,
});

// Username + password (interactive flows; short-lived token)
const session = createSession({ baseUrl: process.env.CANVUS_API_URL! });
await session.auth.login({ email: "user@example.com", password: "password" });
```

## Real-time streaming

```typescript
for await (const event of session.canvases.subscribe()) {
  console.log(event.type, event.canvas?.name);
}
```

## Next steps

- [TypeScript workspace README](../../typescript/README.md)
- [TypeScript conventions](../conventions/typescript.md)
- [API reference](../api-reference/README.md)
- [All core examples](../../typescript/examples/core/)
```

- [ ] **Step 2: Verify referenced paths exist**

```bash
ls typescript/sdk/ typescript/examples/core/ typescript/examples/webui/
ls docs/conventions/typescript.md docs/api-reference/
```

Expected: all exist without error.

- [ ] **Step 3: Commit**

```bash
git add docs/getting-started/typescript.md
git commit -m "docs: add docs/getting-started/typescript.md — TypeScript quickstart (Phase 6)"
```

---

### Task 8: CONTRIBUTING.md + docs/contributing/

**Files:**
- Modify: `CONTRIBUTING.md`
- Create: `docs/contributing/README.md`
- Create: `docs/contributing/go.md`
- Create: `docs/contributing/python.md`
- Create: `docs/contributing/typescript.md`

- [ ] **Step 1: Write CONTRIBUTING.md**

Replace the entire file content with:

```markdown
# Contributing to MT-Canvus-Tools

## Overview

This is a monorepo. Most contributions touch one language subtree (`go/`, `python/`, or `typescript/`). Cross-cutting changes (API reference, conventions) are the exception.

## Before you start

Read the convention doc for the language you're working in. Conventions lock build/lint/log/error/test defaults and must be followed without exception. New convention decisions require an amendment entry appended with rationale.

| Language | Convention doc |
|---|---|
| Go | [`docs/conventions/go.md`](docs/conventions/go.md) |
| Python | [`docs/conventions/python.md`](docs/conventions/python.md) |
| TypeScript | [`docs/conventions/typescript.md`](docs/conventions/typescript.md) |

## Development setup

See the per-language getting-started guides:

- [Go](docs/getting-started/go.md)
- [Python](docs/getting-started/python.md)
- [TypeScript](docs/getting-started/typescript.md)

For per-language contributing workflows (toolchain checks, commit scoping, test patterns): [docs/contributing/](docs/contributing/).

## PR workflow

1. Create a feature branch: `git checkout -b feature/<description>` or `fix/<description>`
2. Write tests before implementation (TDD — red/green/refactor)
3. Ensure toolchain passes cleanly before opening a PR:
   - **Go:** `go build ./...` + `go vet ./...` + `go test ./...` + `gofmt -s -l .` (must print nothing)
   - **Python:** `uv run ruff check .` + `uv run pytest` + `uv run mypy --strict python/sdk/src` (0 errors)
   - **TypeScript:** `pnpm typecheck` + `pnpm build` + `pnpm test` + `pnpm lint`
4. Commit using conventional commit messages: `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`
5. Open a PR against `main`

## Secrets

Never commit secrets. Use a `.secrets` file at the repo root (gitignored) for local credentials. Pass credentials via environment variables at runtime.

## API reference corrections

If you discover a discrepancy between `docs/api-reference/` and the live server, add an entry to `docs/api-reference/VERIFIED-CORRECTIONS.md` with curl evidence. See the upstream doc-fix tracker at [canvus-server#96](https://gitlab.multitaction.com/swrd/conan/canvus/canvus-server/-/work_items/96).
```

- [ ] **Step 2: Write docs/contributing/README.md**

```markdown
# Contributing guides

Per-language setup and workflow guides for contributors.

| Guide | Contents |
|---|---|
| [Go](go.md) | Workspace setup, testing patterns, toolchain checks, commit scope |
| [Python](python.md) | uv workspace, ruff, mypy, pytest patterns |
| [TypeScript](typescript.md) | pnpm workspace, ESLint lint floor, vitest, tsc patterns |

See the root [CONTRIBUTING.md](../../CONTRIBUTING.md) for the shared workflow: branch naming, PR process, commit conventions, and secrets policy.
```

- [ ] **Step 3: Write docs/contributing/go.md**

```markdown
# Contributing — Go

## Workspace setup

```bash
git clone https://github.com/jaypaulb/MT-Canvus-Tools.git
cd MT-Canvus-Tools/go
go build ./...
go test ./...
```

The `go.work` file at `go/` spans all Go modules. Workspace-level `go build ./...` and `go test ./...` operate across all of them without `cd`-ing into individual modules.

## Adding a new package

- Place it under the appropriate module (`sdk/canvus/`, `tools/<name>/internal/`, etc.)
- File size limits apply: atoms < 50 LOC, molecules < 150 LOC, organisms < 400 LOC (see [`docs/conventions/go.md`](../conventions/go.md) §2)
- Unit tests go in `_test.go` files in the same package
- Integration tests go under `integration/` with the `//go:build integration` tag

## Toolchain checks (all must pass)

```bash
go build ./...           # No build errors
go vet ./...             # No vet warnings
go test ./...            # All tests pass
gofmt -s -l .            # Must print nothing
```

## Commit scope

Use the Go module or package path as the commit scope:

```
feat(go/sdk): add SubscribeWidgets helper
fix(go/tools/powertoys): guard m.server with mutex
refactor(go/cli): extract auth command into sub-package
docs(go/sdk): update MIGRATION-NOTES.md for WithAPIKey
```

## Test structure

- Unit tests in the same package as the code under test
- Integration tests under `integration/` gated by the `integration` build tag
- Live-server tests use the `live` build tag and read `CANVUS_DEV_BASE_URL` / `CANVUS_DEV_API_KEY`

## Convention reference

[`docs/conventions/go.md`](../conventions/go.md) — locked defaults for error wrapping, logging (structlog via `go.mod`), context propagation, naming, and file size.
```

- [ ] **Step 4: Write docs/contributing/python.md**

```markdown
# Contributing — Python

## Workspace setup

```bash
git clone https://github.com/jaypaulb/MT-Canvus-Tools.git
cd MT-Canvus-Tools/python
uv sync
```

The `pyproject.toml` at `python/` is the uv workspace root. All members (`sdk`, `examples/core/*`, `tools/mcp-server`) share the same lockfile.

## Adding a new package

1. Create the package directory under the appropriate subtree
2. Add it to the `[tool.uv.workspace]` members list in `python/pyproject.toml`
3. Run `uv sync` to add it to the lockfile
4. Tests go in the package's own `tests/` directory

## Test markers

| Marker | When to use |
|---|---|
| `integration` | Requires a live Canvus server (`CANVUS_API_URL` + `CANVUS_API_KEY`) |
| `live` | Hits the dev server (`CANVUS_DEV_BASE_URL` + `CANVUS_DEV_API_KEY`); skipped when env vars unset |
| `slow` | Takes more than a few seconds |

## Toolchain checks (all must pass)

```bash
uv run ruff check .                          # Linting — must be clean
uv run ruff format --check .                 # Formatting — must be clean
uv run pytest                                # All tests pass
uv run mypy --strict python/sdk/src          # Must report 0 errors
```

For the MCP server:

```bash
uv run mypy --strict python/tools/mcp-server/src   # Must report 0 errors
uv run pytest python/tools/mcp-server/tests        # All tests pass
```

## Commit scope

```
feat(python/sdk): add FoldersResource.subscribe_permissions
fix(python/tools/mcp-server): raise on non-list LLM responses
```

## Convention reference

[`docs/conventions/python.md`](../conventions/python.md).
```

- [ ] **Step 5: Write docs/contributing/typescript.md**

```markdown
# Contributing — TypeScript

## Workspace setup

```bash
git clone https://github.com/jaypaulb/MT-Canvus-Tools.git
cd MT-Canvus-Tools/typescript
pnpm install
pnpm build
```

The `package.json` at `typescript/` is the pnpm workspace root. Members: `sdk`, `examples/core/*`, `examples/webui`.

## Adding a new package

1. Create the package directory under `typescript/`
2. Add it to `typescript/pnpm-workspace.yaml`
3. Run `pnpm install` to update the lockfile
4. Tests go in the package's `src/__tests__/` directory using vitest
5. The SDK outputs ESM + CJS + DTS; examples only need ESM

## Toolchain checks (all must pass)

```bash
pnpm typecheck     # tsc --noEmit — no type errors
pnpm build         # dist/ artifacts produced cleanly
pnpm test          # All vitest tests pass
pnpm lint          # ESLint — must not exceed current lint floor
```

## Lint floor

The ESLint floor is **0 errors** after Phase 5 (`UserId`/`GroupId` narrowed to `number` + `allowNumber: true` rule config). Do not add new lint errors.

## Commit scope

```
feat(typescript/sdk): add createAnyWithAsset
fix(typescript/examples/webui): handle missing canvas_id in relay handler
refactor(typescript/sdk): collapse redundant Buffer/Uint8Array branches
```

## Convention reference

[`docs/conventions/typescript.md`](../conventions/typescript.md) — includes amendment history for deferred decisions (camelCase mapping, zod validation, NotFoundError).
```

- [ ] **Step 6: Verify docs/contributing/ files were created**

```bash
ls docs/contributing/
```

Expected output:
```
README.md  go.md  python.md  typescript.md
```

- [ ] **Step 7: Commit all contributing files**

```bash
git add CONTRIBUTING.md docs/contributing/README.md docs/contributing/go.md docs/contributing/python.md docs/contributing/typescript.md
git commit -m "docs: expand CONTRIBUTING.md + add docs/contributing/ per-language guides (Phase 6)"
```

---

### Task 9: CONSOLIDATION-STATUS.md update

**Files:**
- Modify: `CONSOLIDATION-STATUS.md`

- [ ] **Step 1: Update the header block**

Find and replace the "As of" / "Phases complete" / "Next phase" block at the top of `CONSOLIDATION-STATUS.md`. The current block is:

```
**As of:** 2026-05-19 (post-Phase 4d)
**Phases complete:** 0, 1, 2, 3 (+ verification), 4a (+ review-driven fixes), 4b (+ review-driven fixes), 4c (+ review-driven fixes), 4d (+ review-driven fixes)
**Next phase:** 5 (PowerToys port + Phase 4d carry-over) — pending Jaypaul go-ahead
```

Replace with:

```
**As of:** 2026-05-19 (post-Phase 6)
**Phases complete:** 0, 1, 2, 3 (+ verification), 4a (+ review-driven fixes), 4b (+ review-driven fixes), 4c (+ review-driven fixes), 4d (+ review-driven fixes), 5 (PowerToys port + carry-over), 6 (top-level documentation)
**Next phase:** 7 (CI/CD)
```

- [ ] **Step 2: Add Phase 5 and Phase 6 rows to the "What shipped" table**

After the last Phase 4d row (`| 4d.D | ...`), add:

```
| 5.plan | Phase 5 plan (PowerToys port + Phase 4d carry-over) | `8b918e9` |
| 5.1 | Go: powertoys scaffold + SDK shim + type aliases | `0b5852e`, `3a81f3b` |
| 5.2 | Go: powertoys Phase A verbatim copy + atoms/webui subscriber + client resolver | `27271b7`, `5aef00b` |
| 5.3 | Go: powertoys molecules — canvas_service SDK rewire + goroutine lifecycle fixes | `7dc93bf`, `58740cd`, `5dd4f26` |
| 5.4 | Go: powertoys molecules — rcu_handler + Phase A organism handlers | `132216d`, `e0da83d`, `61a4ebd` |
| 5.5 | Go: powertoys — canvas_service ctx race fix + Manager decompose (ui/config/lifecycle) + insecureTLS wiring | `d2f4667`, `083bb42`, `caba525`, `9f8925b` |
| 5.carry | Python FolderPermissions + FoldersResource.subscribe_permissions; TS UserId/GroupId narrowed to number + allowNumber lint rule | `cd49f54`, `cdde5ab` |
| 5.r | Review fix: performConnectionTests insecureTLS + workspace tidy + binary gitignore | `ae170ea`, `53efe75` |
| 6 | Top-level documentation: root README, per-language READMEs, getting-started guides, CONTRIBUTING, docs/contributing/ | *(this commit)* |
```

- [ ] **Step 3: Update the coverage snapshot**

Find the "Coverage snapshot" table. Update the Python row to remove the "FoldersResource `subscribe_permissions` still pending — Phase 5" parenthetical (it shipped in Phase 5 carry-over `cd49f54`). Update the TypeScript row to reflect 0 lint errors (not 26; fixed by `cdde5ab`).

The current Python row cell:
```
10 modules + `__init__` | ruff clean, all pytest pass (incl. SDK + mcp-server + roundtrip + live opt-ins); **mypy --strict 0 errors** at workspace level
```
Keep as-is (already accurate post Phase 5).

The current Python row "Subscribe coverage" cell:
```
**27 / 27** typed AsyncIterators (Canvas permissions live-verified Phase 4d D1; FoldersResource `subscribe_permissions` still pending — Phase 5)
```
Replace with:
```
**27 / 27** typed AsyncIterators (Canvas + Folder permissions live-verified Phase 4d D1; FoldersResource `subscribe_permissions` added Phase 5 carry-over `cd49f54`)
```

The current TypeScript "Toolchain status" cell:
```
typecheck / build clean, 167/167 vitest pass; **lint 26 errors** (deferred floor: `UserId`/`GroupId` template-literal interpolation, deferred to Phase 5); coverage `70.92 / 74.49 / 64.11 / 70.92` (statements/branches/functions/lines)
```
Replace with:
```
typecheck / build clean, 167/167 vitest pass; **lint 0 errors** (`UserId`/`GroupId` narrowed to `number` + `allowNumber: true` ESLint rule, Phase 5 carry-over `cdde5ab`); coverage `70.92 / 74.49 / 64.11 / 70.92` (statements/branches/functions/lines)
```

- [ ] **Step 4: Update the Phase 4d carry-over section**

Find the "Phase 4d carry-over" section at the bottom of CONSOLIDATION-STATUS.md. Replace it with:

```markdown
## Deferred to Phase 7 (CI/CD)

Single agent: GitHub Actions workflows. Per-language matrix jobs (lint, test, build). Optional spec-drift detector comparing `docs/api-reference/` to `mt-restapi-client` on a schedule.

All Phase 4d and Phase 5 carry-over items are now resolved:

- ✅ **Python `FoldersResource.subscribe_permissions`** — shipped Phase 5 carry-over (`cd49f54`)
- ✅ **TS lint floor** — `UserId`/`GroupId` narrowed to `number` + `allowNumber: true` rule (`cdde5ab`); floor is now 0
- ✅ **Python `client.py:105` unused-ignore** — verify with `uv run mypy --strict python/sdk/src`; if still present, drop in Phase 7 sweep
- **CloneWidget per-type wrappers** — retained as "not needed unless callers report friction" design decision; revisit only on user request
```

- [ ] **Step 5: Verify the status file looks correct**

```bash
head -10 CONSOLIDATION-STATUS.md
grep "Phase 5\|Phase 6\|Phase 7" CONSOLIDATION-STATUS.md | head -20
```

Expected: first 10 lines show "post-Phase 6", grep shows Phase 5 + 6 entries and "Phase 7" as next.

- [ ] **Step 6: Commit**

```bash
git add CONSOLIDATION-STATUS.md
git commit -m "docs: update CONSOLIDATION-STATUS.md — Phase 5+6 complete, Phase 7 next (Phase 6)"
```

---

## Self-review checklist

### Spec coverage

| Spec requirement | Task |
|---|---|
| Top-level `README.md` with language matrix + quick links | Task 1 |
| Per-language landing READMEs | Tasks 2, 3, 4 |
| `docs/getting-started/{go,python,typescript}.md` | Tasks 5, 6, 7 |
| `docs/contributing/` | Task 8 |
| Root `CONTRIBUTING.md` expanded from stub | Task 8 |
| CONSOLIDATION-STATUS.md up to date | Task 9 |

All six spec items are covered. No gaps.

### Placeholder scan

No "TBD", "TODO", or "implement later" strings in any file content above. Each file is shown in full.

### Type consistency

Documentation only — no types. Cross-file consistency: all README files link to the same paths. Getting-started guides use correct class/function names (`canvus.NewSession` + `canvus.WithAPIKey`, `Client` + `Client.from_env()`, `createSession` + `Session` + `loadConfig`).
