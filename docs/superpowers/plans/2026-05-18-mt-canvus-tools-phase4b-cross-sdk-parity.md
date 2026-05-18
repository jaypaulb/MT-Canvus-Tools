# MT-Canvus-Tools Phase 4b — Cross-SDK Parity Sweep Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring the Go, Python, and TypeScript SDKs to functional parity so consumers in any one language have the same capabilities as the others — and port the substantive client-side helpers (`geometry`, `search`, `filters`, `widget_operations`, `export`) from the legacy `CanvusPythonAPI` into all three SDKs as a shared `extras` surface.

**Architecture:** A single audit agent (opus) produces `docs/api-reference/parity-matrix.md` enumerating every public method/type per SDK and marking missing-from-others. Orchestrator then dispatches three parallel implementer agents (one per language) with their language-specific gap list. Each agent fills its gaps in an `extras` submodule that ships alongside the core SDK. Go examples 05/06/07 are refactored at the end to consume the new typed `Subscribe` helpers (the largest known parity gap going in).

**Tech Stack:** Same as Phases 3 and 4a — Go workspace, Python uv workspace, TypeScript pnpm workspace. New: `extras` submodule pattern per SDK.

**Reference design spec:** `/home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/superpowers/specs/2026-05-17-mt-canvus-tools-consolidation-design.md`
**Source of parity items known going in:** `/home/jaypaulb/Projects/gh/MT-Canvus-Tools/CONSOLIDATION-STATUS.md` (§"Deferred to Phase 4b")
**Live dev creds:** `/home/jaypaulb/Projects/gh/MT-Canvus-Tools/.secrets` (gitignored)

---

## Known parity gaps going in (audit will confirm + expand)

| Gap | Has it | Missing it |
|---|---|---|
| Typed `Subscribe` resource methods (canvases / folders / users / widgets + per-type) | TS (all), Python (widgets only) | Go (none) |
| `geometry` helper module (Point/Size/Rectangle, intersection, widget bounding boxes) | Python legacy | Go, TS |
| `search` helper module (cross-canvas search, find-by-text, find-by-type, find-in-area) | Python legacy | Go, TS |
| `filters` helper module (Filter DSL with operators + spatial/text/wildcard factories) | Python legacy | Go, TS |
| `widget_operations` helper module (spatial tolerance, zone manager, batch ops, clusters, density) | Python legacy | Go, TS |
| `export` helper module (WidgetExporter/Importer, canvas↔folder export/import) | Python legacy | Go, TS |
| `CloneWidget` per-type helper variants | (None — single dispatching helper) | (Likely fine as-is; confirmed during audit) |

Audit task 4b.1 may surface additional gaps; the plan accommodates this by passing the matrix into the implementer prompts at dispatch time.

---

## Decision locked: `extras` submodule pattern

The ported helpers are client-side utilities operating on data returned by the core SDK — they do not make new HTTP calls except where stated (`search`, `export` do; the rest are pure functions). To keep the core SDK lean while still distributing the helpers in one package:

- **Go:** Add an `extras` sub-package at `go/sdk/canvus/extras/`. Importable as `import "github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus/extras"`. No separate go.mod — same module as the SDK.
- **Python:** Add an `extras` sub-module at `python/sdk/src/canvus_sdk/extras/`. Importable as `from canvus_sdk.extras import geometry`. Same package, same pyproject.toml.
- **TypeScript:** Add a TS sub-path export at `typescript/sdk/src/extras/`. Importable as `import { geometry } from "@mt-canvus-tools/sdk/extras"`. Requires a new entry in the SDK's `package.json` `exports` map plus a corresponding tsup entry and tsconfig include.

Rationale: keeps a single install/dependency surface per language, lets users tree-shake (TS) or only pay for what they import (Go/Python), preserves the "core SDK = REST transport only" boundary.

---

## File structure (created by this plan)

```
MT-Canvus-Tools/
├── docs/api-reference/
│   └── parity-matrix.md                              # 4b.1 audit output
├── go/sdk/canvus/
│   ├── subscribe.go                                  # NEW — generic subscribe transport
│   ├── canvases.go                                   # MODIFY — add SubscribeCanvases, SubscribeCanvas
│   ├── folders.go                                    # MODIFY — add SubscribeFolders, SubscribeFolder
│   ├── users.go                                      # MODIFY — add SubscribeUsers, SubscribeUser
│   ├── widgets.go                                    # MODIFY — add SubscribeWidgets, SubscribeNotes, etc.
│   └── extras/                                       # NEW — ported legacy helpers
│       ├── doc.go
│       ├── geometry.go
│       ├── filters.go
│       ├── widget_operations.go
│       ├── search.go
│       └── export.go
├── python/sdk/src/canvus_sdk/
│   ├── resources/
│   │   ├── canvases.py                               # MODIFY — add subscribe()
│   │   ├── folders.py                                # MODIFY — add subscribe()
│   │   └── users.py                                  # MODIFY — add subscribe()
│   └── extras/                                       # NEW
│       ├── __init__.py
│       ├── geometry.py                               # COPY-AND-ADAPT from legacy
│       ├── filters.py                                # COPY-AND-ADAPT
│       ├── widget_operations.py                      # COPY-AND-ADAPT
│       ├── search.py                                 # COPY-AND-ADAPT
│       └── export.py                                 # COPY-AND-ADAPT
├── typescript/sdk/
│   ├── package.json                                  # MODIFY — add `./extras` export
│   ├── tsup.config.ts                                # MODIFY — add extras entry
│   └── src/extras/                                   # NEW
│       ├── index.ts
│       ├── geometry.ts
│       ├── filters.ts
│       ├── widget-operations.ts
│       ├── search.ts
│       └── export.ts
└── go/examples/core/
    ├── 05-streaming/main.go                          # REFACTOR — use SubscribeNotes
    ├── 06-llm-integration/cmd/watcher/main.go        # REFACTOR — use SubscribeNotes
    └── 07-webhooks-notifications/main.go             # REFACTOR — use SubscribeWidgets
```

---

## Task 4b.0: Pre-flight

- [ ] **Step 1: Verify HEAD is post-Phase-4a and tree is clean**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
git log -1 --oneline
git status --short
```

Expected: HEAD is at least `e51e3bb` ("docs: Phase 4a completion update + queue Phase 4b/c/d") and tree clean. If dirty, surface to Jaypaul.

- [ ] **Step 2: Verify all three SDKs and example tooling still green**

```bash
go build -C /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go ./sdk/...
go vet -C /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go ./sdk/...
go test -C /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go ./sdk/canvus/...

cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/python && uv sync

cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript && pnpm install
pnpm --filter '@mt-canvus-tools/sdk' run build
pnpm --filter '@mt-canvus-tools/sdk' exec tsc --noEmit
```

Expected: all exit 0. If anything fails, stop and surface.

- [ ] **Step 3: Verify legacy Python source is still readable**

```bash
ls /home/jaypaulb/Projects/gh/CanvusPythonAPI/canvus_api/
wc -l /home/jaypaulb/Projects/gh/CanvusPythonAPI/canvus_api/{geometry,search,filters,widget_operations,export}.py
```

Expected: five .py files listed, totalling ~2,450 lines.

---

## Task 4b.1: Dispatch parity audit (single opus agent)

- [ ] **Step 1: Single Agent tool call**

Use `subagent_type: general-purpose`, `model: opus`, `run_in_background: true`. Prompt:

```
You are auditing the three Canvus SDKs (Go / Python / TypeScript) under /home/jaypaulb/Projects/gh/MT-Canvus-Tools/ for cross-language functional parity. Jaypaul has confirmed (Tier 3 #9 in CONSOLIDATION-STATUS): full parity is required — anything one SDK can do must be doable in the other two.

# Sources
- Go SDK: /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go/sdk/canvus/ (every .go file)
- Python SDK: /home/jaypaulb/Projects/gh/MT-Canvus-Tools/python/sdk/src/canvus_sdk/ (every .py file)
- TS SDK: /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript/sdk/src/ (every .ts file)
- Legacy Python helpers to be ported: /home/jaypaulb/Projects/gh/CanvusPythonAPI/canvus_api/{geometry,search,filters,widget_operations,export}.py
- Live wire reference: /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/VERIFIED-CORRECTIONS.md

# Output
/home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/parity-matrix.md

# Structure
# SDK Parity Matrix

## Method-level parity

### Core SDK methods
For every public method exposed by ANY of the three SDKs: name, resource group, HTTP endpoint covered (or "client-side"), Go / Python / TS columns (✅ with file:line, or ❌). Focus on NON-REST helpers + convenience methods (REST coverage was verified in Phase 3.v).

### Subscribe / streaming methods
Separate sub-table — biggest known gap. Per subscribe-capable endpoint, show which SDKs have typed helpers vs raw transport.

### Constructor / Session ergonomics
from_env, WithAPIKey, withAPIKey, default timeout overrides, request-id injection, retry config — symmetric across languages?

### Error types
Each SDK's error hierarchy. Kind/subclass set the same?

## Helper-module parity (Phase 4b core deliverable)
For each of: geometry, filters, widget_operations, search, export — list functions exposed by the legacy Python module + Go/TS status (always ❌ pre-port).

## Summary
Per SDK: total public methods, missing relative to largest-surface SDK, new methods needed, complexity per gap (trivial / moderate / large).

## Per-SDK work items for Phase 4b
Ordered lists (Go, Python, TS). Each item: method name, file path, complexity, one-line spec.

## Notes
- Python legacy client.py (~112k lines) — spot-check for obvious convenience-method omissions vs migrated SDK.
- CloneWidget per-type vs single-helper interpretation: note whether per-type would be useful or single is sufficient.
- Bidirectional parity: flag any Go method that Python/TS lack.

# Critical rules
- Be EXHAUSTIVE. Cite file:line for every ✅.
- Do NOT modify any SDK source. Audit only.
- Do NOT run git commands. Do NOT commit.

# Completion report
- Total parity gaps surfaced (sum across SDKs)
- Breakdown by category (subscribe / extras / convenience / other)
- Non-trivial design questions raised
- File path of the produced matrix
```

- [ ] **Step 2: Wait for completion**

You'll be notified. Don't poll.

- [ ] **Step 3: Read the audit output**

Open `docs/api-reference/parity-matrix.md`. Skim per-SDK work items. Note total gap counts and complexity distribution.

If the audit surfaces a non-trivial design question (e.g. "should `extras.export` be in core SDK?"), surface to Jaypaul via `AskUserQuestion` before dispatching implementers. Otherwise apply senior-dev judgement and document the call in the implementer prompts.

---

## Task 4b.2: Configure the TypeScript `extras` sub-path export

Foreground scaffolding before agents touch the TS SDK.

- [ ] **Step 1: Update `typescript/sdk/package.json` exports**

Replace the `exports` field with:

```json
"exports": {
  ".": {
    "types": "./dist/index.d.ts",
    "import": "./dist/index.js",
    "require": "./dist/index.cjs"
  },
  "./extras": {
    "types": "./dist/extras/index.d.ts",
    "import": "./dist/extras/index.js",
    "require": "./dist/extras/index.cjs"
  }
}
```

- [ ] **Step 2: Update `typescript/sdk/tsup.config.ts`**

```ts
import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.ts", "src/extras/index.ts"],
  format: ["esm", "cjs"],
  dts: false,
  sourcemap: true,
  clean: true,
  target: "node20",
  splitting: false,
  treeshake: true,
});
```

- [ ] **Step 3: Create the extras placeholder**

```bash
mkdir -p /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript/sdk/src/extras
cat > /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript/sdk/src/extras/index.ts <<'EOF'
// Populated by Phase 4b implementer agent.
export {};
EOF
```

- [ ] **Step 4: Verify SDK still builds**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript
rm -rf sdk/dist
pnpm --filter '@mt-canvus-tools/sdk' run build
ls sdk/dist/ sdk/dist/extras/ 2>&1 | head
```

Expected: both `dist/index.js` AND `dist/extras/index.js` exist plus matching .d.ts files.

---

## Task 4b.3: Dispatch three parallel implementer agents

Send a single message with three Agent tool calls (`subagent_type: general-purpose`, `model: opus`, `run_in_background: true`).

Before dispatching, extract per-SDK work items from the parity matrix produced in 4b.1. Embed the relevant list in each agent's prompt (don't make them re-derive).

### Agent 4b.G — Go SDK gaps + extras port

```
You are filling Go SDK gaps identified in /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/parity-matrix.md (read the "Go work items" section). Jaypaul has confirmed full cross-SDK parity is required.

# Strict boundary
Modify ONLY paths under /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go/sdk/. Do NOT touch examples, docs (other than appending to go/sdk/MIGRATION-NOTES.md), other languages, or the repo root.

# Required deliverables

## 1. Subscribe transport + typed helpers
Create go/sdk/canvus/subscribe.go containing a generic NDJSON streaming primitive:

```go
// subscribe opens a long-lived GET request against `path` with `?subscribe=true`
// and streams decoded T values into the returned channel. Cancel ctx to stop.
// The first emitted value is typically the full snapshot; subsequent values
// are deltas. Empty keepalive lines are skipped silently.
func subscribe[T any](ctx context.Context, s *Session, path string) (<-chan T, <-chan error, error)
```

Then add typed wrappers on Session in the appropriate resource file:
- canvases.go: SubscribeCanvases(ctx) (<-chan []Canvas, <-chan error, error); SubscribeCanvas(ctx, id) (<-chan Canvas, <-chan error, error)
- folders.go: SubscribeFolders, SubscribeFolder
- users.go: SubscribeUsers, SubscribeUser
- widgets.go: SubscribeWidgets(ctx, canvasId), plus SubscribeNotes / SubscribeImages / SubscribeVideos / SubscribePdfs / SubscribeBrowsers / SubscribeAnchors / SubscribeConnectors / SubscribeTables

Mirror the API shape of Python (client.widgets.subscribe) and TS (session.widgets.subscribe) so consumers can reason cross-language. Use Go idioms (channel-of-T + error channel; first frame is the snapshot slice, subsequent frames are individual items).

## 2. Port `extras` sub-package
Create go/sdk/canvus/extras/ with:
- doc.go — package godoc explaining the extras boundary (client-side helpers, not REST transport)
- geometry.go — port Python's geometry.py: Point, Size, Rectangle types + Contains / Touches / Intersects / GetIntersection / GetUnion / WidgetBoundingBox / WidgetContains / WidgetsTouch / WidgetsIntersect / GetWidgetIntersection / GetWidgetUnion / DistanceBetweenWidgets / FindWidgetsInArea / FindWidgetsContainingPoint / GetCanvasBounds
- filters.go — port filters.py: FilterOperator (const block), Filter struct with Apply method, NewSpatialFilter, NewWidgetTypeFilter, NewTextFilter, NewWildcardFilter, CombineFilters
- widget_operations.go — port widget_operations.py: SpatialTolerance, WidgetZoneManager, BatchWidgetOperations, CreateSpatialGroup, FindWidgetClusters, CalculateWidgetDensity
- search.go — port search.py: SearchResult, CrossCanvasSearch, FindWidgetsAcrossCanvases (takes ctx + *Session), FindWidgetsByText, FindWidgetsByType, FindWidgetsInArea. These DO make REST calls.
- export.go — port export.py: ExportConfig, ImportConfig, WidgetExporter, WidgetImporter, ExportWidgetsToFolder, ImportWidgetsFromFolder. These DO touch filesystem and REST.

For every function, write a brief godoc comment that says what it does and references the original Python file as provenance.

Go-idiomatic patterns:
- Pointer receivers where mutation; values where pure
- (T, error) returns, not exception-style
- context.Context as first arg on I/O
- No global state

## 3. Implement other Go work items from the parity matrix
Whatever additional gaps the matrix surfaces. Apply senior-dev judgement on what to port or skip; document skipped items.

## 4. Document
Append to /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go/sdk/MIGRATION-NOTES.md a new section "## Phase 4b parity additions (2026-05-18)" with:
- Every new method/file with one-line spec
- Provenance for ported helpers (Python file/function → Go function map)
- Deferred items with reason

# Critical rules
- Do NOT modify go/examples/ (orchestrator refactors them in Task 4b.5).
- Do NOT run git or `go mod tidy` (orchestrator handles).
- Use `any` not `interface{}`; wrap errors with fmt.Errorf("op: %w", err).
- Aim for ~3000-4500 lines added.

# Completion report
- Files written/modified (count + paths)
- Subscribe helpers: count of methods
- Extras: count per module
- Other parity additions: count
- Deferred items with reason
- Compile risks
```

### Agent 4b.P — Python SDK gaps + extras port

```
You are filling Python SDK gaps identified in /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/parity-matrix.md (read the "Python work items" section).

# Strict boundary
Modify ONLY paths under /home/jaypaulb/Projects/gh/MT-Canvus-Tools/python/sdk/. Do NOT touch examples, other languages, docs (other than MIGRATION-NOTES.md), or the repo root.

# Required deliverables

## 1. Subscribe methods on remaining resources
The current SDK exposes subscribe() on widgets only. Add subscribe() to:
- resources/canvases.py — `async def subscribe(self) -> AsyncIterator[Canvas | list[Canvas]]`; `async def subscribe_one(self, canvas_id) -> AsyncIterator[Canvas]`
- resources/folders.py — `async def subscribe(...)`; `async def subscribe_one(self, folder_id)`
- resources/users.py — `async def subscribe(...)`; `async def subscribe_one(self, user_id)`

Pattern: same NDJSON-line generator the existing widgets.subscribe uses. Yield the parsed value per line; first line is the snapshot (list), subsequent lines are individual deltas. Reuse the transport's stream_lines helper.

## 2. Port `extras` sub-module
Create python/sdk/src/canvus_sdk/extras/ with:
- __init__.py — re-exports
- geometry.py — adapt CanvusPythonAPI/canvus_api/geometry.py. Update imports to canvus_sdk's models. Python 3.11+ syntax (`X | Y`, `list[T]`, `dict[K, V]`). No typing.List/Optional/Union.
- filters.py — adapt filters.py. Same modernisation.
- widget_operations.py — adapt widget_operations.py.
- search.py — adapt search.py. Async functions take `client: Client` (new SDK's Client, not legacy).
- export.py — adapt export.py. Takes a Client.

Preserve docstrings + add a top-line provenance note "Ported from CanvusPythonAPI/canvus_api/<name>.py — see Phase 4b plan."

## 3. Other Python work items
Address what the audit lists.

## 4. Document
Append "## Phase 4b parity additions (2026-05-18)" section to MIGRATION-NOTES.md.

# Critical rules
- Modern Python 3.11+ everywhere (no typing.List, no Optional, use X | None)
- Type-annotate everything; pass mypy --strict for SDK code
- Do NOT run git, ruff, mypy, pytest (orchestrator handles)
- Aim for ~2500-4000 lines added

# Completion report
- Files written/modified
- Subscribe methods: count
- Extras: count per module
- Other parity additions: count
- Compile/type-check risks
```

### Agent 4b.T — TypeScript SDK gaps + extras port

```
You are filling TypeScript SDK gaps identified in /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/parity-matrix.md (read the "TypeScript work items" section).

# Strict boundary
Modify ONLY paths under /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript/sdk/src/. Do NOT touch examples, docs (other than IMPLEMENTATION-NOTES.md), other languages, or the repo root.

The orchestrator already updated package.json + tsup.config.ts + created src/extras/index.ts placeholder in Task 4b.2. Build it out from there.

# Required deliverables

## 1. Port `extras` to TypeScript
Create files under typescript/sdk/src/extras/:
- index.ts — re-exports
- geometry.ts — port geometry.py: Point/Size/Rectangle types + contains, touches, intersects, getIntersection, getUnion, widgetBoundingBox, widgetContains, widgetsTouch, widgetsIntersect, getWidgetIntersection, getWidgetUnion, distanceBetweenWidgets, findWidgetsInArea, findWidgetsContainingPoint, getCanvasBounds
- filters.ts — port filters.py: FilterOperator (string-literal union), Filter class with apply, createSpatialFilter, createWidgetTypeFilter, createTextFilter, createWildcardFilter, combineFilters
- widget-operations.ts — port widget_operations.py: SpatialTolerance, WidgetZoneManager, BatchWidgetOperations, createSpatialGroup, findWidgetClusters, calculateWidgetDensity
- search.ts — port search.py: SearchResult, CrossCanvasSearch, findWidgetsAcrossCanvases (takes Session), findWidgetsByText, findWidgetsByType, findWidgetsInArea
- export.ts — port export.py: ExportConfig, ImportConfig, WidgetExporter, WidgetImporter, exportWidgetsToFolder, importWidgetsFromFolder

For each function, write a TSDoc comment + provenance.

Idiomatic strict-TS:
- Named exports only
- Kebab-case filenames (widget-operations.ts)
- Discriminated unions where Python used isinstance checks
- exactOptionalPropertyTypes-compliant (conditional spread for optional fields)
- `unknown` + narrowing, not `any`

## 2. Other TS work items
Subscribe coverage on TS is already strong (canvases, folders, users, widgets). Confirm matrix doesn't surface additional subscribe gaps. Add anything the audit lists.

## 3. Document
Append "## Phase 4b parity additions (2026-05-18)" section to IMPLEMENTATION-NOTES.md.

# Critical rules
- Strict TS — noUncheckedIndexedAccess is on. No `!`. Use `?? throw` or narrowing.
- Do NOT run git, tsc, vitest, eslint, pnpm install (orchestrator handles).
- Aim for ~2500-4000 lines added.

# Completion report
- Files written/modified
- Extras: count per module
- Other parity additions: count
- Subscribe parity confirmation
- Type-check risks
```

- [ ] **Step 2: Wait for all three agents**

- [ ] **Step 3: If any agent reports BLOCKED, NEEDS_CONTEXT, or DONE_WITH_CONCERNS**

Apply the Jaypaul failure protocol: 3 retries with tighter prompts → GH issue + investigator subagent → never stop and wait.

---

## Task 4b.4: Toolchain checks per language

After all three agents complete:

- [ ] **Step 1: Go**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go
go mod tidy -C sdk
go build ./sdk/...
go vet ./sdk/...
go test ./sdk/canvus/...
find sdk -name "*.go" | xargs gofmt -r 'interface{} -> any' -w
find sdk -name "*.go" | xargs gofmt -s -w
go build ./sdk/...
```

Expected: build + vet + test all exit 0. If failures, apply failure protocol.

- [ ] **Step 2: Python**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/python && uv sync
uvx ruff check sdk/
uvx --with mypy --with pydantic --with pydantic-settings --with structlog --with httpx --with-editable sdk mypy --strict sdk/src/canvus_sdk/
```

Expected: sync clean, ruff clean, mypy clean. Mypy surface issues — fix in place if surgical (≤10 lines), dispatch fix agent if substantial.

- [ ] **Step 3: TypeScript**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript
rm -rf sdk/dist
pnpm install
pnpm --filter '@mt-canvus-tools/sdk' run build
pnpm --filter '@mt-canvus-tools/sdk' exec tsc --noEmit
ls sdk/dist/extras/
```

Expected: clean build with both `dist/index.js` and `dist/extras/index.js` (+ `.d.ts` siblings), clean typecheck, dist/extras populated.

---

## Task 4b.5: Refactor Go examples 05 / 06 / 07 to use new Subscribe helpers

Foreground. The Go agent left examples untouched per its strict boundary.

- [ ] **Step 1: Read the relevant Go examples**

Files:
- `go/examples/core/05-streaming/main.go`
- `go/examples/core/06-llm-integration/cmd/watcher/main.go`
- `go/examples/core/07-webhooks-notifications/main.go`

Each currently builds `http.NewRequestWithContext` against `Session.HTTPClient`.

- [ ] **Step 2: Replace raw HTTP with new typed Subscribe methods**

For example 05 — typical change pattern (adapt to actual signatures from Task 4b.3 Go agent):

```go
// OLD
req, _ := http.NewRequestWithContext(ctx, "GET",
    fmt.Sprintf("%s/canvases/%s/notes?subscribe=true", baseURL, canvasID), nil)
resp, _ := s.HTTPClient().Do(req)
scanner := bufio.NewScanner(resp.Body)
for scanner.Scan() {
    line := scanner.Text()
    if line == "" { continue }
    var notes []canvus.Note
    if err := json.Unmarshal([]byte(line), &notes); err != nil { ... }
    // ... handle batch
}

// NEW
events, errs, err := s.SubscribeNotes(ctx, canvasID)
if err != nil { return fmt.Errorf("subscribe: %w", err) }
for {
    select {
    case notes, ok := <-events:
        if !ok { return nil }
        // ... handle batch
    case err := <-errs:
        return fmt.Errorf("stream: %w", err)
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

- [ ] **Step 3: Remove now-redundant http.NewRequestWithContext + bufio.Scanner code blocks**

Each example file should be 50-100 lines shorter.

- [ ] **Step 4: Re-verify Go examples still build**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go
for d in examples/core/{05-streaming,06-llm-integration,07-webhooks-notifications}/; do
  (cd "$d" && GOMAXPROCS=2 go build -p 1 ./...)
done
```

Expected: zero output (clean build).

- [ ] **Step 5: Update each example's README**

Remove the "How it works — the Go SDK doesn't yet expose a Subscribe helper, so this example uses raw HTTPClient" callout. Replace with one sentence pointing readers at `session.SubscribeNotes(...)` (or the actual method name).

---

## Task 4b.6: Commit per SDK + push

Four commits for clean history (Go SDK, Python SDK, TS SDK, Go-examples refactor).

- [ ] **Step 1: Go SDK commit**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
git add go/sdk/
git commit -m "feat(go/sdk): add Subscribe helpers + extras sub-package

Phase 4b parity sweep. Closes the largest cross-SDK gap: Go now has
typed Subscribe helpers matching the Python and TS surface
(SubscribeCanvases, SubscribeFolders, SubscribeUsers, SubscribeWidgets,
SubscribeNotes, etc.). Adds extras sub-package (geometry, filters,
widget_operations, search, export) ported from the legacy
CanvusPythonAPI helpers. See go/sdk/MIGRATION-NOTES.md for the
provenance table."
```

- [ ] **Step 2: Python SDK commit**

```bash
git add python/sdk/
git commit -m "feat(python/sdk): add subscribe parity + extras module

Phase 4b parity sweep. Adds subscribe() to canvases, folders, and
users resources (widgets already had it). Adds extras submodule
(geometry, filters, widget_operations, search, export) preserving
the legacy CanvusPythonAPI helper API. See
python/sdk/MIGRATION-NOTES.md."
```

- [ ] **Step 3: TypeScript SDK commit**

```bash
git add typescript/sdk/
git commit -m "feat(typescript/sdk): add extras sub-path export

Phase 4b parity sweep. Adds @mt-canvus-tools/sdk/extras with the
ported geometry, filters, widget-operations, search, and export
modules. Subscribe coverage was already strong on the TS side; no
additional subscribe methods needed. See
typescript/sdk/IMPLEMENTATION-NOTES.md."
```

- [ ] **Step 4: Go-examples refactor commit**

```bash
git add go/examples/core/05-streaming go/examples/core/06-llm-integration go/examples/core/07-webhooks-notifications
git commit -m "refactor(go/examples): use new Session.SubscribeNotes/SubscribeWidgets

Replaces the raw http.NewRequestWithContext + bufio.Scanner pattern
in examples 05/06/07 with the typed Subscribe helpers added in
Phase 4b. Each example is ~60-80 lines shorter. READMEs updated to
remove the 'SDK lacks Subscribe' caveat."
```

- [ ] **Step 5: Push**

```bash
git push origin main
git log --oneline | head -8
```

---

## Task 4b.7: Code review gate

- [ ] **Step 1: Dispatch reviewer**

Use `pr-review-toolkit:code-reviewer`, `run_in_background: true`. Prompt:

```
Review the Phase 4b parity sweep in commits <pre-4b>..<post-4b> at /home/jaypaulb/Projects/gh/MT-Canvus-Tools/.

Focus areas:
1. **Cross-language API parity:** Does every Go subscribe helper match its Python and TS counterparts in name, return shape, and semantics?
2. **Extras module fidelity:** Do the ported geometry/filters/widget_operations/search/export functions produce identical results across the three languages given identical inputs? Spot-check at least 2 functions per module.
3. **Idiomatic ports:** Go shouldn't read like Python, TS shouldn't read like Python — each port should use the target language's natural idioms.
4. **Wire-shape compliance:** Anywhere extras code reads widget/canvas fields, does it use the verified underscored keys (and hyphenated exceptions per VERIFIED-CORRECTIONS.md)?
5. **Example refactor:** Are the Go example 05/06/07 refactors functionally equivalent to the pre-refactor code? Same snapshot-vs-delta dedup behaviour?
6. **Silent failures / fallbacks:** Per CLAUDE.md, fail loudly. Flag any try/except swallows, ?? hides, retry loops without escalation.

Report findings categorised by severity (Critical/Important/Nice-to-have) per language, with file:line and suggested fix.
```

Capture pre-4b SHA via `git log -1 --oneline --before="<dispatch time>"` or the known commit `e51e3bb`; post-4b SHA from current HEAD.

- [ ] **Step 2: Triage findings**

- **Critical:** fix inline immediately
- **Important:** batch into follow-up commit `fix(sdk): address Phase 4b review findings`
- **Nice-to-have:** log in Phase 4c backlog

---

## Task 4b.8: Final report + CONSOLIDATION-STATUS update

- [ ] **Step 1: Update CONSOLIDATION-STATUS.md**

Add rows to "What shipped":
- `4b.audit` — parity audit + parity-matrix.md
- `4b.G/P/T` — three SDK parity-add commits
- `4b.examples` — Go examples refactor
- `4b.review` — review-driven fixes (if any)

Move the I7 (Go Subscribe) and "legacy Python helpers parity" entries out of "Deferred to Phase 4b" and into a new "Phase 4b outcome" section with what shipped.

Leave Phase 4c and Phase 4d deferral sections intact.

- [ ] **Step 2: Commit + push**

```bash
git add CONSOLIDATION-STATUS.md
git commit -m "docs: Phase 4b completion update"
git push origin main
```

- [ ] **Step 3: Brief Jaypaul (chat message)**

Summarise: total lines added per SDK, total parity gaps closed, example refactor LOC saved, any deferred items, recommended next action (Plan 4c).

---

## What's next (out of scope for this plan)

- **Plan 4c — Per-Item Refresh:** 11 existing tool/utility repos refreshed against the new SDKs (8 Go, 2 Python, 1 TS). Batched 4-6 agents per round per spec.
- **Plan 4d — Cleanup:** Go SDK dead-code removal (Chesterton-fence check + delete `setToken`, `doRequestWithHeaders`, `warnOnce`, `warnAlways`); TS README sweep for any residual wire-shape rot; back-apply VERIFIED-CORRECTIONS deltas into the source endpoint docs once `canvus-server#96` resolves.

Each will be written when its predecessor completes.
