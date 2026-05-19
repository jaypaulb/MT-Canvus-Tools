# MT-Canvus-Tools — Python

Python uv workspace containing the SDK, examples, and tools for the Canvus platform.

## Structure

| Path | Description |
|---|---|
| [`sdk/`](sdk/) | Python SDK — 147 endpoints, 49 subscribe methods, 10 extras modules |
| [`examples/core/`](examples/core/) | 8 numbered examples (01 auth, 02 auth-flows, 03 widget-CRUD, 04 file-upload, 05 streaming, 06 LLM, 07 webhooks, 08 cross-canvas-clone) |
| [`tools/mcp-server/`](tools/mcp-server/) | MCP server exposing Canvus to Claude Desktop — 16 LLM/brainstorming/correlation/reports tools |

## Workspace setup

```bash
uv sync                                       # Install all workspace dependencies
uv run pytest                                 # Run SDK tests
uv run ruff check .                           # Lint
uv run mypy --strict python/sdk/src           # Type-check SDK (0 errors baseline)
```

Tests live in `sdk/tests/`. Integration tests opt in via `-m integration` (require `CANVUS_API_URL` + `CANVUS_API_KEY`). Live dev-server tests opt in via `-m live`.

## Authentication

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key
```

## Quick example

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

Or using environment-backed config:

```python
async with Client.from_env() as client:
    canvases = await client.canvases.list()
```

## Conventions

[`docs/conventions/python.md`](../docs/conventions/python.md) (monorepo root).

## Getting started

[`docs/getting-started/python.md`](../docs/getting-started/python.md) — step-by-step setup, authentication, running examples.
