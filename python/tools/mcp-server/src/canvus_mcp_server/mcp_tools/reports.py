"""Report-generation MCP tools.

Three tools that produce structured reports from a brainstorming canvas
(Summary, Insight, Export). Legacy implementation in
``canvus_mcp_server/mcp_tools/report_generation.py`` (821 LOC) depends on
the deferred brainstorming-analysis pipeline.

This module preserves the public class names and the MCP registry surface;
the Canvus fetch is wired against the new SDK.
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
    "Phase 4d follow-up: report generation depends on the deferred "
    "brainstorming-analysis and LLM ports (see "
    "canvus_mcp_server.mcp_tools.brainstorming / .llm). Canvus fetch is "
    "wired against the new SDK."
)


class BrainstormingSummaryReportTool(BaseMCPTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            name="brainstorming_summary_report",
            description="Generate a summary report of a brainstorming session.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        canvas = await run_with_api_error_translation(
            self.name,
            "Retrieve canvas",
            lambda: self.client.canvases.get(canvas_id),
        )
        raise MCPToolExecutionError(
            f"{_DEFERRED_MESSAGE} canvas_id={canvas_id}, "
            f"canvas_name={dump_model(canvas).get('name')!r}.",
            self.name,
        )


class BrainstormingInsightReportTool(BaseMCPTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            name="brainstorming_insight_report",
            description="Generate an insight report with LLM-derived patterns and themes.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        await run_with_api_error_translation(
            self.name,
            "Retrieve canvas",
            lambda: self.client.canvases.get(canvas_id),
        )
        raise MCPToolExecutionError(
            f"{_DEFERRED_MESSAGE} canvas_id={canvas_id}.",
            self.name,
        )


class BrainstormingExportTool(BaseMCPTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            name="brainstorming_export",
            description="Export a brainstorming session to JSON/Markdown/CSV.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        await run_with_api_error_translation(
            self.name,
            "Retrieve canvas",
            lambda: self.client.canvases.get(canvas_id),
        )
        raise MCPToolExecutionError(
            f"{_DEFERRED_MESSAGE} canvas_id={canvas_id}.",
            self.name,
        )


__all__ = [
    "BrainstormingExportTool",
    "BrainstormingInsightReportTool",
    "BrainstormingSummaryReportTool",
]
