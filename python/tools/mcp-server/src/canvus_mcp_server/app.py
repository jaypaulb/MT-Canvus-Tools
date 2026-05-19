"""FastAPI application + MCP tool wiring.

The application exposes:

- ``/health`` — basic liveness probe.
- ``/health/llm`` — Ollama probe (no auth).
- ``/tools`` — registry inspection.
- ``/tools/{name}/execute`` — synchronous tool invocation (POST JSON body).
- ``/mcp`` — the fastapi-mcp transport mount (when ``fastapi-mcp`` is
  installed). If the package is unavailable at import time the mount is
  skipped and a warning is logged; the HTTP API still works.

The legacy server's 730-LOC ``auth.py`` JWT system + 523-LOC
``auth_endpoints.py`` is intentionally not ported in this round — it is a
custom session-token store that predates the SDK's resource-grouped auth
helpers. See the README "Deferred from legacy port" section.
"""

from __future__ import annotations

from collections.abc import Callable
from typing import Any

import httpx
import structlog
from fastapi import FastAPI, HTTPException, status

from canvus_sdk import Client

from .config import Settings
from .llm import OllamaClient
from .logging_config import configure_logging
from .mcp_tools import (
    AutoConnectorCreationTool,
    AutomaticConnectorSuggestionTool,
    BrainstormingExportTool,
    BrainstormingInsightReportTool,
    BrainstormingNoteRetrievalTool,
    BrainstormingSummaryReportTool,
    CanvasCopyTool,
    CanvasCreateTool,
    CanvasGetTool,
    CanvasListTool,
    CanvasMoveTool,
    CanvasUpdateTool,
    ConnectorCreateTool,
    ConnectorDeleteTool,
    ConnectorGetTool,
    ConnectorListTool,
    ConnectorUpdateTool,
    ConnectorVisualizationTool,
    ElementRelationshipAnalysisTool,
    GetCanvasPermissionsTool,
    GetCurrentUserTool,
    ImageCreateTool,
    ImageDeleteTool,
    ImageGetTool,
    ImageListTool,
    ImageUpdateTool,
    LLMBrainstormingAnalysisTool,
    LLMBrainstormingEnhancementTool,
    LLMCanvasInsightsTool,
    LLMConnectorSuggestionsTool,
    LLMEnhancedCorrelationTool,
    LLMHealthCheckTool,
    LLMTextAnalysisTool,
    MCPToolError,
    MCPToolRegistry,
    MCPToolValidationError,
    NoteCreateTool,
    NoteDeleteTool,
    NoteGetTool,
    NoteListTool,
    NoteUpdateTool,
    PdfCreateTool,
    PdfGetTool,
    PdfListTool,
    PdfUpdateTool,
    PersonaIdentificationTool,
    UserCreateTool,
    UserListTool,
    UserLoginTool,
    UserLogoutTool,
    VideoCreateTool,
    VideoDeleteTool,
    VideoGetTool,
    VideoListTool,
    VideoUpdateTool,
)

logger = structlog.get_logger(__name__)


def build_client(settings: Settings) -> Client:
    """Construct a :class:`canvus_sdk.Client` from settings."""
    if not settings.api_key:
        raise ValueError("CANVUS_API_KEY environment variable is required")
    return Client(
        settings.api_url,
        settings.api_key,
        verify_ssl=settings.verify_ssl,
    )


def build_registry(client: Client, ollama: OllamaClient) -> MCPToolRegistry:
    """Construct and populate a :class:`MCPToolRegistry`.

    The flat tool list (typed as :class:`BaseMCPTool`) preserves the legacy
    per-family ordering while keeping mypy happy across the heterogeneous
    concrete classes.
    """
    from .mcp_tools.base import BaseMCPTool

    registry = MCPToolRegistry()
    tools: list[BaseMCPTool] = [
        # Canvases
        CanvasGetTool(client),
        CanvasListTool(client),
        CanvasCreateTool(client),
        CanvasUpdateTool(client),
        CanvasCopyTool(client),
        CanvasMoveTool(client),
        # Notes
        NoteCreateTool(client),
        NoteGetTool(client),
        NoteListTool(client),
        NoteUpdateTool(client),
        NoteDeleteTool(client),
        # Images
        ImageCreateTool(client),
        ImageGetTool(client),
        ImageListTool(client),
        ImageUpdateTool(client),
        ImageDeleteTool(client),
        # Videos
        VideoCreateTool(client),
        VideoGetTool(client),
        VideoListTool(client),
        VideoUpdateTool(client),
        VideoDeleteTool(client),
        # PDFs
        PdfCreateTool(client),
        PdfGetTool(client),
        PdfListTool(client),
        PdfUpdateTool(client),
        # Connectors
        ConnectorCreateTool(client),
        ConnectorGetTool(client),
        ConnectorListTool(client),
        ConnectorUpdateTool(client),
        ConnectorDeleteTool(client),
        # Users
        UserLoginTool(client),
        UserLogoutTool(client),
        GetCurrentUserTool(client),
        UserListTool(client),
        UserCreateTool(client),
        GetCanvasPermissionsTool(client),
        # Correlation
        ElementRelationshipAnalysisTool(client, ollama),
        AutomaticConnectorSuggestionTool(client, ollama),
        ConnectorVisualizationTool(client, ollama),
        # LLM
        LLMHealthCheckTool(ollama),
        LLMTextAnalysisTool(ollama),
        LLMConnectorSuggestionsTool(ollama),
        LLMCanvasInsightsTool(ollama),
        LLMEnhancedCorrelationTool(ollama),
        LLMBrainstormingEnhancementTool(ollama),
        # Brainstorming
        BrainstormingNoteRetrievalTool(client, ollama),
        PersonaIdentificationTool(client, ollama),
        LLMBrainstormingAnalysisTool(client, ollama),
        AutoConnectorCreationTool(client, ollama),
        # Reports
        BrainstormingSummaryReportTool(client, ollama),
        BrainstormingInsightReportTool(client, ollama),
        BrainstormingExportTool(client, ollama),
    ]
    for tool in tools:
        registry.register_tool(tool)
    return registry


def create_app(
    settings: Settings | None = None,
    *,
    client: Client | None = None,
    ollama: OllamaClient | None = None,
    registry: MCPToolRegistry | None = None,
) -> FastAPI:
    """Build the FastAPI application.

    The factory makes everything injectable so the test suite can pass in a
    mocked SDK client without touching the network. Callers that don't
    care just call ``create_app()`` with no arguments.
    """
    configure_logging()
    if settings is None:
        settings = Settings()  # type: ignore[call-arg]
    settings.ensure_directories()
    if client is None:
        client = build_client(settings)
    if ollama is None:
        ollama = OllamaClient(settings.build_ollama_config())
    if registry is None:
        registry = build_registry(client, ollama)

    app = FastAPI(
        title="Canvus MCP Server",
        description=(
            "MCP server exposing Canvus operations as MCP tools backed by "
            "the canvus-sdk Python client."
        ),
        version="0.1.0",
        docs_url="/docs" if not settings.is_production else None,
        redoc_url="/redoc" if not settings.is_production else None,
    )
    app.state.settings = settings
    app.state.client = client
    app.state.ollama = ollama
    app.state.registry = registry

    _wire_routes(app, settings, registry)
    _mount_mcp(app)
    return app


def _wire_routes(
    app: FastAPI, settings: Settings, registry: MCPToolRegistry
) -> None:
    @app.get("/health")
    async def health() -> dict[str, Any]:
        return {
            "status": "ok",
            "registered_tools": len(registry),
            "api_url": settings.api_url,
        }

    @app.get("/health/llm")
    async def llm_health() -> dict[str, Any]:
        if not settings.ollama_enabled:
            return {"status": "disabled"}
        url = f"{settings.effective_ollama_base_url}/api/version"
        try:
            async with httpx.AsyncClient(
                timeout=settings.health_check_timeout
            ) as http:
                response = await http.get(url)
                response.raise_for_status()
                return {
                    "status": "healthy",
                    "ollama_url": settings.effective_ollama_base_url,
                    "default_model": settings.default_model,
                    "ollama_response": response.json(),
                }
        except httpx.HTTPError as exc:
            return {
                "status": "unhealthy",
                "ollama_url": settings.effective_ollama_base_url,
                "error": str(exc),
            }

    @app.get("/tools")
    async def list_tools() -> dict[str, Any]:
        return {
            "tools": registry.list_tool_info(),
            "count": len(registry),
        }

    @app.get("/tools/{tool_name}")
    async def get_tool_info(tool_name: str) -> dict[str, Any]:
        tool = registry.get_tool(tool_name)
        if tool is None:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Tool {tool_name!r} not found",
            )
        return tool.get_tool_info()

    @app.post("/tools/{tool_name}/execute")
    async def execute_tool(
        tool_name: str, parameters: dict[str, Any]
    ) -> dict[str, Any]:
        tool = registry.get_tool(tool_name)
        if tool is None:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Tool {tool_name!r} not found",
            )
        try:
            return await registry.execute_tool(tool_name, **parameters)
        except MCPToolValidationError as exc:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=exc.message,
            ) from exc
        except MCPToolError as exc:
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=exc.message,
            ) from exc


def _mount_mcp(app: FastAPI) -> None:
    """Mount the fastapi-mcp router if the package is available.

    fastapi-mcp's import surface has shifted across releases (``MCPServer``
    in 0.3.x, factory in 0.4.x). We attempt the common shapes and fall
    back to logging if none are found — the HTTP API still works.
    """
    # fastapi-mcp ships without a py.typed marker; the dynamic dispatch
    # below survives multiple package versions whose import surface differs.
    try:  # pragma: no cover - import surface depends on installed version
        from fastapi_mcp import MCPServer
    except ImportError:
        try:
            import fastapi_mcp  # type: ignore[import-untyped]
        except ImportError:
            logger.warning("fastapi_mcp not installed; /mcp route not mounted")
            return
        candidate: Callable[[], Any] | None = getattr(
            fastapi_mcp, "create_mcp_server", None
        )
        if candidate is None:
            logger.warning(
                "fastapi_mcp present but no known factory; /mcp not mounted"
            )
            return
        server: Any = candidate()
    else:
        server = MCPServer()
    router = getattr(server, "router", None)
    if router is None:
        logger.warning("fastapi_mcp server has no .router; /mcp not mounted")
        return
    app.include_router(router, prefix="/mcp")


# Async closing helper used by the CLI entrypoint.
async def aclose(app: FastAPI) -> None:
    """Close the underlying SDK and LLM clients cleanly."""
    client: Client | None = getattr(app.state, "client", None)
    if client is not None:
        await client.aclose()
    ollama: OllamaClient | None = getattr(app.state, "ollama", None)
    if ollama is not None:
        await ollama.aclose()


# Help mypy / linters identify the public surface.
__all__ = ["aclose", "build_client", "build_registry", "create_app"]
