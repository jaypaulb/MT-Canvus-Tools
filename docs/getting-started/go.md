# Getting Started — Go

Get up and running with the MT-Canvus-Tools Go SDK.

## Prerequisites

- **Go 1.24 or later** — verify with `go version`
- A running Canvus server and an API key
- `git` (for running monorepo examples)

## Option A: Use the SDK as a dependency

```bash
go get github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus@latest
```

Create `main.go`:

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

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key
go run .
```

## Option B: Run examples from the monorepo

```bash
git clone https://github.com/jaypaulb/MT-Canvus-Tools.git
cd MT-Canvus-Tools/go
go build ./...
```

Set credentials, then run any numbered example:

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key

cd examples/core/01-auth-and-list && go run .
cd examples/core/05-streaming && go run .
```

## Environment variables

| Variable | Description |
|---|---|
| `CANVUS_API_URL` | Full base URL including `/api/v1` suffix (e.g. `https://canvus.example.com/api/v1`) |
| `CANVUS_API_KEY` | Long-lived API key — sent as `Private-Token` header |
| `CANVUS_INSECURE_TLS` | Set to `1` to disable TLS verification for self-signed certificates |

## Authentication modes

```go
// API key (recommended for services and automation)
session := canvus.NewSession(cfg, canvus.WithAPIKey(os.Getenv("CANVUS_API_KEY")))

// Username + password (interactive flows; short-lived token)
session := canvus.NewSession(cfg)
if err := session.Login(ctx, "user@example.com", "password"); err != nil {
    log.Fatal(err)
}
defer session.Logout(ctx)
```

## Real-time streaming

The SDK exposes a typed Subscribe helper for every streamable endpoint:

```go
eventChan, errChan, err := session.SubscribeCanvases(ctx, nil)
if err != nil {
    log.Fatal(err)
}
for ev := range eventChan {
    fmt.Printf("event: %s  canvas: %s\n", ev.Type, ev.Canvas.Name)
}
if err := <-errChan; err != nil {
    log.Println("stream error:", err)
}
```

## Go tools

| Tool | Path | Run |
|---|---|---|
| `canvus` CLI | `go/cli/` | `go run . <command>` |
| db-solver | `go/tools/db-solver/` | `go run .` |
| powertoys | `go/tools/powertoys/` | `go run cmd/powertoys/` |
| translator | `go/tools/translator/` | `go run .` |

## Next steps

- [Go workspace README](../../go/README.md)
- [Go SDK README](../../go/sdk/README.md)
- [Go conventions](../conventions/go.md)
- [API reference](../api-reference/README.md)
- [All core examples](../../go/examples/core/)
