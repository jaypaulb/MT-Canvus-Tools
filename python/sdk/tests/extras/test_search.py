"""Phase 4b §4.2 #19: search tests (mocked via respx)."""

from __future__ import annotations

import pytest
import respx
from httpx import Response

from canvus_sdk import Client
from canvus_sdk.extras import CrossCanvasSearch, find_widgets_by_text, find_widgets_by_type
from canvus_sdk.extras.geometry import Rectangle
from canvus_sdk.extras.search import find_widgets_across_canvases, find_widgets_in_area


@pytest.mark.asyncio
async def test_find_widgets_by_type_filters_results(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("canvases").mock(
            return_value=Response(
                200,
                json=[
                    {"id": "c1", "name": "First", "asset_size": 0},
                ],
            )
        )
        mock.get("canvases/c1/widgets").mock(
            return_value=Response(
                200,
                json=[
                    {"id": "w1", "widget_type": "Note", "state": "normal",
                     "location": {"x": 0, "y": 0}, "size": {"width": 10, "height": 10}},
                    {"id": "w2", "widget_type": "Image", "state": "normal",
                     "location": {"x": 0, "y": 0}, "size": {"width": 10, "height": 10}},
                ],
            )
        )
        results = await find_widgets_by_type(client, "note")
    assert len(results) == 1
    assert results[0].widget_id == "w1"


@pytest.mark.asyncio
async def test_find_widgets_in_area_excludes_outside(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("canvases").mock(
            return_value=Response(
                200,
                json=[{"id": "c1", "name": "C", "asset_size": 0}],
            )
        )
        mock.get("canvases/c1/widgets").mock(
            return_value=Response(
                200,
                json=[
                    {"id": "in", "widget_type": "Note",
                     "location": {"x": 10, "y": 10}, "size": {"width": 5, "height": 5}},
                    {"id": "out", "widget_type": "Note",
                     "location": {"x": 500, "y": 500}, "size": {"width": 5, "height": 5}},
                ],
            )
        )
        results = await find_widgets_in_area(client, Rectangle(0, 0, 100, 100))
    assert {r.widget_id for r in results} == {"in"}


@pytest.mark.asyncio
async def test_find_widgets_by_text_wildcards(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("canvases").mock(
            return_value=Response(
                200, json=[{"id": "c1", "name": "C", "asset_size": 0}]
            )
        )
        mock.get("canvases/c1/widgets").mock(
            return_value=Response(
                200,
                json=[
                    {"id": "w1", "widget_type": "Note", "text": "hello world"},
                    {"id": "w2", "widget_type": "Note", "text": "goodbye"},
                ],
            )
        )
        results = await find_widgets_by_text(client, "hello")
    assert [r.widget_id for r in results] == ["w1"]


@pytest.mark.asyncio
async def test_cross_canvas_search_skips_deleted(client: Client) -> None:
    """With no criteria, results include all non-deleted widgets."""
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("canvases").mock(
            return_value=Response(
                200, json=[{"id": "c1", "name": "C", "asset_size": 0}]
            )
        )
        mock.get("canvases/c1/widgets").mock(
            return_value=Response(
                200,
                json=[
                    {"id": "alive", "widget_type": "Note", "state": "normal"},
                    {"id": "dead", "widget_type": "Note", "state": "deleted"},
                ],
            )
        )
        # Empty dict criteria → no per-field match; only the deleted filter applies.
        results = await find_widgets_across_canvases(client, {})
    assert {r.widget_id for r in results} == {"alive"}


def test_cross_canvas_search_query_parser_handles_strings() -> None:
    parsed = CrossCanvasSearch._parse_query("hello")
    assert parsed == {"text": "*hello*"}
    parsed = CrossCanvasSearch._parse_query({"foo": 1})
    assert parsed == {"foo": 1}
