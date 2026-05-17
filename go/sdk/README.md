# MT-Canvus-Tools — Go SDK

The official Go SDK for the Canvus API, lifted from the legacy `Canvus-Go-API`
module and adapted to the MT-Canvus-Tools monorepo conventions.

## Install

```bash
go get github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus
```

The SDK requires **Go 1.22** or later. A workspace pin is provided in `go.work`
for monorepo development; downstream consumers should depend on a tagged
release.

## Quick example

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func main() {
	cfg := canvus.DefaultSessionConfig()
	cfg.BaseURL = "https://canvus.example.com/api/v1"
	session := canvus.NewSession(cfg, canvus.WithAPIKey("my-key"))

	canvases, err := session.ListCanvases(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range canvases {
		fmt.Println(c.ID, c.Name)
	}
}
```

## Documentation

- Conventions: `docs/conventions/go.md` (monorepo root)
- API reference: `docs/api-reference/` (monorepo root)
- Migration notes: [`MIGRATION-NOTES.md`](MIGRATION-NOTES.md)

## Authentication

Two auth modes are supported:

| Mode | Helper | Use when |
| --- | --- | --- |
| API key | `canvus.WithAPIKey(key)` | Service / automation accounts |
| Username + password | `Session.Login(ctx, user, pw)` | Interactive flows; short-lived token |

The SDK accepts both `username` and `email` in the login payload — see
`MIGRATION-NOTES.md` for the field-name reconciliation history.

## Integration tests

Live-server tests live under `integration/` and are gated by the `integration`
build tag:

```bash
CANVUS_API_KEY=... CANVUS_BASE_URL=https://server/api/v1/ \
  go test -tags=integration ./integration/...
```

Unit tests run as usual:

```bash
go test ./...
```
