# Go Conventions — MT-Canvus-Tools

**Status:** Locked
**Applies to:** All Go code under `go/` in the MT-Canvus-Tools monorepo
**Audience:** Phase 4b refresh agents, future contributors (human or agent)

This document is the binding reference for every Go module in the repo. The decisions here are **defaults that do not require approval to apply**. If a refresh agent encounters a question this doc does not answer, the agent makes a reasonable modern-idiomatic choice and records it in the [Amendments](#amendments) section at the bottom of this file.

---

## 1. Module structure

### Workspace layout

A single Go workspace anchors all modules in the repo:

```
go/
├── go.work                         # Workspace root
├── sdk/                            # github.com/jaypaulb/MT-Canvus-Tools/go/sdk
│   ├── go.mod
│   ├── canvus/                     # Public package: SDK entry point
│   └── internal/                   # SDK-private helpers
├── cli/                            # github.com/jaypaulb/MT-Canvus-Tools/go/cli
│   ├── go.mod
│   ├── cmd/canvus/                 # main package
│   └── internal/
├── examples/
│   ├── core/
│   │   ├── 01-auth-and-list/       # github.com/jaypaulb/MT-Canvus-Tools/go/examples/core/01-auth-and-list
│   │   │   ├── go.mod
│   │   │   └── main.go
│   │   └── ...
│   └── projects/
│       ├── ai-personas/
│       ├── llm-canvas-companion/
│       └── note-mapper/
└── tools/
    ├── mcp-server/
    ├── powertoys/
    ├── translator/
    └── db-solver/
```

### Module paths

All modules live under the path prefix `github.com/jaypaulb/MT-Canvus-Tools/go/`:

| Directory | Module path |
| --- | --- |
| `go/sdk` | `github.com/jaypaulb/MT-Canvus-Tools/go/sdk` |
| `go/cli` | `github.com/jaypaulb/MT-Canvus-Tools/go/cli` |
| `go/examples/core/03-widget-crud` | `github.com/jaypaulb/MT-Canvus-Tools/go/examples/core/03-widget-crud` |
| `go/tools/mcp-server` | `github.com/jaypaulb/MT-Canvus-Tools/go/tools/mcp-server` |

The SDK is imported as `github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus`. Downstream modules pin a workspace-local version via `go.work` during development and a tagged version (`v0.1.0`, etc.) for release.

### `go.work` example

```
go 1.22

use (
    ./sdk
    ./cli
    ./examples/core/01-auth-and-list
    ./examples/core/02-auth-flows
    ./examples/core/03-widget-crud
    ./examples/core/04-file-upload
    ./examples/core/05-streaming
    ./examples/core/06-llm-integration
    ./examples/core/07-webhooks-notifications
    ./examples/core/08-cross-canvas-clone
    ./examples/projects/ai-personas
    ./examples/projects/llm-canvas-companion
    ./examples/projects/note-mapper
    ./tools/mcp-server
    ./tools/powertoys
    ./tools/translator
    ./tools/db-solver
)
```

**Do not commit `go.work.sum`.** Add it to `.gitignore`. Each module's `go.sum` is the canonical record for its dependencies.

### Internal vs. public packages

- `internal/` directories are non-importable outside the parent module. Put helpers, private types, and implementation details here.
- The **public API surface of the SDK is the single package `canvus/`**, mounted at `go/sdk/canvus`. Do not expose sub-packages of the SDK for external consumption — they all live under `go/sdk/internal/` so the import surface stays one line:

  ```go
  import "github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
  ```

- Tools and examples use whatever layered structure makes sense locally. Atomic design (atoms/molecules/organisms) is **permitted but not required** at the tool level. The SDK itself uses flat package layout organised by API resource.

### Why monolithic SDK package?

The existing Canvus-Go-API ships everything under `canvus/`: one type (`Session`) with methods for every endpoint. Sticking with this is intentional — splitting it would force consumers to import five packages to do basic work. The trade-off (a larger single package) is acceptable because the SDK is generated against a single, well-defined API surface and methods naturally group by resource via file naming (`canvases.go`, `notes.go`, `images.go`).

---

## 2. Go version

### Minimum

**Go 1.22.** Declared in every `go.mod`:

```
module github.com/jaypaulb/MT-Canvus-Tools/go/sdk

go 1.22

toolchain go1.23.0
```

The `toolchain` directive ensures contributors pulling the repo with an older Go install automatically download a compatible toolchain. Use the latest patch release of 1.23 as the toolchain.

### CI matrix

CI builds and tests against Go **1.22** and **1.23**. 1.22 is the floor (kept stable for downstream consumers); 1.23 is the head (catches deprecation warnings from the current stable release). When Go 1.24 ships and stabilises, drop 1.22 and add 1.24 in a single PR.

### Why 1.22 and not later?

1.22 introduced loop variable scoping changes and `range` over integers — both are now relied on by idiomatic 2026 Go. Pinning lower (e.g. 1.21) forces SDK code to avoid these features for no benefit; pinning higher (e.g. 1.23+) cuts out enterprise environments that update Go on a 12-month cadence.

---

## 3. Build & tooling

### Build

```bash
# From any module directory
go build ./...

# From the workspace root (builds all modules)
cd go && go build ./...
```

No build tags are required for normal builds. Build tags are reserved for integration tests (see [Testing](#7-testing)) and platform-specific tools (e.g., `//go:build windows` for Windows-only code in `tools/powertoys`).

### Format

```bash
gofmt -s -w .
```

The `-s` flag applies simplification passes (e.g., collapsing redundant type declarations in composite literals). This is the **only** formatter run. `goimports` is enabled via golangci-lint (see [Linting](#4-linting)), so a separate invocation is unnecessary.

A pre-commit Git hook is encouraged but not enforced:

```bash
# .git/hooks/pre-commit
#!/bin/sh
gofmt -s -l . | grep . && { echo "gofmt issues — run: gofmt -s -w ."; exit 1; }
exit 0
```

### Vet

```bash
go vet ./...
```

Run vet on every module before commit. CI fails the build on any vet finding. `go vet` is fast and catches real bugs (printf format mismatches, lock copies, unreachable code); there is no excuse to skip it.

### Module hygiene

```bash
go mod tidy        # Remove unused dependencies, add missing ones
go mod verify      # Verify downloaded modules match go.sum
```

Run `go mod tidy` before every commit that adds or removes an import. CI fails if `go.mod` or `go.sum` is dirty after a tidy.

---

## 4. Linting

### Tool

**golangci-lint v1.60+**, installed via:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.3
```

### Config

A single `.golangci.yml` lives at the **repo root** and applies to every Go module via the workspace. All modules share one set of rules — divergence by module is explicitly disallowed.

```yaml
# .golangci.yml — applies to all Go modules in MT-Canvus-Tools
run:
  timeout: 5m
  go: "1.22"
  tests: true

linters:
  disable-all: true
  enable:
    - errcheck       # Find unchecked errors
    - gosimple       # Simplification suggestions
    - govet          # Standard vet checks
    - ineffassign    # Detect ineffectual assignments
    - staticcheck    # Comprehensive static analysis
    - unused         # Find unused code
    - gofmt          # Enforce gofmt -s
    - goimports      # Import grouping and ordering
    - revive         # Replacement for golint (style)
    - gosec          # Security audit

linters-settings:
  gofmt:
    simplify: true
  goimports:
    local-prefixes: github.com/jaypaulb/MT-Canvus-Tools
  revive:
    severity: warning
    rules:
      - name: exported       # Require godoc on exported identifiers
      - name: var-naming
      - name: error-return
      - name: error-naming
      - name: context-as-argument
      - name: context-keys-type
  gosec:
    excludes:
      - G104   # Audited unhandled errors — errcheck covers this
      - G304   # File path provided as taint — too noisy for CLI/tools

issues:
  exclude-rules:
    # Test files: allow longer functions, missing godoc, gosec false positives on hardcoded test data
    - path: _test\.go
      linters:
        - revive
        - gosec
        - errcheck
  max-issues-per-linter: 0
  max-same-issues: 0
```

### Running

```bash
golangci-lint run ./...
```

CI runs this on every push. A failing lint blocks merge.

### Why this set?

The nine linters above catch ~95% of real-world Go defects without producing noisy false positives. Adding more (e.g., `gocyclo`, `funlen`, `dupl`) is a trap: they generate complaint storms in legitimate code that demand `//nolint` annotations everywhere. If a refresh agent finds a pattern across the codebase that warrants new linter coverage, it should be raised as an amendment and discussed, not silently added.

---

## 5. Logging

### Library

**Standard library `log/slog`.** No third-party logging libraries (zap, zerolog, logrus, etc.) anywhere in the repo.

### Default logger

The SDK and every tool obtain their logger from `slog.Default()`. The CLI entry point (or tool `main()`) configures it once at startup:

```go
package main

import (
    "log/slog"
    "os"
)

func main() {
    setupLogging()
    // ...
}

func setupLogging() {
    var handler slog.Handler
    switch os.Getenv("LOG_FORMAT") {
    case "json":
        handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
            Level: logLevel(),
        })
    default:
        handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
            Level: logLevel(),
        })
    }
    slog.SetDefault(slog.New(handler))
}

func logLevel() slog.Level {
    switch os.Getenv("LOG_LEVEL") {
    case "debug":
        return slog.LevelDebug
    case "warn":
        return slog.LevelWarn
    case "error":
        return slog.LevelError
    default:
        return slog.LevelInfo
    }
}
```

### Production vs. development

- **`LOG_FORMAT=json`** — JSON handler. Use in production, containers, and any deployed tool. Output is structured, parseable by log aggregators.
- **`LOG_FORMAT=` (unset or anything else)** — text handler. Human-readable for local development and CLI usage.

Output goes to **stderr**, never stdout. Stdout is reserved for tool output that callers may pipe (e.g., the CLI's JSON canvas list).

### Levels

Four levels, used as follows:

| Level | When to use |
| --- | --- |
| `Debug` | Wire-level details: request URLs, response sizes, retry attempts. Off by default. |
| `Info` | Significant lifecycle events: server started, batch operation completed, user authenticated. |
| `Warn` | Recoverable problems: retrying a request, falling back from one auth method to another. |
| `Error` | Operations that failed and were surfaced to the caller. The error itself is logged; the caller decides what to do. |

### Idiomatic call sites

Use the structured key/value form, not `Sprintf`:

```go
// Good
slog.Info("canvas created", "canvas_id", canvas.ID, "owner", canvas.Owner)

// Bad — loses structure
slog.Info(fmt.Sprintf("canvas %s created by %s", canvas.ID, canvas.Owner))
```

Attach a logger to a context (`slog.With`) when threading through a long operation:

```go
log := slog.With("operation", "batch_move", "batch_id", batchID)
log.Info("starting", "item_count", len(items))
// ... later
log.Error("item failed", "item_id", id, "error", err)
```

### Library code (SDK)

The SDK logs at `Debug` only. **Anything above Debug from inside the SDK is a bug** — the SDK is a library, not an application, and noisy default output is hostile to embedders. Errors are returned, not logged.

---

## 6. Error handling

### Wrapping

Every error returned from a function should be wrapped with context using `%w`:

```go
// Good
func (s *Session) GetCanvas(ctx context.Context, id string) (*Canvas, error) {
    var canvas Canvas
    if err := s.doRequest(ctx, "GET", "canvases/"+id, nil, &canvas); err != nil {
        return nil, fmt.Errorf("GetCanvas: %w", err)
    }
    return &canvas, nil
}

// Bad — drops the chain, callers can't use errors.Is
func (s *Session) GetCanvas(ctx context.Context, id string) (*Canvas, error) {
    var canvas Canvas
    if err := s.doRequest(ctx, "GET", "canvases/"+id, nil, &canvas); err != nil {
        return nil, fmt.Errorf("GetCanvas: %v", err)
    }
    return &canvas, nil
}

// Also bad — no context, debugging requires guessing
func (s *Session) GetCanvas(ctx context.Context, id string) (*Canvas, error) {
    var canvas Canvas
    if err := s.doRequest(ctx, "GET", "canvases/"+id, nil, &canvas); err != nil {
        return nil, err
    }
    return &canvas, nil
}
```

The prefix should be the function name (`"GetCanvas: %w"`) or a brief operation description (`"decode response: %w"`). It should not duplicate information already in the wrapped error.

### Sentinel errors

For known failure modes that callers will branch on, declare sentinel `var` values:

```go
package canvus

import "errors"

var (
    ErrNotFound       = errors.New("not found")
    ErrUnauthorized   = errors.New("unauthorized")
    ErrConflict       = errors.New("conflict")
    ErrRateLimited    = errors.New("rate limited")
)
```

Callers branch with `errors.Is`:

```go
canvas, err := session.GetCanvas(ctx, id)
if errors.Is(err, canvus.ErrNotFound) {
    return fmt.Errorf("canvas %q does not exist", id)
}
if err != nil {
    return err
}
```

### Structured error types

Define a Go error type **only** when callers need access to structured fields beyond a message. For the SDK, the canonical structured error is `APIError`:

```go
// APIError represents an error returned by the Canvus API.
type APIError struct {
    StatusCode int               // HTTP status code
    Code       string            // Machine-readable error code (e.g., "validation_error")
    Message    string            // Human-readable message
    RequestID  string            // Server-side request ID, if available
    Details    map[string]any    // Additional structured details
}

func (e *APIError) Error() string {
    if e.RequestID != "" {
        return fmt.Sprintf("API error %d (%s): %s [request_id=%s]",
            e.StatusCode, e.Code, e.Message, e.RequestID)
    }
    return fmt.Sprintf("API error %d (%s): %s", e.StatusCode, e.Code, e.Message)
}

// Is allows matching on status code and/or code.
func (e *APIError) Is(target error) bool {
    t, ok := target.(*APIError)
    if !ok {
        return false
    }
    if t.StatusCode != 0 && e.StatusCode != t.StatusCode {
        return false
    }
    if t.Code != "" && e.Code != t.Code {
        return false
    }
    return true
}
```

Callers inspect via `errors.As`:

```go
var apiErr *canvus.APIError
if errors.As(err, &apiErr) {
    if apiErr.StatusCode == 429 {
        // back off
    }
}
```

**Do not** define a new error type when a sentinel will do. **Do not** define a generic `Error` struct with optional fields "in case someone needs them later" — add fields when a real caller needs them.

### What not to do

```go
// Bad: silent fallback hides the failure
canvases, err := session.ListCanvases(ctx, nil)
if err != nil {
    canvases = []Canvas{}   // ← the caller now believes there are zero canvases
}

// Bad: stringly-typed error inspection
if strings.Contains(err.Error(), "not found") {
    // ...
}

// Bad: panic on remote error
canvases, err := session.ListCanvases(ctx, nil)
if err != nil {
    panic(err)   // ← libraries do not panic on operational errors
}
```

The first pattern is the most dangerous because it converts informative failure into expensive silent corruption. See the global "On Fallbacks" rule — fail loudly.

---

## 7. Configuration

### Precedence

Configuration is resolved in this order, highest to lowest:

1. **Command-line flags** (e.g., `--api-key`, `--base-url`)
2. **Environment variables** with the `CANVUS_` prefix
3. **Config file** (YAML, location below)
4. **Built-in defaults**

A later layer never silently overrides an earlier one. If a flag is passed, it wins. If only an env var is set, the env var wins. If nothing is set, the default wins.

### Config file format

**YAML**, parsed with `gopkg.in/yaml.v3`. No TOML, no JSON for configs, no INI. The CLI's default location is `~/.canvus/config.yaml`; tools may use their own locations, documented per-tool.

```yaml
# ~/.canvus/config.yaml
base_url: https://canvus.example.com/api/v1
api_key: redacted-or-loaded-from-env

output: table       # table | json | yaml
timeout: 30         # seconds
insecure: false     # skip TLS verification (dev only)
```

### Environment variables

All env vars use the `CANVUS_` prefix:

| Variable | Maps to |
| --- | --- |
| `CANVUS_BASE_URL` | `base_url` |
| `CANVUS_API_KEY` | `api_key` |
| `CANVUS_USERNAME` | `username` |
| `CANVUS_PASSWORD` | `password` |
| `CANVUS_OUTPUT` | `output` |
| `CANVUS_TIMEOUT` | `timeout` |
| `CANVUS_INSECURE` | `insecure` |
| `CANVUS_CONFIG` | path to alternate config file |

### Loading config

The recommended pattern is a small loader that takes precedence layers as explicit parameters:

```go
// Config holds all settings.
type Config struct {
    BaseURL  string `yaml:"base_url"`
    APIKey   string `yaml:"api_key"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
    Output   string `yaml:"output"`
    Timeout  int    `yaml:"timeout"`
    Insecure bool   `yaml:"insecure"`
}

// Load resolves config from flags, env, file, and defaults in that order.
func Load(flags *Config, configPath string) (*Config, error) {
    cfg := defaults()
    if fileCfg, err := loadFile(configPath); err == nil {
        cfg.merge(fileCfg)
    } else if !os.IsNotExist(err) {
        return nil, fmt.Errorf("read config: %w", err)
    }
    cfg.merge(loadEnv())
    cfg.merge(flags)
    return cfg, cfg.validate()
}
```

Viper is **not** used. It pulls in 30+ transitive dependencies and obscures precedence behind global state. A direct YAML-and-env loader is ~150 lines, fully testable, and easier to debug.

---

## 8. Testing

### Library

**Standard `testing` package + `github.com/stretchr/testify/assert`** for assertions. Testify is the one allowed test helper because its `assert` and `require` packages produce readable failure messages without surprising magic. Do not pull in additional helpers (Ginkgo, GoConvey, etc.).

```go
import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestParseCanvasID(t *testing.T) {
    id, err := ParseCanvasID("c_abc123")
    require.NoError(t, err)            // require: stop on failure
    assert.Equal(t, "abc123", id)      // assert: continue on failure
}
```

Use `require` for preconditions whose failure makes the rest of the test meaningless. Use `assert` for independent checks.

### Table-driven tests

For any function with more than two test cases, use table-driven tests:

```go
func TestNormalizeWidgetType(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        want     string
        wantErr  bool
    }{
        {"lower", "note", "note", false},
        {"upper", "NOTE", "note", false},
        {"mixed", "Image", "image", false},
        {"unknown", "blorp", "", true},
        {"empty", "", "", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NormalizeWidgetType(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

The `name` field becomes the subtest name (`TestNormalizeWidgetType/lower`), making targeted re-runs easy: `go test -run TestNormalizeWidgetType/upper`.

### File layout

- Unit tests live in `*_test.go` files **next to the source** they test (`canvases.go` and `canvases_test.go` in the same package).
- The test file is in the same package (`package canvus`, not `package canvus_test`) so it can exercise unexported helpers. Use `_test` package suffix only when explicitly testing the public API surface.

### Integration tests

Integration tests that require a live Canvus server are gated by a **build tag**:

```go
//go:build integration

package canvus_test

import "testing"

func TestLiveCanvasCRUD(t *testing.T) {
    // ...
}
```

Run with:

```bash
go test -tags=integration ./...
```

Without the tag, `go test ./...` runs only unit tests. CI runs the integration suite in a separate job that only fires on the main branch or when explicitly requested via a PR label.

### Test configuration

Integration tests load credentials from `settings.json` at the module root (gitignored), with the same shape used by Canvus-Go-API today:

```json
{
  "api_base_url": "https://server/api/v1/",
  "api_key": "your-api-key",
  "test_canvas_id": "canvas-id-for-tests"
}
```

A `settings.example.json` is checked in as a template.

### Coverage

| Target | Threshold |
| --- | --- |
| `go/sdk` | **70%** line coverage |
| `go/cli` | No hard target, but tests must exist for command parsing |
| `go/examples/**` | No coverage requirement |
| `go/tools/**` | No coverage requirement (tools can be tested by use, but if shipped to others, aim for 50%+) |

CI fails the SDK job if coverage drops below 70%. Coverage is reported via:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total
```

70% is a deliberate floor, not a goal. The SDK's value is in code that talks to a remote server; mocking the server to chase 95% coverage produces tests that pass while the SDK is broken.

---

## 9. CI

### Workflow

A single GitHub Actions workflow `.github/workflows/go.yml` covers all Go modules:

```yaml
name: Go

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        go: ["1.22", "1.23"]
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}
          cache: true

      - name: Verify modules
        run: |
          cd go
          go work sync
          go mod download

      - name: Lint
        if: matrix.go == '1.23'
        uses: golangci/golangci-lint-action@v6
        with:
          version: v1.60.3
          working-directory: go

      - name: Vet
        working-directory: go
        run: go vet ./...

      - name: Test
        working-directory: go
        run: go test -race -coverprofile=coverage.out ./...

      - name: Coverage gate (SDK)
        if: matrix.go == '1.23'
        working-directory: go/sdk
        run: |
          go test -coverprofile=coverage.out ./...
          pct=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
          awk -v p="$pct" 'BEGIN { if (p+0 < 70) { print "SDK coverage " p "% < 70%"; exit 1 } }'

      - name: Build
        working-directory: go
        run: go build ./...
```

### Notes

- **Ubuntu only.** Cross-platform builds for end-user binaries (Windows, macOS) live in per-tool release workflows (`go/tools/powertoys/.github/workflows/release.yml`), not the shared `go.yml`. Keeping the shared workflow Linux-only cuts CI time roughly by 2/3.
- **Race detector on by default** (`-race`). Race conditions in HTTP client code or batch operations are nasty and cheap to detect.
- **Lint runs once** on the highest Go version in the matrix to keep CI fast; vet and test run on every version.

---

## 10. Code style

### Godoc

Every exported identifier (type, function, method, constant, var) must have a godoc comment. The comment **must start with the identifier name**:

```go
// CreateCanvas creates a new canvas with the given request payload.
// The request may be a CreateCanvasRequest or a map[string]any for advanced cases.
func (s *Session) CreateCanvas(ctx context.Context, req any) (*Canvas, error) {
    // ...
}

// Canvas represents a single Canvus canvas as returned by the API.
type Canvas struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    State string `json:"state"`
}

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("not found")
```

This is not a stylistic preference — `golangci-lint`'s `revive` linter rejects the file without it.

### Receiver names

Short and consistent: 1–2 letters derived from the type name. Every method on the same type uses the same receiver name.

```go
// Good — consistent
func (s *Session) ListCanvases(ctx context.Context) ([]Canvas, error) { ... }
func (s *Session) GetCanvas(ctx context.Context, id string) (*Canvas, error) { ... }

// Bad — inconsistent
func (s *Session) ListCanvases(ctx context.Context) ([]Canvas, error) { ... }
func (sess *Session) GetCanvas(ctx context.Context, id string) (*Canvas, error) { ... }

// Bad — too long, reads like a parameter
func (session *Session) ListCanvases(ctx context.Context) ([]Canvas, error) { ... }
```

Avoid `self`, `this`, `me`. Avoid single-letter generics like `x`.

### No global mutable state

Libraries (SDK, internal packages) **must not** hold mutable state in package-level variables. Configuration, HTTP clients, caches — all attached to a value (`Session`, `Client`, etc.) that callers create and pass around.

```go
// Good
session := canvus.NewSession(baseURL, canvus.WithAPIKey(key))
canvases, _ := session.ListCanvases(ctx)

// Bad
canvus.SetAPIKey(key)             // package-level state, untestable in parallel
canvases, _ := canvus.ListCanvases(ctx)
```

Exceptions: `slog.Default()` (which is itself a configurable global) and `init()` functions that register effectively-constant data (e.g., type adapters). If in doubt, attach to a struct.

### Context

Every public function that does I/O — network calls, file reads, subprocess execution — takes `context.Context` as the **first argument**:

```go
// Good
func (s *Session) GetCanvas(ctx context.Context, id string) (*Canvas, error)

// Bad — caller can't cancel or set a timeout
func (s *Session) GetCanvas(id string) (*Canvas, error)

// Bad — context not first
func (s *Session) GetCanvas(id string, ctx context.Context) (*Canvas, error)
```

Functions that are pure computation (formatters, parsers, geometry helpers) do not take a context.

Do not store a `context.Context` in a struct field. If a long-lived object needs to cancel work, take a context in the method that *starts* the work.

### Naming

- **Package names** are short, lowercase, single-word: `canvus`, `config`, `session`. No underscores. No camelCase.
- **Exported names** are `CamelCase`. Acronyms are uppercase: `URL`, `ID`, `HTTP`, `API` — not `Url`, `Id`, `Http`, `Api`.
- **Test functions** are `TestXxx(t *testing.T)`. Subtest names use snake-case or kebab-case, whichever reads better.
- **Variables** are short within small scope, longer in larger scope. A loop index is `i`; a session passed across 30 lines is `session`.

### Imports

Three groups, separated by blank lines, in this order:

```go
import (
    "context"
    "fmt"
    "net/http"

    "github.com/stretchr/testify/assert"
    "gopkg.in/yaml.v3"

    "github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)
```

1. Standard library
2. Third-party
3. Repo-local (`github.com/jaypaulb/MT-Canvus-Tools/...`)

`goimports` (run via golangci-lint with `local-prefixes` set) handles this automatically. Do not group imports by any other scheme.

### Avoid `interface{}` / `any` in public APIs

The Canvus API genuinely has polymorphic request payloads (widget creation accepts notes, images, PDFs, etc.). When a public function legitimately accepts polymorphism, use `any`, not `interface{}`:

```go
// Good — modern Go (1.18+)
func (s *Session) CreateWidget(ctx context.Context, canvasID string, req any) (Widget, error)

// Bad — legacy spelling, predates 1.18
func (s *Session) CreateWidget(ctx context.Context, canvasID string, req interface{}) (Widget, error)
```

If a strongly typed signature works, prefer it:

```go
// Better still — typed when possible
func (s *Session) CreateNote(ctx context.Context, canvasID string, req CreateNoteRequest) (*Note, error)
```

### File length

There is no hard line limit on Go files. The SDK's `session.go` is 24KB and that is fine — it is one logical unit. But if a file is reaching for 1000+ lines and contains genuinely independent concerns, split it. Use judgment, not a number.

### What not to do

- No `panic` in libraries on operational errors. Panic only on programmer errors (`nil` map access, slice-out-of-bounds in unreachable code).
- No `init()` that does network I/O, file I/O, or anything that can fail.
- No dot imports (`import . "fmt"`).
- No named return values unless they materially clarify the function. `func Foo() (count int, err error)` is fine; `func Foo() (a int, b int, c int)` is not.

---

## 11. Convention amendments

When a Phase 4b refresh agent encounters a decision not covered above, it **must** append the decision and rationale to this section. Format:

### YYYY-MM-DD — One-line summary

**Item:** *Tool, module, or area being refreshed*
**Decision:** *What was chosen*
**Rationale:** *Why — including alternatives considered*

---

<!-- New amendments are appended below this line in reverse chronological order. -->

### 2026-05-18 — WithSubscribeBuffer adds configurable channel capacity (Phase 4d Round B)

**Item:** `go/sdk/canvus/subscribe.go` + `options_subscribe.go` + `options.go`
**Decision:** `SessionConfig.SubscribeBuffer int` (default 4) controls the `make(chan T, N)` capacity in `subscribeStream`. `WithSubscribeBuffer(n int)` sets it; panics if n < 1. The `WithSubscribeBuffer` function lives in the new `options_subscribe.go` to avoid conflicts with parallel B1 work on `options.go`.
**Rationale:** High-throughput consumers (live dashboards, ai-personas) hit backpressure with buffer=4 when the stream produces faster than the consumer drains. Exposing the size as a `SessionConfigOption` keeps defaults unchanged and lets performance-sensitive callers raise it without forking the session type.

---

### 2026-05-18 — CLI tools may use viper precedence instead of FromEnv (Phase 4d)

**Item:** `go/cli` configuration loading (`go/cli/internal/config/config.go`)
**Decision:** The CLI binary uses `github.com/spf13/viper` to resolve configuration in flag > env > file > defaults order, rather than the SDK's `canvus.FromEnv` constructor.
**Rationale:** `FromEnv` is an environment-only constructor; it has no mechanism to accept command-line flag overrides or a YAML config file. CLI binaries inherently require a full precedence chain (flags win over env vars, env vars win over file, file wins over compiled defaults) so that interactive users can override persistent settings at the call site without modifying their environment or config file. `go/cli` is the canonical example of this pattern. Any future CLI-style tool in `go/tools/` may follow the same approach when it accepts flags that override env vars. **Mandatory constraint:** the CLI MUST alias `CANVUS_API_URL` (canonical name, matching the SDK and `.secrets` convention) and `CANVUS_URL` (deprecated alias) via an explicit `viper.BindEnv("url", "CANVUS_API_URL", "CANVUS_URL")` call inside `Load()`, so that env-var precedence matches SDK behaviour and users switching between the SDK and the CLI observe consistent variable names. `go/cli/internal/config/config.go` implements this at the top of `Load()`.
