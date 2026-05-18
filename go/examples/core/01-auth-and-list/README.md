# 01 — Authenticate and List Canvases

The canonical Canvus "hello world": authenticate with a long-lived API key,
list every canvas the principal can read, and print a tab-aligned summary
table to stdout.

If this example runs to completion, your env vars, network path, and API key
are all wired up correctly — every other example in this directory builds on
the same authentication pattern.

## Prerequisites

- Go 1.22 or newer (workspace pulls in 1.23 toolchain automatically).
- A reachable Canvus server with HTTPS enabled.
- A long-lived API key with at least read access. Mint one from the Canvus
  UI under **Settings → Access Tokens**.

## Configuration

| Variable | Required | Description |
| --- | --- | --- |
| `CANVUS_BASE_URL` | yes | Full base URL including `/api/v1/`. The SDK appends paths relative to this. |
| `CANVUS_API_KEY` | yes | Long-lived `Private-Token` value. |
| `LOG_FORMAT` | no | `text` (default) or `json`. |

Copy `.env.example` to `.env` and edit the values:

```bash
cp .env.example .env
$EDITOR .env
set -a; source .env; set +a
```

`direnv` users can drop the same content into `.envrc` for automatic loading.

## Running

From this directory:

```bash
go run .
```

From the workspace root (`go/`):

```bash
go run ./examples/core/01-auth-and-list
```

## Expected output

```
ID                                    NAME              MODE        ACCESS  STATE
3a8a7ce1-0e1f-...                     Daily standup     normal      rw      normal
9e0bf6c4-1c2c-...                     Architecture WIP  presentation rw     normal
...
```

The slog status line `level=INFO msg="listed canvases" count=N` is emitted to
stderr; the table itself is on stdout so it can be piped or redirected without
mixing with logs.

## How it works

1. `mustEnv` reads the two required env vars and returns an explicit error if
   either is empty. We deliberately do **not** print the API key — only its
   absence.
2. `canvus.NewSession` is constructed with `SessionConfig{BaseURL: ...}` and
   the `WithAPIKey` option. The option installs an `http.RoundTripper` that
   appends the `Private-Token` header to every request. It also disables TLS
   verification, which matches the Canvus dev/test convention of using
   self-signed certificates — see the SDK godoc on `WithAPIKey` for the
   rationale.
3. `s.ListCanvases(ctx, nil)` issues a single `GET /api/v1/canvases` request
   and decodes the response into `[]canvus.Canvas`. Passing `nil` for the
   filter returns every canvas the principal can read.
4. `text/tabwriter` renders the table. We `defer w.Flush()` so partial output
   is still emitted if a `Fprintf` errors mid-loop.

## Troubleshooting

| Symptom | Likely cause |
| --- | --- |
| `API error 401` | API key is wrong, revoked, or scoped without read access. |
| `API error 404` | Base URL is wrong — usually missing `/api/v1/`. |
| `tls: failed to verify certificate` | The server's cert is signed by a CA your machine doesn't trust **and** you've passed a custom `*http.Client`. The default `WithAPIKey` client skips verification; the option order matters. |
| Empty table, no error | The API key has zero readable canvases. Verify in the Canvus UI. |
| `connection refused` | Wrong host, wrong port, or VPN/network reachability. Try `curl $CANVUS_BASE_URL/server-info`. |
