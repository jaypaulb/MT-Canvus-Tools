# canvus — Canvus CLI

The `canvus` command is a thin wrapper around the MT-Canvus-Tools Go SDK
(`github.com/jaypaulb/MT-Canvus-Tools/go/sdk`). Command parsing, configuration
loading, and output formatting live in this module; every API call delegates
to a `*canvus.Session` method.

This is the Phase 4c port of the standalone `github.com/jaypaulb/canvus-cli`
repository into the monorepo. See `docs/api-reference/per-item-refresh-audit.md`
(Item 1) for the audit that justified the moves below.

## Layout

```
go/cli/
├── main.go                       # cobra root, persistent flags, init wiring
└── internal/
    ├── commands/
    │   ├── audit/                # canvus audit list
    │   ├── canvas/               # canvus canvas …          + canvas/background/
    │   ├── client/               # canvus client list/get (create/update/delete are stubs)
    │   ├── colorpreset/          # canvus colorpreset …
    │   ├── config/               # canvus config get/set/path/show
    │   ├── folder/               # canvus folder …
    │   ├── group/                # canvus group …  (incl. add-user / remove-user)
    │   ├── mipmap/               # canvus mipmap …
    │   ├── system/               # canvus system get-config / update-config / send-test-email
    │   ├── testing/              # MockSession used by command tests
    │   ├── token/                # canvus token …
    │   ├── upload/               # canvus upload note / asset
    │   ├── user/                 # canvus user …
    │   ├── videoinput/           # canvus videoinput …
    │   ├── videooutput/          # canvus videooutput …  (update via SetVideoOutputSourceByID)
    │   ├── widget/               # canvus widget …  + per-type subdirs (note, image, pdf, video, anchor, browser, connector)
    │   ├── workspace/            # canvus workspace …
    │   ├── login.go              # canvus login (interactive password prompt via golang.org/x/term)
    │   ├── tui.go                # canvus tui (Bubble Tea splash + canvas browser)
    │   └── version.go            # canvus version
    ├── config/                   # ~/.canvus/config.yaml + env + flag merge
    ├── output/                   # json/yaml/table/text formatters + helpers
    ├── session/                  # SDK Session construction; SessionProvider interface for tests
    └── util/                     # slog-backed logger + APIError → exit-code mapping
```

## Build & run

```bash
# From the monorepo root
cd go/cli
go build -o ./canvus .            # binary lands in cli/
./canvus version

# With version info baked in (release-style):
go build \
  -ldflags "-X main.version=v0.2.0 -X main.commit=$(git rev-parse --short HEAD) \
            -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o ./canvus .
```

Don't `go build` from the monorepo root — that drops a binary called `canvus`
next to `go.work`. Either build into `cli/` as above, or use `-o /dev/null`.

## Environment variables

| Variable           | Description                                                  | Default |
|--------------------|--------------------------------------------------------------|---------|
| `CANVUS_API_URL`   | Canvus server URL (canonical; matches SDK + monorepo standard) | —     |
| `CANVUS_URL`       | Deprecated alias for `CANVUS_API_URL` (kept for back-compat) | —       |
| `CANVUS_API_KEY`   | Private-Token value (preferred auth)                         | —       |
| `CANVUS_USERNAME`  | Username for login-based auth                                | —       |
| `CANVUS_PASSWORD`  | Password for login-based auth                                | —       |
| `CANVUS_INSECURE`  | `true` to skip TLS verification (opt-in)                     | `false` |
| `CANVUS_OUTPUT`    | Default output format (`json`\|`yaml`\|`table`\|`text`)      | `table` |
| `CANVUS_VERBOSE`   | `true` to enable Debug-level slog output                     | `false` |
| `LOG_FORMAT`       | `json` to emit slog JSON to stderr                           | text    |
| `LOG_LEVEL`        | Honoured by slog handlers initialised from `util.InitLogger` | info    |

A `~/.canvus/config.yaml` file can supply the same values; the precedence is
flags > env > file > defaults.

## Differences from the legacy `github.com/jaypaulb/canvus-cli`

The source repo has 13,440 Go LOC across 187 files. The port carries the
structural skeleton verbatim and trims only what no longer compiles or is
unreachable against the new SDK / real Canvus API. Behavioural changes:

| Command                                  | Change                                                                                                                                     | Reason                                                                                                                                                   |
|------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| `canvus widget pin <widget-id>`          | Now `canvus widget pin <canvas-id> <widget-id>` and implemented as `UpdateWidget({pinned:true})`                                           | The legacy `/widgets/{id}/pin` endpoint does not exist in the public Canvus API; `pinned` is a widget attribute (see `endpoints/widgets.md`).            |
| `canvus widget unpin <widget-id>`        | Now `canvus widget unpin <canvas-id> <widget-id>` and implemented as `UpdateWidget({pinned:false})`                                        | Same as `pin`.                                                                                                                                           |
| `canvus widget copy <widget-id> <dest>`  | Now `canvus widget copy <source-canvas-id> <widget-id> <target-canvas-id>` and implemented via `CloneWidget`                               | The legacy `/widgets/{id}/copy` endpoint does not exist; the SDK's `CloneWidget` POSTs to the destination canvas's type-specific create endpoint.        |
| `canvus widget move <widget-id> <dest>`  | Now `canvus widget move <source-canvas-id> <widget-id> <target-canvas-id>` and implemented as `CloneWidget` followed by `DeleteWidget`     | The legacy `/widgets/{id}/move` endpoint does not exist; the SDK exposes no atomic cross-canvas move.                                                    |
| `canvus videooutput update`              | First positional renamed `<canvas-id>` → `<client-id>`; implemented via `SetVideoOutputSourceByID`                                         | The endpoint is keyed on the hosting client, not on a canvas; the new SDK exposes the operation only through `SetVideoOutputSourceByID`.                 |
| `canvus client create / update / delete` | Marked **hidden** and return `"…not supported by the Canvus API (Phase 4d gap)"`                                                           | The Canvus API has no client-mutation endpoints — clients are processes the server discovers, not records callers POST. The commands are preserved as command-tree stubs for legacy compatibility; if a real endpoint surfaces we will wire it in Phase 4d. |
| `cmd/canvus/main.go`                     | Hoisted to `go/cli/main.go` (single binary, no nested `cmd/`)                                                                              | Monorepo convention.                                                                                                                                     |
| `internal/util/logging.go`               | Rewritten as a thin shim over `log/slog`. `InitLogger` calls `slog.SetDefault` so the SDK and CLI agree on level/format.                   | `docs/conventions/go.md` §5 — no third-party loggers; slog only.                                                                                         |
| `internal/session/session.go`            | Uses `WithAPIKey` directly; the insecure-TLS path is opt-in (only fires when `cfg.Insecure` is set) instead of constructing a transport unconditionally. | `docs/conventions/go.md` and the audit's "trim bespoke insecure-TLS HTTP-client path now that `WithAPIKey` defaults to it".                              |

The `test/` directory from the source (live-server integration smoke tests) is
**not** ported. Unit tests are colocated with the source they exercise per
`docs/conventions/go.md` §8.

## Known limits

- No `import` / `export` subcommands yet. The SDK exposes
  `ExportWidgetsToFolder` / `ImportWidgetsToRegion`; wiring these into a
  `canvus canvas export <id> <dir>` / `canvus canvas import <id> <dir>` pair
  is the most natural next addition.
- The TUI (`canvus tui`, 319 LOC) was hoisted verbatim. The audit flagged it
  as borderline-large; it is not on the critical path for Phase 4c.
- `MockSession` (`internal/commands/testing/mock_session.go`, ~400 LOC) is a
  hand-rolled mock. The audit notes "could be replaced by interface-based
  test seams"; it is left in place because it works and is exercised by the
  existing test suite.

## SDK gaps to address in Phase 4d

- `CreateClient` / `UpdateClient` / `DeleteClient` — verify whether the
  Canvus API exposes these and either un-stub the commands or add them to
  the SDK.
- `UpdateVideoOutput` — confirm that `SetVideoOutputSourceByID` covers every
  field the legacy CLI's `--json` flag was meant to update; if not, add a
  generic update helper to the SDK.

## Tests

```bash
# From the monorepo go/ root
GOMAXPROCS=2 go test -count=1 ./cli/...
GOMAXPROCS=2 go vet ./cli/...
GOMAXPROCS=2 go test -count=1 -race ./cli/...
```

87 test functions across 25 test files at the time of the port; all pass
unit and race builds. Live-server tests are intentionally not part of this
suite.
