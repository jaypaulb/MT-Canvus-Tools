"""Phase 4b §4.2 #16: geometry port.

Spatial utilities for working with widget positions, sizes, and rectangular
relationships in canvas pixel-space.

Ported from ``CanvusPythonAPI/canvus_api/geometry.py`` with the following
type-system upgrades for the new SDK:

- Dataclasses are typed (``float`` not ``int | float``) and ``frozen=True``.
- All return / parameter types are explicit; no implicit ``Optional``.
- The widget-shaped helpers accept any object exposing ``.location`` and
  ``.size`` mappings (i.e. duck-typed across :mod:`canvus_sdk.models`
  widget classes and raw dicts).

The Canvus pixel convention is preserved: ``y`` grows DOWN (top-left origin).
"""

from __future__ import annotations

import math
from collections.abc import Iterable, Mapping, Sequence
from dataclasses import dataclass
from typing import Any, Protocol, runtime_checkable

__all__ = [
    "Point",
    "Rectangle",
    "Size",
    "WidgetLike",
    "contains",
    "distance_between_widgets",
    "find_widgets_containing_point",
    "find_widgets_in_area",
    "get_canvas_bounds",
    "get_intersection",
    "get_union",
    "get_widget_intersection",
    "get_widget_union",
    "intersects",
    "touches",
    "widget_bounding_box",
    "widget_contains",
    "widgets_intersect",
    "widgets_touch",
]


# ---- primitive types -------------------------------------------------------


@dataclass(frozen=True, slots=True)
class Point:
    """2D point in canvas pixel-space."""

    x: float
    y: float


@dataclass(frozen=True, slots=True)
class Size:
    """2D size in canvas pixel-space. Width / height must be non-negative."""

    width: float
    height: float

    def __post_init__(self) -> None:
        if self.width < 0 or self.height < 0:
            raise ValueError("Size dimensions cannot be negative")


@dataclass(frozen=True, slots=True)
class Rectangle:
    """Axis-aligned rectangle defined by top-left position and size."""

    x: float
    y: float
    width: float
    height: float

    def __post_init__(self) -> None:
        if self.width < 0 or self.height < 0:
            raise ValueError("Rectangle dimensions cannot be negative")

    @property
    def left(self) -> float:
        return self.x

    @property
    def right(self) -> float:
        return self.x + self.width

    @property
    def top(self) -> float:
        return self.y

    @property
    def bottom(self) -> float:
        return self.y + self.height

    @property
    def center(self) -> Point:
        return Point(self.x + self.width / 2.0, self.y + self.height / 2.0)

    @property
    def size(self) -> Size:
        return Size(self.width, self.height)

    @property
    def position(self) -> Point:
        return Point(self.x, self.y)


# ---- widget protocol -------------------------------------------------------


@runtime_checkable
class WidgetLike(Protocol):
    """Structural protocol: anything with a location and a size mapping."""

    location: Mapping[str, float] | object
    size: Mapping[str, float] | object


def _to_xy(obj: Any) -> tuple[float, float]:
    if isinstance(obj, Mapping):
        return float(obj.get("x", 0.0)), float(obj.get("y", 0.0))
    if hasattr(obj, "x") and hasattr(obj, "y"):
        return float(obj.x), float(obj.y)
    raise ValueError(f"expected location-like {{x,y}}, got {type(obj).__name__}")


def _to_wh(obj: Any) -> tuple[float, float]:
    if isinstance(obj, Mapping):
        return float(obj.get("width", 0.0)), float(obj.get("height", 0.0))
    if hasattr(obj, "width") and hasattr(obj, "height"):
        return float(obj.width), float(obj.height)
    raise ValueError(f"expected size-like {{width,height}}, got {type(obj).__name__}")


# ---- rectangle operations --------------------------------------------------


def contains(outer: Rectangle, inner: Rectangle) -> bool:
    """True iff ``outer`` fully contains ``inner`` (edges may touch)."""
    return (
        outer.left <= inner.left
        and outer.right >= inner.right
        and outer.top <= inner.top
        and outer.bottom >= inner.bottom
    )


def touches(a: Rectangle, b: Rectangle) -> bool:
    """True iff rectangles touch (shared edge) or overlap."""
    return not (a.right < b.left or a.left > b.right or a.bottom < b.top or a.top > b.bottom)


def intersects(a: Rectangle, b: Rectangle) -> bool:
    """True iff rectangles have overlapping area (touching edges only is False)."""
    return not (
        a.right <= b.left or a.left >= b.right or a.bottom <= b.top or a.top >= b.bottom
    )


def get_intersection(a: Rectangle, b: Rectangle) -> Rectangle | None:
    """Return the intersection rectangle, or ``None`` if disjoint."""
    if not intersects(a, b):
        return None
    left = max(a.left, b.left)
    top = max(a.top, b.top)
    right = min(a.right, b.right)
    bottom = min(a.bottom, b.bottom)
    return Rectangle(left, top, right - left, bottom - top)


def get_union(a: Rectangle, b: Rectangle) -> Rectangle:
    """Return the minimum bounding rectangle of ``a`` and ``b``."""
    left = min(a.left, b.left)
    top = min(a.top, b.top)
    right = max(a.right, b.right)
    bottom = max(a.bottom, b.bottom)
    return Rectangle(left, top, right - left, bottom - top)


# ---- widget-aware operations -----------------------------------------------


def widget_bounding_box(widget: object) -> Rectangle:
    """Return the axis-aligned bounding rectangle for a widget.

    Connectors are handled specially when both ``src`` and ``dst`` exist with
    ``rel_location`` mappings.
    """
    src = getattr(widget, "src", None)
    dst = getattr(widget, "dst", None)
    if src is not None and dst is not None and hasattr(src, "rel_location"):
        src_x, src_y = _to_xy(src.rel_location)
        dst_x, dst_y = _to_xy(dst.rel_location)
        min_x, max_x = min(src_x, dst_x), max(src_x, dst_x)
        min_y, max_y = min(src_y, dst_y), max(src_y, dst_y)
        padding = 10.0
        return Rectangle(
            x=min_x - padding,
            y=min_y - padding,
            width=(max_x - min_x) + 2.0 * padding,
            height=(max_y - min_y) + 2.0 * padding,
        )
    location = getattr(widget, "location", None)
    size = getattr(widget, "size", None)
    if location is None or size is None:
        raise ValueError(
            f"widget {type(widget).__name__!r} lacks location/size; cannot compute "
            "bounding box",
        )
    x, y = _to_xy(location)
    w, h = _to_wh(size)
    return Rectangle(x=x, y=y, width=w, height=h)


def widget_contains(outer: object, inner: object) -> bool:
    """True iff ``outer``'s bounding box contains ``inner``'s."""
    return contains(widget_bounding_box(outer), widget_bounding_box(inner))


def widgets_touch(a: object, b: object) -> bool:
    """True iff two widgets' bounding boxes touch or overlap."""
    return touches(widget_bounding_box(a), widget_bounding_box(b))


def widgets_intersect(a: object, b: object) -> bool:
    """True iff two widgets' bounding boxes have overlapping area."""
    return intersects(widget_bounding_box(a), widget_bounding_box(b))


def get_widget_intersection(a: object, b: object) -> Rectangle | None:
    """Return the intersection rectangle of two widgets, or ``None``."""
    return get_intersection(widget_bounding_box(a), widget_bounding_box(b))


def get_widget_union(a: object, b: object) -> Rectangle:
    """Return the minimum bounding rectangle of two widgets."""
    return get_union(widget_bounding_box(a), widget_bounding_box(b))


def distance_between_widgets(a: object, b: object) -> float:
    """Return the cartesian gap distance between two widgets' bounding boxes.

    Returns ``0.0`` if the widgets overlap. When the rectangles are separated
    on both axes, returns the euclidean distance between the nearest corners
    (``math.hypot(dx, dy)``). When separated on a single axis only, returns
    that axis's gap directly. Symmetric across Go / Python / TS SDK extras.

    Note: this deliberately differs from the legacy ``CanvusPythonAPI`` helper,
    which returned ``min(dx, dy)`` in the both-positive case — the legacy
    answer was a single-axis projection, not a cartesian distance.
    """
    ra = widget_bounding_box(a)
    rb = widget_bounding_box(b)
    if intersects(ra, rb):
        return 0.0
    dx = max(0.0, max(ra.left - rb.right, rb.left - ra.right))
    dy = max(0.0, max(ra.top - rb.bottom, rb.top - ra.bottom))
    if dx > 0.0 and dy > 0.0:
        return math.hypot(dx, dy)
    return max(dx, dy)


def find_widgets_in_area(
    widgets: Iterable[object],
    area: Rectangle,
) -> list[object]:
    """Return every widget whose bounding box intersects ``area``."""
    return [w for w in widgets if intersects(widget_bounding_box(w), area)]


def find_widgets_containing_point(
    widgets: Iterable[object],
    point: Point,
) -> list[object]:
    """Return every widget whose bounding box contains ``point``."""
    out: list[object] = []
    for widget in widgets:
        box = widget_bounding_box(widget)
        if box.left <= point.x <= box.right and box.top <= point.y <= box.bottom:
            out.append(widget)
    return out


def get_canvas_bounds(widgets: Sequence[object]) -> Rectangle | None:
    """Return the minimum bounding rectangle for a non-empty widget list."""
    if not widgets:
        return None
    bounds = widget_bounding_box(widgets[0])
    for widget in widgets[1:]:
        bounds = get_union(bounds, widget_bounding_box(widget))
    return bounds
