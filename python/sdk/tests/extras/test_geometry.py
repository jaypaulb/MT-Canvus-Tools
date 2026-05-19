"""Phase 4b §4.2 #16: geometry tests."""

from __future__ import annotations

import math

import pytest

from canvus_sdk.extras.geometry import (
    Point,
    Rectangle,
    Size,
    contains,
    distance_between_widgets,
    find_widgets_containing_point,
    find_widgets_in_area,
    get_canvas_bounds,
    get_intersection,
    get_union,
    intersects,
    touches,
    widget_bounding_box,
    widget_contains,
    widgets_intersect,
    widgets_touch,
)


def test_rectangle_negative_size_rejected() -> None:
    with pytest.raises(ValueError, match="dimensions cannot be negative"):
        Rectangle(0.0, 0.0, -1.0, 10.0)


def test_size_negative_rejected() -> None:
    with pytest.raises(ValueError):
        Size(-1.0, 10.0)


def test_contains_and_touches_and_intersects() -> None:
    outer = Rectangle(0.0, 0.0, 100.0, 100.0)
    inner = Rectangle(10.0, 10.0, 80.0, 80.0)
    edge = Rectangle(100.0, 0.0, 50.0, 100.0)  # touches outer right edge
    far = Rectangle(200.0, 0.0, 10.0, 10.0)

    assert contains(outer, inner)
    assert not contains(inner, outer)
    assert touches(outer, edge)
    assert intersects(outer, inner)
    assert not intersects(outer, edge)  # only touching, no overlap
    assert not touches(outer, far)


def test_get_intersection_and_union() -> None:
    a = Rectangle(0.0, 0.0, 50.0, 50.0)
    b = Rectangle(25.0, 25.0, 50.0, 50.0)
    inter = get_intersection(a, b)
    assert inter == Rectangle(25.0, 25.0, 25.0, 25.0)

    union = get_union(a, b)
    assert union == Rectangle(0.0, 0.0, 75.0, 75.0)

    disjoint = Rectangle(200.0, 200.0, 10.0, 10.0)
    assert get_intersection(a, disjoint) is None


class _W:
    """A bare widget stand-in (no pydantic to keep the test minimal)."""

    def __init__(self, x: float, y: float, w: float, h: float, wid: str = "x") -> None:
        self.location = {"x": x, "y": y}
        self.size = {"width": w, "height": h}
        self.id = wid


def test_widget_bounding_box_and_relations() -> None:
    a = _W(0, 0, 100, 100, "a")
    b = _W(10, 10, 50, 50, "b")
    rect_a = widget_bounding_box(a)
    assert rect_a == Rectangle(0, 0, 100, 100)
    assert widget_contains(a, b)
    assert widgets_touch(a, b)
    assert widgets_intersect(a, b)


def test_distance_zero_when_overlapping() -> None:
    a = _W(0, 0, 100, 100)
    b = _W(50, 50, 100, 100)
    assert distance_between_widgets(a, b) == 0.0


def test_distance_diagonal() -> None:
    a = _W(0, 0, 10, 10)
    b = _W(30, 40, 10, 10)
    # gap (20, 30) → hypot(20, 30)
    assert distance_between_widgets(a, b) == pytest.approx(math.hypot(20.0, 30.0))


def test_find_widgets_in_area_and_containing_point() -> None:
    widgets = [_W(0, 0, 10, 10, "a"), _W(50, 50, 10, 10, "b"), _W(100, 100, 5, 5, "c")]
    area = Rectangle(0, 0, 70, 70)
    in_area = find_widgets_in_area(widgets, area)
    assert {w.id for w in in_area} == {"a", "b"}

    point = Point(5.0, 5.0)
    hits = find_widgets_containing_point(widgets, point)
    assert {w.id for w in hits} == {"a"}


def test_get_canvas_bounds_empty_and_some() -> None:
    assert get_canvas_bounds([]) is None
    widgets = [_W(0, 0, 10, 10), _W(50, 50, 10, 10)]
    bounds = get_canvas_bounds(widgets)
    assert bounds == Rectangle(0, 0, 60, 60)
