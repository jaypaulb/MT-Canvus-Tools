"""Centralised MCP tool registry.

The registry is the single point of access for tool discovery, lookup, and
execution. The class is preserved verbatim from the legacy package — it is
the user-facing contract for any FastAPI / fastapi-mcp wiring layer.
"""

from __future__ import annotations

from collections.abc import Iterator
from typing import Any

import structlog

from .base import BaseMCPTool, MCPToolError

logger = structlog.get_logger(__name__)


class MCPToolRegistry:
    """Central registry mapping tool names to instances."""

    def __init__(self) -> None:
        self._tools: dict[str, BaseMCPTool] = {}
        self._tool_classes: dict[str, type[BaseMCPTool]] = {}
        self.logger = structlog.get_logger("canvus_mcp_server.registry")

    def register_tool(self, tool: BaseMCPTool) -> None:
        """Register a tool instance.

        Raises:
            TypeError: ``tool`` is not a :class:`BaseMCPTool`.
            ValueError: a tool with the same name is already registered.
        """
        if not isinstance(tool, BaseMCPTool):
            raise TypeError(
                f"Tool must be an instance of BaseMCPTool, got {type(tool).__name__}"
            )
        if tool.name in self._tools:
            raise ValueError(f"Tool {tool.name!r} is already registered")
        self._tools[tool.name] = tool
        self._tool_classes[tool.name] = tool.__class__
        self.logger.info(
            "registered tool", tool=tool.name, cls=tool.__class__.__name__
        )

    def register_tool_class(
        self, tool_class: type[BaseMCPTool], **kwargs: Any
    ) -> None:
        """Instantiate then register a tool class."""
        if not issubclass(tool_class, BaseMCPTool):
            raise TypeError(
                f"Tool class must subclass BaseMCPTool, got {tool_class.__name__}"
            )
        self.register_tool(tool_class(**kwargs))

    def unregister_tool(self, tool_name: str) -> None:
        """Remove a tool by name.

        Raises:
            KeyError: tool not registered.
        """
        if tool_name not in self._tools:
            raise KeyError(f"Tool {tool_name!r} not found in registry")
        del self._tools[tool_name]
        del self._tool_classes[tool_name]
        self.logger.info("unregistered tool", tool=tool_name)

    def get_tool(self, tool_name: str) -> BaseMCPTool | None:
        """Return the registered instance or ``None``."""
        return self._tools.get(tool_name)

    def get_tool_class(self, tool_name: str) -> type[BaseMCPTool] | None:
        """Return the registered class or ``None``."""
        return self._tool_classes.get(tool_name)

    def list_tools(self) -> list[str]:
        """All registered tool names."""
        return list(self._tools.keys())

    def list_tool_info(self) -> list[dict[str, Any]]:
        """Metadata for all tools."""
        return [tool.get_tool_info() for tool in self._tools.values()]

    async def execute_tool(self, tool_name: str, **kwargs: Any) -> dict[str, Any]:
        """Execute a registered tool.

        Raises:
            KeyError: tool not registered.
            MCPToolError: validation or execution failure.
        """
        tool = self.get_tool(tool_name)
        if tool is None:
            raise KeyError(f"Tool {tool_name!r} not found in registry")
        try:
            if not tool.validate_input(**kwargs):
                raise MCPToolError(
                    f"Input validation failed for tool {tool_name!r}", tool_name
                )
            result = await tool.execute(**kwargs)
            return tool.format_response(result)
        except MCPToolError:
            raise
        except Exception as exc:
            self.logger.error(
                "tool execution failed", tool=tool_name, error=str(exc)
            )
            raise MCPToolError(
                f"Tool execution failed: {exc}", tool_name
            ) from exc

    def clear(self) -> None:
        """Drop all tools."""
        self._tools.clear()
        self._tool_classes.clear()
        self.logger.info("cleared registry")

    def __len__(self) -> int:
        return len(self._tools)

    def __contains__(self, tool_name: object) -> bool:
        return isinstance(tool_name, str) and tool_name in self._tools

    def __iter__(self) -> Iterator[str]:
        return iter(self._tools.keys())


# Module-level singleton, mirroring the legacy package.
tool_registry: MCPToolRegistry = MCPToolRegistry()


__all__ = ["MCPToolRegistry", "tool_registry"]
