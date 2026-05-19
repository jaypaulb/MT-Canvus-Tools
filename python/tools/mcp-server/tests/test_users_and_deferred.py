"""Tests for user-management tools and the Phase 4d LLM-port tools.

The previous version of this module asserted the Phase 4d deferral
errors; those tools are now real implementations backed by an injected
:class:`OllamaClient` (Phase 4d Round D). The LLM tests below use
``respx`` to mock the Ollama HTTP surface — no network access required.
"""

from __future__ import annotations

import json
from collections.abc import AsyncIterator

import httpx
import pytest
import pytest_asyncio
import respx
from canvus_mcp_server.llm import OllamaClient, OllamaConfig
from canvus_mcp_server.mcp_tools.base import (
    MCPToolExecutionError,
    MCPToolValidationError,
)
from canvus_mcp_server.mcp_tools.brainstorming import (
    BrainstormingNoteRetrievalTool,
)
from canvus_mcp_server.mcp_tools.correlation import (
    ElementRelationshipAnalysisTool,
)
from canvus_mcp_server.mcp_tools.llm import LLMHealthCheckTool, LLMTextAnalysisTool
from canvus_mcp_server.mcp_tools.reports import BrainstormingSummaryReportTool
from canvus_mcp_server.mcp_tools.users import (
    GetCanvasPermissionsTool,
    GetCurrentUserTool,
    UserCreateTool,
    UserListTool,
    UserLoginTool,
    UserLogoutTool,
)

from canvus_sdk import Client

CANVAS_ID = "00000000-0000-0000-0000-000000000000"
OLLAMA_URL = "http://localhost:11434"


@pytest_asyncio.fixture
async def ollama() -> AsyncIterator[OllamaClient]:
    cfg = OllamaConfig(
        base_url=OLLAMA_URL,
        model="gemma3:2b",
        timeout_seconds=5.0,
        max_retries=1,
    )
    c = OllamaClient(cfg)
    try:
        yield c
    finally:
        await c.aclose()


# ---- user tools (unchanged behaviour) -------------------------------------


async def test_user_login_accepts_username_alias(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    respx_mock.post("/users/login").mock(
        return_value=httpx.Response(
            200,
            json={
                "token": "ck_x",
                "user": {"id": 1, "email": "a@b", "name": "Alice"},
            },
        )
    )
    tool = UserLoginTool(client)
    result = await tool.execute(username="a@b", password="pw")
    assert result["success"] is True


async def test_user_logout(client: Client, respx_mock: respx.MockRouter) -> None:
    respx_mock.post("/users/logout").mock(
        return_value=httpx.Response(200, json={})
    )
    result = await UserLogoutTool(client).execute()
    assert result["success"] is True


async def test_get_current_user_uses_users_current_endpoint(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    respx_mock.get("/users/current").mock(
        return_value=httpx.Response(
            200, json={"id": 7, "email": "me@x", "name": "Me"}
        )
    )
    result = await GetCurrentUserTool(client).execute()
    assert result["user"]["id"] == 7


async def test_user_list(client: Client, respx_mock: respx.MockRouter) -> None:
    respx_mock.get("/users").mock(
        return_value=httpx.Response(
            200,
            json=[
                {"id": 1, "email": "a@b", "name": "Alice"},
                {"id": 2, "email": "c@d", "name": "Bob"},
            ],
        )
    )
    result = await UserListTool(client).execute()
    assert result["count"] == 2


async def test_user_create_validates_payload(client: Client) -> None:
    with pytest.raises(MCPToolValidationError):
        UserCreateTool(client).validate_input()


async def test_get_canvas_permissions(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    respx_mock.get(f"/canvases/{CANVAS_ID}/permissions").mock(
        return_value=httpx.Response(200, json={"editors_can_share": True})
    )
    result = await GetCanvasPermissionsTool(client).execute(canvas_id=CANVAS_ID)
    assert result["permissions"]["editors_can_share"] is True


# ---- Phase 4d Round D LLM tools (now real) --------------------------------


async def test_llm_health_check_reports_models(ollama: OllamaClient) -> None:
    with respx.mock(base_url=OLLAMA_URL, assert_all_called=False) as router:
        router.get("/api/tags").mock(
            return_value=httpx.Response(
                200, json={"models": [{"name": "gemma3:2b"}]}
            )
        )
        result = await LLMHealthCheckTool(ollama).execute()
    assert result["service_healthy"] is True
    assert result["status"] == "healthy"
    assert "gemma3:2b" in result["available_models"]
    assert result["configured_model_available"] is True


async def test_llm_health_check_propagates_transport_failure(
    ollama: OllamaClient,
) -> None:
    with respx.mock(base_url=OLLAMA_URL, assert_all_called=False) as router:
        router.get("/api/tags").mock(side_effect=httpx.ConnectError("nope"))
        with pytest.raises(MCPToolExecutionError) as exc:
            await LLMHealthCheckTool(ollama).execute()
    assert "LLM health check failed" in exc.value.message


async def test_llm_text_analysis_keywords(ollama: OllamaClient) -> None:
    with respx.mock(base_url=OLLAMA_URL, assert_all_called=False) as router:
        router.post("/api/generate").mock(
            return_value=httpx.Response(
                200, json={"response": json.dumps(["alpha", "beta"])}
            )
        )
        result = await LLMTextAnalysisTool(ollama).execute(
            text1="some content", analysis_type="keywords"
        )
    assert result["keywords"] == ["alpha", "beta"]
    assert result["analysis_type"] == "keywords"


async def test_correlation_tool_runs_against_empty_canvas(
    client: Client,
    ollama: OllamaClient,
    respx_mock: respx.MockRouter,
) -> None:
    respx_mock.get(f"/canvases/{CANVAS_ID}").mock(
        return_value=httpx.Response(200, json={"id": CANVAS_ID, "name": "x"})
    )
    respx_mock.get(f"/canvases/{CANVAS_ID}/notes").mock(
        return_value=httpx.Response(200, json=[])
    )
    respx_mock.get(f"/canvases/{CANVAS_ID}/connectors").mock(
        return_value=httpx.Response(200, json=[])
    )
    with respx.mock(base_url=OLLAMA_URL, assert_all_called=False) as ollama_router:
        ollama_router.post("/api/generate").mock(
            return_value=httpx.Response(200, json={"response": "[]"})
        )
        result = await ElementRelationshipAnalysisTool(client, ollama).execute(
            canvas_id=CANVAS_ID
        )
    assert result["note_count"] == 0
    assert result["spatial_relationships"] == []
    assert result["semantic_relationships"] == []


async def test_brainstorming_tool_validation(
    client: Client, ollama: OllamaClient
) -> None:
    tool = BrainstormingNoteRetrievalTool(client, ollama)
    with pytest.raises(MCPToolValidationError):
        tool.validate_input()


async def test_report_summary_runs_against_empty_canvas(
    client: Client,
    ollama: OllamaClient,
    respx_mock: respx.MockRouter,
) -> None:
    respx_mock.get(f"/canvases/{CANVAS_ID}").mock(
        return_value=httpx.Response(200, json={"id": CANVAS_ID, "name": "x"})
    )
    respx_mock.get(f"/canvases/{CANVAS_ID}/notes").mock(
        return_value=httpx.Response(200, json=[])
    )
    with respx.mock(base_url=OLLAMA_URL, assert_all_called=False) as ollama_router:
        ollama_router.post("/api/generate").mock(
            return_value=httpx.Response(
                200,
                json={
                    "response": json.dumps(
                        {
                            "executive_summary": "Nothing to summarise.",
                            "main_themes": [],
                            "top_ideas": [],
                            "next_steps": [],
                        }
                    )
                },
            )
        )
        result = await BrainstormingSummaryReportTool(client, ollama).execute(
            canvas_id=CANVAS_ID
        )
    assert result["canvas_id"] == CANVAS_ID
    assert result["note_count"] == 0
    assert result["summary"]["executive_summary"] == "Nothing to summarise."
