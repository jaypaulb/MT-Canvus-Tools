"""Phase 4b §4.2 #20: import port.

Recreate widgets from an :class:`ExportedWidgetSet`, scaling them to fit a
new region. Mirrors Go's ``import.go`` (``ImportWidgetsToRegion``).

The module is named ``import_`` because ``import`` is a Python keyword. The
public API re-exports through :mod:`canvus_sdk.extras` so callers see the
clean name :func:`canvus_sdk.extras.import_widgets_from_folder`.
"""

from __future__ import annotations

from pathlib import Path
from typing import TYPE_CHECKING, Any

from ..errors import ValidationError
from .export import ExportedWidgetSet
from .geometry import Rectangle

if TYPE_CHECKING:
    from ..client import Client

__all__ = [
    "WidgetImporter",
    "import_widgets_from_folder",
]


class WidgetImporter:
    """Recreate widgets from an :class:`ExportedWidgetSet`."""

    def __init__(self, client: Client) -> None:
        self.client = client

    async def import_widgets_from_folder(
        self,
        canvas_id: str,
        folder: str | Path,
        target_region: Rectangle,
    ) -> list[str]:
        """Read ``folder/export.json``, recreate widgets, return new IDs."""
        exported = ExportedWidgetSet.load(folder)
        return await self.import_widgets_to_region(canvas_id, exported, target_region)

    async def import_widgets_to_region(
        self,
        canvas_id: str,
        exported: ExportedWidgetSet,
        target_region: Rectangle,
    ) -> list[str]:
        """Recreate widgets, scaled+translated into ``target_region``."""
        if not exported.widgets:
            return []
        if exported.region is None:
            raise ValidationError(
                "import_widgets_to_region: ExportedWidgetSet has no source region",
            )
        if exported.region.width <= 0.0 or exported.region.height <= 0.0:
            raise ValidationError(
                "import_widgets_to_region: source region has zero or negative size",
            )
        scale_x = target_region.width / exported.region.width
        scale_y = target_region.height / exported.region.height
        dx = target_region.x - exported.region.x * scale_x
        dy = target_region.y - exported.region.y * scale_y

        new_ids: list[str] = []
        for widget in exported.widgets:
            widget_id = str(widget.get("id") or "")
            widget_type = str(widget.get("widget_type") or "").lower()
            scaled = _apply_transform(widget, scale_x, scale_y, dx, dy)

            if widget_type == "image":
                data = exported.asset_bytes(widget_id)
                created = await self.client.widgets.images.upload(
                    canvas_id,
                    data,
                    f"imported_image_{widget_id}.jpg",
                    content_type="image/jpeg",
                    metadata=_asset_meta(scaled),
                )
                new_ids.append(_get_str_id(created))
                continue
            if widget_type == "pdf":
                data = exported.asset_bytes(widget_id)
                created_pdf = await self.client.widgets.pdfs.upload(
                    canvas_id,
                    data,
                    f"imported_{widget_id}.pdf",
                    content_type="application/pdf",
                    metadata=_asset_meta(scaled),
                )
                new_ids.append(_get_str_id(created_pdf))
                continue
            if widget_type == "video":
                data = exported.asset_bytes(widget_id)
                created_video = await self.client.widgets.videos.upload(
                    canvas_id,
                    data,
                    f"imported_video_{widget_id}.mp4",
                    content_type="video/mp4",
                    metadata=_asset_meta(scaled),
                )
                new_ids.append(_get_str_id(created_video))
                continue

            # Non-asset widget: round-trip via the generic create_any dispatcher.
            payload = dict(scaled)
            # Strip server-generated keys; the new server will assign fresh ones.
            for stripped in ("id", "created_at", "modified_at", "state"):
                payload.pop(stripped, None)
            created_dict = await self.client.widgets.create_any(canvas_id, payload)
            new_ids.append(str(created_dict.get("id") or ""))
        return new_ids


async def import_widgets_from_folder(
    client: Client,
    canvas_id: str,
    folder: str | Path,
    target_region: Rectangle,
) -> list[str]:
    """Module-level convenience wrapper around :class:`WidgetImporter`."""
    return await WidgetImporter(client).import_widgets_from_folder(
        canvas_id, folder, target_region
    )


def _apply_transform(
    widget: dict[str, Any],
    scale_x: float,
    scale_y: float,
    dx: float,
    dy: float,
) -> dict[str, Any]:
    out = dict(widget)
    location = out.get("location")
    if isinstance(location, dict):
        out["location"] = {
            "x": float(location.get("x", 0.0)) * scale_x + dx,
            "y": float(location.get("y", 0.0)) * scale_y + dy,
        }
    size = out.get("size")
    if isinstance(size, dict):
        out["size"] = {
            "width": float(size.get("width", 0.0)) * scale_x,
            "height": float(size.get("height", 0.0)) * scale_y,
        }
    return out


def _asset_meta(widget: dict[str, Any]) -> dict[str, Any]:
    meta: dict[str, Any] = {
        "title": str(widget.get("id") or ""),
        "widget_type": str(widget.get("widget_type") or "").lower(),
    }
    if isinstance(widget.get("location"), dict):
        meta["location"] = dict(widget["location"])
    if isinstance(widget.get("size"), dict):
        meta["size"] = dict(widget["size"])
    return meta


def _get_str_id(obj: Any) -> str:
    """Coerce a pydantic model or dict to its string ``id`` field."""
    wid: Any = None
    if hasattr(obj, "id"):
        wid = obj.id
    elif isinstance(obj, dict):
        wid = obj.get("id")
    return str(wid) if wid is not None else ""
