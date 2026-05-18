"""Widget models for every supported Canvus widget type."""

from __future__ import annotations

from typing import Any

from pydantic import Field

from ._base import CanvusModel
from .common import GridSize, RelativeLocation


class BaseWidget(CanvusModel):
    """Base class for every canvas widget.

    Fields that are server-generated (``id``, ``state``, ``widget_type``,
    ``created_at``, ``modified_at``) are present on every read response.
    Client-side creates may omit them; the server will populate them.
    """

    id: str | None = None
    widget_type: str | None = None
    state: str | None = None
    location: dict[str, float] = Field(default_factory=lambda: {"x": 0.0, "y": 0.0})
    size: dict[str, float] = Field(default_factory=lambda: {"width": 0.0, "height": 0.0})
    depth: float = 0.0
    scale: float = 1.0
    pinned: bool = False
    parent_id: str | None = None
    created_at: str | None = None
    modified_at: str | None = None


# ---- concrete widget types -------------------------------------------------


class Note(BaseWidget):
    """A text note widget."""

    widget_type: str = "Note"
    text: str = ""
    title: str | None = None
    text_color: str = "#000000ff"
    background_color: str = "#ffffffff"
    auto_text_color: bool = True


class Image(BaseWidget):
    """An image widget. ``hash`` is the server-generated asset identifier."""

    widget_type: str = "Image"
    hash: str | None = None
    original_filename: str | None = None
    title: str | None = None
    mime_type: str | None = None
    file_size: int | None = None


class Video(BaseWidget):
    """A video widget."""

    widget_type: str = "Video"
    hash: str | None = None
    original_filename: str | None = None
    title: str | None = None
    mime_type: str | None = None
    file_size: int | None = None
    playback_position: float = 0.0
    playback_state: str = "STOPPED"
    muted: bool = False
    duration: str | None = None


class PDF(BaseWidget):
    """A PDF widget. ``widget_type`` is ``Pdf`` (note the capital ``P``)."""

    widget_type: str = "Pdf"
    hash: str | None = None
    original_filename: str | None = None
    title: str | None = None
    mime_type: str | None = None
    file_size: int | None = None
    index: int = 0
    page_count: int | None = None


class Browser(BaseWidget):
    """An embedded-browser widget."""

    widget_type: str = "Browser"
    url: str = ""
    title: str | None = None
    transparent_mode: bool = False
    main_frame_scroll_offset: dict[str, float] = Field(
        default_factory=lambda: {"x": 0.0, "y": 0.0},
    )


class Anchor(BaseWidget):
    """A named navigation anchor on the canvas."""

    widget_type: str = "Anchor"
    anchor_index: int = 0
    anchor_name: str = "New anchor"


class ConnectorEndpoint(CanvusModel):
    """One end (src or dst) of a connector."""

    id: str
    rel_location: RelativeLocation = Field(default_factory=RelativeLocation)
    auto_location: bool = False
    tip: str = "none"


class Connector(BaseWidget):
    """A line / curve / arrow between two widgets."""

    widget_type: str = "Connector"
    src: ConnectorEndpoint | None = None
    dst: ConnectorEndpoint | None = None
    line_color: str = "#e7e7f2ff"
    line_width: float = 5.0
    type: str = "curve"


class Table(BaseWidget):
    """A table widget.

    Note:
        Per API changelog §4, the server does not serialise ``column_widths`` or
        ``row_heights`` — both are intentionally omitted from this model.
        Per changelog §5, ``grid_size`` is set at creation and cannot be
        changed; including it in a PATCH payload is silently ignored.
    """

    widget_type: str = "Table"
    title: str | None = None
    grid_size: GridSize | None = None


class TableCell(CanvusModel):
    """A single cell in a table widget."""

    column: int = 0
    row: int = 0
    text: str = ""
    background_color: str | None = None
    text_color: str | None = None


class VideoInput(BaseWidget):
    """A canvas-scoped video-input widget."""

    widget_type: str = "VideoInput"
    name: str | None = None
    source: str | None = None
    resolution: str | None = None
    config: dict[str, Any] = Field(default_factory=dict)


class IPVideo(BaseWidget):
    """An IP video stream widget.

    Note:
        Per API changelog §2, this widget type **cannot be created** via the
        REST API. Only GET / PATCH / DELETE are supported. Creation must
        happen from the Canvus desktop client.

        Per live-server verification (v1.2), the wire uses ``host-id`` (hyphenated)
        for the host identifier; other fields use underscores.
    """

    widget_type: str = "IpVideo"
    host_id: str | None = Field(default=None, alias="host-id")
    name: str | None = None
    source: str | None = None
    title: str | None = None


class RDPConnection(BaseWidget):
    """A remote-desktop connection widget.

    Note:
        Per API changelog §2, this widget type **cannot be created** via the
        REST API. Only GET / PATCH / DELETE are supported.

        Per API changelog §3, the actual wire field naming for ``host_id``,
        ``content_id``, ``connection_name``, and ``host_site`` (hyphens vs
        underscores) is pending verification against a live server. This
        model uses snake_case and aliases the hyphenated variants; if a
        future API release standardises on one form, only the aliases need
        to change.
    """

    widget_type: str = "RdpConnection"
    title: str | None = None
    connection_name: str | None = Field(default=None, alias="connection-name")
    host_id: str | None = Field(default=None, alias="host-id")
    host_site: str | None = Field(default=None, alias="host-site")
    content_id: str | None = Field(default=None, alias="content-id")


class UploadItem(CanvusModel):
    """An entry returned by ``GET /canvases/{id}/uploads-folder``."""

    id: str
    name: str | None = None
    type: str | None = None
    size: int | None = None
    hash: str | None = None
    uploaded_at: str | None = None


class Widget(BaseWidget):
    """Generic widget container for types not modelled explicitly.

    The :meth:`canvus_sdk.resources.widgets.WidgetsResource.list` method uses
    this as its untyped fallback when an unknown ``widget_type`` is returned.
    """

    config: dict[str, Any] = Field(default_factory=dict)


__all__ = [
    "Anchor",
    "BaseWidget",
    "Browser",
    "Connector",
    "ConnectorEndpoint",
    "IPVideo",
    "Image",
    "Note",
    "PDF",
    "RDPConnection",
    "Table",
    "TableCell",
    "UploadItem",
    "Video",
    "VideoInput",
    "Widget",
]
