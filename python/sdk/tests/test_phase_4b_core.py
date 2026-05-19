"""Phase 4b §4.2 #1-#15 tests for core-SDK additions."""

from __future__ import annotations

import json

import pytest
import respx
from httpx import Response

from canvus_sdk import Client, UnsupportedOperationError, ValidationError

# ---- foundations ----------------------------------------------------------


@pytest.mark.asyncio
async def test_auth_get_current_user_returns_user(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("users/current").mock(
            return_value=Response(
                200,
                json={"id": 17, "email": "x@y.example", "name": "Alice"},
            )
        )
        user = await client.auth.get_current_user()
    assert user.id == 17


def test_validation_error_carries_issues() -> None:
    err = ValidationError(
        "bad payload",
        issues=[{"path": "name", "message": "required"}],
    )
    assert err.issues == [{"path": "name", "message": "required"}]
    # mutate result; the error's copy should be unaffected
    err.issues[0]["message"] = "changed"
    assert err.issues[0]["message"] == "changed"  # self mutation works
    err2 = ValidationError("again", issues=err.issues)
    err2.issues.clear()
    assert err.issues  # original untouched


# ---- generic widget CRUD (Phase 4b §4.2 #1-#4) ----------------------------


@pytest.mark.asyncio
async def test_create_any_dispatches_to_notes_path(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        route = mock.post("canvases/c1/notes").mock(
            return_value=Response(200, json={"id": "n1", "widget_type": "Note"})
        )
        result = await client.widgets.create_any(
            "c1",
            {"widget_type": "Note", "text": "hello"},
        )
    assert result["id"] == "n1"
    body = json.loads(route.calls[0].request.content)
    assert body == {"widget_type": "Note", "text": "hello"}


@pytest.mark.asyncio
async def test_create_any_rejects_ip_video(client: Client) -> None:
    with pytest.raises(UnsupportedOperationError):
        await client.widgets.create_any("c1", {"widget_type": "ip_video"})


@pytest.mark.asyncio
async def test_update_any_strips_grid_size_for_tables(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        route = mock.patch("canvases/c1/tables/t1").mock(
            return_value=Response(200, json={"id": "t1", "widget_type": "Table"})
        )
        with pytest.warns(UserWarning, match="grid_size"):
            await client.widgets.update_any(
                "c1",
                "t1",
                {"widget_type": "table", "grid_size": {"columns": 4, "rows": 4}},
            )
    body = json.loads(route.calls[0].request.content)
    assert "grid_size" not in body


@pytest.mark.asyncio
async def test_delete_any_dispatches_to_typed_endpoint(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.delete("canvases/c1/images/img1").mock(return_value=Response(204))
        await client.widgets.delete_any("c1", "img1", "image")


@pytest.mark.asyncio
async def test_patch_parent_id_posts_to_generic_widgets_path(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        route = mock.patch("canvases/c1/widgets/w1").mock(
            return_value=Response(200, json={"id": "w1", "parent_id": "p1"})
        )
        widget = await client.widgets.patch_parent_id("c1", "w1", "p1")
    assert widget.parent_id == "p1"
    body = json.loads(route.calls[0].request.content)
    assert body == {"parent_id": "p1"}


# ---- trash (Phase 4b §4.2 #5-#6) ------------------------------------------


@pytest.mark.asyncio
async def test_canvases_trash_uses_user_id_from_current(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("users/current").mock(
            return_value=Response(
                200, json={"id": 42, "email": "x@y", "name": "x"}
            )
        )
        route = mock.post("canvases/c1/move").mock(
            return_value=Response(
                200,
                json={"id": "c1", "name": "x", "folder_id": "trash.42",
                      "asset_size": 0, "in_trash": True},
            )
        )
        canvas = await client.canvases.trash("c1")
    assert canvas.folder_id == "trash.42"
    body = json.loads(route.calls[0].request.content)
    assert body == {"folder_id": "trash.42"}


@pytest.mark.asyncio
async def test_folders_trash_uses_user_id(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("users/current").mock(
            return_value=Response(
                200, json={"id": 7, "email": "x@y", "name": "x"}
            )
        )
        mock.post("canvas-folders/f1/move").mock(
            return_value=Response(
                200, json={"id": "f1", "name": "x", "folder_id": "trash.7"}
            )
        )
        folder = await client.folders.trash("f1")
    assert folder.folder_id == "trash.7"


# ---- color preset decomposition (Phase 4b §4.2 #8) ------------------------


@pytest.mark.asyncio
async def test_list_color_presets_returns_sorted_names(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("canvases/c1/color-presets").mock(
            return_value=Response(200, json={"blue": "#0000ffff", "red": "#ff0000ff"})
        )
        names = await client.canvases.list_color_presets("c1")
    assert names == ["blue", "red"]


@pytest.mark.asyncio
async def test_create_color_preset_writes_merged_payload(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("canvases/c1/color-presets").mock(
            return_value=Response(200, json={"red": {"value": "FF0000FF"}})
        )
        route = mock.patch("canvases/c1/color-presets").mock(
            return_value=Response(
                200,
                json={"red": {"value": "FF0000FF"}, "blue": {"value": "0000FFFF"}},
            )
        )
        result = await client.canvases.create_color_preset(
            "c1", "blue", {"value": "0000FFFF"}
        )
    assert result == {"value": "0000FFFF"}
    body = json.loads(route.calls[0].request.content)
    assert body == {"red": {"value": "FF0000FF"}, "blue": {"value": "0000FFFF"}}


@pytest.mark.asyncio
async def test_delete_color_preset_drops_key(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("canvases/c1/color-presets").mock(
            return_value=Response(200, json={"red": "x", "blue": "y"})
        )
        route = mock.patch("canvases/c1/color-presets").mock(
            return_value=Response(200, json={"blue": "y"})
        )
        await client.canvases.delete_color_preset("c1", "red")
    body = json.loads(route.calls[0].request.content)
    assert body == {"blue": "y"}


# ---- server convenience (Phase 4b §4.2 #9-#12) ----------------------------


@pytest.mark.asyncio
async def test_get_config_raw_flattens_nested_object(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("server-config").mock(
            return_value=Response(
                200,
                json={
                    "server_name": "example",
                    "authentication": {"password": {"enabled": True}},
                },
            )
        )
        elems = await client.server.get_config_raw()
    flattened = {e["setting-key"]: e["setting-value"] for e in elems}
    assert flattened == {
        "server_name": "example",
        "authentication.password.enabled": True,
    }


@pytest.mark.asyncio
async def test_set_video_output_source_by_index_patches_path(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        route = mock.patch("clients/cli1/video-outputs/0").mock(
            return_value=Response(200, json={"id": "0", "source": "src-a"})
        )
        await client.server.set_video_output_source_by_index("cli1", 0, "src-a")
    body = json.loads(route.calls[0].request.content)
    assert body == {"source": "src-a"}


@pytest.mark.asyncio
async def test_toggle_workspace_info_panel_inverts(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("clients/cli1/workspaces/ws1").mock(
            return_value=Response(200, json={"id": "ws1", "info_panel_visible": True})
        )
        route = mock.patch("clients/cli1/workspaces/ws1").mock(
            return_value=Response(200, json={"id": "ws1", "info_panel_visible": False})
        )
        ws = await client.server.toggle_workspace_info_panel("cli1", "ws1")
    assert ws.info_panel_visible is False
    body = json.loads(route.calls[0].request.content)
    assert body == {"info_panel_visible": False}


@pytest.mark.asyncio
async def test_set_workspace_viewport_from_widget(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("canvases/c1/widgets/w1").mock(
            return_value=Response(
                200,
                json={
                    "id": "w1",
                    "location": {"x": 100, "y": 200},
                    "size": {"width": 50, "height": 50},
                },
            )
        )
        route = mock.patch("clients/cli1/workspaces/ws1").mock(
            return_value=Response(200, json={"id": "ws1"})
        )
        await client.server.set_workspace_viewport(
            "cli1",
            "ws1",
            widget_canvas_id="c1",
            widget_id="w1",
            margin=10.0,
        )
    body = json.loads(route.calls[0].request.content)
    assert body == {
        "view_rectangle": {"x": 90.0, "y": 190.0, "width": 70.0, "height": 70.0}
    }


@pytest.mark.asyncio
async def test_set_workspace_viewport_requires_args(client: Client) -> None:
    with pytest.raises(ValidationError):
        await client.server.set_workspace_viewport("cli1", "ws1")


# ---- subscribe (Phase 4b §4.2 #13) ---------------------------------------


@pytest.mark.asyncio
async def test_canvases_subscribe_yields_typed_models(client: Client) -> None:
    ndjson = b'{"id":"c1","name":"first","asset_size":0}\n{"id":"c2","name":"second","asset_size":0}\n'
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("canvases").mock(
            return_value=Response(
                200, content=ndjson, headers={"content-type": "application/x-ndjson"}
            )
        )
        canvases = [c async for c in client.canvases.subscribe()]
    assert [c.id for c in canvases] == ["c1", "c2"]
