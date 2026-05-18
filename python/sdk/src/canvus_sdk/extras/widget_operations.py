"""Phase 4b §4.2 #18: widget_operations port.

Spatial zone management, batch widget edits, and cluster/density analysis.

Ported from ``CanvusPythonAPI/canvus_api/widget_operations.py``. Differences
from legacy:

- ``WidgetZone`` is local to this module — the legacy ``WidgetZone`` lived in
  ``models.py`` and bundled custom Pydantic logic; here it is a frozen
  dataclass with the same field surface (``id``, ``name``, ``description``,
  ``location``, ``size``).
- All ``operation`` dicts produced by :class:`BatchWidgetOperations` are typed
  as ``WidgetOperation`` TypedDicts for static checkers.
- Connector handling uses duck typing instead of an ``isinstance(Connector)``
  check, which would import the model class and create a circular dependency.
"""

from __future__ import annotations

from collections.abc import Iterable, Sequence
from dataclasses import dataclass, field
from typing import Any, TypedDict

from .geometry import (
    Rectangle,
    contains,
    intersects,
    touches,
    widget_bounding_box,
    widget_contains,
    widgets_touch,
)

__all__ = [
    "BatchWidgetOperations",
    "SpatialTolerance",
    "WidgetOperation",
    "WidgetZone",
    "WidgetZoneManager",
    "calculate_widget_density",
    "create_spatial_group",
    "find_widget_clusters",
]


# ---- types -----------------------------------------------------------------


@dataclass(frozen=True, slots=True)
class SpatialTolerance:
    """Tolerance knobs used by zone/cluster operations."""

    position_tolerance: float = 5.0
    size_tolerance: float = 2.0
    overlap_tolerance: float = 1.0
    distance_tolerance: float = 10.0


@dataclass(frozen=True, slots=True)
class WidgetZone:
    """A named bounding region on a canvas. Pure data; no I/O."""

    id: str
    name: str
    location: dict[str, float]
    size: dict[str, float]
    description: str | None = None
    contents: list[Any] = field(default_factory=list)


class WidgetOperation(TypedDict):
    """One entry in the result of a :class:`BatchWidgetOperations` call."""

    widget_id: str
    operation: str
    payload: dict[str, Any]


# ---- zone management -------------------------------------------------------


class WidgetZoneManager:
    """Compute :class:`WidgetZone` objects and query widgets relative to them."""

    def __init__(self, tolerance: SpatialTolerance | None = None) -> None:
        self.tolerance = tolerance or SpatialTolerance()

    def create_zone_from_widgets(
        self,
        widgets: Sequence[object],
        name: str,
        description: str | None = None,
        padding: float = 20.0,
    ) -> WidgetZone:
        """Return a :class:`WidgetZone` enclosing every input widget with ``padding``."""
        if not widgets:
            raise ValueError("Cannot create zone from empty widget list")
        bounds = self._calculate_widget_bounds(widgets)
        return WidgetZone(
            id=f"zone_{len(widgets)}_{hash(name) & 0xFFFFFFFF:08x}",
            name=name,
            description=description,
            location={"x": bounds.x - padding, "y": bounds.y - padding},
            size={
                "width": bounds.width + 2.0 * padding,
                "height": bounds.height + 2.0 * padding,
            },
        )

    def widgets_in_zone(
        self,
        widgets: Iterable[object],
        zone: WidgetZone,
    ) -> list[object]:
        """Return widgets whose bounding box is fully inside ``zone``."""
        zone_rect = _zone_rect(zone)
        out: list[object] = []
        for w in widgets:
            try:
                if contains(zone_rect, widget_bounding_box(w)):
                    out.append(w)
            except (ValueError, AttributeError):
                continue
        return out

    def widgets_touching_zone(
        self,
        widgets: Iterable[object],
        zone: WidgetZone,
    ) -> list[object]:
        """Return widgets whose bounding box touches or overlaps ``zone``."""
        zone_rect = _zone_rect(zone)
        out: list[object] = []
        for w in widgets:
            try:
                if touches(zone_rect, widget_bounding_box(w)):
                    out.append(w)
            except (ValueError, AttributeError):
                continue
        return out

    @staticmethod
    def _calculate_widget_bounds(widgets: Sequence[object]) -> Rectangle:
        min_x = float("inf")
        min_y = float("inf")
        max_x = float("-inf")
        max_y = float("-inf")
        any_valid = False
        for w in widgets:
            try:
                rect = widget_bounding_box(w)
            except (ValueError, AttributeError):
                continue
            any_valid = True
            min_x = min(min_x, rect.x)
            min_y = min(min_y, rect.y)
            max_x = max(max_x, rect.x + rect.width)
            max_y = max(max_y, rect.y + rect.height)
        if not any_valid:
            raise ValueError("No valid widgets found for bounds calculation")
        return Rectangle(x=min_x, y=min_y, width=max_x - min_x, height=max_y - min_y)


def _zone_rect(zone: WidgetZone) -> Rectangle:
    return Rectangle(
        x=float(zone.location.get("x", 0.0)),
        y=float(zone.location.get("y", 0.0)),
        width=float(zone.size.get("width", 0.0)),
        height=float(zone.size.get("height", 0.0)),
    )


# ---- batch operations ------------------------------------------------------


class BatchWidgetOperations:
    """Compose update payloads for moving / resizing many widgets at once.

    These methods generate payload dicts but do NOT call the server. Pair
    with :meth:`canvus_sdk.resources.widgets.WidgetsResource.update_any` (or
    the typed per-type ``update``) to apply.
    """

    def __init__(self, tolerance: SpatialTolerance | None = None) -> None:
        self.tolerance = tolerance or SpatialTolerance()

    def move_widgets(
        self,
        widgets: Iterable[object],
        offset_x: float,
        offset_y: float,
    ) -> list[WidgetOperation]:
        """Return one ``move`` operation per widget, applying the offset."""
        out: list[WidgetOperation] = []
        for widget in widgets:
            widget_id = _widget_id(widget)
            if widget_id is None:
                continue
            if _is_connector(widget):
                src, dst = widget.src, widget.dst  # type: ignore[attr-defined]
                src_loc = _rel(src.rel_location)
                dst_loc = _rel(dst.rel_location)
                out.append(
                    {
                        "widget_id": widget_id,
                        "operation": "move",
                        "payload": {
                            "src": {
                                "rel_location": {
                                    "x": src_loc[0] + offset_x,
                                    "y": src_loc[1] + offset_y,
                                }
                            },
                            "dst": {
                                "rel_location": {
                                    "x": dst_loc[0] + offset_x,
                                    "y": dst_loc[1] + offset_y,
                                }
                            },
                        },
                    }
                )
                continue
            location = getattr(widget, "location", None)
            if not isinstance(location, dict):
                continue
            new_loc = {
                "x": float(location.get("x", 0.0)) + offset_x,
                "y": float(location.get("y", 0.0)) + offset_y,
            }
            out.append(
                {
                    "widget_id": widget_id,
                    "operation": "move",
                    "payload": {"location": new_loc},
                }
            )
        return out

    def resize_widgets(
        self,
        widgets: Iterable[object],
        scale_factor: float,
    ) -> list[WidgetOperation]:
        """Return one ``resize`` operation per widget. Connectors scale ``line_width``."""
        out: list[WidgetOperation] = []
        for widget in widgets:
            widget_id = _widget_id(widget)
            if widget_id is None:
                continue
            if _is_connector(widget):
                line_width = getattr(widget, "line_width", None)
                if line_width is None:
                    continue
                out.append(
                    {
                        "widget_id": widget_id,
                        "operation": "resize",
                        "payload": {"line_width": float(line_width) * scale_factor},
                    }
                )
                continue
            size = getattr(widget, "size", None)
            if not isinstance(size, dict):
                continue
            out.append(
                {
                    "widget_id": widget_id,
                    "operation": "resize",
                    "payload": {
                        "size": {
                            "width": float(size.get("width", 0.0)) * scale_factor,
                            "height": float(size.get("height", 0.0)) * scale_factor,
                        }
                    },
                }
            )
        return out

    def widgets_contain_id(
        self,
        widgets: Sequence[object],
        target_id: str,
    ) -> list[object]:
        """Return widgets that fully contain the widget with ``target_id``."""
        target = next((w for w in widgets if _widget_id(w) == target_id), None)
        if target is None:
            return []
        out: list[object] = []
        for widget in widgets:
            if _widget_id(widget) == target_id:
                continue
            try:
                if widget_contains(widget, target):
                    out.append(widget)
            except (ValueError, AttributeError):
                continue
        return out

    def widgets_touch_id(
        self,
        widgets: Sequence[object],
        target_id: str,
    ) -> list[object]:
        """Return widgets whose bounding box touches the widget with ``target_id``."""
        target = next((w for w in widgets if _widget_id(w) == target_id), None)
        if target is None:
            return []
        out: list[object] = []
        for widget in widgets:
            if _widget_id(widget) == target_id:
                continue
            try:
                if widgets_touch(widget, target):
                    out.append(widget)
            except (ValueError, AttributeError):
                continue
        return out


# ---- grouping / clustering / density --------------------------------------


def create_spatial_group(
    widgets: Sequence[object],
    tolerance: float = 10.0,
) -> list[list[object]]:
    """Greedily group widgets whose bounding-box origins lie within ``tolerance``."""
    if not widgets:
        return []
    groups: list[list[object]] = []
    processed: set[str] = set()
    for widget in widgets:
        wid = _widget_id(widget)
        if wid is None or wid in processed:
            continue
        group: list[object] = [widget]
        processed.add(wid)
        changed = True
        while changed:
            changed = False
            for other in widgets:
                other_id = _widget_id(other)
                if other_id is None or other_id in processed:
                    continue
                for member in group:
                    try:
                        r1 = widget_bounding_box(member)
                        r2 = widget_bounding_box(other)
                    except (ValueError, AttributeError):
                        continue
                    if abs(r1.x - r2.x) <= tolerance and abs(r1.y - r2.y) <= tolerance:
                        group.append(other)
                        processed.add(other_id)
                        changed = True
                        break
        groups.append(group)
    return groups


def find_widget_clusters(
    widgets: Sequence[object],
    min_cluster_size: int = 2,
    tolerance: float = 20.0,
) -> list[list[object]]:
    """Return :func:`create_spatial_group` results filtered to clusters of N+ widgets."""
    return [
        group
        for group in create_spatial_group(widgets, tolerance)
        if len(group) >= min_cluster_size
    ]


def calculate_widget_density(
    widgets: Iterable[object],
    area: Rectangle,
) -> float:
    """Return widget-count-per-pixel-squared for widgets intersecting ``area``."""
    if area.width <= 0.0 or area.height <= 0.0:
        return 0.0
    area_size = area.width * area.height
    count = 0
    for widget in widgets:
        try:
            if intersects(area, widget_bounding_box(widget)):
                count += 1
        except (ValueError, AttributeError):
            continue
    return count / area_size if area_size > 0.0 else 0.0


# ---- helpers ---------------------------------------------------------------


def _widget_id(widget: object) -> str | None:
    wid = getattr(widget, "id", None)
    if isinstance(wid, str):
        return wid
    if isinstance(wid, int):
        return str(wid)
    return None


def _is_connector(widget: object) -> bool:
    return hasattr(widget, "src") and hasattr(widget, "dst")


def _rel(obj: Any) -> tuple[float, float]:
    if isinstance(obj, dict):
        return float(obj.get("x", 0.0)), float(obj.get("y", 0.0))
    if hasattr(obj, "x") and hasattr(obj, "y"):
        return float(obj.x), float(obj.y)
    return 0.0, 0.0
