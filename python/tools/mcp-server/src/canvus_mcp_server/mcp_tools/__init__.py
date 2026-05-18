"""MCP tool registry — one module per resource family.

Each submodule defines one or more :class:`BaseMCPTool` subclasses; the
registry pattern is the user-visible MCP surface and is preserved across the
port from the legacy ``canvus-mcp-server`` package.

The legacy 3,169-LOC ``canvas.py`` is decomposed here into per-resource
modules (``canvases``, ``notes``, ``images``, ``videos``, ``pdfs``,
``connectors``) per the Phase 4c audit recommendation.
"""

from __future__ import annotations

from .base import (
    BaseMCPTool,
    MCPToolError,
    MCPToolExecutionError,
    MCPToolValidationError,
)
from .brainstorming import (
    AutoConnectorCreationTool,
    BrainstormingNoteRetrievalTool,
    LLMBrainstormingAnalysisTool,
    PersonaIdentificationTool,
)
from .canvases import (
    CanvasCopyTool,
    CanvasCreateTool,
    CanvasGetTool,
    CanvasListTool,
    CanvasMoveTool,
    CanvasUpdateTool,
)
from .connectors import (
    ConnectorCreateTool,
    ConnectorDeleteTool,
    ConnectorGetTool,
    ConnectorListTool,
    ConnectorUpdateTool,
)
from .correlation import (
    AutomaticConnectorSuggestionTool,
    ConnectorVisualizationTool,
    ElementRelationshipAnalysisTool,
)
from .images import (
    ImageCreateTool,
    ImageDeleteTool,
    ImageGetTool,
    ImageListTool,
    ImageUpdateTool,
)
from .llm import (
    LLMBrainstormingEnhancementTool,
    LLMCanvasInsightsTool,
    LLMConnectorSuggestionsTool,
    LLMEnhancedCorrelationTool,
    LLMHealthCheckTool,
    LLMTextAnalysisTool,
)
from .notes import (
    NoteCreateTool,
    NoteDeleteTool,
    NoteGetTool,
    NoteListTool,
    NoteUpdateTool,
)
from .pdfs import (
    PdfCreateTool,
    PdfGetTool,
    PdfListTool,
    PdfUpdateTool,
)
from .registry import MCPToolRegistry, tool_registry
from .reports import (
    BrainstormingExportTool,
    BrainstormingInsightReportTool,
    BrainstormingSummaryReportTool,
)
from .users import (
    GetCanvasPermissionsTool,
    GetCurrentUserTool,
    UserCreateTool,
    UserListTool,
    UserLoginTool,
    UserLogoutTool,
)
from .videos import (
    VideoCreateTool,
    VideoDeleteTool,
    VideoGetTool,
    VideoListTool,
    VideoUpdateTool,
)

__all__ = [
    "AutoConnectorCreationTool",
    "AutomaticConnectorSuggestionTool",
    "BaseMCPTool",
    "BrainstormingExportTool",
    "BrainstormingInsightReportTool",
    "BrainstormingNoteRetrievalTool",
    "BrainstormingSummaryReportTool",
    "CanvasCopyTool",
    "CanvasCreateTool",
    "CanvasGetTool",
    "CanvasListTool",
    "CanvasMoveTool",
    "CanvasUpdateTool",
    "ConnectorCreateTool",
    "ConnectorDeleteTool",
    "ConnectorGetTool",
    "ConnectorListTool",
    "ConnectorUpdateTool",
    "ConnectorVisualizationTool",
    "ElementRelationshipAnalysisTool",
    "GetCanvasPermissionsTool",
    "GetCurrentUserTool",
    "ImageCreateTool",
    "ImageDeleteTool",
    "ImageGetTool",
    "ImageListTool",
    "ImageUpdateTool",
    "LLMBrainstormingAnalysisTool",
    "LLMBrainstormingEnhancementTool",
    "LLMCanvasInsightsTool",
    "LLMConnectorSuggestionsTool",
    "LLMEnhancedCorrelationTool",
    "LLMHealthCheckTool",
    "LLMTextAnalysisTool",
    "MCPToolError",
    "MCPToolExecutionError",
    "MCPToolRegistry",
    "MCPToolValidationError",
    "NoteCreateTool",
    "NoteDeleteTool",
    "NoteGetTool",
    "NoteListTool",
    "NoteUpdateTool",
    "PdfCreateTool",
    "PdfGetTool",
    "PdfListTool",
    "PdfUpdateTool",
    "PersonaIdentificationTool",
    "UserCreateTool",
    "UserListTool",
    "UserLoginTool",
    "UserLogoutTool",
    "VideoCreateTool",
    "VideoDeleteTool",
    "VideoGetTool",
    "VideoListTool",
    "VideoUpdateTool",
    "tool_registry",
]
