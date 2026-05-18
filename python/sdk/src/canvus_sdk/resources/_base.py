"""Shared base class and helpers for resource modules."""

from __future__ import annotations

import json
from collections.abc import AsyncIterator, Iterable, Mapping
from typing import Any, TypeVar

from pydantic import BaseModel

from .._http import Transport

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

        Args:
            model_cls: pydantic model class for each yielded item.
            path: API path (without leading ``/``).
            params: Optional query string. ``subscribe=true`` is added
                automatically.

        Yields:
            Validated ``model_cls`` instances.
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
                    yield model_cls.model_validate(item)
            else:
                yield model_cls.model_validate(payload)

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
