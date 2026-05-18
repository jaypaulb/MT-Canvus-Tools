"""Shared primitive models used across multiple resource groups."""

from __future__ import annotations

from ._base import CanvusModel


class Location(CanvusModel):
    """A 2D pixel-space location on a canvas."""

    x: float = 0.0
    y: float = 0.0


class Size(CanvusModel):
    """A 2D pixel-space dimension."""

    width: float = 0.0
    height: float = 0.0


class RelativeLocation(CanvusModel):
    """A normalised attachment point inside a widget (0.0-1.0 on each axis)."""

    x: float = 0.5
    y: float = 0.5


class GridSize(CanvusModel):
    """Row/column count for a table widget.

    Note:
        ``grid_size`` is set at table creation and silently ignored on PATCH
        (per API changelog §5).
    """

    columns: int = 1
    rows: int = 1


class ViewRectangle(CanvusModel):
    """A pixel-space rectangle on a canvas, used by workspace view state."""

    x: float = 0.0
    y: float = 0.0
    width: float = 0.0
    height: float = 0.0


__all__ = [
    "GridSize",
    "Location",
    "RelativeLocation",
    "Size",
    "ViewRectangle",
]
