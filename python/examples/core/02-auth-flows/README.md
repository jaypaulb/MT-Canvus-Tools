# 02 — auth-flows

Demonstrates the three authentication paths the Canvus API supports:

1. **API key** — long-lived `Private-Token` header. Recommended for daemons.
2. **Email + password login** — `POST /users/login` returns a short-lived
   session token plus a user object.
3. **Programmatic access tokens** — admin-issued per-user tokens with
   optional scopes / expiry. Useful for delegated automation.

Each flow runs in sequence; a missing env var skips that flow with a log
warning, the other flows still execute.

## Prerequisites

| Env var           | Drives flow | Required?                  |
| ----------------- | ----------- | -------------------------- |
| `CANVUS_API_URL`  | all         | yes                        |
| `CANVUS_API_KEY`  | 1 + 3       | yes for flows 1 and 3      |
| `CANVUS_EMAIL`    | 2 + 3       | yes for flow 2 (and 3)     |
| `CANVUS_PASSWORD` | 2 + 3       | yes for flow 2 (and 3)     |

Flow 3 (token lifecycle) needs a user id — we obtain it from the login flow,
since `/users/me` was removed in v1.2 (see `VERIFIED-CORRECTIONS.md §5`).

## Run

```bash
cd python
python examples/core/02-auth-flows/main.py
```

## Expected output (abbreviated)

```
flow_api_key ok                       canvas_count=7
flow_login ok                         user_id=1000 has_token=True
flow_token_lifecycle created          token_id=42 plain_token_prefix=ck_AB…
flow_token_lifecycle listed           count=3
flow_token_lifecycle deleted          token_id=42
```

## Important verified-truth caveats

- The login body must contain **only** `email` (and `password`). Adding
  `username` causes the server to respond
  `"Login request must have either email and password or token"`. The SDK's
  `client.auth.login(email=...)` keyword-only signature reflects this.
- User ids are integers, not UUIDs (`MIGRATION-NOTES.md §5.5`). The SDK
  accepts them as strings for the API surface but the wire value is numeric.

## Troubleshooting

- **`AuthError` on flow 1:** API key revoked or wrong server.
- **`AuthError` on flow 2:** wrong email / password.
- **`AuthError` on flow 3 create:** API key does not have admin scope; only
  admin or the user-owner can mint tokens.
