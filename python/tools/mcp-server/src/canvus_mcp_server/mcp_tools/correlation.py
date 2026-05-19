"""Correlation analysis MCP tools.

Phase 4d port of the legacy ``correlation_tools.py`` (510 LOC) and
``correlation_analysis.py`` (606 LOC). The port favours readable, focused
algorithms over byte-for-byte fidelity; the LLM-driven semantic pass
replaces the legacy custom embedding heuristics, which depended on
sklearn — out of scope for the Phase 4d port.
"""

from __future__ import annotations

import json
import math
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


async def _fetch_canvas_payload(
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


def _spatial_distance(a: dict[str, Any], b: dict[str, Any]) -> float:
    """Euclidean distance between two widgets' positions; ``inf`` if missing."""
    ap = a.get("position") if isinstance(a, dict) else None
    bp = b.get("position") if isinstance(b, dict) else None
    if not isinstance(ap, dict) or not isinstance(bp, dict):
        return math.inf
    try:
        return math.hypot(
            float(ap.get("x", 0)) - float(bp.get("x", 0)),
            float(ap.get("y", 0)) - float(bp.get("y", 0)),
        )
    except (TypeError, ValueError):
        return math.inf


class ElementRelationshipAnalysisTool(BaseMCPTool):
    """Analyse semantic + spatial relationships between canvas elements."""

    def __init__(self, client: Client, ollama: OllamaClient) -> None:
        super().__init__(
            name="element_relationship_analysis",
            description="Analyse semantic and spatial relationships between canvas elements.",
        )
        self.client = client
        self._ollama = ollama

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        model = optional_str(kwargs, "model", self.name)
        payload = await _fetch_canvas_payload(self.client, self.name, canvas_id)
        notes = payload["notes"] if isinstance(payload["notes"], list) else []
        connectors = (
            payload["connectors"] if isinstance(payload["connectors"], list) else []
        )

        # Spatial clusters: simple proximity threshold.
        threshold = float(kwargs.get("spatial_threshold", 500.0) or 500.0)
        spatial_pairs: list[dict[str, Any]] = []
        for i, a in enumerate(notes):
            if not isinstance(a, dict):
                continue
            for b in notes[i + 1:]:
                if not isinstance(b, dict):
                    continue
                d = _spatial_distance(a, b)
                if d <= threshold:
                    spatial_pairs.append(
                        {
                            "source_id": a.get("id"),
                            "target_id": b.get("id"),
                            "distance": d,
                        }
                    )

        # LLM semantic pass.
        content_blocks = [
            {"id": n.get("id"), "text": n.get("text") or n.get("content") or ""}
            for n in notes
            if isinstance(n, dict)
        ]
        prompt = (
            "For these canvas notes, identify the strongest semantic "
            "relationships. Respond ONLY with a JSON array of "
            "{source_id, target_id, relationship_type, strength (0..1)}.\n\n"
            f"Notes:\n{json.dumps(content_blocks, default=str)[:8000]}"
        )
        try:
            raw = await self._ollama.generate(
                prompt,
                model=model,
                system="You analyse semantic relationships. Reply with JSON only.",
                temperature=0.3,
            )
            semantic = _safe_json(raw)
        except OllamaError as exc:
            raise MCPToolExecutionError(
                f"semantic analysis failed: {exc}", self.name
            ) from exc
        except json.JSONDecodeError as exc:
            raise MCPToolExecutionError(
                f"semantic response was not valid JSON: {exc.msg}",
                self.name,
            ) from exc

        if not isinstance(semantic, list):
            semantic = []

        return {
            "canvas_id": canvas_id,
            "note_count": len(notes),
            "connector_count": len(connectors),
            "spatial_relationships": spatial_pairs,
            "semantic_relationships": semantic,
            "model": model or self._ollama.config.model,
        }


class AutomaticConnectorSuggestionTool(BaseMCPTool):
    """Suggest connectors between related canvas elements."""

    def __init__(self, client: Client, ollama: OllamaClient) -> None:
        super().__init__(
            name="automatic_connector_suggestion",
            description="Suggest visual connectors between related canvas elements.",
        )
        self.client = client
        self._ollama = ollama

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        model = optional_str(kwargs, "model", self.name)
        payload = await _fetch_canvas_payload(self.client, self.name, canvas_id)
        notes = payload["notes"] if isinstance(payload["notes"], list) else []
        existing = payload["connectors"] if isinstance(payload["connectors"], list) else []

        existing_pairs: set[tuple[Any, Any]] = set()
        for c in existing:
            if not isinstance(c, dict):
                continue
            src = c.get("src", {})
            dst = c.get("dst", {})
            if isinstance(src, dict) and isinstance(dst, dict):
                existing_pairs.add((src.get("id"), dst.get("id")))

        if len(notes) < 2:
            return {
                "canvas_id": canvas_id,
                "suggestions": [],
                "model": model or self._ollama.config.model,
            }

        prompt = (
            "Suggest connector pairs between these notes. Respond ONLY with a "
            "JSON array of {source_id, target_id, connection_type, "
            "confidence (0..1), reasoning}. Include only confidence > 0.6.\n\n"
            f"Notes:\n{json.dumps(notes, default=str)[:8000]}"
        )
        try:
            raw = await self._ollama.generate(
                prompt,
                model=model,
                system="You suggest connectors. Reply with JSON only.",
                temperature=0.4,
            )
            suggestions = _safe_json(raw)
        except OllamaError as exc:
            raise MCPToolExecutionError(
                f"connector suggestion failed: {exc}", self.name
            ) from exc
        except json.JSONDecodeError as exc:
            raise MCPToolExecutionError(
                f"suggestion response was not valid JSON: {exc.msg}",
                self.name,
            ) from exc

        if not isinstance(suggestions, list):
            suggestions = []

        novel = [
            s for s in suggestions
            if isinstance(s, dict)
            and (s.get("source_id"), s.get("target_id")) not in existing_pairs
        ]
        return {
            "canvas_id": canvas_id,
            "existing_connector_count": len(existing),
            "suggestions": novel,
            "model": model or self._ollama.config.model,
        }


class ConnectorVisualizationTool(BaseMCPTool):
    """Produce a simple node/edge graph representation of the canvas."""

    def __init__(self, client: Client, ollama: OllamaClient) -> None:
        super().__init__(
            name="connector_visualization",
            description="Generate a visual representation of the connector graph for a canvas.",
        )
        self.client = client
        # Ollama unused — the visualisation is purely structural — but the
        # constructor signature matches the family for uniform wiring.
        self._ollama = ollama

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        payload = await _fetch_canvas_payload(self.client, self.name, canvas_id)
        notes = payload["notes"] if isinstance(payload["notes"], list) else []
        connectors = (
            payload["connectors"] if isinstance(payload["connectors"], list) else []
        )

        nodes: list[dict[str, Any]] = []
        for note in notes:
            if not isinstance(note, dict):
                continue
            nodes.append(
                {
                    "id": note.get("id"),
                    "label": (note.get("text") or note.get("content") or "")[:80],
                    "color": note.get("background_color") or note.get("color"),
                    "position": note.get("position"),
                }
            )

        edges: list[dict[str, Any]] = []
        for c in connectors:
            if not isinstance(c, dict):
                continue
            src = c.get("src") if isinstance(c.get("src"), dict) else {}
            dst = c.get("dst") if isinstance(c.get("dst"), dict) else {}
            edges.append(
                {
                    "id": c.get("id"),
                    "source_id": src.get("id"),
                    "target_id": dst.get("id"),
                    "line_color": c.get("line_color"),
                }
            )

        return {
            "canvas_id": canvas_id,
            "node_count": len(nodes),
            "edge_count": len(edges),
            "nodes": nodes,
            "edges": edges,
        }


__all__ = [
    "AutomaticConnectorSuggestionTool",
    "ConnectorVisualizationTool",
    "ElementRelationshipAnalysisTool",
]
