# Canvus Python Examples — Core Set

Eight canonical examples that exercise the major capabilities of the
`canvus-sdk` Python package. Each lives in its own uv workspace member and
is runnable as a standalone script.

## Index

| #  | Slug                          | Purpose                                                                | Key env vars                                                                                  | Complexity |
| -- | ----------------------------- | ---------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- | ---------- |
| 01 | `01-auth-and-list`            | Smoke test: authenticate, list canvases, print table.                  | `CANVUS_API_URL`, `CANVUS_API_KEY`                                                            | *          |
| 02 | `02-auth-flows`               | Three auth modes (API key / login / programmatic token).               | `CANVUS_API_URL`, `CANVUS_API_KEY`, `CANVUS_EMAIL`, `CANVUS_PASSWORD`                         | **         |
| 03 | `03-widget-crud`              | Full lifecycle on a note widget incl. delete verification.             | `CANVUS_API_URL`, `CANVUS_API_KEY`, `CANVUS_CANVAS_ID`                                        | **         |
| 04 | `04-file-upload`              | Upload PNG → image widget → reposition → cleanup. Self-contained PNG.  | `CANVUS_API_URL`, `CANVUS_API_KEY`, `CANVUS_CANVAS_ID`, `CANVUS_IMAGE_PATH?`                  | **         |
| 05 | `05-streaming`                | Subscribe to NDJSON updates for a fixed duration.                      | `CANVUS_API_URL`, `CANVUS_API_KEY`, `CANVUS_CANVAS_ID`, `STREAM_DURATION_SECONDS?`            | ***        |
| 06 | `06-llm-integration`          | Paired watcher + responder bridging notes with a local Ollama LLM.     | `CANVUS_*`, `OLLAMA_URL`, `OLLAMA_MODEL`                                                      | ****       |
| 07 | `07-webhooks-notifications`   | Bridge subscribe events to an outbound webhook with retry/backoff.     | `CANVUS_*`, `WEBHOOK_URL`                                                                     | ***        |
| 08 | `08-cross-canvas-clone`       | Clone a widget across canvases via the per-type create endpoints.      | `CANVUS_API_URL`, `CANVUS_API_KEY`, `CANVUS_CANVAS_ID`, `CANVUS_SOURCE_WIDGET_ID`, `CANVUS_DEST_CANVAS_ID` | ***        |

Complexity scale: `*` trivial / `**` straightforward / `***` async + control
flow / `****` two binaries + external service.

## Conventions shared by every example

- **Python 3.11+**, async-first (`async def main()` + `asyncio.run(main())`).
- **Config:** `pydantic-settings` with `env_prefix="CANVUS_"`. Examples 06 /
  07 read a small number of non-`CANVUS_` vars (Ollama, webhook URL,
  stream duration) directly from `os.environ`.
- **Logging:** `structlog` configured via the SDK's `configure_logging()`.
  `LOG_FORMAT=json` flips to a JSON renderer; default is colourised console.
- **HTTP outside the SDK:** `httpx.AsyncClient`. No `requests`, no
  `aiohttp`.
- **Exit codes:** `0` success, `1` runtime/API failure, `2` config invalid.

## Running an example

From the `python/` workspace root, sync once:

```bash
cd python
uv sync
```

Then run the script directly (each example is a self-contained `main.py`):

```bash
# Drop env vars in a .env next to main.py, OR export them in the shell.
python examples/core/01-auth-and-list/main.py
```

The `.env.example` next to each `main.py` lists every variable that example
reads. For local dev a tool like `direnv` keeps your shell tidy.

## Workspace membership

All eight examples are registered as uv workspace members in
`python/pyproject.toml`. `canvus-sdk` is declared as a workspace source so
local SDK edits propagate to every example without reinstall.
