# Phase 4d — Cleanup Sweep + MCP LLM Tools Port Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close out all 17 deferred items accumulated through Phases 3.t → 4c, leaving the monorepo at a clean v0.1-tag-ready state. The work is a mix of one-line tweaks, SDK option additions, baseline reductions, doc back-applications, live verification, and one major port (mcp-server-python's 13 LLM/correlation/brainstorming/reports tools + Ollama→httpx client rewrite + SQLite cache + PDF processing pipelines).

**Architecture:** Items batched by independence and risk. Four rounds, mostly parallel within a round. SDK additions (`WithVerifyTLS`, `users.current()`, `WithSubscribeBuffer`, `createAnyWithAsset`) land before their consumer refactors. Cross-cutting refactors (shared Gemini helper) wait for their dependency items. Live verification and MCP LLM port run in the final round in parallel — they touch disjoint surfaces.

**Tech Stack:** No new tech. All work happens in the existing Go workspace (`go.work`), Python uv workspace (`python/pyproject.toml`), TypeScript pnpm workspace (`typescript/pnpm-workspace.yaml`). MCP LLM port specifically replaces the legacy `aiohttp` Ollama client with the monorepo-standard `httpx` async client + `structlog` + `pydantic` settings (no module-scoped globals).

**Scope decisions (locked in conversation 2026-05-18):**

- **MCP LLM tools port: IN.** Phase 4d covers the full 13-tool re-implementation. Risks acknowledged: this round alone is multiple days of agent work. Mitigated by isolating it as Round D's sole opus implementer.
- **Live-verify permissions subscribe: IN.** Run `SubscribeCanvasPermissions` + `SubscribeFolderPermissions` against `dev-mtcs.multitaction.com`. On pass: drop the "not yet live-verified" warnings from Go SDK + Python SDK. On fail: escalate to documented gaps in parity-matrix §5.5.
- **CloneWidget per-type wrappers: DROPPED.** No friction reported. Honour parity-matrix §5.2 recommendation. Removed from 4d list.

---

## Item inventory (17 items, all from CONSOLIDATION-STATUS.md "Deferred to Phase 4d")

| # | Item | Source | Files touched | Size |
|---|---|---|---|---|
| 1 | Go SDK dead code removal | Tier 3 #10 | `go/sdk/canvus/session.go` (drop `setToken`), `go/sdk/canvus/warnings.go` (drop `warnAlways`) | tiny |
| 2 | TS SDK README inaccuracies sweep | 4a fallout | 6 TS example READMEs with "Expected output" referencing wire-shape fields that don't exist on server (cascade from C1) | tiny |
| 3 | Spec doc errors back-apply | VERIFIED-CORRECTIONS | `docs/api-reference/endpoints/server.md`, `docs/api-reference/endpoints/auth.md` | tiny |
| 4 | Reclaim TS coverage thresholds | 4b review | `typescript/sdk/vitest.config.ts` (50/60/40/50 → 60/70/60/60) + new focused tests for `auth.ts`, `users.ts`, `server.ts` resource error mapping | medium |
| 5 | Live-verify permissions subscribe | 4b parity §5.5 | Go `go/sdk/canvus/permissions.go` warnings + Python `python/sdk/src/canvus_sdk/resources/permissions.py` docstring + parity-matrix update | small (run + remove warnings) |
| 6 | Asset roundtrip integration tests | 4b review | `typescript/sdk/tests/export-import-roundtrip.test.ts`, `python/sdk/tests/test_export_import_roundtrip.py` | medium |
| 7 | Python mypy strict baseline reduction | 4b review | Several `python/sdk/src/canvus_sdk/resources/*.py` — `list[T] annotation vs .list() method` pattern (38 → ≤30) | small |
| 8 | TS lint baseline reduction | 4b review | `typescript/sdk/src/**/*.ts` — `UserId` template-literal interpolation, `request<void>` returns, `ReadonlyArray<T>` syntax, tsconfig-include for tests (64 → ≤30) | medium |
| 9 | `WithSubscribeBuffer(int)` option | 4b review | Go `go/sdk/canvus/options.go` + Python `python/sdk/src/canvus_sdk/client.py` + TS `typescript/sdk/src/session.ts` (currently hardcoded buffer=4) | small |
| 10 | TS `createAnyWithAsset` | 4b review | `typescript/sdk/src/resources/widgets.ts` (or document the asymmetry in `docs/conventions/typescript.md`) | small |
| 11 | note-mapper Gemini dep migration | 4c review | `go/examples/projects/note-mapper/go.mod`, `go/examples/projects/note-mapper/internal/llm/extract.go`, `go/examples/projects/note-mapper/cmd/main.go` — migrate `github.com/google/generative-ai-go v0.19.0` → `google.golang.org/genai v1.34.0` | small |
| 12 | Extract shared Gemini helper | 4c review (rule-of-three) | New `go/examples/internal/llm/gemini.go` (or `go/sdk/extras/llm/gemini.go` if promoted SDK-side); consumers: translator, ai-personas, note-mapper | medium |
| 13 | Go SDK `WithVerifyTLS(bool)` option | 4c review | `go/sdk/canvus/options.go` + remove 4 callsites: `go/tools/db-solver/internal/commands/session.go:15`, `go/tools/db-solver/internal/commands/lookup_hash.go:242`, `go/cli/internal/session/session.go:42`, `go/cli/internal/commands/login.go:111` | small-medium |
| 14 | ai-personas Subscribe migration | 4c review | `go/examples/projects/ai-personas/internal/qa/wait.go` — replace 500ms `GetNote` poll with `SubscribeWidget` | small |
| 15 | mcp-server-python LLM tools port | 4c deferral | `python/tools/mcp-server/src/canvus_mcp_server/mcp_tools/{llm,brainstorming,correlation,reports}.py` (13 tools) + new `python/tools/mcp-server/src/canvus_mcp_server/llm/ollama.py` (httpx port of 541-LOC client) + SQLite cache + PDF processing pipelines | **LARGE** |
| 16 | mcp-server-python `/users/current` escape hatch | 4c review | Python SDK `python/sdk/src/canvus_sdk/resources/users.py` add typed `current()` + mcp-server consumer `python/tools/mcp-server/src/canvus_mcp_server/mcp_tools/users.py:116` swap | tiny |
| 17 | Document cli `FromEnv` deviation | 4c review | `docs/conventions/go.md §11` — dated Amendment explaining viper precedence in `go/cli/internal/config/config.go` | tiny |

**Round allocation:**

- **Round A — Tiny independent cleanups, 5 parallel sonnet agents (items 1, 11, 14, 16, 17):**
  - A1: Go SDK dead code (item 1)
  - A2: note-mapper Gemini migration (item 11)
  - A3: ai-personas Subscribe migration (item 14)
  - A4: Python SDK `users.current()` + mcp-server consumer (item 16)
  - A5: cli FromEnv docs Amendment (item 17)

- **Round B — SDK additions + small refactors, 4 parallel sonnet agents (items 9, 10, 13 + items 2, 3):**
  - B1: Go SDK `WithVerifyTLS(bool)` + 4 callsite refactors (item 13)
  - B2: `WithSubscribeBuffer(int)` across Go/Python/TS (item 9)
  - B3: TS `createAnyWithAsset` decision (item 10)
  - B4: Docs sweeps — TS SDK READMEs (item 2) + spec doc back-apply (item 3) merged into one agent (both are pure doc edits)

- **Round C — Refactor + coverage + baseline, 5 parallel mixed sonnet/opus (items 4, 6, 7, 8, 12):**
  - C1: Extract shared Gemini helper (item 12, opus — depends on Round A item 11 landing)
  - C2: Python mypy baseline reduction (item 7, sonnet)
  - C3: TS lint baseline reduction (item 8, sonnet)
  - C4: Reclaim TS coverage thresholds (item 4, opus)
  - C5: Asset roundtrip integration tests TS + Python (item 6, opus)

- **Round D — Live verification + MCP LLM port, 2 parallel opus (items 5, 15):**
  - D1: Live-verify permissions subscribe (item 5, opus, env access)
  - D2: MCP LLM tools port (item 15, opus, LARGE)

Total: 4 rounds, 5+4+5+2 = 16 implementer agents (item 4d.B4 bundles items 2+3, item 4d.A4 bundles SDK + consumer, item 4d.B1 bundles SDK + 4 consumers, item 4d.B2 spans 3 SDKs).

Per-round review gate after each round (opus, single agent). Final review gate after Round D.

---

## Task 4d.0: Pre-flight

- [ ] **Step 1: Verify HEAD is post-Phase-4c**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
git status
git log --oneline | head -5
```

Expected: HEAD at `1820a01` ("docs: Phase 4c completion update + 4d tracker refresh + Phase 5 carry-forward"), working tree clean.

- [ ] **Step 2: Toolchain baseline snapshot**

Record current numbers so Round C agents have a clear delta target:

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/python
uv run mypy --strict src/ 2>&1 | tail -5

cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools/typescript
pnpm --filter @mt-canvus-tools/sdk lint 2>&1 | tail -5
pnpm --filter @mt-canvus-tools/sdk test:coverage 2>&1 | tail -20
```

Expected baselines (from CONSOLIDATION-STATUS): Python mypy strict = 38 errors; TS lint = 64 errors; TS coverage thresholds = 50/60/40/50 (lines/statements/functions/branches).

- [ ] **Step 3: Confirm `.secrets` file still present for live verification (Round D1)**

```bash
ls -la /home/jaypaulb/Projects/gh/MT-Canvus-Tools/.secrets
```

Expected: file exists, gitignored (not tracked). Contains `CANVUS_API_URL=https://dev-mtcs.multitaction.com/api/v1/` and `CANVUS_API_KEY=...`. If missing, STOP and ask Jaypaul before proceeding to Round D.

---

## Task 4d.1: Round A — Tiny independent cleanups (5 parallel sonnet implementers)

All five items are independent: different files, no shared state, no dependency chain. Dispatch all five in parallel.

### Task 4d.A1: Go SDK dead code removal

**Files:**
- Modify: `go/sdk/canvus/session.go` (remove `setToken` method, lines ~212-222)
- Modify: `go/sdk/canvus/warnings.go` (remove `warnAlways` function, lines ~114-125)

- [ ] **Implementer prompt template:**

```
You are removing two pieces of dead code from the Go SDK in MT-Canvus-Tools.

Verified dead (grep showed zero call sites outside the definitions themselves):
- `tokenManager.setToken` at go/sdk/canvus/session.go:212 — defined but never called. Token mutation goes through `clearToken` + `refreshToken` only.
- `warnAlways` at go/sdk/canvus/warnings.go:114 — defined but never called. All call sites use `warnOnce`.

Tasks:
1. Read both files in full first to understand the surrounding context.
2. Re-verify no callers: `cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools && grep -rn "setToken\|warnAlways" go/ --include="*.go" | grep -v "^.*\.go:[0-9]*:.*//"`. Expect only the definitions to remain.
3. Delete `setToken` method (the entire function block, including its doc comment).
4. Delete `warnAlways` function (entire function block, including its doc comment).
5. Run `cd go/sdk/canvus && go build ./... && go vet ./... && go test ./...`. Expect all clean.
6. Run `gofmt -s -l .` from `go/sdk/canvus/`. Expect empty output.
7. Stage exactly the two files: `git add go/sdk/canvus/session.go go/sdk/canvus/warnings.go`.
8. Commit: `chore(sdk/go): remove dead setToken and warnAlways (Phase 4d Round A)`.

Reference: Tier 3 #10 in CONSOLIDATION-STATUS.md. Phase 4b Chesterton-fenced `doRequestWithHeaders` already (it IS the subscribe primitive backbone — DO NOT touch it).

Report DONE with the commit SHA, or BLOCKED with reason.
```

### Task 4d.A2: note-mapper Gemini dep migration

**Files:**
- Modify: `go/examples/projects/note-mapper/go.mod`
- Modify: `go/examples/projects/note-mapper/internal/llm/extract.go`
- Modify: `go/examples/projects/note-mapper/cmd/main.go` (if it imports the Gemini API directly)

- [ ] **Implementer prompt template:**

```
You are migrating note-mapper from the deprecated `github.com/google/generative-ai-go v0.19.0` to `google.golang.org/genai v1.34.0` so it matches translator and ai-personas (which are already on the new lib).

Context: rule-of-three triggered. translator/ai-personas already use the new lib. Phase 4c final review flagged note-mapper as the odd one out.

Tasks:
1. Read go/examples/projects/note-mapper/go.mod to confirm current state.
2. Read go/examples/projects/note-mapper/internal/llm/extract.go and cmd/main.go for current usage patterns.
3. Read go/tools/translator/internal/gemini.go AND go/examples/projects/ai-personas/internal/ai/gemini.go for the target API shape. (Do NOT extract a shared helper — that is Round C's job. Just match their API usage style.)
4. Update go.mod: drop `github.com/google/generative-ai-go` + its `cloud.google.com/go/ai` indirect dependencies that are no longer needed. Add `google.golang.org/genai v1.34.0`. Run `go mod tidy` from inside note-mapper directory.
5. Rewrite the Gemini call sites in extract.go (and cmd/main.go if applicable) using the new lib's API. Preserve the function signatures and behaviour exactly — this is a library swap, not a feature change.
6. Verify: `cd go/examples/projects/note-mapper && go build ./... && go vet ./... && go test ./...`. Expect all clean.
7. Run `gofmt -s -l .` from `go/examples/projects/note-mapper/`. Expect empty.
8. Stage the modified files. Commit: `chore(examples/note-mapper): migrate to google.golang.org/genai (Phase 4d Round A)`.

Watch out: the new lib's auth pattern is slightly different (uses `genai.NewClient(ctx, &genai.ClientConfig{APIKey: key})` instead of the old `genai.NewClient(ctx, option.WithAPIKey(key))`). Check translator's gemini.go for the exact pattern.

Report DONE with commit SHA, or BLOCKED with reason.
```

### Task 4d.A3: ai-personas Subscribe migration

**Files:**
- Modify: `go/examples/projects/ai-personas/internal/qa/wait.go`

- [ ] **Implementer prompt template:**

```
You are replacing a 500ms GetNote poll loop in ai-personas with the Phase 4b SubscribeWidget helper.

Background: ai-personas' qa/wait.go currently polls `GetNote(...)` every 500ms with a 10s error tolerance. Phase 4b shipped `SubscribeWidget` (`go/sdk/canvus/widgets.go`, parity-matrix §4.1 #7) which emits NDJSON events for the same widget. Switching cuts request volume, latency, and the documented "deferral" warning.

Tasks:
1. Read go/examples/projects/ai-personas/internal/qa/wait.go in full.
2. Read go/sdk/canvus/widgets.go's `SubscribeWidget` method signature + usage example (find a Go example/SDK call site to model after, e.g., `go/examples/core/05-event-subscription/main.go` or similar).
3. Rewrite the wait function:
   - Use SubscribeWidget with the existing context.
   - Apply the snapshot-drain dedup pattern (Phase 4b: settle-timer of 2s, env-overridable via the parent process env) — match Round 2 / 3 implementer patterns used in `go/examples/core/`.
   - Preserve the function's external behaviour: same signature, same return semantics. Internal polling structure replaced.
   - Keep error tolerance similar (10s deadline before declaring failure) but rebuild it around stream cancellation, not retry counting.
4. Re-run the package's tests. If wait.go has tests, ensure they still pass — they may need to be rewritten to drive the stream instead of mocking GetNote.
5. `go build ./... && go vet ./... && go test ./...` from inside go/examples/projects/ai-personas/. Clean.
6. `gofmt -s -l .` — empty.
7. Stage. Commit: `refactor(examples/ai-personas): wait.go uses SubscribeWidget instead of poll (Phase 4d Round A)`.

If you find wait.go's polling behaviour serves a purpose not visible from the code alone (e.g., test infra reasons), surface as BLOCKED rather than half-implementing.

Report DONE with commit SHA, or BLOCKED with reason.
```

### Task 4d.A4: Python SDK `users.current()` + mcp-server consumer

**Files:**
- Modify: `python/sdk/src/canvus_sdk/resources/users.py` (add `current()` method)
- Modify: `python/sdk/tests/test_users.py` (add test for `current()` — mock the transport like other tests)
- Modify: `python/tools/mcp-server/src/canvus_mcp_server/mcp_tools/users.py` (line ~116 — swap escape hatch for typed call)

- [ ] **Implementer prompt template:**

```
You are closing the "Python SDK escape hatch" that mcp-server uses to call `GET /users/current`.

Current state:
- python/tools/mcp-server/src/canvus_mcp_server/mcp_tools/users.py:116 reaches into private API: `client._transport.request("GET", "users/current")`.
- python/sdk/src/canvus_sdk/resources/users.py has no `current()` method despite the Go SDK exposing `GetCurrentUser` and canvases.py at python/sdk/src/canvus_sdk/resources/canvases.py:92 having an analogous one-line typed method.

Tasks:
1. Read python/sdk/src/canvus_sdk/resources/canvases.py:80-100 to see the one-line typed-method pattern used elsewhere.
2. Read python/sdk/src/canvus_sdk/resources/users.py in full.
3. Add an `async def current(self) -> User:` method on the Users resource class. Mirror the canvases.py pattern: `data = await self._transport.request("GET", "users/current"); return User.model_validate(data)`.
4. Update python/sdk/src/canvus_sdk/__init__.py if needed (resource methods are auto-exposed via the client; no extra wiring usually required — check).
5. Add a focused unit test in python/sdk/tests/test_users.py modelled on the existing tests in that file. Mock the transport, return a valid User dict, assert `client.users.current()` returns a User instance.
6. Update python/tools/mcp-server/src/canvus_mcp_server/mcp_tools/users.py:116 to call `client.users.current()` directly. Remove the lambda + private `_transport` access.
7. Add a "Phase 4d cleanup" comment NEAR the swapped line ONLY IF the previous comment referenced the gap; otherwise leave clean (per CLAUDE.md "no churn comments").
8. Verify:
   - `cd python && uv run pytest sdk/tests/test_users.py -v` (all pass, new test included)
   - `cd python && uv run mypy --strict sdk/src/canvus_sdk/resources/users.py` (no new errors)
   - `cd python && uv run ruff check sdk/ tools/mcp-server/` (clean)
9. Stage all 3 files + any __init__ change. Commit: `feat(sdk/python): add users.current() and migrate mcp-server off escape hatch (Phase 4d Round A)`.

Report DONE with commit SHA, or BLOCKED.
```

### Task 4d.A5: cli FromEnv deviation docs Amendment

**Files:**
- Modify: `docs/conventions/go.md` (add Amendment under §11 "Configuration loading")

- [ ] **Implementer prompt template:**

```
You are adding a dated Amendment to docs/conventions/go.md §11 documenting that go/cli intentionally uses viper precedence (flag > env > file > defaults) instead of the SDK's FromEnv constructor.

Context:
- docs/conventions/go.md §11 mandates `FromEnv` for env-driven config.
- go/cli/internal/config/config.go uses viper directly because the CLI needs flag/file precedence that FromEnv doesn't provide.
- Phase 4c final review flagged the deviation.

Tasks:
1. Read docs/conventions/go.md §11 in full to see the existing Amendment style (other sections of the convention already have dated Amendments — find one and match its format).
2. Read go/cli/internal/config/config.go to confirm the precedence chain and the explicit `viper.BindEnv("url", "CANVUS_API_URL", "CANVUS_URL")` call inside Load().
3. Add an Amendment under §11. Format:
   - Date stamp: `**Amendment 2026-05-18 (Phase 4d):**`
   - Subject: `CLI tools may use viper precedence instead of FromEnv`
   - Rationale: 2-3 sentences explaining that CLI binaries need flag-overrides-env-overrides-file, while FromEnv is env-only. Reference go/cli as the canonical example.
   - Mandatory constraint: the CLI MUST still alias `CANVUS_API_URL` (canonical) + `CANVUS_URL` (deprecated) via explicit BindEnv so env precedence matches SDK behaviour. cli does this.
4. Stage. Commit: `docs(conventions/go): amend §11 — CLI viper precedence (Phase 4d Round A)`.

NO code changes. This is pure documentation.

Report DONE with commit SHA.
```

- [ ] **Step 1: Dispatch all 5 Round A implementer agents in parallel** (single message, 5 Agent tool calls). Sonnet model. NOT background.
- [ ] **Step 2: As each returns, capture commit SHA in your local notes.**
- [ ] **Step 3: Dispatch Round A spec-compliance reviewer.** Single opus agent, given all 5 commits + each task's spec. Look for: scope creep (did anyone touch files outside their spec?), missed steps, half-implementations.
- [ ] **Step 4: Dispatch Round A code-quality reviewer.** Single opus agent, same scope. Look for: dead code introduction, style violations, missing tests, dependency-graph regressions.
- [ ] **Step 5: Fix any Critical findings inline** (dispatch fix agent per finding). Re-review until ✅.
- [ ] **Step 6: Final per-round sanity:**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
git log --oneline | head -8
cd go && go build ./... && go vet ./... && go test ./... 2>&1 | tail -10
cd ../python && uv run pytest 2>&1 | tail -5
```

All green.

---

## Task 4d.2: Round B — SDK additions + small refactors (4 parallel sonnet implementers)

### Task 4d.B1: Go SDK `WithVerifyTLS(bool)` option + 4 callsite refactors

**Files:**
- Modify: `go/sdk/canvus/options.go` (add `WithVerifyTLS(bool) SessionOption`)
- Modify: `go/sdk/canvus/session.go` (apply option in `NewSession` HTTPClient setup)
- Modify: `go/sdk/canvus/options_test.go` (test the option)
- Modify: `go/cli/internal/session/session.go:42` (remove hand-rolled insecure http.Client; use option)
- Modify: `go/cli/internal/commands/login.go:108-111` (remove hand-rolled insecure http.Client; use option)
- Modify: `go/tools/db-solver/internal/commands/session.go:11-15` (remove `insecureHTTPClient()` helper; use option)
- Modify: `go/tools/db-solver/internal/commands/lookup_hash.go:242` (same callsite refactor)

- [ ] **Implementer prompt template:**

```
You are adding a first-class WithVerifyTLS(bool) option to the Go SDK and removing 4 hand-built workaround sites in cli + db-solver.

Context: Phase 4c surfaced that 4 sites built their own `*http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}`. SDK didn't expose a clean toggle.

Tasks:
1. Read go/sdk/canvus/options.go in full to see the existing option-builder pattern (WithAPIKey, WithBaseURL, WithRequestIDFunc etc).
2. Add `WithVerifyTLS(verify bool) SessionOption`:
   - Default behaviour (no call) MUST equal verify=true. This MUST NOT change behaviour for any existing user.
   - When verify=false, the SDK's internal http.Client uses a Transport with `&tls.Config{InsecureSkipVerify: true}`. NEVER swap the user-supplied HTTPClient if one was set via WithHTTPClient — the option is mutually exclusive; document that.
3. Add a unit test in go/sdk/canvus/options_test.go: create a session with WithVerifyTLS(false), assert the http.Client.Transport has the insecure TLS config. Also assert default (no option) does NOT have the insecure config.
4. Refactor the 4 callsites:
   - go/cli/internal/session/session.go: replace the `if cfg.Insecure { sessionCfg.HTTPClient = ... }` block with `if cfg.Insecure { opts = append(opts, canvus.WithVerifyTLS(false)) }`.
   - go/cli/internal/commands/login.go:108-111: same pattern.
   - go/tools/db-solver/internal/commands/session.go: delete the `insecureHTTPClient()` helper function entirely. Update its caller.
   - go/tools/db-solver/internal/commands/lookup_hash.go:242: same callsite refactor.
5. Verify all 4 packages build, vet, test clean (sdk/canvus, cli/, tools/db-solver/).
6. gofmt -s -l empty for all modified files.
7. Stage all 7 files. Commit: `feat(sdk/go): add WithVerifyTLS option; remove 4 hand-rolled insecure clients (Phase 4d Round B)`.

IMPORTANT: The behavioural change is opt-in. Default = verify TLS. Make sure NO existing test breaks.

Report DONE with commit SHA, or BLOCKED.
```

### Task 4d.B2: `WithSubscribeBuffer(int)` across Go/Python/TS

**Files:**
- Modify: `go/sdk/canvus/options.go` + `go/sdk/canvus/subscribe.go` (or wherever the buffer=4 hardcode lives)
- Modify: `python/sdk/src/canvus_sdk/client.py` (settings/options entry)
- Modify: `typescript/sdk/src/session.ts` (option in SessionConfig)
- Tests in all 3 SDKs

- [ ] **Implementer prompt template:**

```
You are adding a SubscribeBuffer (channel/queue size) configuration option to all 3 SDKs.

Current state: each SDK hardcodes buffer=4 inside its subscribe primitive. High-throughput consumers (live dashboards, ai-personas) hit backpressure.

Tasks:
1. Find the hardcoded buffer=4 in each SDK:
   - Go: grep for `make(chan` in go/sdk/canvus/subscribe.go and related files. Confirm capacity 4.
   - Python: grep for `Queue(maxsize=` or `asyncio.Queue(4)` in python/sdk/src/canvus_sdk/transport.py or related.
   - TS: grep for buffer/queue size in typescript/sdk/src/transport.ts.
2. Add option:
   - Go: `WithSubscribeBuffer(size int) SessionOption` in options.go. Validate size >= 1; panic on <1 (or return error).
   - Python: add `subscribe_buffer: int = 4` to client/settings constructor (mirror existing options like `request_timeout`).
   - TS: add `subscribeBuffer?: number` to SessionConfig (typescript/sdk/src/session.ts).
3. Plumb through to the subscribe primitive in each SDK. Default behaviour remains buffer=4.
4. Add one unit test per SDK: configure size=16, assert internal buffer is 16. Configure default, assert 4.
5. Update conventions doc IF the option is genuinely new public surface — add a 2-line bullet under §"Session options" in docs/conventions/{go,python,typescript}.md.
6. Verify each SDK: build, lint, tests clean.
7. Stage all changes. Commit: `feat(sdk): add SubscribeBuffer option in Go/Python/TS (Phase 4d Round B)`.

Watch out: do not change the buffer in places where it's part of an internal protocol-specific invariant (e.g., NDJSON dedup window). If you find such a constraint, document why the new option does not apply there and leave that buffer hardcoded.

Report DONE with commit SHA, or BLOCKED.
```

### Task 4d.B3: TS `createAnyWithAsset` decision

**Files (option A — implement):**
- Modify: `typescript/sdk/src/resources/widgets.ts` (add `createAnyWithAsset` method)
- Modify: `typescript/sdk/tests/widgets.test.ts` (new test)
- Modify: `docs/api-reference/parity-matrix.md` (mark asymmetry resolved)

**Files (option B — document):**
- Modify: `docs/conventions/typescript.md` (Amendment explaining the asymmetry vs Go's `CreateWidget(io.Reader)`)
- Modify: `docs/api-reference/parity-matrix.md` (mark asymmetry intentional + document workaround)

- [ ] **Implementer prompt template:**

```
You are deciding between adding TS createAnyWithAsset for parity with Go's CreateWidget(io.Reader), or documenting the asymmetry as permanent.

Context from Phase 4b review: Go's CreateWidget accepts an io.Reader for image/pdf/video uploads. TS createAny only sends JSON. Either we add a multipart variant, or we ship the asymmetry.

Tasks:
1. Read go/sdk/canvus/widgets.go to see CreateWidget's signature + the multipart-form wire format it uses.
2. Read typescript/sdk/src/resources/widgets.ts to see createAny's current shape.
3. Read typescript/sdk/src/transport.ts to see how multipart/FormData is handled (already supported for individual createImage/createPDF/createVideo? — confirm).
4. Decision: if the transport already handles FormData and there's a clean way to plug it into createAny, implement option A. If the multipart path is type-tangled enough to risk corruption of existing methods, document option B.
5. If A: add `createAnyWithAsset(canvasId: string, payload: WidgetCreate, blob: Blob | Buffer | ReadableStream, contentType: string): Promise<Widget>`. Mirror Go's signature semantics. Add a vitest test. Mark parity-matrix asymmetry resolved.
6. If B: add Amendment to docs/conventions/typescript.md (dated 2026-05-18 Phase 4d), explaining the trade-off (typed asymmetry + clean transport types) and pointing TS consumers to `createImage`/`createPDF`/`createVideo` for asset uploads.
7. Either way, verify SDK builds, lints, tests pass.
8. Commit: `feat(sdk/ts): add createAnyWithAsset for parity (Phase 4d Round B)` OR `docs(conventions/ts): document createAnyWithAsset asymmetry (Phase 4d Round B)`.

Surface the decision in your final report.

Report DONE with commit SHA + which option you chose + why.
```

### Task 4d.B4: Docs sweeps — TS SDK READMEs + spec doc back-apply (bundled)

**Files:**
- Modify: 6 TS example READMEs under `typescript/examples/core/*/README.md` — find the "Expected output" sections that reference non-existent server fields (cascade from Phase 4a C1).
- Modify: `docs/api-reference/endpoints/server.md`
- Modify: `docs/api-reference/endpoints/auth.md`

Both back-applied from `docs/api-reference/VERIFIED-CORRECTIONS.md` (Tier 1 #2/#4 + License GET correction).

- [ ] **Implementer prompt template:**

```
You are performing two pure-doc back-applications and one README sweep.

# Part 1: TS example READMEs cleanup
Phase 4a C1 fix rewrote 21 TS SDK source files to use verified-correct wire-shape fields. Six TS example READMEs in typescript/examples/core/*/README.md still contain "Expected output" snippets referencing the old (wrong) field names.

Tasks:
1. ls typescript/examples/core/ to find all 8 example dirs.
2. For each, grep its README.md for field names that don't exist on the current wire shape. Cross-reference with the actual types in typescript/sdk/src/types/*.ts.
3. Fix each occurrence. Examples of likely issues: `host_id` (should be `host-id` in IPVideo wire shape but `hostId` in TS), nested ServerConfig shape, License field renames.
4. If any README references fields that DO match the current TS types, leave it.

# Part 2: Spec doc back-apply
docs/api-reference/VERIFIED-CORRECTIONS.md documents the actual server behaviour discovered via live curl in Phase 3.t. Two endpoint docs still show the old (wrong) info:
- docs/api-reference/endpoints/server.md (License install path was wrong; License GET field names; ServerConfig shape)
- docs/api-reference/endpoints/auth.md (Login body shape — server REJECTS `username` field)

Tasks:
5. Read VERIFIED-CORRECTIONS.md in full first.
6. Read both endpoint docs.
7. Apply the verified corrections to each, preserving the existing doc style. Add a `> Verified against dev-mtcs.multitaction.com 2026-05-18` blockquote at the top of any section where you changed the spec.
8. Update docs/api-reference/VERIFIED-CORRECTIONS.md to mark each section as "back-applied". Don't delete the corrections themselves — they remain the historical record.

# Verification
9. No code-build steps; this is pure docs. But run `grep -rn "TBD\|TODO\|FIXME" docs/api-reference/endpoints/server.md docs/api-reference/endpoints/auth.md` and confirm no new placeholders introduced.
10. Stage all modified files. Commit: `docs: TS example READMEs + spec doc back-apply from VERIFIED-CORRECTIONS (Phase 4d Round B)`.

Report DONE with commit SHA + count of files touched.
```

- [ ] **Step 1: Dispatch all 4 Round B implementer agents in parallel.** Sonnet model.
- [ ] **Step 2: Capture SHAs.**
- [ ] **Step 3: Round B spec-compliance reviewer** (opus, all 4 commits).
- [ ] **Step 4: Round B code-quality reviewer** (opus, same scope).
- [ ] **Step 5: Fix Critical findings inline.**
- [ ] **Step 6: Sanity check:**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
cd go && go build ./... && go vet ./... && go test ./... 2>&1 | tail -10
cd ../python && uv run pytest 2>&1 | tail -5
cd ../typescript && pnpm --filter @mt-canvus-tools/sdk test 2>&1 | tail -5
```

All green.

---

## Task 4d.3: Round C — Refactor + coverage + baseline (5 parallel mixed implementers)

C1 depends on Round A's item 11 (note-mapper on new genai). Verify before dispatch:

```bash
grep -n "google.golang.org/genai" /home/jaypaulb/Projects/gh/MT-Canvus-Tools/go/examples/projects/note-mapper/go.mod
```

Expected: line present. If missing, halt and investigate Round A.

### Task 4d.C1: Extract shared Gemini helper (opus)

**Files:**
- Create: `go/examples/internal/llm/gemini.go` (new helper)
- Create: `go/examples/internal/llm/gemini_test.go`
- Modify: `go/tools/translator/internal/gemini.go` (use new helper; trim duplication)
- Modify: `go/examples/projects/ai-personas/internal/ai/gemini.go` (use new helper)
- Modify: `go/examples/projects/note-mapper/internal/llm/extract.go` (use new helper)
- Modify: `go/examples/projects/note-mapper/cmd/main.go` (if it constructed clients directly)

**Decision:** Place the helper under `go/examples/internal/llm/` (private to examples), not `go/sdk/extras/llm/`. The SDK is API-only; LLM integration is consumer-side. If a future tool needs it, it can be promoted via Chesterton-fence review then.

- [ ] **Implementer prompt template:**

```
You are extracting a shared Gemini helper from 3 duplicate copies (translator, ai-personas, note-mapper) into go/examples/internal/llm/gemini.go.

Pre-conditions (verify FIRST):
1. note-mapper MUST already be on google.golang.org/genai (Round A item 11). Check go/examples/projects/note-mapper/go.mod.
2. translator + ai-personas both already use google.golang.org/genai (confirmed during Phase 4c).

Tasks:
1. Read all three current helpers:
   - go/tools/translator/internal/gemini.go (look at TranslateText, buildTranslatePrompt, formatTranslation)
   - go/examples/projects/ai-personas/internal/ai/gemini.go
   - go/examples/projects/note-mapper/internal/llm/extract.go
2. Identify the genuinely shared surface — likely a `Client` wrapper with `NewClient(ctx, apiKey, model) (*Client, error)` and `Complete(ctx context.Context, prompt string) (string, error)` (or `CompleteJSON(ctx, prompt, &target) error`). DO NOT pull in the consumer-specific prompt-building logic — those stay in each consumer.
3. Design the helper:
   - Package: `package llm` under `go/examples/internal/llm/`.
   - Exported types: `Client` (with internal genai.Client wrapped), `Config { APIKey, Model, Temperature }`.
   - Exported funcs: `NewClient(ctx, Config) (*Client, error)`, `(*Client).Complete(ctx, prompt) (string, error)`, `(*Client).Close() error`.
   - Should be testable without a live key — at minimum, expose enough for callers to unit-test their prompt-builders. The genai client itself doesn't need mocking for the helper.
4. Write the helper. Aim for <150 LOC (molecule-tier per atomic design).
5. Write unit tests in gemini_test.go — focus on Config validation (empty APIKey, empty Model rejection), client construction. The actual Complete call doesn't need a live key in unit tests (only integration tests live elsewhere or run with key).
6. Refactor each consumer to use the new helper. Each consumer keeps its own prompt-building + response-parsing logic. The duplication you remove is ONLY the genai.NewClient / GenerateContent / response.Text() extraction boilerplate.
7. Update go.mod / go.sum across affected modules (you may need go work sync or per-module go mod tidy).
8. Verify each affected module builds + tests clean:
   - go/tools/translator/
   - go/examples/projects/ai-personas/
   - go/examples/projects/note-mapper/
   - go/examples/internal/llm/ (the new helper)
9. gofmt -s -l empty across all modified files.
10. Stage. Commit: `refactor(examples): extract shared Gemini helper (Phase 4d Round C)`.

DO NOT create go/sdk/extras/llm/. Per consolidation-design, the SDK is API-only. Future SDK promotion is a separate decision.

Report DONE with commit SHA + LOC delta (how many lines removed across consumers, how many added in helper).
```

### Task 4d.C2: Python mypy strict baseline reduction (sonnet)

**Files:**
- Modify: several `python/sdk/src/canvus_sdk/resources/*.py` (the files with the `list[T] annotation vs .list() method` pattern)

- [ ] **Implementer prompt template:**

```
You are reducing the Python SDK's mypy --strict baseline from 38 errors to ≤30.

Background: from CONSOLIDATION-STATUS Phase 4d tracker, "4 errors share the same `list[T] annotation vs .list() method` pattern". That pattern alone gets you 4 below baseline; you have headroom for 4 more arbitrary fixes.

Tasks:
1. Run baseline: `cd python && uv run mypy --strict src/ 2>&1 | tee /tmp/mypy-baseline.txt`. Count "error:" lines. Confirm 38 (or accept current count).
2. Read /tmp/mypy-baseline.txt to identify the `list[T] vs .list()` pattern errors. Find ~4. The common fix is either:
   - Rename the resource method from `.list()` to `.all()` or `.list_items()` (if breaking change acceptable for v0.x). NOT this choice — breaking SDK surface for cosmetic mypy is wrong.
   - Use `typing.List[T]` annotation instead of `list[T]` in the files where the conflict occurs. Cleaner.
3. Apply Option B: change `list[T]` to `typing.List[T]` in the conflicting files. Add `from typing import List` if missing.
4. Re-run mypy. Confirm those 4 errors gone. Take stock of remaining errors.
5. Look for 4 more easy wins:
   - Missing return type annotations (`def foo():` → `def foo() -> None:`).
   - Implicit Optional (`def foo(x: str = None)` → `x: str | None = None`).
   - Unnecessary cast that mypy flags.
   AVOID structural rewrites or anything that changes runtime behaviour.
6. Re-run mypy. Confirm count ≤30.
7. Run pytest to confirm no behavioural regression: `cd python && uv run pytest`.
8. Stage modified files. Commit: `refactor(sdk/python): reduce mypy --strict baseline from 38 to <count> (Phase 4d Round C)`.

If you cannot get under 30 without rewriting structure, STOP. Commit what you have, surface in your report: "Reached <count>; further reduction requires structural changes." That is acceptable.

Report DONE with commit SHA + new baseline number.
```

### Task 4d.C3: TS lint baseline reduction (sonnet)

**Files:**
- Modify: various `typescript/sdk/src/**/*.ts`
- Possibly: `typescript/sdk/tsconfig.json` (test include)

- [ ] **Implementer prompt template:**

```
You are reducing the TS SDK lint baseline from 64 errors to ≤30.

Background per CONSOLIDATION-STATUS: "bulk of remaining errors are `UserId` template-literal interpolation, `request<void>` returns, `ReadonlyArray<T>` syntax, tsconfig-include for tests. ~1 day cleanup."

Tasks:
1. Run baseline: `cd typescript && pnpm --filter @mt-canvus-tools/sdk lint 2>&1 | tee /tmp/lint-baseline.txt`. Count errors.
2. Categorize the 64 errors:
   - UserId template literal: should be `\`${userId}\`` or use the branded type's `.toString()`. Apply the pattern that matches surrounding code.
   - `request<void>` returns: change to `request<undefined>` or remove the type-param to let it infer. Match existing patterns.
   - `ReadonlyArray<T>` syntax: change to `readonly T[]` (the modern syntax — confirm in tsconfig/eslint config which is enforced).
   - tsconfig-include for tests: extend tsconfig.json to include test glob, or create tsconfig.test.json. Pick whichever pattern matches existing repo style.
3. Apply fixes in batches by category. Re-run lint after each batch.
4. Run `pnpm --filter @mt-canvus-tools/sdk typecheck && pnpm --filter @mt-canvus-tools/sdk build && pnpm --filter @mt-canvus-tools/sdk test`. All green.
5. Confirm lint ≤30 errors. If you hit a stubborn category you can't fix safely, stop and commit what you have.
6. Stage modified files. Commit: `refactor(sdk/ts): reduce lint baseline from 64 to <count> (Phase 4d Round C)`.

Behavioural changes prohibited. Pure lint/style. If any fix triggers a typecheck or test failure, revert that specific fix and pick another.

Report DONE with commit SHA + new baseline number.
```

### Task 4d.C4: Reclaim TS coverage thresholds (opus)

**Files:**
- Modify: `typescript/sdk/vitest.config.ts` (raise thresholds: 50/60/40/50 → at least 60/70/60/60)
- Add: `typescript/sdk/tests/auth.test.ts` (focused tests for auth.ts error mapping)
- Add: `typescript/sdk/tests/users.test.ts` (focused error-mapping tests)
- Add: `typescript/sdk/tests/server.test.ts` (focused error-mapping tests)

- [ ] **Implementer prompt template:**

```
You are reclaiming TS SDK coverage thresholds from 50/60/40/50 (lines/statements/functions/branches) to at least 60/70/60/60, by adding focused tests for auth.ts, users.ts, server.ts.

Background: Phase 4b agent pragmatically relaxed thresholds to ship. Phase 4d closes that gap.

Tasks:
1. Run baseline coverage: `cd typescript && pnpm --filter @mt-canvus-tools/sdk test:coverage 2>&1 | tee /tmp/coverage-baseline.txt`. Capture per-file numbers.
2. Identify auth.ts, users.ts, server.ts (resource layer) current coverage. The error-mapping branches (401 → AuthError, 404 → NotFoundError, 429 → RateLimitError, 5xx → ServerError) are likely uncovered.
3. Write focused vitest tests:
   - tests/auth.test.ts: mock transport, simulate each error status, assert correct error type thrown. Test successful login response parse.
   - tests/users.test.ts: same pattern for users.list / users.get / users.create. Mock errors at transport level.
   - tests/server.test.ts: same for server.config / server.installLicense / server.sendTestEmail.
4. Aim for ≥80% coverage on each of these three resource files.
5. Re-run coverage. Confirm at least 60/70/60/60 globally.
6. Raise thresholds in vitest.config.ts to match what you achieved (don't set them artificially low — target your actual numbers).
7. Run full test suite. All pass.
8. Stage all test files + vitest.config.ts. Commit: `test(sdk/ts): reclaim coverage thresholds (Phase 4d Round C)`.

If you cannot get coverage to ≥60/70/60/60 without adding tests for files NOT in spec, STOP and surface that — the thresholds may need a different baseline.

Report DONE with commit SHA + per-file coverage delta.
```

### Task 4d.C5: Asset roundtrip integration tests TS + Python (opus)

**Files:**
- Add: `typescript/sdk/tests/export-import-roundtrip.test.ts`
- Add: `python/sdk/tests/test_export_import_roundtrip.py`

- [ ] **Implementer prompt template:**

```
You are adding asset roundtrip integration tests for TS and Python — Go already has one.

Background: Phase 4b shipped export/import in all three SDKs with canonical schema parity (Critical fix in e50001c). TS coverage was image-write only, not full roundtrip. Python had no roundtrip test.

Tasks:
1. Read go/sdk/canvus/export_test.go (or wherever the Go roundtrip integration test lives) to understand the canonical pattern: export a small canvas to a tmpdir, parse export.json, re-import to a different canvas ID (or same with --merge), verify widget count + asset SHAs match.
2. Decide test mode: full live-server integration test (requires .secrets) or schema-only roundtrip (read fixture export.json, write parsed widgets back, diff). Recommend SCHEMA-ONLY for unit-style coverage. Live integration is Round D1's territory (live verification).
3. Write typescript/sdk/tests/export-import-roundtrip.test.ts:
   - Build a synthetic export bundle in-memory (matching canonical schema: {widgets, assets, region} with flat asset siblings).
   - Pass through the importer.
   - Assert widget shapes preserved, asset file mapping preserved.
   - Use the same fixture pattern as existing vitest tests.
4. Write python/sdk/tests/test_export_import_roundtrip.py:
   - Same structure, pytest-style.
   - Use the canvus_sdk.extras.export and .import_ modules.
   - Use tmp_path fixture for the bundle dir.
5. Run both SDK test suites. All pass.
6. Stage. Commit: `test(sdk): add export/import roundtrip tests for TS and Python (Phase 4d Round C)`.

If you find the export/import implementations diverge from Go's canonical schema, STOP and surface as a Critical finding — Phase 4b fix should have closed that gap.

Report DONE with commit SHA.
```

- [ ] **Step 1: Dispatch all 5 Round C implementer agents in parallel.** Mixed sonnet/opus per task above.
- [ ] **Step 2: Capture SHAs.**
- [ ] **Step 3: Round C spec-compliance reviewer** (opus, all 5 commits).
- [ ] **Step 4: Round C code-quality reviewer** (opus, same scope).
- [ ] **Step 5: Fix Critical findings inline.**
- [ ] **Step 6: Sanity check:**

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools
cd go && go build ./... && go vet ./... && go test ./... 2>&1 | tail -10
cd ../python && uv run pytest && uv run mypy --strict src/ 2>&1 | tail -5
cd ../typescript && pnpm --filter @mt-canvus-tools/sdk lint && pnpm --filter @mt-canvus-tools/sdk test:coverage 2>&1 | tail -10
```

All green. Capture new mypy/lint/coverage numbers for the CONSOLIDATION-STATUS update.

---

## Task 4d.4: Round D — Live verification + MCP LLM port (2 parallel opus)

These are independent: D1 touches Go + Python SDK warnings + parity-matrix; D2 lives entirely under `python/tools/mcp-server/`. Dispatch in parallel.

### Task 4d.D1: Live-verify permissions subscribe

**Files:**
- Use: `.secrets` (env source for live API)
- Modify (on PASS): `go/sdk/canvus/permissions.go` (drop "not yet live-verified" warnings); `python/sdk/src/canvus_sdk/resources/permissions.py` (same); `docs/api-reference/parity-matrix.md` §5.5 (mark verified)
- Modify (on FAIL): same files, but with "VERIFIED FAILING" warnings + escalation note

- [ ] **Implementer prompt template:**

```
You are live-verifying SubscribeCanvasPermissions and SubscribeFolderPermissions against dev-mtcs.multitaction.com. On pass, drop the "not yet verified" warnings. On fail, escalate to documented gap with reproduction steps.

Pre-conditions:
1. .secrets file at /home/jaypaulb/Projects/gh/MT-Canvus-Tools/.secrets must contain CANVUS_API_URL + CANVUS_API_KEY.
2. test_canvas_id from .secrets must exist on the server.

Verification procedure (write as a single Go integration test file, then a single Python one):

# Go test
1. Create go/sdk/canvus/permissions_live_test.go (with build tag `//go:build live` so it doesn't run in default CI).
2. Test plan:
   a. Open SubscribeCanvasPermissions for the test canvas.
   b. Drain initial snapshot (settle 2s).
   c. Use a separate session (or curl exec) to PATCH the canvas permissions: add a sentinel user/group at a known permission level.
   d. Receive the change event from the subscription. Assert it matches.
   e. Revert the permission change.
   f. Close the subscription. Assert no leak.
3. Run: `cd go/sdk/canvus && go test -tags=live -run TestSubscribeCanvasPermissionsLive -v`. Pass/fail.
4. Repeat for SubscribeFolderPermissions against a test folder (create one if needed; use trash to clean up).

# Python test
5. Create python/sdk/tests/test_permissions_live.py with a `@pytest.mark.live` marker (and update pyproject.toml or pytest.ini to register the marker if not present).
6. Same test plan as Go, using the Python async iterator pattern.
7. Run: `cd python && uv run pytest sdk/tests/test_permissions_live.py -m live -v`.

# Outcome handling
- IF BOTH PASS:
   a. Edit go/sdk/canvus/permissions.go: remove the warnOnce / docstring warnings about "not yet live-verified" for both Subscribe* methods.
   b. Edit python/sdk/src/canvus_sdk/resources/permissions.py: remove the equivalent `.. warning::` rst block.
   c. Update docs/api-reference/parity-matrix.md §5.5 from "deferred to Phase 4d live-verification" to "verified 2026-05-18 against dev-mtcs.multitaction.com".
   d. Commit: `verify(sdk): live-verify Subscribe{Canvas,Folder}Permissions; drop warnings (Phase 4d Round D)`.
- IF EITHER FAILS:
   a. KEEP the warnings.
   b. Update the warning text to: "VERIFIED FAILING against dev-mtcs.multitaction.com 2026-05-18 — events do not fire on permission change. Use polling instead. See parity-matrix.md §5.5."
   c. Update parity-matrix §5.5 with failure mode + reproduction steps from your test.
   d. File a tracker entry for the canvus-server issue (in CONSOLIDATION-STATUS, under "Upstream doc fix tracker" — but DO NOT push to canvus-server#96 directly; surface to Jaypaul in your DONE report).
   e. Commit: `docs(sdk): document Subscribe{Canvas,Folder}Permissions failure on dev-mtcs (Phase 4d Round D)`.
- IF MIXED (one pass, one fail): handle per-helper.

Critical: NEVER report PASS without seeing the change event actually arrive in your test. A "no errors" run is not a pass; receipt of the change event is.

Per CLAUDE.md: report failures with raw error + theory. Stop and ask Jaypaul if the test infra fails (e.g., can't connect to server) before committing anything.

Report DONE with commit SHA + per-helper outcome + brief evidence (event JSON received).
```

### Task 4d.D2: MCP LLM tools port (LARGE)

**Files:**
- Modify: `python/tools/mcp-server/src/canvus_mcp_server/mcp_tools/llm.py` (currently 5-6 MCPToolExecutionError stubs)
- Modify: `python/tools/mcp-server/src/canvus_mcp_server/mcp_tools/brainstorming.py` (stubs)
- Modify: `python/tools/mcp-server/src/canvus_mcp_server/mcp_tools/correlation.py` (stubs)
- Modify: `python/tools/mcp-server/src/canvus_mcp_server/mcp_tools/reports.py` (stubs)
- Create: `python/tools/mcp-server/src/canvus_mcp_server/llm/__init__.py`
- Create: `python/tools/mcp-server/src/canvus_mcp_server/llm/ollama.py` (httpx port of the legacy 541-LOC aiohttp client)
- Create: `python/tools/mcp-server/src/canvus_mcp_server/llm/cache.py` (SQLite cache, if currently a separate module)
- Create: `python/tools/mcp-server/src/canvus_mcp_server/llm/pdf.py` (PDF processing pipeline)
- Update: `python/tools/mcp-server/pyproject.toml` (httpx already present; possibly add aiosqlite, pypdf, etc.)
- Update: `python/tools/mcp-server/src/canvus_mcp_server/settings.py` (add Ollama config fields if not already present)
- Tests: `python/tools/mcp-server/tests/test_llm_*.py`, `test_brainstorming.py`, `test_correlation.py`, `test_reports.py`, `test_ollama_client.py`

- [ ] **Implementer prompt template:**

```
You are porting the legacy canvus-mcp-server's LLM toolchain (13 tools across 4 files + 541-LOC Ollama client + SQLite cache + PDF processing) into the monorepo's mcp-server-python item, rewritten against monorepo conventions.

# Source for reference
Legacy code lives at: /home/jaypaulb/Projects/gh/canvus-mcp-server/ (archived in Phase 4c; READ-ONLY).
- src/llm_client.py (~541 LOC, aiohttp-based) — port to httpx.
- src/cache.py (SQLite cache for LLM responses).
- src/pdf_handler.py (PDF download + parse + chunked LLM summarize).
- src/mcp_tools/llm.py, brainstorming.py, correlation.py, reports.py — 13 tools total. Read each to understand the contract.

# Destination
python/tools/mcp-server/src/canvus_mcp_server/ in the monorepo. Existing files: llm.py, brainstorming.py, correlation.py, reports.py — currently contain MCPToolExecutionError stubs preserving the public tool surface.

# Conventions (NON-NEGOTIABLE)
1. `httpx` async client, NOT `aiohttp`. The convention is documented in docs/conventions/python.md.
2. NO module-scoped globals. Configuration comes via a pydantic Settings constructor (mirror the existing canvus_mcp_server settings.py pattern).
3. structlog for logging, NOT print() or logging.getLogger().
4. mypy --strict must pass on all new code (no `# type: ignore` without explicit reason comment).
5. ruff must pass clean.
6. Errors raised as MCPToolExecutionError (preserving public surface) with structured detail dicts — never silent fallback.

# Architecture
- Create a `canvus_mcp_server.llm` subpackage:
  - `ollama.py`: OllamaClient class. Constructor takes OllamaConfig (BaseURL, Model, Timeout, Retries). Methods: `async def generate(prompt: str, **opts) -> str`, `async def chat(messages: list[Message]) -> str`, `async def embed(text: str) -> list[float]` (if used). Uses httpx.AsyncClient with proper connection reuse (per-instance, NOT per-call construction).
  - `cache.py`: SQLite cache wrapping aiosqlite. Key by SHA256(model + prompt). TTL configurable. Method: `async def get_or_compute(key, factory)`.
  - `pdf.py`: PDF fetch (httpx) → parse (pypdf or pymupdf, pick whichever the legacy used unless it has known security issues) → chunk → multi-call summarize via OllamaClient with structured result merge.
- Each tool file (llm.py, brainstorming.py, correlation.py, reports.py) replaces its stubs with real implementations that depend on the new subpackage via dependency injection (pass OllamaClient + cache to the tool factory, not module-globals).

# Tools to port (read legacy contracts FIRST, preserve names + signatures)
Approx (verify against legacy source):
- llm.py: chat_with_canvus_context, summarize_canvas, extract_key_themes
- brainstorming.py: generate_ideas, refine_ideas, brainstorm_session
- correlation.py: correlate_notes, find_similar_widgets, semantic_search
- reports.py: generate_canvas_report, export_summary_pdf, executive_summary, weekly_report

# Tasks
1. Read python/tools/mcp-server/ in full to understand current shape. Note all stubs.
2. Read /home/jaypaulb/Projects/gh/canvus-mcp-server/src/ — focus on llm_client.py, cache.py, pdf_handler.py, mcp_tools/*. Take notes on:
   - Each public function's signature
   - Each function's contract (what does it return? what errors?)
   - Any test fixtures useful as reference
3. Read python/tools/mcp-server/src/canvus_mcp_server/settings.py to see how Settings are constructed currently. You will extend it.
4. Read docs/conventions/python.md for all relevant convention rules.
5. Stage your work in this order (commit at the end of each substantial chunk so review can happen mid-flight):
   a. ollama.py + ollama_test.py (focused unit tests with httpx mock_transport)
   b. cache.py + cache_test.py
   c. pdf.py + pdf_test.py
   d. Settings extension
   e. Tool file ports, one file at a time, with tests per tool
6. Per file, run: `cd python && uv run ruff check tools/mcp-server/src/canvus_mcp_server/llm/ && uv run mypy --strict tools/mcp-server/src/canvus_mcp_server/llm/ && uv run pytest tools/mcp-server/tests/`.
7. After all tools ported, run the full test suite + a smoke test that imports the server and lists all available tools (confirm no tool is still a stub).
8. Update python/tools/mcp-server/README.md: remove the "LLM tools are stubs" section. Add a "Configuration" section documenting the new Ollama env vars.
9. Per-file commits with explicit Phase 4d tags. Example:
   - `feat(tools/mcp-server): port OllamaClient to httpx (Phase 4d Round D)`
   - `feat(tools/mcp-server): port LLM response SQLite cache (Phase 4d Round D)`
   - `feat(tools/mcp-server): port PDF processing pipeline (Phase 4d Round D)`
   - `feat(tools/mcp-server): port llm.py tools (Phase 4d Round D)`
   - `feat(tools/mcp-server): port brainstorming.py tools (Phase 4d Round D)`
   - `feat(tools/mcp-server): port correlation.py tools (Phase 4d Round D)`
   - `feat(tools/mcp-server): port reports.py tools (Phase 4d Round D)`
   - `docs(tools/mcp-server): document LLM configuration (Phase 4d Round D)`

# Scope guards
- Do NOT touch the canvus_mcp_server/mcp_tools/users.py escape hatch — that was Round A item 16.
- Do NOT modify python/sdk/. The SDK is feature-complete for this port; you consume it.
- If a legacy LLM tool depends on a feature missing from the monorepo Python SDK, STOP and surface as BLOCKED with the gap details — DO NOT add the gap to the SDK in this task.
- If you cannot replicate a legacy behaviour exactly (e.g., legacy used aiohttp's specific connection reuse semantics), document the deviation in the tool's docstring with a "Behavioural diff from legacy:" note.

# Reporting
Long task. As you commit each chunk, capture the SHA. Final DONE report includes:
- All commit SHAs (chronological)
- Per-tool status (ported / stubbed / blocked)
- Test counts before vs after
- New mypy / ruff numbers for the mcp-server subdirectory
- Any deviations from legacy behaviour
- Any new Phase 4e items surfaced

If you BLOCK at any point, surface immediately with the specific gap. Do not silently stub something that should be real.
```

- [ ] **Step 1: Dispatch D1 and D2 in parallel** (single message, 2 Agent tool calls). Both opus.
- [ ] **Step 2: D2 will return many commit SHAs across many sub-commits.** Capture all of them.
- [ ] **Step 3: Round D spec-compliance reviewer** (opus). Given D1 + D2's commit list + each task's spec.
- [ ] **Step 4: Round D code-quality reviewer** (opus). Same scope. For D2 specifically, focus on: convention adherence (httpx vs aiohttp, no globals, structlog), test coverage of new LLM subpackage, no broken stubs left behind, README documents the new configuration.
- [ ] **Step 5: Fix Critical findings inline.**
- [ ] **Step 6: Sanity check** (this will take a few minutes):

```bash
cd /home/jaypaulb/Projects/gh/MT-Canvus-Tools

# Full workspace verify
cd go && go build ./... && go vet ./... && go test ./... 2>&1 | tail -10
cd ../python && uv run pytest && uv run ruff check && uv run mypy --strict src/ 2>&1 | tail -5
cd ../typescript && pnpm --filter @mt-canvus-tools/sdk lint && pnpm --filter @mt-canvus-tools/sdk test 2>&1 | tail -10

# MCP server-specific
cd python && uv run python -c "from canvus_mcp_server.server import build_server; s = build_server(); print(f'{len(s.list_tools())} tools registered')"
```

All green. MCP server should now report all 13+ tools registered (vs the previous count which had stubs).

---

## Task 4d.5: Final review gate + CONSOLIDATION-STATUS update

- [ ] **Step 1: Dispatch final reviewer** (opus, background). Prompt:

```
You are the final review gate for Phase 4d of MT-Canvus-Tools consolidation.

Scope: every commit between `1820a01` (Phase 4c completion) and HEAD.

Review goals:
1. SCOPE: Did Phase 4d close all 17 items from CONSOLIDATION-STATUS "Deferred to Phase 4d"? Item-by-item check.
2. QUALITY: Any commits introduce new convention violations? New baseline regressions in mypy/lint/coverage?
3. ARCHITECTURE: Any cross-cutting issues — leaky abstractions, broken parity, missing tests?
4. SDK HEALTH: Are all 3 SDKs still at 147/147 endpoint coverage? Subscribe still 36/27/27? Asset roundtrip tested all 3 ways?
5. MCP SERVER: All originally-stubbed tools real? Convention adherence (httpx, structlog, no module globals, mypy strict)?
6. LIVE VERIFY: D1 result clearly recorded and reflected in parity-matrix?
7. DOCS: CONSOLIDATION-STATUS updates this round? VERIFIED-CORRECTIONS back-applies recorded? READMEs accurate?

Output:
- For each item: ✅ shipped / ⚠️ partial / ❌ missed
- Critical findings (must fix to call 4d done)
- Important findings (defer to Phase 5 tracker — capture clearly)
- Nice-to-have findings (defer to Phase 5 tracker)
- Overall verdict: PHASE 4D COMPLETE / NEEDS FIXES

Reference docs:
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/CONSOLIDATION-STATUS.md
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/superpowers/plans/2026-05-18-mt-canvus-tools-phase4d-cleanup-and-mcp-llm-port.md
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/parity-matrix.md
- /home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/VERIFIED-CORRECTIONS.md
```

- [ ] **Step 2: Address Critical findings.** Dispatch fix agents per finding if any.

- [ ] **Step 3: Update CONSOLIDATION-STATUS.md.**
  - Header: post-Phase 4d
  - Add "What shipped" rows for every Phase 4d commit
  - Add new "Phase 4d outcome — DONE" section: per-item table (17 rows), toolchain status, review-gate verdict
  - Move any Phase 4d deferrals (Important/Nice-to-have findings) to a new "Deferred to Phase 5" subsection
  - If permissions live-verify FAILED, add to parity-matrix §5.5 + tracker
  - Refresh the coverage snapshot table with new mypy/lint/coverage numbers

- [ ] **Step 4: Commit + push the status update.**

```bash
git add CONSOLIDATION-STATUS.md
git commit -m "docs: Phase 4d completion update + Phase 5 tracker"
git push origin main
```

- [ ] **Step 5: Announce phase complete.** Brief summary to Jaypaul: commit count, items shipped, deferrals, suggested next phase (Phase 5 = PowerToys, per CONSOLIDATION-STATUS).

---

## Risk register

- **Round D2 (MCP LLM port) blocking risk.** This is the largest single task across all four phases. If the implementer agent stalls or BLOCKS partway, the fix is to break it into smaller chunks per-tool-file and re-dispatch. The agent prompt already mandates per-chunk commits so partial progress survives.
- **Live verification (D1) might genuinely fail.** Server might not emit permissions subscribe events. If so, the OUTCOME is documented (failing helpers stay warned), not a phase failure.
- **Baseline reductions (C2, C3) might floor.** If mypy can't get below ~32 or lint below ~35 without structural rewrites, accept the higher number and document why in the commit.
- **Shared Gemini helper (C1) depends on Round A item 11.** Verify A2 landed before dispatching C1. If A2 stalled, C1 must wait.

## Self-review

Before dispatching anything: walk the spec → walk this plan → confirm every item in the 17-item table has a Round assignment, a file list, and an implementer prompt section. Confirm no placeholders ("TBD", "fill in details") remain. Confirm dependency order respects A2→C1.

Spot-checked: every item present. ✅
