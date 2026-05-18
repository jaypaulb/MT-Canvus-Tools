"""Sanity tests for the MCP tool base classes and the registry."""

from __future__ import annotations

from typing import Any

import pytest
from canvus_mcp_server.mcp_tools.base import (
    BaseMCPTool,
    MCPToolError,
    MCPToolValidationError,
)
from canvus_mcp_server.mcp_tools.registry import MCPToolRegistry


class _Echo(BaseMCPTool):
    def __init__(self) -> None:
        super().__init__(name="echo", description="echo back")

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        return {"echoed": kwargs}


class _RaisesValidation(BaseMCPTool):
    def __init__(self) -> None:
        super().__init__(name="bad_input", description="raises on validate")

    def validate_input(self, **kwargs: Any) -> bool:
        raise MCPToolValidationError("nope", self.name)

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        return {}


class _Boom(BaseMCPTool):
    def __init__(self) -> None:
        super().__init__(name="boom", description="raises in execute")

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        raise RuntimeError("kaboom")


def test_registry_register_and_get() -> None:
    registry = MCPToolRegistry()
    tool = _Echo()
    registry.register_tool(tool)
    assert "echo" in registry
    assert registry.get_tool("echo") is tool
    assert registry.get_tool_class("echo") is _Echo


def test_registry_register_duplicate_raises() -> None:
    registry = MCPToolRegistry()
    registry.register_tool(_Echo())
    with pytest.raises(ValueError, match="already registered"):
        registry.register_tool(_Echo())


def test_registry_register_non_tool_raises() -> None:
    registry = MCPToolRegistry()
    with pytest.raises(TypeError):
        registry.register_tool("not a tool")  # type: ignore[arg-type]


def test_registry_unregister_missing_raises() -> None:
    registry = MCPToolRegistry()
    with pytest.raises(KeyError):
        registry.unregister_tool("nope")


def test_registry_list_tool_info_returns_metadata() -> None:
    registry = MCPToolRegistry()
    registry.register_tool(_Echo())
    info = registry.list_tool_info()
    assert info == [
        {
            "name": "echo",
            "description": "echo back",
            "version": "1.0.0",
            "class": "_Echo",
        }
    ]


async def test_registry_execute_wraps_validation_errors() -> None:
    registry = MCPToolRegistry()
    registry.register_tool(_RaisesValidation())
    with pytest.raises(MCPToolValidationError):
        await registry.execute_tool("bad_input")


async def test_registry_execute_wraps_runtime_errors() -> None:
    registry = MCPToolRegistry()
    registry.register_tool(_Boom())
    with pytest.raises(MCPToolError) as ctx:
        await registry.execute_tool("boom")
    # The original cause is chained so debuggers see the real failure.
    assert isinstance(ctx.value.__cause__, RuntimeError)


async def test_registry_execute_passes_kwargs() -> None:
    registry = MCPToolRegistry()
    registry.register_tool(_Echo())
    result = await registry.execute_tool("echo", a=1, b="two")
    # format_response wraps the raw dict.
    assert result["success"] is True
    assert result["data"] == {"echoed": {"a": 1, "b": "two"}}
    assert result["tool"] == "echo"
