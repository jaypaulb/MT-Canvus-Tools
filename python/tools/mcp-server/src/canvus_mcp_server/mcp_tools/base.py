"""Base MCP tool classes and exception hierarchy.

This module defines the abstract :class:`BaseMCPTool` that every MCP tool
inherits from, plus the tool-error hierarchy. The pattern is preserved
verbatim from the legacy ``canvus-mcp-server`` package — it is the
user-facing tool contract.
"""

from __future__ import annotations

from abc import ABC, abstractmethod
from typing import Any

import structlog


class MCPToolError(Exception):
    """Base exception for MCP tool errors."""

    def __init__(self, message: str, tool_name: str | None = None) -> None:
        self.message = message
        self.tool_name = tool_name
        super().__init__(self.message)


class MCPToolValidationError(MCPToolError):
    """Raised when tool input validation fails."""


class MCPToolExecutionError(MCPToolError):
    """Raised when tool execution fails."""


class BaseMCPTool(ABC):
    """Abstract base class for all MCP tools.

    Attributes:
        name: Unique tool name (snake_case).
        description: Human-readable description.
        version: Tool version string.
    """

    def __init__(
        self, name: str, description: str, version: str = "1.0.0"
    ) -> None:
        self.name = name
        self.description = description
        self.version = version
        self.logger = structlog.get_logger(f"canvus_mcp_server.tool.{name}")

    @abstractmethod
    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        """Execute the tool.

        Args:
            **kwargs: Tool-specific parameters.

        Returns:
            Result dictionary.

        Raises:
            MCPToolValidationError: invalid input.
            MCPToolExecutionError: execution failed.
        """

    def validate_input(self, **kwargs: Any) -> bool:
        """Validate input. Override for tool-specific rules.

        Args:
            **kwargs: Parameters to validate.

        Returns:
            True if validation passes.

        Raises:
            MCPToolValidationError: on failure.
        """
        return True

    def format_response(self, data: Any) -> dict[str, Any]:
        """Wrap raw data in the canonical response envelope."""
        return {
            "success": True,
            "data": data,
            "tool": self.name,
            "version": self.version,
        }

    def get_tool_info(self) -> dict[str, Any]:
        """Return metadata about the tool."""
        return {
            "name": self.name,
            "description": self.description,
            "version": self.version,
            "class": self.__class__.__name__,
        }

    def __str__(self) -> str:
        return f"{self.__class__.__name__}(name={self.name!r}, version={self.version!r})"

    def __repr__(self) -> str:
        return (
            f"{self.__class__.__name__}(name={self.name!r}, "
            f"description={self.description!r}, version={self.version!r})"
        )


__all__ = [
    "BaseMCPTool",
    "MCPToolError",
    "MCPToolExecutionError",
    "MCPToolValidationError",
]
