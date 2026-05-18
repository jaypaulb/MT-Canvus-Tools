"""Tests for the canvas-resource MCP tools.

The tools are invoked with a real :class:`canvus_sdk.Client`; the HTTP
boundary is mocked via :mod:`respx` so we exercise the real SDK
serialisation, validation, and error-translation paths.
"""

from __future__ import annotations

import httpx
import pytest
import respx
from canvus_sdk import Client

from canvus_mcp_server.mcp_tools.base import (
    MCPToolExecutionError,
    MCPToolValidationError,
)
from canvus_mcp_server.mcp_tools.canvases import (
    CanvasCopyTool,
    CanvasCreateTool,
    CanvasGetTool,
    CanvasListTool,
    CanvasMoveTool,
    CanvasUpdateTool,
)

CANVAS_ID = "11111111-1111-1111-1111-111111111111"
NEW_CANVAS_ID = "22222222-2222-2222-2222-222222222222"


def _canvas(*, cid: str = CANVAS_ID, name: str = "demo") -> dict[str, object]:
    return {"id": cid, "name": name, "folder_id": "root"}


async def test_canvas_get_returns_dumped_model(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    respx_mock.get(f"/canvases/{CANVAS_ID}").mock(
        return_value=httpx.Response(200, json=_canvas())
    )
    tool = CanvasGetTool(client)
    result = await tool.execute(canvas_id=CANVAS_ID)
    assert result["success"] is True
    assert result["canvas_id"] == CANVAS_ID
    assert result["canvas_data"]["id"] == CANVAS_ID


async def test_canvas_get_validates_canvas_id(client: Client) -> None:
    tool = CanvasGetTool(client)
    with pytest.raises(MCPToolValidationError):
        tool.validate_input()
    with pytest.raises(MCPToolValidationError):
        tool.validate_input(canvas_id="")


async def test_canvas_list_returns_count(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    respx_mock.get("/canvases").mock(
        return_value=httpx.Response(
            200, json=[_canvas(cid="a", name="A"), _canvas(cid="b", name="B")]
        )
    )
    tool = CanvasListTool(client)
    result = await tool.execute()
    assert result["count"] == 2
    assert {c["id"] for c in result["canvases"]} == {"a", "b"}


async def test_canvas_create_sends_payload(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    route = respx_mock.post("/canvases").mock(
        return_value=httpx.Response(201, json=_canvas(name="new"))
    )
    tool = CanvasCreateTool(client)
    result = await tool.execute(name="new", description="hello")
    assert route.called
    sent = route.calls.last.request.content.decode()
    assert '"name":"new"' in sent
    assert '"description":"hello"' in sent
    assert result["canvas_id"] == CANVAS_ID


async def test_canvas_create_validates_name(client: Client) -> None:
    tool = CanvasCreateTool(client)
    with pytest.raises(MCPToolValidationError):
        tool.validate_input()
    with pytest.raises(MCPToolValidationError):
        tool.validate_input(name="   ")


async def test_canvas_update_requires_a_field(client: Client) -> None:
    tool = CanvasUpdateTool(client)
    with pytest.raises(MCPToolValidationError):
        tool.validate_input(canvas_id=CANVAS_ID)


async def test_canvas_update_succeeds(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    respx_mock.patch(f"/canvases/{CANVAS_ID}").mock(
        return_value=httpx.Response(200, json=_canvas(name="renamed"))
    )
    tool = CanvasUpdateTool(client)
    result = await tool.execute(canvas_id=CANVAS_ID, name="renamed")
    assert result["update_result"]["name"] == "renamed"
    assert result["updated_fields"] == {"name": "renamed"}


async def test_canvas_copy_returns_new_id(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    respx_mock.post(f"/canvases/{CANVAS_ID}/copy").mock(
        return_value=httpx.Response(200, json=_canvas(cid=NEW_CANVAS_ID, name="copy"))
    )
    tool = CanvasCopyTool(client)
    result = await tool.execute(canvas_id=CANVAS_ID, new_name="copy")
    assert result["new_canvas_id"] == NEW_CANVAS_ID
    assert result["new_canvas_name"] == "copy"


async def test_canvas_move_succeeds(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    respx_mock.post(f"/canvases/{CANVAS_ID}/move").mock(
        return_value=httpx.Response(200, json=_canvas())
    )
    tool = CanvasMoveTool(client)
    result = await tool.execute(canvas_id=CANVAS_ID, folder_id="other")
    assert result["destination"] == "other"


async def test_canvas_move_accepts_legacy_destination_alias(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    respx_mock.post(f"/canvases/{CANVAS_ID}/move").mock(
        return_value=httpx.Response(200, json=_canvas())
    )
    tool = CanvasMoveTool(client)
    result = await tool.execute(canvas_id=CANVAS_ID, destination="other")
    assert result["destination"] == "other"


async def test_canvas_get_translates_api_error(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    respx_mock.get(f"/canvases/{CANVAS_ID}").mock(
        return_value=httpx.Response(404, json={"message": "no such canvas"})
    )
    tool = CanvasGetTool(client)
    with pytest.raises(MCPToolExecutionError) as exc:
        await tool.execute(canvas_id=CANVAS_ID)
    assert "Retrieve canvas" in exc.value.message
