# MT-Canvus-Tools Foundation Implementation Plan (Phase 0–2)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish the `MT-Canvus-Tools` monorepo on GitHub with an authoritative API reference and locked conventions docs, so Phase 3+ agents can run autonomously against a stable foundation.

**Architecture:** Foreground bootstrap creates the repo skeleton and pushes to GitHub. Two waves of parallel agent dispatches then populate `docs/api-reference/` (Phase 1) and `docs/conventions/` + `coverage-matrix.md` (Phase 2). Each phase commits before the next begins; no agent waits on another within a phase.

**Tech Stack:** git, `gh` CLI, GitHub Actions (Phase 7), Bash, Claude Agent tool (Explore + Plan subagent types).

**Reference design spec:** `/home/jaypaulb/Projects/gh/docs/superpowers/specs/2026-05-17-mt-canvus-tools-consolidation-design.md`

---

## File Structure (created by this plan)

```
/home/jaypaulb/Projects/gh/MT-Canvus-Tools/
├── README.md                                    # Phase 0
├── LICENSE                                      # Phase 0 (license tbd by Jaypaul)
├── CONTRIBUTING.md                              # Phase 0 (stub)
├── .gitignore                                   # Phase 0
├── docs/
│   ├── superpowers/specs/
│   │   └── 2026-05-17-mt-canvus-tools-consolidation-design.md   # Phase 0 (copy)
│   ├── superpowers/plans/
│   │   └── 2026-05-17-mt-canvus-tools-foundation.md             # Phase 0 (copy of this file)
│   ├── api-reference/
│   │   ├── README.md                            # Phase 1
│   │   ├── SOURCE.md                            # Phase 1 (spec-freeze SHAs)
│   │   ├── authentication.md                    # Phase 1
│   │   ├── widget-types.md                      # Phase 1
│   │   ├── streaming.md                         # Phase 1
│   │   ├── changelog.md                         # Phase 1
│   │   ├── endpoints/                           # Phase 1 (one .md per resource group)
│   │   │   ├── canvases.md
│   │   │   ├── widgets.md
│   │   │   ├── auth.md
│   │   │   ├── users.md
│   │   │   ├── folders.md
│   │   │   ├── assets.md
│   │   │   └── server.md
│   │   ├── examples/                            # Phase 1 (request/response pairs)
│   │   └── coverage-matrix.md                   # Phase 2
│   └── conventions/
│       ├── go.md                                # Phase 2
│       ├── python.md                            # Phase 2
│       └── typescript.md                        # Phase 2
└── (language dirs scaffolded in Phase 3+)
```

---

## Phase 0: Bootstrap

Foreground orchestrator work. No agent dispatches.

### Task 0.1: Confirm licence choice

- [ ] **Step 1: Ask Jaypaul for licence decision**

Use `AskUserQuestion` with options: MIT, Apache-2.0, Proprietary (no licence file).

Capture the answer in a variable used in Task 0.4.

### Task 0.2: Create the local directory and scaffold structure

**Files:**
- Create: `/home/jaypaulb/Projects/gh/MT-Canvus-Tools/` (and subdirs)

- [ ] **Step 1: Create directory tree**

```bash
cd /home/jaypaulb/Projects/gh
mkdir -p MT-Canvus-Tools/docs/{api-reference/{endpoints,examples},conventions,getting-started,contributing}
mkdir -p MT-Canvus-Tools/docs/superpowers/{specs,plans}
```

- [ ] **Step 2: Verify the tree**

```bash
find /home/jaypaulb/Projects/gh/MT-Canvus-Tools -type d | sort
```

Expected output includes: `docs`, `docs/api-reference`, `docs/api-reference/endpoints`, `docs/api-reference/examples`, `docs/conventions`, `docs/getting-started`, `docs/contributing`, `docs/superpowers/specs`, `docs/superpowers/plans`.

### Task 0.3: Initialise git and create base files

**Files:**
- Create: `MT-Canvus-Tools/.gitignore`
- Create: `MT-Canvus-Tools/README.md`
- Create: `MT-Canvus-Tools/CONTRIBUTING.md`

- [ ] **Step 1: Initialise git**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
git init -b main
```

Expected: `Initialized empty Git repository in /home/jaypaulb/Projects/gh/MT-Canvus-Tools/.git/`

- [ ] **Step 2: Write .gitignore**

```gitignore
# OS / editor
.DS_Store
*.swp
.vscode/
.idea/

# Go
*.exe
*.dll
*.so
*.dylib
*.test
*.out
vendor/

# Python
__pycache__/
*.py[cod]
*.egg-info/
.venv/
.pytest_cache/
.ruff_cache/
.mypy_cache/

# Node / TypeScript
node_modules/
dist/
build/
*.tsbuildinfo
.pnpm-store/

# Secrets / local config
.env
.env.local
settings.json
*.local
```

- [ ] **Step 3: Write README.md**

```markdown
# MT-Canvus-Tools

A monorepo of SDKs, tools, and examples for building on the Canvus collaborative infinite canvas platform.

## Languages

| Language | SDK | Examples | Tools |
|---|---|---|---|
| Go | `go/sdk/` | `go/examples/` | `go/cli/`, `go/tools/` |
| Python | `python/sdk/` | `python/examples/` | `python/tools/` |
| TypeScript | `typescript/sdk/` | `typescript/examples/` | — |

## Status

🚧 Under active consolidation. See [the consolidation spec](docs/superpowers/specs/2026-05-17-mt-canvus-tools-consolidation-design.md).

## Documentation

- [API Reference](docs/api-reference/README.md) — canonical Canvus REST API spec
- [Conventions](docs/conventions/) — language-specific build, lint, and code conventions
- [Getting Started](docs/getting-started/) — per-language quickstarts
- [Contributing](CONTRIBUTING.md)
```

- [ ] **Step 4: Write CONTRIBUTING.md (stub)**

```markdown
# Contributing to MT-Canvus-Tools

> This file will be expanded in Phase 6 of the consolidation.

For now: see the language-specific READMEs under `go/`, `python/`, and `typescript/` once those phases are complete.

## Convention docs

Each language has its own convention doc in `docs/conventions/`. These lock build/lint/log/error/test defaults so contributions stay consistent. New convention decisions must be appended to the relevant doc with rationale.
```

### Task 0.4: Write the licence file

**Files:**
- Create: `MT-Canvus-Tools/LICENSE`

- [ ] **Step 1: Write LICENSE per Jaypaul's choice from Task 0.1**

If MIT:
```
MIT License

Copyright (c) 2026 Jaypaul Barrow

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

If Apache-2.0: use the standard Apache 2.0 text (192 lines). Source: https://www.apache.org/licenses/LICENSE-2.0.txt

If Proprietary: write `LICENSE` with text:
```
Copyright (c) 2026 Jaypaul Barrow. All rights reserved.

This software is proprietary. Unauthorised copying, modification, distribution,
or use of this software, via any medium, is strictly prohibited without prior
written permission from the copyright holder.
```

### Task 0.5: Copy design spec and this plan into the new repo

**Files:**
- Copy: `docs/superpowers/specs/2026-05-17-mt-canvus-tools-consolidation-design.md`
- Copy: `docs/superpowers/plans/2026-05-17-mt-canvus-tools-foundation.md`

- [ ] **Step 1: Copy spec and plan**

```bash
cp /home/jaypaulb/Projects/gh/docs/superpowers/specs/2026-05-17-mt-canvus-tools-consolidation-design.md \
   /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/superpowers/specs/

cp /home/jaypaulb/Projects/gh/docs/superpowers/plans/2026-05-17-mt-canvus-tools-foundation.md \
   /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/superpowers/plans/
```

### Task 0.6: Create initial commit

- [ ] **Step 1: Stage and commit**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
git add README.md LICENSE CONTRIBUTING.md .gitignore docs/
git commit -m "chore: bootstrap MT-Canvus-Tools monorepo

Initial structure for the Canvus API consumer consolidation per
docs/superpowers/specs/2026-05-17-mt-canvus-tools-consolidation-design.md"
```

- [ ] **Step 2: Verify**

```bash
git log --oneline
git status
```

Expected: one commit on `main`, working tree clean.

### Task 0.7: Create GitHub repo and push

- [ ] **Step 1: Create private GH repo**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
gh repo create jaypaulb/MT-Canvus-Tools --private --source=. --description "MultiTaction Canvus SDKs, tools, and examples (Go, Python, TypeScript)" --push
```

Expected output: `✓ Created repository jaypaulb/MT-Canvus-Tools on GitHub` followed by a push log.

- [ ] **Step 2: Verify remote**

```bash
gh repo view jaypaulb/MT-Canvus-Tools --json url,visibility
```

Expected: `"visibility": "PRIVATE"` and the repo URL.

---

## Phase 1: API Spec Extraction

Three agents dispatched in parallel via the Agent tool (`subagent_type: general-purpose`, `model: haiku` for cost — the work is heavy reading + structured writing, not deep reasoning). All three sent in a single message so they run concurrently.

> **Note on subagent type:** `Explore` and `Plan` subagent types cannot use `Write`. Phase 1 needs to write Markdown files, so `general-purpose` is the correct subagent type with `haiku` model for cost.

### Task 1.1: Dispatch all three Phase-1 Explore agents in parallel

**Files (output):**
- Created by Agent 1.1: `MT-Canvus-Tools/docs/api-reference/endpoints/*.md`, `authentication.md`, `widget-types.md`, `streaming.md`, `README.md`
- Created by Agent 1.2: `MT-Canvus-Tools/docs/api-reference/examples/*.md`
- Created by Agent 1.3: `MT-Canvus-Tools/docs/api-reference/changelog.md`

- [ ] **Step 1: Send single message with three Agent tool calls**

Agent 1.1 prompt:

```
You are an Explore agent extracting the Canvus REST API endpoint catalogue from the C++ canonical implementation. The user is Jaypaul; you are working autonomously and will not be able to ask questions.

Read sources:
- /home/jaypaulb/Projects/gl/conan/canvus/mt-restapi-client/RestApiClient/include/mt-restapi-client/
- /home/jaypaulb/Projects/gl/conan/canvus/mt-restapi-client/RestApiClient/src/
- /home/jaypaulb/Projects/gl/conan/canvus/mt-restapi-client/README.md
- /home/jaypaulb/Projects/gl/knowledge-base.wiki/Canvus/developers/api-reference/rest-api/ (existing public docs, for endpoint structure reference; treat as possibly out of date)

Produce these output files under /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/ :

1. README.md — index of what's in this folder, how the catalogue is organised, and the source SHA freeze.

2. endpoints/canvases.md — every endpoint on /api/v1/canvases/* . For each: HTTP verb, full path with path params, query params (name, type, required, description), request body schema (JSON, with field types and required/optional), response body schema, auth required (api-key / login-token / none), streaming-capable (yes/no), status (implemented / experimental / deprecated).

3. endpoints/widgets.md — every endpoint on /api/v1/canvases/{canvasId}/{widgetType}/* for each widget type (notes, images, videos, browsers, pdfs, anchors, connectors). Same per-endpoint format.

4. endpoints/auth.md — login, logout, token operations.

5. endpoints/users.md — user CRUD, group operations.

6. endpoints/folders.md — folder/asset organisation endpoints.

7. endpoints/assets.md — file upload, asset retrieval.

8. endpoints/server.md — server info, audit, admin operations.

9. authentication.md — narrative explanation of the two auth modes (API key + login token), how to obtain each, how to send them, scope/expiry rules.

10. widget-types.md — narrative explanation of each widget type's data model and lifecycle.

11. streaming.md — narrative explanation of the subscribe=true pattern, how the server streams updates, how clients should consume the stream.

Output format for endpoint entries (use this exact template):

### `VERB /api/v1/path/{param}`

**Auth:** api-key | login-token | none
**Streaming:** yes (subscribe=true) | no
**Status:** implemented | experimental | deprecated

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| param | string | description |

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|

**Request body:** none | JSON schema (table or example)

**Response (200):** JSON schema (table or example)

**Errors:** common error codes (400, 401, 403, 404, 409, 500) with conditions

---

When uncertain about a field's type or behaviour, infer from C++ source and add a "Source uncertainty" callout in the entry. Do NOT invent endpoints not in source.

Do not modify files outside /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/ . Do not run git commands. Do not commit.

When complete, report: number of files written, total endpoint count by resource group, list of "Source uncertainty" callouts.
```

Agent 1.2 prompt:

```
You are an Explore agent extracting behavioural examples from the Canvus REST API integration tests. The user is Jaypaul; you are working autonomously and will not be able to ask questions.

Read sources:
- /home/jaypaulb/Projects/gl/conan/canvus/mt-restapi-tests/tests/
- /home/jaypaulb/Projects/gl/conan/canvus/mt-restapi-tests/testdata/
- /home/jaypaulb/Projects/gl/conan/canvus/mt-restapi-tests/README.md

Produce one Markdown file per resource group under /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/examples/ :

- canvases.md
- widgets.md
- auth.md
- users.md
- folders.md
- assets.md
- streaming.md

For each: extract working request/response pairs from the test source (Go test code, fixtures, expected JSON). For each example use this format:

### Example: <short description, e.g. "Create a sticky note with default colour">

**Source:** path/to/test/file.go:line

**Request:**
\`\`\`http
POST /api/v1/canvases/abc123/notes HTTP/1.1
Private-Token: <api-key>
Content-Type: application/json

{
  "text": "Hello",
  "location": {"x": 100, "y": 200},
  ...
}
\`\`\`

**Response (201):**
\`\`\`json
{
  "id": "...",
  "type": "note",
  ...
}
\`\`\`

**Notes:** any preconditions, side effects, or gotchas observed in the test.

Skip examples that exercise internal/undocumented behaviour. Prefer happy-path examples; include 1-2 error-path examples per resource group where the test demonstrates a documented failure mode.

Do not modify files outside /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/examples/ . Do not run git commands. Do not commit.

When complete, report: number of examples extracted per file.
```

Agent 1.3 prompt:

```
You are an Explore agent extracting pending API documentation changes. The user is Jaypaul; you are working autonomously and will not be able to ask questions.

Read source:
- /home/jaypaulb/Projects/gl/conan/canvus/mt-restapi-client/doc-updates-for-developer-site.md

Produce /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/changelog.md with this structure:

# API Changelog

> Pending documentation updates extracted from `mt-restapi-client/doc-updates-for-developer-site.md` as of <date>.
> Each entry describes a change from the current public developer documentation that downstream SDKs must reflect.

## Updates

### <Update title>

**Type:** new endpoint | changed behaviour | new parameter | deprecation | clarification
**Endpoint(s) affected:** `VERB /api/v1/path`
**Source location in doc-updates note:** line N

**Summary:** one-line summary.

**Detail:** the full change as described in the source note. Quote verbatim where unambiguous; paraphrase for clarity where the source is rambling.

**Action required for SDKs:** explicit list of what Go/Python/TypeScript SDKs need to do to implement this change.

---

Process every section in the source note. Order entries by significance (new endpoints first, deprecations last).

Do not modify files outside /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/changelog.md . Do not run git commands. Do not commit.

When complete, report: number of updates extracted, count by type.
```

Send all three Agent tool calls in a single assistant message.

- [ ] **Step 2: Verify all three agents completed**

After all three return, list output files:

```bash
find /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference -type f -name "*.md" | sort
```

Expected: README.md, SOURCE.md (created in Task 1.2), authentication.md, widget-types.md, streaming.md, changelog.md, endpoints/*.md (7 files), examples/*.md (7 files). Total ≥ 18 markdown files.

- [ ] **Step 3: Quick sanity scan of each output**

```bash
wc -l /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/endpoints/*.md
```

Expected: each endpoint file is > 50 lines (a 20-line endpoint file means the agent stub'd out).

If any agent produced thin output, re-dispatch it with the same prompt before proceeding.

### Task 1.2: Spec-freeze the source SHAs

**Files:**
- Create: `MT-Canvus-Tools/docs/api-reference/SOURCE.md`

- [ ] **Step 1: Capture source SHAs**

```bash
cd /home/jaypaulb/Projects/gl/conan/canvus/mt-restapi-client
CLIENT_SHA=$(git rev-parse HEAD 2>/dev/null || echo "not-a-git-repo")
CLIENT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "n/a")

cd /home/jaypaulb/Projects/gl/conan/canvus/mt-restapi-tests
TESTS_SHA=$(git rev-parse HEAD 2>/dev/null || echo "not-a-git-repo")
TESTS_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "n/a")

echo "Client: $CLIENT_BRANCH @ $CLIENT_SHA"
echo "Tests:  $TESTS_BRANCH @ $TESTS_SHA"
```

Record the printed SHAs.

- [ ] **Step 2: Write SOURCE.md**

Replace `<CLIENT_SHA>`, `<CLIENT_BRANCH>`, `<TESTS_SHA>`, `<TESTS_BRANCH>` with values from Step 1 and write:

```markdown
# Source Freeze

This API reference was extracted from the following source-of-truth commits. Downstream SDK phases reference *this captured state*, not a moving upstream target.

| Source | Path | Branch | SHA |
|---|---|---|---|
| `mt-restapi-client` (C++ canonical impl) | `gl/conan/canvus/mt-restapi-client` | `<CLIENT_BRANCH>` | `<CLIENT_SHA>` |
| `mt-restapi-tests` (Go integration tests) | `gl/conan/canvus/mt-restapi-tests` | `<TESTS_BRANCH>` | `<TESTS_SHA>` |

Extracted: <YYYY-MM-DD>

## Re-extraction policy

A future post-v1 phase may re-extract from newer SHAs and produce a diff against this captured spec to surface upstream drift. Until then, all SDKs and tools in this monorepo are pinned to the behaviour documented from these SHAs.
```

### Task 1.3: Commit Phase-1 output

- [ ] **Step 1: Stage and commit**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
git add docs/api-reference/
git status
```

Verify staged files: README.md, SOURCE.md, authentication.md, widget-types.md, streaming.md, changelog.md, endpoints/*.md, examples/*.md.

- [ ] **Step 2: Commit**

```bash
git commit -m "docs(api-reference): extract canonical API spec from mt-restapi-client

Phase 1 of the consolidation: API endpoint catalogue, behavioural examples,
and pending changelog extracted from mt-restapi-client and mt-restapi-tests.

Source SHAs captured in docs/api-reference/SOURCE.md."

git push origin main
```

---

## Phase 2: SDK Audit + Conventions Doc

Five agents dispatched in parallel via the Agent tool (`subagent_type: general-purpose`, `model: opus` for deep audit + design reasoning). All five sent in a single message.

> **Note on subagent type:** Same as Phase 1 — `general-purpose` is required so agents can `Write` their output files.

### Task 2.1: Dispatch all five Phase-2 Plan agents in parallel

**Files (output):**
- Created by Agent 2.1 + 2.2: `MT-Canvus-Tools/docs/api-reference/coverage-matrix.md`
- Created by Agent 2.3: `MT-Canvus-Tools/docs/conventions/go.md`
- Created by Agent 2.4: `MT-Canvus-Tools/docs/conventions/python.md`
- Created by Agent 2.5: `MT-Canvus-Tools/docs/conventions/typescript.md`

**Coordination note:** Agents 2.1 and 2.2 both write to `coverage-matrix.md`. To avoid write conflict, each writes to a *separate* intermediate file, then the orchestrator merges them in Task 2.2. Agent 2.1 writes `_coverage-go.md`; Agent 2.2 writes `_coverage-python.md`. Orchestrator merges and deletes intermediates.

- [ ] **Step 1: Send single message with five Agent tool calls**

Agent 2.1 prompt:

```
You are a Plan agent auditing the existing Canvus Go SDK against the freshly extracted canonical API spec. The user is Jaypaul; you are working autonomously and will not be able to ask questions.

Read sources:
- Spec: /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/ (all files)
- Spec freeze SHAs: /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/SOURCE.md
- Existing Go SDK: /home/jaypaulb/Projects/gh/Canvus-Go-API/ (all .go files)
- Pending changes: /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/changelog.md

For every endpoint in the spec, determine the Go SDK status:
- **implemented** — Go SDK method exists, signature matches current spec, behaviour matches
- **missing** — no Go SDK method for this endpoint
- **outdated** — Go SDK method exists but signature/behaviour drifted from current spec (specify what changed)
- **deprecated** — Go SDK has the method but the spec marks the endpoint deprecated

Produce /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/_coverage-go.md with this format:

# Go SDK Coverage

Audited: <date>
Spec freeze: <commit SHA from SOURCE.md>
SDK source: /home/jaypaulb/Projects/gh/Canvus-Go-API @ <SDK commit SHA>

## Summary

| Resource group | Total endpoints | Implemented | Missing | Outdated | Deprecated |
|---|---|---|---|---|---|
| Canvases | N | N | N | N | N |
| Widgets | N | N | N | N | N |
| (etc) | | | | | |

## Per-endpoint detail

### Canvases

| Endpoint | Status | Go method | Notes |
|---|---|---|---|
| `GET /api/v1/canvases` | implemented | `Session.ListCanvases` | — |
| `POST /api/v1/canvases/{id}/widgets/clone` | missing | — | New endpoint per changelog #N |
| ... | | | |

### Widgets

| Endpoint | Status | Go method | Notes |
|---|---|---|---|

(continue for all resource groups)

## Phase 3 work items

Ordered list of remediation tasks the Phase 3 Go SDK agent must execute. For each: file path in /home/jaypaulb/Projects/gh/Canvus-Go-API/ , what to add/change/remove, why.

Do not modify files outside /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/_coverage-go.md . Do not run git commands. Do not commit.

When complete, report: total endpoints, percent implemented, count of work items for Phase 3.
```

Agent 2.2 prompt:

```
You are a Plan agent auditing the existing Canvus Python SDK against the freshly extracted canonical API spec. The user is Jaypaul; you are working autonomously and will not be able to ask questions.

Read sources:
- Spec: /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/ (all files)
- Spec freeze SHAs: /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/SOURCE.md
- Existing Python SDK: /home/jaypaulb/Projects/gh/CanvusPythonAPI/ (all .py files)
- Pending changes: /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/changelog.md

For every endpoint in the spec, determine the Python SDK status using the same taxonomy as the Go audit (implemented / missing / outdated / deprecated).

Produce /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/_coverage-python.md with the same format as the Go audit but for Python. Replace `Go method` column with `Python method` (use dotted path: `module.Class.method`).

End with a "Phase 3 work items" section listing remediation tasks for the Phase 3 Python SDK agent.

Do not modify files outside /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/_coverage-python.md . Do not run git commands. Do not commit.

When complete, report: total endpoints, percent implemented, count of work items for Phase 3.
```

Agent 2.3 prompt:

```
You are a Plan agent authoring the Go conventions doc for the MT-Canvus-Tools monorepo. The user is Jaypaul; you are working autonomously and will not be able to ask questions.

This doc locks the build/lint/log/error/test/CI defaults for all Go code in the repo. Phase 4b refresh agents follow this doc without prompting Jaypaul. Pick standard, modern, idiomatic Go practice as of 2026.

Read for context:
- Spec: /home/jaypaulb/Projects/gh/docs/superpowers/specs/2026-05-17-mt-canvus-tools-consolidation-design.md
- Existing Go code patterns: /home/jaypaulb/Projects/gh/Canvus-Go-API/, /home/jaypaulb/Projects/gh/canvus-cli/, /home/jaypaulb/Projects/gh/CanvusPowerToys/

Produce /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/conventions/go.md with these sections:

# Go Conventions

## Module structure
- Go workspace at `go/go.work`, modules at `go/sdk`, `go/cli`, `go/examples/core/*`, `go/examples/projects/*`, `go/tools/*`
- Module paths under `github.com/jaypaulb/MT-Canvus-Tools/go/...`
- Internal packages: `internal/` for non-exported helpers
- Public API surface: top-level packages only

## Go version
- Minimum: 1.22 (specify in go.mod toolchain directive)
- CI matrix: 1.22, 1.23 (current stable)

## Build & tooling
- Build: `go build ./...`
- Format: `gofmt -s -w .` (no goimports needed beyond this)
- Vet: `go vet ./...`

## Linting
- Tool: `golangci-lint` v1.60+
- Config: `.golangci.yml` at repo root with these linters enabled: errcheck, gosimple, govet, ineffassign, staticcheck, unused, gofmt, goimports, revive, gosec
- All Go modules share this config

## Logging
- Standard library `log/slog` with `slog.Default()`
- JSON handler in production, text handler in development (driven by env var `LOG_FORMAT`)
- Levels: Debug, Info, Warn, Error
- No third-party logging libraries

## Error handling
- Wrap errors with `fmt.Errorf("context: %w", err)`
- Sentinel errors for known failure modes: `var ErrNotFound = errors.New("not found")`
- Use `errors.Is` and `errors.As` for inspection
- Don't define error types unless callers need structured fields
- HTTP API errors: typed struct `*APIError` with status code and message

## Configuration
- Order of precedence: command-line flags > env vars > config file > defaults
- Config file format: YAML, parsed with `gopkg.in/yaml.v3`
- Env var prefix: `CANVUS_` (e.g., `CANVUS_API_KEY`, `CANVUS_BASE_URL`)

## Testing
- Standard library `testing` + `github.com/stretchr/testify/assert`
- Table-driven tests for any function with >2 test cases
- Test files: `*_test.go` next to source
- Integration tests: build tag `// +build integration`, run with `go test -tags=integration`
- Coverage target: 70% line coverage for SDK code, no hard target for examples/tools

## CI
- GitHub Actions workflow `.github/workflows/go.yml`
- Matrix on Go 1.22 and 1.23, Ubuntu only
- Steps: checkout, setup-go, cache modules, lint, vet, test (with coverage), build

## Code style
- Public functions documented with godoc comments starting with the function name
- Receiver names: short and consistent (1-2 letters)
- No global mutable state in libraries; SDK uses a `Session` value passed by caller
- Context: every public function that does I/O takes `ctx context.Context` as first arg

## Convention amendments

When a Phase 4b refresh agent encounters a new decision not covered above, it MUST append the decision and rationale to the "Amendments" section below. Format:

### YYYY-MM-DD — <one-line summary>

**Item:** <tool or example being refreshed>
**Decision:** <what was chosen>
**Rationale:** <why>

Do not modify files outside /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/conventions/go.md . Do not run git commands. Do not commit.

When complete, report: section count, total length in lines.
```

Agent 2.4 prompt:

```
You are a Plan agent authoring the Python conventions doc for the MT-Canvus-Tools monorepo. The user is Jaypaul; you are working autonomously and will not be able to ask questions.

This doc locks the build/lint/log/error/test/CI defaults for all Python code in the repo. Phase 4b refresh agents follow this doc without prompting Jaypaul. Pick standard, modern, idiomatic Python practice as of 2026.

Read for context:
- Spec: /home/jaypaulb/Projects/gh/docs/superpowers/specs/2026-05-17-mt-canvus-tools-consolidation-design.md
- Existing Python code patterns: /home/jaypaulb/Projects/gh/CanvusPythonAPI/, /home/jaypaulb/Projects/gh/canvus-mcp-server/

Produce /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/conventions/python.md following the same outline as the Go conventions doc (module structure / version / build / linting / logging / error handling / config / testing / CI / code style / amendments), but with Python idioms:

- Workspace: uv workspace at `python/pyproject.toml`, member packages under `python/sdk`, `python/examples/core/*`, `python/tools/*`
- Python version: 3.11 minimum, CI on 3.11 and 3.12
- Build/install: `uv` (uv.lock at repo root)
- Format/lint: `ruff` (replaces black + isort + flake8)
- Type checking: `mypy --strict` for SDK; permissive for examples
- Logging: `structlog` configured with JSON renderer for prod, ConsoleRenderer for dev
- Error handling: custom exception hierarchy under `canvus_sdk.errors`, with `CanvusError` base
- Config: pydantic-settings with env prefix `CANVUS_`
- Testing: pytest + pytest-asyncio + pytest-cov, coverage target 70% for SDK
- HTTP client: httpx (sync + async)
- Async: prefer async-first SDK, expose sync wrappers
- CI: GitHub Actions `.github/workflows/python.yml` with matrix

Include the same "Convention amendments" section as the Go doc.

Do not modify files outside /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/conventions/python.md . Do not run git commands. Do not commit.

When complete, report: section count, total length in lines.
```

Agent 2.5 prompt:

```
You are a Plan agent authoring the TypeScript conventions doc for the MT-Canvus-Tools monorepo. The user is Jaypaul; you are working autonomously and will not be able to ask questions.

This doc locks the build/lint/log/error/test/CI defaults for all TypeScript code in the repo. Phase 4b refresh agents follow this doc without prompting Jaypaul. Pick standard, modern, idiomatic TypeScript practice as of 2026. Note: there is no existing TS SDK — this is greenfield.

Read for context:
- Spec: /home/jaypaulb/Projects/gh/docs/superpowers/specs/2026-05-17-mt-canvus-tools-consolidation-design.md
- API reference (target spec): /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/

Produce /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/conventions/typescript.md following the same outline as the Go conventions doc (module structure / version / build / linting / logging / error handling / config / testing / CI / code style / amendments), with these TypeScript idioms:

- Workspace: pnpm workspaces at `typescript/package.json`, members `typescript/sdk`, `typescript/examples/core/*`, `typescript/examples/webui`
- Node version: 20 LTS minimum, CI on 20 and 22
- Package manager: pnpm 9+
- Build: `tsc` with project references; output to `dist/`
- TS config: `typescript/tsconfig.base.json` with `strict: true`, `noUncheckedIndexedAccess: true`, `target: ES2022`, `module: ESNext`, `moduleResolution: Bundler`
- SDK is dual-published: ESM + CJS via tsup
- Linting: eslint v9 (flat config) + @typescript-eslint, prettier for formatting
- Logging: `pino` (JSON in prod, pino-pretty in dev)
- Error handling: custom error classes extending `Error`; `CanvusError` base class; structured `cause` chains
- Config: env vars via `zod` schema parsing; prefix `CANVUS_`
- HTTP client: `undici` (Node 20+ built-in fetch is fine for SDK; undici for streaming)
- Streaming (Canvus subscribe=true): use async iterators
- Testing: `vitest` + `@vitest/coverage-v8`, coverage target 70% for SDK
- CI: GitHub Actions `.github/workflows/typescript.yml` with matrix

Include the same "Convention amendments" section.

Do not modify files outside /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/conventions/typescript.md . Do not run git commands. Do not commit.

When complete, report: section count, total length in lines.
```

Send all five Agent tool calls in a single assistant message.

- [ ] **Step 2: Verify all five agents completed**

```bash
ls -la /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/_coverage-*.md
ls -la /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/conventions/
```

Expected: `_coverage-go.md`, `_coverage-python.md`, `go.md`, `python.md`, `typescript.md`.

### Task 2.2: Merge coverage matrices

**Files:**
- Create: `MT-Canvus-Tools/docs/api-reference/coverage-matrix.md`
- Delete: `_coverage-go.md`, `_coverage-python.md`

- [ ] **Step 1: Merge into a single coverage-matrix.md**

Write `coverage-matrix.md` with this structure:

```markdown
# SDK Coverage Matrix

Audited: <date>
Spec freeze: see [SOURCE.md](SOURCE.md)

## Summary

| Resource group | Total | Go: impl/miss/old/dep | Python: impl/miss/old/dep |
|---|---|---|---|

(populate from the two intermediate files)

## Per-endpoint detail

### Canvases

| Endpoint | Go status | Go method | Python status | Python method | Notes |
|---|---|---|---|---|---|

(merge row-by-row from intermediate files; one combined row per endpoint)

### Widgets

(... continue for all resource groups)

## Phase 3 work items

### Go SDK

(append Go agent's Phase 3 work items section)

### Python SDK

(append Python agent's Phase 3 work items section)

### TypeScript SDK

All endpoints from the spec must be implemented from scratch. Phase 3.3 agent will use this matrix's endpoint list as its work item.
```

- [ ] **Step 2: Delete intermediate files**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
rm docs/api-reference/_coverage-go.md docs/api-reference/_coverage-python.md
```

### Task 2.3: Commit Phase-2 output

- [ ] **Step 1: Stage and commit**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
git add docs/api-reference/coverage-matrix.md docs/conventions/
git status
```

Verify staged files: coverage-matrix.md, go.md, python.md, typescript.md. Verify intermediate `_coverage-*.md` files are NOT present.

- [ ] **Step 2: Commit**

```bash
git commit -m "docs: audit existing SDKs and lock language conventions

Phase 2 of the consolidation:
- Coverage matrix (Go + Python vs current spec) at docs/api-reference/coverage-matrix.md
- Convention docs locking build/lint/log/error/test/CI defaults per language
- These lock the autonomy contract for Phase 4b refresh agents"

git push origin main
```

---

## Phase 2 Completion Gate

- [ ] **Step 1: Verify Phase 0–2 deliverables are all present**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
find docs -type f -name "*.md" | sort
git log --oneline
```

Expected files (must all exist):

- `docs/superpowers/specs/2026-05-17-mt-canvus-tools-consolidation-design.md`
- `docs/superpowers/plans/2026-05-17-mt-canvus-tools-foundation.md`
- `docs/api-reference/README.md`
- `docs/api-reference/SOURCE.md`
- `docs/api-reference/authentication.md`
- `docs/api-reference/widget-types.md`
- `docs/api-reference/streaming.md`
- `docs/api-reference/changelog.md`
- `docs/api-reference/coverage-matrix.md`
- `docs/api-reference/endpoints/canvases.md` (and 6 more)
- `docs/api-reference/examples/canvases.md` (and 6 more)
- `docs/conventions/go.md`
- `docs/conventions/python.md`
- `docs/conventions/typescript.md`

Expected git log:
- `chore: bootstrap MT-Canvus-Tools monorepo`
- `docs(api-reference): extract canonical API spec from mt-restapi-client`
- `docs: audit existing SDKs and lock language conventions`

- [ ] **Step 2: Verify GH push succeeded**

```bash
gh repo view jaypaulb/MT-Canvus-Tools --json url,visibility,pushedAt
```

- [ ] **Step 3: Surface completion to Jaypaul**

Report:
- Files written (count by category)
- Endpoint coverage percentage (from coverage-matrix.md)
- Convention doc decisions made
- Any "Source uncertainty" callouts from Phase 1 that need Jaypaul's review
- Confirmation that the foundation is ready for Phase 3 dispatch

---

## What comes next (out of scope for this plan)

After this plan completes, the orchestrator generates the **next plan**: Phase 3 SDK Migration. That plan will be written using the matured `docs/api-reference/` and `docs/conventions/` artefacts produced here, and will dispatch three parallel code-writing agents (one per language) in worktree isolation.

Subsequent plans:
- Plan 3: Phase 4a Core Examples (3 parallel agents)
- Plan 4: Phase 4b Per-Item Refresh (11 items, batched 4–6)
- Plan 5: Phase 5–7 Wrap-up (custom menu, top-level docs, CI)
- Plan 6: Phase 8 Public Release (foreground)

Each subsequent plan is written when its predecessor completes, so it can reference the actual artefacts the predecessor produced rather than speculating.
