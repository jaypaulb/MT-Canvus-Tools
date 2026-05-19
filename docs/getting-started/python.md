# Getting Started — Python

Get up and running with the MT-Canvus-Tools Python SDK.

## Prerequisites

- **Python 3.11 or later** — verify with `python --version`
- **uv** — install with `pip install uv` or see the [uv docs](https://docs.astral.sh/uv/)
- A running Canvus server and an API key

## Option A: Use the SDK as a dependency

> **Note:** Until the package is published on PyPI, use Option B or install directly from the repo:
> `uv add "canvus-sdk @ git+https://github.com/jaypaulb/MT-Canvus-Tools.git#subdirectory=python/sdk"`

Create `main.py`:

```python
import asyncio
import os
from canvus_sdk import Client

async def main():
    async with Client(
        base_url=os.environ["CANVUS_API_URL"],
        api_key=os.environ["CANVUS_API_KEY"],
    ) as client:
        canvases = await client.canvases.list()
        for canvas in canvases:
            print(canvas.id, canvas.name)

asyncio.run(main())
```

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key
uv run python main.py
```

## Option B: Run examples from the monorepo

```bash
git clone https://github.com/jaypaulb/MT-Canvus-Tools.git
cd MT-Canvus-Tools/python
uv sync
```

Set credentials, then run any numbered example:

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key

uv run python examples/core/01-auth-and-list/main.py
uv run python examples/core/05-streaming/main.py
```

## Environment variables

| Variable | Description |
|---|---|
| `CANVUS_API_URL` | Full base URL including `/api/v1` suffix |
| `CANVUS_API_KEY` | Long-lived API key — sent as `Private-Token` header |

Environment-backed config is available via `Client.from_env()` (reads `CANVUS_API_URL` and `CANVUS_API_KEY`):

```python
async with Client.from_env() as client:
    canvases = await client.canvases.list()
```

## Authentication modes

```python
# API key (recommended for services and automation)
async with Client(
    base_url=os.environ["CANVUS_API_URL"],
    api_key=os.environ["CANVUS_API_KEY"],
) as client:
    ...

# Username + password (interactive flows; short-lived token)
async with Client(base_url=os.environ["CANVUS_API_URL"]) as client:
    await client.auth.login(email="user@example.com", password="password")
    ...
```

## Real-time streaming

```python
async with Client.from_env() as client:
    async for canvas in client.canvases.subscribe():
        print(canvas.id, canvas.name)
```

## Python tools

| Tool | Path | Run |
|---|---|---|
| MCP server | `python/tools/mcp-server/` | `uv run python -m canvus_mcp_server` |

## Next steps

- [Python workspace README](../../python/README.md)
- [Python conventions](../conventions/python.md)
- [API reference](../api-reference/README.md)
- [All core examples](../../python/examples/core/)
