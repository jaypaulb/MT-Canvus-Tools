"""Correlation analysis MCP tools.

These three tools (ElementRelationshipAnalysis, AutomaticConnectorSuggestion,
ConnectorVisualization) perform pure in-process analysis on canvas data
fetched via the SDK. The legacy implementations live in
``/home/jaypaulb/Projects/gh/canvus-mcp-server/canvus_mcp_server/mcp_tools/correlation_tools.py``
(510 LOC); the algorithm bodies are non-Canvus and are scheduled for
Phase 4d port (they depend on the legacy ``correlation_analysis.py``
module — 606 LOC of clustering, similarity, and visualisation logic that
also needs to migrate).

This module preserves the public class names and the MCP tool registry
surface; calling :meth:`execute` raises a clear, actionable error pointing
at the deferred work so MCP clients receive structured feedback rather
than ``ImportError``.
"""

from __future__ import annotations

from typing import Any

from canvus_sdk import Client

from ._helpers import (
    dump_model,
    require_str,
    run_with_api_error_translation,
)
from .base import BaseMCPTool, MCPToolExecutionError

_DEFERRED_MESSAGE = (
    "Phase 4d follow-up: the LLM-driven analysis body for this tool is "
    "scheduled for the next refresh round. The Canvus-side data fetch is "
    "wired against the new SDK; only the in-process analysis algorithm "
    "(originally in canvus_mcp_server/correlation_analysis.py) is pending."
)


async def _fetch_canvas_payload(client: Client, tool: str, canvas_id: str) -> dict[str, Any]:
    """Fetch the canvas + its notes + connectors for analysis tools."""
    canvas = await run_with_api_error_translation(
        tool,
        "Retrieve canvas",
        lambda: client.canvases.get(canvas_id),
    )
    notes = await run_with_api_error_translation(
        tool,
        "List notes",
        lambda: client.widgets.notes.list(canvas_id),
    )
    connectors = await run_with_api_error_translation(
        tool,
        "List connectors",
        lambda: client.widgets.connectors.list(canvas_id),
    )
    return {
        "canvas": dump_model(canvas),
        "notes": dump_model(notes),
        "connectors": dump_model(connectors),
    }


class ElementRelationshipAnalysisTool(BaseMCPTool):
    """Analyse the relationships between elements on a canvas."""

    def __init__(self, client: Client) -> None:
        super().__init__(
            name="element_relationship_analysis",
            description="Analyse semantic and spatial relationships between canvas elements.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        payload = await _fetch_canvas_payload(self.client, self.name, canvas_id)
        raise MCPToolExecutionError(
            f"{_DEFERRED_MESSAGE} canvas_id={canvas_id}, "
            f"fetched_notes={len(payload['notes'])}, "
            f"fetched_connectors={len(payload['connectors'])}.",
            self.name,
        )


class AutomaticConnectorSuggestionTool(BaseMCPTool):
    """Suggest connectors between related canvas elements."""

    def __init__(self, client: Client) -> None:
        super().__init__(
            name="automatic_connector_suggestion",
            description="Suggest visual connectors between related canvas elements.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        payload = await _fetch_canvas_payload(self.client, self.name, canvas_id)
        raise MCPToolExecutionError(
            f"{_DEFERRED_MESSAGE} canvas_id={canvas_id}, "
            f"fetched_notes={len(payload['notes'])}.",
            self.name,
        )


class ConnectorVisualizationTool(BaseMCPTool):
    """Visualise the connector graph for a canvas."""

    def __init__(self, client: Client) -> None:
        super().__init__(
            name="connector_visualization",
            description="Generate a visual representation of the connector graph for a canvas.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        payload = await _fetch_canvas_payload(self.client, self.name, canvas_id)
        raise MCPToolExecutionError(
            f"{_DEFERRED_MESSAGE} canvas_id={canvas_id}, "
            f"fetched_connectors={len(payload['connectors'])}.",
            self.name,
        )


__all__ = [
    "AutomaticConnectorSuggestionTool",
    "ConnectorVisualizationTool",
    "ElementRelationshipAnalysisTool",
]
