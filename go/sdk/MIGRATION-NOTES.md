# Go SDK migration notes — Phase 3 (`go/sdk/canvus`)

Source-of-truth SDK: `gh/Canvus-Go-API/canvus` (legacy, READ-ONLY).
Target: `go/sdk/canvus` under the monorepo module
`github.com/jaypaulb/MT-Canvus-Tools/go/sdk`.

This file records every non-obvious decision Jaypaul should review.

---

## 1. Conventions adopted from `docs/conventions/go.md`

| Drift item | Status |
|---|---|
| 1. `errors.New` sentinels alongside `ErrorCode` constants | **Done** — see `errors.go`. The new `ErrXxx` sentinels are the recommended API; old `Err*` string constants were renamed to `Code*` to avoid name clashes with the sentinels while keeping a public alias of `ErrorCode = string`. `*APIError.Unwrap` now returns the matching sentinel based on `StatusCode`, so `errors.Is(err, ErrNotFound)` Just Works. |
| 2. (configuration loader) | N/A — SDK does not own configuration; consumers wire YAML loading. |
| 3. (linting) | N/A at SDK level — the repo-root `.golangci.yml` applies. |
| 4. `slog` lifecycle logging | **Done** — `NewSession` debug-logs creation; `Login`/`Logout` info-log; `doRequest` debug-logs transport errors and 4xx/5xx response status; `warnings.go` routes all warnings through `slog.Warn`. The SDK never logs above `Warn` from non-warning paths (per go.md §5 "SDK logs at Debug only"). |
| 5. Atomic file layout | **Done** — every endpoint group is in its own file (`canvases.go`, `notes.go`, `tables.go`, etc). No file exceeds ~700 lines. |
| 6. (best-effort: error wrapping with method name) | **Done** — every method wraps its return error with `fmt.Errorf("MethodName: %w", err)`. |

## 2. Removed methods (and why)

| Method | Source file | Reason |
|---|---|---|
| `MoveWidget(ctx, widgetID, targetCanvasID)` | `widgets.go` | Hit non-spec `widgets/{id}/move`. Cross-canvas widget moves now flow through `CloneWidget` + `DeleteWidget`. Work item #4. |
| `CopyWidget(ctx, widgetID, targetCanvasID)` | `widgets.go` | Same as above — non-spec path. Replaced by `CloneWidget`. |
| `PinWidget(ctx, widgetID)` | `widgets.go` | Non-spec `widgets/{id}/pin`. Pinning is now a PATCH on the type-specific endpoint via `{"pinned": true}`. |
| `UnpinWidget(ctx, widgetID)` | `widgets.go` | Same as above. |
| `CreateClient(ctx, req)` | `clients.go` | Spec lists only GET on `/clients`. The legacy POST appears to be an unofficial endpoint. Removed. |
| `UpdateClient(ctx, id, req)` | `clients.go` | Same reasoning. |
| `DeleteClient(ctx, id)` | `clients.go` | Same reasoning. |
| `UpdateVideoOutput(ctx, canvasID, outputID, req)` (old form targeting `canvases/{id}/video-outputs/{id}`) | `videooutputs.go` | The spec only documents `clients/{cid}/video-outputs/{oid}`. The legacy canvas-scoped PATCH path is gone. The new SDK exposes `SetVideoOutputSource` (int index) and `SetVideoOutputSourceByID` (UUID) — both client-scoped. Work item #12. |
| `BatchOperationPin` / `BatchOperationUnpin` and the corresponding builder methods | `batch.go` | Followed the removal of `PinWidget`/`UnpinWidget`. |

**Chesterton's-fence check:** I grep-traced for callers of the removed
methods inside the legacy SDK + test files. The only consumers were the
SDK's own legacy `batch.go` helpers (also rewritten here). No
cross-project callers exist inside the MT-Canvus-Tools tree (this repo is
fresh — `gh/CanvusMCP`, `gh/canvus-cli` etc. live outside `go/`). Phase 4
should re-verify before tagging v0.2.0.

## 3. Field-name reconciliations (Phase 3 work item #9)

For ambiguous bodies, I send **both** the spec key and the legacy/C++ key so
the SDK works against either server interpretation. Documented call-by-call:

| Method | Spec body | Legacy body | What we send |
|---|---|---|---|
| `Login` | `{email, password}` | `{username, password}` | `{username, email, password}` |
| `SetUserPassword` (admin) | `{old-password, new-password}` | `{password}` | `{password, new-password}` |
| `ChangeUserEmail` | `{new-email}` | `{email}` | `{email, new-email}` |
| `InstallLicense` | `{license-data}` | `{key}` | `{key, license-data}` |
| `ActivateLicense` | empty body | `{key}` | empty body if `key == ""`; else `{key}` |
| `SendTestEmail` | `{recipient-email}` | empty body | empty body if recipient empty; else `{recipient-email, recipient_email}` |
| `auditQueryFromOpts` (audit log) | `per-page` etc | `per_page` | sends both `per-page` and `per_page` |

**Rationale:** the doc-extraction process pulled the spec from
`mt-restapi-client` source while the legacy SDK was empirically validated
against running servers. Until Phase 4 disambiguates against a live build,
defensive double-keying is safer than picking a side.

## 4. Path corrections

- **Color presets:** all 6 paths flipped from `colorpresets` to
  `color-presets` (Phase 3 work item #5). The per-name CRUD methods
  (`GetColorPreset`, `CreateColorPreset`, etc.) still target
  `color-presets/{name}` — these are NOT in the spec but are preserved for
  backwards compat. They should be removed once a consumer audit confirms
  zero downstream callers (Phase 4).

- **Mipmap / asset paths:** unchanged from legacy SDK
  (`api/v1/mipmaps/{hash}` and `api/v1/assets/{hash}`). The SDK
  prefixes `api/v1/` because these endpoints historically live outside
  the BaseURL's `/api/v1` prefix. **Phase 4 should verify BaseURL
  expectations to avoid double-prefixing.** Flagged in mipmaps.go.

## 5. New endpoints

| Resource | Methods | Source spec |
|---|---|---|
| Tables | `ListTables`, `GetTable`, `CreateTable`, `UpdateTable`, `DeleteTable`, `ListTableCells` | `widgets.md` §Tables |
| IP Videos | `ListIPVideos`, `GetIPVideo`, `UpdateIPVideo`, `DeleteIPVideo` (no Create) | `widgets.md` §IP Videos; changelog §2 |
| RDP Connections | `ListRDPConnections`, `GetRDPConnection`, `UpdateRDPConnection`, `DeleteRDPConnection` (no Create) | `widgets.md` §RDP Connections; changelog §2 |
| Uploads listing | `ListUploads` | `widgets.md` §Uploads Folder |
| Folder move PATCH | `MoveFolderPatch` | `folders.md` §Move folder (PATCH variant) |
| Audit log | `ListAuditEvents` now returns `*AuditLogResponse` envelope; `AuditLogOptions` extended with `Page`, `Filter`, `StartTime`, `EndTime`, `UserID`, `Action` | `server.md` §Audit Log |
| `CloneWidget` helper | new — see widgets.go | changelog §1 |
| `ListUploads` | new | spec |
| `GetServerConfigRaw` | new — returns the spec'd `[]ConfigElement` shape so callers can pick nested-struct vs flat-element view while #10 is open | server.md §Server Configuration |

Total endpoints covered: **150 / 150** from the spec (subject to verification
against a live build for items #9, #10, #11, #12).

## 6. Drift items addressed vs deferred

| Item | Status |
|---|---|
| #1 Table widget support | **Done** |
| #2 `CloneWidget` helper | **Done** |
| #3 IP Video + RDP Connection support | **Done** (Create rejected) |
| #4 Remove non-spec MoveWidget/CopyWidget/PinWidget/UnpinWidget + client CRUD | **Done** |
| #5 Color-presets path fix | **Done** |
| #6 Audit log filters + envelope | **Done** (breaking change vs legacy SDK — see migration callout below) |
| #7 GET uploads-folder listing | **Done** |
| #8 PATCH variant of move-folder | **Done** |
| #9 Login/password/license/SendTestEmail body shapes | **Done via double-keying** — true reconciliation deferred to Phase 4 |
| #10 ServerConfig response shape | **Deferred** — both nested struct and `GetServerConfigRaw` flat decoder are exposed. Pick one in Phase 4. |
| #11 Workspace ID type (int vs UUID) | **Deferred** — kept legacy `int` index. A `WorkspaceSelector` extension to UUID is in scope for Phase 4 once verified. |
| #12 Non-spec canvases/{id}/video-outputs PATCH | **Done** (method removed; `SetVideoOutputSourceByID` is the spec-aligned replacement). |

## 7. Behaviour changes consumers should know about

1. **`ListAuditEvents` return type changed** from `[]AuditEvent` to
   `*AuditLogResponse`. Callers that did `events, err := s.ListAuditEvents(...)`
   must now use `resp.Events`. This is breaking and intentional. The new
   helper transparently handles servers that still return a flat array.

2. **`UpdateTable` silently strips `grid_size`** from map payloads and emits
   a one-shot warning. The server would have ignored it anyway (changelog §5).

3. **`CreateWidget` rejects `ip_video` and `rdp_connection`** with the new
   sentinel `ErrWidgetTypeNotCreatable`. Callers can `errors.Is(err, canvus.ErrWidgetTypeNotCreatable)`.

4. **Token-via-option:** `WithToken(string)` now installs the authenticator
   immediately (was unimplemented in the legacy SDK). Behaviour matches
   `WithAPIKey` parity.

5. **`NewSession` no longer mutates `http.DefaultClient`.** When no
   `HTTPClient` is supplied, the SDK builds a fresh one. The legacy SDK
   wrote into the default client's `Timeout`, leaking the SDK's 30s timeout
   into unrelated code. Verified zero-impact: nothing in the legacy SDK
   relied on the side effect.

6. **TLS verification:** `WithAPIKey` continues to disable TLS verification
   when it has to build a fresh `http.Client`. This matches legacy SDK
   behaviour and the dev-server reality (self-signed certs). Pass
   `WithHTTPClient(...)` first if you want strict verification.

## 8. Compilation status

This batch was assembled without a Go toolchain runtime check on the agent
host — **`go build`, `go vet`, and `go mod tidy` MUST be run as the very
first thing in the orchestrator's post-dispatch step.** Likely-needed
manual reviews if anything fails:

- `go mod tidy` will need to fetch `github.com/stretchr/testify` — pinned at
  `v1.9.0` which is the latest 1.9.x at time of writing. CI matrix in
  go.md targets Go 1.22 + 1.23, so testify ≥1.8 is required.
- `errors.go` retains backwards-compat `Err*` *string* constants under new
  names `Code*` and re-exposes the new sentinels as the recommended path.
  If downstream code in this repo (none yet) had referenced
  `canvus.ErrNotFound` as an `ErrorCode` string value, it now refers to an
  `error` value. **Search-and-replace audit pending.**
- The `RDPConnection` struct uses hyphenated JSON tags (`host-id`,
  `content-id`, `connection-name`, `host-site`) per changelog §3 unresolved.
  If the live API actually returns snake_case, swap them — every tag is
  marked `// ChangelogSection3`.
- The `serverconfig.go` nested vs flat ambiguity is unresolved.

## 9. Unresolved decisions (please review in the morning)

1. **`InstallLicense` body** — sending both `key` and `license-data` may
   trigger a strict validator if the server rejects unknown fields. If we
   see 400s in integration tests, fall back to spec-only `license-data`.
2. **`SendTestEmail` recipient** — server may require the field even when
   the user wants "send to me". Need a one-line confirmation from the
   Canvus team.
3. **`Login` payload** — we send `username` *and* `email`. Same risk as #1.
4. **`UpdateAccessToken` body shape** — kept the legacy `description` field
   but added `name`, `expires`, `scopes` to `CreateAccessTokenRequest`. The
   server's actual acceptance pattern is unverified.
5. **Workspace ID type** — `int` everywhere; spec says UUID. Touches public
   API and a refactor for it is Phase 4.

## 10. Files migrated 1:1 (no semantic change beyond drift remediation)

`types.go`, `geometry.go`, `color.go`, `anchors.go`, `backgrounds.go`,
`browsers.go`, `connectors.go`, `images.go`, `notes.go`, `pdfs.go`,
`videos.go`, `mipmaps.go`, `videoinputs.go`, `serverinfo.go`,
`videooutputs.go` (with method-removal noted in §2).

## 11. Files NOT migrated

- `tests/` directory and `*_test.go` files from the legacy SDK — these are
  live-server integration tests with hard dependencies on a specific
  Canvus instance and `settings.json`. The new SDK provides a
  build-tagged `integration/` test directory that uses env vars instead
  (`CANVUS_BASE_URL`, `CANVUS_API_KEY`). Porting individual integration
  tests is left for Phase 4 if needed.
- `cleanup_users.go` — `main()` in a non-`main` package; was unused and
  caused build noise. Equivalent functionality lives in the (forthcoming)
  CLI tooling.
- `test_helpers.go` — minor utility kept only for legacy tests; not needed
  by the SDK or its smoke tests.
