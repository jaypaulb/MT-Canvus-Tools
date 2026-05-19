"""Shared base class and helpers for resource modules."""

from __future__ import annotations

import asyncio
import contextlib
import dataclasses
import json
from collections.abc import AsyncIterator, Iterable, Mapping
from typing import Any, TypeVar

from pydantic import BaseModel

from .._http import Transport


@dataclasses.dataclass
class _StreamError:
    exc: Exception

ModelT = TypeVar("ModelT", bound=BaseModel)


class Resource:
    """Common scaffolding for all resource classes.

    Resources hold a reference to the shared :class:`Transport` instance owned
    by the parent :class:`Client`. They never own connection state of their
    own.
    """

    def __init__(self, transport: Transport) -> None:
        self._transport = transport

    @staticmethod
    def _parse(model_cls: type[ModelT], data: Any) -> ModelT:
        """Validate ``data`` against ``model_cls``, raising the SDK error type."""
        return model_cls.model_validate(data)

    @staticmethod
    def _parse_list(model_cls: type[ModelT], data: Any) -> list[ModelT]:
        """Validate a list payload into a list of ``model_cls`` instances."""
        if data is None:
            return []
        if not isinstance(data, Iterable):
            raise TypeError(
                f"expected list response, got {type(data).__name__}",
            )
        return [model_cls.model_validate(item) for item in data]

    async def _typed_subscribe(
        self,
        model_cls: type[ModelT],
        path: str,
        *,
        params: Mapping[str, Any] | None = None,
    ) -> AsyncIterator[ModelT]:
        """Subscribe to a streaming endpoint and yield typed models.

        Phase 4b §4.2 #13-#14: thin async-iterator wrapper around
        :meth:`Transport.stream_lines` that JSON-decodes each NDJSON line and
        validates it against ``model_cls``. Lines that fail JSON decoding are
        silently skipped (matches the existing widget-subscribe behaviour);
        lines that decode to a list (server may batch updates) are yielded
        one item at a time.

        Phase 4d Round B: uses an ``asyncio.Queue`` of capacity
        ``transport.subscribe_buffer`` (default 4) to decouple the stream
        reader from the consumer. High-throughput callers (live dashboards,
        ai-personas) can raise ``subscribe_buffer`` on the :class:`Client` to
        absorb bursts without applying backpressure to the HTTP response body.

        Args:
            model_cls: pydantic model class for each yielded item.
            path: API path (without leading ``/``).
            params: Optional query string. ``subscribe=true`` is added
                automatically.

        Yields:
            Validated ``model_cls`` instances.
        """
        _sentinel = object()
        queue: asyncio.Queue[ModelT | _StreamError | object] = asyncio.Queue(
            maxsize=self._transport.subscribe_buffer
        )
        query: dict[str, Any] = dict(params) if params else {}
        query.setdefault("subscribe", "true")

        async def _reader() -> None:
            try:
                async for line in self._transport.stream_lines("GET", path, params=query):
                    try:
                        payload = json.loads(line)
                    except json.JSONDecodeError:
                        continue
                    if isinstance(payload, list):
                        for item in payload:
                            await queue.put(model_cls.model_validate(item))
                    else:
                        await queue.put(model_cls.model_validate(payload))
            except Exception as exc:
                await queue.put(_StreamError(exc))
            finally:
                await queue.put(_sentinel)

        task = asyncio.ensure_future(_reader())
        try:
            while True:
                item = await queue.get()
                if item is _sentinel:
                    break
                if isinstance(item, _StreamError):
                    raise item.exc
                yield item  # type: ignore[misc]
        finally:
            task.cancel()
            with contextlib.suppress(asyncio.CancelledError):
                await task

    async def _raw_subscribe(
        self,
        path: str,
        *,
        params: Mapping[str, Any] | None = None,
    ) -> AsyncIterator[dict[str, Any]]:
        """Subscribe and yield raw decoded JSON dicts (no model validation).

        Use for endpoints with no first-class model (e.g. ``server-config``
        emits arbitrary key/value blobs).
        """
        query: dict[str, Any] = dict(params) if params else {}
        query.setdefault("subscribe", "true")
        async for line in self._transport.stream_lines("GET", path, params=query):
            try:
                payload = json.loads(line)
            except json.JSONDecodeError:
                continue
            if isinstance(payload, list):
                for item in payload:
                    if isinstance(item, dict):
                        yield item
            elif isinstance(payload, dict):
                yield payload


__all__ = ["Resource"]
