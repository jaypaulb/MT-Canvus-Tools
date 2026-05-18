"""Tests for the widget MCP tools (notes + asset-shaped images/videos/pdfs)."""

from __future__ import annotations

from pathlib import Path

import httpx
import pytest
import respx
from canvus_sdk import Client

from canvus_mcp_server.mcp_tools.base import (
    MCPToolExecutionError,
    MCPToolValidationError,
)
from canvus_mcp_server.mcp_tools.connectors import (
    ConnectorCreateTool,
    ConnectorDeleteTool,
    ConnectorListTool,
)
from canvus_mcp_server.mcp_tools.images import (
    ImageCreateTool,
    ImageListTool,
)
from canvus_mcp_server.mcp_tools.notes import (
    NoteCreateTool,
    NoteDeleteTool,
    NoteGetTool,
    NoteListTool,
    NoteUpdateTool,
)
from canvus_mcp_server.mcp_tools.pdfs import PdfListTool
from canvus_mcp_server.mcp_tools.videos import VideoListTool

CANVAS_ID = "00000000-0000-0000-0000-000000000000"
NOTE_ID = "11111111-1111-1111-1111-111111111111"


def _note(*, nid: str = NOTE_ID, text: str = "hello") -> dict[str, object]:
    return {
        "id": nid,
        "widget_type": "Note",
        "text": text,
        "location": {"x": 0.0, "y": 0.0},
    }


# ---- notes ---------------------------------------------------------------


async def test_note_create_validates_payload(client: Client) -> None:
    tool = NoteCreateTool(client)
    with pytest.raises(MCPToolValidationError):
        tool.validate_input(canvas_id=CANVAS_ID)
    with pytest.raises(MCPToolValidationError):
        tool.validate_input(canvas_id=CANVAS_ID, payload={})


async def test_note_create_round_trip(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    respx_mock.post(f"/canvases/{CANVAS_ID}/notes").mock(
        return_value=httpx.Response(201, json=_note())
    )
    tool = NoteCreateTool(client)
    result = await tool.execute(canvas_id=CANVAS_ID, payload={"text": "hello"})
    assert result["note_id"] == NOTE_ID
    assert result["note_data"]["text"] == "hello"


async def test_note_get_list_update_delete(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    respx_mock.get(f"/canvases/{CANVAS_ID}/notes/{NOTE_ID}").mock(
        return_value=httpx.Response(200, json=_note())
    )
    respx_mock.get(f"/canvases/{CANVAS_ID}/notes").mock(
        return_value=httpx.Response(200, json=[_note(), _note(nid="other")])
    )
    respx_mock.patch(f"/canvases/{CANVAS_ID}/notes/{NOTE_ID}").mock(
        return_value=httpx.Response(200, json=_note(text="updated"))
    )
    respx_mock.delete(f"/canvases/{CANVAS_ID}/notes/{NOTE_ID}").mock(
        return_value=httpx.Response(204)
    )

    get_result = await NoteGetTool(client).execute(
        canvas_id=CANVAS_ID, note_id=NOTE_ID
    )
    list_result = await NoteListTool(client).execute(canvas_id=CANVAS_ID)
    update_result = await NoteUpdateTool(client).execute(
        canvas_id=CANVAS_ID, note_id=NOTE_ID, payload={"text": "updated"}
    )
    delete_result = await NoteDeleteTool(client).execute(
        canvas_id=CANVAS_ID, note_id=NOTE_ID
    )

    assert get_result["note_data"]["text"] == "hello"
    assert list_result["count"] == 2
    assert update_result["note_data"]["text"] == "updated"
    assert delete_result["status"] == "deleted"


# ---- assets --------------------------------------------------------------


async def test_image_create_uploads_file_bytes(
    client: Client, respx_mock: respx.MockRouter, tmp_path: Path
) -> None:
    tmp_image = tmp_path / "x.png"
    tmp_image.write_bytes(b"PNGDATA")
    respx_mock.post(f"/canvases/{CANVAS_ID}/images").mock(
        return_value=httpx.Response(
            201, json={"id": "img1", "widget_type": "Image"}
        )
    )
    tool = ImageCreateTool(client)
    result = await tool.execute(
        canvas_id=CANVAS_ID, file_path=str(tmp_image), x=10, y=20
    )
    assert result["image_id"] == "img1"
    assert result["position"] == {"x": 10.0, "y": 20.0}


async def test_image_create_rejects_missing_file(
    client: Client, tmp_path: Path
) -> None:
    tool = ImageCreateTool(client)
    with pytest.raises(MCPToolValidationError):
        await tool.execute(
            canvas_id=CANVAS_ID, file_path=str(tmp_path / "nope.png")
        )


async def test_asset_list_tools(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    respx_mock.get(f"/canvases/{CANVAS_ID}/images").mock(
        return_value=httpx.Response(200, json=[])
    )
    respx_mock.get(f"/canvases/{CANVAS_ID}/videos").mock(
        return_value=httpx.Response(200, json=[])
    )
    respx_mock.get(f"/canvases/{CANVAS_ID}/pdfs").mock(
        return_value=httpx.Response(200, json=[])
    )
    assert (
        await ImageListTool(client).execute(canvas_id=CANVAS_ID)
    )["count"] == 0
    assert (
        await VideoListTool(client).execute(canvas_id=CANVAS_ID)
    )["count"] == 0
    assert (
        await PdfListTool(client).execute(canvas_id=CANVAS_ID)
    )["count"] == 0


# ---- connectors ----------------------------------------------------------


async def test_connector_create_builds_endpoints(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    route = respx_mock.post(f"/canvases/{CANVAS_ID}/connectors").mock(
        return_value=httpx.Response(
            201, json={"id": "c1", "widget_type": "Connector"}
        )
    )
    tool = ConnectorCreateTool(client)
    result = await tool.execute(
        canvas_id=CANVAS_ID,
        source_id="src",
        target_id="dst",
        color="#ff0000",
    )
    body = route.calls.last.request.content.decode()
    assert '"src"' in body and '"dst"' in body
    assert '"line_color":"#ff0000"' in body
    assert result["connector_id"] == "c1"


async def test_connector_list_and_delete(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    respx_mock.get(f"/canvases/{CANVAS_ID}/connectors").mock(
        return_value=httpx.Response(200, json=[])
    )
    respx_mock.delete(f"/canvases/{CANVAS_ID}/connectors/c1").mock(
        return_value=httpx.Response(204)
    )
    list_result = await ConnectorListTool(client).execute(canvas_id=CANVAS_ID)
    delete_result = await ConnectorDeleteTool(client).execute(
        canvas_id=CANVAS_ID, connector_id="c1"
    )
    assert list_result["count"] == 0
    assert delete_result["status"] == "deleted"


# ---- error translation ---------------------------------------------------


async def test_api_error_is_translated(
    client: Client, respx_mock: respx.MockRouter
) -> None:
    respx_mock.get(f"/canvases/{CANVAS_ID}/notes").mock(
        return_value=httpx.Response(500, json={"message": "boom"})
    )
    with pytest.raises(MCPToolExecutionError) as exc:
        await NoteListTool(client).execute(canvas_id=CANVAS_ID)
    assert "List notes" in exc.value.message
