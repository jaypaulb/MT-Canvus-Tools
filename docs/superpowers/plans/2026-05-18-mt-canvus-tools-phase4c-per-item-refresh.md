# Phase 4c — Per-Item Refresh Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port the 11 external Canvus tool/utility/example repos into the MT-Canvus-Tools monorepo, refreshing each against the new Go/Python/TypeScript SDKs.

**Architecture:** Each item gets its own directory under `go/cli/`, `go/tools/`, `go/examples/projects/`, `python/tools/`, or `typescript/examples/` per the consolidation design spec. Implementer agents are licensed to do a full assessment + refactor per item (Jaypaul's directive: "treat the prior code as the MVP we are building a better version of") — not just swap SDK imports. Items are batched 4 at a time, mixed-size per batch, with a review gate after each round.

**Tech Stack:** Go workspace (`go.work`), Python uv workspace (`python/pyproject.toml`), TypeScript pnpm workspace (`typescript/pnpm-workspace.yaml`). All items use the new `@mt-canvus-tools/sdk` (TS), `canvus-sdk` (Python), and `github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus` (Go) packages — none should reference the legacy `Canvus-Go-API` or `CanvusPythonAPI` modules.

**Scope decisions (locked in conversation 2026-05-18):**
- **Full assessment + refactor** per item — treat prior code as MVP, build a better version. Drop dead code, simplify, modernize, fix bugs.
- **No live-server smoke tests per item** — rely on SDK integration tests + each item's unit tests. Phase 3.t live verification holds.
- **ARCHIVED.md notes** dropped in each source repo after port lands.

---

## Item inventory (from consolidation-design spec)

| # | Source repo | Destination | Lang | Source LOC | Source files | Notes |
|---|---|---|---|---|---|---|
| 1 | `canvus-cli` | `go/cli/` | Go | 13,440 | 187 | Large; many subcommands. |
| 2 | `CanvusPowerToys` | `go/tools/powertoys/` | Go | 20,173 | 87 | Largest item; many utilities. |
| 3 | `CanvusTranslator` | `go/tools/translator/` | Go | 890 | 6 | Tiny. |
| 4 | `Canvus-Server-db-solver` | `go/tools/db-solver/` | Go | 9,906 | 67 | Large; DB-heavy. |
| 5 | `AI-personas` | `go/examples/projects/ai-personas/` | Go | 5,206 | 21 | Medium; LLM-driven. |
| 6 | `CanvusAPI-LLMDemo` | `go/examples/projects/llm-canvas-companion/` | Go | 4,466 | 11 | Medium; **rename**. |
| 7 | `CanvusNoteMapper` | `go/examples/projects/note-mapper/` | Go | 2,999 | 17 | Medium. |
| 8 | `canvus-mcp-server` | `python/tools/mcp-server/` | Python | ~5-10k real | many | Large; full MCP server. |
| 9 | `Canvus-Local-LLM` | `python/tools/local-llm/` | Python | 772 | 8 | Tiny. |
| 10 | `CanvusWebUI` | `typescript/examples/webui/` | TS | 6,521 | 15 | **Rewrite** — source is JavaScript. |

**Original spec listed an 11th item, `CanvusMCP` → `go/tools/mcp-server/`. Confirmed during Task 4c.0: no `CanvusMCP` repo exists upstream (no local copy, no GitHub repo under that name or alternates). Python `canvus-mcp-server` (item #8) covers the MCP use case. Phase 4c scope is therefore 10 items, not 11.**

**Batch allocation** (mixed sizes per round so each round produces a meaningful slice):

- **Round 1 (small validators, 4 items):** #3 translator (Go), #9 local-llm (Python), #7 note-mapper (Go), #6 llm-canvas-companion (Go).
- **Round 2 (medium + rewrites, 3 items):** #5 ai-personas (Go), #10 webui (TS rewrite), #8 mcp-server-python.
- **Round 3 (large items, 3 items):** #1 cli (Go), #4 db-solver (Go), #2 powertoys (Go).

Total: 10 items, 3 rounds, 4 + 3 + 3 agents (max parallel = 4, under the runaway-guard threshold).

---

## Task 4c.0: Pre-flight + workspace scaffolding

- [ ] **Step 1: Verify HEAD is post-Phase-4b**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
git status
git log --oneline | head -5
```

Expected: HEAD at least at `e1aaf55` ("docs: Phase 4b completion update + Phase 4d cleanup tracker refresh"), working tree clean.

- [ ] **Step 2: Confirm `CanvusMCP` Go source availability — RESOLVED 2026-05-18**

Not present locally. Searched `gh repo view jaypaulb/CanvusMCP` (404) and `gh search repos --owner jaypaulb mcp` (empty). The original consolidation spec listed it speculatively but no such repo was ever built. Python `canvus-mcp-server` (item #8) covers the MCP use case. **Item dropped from scope; Phase 4c is 10 items.**

- [ ] **Step 3: Scaffold destination directories**

```bash
mkdir -p go/cli go/tools go/examples/projects
mkdir -p python/tools
mkdir -p typescript/examples
```

(Most of these may already exist as siblings of `go/examples/core/`, `python/examples/core/`, `typescript/examples/core/`.)

- [ ] **Step 4: Pre-flight commit**

```bash
git add -A
git diff --cached --stat
# If nothing to commit, skip. Otherwise:
git commit -m "chore(phase4c): pre-flight directory scaffolding"
```

---

## Task 4c.1: Per-item assessment audit (single opus agent, background)

The Phase 4b experience taught us: audit before dispatching implementers. Each item's "full assessment + refactor" license could easily explode scope. The audit bounds each item's work before agents commit time to it.

- [ ] **Step 1: Single Agent tool call**

`subagent_type: general-purpose`, `model: opus`, `run_in_background: true`. Prompt:

```
You are auditing 11 Canvus tool/utility/example repos in /home/jaypaulb/Projects/gh/ for porting into the MT-Canvus-Tools monorepo. Jaypaul has approved a full-assessment + refactor approach: each port treats the source as the MVP and builds a better version. Your job is to bound the scope per item BEFORE implementer agents start.

# Source repos to audit (read-only)
| # | Source repo path | Destination in monorepo | Lang |
|---|---|---|---|
| 1 | /home/jaypaulb/Projects/gh/canvus-cli/ | go/cli/ | Go |
| 2 | /home/jaypaulb/Projects/gh/CanvusMCP/ (if present, else flag) | go/tools/mcp-server/ | Go |
| 3 | /home/jaypaulb/Projects/gh/CanvusPowerToys/ | go/tools/powertoys/ | Go |
| 4 | /home/jaypaulb/Projects/gh/CanvusTranslator/ | go/tools/translator/ | Go |
| 5 | /home/jaypaulb/Projects/gh/Canvus-Server-db-solver/ | go/tools/db-solver/ | Go |
| 6 | /home/jaypaulb/Projects/gh/AI-personas/ | go/examples/projects/ai-personas/ | Go |
| 7 | /home/jaypaulb/Projects/gh/CanvusAPI-LLMDemo/ | go/examples/projects/llm-canvas-companion/ | Go |
| 8 | /home/jaypaulb/Projects/gh/CanvusNoteMapper/ | go/examples/projects/note-mapper/ | Go |
| 9 | /home/jaypaulb/Projects/gh/canvus-mcp-server/ | python/tools/mcp-server/ | Python |
| 10 | /home/jaypaulb/Projects/gh/Canvus-Local-LLM/ | python/tools/local-llm/ | Python |
| 11 | /home/jaypaulb/Projects/gh/CanvusWebUI/ | typescript/examples/webui/ | TS (rewrite from JS) |

# Reference docs
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/conventions/{go,python,typescript}.md
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/VERIFIED-CORRECTIONS.md
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/parity-matrix.md (for SDK surface)

# Output
/home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/per-item-refresh-audit.md

# Per-item structure
For each item, produce a section:

## Item N — <source name> → <destination>
**Source LOC** | **File count** | **Effective LOC** (excluding generated/vendored)
**Purpose** — one paragraph: what does this tool actually do?
**Current SDK usage** — which legacy SDK calls are made? (cite source file:line)
**SDK call audit** — for each legacy SDK call, what's the equivalent in the new SDK? Are there gaps? (cross-reference parity-matrix.md)
**Architecture quality** — concrete assessment: dead code? duplication? questionable patterns? hard-to-test logic?
**Refresh complexity estimate** — trivial / moderate / large / major-rewrite
**Recommended port approach** — one paragraph: what should the implementer agent do? Drop X module entirely? Restructure Y? Just swap imports?
**Recommended model** — sonnet (mechanical) / opus (architectural judgement)
**External dependencies** — non-Canvus deps to preserve (e.g. ollama, anthropic SDK, sqlite, etc.)
**README contents** — what should the new README cover that the old one didn't?

# After per-item sections
## Cross-item observations
Patterns that span multiple items — shared helpers worth extracting? Common refactors? Duplicate code across two tools that should be unified?

## Items needing pre-port clarification
List any items where the source repo's intent is unclear or the spec mapping is ambiguous (e.g. CanvusMCP missing locally, two repos that overlap, etc.).

## Recommended batch reshuffle
The plan has Round 1 (translator, local-llm, note-mapper, llm-canvas-companion), Round 2 (ai-personas, webui, mcp-go, mcp-python), Round 3 (cli, db-solver, powertoys). Validate or propose changes.

# Critical rules
- READ ONLY. Do not modify source repos. Do not modify the monorepo.
- Cite file:line for every concrete claim.
- Be surgical. The audit's job is to bound scope, not to do the work.
- If a source repo is genuinely tiny (<1k LOC) AND uses the SDK trivially, you can recommend a one-pass refresh with sonnet — don't over-engineer.

# Length target
~3000-5000 words total. Per-item sections should be ~200-400 words each.
```

- [ ] **Step 2: Wait for completion**

You'll be notified. Don't poll.

- [ ] **Step 3: Read the audit + surface scope concerns**

Open `docs/api-reference/per-item-refresh-audit.md`. Skim:
- Total estimated complexity across all items
- Items flagged "major-rewrite" — surface to Jaypaul via `AskUserQuestion` before committing implementer time
- Cross-item observations — any shared utilities that should be a one-time extraction before per-item work?
- Batch reshuffle — apply the audit's recommendations if it found a better grouping

If the audit recommends changing batch order, update `Task 4c.3 / 4c.5 / 4c.7` headers accordingly. If a single item is estimated >8 hours of agent time, consider splitting it into a focused sub-task.

---

## Task 4c.2: Pre-port shared-helper extraction (conditional)

Only execute if the audit (4c.1) identifies shared utilities used by 2+ items that should live in the SDK extras rather than be copy-pasted across tools.

- [ ] **Step 1: Decide based on audit**

If audit found N shared helpers worth extracting: dispatch a focused agent to add them to the appropriate extras subpackage(s), commit, push. This keeps per-item refreshes free of "I had to re-port X" cycles.

If audit found nothing: skip this task and proceed directly to Task 4c.3.

---

## Task 4c.3: Batch 1 — small validators (4 items in parallel)

Items: `translator` (Go), `local-llm` (Python), `note-mapper` (Go), `llm-canvas-companion` (Go).

**Why these first:** smallest source LOC means lowest scope risk. Running them first validates the per-item agent contract (Step 2 below) and proves the workspace integration works before larger items pile in.

- [ ] **Step 1: Dispatch 4 implementer agents in parallel**

All 4 in a single message (single response with 4 `Agent` tool uses). Each agent prompt follows the **per-item implementer contract** template:

```
You are the Phase 4c implementer for the <item-name> tool.

# Required reading (in order)
1. /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/per-item-refresh-audit.md — Item N section. Your scope is bounded by what the audit prescribed.
2. /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/conventions/<lang>.md — language conventions for the destination SDK.
3. /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/VERIFIED-CORRECTIONS.md — wire-shape truth.
4. /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/parity-matrix.md §1 — new SDK method catalogue.
5. The source repo at /home/jaypaulb/Projects/gh/<source-name>/ — read every meaningful source file.

# Senior-dev license
Jaypaul has approved a full-assessment + refactor approach: treat the prior code as the MVP and build a better version. You may:
- Drop genuinely dead code (Chesterton fence: explain why it's dead in commit/PR notes).
- Rename for clarity.
- Restructure modules where the original layout is genuinely confusing.
- Fix bugs surfaced during the port.
- Replace bespoke utilities with SDK helpers where the SDK has parity.

You may NOT:
- Drop features without explicit audit approval.
- Rename CLI flags / public APIs in ways that break user-visible behavior unless the audit calls it out.
- Add new features the audit didn't anticipate.

# Implementation contract

## Step 1: Audit the source
Re-read the source repo's CLAUDE.md / README.md / top-level code. Confirm you agree with the audit's complexity estimate. If you find scope larger than estimated, surface to the orchestrator (return BLOCKED status with specifics) — do not silently expand.

## Step 2: Scaffold destination
Create the destination directory inside the monorepo (e.g. `go/tools/translator/`). Add a minimal module manifest (`go.mod`, `pyproject.toml`, or `package.json`) wired to the appropriate workspace.

## Step 3: Port code
Port logic file-by-file. For each Canvus SDK call:
- Old Go SDK: `github.com/jaypaulb/Canvus-Go-API/...` → New Go SDK: `github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus`
- Old Python SDK: `canvus_api` → New Python SDK: `canvus_sdk`
- TS: rewrite using `@mt-canvus-tools/sdk`

Use SDK helpers wherever the old code reimplemented something (e.g. raw HTTP for subscribe → `session.SubscribeNotes(...)`; manual geometry calcs → `extras.WidgetBoundingBox(...)`).

## Step 4: Tests
Port existing tests; add unit tests for any new helper code you write. Each item should have at least basic test coverage of its core logic. Do NOT add live-server smoke tests (Jaypaul's directive).

## Step 5: Workspace registration
Add the item's module to the appropriate workspace file:
- Go: append to `go/go.work` `use (...)` block.
- Python: append to `python/pyproject.toml` `[tool.uv.workspace] members` list.
- TS: append to `typescript/pnpm-workspace.yaml` `packages` list.

## Step 6: README
Author a fresh README.md for the destination explaining:
- What the tool does.
- How to build / run.
- What's different from the old repo (briefly — for future-Jaypaul orientation).
- Configuration (env vars, flags).
- Known limitations.

## Step 7: Toolchain verify
Run language-appropriate toolchain checks:
- Go: `GOMAXPROCS=2 go build -p 1 ./...`, `go vet ./...`, `go test ./...`.
- Python: `uv run --extra dev ruff check`, `uv run --extra dev pytest`.
- TS: `pnpm typecheck`, `pnpm build`, `pnpm test`.

## Step 8: Report
Report status, files touched, test counts, any deviations from audit recommendation, and any new design questions surfaced.

# Critical rules
- Do NOT touch other items' source or destination directories.
- Do NOT modify any SDK source (`go/sdk/`, `python/sdk/`, `typescript/sdk/`). If you find a real SDK gap, surface it for Phase 4d.
- Do NOT touch any source repo in `/home/jaypaulb/Projects/gh/<source-name>/`. Read-only.
- Do NOT commit — the orchestrator commits after the batch + review gate.
- Cite the parity-matrix work-item or audit recommendation in code comments at the top of significantly-rewritten files.
```

Per-item parameters:
- `translator`: `subagent_type=general-purpose`, `model=sonnet` (small, mechanical), `run_in_background=true`
- `local-llm`: `subagent_type=general-purpose`, `model=sonnet`, `run_in_background=true`
- `note-mapper`: `subagent_type=general-purpose`, `model=sonnet`, `run_in_background=true`
- `llm-canvas-companion`: `subagent_type=general-purpose`, `model=sonnet`, `run_in_background=true`

Override to `opus` if the audit recommended it.

- [ ] **Step 2: Wait for all 4 to complete**

You'll be notified. Don't poll. While waiting, you can begin reading the audit's Round 2 items so you're ready to dispatch them immediately.

- [ ] **Step 3: Per-item toolchain re-verify (orchestrator)**

For each completed item, run the toolchain commands yourself to confirm the agent's "clean" claim. Phase 4b had several cases where agents' self-reports were directionally wrong.

- [ ] **Step 4: Dispatch review-gate agent (background opus)**

`subagent_type: pr-review-toolkit:code-reviewer`, `model: opus`, `run_in_background: true`. Prompt covers all 4 items at once with file:line citations. Categorise: Critical / Important / Nice-to-have.

- [ ] **Step 5: Apply review fixes inline (or dispatch fix agent for big ones)**

Same pattern as Phase 4b: Critical and easy Importants go inline. Larger Importants get a fix agent. Deferred items added to the 4d tracker.

- [ ] **Step 6: Commit Round 1**

One commit per item (4 commits) for clean history. Each commit message:

```
feat(<lang>/<destination>): port <source-name> from gh/ (Phase 4c Round 1)

<summary of what changed vs. the source — refactors, dead code dropped,
SDK helpers adopted, tests added>

Source: github.com/jaypaulb/<source-name>
ARCHIVED.md follow-up in Task 4c.9.
```

Push after all 4 are committed.

---

## Task 4c.4: Batch 1 retrospective

- [ ] **Step 1: 30-second retro before dispatching Round 2**

What went well? What got stuck? Did any agents surface SDK gaps? Update the Round 2 prompts if anything specific needs adjusting (e.g. "always run `go vet` before declaring done", "explicit reminder to register in `go.work`").

If any major design question came out of Round 1, surface to Jaypaul before Round 2 fires.

---

## Task 4c.5: Batch 2 — medium items + TS rewrite (3 items in parallel)

Items: `ai-personas` (Go), `webui` (TS rewrite), `mcp-server` (Python).

**Why this batch:** webui is the only TS item and the only rewrite (source is JavaScript) — runs solo for its language so it doesn't conflict. Python mcp-server is the largest Python item. (The originally-planned Go mcp-server item was dropped — see 4c.0 Step 2.)

- [ ] **Step 1: Dispatch 3 implementer agents in parallel**

Use the same per-item implementer contract as Task 4c.3 Step 1. Per-item model:
- `ai-personas`: `opus` (LLM orchestration architecture)
- `webui` (TS rewrite): `opus` (full rewrite, architectural judgement)
- `mcp-server-python`: `opus` (large, full MCP server)

- [ ] **Step 3: Wait + per-item toolchain re-verify (orchestrator)**

Same as Task 4c.3 Step 3.

- [ ] **Step 4: Dispatch review-gate agent (background opus)**

Same as Task 4c.3 Step 4 but for Round 2 items.

- [ ] **Step 5: Apply review fixes**

Same pattern.

- [ ] **Step 6: Commit Round 2**

Same pattern: 4 commits, push after all 4.

---

## Task 4c.6: Batch 2 retrospective + Round 3 prep

- [ ] **Step 1: Retro**

Same questions as 4c.4 Step 1.

- [ ] **Step 2: Round 3 sizing check**

Round 3 has the three largest items (cli ~13k LOC, db-solver ~10k LOC, powertoys ~20k LOC). Decide:
- All 3 in parallel? Watch agent runtime — these are 2-4 hour items each.
- Or sequence them 1 at a time with reviews between?

Default: 3 in parallel (matches the spec's 4-6 batching guidance and stays under runaway-guard). If Round 1 + Round 2 exposed problems with very large parallel runs, sequence them instead.

---

## Task 4c.7: Batch 3 — large items (3 items in parallel)

Items: `cli` (Go), `db-solver` (Go), `powertoys` (Go).

- [ ] **Step 1: Dispatch 3 implementer agents in parallel**

All three get `opus` — large items benefit from architectural judgement. Same per-item implementer contract.

- [ ] **Step 2: Wait + per-item toolchain re-verify (orchestrator)**

Same as Task 4c.3 Step 3. These items are larger; expect 1-3 hours per agent.

- [ ] **Step 3: Dispatch review-gate agent (background opus)**

Same as Task 4c.3 Step 4 but for Round 3.

- [ ] **Step 4: Apply review fixes**

Same pattern.

- [ ] **Step 5: Commit Round 3**

3 commits, push.

---

## Task 4c.8: Workspace integration verification

After all 11 items landed, verify everything builds together from the monorepo root.

- [ ] **Step 1: Go workspace cross-build**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go
GOMAXPROCS=2 go build -p 1 ./...
GOMAXPROCS=2 go vet ./...
GOMAXPROCS=2 go test -count=1 ./...
```

Expected: every module in `go.work` builds clean and tests pass.

- [ ] **Step 2: Python workspace cross-build**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/python
uv sync --all-packages --extra dev
uv run --all-packages --extra dev ruff check
uv run --all-packages --extra dev pytest
```

Expected: every member in the uv workspace passes ruff and pytest.

- [ ] **Step 3: TS workspace cross-build**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript
pnpm install
pnpm -r build
pnpm -r typecheck
pnpm -r test
```

Expected: every workspace package builds, typechecks, and tests pass.

- [ ] **Step 4: Fix any cross-package breakage**

If a previously-passing toolchain check breaks at the workspace level (likely cause: an item's `go.mod`/`pyproject.toml`/`package.json` references a version constraint that conflicts with another item), fix in place.

- [ ] **Step 5: Commit integration fixes (if any)**

```bash
git add <files>
git commit -m "chore(phase4c): workspace integration fixes"
git push
```

---

## Task 4c.9: ARCHIVED.md notes in source repos

For each ported source repo, drop a one-line `ARCHIVED.md` pointer to the new monorepo location.

- [ ] **Step 1: Loop through the 11 source repos**

For each item N with source `<source-path>` and destination `<dest-path>`:

```bash
cat > /home/jaypaulb/Projects/gh/<source-path>/ARCHIVED.md <<EOF
# Archived

This repo has been refreshed into the MT-Canvus-Tools monorepo:

→ https://github.com/jaypaulb/MT-Canvus-Tools/tree/main/<dest-path>

Phase 4c (per-item refresh), 2026-05-18.

The new location uses the consolidated MT-Canvus-Tools SDKs and follows
the conventions in \`docs/conventions/\`. Open issues/PRs against the
monorepo, not this archive.
EOF
```

- [ ] **Step 2: Commit ARCHIVED.md notes (per source repo)**

Each source repo is a separate git repository. For each:

```bash
cd /home/jaypaulb/Projects/gh/<source-path>
git add ARCHIVED.md
git commit -m "chore: archive — refreshed into MT-Canvus-Tools/<dest-path>"
git push origin <default-branch>
```

If a source repo has uncommitted local changes, surface to Jaypaul — don't blindly stash or override.

If a source repo has no GitHub remote (or push fails), note it and proceed with the others — surface the failures at the end.

---

## Task 4c.10: Final review gate + CONSOLIDATION-STATUS update

- [ ] **Step 1: Dispatch final review-gate agent (background opus)**

`subagent_type: pr-review-toolkit:code-reviewer`, `model: opus`, `run_in_background: true`. Prompt:

```
Phase 4c per-item refresh final review. 11 items have been ported into the MT-Canvus-Tools monorepo:
- go/cli/, go/tools/{mcp-server,powertoys,translator,db-solver}/
- go/examples/projects/{ai-personas,llm-canvas-companion,note-mapper}/
- python/tools/{mcp-server,local-llm}/
- typescript/examples/webui/

# Commits to review
git log --oneline e1aaf55..HEAD

# Required reading
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/per-item-refresh-audit.md — the original scope per item.
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/conventions/{go,python,typescript}.md.
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/CONSOLIDATION-STATUS.md.

# Review focus
1. Per-item: did each port match its audit's scope, or did the agent over/under-build?
2. SDK usage: every legacy SDK import gone? Any tool reimplementing functionality that the SDK now provides?
3. Workspace integration: every item registered in go.work / pyproject.toml / pnpm-workspace.yaml? Cross-build clean?
4. Convention adherence: each item follows its language's conventions (file layout, error handling, logging, test style)?
5. README quality: each destination has a meaningful README — not a stub?
6. Tests: each item has unit tests for its core logic; ports preserved any meaningful pre-existing tests.
7. Cross-item duplication: any helper code copy-pasted across 2+ tools that should be extracted to extras?

# Output structure
Same as Phase 4b review: per-area verdict, Critical/Important/Nice-to-have list with file:line, overall verdict (✅/⚠/❌), Phase 4d cleanup tracker additions.

# Length: 3000 words max.
```

- [ ] **Step 2: Apply Critical and Important fixes from review**

Same pattern as Phase 4b.

- [ ] **Step 3: Update CONSOLIDATION-STATUS.md**

Mark Phase 4c complete. Add a per-item commit table. Move 4c-deferred items into the Phase 4d tracker.

- [ ] **Step 4: Commit + push the status update**

```bash
git add CONSOLIDATION-STATUS.md
git commit -m "docs: Phase 4c completion update + 4d tracker refresh"
git push
```

---

## Risks & mitigations

| Risk | Mitigation |
|---|---|
| `CanvusMCP` (Go) genuinely doesn't exist | Audit Task 4c.1 surfaces; either clone from GitHub if found, or skip and document. Not a blocker. |
| Source repo dependencies break under newer Go/Python/TS toolchain | Each implementer agent must run the toolchain — failures surface immediately, not at integration time. |
| `canvus-mcp-server` Python item is genuinely 5-10k+ real LOC | Audit estimate per Task 4c.1 — if "major-rewrite", split into a focused sub-task and possibly run alone. |
| Agent over-scopes "full assessment + refactor" license and rewrites half the codebase | Audit bounds scope per item; implementer is instructed to return BLOCKED if scope grows beyond audit. Review gate catches over-builds. |
| Cross-item dependency loops (e.g. powertoys depends on a helper in cli) | Audit Task 4c.1 cross-item observations catches; pre-port shared-helper extraction (Task 4c.2) resolves. |
| Workspace integration breaks at Task 4c.8 due to package-version drift | Fix inline; document the resolution in the integration commit message so future items don't repeat. |

---

## What this plan does NOT include

- **`CanvusCustomMenuExample + canvus-menu`** — these are scheduled for Phase 5 (custom-menu consolidation), not 4c.
- **`CanvusServerInstaller`** — dropped per the original spec ("containerised deployment replaces").
- **`CanvusLite`, `CanvusConsole`** — Flutter apps; not in the consolidation scope.
- **Reclaiming TS coverage thresholds, fixing pre-existing lint debt, mypy strict cleanup** — Phase 4d.
- **Live-server verification of permissions-subscribe** — Phase 4d per the 4b review.
- **Plug into CI/CD** — Phase 7.

---

## Done criteria

- 11 items ported and committed.
- Each item has a passing build + test pass in its language.
- `go.work`, `python/pyproject.toml`, `typescript/pnpm-workspace.yaml` updated.
- Workspace-level cross-build (Task 4c.8) clean for all three languages.
- `ARCHIVED.md` notes pushed to each ported source repo (or failures documented).
- `CONSOLIDATION-STATUS.md` reflects Phase 4c done.
- Phase 4d cleanup tracker updated with anything 4c review surfaced and deferred.
- Final review gate verdict is ✅ ship or ⚠ ship-with-deferrals.

Phase 4d (cleanup) ready to plan when Jaypaul gives the go-ahead.
