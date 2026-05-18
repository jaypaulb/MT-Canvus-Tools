# Example 02 — auth-flows

Three auth flows in a single program, run in sequence.

## Purpose

Most Canvus clients use API keys, but the API supports two other auth
paths that show up in production:

1. **API key** — `Private-Token` header. The SDK's default flow.
2. **Email + password login** — `POST /users/login` exchanges
   credentials for a session token, which is then sent as
   `Private-Token` on subsequent requests.
3. **Access-token CRUD** — mint, list, and revoke a programmatic API
   token tied to the logged-in user. The minted token's secret is
   returned **only at creation time** and never again.

The example runs all three independently. If one fails (for example,
no `CANVUS_EMAIL`/`CANVUS_PASSWORD` is set), the others still run.

## Prerequisites

- Node 20+ and pnpm 9+
- An API key with read access (mode 1)
- A user account with email + password (modes 2 and 3, optional)

## Configuration

| Variable           | Required        | Description                                       |
| ------------------ | --------------- | ------------------------------------------------- |
| `CANVUS_API_URL`  | yes             | API base URL, e.g. `https://server/api/v1/`       |
| `CANVUS_API_KEY`   | yes             | Long-lived API key (mode 1)                       |
| `CANVUS_EMAIL`     | mode 2 + 3      | Login email                                       |
| `CANVUS_PASSWORD`  | mode 2 + 3      | Login password                                    |
| `LOG_LEVEL`        | no              | pino log level                                    |
| `LOG_FORMAT`       | no              | `pretty` for human-readable output                |

## Run

```bash
pnpm dev
# or
pnpm build && node --env-file=.env dist/index.js
```

## Expected output

A line per mode, structured JSON in production (or coloured pretty
output with `LOG_FORMAT=pretty`):

```
{"level":30,"component":"example-auth-flows","mode":"api-key","msg":"trying API-key auth"}
{"level":30,"component":"example-auth-flows","mode":"api-key","canvasCount":12,"msg":"API-key auth succeeded"}
{"level":30,"component":"example-auth-flows","mode":"login","email":"alice@example.com","msg":"trying login auth"}
{"level":30,"component":"example-auth-flows","mode":"login","userId":1000,"email":"alice@example.com","isAdmin":false,"msg":"login auth succeeded"}
{"level":30,"component":"example-auth-flows","mode":"token","tokenId":"...","description":"auth-flows-demo-1715942400000","msg":"minted access token"}
{"level":30,"component":"example-auth-flows","mode":"token","visibleTokenCount":1,"msg":"listed access tokens"}
{"level":30,"component":"example-auth-flows","mode":"token","tokenId":"...","msg":"revoked access token"}
```

## How it works

- **`runApiKey`** builds a session with `apiKey` set and calls
  `session.canvases.list()`. The SDK adds `Private-Token: <key>` to
  every request.
- **`runLogin`** builds an anonymous session, calls `session.auth.login`,
  then builds a second session using the returned session token. This
  shows the explicit two-step pattern; in a longer-lived app you would
  hold the token in a credential store.
- **`runTokenLifecycle`** uses the logged-in session to create an
  access token, list current tokens, and delete the one just created.
  The `token` field in the `createAccessToken` response is the secret;
  store it securely if you intend to use it.

## Verified corrections

- The login body uses **`email` only**. Sending an additional `username`
  field causes the server to reject the request with
  `"Login request must have either email and password or token"`.
  See `docs/api-reference/VERIFIED-CORRECTIONS.md` §4.
- `User.id` is an **integer**, not a UUID, and the field is named
  `id` (not `user-id`). The SDK type reflects this. See
  `VERIFIED-CORRECTIONS.md` §6.
- User fields use underscored names: `name` (not `full-name`),
  `admin` (not `is-admin`), `blocked` (not `is-blocked`).
- Access tokens use `id` (an opaque string) and `description` (the
  user-facing label). Older SDK drafts called these `token-id` and
  `name`; they do not exist on the wire.

## Troubleshooting

- **401 on mode 1** — wrong or missing API key.
- **401 on mode 2** — wrong email or password.
- **Mode 3 skipped** — needs a logged-in session, which needs
  `CANVUS_EMAIL` + `CANVUS_PASSWORD`.
- **Access denied on token create** — the API key user must have
  permission to manage tokens for the target user (typically only the
  user themselves, or an admin).
