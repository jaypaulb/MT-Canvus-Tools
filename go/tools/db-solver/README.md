# db-solver

A Cobra CLI tool that recovers from corrupted Canvus-Server installations.

It queries the Canvus API for all media assets across every canvas, compares the
resulting hash list against the on-disk asset store, looks up missing hashes in
the Canvus PostgreSQL database (`asset_files` table), and restores files from
configured backup folders.

## Monorepo path

`go/tools/db-solver/` — module `github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver`

## Prerequisites

| Requirement | Notes |
|---|---|
| Canvus Server | Accessible from the machine running the tool (typically run directly on the server) |
| PostgreSQL | `asset_files` table reachable; credentials loaded from `mt-canvus-server.ini` |
| Assets folder | The live Canvus asset store (read + write for restoration) |
| Backup folder | One or more backup copies of the asset store (read-only) |

## Build

```bash
# From the workspace root:
cd go
go build ./tools/db-solver/...

# Binary appears at: $GOPATH/bin/db-solver (or run with go run)
go run ./tools/db-solver/cmd/db-solver -- --help
```

## Configuration

The tool loads configuration in this priority order:

1. `--config <path>` flag
2. `CANVUS_API_URL` / `CANVUS_USERNAME` / `CANVUS_PASSWORD` environment variables
3. `./config.yaml` (or `$CANVUS_CONFIG`)
4. Interactive prompts (if no file or env vars are found)

### Minimal `config.yaml`

```yaml
canvus_server:
  url: "https://localhost:443"
  username: "admin@example.com"
  password: "changeme"
  insecuretls: true            # true for self-signed certificates (common)

paths:
  assetsfolder: "/var/lib/mt-canvus-server/assets"
  backuprootfolder: "/var/backups/canvus"
  outputfolder: "./reports"

logging:
  level: "info"                # debug | info | warn | error
  verbose: false
  logtofile: true
  logfile: "canvus-server-db-solver.log"

performance:
  maxconcurrentapi: 10
  maxconcurrentfiles: 20
```

### Environment variables

| Variable | Config equivalent |
|---|---|
| `CANVUS_API_URL` | `canvus_server.url` |
| `CANVUS_BASE_URL` | `canvus_server.url` (fallback) |
| `CANVUS_USERNAME` | `canvus_server.username` |
| `CANVUS_PASSWORD` | `canvus_server.password` |

## Subcommands

### `discover`

Scans all canvases via the Canvus API, compares discovered asset hashes against
the local assets folder, and produces:

- `reports/missing_assets_report.txt` — detailed canvas-grouped report
- `reports/missing_assets.csv` — spreadsheet-ready export

```bash
db-solver discover --config config.yaml
```

### `run`

Runs the full workflow: discover → backup search → report.

```bash
db-solver run --config config.yaml
```

### `lookup-hash`

Processes assets that have no hash value (typically very old widgets uploaded
before the server started computing hashes). Uses the PostgreSQL `asset_files`
table to resolve original filenames to hash pairs, then searches the assets and
backup folders.

```bash
db-solver lookup-hash --config config.yaml [flags]
```

Flags:

| Flag | Default | Description |
|---|---|---|
| `--dry-run` | false | Report what would be restored without copying files |
| `--ini-path` | auto-detect | Path to `mt-canvus-server.ini` (for DB credentials) |
| `--skip-archived` | true | Skip canvases in the trash during discovery |
| `--low-memory` | false | Search files on-demand instead of pre-building catalogs (slower, less RAM) |

The tool auto-detects `mt-canvus-server.ini` in the following locations:

- Linux: `/etc/MultiTaction/canvus/mt-canvus-server.ini`, `/opt/MultiTaction/…`
- Windows: `C:\ProgramData\MultiTaction\Canvus\mt-canvus-server.ini`

## Recovery workflow

1. **Run `discover`** to see which assets are missing from the on-disk store.
2. **Inspect the reports** in `./reports/`. Each missing asset shows whether a
   backup copy was found and where it lives.
3. **Run `lookup-hash`** (optionally with `--dry-run` first) to process assets
   that lack hashes — these require a Postgres lookup to identify the correct file.
4. **Restore manually** from the reported backup paths, or implement the TODO
   restoration step in `restorer.go` for automated recovery.

## Operational guardrails

- Use `--dry-run` before any restoration to verify the tool finds the right files.
- The server-side OOM-risk validation (downloading every asset to check existence)
  is intentionally disabled; actual validation happens during the filesystem scan.
- `InsecureTLS: true` is the default because Canvus server deployments commonly
  use self-signed certificates. Set `insecuretls: false` in production when you
  have a valid certificate chain.

## Diff vs. source (`Canvus-Server-db-solver`)

| Area | Change |
|---|---|
| Vendored SDK (`src/pkg/canvus/`) | **Dropped entirely.** ~6,000 LOC gone. Replaced by six calls to `github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus`. |
| Import paths | `github.com/jaypaulb/canvus-server-db-solver/...` → `github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/...` |
| `lookup_hash.go` (744 LOC) | Split into three files: orchestration, per-asset processing, and report generation. |
| `internal/config` | Replaced Viper with a direct YAML+env loader (removes ~10 transitive deps). |
| `internal/logging` | Replaced custom `log.Logger` with `log/slog` underneath the same `Logger` API. |
| `src/` prefix | Flattened — source lived under `src/cmd/` and `src/internal/`; destination uses `cmd/` and `internal/` directly. |
| Binaries / releases | Not ported — `bin/`, `releases/`, `public/releases/` dropped. |
| `kpmg-paris-report/` | Sample data — not ported. |
| `restore_assets.ps1` | PowerShell helper — not ported (functionality documented in this README). |
| `docs/` planning material | Not ported. |
| `go.mod` | Switched from Go 1.24.5 (declared in source `go.mod`) to Go 1.22 (monorepo floor). |

## Known limits / Phase 4d candidates

- The restoration step in `lookup-hash` logs the target path but does not copy
  the file yet (marked `// TODO`). Full restore is implemented in `restorer.go`
  and can be wired up.
- The `--low-memory` on-demand search path does not fall back to recursive search
  in the assets folder (the pre-built catalog path does). Parity improvement
  deferred.
- No integration tests: connecting to a live Canvus server or Postgres requires
  external infrastructure. If you add them, gate behind `//go:build integration`.
