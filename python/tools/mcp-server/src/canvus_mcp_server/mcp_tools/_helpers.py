"""Shared helpers for MCP tool implementations.

These helpers absorb the recurring boilerplate around the new
``canvus-sdk`` (resource-namespace style, async-first, pydantic models)
so individual tool modules stay readable.
"""

from __future__ import annotations

from collections.abc import Callable
from typing import Any, TypeVar

from pydantic import BaseModel

from canvus_sdk import APIError

from .base import MCPToolExecutionError, MCPToolValidationError

T = TypeVar("T")


def dump_model(value: Any) -> Any:
    """Best-effort pydantic model -> dict conversion.

    Lists/tuples are recursed; pydantic models become dicts via
    ``model_dump()``; everything else is returned as-is.
    """
    if isinstance(value, BaseModel):
        return value.model_dump(mode="json")
    if isinstance(value, list | tuple):
        return [dump_model(v) for v in value]
    if isinstance(value, dict):
        return {k: dump_model(v) for k, v in value.items()}
    return value


def require_str(kwargs: dict[str, Any], key: str, tool_name: str) -> str:
    """Pop a non-empty string ``key`` from ``kwargs`` or raise.

    Raises:
        MCPToolValidationError: missing, wrong type, or empty.
    """
    value = kwargs.get(key)
    if value is None:
        raise MCPToolValidationError(f"{key} parameter is required", tool_name)
    if not isinstance(value, str):
        raise MCPToolValidationError(f"{key} must be a string", tool_name)
    if not value.strip():
        raise MCPToolValidationError(f"{key} cannot be empty", tool_name)
    return value


def optional_str(kwargs: dict[str, Any], key: str, tool_name: str) -> str | None:
    """Pop an optional string ``key`` from ``kwargs`` or raise on wrong type."""
    value = kwargs.get(key)
    if value is None:
        return None
    if not isinstance(value, str):
        raise MCPToolValidationError(
            f"{key} must be a string if provided", tool_name
        )
    return value


def optional_number(
    kwargs: dict[str, Any], key: str, tool_name: str
) -> float | None:
    """Pop an optional numeric ``key`` (int or float) from ``kwargs``."""
    value = kwargs.get(key)
    if value is None:
        return None
    if not isinstance(value, int | float) or isinstance(value, bool):
        raise MCPToolValidationError(
            f"{key} must be a number if provided", tool_name
        )
    return float(value)


def optional_dict(
    kwargs: dict[str, Any], key: str, tool_name: str
) -> dict[str, Any] | None:
    """Pop an optional dict ``key`` from ``kwargs``."""
    value = kwargs.get(key)
    if value is None:
        return None
    if not isinstance(value, dict):
        raise MCPToolValidationError(
            f"{key} must be an object if provided", tool_name
        )
    return value


async def run_with_api_error_translation(
    tool_name: str,
    operation: str,
    coro_factory: Callable[[], Any],
) -> Any:
    """Invoke ``coro_factory()`` and translate :class:`APIError` to MCP error.

    The helper does NOT swallow unrelated exceptions — those propagate to
    the registry's outer handler which logs them. Per the python conventions
    in ``docs/conventions/python.md`` §Error handling, silent fallbacks are
    forbidden.
    """
    try:
        return await coro_factory()
    except APIError as exc:
        raise MCPToolExecutionError(
            f"{operation} failed: {exc}", tool_name
        ) from exc


__all__ = [
    "dump_model",
    "optional_dict",
    "optional_number",
    "optional_str",
    "require_str",
    "run_with_api_error_translation",
]
