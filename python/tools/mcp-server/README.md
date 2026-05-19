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
| `OLLAMA_TIMEOUT` | `30` | Per-request timeout (seconds). |
| `OLLAMA_MAX_RETRIES` | `3` | Retry budget on transport / 5xx errors. |
| `DEFAULT_MODEL` | `gemma3:2b` | Default Ollama model. |
| `MODEL_TEMPERATURE` | `0.7` | Default sampling temperature. |
| `MODEL_MAX_TOKENS` | `2048` | Default `num_predict` (max tokens). |
| `MODEL_TOP_P` | `0.9` | Default top-p sampling threshold. |
| `HEALTH_CHECK_TIMEOUT` | `10` | Ollama probe timeout (`/health/llm`). |

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
| Correlation | `element_relationship_analysis`, `automatic_connector_suggestion`, `connector_visualization` |
| LLM | `llm_health_check`, `llm_text_analysis`, `llm_connector_suggestions`, `llm_canvas_insights`, `llm_enhanced_correlation`, `llm_brainstorming_enhancement` |
| Brainstorming | `brainstorming_note_retrieval`, `persona_identification`, `llm_brainstorming_analysis`, `auto_connector_creation` |
| Reports | `brainstorming_summary_report`, `brainstorming_insight_report`, `brainstorming_export` |

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

## LLM configuration

The LLM-orchestration tools (correlation / LLM / brainstorming / reports
families) talk to an [Ollama](https://ollama.com) instance via the
in-process `OllamaClient` (`canvus_mcp_server.llm.OllamaClient`). The
client is constructed once at app start-up from `Settings.build_ollama_config()`
and injected into every LLM-aware tool — there are no module-level globals.

| Convention | Notes |
|---|---|
| HTTP transport | `httpx.AsyncClient` (one per `OllamaClient`). |
| Logging | `structlog` at `INFO` / `WARNING`. |
| Errors | `OllamaError` / `OllamaConnectionError` / `OllamaInferenceError`; tools translate to `MCPToolExecutionError`. |
| Silent fallbacks | None. Legacy `LLM_FALLBACK_ENABLED` / `LLM_OFFLINE_MODE` paths are intentionally not ported (see `docs/conventions/python.md` §Error handling). |
| Retries | Per `OLLAMA_MAX_RETRIES`, with linear back-off. |

## SQLite cache

PDF-text extraction and LLM responses are cached to a SQLite database
at `DATABASE_PATH` (default `~/.canvus-mcp-server/pdf_cache.db`) via the
async `aiosqlite` backend. Cache keys are `SHA256(model + prompt)`; TTL
is governed by `PDF_CACHE_TTL`. Backend failures raise `CacheError`
rather than silently re-computing.

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

**Ported in Phase 4d Round D (LLM toolchain):**
- 16 LLM-orchestration tools (`llm.py`, `brainstorming.py`, `reports.py`,
  `correlation.py`) — previously deferred stubs are now full working
  implementations backed by the new `OllamaClient`.
- `canvus_mcp_server.llm.OllamaClient` — `httpx`-based replacement for
  the 541-LOC `aiohttp` legacy client; no module-level globals.
- `canvus_mcp_server.llm.LLMCache` — `aiosqlite`-based replacement for
  the threading-locked sync legacy cache.
- `canvus_mcp_server.llm.PDFProcessor` — `httpx` + `pdfplumber` PDF
  fetch / extract / chunk / summarise pipeline.

**Behavioural diffs from legacy** (documented in tool docstrings):
- `LLM_FALLBACK_ENABLED` / `LLM_OFFLINE_MODE` silent-fallback paths are
  intentionally not ported per `docs/conventions/python.md` §Error
  handling. Failures raise structured `MCPToolExecutionError`.
- `AutoConnectorCreationTool` is suggestion-only (`dry_run=True`) — the
  legacy direct-mutation path is deferred to Phase 4e behind a
  `dry_run=False` opt-in.
- `BrainstormingExportTool` ships JSON / Markdown / CSV; legacy PDF and
  PNG renderings (matplotlib + reportlab) are dropped.
- `ElementRelationshipAnalysisTool` replaces legacy sklearn-based
  clustering / TF-IDF with proximity threshold + LLM semantic pass.

**Still deferred:**
- Custom JWT auth layer (legacy `auth.py` 730 LOC + `auth_endpoints.py`
  523 LOC) — a per-MCP-client session/permission system distinct from the
  Canvus API auth. Phase 4e should decide whether to revive this or rely
  on transport-level auth (Claude Desktop's MCP transport is local + the
  user already owns the API key).
- Full `health_checks.py` (451 LOC) — kept as a thin `/health/llm` probe.

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
