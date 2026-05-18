"""Brainstorming-analysis MCP tools.

Four tools (PersonaIdentification, BrainstormingNoteRetrieval,
LLMBrainstormingAnalysis, AutoConnectorCreation) that combine Canvus data
fetching with LLM orchestration. Legacy implementation lives in
``canvus_mcp_server/mcp_tools/brainstorming_analysis.py`` (902 LOC) and
depends on the deferred Ollama client (see :mod:`canvus_mcp_server.mcp_tools.llm`).

This module preserves the public class names and the MCP registry surface;
the Canvus fetch is wired against the new SDK so the tools surface a
structured deferral error rather than ``ImportError``.
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
    "Phase 4d follow-up: LLM-driven brainstorming analysis (legacy "
    "brainstorming_analysis.py, 902 LOC) depends on the Ollama client port. "
    "Canvus-side data fetch is wired against the new SDK."
)


async def _fetch_brainstorm_payload(
    client: Client, tool: str, canvas_id: str
) -> dict[str, Any]:
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
    return {"canvas": dump_model(canvas), "notes": dump_model(notes)}


class BrainstormingNoteRetrievalTool(BaseMCPTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            name="brainstorming_note_retrieval",
            description="Retrieve and structure notes for brainstorming analysis.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        payload = await _fetch_brainstorm_payload(self.client, self.name, canvas_id)
        # This particular tool only retrieves & structures — no LLM call.
        # The legacy implementation grouped notes by colour and spatial
        # cluster; that algorithm is part of the Phase 4d port.
        raise MCPToolExecutionError(
            f"{_DEFERRED_MESSAGE} canvas_id={canvas_id}, "
            f"fetched_notes={len(payload['notes'])}.",
            self.name,
        )


class PersonaIdentificationTool(BaseMCPTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            name="persona_identification",
            description="Identify discussion personas from a brainstorming canvas.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        await _fetch_brainstorm_payload(self.client, self.name, canvas_id)
        raise MCPToolExecutionError(
            f"{_DEFERRED_MESSAGE} canvas_id={canvas_id}.",
            self.name,
        )


class LLMBrainstormingAnalysisTool(BaseMCPTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            name="llm_brainstorming_analysis",
            description="Run end-to-end LLM analysis on a brainstorming canvas.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        await _fetch_brainstorm_payload(self.client, self.name, canvas_id)
        raise MCPToolExecutionError(
            f"{_DEFERRED_MESSAGE} canvas_id={canvas_id}.",
            self.name,
        )


class AutoConnectorCreationTool(BaseMCPTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            name="auto_connector_creation",
            description="Automatically create connectors between related brainstorming notes.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        await _fetch_brainstorm_payload(self.client, self.name, canvas_id)
        raise MCPToolExecutionError(
            f"{_DEFERRED_MESSAGE} canvas_id={canvas_id}.",
            self.name,
        )


__all__ = [
    "AutoConnectorCreationTool",
    "BrainstormingNoteRetrievalTool",
    "LLMBrainstormingAnalysisTool",
    "PersonaIdentificationTool",
]
