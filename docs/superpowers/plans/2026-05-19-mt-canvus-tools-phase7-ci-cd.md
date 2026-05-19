# MT-Canvus-Tools Phase 7 — CI/CD Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add GitHub Actions CI for all three language workspaces (Go, Python, TypeScript), remove one pre-existing Python mypy unused-ignore, and update CONSOLIDATION-STATUS.md.

**Architecture:** Three independent per-language workflow files with path-scoped triggers (Go changes don't run Python CI, etc.). The Go workflow iterates over workspace modules via `awk` on `go.work` because `go build ./...` from the workspace root fails (no root go.mod). Python uses uv; TypeScript uses pnpm. Each workflow runs lint → build → vet → test in a single job.

**Tech Stack:** GitHub Actions (`actions/checkout@v4`, `actions/setup-go@v5`, `astral-sh/setup-uv@v4`, `pnpm/action-setup@v4`, `actions/setup-node@v4`), Go 1.24, Python 3.11 (via uv), Node 20 (via pnpm).

**Branch:** Create `feature/phase7-ci-cd` from `main` before starting.

---

## File Map

| Action | Path |
|---|---|
| Create | `.github/workflows/ci-go.yml` |
| Create | `.github/workflows/ci-python.yml` |
| Create | `.github/workflows/ci-typescript.yml` |
| Modify | `python/sdk/src/canvus_sdk/client.py:105` (remove unused `type: ignore`) |
| Modify | `CONSOLIDATION-STATUS.md` |

---

### Task 1: Create Go CI workflow

**Files:**
- Create: `.github/workflows/ci-go.yml`

**Context:**

- The workspace root is `go/` with `go.work`. Running `go build ./...` from `go/` fails with "pattern ./...: directory prefix . does not contain modules listed in go.work" — there is no `go.mod` at the workspace root. Each of the 17 workspace modules must be addressed individually.
- Powertoys (`go/tools/powertoys/`) uses Fyne and getlantern/systray which require CGO and X11/GL headers. Without them, `go build` fails at the C compilation step. Install them with apt-get before the Go setup step.
- Live tests use `//go:build live` tag. Omitting `-tags live` from `go test` is sufficient to skip them — no extra flag needed.
- `go.work.sum` is at `go/go.work.sum` (gitignored per `.gitignore` line 16). The Go module cache is shared; use `setup-go@v5` with `cache-dependency-path: go/go.work.sum` for dependency caching.

- [ ] **Step 1: Create `.github/workflows/` directory and the workflow file**

```bash
mkdir -p .github/workflows
```

Create `.github/workflows/ci-go.yml` with this exact content:

```yaml
name: CI — Go

on:
  push:
    branches: [main]
    paths:
      - 'go/**'
      - '.github/workflows/ci-go.yml'
  pull_request:
    paths:
      - 'go/**'
      - '.github/workflows/ci-go.yml'

jobs:
  go:
    name: Format / Build / Vet / Test
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: go

    steps:
      - uses: actions/checkout@v4

      - name: Install system dependencies (Fyne + systray)
        run: sudo apt-get update -qq && sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev libgtk-3-dev

      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache-dependency-path: go/go.work.sum

      - name: Format check
        run: |
          bad=$(gofmt -s -l .)
          if [ -n "$bad" ]; then
            echo "Files with formatting issues:"
            echo "$bad"
            exit 1
          fi

      - name: Build + Vet all workspace modules
        run: |
          set -euo pipefail
          awk '/^\t\./{print substr($0,2)}' go.work | while IFS= read -r dir; do
            echo "--- $dir ---"
            (cd "$dir" && go build ./... && go vet ./...)
          done

      - name: Test all workspace modules
        run: |
          set -euo pipefail
          awk '/^\t\./{print substr($0,2)}' go.work | while IFS= read -r dir; do
            echo "--- $dir ---"
            (cd "$dir" && go test -count=1 -short ./...)
          done
```

- [ ] **Step 2: Verify YAML syntax**

```bash
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci-go.yml')); print('YAML OK')"
```

Expected: `YAML OK`

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci-go.yml
git commit -m "ci: add Go CI workflow (format, build, vet, test all workspace modules)"
```

---

### Task 2: Create Python CI workflow

**Files:**
- Create: `.github/workflows/ci-python.yml`

**Context:**

- Python workspace root: `python/` with `pyproject.toml` containing `[tool.uv.workspace]`. Run all commands from this directory.
- `uv sync --all-extras` syncs all workspace members and their optional dev dependencies (mypy, pytest, ruff, etc.).
- `ruff check .` from `python/` respects the `[tool.ruff]` config in `python/pyproject.toml`.
- **mypy SDK:** run from `python/sdk/` (step `working-directory: python/sdk`) so it picks up `sdk/pyproject.toml` which has `[tool.mypy]` with `warn_unused_ignores = true`. Running from `python/` skips that check.
- **mypy MCP server:** run from `python/tools/mcp-server/` (step `working-directory: python/tools/mcp-server`) to pick up the pydantic plugin config in `tools/mcp-server/pyproject.toml`.
- `uv run` from any subdirectory of the workspace finds the workspace venv by walking up the directory tree.
- Pytest `testpaths = ["sdk/tests"]` in root pyproject.toml only covers the SDK. Pass `sdk/tests tools/mcp-server/tests` explicitly to cover both. The marker `-m "not live and not integration"` skips tests requiring a live Canvus server.

- [ ] **Step 1: Create `.github/workflows/ci-python.yml`**

```yaml
name: CI — Python

on:
  push:
    branches: [main]
    paths:
      - 'python/**'
      - '.github/workflows/ci-python.yml'
  pull_request:
    paths:
      - 'python/**'
      - '.github/workflows/ci-python.yml'

jobs:
  python:
    name: Lint / Type-check / Test
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: python

    steps:
      - uses: actions/checkout@v4

      - uses: astral-sh/setup-uv@v4
        with:
          version: "latest"

      - name: Sync workspace (all extras)
        run: uv sync --all-extras

      - name: Ruff check
        run: uv run ruff check .

      - name: Mypy — SDK (strict, with warn_unused_ignores)
        working-directory: python/sdk
        run: uv run mypy --strict src

      - name: Mypy — MCP server (strict, with pydantic plugin)
        working-directory: python/tools/mcp-server
        run: uv run mypy --strict src

      - name: Pytest (excluding live/integration)
        run: uv run pytest sdk/tests tools/mcp-server/tests -m "not live and not integration"
```

- [ ] **Step 2: Verify YAML syntax**

```bash
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci-python.yml')); print('YAML OK')"
```

Expected: `YAML OK`

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci-python.yml
git commit -m "ci: add Python CI workflow (ruff, mypy strict, pytest)"
```

---

### Task 3: Create TypeScript CI workflow

**Files:**
- Create: `.github/workflows/ci-typescript.yml`

**Context:**

- TypeScript workspace root: `typescript/` with `pnpm-workspace.yaml` listing `sdk`, `examples/core/*`, `examples/webui`.
- Package manager is `pnpm@9.12.0` (declared in `package.json` `packageManager` field). Use `pnpm/action-setup@v4` to install it; `actions/setup-node@v4` with `cache: 'pnpm'` for caching.
- The `cache-dependency-path` in `setup-node` is relative to the repo root, not the working directory: use `typescript/pnpm-lock.yaml`.
- **Build order matters:** Examples reference `@mt-canvus-tools/sdk` and need `dist/` to exist for typecheck and test. Build the SDK with `pnpm --filter @mt-canvus-tools/sdk build` before running workspace-wide commands.
- `pnpm -r typecheck` and `pnpm -r lint` and `pnpm -r test` run recursively across all workspace packages. pnpm respects topological dependency order.
- The SDK is the only package with a coverage-gated test run (`vitest --coverage` with thresholds at 65/70/60/65). The webui also has vitest tests. Core examples have no tests (no test script output = success for `pnpm -r test`).
- `--frozen-lockfile` on install ensures the lockfile is not modified by CI.

- [ ] **Step 1: Create `.github/workflows/ci-typescript.yml`**

```yaml
name: CI — TypeScript

on:
  push:
    branches: [main]
    paths:
      - 'typescript/**'
      - '.github/workflows/ci-typescript.yml'
  pull_request:
    paths:
      - 'typescript/**'
      - '.github/workflows/ci-typescript.yml'

jobs:
  typescript:
    name: Build / Typecheck / Lint / Test
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: typescript

    steps:
      - uses: actions/checkout@v4

      - uses: pnpm/action-setup@v4
        with:
          version: 9

      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'pnpm'
          cache-dependency-path: typescript/pnpm-lock.yaml

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Build SDK
        run: pnpm --filter @mt-canvus-tools/sdk build

      - name: Typecheck all packages
        run: pnpm -r typecheck

      - name: Lint all packages
        run: pnpm -r lint

      - name: Test all packages
        run: pnpm -r test
```

- [ ] **Step 2: Verify YAML syntax**

```bash
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci-typescript.yml')); print('YAML OK')"
```

Expected: `YAML OK`

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci-typescript.yml
git commit -m "ci: add TypeScript CI workflow (build SDK, typecheck, lint, test)"
```

---

### Task 4: Remove Python unused `type: ignore`

**Files:**
- Modify: `python/sdk/src/canvus_sdk/client.py:105`

**Context:**

`Settings()` is a `pydantic_settings.BaseSettings` subclass. `api_url` and `api_key` are declared as required fields with no default (`Field(...)`). At the time the ignore was added, mypy didn't understand that `pydantic_settings` populates fields from environment variables, making `Settings()` valid at runtime. The `# type: ignore[call-arg]` suppressed the false "Missing positional arguments" error.

As of the current pydantic-settings version, mypy no longer raises `call-arg` for `BaseSettings` subclasses — the library ships inline type stubs that teach mypy about the env-variable injection. Running `uv run mypy --strict src` from `python/sdk/` now reports `Unused "type: ignore" comment [unused-ignore]`. The ignore needs to be removed.

- [ ] **Step 1: Verify the ignore is unused (confirm the error)**

```bash
cd python/sdk && uv run mypy --strict src 2>&1 | grep -E "unused|error|Success"
```

Expected output contains:
```
src/canvus_sdk/client.py:105: error: Unused "type: ignore" comment  [unused-ignore]
Found 1 error in 1 file (checked 33 source files)
```

If output is `Success: no issues found`, the ignore is still needed — skip to Step 4.

- [ ] **Step 2: Remove the `type: ignore` comment**

File: `python/sdk/src/canvus_sdk/client.py`, line 105.

Before:
```python
        cfg = settings or Settings()  # type: ignore[call-arg]
```

After:
```python
        cfg = settings or Settings()
```

- [ ] **Step 3: Verify mypy clean from both run locations**

```bash
# From sdk directory (strict config with warn_unused_ignores)
cd python/sdk && uv run mypy --strict src 2>&1 | tail -2
```

Expected: `Success: no issues found in 33 source files`

```bash
# From workspace root (as CI will run it)
cd python && uv run mypy --strict sdk/src 2>&1 | tail -2
```

Expected: `Success: no issues found in 33 source files`

- [ ] **Step 4: Commit**

```bash
git add python/sdk/src/canvus_sdk/client.py
git commit -m "fix(python): remove unused type: ignore[call-arg] on Settings() — pydantic-settings stubs now teach mypy about env-variable injection"
```

---

### Task 5: Update CONSOLIDATION-STATUS.md

**Files:**
- Modify: `CONSOLIDATION-STATUS.md`

**Context:**

Update the status header to reflect Phase 7 complete, add Phase 7 to the shipped table, and clear the "Deferred to Phase 7" section.

- [ ] **Step 1: Update the header block**

Find:
```markdown
**As of:** 2026-05-19 (post-Phase 6)
**Phases complete:** 0, 1, 2, 3 (+ verification), 4a (+ review-driven fixes), 4b (+ review-driven fixes), 4c (+ review-driven fixes), 4d (+ review-driven fixes), 5 (PowerToys port + carry-over), 6 (top-level documentation)
**Next phase:** 7 (CI/CD)
```

Replace with:
```markdown
**As of:** 2026-05-19 (post-Phase 7)
**Phases complete:** 0, 1, 2, 3 (+ verification), 4a (+ review-driven fixes), 4b (+ review-driven fixes), 4c (+ review-driven fixes), 4d (+ review-driven fixes), 5 (PowerToys port + carry-over), 6 (top-level documentation), 7 (CI/CD)
**Next phase:** 8 (TBD)
```

- [ ] **Step 2: Add Phase 7 rows to the shipped table**

After the `| 6 | ...` row, add Phase 7 rows. Use the actual commit hashes from Task 1–4 once they are committed. Placeholder format (fill in real hashes at commit time):

```markdown
| 7.1 | GitHub Actions: Go CI workflow (format / build / vet / test all workspace modules) | `<hash-from-task-1>` |
| 7.2 | GitHub Actions: Python CI workflow (ruff / mypy strict / pytest) | `<hash-from-task-2>` |
| 7.3 | GitHub Actions: TypeScript CI workflow (build SDK / typecheck / lint / test) | `<hash-from-task-3>` |
| 7.fix | Python: remove unused type: ignore[call-arg] on Settings() | `<hash-from-task-4>` |
```

- [ ] **Step 3: Update the "Deferred to Phase 7" section**

Find the section:
```markdown
## Deferred to Phase 7 (CI/CD)

Single agent: GitHub Actions workflows. Per-language matrix jobs (lint, test, build). Optional spec-drift detector comparing `docs/api-reference/` to `mt-restapi-client` on a schedule.

All Phase 4d and Phase 5 carry-over items are now resolved:

- ✅ **Python `FoldersResource.subscribe_permissions`** — shipped Phase 5 carry-over (`cd49f54`)
- ✅ **TS lint floor** — `UserId`/`GroupId` narrowed to `number` + `allowNumber: true` rule (`cdde5ab`); floor is now 0
- **Python `client.py:105` unused-ignore** — `# type: ignore[call-arg]` on `Settings()` — verify still needed with `uv run mypy --strict sdk/src`; drop in Phase 7 sweep if resolved
- **CloneWidget per-type wrappers** — retained as "not needed unless callers report friction" design decision; revisit only on user request
```

Replace with:
```markdown
## Phase 7 outcome — DONE

GitHub Actions CI for all three language workspaces.

| Workflow | Jobs |
|---|---|
| `ci-go.yml` | format check (`gofmt -s -l`), build + vet, test (all 17 workspace modules) |
| `ci-python.yml` | ruff check, mypy --strict SDK, mypy --strict MCP server, pytest (not live/integration) |
| `ci-typescript.yml` | pnpm install, build SDK, typecheck, lint, test (all workspace packages) |

All Phase 4d and Phase 5 carry-over items resolved:

- ✅ **Python `FoldersResource.subscribe_permissions`** — shipped Phase 5 carry-over (`cd49f54`)
- ✅ **TS lint floor** — `UserId`/`GroupId` narrowed to `number` + `allowNumber: true` rule (`cdde5ab`); floor is now 0
- ✅ **Python `client.py:105` unused-ignore** — `type: ignore[call-arg]` removed; pydantic-settings stubs now teach mypy about env-variable injection (Phase 7 `<hash-from-task-4>`)
- **CloneWidget per-type wrappers** — retained as "not needed unless callers report friction" design decision; revisit only on user request

**Spec-drift detector** (optional, deferred): A scheduled workflow comparing `docs/api-reference/` against `mt-restapi-client` in `gitlab.multitaction.com` requires a GitLab PAT secret and a custom diff script. Deferred to a future phase when CI is more mature.
```

- [ ] **Step 4: Verify the file is well-formed Markdown**

```bash
python3 -c "
with open('CONSOLIDATION-STATUS.md') as f:
    content = f.read()
# Check key markers are present
assert '7 (CI/CD)' in content, 'Phase 7 missing from phases complete'
assert '7.1' in content, 'Phase 7 shipped row missing'
assert 'Phase 7 outcome' in content, 'Phase 7 outcome section missing'
print('CONSOLIDATION-STATUS.md OK')
"
```

Expected: `CONSOLIDATION-STATUS.md OK`

- [ ] **Step 5: Commit**

```bash
git add CONSOLIDATION-STATUS.md
git commit -m "docs: update CONSOLIDATION-STATUS for Phase 7 (CI/CD complete)"
```

---

## Self-Review Notes

**Spec coverage:** All 5 scope items addressed — 3 CI workflows, 1 Python cleanup, 1 status update.

**Go module iteration:** `awk '/^\t\./{print substr($0,2)}' go.work` extracts module paths from `go.work` `use (...)` block. Lines inside `use (...)` are tab-indented (`\t./cli`, etc.). The `awk` strips the leading tab; the `while` loop then `cd`s into each directory. This is more reliable than `grep $'^\t\.'` (ANSI-C quoting has subtle shell differences across CI environments).

**Python mypy run locations:** CI mypy steps use step-level `working-directory: python/sdk` and `python/tools/mcp-server` (overriding job-level `working-directory: python`). This picks up the per-package `[tool.mypy]` configs (including `warn_unused_ignores` for SDK and pydantic plugin for MCP server).

**TypeScript build order:** SDK must be built before typecheck/lint/test of examples that import it. `pnpm --filter @mt-canvus-tools/sdk build` runs first; subsequent `pnpm -r` commands run in topological order.

**Spec-drift detector:** Explicitly deferred — requires GitLab PAT for private repo access. Documented in Phase 7 outcome section.
