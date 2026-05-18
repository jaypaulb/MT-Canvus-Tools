# canvus-mcp-server

A FastAPI + [fastapi-mcp](https://github.com/jlowin/fastmcp) server that
exposes Canvus operations as MCP (Model Context Protocol) tools, backed by
the new `canvus-sdk` Python client.

This is the Phase 4c port of the standalone
[`canvus-mcp-server`](https://github.com/jaypaulb/canvus-mcp-server) into
the `MT-Canvus-Tools` monorepo.

## Quick start

```bash
# From the python/ workspace root:
uv sync --extra dev

# Run the server:
CANVUS_API_URL=https://your-canvus-server/api/v1/ \
CANVUS_API_KEY=ck_... \
uv run python -m canvus_mcp_server
```

By default the server listens on `0.0.0.0:8000`. Override with
`CANVUS_MCP_SERVER_HOST` / `CANVUS_MCP_SERVER_PORT`.

## Environment variables

| Variable | Default | Purpose |
|---|---|---|
| `CANVUS_API_URL` | required | Base URL of the Canvus server. |
| `CANVUS_API_KEY` | required | Private-Token sent on every Canvus request. |
| `CANVUS_VERIFY_SSL` | `true` | Verify the Canvus server's TLS certificate. |
| `CANVUS_MCP_SERVER_HOST` | `0.0.0.0` | HTTP bind host. |
| `CANVUS_MCP_SERVER_PORT` | `8000` | HTTP bind port. |
| `CANVUS_MCP_SERVER_DEBUG` | `false` | Enable `/docs` + `/redoc`. |
| `CACHE_ENABLED` | `true` | Enable the SQLite PDF cache. |
| `DATABASE_PATH` | `~/.canvus-mcp-server/pdf_cache.db` | SQLite cache file. |
| `PDF_CACHE_TTL` | `86400` | Cache TTL in seconds. |
| `LOG_LEVEL` | `INFO` | Structlog log level. |
| `LOG_FORMAT` | `console` (TTY) / `json` | Renderer. |
| `OLLAMA_ENABLED` | `true` | Enable LLM features via Ollama. |
| `OLLAMA_HOST` | `localhost` | Ollama daemon host. |
| `OLLAMA_PORT` | `11434` | Ollama daemon port. |
| `OLLAMA_BASE_URL` | derived | Full URL; overrides host/port if set. |
| `DEFAULT_MODEL` | `gemma3:2b` | Default Ollama model. |

## MCP tool inventory

The server registers **41 tools** across nine resource families. All tools
follow the `BaseMCPTool` contract (`name`, `description`, `validate_input`,
`execute`) and are accessible via the registry at `/tools` and via the
`/mcp` MCP transport.

| Family | Tools |
|---|---|
| Canvases | `canvas_get`, `canvas_list`, `canvas_create`, `canvas_update`, `canvas_copy`, `canvas_move` |
| Notes | `note_create`, `note_get`, `note_list`, `note_update`, `note_delete` |
| Images | `image_create`, `image_get`, `image_list`, `image_update`, `image_delete` |
| Videos | `video_create`, `video_get`, `video_list`, `video_update`, `video_delete` |
| PDFs | `pdf_create`, `pdf_get`, `pdf_list`, `pdf_update` |
| Connectors | `connector_create`, `connector_get`, `connector_list`, `connector_update`, `connector_delete` |
| Users | `user_login`, `user_logout`, `get_current_user`, `user_list`, `user_create`, `get_canvas_permissions` |
| Correlation (Phase 4d body) | `element_relationship_analysis`, `automatic_connector_suggestion`, `connector_visualization` |
| LLM (Phase 4d body) | `llm_health_check`, `llm_text_analysis`, `llm_connector_suggestions`, `llm_canvas_insights`, `llm_enhanced_correlation`, `llm_brainstorming_enhancement` |
| Brainstorming (Phase 4d body) | `brainstorming_note_retrieval`, `persona_identification`, `llm_brainstorming_analysis`, `auto_connector_creation` |
| Reports (Phase 4d body) | `brainstorming_summary_report`, `brainstorming_insight_report`, `brainstorming_export` |

## HTTP API

- `GET /health` — liveness, registered-tool count.
- `GET /health/llm` — Ollama probe (no auth).
- `GET /tools` — registry inspection.
- `GET /tools/{name}` — tool metadata.
- `POST /tools/{name}/execute` — synchronous tool execution (body is the
  tool's `kwargs` as JSON).
- `/mcp/*` — fastapi-mcp transport (mounted if `fastapi-mcp` is installed).

## Connecting from Claude Desktop / MCP clients

Add to your MCP client configuration (e.g. `~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "canvus": {
      "command": "uv",
      "args": [
        "run",
        "--directory", "/path/to/MT-Canvus-Tools/python",
        "python", "-m", "canvus_mcp_server"
      ],
      "env": {
        "CANVUS_API_URL": "https://your-canvus-server/api/v1/",
        "CANVUS_API_KEY": "ck_..."
      }
    }
  }
}
```

## SQLite cache

PDF-text extraction is cached to a SQLite database at `DATABASE_PATH`
(default `~/.canvus-mcp-server/pdf_cache.db`). The cache layer itself is
non-Canvus infrastructure preserved from the legacy server; its full port
is scheduled for Phase 4d alongside the PDF-processing pipeline.

## Ollama setup

LLM features require a running [Ollama](https://ollama.com) instance and a
chat model. Default expected model: `gemma3:2b`.

```bash
brew install ollama   # or your platform's installer
ollama serve &
ollama pull gemma3:2b
```

Then point the server at it via `OLLAMA_HOST` / `OLLAMA_PORT` or
`OLLAMA_BASE_URL`.

## What's wired vs. deferred from the legacy port

Per the Phase 4c per-item audit (item #8) the legacy server is ~12k LOC of
production code plus ~13k LOC of tests. This round delivers:

**Wired against the new SDK (production-ready):**
- The full Canvus-touching MCP tool surface — 32 tools across canvases,
  notes, images, videos, PDFs, connectors, users.
- Resource-namespace SDK call style (`client.canvases.list(...)` etc.)
  per the new SDK's contract.
- The 3,169-LOC `canvas.py` god-file decomposed into per-resource modules
  (`canvases.py`, `notes.py`, `images.py`, `videos.py`, `pdfs.py`,
  `connectors.py`, `users.py`) — the audit's #1 decomposition target.
- Shared asset-tool helper (`_asset_tools.py`) factoring out
  near-identical Image / Video / PDF CRUD bodies (the rule of three is
  satisfied: three real implementations existed).
- FastAPI app factory with HTTP `/tools/{name}/execute` shim that
  surfaces validation errors as 400 and execution errors as 500.
- Settings via `pydantic-settings` mirroring the SDK convention; no
  module-scoped mutable global state.
- Structlog logging per the monorepo convention.
- mypy `--strict` clean on all new code (per the Round 1+2 lessons).

**Deferred to Phase 4d (tool surface preserved, body raises a structured
`MCPToolExecutionError` referencing the deferral):**
- 9 LLM-orchestration tools (`llm.py`, `brainstorming.py`, `reports.py`,
  `correlation.py`) — these depend on the legacy 541-LOC `llm_client.py`
  Ollama wrapper, the 902-LOC `brainstorming_analysis.py`, the 821-LOC
  `report_generation.py`, the 606-LOC `correlation_analysis.py`, plus a
  large catalogue of `settings.*` LLM flags. The Canvus-side data fetch
  is wired against the new SDK so the deferral is purely the in-process
  algorithm body.
- Custom JWT auth layer (legacy `auth.py` 730 LOC + `auth_endpoints.py`
  523 LOC) — a per-MCP-client session/permission system distinct from the
  Canvus API auth. Phase 4d should decide whether to revive this or rely
  on transport-level auth (Claude Desktop's MCP transport is local + the
  user already owns the API key).
- SQLite cache layer body (legacy `caching.py` 467 LOC) — the settings
  surface is preserved; the cache implementation port is paired with the
  PDF-processing port.
- PDF processing pipeline (legacy `pdf_processing.py` 714 LOC) — depends
  on the SQLite cache port and Ollama port; only invoked by the
  brainstorming / report tools currently deferred.
- Health-check sub-system (legacy `health_checks.py` 451 LOC) — kept as a
  thin `/health/llm` probe for now.

These deferrals do **not** change the user-visible MCP tool registry —
all 41 legacy tools remain registered and discoverable, so MCP clients
that enumerate the catalogue still see the full surface.

## Tests

```bash
# From python/:
uv run --extra dev pytest tools/mcp-server/tests
```

Tests exercise the real :class:`canvus_sdk.Client` with the HTTP boundary
mocked via [`respx`](https://lundberg.github.io/respx/) — no live server
or duplicate-helper test shims (per the Round 1+2 conventions).

## Linting and type-checking

```bash
# From python/:
uv run --extra dev ruff check tools/mcp-server/src
uv run --extra dev mypy --strict tools/mcp-server/src
```
