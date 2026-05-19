"""LLM-enhanced MCP tools.

Phase 4d port of the legacy ``canvus_mcp_server/mcp_tools/llm_tools.py``
(750 LOC). Tools take an injected :class:`OllamaClient` (no module-level
globals) and surface failures via :class:`MCPToolExecutionError` — the
legacy silent-fallback ``LLM_FALLBACK_ENABLED`` / ``LLM_OFFLINE_MODE`` paths
are intentionally not ported (see ``docs/conventions/python.md`` §Error
handling — no silent fallbacks).
"""

from __future__ import annotations

import json
from typing import Any

import structlog

from ..llm import OllamaClient, OllamaError
from ._helpers import optional_str, require_str
from .base import BaseMCPTool, MCPToolExecutionError

logger = structlog.get_logger(__name__)


def _safe_json_loads(raw: str) -> Any:
    """Parse JSON, stripping common ``` fences."""
    text = raw.strip()
    if text.startswith("```"):
        lines = text.splitlines()
        if lines and lines[0].startswith("```"):
            lines = lines[1:]
        if lines and lines[-1].startswith("```"):
            lines = lines[:-1]
        text = "\n".join(lines).strip()
    return json.loads(text)


class LLMHealthCheckTool(BaseMCPTool):
    """Check Ollama health and report installed models."""

    def __init__(self, ollama: OllamaClient) -> None:
        super().__init__(
            name="llm_health_check",
            description="Check LLM service health and model availability.",
        )
        self._ollama = ollama

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        try:
            is_healthy = await self._ollama.health_check()
            models = await self._ollama.get_available_models()
        except OllamaError as exc:
            raise MCPToolExecutionError(
                f"LLM health check failed: {exc}", self.name
            ) from exc
        configured = self._ollama.config.model
        return {
            "service_healthy": is_healthy,
            "available_models": models,
            "configured_model": configured,
            "configured_model_available": configured in models,
            "status": "healthy" if is_healthy else "unhealthy",
        }


class LLMTextAnalysisTool(BaseMCPTool):
    """Run similarity / keyword / summary analysis on free text."""

    _ALLOWED_TYPES = {"similarity", "keywords", "summary"}

    def __init__(self, ollama: OllamaClient) -> None:
        super().__init__(
            name="llm_text_analysis",
            description="LLM-powered text analysis (similarity, keywords, summary).",
        )
        self._ollama = ollama

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "text1", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        text1 = require_str(kwargs, "text1", self.name)
        text2 = optional_str(kwargs, "text2", self.name)
        analysis_type = (optional_str(kwargs, "analysis_type", self.name) or "similarity").lower()
        model = optional_str(kwargs, "model", self.name)
        if analysis_type not in self._ALLOWED_TYPES:
            raise MCPToolExecutionError(
                f"Unsupported analysis_type {analysis_type!r}; "
                f"expected one of {sorted(self._ALLOWED_TYPES)}",
                self.name,
            )

        try:
            if analysis_type == "similarity":
                if not text2:
                    raise MCPToolExecutionError(
                        "text2 is required for similarity analysis", self.name
                    )
                result = await self._similarity(text1, text2, model)
            elif analysis_type == "keywords":
                result = await self._keywords(text1, model)
            else:
                result = await self._summary(text1, model)
        except OllamaError as exc:
            raise MCPToolExecutionError(
                f"LLM text analysis failed: {exc}", self.name
            ) from exc

        result.update(
            {
                "analysis_type": analysis_type,
                "model": model or self._ollama.config.model,
                "text_length": len(text1),
            }
        )
        return result

    async def _similarity(
        self, text1: str, text2: str, model: str | None
    ) -> dict[str, Any]:
        prompt = (
            "Compare these two texts and respond ONLY with JSON of the form "
            "{\"similarity_score\": <0..1>, \"common_themes\": [...], "
            "\"differences\": [...], \"relationship\": \"...\", "
            "\"confidence\": <0..1>}.\n\n"
            f"Text 1:\n{text1}\n\nText 2:\n{text2}"
        )
        raw = await self._ollama.generate(
            prompt,
            model=model,
            system="You are an expert text analyst. Reply with JSON only.",
            temperature=0.3,
        )
        try:
            parsed = _safe_json_loads(raw)
        except json.JSONDecodeError as exc:
            raise MCPToolExecutionError(
                f"similarity response was not valid JSON: {exc.msg}",
                self.name,
            ) from exc
        return {"similarity": parsed}

    async def _keywords(self, text: str, model: str | None) -> dict[str, Any]:
        prompt = (
            "Extract the most important keywords from this text. "
            "Respond ONLY with a JSON array of strings.\n\nText:\n" + text
        )
        raw = await self._ollama.generate(
            prompt,
            model=model,
            system="You are a keyword extraction expert. Reply with JSON array only.",
            temperature=0.2,
        )
        try:
            parsed = _safe_json_loads(raw)
        except json.JSONDecodeError as exc:
            raise MCPToolExecutionError(
                f"keywords response was not valid JSON: {exc.msg}",
                self.name,
            ) from exc
        if not isinstance(parsed, list):
            raise MCPToolExecutionError(
                "keywords response was not a JSON array", self.name
            )
        return {"keywords": parsed}

    async def _summary(self, text: str, model: str | None) -> dict[str, Any]:
        prompt = (
            "Summarise the following text in 3-5 sentences, preserving key "
            "facts. Do not invent details.\n\nText:\n" + text
        )
        raw = await self._ollama.generate(
            prompt,
            model=model,
            system="You are a careful, factual summariser.",
            temperature=0.3,
        )
        return {"summary": raw.strip()}


class LLMConnectorSuggestionsTool(BaseMCPTool):
    """Suggest connectors between provided canvas elements via LLM."""

    def __init__(self, ollama: OllamaClient) -> None:
        super().__init__(
            name="llm_connector_suggestions",
            description="LLM-driven suggestions for connectors between widgets.",
        )
        self._ollama = ollama

    def validate_input(self, **kwargs: Any) -> bool:
        elements = kwargs.get("elements")
        if not isinstance(elements, list):
            raise MCPToolExecutionError(
                "elements must be a list of objects", self.name
            )
        if len(elements) < 2:
            raise MCPToolExecutionError(
                "At least 2 elements required for connector suggestions",
                self.name,
            )
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        elements = list(kwargs.get("elements", []))
        model = optional_str(kwargs, "model", self.name)

        prompt = (
            "Given these canvas elements, propose connector pairs (source_id, "
            "target_id) with a connection_type and confidence (0..1). "
            "Respond ONLY with a JSON array of objects.\n\n"
            f"Elements:\n{json.dumps(elements, indent=2)}"
        )
        try:
            raw = await self._ollama.generate(
                prompt,
                model=model,
                system="You analyse relationships between ideas. Reply with JSON only.",
                temperature=0.4,
            )
            parsed = _safe_json_loads(raw)
        except OllamaError as exc:
            raise MCPToolExecutionError(
                f"LLM connector suggestions failed: {exc}", self.name
            ) from exc
        except json.JSONDecodeError as exc:
            raise MCPToolExecutionError(
                f"connector suggestions response was not valid JSON: {exc.msg}",
                self.name,
            ) from exc

        if not isinstance(parsed, list):
            raise MCPToolExecutionError(
                "connector suggestions response was not a JSON array",
                self.name,
            )
        return {
            "suggestions": parsed,
            "total_elements": len(elements),
            "model": model or self._ollama.config.model,
            "suggestions_count": len(parsed),
        }


class LLMCanvasInsightsTool(BaseMCPTool):
    """Generate high-level insights about a canvas payload."""

    def __init__(self, ollama: OllamaClient) -> None:
        super().__init__(
            name="llm_canvas_insights",
            description="Generate LLM insights about an entire canvas.",
        )
        self._ollama = ollama

    def validate_input(self, **kwargs: Any) -> bool:
        if "canvas_data" not in kwargs and "canvas_id" not in kwargs:
            raise MCPToolExecutionError(
                "either canvas_data or canvas_id must be provided", self.name
            )
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_data = kwargs.get("canvas_data")
        if canvas_data is None:
            # Convenience: accept just an id and ask the caller to pass data
            # — the tool itself is SDK-agnostic.
            canvas_data = {"canvas_id": require_str(kwargs, "canvas_id", self.name)}
        model = optional_str(kwargs, "model", self.name)
        elements = canvas_data.get("elements", []) if isinstance(canvas_data, dict) else []

        notes_count = sum(
            1 for e in elements if isinstance(e, dict) and e.get("type") == "note"
        )
        images_count = sum(
            1 for e in elements if isinstance(e, dict) and e.get("type") == "image"
        )
        videos_count = sum(
            1 for e in elements if isinstance(e, dict) and e.get("type") == "video"
        )

        prompt = (
            "Analyse this canvas and produce insights. Respond ONLY with JSON: "
            "{\"main_themes\": [...], \"content_analysis\": \"...\", "
            "\"structure_assessment\": \"...\", "
            "\"improvement_suggestions\": [...], \"overall_purpose\": \"...\", "
            "\"effectiveness_score\": <0..1>}.\n\n"
            f"Canvas data:\n{json.dumps(canvas_data, default=str)[:8000]}"
        )
        try:
            raw = await self._ollama.generate(
                prompt,
                model=model,
                system="You analyse visual content. Reply with JSON only.",
                temperature=0.4,
            )
            insights = _safe_json_loads(raw)
        except OllamaError as exc:
            raise MCPToolExecutionError(
                f"LLM canvas insights failed: {exc}", self.name
            ) from exc
        except json.JSONDecodeError as exc:
            raise MCPToolExecutionError(
                f"canvas insights response was not valid JSON: {exc.msg}",
                self.name,
            ) from exc

        return {
            "insights": insights,
            "canvas_stats": {
                "total_elements": len(elements),
                "notes_count": notes_count,
                "images_count": images_count,
                "videos_count": videos_count,
            },
            "model": model or self._ollama.config.model,
        }


class LLMEnhancedCorrelationTool(BaseMCPTool):
    """Combine simple keyword analysis with an LLM semantic pass."""

    def __init__(self, ollama: OllamaClient) -> None:
        super().__init__(
            name="llm_enhanced_correlation",
            description="LLM-enhanced correlation analysis between canvas elements.",
        )
        self._ollama = ollama

    def validate_input(self, **kwargs: Any) -> bool:
        elements = kwargs.get("elements")
        if not isinstance(elements, list) or not elements:
            raise MCPToolExecutionError(
                "elements must be a non-empty list", self.name
            )
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        elements = list(kwargs.get("elements", []))
        model = optional_str(kwargs, "model", self.name)

        traditional = self._traditional(elements)

        joined = "\n".join(
            str(e.get("content", "")) for e in elements
            if isinstance(e, dict) and e.get("content")
        )
        prompt = (
            "Analyse these canvas elements for semantic relationships. "
            "Respond ONLY with JSON: {\"semantic_themes\": [...], "
            "\"relationship_patterns\": [...], \"content_clusters\": [...], "
            "\"semantic_coherence\": <0..1>, \"complexity_assessment\": \"...\"}.\n\n"
            f"Content:\n{joined[:8000]}"
        )
        try:
            raw = await self._ollama.generate(
                prompt,
                model=model,
                system="You are a semantic analyst. Reply with JSON only.",
                temperature=0.3,
            )
            llm_results = _safe_json_loads(raw)
        except OllamaError as exc:
            raise MCPToolExecutionError(
                f"enhanced correlation LLM call failed: {exc}", self.name
            ) from exc
        except json.JSONDecodeError as exc:
            raise MCPToolExecutionError(
                f"enhanced correlation response was not valid JSON: {exc.msg}",
                self.name,
            ) from exc

        combined_themes = sorted(
            set(traditional.get("top_keywords", []))
            | set(
                llm_results.get("semantic_themes", [])
                if isinstance(llm_results, dict)
                else []
            )
        )
        return {
            "traditional_analysis": traditional,
            "llm_analysis": llm_results,
            "combined_results": {
                "combined_themes": combined_themes,
                "analysis_confidence": 0.8,
                "insights_summary": (
                    f"Analysed {len(elements)} elements; "
                    f"{traditional['note_count']} notes."
                ),
            },
            "model": model or self._ollama.config.model,
            "analysis_method": "enhanced_correlation",
        }

    @staticmethod
    def _traditional(elements: list[Any]) -> dict[str, Any]:
        notes = [
            e for e in elements
            if isinstance(e, dict) and e.get("type") == "note"
        ]
        all_content = " ".join(str(e.get("content", "")) for e in notes)
        word_freq: dict[str, int] = {}
        for raw in all_content.lower().split():
            if len(raw) > 3:
                word_freq[raw] = word_freq.get(raw, 0) + 1
        top = sorted(word_freq.items(), key=lambda x: x[1], reverse=True)[:10]
        return {
            "total_elements": len(elements),
            "note_count": len(notes),
            "top_keywords": [w for w, _ in top],
            "keyword_frequencies": dict(top),
            "analysis_type": "traditional",
        }


class LLMBrainstormingEnhancementTool(BaseMCPTool):
    """LLM-driven enhancement of a brainstorming session payload."""

    _ALLOWED_TYPES = {"analysis", "ideas", "connections", "summary"}

    def __init__(self, ollama: OllamaClient) -> None:
        super().__init__(
            name="llm_brainstorming_enhancement",
            description="LLM-driven enhancement of brainstorming sessions.",
        )
        self._ollama = ollama

    def validate_input(self, **kwargs: Any) -> bool:
        session_data = kwargs.get("session_data")
        if not isinstance(session_data, dict):
            raise MCPToolExecutionError(
                "session_data must be an object", self.name
            )
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        session_data = kwargs.get("session_data", {})
        enhancement_type = (
            optional_str(kwargs, "enhancement_type", self.name) or "analysis"
        ).lower()
        model = optional_str(kwargs, "model", self.name)
        if enhancement_type not in self._ALLOWED_TYPES:
            raise MCPToolExecutionError(
                f"Unsupported enhancement_type {enhancement_type!r}; "
                f"expected one of {sorted(self._ALLOWED_TYPES)}",
                self.name,
            )

        notes = session_data.get("notes", [])
        try:
            if enhancement_type == "analysis":
                result = self._analysis(notes, session_data)
            elif enhancement_type == "ideas":
                result = {
                    "new_ideas": await self._generate_ideas(notes, model),
                }
            elif enhancement_type == "connections":
                result = {
                    "suggested_connections": await self._suggest_connections(
                        notes, model
                    ),
                }
            else:
                result = {
                    "session_summary": await self._summary(notes, session_data, model),
                }
        except OllamaError as exc:
            raise MCPToolExecutionError(
                f"brainstorming enhancement failed: {exc}", self.name
            ) from exc

        result.update(
            {
                "enhancement_type": enhancement_type,
                "model": model or self._ollama.config.model,
            }
        )
        return result

    @staticmethod
    def _analysis(notes: list[Any], session_data: dict[str, Any]) -> dict[str, Any]:
        color_counts: dict[str, int] = {}
        for note in notes:
            if not isinstance(note, dict):
                continue
            color = str(note.get("color", "unknown"))
            color_counts[color] = color_counts.get(color, 0) + 1
        most_active = (
            max(color_counts.items(), key=lambda x: x[1])[0]
            if color_counts
            else "unknown"
        )
        return {
            "total_notes": len(notes),
            "participant_count": len(color_counts),
            "notes_per_participant": color_counts,
            "most_active_participant": most_active,
            "session_duration": session_data.get("duration"),
            "session_topic": session_data.get("topic", "Brainstorming Session"),
        }

    async def _generate_ideas(
        self, notes: list[Any], model: str | None
    ) -> list[Any]:
        if not notes:
            return []
        content = "\n".join(
            str(n.get("content", "")) for n in notes if isinstance(n, dict)
        )
        prompt = (
            "Generate 5-10 new related ideas. Respond ONLY with a JSON array "
            "of objects with keys idea, category, complexity, potential_impact.\n\n"
            f"Existing ideas:\n{content}"
        )
        raw = await self._ollama.generate(
            prompt,
            model=model,
            system="You are a creative brainstorming facilitator. Reply with JSON only.",
            temperature=0.7,
        )
        try:
            parsed = _safe_json_loads(raw)
        except json.JSONDecodeError as exc:
            raise MCPToolExecutionError(
                f"ideas response was not valid JSON: {exc.msg}", self.name
            ) from exc
        if not isinstance(parsed, list):
            raise MCPToolExecutionError(
                "ideas response was not a JSON array", self.name
            )
        return parsed

    async def _suggest_connections(
        self, notes: list[Any], model: str | None
    ) -> list[Any]:
        if len(notes) < 2:
            return []
        prompt = (
            "Identify the strongest pairwise connections between these "
            "brainstorming notes. Respond ONLY with a JSON array of objects "
            "{source_id, target_id, connection_type, strength (0..1), reasoning}. "
            "Include only strength > 0.5.\n\n"
            f"Notes:\n{json.dumps(notes, default=str)[:8000]}"
        )
        raw = await self._ollama.generate(
            prompt,
            model=model,
            system="You identify relationships between ideas. Reply with JSON only.",
            temperature=0.4,
        )
        try:
            parsed = _safe_json_loads(raw)
        except json.JSONDecodeError as exc:
            raise MCPToolExecutionError(
                f"connections response was not valid JSON: {exc.msg}",
                self.name,
            ) from exc
        if not isinstance(parsed, list):
            raise MCPToolExecutionError(
                "connections response was not a JSON array", self.name
            )
        return [
            c for c in parsed
            if isinstance(c, dict) and float(c.get("strength", 0) or 0) > 0.5
        ]

    async def _summary(
        self, notes: list[Any], session_data: dict[str, Any], model: str | None
    ) -> dict[str, Any]:
        session_info = {
            "total_notes": len(notes),
            "participant_count": len(
                {
                    n.get("color") for n in notes
                    if isinstance(n, dict) and n.get("color")
                }
            ),
            "session_duration": session_data.get("duration"),
            "session_topic": session_data.get("topic", "Brainstorming Session"),
        }
        prompt = (
            "Generate 3-5 actionable next steps for this brainstorming session. "
            "Respond ONLY with a JSON array of objects "
            "{action, priority, assignee, timeline, description}.\n\n"
            f"Session info:\n{json.dumps(session_info)}\n\n"
            f"Notes sample:\n"
            + "\n".join(
                str(n.get("content", "")) for n in notes[:10] if isinstance(n, dict)
            )
        )
        raw = await self._ollama.generate(
            prompt,
            model=model,
            system="You are a project manager. Reply with JSON only.",
            temperature=0.5,
        )
        try:
            action_items = _safe_json_loads(raw)
        except json.JSONDecodeError as exc:
            raise MCPToolExecutionError(
                f"summary response was not valid JSON: {exc.msg}",
                self.name,
            ) from exc

        return {
            "session_info": session_info,
            "action_items": action_items,
            "notable_ideas": [
                str(n.get("content", ""))
                for n in notes[:5]
                if isinstance(n, dict)
            ],
        }


__all__ = [
    "LLMBrainstormingEnhancementTool",
    "LLMCanvasInsightsTool",
    "LLMConnectorSuggestionsTool",
    "LLMEnhancedCorrelationTool",
    "LLMHealthCheckTool",
    "LLMTextAnalysisTool",
]
