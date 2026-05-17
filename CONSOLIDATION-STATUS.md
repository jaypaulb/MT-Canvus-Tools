# MT-Canvus-Tools Consolidation Status

**As of:** 2026-05-18 (morning)
**Phases complete:** 0, 1, 2, 3
**Next phase:** 4 (Core Examples) — plan to be written based on your decisions below

---

## What shipped overnight

| Phase | Output | Commit |
|---|---|---|
| 0 | Repo bootstrap (private, MIT) on GitHub | `5f1eaf4`, `7616ea3` |
| 1 | Canonical API reference: 150 endpoints, 82 examples, 5 changelog entries | `a2dc856` |
| 2 | SDK coverage audit + 3 locked conventions docs | `31e42a6` |
| 2 plan | Phase 3 SDK migration plan | `894a6cb` |
| 3.1 | Go SDK migrated (build ✓, vet ✓, unit tests pass) | `a0097d3` |
| 3.2 | Python SDK migrated (src-layout, restructured) | `2b04093` |
| 3.3 | TypeScript SDK greenfield | `568e746` |
| 3.v | Coverage verification + Go `CopyFolderPatch` fix | (this commit) |

**Repo:** `github.com/jaypaulb/MT-Canvus-Tools` (private)
**Total size:** ~145 files, ~26,000 lines of docs + code

### Coverage snapshot (verified endpoint-by-endpoint)

| SDK | Implementable coverage | Files | Lines | Toolchain status |
|---|---|---|---|---|
| Go | **147 / 147** | 42 | 5,684 | `go build` + `go vet` + unit tests clean |
| Python | **147 / 147** | 23 | 3,895 | Not run (no toolchain check) |
| TypeScript | **147 / 147** | 25 | 3,430 | Not run (no toolchain check) |

**Three endpoints are deliberately omitted across all SDKs** (per changelog §1 + §2, identical reason in each):
- `POST /api/v1/canvases/{id}/widgets/clone` — replaced by per-type `clone()` helpers (changelog §1)
- `POST /api/v1/canvases/{id}/ip-videos` — server rejects (changelog §2)
- `POST /api/v1/canvases/{id}/rdp-connections` — server rejects (changelog §2)

Spec total is 150 endpoints; implementable surface after these omissions is 147.

**Reconciliation against initial self-reports:** Go agent over-reported (claimed 150/150; was 146/147, missed `PATCH /canvas-folders/{id}/copy` — fixed this morning). Python agent under-reported (claimed ~140; was actually 147/147 — methodology artifact, not a gap). TS agent under-reported (claimed 148/150; was actually 147/147 — double-counted a deprecated endpoint). Full endpoint-by-endpoint audit at `docs/api-reference/coverage-verification.md`.

---

## Decisions you need to make this morning

Ordered by impact. All are documented in the per-SDK MIGRATION-NOTES.md / IMPLEMENTATION-NOTES.md.

### Tier 1: Need live-server answers before Phase 4

1. **RDP / IP Video field naming (changelog §3, §5).** C++ source uses hyphens (`host-id`), public docs use underscores (`host_id`). Go SDK followed C++ (hyphens). Python SDK followed docs. TS SDK followed docs but flagged uncertainty. **A 5-minute `curl` against dev-mtcs.multitaction.com will resolve all three.**

2. **Login / InstallLicense / SendTestEmail body shapes.** Go agent applied "defensive double-keying" (sends both spec and legacy keys); if the server rejects unknown fields with 400, this breaks. Python and TS sent spec-only. **Same `curl` session can verify.**

3. **Streaming wire format.** Python and TS assume NDJSON. If the server actually emits SSE, the streaming layer needs swapping for `httpx-sse` / `eventsource-parser`. Affects all three SDKs.

4. **ServerConfig wire shape.** Spec describes a flat element array; observed servers return nested objects. Go + Python both ship defensive decoders (handle both). TS types as open `{[key: string]: unknown}`. Tighten with real responses.

### Tier 2: Architectural calls made autonomously — easy to reverse now, hard later

5. **TS preserves wire shape (no camelCase mapping).** `canvas["canvas-name"]` not `canvas.canvasName`. Easy to add a mapper now; painful once 8 example apps and the WebUI bake in the snake/hyphen keys.

6. **TS: zod for config only, not response validation.** v0.2 can add response validation. Deferring is the conservative call.

7. **TS: `NotFoundError` added as 5th error kind** beyond the conventions doc's four. 404 is too often normal control flow to collapse into generic `APIError`. Update `docs/conventions/typescript.md` Amendments if you concur.

8. **Go: `ListAuditEvents` return type changed** from `[]AuditEvent` to `*AuditLogResponse` (envelope). Includes flat-array fallback. Breaking change for downstream callers — they must use `.Events`.

### Tier 3: Scope omissions to confirm

9. **Python: legacy helpers not migrated** (`geometry.py`, `search.py`, `filters.py`, `export.py`, `widget_operations.py`, circular-parenting guard). Agent classified them as out-of-scope for SDK transport. If you want them, recommend a `python/sdk-extras/` workspace member rather than re-bloating the core SDK.

10. **Go: dead code left in place** (`setToken`, `doRequestWithHeaders`, `warnOnce`, `warnAlways` — flagged unused by the linter). Removing now risks breaking a path I can't see; Phase 4 can decide after Chesterton fence inspection.

---

## Process notes

- **Failure protocol triggers:** none. All 3 Phase-1 agents and all 5 Phase-2 agents and all 3 Phase-3 agents produced substantive output on first attempt. No retries, no GH issues opened.
- **Source uncertainty callouts** during Phase 1: none flagged explicitly (agents used `canvus-server` cross-reference per your guidance).
- **Convention defaults:** locked exactly as you confirmed. No overrides applied.
- **`.secrets` file:** present at repo root with dev-mtcs creds, gitignored via `**/.secrets` pattern.

## Phase 4 readiness

The Phase 4 plan (Core Examples: 8 examples × 3 languages = 24 small projects) is NOT yet written. The right time to write it is **after you've made the Tier 1 decisions above** — the example apps will exercise streaming + auth body shapes + RDP/IP fields, so getting those locked first avoids rework.

Recommended sequence:
1. Read this file (5 min)
2. Run live-server `curl` smoke test against dev-mtcs to resolve Tier 1 (15 min)
3. Tell next Claude what you decided
4. Next Claude writes Phase 4 plan + dispatches

The full design spec, all plans, all status files, and all per-SDK NOTES.md docs are at:
- `docs/superpowers/specs/2026-05-17-mt-canvus-tools-consolidation-design.md`
- `docs/superpowers/plans/2026-05-17-mt-canvus-tools-foundation.md`
- `docs/superpowers/plans/2026-05-17-mt-canvus-tools-phase3-sdk-migration.md`
- `go/sdk/MIGRATION-NOTES.md`
- `python/sdk/MIGRATION-NOTES.md`
- `typescript/sdk/IMPLEMENTATION-NOTES.md`
