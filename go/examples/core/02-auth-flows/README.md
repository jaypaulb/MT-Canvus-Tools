# 02 — Authentication Flows

Three auth paths, side by side:

1. **API key** — static `Private-Token` header. Best for daemons, CI, and
   anything else that should not store user credentials.
2. **Email + password login** — exchanges credentials for a short-lived
   bearer token via `POST /api/v1/users/login`. Best for interactive tools.
3. **Access-token CRUD** — once logged in, mint a programmatic token for
   later use, list to confirm, then delete to leave the campsite clean.

Each path runs independently. If you only set the API-key env vars, paths 2
and 3 are skipped with a warning — you still get a useful smoke test.

## Prerequisites

- A Canvus server reachable on `CANVUS_BASE_URL`.
- One or both of: a long-lived API key, and an email/password pair valid on
  that server.

## Configuration

| Variable | Required for | Description |
| --- | --- | --- |
| `CANVUS_BASE_URL` | all paths | Full base URL including `/api/v1/`. |
| `CANVUS_API_KEY` | path 1 | Long-lived API key. |
| `CANVUS_EMAIL` | paths 2 + 3 | Account email address. |
| `CANVUS_PASSWORD` | paths 2 + 3 | Account password. |
| `LOG_FORMAT` | — | `text` (default) or `json`. |

## Running

```bash
go run .
```

## Expected output (truncated)

```
level=INFO msg="=== auth path 1: API key ==="
level=INFO msg="api-key auth succeeded" canvas_count=12
level=INFO msg="=== auth path 2: email + password login ==="
level=INFO msg="login succeeded" user_id=1000
level=INFO msg="fetched self" user_id=1000 email=user@example.com name="Example User"
level=INFO msg="=== auth path 3: access-token CRUD ==="
level=INFO msg="access token created" token_id=tok_xxx name="example-02-auth-flows ..."
level=INFO msg="listed access tokens" count=3 found_new_token=true
level=INFO msg="access token deleted" token_id=tok_xxx
```

## How it works

### Path 1 — API key
`canvus.NewSession(cfg, canvus.WithAPIKey(key))` installs an HTTP round
tripper that appends `Private-Token: <key>` to every outbound request. No
explicit `Login` call is needed.

### Path 2 — Email + password
A bare `canvus.NewSession(cfg)` (no options) is anonymous until `s.Login(ctx,
email, password)` is called. The SDK posts `{"email": ..., "password": ...}`
to `/users/login`.

Per
[`VERIFIED-CORRECTIONS.md` §4](../../../docs/api-reference/VERIFIED-CORRECTIONS.md),
the live server rejects login bodies that include any **other** field — in
particular, defensive `{email, username, password}` payloads return
`{"msg":"Login request must have either email and password or token"}`. The
SDK sends the documented fields only; do not "improve" this by adding
`username`.

After login the SDK swaps in a `TokenAuthenticator` and remembers the
authenticated user ID, accessible via `s.UserID()`.

### Path 3 — Access-token CRUD
Built on top of the path-2 session. We `CreateAccessToken` with a uniquely
named request (timestamp suffix avoids name collisions if you run the
example repeatedly), `ListAccessTokens` to verify the new ID appears, then
`DeleteAccessToken` to clean up.

The freshly created token includes a `plain_token` field — that is the only
moment you ever see the raw secret. Store it now or never. (This example
discards it; it's a demo.)

## Troubleshooting

| Symptom | Likely cause |
| --- | --- |
| Login returns `Invalid username or password` | Wrong credentials, or the account is blocked / pending approval. |
| Login returns `Login request must have either email and password or token` | A custom build of the SDK is sending extra fields. The standard SDK does not — file an issue. |
| `CreateAccessToken: API error 403` | The logged-in user lacks permission to mint tokens. Some servers restrict this to admins. |
| `runTokenLifecycle` never runs | Path 2 was skipped or failed; path 3 depends on it. Check the path-2 warning. |
