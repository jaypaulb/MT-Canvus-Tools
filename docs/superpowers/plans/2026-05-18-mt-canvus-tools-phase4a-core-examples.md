# MT-Canvus-Tools Phase 4a — Core Examples Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author the 8 canonical Canvus examples in each of Go, Python, and TypeScript (24 total small projects), each self-contained, runnable end-to-end against `dev-mtcs.multitaction.com`, and exercising the API surface the freshly-verified SDKs now cover.

**Architecture:** Foreground orchestrator runs SDK toolchain bring-up (so agents inherit working dep trees), then dispatches three parallel general-purpose agents (one per language, `opus` model for design care). Each agent owns its `<lang>/examples/core/` subtree exclusively and writes all 8 examples in dependency order (01 → 08). Orchestrator then runs per-language build/typecheck, commits per language, and dispatches a code-review subagent over the diff.

**Tech Stack:** Go workspace + go.work; Python uv workspace; TypeScript pnpm workspace; dotenv-style env loading (per-language idiomatic); Ollama (Example 06 only); standard HTTP for Example 07 webhook delivery.

**Reference design spec:** `/home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/superpowers/specs/2026-05-17-mt-canvus-tools-consolidation-design.md` (Phase 4a section, line 219).

**Verified live server creds:** `/home/jaypaulb/Projects/gh/MT-Canvus-Tools/.secrets` (gitignored). Contains `CANVUS_DEV_BASE_URL` and `CANVUS_DEV_API_KEY`. Test canvas-id: `e36c286c-d447-4b99-a956-678dc118c774` (has notes; safe to read).

---

## The 8 Canonical Examples (per spec, mirrored across all 3 languages)

| # | Slug | Purpose | Key SDK surface exercised |
|---|---|---|---|
| 01 | `auth-and-list` | Authenticate, list canvases | Session/Client constructor + canvas list |
| 02 | `auth-flows` | API key vs email/password login vs access-token lifecycle | Auth resource + token CRUD |
| 03 | `widget-crud` | Create / update / delete a sticky note | Notes resource (full CRUD) |
| 04 | `file-upload` | Upload an image and place it as a widget | Assets resource (multipart) + Images resource |
| 05 | `streaming` | Subscribe to a canvas's widgets, react to events | Streaming subscribe + NDJSON parse |
| 06 | `llm-integration` | Paired watcher + responder via Ollama | Streaming + Notes create + external HTTP |
| 07 | `webhooks-notifications` | Outbound notification pattern (subscribe → POST to user-configured webhook) | Streaming + outbound HTTP |
| 08 | `cross-canvas-clone` | Clone a widget across canvases per changelog §1 | Per-type clone helper (the SDK feature added in Phase 3) |

---

## File Structure (created by this plan)

```
MT-Canvus-Tools/
├── go/
│   ├── go.work                              # MODIFIED: extend `use` to include examples
│   └── examples/core/
│       ├── 01-auth-and-list/{main.go, go.mod, README.md, .env.example}
│       ├── 02-auth-flows/{main.go, go.mod, README.md, .env.example}
│       ├── 03-widget-crud/{main.go, go.mod, README.md, .env.example}
│       ├── 04-file-upload/{main.go, go.mod, README.md, sample.png, .env.example}
│       ├── 05-streaming/{main.go, go.mod, README.md, .env.example}
│       ├── 06-llm-integration/{cmd/watcher/main.go, cmd/responder/main.go, go.mod, README.md, .env.example}
│       ├── 07-webhooks-notifications/{main.go, go.mod, README.md, .env.example}
│       └── 08-cross-canvas-clone/{main.go, go.mod, README.md, .env.example}
├── python/
│   ├── pyproject.toml                       # MODIFIED: add examples to uv workspace members
│   └── examples/core/
│       ├── 01-auth-and-list/{main.py, pyproject.toml, README.md, .env.example}
│       ├── 02-auth-flows/...
│       ├── ... (same structure as Go)
│       └── 06-llm-integration/{watcher.py, responder.py, pyproject.toml, README.md, .env.example}
└── typescript/
    ├── pnpm-workspace.yaml                  # MODIFIED: add examples/core/* to workspace
    └── examples/core/
        ├── 01-auth-and-list/{src/index.ts, package.json, tsconfig.json, README.md, .env.example}
        ├── ... (same structure)
        └── 06-llm-integration/{src/watcher.ts, src/responder.ts, package.json, tsconfig.json, README.md, .env.example}
```

---

## Concurrency Model

Three agents own non-overlapping subtrees:
- Go agent: `MT-Canvus-Tools/go/` (writes under `examples/core/`, modifies only `go.work`)
- Python agent: `MT-Canvus-Tools/python/` (writes under `examples/core/`, modifies only `pyproject.toml`)
- TypeScript agent: `MT-Canvus-Tools/typescript/` (writes under `examples/core/`, modifies only `pnpm-workspace.yaml`)

No worktree isolation needed — distinct subtrees. Orchestrator handles all `git add`, `git commit`, `git push`.

---

## Task 4a.0: Pre-flight checks

- [ ] **Step 1: Verify HEAD is the post-verification commit and tree is clean**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
git log -1 --oneline
git status --short
```

Expected: HEAD is at least `115ad3f` ("fix(sdks): resolve Tier-1/Tier-2 deltas") and working tree is clean.

If dirty, stop and surface to Jaypaul.

- [ ] **Step 2: Verify Go SDK builds + tests**

```bash
go build -C /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go ./sdk/...
go vet -C /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go ./sdk/...
go test -C /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go ./sdk/canvus/...
```

Expected: all three commands exit 0, tests pass.

- [ ] **Step 3: Verify .secrets is readable and has expected keys**

```bash
test -r /home/jaypaulb/Projects/gh/MT-Canvus-Tools/.secrets && \
  grep -q "^CANVUS_DEV_BASE_URL=" /home/jaypaulb/Projects/gh/MT-Canvus-Tools/.secrets && \
  grep -q "^CANVUS_DEV_API_KEY=" /home/jaypaulb/Projects/gh/MT-Canvus-Tools/.secrets && \
  echo "secrets OK"
```

Expected: `secrets OK` printed.

- [ ] **Step 4: Verify test canvas is reachable**

```bash
set -a; source /home/jaypaulb/Projects/gh/MT-Canvus-Tools/.secrets; set +a
curl -sS -o /dev/null -w "%{http_code}\n" \
  -H "Private-Token: $CANVUS_DEV_API_KEY" \
  "${CANVUS_DEV_BASE_URL}canvases/e36c286c-d447-4b99-a956-678dc118c774"
```

Expected: `200`.

---

## Task 4a.1: Bring Python and TypeScript SDK dep trees online

(Go was already brought online during Phase 3 — `go mod tidy` ran successfully.)

- [ ] **Step 1: Python uv workspace sync**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/python && uv sync 2>&1 | tail -20
```

Expected: lockfile created/updated, no resolution errors. If `uv` is not installed, surface to Jaypaul — do NOT fall back to `pip`.

If sync fails on a dependency conflict, stop and surface to Jaypaul.

- [ ] **Step 2: TypeScript pnpm workspace install**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript && pnpm install 2>&1 | tail -20
```

Expected: lockfile created/updated, no peer-dep warnings that look like errors. If `pnpm` is not installed, surface to Jaypaul.

- [ ] **Step 3: TypeScript SDK typecheck**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript && pnpm --filter '@mt-canvus-tools/sdk' exec tsc --noEmit 2>&1 | head -30
```

Expected: zero output (clean typecheck). If errors, stop and surface — Phase 4a examples will fail without a typechecking SDK.

- [ ] **Step 4: Commit lockfiles if newly generated**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
git add python/uv.lock typescript/pnpm-lock.yaml 2>/dev/null || true
git status --short
```

If new lockfiles staged:

```bash
git commit -m "chore: generate Python uv.lock and TypeScript pnpm-lock.yaml for SDK dep trees"
git push origin main
```

---

## Task 4a.2: Dispatch three example-authoring agents in parallel

Send a single assistant message with three `Agent` tool calls. Use `subagent_type: general-purpose`, `model: opus`, `run_in_background: true`.

### Agent 4a.1 prompt — Go examples

```
You are authoring the 8 canonical Canvus examples in Go. Jaypaul is awake but waiting on three parallel agents — be precise, idiomatic, and complete. Do NOT ask questions; use your judgement and document non-obvious choices in the per-example README.

# Strict boundary
You may modify ONLY these paths:
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go/go.work
- Anything under /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go/examples/

Do NOT touch /go/sdk/, /python/, /typescript/, /docs/, or the repo root.

# Inputs to read
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/conventions/go.md
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/VERIFIED-CORRECTIONS.md (wire-shape truth)
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/changelog.md
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go/sdk/canvus/*.go (consume this SDK)
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go/sdk/MIGRATION-NOTES.md (decision context)

# Output structure
For each example slug, create directory go/examples/core/<slug>/ containing:
- main.go (or cmd/<bin>/main.go for example 06)
- go.mod (module github.com/jaypaulb/MT-Canvus-Tools/go/examples/core/<slug>, go 1.22, toolchain go1.23.0)
- README.md (purpose, prerequisites, env-var reference, run command, expected output, walkthrough)
- .env.example (env var template — never include actual creds)

After creating all eight, modify /go/go.work to add each example to the `use` directive list.

# Workspace integration
Final go.work shape:
```
go 1.22
toolchain go1.23.0
use (
    ./sdk
    ./examples/core/01-auth-and-list
    ./examples/core/02-auth-flows
    ./examples/core/03-widget-crud
    ./examples/core/04-file-upload
    ./examples/core/05-streaming
    ./examples/core/06-llm-integration
    ./examples/core/07-webhooks-notifications
    ./examples/core/08-cross-canvas-clone
)
```

Each example's go.mod requires the SDK via `github.com/jaypaulb/MT-Canvus-Tools/go/sdk v0.0.0-00010101000000-000000000000` (workspace replaces the version).

# Conventions (per docs/conventions/go.md)
- slog for logging (text handler dev, JSON handler if LOG_FORMAT=json)
- context.Context on every I/O
- Errors wrapped with fmt.Errorf("operation: %w", err)
- Env vars via os.Getenv with explicit fallback/error if missing
- Sentinel errors imported from sdk where applicable
- main() returns via run() pattern: `func main() { if err := run(); err != nil { slog.Error("...", "err", err); os.Exit(1) } }`
- godoc comments on exported symbols

# Env var conventions (all examples)
- CANVUS_BASE_URL (required) — e.g. https://dev-mtcs.multitaction.com/api/v1/
- CANVUS_API_KEY (required for examples 01, 03, 04, 05, 06, 07, 08)
- CANVUS_EMAIL + CANVUS_PASSWORD (required for example 02 login path)
- CANVUS_CANVAS_ID (required for examples 03, 05, 06, 07; optional for 04 — defaults to first available canvas with write access)
- CANVUS_DEST_CANVAS_ID + CANVUS_SOURCE_WIDGET_ID (required for example 08)
- CANVUS_IMAGE_PATH (optional for example 04 — defaults to bundled sample.png)
- OLLAMA_URL + OLLAMA_MODEL (required for example 06 — default http://localhost:11434 and llama3.2)
- WEBHOOK_URL (required for example 07)
- LOG_FORMAT (optional: "text" default, "json" for prod)

Load env via standard library os.Getenv; do NOT add a dotenv dependency (idiomatic Go is plain env).

# Per-example specifications

## 01-auth-and-list
**Purpose:** Smoke test: authenticate with API key, list canvases, print summary table.

**main.go skeleton:**
```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"
    "text/tabwriter"

    "github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func main() {
    if err := run(); err != nil {
        slog.Error("auth-and-list failed", "err", err)
        os.Exit(1)
    }
}

func run() error {
    baseURL := mustEnv("CANVUS_BASE_URL")
    apiKey := mustEnv("CANVUS_API_KEY")

    s := canvus.NewSession(baseURL, canvus.WithAPIKey(apiKey))
    ctx := context.Background()

    canvases, err := s.ListCanvases(ctx)
    if err != nil {
        return fmt.Errorf("list canvases: %w", err)
    }

    w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
    fmt.Fprintln(w, "ID\tNAME\tMODE\tACCESS\tSTATE")
    for _, c := range canvases {
        fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", c.ID, c.Name, c.Mode, c.Access, c.State)
    }
    return w.Flush()
}

func mustEnv(k string) string {
    v := os.Getenv(k)
    if v == "" {
        slog.Error("missing required env var", "name", k)
        os.Exit(2)
    }
    return v
}
```

**README sections:** Purpose, Prerequisites, Configuration, Run, Expected Output (sample), How It Works (3-paragraph walkthrough), Troubleshooting (401 = wrong key; 404 = wrong base URL; empty list = no read access).

## 02-auth-flows
**Purpose:** Demonstrate three auth modes side-by-side:
1. **API-key** path (Private-Token header)
2. **Email + password login** path (POST /users/login → token)
3. **Access-token CRUD** (create + list + delete a programmatic token tied to the logged-in user)

Each path prints which auth mode succeeded and one trivial authenticated call (`/users/{id}` for the logged-in user). Use separate `runAPIKey`, `runLogin`, and `runTokenLifecycle` functions, called in sequence. Each function logs success or failure but the program proceeds to the next mode so a partial auth setup is still demonstrable. Login uses `email` field ONLY (no `username` — verified breaks server).

## 03-widget-crud
**Purpose:** Full lifecycle on a sticky note.

Steps:
1. Create a note on CANVUS_CANVAS_ID with `text: "hello from Go @ <ISO timestamp>"`, default colour
2. Print the new widget id
3. Patch the note: `text: "updated from Go @ <ISO timestamp>"`, `background-color: "#3aaa34ff"`
4. Print the updated widget
5. Delete the note
6. Verify delete by GET → expect NotFound

Use the SDK's notes resource. On error in any step, print the error and DO NOT attempt subsequent steps (so the user can debug the failing step).

## 04-file-upload
**Purpose:** Upload a local PNG and place it as an image widget on the canvas.

Steps:
1. Open file at CANVUS_IMAGE_PATH (default ./sample.png — write a tiny 256x256 PNG sample.png with a solid colour using image/png stdlib so the example is self-contained)
2. Upload via SDK's multipart upload (POST /canvases/{id}/images — multipart with the file in `data` field)
3. Server returns the image widget object with id and asset hash
4. Patch the image to position {x: 100, y: 200}, scale 0.5
5. Print the widget id and the on-screen position

If CANVUS_KEEP_WIDGET is set to "1", leave the widget on the canvas; otherwise delete it as cleanup (with a defer).

## 05-streaming
**Purpose:** Subscribe to a canvas's notes endpoint, print events for the configured duration.

Steps:
1. Open subscribe stream: GET /canvases/{id}/notes?subscribe=true
2. Set up signal.Notify for SIGINT — context cancellation drains and exits cleanly
3. Loop: read NDJSON line, parse to []Note, print summary (count, first-id, first-text)
4. After STREAM_DURATION_SECONDS (default 30) elapses or SIGINT received, cancel context, close stream, print final stats

Document the keep-alive line behaviour observed in the SDK — empty newlines should be skipped silently.

## 06-llm-integration
**Purpose:** Paired apps demonstrating Canvus + Ollama.

**Two binaries:**
- `cmd/watcher/main.go`: subscribes to notes on CANVUS_CANVAS_ID. When a NEW note appears whose text starts with `?` (a "question note"), POST `{model, prompt: <note text minus leading ?>, stream: false}` to `OLLAMA_URL/api/generate`. Capture the response. Then add the response as a sibling note positioned right of the question note (offset +400px x). The new "answer note" gets `background_color: "#1d71b8ff"`.
- `cmd/responder/main.go`: lightweight diagnostic binary that polls `OLLAMA_URL/api/tags` and prints available models. Used to verify Ollama is reachable before running watcher.

README explains: prerequisites (Ollama running locally), how to start watcher in one terminal, how to use responder to verify Ollama, expected behaviour (write `?What is 2+2?` as a note in Canvus and watch the answer appear).

## 07-webhooks-notifications
**Purpose:** Bridge Canvus events to an outbound HTTP webhook (the Canvus API does not have built-in webhooks; this example demonstrates the *pattern* for building one on top of streaming subscribe).

Steps:
1. Subscribe to a canvas's widgets endpoint
2. For each new widget event (filtered: only "created" — initial snapshot ignored), POST to WEBHOOK_URL with a JSON body `{event: "widget.created", canvas_id, widget_id, widget_type, timestamp}`
3. Log each POST: status code, latency
4. Backoff + retry on non-2xx (up to 3 attempts with 2s/4s/8s delays)

Tip: for testing, point WEBHOOK_URL at https://webhook.site/ — the README should suggest this.

## 08-cross-canvas-clone
**Purpose:** Demonstrate the per-type clone helper (the SDK feature added in Phase 3 per changelog §1).

Steps:
1. GET widget CANVUS_SOURCE_WIDGET_ID from CANVUS_CANVAS_ID (call this `srcWidget`)
2. Determine the widget type from `srcWidget.widget_type`
3. Use the corresponding SDK clone helper (e.g. `s.CloneNote(ctx, srcCanvasID, srcWidgetID, destCanvasID)`). If the SDK has a dispatching helper `s.CloneWidget(...)`, use that and document its lookup logic.
4. Print the new widget id on the destination canvas
5. If CANVUS_CLEANUP=1, delete the cloned widget after a 2-second pause

The README must explicitly call out that the legacy `POST /api/v1/canvases/{id}/widgets/clone` endpoint returns 501 and is deliberately not used.

# Output cadence (per example)
- main.go: implements the spec above with the conventions baked in.
- README.md: ~80-120 lines per example (purpose, prerequisites, configuration table, run command, expected output sample, walkthrough, troubleshooting).
- go.mod: minimal — `github.com/jaypaulb/MT-Canvus-Tools/go/sdk` requirement + any small deps (e.g. example 06 may need standard `net/http`; nothing exotic).
- .env.example: every env var the example reads, with descriptive comments and sample values (no real creds).

# Final deliverable: examples-summary
Create /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go/examples/core/README.md indexing the eight examples in a table (slug, purpose, key env vars, complexity rating 1-5 stars), plus a "Running the examples" section that explains the env-loading pattern, where to get .env values, and a `direnv`-style suggestion for local dev.

# Critical rules
- Do NOT run git commands. Do NOT commit.
- Do NOT run `go build` / `go test` on the examples — orchestrator does that.
- Each example MUST compile in principle. Use the EXACT SDK names from /go/sdk/canvus/. If you can't find a method, surface in your completion report rather than inventing one.
- Aim for ~1500-2500 lines of Go across all 8 examples (200-300 lines each, except 06 which is bigger).

# Completion report (one structured response)
- Files written (count per example)
- Total lines of Go across all examples
- SDK methods used per example (so we can spot any imagined methods)
- Any compilation risks (uncertainty about a method signature)
- Confirmation that go.work is updated with all 8 new use directives
- Top 3 architectural choices made (e.g. how 06 watcher dedupes already-seen notes)
```

### Agent 4a.2 prompt — Python examples

```
You are authoring the 8 canonical Canvus examples in Python. Jaypaul is awake but waiting on three parallel agents — be precise, idiomatic, and complete. Do NOT ask questions; use your judgement and document non-obvious choices in the per-example README.

# Strict boundary
You may modify ONLY these paths:
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/python/pyproject.toml (add examples to workspace members)
- Anything under /home/jaypaulb/Projects/gh/MT-Canvus-Tools/python/examples/

Do NOT touch /python/sdk/, /go/, /typescript/, /docs/, or the repo root.

# Inputs to read
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/conventions/python.md
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/VERIFIED-CORRECTIONS.md
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/changelog.md
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/python/sdk/src/canvus_sdk/ (consume this SDK)
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/python/sdk/MIGRATION-NOTES.md

# Output structure
For each example slug, create directory python/examples/core/<slug>/ containing:
- main.py (or watcher.py + responder.py for example 06)
- pyproject.toml (uv-compatible, depends on canvus-sdk via workspace path)
- README.md
- .env.example

After creating all eight, modify /python/pyproject.toml's `[tool.uv.workspace] members` array to include each example.

# Conventions (per docs/conventions/python.md)
- Python 3.11+ syntax: `X | Y`, `list[T]`, `dict[K, V]`, `Self`, no `typing.List`
- structlog with env-driven renderer
- pydantic-settings for env via CANVUS_ prefix
- httpx for any non-SDK HTTP (example 06 Ollama, example 07 webhook)
- Async-first: prefer `async def main()` invoked via `asyncio.run`
- Type-annotate everything
- Module docstring at top of each main.py describing the example

# Env var conventions (same as Go example agent — see that agent's prompt for the full list)
Use pydantic-settings BaseSettings with prefix CANVUS_ for canvus vars; load other vars (OLLAMA_*, WEBHOOK_URL, LOG_FORMAT, STREAM_DURATION_SECONDS) via os.environ directly.

# Per-example specifications

(All 8 examples have the SAME purpose and step list as the Go agent's prompt — see the design spec at /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/superpowers/specs/2026-05-17-mt-canvus-tools-consolidation-design.md and the Phase 4a plan at /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/superpowers/plans/2026-05-18-mt-canvus-tools-phase4a-core-examples.md for the per-example specs.)

Key idiomatic Python deltas from the Go versions:
- Use `async with client.session() as session:` for context-managed lifecycle
- Pretty-print canvas listings via `rich` table OR plain `print` with `f"{c.id:<40} {c.name:<30} {c.mode}"` — pick one and be consistent. Do NOT add `rich` as a dep unless you use it everywhere; prefer plain print for examples to keep dep tree minimal.
- Example 04 (file-upload): use httpx multipart through the SDK's upload method; generate sample.png via Pillow if needed — but prefer pure-stdlib via `struct.pack` + bytes for a minimal valid PNG so we don't add Pillow as a dep just for one example
- Example 05 (streaming): use `async for event in session.canvases.subscribe(canvas_id):` if the SDK exposes that; otherwise call the raw stream method
- Example 06 (llm-integration): two separate .py files (watcher.py + responder.py), entry points declared in pyproject.toml as `[project.scripts]`
- Example 08: use the SDK's clone helper exactly as Go does

# Output cadence per example
- main.py / watcher.py / responder.py: implement spec, idiomatic 3.11+ Python, fully typed
- pyproject.toml: minimal, build-system = hatchling, deps = canvus-sdk + httpx (for examples 06, 07) + structlog
- README.md: same sections as the Go README, just Python-flavoured
- .env.example: same env var names as Go examples (already shared CANVUS_ prefix)

# Workspace integration
In /home/jaypaulb/Projects/gh/MT-Canvus-Tools/python/pyproject.toml, ensure `[tool.uv.workspace] members` lists:
```toml
members = [
    "sdk",
    "examples/core/01-auth-and-list",
    "examples/core/02-auth-flows",
    "examples/core/03-widget-crud",
    "examples/core/04-file-upload",
    "examples/core/05-streaming",
    "examples/core/06-llm-integration",
    "examples/core/07-webhooks-notifications",
    "examples/core/08-cross-canvas-clone",
]
```

Each example's pyproject.toml depends on `canvus-sdk` via:
```toml
[tool.uv.sources]
canvus-sdk = { workspace = true }
```

# Final deliverable: examples-summary
Create /home/jaypaulb/Projects/gh/MT-Canvus-Tools/python/examples/core/README.md indexing all 8 examples.

# Critical rules
- Do NOT run git commands. Do NOT commit.
- Do NOT run `uv sync` / `pytest` / `mypy` — orchestrator handles.
- Use EXACT SDK method names from /python/sdk/src/canvus_sdk/. Don't invent methods.
- Aim for ~1200-2200 lines of Python across all 8 examples.

# Completion report
- Files written (count per example)
- Total lines of Python
- SDK methods used per example
- Any imagined-method risks
- Confirmation that python/pyproject.toml workspace members are updated
- Top 3 architectural choices
```

### Agent 4a.3 prompt — TypeScript examples

```
You are authoring the 8 canonical Canvus examples in TypeScript. Jaypaul is awake but waiting on three parallel agents — be precise, idiomatic, and complete. Do NOT ask questions; use your judgement and document non-obvious choices in the per-example README.

# Strict boundary
You may modify ONLY these paths:
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript/pnpm-workspace.yaml
- Anything under /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript/examples/

Do NOT touch /typescript/sdk/, /go/, /python/, /docs/, or the repo root.

# Inputs to read
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/conventions/typescript.md (and the Amendments section)
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/VERIFIED-CORRECTIONS.md
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/changelog.md
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript/sdk/src/ (consume this SDK)
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript/sdk/IMPLEMENTATION-NOTES.md

# Output structure
For each example slug, create directory typescript/examples/core/<slug>/ containing:
- src/index.ts (or src/watcher.ts + src/responder.ts for example 06)
- package.json (name: @mt-canvus-tools/example-<slug>, type: module, depends on @mt-canvus-tools/sdk via workspace:*)
- tsconfig.json (extends ../../../tsconfig.base.json; outDir dist; references the sdk)
- README.md
- .env.example

After creating all eight, modify /typescript/pnpm-workspace.yaml to include each example.

# Conventions (per docs/conventions/typescript.md)
- ESM imports only, named exports
- kebab-case filenames
- Node 20+ — use native fetch where SDK doesn't already wrap (examples 06, 07 use it for Ollama/webhook)
- Async/await everywhere
- zod for env-var parsing (matching the SDK's loadConfig pattern)
- pino for logging (json in prod, pino-pretty in dev driven by LOG_FORMAT)
- TSDoc comments on exported symbols
- Strict TS, no `any`

# Env var conventions (same names as Go/Python agents — shared CANVUS_ prefix)
Load + validate via zod schema. Pattern:
```ts
import { z } from "zod";

const envSchema = z.object({
  CANVUS_BASE_URL: z.string().url(),
  CANVUS_API_KEY: z.string().min(1),
  // ... per-example additions
});

function loadEnv() {
  const parsed = envSchema.safeParse(process.env);
  if (!parsed.success) {
    console.error("Missing or invalid env:", parsed.error.flatten().fieldErrors);
    process.exit(2);
  }
  return parsed.data;
}
```

# Per-example specifications

(Identical purpose + steps as Go and Python agent prompts. See design spec + this plan for details.)

TypeScript-idiomatic deltas:
- Example 04 file-upload: use `fs.readFile` + the SDK's multipart upload method. Bundle a tiny sample.png (write it as a base64 const in the example so users don't need to find one)
- Example 05 streaming: use `for await (const event of session.canvases.subscribe(canvasId))` matching the SDK's async-iterator pattern
- Example 06 llm-integration: TWO TypeScript entry points (src/watcher.ts + src/responder.ts), each exposed via package.json `"bin"` field. Use `tsx` for dev execution; build via `tsup` for distribution.
- Example 08: `await session.widgets.clone({ destCanvasId, sourceCanvasId, sourceWidgetId })`

# Output cadence per example
- src/<entry>.ts: implementation
- package.json: minimal — `@mt-canvus-tools/sdk` workspace dep + pino + zod; example 06 needs nothing extra beyond fetch (Node 20+ native); example 07 same
- tsconfig.json: extends the base, sets outDir/rootDir, includes the sdk via project references
- README.md: same sections as Go/Python
- .env.example: same env names

# Workspace integration
Update /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript/pnpm-workspace.yaml to:
```yaml
packages:
  - sdk
  - examples/core/*
```

(If `examples/core/*` is already there, leave it. Use a glob if the previous file used one.)

# Final deliverable: examples-summary
Create /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript/examples/core/README.md indexing all 8 examples.

# Critical rules
- Do NOT run git commands. Do NOT commit.
- Do NOT run pnpm install / tsc / vitest / eslint — orchestrator handles.
- Use EXACT SDK exported names from /typescript/sdk/src/. Check src/index.ts re-exports first.
- Strict TS — `noUncheckedIndexedAccess` is on. No `!` shortcuts.
- Aim for ~1500-2500 lines of TS across all 8 examples.

# Completion report
- Files written (count per example)
- Total lines of TS
- SDK exports used per example (helps spot imagined APIs)
- Any uncertain SDK method signatures
- Confirmation that pnpm-workspace.yaml includes examples/core
- Top 3 architectural choices
```

- [ ] **Step 1: Send single message with three Agent tool calls**

- [ ] **Step 2: Wait for completion notifications**

You will be notified automatically when each agent completes. Do NOT poll.

---

## Task 4a.3: Verify outputs (inventory)

After all 3 agents return:

- [ ] **Step 1: Per-language file inventory**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
for lang in go python typescript; do
  echo "=== $lang ==="
  find $lang/examples/core -type f | sort
  echo "--- line counts ---"
  find $lang/examples/core -type f \( -name "*.go" -o -name "*.py" -o -name "*.ts" \) -exec wc -l {} + | tail -1
done
```

Expected: each language has ~10-15 files per example (main + go.mod/pyproject.toml/package.json + README + .env.example + maybe tsconfig.json). Lines: Go ~1500-2500, Python ~1200-2200, TS ~1500-2500.

- [ ] **Step 2: Confirm workspace integrations**

```bash
grep -A 10 "use (" /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go/go.work
grep -A 12 "members" /home/jaypaulb/Projects/gh/MT-Canvus-Tools/python/pyproject.toml
cat /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript/pnpm-workspace.yaml
```

Expected: all 8 example dirs listed in each workspace config.

If any language is missing entries, the agent under-delivered. Trigger failure-protocol (retry that agent with a tighter prompt; if still failing on retry 3, open GH issue and continue per the documented protocol).

---

## Task 4a.4: Per-language toolchain checks

- [ ] **Step 1: Go — build all examples**

```bash
go build -C /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go ./examples/core/... 2>&1 | head -50
```

Expected: zero output (clean build). On errors:
1. Read the error, identify the file + line.
2. Determine if it's a typo in the agent's code or a real SDK gap.
3. Fix if surgical (≤5 lines); otherwise dispatch a fix subagent with the precise error.

```bash
go vet -C /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go ./examples/core/... 2>&1 | head -20
```

Then a modernise sweep (same as Phase 3):
```bash
find /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go/examples -name "*.go" | xargs gofmt -r 'interface{} -> any' -w
find /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go/examples -name "*.go" | xargs gofmt -s -w
```

- [ ] **Step 2: Python — uv sync + mypy + ruff**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/python && uv sync 2>&1 | tail -10
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/python && uv run ruff check examples/core/ 2>&1 | tail -20
```

Expected: sync clean, ruff clean (or only warnings that ruff doesn't enforce by default).

Skip mypy for examples (per python.md: "permissive for examples").

- [ ] **Step 3: TypeScript — typecheck all examples**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript && pnpm install 2>&1 | tail -10
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript && pnpm -r --filter './examples/core/*' exec tsc --noEmit 2>&1 | head -50
```

Expected: clean typecheck across all 8 example packages.

- [ ] **Step 4: On any toolchain failure**

Follow Jaypaul's failure protocol (`~/.claude/projects/-home-jaypaulb-Projects-gh/memory/feedback_agent_failure_protocol.md`):
- 3 retries with progressively tighter prompts
- After retries exhausted: open a GH issue + dispatch investigator subagent
- If non-blocking, continue the main run while the investigator works
- Never stop and wait

---

## Task 4a.5: Commit per language

Three commits, one per language, for clean history.

- [ ] **Step 1: Go commit**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
git add go/
git status --short | head -30
```

Verify staged files: `go/go.work` modified, `go/examples/core/**` added.

```bash
git commit -m "feat(go/examples/core): author 8 canonical examples

Phase 4a of consolidation. Adds 8 mirrored canonical examples:
01-auth-and-list, 02-auth-flows, 03-widget-crud, 04-file-upload,
05-streaming, 06-llm-integration (paired watcher+responder via Ollama),
07-webhooks-notifications, 08-cross-canvas-clone.

Each example is a self-contained Go module added to the workspace via
go.work. Run end-to-end against dev-mtcs.multitaction.com using the
shared CANVUS_* env vars (see each example's .env.example)."
```

- [ ] **Step 2: Python commit**

```bash
git add python/
git status --short | head -30
git commit -m "feat(python/examples/core): author 8 canonical examples

Phase 4a of consolidation. Mirrors the Go example set in idiomatic
async Python 3.11+ using structlog, httpx, and the canvus-sdk
workspace member. Examples registered as uv workspace members.

Same env-var contract as the Go and TS example sets."
```

- [ ] **Step 3: TypeScript commit**

```bash
git add typescript/
git status --short | head -30
git commit -m "feat(typescript/examples/core): author 8 canonical examples

Phase 4a of consolidation. Mirrors the Go and Python example sets in
strict TypeScript via the @mt-canvus-tools/sdk workspace package.
Each example builds via tsup; runs via tsx in dev. zod validates env;
pino for logging. Examples added to pnpm-workspace.yaml.

Same env-var contract as the Go and Python example sets."
```

- [ ] **Step 4: Push**

```bash
git push origin main
git log --oneline | head -10
```

---

## Task 4a.6: Optional code-review gate

Per the design spec QA-gate strategy (line 264), dispatch the `pr-review-toolkit:code-reviewer` agent on the Phase-4a diff.

- [ ] **Step 1: Dispatch the reviewer (background, opus)**

```
Agent({
  description: "Review Phase 4a core examples",
  subagent_type: "pr-review-toolkit:code-reviewer",
  prompt: "Review the Phase 4a core examples added in commits between <pre-4a-sha> and <post-4a-sha>. Focus areas: (1) Do the examples actually exercise the SDK they consume? (2) Are env-var handling and error-handling patterns idiomatic per docs/conventions/<lang>.md? (3) Are READMEs accurate to the code (no doc rot)? (4) Do parallel examples (same number across languages) have parity in capability? (5) Any silent failures or fallbacks that hide real errors? Report findings categorised by language, with concrete suggested fixes. Do not modify code yourself.",
  run_in_background: true
})
```

Capture `<pre-4a-sha>` from `git log --oneline | grep '115ad3f' | awk '{print $1}'` (the Phase 3.t commit), `<post-4a-sha>` from the post-Push HEAD.

- [ ] **Step 2: On review completion, triage findings**

For each finding:
- **Critical** (broken example, wrong SDK use): fix inline immediately
- **Important** (style, idiom): batch fixes into a follow-up commit
- **Nice-to-have**: note in Phase 4b backlog

Commit fix(es) as `fix(examples/core): address Phase 4a review findings`.

---

## Task 4a.7: Final report

- [ ] **Step 1: Update CONSOLIDATION-STATUS.md** with Phase 4a outcome

Add a row to the "What shipped" table:

```
| 4a | Core examples: 8 × 3 = 24 projects | <sha> |
```

Update the "Coverage snapshot" section to add an "Examples" column or section noting that every SDK now has 8 working examples.

Add a "Ready for Phase 4b/4c/4d" section noting that the cross-SDK parity sweep + per-item refresh + dead-code cleanup are queued as separate plans.

- [ ] **Step 2: Commit + push status update**

```bash
git add CONSOLIDATION-STATUS.md
git commit -m "docs: Phase 4a completion update in CONSOLIDATION-STATUS"
git push origin main
```

- [ ] **Step 3: Final report to Jaypaul (chat message)**

Brief summary: count of files/lines per language, count of SDK methods exercised, any deferred items, link to GH compare view, recommended next action (Phase 4b plan).

---

## What's next (out of scope for this plan)

This plan covers Phase 4a only. Three sibling plans queued:

- **Plan 4b — Cross-SDK Parity Sweep:** Audit each SDK against the others (Go vs Python vs TS). Port any missing utilities (e.g. Python `geometry.py`, `search.py`, `filters.py`, `export.py`, `widget_operations.py`) into a `sdk-extras` subpackage per language. Same shape: 3 parallel agents.
- **Plan 4c — Per-Item Refresh:** 11 existing tool/utility repos refreshed against the new SDKs (per spec Phase 4b). Batched 4-6 agents per round.
- **Plan 4d — Go dead-code cleanup:** Small foreground task — Chesterton fence inspection then deletion of `setToken`, `doRequestWithHeaders`, `warnOnce`, `warnAlways` (per Tier 3 #10).

Each will be written as its own plan when this one completes, so it can reference Phase 4a's actual artefacts.
