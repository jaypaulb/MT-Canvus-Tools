"""Async SQLite cache for LLM-derived artefacts.

Replaces the legacy ``caching.py`` (467 LOC, sync ``sqlite3`` + threading
lock). The async port uses :mod:`aiosqlite` so cache reads never block
the FastAPI event loop.

Cache keys are derived deterministically from ``(model, prompt)`` via
SHA-256; entries carry a configurable TTL.
"""

from __future__ import annotations

import hashlib
import time
from collections.abc import Awaitable, Callable
from dataclasses import dataclass
from pathlib import Path

import aiosqlite
import structlog

logger = structlog.get_logger(__name__)


class CacheError(Exception):
    """Cache backend failure (disk error, schema mismatch, etc.)."""


@dataclass(frozen=True)
class LLMCacheConfig:
    """Cache runtime configuration.

    Attributes:
        database_path: SQLite file path. Parent directory must exist or
            be creatable by the caller (see :meth:`Settings.ensure_directories`).
        ttl_seconds: Entry lifetime in seconds. ``0`` disables expiry.
        timeout_seconds: ``aiosqlite.connect`` timeout.
    """

    database_path: Path
    ttl_seconds: int = 86_400
    timeout_seconds: float = 30.0


class LLMCache:
    """SHA-256-keyed, TTL'd cache for textual LLM responses.

    The cache is *intentionally* not used as an error-recovery fallback —
    a cache miss is a normal flow that triggers re-computation, but a
    backend failure raises :class:`CacheError` per the project's
    no-silent-fallback convention.
    """

    def __init__(self, config: LLMCacheConfig) -> None:
        self._config = config
        self._initialised = False
        self._log = logger.bind(
            component="llm_cache", database=str(config.database_path)
        )

    @staticmethod
    def make_key(model: str, prompt: str) -> str:
        """Return the deterministic cache key for ``(model, prompt)``."""
        digest = hashlib.sha256()
        digest.update(model.encode("utf-8"))
        digest.update(b"\x00")
        digest.update(prompt.encode("utf-8"))
        return digest.hexdigest()

    async def initialise(self) -> None:
        """Create the schema if missing. Idempotent."""
        if self._initialised:
            return
        try:
            self._config.database_path.parent.mkdir(parents=True, exist_ok=True)
            async with aiosqlite.connect(
                self._config.database_path,
                timeout=self._config.timeout_seconds,
            ) as conn:
                await conn.execute(
                    """
                    CREATE TABLE IF NOT EXISTS llm_responses (
                        key TEXT PRIMARY KEY,
                        model TEXT NOT NULL,
                        value TEXT NOT NULL,
                        created_at REAL NOT NULL
                    )
                    """
                )
                await conn.commit()
        except (aiosqlite.Error, OSError) as exc:
            raise CacheError(
                f"failed to initialise LLM cache at {self._config.database_path}: {exc}"
            ) from exc
        self._initialised = True

    async def get(self, key: str) -> str | None:
        """Return the cached value or ``None`` on miss / expiry.

        Raises:
            CacheError: backend failure (not a cache miss).
        """
        await self.initialise()
        try:
            async with aiosqlite.connect(
                self._config.database_path,
                timeout=self._config.timeout_seconds,
            ) as conn, conn.execute(
                "SELECT value, created_at FROM llm_responses WHERE key = ?",
                (key,),
            ) as cursor:
                row = await cursor.fetchone()
        except aiosqlite.Error as exc:
            raise CacheError(f"cache read failed: {exc}") from exc

        if row is None:
            return None
        value, created_at = row
        if self._config.ttl_seconds > 0:
            age = time.time() - float(created_at)
            if age > self._config.ttl_seconds:
                return None
        return str(value)

    async def set(self, key: str, model: str, value: str) -> None:
        """Insert or overwrite a cache entry.

        Raises:
            CacheError: backend failure.
        """
        await self.initialise()
        now = time.time()
        try:
            async with aiosqlite.connect(
                self._config.database_path,
                timeout=self._config.timeout_seconds,
            ) as conn:
                await conn.execute(
                    """
                    INSERT INTO llm_responses (key, model, value, created_at)
                    VALUES (?, ?, ?, ?)
                    ON CONFLICT(key) DO UPDATE SET
                        model = excluded.model,
                        value = excluded.value,
                        created_at = excluded.created_at
                    """,
                    (key, model, value, now),
                )
                await conn.commit()
        except aiosqlite.Error as exc:
            raise CacheError(f"cache write failed: {exc}") from exc

    async def get_or_compute(
        self,
        key: str,
        model: str,
        factory: Callable[[], Awaitable[str]],
    ) -> str:
        """Return cached value or compute, cache, and return it.

        The ``factory`` is awaited only on a cache miss. Backend failures
        from the cache are surfaced — they do not silently fall through
        to ``factory``.

        Raises:
            CacheError: cache backend failure.
            Any exception raised by ``factory``.
        """
        cached = await self.get(key)
        if cached is not None:
            self._log.debug("cache hit", key=key)
            return cached
        self._log.debug("cache miss", key=key)
        value = await factory()
        await self.set(key, model, value)
        return value

    async def purge_expired(self) -> int:
        """Delete all rows older than ``ttl_seconds``. Returns rows removed.

        Raises:
            CacheError: backend failure.
        """
        if self._config.ttl_seconds <= 0:
            return 0
        await self.initialise()
        cutoff = time.time() - self._config.ttl_seconds
        try:
            async with aiosqlite.connect(
                self._config.database_path,
                timeout=self._config.timeout_seconds,
            ) as conn:
                cursor = await conn.execute(
                    "DELETE FROM llm_responses WHERE created_at < ?",
                    (cutoff,),
                )
                await conn.commit()
                return cursor.rowcount or 0
        except aiosqlite.Error as exc:
            raise CacheError(f"cache purge failed: {exc}") from exc


__all__ = ["CacheError", "LLMCache", "LLMCacheConfig"]
