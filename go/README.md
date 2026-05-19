# MT-Canvus-Tools — Go

Go workspace containing the SDK, CLI, examples, and tools for the Canvus platform. Managed by `go.work` — all modules build and test from this directory.

## Structure

| Path | Description |
|---|---|
| [`sdk/`](sdk/) | Go SDK — 147 endpoints, 44 typed Subscribe helpers, extras subpackage |
| [`cli/`](cli/) | `canvus` CLI — thin wrapper around the SDK |
| [`examples/core/`](examples/core/) | 8 numbered examples (01 auth, 02 auth-flows, 03 widget-CRUD, 04 file-upload, 05 streaming, 06 LLM, 07 webhooks, 08 cross-canvas-clone) |
| [`examples/projects/`](examples/projects/) | 3 project examples (note-mapper, llm-canvas-companion, ai-personas) |
| [`tools/db-solver/`](tools/db-solver/) | Database import/export helper |
| [`tools/powertoys/`](tools/powertoys/) | Desktop tray app with WebUI and canvas relay tools |
| [`tools/translator/`](tools/translator/) | Real-time note translation via LLM |
| [`internal/llm/`](internal/llm/) | Shared Gemini helper used by project examples |

## Workspace commands

Run from this directory (`go/`) to operate across all modules:

```bash
go build ./...        # Build every module
go test ./...         # Test every module
go vet ./...          # Vet every module
gofmt -s -l .         # Check formatting — must print nothing
```

## Authentication

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key
```

The SDK sends the key as the `Private-Token` request header.

## Quick example

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func main() {
	cfg := canvus.DefaultSessionConfig()
	cfg.BaseURL = os.Getenv("CANVUS_API_URL")
	session := canvus.NewSession(cfg, canvus.WithAPIKey(os.Getenv("CANVUS_API_KEY")))

	canvases, err := session.ListCanvases(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range canvases {
		fmt.Printf("%s  %s\n", c.ID, c.Name)
	}
}
```

## Conventions

[`docs/conventions/go.md`](../docs/conventions/go.md) (monorepo root) — locked defaults for import order, error wrapping, logging, context propagation, and test structure.

## Getting started

[`docs/getting-started/go.md`](../docs/getting-started/go.md) — step-by-step setup, authentication, running examples.
