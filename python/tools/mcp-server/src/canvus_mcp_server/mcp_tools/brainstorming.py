"""Brainstorming-analysis MCP tools.

Phase 4d port of the legacy ``brainstorming_analysis.py`` (902 LOC).
Tools combine SDK-side note retrieval with LLM analysis via an injected
:class:`OllamaClient`.
"""

from __future__ import annotations

import json
from typing import Any

import structlog
from canvus_sdk import Client

from ..llm import OllamaClient, OllamaError
from ._helpers import (
    dump_model,
    optional_str,
    require_str,
    run_with_api_error_translation,
)
from .base import BaseMCPTool, MCPToolExecutionError

logger = structlog.get_logger(__name__)


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


def _group_by_color(notes: list[Any]) -> dict[str, list[Any]]:
    grouped: dict[str, list[Any]] = {}
    for note in notes:
        if not isinstance(note, dict):
            continue
        color = str(note.get("background_color") or note.get("color") or "unknown")
        grouped.setdefault(color, []).append(note)
    return grouped


def _safe_json(raw: str) -> Any:
    text = raw.strip()
    if text.startswith("```"):
        lines = text.splitlines()
        if lines and lines[0].startswith("```"):
            lines = lines[1:]
        if lines and lines[-1].startswith("```"):
            lines = lines[:-1]
        text = "\n".join(lines).strip()
    return json.loads(text)


class BrainstormingNoteRetrievalTool(BaseMCPTool):
    """Retrieve and group notes from a brainstorming canvas."""

    def __init__(self, client: Client, ollama: OllamaClient) -> None:
        super().__init__(
            name="brainstorming_note_retrieval",
            description="Retrieve and structure notes for brainstorming analysis.",
        )
        self.client = client
        # ollama unused here but kept for constructor uniformity with other
        # brainstorming tools that all share app.py wiring.
        self._ollama = ollama

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        filter_by_color = optional_str(kwargs, "filter_by_color", self.name)
        group_by = (optional_str(kwargs, "group_by", self.name) or "color").lower()

        payload = await _fetch_brainstorm_payload(self.client, self.name, canvas_id)
        notes = payload["notes"] if isinstance(payload["notes"], list) else []

        if filter_by_color:
            notes = [
                n for n in notes
                if isinstance(n, dict)
                and str(n.get("background_color") or n.get("color") or "") == filter_by_color
            ]

        if group_by == "color":
            grouped = _group_by_color(notes)
        else:
            grouped = {"all": notes}

        return {
            "canvas_id": canvas_id,
            "total_notes": len(notes),
            "grouped_notes": grouped,
            "group_by": group_by,
            "filters_applied": {"color": filter_by_color},
        }


class PersonaIdentificationTool(BaseMCPTool):
    """Identify per-colour personas in a brainstorming canvas."""

    def __init__(self, client: Client, ollama: OllamaClient) -> None:
        super().__init__(
            name="persona_identification",
            description="Identify discussion personas from a brainstorming canvas.",
        )
        self.client = client
        self._ollama = ollama

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        model = optional_str(kwargs, "model", self.name)
        payload = await _fetch_brainstorm_payload(self.client, self.name, canvas_id)
        notes = payload["notes"] if isinstance(payload["notes"], list) else []
        grouped = _group_by_color(notes)

        prompt = (
            "For each colour group below, infer a persona — a short label and "
            "1-2 sentence description of the perspective they represent. "
            "Respond ONLY with a JSON object keyed by colour, with values "
            "{persona_name, description, key_themes (array)}.\n\n"
            + json.dumps(
                {
                    color: [str(n.get("text") or n.get("content") or "") for n in group]
                    for color, group in grouped.items()
                },
                default=str,
            )[:8000]
        )
        try:
            raw = await self._ollama.generate(
                prompt,
                model=model,
                system="You identify discussion personas. Reply with JSON only.",
                temperature=0.4,
            )
            personas = _safe_json(raw)
        except OllamaError as exc:
            raise MCPToolExecutionError(
                f"persona identification failed: {exc}", self.name
            ) from exc
        except json.JSONDecodeError as exc:
            raise MCPToolExecutionError(
                f"persona response was not valid JSON: {exc.msg}",
                self.name,
            ) from exc

        return {
            "canvas_id": canvas_id,
            "personas": personas,
            "color_distribution": {k: len(v) for k, v in grouped.items()},
            "model": model or self._ollama.config.model,
        }


class LLMBrainstormingAnalysisTool(BaseMCPTool):
    """End-to-end LLM analysis of a brainstorming canvas."""

    def __init__(self, client: Client, ollama: OllamaClient) -> None:
        super().__init__(
            name="llm_brainstorming_analysis",
            description="Run end-to-end LLM analysis on a brainstorming canvas.",
        )
        self.client = client
        self._ollama = ollama

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        model = optional_str(kwargs, "model", self.name)
        payload = await _fetch_brainstorm_payload(self.client, self.name, canvas_id)
        notes = payload["notes"] if isinstance(payload["notes"], list) else []
        canvas = payload["canvas"]

        content = "\n".join(
            f"- {n.get('text') or n.get('content') or ''}"
            for n in notes
            if isinstance(n, dict)
        )
        prompt = (
            "Analyse this brainstorming session and respond ONLY with JSON: "
            "{\"main_themes\": [...], \"key_insights\": [...], "
            "\"gaps\": [...], \"recommendations\": [...], "
            "\"overall_assessment\": \"...\"}.\n\n"
            f"Notes ({len(notes)} total):\n{content[:8000]}"
        )
        try:
            raw = await self._ollama.generate(
                prompt,
                model=model,
                system="You analyse brainstorming sessions. Reply with JSON only.",
                temperature=0.4,
            )
            analysis = _safe_json(raw)
        except OllamaError as exc:
            raise MCPToolExecutionError(
                f"brainstorming analysis failed: {exc}", self.name
            ) from exc
        except json.JSONDecodeError as exc:
            raise MCPToolExecutionError(
                f"analysis response was not valid JSON: {exc.msg}",
                self.name,
            ) from exc

        return {
            "canvas_id": canvas_id,
            "canvas_name": canvas.get("name") if isinstance(canvas, dict) else None,
            "note_count": len(notes),
            "analysis": analysis,
            "model": model or self._ollama.config.model,
        }


class AutoConnectorCreationTool(BaseMCPTool):
    """Suggest (and optionally create) connectors between related notes."""

    def __init__(self, client: Client, ollama: OllamaClient) -> None:
        super().__init__(
            name="auto_connector_creation",
            description="Automatically create connectors between related brainstorming notes.",
        )
        self.client = client
        self._ollama = ollama

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        model = optional_str(kwargs, "model", self.name)
        # dry_run defaults to True — we surface suggestions but do not
        # mutate the canvas unless the caller opts in. Connector creation
        # is deferred to a follow-up because the legacy implementation
        # depends on widget endpoints that vary between server versions.
        dry_run = bool(kwargs.get("dry_run", True))

        payload = await _fetch_brainstorm_payload(self.client, self.name, canvas_id)
        notes = payload["notes"] if isinstance(payload["notes"], list) else []
        if len(notes) < 2:
            return {
                "canvas_id": canvas_id,
                "suggestions": [],
                "created_connectors": [],
                "model": model or self._ollama.config.model,
            }

        prompt = (
            "Identify the strongest pairwise connections between these notes. "
            "Respond ONLY with a JSON array of "
            "{source_id, target_id, connection_type, strength (0..1), reasoning}. "
            "Include only strength > 0.6.\n\n"
            f"Notes:\n{json.dumps(notes, default=str)[:8000]}"
        )
        try:
            raw = await self._ollama.generate(
                prompt,
                model=model,
                system="You identify relationships between notes. Reply with JSON only.",
                temperature=0.4,
            )
            suggestions = _safe_json(raw)
        except OllamaError as exc:
            raise MCPToolExecutionError(
                f"connector suggestion failed: {exc}", self.name
            ) from exc
        except json.JSONDecodeError as exc:
            raise MCPToolExecutionError(
                f"suggestions response was not valid JSON: {exc.msg}",
                self.name,
            ) from exc

        if not isinstance(suggestions, list):
            raise MCPToolExecutionError(
                "suggestions response was not a JSON array", self.name
            )

        return {
            "canvas_id": canvas_id,
            "suggestions": suggestions,
            "created_connectors": [],
            "dry_run": dry_run,
            "model": model or self._ollama.config.model,
            "note": (
                "Behavioural diff from legacy: connector creation is "
                "suggestion-only (dry_run defaults to True). Phase 4e: "
                "wire SDK connector-create once mutation policy is set."
            ),
        }


__all__ = [
    "AutoConnectorCreationTool",
    "BrainstormingNoteRetrievalTool",
    "LLMBrainstormingAnalysisTool",
    "PersonaIdentificationTool",
]
