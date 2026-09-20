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

## Request safety and compatibility (issue #6)

- Use one session per actor. Authentication is selected once per logical request, including read retries. A supplied token (including TokenStore bootstrap) takes precedence over the API key; explicit successful `Login` replaces bootstrap authority. A 401 does not refresh into a different identity or fall back to service/guest. Explicit `Logout` clears authority and UserID; it does not restore the API key. Explicit login/logout should not be interleaved across different actors on a shared session.
- Configuration and supplied `http.Client` values are copied. Caller-supplied transports and cookie jars remain caller-owned references; do not use an auth-injecting transport or shared authenticated cookie jar to mix actors. Auth injection happens in SDK request construction, not on arbitrary calls through the exported `HTTPClient`. TLS verification defaults are unchanged.
- `DefaultSessionConfig()` gives three retries for bodyless GET/HEAD requests. `MaxRetries=0` now means **no retries**, including in a plain `SessionConfig` literal. Negative budgets fail before ordinary requests are sent. Retry waits are cancellable.
- Writes/uploads and supported batch mutations are never automatically replayed. `BatchConfig.RetryAttempts` and `RetryDelay` are deprecated/ignored; `BatchResult.Retries` is zero. Reconcile effects before explicitly trying a mutation again.
- Redirects are not followed by default, preventing credential forwarding and implicit write replay. A supplied `CheckRedirect` is an explicit caller-owned policy; review it for origin/credential/replay safety.
- A 2xx response does not have to echo requested fields exactly. Server-normalized values and omitted optional fields are accepted. If the response cannot be read/decoded (including an empty body when a typed result is required), `errors.As` exposes `*canvus.AcceptedResponseError` with HTTP status, response request ID and resource ID when available from complete JSON. `Unwrap` preserves the underlying error. Raw bodies are not retained in this error.
- **Accepted is not the same as fully applied:** native asynchronous normalization may happen later. Conversely, a transport error without an HTTP response does not prove a mutation was never applied. Do not automatically recreate after either situation.

Python/TypeScript parity and exclusions are recorded in `docs/api-reference/changelog.md`. Stream lifetime/presence and camera helpers are not fixed by this change.

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
