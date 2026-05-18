# Example 01 — auth-and-list

Smoke test for the SDK and your server credentials.

## Purpose

Authenticates with an API key, lists every canvas visible to the
caller, and prints a four-column summary table to stdout. If this
example runs without errors, the rest of the examples in this
directory should work too.

## Prerequisites

- Node 20+ and pnpm 9+
- A reachable Canvus server with an API key that has at least read
  access to one or more canvases

## Configuration

| Variable           | Required | Description                                       |
| ------------------ | -------- | ------------------------------------------------- |
| `CANVUS_BASE_URL`  | yes      | API base URL, e.g. `https://server/api/v1/`       |
| `CANVUS_API_KEY`   | yes      | Long-lived API key (Private-Token header)         |
| `LOG_LEVEL`        | no       | pino log level (`debug`, `info`, `warn`, `error`) |
| `LOG_FORMAT`       | no       | `pretty` for human-readable dev output            |

Copy `.env.example` to `.env` and fill in values, then run with
`node --env-file=.env` (or use a tool like `direnv`).

## Run

From the example directory:

```bash
pnpm install         # one-time, from the typescript/ workspace root
pnpm dev             # runs via tsx (no build step)
# or
pnpm build && node --env-file=.env dist/index.js
```

## Expected output

```
ID                                    NAME                                      OWNER                   MODIFIED
----------------------------------------------------------------------------------------------------------------
1a2b3c4d-...-1234                     My First Canvas                           alice@example.com       2026-05-17T12:34:56Z
...
```

## How it works

The example wires the SDK to your env vars in three steps:

1. **Env loading** — `zod` parses `process.env` against a strict schema.
   Missing or malformed vars exit with status 2 and a structured error.
2. **Session creation** — `createSession({ baseUrl, apiKey })` returns a
   `Session` with all resource namespaces (`session.canvases`,
   `session.widgets`, etc.) wired up.
3. **List + print** — `session.canvases.list()` issues a single GET to
   `/api/v1/canvases` and returns the wire shape verbatim (kebab-case
   keys). The rendering code truncates long names so the table fits in
   a typical 120-column terminal.

`console.table` would have been simpler but cannot truncate UUIDs and
produces unreadable output once any column exceeds the column width.
The manual padding here is deliberate.

## Troubleshooting

- **401 Unauthorized** — your `CANVUS_API_KEY` is missing or wrong.
- **404 Not Found** — your `CANVUS_BASE_URL` does not end in
  `/api/v1/`, or the server is on a different path.
- **Empty list** — the API key authenticates a user with no canvas
  read access. Confirm the user's permissions in the Canvus admin UI.
- **TLS errors** — for dev servers with self-signed certs, set
  `NODE_TLS_REJECT_UNAUTHORIZED=0` (insecure; never in production).
