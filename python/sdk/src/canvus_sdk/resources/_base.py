"""Shared base class and helpers for resource modules."""

from __future__ import annotations

from collections.abc import Iterable
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


__all__ = ["Resource"]
