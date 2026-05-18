# MT-Canvus-Tools Consolidation Status

**As of:** 2026-05-18 (mid-morning, post-verification)
**Phases complete:** 0, 1, 2, 3 (+ Tier-1/Tier-2 verification & patch pass)
**Next phase:** 4 (Core Examples)
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
| 3.t | Live-server verification + Tier-1/Tier-2 fixes + `VERIFIED-CORRECTIONS.md` | (this commit) |

**Repo:** `github.com/jaypaulb/MT-Canvus-Tools` (private)

### Coverage snapshot (verified endpoint-by-endpoint)

| SDK | Implementable coverage | Toolchain status |
|---|---|---|
| Go | **147 / 147** | `go build` + `go vet` + unit tests clean |
| Python | **147 / 147** | Not run (no toolchain check) |
| TypeScript | **147 / 147** | Not run (no toolchain check) |

Three endpoints deliberately omitted across all SDKs per changelog §1 + §2.

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

## Ready for Phase 4

The Phase 4 plan needs to cover:
1. **Core Examples (4a):** 8 canonical examples × 3 languages = 24 small projects. These should be written using the freshly verified SDKs and serve as the parity contract for the SDKs themselves.
2. **Cross-SDK parity sweep:** audit each SDK against the others; port any helpers / utilities that exist in one but not all. Likely lives in a `sdk-extras` subpackage per language to keep core SDKs lean.
3. **Per-Item Refresh (4b):** 11 existing utility/tool repos refreshed against the new SDKs.
4. **Phase 4 should also remove the Go dead code** flagged in Tier 3 #10 after grep confirms no callers.

Tell me when to write Plan 4 and how aggressive you want me to be (foreground/background, batch size, etc.).
