# MT-Canvus-Tools Consolidation Status

**As of:** 2026-05-19 (post-Phase 6)
**Phases complete:** 0, 1, 2, 3 (+ verification), 4a (+ review-driven fixes), 4b (+ review-driven fixes), 4c (+ review-driven fixes), 4d (+ review-driven fixes), 5 (PowerToys port + carry-over), 6 (top-level documentation)
**Next phase:** 7 (CI/CD)
**Upstream doc fix tracker:** [`canvus-server#96`](https://gitlab.multitaction.com/swrd/conan/canvus/canvus-server/-/work_items/96)

---

## What shipped

| Phase | Output | Commit |
|---|---|---|
| 0 | Repo bootstrap (private, MIT) on GitHub | `5f1eaf4`, `7616ea3` |
| 1 | Canonical API reference: 150 endpoints, 82 examples, 5 changelog entries | `a2dc856` |
| 2 | SDK coverage audit + 3 locked conventions docs | `31e42a6` |
| 2 plan | Phase 3 SDK migration plan | `894a6cb` |
| 3.1 | Go SDK migrated | `a0097d3` |
| 3.2 | Python SDK migrated | `2b04093` |
| 3.3 | TypeScript SDK greenfield | `568e746` |
| 3.v | Coverage verification + Go `CopyFolderPatch` fix | `5e0cc95` |
| 3.t | Live-server verification + Tier-1/Tier-2 fixes + `VERIFIED-CORRECTIONS.md` | `115ad3f` |
| 4a plan | Phase 4a Core Examples plan | `611db46` |
| 4a.pre | TS SDK exactOptionalPropertyTypes fixes + lockfile gen | `e51351f` |
| 4a.1 | Go examples (8 canonical) | `a018cf3` |
| 4a.2 | Python examples (8 canonical) | `30f2f46` |
| 4a.3 | TypeScript examples (8 canonical) | `97297f7` |
| 4a.r | Review-driven fixes (C1-C4 + I1/I3/I4/I5/I8) | `0770701` |
| 4b plan | Phase 4b Cross-SDK Parity Sweep plan + go.work fix | `848434c` |
| 4b.audit | Parity matrix (223 gaps surfaced) | `45f7fec` |
| 4b.1 | Go SDK: 36 Subscribe helpers + extras subpackage (geometry/filters/zones/batch_widgets/search) | `a4a9f58` |
| 4b.2 | Python SDK: subscribe completion + 10-module extras + generic widget CRUD + trash + server convenience | `4a2667b` |
| 4b.3 | TS SDK: extras subpackage + retry/circuit-breaker + verifyTls fix + lifecycle + Rate/Server errors | `3e22b19` |
| 4b.4 | Go examples 05/06/07 refactored to typed Subscribe helpers | `678dbcf` |
| 4b.r | Review-driven fixes (export schema parity + hypot alignment + lint regressions + Python warning) | `e50001c` |
| 4c plan | Phase 4c Per-Item Refresh plan (10 items → 9 items → 8 shipped) | `16242ca`, `8b2590d`, `7540a4b`, `d42ca79` |
| 4c.audit | Per-item refresh audit (10 items, scope-bounded per item) | `7540a4b` |
| 4c.1.1 | Go: translator port | `83b233d` |
| 4c.1.2 | Go: note-mapper port | `1b27679` |
| 4c.1.3 | Go: llm-canvas-companion port (renamed from CanvusAPI-LLMDemo) | `cbe046b` |
| 4c.1.4 | Go: db-solver port | `9b35867` |
| 4c.2.1 | Go: ai-personas port | `bb7f436` |
| 4c.2.2 | TS: webui rewrite (Hono) | `7573af0` |
| 4c.3.1 | Go: canvus-cli port | `2e15a32` |
| 4c.3.2 | Python: mcp-server port | `e51f341` |
| 4c.r | Review-driven fixes (Round 3: env-var alignment + gofmt sweep + ruff tests) | `34d8965` |
| 4d plan | Phase 4d Cleanup + MCP LLM Port plan | `3fcedc9` |
| 4d.A | Round A cleanups: Go SDK dead code, note-mapper genai, ai-personas Subscribe, Python users.current, go conventions §11, gofmt drift, wait_test | `b478961`, `49936b8`, `0f9b149`, `76e995d`, `4425fd2`, `d832f69`, `96b7699` |
| 4d.B | Round B SDK additions: WithVerifyTLS, SubscribeBuffer, createAnyWithAsset doc, TS READMEs + spec back-applies, review fixes | `636da9f`, `25ffede`, `ad2ac3f`, `e1dc39d`, `ba3759d` |
| 4d.C | Round C refactor + coverage + baseline: TS coverage reclaim, export/import roundtrip tests, Python mypy 39→0, TS lint 64→26, Gemini helper extraction, Generic refactor, import_ var fix | `cf874e9`, `b26096d`, `b5cb116`, `798ee09`, `1bee991`, `9173fa8`, `bcdd7d5` |
| 4d.D | Round D live verify + MCP LLM port (16 tools): OllamaClient httpx, SQLite cache, PDF pipeline, Settings, llm/brainstorming/correlation/reports tools, permissions live-verify, mypy/ruff/tests, LLM docs, correlation silent-fallback fix | `f6e5327`, `761c1a6`, `c3b626b`, `b757ec5`, `fe8e162`, `224fad1`, `e0c72ec`, `c7714b8`, `1807060`, `e60b75e`, `ec436a2`, `5ec29ae` |
| 5.plan | Phase 5 plan (PowerToys port + Phase 4d carry-over) | `8b918e9` |
| 5.1 | Go: powertoys scaffold + SDK shim + type aliases | `0b5852e`, `3a81f3b` |
| 5.2 | Go: powertoys Phase A verbatim copy + atoms/webui subscriber + client resolver | `27271b7`, `5aef00b` |
| 5.3 | Go: powertoys molecules — canvas_service SDK rewire + goroutine lifecycle fixes | `7dc93bf`, `58740cd`, `5dd4f26` |
| 5.4 | Go: powertoys molecules — rcu_handler + Phase A organism handlers | `132216d`, `e0da83d`, `61a4ebd` |
| 5.5 | Go: powertoys — canvas_service ctx race fix + Manager decompose (ui/config/lifecycle) + insecureTLS wiring | `d2f4667`, `083bb42`, `caba525`, `9f8925b` |
| 5.carry | Python FolderPermissions + FoldersResource.subscribe_permissions; TS UserId/GroupId narrowed to number + allowNumber lint rule | `cd49f54`, `cdde5ab` |
| 5.r | Review fix: performConnectionTests insecureTLS + workspace tidy + binary gitignore | `ae170ea`, `53efe75` |
| 6 | Top-level documentation: root README, per-language READMEs, getting-started guides, CONTRIBUTING, docs/contributing/ | `3bc0f38`, `810e5be`, `f4f0e34`, `bd3e01f`, `49f61a9`, `a48d088`, `57ff572`, `fa332dd`, `640e56e`, `882f9b3`, `09db7ec`, `b4aae25`, `73607f6` |

**Repo:** `github.com/jaypaulb/MT-Canvus-Tools` (private)

### Coverage snapshot (verified endpoint-by-endpoint, post-Phase 4d)

| SDK | Endpoint coverage | Subscribe coverage | Extras subpackage | Toolchain status |
|---|---|---|---|---|
| Go | **147 / 147** | **36 / 36** typed helpers (permissions live-verified Phase 4d D1) | 5 modules (geometry, filters, zones, batch_widgets, search) | build / vet / test clean; `gofmt -s -l` empty |
| Python | **147 / 147** | **49** typed subscribe methods (Canvas + Folder permissions live-verified Phase 4d D1; FoldersResource `subscribe_permissions` added Phase 5 carry-over `cd49f54`) | 10 modules + `__init__` | ruff clean, all pytest pass (incl. SDK + mcp-server + roundtrip + live opt-ins); **mypy --strict 0 errors** at workspace level |
| TypeScript | **147 / 147** | **27 / 27** typed async iterators | 9 modules + index | typecheck / build clean, 167/167 vitest pass; **lint 0 errors** (`UserId`/`GroupId` narrowed to `number` + `allowNumber: true` ESLint rule, Phase 5 carry-over `cdde5ab`); coverage `70.92 / 74.49 / 64.11 / 70.92` (statements/branches/functions/lines) |

Three endpoints deliberately omitted across all SDKs per changelog §1 + §2.
Permissions-subscribe helpers in Go (`SubscribeCanvasPermissions`, `SubscribeFolderPermissions`) and Python (`canvases.subscribe_permissions`) are now live-verified against `dev-mtcs.multitaction.com` (Phase 4d Task D1, 2026-05-19). Live tests at `go/sdk/canvus/permissions_subscribe_live_test.go` (build tag `live`) and `python/sdk/tests/test_permissions_live.py` (`pytest -m live`). The Python `FoldersResource` does not yet expose `subscribe_permissions` — captured as a Phase 5 follow-up parity item, not a verification gap (see parity-matrix §5.5).

MCP server (Python, `python/tools/mcp-server/`) toolchain: **mypy --strict 0 errors**, ruff clean, 44/44 pytest pass. All 16 originally-stubbed LLM/brainstorming/correlation/reports tools now real implementations (Phase 4d Round D2).

---

## Tier 1 — RESOLVED via live curl session against `dev-mtcs.multitaction.com`

| # | Question | Resolution |
|---|---|---|
| 1 | RDP / IP Video field naming (hyphens vs underscores) | **HYBRID.** Server uses hyphens for `host-id` (IPVideo & RDP), `connection-name`, `content-id`; other fields stay underscored. Go was right (followed C++ source). Python and TS patched to match. |
| 2 | Login / InstallLicense / SendTestEmail body shapes | **Login:** server REJECTS unknown fields — `username` causes failure. Go's defensive double-keying patched out. **InstallLicense:** path was wrong (`/license/install` 404s; correct path is `POST /license`); body field was wrong (`key`/`license-data` rejected; correct is `license`). All 3 SDKs patched. **SendTestEmail:** server ignores body field name (SMTP not configured on dev), all variants safe. No change needed. |
| 3 | Streaming wire format (NDJSON or SSE?) | **NDJSON over HTTP `?subscribe`.** No change needed. WebSocket at `/ws` is canvus-web proxy layer (bridges NDJSON streams to browser SPA), NOT part of `/api/v1/*` surface. SDKs are correct as written. |
| 4 | ServerConfig wire shape (flat array vs nested) | **Nested object.** Spec's flat element-array form doesn't exist on v1.2. Go + Python defensive decoders still work; TS uses open type. Documented in `VERIFIED-CORRECTIONS.md` §3. PATCH write API still uses element-array shape. |

**Additional doc errors found during verification** (all in [`canvus-server#96`](https://gitlab.multitaction.com/swrd/conan/canvus/canvus-server/-/work_items/96)):
- License GET response uses underscores not hyphens; missing `status`/`clients` fields; undocumented `edition`/`type` fields
- User IDs are integers, not UUIDs (spec wrong)
- `/users/me` convenience endpoint does not exist (spec wrong)
- Audit log returns flat array, not envelope (spec wrong)
- Field-naming convention is genuinely hybrid (not documented anywhere)

---

## Tier 2 — RESOLVED with senior-dev calls (80% industry-standard)

| # | Question | Decision | Rationale |
|---|---|---|---|
| 5 | TS camelCase mapping | **Defer to v0.2** | Hybrid wire shape is messier than expected; a careful mapper with test coverage is safer than a quick sweep. Documented in TS conventions Amendments. |
| 6 | TS zod response validation | **Defer to v0.2** | Stripe, Octokit, AWS SDK v3 don't ship runtime validation. ~150 endpoints + 11 widget kinds = multi-day work. TypeScript types provide compile-time safety. Documented in TS conventions Amendments. |
| 7 | TS NotFoundError as 5th error kind | **Keep** | Stripe, AWS, GitHub all distinguish 404 separately. 404 is often normal control flow. Formalised in TS conventions Amendments. |
| 8 | Go `ListAuditEvents` envelope return | **Reverted to flat array** | Live server returns flat array; envelope was a documentation phantom. Go agent's "breaking change" was based on bad spec — now matches Python + TS contract. |

---

## Tier 3 — Deferred to Phase 4

| # | Item | Decision |
|---|---|---|
| 9 | **Cross-SDK parity** for legacy helpers (`geometry`, `search`, `filters`, `export`, `widget_operations`, circular-parenting guard) | Required: full parity across Go / Python / TS. Phase 4 plan will explicitly include a parity-enforcement task that audits each SDK against the others and ports any missing utilities. Recommend a `sdk-extras` subpackage per language for helpers beyond pure REST transport. |
| 10 | Go dead code (`setToken`, `doRequestWithHeaders`, `warnOnce`, `warnAlways`) | Deferred to Phase 4 after Chesterton fence inspection. |

---

## Tier-1/Tier-2 SDK patches applied

**Go SDK** (3 files modified, all changes verified `go build` + `go vet` + tests pass):
- `session.go`: removed `username` field from Login body
- `license.go`: path `/license`, body `{"license":...}`, response struct matches actual server shape
- `auditlog.go`: reverted to flat `[]AuditEvent` return; `AuditLogResponse` envelope type removed; field tags use underscores

**Python SDK** (4 files modified):
- `models/widgets.py`: `IPVideo.host_id` aliased to `"host-id"` (RDP already had aliases)
- `models/server.py`: `LicenseInfo` rewritten to match server shape with underscored fields
- `models/audit.py`: `AuditLogEntry` rewritten — `author_id`/`created_at`/`target_id`/`target_type`/`ip_address` fields
- `resources/server.py`: install_license path + body fixed; `user_id` reverted from `str` to `int`

**TypeScript SDK** (3 source files + 1 conventions file modified):
- `types/widget.ts`: IpVideo + RdpConnection use bracketed hyphenated keys
- `types/server.ts`: License + AuditEntry + InstallLicenseRequest match server shape
- `types/user.ts`: User.id is `number` (was `Uuid`)
- `docs/conventions/typescript.md` Amendments: NotFoundError + camelCase deferral + zod deferral + audit-log flat-array decision

**Spec docs:**
- `docs/api-reference/VERIFIED-CORRECTIONS.md` created (full delta vs as-extracted spec, with curl evidence and GitLab issue link)
- `docs/api-reference/changelog.md` gains a 2026-05-18 verification addendum pointer

---

## Process notes

- **Failure protocol triggers:** still zero. All fix agents produced surgical, on-spec output first try.
- **`.secrets` file:** present at repo root with dev-mtcs creds, gitignored.
- **Convention defaults:** all locked as confirmed. NotFoundError formalised as the only convention amendment.

---

## Phase 4a outcome — DONE

24 example projects shipped (8 × 3 languages). Total: ~4,500 lines of example code across 110 files.

| Lang | Examples | LOC | Toolchain status |
|---|---|---|---|
| Go | 8 mirrored modules under `go/examples/core/` | 1,692 | `go build` + `go vet` clean |
| Python | 8 uv workspace members under `python/examples/core/` | 1,404 | `uv sync` + `ruff check` clean |
| TypeScript | 8 pnpm workspace members under `typescript/examples/core/` | 1,370 | SDK builds (esm+cjs+dts); examples `tsc --noEmit` clean |

A code-review pass flagged 13 findings (4 Critical + 9 Important). 11 were fixed in commit `0770701`:
- **C1** TS SDK wire-shape types (21 files rewritten by a dispatched fix agent — hyphenated keys replaced with verified underscored keys; specific hyphenated exceptions kept)
- **C2/C3** Python + TS LLM watchers' snapshot-vs-delta dedup
- **C4 + I9** Env-var unification across all 3 languages to `CANVUS_API_URL`
- **I1** Dropped redundant `widget_type` from Go PATCH bodies
- **I3** Python example 02 control-flow cleanup
- **I4** Python User/Group ID types changed from str → int
- **I5** Python 04 redundant env check removal
- **I8** TS example 08 arrow-wrap for SDK method references

Live-server verification cred file: `.secrets` (gitignored). Test canvas reachable at `e36c286c-d447-4b99-a956-678dc118c774` on `dev-mtcs.multitaction.com`.

---

## Phase 4b outcome — DONE

Cross-SDK parity sweep. Scope expanded mid-phase: the audit surfaced **223 gaps** versus the plan's anticipated ~5 helpers + Go Subscribe. Jaypaul selected full-parity (option 1) at the scope gate; all 223 closed across 4 commits + 1 review-driven fix commit.

| Lang | Work items | Result |
|---|---|---|
| Go | §4.1 #1-16 — 36 Subscribe helpers (every streamable endpoint) + 5 extras modules (geometry, filters, zones=WidgetZoneManager, batch_widgets, search=CrossCanvasSearch) + FromEnv + GetCurrentUser + WithRequestIDFunc + WithConnectTimeout + ErrUnsupportedOperation | `go build` / `go vet` / `go test` clean. All 8 examples build clean. |
| Python | §4.2 #1-23 — subscribe completion (~28 endpoints), 10 extras modules (geometry, filters, widget_operations, search, export, import_, batch, color, warnings + __init__), generic widget CRUD (create_any/update_any/delete_any/patch_parent_id), trash helpers, color-preset decomposition, server convenience (config_raw, set_video_output_source_by_index, toggle_workspace_*, set_workspace_viewport), get_current_user, ValidationError.issues, subscribe_permissions live-verification warning | ruff clean, 102/102 pytest pass. |
| TypeScript | §4.3 #1-22 — 9 extras modules, retry layer + circuit breaker (5 consecutive failures → open, 30s reset), verifyTls via undici.Agent, requestIdProvider, Session.close(), RateLimitError/ServerError/UnsupportedOperationError, generic widget CRUD (createAny/updateAny/deleteAny/patchParentId), trash, color-preset decomposition, server convenience, currentUser | typecheck/build clean, 82/82 vitest pass. |
| Go examples | Examples 05/06/07 refactored to typed Subscribe; ~60-100 lines shorter each; settle-timer snapshot-drain dedup (default 2s, env-overridable) replaces frame-boundary tracking. READMEs updated. | All build clean. |

A code-review gate (background opus agent) flagged 2 Critical and 4 Important findings against the 4 SDK commits. All 6 fixed in `e50001c`:

- **Critical: Export schema divergence.** TS shipped `manifest.json` with a different schema than Go/Python's `export.json`. Cross-runtime import was broken. TS rewritten to canonical Go schema `{widgets, assets, region}` with flat asset siblings (`image_<id>.jpg`, `pdf_<id>.pdf`, `video_<id>.mp4`). New tests cover schema, asset write, and import round-trip.
- **Critical: `distance_between_widgets` three-way divergence.** Go and Python used `math.Hypot(dx, dy)`; TS used `Math.min`. TS docstring falsely claimed Python parity. Canonical is hypot (cartesian distance); TS fixed; all three SDKs now have explicit docstrings noting the deliberate divergence from the legacy CanvusPythonAPI `min`-based helper. Cross-SDK parity tested with the 3-4-5 triangle case.
- **Important: TS lint regressions (4)** in `transport.ts` — fixed by collapsing redundant Buffer-vs-Uint8Array branches (Buffer extends Uint8Array under Node).
- **Important: Python `subscribe_permissions` missing warning** — added `.. warning::` block matching Go's parity-matrix §5.5 caveat.

Other code-review findings deferred to Phase 4d per the review's recommendation (TS coverage threshold reclaim, full test coverage on smoke-only extras tests, permissions-subscribe live verification, channel-buffer config, createAnyWithAsset for TS).

Live-server verification was NOT re-run in 4b — toolchain coverage + the existing `.secrets` env cred file remain valid for Phase 4c per-item refresh.

---

## Phase 4c outcome — DONE

Per-item refresh of the external Canvus tool/utility/example repos into the monorepo. Original spec listed 11 items; final scope **8 items** after three documented scope reductions.

| Item | Source repo | Destination | Lang | Round |
|---|---|---|---|---|
| translator | `CanvusTranslator` | `go/tools/translator/` | Go | 1 |
| note-mapper | `CanvusNoteMapper` | `go/examples/projects/note-mapper/` | Go | 1 |
| llm-canvas-companion | `CanvusAPI-LLMDemo` (renamed) | `go/examples/projects/llm-canvas-companion/` | Go | 1 |
| db-solver | `Canvus-Server-db-solver` | `go/tools/db-solver/` | Go | 1 |
| ai-personas | `AI-personas` | `go/examples/projects/ai-personas/` | Go | 2 |
| webui (rewrite) | `CanvusWebUI` | `typescript/examples/webui/` | TS (Hono) | 2 |
| cli | `canvus-cli` | `go/cli/` | Go | 3 |
| mcp-server (Python) | `canvus-mcp-server` | `python/tools/mcp-server/` | Python | 3 |

**Scope reductions** (documented during pre-flight + audit + Round 3 dispatch):
- `CanvusMCP` (Go) → `go/tools/mcp-server/` — dropped (no upstream repo, never built; the Python `canvus-mcp-server` covers the MCP use case).
- `Canvus-Local-LLM` (Python) → `python/tools/local-llm/` — dropped (superseded by Go `llm-canvas-companion` + Phase 4a's `python/examples/core/06-llm-integration`; three impls of the same Canvus+LLM idea would be redundant).
- `CanvusPowerToys` (Go) → `go/tools/powertoys/` — **deferred to Phase 5**. The Round 3 implementer agent BLOCKED on architectural decisions: the source is significantly bigger and more wired-up than the audit estimated (1,100-LOC `manager.go` god-organism + local-typed `Widget`/`Location`/`Size` referenced from 10+ files). Two viable paths (pragmatic shim vs full SDK type migration) need a dedicated Phase 5 plan.

Per-language toolchain status (post-Round-3 review fixes):

| Lang | Items | Toolchain |
|---|---|---|
| Go | translator, note-mapper, llm-canvas-companion, db-solver, ai-personas, cli (6) | build / vet / test green per-module; `gofmt -s -l` empty across all 4c items. |
| Python | mcp-server (1) | ruff clean, mypy --strict clean on new code, 42/42 pytest pass. |
| TS | webui (1) | typecheck clean, build clean, 24/24 vitest pass. |

Three per-round review gates ran (after each batch). Each surfaced 1-5 Critical/Important findings; all addressed inline within the round (Round 1: stray binaries, fake tests, dead-code temp file, test colocation; Round 2: orphan rcu.html, silent error swallow in qa/wait.go, RcuStore extraction; Round 3: env-var alignment `CANVUS_URL`→`CANVUS_API_URL`, gofmt sweep, ruff on Python tests).

Final review gate (`af2ee7df`) verdict: ✅ phase complete with deferrals. Zero Critical. 4 Important + 2 Nice-to-have items all destined for Phase 4d (captured below).

Live-server verification was NOT re-run in 4c — toolchain coverage + per-item unit tests + the existing `.secrets` env cred file from Phase 3.t remain valid.

ARCHIVED.md notes pushed to all 8 source repos (commit `chore: archive — refreshed into MT-Canvus-Tools (Phase 4c)` on each repo's default branch). 4 ported READMEs that lacked source-repo back-links updated in the Task 4c.10 sweep.

mcp-server-python deferrals (preserved as structured `MCPToolExecutionError` stubs at runtime, NOT silent failures):
- 13 LLM/correlation/brainstorming/reports tool bodies — depend on the legacy 541-LOC Ollama client with module-scoped globals + aiohttp (conflicts with httpx convention). Public surfaces preserved.
- Custom JWT auth layer (730+523 LOC) — distinct from Canvus API auth.
- SQLite cache + PDF processing pipelines — paired with the LLM port.

cli behavioural diffs from source (all documented in `go/cli/README.md`):
- `widget pin/unpin/copy/move` now route through `UpdateWidget`/`CloneWidget`/clone-then-delete because the legacy `/widgets/{id}/{pin,copy,move}` endpoints don't exist on the canvus-server (phantom endpoints in the legacy CLI).
- `client create/update/delete` hidden + return "Phase 4d gap" — pending API verification.
- `system send-test-email` and `group add-user/remove-user` adjusted to new SDK signatures.

---

## Phase 4d outcome — DONE

Cleanup sweep + MCP LLM port. 32 commits between `1820a01` (post-Phase 4c) and HEAD. Executed in 4 rounds + plan commit + final review.

### Per-item delivery (17 items)

| # | Item | Status | Commit(s) |
|---|---|---|---|
| 1 | Go SDK dead code removal (`setToken`, `warnAlways`) | ✅ shipped | `b478961` |
| 2 | TS SDK README inaccuracies sweep | ✅ shipped (no-change verified) | `25ffede` |
| 3 | Spec doc back-applies from VERIFIED-CORRECTIONS (`server.md`, `auth.md`) | ✅ shipped | `25ffede` |
| 4 | Reclaim TS coverage thresholds | ✅ shipped (167/167 pass, 70.92/74.49/64.11/70.92) | `b26096d` |
| 5 | Live-verify `SubscribeCanvasPermissions`/`SubscribeFolderPermissions` (Go) + `canvases.subscribe_permissions` (Python) | ✅ shipped (Canvas PASS, Folder PASS; warnings dropped) | `1807060` |
| 6 | Asset roundtrip integration tests (TS + Python) | ✅ shipped | `cf874e9` |
| 7 | Reduce Python mypy --strict baseline below 38 | ✅ shipped (39 → 0 at workspace level) | `b5cb116`, `9173fa8`, `bcdd7d5` |
| 8 | Reduce TS lint baseline below 64 | ✅ shipped (64 → 26; remaining 26 are `UserId`/`GroupId` template-literal floor) | `798ee09` |
| 9 | `WithSubscribeBuffer` / `subscribe_buffer` option across Go/Python/TS | ✅ shipped (incl. validation + env wiring) | `e1dc39d`, `ba3759d` |
| 10 | TS `createAnyWithAsset` for symmetry with Go | ✅ shipped (Option B: documented asymmetry in `docs/conventions/typescript.md`) | `636da9f` |
| 11 | CloneWidget per-type wrappers | ⏸ no-op (parity-matrix §5.2 recommends NOT adding sugar — design decision retained) | n/a |
| 12 | note-mapper Gemini dep migration to `google.golang.org/genai` | ✅ shipped | `4425fd2` |
| 13 | Shared Gemini helper extraction (`go/internal/llm/gemini.go`) | ✅ shipped | `1bee991` |
| 14 | Go SDK `WithVerifyTLS(bool)` option + remove 4 hand-rolled insecure clients | ✅ shipped (incl. `WithAPIKey` split: auth ≠ transport) | `ad2ac3f`, `ba3759d` |
| 15 | ai-personas `qa/wait.go` Subscribe migration (poll → `SubscribeNote`) | ✅ shipped (+ `wait_test.go` covering 4 scenarios) | `76e995d`, `96b7699` |
| 16 | mcp-server LLM tools port (13+ tools across llm/brainstorming/correlation/reports + SQLite cache + PDF pipeline + Ollama httpx client + Settings extension) | ✅ shipped (16 tools real, mypy --strict 0, ruff clean, 44/44 pytest pass) | `f6e5327`, `761c1a6`, `c3b626b`, `b757ec5`, `fe8e162`, `224fad1`, `e0c72ec`, `c7714b8`, `e60b75e`, `ec436a2`, `5ec29ae` |
| 17 | mcp-server `/users/current` escape hatch (Python SDK `users.current()`) | ✅ shipped | `0f9b149` |

**16 of 17 items shipped; 1 is a documented no-op design decision.**

### Round-by-round review-gate summary

- **Round A** (7 commits): spec ✅ / quality ⚠️ → fixed inline (`96b7699` added `wait_test.go`).
- **Round B** (5 commits): spec ✅ / quality ⚠️ → fixed inline (`ba3759d` resolved Critical `WithAPIKey` TLS-split + Python stream error swallow + TS env-wire).
- **Round C** (7 commits): clean (1 minor mypy regression surfaced + fixed in `bcdd7d5`).
- **Round D** (11 commits): live-verify D1 PASS; D2 LLM port quality review surfaced 1 silent-fallback in `correlation.py` → fixed inline (`5ec29ae`).
- **Final review (this gate)**: ZERO Critical findings. Phase 4d **COMPLETE**.

### D1 live-verify result (2026-05-19, dev-mtcs.multitaction.com)

- **Canvas permissions subscribe** (Go + Python): PASS. Initial snapshot followed by change event on `POST /canvases/{id}/permissions`.
- **Folder permissions subscribe** (Go): PASS. Initial snapshot followed by change event on `POST /canvas-folders/{id}/permissions`.
- **Python `FoldersResource.subscribe_permissions`**: GAP — Python SDK does not yet expose this on `FoldersResource` (Canvas-level helper exists). Captured as Phase 5 follow-up parity item; not a verification failure.
- "Not yet live-verified" warnings dropped from parity-matrix §5.5.

### Phase 4d carry-over to Phase 5

Deferred items surfaced during Phase 4d, captured below for Phase 5 planning:

- **Python `FoldersResource.subscribe_permissions`** — add helper to mirror `CanvasesResource.subscribe_permissions` for folder-permission events. One-method parity fill.
- **TS lint floor (26 errors)** — `UserId`/`GroupId` template-literal interpolation in `resources/users.ts`. Requires either (a) widening the branded-type interpolation in `@typescript-eslint/restrict-template-expressions` config, or (b) renaming the branded types to `string & {…}` aliases. Design decision deferred.
- **Python `client.py:105` unused-ignore** — `# type: ignore[call-arg]` on `Settings()` no longer needed at sdk-only mypy run (workspace-level mypy is clean). Drop comment in next sweep.
- **CloneWidget per-type wrappers** — retained as a "not needed unless callers report friction" deferral; surface again only on user request.

---

## Deferred to Phase 7 (CI/CD)

Single agent: GitHub Actions workflows. Per-language matrix jobs (lint, test, build). Optional spec-drift detector comparing `docs/api-reference/` to `mt-restapi-client` on a schedule.

All Phase 4d and Phase 5 carry-over items are now resolved:

- ✅ **Python `FoldersResource.subscribe_permissions`** — shipped Phase 5 carry-over (`cd49f54`)
- ✅ **TS lint floor** — `UserId`/`GroupId` narrowed to `number` + `allowNumber: true` rule (`cdde5ab`); floor is now 0
- **Python `client.py:105` unused-ignore** — `# type: ignore[call-arg]` on `Settings()` — verify still needed with `uv run mypy --strict sdk/src`; drop in Phase 7 sweep if resolved
- **CloneWidget per-type wrappers** — retained as "not needed unless callers report friction" design decision; revisit only on user request
