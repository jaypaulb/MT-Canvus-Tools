# MT-Canvus-Tools Phase 3: SDK Migration Plan

**Goal:** Populate `go/sdk/`, `python/sdk/`, and `typescript/sdk/` with refreshed/migrated/greenfield SDK code grounded in the locked spec and conventions from Phase 0–2.

**Architecture:** Three independent SDK agents dispatched in parallel via the Agent tool (`subagent_type: general-purpose`, `model: opus`). Each agent owns its language directory exclusively; they do not share files. Orchestrator handles all git operations.

**Inputs (frozen by prior phases):**
- API spec: `MT-Canvus-Tools/docs/api-reference/` (20 files)
- Conventions: `MT-Canvus-Tools/docs/conventions/{go,python,typescript}.md`
- Coverage matrix: `MT-Canvus-Tools/docs/api-reference/coverage-matrix.md` (38 work items)
- Changelog: `MT-Canvus-Tools/docs/api-reference/changelog.md` (5 pending updates)

**Source repos (read-only references; do not modify):**
- Go: `/home/jaypaulb/Projects/gh/Canvus-Go-API/`
- Python: `/home/jaypaulb/Projects/gh/CanvusPythonAPI/`
- TypeScript: none (greenfield)

**Out of scope for this plan:** CLI (`go/cli/`), examples, tools, top-level READMEs, CI workflows. Those land in Phase 4+.

---

## File Structure (created by this plan)

```
MT-Canvus-Tools/
├── go/
│   ├── go.work
│   └── sdk/
│       ├── go.mod
│       ├── go.sum
│       ├── doc.go
│       ├── session.go
│       ├── client.go            # internal HTTP transport
│       ├── errors.go
│       ├── logging.go
│       ├── canvases.go
│       ├── widgets.go
│       ├── notes.go
│       ├── images.go
│       ├── videos.go
│       ├── browsers.go
│       ├── pdfs.go
│       ├── anchors.go
│       ├── connectors.go
│       ├── tables.go            # NEW (was missing)
│       ├── ipvideos.go          # NEW
│       ├── rdpconnections.go    # NEW
│       ├── videoinputs.go
│       ├── auth.go
│       ├── users.go
│       ├── groups.go
│       ├── folders.go
│       ├── assets.go
│       ├── auditlog.go
│       ├── server.go
│       ├── workspaces.go
│       ├── colorpresets.go
│       └── streaming.go
├── python/
│   ├── pyproject.toml           # uv workspace root
│   └── sdk/
│       ├── pyproject.toml
│       ├── README.md
│       └── src/canvus_sdk/
│           ├── __init__.py
│           ├── client.py        # async Client + sync wrapper namespace
│           ├── transport.py     # httpx wrappers
│           ├── errors.py        # CanvusError hierarchy
│           ├── logging.py       # structlog config
│           ├── config.py        # pydantic-settings
│           ├── models/          # pydantic data models
│           │   ├── __init__.py
│           │   ├── canvas.py
│           │   ├── widget.py    # union types per widget kind
│           │   ├── user.py
│           │   ├── folder.py
│           │   └── asset.py
│           └── resources/       # endpoint groupings
│               ├── __init__.py
│               ├── canvases.py
│               ├── widgets.py
│               ├── auth.py
│               ├── users.py
│               ├── folders.py
│               ├── assets.py
│               ├── server.py
│               └── streaming.py
└── typescript/
    ├── package.json             # pnpm workspace root
    ├── pnpm-workspace.yaml
    ├── tsconfig.base.json
    └── sdk/
        ├── package.json
        ├── tsconfig.json
        ├── tsup.config.ts
        ├── README.md
        └── src/
            ├── index.ts
            ├── session.ts
            ├── transport.ts
            ├── errors.ts
            ├── logging.ts
            ├── config.ts
            ├── streaming.ts
            ├── types/
            │   ├── canvas.ts
            │   ├── widget.ts
            │   ├── user.ts
            │   ├── folder.ts
            │   └── asset.ts
            └── resources/
                ├── canvases.ts
                ├── widgets.ts
                ├── auth.ts
                ├── users.ts
                ├── folders.ts
                ├── assets.ts
                ├── server.ts
                └── streaming.ts
```

---

## Concurrency model

All 3 SDK agents own non-overlapping subtrees (`go/`, `python/`, `typescript/`). They can run fully in parallel without worktree isolation, but each agent is explicitly instructed:

> You MUST NOT touch any file outside your assigned language directory. The repo root, `.github/`, `docs/`, and other language subtrees are owned by other agents and must not be modified.

Orchestrator handles all `git add`, `git commit`, `git push`. Agents write files only.

---

## Task 3.0: Pre-flight checks

- [ ] **Verify source repos are clean and on default branch**

```bash
for repo in Canvus-Go-API CanvusPythonAPI; do
  echo "=== $repo ==="
  git -C /home/jaypaulb/Projects/gh/$repo status --short
  git -C /home/jaypaulb/Projects/gh/$repo log -1 --oneline
done
```

If either repo has uncommitted changes, surface to Jaypaul before dispatch.

- [ ] **Verify Phase 2 commit is pushed**

```bash
git -C /home/jaypaulb/Projects/gh/MT-Canvus-Tools log --oneline -1
git -C /home/jaypaulb/Projects/gh/MT-Canvus-Tools status --short
```

---

## Task 3.1: Dispatch 3 SDK agents in parallel

Send a single message with three Agent tool calls. Each agent receives the full prompt below (one prompt block per agent).

### Agent 3.1: Go SDK migration

```
You are migrating the existing Canvus Go SDK into the MT-Canvus-Tools monorepo, applying drift remediations, and adding missing endpoints. Jaypaul is asleep and CANNOT answer questions — make best-judgement calls and document them in commit-ready notes.

# Strict boundary
You MUST NOT modify any file outside /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go/ . The repo root, docs/, .github/, python/, typescript/ are off-limits.

# Inputs to read in full
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/conventions/go.md — locked conventions
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/coverage-matrix.md — 12 Phase 3 Go work items
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/changelog.md — 5 pending API updates
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/endpoints/ — endpoint specs (read what's relevant per file)
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/authentication.md, widget-types.md, streaming.md
- /home/jaypaulb/Projects/gh/Canvus-Go-API/ — existing SDK to migrate (READ-ONLY)

# Output location
/home/jaypaulb/Projects/gh/MT-Canvus-Tools/go/

# Required deliverables

## 1. Workspace + module scaffold
- go/go.work (lists ./sdk)
- go/sdk/go.mod with module path github.com/jaypaulb/MT-Canvus-Tools/go/sdk, go 1.22, toolchain go1.23.0
- go/sdk/doc.go with package-level godoc
- go/sdk/README.md (concise: install, quick example, link to docs)

## 2. Migrated SDK files
For every Go file in Canvus-Go-API/canvus/:
  a. Copy logically (don't preserve filename if convention dictates otherwise)
  b. Update module path imports
  c. Apply drift remediations from go.md (drift items 1, 4, 5 are mandatory; 2, 3 N/A for SDK; 6 best-effort)
  d. Replace ErrorCode string consts with errors.New sentinels alongside existing patterns
  e. Add slog logging at appropriate I/O boundaries (NewSession, request errors)

## 3. Missing endpoints (per coverage-matrix.md Go work items)
- tables.go — full CRUD + grid_size silent-ignore handling per changelog §5
- ipvideos.go — GET/PATCH/DELETE only; reject POST per changelog §2
- rdpconnections.go — same as ipvideos
- CloneWidget helper (across-canvas) using standard create endpoints with source_canvas_id/source_widget_id per changelog §1
- Audit log filters + envelope decoding (events/total-count/page/per-page)
- Color presets path correction (color-presets not colorpresets)
- Remove non-spec methods: CreateClient/UpdateClient/DeleteClient (after a grep confirms zero internal callers)

## 4. Outdated endpoint fixes (8 endpoints from work items)
Apply field-name corrections; for ambiguous ones (login `username` vs spec's `email`; license `key` vs spec's `license-data`; send-test-email body), prefer the SOURCE-OF-TRUTH side (C++ canonical impl wins over public docs). Document the call in a comment.

## 5. Testing scaffold
- go/sdk/session_test.go with table-driven unit tests for NewSession + a smoke test
- go/sdk/integration/ folder with one //go:build integration tagged test calling the live API (use env vars CANVUS_API_KEY, CANVUS_BASE_URL — do NOT hardcode the dev creds)
- DO NOT run go test (no live server available in your sandbox)

## 6. Decisions log
Write go/sdk/MIGRATION-NOTES.md listing every non-obvious decision you made:
- Removed methods (with reason)
- Field-name conflicts resolved (which source you trusted, why)
- Drift items addressed vs deferred
- Any "Source uncertainty" left unresolved (Phase 4 will tackle)

# Critical rules
- Do NOT run git commands. Do NOT commit.
- Do NOT run go mod tidy yet (orchestrator does that post-dispatch).
- Code must compile in principle — use correct types, exported names per Go convention. If you cannot verify compilation, add a "needs go mod tidy + manual review" callout in MIGRATION-NOTES.md.
- Aim for ~3000-4500 lines of Go across all files. Thinner = incomplete coverage.

# Completion report
- Files written (count, total lines)
- Endpoints covered (count vs spec's 150)
- Work items resolved vs deferred (cite item numbers)
- Top 3 risks / unresolved decisions for Jaypaul's morning review
```

### Agent 3.2: Python SDK migration

```
You are migrating the existing Canvus Python SDK into the MT-Canvus-Tools monorepo, applying drift remediations, and adding missing endpoints. Jaypaul is asleep and CANNOT answer questions — make best-judgement calls and document them.

# Strict boundary
You MUST NOT modify any file outside /home/jaypaulb/Projects/gh/MT-Canvus-Tools/python/ . The repo root, docs/, .github/, go/, typescript/ are off-limits.

# Inputs to read in full
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/conventions/python.md — locked conventions
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/coverage-matrix.md — 23 Phase 3 Python work items
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/changelog.md — 5 pending API updates
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/endpoints/ — endpoint specs
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/authentication.md, widget-types.md, streaming.md
- /home/jaypaulb/Projects/gh/CanvusPythonAPI/ — existing SDK to migrate (READ-ONLY)

# Output location
/home/jaypaulb/Projects/gh/MT-Canvus-Tools/python/

# Required deliverables

## 1. Workspace + package scaffold
- python/pyproject.toml — uv workspace root listing ["sdk"]
- python/sdk/pyproject.toml — PEP 621, requires-python >=3.11, hatchling backend, ruff+mypy+pytest configs per conventions
- python/sdk/README.md
- python/sdk/src/canvus_sdk/__init__.py exposing Client, CanvusError, key models

## 2. Migrated + restructured code
- Move CanvusPythonAPI/canvus_api/ → python/sdk/src/canvus_sdk/ (src layout)
- Rename exception base CanvusAPIError → CanvusError; add APIError subclass for HTTP-level
- Replace stdlib logging with structlog (apply python.md §Logging pattern)
- Drop aiohttp; httpx-only (sync + async)
- Restructure into models/ (pydantic v2) and resources/ (endpoint groupings)
- Single CANVUS_ env prefix (drop the project-specific suffixes)
- Type-annotate everything to pass mypy --strict (user_id: str, not int — see work item 21)

## 3. Missing endpoints (per coverage-matrix.md Python work items)
- Tables widget family (full CRUD + grid_size silent-ignore per changelog §5)
- IP Video + RDP families (GET/PATCH/DELETE; reject POST per changelog §2)
- clone_widget() helper using standard create endpoints with source_canvas_id/source_widget_id
- Audit log filters + envelope decoding
- change_email, force_reset_password, update_group, reload_certs, open_canvas_in_workspace
- Other items per work-items list

## 4. Outdated endpoint fixes
- set_canvas_permissions body shape
- send_test_email recipient param
- install_offline_license field name
- update_video_output path correction

## 5. Testing scaffold
- python/sdk/tests/test_client.py with pytest unit tests for Client init + smoke
- python/sdk/tests/integration/ with one @pytest.mark.integration test reading CANVUS_API_KEY/CANVUS_BASE_URL from env (do NOT hardcode dev creds)
- DO NOT run pytest

## 6. Decisions log
Write python/sdk/MIGRATION-NOTES.md listing every non-obvious decision:
- Sync vs async API surface choice
- Models: full pydantic v2 vs lightweight TypedDict
- Drift items resolved vs deferred
- Unresolved "Source uncertainty" items

# Critical rules
- Do NOT run git commands. Do NOT commit.
- Do NOT run uv sync, pip install, ruff, mypy, pytest (orchestrator handles).
- Use modern Python 3.11+ syntax: `X | Y`, `list[T]`, `dict[K, V]`, `Self`, no `typing.List`.
- Aim for ~3000-4500 lines of Python across all files.

# Completion report
- Files written (count, total lines)
- Endpoints covered (count vs spec's 150)
- Work items resolved vs deferred (cite item numbers)
- Top 3 risks / unresolved decisions for Jaypaul's morning review
```

### Agent 3.3: TypeScript SDK greenfield

```
You are writing a greenfield TypeScript SDK for the Canvus REST API in the MT-Canvus-Tools monorepo. There is no existing TS SDK to migrate. Jaypaul is asleep and CANNOT answer questions — make best-judgement calls and document them.

# Strict boundary
You MUST NOT modify any file outside /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript/ . The repo root, docs/, .github/, go/, python/ are off-limits.

# Inputs to read in full
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/conventions/typescript.md — locked conventions
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/coverage-matrix.md (TS section)
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/changelog.md — 5 pending API updates
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/endpoints/ — full endpoint specs (read all 7)
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/authentication.md, widget-types.md, streaming.md

# Output location
/home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript/

# Required deliverables

## 1. Workspace + package scaffold
- typescript/package.json (private, workspace root)
- typescript/pnpm-workspace.yaml
- typescript/tsconfig.base.json with strict + noUncheckedIndexedAccess + ES2022 + Bundler
- typescript/sdk/package.json: name @mt-canvus-tools/sdk, dual ESM+CJS via tsup, version 0.1.0
- typescript/sdk/tsconfig.json
- typescript/sdk/tsup.config.ts
- typescript/sdk/README.md
- typescript/sdk/.eslintrc / eslint.config.js (flat)
- typescript/.prettierrc

## 2. Core SDK source (typescript/sdk/src/)

### Transport + session
- src/transport.ts — fetch wrapper with auth header injection, error mapping to CanvusError hierarchy
- src/session.ts — Session class (created via createSession({baseUrl, apiKey})) exposing resource namespaces

### Types (src/types/)
Full data models for: Canvas, Widget union (note/image/video/browser/pdf/anchor/connector/table/ipvideo/rdpconnection/videoinput), User, Group, Folder, Asset, ServerInfo, AuditEntry. Use TypeScript discriminated unions for the Widget kind.

### Errors (src/errors.ts)
- CanvusError extends Error (base)
- APIError extends CanvusError (HTTP-level, with status + payload)
- ValidationError, AuthError, NetworkError, NotFoundError

### Logging (src/logging.ts)
- pino logger with env-driven transport (json vs pino-pretty)
- Single configured instance per package

### Config (src/config.ts)
- zod schema for CANVUS_* env vars
- loadConfig() returning typed Config

### Streaming (src/streaming.ts)
- Async iterator using undici for true streaming/backpressure
- subscribe(): AsyncIterable<Event> consumed as `for await (const event of session.subscribeCanvas(id))`
- NDJSON parsing + keep-alive ping handling

### Resources (src/resources/)
Implement EVERY endpoint from the spec. One file per resource group:
- canvases.ts (17 endpoints)
- widgets.ts (64 endpoints across all widget types)
- auth.ts (13)
- users.ts (19)
- folders.ts (12)
- assets.ts (3)
- server.ts (22)
- streaming.ts (in addition to streaming.ts at root, this exposes subscribe helpers per resource)

Each resource exposes typed methods on the Session class via `session.canvases.list()`, `session.canvases.get(id)`, etc.

Implement the changelog updates correctly:
- §1: cloneWidget() helper using standard create endpoints with sourceCanvasId/sourceWidgetId
- §2: omit POST methods for IP Video + RDP types
- §3: omit columnWidths/rowHeights from table responses
- §4: silent-ignore behaviour on table grid_size PATCH
- §5: flag the RDP field-naming uncertainty in a comment

### Index (src/index.ts)
- Re-export createSession, Session, all types, all errors

## 3. Testing scaffold
- typescript/sdk/tests/session.test.ts — vitest unit tests with mocked fetch
- typescript/sdk/tests/integration/canvas.test.ts — integration test reading CANVUS_API_KEY/CANVUS_BASE_URL from env (do NOT hardcode dev creds)
- typescript/sdk/vitest.config.ts

## 4. Decisions log
Write typescript/sdk/IMPLEMENTATION-NOTES.md listing:
- Validation strategy (zod runtime validation vs trust-the-server)
- Browser/Deno support claims and what's tested
- Streaming implementation choices
- Any "Source uncertainty" deferred for Phase 4

# Critical rules
- Do NOT run git commands. Do NOT commit.
- Do NOT run pnpm install, tsc, vitest, eslint (orchestrator handles).
- Use ESM imports, named exports only, kebab-case filenames.
- Strict TS: no `any`, prefer `unknown` + narrowing.
- Aim for ~4500-6000 lines of TS across all files (greenfield, 150 endpoints).

# Completion report
- Files written (count, total lines)
- Endpoints implemented (count vs spec's 150)
- Architectural decisions made autonomously (zod yes/no, validation depth, etc.)
- Top 3 risks / unresolved decisions for Jaypaul's morning review
```

---

## Task 3.2: Verify agent outputs

After all 3 agents complete:

- [ ] **Inventory each language directory**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
for lang in go python typescript; do
  echo "=== $lang ==="
  find $lang -type f \( -name "*.go" -o -name "*.py" -o -name "*.ts" -o -name "*.md" -o -name "*.json" -o -name "*.toml" -o -name "*.yaml" -o -name "*.yml" \) | wc -l
  find $lang -type f \( -name "*.go" -o -name "*.py" -o -name "*.ts" \) -exec wc -l {} + | tail -1
done
```

Expected: Go ~30 files / ~3500 lines; Python ~20 files / ~3500 lines; TypeScript ~25 files / ~5000 lines.

- [ ] **Read each MIGRATION-NOTES.md / IMPLEMENTATION-NOTES.md** to capture decision rationale for the completion report.

If any agent produced thin output (<50% of expected lines, or missing entire deliverables sections), apply the failure protocol: retry up to 3 times with tighter prompts; if still failing, create a GH issue and continue.

---

## Task 3.3: Commit + push per-language

Three commits, one per language, for clean history.

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools

# Go
git add go/
git commit -m "feat(go/sdk): migrate Canvus-Go-API into monorepo with refresh

Phase 3 of consolidation. Migrated existing SDK code, applied drift
remediations per docs/conventions/go.md, added missing endpoints per
coverage-matrix.md (tables, ipvideos, rdpconnections, clone helper,
audit log envelope). See go/sdk/MIGRATION-NOTES.md for decision log."

# Python
git add python/
git commit -m "feat(python/sdk): migrate CanvusPythonAPI into monorepo with refresh

Phase 3 of consolidation. Migrated to src layout, renamed exception
base to CanvusError, switched to httpx-only + structlog + ruff,
added missing endpoints per coverage-matrix.md. See
python/sdk/MIGRATION-NOTES.md for decision log."

# TypeScript
git add typescript/
git commit -m "feat(typescript/sdk): greenfield TypeScript SDK for Canvus API

Phase 3 of consolidation. Greenfield implementation of 150 endpoints
following docs/conventions/typescript.md (pnpm + tsup + pino + undici
+ vitest). See typescript/sdk/IMPLEMENTATION-NOTES.md for decisions."

git push origin main
```

---

## Task 3.4: Update HANDOFF + final report

- [ ] Write `CONSOLIDATION-STATUS.md` at MT-Canvus-Tools root summarising:
  - Phase 0–3 completion state
  - Outstanding decisions surfaced by agents (collect from all NOTES.md files)
  - Phase 4 readiness checklist
  - Outstanding GH issues opened during failure recovery

- [ ] Update `/home/jaypaulb/Projects/gh/docs/superpowers/HANDOFF.md` with new state.

- [ ] Final report to Jaypaul (in chat, last message before stop):
  - 3-line summary of what shipped
  - Outstanding decisions requiring his review (max 10 bullets)
  - GH issues opened (if any)
  - Recommended next action

---

## What's next (out of scope)

Phase 4a (Core Examples): 8 canonical examples × 3 languages = 24 small projects. Written using the SDKs from Phase 3. Separate plan.

Phase 4b (Per-Item Refresh): 11 existing utility/tool repos refreshed against the new SDKs. Separate plan.
