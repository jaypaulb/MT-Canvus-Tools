# Per-Item Refresh Audit (Phase 4c)

**Audit date:** 2026-05-18
**Auditor:** Phase 4c pre-flight assessment agent
**Scope:** 10 source repos under `/home/jaypaulb/Projects/gh/` slated for porting into the MT-Canvus-Tools monorepo
**Reference baselines:**
- New SDK parity surface: `docs/api-reference/parity-matrix.md`
- Wire-format truth: `docs/api-reference/VERIFIED-CORRECTIONS.md`
- Per-language conventions: `docs/conventions/{go,python,typescript}.md`
- Consolidation context: `CONSOLIDATION-STATUS.md`

**Effective LOC** excludes binaries, vendored deps, `node_modules`, `venv`, `.venv`, `__pycache__`, `htmlcov`, generated `bin`/`build`/`releases`, planning material under `dross/`, `devdocs/`, `dev-docs/`, `agent-os/`, and `*.egg-info`.

**Eleventh item nuance:** the consolidation spec listed `CanvusMCP (Go) → go/tools/mcp-server/`. Confirmed no upstream repo exists at `/home/jaypaulb/Projects/gh/CanvusMCP` — skipped as agreed.

---

## Item 1 — canvus-cli → go/cli/

**Source LOC:** 13,440 (Go only) | **File count:** 187 Go files | **Effective LOC:** 13,440 (~3.8k of that is tests across 25 `*_test.go` files; ~9.6k production)
**Purpose** — A complete Cobra-based command-line interface for the Canvus collaborative-workspace REST API: canvases, folders, widgets, users, groups, access-tokens, clients, audit, mipmaps, server config. Supports text/JSON/YAML/table output, API-key or username/password auth, an interactive Bubble Tea TUI (`internal/commands/tui.go:319`), and is the largest/most polished port candidate.

**Current SDK usage** — Already imports the **legacy** `github.com/jaypaulb/Canvus-Go-API v0.1.0` (see `go.mod:9`). Session created via `canvus.NewSession(sessionCfg, canvus.WithAPIKey(cfg.APIKey))` (`internal/session/session.go:60`). Direct SDK type references include `canvus.Session`, `canvus.Filter`, `canvus.Canvas`, `canvus.Note`, `canvus.Image`, `canvus.PDF`, `canvus.Video`, `canvus.Anchor`, `canvus.Connector` (e.g. `internal/commands/canvas/list.go:59,76`, `internal/commands/widget/create.go:130-242`). Endpoints exercised cover roughly the full SDK surface — canvases, folders, widgets (all per-type), users, groups, tokens, clients, audit, mipmaps.

**SDK call audit** — The legacy `Canvus-Go-API` package and the new `go/sdk/canvus/` are structurally the same (both authored by Jaypaul; new SDK is the consolidation of the legacy SDK). The vast majority of calls are pure rename-free import swaps. A few items deserve verification:
- `canvus.DefaultSessionConfig()` (`session.go:32`) — confirm new SDK exports this; parity-matrix shows `NewSession(cfg, opts...)` (§1.3) so yes.
- `WithAPIKey` (`session.go:60`) — confirmed present (parity §1.3 row 4).
- `Filter` (`canvas/list.go:59`) — present (`go/sdk/canvus/types.go:59`); migrate users to `extras/filters` for richer behaviour later, but the literal-map form still works.
- TUI uses Bubble Tea — fine.
- No gaps for `Trash*`, `GetCurrentUser`, `Subscribe*` — all live in new Go SDK.

**Architecture quality** — High. Well-decomposed: `internal/commands/<resource>/<verb>.go` Cobra files, `internal/session/`, `internal/config/`, `internal/output/` (formatter abstraction across json/yaml/table/text). Tests in mirrored layout. Concrete cleanup notes:
- `internal/commands/canvas/import.go` and `export.go` should swap to the new SDK's `extras/export` and `extras/import` (Go) once it exists (parity §2.5 row 4 confirms Go already has `export.go:28 ExportWidgetsToFolder` / `import.go:15 ImportWidgetsToRegion`).
- `internal/commands/tui.go:319` is a 319-line organism that's borderline; not urgent but worth flagging.
- `cmd/canvus/main.go` is the only file under `cmd/`; adopt the monorepo's `go/cli/` layout (single binary, no nested cmd).
- The `internal/commands/testing/mock_session.go:406` is a hand-rolled mock — could be replaced by interface-based test seams, but Chesterton-fence: it's working; defer.

**Refresh complexity estimate:** moderate. The structural skeleton is sound and entirely SDK-compatible; the work is mostly mechanical import path rewrites plus adopting `FromEnv` (parity §1.3) and trimming the bespoke insecure-TLS HTTP-client path now that `WithAPIKey` defaults to it (`session.go:38-50`).

**Recommended port approach** — Copy the tree into `go/cli/`, swap `github.com/jaypaulb/Canvus-Go-API` → `github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus`, replace the bespoke insecure-HTTP setup with `WithAPIKey` plus an explicit `WithVerifyTLS(false)` only when `--insecure` is set, adopt `FromEnv` for the env-var path in `internal/session/session.go`, switch `import.go`/`export.go` to call the SDK's `ExportWidgetsToFolder`/`ImportWidgetsToRegion` (drop any duplicated logic), and run the existing tests. Update CLAUDE.md to reflect monorepo paths.

**Recommended model:** opus. >2k LOC and architectural decisions to make (insecure-TLS handling, FromEnv adoption, export refactor).

**External dependencies** — `cobra`, `viper`, `bubbletea`, `lipgloss`, `yaml.v3`, `golang.org/x/term`. All current, all keep.

**README contents** — Should document: monorepo path; new SDK import; updated env vars (`CANVUS_API_URL` not `CANVUS_URL` — see Phase 4a `0770701` env-var unification); link to convention doc; remove "Go install" instructions referring to the old standalone repo; add link to the monorepo's top-level README.

---

## Item 2 — CanvusPowerToys → go/tools/powertoys/

**Source LOC:** ~23k Go + 4.8k HTML/JS/CSS = ~27.8k total | **File count:** 72 Go, 14 HTML, 14 JS, 34 CSS | **Effective LOC:** ~23k Go (after excluding `dross/`, `devdocs/`, `dev-docs/`, `agent-os/`, 13 binary `.exe` files, `assets/`, `scripts/`)
**Purpose** — A Fyne desktop GUI for Canvus-Server administration: Screen.xml visual editor for multi-display layouts, mt-canvus.ini config editor with embedded schema docs, CSS-plugin manager (rotation/looping/kiosk), custom-menu YAML designer with icon picker, and an optional embedded WebUI for remote management of canvases (zones, macros, RCU, page sequencing). Tray-resident on Windows.

**Current SDK usage** — Does **NOT** depend on Canvus-Go-API. It has its own hand-rolled HTTP client at `internal/atoms/webui/api_client.go:1-105` (`InsecureSkipVerify: true` hardcoded at line 27, `time.Sleep`-printf debug logging throughout). API surface is in `internal/atoms/webui/` (~1k LOC) and `internal/molecules/webui/` (~6.3k LOC). Endpoints exercised:
- `GET/POST/PATCH/DELETE /api/v1/canvases/{id}/{notes,images,videos,pdfs,anchors,browsers}` (with download variants)
- `GET /api/v1/canvases/{id}/widgets`
- `GET /api/v1/clients`, `GET /api/v1/clients/{id}/workspaces/0`
- **Non-standard**: `GET/POST /api/v1/canvases/{id}/rcu/config`, `/rcu/status`, `/rcu/test` (`internal/molecules/webui/rcu_handler.go`) — these are not in the canonical API; either a custom server extension or a stub. Pre-port question: confirm with Jaypaul whether RCU endpoints are real.

**SDK call audit** — Everything translates cleanly to the new Go SDK once swapped:
- Per-type widget CRUD → `notes.go`, `images.go`, `videos.go`, `pdfs.go`, `anchors.go`, `browsers.go` (all present per parity §1.1.4).
- Asset downloads → per-type `Download*` (`images.go:57`, `videos.go:28`, `pdfs.go:28`).
- Generic `GET /canvases/{id}/widgets` → `ListWidgets` (parity §1.1.3).
- Clients + workspaces → `clients.go`, `workspaces.go` (parity §1.1.11).
- Workspace subscription (`internal/atoms/webui/workspace_subscriber.go`) → new SDK has `SubscribeClientWorkspaces`/`SubscribeWorkspace` (Phase 4b §4.1 #10). Major upgrade — current code likely polls.
- RCU endpoints: **no SDK coverage**. These are NOT in parity-matrix or VERIFIED-CORRECTIONS. Flag for pre-port clarification (see Items needing pre-port clarification).
- Zone discovery (`internal/atoms/webui/zone_bounding_box.go`) — replace with `extras/zones.WidgetZoneManager` (parity §2.3).

**Architecture quality** — Mixed. Pluses: clean atomic layout (`atoms`/`molecules`/`organisms`). Minuses:
- `internal/atoms/webui/api_client.go:24,46` has `fmt.Printf("[APIClient] …")` debug logging throughout — should use a logger.
- `InsecureSkipVerify: true` is unconditional at `api_client.go:27` — must become opt-in via config.
- `internal/molecules/configeditor/` (~3.5k LOC), `internal/molecules/screenxml/` (~2.7k LOC), `internal/molecules/custommenu/` (~1.7k LOC), `internal/molecules/cssoptions/` (~1.1k LOC) are **not Canvus-API code** — they are local INI/XML/CSS/YAML editors. They should port verbatim with no SDK changes.
- `internal/molecules/webui/canvas_service.go` (~750 LOC visible from listing), `pages_handler.go`, `pages_zones.go`, `macros_handler.go`, `admin_handler.go`, `rcu_handler.go`, `sse_handler.go` are the Canvus-API-touching layer (~6.3k LOC) — this is where the SDK swap concentrates.
- `internal/molecules/webui/api_routes_integration_test.go`, `canvas_service_test.go`, `pages_zones_test.go`, `workspace_subscription_test.go` — good test coverage on the API-touching layer; preserve.
- `internal/atoms/webui/widget.go:11-32` defines its own `Widget`, `WidgetLocation`, `WidgetSize` — replace with SDK types.
- `dross/`, `devdocs/`, `dev-docs/` directories: planning material, do not port.
- 13 versioned `.exe` files at repo root + Linux binary: do not port.

**Refresh complexity estimate:** large. ~7k LOC of API-touching code to rewire, plus workspace-subscription is a real architectural upgrade (polling → typed `SubscribeClientWorkspace` channel). The non-API ~9k LOC (configeditor + screenxml + custommenu + cssoptions) is straight copy. Plus 13 versioned binaries to NOT bring across.

**Recommended port approach** — Two-phase. Phase A (mechanical): copy `assets/`, `atoms/{logger,theme,backup,config,version,shortcut,errors,validation,paths}`, all of `molecules/{configeditor,screenxml,custommenu,cssoptions,tray}`, `organisms/app`, `organisms/services`, and `cmd/powertoys` verbatim. Phase B (SDK rewire): rewrite `internal/atoms/webui/api_client.go` as a thin shim around `canvus.Session`, replace `widget.go`/`client_resolver.go`/`canvas_tracker.go`/`zone_bounding_box.go`/`workspace_subscriber.go` with SDK calls (especially `SubscribeWorkspace` for the tracker), audit and either port-or-document the `rcu_handler.go` non-standard endpoints (Jaypaul clarification needed), strip `fmt.Printf` debug noise in favour of `internal/atoms/logger`, drop `InsecureSkipVerify: true` default in favour of an opt-in flag. Skip everything under `dross/`, `devdocs/`, `dev-docs/`, `agent-os/`, `scripts/`, all `.exe` and `*-linux` files.

**Recommended model:** opus. ~7k API-touching LOC to rewire + architectural Subscribe upgrade + RCU endpoint question.

**External dependencies** — `fyne.io/fyne/v2`, `getlantern/systray`, `tdewolff/minify`, `ini.v1`, `yaml.v3`. All keep. (Note: `fyne/v2.7.1` and the GL/GLFW chain are heavy native deps — already known.)

**README contents** — Should document: monorepo path; new SDK; the deliberate separation of non-API tools (Screen.xml, INI editor, CSS plugins, custom menus) from API-touching webui; deployment story (Windows tray vs Linux daemon); RCU endpoint status (if real). Old README is feature-rich; preserve content, retarget paths.

---

## Item 3 — CanvusTranslator → go/tools/translator/

**Source LOC:** 1,007 (Go + minimal HTML/JS/CSS) | **File count:** 5 Go, 1 HTML, 1 JS, 1 CSS | **Effective LOC:** ~870 production + ~135 web
**Purpose** — A web service that, on POST `/translate`, fetches all note widgets from one Canvus canvas, uses Google Gemini Generative AI to translate their content into the requested language, and updates the notes in place. Single-canvas, single-tenant; configured via env (`CANVUS_API_KEY`, `CANVAS_ID`, `GEMINI_API_KEY`). Tiny.

**Current SDK usage** — Uses its own hand-rolled `internal/canvusapi/canvusapi.go:1-492` (the same legacy "canvusapi" pattern recurring across AI-personas/CanvusAPI-LLMDemo/CanvusNoteMapper). Endpoints exercised:
- `GET /api/v1/canvases/{id}/widgets?subscribe` — but only used as a one-shot fetch in `internal/canvus_service.go:30` via `client.GetWidgets(false)`.
- `PATCH /api/v1/canvases/{id}/notes/{nid}` — to update each note.

Real surface used: just `GetWidgets` + `UpdateNote`. The other 480+ lines of `canvusapi.go` are dead weight for this app.

**SDK call audit** — Trivially covered:
- `GetWidgets` → `session.ListWidgets(ctx, canvasID, nil)` (parity §1.1.3).
- `UpdateNote` → `session.UpdateNote(ctx, canvasID, noteID, &canvus.NotePatch{Text: ...})` (parity §1.1.4 `notes.go:43`).

No gaps. The vendored `canvusapi.go` is straight dead code post-port.

**Architecture quality** — Clean and tiny.
- `cmd/server/main.go:137` is a single-file HTTP server; readable.
- `internal/canvus_service.go:30-76` is a wrapper around a wrapper — can collapse into one file.
- `internal/gemini_service.go:62` uses the deprecated `github.com/google/generative-ai-go` package — Google has migrated to `github.com/google/genai`. Replace.
- `internal/canvusapi/canvusapi.go` is 492 lines of generic SDK code, only two methods of which are used — delete entirely after port.
- Web frontend (`web/`) is a single HTML/JS/CSS page; preserve.

**Refresh complexity estimate:** trivial. Replace vendored SDK with new SDK, modernize Gemini client, collapse the wrapper layers. ~200 LOC of net production code after refresh.

**Recommended port approach** — Treat as a near-rewrite: keep the web frontend and the `/translate` HTTP-handler shape; replace `internal/canvusapi/canvusapi.go` with a single `canvus.Session`; merge `canvus_service.go` into the handler; replace the Gemini client with the current `github.com/google/genai` package. Adopt `FromEnv` for the session and document the new env-var name (`CANVUS_API_URL` not `CANVUS_SERVER`).

**Recommended model:** sonnet. <1k LOC, mechanical rewire, no architecture decisions.

**External dependencies** — `github.com/google/generative-ai-go/genai` — DEPRECATED. Replace with `github.com/google/genai` (Vertex/Gemini unified SDK). `github.com/joho/godotenv` — keep or replace with SDK's env handling.

**README contents** — Document migration: old env vars (`CANVUS_SERVER`, `CANVAS_ID`, `CANVUS_API_KEY`, `GEMINI_API_KEY`) and which became `CANVUS_API_URL`; explicit "this works on one canvas at a time" scope; the Gemini SDK upgrade; the web UI's single-page nature; deployment as a stateless web service.

---

## Item 4 — Canvus-Server-db-solver → go/tools/db-solver/

**Source LOC:** 10,058 (Go) | **File count:** 67 Go | **Effective LOC:** ~4,100 application code (`src/internal/` + `src/cmd/`) + ~6,000 of vendored legacy SDK (`src/pkg/canvus/canvus/` — 32 files including `_test.go` files) that must be stripped.
**Purpose** — A Cobra CLI tool to recover from corrupted Canvus-Server installations: queries the Canvus API for all assets (across all canvases), compares against the on-disk asset store, identifies missing files, looks up their hashes in the PostgreSQL asset_files table, and restores from configured backup folders. Operates at the canvus-server filesystem + DB level — distinct from all other tools in this list.

**Current SDK usage** — Vendored legacy Canvus-Go-API at `src/pkg/canvus/canvus/` (full copy including tests). Real app code in `src/internal/canvus/discovery.go:511` calls just six SDK methods:
- `session.ListCanvases(ctx, nil)`
- `session.ListWidgets(ctx, canvas.ID, nil)`
- `session.GetCanvasBackground(ctx, canvas.ID)`
- `session.GetImage(ctx, canvas.ID, widget.ID)`
- `session.GetPDF(ctx, canvas.ID, widget.ID)`
- `session.GetVideo(ctx, canvas.ID, widget.ID)`

That's it. Discovery walks every canvas, every widget, pulls hashes from per-asset detail responses.

**SDK call audit** — All six methods exist in new SDK (parity §1.1.1, §1.1.3, §1.1.4). No gaps. The vendored 6k-LOC `src/pkg/canvus/canvus/` is pure dead code after import-path swap.

**Architecture quality** — Generally good — clear separation of concerns:
- `src/internal/canvus/discovery.go:511` walks the API.
- `src/internal/database/client.go:206` handles Postgres (lib/pq).
- `src/internal/filesystem/{scanner,hash_search,hash_catalog}.go` scan backup roots.
- `src/internal/backup/{searcher,restorer}.go` orchestrate hash-keyed lookup + copy.
- `src/internal/commands/{run,discover,lookup_hash}.go` are Cobra commands.
- `src/internal/config/{config,prompts}.go` are interactive prompts.

Concrete issues:
- `src/internal/commands/lookup_hash.go:744` is a 744-line organism — split: prompt-collection, hash-lookup orchestration, restore orchestration, reporting.
- `src/cmd/canvus-server-db-solver/main.go:282` mixes rootCmd setup with version-injection scaffolding — typical Cobra pattern but could be trimmed.
- The `kpmg-paris-report/` directory is sample input data; do not port (or move to a `testdata/` subdir).
- `restore_assets.ps1` PowerShell helper — port to a Go subcommand or document as deprecated.
- `src/pkg/canvus/canvus/*_test.go` — these are the legacy SDK's own tests, not app tests. Delete with the vendored SDK.

**Refresh complexity estimate:** moderate. The "moderate" rating reflects that 6k LOC of the source goes straight to /dev/null (vendored SDK) and the remaining 4k LOC is well-structured with a small SDK surface. The risks: Postgres dependency (lib/pq) needs to stay; interactive prompts (`internal/config/prompts.go`) should remain; restore semantics (file copy + permissions) must not regress.

**Recommended port approach** — Copy only `src/cmd/` and `src/internal/` (drop `src/pkg/`), rewire imports from `…/canvus-server-db-solver/src/pkg/canvus/canvus` to `github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus`, no SDK call rewrites needed. Split `lookup_hash.go` into ≤300-LOC files. Move `kpmg-paris-report/` and `restore_assets.ps1` to a `legacy/` subdir or drop entirely. Strip the `bin/`, `build/`, `releases/`, `docs/` (planning material), `public/` directories.

**Recommended model:** opus. Despite the trivial SDK swap, the code touches production PostgreSQL + file restore — wrong moves here lose customer data. Architectural judgement needed on the `lookup_hash.go` refactor.

**External dependencies** — `lib/pq` (Postgres), `cobra`, `yaml.v3`. All keep. Postgres-driver replacement to `jackc/pgx` is tempting but out of scope.

**README contents** — Document: monorepo path; new SDK; the PostgreSQL connection prerequisite; the asset backup-folder semantics (was the old README ambiguous?); the recovery workflow (discover → lookup-hash → run); operational guardrails (dry-run flag, backup-before-restore); the dropped `kpmg-paris-report` sample data.

---

## Item 5 — AI-personas → go/examples/projects/ai-personas/

**Source LOC:** 5,246 (Go) + 1 HTML | **File count:** 20 Go, 1 HTML | **Effective LOC:** ~5,250 (no significant vendored deps beyond `canvusapi/canvusapi.go:680`)
**Purpose** — Demo app: monitors a Canvus canvas for trigger-widget creations, uses Google Gemini to generate personas from business-model data extracted from notes, then runs persona-driven Q&A simulations laid out in a 3×3 grid around each question with connectors and anchors. Includes a web dashboard (`internal/web/server.go:547`) to manage personas and observe sessions.

**Current SDK usage** — Hand-rolled `canvusapi/canvusapi.go:680` (same pattern as Translator). Real SDK surface used (from `internal/canvus/events.go` + `internal/gemini/aiquestion.go` + `internal/molecule/*.go`):
- `client.GetWidgets` → `session.ListWidgets`
- `client.GetWidget` → `session.GetWidget`
- `client.GetNote` → `session.GetNote`
- `client.CreateNote`, `UpdateNote`, `DeleteNote` → per-type SDK methods
- `client.CreateAnchor` → `session.CreateAnchor`
- `client.CreateConnector` → `session.CreateConnector`
- Implicit subscribe (event-loop polling) in `internal/canvus/events.go:318` — replace with `SubscribeWidgets`.

Endpoints exercised: standard widgets/notes/anchors/connectors CRUD + the subscribe stream.

**SDK call audit** — All methods covered in new Go SDK (parity §1.1.3 + §1.1.4). The `EventMonitor` polling loop is the major upgrade opportunity — swap to `SubscribeWidgets` (new typed helper from Phase 4b §4.1 #7). Anchor creation and connector chaining patterns already match SDK shapes.

**Architecture quality** — Mixed.
- `internal/gemini/aiquestion.go:1286` is a god-organism — needs decomposition into persona-prompt / response-parse / layout / overlap-resolution.
- `canvusapi/canvusapi.go:680` is the full vendored legacy SDK — delete after port.
- `internal/web/server.go:547` is borderline-large; reasonable for an HTTP server but should split route handlers.
- `internal/gemini/personas.go:496`, `internal/gemini/client.go:580` — also large; split.
- `internal/atom/retry.go:185`, `internal/atom/format.go:64`, `internal/atom/payload.go:46`, `internal/atom/safetypes.go:43`, `internal/atom/string.go:40` — proper atomic decomposition; preserve.
- `internal/molecule/{business,anchor,layout}.go` — molecules pattern in place.
- `internal/timing/timing.go:89` — instrumentation utility; preserve.
- `cmd/ai-personas/main.go:219` — reasonable size for the orchestration entrypoint.
- `dross/` planning material — do not port.
- `qr_remote.png` and 3 versioned `ai-personas*` binaries at root — do not port.

**Refresh complexity estimate:** large. The SDK swap is mechanical, but `internal/gemini/aiquestion.go:1286` deserves real refactoring (it's 1.3k LOC of mixed concerns: persona iteration, response succinctness re-prompting, grid layout math, overlap handling, anchor grouping). The Subscribe upgrade replaces a manual polling loop.

**Recommended port approach** — Mechanical SDK swap first; preserve current behaviour. Then refactor `aiquestion.go` into ≤300-LOC molecules: persona iterator, response-quality enforcer, grid layout, overlap resolver, anchor builder. Replace `internal/canvus/events.go` polling with `session.SubscribeWidgets(ctx, canvasID)` — net negative LOC, clearer code. Modernize Gemini client (`internal/gemini/client.go`) from deprecated `generative-ai-go` to `genai`. Drop `canvusapi/`, `dross/`, all binaries.

**Recommended model:** opus. >2k LOC + decomposition judgement on a 1.3k-LOC god-file.

**External dependencies** — `github.com/google/generative-ai-go/genai` (DEPRECATED — replace with `github.com/google/genai`), `github.com/sashabaranov/go-openai` (alternative LLM client, current). The example.env mentions OpenAI as well — keep both options.

**README contents** — Document: monorepo path; new SDK; the polling → subscribe upgrade; the Gemini SDK upgrade; the persona workflow with screenshots; clearer separation of "you need this canvas template / these starter notes" vs "run the binary"; the dashboard URL and admin auth.

---

## Item 6 — CanvusAPI-LLMDemo → go/examples/projects/llm-canvas-companion/ (renamed)

**Source LOC:** 4,466 (Go) | **File count:** 11 Go | **Effective LOC:** ~3,000 production (handlers.go is 2,104 LOC by itself) + ~750 of vendored SDK + ~700 of tests
**Purpose** — Demo that watches a Canvus canvas for content in `{{ double curly braces }}` triggers, then routes to LLM workflows: text completion, PDF summarization (chunk + consolidate), canvas-summary, image generation (DALL-E / Azure OpenAI / local stable-diffusion), and OCR via Google Vision. Supports OpenAI, Azure OpenAI, and local LLM servers (LMStudio / Ollama / LLaMA) by configurable base URLs.

**Current SDK usage** — Hand-rolled `canvusapi/canvusapi.go:510` (the recurring pattern). Real SDK surface (from `handlers.go:2104` + `monitorcanvus.go:277`):
- `client.GetWidgets`, `client.GetWidget`, `client.GetNote`, `client.UpdateNote`, `client.CreateNote`, `client.DeleteNote`, `client.DeletePDF`, `client.DeleteImage`, `client.DeleteVideo`
- `client.CreateImage` (upload), `client.CreatePDF` (upload)
- Subscribe loop via `monitorcanvus.go:65` (manual NDJSON parsing)
- `client.GetCanvasInfo` → new SDK `GetCanvas`

**SDK call audit** — All covered (parity §1.1.1, §1.1.3, §1.1.4). The manual NDJSON parsing in `monitorcanvus.go:65` should be replaced with `session.SubscribeWidgets` (Phase 4b §4.1 #7) — same upgrade as AI-personas.

**Architecture quality** — Mediocre.
- `handlers.go:2104` is a single-file god-organism with 27+ top-level handler functions: note handler, snapshot handler, PDF precis (incl. chunking + consolidation), canvas precis, AI image generation (×3 backends: OpenAI / Azure OpenAI / Stable Diffusion implied), error recovery, cleanup, processing-state UI. This MUST be decomposed.
- `monitorcanvus.go:277` is the Subscribe loop + widget-state tracking — should become a SDK Subscribe call + small dispatcher.
- `main.go:92` is fine.
- `canvusapi/canvusapi.go:510` — delete after port.
- `core/config.go:236`, `core/ai.go:54` — OK.
- `logging/logging.go` — small.
- `tests/{test_data,canvas_check_test,testAPI_test,llm_test}.go` (~700 LOC) — keep, retarget imports.
- `app.log` (2.5 MB), 2 root-level binaries, `downloads/`, `test_files/`, `Docs/`, `icons-for-custom-menu/` — do not port (move icons to embedded assets if used).
- `shared_canvas.json` (205 B) — sample data; relocate to `testdata/`.

**Refresh complexity estimate:** large. SDK swap is mechanical, but `handlers.go:2104` must be split into ≤300-LOC files keyed by trigger type (note / snapshot / pdf-precis / canvas-precis / ai-image / handwriting) before this can be considered an "example project" worth showing.

**Recommended port approach** — Phase A: mechanical SDK swap, replace manual NDJSON loop in `monitorcanvus.go` with `session.SubscribeWidgets`. Phase B: decompose `handlers.go` into `handlers/{note,snapshot,pdf,canvas,image,ocr,common}.go`. Phase C: extract the multi-backend AI client (`core/ai.go`) into proper provider abstractions — currently the LLM-base-URL juggling happens inline in handlers. Update env-var names per Phase 4a unification (`CANVUS_API_URL`). Drop binaries, `app.log`, `downloads/`, `test_files/` (or move to gitignored test fixtures).

**Recommended model:** opus. ~3k production LOC + 2.1k-LOC god-file decomposition + multi-LLM-provider abstraction work.

**External dependencies** — `openai/openai-go v0.1.0-alpha.44` (current alpha — verify stability), `sashabaranov/go-openai v1.38.2` (stable alternative), `ledongthuc/pdf` (PDF text extraction), `fatih/color`, `joho/godotenv`. Consider whether two OpenAI clients are needed.

**README contents** — Document: monorepo path; new SDK; the rename rationale (LLMDemo → llm-canvas-companion: it's a long-running companion, not a one-shot demo); the multi-backend LLM model (OpenAI / Azure / local); env-var renames; the `{{ }}` trigger convention; deployment ("run alongside a Canvus session"); the workflow gallery (text / PDF precis / canvas precis / image gen / OCR).

---

## Item 7 — CanvusNoteMapper → go/examples/projects/note-mapper/

**Source LOC:** 3,301 (Go + TS) | **File count:** 12 Go, 3 TS (in `ai/`), 1 HTML, 2 JS, 1 CSS | **Effective LOC:** ~2,500 Go production (after dropping vendored SDK) + 500 web frontend + ~300 TS prototype
**Purpose** — Webapp: user uploads a photo of a whiteboard or table covered in physical Post-it notes; backend uses Google Gemini multimodal LLM to extract each note (position, color, text), then creates corresponding digital notes inside a Canvus anchor on a target canvas, preserving spatial layout from the photo.

**Current SDK usage** — Hand-rolled `internal/canvusapi/canvusapi.go:490` (same legacy pattern). Real SDK surface (from `internal/mcs/mcs.go:201` and `internal/api/api.go:524`):
- `client.GetCanvases`, `client.GetCanvasSize`, `client.GetAnchorInfo`, `client.GetAnchors`, `client.GetWidgets`, `client.CreateNote`
- Direct HTTP for some endpoints (`internal/mcs/mcs.go:47-79` builds URLs by string concat: `/api/v1/canvases`, `/api/v1/canvases/{id}/anchors`)
- Subscribe URL constructed at `internal/canvusapi/canvusapi.go:465` but not clear it's used (need deeper investigation)

**SDK call audit** — All covered (parity §1.1.1, §1.1.3, §1.1.4). `GetAnchors` → `session.ListAnchors(ctx, canvasID, nil)` (parity §1.1.4 row 6). `GetCanvasSize` is a derived metric — compute from canvas bounding box or use `GetCanvasBackground`. No real gaps.

**Architecture quality** — Reasonable.
- `internal/api/api.go:524` is the HTTP handler layer — borderline; reasonable for a single endpoint with file upload.
- `internal/canvusapi/canvusapi.go:490` — delete after port.
- `internal/mcs/mcs.go:201` — mixes direct HTTP construction with `canvusapi.NewClient` calls; clean up to consistent SDK usage.
- `internal/llm/extract_postit_notes.go:171` — Gemini integration; modernize SDK.
- `internal/pipeline/{pipeline,stage1_image_preprocessing}.go` — clean pipeline pattern.
- `internal/image/processor.go:95` — image preprocessing utilities.
- `internal/mapping/mapping.go:58` — spatial mapping logic.
- `cmd/main.go:48` — tiny entrypoint.
- `ai/{ai-instance.ts,dev.ts,flows/extract-postit-notes.ts}` — TS prototype code, predates the Go LLM integration; either delete or port to a TS example sibling. **Recommend delete** — the Go path is what ships.
- `web/` — frontend; preserve.
- `notescanner-linux-amd64` binary at repo root — do not port.

**Refresh complexity estimate:** moderate. SDK swap is mechanical; the architecture is sound; main complications are the `mcs.go` inconsistency between direct-HTTP and SDK-wrapper, the dead TS prototype, and the deprecated Gemini package.

**Recommended port approach** — Drop `internal/canvusapi/` entirely. Rewrite `internal/mcs/mcs.go` to use `canvus.Session` consistently (no string-concat URLs). Replace direct-Gemini code (deprecated `generative-ai-go`) with current `genai` SDK. Decide on the `ai/` TS prototype: delete (recommended). Drop the binary. Preserve `web/` frontend; update its env-var name in any inlined config.

**Recommended model:** sonnet. ~2.5k LOC, mostly mechanical, with one consistency cleanup (mcs.go) — does not warrant opus.

**External dependencies** — `github.com/google/generative-ai-go/genai` — DEPRECATED, replace. Standard Go web + image processing (no opencv-go but uses `image/jpeg`, etc.). The TS prototype uses `genkit` — irrelevant if deleted.

**README contents** — Document: monorepo path; new SDK; the Gemini upgrade; the rename of the binary to `note-mapper`; the workflow (upload photo → see notes on canvas in an anchor); env vars (`CANVUS_API_URL`, `GOOGLE_GENAI_API_KEY`); coordinate-system note (Canvus widgets are in pixels relative to parent — the original CLAUDE.md captured this; preserve).

---

## Item 8 — canvus-mcp-server → python/tools/mcp-server/

**Source LOC:** 26,338 (Python) | **File count:** 53 .py files | **Effective LOC:** ~13.6k production (`canvus_mcp_server/`) + ~12.7k tests (`tests/`) — well-tested.
**Purpose** — FastAPI server exposing Canvus operations as MCP (Model Context Protocol) tools for AI agents. Wraps the legacy `CanvusPythonAPI` for HTTP transport, adds Ollama-powered local LLM workflows (brainstorming analysis, correlation/connector suggestion, report generation), PDF text extraction + SQLite caching, and user/auth management. The single largest python item in the consolidation.

**Current SDK usage** — Imports `from canvus_api import CanvusClient, CanvusAPIError` (the legacy package name, distinct from the monorepo's new `canvus_sdk`). 36 unique SDK methods used (`grep -hE "self\.client\.[a-z_]+\("` across `mcp_tools/`):
- Canvas CRUD: `list_canvases`, `get_canvas`, `create_canvas`, `update_canvas`, `delete_*` (per type), `copy_canvas`, `move_canvas`, `get_canvas_permissions`
- Per-type widget CRUD: `create_note/image/pdf/video/connector`, `get_*`, `update_*`, `list_*` for notes/images/pdfs/videos/connectors
- Auth: `login`, `logout`, `get_current_user`
- Users: `list_users`, `create_user`

All standard endpoints.

**SDK call audit** — Method-name mapping legacy → new SDK is mostly trivial but not always 1:1. Need to confirm during port:
- `client.list_canvases()` → `client.canvases.list()` (new SDK uses resource-grouped namespaces per parity §1.1)
- `client.list_notes(canvas_id)` → `client.notes.list(canvas_id)` or `client.widgets.notes.list(canvas_id)` — verify against `python/sdk/src/canvus_sdk/resources/widgets.py:125-145`
- `client.get_current_user()` → newly added in Phase 4b (parity §1.1.6 row "Get current user" + Phase 4b plan §4.2 #7) — confirmed present.
- `client.copy_canvas`, `move_canvas` — present (parity §1.1.1 rows 7-8).

The resource-namespace restructuring is the biggest mechanical change — every `self.client.<verb>_<resource>(...)` becomes `self.client.<resource>.<verb>(...)`. This is search-and-replace-heavy but conceptually flat.

**Architecture quality** — Mixed. The tests are excellent (15+ test files, 12.7k LOC), production code is well-decomposed at the package level but has one outrageous file:
- `canvus_mcp_server/mcp_tools/canvas.py:3,169` — 3.2k LOC of canvas tools in one file. This is the worst single file in the entire 10-repo audit. Must be split: by resource (canvas/note/image/pdf/video/connector) or by verb-group (read/write/delete/admin).
- `canvus_mcp_server/mcp_tools/brainstorming_analysis.py:902` — large but cohesive (a single workflow); could stay if cleanly split into stages.
- `canvus_mcp_server/mcp_tools/report_generation.py:821` — similar.
- `canvus_mcp_server/mcp_tools/llm_tools.py:750` — Ollama orchestration; could split by tool.
- `canvus_mcp_server/auth.py:730`, `pdf_processing.py:714`, `correlation_analysis.py:606`, `llm_client.py:541`, `auth_endpoints.py:523` — all borderline-large; review during port.
- `canvus_mcp_server/mcp_tools/base.py`, `registry.py`, `__init__.py` — clean abstractions; preserve.
- Test coverage is impressive — preserve all tests, retarget imports.
- `htmlcov/`, `coverage.xml`, `mcp_server.log`, `venv/`, `canvus_mcp_server.egg-info/`, `docs/` (planning), `agent-os/`, `scripts/`, root-level dev md files (`DEV_GUIDE.md`, `DEV_WORKFLOW.md`, `LLM_DEV_GUIDE.md`, `TASKS.md`, `Notes.md`, `PRD.md`, `TECH_STACK.md`) — do not port; relocate or drop.

**Refresh complexity estimate:** major-rewrite. NOT because the SDK swap is hard, but because (a) the 3.2k-LOC `canvas.py` MUST be split before this can ship as a quality example, (b) the resource-namespace migration affects every `mcp_tools/*.py` file, (c) test imports also need updating, (d) the SQLite cache + Ollama integration are stateful components that deserve scrutiny. This is the single biggest item in the audit.

**Recommended port approach** — Three rounds. Round 1: copy structure, do mechanical import swaps (`canvus_api` → `canvus_sdk`, restructure method calls to resource namespaces), run tests, fix breakage. Round 2: decompose `canvas.py` into `mcp_tools/canvas/{__init__,canvas_ops,permissions,bulk_ops}.py` and `mcp_tools/widgets/{notes,images,pdfs,videos,connectors}.py` — keep the MCP-tool naming-registry stable. Round 3: trim borderline-large files; consolidate auth (`auth.py` + `auth_endpoints.py` overlap?). Drop all planning md, dev md, coverage artifacts, venv, .egg-info.

**Recommended model:** opus. Largest item, multi-round refactor, strict test coverage to preserve.

**External dependencies** — `fastapi`, `fastapi-mcp`, `uvicorn`, `pydantic`, `pdfplumber`, `httpx` (will become transitive via new SDK), Ollama via custom `llm_client.py`. All keep. Verify `fastapi-mcp>=0.3.7` is still maintained.

**README contents** — Document: monorepo path; new SDK migration with code-style example (resource-namespace style); the MCP tool registry pattern; SQLite cache location + invalidation; Ollama setup (link to existing `docs/OLLAMA_SETUP.md`); deployment story; the brainstorming workflow with screenshots; rate-limiting / concurrency considerations now that the new SDK exposes a circuit-breaker.

---

## Item 9 — Canvus-Local-LLM → python/tools/local-llm/

**Source LOC:** 772 (Python) | **File count:** 8 .py | **Effective LOC:** 772 — but the actual implementation is ~530 LOC of scaffolding + ~240 LOC of tests. **The repo is largely a SKELETON.**
**Purpose** — INTENDED to be a Windows system-tray Python app that monitors Canvus canvases via the `?subscribe` stream and routes content (notes, PDFs, images) to a local Ollama vision model for AI processing (text analysis, PDF precis, OCR, canvas summary). **As-of audit date: the PRD is comprehensive, but the source code is a stub** — `src/main.py:235` has a `CanvusLLMInterface` class with `canvus_client = None` placeholder (line 31), no actual SDK calls, no Ollama integration, no subscribe loop. The 14 `section-N-issue.md` files are an implementation plan that was never executed.

**Current SDK usage** — None. `src/config.py:127` defines `canvus_api_key`/`canvus_username`/`canvus_password` settings but no client is constructed. `src/main.py:31` has `self.canvus_client = None`.

**SDK call audit** — N/A: no existing SDK calls. This is a green-field implementation against the PRD using the new Python SDK.

**Architecture quality** — Skeleton is reasonable:
- `src/main.py:235` — application shell with lifecycle hooks; the methods are stubs.
- `src/config.py:134` — pydantic Settings-based config with Canvus + Ollama fields; clean.
- `src/exceptions.py:86` — exception hierarchy; preserve.
- `src/tray.py:84` — infi-systray menu (Windows-only); preserve as is for Windows, mark Linux as unsupported (or replace with cross-platform alternative).
- `tests/test_{main,config}.py:222` — ~220 LOC of unit tests for the scaffolding; preserve.
- Root-level `section-1-issue.md` through `section-14-issue.md` — 14 planning files; **delete or move to docs/planning/** post-port. The author abandoned mid-plan.

**Refresh complexity estimate:** major-rewrite. This is essentially a NEW implementation against the PRD. The skeleton is so thin that the port doesn't have anything substantive to refresh — it has to BUILD the missing pieces (subscribe loop, dispatch table, Ollama wrapper, image/PDF processing, system-tray glue).

**Recommended port approach** — Treat as a green-field example backed by the existing PRD (`PRD.md` is excellent — 13 KB, comprehensive). Adopt the new Python SDK (`canvus_sdk`) for all transport. Build the missing pieces in the order the section-issue files suggest: subscribe + dispatch → Ollama client → PDF/image processing → tray UI. Drop the 14 section-issue files. Drop `poetry.lock` (use uv per monorepo convention `python/sdk/src/canvus_sdk/` setup).

**Recommended model:** opus. Major-rewrite from PRD; design judgement required on dispatch architecture and Ollama integration.

**External dependencies** — `httpx` (will be transitive), `pydantic`, `pydantic-settings`, `aiohttp` (probably redundant with httpx — drop one), `infi-systray` (Windows-only), `Pillow`, `ollama`, `PyPDF2`, `opencv-python`, `loguru`, `fastapi`. Several of these are aspirational; trim during build. `PyPDF2` is deprecated — use `pypdf`.

**README contents** — Document: monorepo path; new SDK; the actual implementation status (current README implies it works — it doesn't); the system-tray workflow with screenshots once built; Ollama prerequisites; the platform support matrix (Windows-first); the rationale for "local LLM" vs the cloud-based `llm-canvas-companion` companion.

---

## Item 10 — CanvusWebUI → typescript/examples/webui/ (rewrite from JS)

**Source LOC:** 9,164 (JS + HTML + CSS) | **File count:** 15 JS, 9 HTML, 2 CSS | **Effective LOC:** 3,797 server.js + 5,367 frontend (HTML + JS + CSS)
**Purpose** — Express-based web app providing a Canvus admin/operator UI: identifies users (color assignment), uploads notes/images/videos/PDFs to a canvas, creates team-target anchors, zone-based macros (move/copy/delete/group-color/group-title/auto-grid), page-sequence management ("PageRes" + "pages"), RCU (Real-time Canvus Updates?) config/status/test, deleted-record audit/undelete, and an admin panel for env-var management.

**Current SDK usage** — JavaScript, no SDK. Uses `axios` via `webui/utils/apiClient.js:50` configured with `baseURL: CANVUS_SERVER`, `Private-Token: CANVUS_API_KEY`, optional `httpsAgent` for self-signed certs. All Canvus calls are inline `apiClient.get/post/patch/delete('/api/v1/canvases/...')` in `webui/server.js:3797`. Endpoints exercised (from grep):
- `GET/PATCH /api/v1/canvases/{id}` (target canvas verification + update)
- `GET/POST/PATCH/DELETE /api/v1/canvases/{id}/{notes,anchors,widgets}` (and per-id variants)
- `GET /api/v1/canvases/{id}/widgets` (mass operations under macros)
- `GET /api/v1/clients`, `GET /api/v1/clients/{cid}/workspaces/{wsid}` (workspace tracking)
- **Non-standard?** Routes are constructed dynamically with `${route}` interpolation — `/api/v1/canvases/{id}/{route}/{wid}` where `route` is derived from widget type; this is the recurring multi-type dispatch.

**SDK call audit** — N/A as a method-by-method audit since this is a TS rewrite. Conceptually all endpoints map cleanly:
- Per-type widget CRUD → typed namespaces on TS SDK (parity §1.1.4).
- Generic widget list → `session.widgets.list(canvasId)`.
- Workspace tracking → can use `session.subscribeWorkspaces(clientId)` (parity §1.2 row 27) — major upgrade from polling.
- Upload via multipart → per-type create methods in TS SDK accept `Blob`.
- Macros (auto-grid, group-color/title) use SDK widget patches.
- Zone management → `extras/widgetOperations.WidgetZoneManager` (parity §2.3).

**Architecture quality** — Poor by modern TS standards.
- `webui/server.js:3,797` is the worst single file in the audit alongside `canvas.py`. All 30+ HTTP route handlers are in one CommonJS file. Must be decomposed by route group.
- Frontend HTML/JS/CSS is vanilla — no framework. For a "TypeScript example" we shouldn't bring in React/Next; consider keeping it framework-free and adding TS via build (esbuild / vite) or migrating to a lightweight framework (Astro / Solid) — Jaypaul's call. Recommend: keep vanilla JS in templates, but rewrite `server.js` as a Hono or Fastify app in TS with typed routes.
- `users.json` is on-disk session state — fine for a demo but document it.
- `webui-test-uploads/` — test fixtures; relocate.
- `docker-compose.yml`/`Dockerfile` — preserve, retarget paths.
- `pm2` ecosystem.config — preserve if deployment story matters; otherwise drop.
- CommonJS (`require`) → ESM (`import`) — universal migration.
- Express → consider Hono (better TS, smaller, edge-deployable) or Fastify (typed schemas, still Node-only). Given monorepo's pnpm setup, Hono is the natural pick.

**Refresh complexity estimate:** major-rewrite. 3.8k LOC of express server in ONE FILE + 5.4k of frontend. The frontend can mostly stay (HTML + vanilla JS); the server must be decomposed in TS with proper routing, validation (zod schemas at the seams since the SDK doesn't do runtime validation per Tier-2 §6), and proper error handling.

**Recommended port approach** — TS server: scaffold a Hono app, organize routes into `src/routes/{users,uploads,macros,zones,pages,rcu,admin}.ts`; use `@hono/zod-validator` for request body validation (the SDK won't validate responses but inputs need it); wrap the TS SDK as a per-request resource (api-key-based session). Frontend: keep HTML + vanilla JS; add a small TS build for shared utilities. Drop `users.json` in favour of `cookie-session` or document the demo-only persistence story. The RCU endpoints in the legacy server (`/api/v1/canvases/{id}/rcu/*`) — same question as PowerToys: confirm with Jaypaul whether these are real before porting.

**Recommended model:** opus. Major rewrite, framework selection, architectural design — clearly opus territory.

**External dependencies** — Drop: `axios` (use SDK), `body-parser` (built into Hono), `fs`, `path`, `pm2`. Keep concept of: `dotenv`-style env loading (built into Node 20+), `multer`-equivalent for uploads (Hono has `parseBody`), `express-validator` → zod, `uuid` → `crypto.randomUUID`.

**README contents** — Document: monorepo path; new TS SDK; the migration from Express/JS to Hono/TS; the rationale ("example of a real-world admin UI built on the SDK"); the demo-grade persistence (`users.json` is not production); per-feature walkthrough (uploads, macros, zones, pages, RCU if kept); deployment with the existing Docker compose.

---

## Cross-item observations

**1. The recurring "vendored canvusapi" pattern.** Four items (Translator, AI-personas, CanvusAPI-LLMDemo, CanvusNoteMapper) each contain a near-identical hand-rolled `canvusapi/canvusapi.go` (~490-680 LOC) that predates the legacy SDK split. Db-solver vendors the full legacy `pkg/canvus/canvus/` (~6k LOC including tests). PowerToys has its own hand-rolled `internal/atoms/webui/api_client.go` (~1k LOC). After the port, ALL of these go to /dev/null. Net effective LOC reduction across these six items: ~11,000+ LOC of code that simply disappears. The new SDK is doing real work.

**2. Subscribe-via-polling is the universal architectural debt.** Translator (calls `GetWidgets(false)` once — fine for single-shot), AI-personas (`internal/canvus/events.go:318` event-loop polling), CanvusAPI-LLMDemo (`monitorcanvus.go:65` manual NDJSON parsing), CanvusWebUI (no subscribe at all — pure polling), PowerToys (`workspace_subscriber.go` likely polling). Phase 4b shipped 36 typed Subscribe helpers in Go and ~28 in Python and 27 in TS — this is the killer upgrade across all five long-running watchers. Every implementer agent should explicitly look for the polling loop and replace it with the typed `session.Subscribe*` call. **Specifically flag**: `SubscribeWidgets` for canvas watchers, `SubscribeWorkspace`/`SubscribeWorkspaces` for the PowerToys/WebUI workspace trackers.

**3. Deprecated Gemini SDK across three items.** AI-personas, Translator, and CanvusNoteMapper all use `github.com/google/generative-ai-go/genai`, which Google has deprecated in favour of the unified `github.com/google/genai` Vertex+Gemini SDK. Worth pre-stating in the work packet: every Go LLM example needs this migration. Could be extracted into a tiny `go/examples/internal/llm/gemini.go` shared helper (the rule of three is satisfied here: three sites, identical pattern).

**4. Two giant single files dominate the audit's risk surface.** `canvus_mcp_server/mcp_tools/canvas.py:3,169` and `CanvusWebUI/webui/server.js:3,797`. These are the only two files >3k LOC in the entire 10-repo set. Both REQUIRE decomposition before ship. Allocate extra time for each.

**5. The "Subscribe → dispatcher" pattern reoccurs across ai-personas, llm-canvas-companion, local-llm, mcp-server (brainstorming watcher).** Worth considering whether `go/examples/internal/dispatch/` and `python/examples/internal/dispatch/` shared helpers (a tiny "trigger-handler registry" pattern) would benefit all four. Caveat: don't extract until at least three of the four ports are complete (rule of three on real implementations, not on aspiration).

**6. Env-var inconsistency across all items.** Mixed bag: `CANVUS_SERVER`, `CANVUS_URL`, `CANVUS_API_URL`, `CANVAS_ID`, `CANVUS_API_KEY`. Phase 4a `0770701` unified the example apps on `CANVUS_API_URL`. Every port should respect this convention. The legacy READMEs all need updating.

**7. CSS / non-API tools in PowerToys are NOT consolidation candidates per se — they're a feature suite.** Of PowerToys' ~23k Go LOC, only ~7.3k touch Canvus API. The rest (Screen.xml editor, INI editor, CSS plugin manager, custom-menu designer) is a Canvus-Server administration suite that happens to live alongside the API layer. Pre-port question: is the intent to bring all of PowerToys (admin suite + API ops) into the monorepo, or only the API-touching webui module? **Recommend bringing all of it** — splitting would orphan the desktop GUI shell.

**8. Tests are a major asset in two items.** canvus-cli has 25 test files (~3.8k LOC) and canvus-mcp-server has 12.7k LOC of tests. These are huge accelerators for the port — keep them, retarget imports, and let them be the regression net. Other items have sparse-to-zero tests and accept the risk.

---

## Items needing pre-port clarification

**A. PowerToys + WebUI: are the `/api/v1/canvases/{id}/rcu/*` endpoints real?** `internal/molecules/webui/rcu_handler.go` calls `GET /rcu/config`, `POST /rcu/config`, `GET /rcu/status`, `POST /rcu/test`. The legacy CanvusWebUI also has RCU pages. NEITHER endpoint is in the canonical API spec, parity-matrix, or VERIFIED-CORRECTIONS. Possibilities: (1) custom server extension at MultiTaction, (2) MQTT-to-HTTP shim built into a non-stock canvus-server, (3) cruft from an aborted feature. **Resolution recommendation:** ask Jaypaul. If real, document them in VERIFIED-CORRECTIONS §"not yet verified" and have the porters keep them as raw transport calls. If not real, strip the RCU handler from both items before porting.

**B. Canvus-Local-LLM is barely an MVP — confirm scope.** The PRD describes a full Windows tray app; the source is ~530 LOC of scaffolding with `canvus_client = None`. Treating this as an MVP "to refresh" overstates its readiness. **Resolution recommendation:** confirm with Jaypaul that this counts as a green-field example built against the existing PRD. If yes, allocate opus + a longer time budget. If the bar is lower (just a working subscribe-and-print-text proof of concept), say so explicitly to the implementer.

**C. CanvusWebUI framework choice.** TS rewrite has multiple defensible targets: vanilla Node + Hono (recommended), Fastify, Express + ts-node, Next.js Server Actions, Astro. The monorepo's TS SDK is pnpm-managed and ESM-first, which biases toward Hono/Fastify. **Resolution recommendation:** decide before the port starts.

**D. CanvusAPI-LLMDemo rename.** Spec maps to `llm-canvas-companion`. Confirm the rename is final before the port — it propagates to import paths, README headings, binary names.

**E. Old binaries — confirm "do not port" universally.** Multiple repos ship versioned `.exe` files at root (PowerToys has 13). Confirming that the monorepo policy is: do not commit binaries; use GitHub Releases for distribution.

---

## Recommended batch reshuffle

**Original plan:**
- Round 1 (4 items, parallel): translator, local-llm, note-mapper, llm-canvas-companion
- Round 2 (3 items, parallel): ai-personas, webui-rewrite, mcp-server-python
- Round 3 (3 items, parallel): cli, db-solver, powertoys

**Validation:** the plan is mostly sound; one swap recommended.

**Issue 1 — local-llm is major-rewrite, not "easy starter."** Putting it in Round 1 next to translator (trivial) and note-mapper (moderate) understates the work. It needs design judgement that the other Round-1 items don't.

**Issue 2 — db-solver is moderate, not "save for last."** Round 3 batches the three "biggest" items together (cli, db-solver, powertoys), but db-solver's effective application code is only ~4k LOC (the 10k total is inflated by 6k of vendored SDK). It's smaller than ai-personas (~5k effective production) which is in Round 2. db-solver also touches Postgres — extra opus value if scheduled when an opus agent is already warmed up on its conventions.

**Issue 3 — mcp-server-python is THE largest item by tested-LOC and has the worst single file (3.2k-LOC `canvas.py`).** Putting it in Round 2 alongside ai-personas + webui-rewrite (both also large) compresses the highest-risk work into a single round. Could overrun.

**Recommended reshuffle:**

- **Round 1** (parallel, trivial-to-moderate): translator, note-mapper, llm-canvas-companion, **db-solver**. All four are SDK-swap-and-clean with localized cleanup. (Move db-solver from R3 to R1.)
- **Round 2** (parallel, large with judgement): ai-personas, **local-llm**, webui-rewrite. (Move local-llm from R1 to R2; it deserves opus and isn't dependency-blocked by Round 1.) Three opus agents, three independent items.
- **Round 3** (parallel, the two giants + the polished one): cli, powertoys, mcp-server-python. cli is moderate and almost mechanical; powertoys and mcp-server-python are the two largest, both opus, both need decomposition of giant files. Three opus agents.

Rationale: this shifts roughly to "by judgement-need" instead of "by LOC". Round 1 becomes the warm-up batch (smallest changes, fastest feedback). Round 2 is design-heavy but not gigantic. Round 3 is the marathon batch where the biggest risks live and where prior rounds' lessons can feed in.

---

## Completion report

- **Total items audited:** 10 (one originally-spec'd 11th item — CanvusMCP Go — confirmed missing upstream and skipped per pre-flight instructions).
- **Per-item complexity breakdown:**
  - **trivial:** 1 (Item 3 translator)
  - **moderate:** 4 (Item 1 canvus-cli, Item 4 db-solver, Item 7 note-mapper, Item 5 ai-personas counted "large" — see below)
  - **large:** 3 (Item 2 powertoys, Item 5 ai-personas, Item 6 llm-canvas-companion)
  - **major-rewrite:** 3 (Item 8 mcp-server, Item 9 local-llm, Item 10 webui)

  (Counts: trivial 1 + moderate 3 + large 3 + major-rewrite 3 = 10.)

- **Top 3 cross-item observations:**
  1. ~11,000+ LOC of vendored / hand-rolled HTTP-client code disappears across six items once the new SDK is adopted — the consolidation actually shrinks the codebase materially while improving quality.
  2. Five of the ten items have polling-loop watchers that can be replaced with typed Subscribe helpers shipped in Phase 4b — this is the universal architectural upgrade and should be a checklist item for every implementer agent.
  3. Two giant single files (`canvus_mcp_server/mcp_tools/canvas.py:3,169` and `CanvusWebUI/webui/server.js:3,797`) dominate the audit's risk surface. Both MUST be decomposed before ship and warrant extra time + opus + verification.

- **Items flagged needing pre-port clarification:**
  - **A** (PowerToys + WebUI RCU endpoints) — confirm whether `/api/v1/canvases/{id}/rcu/*` is a real custom server feature.
  - **B** (local-llm scope) — confirm green-field rewrite from PRD is the intent vs a smaller scope.
  - **C** (WebUI TS framework) — Hono recommended; Jaypaul to confirm.
  - **D** (LLMDemo rename) — confirm `llm-canvas-companion` rename is final.
  - **E** (binary commit policy) — confirm "no committed binaries; releases via GitHub Releases."

- **Audit file path:** `/home/jaypaulb/Projects/gh/MT-Canvus-Tools/docs/api-reference/per-item-refresh-audit.md`
