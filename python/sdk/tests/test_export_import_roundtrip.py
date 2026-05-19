"""Phase 4d Round C5: schema-only export/import roundtrip test.

Builds a synthetic export bundle (canonical ``{widgets, assets, region}``
schema with flat sibling asset files) on disk and runs it back through the
Python ``WidgetImporter``. Verifies that the cross-runtime wire shape is
honoured: widget shapes flow through to the create call, and asset file
mapping is preserved (importer reads the correct sibling file).

No live server is touched — :mod:`respx` mocks the transport.
"""

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
    WidgetImporter,
)


@pytest.mark.asyncio
async def test_roundtrip_note_and_image_preserves_schema(
    client: Client, tmp_path: Path
) -> None:
    """Build a bundle on disk, reimport, verify widget shapes + asset map."""
    folder = tmp_path / "bundle"
    folder.mkdir()

    # Step 1: write the canonical bundle — flat sibling image file + manifest.
    image_bytes = bytes([0xFF, 0xD8, 0xFF, 0xE0, 0x10, 0x4A, 0x46, 0x49, 0x46])  # JPEG magic
    (folder / "image_img-1.jpg").write_bytes(image_bytes)
    manifest = {
        "widgets": [
            {
                "id": "note-1",
                "widget_type": "note",
                "parent_id": "src-canvas",
                "location": {"x": 10.0, "y": 20.0},
                "size": {"width": 100.0, "height": 50.0},
                "text": "roundtrip",
            },
            {
                "id": "img-1",
                "widget_type": "image",
                "parent_id": "src-canvas",
                "location": {"x": 50.0, "y": 60.0},
                "size": {"width": 200.0, "height": 200.0},
            },
        ],
        "assets": {"img-1": "image_img-1.jpg"},
        "region": {"x": 0.0, "y": 0.0, "width": 1000.0, "height": 1000.0},
    }
    (folder / "export.json").write_text(json.dumps(manifest, indent=2, sort_keys=True))

    # Step 2: mock the create endpoints (note POST + image multipart upload).
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        note_route = mock.post("canvases/dst-canvas/notes").mock(
            return_value=Response(
                200,
                json={
                    "id": "note-1-new",
                    "widget_type": "Note",
                    "parent_id": "dst-canvas",
                    "location": {"x": 50.0, "y": 50.0},
                    "size": {"width": 200.0, "height": 100.0},
                    "text": "roundtrip",
                },
            )
        )
        image_route = mock.post("canvases/dst-canvas/images").mock(
            return_value=Response(
                200,
                json={
                    "id": "img-1-new",
                    "widget_type": "Image",
                    "parent_id": "dst-canvas",
                    "location": {"x": 250.0, "y": 300.0},
                    "size": {"width": 400.0, "height": 400.0},
                },
            )
        )

        # Step 3: drive the importer. Target region is 2x source so coordinates
        # scale by 2.
        new_ids = await WidgetImporter(client).import_widgets_from_folder(
            "dst-canvas",
            folder,
            Rectangle(0.0, 0.0, 2000.0, 2000.0),
        )

    # Step 4: roundtrip assertions.
    assert new_ids == ["note-1-new", "img-1-new"]

    # Note: schema preserved (widget_type, text flowed through), spatial fields
    # transformed by the importer (scale_x=scale_y=2, dx=dy=0).
    note_body = json.loads(note_route.calls[0].request.content)
    assert note_body["widget_type"] == "note"
    assert note_body["text"] == "roundtrip"
    assert note_body["location"] == {"x": 20.0, "y": 40.0}
    assert note_body["size"] == {"width": 200.0, "height": 100.0}
    # Importer strips server-generated keys before POST.
    assert "id" not in note_body
    assert "state" not in note_body
    assert "created_at" not in note_body
    assert "modified_at" not in note_body

    # Image: multipart upload — the asset file mapping resolved correctly so
    # the on-disk image bytes appear in the multipart body.
    img_req = image_route.calls[0].request
    assert img_req.headers["content-type"].startswith("multipart/form-data")
    body_bytes = img_req.content
    assert image_bytes in body_bytes, "asset file mapping failed: image bytes missing from upload"
    # The metadata block embeds the transformed location/size and the
    # importer's filename convention.
    assert b'"widget_type": "image"' in body_bytes
    assert b"imported_image_img-1.jpg" in body_bytes


def test_exported_widget_set_roundtrip_via_dict(tmp_path: Path) -> None:
    """Wire-shape roundtrip: dict -> ExportedWidgetSet -> dict is stable."""
    folder = tmp_path / "manifest-only"
    folder.mkdir()
    data: dict[str, object] = {
        "widgets": [
            {
                "id": "w1",
                "widget_type": "Note",
                "location": {"x": 1.5, "y": 2.5},
                "size": {"width": 30.0, "height": 40.0},
                "text": "schema-check",
            }
        ],
        "assets": {"a1": "image_a1.jpg"},
        "region": {"x": 0.0, "y": 0.0, "width": 500.0, "height": 500.0},
    }
    (folder / "export.json").write_text(json.dumps(data))

    loaded = ExportedWidgetSet.load(folder)
    out = loaded.to_dict()

    # Top-level keys are exactly {widgets, assets, region}.
    assert set(out.keys()) == {"widgets", "assets", "region"}
    assert out["assets"] == {"a1": "image_a1.jpg"}
    assert out["region"] == {"x": 0.0, "y": 0.0, "width": 500.0, "height": 500.0}
    assert isinstance(out["widgets"], list)
    assert len(out["widgets"]) == 1
    w = out["widgets"][0]
    assert isinstance(w, dict)
    assert w["id"] == "w1"
    assert w["widget_type"] == "Note"
    assert w["location"] == {"x": 1.5, "y": 2.5}


def test_canonical_asset_filename_convention(tmp_path: Path) -> None:
    """Cross-runtime: asset filenames follow ``<type>_<widget_id>.<ext>``."""
    folder = tmp_path / "ex"
    folder.mkdir()
    (folder / "image_abc.jpg").write_bytes(b"jpg")
    (folder / "pdf_def.pdf").write_bytes(b"pdf")
    (folder / "video_ghi.mp4").write_bytes(b"mp4")
    manifest = {
        "widgets": [
            {"id": "abc", "widget_type": "image"},
            {"id": "def", "widget_type": "pdf"},
            {"id": "ghi", "widget_type": "video"},
        ],
        "assets": {
            "abc": "image_abc.jpg",
            "def": "pdf_def.pdf",
            "ghi": "video_ghi.mp4",
        },
        "region": {"x": 0, "y": 0, "width": 10, "height": 10},
    }
    (folder / "export.json").write_text(json.dumps(manifest))

    loaded = ExportedWidgetSet.load(folder)
    assert loaded.asset_bytes("abc") == b"jpg"
    assert loaded.asset_bytes("def") == b"pdf"
    assert loaded.asset_bytes("ghi") == b"mp4"
