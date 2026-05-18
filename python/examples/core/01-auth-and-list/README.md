# 01 — auth-and-list

Smoke test for the Canvus Python SDK: authenticate with a long-lived API key
and print every canvas the caller can see, as a fixed-width table.

## Prerequisites

- Python 3.11+.
- A working Canvus server URL.
- A Canvus API key (`CANVUS_API_KEY`) belonging to a user with at least
  view access on one or more canvases.

## Configuration

| Env var          | Required | Purpose                                                   |
| ---------------- | -------- | --------------------------------------------------------- |
| `CANVUS_API_URL` | yes      | Base URL (with or without `/api/v1/`).                    |
| `CANVUS_API_KEY` | yes      | API token sent in the `Private-Token` header.             |
| `LOG_FORMAT`     | no       | `console` (default at a TTY) or `json` for log shipping.  |
| `LOG_LEVEL`      | no       | `INFO` (default), `DEBUG`, etc.                           |

## Run

From the repo root:

```bash
cd python
uv run --package canvus-example-01-auth-and-list python examples/core/01-auth-and-list/main.py
```

Or set the env vars then invoke directly:

```bash
export CANVUS_API_URL=https://dev-mtcs.multitaction.com/api/v1/
export CANVUS_API_KEY=ck_...
python examples/core/01-auth-and-list/main.py
```

## Expected output

```
ID                                    NAME                MODE    ACCESS  STATE
------------------------------------  ------------------  ------  ------  ------
ab12...                               Sprint planning     normal  edit    normal
cd34...                               Board demo          demo    view    normal
```

## How it works

1. `Settings()` reads `CANVUS_API_URL` and `CANVUS_API_KEY` (pydantic-settings,
   prefix `CANVUS_`). Missing values surface as a validation error and the
   process exits 2.
2. `Client.from_env(settings)` builds the async client. `async with` ensures
   the underlying HTTP transport is closed on every exit path.
3. `client.canvases.list()` issues `GET /canvases`. The response is parsed
   into `Canvas` models; we render the relevant fields in a fixed-width
   table.

## Troubleshooting

- **`AuthError` (401):** API key is wrong or revoked.
- **`NotFoundError` (404):** Base URL is wrong (often missing `/api/v1/` or
  pointing at the web UI rather than the API host).
- **Empty list:** The token is valid but the user has no canvases. Create
  one in the UI, or invite the user to an existing canvas.
