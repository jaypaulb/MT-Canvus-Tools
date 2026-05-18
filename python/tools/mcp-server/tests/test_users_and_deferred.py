"""Tests for user-management tools and the Phase 4d deferred tools."""

from __future__ import annotations

import httpx
import pytest
import respx
from canvus_sdk import Client

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
from canvus_mcp_server.mcp_tools.llm import LLMHealthCheckTool
from canvus_mcp_server.mcp_tools.reports import BrainstormingSummaryReportTool
from canvus_mcp_server.mcp_tools.users import (
    GetCanvasPermissionsTool,
    GetCurrentUserTool,
    UserCreateTool,
    UserListTool,
    UserLoginTool,
    UserLogoutTool,
)

CANVAS_ID = "00000000-0000-0000-0000-000000000000"


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
        return_value=httpx.Response(200, json={"id": 7, "email": "me@x"})
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


# ---- deferred Phase 4d tools surface MCPToolExecutionError ----------------


async def test_llm_health_check_surfaces_deferred_error(client: Client) -> None:
    with pytest.raises(MCPToolExecutionError) as exc:
        await LLMHealthCheckTool().execute()
    assert "Phase 4d" in exc.value.message


async def test_correlation_tool_fetches_canvas_then_defers(
    client: Client, respx_mock: respx.MockRouter
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
    with pytest.raises(MCPToolExecutionError) as exc:
        await ElementRelationshipAnalysisTool(client).execute(canvas_id=CANVAS_ID)
    assert "Phase 4d" in exc.value.message
    assert "fetched_notes=0" in exc.value.message


async def test_brainstorming_tool_validation(client: Client) -> None:
    tool = BrainstormingNoteRetrievalTool(client)
    with pytest.raises(MCPToolValidationError):
        tool.validate_input()


async def test_report_tool_fetches_canvas_then_defers(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    respx_mock.get(f"/canvases/{CANVAS_ID}").mock(
        return_value=httpx.Response(200, json={"id": CANVAS_ID, "name": "x"})
    )
    with pytest.raises(MCPToolExecutionError) as exc:
        await BrainstormingSummaryReportTool(client).execute(canvas_id=CANVAS_ID)
    assert "Phase 4d" in exc.value.message
