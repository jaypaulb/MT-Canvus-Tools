"""Phase 4b §4.2 #20: export/import tests."""

from __future__ import annotations

import json
from pathlib import Path

import pytest
import respx
from httpx import Response

from canvus_sdk import Client
from canvus_sdk.extras import (
    ExportedWidgetSet,
    Rectangle,
    WidgetExporter,
    WidgetImporter,
)


@pytest.mark.asyncio
async def test_export_writes_export_json_and_image_asset(
    client: Client, tmp_path: Path
) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("canvases/c1/widgets/w1").mock(
            return_value=Response(
                200,
                json={
                    "id": "w1",
                    "widget_type": "Image",
                    "location": {"x": 0.0, "y": 0.0},
                    "size": {"width": 100.0, "height": 100.0},
                },
            )
        )
        mock.get("canvases/c1/images/w1/download").mock(
            return_value=Response(200, content=b"FAKE_IMAGE_BYTES")
        )
        folder = await WidgetExporter(client).export_widgets_to_folder(
            "c1",
            ["w1"],
            Rectangle(0, 0, 100, 100),
            base_folder=tmp_path / "export",
        )
    assert (folder / "export.json").exists()
    assert (folder / "image_w1.jpg").read_bytes() == b"FAKE_IMAGE_BYTES"
    manifest = json.loads((folder / "export.json").read_text())
    assert manifest["assets"] == {"w1": "image_w1.jpg"}
    assert manifest["region"] == {"x": 0.0, "y": 0.0, "width": 100.0, "height": 100.0}


def test_exported_widget_set_roundtrip(tmp_path: Path) -> None:
    folder = tmp_path / "ex"
    folder.mkdir()
    data = {
        "widgets": [{"id": "w1", "widget_type": "Note"}],
        "assets": {},
        "region": {"x": 0, "y": 0, "width": 10, "height": 10},
    }
    (folder / "export.json").write_text(json.dumps(data))
    loaded = ExportedWidgetSet.load(folder)
    assert loaded.region is not None and loaded.region.width == 10.0
    assert loaded.widgets[0]["id"] == "w1"
    serialised = loaded.to_dict()
    assert serialised["region"] == {"x": 0.0, "y": 0.0, "width": 10.0, "height": 10.0}


@pytest.mark.asyncio
async def test_import_note_uses_create_any(client: Client, tmp_path: Path) -> None:
    folder = tmp_path / "ex"
    folder.mkdir()
    data = {
        "widgets": [
            {
                "id": "old",
                "widget_type": "note",
                "location": {"x": 0.0, "y": 0.0},
                "size": {"width": 10.0, "height": 10.0},
                "text": "hello",
            }
        ],
        "assets": {},
        "region": {"x": 0.0, "y": 0.0, "width": 100.0, "height": 100.0},
    }
    (folder / "export.json").write_text(json.dumps(data))

    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        route = mock.post("canvases/c1/notes").mock(
            return_value=Response(200, json={"id": "new", "widget_type": "Note"})
        )
        new_ids = await WidgetImporter(client).import_widgets_from_folder(
            "c1", folder, Rectangle(50, 50, 200, 200)
        )
    assert new_ids == ["new"]
    body = json.loads(route.calls[0].request.content)
    # Scaled location: x=0*2+50, y=0*2+50
    assert body["location"] == {"x": 50.0, "y": 50.0}
    assert body["size"] == {"width": 20.0, "height": 20.0}
    # Server-generated keys should have been stripped before POST.
    assert "id" not in body
