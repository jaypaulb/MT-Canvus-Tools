# Contributing — Python

## Workspace setup

```bash
git clone https://github.com/jaypaulb/MT-Canvus-Tools.git
cd MT-Canvus-Tools/python
uv sync
```

The `pyproject.toml` at `python/` is the uv workspace root. All members share the same lockfile.

## Adding a new package

1. Create the package directory
2. Add it to `[tool.uv.workspace]` members in `python/pyproject.toml`
3. Run `uv sync`
4. Tests go in the package's own `tests/` directory

## Test markers

| Marker | When to use |
|---|---|
| `integration` | Requires a live Canvus server (`CANVUS_API_URL` + `CANVUS_API_KEY`) |
| `live` | Hits the dev server (`CANVUS_DEV_BASE_URL` + `CANVUS_DEV_API_KEY`); skipped when env vars unset |
| `slow` | Takes more than a few seconds |

## Toolchain checks (all must pass)

```bash
uv run ruff check .                    # Linting — must be clean
uv run ruff format --check .           # Formatting — must be clean
uv run pytest                          # All tests pass
uv run mypy --strict sdk/src           # Must report 0 errors
```

## Commit scope

```
feat(python/sdk): add FoldersResource.subscribe_permissions
fix(python/tools/mcp-server): raise on non-list LLM responses
```

## Convention reference

[`docs/conventions/python.md`](../conventions/python.md).
