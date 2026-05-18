"""Phase 4b §4.2 #20: export port.

Export a set of widgets from a canvas to a directory on disk. Output format
matches Go's ``export.go`` so exports are portable across SDKs.

On-disk layout::

    <export_folder>/
        export.json      # {widgets, assets, region}
        image_<id>.jpg   # one per image widget
        pdf_<id>.pdf     # one per PDF widget
        video_<id>.mp4   # one per video widget

The ``assets`` map in ``export.json`` is ``{widget_id: filename}``. Non-asset
widget types (notes, anchors, browsers, connectors, tables) appear only in
``widgets`` — the importer recreates them from the serialised model.
"""

from __future__ import annotations

import datetime as dt
import json
from collections.abc import Sequence
from dataclasses import dataclass, field
from pathlib import Path
from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from ..client import Client

from .geometry import Rectangle

__all__ = [
    "ExportedWidgetSet",
    "WidgetExporter",
    "export_widgets_to_folder",
]


@dataclass(slots=True)
class ExportedWidgetSet:
    """In-memory form of an export.json plus its sibling asset files."""

    widgets: list[dict[str, Any]] = field(default_factory=list)
    assets: dict[str, str] = field(default_factory=dict)
    region: Rectangle | None = None
    base_path: Path | None = None

    def to_dict(self) -> dict[str, Any]:
        """Serialise to the export.json wire shape."""
        return {
            "widgets": self.widgets,
            "assets": self.assets,
            "region": _rect_to_dict(self.region) if self.region else None,
        }

    @classmethod
    def from_dict(
        cls,
        data: dict[str, Any],
        *,
        base_path: Path | None = None,
    ) -> ExportedWidgetSet:
        """Rebuild from the export.json wire shape."""
        region: Rectangle | None = None
        raw_region = data.get("region")
        if isinstance(raw_region, dict):
            region = Rectangle(
                x=float(raw_region.get("x", 0.0)),
                y=float(raw_region.get("y", 0.0)),
                width=float(raw_region.get("width", 0.0)),
                height=float(raw_region.get("height", 0.0)),
            )
        widgets_obj = data.get("widgets") or []
        assets_obj = data.get("assets") or {}
        return cls(
            widgets=[dict(w) for w in widgets_obj if isinstance(w, dict)],
            assets={str(k): str(v) for k, v in assets_obj.items()},
            region=region,
            base_path=base_path,
        )

    @classmethod
    def load(cls, folder: str | Path) -> ExportedWidgetSet:
        """Read an export folder produced by :meth:`WidgetExporter.export_widgets_to_folder`."""
        path = Path(folder)
        manifest_path = path / "export.json"
        with manifest_path.open("r", encoding="utf-8") as fh:
            data = json.load(fh)
        return cls.from_dict(data, base_path=path)

    def asset_bytes(self, widget_id: str) -> bytes:
        """Return the on-disk asset bytes for ``widget_id``.

        Raises:
            FileNotFoundError: ``widget_id`` has no asset entry or the file
                is missing.
        """
        if widget_id not in self.assets:
            raise FileNotFoundError(
                f"export: no asset recorded for widget {widget_id!r}",
            )
        if self.base_path is None:
            raise FileNotFoundError(
                "export: cannot resolve asset bytes — ExportedWidgetSet has no base_path",
            )
        asset_path = self.base_path / self.assets[widget_id]
        return asset_path.read_bytes()


class WidgetExporter:
    """Exports widgets + assets from a canvas region to a folder on disk."""

    def __init__(self, client: Client) -> None:
        self.client = client

    async def export_widgets_to_folder(
        self,
        canvas_id: str,
        widget_ids: Sequence[str],
        region: Rectangle,
        *,
        shared_canvas_id: str | None = None,
        base_folder: str | Path | None = None,
    ) -> Path:
        """Export the named widgets to a folder. Returns the folder path.

        Args:
            canvas_id: Source canvas.
            widget_ids: IDs of widgets to export. Order is preserved in the
                manifest.
            region: Source ``Rectangle`` covered by the selection (used by
                the importer to compute scale/translate when placing into a
                new region).
            shared_canvas_id: If set, widgets whose ``parent_id`` equals
                this value have ``parent_id`` blanked in the export so they
                re-root cleanly on a different canvas (matches Go behaviour).
            base_folder: Output directory. Defaults to
                ``export/<UTC-timestamp>``.

        Returns:
            The resolved :class:`pathlib.Path` of the export folder.
        """
        folder = Path(
            base_folder
            if base_folder is not None
            else Path("export") / dt.datetime.now(tz=dt.UTC).strftime("%Y%m%d_%H%M%S")
        )
        folder.mkdir(parents=True, exist_ok=True)

        widgets_out: list[dict[str, Any]] = []
        assets: dict[str, str] = {}

        for widget_id in widget_ids:
            widget = await self.client.widgets.get(canvas_id, widget_id)
            widget_dict = widget.model_dump(mode="json", by_alias=False)
            if shared_canvas_id and widget_dict.get("parent_id") == shared_canvas_id:
                widget_dict["parent_id"] = None
            widgets_out.append(widget_dict)

            widget_type = str(widget_dict.get("widget_type") or "").lower()
            filename: str | None = None
            data: bytes | None = None
            if widget_type == "image":
                data = await self.client.widgets.images.download(canvas_id, widget_id)
                filename = f"image_{widget_id}.jpg"
            elif widget_type == "pdf":
                data = await self.client.widgets.pdfs.download(canvas_id, widget_id)
                filename = f"pdf_{widget_id}.pdf"
            elif widget_type == "video":
                data = await self.client.widgets.videos.download(canvas_id, widget_id)
                filename = f"video_{widget_id}.mp4"

            if filename is not None and data is not None:
                (folder / filename).write_bytes(data)
                assets[widget_id] = filename

        manifest = {
            "widgets": widgets_out,
            "assets": assets,
            "region": _rect_to_dict(region),
        }
        (folder / "export.json").write_text(
            json.dumps(manifest, indent=2, sort_keys=True),
            encoding="utf-8",
        )
        return folder


async def export_widgets_to_folder(
    client: Client,
    canvas_id: str,
    widget_ids: Sequence[str],
    region: Rectangle,
    *,
    shared_canvas_id: str | None = None,
    base_folder: str | Path | None = None,
) -> Path:
    """Module-level convenience wrapper around :class:`WidgetExporter`."""
    return await WidgetExporter(client).export_widgets_to_folder(
        canvas_id,
        widget_ids,
        region,
        shared_canvas_id=shared_canvas_id,
        base_folder=base_folder,
    )


def _rect_to_dict(rect: Rectangle) -> dict[str, float]:
    return {"x": rect.x, "y": rect.y, "width": rect.width, "height": rect.height}
