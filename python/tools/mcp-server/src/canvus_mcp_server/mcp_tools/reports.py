"""Report-generation MCP tools.

Phase 4d port of the legacy ``report_generation.py`` (821 LOC). Reports
combine SDK-side canvas data with LLM-generated narrative via an
injected :class:`OllamaClient`.
"""

from __future__ import annotations

import csv
import io
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


async def _fetch_canvas_and_notes(
    client: Client, tool: str, canvas_id: str
) -> tuple[Any, list[Any]]:
    canvas = await run_with_api_error_translation(
        tool, "Retrieve canvas", lambda: client.canvases.get(canvas_id)
    )
    notes = await run_with_api_error_translation(
        tool, "List notes", lambda: client.widgets.notes.list(canvas_id)
    )
    notes_payload = dump_model(notes)
    return dump_model(canvas), notes_payload if isinstance(notes_payload, list) else []


class BrainstormingSummaryReportTool(BaseMCPTool):
    """Generate a structured summary of a brainstorming canvas."""

    def __init__(self, client: Client, ollama: OllamaClient) -> None:
        super().__init__(
            name="brainstorming_summary_report",
            description="Generate a summary report of a brainstorming session.",
        )
        self.client = client
        self._ollama = ollama

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        model = optional_str(kwargs, "model", self.name)
        canvas, notes = await _fetch_canvas_and_notes(self.client, self.name, canvas_id)

        body = "\n".join(
            f"- {n.get('text') or n.get('content') or ''}"
            for n in notes
            if isinstance(n, dict)
        )
        prompt = (
            "Produce a structured summary of this brainstorming session. "
            "Respond ONLY with JSON: {\"executive_summary\": \"...\", "
            "\"main_themes\": [...], \"top_ideas\": [...], "
            "\"next_steps\": [...]}.\n\n"
            f"Canvas: {canvas.get('name') if isinstance(canvas, dict) else ''}\n"
            f"Notes ({len(notes)} total):\n{body[:8000]}"
        )
        try:
            raw = await self._ollama.generate(
                prompt,
                model=model,
                system="You produce concise meeting summaries. Reply with JSON only.",
                temperature=0.3,
            )
            summary = _safe_json(raw)
        except OllamaError as exc:
            raise MCPToolExecutionError(
                f"summary generation failed: {exc}", self.name
            ) from exc
        except json.JSONDecodeError as exc:
            raise MCPToolExecutionError(
                f"summary response was not valid JSON: {exc.msg}",
                self.name,
            ) from exc

        return {
            "canvas_id": canvas_id,
            "canvas_name": canvas.get("name") if isinstance(canvas, dict) else None,
            "note_count": len(notes),
            "summary": summary,
            "model": model or self._ollama.config.model,
        }


class BrainstormingInsightReportTool(BaseMCPTool):
    """Generate a deeper insight report with patterns and gaps."""

    def __init__(self, client: Client, ollama: OllamaClient) -> None:
        super().__init__(
            name="brainstorming_insight_report",
            description="Generate an insight report with LLM-derived patterns and themes.",
        )
        self.client = client
        self._ollama = ollama

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        model = optional_str(kwargs, "model", self.name)
        canvas, notes = await _fetch_canvas_and_notes(self.client, self.name, canvas_id)

        body = "\n".join(
            f"- {n.get('text') or n.get('content') or ''}"
            for n in notes
            if isinstance(n, dict)
        )
        prompt = (
            "Produce an insight report: surface non-obvious patterns, hidden "
            "themes, conflicts, and blindspots. Respond ONLY with JSON: "
            "{\"patterns\": [...], \"hidden_themes\": [...], "
            "\"conflicts\": [...], \"blindspots\": [...], "
            "\"recommendations\": [...], \"confidence\": <0..1>}.\n\n"
            f"Notes ({len(notes)} total):\n{body[:8000]}"
        )
        try:
            raw = await self._ollama.generate(
                prompt,
                model=model,
                system="You are an insightful analyst. Reply with JSON only.",
                temperature=0.5,
            )
            insights = _safe_json(raw)
        except OllamaError as exc:
            raise MCPToolExecutionError(
                f"insight generation failed: {exc}", self.name
            ) from exc
        except json.JSONDecodeError as exc:
            raise MCPToolExecutionError(
                f"insight response was not valid JSON: {exc.msg}",
                self.name,
            ) from exc

        return {
            "canvas_id": canvas_id,
            "canvas_name": canvas.get("name") if isinstance(canvas, dict) else None,
            "note_count": len(notes),
            "insights": insights,
            "model": model or self._ollama.config.model,
        }


class BrainstormingExportTool(BaseMCPTool):
    """Export a brainstorming session to JSON / Markdown / CSV.

    Behavioural diff from legacy: the legacy implementation also produced
    PDF and PNG renderings via matplotlib + reportlab. Those binary
    formats are excluded here — the structured formats cover all known
    downstream consumers (per Phase 4d audit).
    """

    _ALLOWED_FORMATS = {"json", "markdown", "csv"}

    def __init__(self, client: Client, ollama: OllamaClient) -> None:
        super().__init__(
            name="brainstorming_export",
            description="Export a brainstorming session to JSON/Markdown/CSV.",
        )
        self.client = client
        # Ollama only used if include_summary=True
        self._ollama = ollama

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        fmt = (optional_str(kwargs, "format", self.name) or "json").lower()
        if fmt not in self._ALLOWED_FORMATS:
            raise MCPToolExecutionError(
                f"Unsupported format {fmt!r}; "
                f"expected one of {sorted(self._ALLOWED_FORMATS)}",
                self.name,
            )
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        fmt = (optional_str(kwargs, "format", self.name) or "json").lower()
        include_summary = bool(kwargs.get("include_summary", False))
        model = optional_str(kwargs, "model", self.name)
        canvas, notes = await _fetch_canvas_and_notes(self.client, self.name, canvas_id)

        export_data: dict[str, Any] = {
            "canvas_id": canvas_id,
            "canvas_name": canvas.get("name") if isinstance(canvas, dict) else None,
            "notes": notes,
        }

        if include_summary and notes:
            body = "\n".join(
                f"- {n.get('text') or n.get('content') or ''}"
                for n in notes
                if isinstance(n, dict)
            )
            prompt = (
                "Summarise this brainstorming session in 3-5 sentences. "
                "Plain prose, no JSON.\n\n" + body[:8000]
            )
            try:
                export_data["summary"] = (
                    await self._ollama.generate(
                        prompt,
                        model=model,
                        system="You are a careful summariser.",
                        temperature=0.3,
                    )
                ).strip()
            except OllamaError as exc:
                raise MCPToolExecutionError(
                    f"summary for export failed: {exc}", self.name
                ) from exc

        if fmt == "json":
            content = json.dumps(export_data, indent=2, default=str)
        elif fmt == "markdown":
            content = self._to_markdown(export_data)
        else:
            content = self._to_csv(export_data)

        return {
            "canvas_id": canvas_id,
            "format": fmt,
            "content": content,
            "note_count": len(notes),
        }

    @staticmethod
    def _to_markdown(data: dict[str, Any]) -> str:
        lines = [f"# {data.get('canvas_name') or data.get('canvas_id')}", ""]
        if data.get("summary"):
            lines.extend(["## Summary", "", str(data["summary"]), ""])
        lines.append("## Notes")
        lines.append("")
        for note in data.get("notes", []):
            if not isinstance(note, dict):
                continue
            text = note.get("text") or note.get("content") or ""
            color = note.get("background_color") or note.get("color") or ""
            lines.append(f"- ({color}) {text}")
        return "\n".join(lines)

    @staticmethod
    def _to_csv(data: dict[str, Any]) -> str:
        buf = io.StringIO()
        writer = csv.writer(buf)
        writer.writerow(["id", "color", "text", "position_x", "position_y"])
        for note in data.get("notes", []):
            if not isinstance(note, dict):
                continue
            pos = note.get("position") if isinstance(note.get("position"), dict) else {}
            writer.writerow(
                [
                    note.get("id", ""),
                    note.get("background_color") or note.get("color") or "",
                    note.get("text") or note.get("content") or "",
                    pos.get("x", ""),
                    pos.get("y", ""),
                ]
            )
        return buf.getvalue()


__all__ = [
    "BrainstormingExportTool",
    "BrainstormingInsightReportTool",
    "BrainstormingSummaryReportTool",
]
