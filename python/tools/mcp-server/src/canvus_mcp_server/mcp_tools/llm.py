"""LLM-enhanced MCP tools.

Six tools wrap an Ollama client (originally in legacy
``canvus_mcp_server/llm_client.py`` — 541 LOC) to perform text analysis,
canvas insights, brainstorming enhancement, and connector suggestion.

The Ollama client itself is non-Canvus infrastructure but depends on
~14 settings keys (``LLM_TIMEOUT``, ``LLM_FALLBACK_ENABLED``,
``LLM_OFFLINE_MODE``, ``MODEL_TEMPERATURE``, etc.) and contains
module-scoped mutable global state — neither survives the Round 1+2
lessons baseline (mypy --strict + no module globals + no silent error
swallow). The port is scheduled for Phase 4d as a focused workstream.

This module preserves the public class names and the MCP tool registry
surface; calling :meth:`execute` raises a clear deferral error.
"""

from __future__ import annotations

from typing import Any

from ._helpers import require_str
from .base import BaseMCPTool, MCPToolExecutionError

_DEFERRED_MESSAGE = (
    "Phase 4d follow-up: the Ollama client (legacy llm_client.py, 541 LOC) "
    "must be re-implemented against the new SDK's structlog / settings / "
    "no-globals conventions before these tools can run. Tool surface and "
    "MCP registry entry preserved so MCP clients see a structured error."
)


class _LLMDeferredTool(BaseMCPTool):
    """Shared base for LLM tools whose body is deferred to Phase 4d."""

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        raise MCPToolExecutionError(_DEFERRED_MESSAGE, self.name)


class LLMHealthCheckTool(_LLMDeferredTool):
    def __init__(self) -> None:
        super().__init__(
            name="llm_health_check",
            description="Check LLM service health and model availability.",
        )


class LLMTextAnalysisTool(_LLMDeferredTool):
    def __init__(self) -> None:
        super().__init__(
            name="llm_text_analysis",
            description="LLM-powered text analysis of canvas content.",
        )

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "text", self.name)
        return True


class LLMConnectorSuggestionsTool(_LLMDeferredTool):
    def __init__(self) -> None:
        super().__init__(
            name="llm_connector_suggestions",
            description="LLM-driven suggestions for connectors between widgets.",
        )

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True


class LLMCanvasInsightsTool(_LLMDeferredTool):
    def __init__(self) -> None:
        super().__init__(
            name="llm_canvas_insights",
            description="Generate LLM insights about an entire canvas.",
        )

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True


class LLMEnhancedCorrelationTool(_LLMDeferredTool):
    def __init__(self) -> None:
        super().__init__(
            name="llm_enhanced_correlation",
            description="LLM-enhanced correlation analysis between canvas elements.",
        )

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True


class LLMBrainstormingEnhancementTool(_LLMDeferredTool):
    def __init__(self) -> None:
        super().__init__(
            name="llm_brainstorming_enhancement",
            description="LLM-driven enhancement of brainstorming sessions.",
        )

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True


__all__ = [
    "LLMBrainstormingEnhancementTool",
    "LLMCanvasInsightsTool",
    "LLMConnectorSuggestionsTool",
    "LLMEnhancedCorrelationTool",
    "LLMHealthCheckTool",
    "LLMTextAnalysisTool",
]
