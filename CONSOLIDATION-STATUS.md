# MT-Canvus-Tools Consolidation Status

**As of:** 2026-05-18 (late evening, post-Phase 4c)
**Phases complete:** 0, 1, 2, 3 (+ verification), 4a (+ review-driven fixes), 4b (+ review-driven fixes), 4c (+ review-driven fixes)
**Next phase:** 4d (cleanup) or 5 (PowerToys port) — pending Jaypaul go-ahead
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

**Repo:** `github.com/jaypaulb/MT-Canvus-Tools` (private)

### Coverage snapshot (verified endpoint-by-endpoint)

| SDK | Endpoint coverage | Subscribe coverage | Extras subpackage | Toolchain status |
|---|---|---|---|---|
| Go | **147 / 147** | **36 / 36** typed helpers | 5 modules (geometry, filters, zones, batch_widgets, search) | build / vet / test clean |
| Python | **147 / 147** | **~28 / 28** typed AsyncIterators | 10 modules + `__init__` (geometry, filters, widget_operations, search, export, import_, batch, color, warnings) | ruff clean, 102/102 pytest pass; mypy strict 39 errors (pre-existing baseline 38) |
| TypeScript | **147 / 147** | **27 / 27** typed async iterators | 9 modules + index (geometry, filters, widgetOperations, search, export, import, batch, color, warnings) | typecheck / build clean, 82/82 vitest pass; lint 64 pre-existing errors (baseline 85) |

Three endpoints deliberately omitted across all SDKs per changelog §1 + §2.
Permissions-subscribe helpers in Go (`SubscribeCanvasPermissions`, `SubscribeFolderPermissions`) and Python (`subscribe_permissions`) ship behind documented "not yet live-verified" warnings (parity-matrix §5.5); deferred to Phase 4d live-verification.

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

## Deferred to Phase 4d (cleanup)

Pre-Phase-4b items:
- Go SDK dead code removal (`setToken` — `doRequestWithHeaders` was Chesterton-fence-resolved during 4b; it IS the backbone of the new subscribeStream primitive) — Tier 3 #10 remainder.
- TS SDK README inaccuracies sweep — six READMEs had "Expected output" sections referencing wire-shape fields that don't exist on the server (cascade from C1).
- Spec doc errors at `docs/api-reference/endpoints/server.md` and `auth.md` (per VERIFIED-CORRECTIONS) — currently captured as a delta in `VERIFIED-CORRECTIONS.md`; can be back-applied after `canvus-server#96` resolves.

Added by Phase 4b review gate:
- **Reclaim TS coverage thresholds** (50/60/40/50 → 60/70/60/60 or higher) by adding focused unit tests for `auth.ts`, `users.ts`, `server.ts` resource error mapping. Current relaxation is the agent's pragmatic call to ship; needs proper coverage in 4d.
- **Live-verification of `SubscribeCanvasPermissions` / `SubscribeFolderPermissions`** against `dev-mtcs.multitaction.com`; remove "not yet verified" warnings if pass, escalate to skipped helpers if fail.
- **Asset roundtrip integration tests** for export/import in TS and Python (current TS tests cover image-write but not the full roundtrip; Go's existing integration test stands).
- **Reduce Python mypy strict baseline below 38** — 4 errors share the same `list[T] annotation vs .list() method` pattern; either rename methods or globally fix via `from builtins import list as _list`.
- **Reduce TS lint baseline below 64** — bulk of remaining errors are `UserId` template-literal interpolation, `request<void>` returns, `ReadonlyArray<T>` syntax, tsconfig-include for tests. ~1 day cleanup.
- **`WithSubscribeBuffer(int)` option in Go** and equivalent in Python/TS for high-throughput consumers — currently hardcoded buffer=4.
- **`createAnyWithAsset(canvasId, payload, blob, contentType)` in TS** for symmetry with Go's `CreateWidget(io.Reader)` — or document the asymmetry permanently in conventions.
- **CloneWidget per-type wrappers** — all three SDKs implement clone as single method with type-param (matches changelog §1 literal text). Not needed unless callers report friction; parity-matrix §5.2 recommends NOT adding sugar.

Added by Phase 4c final-review gate:
- **Align note-mapper Gemini dependency** — migrate `go/examples/projects/note-mapper/go.mod` from deprecated `github.com/google/generative-ai-go v0.19.0` to `google.golang.org/genai v1.34.0`, matching translator + ai-personas.
- **Extract shared Gemini helper** — three-item duplication (translator, ai-personas, note-mapper) satisfies rule-of-three. Target: `go/examples/internal/llm/gemini.go` (or a new `go/sdk/extras/llm/` if promoted SDK-side).
- **Add Go SDK `WithVerifyTLS(bool)` option** — four sites worked around the gap via hand-built insecure `*http.Client` (`go/tools/db-solver/internal/commands/session.go:15`, `go/tools/db-solver/internal/commands/lookup_hash.go:242`, `go/cli/internal/session/session.go:42`, `go/cli/internal/commands/login.go:111`). Add a first-class option to `go/sdk/canvus/options.go`; remove the workarounds.
- **ai-personas `qa/wait.go` Subscribe migration** — replace the 500ms `GetNote` poll loop with `SubscribeWidget` (Phase 4b §4.1 #7). Low priority — currently bounded by 10s error tolerance + documented as deferral.
- **mcp-server-python LLM tools port** — 13 deferred tools across `python/tools/mcp-server/src/canvus_mcp_server/mcp_tools/{llm,brainstorming,correlation,reports}.py`. Re-implement the legacy 541-LOC Ollama client against monorepo conventions: `httpx` (not `aiohttp`), settings constructor (not module globals), structlog, mypy --strict clean. Also: SQLite cache + PDF processing pipelines paired with the LLM port.
- **mcp-server-python `/users/current` escape hatch** — `python/tools/mcp-server/src/canvus_mcp_server/mcp_tools/users.py:116` reaches into SDK private `client._transport.request()`. Add typed `client.users.current()` to the Python SDK (one-line mirror of `canvases.py:92`) and update the consumer.
- **Document cli `FromEnv` deviation** — `go/cli/internal/config/config.go` uses viper flag/file/env precedence instead of the SDK's `FromEnv`. Per `docs/conventions/go.md §11` this should be added as a dated Amendment.

---

## Deferred to Phase 5 (CanvusPowerToys port)

Standalone phase, separate plan to be written when Jaypaul gives the go-ahead. Decisions already locked in (in `docs/superpowers/plans/2026-05-18-mt-canvus-tools-phase4c-per-item-refresh.md` and conversation 2026-05-18):

- **Opt-in TLS insecure mode** via BOTH `--insecure-tls` flag AND `CANVUS_INSECURE_TLS` env var. Will use the new SDK `WithVerifyTLS(bool)` option once Phase 4d ships it.
- **Split client architecture** — SDK-backed `APIClient` for canvus-server calls + separate `RCUClient` struct targeting `http://127.0.0.1:<webui-port>` with `WEBUI_PWD` Bearer auth (matching the Phase 4c Round 2 webui server's auth gate at `typescript/examples/webui/src/routes/rcu.ts`).
- **Decompose `manager.go`** (1,100-LOC god-organism) and local-typed `Widget`/`Location`/`Size` (referenced from 10+ files) — the work that caused the Round 3 implementer to BLOCK rather than half-implement.
- **RCU endpoint reality check** — `docs/api-reference/per-item-refresh-audit.md:423` open question. The webui Round 2 implementer noted source `rcu.html` was actually "Remote Content Upload" (user-facing upload form), NOT an admin API. Phase 5 confirms whether `/api/v1/canvases/{id}/rcu/*` is a real custom server feature before porting both items' RCU handlers.
