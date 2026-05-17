# MT-Canvus-Tools Consolidation — Design Spec

**Date:** 2026-05-17
**Status:** Approved for implementation planning
**Owner:** Jaypaul Barrow

---

## Context

The Canvus ecosystem currently spans 13+ separate GitHub repositories — two SDKs (Go, Python), a CLI, two MCP servers, multiple tools, multiple example integrations, web demos, and configuration assets. Each repo evolved independently against the Canvus REST API; the SDKs have drifted from the current authoritative implementation in `gl/conan/canvus/mt-restapi-client/`, and there is no unified entry point for developers wanting to build on Canvus.

This spec consolidates the Canvus API-consumer projects into a single monorepo, `MT-Canvus-Tools`, organised by language. Every artifact (SDK, example, tool) is re-aligned to the current API spec and modernised against current best practices during the migration. Flutter applications (CanvusLite, CanvusConsole) remain as standalone product repos and are out of scope. Server-side installers (CanvusServerInstaller) are dropped — containerised deployment has superseded them.

The intended outcome is a single discoverable home for Canvus integration work, with parallel canonical examples in Go / Python / TypeScript, language-isolated workspaces, a maintained API reference derived from the canonical C++ implementation, and a clear convention set so future contributors (human or agent) can extend the repo without one-off decisions.

---

## Scope

### In scope
- Two existing SDKs migrated and updated: **Canvus-Go-API**, **CanvusPythonAPI**
- One new SDK authored from spec: **TypeScript SDK**
- Existing Go tools, CLI, MCP server, and project-tier examples
- Existing Python MCP server and tools
- Existing Node example (CanvusWebUI) ported to typescript/examples
- Custom-menu configuration assets consolidated into a single folder
- A new canonical example set mirrored across all three languages
- Authoritative API reference derived from `gl/conan/canvus/mt-restapi-client/`

### Out of scope
- **CanvusLite** (Flutter product app) — stays standalone
- **CanvusConsole** (Flutter product app) — stays standalone
- **CanvusServerInstaller** (shell + PowerShell installers) — dropped, containerised deployment replaces it
- The C++ `mt-restapi-client` itself (it is the *source* of the API spec, not a consumer SDK)
- C#, Java/Kotlin, Rust, Swift, Ruby SDKs — deferred to a future version when demand is established

### Languages (v1)
1. **Go** — port-and-update existing SDK + 8 dependent projects
2. **Python** — port-and-update existing SDK + 2 dependent tools
3. **TypeScript** — greenfield SDK, hand-written from API spec; covers Node, browser, and Deno consumers

---

## Repo Structure

```
MT-Canvus-Tools/
├── README.md
├── LICENSE
├── CONTRIBUTING.md
├── docs/
│   ├── api-reference/
│   │   ├── README.md
│   │   ├── endpoints/                # One file per resource group
│   │   ├── authentication.md
│   │   ├── widget-types.md
│   │   ├── streaming.md
│   │   ├── changelog.md              # Captures doc-updates-for-developer-site changes
│   │   └── coverage-matrix.md        # Endpoint × SDK implementation status
│   ├── conventions/
│   │   ├── go.md                     # Build, lint, log, error, config, test, CI defaults
│   │   ├── python.md
│   │   └── typescript.md
│   ├── getting-started/
│   │   ├── go.md
│   │   ├── python.md
│   │   └── typescript.md
│   └── contributing/
├── go/
│   ├── README.md
│   ├── go.work                       # Go workspace tying modules together
│   ├── sdk/                          # ← Canvus-Go-API (updated to spec)
│   ├── cli/                          # ← canvus-cli (refreshed)
│   ├── examples/
│   │   ├── core/                     # 8 canonical examples (mirrored across languages)
│   │   │   ├── 01-auth-and-list/
│   │   │   ├── 02-auth-flows/
│   │   │   ├── 03-widget-crud/
│   │   │   ├── 04-file-upload/
│   │   │   ├── 05-streaming/
│   │   │   ├── 06-llm-integration/
│   │   │   ├── 07-webhooks-notifications/
│   │   │   └── 08-cross-canvas-clone/
│   │   └── projects/                 # Tutorial-grade applications (Go-specific)
│   │       ├── ai-personas/          # ← AI-personas (refreshed)
│   │       ├── llm-canvas-companion/ # ← CanvusAPI-LLMDemo (renamed, refreshed)
│   │       └── note-mapper/          # ← CanvusNoteMapper (refreshed)
│   └── tools/
│       ├── mcp-server/               # ← CanvusMCP (refreshed)
│       ├── powertoys/                # ← CanvusPowerToys (refreshed)
│       ├── translator/               # ← CanvusTranslator (refreshed)
│       └── db-solver/                # ← Canvus-Server-db-solver (refreshed)
├── python/
│   ├── README.md
│   ├── pyproject.toml                # uv workspace
│   ├── sdk/                          # ← CanvusPythonAPI (updated to spec)
│   ├── examples/
│   │   └── core/                     # Same 8 canonical examples (Python implementations)
│   └── tools/
│       ├── mcp-server/               # ← canvus-mcp-server (refreshed)
│       └── local-llm/                # ← Canvus-Local-LLM (refreshed)
├── typescript/
│   ├── README.md
│   ├── package.json                  # pnpm workspaces
│   ├── tsconfig.base.json
│   ├── sdk/                          # NEW — hand-written from mt-restapi-client spec
│   ├── examples/
│   │   ├── core/                     # Same 8 canonical examples (TS implementations)
│   │   └── webui/                    # ← CanvusWebUI (rewritten on new TS SDK)
└── custom-menu/                      # ← CanvusCustomMenuExample + canvus-menu merged
    ├── README.md
    ├── examples/
    └── docs/
```

---

## Migration Map

| Source repo | Destination | Action |
|---|---|---|
| Canvus-Go-API | `go/sdk/` | Copy → align to current API spec → modernise |
| canvus-cli | `go/cli/` | Copy → refresh SDK calls + best practices |
| CanvusMCP | `go/tools/mcp-server/` | Copy → refresh |
| CanvusPowerToys | `go/tools/powertoys/` | Copy → refresh |
| CanvusTranslator | `go/tools/translator/` | Copy → refresh |
| Canvus-Server-db-solver | `go/tools/db-solver/` | Copy → refresh |
| AI-personas | `go/examples/projects/ai-personas/` | Copy → refresh |
| CanvusAPI-LLMDemo | `go/examples/projects/llm-canvas-companion/` | Copy → rename → refresh |
| CanvusNoteMapper | `go/examples/projects/note-mapper/` | Copy → refresh |
| CanvusPythonAPI | `python/sdk/` | Copy → align to current API spec → modernise |
| canvus-mcp-server | `python/tools/mcp-server/` | Copy → refresh |
| Canvus-Local-LLM | `python/tools/local-llm/` | Copy → refresh |
| CanvusWebUI | `typescript/examples/webui/` | Rewrite using new TS SDK |
| CanvusCustomMenuExample + canvus-menu | `custom-menu/` | Merge → dedupe → unified README |
| CanvusServerInstaller | — | **Dropped** (containerised deployment replaces) |
| CanvusLite | — | **Out of scope** (standalone product) |
| CanvusConsole | — | **Out of scope** (standalone product) |

**Migration method:** fresh copy, no git history preservation. Each subfolder receives a single "Imported from `<repo>@<sha>` — see migration-source.md for upstream history" commit followed by refresh commits.

---

## Core Examples (mirrored across all 3 languages)

Each language implements the same 8 examples with idiomatic conventions per the conventions doc:

| # | Name | Purpose |
|---|---|---|
| 01 | auth-and-list | Minimal viable: authenticate, list canvases |
| 02 | auth-flows | Login + API key + token-refresh patterns |
| 03 | widget-crud | Create, update, delete a sticky note |
| 04 | file-upload | Upload an image and place as a widget |
| 05 | streaming | Subscribe to a canvas, react to widget changes |
| 06 | llm-integration | Paired apps: watcher + responder via external LLM |
| 07 | webhooks-notifications | Outbound notification subscriptions |
| 08 | cross-canvas-clone | New endpoint per `doc-updates-for-developer-site.md` |

Total: **24 small projects** (8 × 3 languages).

---

## Phase Breakdown

```
Phase 0  Bootstrap                        foreground          ~1hr
Phase 1  API Spec → docs/api-reference/   3× Explore (Haiku)  ~3hr
Phase 2  SDK Audit + Conventions Doc      5× Plan (Opus)      ~3hr
Phase 3  SDK Migration & Update           3× code (Sonnet)    ~6hr
Phase 4a Core Examples Authoring          3× code (Sonnet)    ~4hr
Phase 4b Per-Item Autonomous Refresh      11× code, batched   ~10hr
Phase 5  Custom Menu Consolidation        1× code             ~1hr
Phase 6  Top-level Docs                   1× code             ~2hr
Phase 7  CI/CD                            1× code             ~2hr
Phase 8  Public Release                   foreground          ~30min
```

### Phase 0 — Bootstrap
- Create `MT-Canvus-Tools/` directory locally under `/home/jaypaulb/Projects/gh/`
- Initialise git, scaffold top-level dirs from the structure above
- **Decision required from Jaypaul at bootstrap:** licence choice (MIT vs proprietary vs Apache-2.0) — affects `LICENSE` file and per-source-file headers
- Create initial `README.md`, `LICENSE`, `CONTRIBUTING.md` stubs
- Push to `github.com/jaypaulb/MT-Canvus-Tools` (private at start; flip to public at Phase 8)
- Copy this design spec into `docs/superpowers/specs/`
- Commit & push

### Phase 1 — API Spec as Source of Truth
Three Explore agents working in parallel against the canonical sources:

- **Agent 1.1 — Endpoint catalog:** Read `gl/conan/canvus/mt-restapi-client/RestApiClient/` source. Produce `docs/api-reference/endpoints/*.md`, one file per resource group (canvases, widgets, auth, users, folders, assets, server). Each endpoint: HTTP verb, path, params, request schema, response schema, auth requirements, streaming support.
- **Agent 1.2 — Behavioural examples:** Read `gl/conan/canvus/mt-restapi-tests/tests/`. Produce `docs/api-reference/examples/` — known-good request/response pairs derived from integration tests.
- **Agent 1.3 — Pending changes:** Read `gl/conan/canvus/mt-restapi-client/doc-updates-for-developer-site.md`. Produce `docs/api-reference/changelog.md` capturing what has changed since the public developer docs were last updated.

Output: `docs/api-reference/` is the canonical, language-agnostic spec for all downstream phases.

**Spec freeze:** At the end of Phase 1, capture the `mt-restapi-client` and `mt-restapi-tests` commit SHAs into `docs/api-reference/SOURCE.md`. Subsequent phases reference *this captured spec*, not a moving upstream target. A later phase (post-v1) can re-extract and diff to catch upstream drift.

### Phase 2 — SDK Audit + Conventions
Five Plan agents in parallel:

- **2.1 Go SDK audit** — for each endpoint in `docs/api-reference/`, mark Go SDK status: implemented / missing / outdated / deprecated. Output: Go column of `coverage-matrix.md`.
- **2.2 Python SDK audit** — same for Python. Output: Python column.
- **2.3 Conventions/Go** — author `docs/conventions/go.md`: module structure, build (go workspaces), lint (golangci-lint config), log (slog), error (wrapping + sentinel errors), config (env + file), test (stdlib + testify), CI (GH Actions matrix on supported Go versions).
- **2.4 Conventions/Python** — author `docs/conventions/python.md`: pyproject.toml + uv workspaces, ruff lint, structlog, pytest, typing strictness, CI matrix.
- **2.5 Conventions/TypeScript** — author `docs/conventions/typescript.md`: pnpm workspaces, tsconfig strict, eslint+prettier, vitest, pino logging, CI matrix on Node versions.

The conventions doc is **the autonomy contract** for Phase 4b. Locks in defaults so per-item refresh agents need no prompts.

### Phase 3 — SDK Migration & Update
Three code-writing agents in parallel, each in its own worktree:

- **3.1 Go SDK track** (`feature/go-sdk` branch): copy `Canvus-Go-API` → `go/sdk/`. Update Go module path. Address every Phase-2.1 finding (add missing endpoints, fix outdated, remove deprecated). Update tests against current API behaviour. Apply Phase-2.3 conventions. Commit incrementally.
- **3.2 Python SDK track** (`feature/python-sdk` branch): copy `CanvusPythonAPI` → `python/sdk/`. Address Phase-2.2 findings. Apply Phase-2.4 conventions.
- **3.3 TypeScript SDK track** (`feature/typescript-sdk` branch): hand-write from `docs/api-reference/`. Use Python SDK as design reference, **not source**. Cover all endpoints. Apply Phase-2.5 conventions. Author baseline tests.

After each track completes: `pr-review-toolkit:code-reviewer` (Opus) reviews; reviewer findings either merged-in or filed as follow-ups. Then feature branch fast-forward-merged to main.

### Phase 4a — Core Examples Authoring
Three code-writing agents in parallel, one per language. Each authors **all 8 canonical examples** in its language, in dependency order (01 first, 08 last). Each example:
- Self-contained module/project
- README explaining purpose + walkthrough
- Minimal config (env vars only)
- Idiomatic code per language conventions
- Runs end-to-end against a configured Canvus server

### Phase 4b — Per-Item Autonomous Refresh
Eleven items, batched 4–6 at a time to respect runaway-guard. Each item is a single agent task with the contract:

> *"Refresh `<source>` into `<target>`. Apply current API spec from `docs/api-reference/`. Apply conventions from `docs/conventions/<lang>.md`. Modernise dependencies, error handling, logging, config. Update the item's README. Run build + tests. Commit on `feature/refresh-<item>` branch. Do NOT prompt; pick best-practice defaults and document choices in the item's README."*

**Items:**

Go (8): cli, mcp-server, powertoys, translator, db-solver, ai-personas, llm-canvas-companion, note-mapper.
Python (2): mcp-server, local-llm.
TypeScript (1): webui (this is a rewrite, not a refresh — the original is JavaScript/Node).

After each batch: reviewer agent → merge approved branches.

### Phase 5 — Custom Menu Consolidation
Single agent: merge `CanvusCustomMenuExample/` + `canvus-menu/canvus-custom-menu/` into `custom-menu/`. Dedupe overlap. Author unified README explaining what custom menus are and how to configure them.

### Phase 6 — Top-level Documentation
Single agent: author top-level `README.md` (language matrix, quick links), per-language landing READMEs (consolidating any Phase-3 partials), `docs/getting-started/{go,python,typescript}.md`, and `docs/contributing/`.

### Phase 7 — CI/CD
Single agent: author GitHub Actions workflows. Per-language matrix jobs (lint, test, build). Optional spec-drift detector that compares `docs/api-reference/` to `mt-restapi-client` on a schedule.

### Phase 8 — Public Release
Foreground (Jaypaul): final review. Flip repo visibility to public. Tag `v1.0.0`. Archive source repos with a `MOVED-TO-MT-CANVUS-TOOLS.md` notice.

---

## Multi-Agent Dispatch Strategy

| Phase | Agent Type | Model | Count | Parallelism |
|---|---|---|---|---|
| 1 | Explore | Haiku | 3 | All in one message |
| 2 | Plan | Opus | 5 | All in one message |
| 3 | general-purpose, `isolation:worktree` | Sonnet 4.6 | 3 | All in one message |
| 4a | general-purpose | Sonnet 4.6 | 3 | All in one message |
| 4b | general-purpose, `isolation:worktree` | Sonnet 4.6 | 11 | **Batched 4–6 per round** |
| 5–7 | general-purpose | Sonnet 4.6 | 1 each | Sequential |
| QA gates | `pr-review-toolkit:code-reviewer` | Opus | per phase | After 3, 4a, 4b |

**Worktree isolation** for Phase 3 and 4b: each agent gets its own branch + checkout. Branches merged to main only after reviewer approval.

**Runaway-guard safety:** CLAUDE.md warns at 80 concurrent bash processes, blocks at 200. Phase 4b's 11 items batched at 4–6 keeps total well under threshold.

**Autonomy contract for Phase 4b agents:**
- MUST read `docs/api-reference/` and `docs/conventions/<lang>.md` before coding
- MUST commit and push their work
- MUST NOT spawn `AskUserQuestion`
- IF a new convention decision is required: pick LLM-best-practice option AND append rationale to `docs/conventions/<lang>.md`
- IF a build/test failure persists after one fix attempt: surface via task status, do not silently retry

**Failure handling:** Failed agents leave their task `in_progress` and branch unmerged. Orchestrator (Claude main session) inspects, decides retry/replan/skip per CLAUDE.md "On Failure" protocol, surfaces to Jaypaul before destructive action.

---

## Conventions Doc (autonomy contract)

`docs/conventions/{go,python,typescript}.md` is authored in Phase 2 and locks the defaults Phase 4b agents follow. Each file covers:

- **Module/package layout** (workspaces, internal vs public APIs)
- **Build tooling** (go workspaces, uv, pnpm)
- **Linting** (golangci-lint, ruff, eslint+prettier)
- **Logging** (slog, structlog, pino)
- **Error handling** (wrapping patterns, sentinel errors, typed errors)
- **Configuration** (env + file precedence)
- **Testing** (framework, coverage threshold)
- **CI matrix** (supported language versions)

Phase 2 produces these autonomously, no checkpoint. Phase 3 SDKs are the first consumers — they validate the conventions are workable before Phase 4b scales them out.

---

## Risk & Rollback

| Risk | Mitigation |
|---|---|
| API spec extraction misreads C++ code | Phase 1 cross-references with `mt-restapi-tests` behaviour; Phase 2 audit catches gaps |
| Conventions doc encodes a bad default | Phase 3 SDK work surfaces it; conventions doc is amendable; cost is rework, not corruption |
| Phase 4b agent makes wrong refresh call | Worktree isolation means damage stays on feature branch; reviewer gate before merge |
| Runaway process explosion | Batching enforces concurrency cap below CLAUDE.md threshold |
| Existing repos archived prematurely | Phase 8 archival only after Jaypaul's final review; source repos retained read-only |
| TS SDK hand-write incorrect | Phase 4a core examples exercise every major API path; failures caught before tools depend on it |

**Rollback strategy:** Each phase commits to a feature branch and merges via reviewer gate. To roll back a phase, revert its merge commit on main. Source repos remain untouched until Phase 8.

---

## Success Criteria

The consolidation is complete when:

1. `MT-Canvus-Tools` exists on GitHub with the structure above
2. `docs/api-reference/` is the authoritative spec, derived from `mt-restapi-client`
3. Go, Python, and TypeScript SDKs all compile, lint, and pass their tests
4. All 24 core examples (8 × 3 languages) run successfully against a configured Canvus server
5. Every migrated tool/example builds and its README documents how to run it
6. `docs/conventions/{lang}.md` exists for each language and is referenced by the language-level READMEs
7. CI passes on `main` for all languages
8. Source repos in `gh/` carry a `MOVED-TO-MT-CANVUS-TOOLS.md` notice
9. `v1.0.0` is tagged and the repo is public

---

## Out-of-Scope Follow-ups (post-v1)

- C# / .NET SDK + examples (enterprise Windows customers)
- Java/Kotlin SDK (Android + enterprise)
- Rust SDK
- OpenAPI spec generation from `docs/api-reference/`
- Spec-drift CI job comparing `docs/api-reference/` to live `mt-restapi-client`
- Auto-publish to package registries (pkg.go.dev, PyPI, npm)
- Per-SDK semantic versioning automation

---

## Critical File References

- **Canonical API source:** `/home/jaypaulb/Projects/gl/conan/canvus/mt-restapi-client/RestApiClient/`
- **Integration test behaviour:** `/home/jaypaulb/Projects/gl/conan/canvus/mt-restapi-tests/tests/`
- **Pending doc changes:** `/home/jaypaulb/Projects/gl/conan/canvus/mt-restapi-client/doc-updates-for-developer-site.md`
- **Existing API doc reference:** `/home/jaypaulb/Projects/gl/knowledge-base.wiki/Canvus/developers/api-reference/rest-api/`
- **Source repos to migrate:** all under `/home/jaypaulb/Projects/gh/`
- **Target repo (to create):** `/home/jaypaulb/Projects/gh/MT-Canvus-Tools/`
- **Remote target:** `github.com/jaypaulb/MT-Canvus-Tools`

---

## Verification

End-to-end verification once implementation is complete:

1. Clone fresh: `git clone https://github.com/jaypaulb/MT-Canvus-Tools.git`
2. Each language: follow `docs/getting-started/<lang>.md` from a clean machine — should reach a working core example in under 10 minutes
3. Each tool: cd into its directory, follow its README, build + run successfully
4. CI green on `main`
5. API reference `coverage-matrix.md` shows 100% (or documented exceptions) for Go and Python; TypeScript shows full coverage of v1 endpoint scope
