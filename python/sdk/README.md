# canvus-sdk

Async-first Python SDK for the Canvus REST API.

- **Async-first.** Every resource method is `async def`. A synchronous facade is available via `client.sync` for non-async callers; it runs the underlying coroutines on a dedicated worker-thread loop so it works correctly inside Jupyter and other already-async contexts.
- **Typed.** All public methods are fully type-annotated; the package ships `py.typed` so consumer type checkers pick up the SDK's types without configuration. Internal type-checking is `mypy --strict`.
- **HTTP via `httpx`.** One library for both sync and async surfaces. The legacy `aiohttp` dependency is gone.
- **Pydantic v2 models.** All wire types are exposed as Pydantic models; field aliases handle the hyphenated wire keys the Canvus API uses in places.
- **Structured logging via `structlog`.** No `print()` anywhere in library code. Call `configure_logging()` once during application startup if you want the SDK's logs emitted with sensible defaults.

## Install

```bash
uv add canvus-sdk
```

Or, when working in this monorepo, the workspace lockfile already includes the SDK — no extra install needed.

## Quick start

```python
import asyncio
from canvus_sdk import Client

async def main() -> None:
    async with Client(base_url="https://canvus.example.com", api_key="ck_...") as client:
        canvases = await client.canvases.list()
        print(f"{len(canvases)} canvases")

asyncio.run(main())
```

### Environment-backed configuration

The SDK reads `CANVUS_API_URL` and `CANVUS_API_KEY` (with the `CANVUS_` prefix) from the environment or a `.env` file via `pydantic-settings`:

```bash
export CANVUS_API_URL=https://canvus.example.com
export CANVUS_API_KEY=ck_...
```

```python
async with Client.from_env() as client:
    server_info = await client.server.get_info()
```

### Sync usage

```python
client = Client.from_env()
canvases = client.sync.canvases.list()
note = client.sync.widgets.notes.create(
    canvases[0].id,
    {"text": "hello", "location": {"x": 100, "y": 100}},
)
client.sync.close()
```

## Resource map

| Attribute | Purpose |
|---|---|
| `client.canvases` | Canvases, backgrounds, color-presets, permissions, preview |
| `client.folders` | Canvas folders + permissions |
| `client.widgets` | All widget types (notes, images, videos, pdfs, browsers, anchors, connectors, tables, video-inputs, ip-videos, rdp-connections), uploads-folder, clone helper |
| `client.users` | User CRUD + lifecycle (block / unblock / approve / change-email / force-reset) |
| `client.groups` | Group CRUD + member management |
| `client.auth` | Login (password, SAML), logout, password reset, registration, access tokens |
| `client.server` | Server info / config, license, audit log, clients & workspaces, client video outputs/inputs |
| `client.assets` | Asset & mipmap downloads by hash |

## Errors

All exceptions derive from `CanvusError`. HTTP-level failures derive from `APIError` (with `status_code`, `response_body`, and `request_id` attributes). `AuthError`, `NotFoundError`, `RateLimitError`, `ServerError`, and `UnsupportedOperationError` cover the common cases.

## Conventions

This SDK follows `docs/conventions/python.md` in the monorepo. The migration from the legacy `CanvusPythonAPI` and the API drift fixes applied to this package are documented in [`MIGRATION-NOTES.md`](MIGRATION-NOTES.md).
