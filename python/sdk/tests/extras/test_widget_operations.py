"""Phase 4b §4.2 #18: widget_operations tests."""

from __future__ import annotations

from canvus_sdk.extras.geometry import Rectangle
from canvus_sdk.extras.widget_operations import (
    BatchWidgetOperations,
    SpatialTolerance,
    WidgetZoneManager,
    calculate_widget_density,
    find_widget_clusters,
)


class _W:
    def __init__(self, x: float, y: float, w: float, h: float, wid: str) -> None:
        self.location = {"x": x, "y": y}
        self.size = {"width": w, "height": h}
        self.id = wid


def test_create_zone_from_widgets_adds_padding() -> None:
    mgr = WidgetZoneManager()
    widgets = [_W(0, 0, 100, 100, "a"), _W(200, 50, 50, 50, "b")]
    zone = mgr.create_zone_from_widgets(widgets, "test", padding=10.0)
    assert zone.name == "test"
    assert zone.location == {"x": -10.0, "y": -10.0}
    assert zone.size == {"width": 270.0, "height": 120.0}


def test_widgets_in_zone_and_touching_zone() -> None:
    mgr = WidgetZoneManager(SpatialTolerance())
    widgets = [
        _W(10, 10, 20, 20, "inside"),
        _W(200, 200, 20, 20, "outside"),
        _W(95, 95, 20, 20, "overlap"),
    ]
    zone = mgr.create_zone_from_widgets([_W(0, 0, 100, 100, "ref")], "zone", padding=0.0)
    inside = mgr.widgets_in_zone(widgets, zone)
    assert [w.id for w in inside if isinstance(getattr(w, "id", None), str)] == ["inside"]
    touching = mgr.widgets_touching_zone(widgets, zone)
    touching_ids = {w.id for w in touching if isinstance(getattr(w, "id", None), str)}
    assert "inside" in touching_ids and "overlap" in touching_ids


def test_batch_move_widgets_generates_payloads() -> None:
    ops = BatchWidgetOperations().move_widgets(
        [_W(0, 0, 10, 10, "a"), _W(50, 50, 10, 10, "b")],
        offset_x=5.0,
        offset_y=-3.0,
    )
    assert len(ops) == 2
    assert ops[0]["widget_id"] == "a"
    assert ops[0]["payload"]["location"] == {"x": 5.0, "y": -3.0}


def test_batch_resize_scales_size() -> None:
    ops = BatchWidgetOperations().resize_widgets(
        [_W(0, 0, 100, 50, "a")], scale_factor=2.0
    )
    assert ops[0]["payload"]["size"] == {"width": 200.0, "height": 100.0}


def test_find_widget_clusters_min_size() -> None:
    # widgets close together → cluster
    widgets = [
        _W(0, 0, 10, 10, "a"),
        _W(5, 5, 10, 10, "b"),
        _W(500, 500, 10, 10, "c"),
    ]
    clusters = find_widget_clusters(widgets, min_cluster_size=2, tolerance=20.0)
    assert len(clusters) == 1
    assert {w.id for w in clusters[0]} == {"a", "b"}


def test_calculate_widget_density() -> None:
    area = Rectangle(0, 0, 100, 100)
    widgets = [_W(0, 0, 10, 10, "a"), _W(50, 50, 10, 10, "b")]
    density = calculate_widget_density(widgets, area)
    assert density == 2 / (100 * 100)
