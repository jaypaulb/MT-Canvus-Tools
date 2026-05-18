# Python Conventions — MT-Canvus-Tools

> **Status:** Locked defaults for all Python code in this monorepo.
> **Scope:** `python/sdk`, `python/examples/core/*`, `python/tools/*`, and any future Python member packages of the `python/` uv workspace.
> **Audience:** Phase 4b refresh agents and human contributors. Follow these conventions without prompting Jaypaul. If you need to deviate, append a dated entry to [Convention amendments](#convention-amendments) instead of forking the rule.

This document fixes the tooling so every Python tool, example, and the SDK behave identically in build, lint, type-check, log, error, test, and CI. Predictability over preference.

---

## Table of contents

1. [Module structure](#module-structure)
2. [Python version](#python-version)
3. [Build / install](#build--install)
4. [Format / lint](#format--lint)
5. [Type checking](#type-checking)
6. [Logging](#logging)
7. [Error handling](#error-handling)
8. [Configuration](#configuration)
9. [Testing](#testing)
10. [HTTP client](#http-client)
11. [Async](#async)
12. [CI](#ci)
13. [Code style](#code-style)
14. [Convention amendments](#convention-amendments)

---

## Module structure

### Workspace

The entire Python ecosystem lives under `python/` as a single **`uv` workspace**. One lockfile, one virtual environment, one resolver run. Member packages are referenced by path from the root `pyproject.toml`.

```
python/
├── pyproject.toml            # workspace root; declares members + dev deps
├── uv.lock                   # single resolved lockfile for the whole workspace
├── README.md
├── sdk/                      # ← canvus-sdk (the SDK package)
│   ├── pyproject.toml
│   └── src/
│       └── canvus_sdk/
│           ├── __init__.py
│           └── ...
├── examples/
│   └── core/
│       ├── 01-hello-canvas/
│       │   ├── pyproject.toml
│       │   └── src/
│       │       └── hello_canvas/
│       │           └── ...
│       └── 02-watch-canvas/
│           └── ...
└── tools/
    ├── mcp-server/
    │   ├── pyproject.toml
    │   └── src/
    │       └── canvus_mcp_server/
    │           └── ...
    └── local-llm/
        └── ...
```

Workspace root `pyproject.toml`:

```toml
[tool.uv.workspace]
members = [
    "sdk",
    "examples/core/*",
    "tools/*",
]

[tool.uv.sources]
canvus-sdk = { workspace = true }
```

Every member that consumes the SDK declares `canvus-sdk` as a workspace source so local edits propagate without `pip install -e` rituals:

```toml
[project]
dependencies = ["canvus-sdk"]

[tool.uv.sources]
canvus-sdk = { workspace = true }
```

### Package layout

**Src-layout, always.** Source under `src/<package_name>/`, tests under `tests/` at the package root. This makes import-time bugs impossible to mask via the implicit current-directory import, and matches the modern Python packaging consensus.

```
sdk/
├── pyproject.toml
├── src/
│   └── canvus_sdk/
│       ├── __init__.py
│       ├── client.py
│       ├── errors.py
│       ├── models/
│       └── resources/
└── tests/
    ├── conftest.py
    ├── unit/
    └── integration/
```

No `setup.py`, no `setup.cfg`, no flat layout. PEP 621 metadata only, declared in each member's `pyproject.toml`.

### Naming

- Package directory: `snake_case` (e.g. `canvus_sdk`, `canvus_mcp_server`).
- Distribution name on PyPI / in `[project].name`: `kebab-case` (e.g. `canvus-sdk`, `canvus-mcp-server`).
- Import name and distribution name should be obviously related — one is the kebab-version of the other.

---

## Python version

- **Minimum supported:** **Python 3.11**.
- **CI matrix:** Python **3.11** and **3.12**.
- **`requires-python`** in every member `pyproject.toml`:

  ```toml
  [project]
  requires-python = ">=3.11"
  ```

Why 3.11 floor: `Self` typing, `tomllib` in stdlib, `ExceptionGroup`, faster startup, and `StrEnum` are all assumed throughout the codebase. Anything older is not supported and code must not contain conditional shims for it.

3.13 is not yet in the matrix; it can be added by amendment once SDK dependencies (notably `httpx`, `pydantic`, `structlog`) ship clean wheels across the matrix.

---

## Build / install

### Tooling

- **`uv`** is the only supported package manager and resolver. No `pip`, no `poetry`, no `pip-tools`, no `pipenv` in any project documentation, scripts, or CI step.
- **`uv.lock`** lives at the workspace root (`python/uv.lock`) and is committed.
- **`hatchling`** is the build backend for every member's wheel/sdist. It plays well with src-layout and PEP 621.

Each member's `pyproject.toml`:

```toml
[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[project]
name = "canvus-sdk"
version = "0.1.0"
description = "Async-first Python SDK for the Canvus REST API."
readme = "README.md"
requires-python = ">=3.11"
license = { text = "MIT" }
authors = [{ name = "MultiTaction", email = "support@multitaction.com" }]
dependencies = [
    "httpx>=0.27",
    "pydantic>=2.7",
    "pydantic-settings>=2.3",
    "structlog>=24.1",
]

[tool.hatch.build.targets.wheel]
packages = ["src/canvus_sdk"]
```

### Common commands

Run from `python/` (the workspace root):

```bash
uv sync                     # install all workspace members + dev deps into .venv
uv sync --all-extras        # include optional extras across all members
uv run pytest               # run pytest within the workspace env
uv run ruff check .         # lint
uv run mypy sdk/src         # type-check the SDK
uv build --package canvus-sdk   # build wheel + sdist for a single member
```

Members must NOT carry their own `.venv/`. The single workspace environment is the only environment.

### No `setup.py`

No `setup.py` files anywhere. No dynamic build logic. If a member needs generated code, generate it as a pre-commit step or a `make` target, not inside the build backend.

---

## Format / lint

### Tool

**`ruff`** is the only formatter and linter. It replaces black, isort, flake8, pyupgrade, and bugbear. One tool, one config, one cache.

### Config

Declared once at the workspace root and inherited by all members (Ruff walks upward to discover config). Members may extend but should not contradict.

```toml
# python/pyproject.toml
[tool.ruff]
line-length = 100
target-version = "py311"
src = ["sdk/src", "tools/*/src", "examples/core/*/src"]

[tool.ruff.lint]
select = ["E", "W", "F", "I", "B", "UP", "N", "C4", "SIM", "RUF"]
ignore = [
    "E501",   # line length handled by formatter
]

[tool.ruff.lint.per-file-ignores]
"tests/**" = ["B011", "N802"]   # assert False and test_ method names allowed in tests
"examples/**" = ["N803", "N806"] # examples may use API-mirror naming

[tool.ruff.format]
quote-style = "double"
indent-style = "space"
docstring-code-format = true
```

### Selected rule sets

| Code | Set                       | Why                                                                |
| ---- | ------------------------- | ------------------------------------------------------------------ |
| `E`  | pycodestyle errors        | Baseline PEP 8 errors                                              |
| `W`  | pycodestyle warnings      | Baseline PEP 8 warnings                                            |
| `F`  | pyflakes                  | Real bugs: unused imports, undefined names                         |
| `I`  | isort                     | Import ordering and grouping                                       |
| `B`  | flake8-bugbear            | Likely bugs (mutable default args, `except:` clauses, etc.)        |
| `UP` | pyupgrade                 | Keep syntax modern: `X \| Y`, `list[int]`, f-strings               |
| `N`  | pep8-naming               | Enforce naming conventions                                         |
| `C4` | flake8-comprehensions     | Idiomatic comprehension usage                                      |
| `SIM`| flake8-simplify           | Catch redundant `if`, double negatives, `dict.get` patterns        |
| `RUF`| ruff-specific             | Ruff's own quality lints                                           |

### Formatter

`ruff format` replaces `black`. The two are line-compatible (Ruff's formatter is a black-compatible reimplementation). Line length **100**, not 88 — the extra width pays off for typed signatures and async wrappers.

### Pre-commit / on-save

- Editors: format on save with `ruff format`, organize imports with `ruff check --fix --select I`.
- Local pre-commit hook (optional but recommended): `uv run ruff format --check . && uv run ruff check .`.
- CI fails on `ruff format --check` or `ruff check` violations.

---

## Type checking

### Tool

**`mypy`** with **`--strict`** for the SDK. Permissive (default `mypy`, no `--strict`) for everything else.

### SDK config

```toml
# sdk/pyproject.toml
[tool.mypy]
python_version = "3.11"
strict = true
warn_unused_ignores = true
warn_redundant_casts = true
disallow_any_generics = true
disallow_untyped_decorators = true
no_implicit_reexport = true
plugins = ["pydantic.mypy"]
```

The SDK is the contract surface. Every public symbol must have explicit, accurate types. `Any` is a smell; if unavoidable, justify it inline with `# type: ignore[<code>]  # reason: ...`.

### Examples / tools config

```toml
# examples/core/01-hello-canvas/pyproject.toml
[tool.mypy]
python_version = "3.11"
ignore_missing_imports = true
# strict not enabled
```

Examples optimise for readability. Tools choose: if a tool is critical infrastructure, opt into `strict = true`; if it's a one-off integration, default suffices.

### Running

```bash
uv run mypy sdk/src                 # SDK only
uv run mypy tools/mcp-server/src    # one tool
uv run mypy .                       # everything (slower, used in CI)
```

---

## Logging

### Tool

**`structlog`** with two renderers:

- **JSON renderer** in production (machine-parseable, ships to log aggregators clean).
- **`ConsoleRenderer`** in development (colorised, human-readable).

The choice is driven by the `LOG_FORMAT` env var (`json` | `console`). Default is `console` if the process is attached to a TTY, `json` otherwise.

### Configuration

A single `configure_logging()` function lives in each package's `logging_config.py` and is called once during startup. It reads `LOG_FORMAT` and `LOG_LEVEL` from the environment.

```python
# canvus_sdk/logging_config.py
from __future__ import annotations

import logging
import os
import sys

import structlog


def configure_logging() -> None:
    """Configure structlog once for the process. Idempotent."""
    log_level = os.environ.get("LOG_LEVEL", "INFO").upper()
    log_format = os.environ.get("LOG_FORMAT", "console" if sys.stderr.isatty() else "json")

    timestamper = structlog.processors.TimeStamper(fmt="iso", utc=True)
    shared_processors: list[structlog.types.Processor] = [
        structlog.contextvars.merge_contextvars,
        structlog.processors.add_log_level,
        structlog.processors.StackInfoRenderer(),
        timestamper,
    ]

    if log_format == "json":
        renderer: structlog.types.Processor = structlog.processors.JSONRenderer()
    else:
        renderer = structlog.dev.ConsoleRenderer(colors=True)

    structlog.configure(
        processors=[
            *shared_processors,
            structlog.processors.format_exc_info,
            renderer,
        ],
        wrapper_class=structlog.make_filtering_bound_logger(
            getattr(logging, log_level, logging.INFO),
        ),
        context_class=dict,
        logger_factory=structlog.PrintLoggerFactory(file=sys.stderr),
        cache_logger_on_first_use=True,
    )
```

### Acquiring loggers

```python
import structlog

logger = structlog.get_logger(__name__)


async def fetch_canvas(canvas_id: str) -> Canvas:
    logger.info("fetching canvas", canvas_id=canvas_id)
    try:
        return await client.get_canvas(canvas_id)
    except APIError as e:
        logger.error("canvas fetch failed", canvas_id=canvas_id, status=e.status_code)
        raise
```

- Logger acquired via `structlog.get_logger(__name__)`.
- Keys are snake_case nouns; values are plain data (no `repr` strings).
- The message is a short verb phrase, not a sentence; structured key/value pairs carry the data.
- Do not interpolate values into the message string. `logger.info("created widget", widget_id=wid)`, not `logger.info(f"created widget {wid}")`.

### Levels

| Level     | Use                                                                |
| --------- | ------------------------------------------------------------------ |
| `DEBUG`   | Internal state, request payloads, retry counters.                  |
| `INFO`    | Normal lifecycle: connect, dispatch, complete.                     |
| `WARNING` | Recoverable anomaly: retry triggered, fell back, deprecated input. |
| `ERROR`   | Operation failed but process continues.                            |
| `CRITICAL`| Process cannot continue.                                           |

### What never to log

- Raw API keys, passwords, OAuth tokens, session cookies.
- Full request bodies that may contain user content. Log a hash or length instead.
- PII fields. If you cannot tell whether something is PII, treat it as PII.

---

## Error handling

### Hierarchy

All exceptions raised by the SDK derive from `canvus_sdk.errors.CanvusError`. Tools and examples may either re-export these or define their own subclasses of the appropriate SDK base.

```python
# canvus_sdk/errors.py
from __future__ import annotations


class CanvusError(Exception):
    """Base class for every error raised by canvus_sdk."""


class APIError(CanvusError):
    """HTTP-level error returned by the Canvus API."""

    def __init__(
        self,
        message: str,
        *,
        status_code: int,
        response_body: str | None = None,
        request_id: str | None = None,
    ) -> None:
        super().__init__(message)
        self.status_code = status_code
        self.response_body = response_body
        self.request_id = request_id


class AuthError(APIError):
    """401 / 403 from the API."""


class NotFoundError(APIError):
    """404 from the API."""


class RateLimitError(APIError):
    """429 from the API; check `retry_after`."""

    def __init__(self, *args: object, retry_after: float | None = None, **kwargs: object) -> None:
        super().__init__(*args, **kwargs)
        self.retry_after = retry_after


class ServerError(APIError):
    """5xx from the API."""


class ValidationError(CanvusError):
    """Local payload validation failed before the request was sent."""


class TransportError(CanvusError):
    """Networking layer failure (DNS, TLS, connection reset)."""
```

### Rules

1. **Always chain causes.** When catching to translate, use `raise NewError(...) from e`. Never lose the original traceback.

   ```python
   try:
       resp = await self._http.get(url)
   except httpx.TransportError as e:
       raise TransportError(f"failed to reach {url}") from e
   ```

2. **No bare `except:`.** Ever. Catch the specific exception class.

3. **No swallowing.** If you catch and don't re-raise, log at `WARNING` or `ERROR` and include enough context to understand the impact.

4. **No silent fallbacks.** A fallback is a feature, not an error handler. If `cache.get()` returns `None`, that's a normal cache miss, not an exception path. If the cache backend itself fails, raise — let the caller decide whether to degrade.

5. **Exceptions are for exceptional flow.** Use return values for expected outcomes (e.g. `find_by_name` returning `None` when not found, rather than raising).

6. **Public surface is documented.** Every public function lists its raised exceptions in the docstring `Raises:` section (see [Code style](#code-style)).

---

## Configuration

### Tool

**`pydantic-settings`** with env prefix `CANVUS_`. Precedence (highest first):

1. Environment variables.
2. `.env` file in the working directory.
3. Defaults in the `Settings` class.

### Pattern

```python
# canvus_sdk/config.py
from __future__ import annotations

from pydantic import Field, HttpUrl
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Canvus SDK runtime configuration."""

    model_config = SettingsConfigDict(
        env_prefix="CANVUS_",
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    api_url: HttpUrl = Field(..., description="Base URL for the Canvus API (e.g. https://canvus.example.com/api/v1/)")
    api_key: str = Field(..., description="Long-lived API token sent in the Private-Token header")
    verify_ssl: bool = True
    request_timeout_seconds: float = 30.0
    log_level: str = "INFO"
    log_format: str = "console"
```

Environment:

```
CANVUS_API_URL=https://canvus.example.com/api/v1/
CANVUS_API_KEY=ck_...
CANVUS_REQUEST_TIMEOUT_SECONDS=60
```

### Rules

- Field names: `snake_case` in code; env var auto-derived to `CANVUS_SNAKE_CASE`.
- Required fields have no default. If a value is genuinely optional, type it as `T | None = None`.
- Validation happens once at construction; the `Settings` object is then treated as frozen.
- Never read `os.environ` directly from non-config modules. Route everything through `Settings`.
- Examples and tools each define their own `Settings` subclass if they need extra fields; they should not import the SDK's `Settings` and monkey-patch it.

---

## Testing

### Tools

- **`pytest`** — runner.
- **`pytest-asyncio`** — `asyncio_mode = "auto"`, so `async def test_*` works without per-test decorators.
- **`pytest-cov`** — coverage measurement.
- **`respx`** — for mocking `httpx` in SDK unit tests (one tool, consistently).

### Layout

`tests/` lives at the package root and mirrors the `src/` layout:

```
sdk/
├── src/
│   └── canvus_sdk/
│       ├── client.py
│       └── resources/
│           └── canvases.py
└── tests/
    ├── conftest.py
    ├── unit/
    │   ├── test_client.py
    │   └── resources/
    │       └── test_canvases.py
    └── integration/
        └── test_canvases_live.py
```

### Markers

```toml
# python/pyproject.toml (workspace root)
[tool.pytest.ini_options]
minversion = "7.0"
addopts = [
    "--strict-markers",
    "--strict-config",
    "-ra",
]
testpaths = ["sdk/tests", "tools/*/tests", "examples/core/*/tests"]
asyncio_mode = "auto"
markers = [
    "integration: requires a live Canvus server (CANVUS_API_URL + CANVUS_API_KEY)",
    "slow: takes more than a few seconds; deselect with -m 'not slow'",
]
```

Integration tests are skipped by default:

```python
import os
import pytest

pytestmark = pytest.mark.integration

@pytest.fixture(autouse=True)
def _skip_if_no_creds() -> None:
    if not os.environ.get("CANVUS_API_KEY"):
        pytest.skip("CANVUS_API_KEY not set; skipping integration test")
```

Run:

```bash
uv run pytest                            # unit only (integration auto-skipped)
uv run pytest -m "integration"           # only integration
uv run pytest -m "not slow"              # exclude slow
```

### Coverage

- **SDK target: 70%** line coverage. Hard gate in CI.
- **Examples / tools: no hard target.** Coverage is collected for visibility but does not fail CI.

```toml
# sdk/pyproject.toml
[tool.coverage.run]
branch = true
source = ["canvus_sdk"]
omit = ["*/tests/*"]

[tool.coverage.report]
fail_under = 70
show_missing = true
exclude_lines = [
    "pragma: no cover",
    "if TYPE_CHECKING:",
    "raise NotImplementedError",
    "@(abc\\.)?abstractmethod",
]
```

### Style

- One assertion per test where practical; readability over cleverness.
- Use `pytest.mark.parametrize` for combinatorial cases instead of loops.
- Fixtures live in the nearest `conftest.py`. Don't reach across packages for fixtures — duplicate trivially.
- Mock at the boundary (HTTP via `respx`), not at internal seams. Tests should exercise real SDK logic.

---

## HTTP client

### Tool

**`httpx`** — both sync (`httpx.Client`) and async (`httpx.AsyncClient`). One library for both surfaces means one timeout model, one auth model, one retry model.

### Defaults

```python
import httpx

DEFAULT_TIMEOUT = httpx.Timeout(connect=5.0, read=30.0, write=30.0, pool=5.0)
DEFAULT_LIMITS = httpx.Limits(max_connections=100, max_keepalive_connections=20)

client = httpx.AsyncClient(
    base_url=str(settings.api_url),
    headers={"Private-Token": settings.api_key},
    timeout=DEFAULT_TIMEOUT,
    limits=DEFAULT_LIMITS,
    verify=settings.verify_ssl,
)
```

### Rules

- Always use a long-lived client (one per `Session` / `Client` instance); never construct per-request.
- Always set explicit timeouts. No infinite reads.
- Retry policy is implemented above `httpx` (see SDK retry module), not via `transport=` hacks.
- For streaming endpoints, use `client.stream("GET", ...)` with `async for line in response.aiter_lines()`.

---

## Async

### Posture

**Async-first SDK.** All resource methods are `async def` by default. The async surface is the source of truth.

```python
class Client:
    async def get_canvas(self, canvas_id: str) -> Canvas: ...
    async def list_canvases(self) -> list[Canvas]: ...
    async def stream_canvas(self, canvas_id: str) -> AsyncIterator[CanvasEvent]: ...
```

### Sync wrappers

A `Client.sync` namespace exposes sync versions of every method, implemented by running the coroutine on a dedicated event loop. This avoids forcing every consumer to adopt `asyncio` for a one-shot script.

```python
class Client:
    @property
    def sync(self) -> SyncClient:
        return SyncClient(self)


class SyncClient:
    def __init__(self, async_client: Client) -> None:
        self._async = async_client
        self._loop = _ensure_thread_loop()  # dedicated worker-thread loop

    def get_canvas(self, canvas_id: str) -> Canvas:
        return asyncio.run_coroutine_threadsafe(
            self._async.get_canvas(canvas_id), self._loop
        ).result()
```

### Rules

- Hand-write the async signatures; do not generate sync from async via `unasync`. The hand-written sync surface stays tiny because it just delegates.
- Streaming endpoints expose only the async form; document this in the docstring.
- Sync wrappers must NOT call `asyncio.run()` — that breaks inside Jupyter and inside existing event loops. Use a dedicated worker-thread loop.
- Cancellation: `httpx` cancellation propagates correctly through `asyncio.CancelledError`; do not catch and swallow it.

---

## CI

### Workflow

A single GitHub Actions workflow at `.github/workflows/python.yml` covers every Python member.

```yaml
name: Python

on:
  push:
    branches: [main]
    paths:
      - "python/**"
      - ".github/workflows/python.yml"
  pull_request:
    paths:
      - "python/**"
      - ".github/workflows/python.yml"

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        python-version: ["3.11", "3.12"]
    defaults:
      run:
        working-directory: python
    steps:
      - uses: actions/checkout@v4

      - name: Install uv
        uses: astral-sh/setup-uv@v3
        with:
          enable-cache: true

      - name: Set up Python ${{ matrix.python-version }}
        run: uv python install ${{ matrix.python-version }}

      - name: Sync workspace
        run: uv sync --all-extras --dev

      - name: Lint (ruff check)
        run: uv run ruff check .

      - name: Format check (ruff format)
        run: uv run ruff format --check .

      - name: Type-check SDK (mypy --strict)
        run: uv run mypy sdk/src

      - name: Type-check tools
        run: uv run mypy tools/*/src

      - name: Tests with coverage
        run: uv run pytest --cov --cov-report=term --cov-report=xml

      - name: Upload coverage
        if: matrix.python-version == '3.12'
        uses: codecov/codecov-action@v4
        with:
          files: python/coverage.xml
          flags: python
```

### Rules

- Both 3.11 and 3.12 must be green for a PR to merge.
- Coverage upload runs only on the highest matrix version to avoid double-counting.
- Integration tests do NOT run in CI by default. A separate manually-triggered workflow runs them against a staging Canvus instance with secrets.
- Caching uses `astral-sh/setup-uv@v3`'s built-in cache — no manual cache configuration.

---

## Code style

### Type hints

- **All public function signatures** carry type hints — arguments and return type.
- **Internal helpers** are typed when they touch domain types; trivial one-liners may be untyped.
- Prefer modern syntax: `list[int]`, `dict[str, Any]`, `X | None`. Never `List`, `Dict`, `Optional` (Ruff `UP` enforces this).
- Use `from __future__ import annotations` at the top of every module to defer evaluation. This keeps forward references cheap and avoids circular-import headaches.
- Type aliases for repeated complex types:

  ```python
  from typing import TypeAlias

  CanvasID: TypeAlias = str
  WidgetPayload: TypeAlias = dict[str, Any]
  ```

### Docstrings

- **Google style** for all public APIs. The format is unambiguous and renders well in both Sphinx and most editors.

```python
async def update_note(
    self,
    canvas_id: str,
    note_id: str,
    *,
    text: str | None = None,
    color: str | None = None,
) -> Note:
    """Update a note widget on a canvas.

    Args:
        canvas_id: ID of the canvas containing the note.
        note_id: ID of the note widget.
        text: New text content. If `None`, text is unchanged.
        color: New fill color (hex string). If `None`, color is unchanged.

    Returns:
        The updated `Note` model.

    Raises:
        NotFoundError: The canvas or note does not exist.
        AuthError: The current credentials cannot modify this canvas.
        APIError: Any other non-success HTTP response.
    """
```

- Module docstrings: one sentence summarising purpose, followed by a paragraph if the module is non-trivial.
- Examples and tools may use shorter docstrings; the SDK is the strict surface.

### Exceptions

- No bare `except:`. Always `except SpecificError:` or `except (ErrA, ErrB):`.
- Catch the narrowest exception that makes sense. `except Exception:` is permissible only at the outermost boundary (e.g. CLI main, request handler) where the goal is to log-and-fail rather than handle.
- Re-raise with `from e` whenever wrapping.

### Constants

- Module-level constants in `UPPER_SNAKE_CASE`:

  ```python
  DEFAULT_TIMEOUT_SECONDS = 30.0
  MAX_RETRIES = 3
  WIDGET_TYPES: frozenset[str] = frozenset({"note", "image", "video", "pdf"})
  ```

- Configuration values that change per environment live in `Settings`, not as module constants.

### Imports

- Three ordered groups (Ruff `I` enforces): stdlib, third-party, first-party. Blank line between groups.
- Absolute imports for first-party (`from canvus_sdk.errors import APIError`); never relative-from-root (`from ..errors import APIError`) inside the SDK package — relative is permitted only within a tightly-scoped subpackage.
- `from __future__ import annotations` is always first.

### General

- No `print()` in library code. Use `structlog`.
- No `assert` for runtime checks in library code. `assert` is for tests and for type narrowing where mypy benefits.
- Prefer `pathlib.Path` over `os.path`.
- Prefer dataclasses or Pydantic models over `dict`-shaped payloads for anything that crosses module boundaries.
- Functions over 50 lines deserve scrutiny; over 100 lines almost certainly need splitting.

---

## Convention amendments

Append entries here when a Phase 4b refresh, an SDK design decision, or a downstream tool surfaces something the locked defaults do not cover, or when an existing default is consciously overridden. Format mirrors `docs/conventions/go.md`. Do not edit historical entries; supersede with a new one.

### 2026-05-18 — subscribe_buffer adds asyncio.Queue capacity control (Phase 4d Round B)

**Item:** `python/sdk/src/canvus_sdk/client.py`, `_http.py`, `config.py`, `resources/_base.py`
**Decision:** `subscribe_buffer: int = 4` is added to `Client.__init__`, `Settings`, and `Transport`. `_typed_subscribe` now wraps the raw async-generator in an `asyncio.Queue(maxsize=subscribe_buffer)` backed by a background reader task. This converts the subscribe primitive from a pure pull-model generator to a push-buffered queue without changing the public `AsyncIterator[T]` return type.
**Rationale:** The Python subscribe was previously a pure async generator (pull-based, no internal buffer). High-throughput consumers that do slow per-item processing held back the HTTP read loop. Adding the queue decouples stream reading from item consumption. Storing `subscribe_buffer` on `Transport` (rather than threading it through every `Resource` subclass) minimises the change surface.

---

### YYYY-MM-DD — \<one-line summary>

**Item:** \<tool or example being refreshed>
**Decision:** \<what was chosen>
**Rationale:** \<why>

<!-- example:
### 2026-06-01 — MCP server adopts httpx-sse

**Item:** `python/tools/mcp-server`
**Decision:** Add `httpx-sse>=0.4` as a runtime dependency and use it for Canvus streaming subscriptions instead of hand-rolled line parsing.
**Rationale:** The hand-rolled parser drifted from the SSE spec around comment lines; `httpx-sse` is small, well-typed, and integrates with the existing `AsyncClient` instance with no extra connection management.
-->
