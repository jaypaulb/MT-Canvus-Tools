"""All widget-family endpoints.

Each widget type has its own CRUD set under
``/canvases/{canvas_id}/{widget_path}``. To keep this file maintainable we
factor each type's per-type endpoints into a small helper class but expose
all of them as public attributes on :class:`WidgetsResource`.

Per API changelog §1, **cross-canvas cloning** is implemented by the standard
create endpoints: include ``source_canvas_id`` and ``source_widget_id`` in the
POST body and the server clones the source widget into the destination canvas.
The deprecated ``POST /canvases/{id}/widgets/clone`` endpoint is intentionally
NOT exposed by this SDK — use :meth:`WidgetsResource.clone` instead.

Per API changelog §2, IP Video and RDP Connection widgets cannot be created
via the REST API. :class:`IPVideosResource` and :class:`RDPConnectionsResource`
expose only GET / PATCH / DELETE.
"""

from __future__ import annotations

import builtins
import json
import warnings
from collections.abc import AsyncIterator
from typing import Any, ClassVar, Generic, TypeVar, cast

from pydantic import BaseModel

from ..errors import UnsupportedOperationError
from ..models import (
    PDF,
    Anchor,
    Browser,
    Connector,
    Image,
    IPVideo,
    Note,
    RDPConnection,
    Table,
    TableCell,
    UploadItem,
    Video,
    VideoInput,
    Widget,
)
from ._base import Resource

_ModelT = TypeVar("_ModelT", bound=BaseModel)

# ---- valid widget type paths for clone_widget() ----------------------------

CLONE_WIDGET_PATHS: dict[str, str] = {
    "note": "notes",
    "notes": "notes",
    "image": "images",
    "images": "images",
    "video": "videos",
    "videos": "videos",
    "pdf": "pdfs",
    "pdfs": "pdfs",
    "browser": "browsers",
    "browsers": "browsers",
    "anchor": "anchors",
    "anchors": "anchors",
    "table": "tables",
    "tables": "tables",
}

# ---- Phase 4b §4.2 #1-#3: widget_type → URL path segment for create/update/delete
# Covers every widget family the server exposes per /widget-types/. The
# IPVideo and RDP-Connection entries are present so ``update_any`` and
# ``delete_any`` work; ``create_any`` rejects them via UnsupportedOperationError
# before dispatching.

WIDGET_TYPE_TO_PATH: dict[str, str] = {
    "note": "notes",
    "notes": "notes",
    "image": "images",
    "images": "images",
    "video": "videos",
    "videos": "videos",
    "pdf": "pdfs",
    "pdfs": "pdfs",
    "browser": "browsers",
    "browsers": "browsers",
    "anchor": "anchors",
    "anchors": "anchors",
    "connector": "connectors",
    "connectors": "connectors",
    "table": "tables",
    "tables": "tables",
    "videoinput": "video-inputs",
    "video-input": "video-inputs",
    "video-inputs": "video-inputs",
    "video_input": "video-inputs",
    "ipvideo": "ip-videos",
    "ip-video": "ip-videos",
    "ip-videos": "ip-videos",
    "ip_video": "ip-videos",
    "rdpconnection": "rdp-connections",
    "rdp-connection": "rdp-connections",
    "rdp-connections": "rdp-connections",
    "rdp_connection": "rdp-connections",
}

# Widget types the server's createElement whitelist rejects (per changelog §2).
UNCREATABLE_WIDGET_TYPES: frozenset[str] = frozenset({
    "ipvideo", "ip-video", "ip-videos", "ip_video",
    "rdpconnection", "rdp-connection", "rdp-connections", "rdp_connection",
})


def _resolve_widget_path(widget_type: str) -> str:
    """Map a widget_type string (singular or plural, hyphen or underscore) to its URL segment."""
    key = widget_type.lower()
    path = WIDGET_TYPE_TO_PATH.get(key)
    if path is None:
        raise ValueError(
            f"widget_type {widget_type!r} is not recognised. "
            f"Allowed: {sorted(set(WIDGET_TYPE_TO_PATH.values()))}",
        )
    return path


class _TypedSubResource(Resource, Generic[_ModelT]):
    """Mixin providing CRUD helpers for a single widget type.

    Subclasses set :cvar:`_path` (URL segment) and :cvar:`_model` (Pydantic
    class). The standard CRUD methods are defined once here.
    """

    _path: ClassVar[str] = ""
    _model: ClassVar[type[Any]] = dict  # overridden per subclass; type[Any] because ClassVar can't carry TypeVar

    def _get_model(self) -> type[_ModelT]:
        return cast(type[_ModelT], self._model)

    async def _list_impl(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> list[_ModelT]:
        data = await self._transport.request(
            "GET", f"canvases/{canvas_id}/{self._path}", params=params
        )
        return self._parse_list(self._get_model(), data)

    async def _get_impl(self, canvas_id: str, widget_id: str) -> _ModelT:
        data = await self._transport.request(
            "GET", f"canvases/{canvas_id}/{self._path}/{widget_id}"
        )
        return self._parse(self._get_model(), data)

    async def _create_impl(
        self, canvas_id: str, payload: dict[str, Any]
    ) -> _ModelT:
        data = await self._transport.request(
            "POST", f"canvases/{canvas_id}/{self._path}", json_body=payload
        )
        return self._parse(self._get_model(), data)

    async def _patch_impl(
        self, canvas_id: str, widget_id: str, payload: dict[str, Any]
    ) -> _ModelT:
        data = await self._transport.request(
            "PATCH",
            f"canvases/{canvas_id}/{self._path}/{widget_id}",
            json_body=payload,
        )
        return self._parse(self._get_model(), data)

    async def _delete_impl(self, canvas_id: str, widget_id: str) -> None:
        await self._transport.request(
            "DELETE", f"canvases/{canvas_id}/{self._path}/{widget_id}"
        )

    async def _download_impl(self, canvas_id: str, widget_id: str) -> bytes:
        return await self._transport.request_bytes(
            "GET", f"canvases/{canvas_id}/{self._path}/{widget_id}/download"
        )


# ---- notes -----------------------------------------------------------------


class NotesResource(_TypedSubResource[Note]):
    _path = "notes"
    _model = Note

    async def list(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> list[Note]:
        return await self._list_impl(canvas_id, params=params)

    async def get(self, canvas_id: str, note_id: str) -> Note:
        return await self._get_impl(canvas_id, note_id)

    async def create(self, canvas_id: str, payload: dict[str, Any]) -> Note:
        return await self._create_impl(canvas_id, payload)

    async def update(
        self, canvas_id: str, note_id: str, payload: dict[str, Any]
    ) -> Note:
        return await self._patch_impl(canvas_id, note_id, payload)

    async def delete(self, canvas_id: str, note_id: str) -> None:
        await self._delete_impl(canvas_id, note_id)

    def subscribe(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[Note]:
        """Subscribe to all notes on a canvas (Phase 4b §4.2 #13)."""
        return self._typed_subscribe(
            Note, f"canvases/{canvas_id}/notes", params=params
        )

    def subscribe_one(
        self, canvas_id: str, note_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[Note]:
        """Subscribe to one note (Phase 4b §4.2 #13)."""
        return self._typed_subscribe(
            Note, f"canvases/{canvas_id}/notes/{note_id}", params=params
        )


# ---- images / videos / pdfs (multipart-capable) ----------------------------


class _AssetWidgetMixin(_TypedSubResource[_ModelT]):
    """Helper for asset-backed widgets that share an upload + download surface."""

    async def upload(
        self,
        canvas_id: str,
        file_bytes: bytes,
        filename: str,
        *,
        content_type: str = "application/octet-stream",
        metadata: dict[str, Any] | None = None,
    ) -> _ModelT:
        """Create a widget by uploading a file as multipart/form-data."""
        files: dict[str, Any] = {"data": (filename, file_bytes, content_type)}
        if metadata is not None:
            files["json"] = (
                None,
                json.dumps(metadata),
                "application/json",
            )
        data = await self._transport.request(
            "POST", f"canvases/{canvas_id}/{self._path}", files=files
        )
        return self._parse(self._get_model(), data)


class ImagesResource(_AssetWidgetMixin[Image]):
    _path = "images"
    _model = Image

    async def list(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> list[Image]:
        return await self._list_impl(canvas_id, params=params)

    async def get(self, canvas_id: str, image_id: str) -> Image:
        return await self._get_impl(canvas_id, image_id)

    async def create(self, canvas_id: str, payload: dict[str, Any]) -> Image:
        return await self._create_impl(canvas_id, payload)

    async def update(
        self, canvas_id: str, image_id: str, payload: dict[str, Any]
    ) -> Image:
        return await self._patch_impl(canvas_id, image_id, payload)

    async def delete(self, canvas_id: str, image_id: str) -> None:
        await self._delete_impl(canvas_id, image_id)

    async def download(self, canvas_id: str, image_id: str) -> bytes:
        return await self._download_impl(canvas_id, image_id)

    def subscribe(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[Image]:
        """Subscribe to all images on a canvas (Phase 4b §4.2 #13)."""
        return self._typed_subscribe(
            Image, f"canvases/{canvas_id}/images", params=params
        )

    def subscribe_one(
        self, canvas_id: str, image_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[Image]:
        """Subscribe to one image."""
        return self._typed_subscribe(
            Image, f"canvases/{canvas_id}/images/{image_id}", params=params
        )


class VideosResource(_AssetWidgetMixin[Video]):
    _path = "videos"
    _model = Video

    async def list(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> list[Video]:
        return await self._list_impl(canvas_id, params=params)

    async def get(self, canvas_id: str, video_id: str) -> Video:
        return await self._get_impl(canvas_id, video_id)

    async def create(self, canvas_id: str, payload: dict[str, Any]) -> Video:
        return await self._create_impl(canvas_id, payload)

    async def update(
        self, canvas_id: str, video_id: str, payload: dict[str, Any]
    ) -> Video:
        return await self._patch_impl(canvas_id, video_id, payload)

    async def delete(self, canvas_id: str, video_id: str) -> None:
        await self._delete_impl(canvas_id, video_id)

    async def download(self, canvas_id: str, video_id: str) -> bytes:
        return await self._download_impl(canvas_id, video_id)

    def subscribe(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[Video]:
        """Subscribe to all videos on a canvas (Phase 4b §4.2 #13)."""
        return self._typed_subscribe(
            Video, f"canvases/{canvas_id}/videos", params=params
        )

    def subscribe_one(
        self, canvas_id: str, video_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[Video]:
        """Subscribe to one video."""
        return self._typed_subscribe(
            Video, f"canvases/{canvas_id}/videos/{video_id}", params=params
        )


class PDFsResource(_AssetWidgetMixin[PDF]):
    _path = "pdfs"
    _model = PDF

    async def list(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> list[PDF]:
        return await self._list_impl(canvas_id, params=params)

    async def get(self, canvas_id: str, pdf_id: str) -> PDF:
        return await self._get_impl(canvas_id, pdf_id)

    async def create(self, canvas_id: str, payload: dict[str, Any]) -> PDF:
        return await self._create_impl(canvas_id, payload)

    async def update(
        self, canvas_id: str, pdf_id: str, payload: dict[str, Any]
    ) -> PDF:
        return await self._patch_impl(canvas_id, pdf_id, payload)

    async def delete(self, canvas_id: str, pdf_id: str) -> None:
        await self._delete_impl(canvas_id, pdf_id)

    async def download(self, canvas_id: str, pdf_id: str) -> bytes:
        return await self._download_impl(canvas_id, pdf_id)

    def subscribe(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[PDF]:
        """Subscribe to all PDFs on a canvas (Phase 4b §4.2 #13)."""
        return self._typed_subscribe(
            PDF, f"canvases/{canvas_id}/pdfs", params=params
        )

    def subscribe_one(
        self, canvas_id: str, pdf_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[PDF]:
        """Subscribe to one PDF."""
        return self._typed_subscribe(
            PDF, f"canvases/{canvas_id}/pdfs/{pdf_id}", params=params
        )


# ---- browsers / anchors / connectors --------------------------------------


class BrowsersResource(_TypedSubResource[Browser]):
    _path = "browsers"
    _model = Browser

    async def list(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> list[Browser]:
        return await self._list_impl(canvas_id, params=params)

    async def get(self, canvas_id: str, browser_id: str) -> Browser:
        return await self._get_impl(canvas_id, browser_id)

    async def create(self, canvas_id: str, payload: dict[str, Any]) -> Browser:
        return await self._create_impl(canvas_id, payload)

    async def update(
        self, canvas_id: str, browser_id: str, payload: dict[str, Any]
    ) -> Browser:
        return await self._patch_impl(canvas_id, browser_id, payload)

    async def delete(self, canvas_id: str, browser_id: str) -> None:
        await self._delete_impl(canvas_id, browser_id)

    def subscribe(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[Browser]:
        """Subscribe to all browsers on a canvas (Phase 4b §4.2 #13)."""
        return self._typed_subscribe(
            Browser, f"canvases/{canvas_id}/browsers", params=params
        )

    def subscribe_one(
        self, canvas_id: str, browser_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[Browser]:
        """Subscribe to one browser."""
        return self._typed_subscribe(
            Browser, f"canvases/{canvas_id}/browsers/{browser_id}", params=params
        )


class AnchorsResource(_TypedSubResource[Anchor]):
    _path = "anchors"
    _model = Anchor

    async def list(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> list[Anchor]:
        return await self._list_impl(canvas_id, params=params)

    async def get(self, canvas_id: str, anchor_id: str) -> Anchor:
        return await self._get_impl(canvas_id, anchor_id)

    async def create(self, canvas_id: str, payload: dict[str, Any]) -> Anchor:
        return await self._create_impl(canvas_id, payload)

    async def update(
        self, canvas_id: str, anchor_id: str, payload: dict[str, Any]
    ) -> Anchor:
        return await self._patch_impl(canvas_id, anchor_id, payload)

    async def delete(self, canvas_id: str, anchor_id: str) -> None:
        await self._delete_impl(canvas_id, anchor_id)

    def subscribe(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[Anchor]:
        """Subscribe to all anchors on a canvas (Phase 4b §4.2 #13)."""
        return self._typed_subscribe(
            Anchor, f"canvases/{canvas_id}/anchors", params=params
        )

    def subscribe_one(
        self, canvas_id: str, anchor_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[Anchor]:
        """Subscribe to one anchor."""
        return self._typed_subscribe(
            Anchor, f"canvases/{canvas_id}/anchors/{anchor_id}", params=params
        )


class ConnectorsResource(_TypedSubResource[Connector]):
    _path = "connectors"
    _model = Connector

    async def list(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> list[Connector]:
        return await self._list_impl(canvas_id, params=params)

    async def get(self, canvas_id: str, connector_id: str) -> Connector:
        return await self._get_impl(canvas_id, connector_id)

    async def create(self, canvas_id: str, payload: dict[str, Any]) -> Connector:
        return await self._create_impl(canvas_id, payload)

    async def update(
        self, canvas_id: str, connector_id: str, payload: dict[str, Any]
    ) -> Connector:
        return await self._patch_impl(canvas_id, connector_id, payload)

    async def delete(self, canvas_id: str, connector_id: str) -> None:
        await self._delete_impl(canvas_id, connector_id)

    def subscribe(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[Connector]:
        """Subscribe to all connectors on a canvas (Phase 4b §4.2 #13)."""
        return self._typed_subscribe(
            Connector, f"canvases/{canvas_id}/connectors", params=params
        )

    def subscribe_one(
        self, canvas_id: str, connector_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[Connector]:
        """Subscribe to one connector."""
        return self._typed_subscribe(
            Connector,
            f"canvases/{canvas_id}/connectors/{connector_id}",
            params=params,
        )


# ---- tables ---------------------------------------------------------------


class TablesResource(_TypedSubResource[Table]):
    """Table widget endpoints (per Phase 3 Python work item #1).

    Per API changelog §4, the server does NOT serialise ``column_widths`` or
    ``row_heights`` — they are intentionally absent from the :class:`Table`
    model. Per changelog §5, ``grid_size`` is set at creation and is
    silently ignored on PATCH; :meth:`update` emits a ``UserWarning`` if a
    caller passes it.
    """

    _path = "tables"
    _model = Table

    async def list(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> list[Table]:
        return await self._list_impl(canvas_id, params=params)

    async def get(self, canvas_id: str, table_id: str) -> Table:
        return await self._get_impl(canvas_id, table_id)

    async def create(self, canvas_id: str, payload: dict[str, Any]) -> Table:
        return await self._create_impl(canvas_id, payload)

    async def update(
        self, canvas_id: str, table_id: str, payload: dict[str, Any]
    ) -> Table:
        if "grid_size" in payload or "grid-size" in payload:
            warnings.warn(
                "grid_size cannot be modified after creation and will be "
                "silently ignored by the server (see API changelog §5).",
                UserWarning,
                stacklevel=2,
            )
        return await self._patch_impl(canvas_id, table_id, payload)

    async def delete(self, canvas_id: str, table_id: str) -> None:
        await self._delete_impl(canvas_id, table_id)

    async def list_cells(self, canvas_id: str, table_id: str) -> builtins.list[TableCell]:
        """Return the cells inside a table."""
        data = await self._transport.request(
            "GET", f"canvases/{canvas_id}/tables/{table_id}/cells"
        )
        return self._parse_list(TableCell, data)

    def subscribe(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[Table]:
        """Subscribe to all tables on a canvas (Phase 4b §4.2 #13)."""
        return self._typed_subscribe(
            Table, f"canvases/{canvas_id}/tables", params=params
        )

    def subscribe_one(
        self, canvas_id: str, table_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[Table]:
        """Subscribe to one table."""
        return self._typed_subscribe(
            Table, f"canvases/{canvas_id}/tables/{table_id}", params=params
        )

    def subscribe_cells(
        self,
        canvas_id: str,
        table_id: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> AsyncIterator[TableCell]:
        """Subscribe to a table's cells (Phase 4b §4.2 #14)."""
        return self._typed_subscribe(
            TableCell,
            f"canvases/{canvas_id}/tables/{table_id}/cells",
            params=params,
        )


# ---- video inputs (canvas-scoped) ------------------------------------------


class VideoInputsResource(_TypedSubResource[VideoInput]):
    _path = "video-inputs"
    _model = VideoInput

    async def list(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> list[VideoInput]:
        return await self._list_impl(canvas_id, params=params)

    async def get(self, canvas_id: str, widget_id: str) -> VideoInput:
        """Get a single video-input widget (added in Phase 3 Python item #5)."""
        return await self._get_impl(canvas_id, widget_id)

    async def create(self, canvas_id: str, payload: dict[str, Any]) -> VideoInput:
        return await self._create_impl(canvas_id, payload)

    async def update(
        self, canvas_id: str, widget_id: str, payload: dict[str, Any]
    ) -> VideoInput:
        """Update a video-input widget (added in Phase 3 Python item #5)."""
        return await self._patch_impl(canvas_id, widget_id, payload)

    async def delete(self, canvas_id: str, widget_id: str) -> None:
        await self._delete_impl(canvas_id, widget_id)

    def subscribe(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[VideoInput]:
        """Subscribe to all video-inputs on a canvas (Phase 4b §4.2 #13)."""
        return self._typed_subscribe(
            VideoInput, f"canvases/{canvas_id}/video-inputs", params=params
        )

    def subscribe_one(
        self, canvas_id: str, widget_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[VideoInput]:
        """Subscribe to one video-input."""
        return self._typed_subscribe(
            VideoInput,
            f"canvases/{canvas_id}/video-inputs/{widget_id}",
            params=params,
        )


# ---- ip-videos -------------------------------------------------------------


class IPVideosResource(_TypedSubResource[IPVideo]):
    """IP video widget endpoints.

    GET / PATCH / DELETE only. Create is unsupported (changelog §2).
    """

    _path = "ip-videos"
    _model = IPVideo

    async def list(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> list[IPVideo]:
        return await self._list_impl(canvas_id, params=params)

    async def get(self, canvas_id: str, widget_id: str) -> IPVideo:
        return await self._get_impl(canvas_id, widget_id)

    async def update(
        self, canvas_id: str, widget_id: str, payload: dict[str, Any]
    ) -> IPVideo:
        return await self._patch_impl(canvas_id, widget_id, payload)

    async def delete(self, canvas_id: str, widget_id: str) -> None:
        await self._delete_impl(canvas_id, widget_id)

    async def create(self, *args: Any, **kwargs: Any) -> IPVideo:
        """Always raises — IP Video widgets cannot be created via the REST API."""
        raise UnsupportedOperationError(
            "IP Video widgets can only be created from the Canvus desktop "
            "client; the REST API does not support POST on /ip-videos "
            "(see API changelog §2).",
        )

    def subscribe(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[IPVideo]:
        """Subscribe to all IP videos on a canvas (Phase 4b §4.2 #13)."""
        return self._typed_subscribe(
            IPVideo, f"canvases/{canvas_id}/ip-videos", params=params
        )

    def subscribe_one(
        self, canvas_id: str, widget_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[IPVideo]:
        """Subscribe to one IP video."""
        return self._typed_subscribe(
            IPVideo,
            f"canvases/{canvas_id}/ip-videos/{widget_id}",
            params=params,
        )


# ---- rdp-connections -------------------------------------------------------


class RDPConnectionsResource(_TypedSubResource[RDPConnection]):
    """RDP connection widget endpoints.

    GET / PATCH / DELETE only. Create is unsupported (changelog §2).
    Per changelog §3, the actual wire field names for ``host_id`` etc. (hyphen
    vs underscore) are pending live-server verification.
    """

    _path = "rdp-connections"
    _model = RDPConnection

    async def list(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> list[RDPConnection]:
        return await self._list_impl(canvas_id, params=params)

    async def get(self, canvas_id: str, widget_id: str) -> RDPConnection:
        return await self._get_impl(canvas_id, widget_id)

    async def update(
        self, canvas_id: str, widget_id: str, payload: dict[str, Any]
    ) -> RDPConnection:
        return await self._patch_impl(canvas_id, widget_id, payload)

    async def delete(self, canvas_id: str, widget_id: str) -> None:
        await self._delete_impl(canvas_id, widget_id)

    async def create(self, *args: Any, **kwargs: Any) -> RDPConnection:
        """Always raises — RDP Connection widgets cannot be created via REST."""
        raise UnsupportedOperationError(
            "RDP Connection widgets can only be created from the Canvus "
            "desktop client; the REST API does not support POST on "
            "/rdp-connections (see API changelog §2).",
        )

    def subscribe(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[RDPConnection]:
        """Subscribe to all RDP connections on a canvas (Phase 4b §4.2 #13)."""
        return self._typed_subscribe(
            RDPConnection,
            f"canvases/{canvas_id}/rdp-connections",
            params=params,
        )

    def subscribe_one(
        self, canvas_id: str, widget_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[RDPConnection]:
        """Subscribe to one RDP connection."""
        return self._typed_subscribe(
            RDPConnection,
            f"canvases/{canvas_id}/rdp-connections/{widget_id}",
            params=params,
        )


# ---- top-level WidgetsResource --------------------------------------------


class WidgetsResource(Resource):
    """Container for every widget-family resource plus generic widget ops.

    Attributes:
        notes: :class:`NotesResource`
        images: :class:`ImagesResource`
        videos: :class:`VideosResource`
        pdfs: :class:`PDFsResource`
        browsers: :class:`BrowsersResource`
        anchors: :class:`AnchorsResource`
        connectors: :class:`ConnectorsResource`
        tables: :class:`TablesResource`
        video_inputs: :class:`VideoInputsResource`
        ip_videos: :class:`IPVideosResource`
        rdp_connections: :class:`RDPConnectionsResource`
    """

    def __init__(self, transport: Any) -> None:
        super().__init__(transport)
        self.notes = NotesResource(transport)
        self.images = ImagesResource(transport)
        self.videos = VideosResource(transport)
        self.pdfs = PDFsResource(transport)
        self.browsers = BrowsersResource(transport)
        self.anchors = AnchorsResource(transport)
        self.connectors = ConnectorsResource(transport)
        self.tables = TablesResource(transport)
        self.video_inputs = VideoInputsResource(transport)
        self.ip_videos = IPVideosResource(transport)
        self.rdp_connections = RDPConnectionsResource(transport)

    # ---- generic widget endpoints ------------------------------------------

    async def list(
        self,
        canvas_id: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> list[Widget]:
        """List every widget on a canvas (mixed types).

        Returns a list of :class:`Widget` (the generic container). Callers
        that want type-specific models should call the per-type resource
        (e.g. ``client.widgets.notes.list(canvas_id)``).
        """
        data = await self._transport.request(
            "GET", f"canvases/{canvas_id}/widgets", params=params
        )
        return self._parse_list(Widget, data)

    async def get(self, canvas_id: str, widget_id: str) -> Widget:
        """Get one widget by ID via the generic endpoint."""
        data = await self._transport.request(
            "GET", f"canvases/{canvas_id}/widgets/{widget_id}"
        )
        return self._parse(Widget, data)

    # ---- generic create/update/delete (Phase 4b §4.2 #1-#3) ----------------

    async def create_any(
        self,
        canvas_id: str,
        payload: dict[str, Any],
    ) -> dict[str, Any]:
        """Create a widget when the caller already has a wire-shape payload.

        Phase 4b §4.2 #1: generic dispatcher mirroring Go's
        ``widgets.go:53 CreateWidget``. The typed per-type APIs
        (``widgets.notes.create`` etc.) remain the recommended surface; this
        method exists for callers that already hold a ``payload`` containing
        ``widget_type`` and want the SDK to route it for them.

        Args:
            canvas_id: ID of the destination canvas.
            payload: Wire-shape body. Must contain ``widget_type``.

        Returns:
            Raw decoded response dict (type-specific; callers may validate
            against the matching :mod:`canvus_sdk.models` class).

        Raises:
            ValueError: ``payload`` lacks ``widget_type`` or it is unknown.
            UnsupportedOperationError: ``widget_type`` is IP Video or RDP
                Connection — these can only be created from the Canvus
                desktop client (changelog §2).
        """
        widget_type_obj = payload.get("widget_type") or payload.get("widget-type")
        if not isinstance(widget_type_obj, str):
            raise ValueError(
                "create_any: payload must include a string 'widget_type' field",
            )
        widget_type = widget_type_obj
        key = widget_type.lower()
        if key in UNCREATABLE_WIDGET_TYPES:
            raise UnsupportedOperationError(
                f"widget_type {widget_type!r} cannot be created via the REST API "
                "(see API changelog §2).",
            )
        path = _resolve_widget_path(widget_type)
        data = await self._transport.request(
            "POST",
            f"canvases/{canvas_id}/{path}",
            json_body=payload,
        )
        return data if isinstance(data, dict) else {"raw": data}

    async def update_any(
        self,
        canvas_id: str,
        widget_id: str,
        payload: dict[str, Any],
    ) -> dict[str, Any]:
        """Update a widget when the caller already has a wire-shape payload.

        Phase 4b §4.2 #2: generic dispatcher mirroring Go's
        ``widgets.go:120 UpdateWidget``. The payload must include
        ``widget_type`` so this method knows which type-specific endpoint to
        hit. For tables, ``grid_size`` is stripped before sending (silently
        ignored by the server per changelog §5).

        Args:
            canvas_id: Owning canvas ID.
            widget_id: Widget ID to update.
            payload: Wire-shape body. Must contain ``widget_type``.

        Returns:
            Raw decoded response dict.

        Raises:
            ValueError: ``payload`` lacks ``widget_type`` or it is unknown.
        """
        widget_type_obj = payload.get("widget_type") or payload.get("widget-type")
        if not isinstance(widget_type_obj, str):
            raise ValueError(
                "update_any: payload must include a string 'widget_type' field",
            )
        widget_type = widget_type_obj
        path = _resolve_widget_path(widget_type)
        body: dict[str, Any] = dict(payload)
        if path == "tables":
            for stripped in ("grid_size", "grid-size"):
                if stripped in body:
                    warnings.warn(
                        "grid_size cannot be modified after creation and will be "
                        "silently ignored by the server (see API changelog §5).",
                        UserWarning,
                        stacklevel=2,
                    )
                    body.pop(stripped, None)
        data = await self._transport.request(
            "PATCH",
            f"canvases/{canvas_id}/{path}/{widget_id}",
            json_body=body,
        )
        return data if isinstance(data, dict) else {"raw": data}

    async def delete_any(
        self,
        canvas_id: str,
        widget_id: str,
        widget_type: str,
    ) -> None:
        """Delete a widget by ID + widget_type.

        Phase 4b §4.2 #3: generic dispatcher mirroring Go's
        ``widgets.go:193 DeleteWidget``. ``widget_type`` must be supplied by
        the caller (the server response is empty so we cannot infer it).
        """
        path = _resolve_widget_path(widget_type)
        await self._transport.request(
            "DELETE",
            f"canvases/{canvas_id}/{path}/{widget_id}",
        )

    async def patch_parent_id(
        self,
        canvas_id: str,
        widget_id: str,
        parent_id: str,
    ) -> Widget:
        """Re-parent a widget by PATCHing its parent_id.

        Phase 4b §4.2 #4: mirrors Go's ``widgets.go:224 PatchParentID``. The
        server accepts ``PATCH /canvases/{cid}/widgets/{wid}`` as a generic
        shortcut even though the read-only ``/widgets`` path doesn't formally
        document this. Prefer the typed per-type ``update`` helpers where
        feasible.
        """
        data = await self._transport.request(
            "PATCH",
            f"canvases/{canvas_id}/widgets/{widget_id}",
            json_body={"parent_id": parent_id},
        )
        return self._parse(Widget, data)

    async def clone(
        self,
        dest_canvas_id: str,
        source_canvas_id: str,
        source_widget_id: str,
        widget_type: str,
        *,
        location: dict[str, float] | None = None,
        extra: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        """Clone a widget into ``dest_canvas_id`` (per changelog §1).

        Internally issues
        ``POST /canvases/{dest_canvas_id}/{widget_path}`` with a body that
        includes ``source_canvas_id``, ``source_widget_id``, and the optional
        ``location`` override. The deprecated
        ``POST /canvases/{id}/widgets/clone`` endpoint is intentionally not
        called — server-side cloning is implemented as a behaviour of the
        per-type create endpoints.

        Args:
            dest_canvas_id: ID of the canvas to clone INTO. Caller must have
                edit access.
            source_canvas_id: ID of the canvas to clone FROM. Caller must have
                at least view access.
            source_widget_id: ID of the widget to clone.
            widget_type: Singular or plural type name — one of ``note``,
                ``image``, ``video``, ``pdf``, ``browser``, ``anchor``, or
                ``table`` (singular or plural forms accepted).
            location: Optional ``{"x": ..., "y": ...}`` pixel coordinates to
                override the source widget's position on the destination.
            extra: Optional additional body fields merged into the request.

        Returns:
            Raw decoded response dict (type-specific; callers should validate
            via the appropriate model if needed).

        Raises:
            ValueError: If ``widget_type`` is not a cloneable widget family.
        """
        path_segment = CLONE_WIDGET_PATHS.get(widget_type.lower())
        if path_segment is None:
            raise ValueError(
                f"widget_type {widget_type!r} is not cloneable. "
                f"Allowed: {sorted(set(CLONE_WIDGET_PATHS.values()))}",
            )
        body: dict[str, Any] = {
            "source_canvas_id": source_canvas_id,
            "source_widget_id": source_widget_id,
        }
        if location is not None:
            body["location"] = location
        if extra:
            body.update(extra)
        data = await self._transport.request(
            "POST",
            f"canvases/{dest_canvas_id}/{path_segment}",
            json_body=body,
        )
        return data if isinstance(data, dict) else {"raw": data}

    # ---- uploads-folder ----------------------------------------------------

    async def list_uploads_folder(self, canvas_id: str) -> builtins.list[UploadItem]:
        """List items in the canvas uploads folder (Phase 3 Python item #6)."""
        data = await self._transport.request(
            "GET", f"canvases/{canvas_id}/uploads-folder"
        )
        return self._parse_list(UploadItem, data)

    async def upload_to_uploads_folder(
        self,
        canvas_id: str,
        file_bytes: bytes,
        filename: str,
        *,
        content_type: str = "application/octet-stream",
        metadata: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        """Upload a file (note or asset) to the canvas uploads folder."""
        files: dict[str, Any] = {"data": (filename, file_bytes, content_type)}
        if metadata is not None:
            files["json"] = (
                None,
                json.dumps(metadata),
                "application/json",
            )
        data = await self._transport.request(
            "POST", f"canvases/{canvas_id}/uploads-folder", files=files
        )
        return data if isinstance(data, dict) else {"raw": data}

    # ---- subscription / streaming ------------------------------------------

    async def subscribe(
        self,
        canvas_id: str,
        *,
        widget_type: str | None = None,
        params: dict[str, Any] | None = None,
    ) -> AsyncIterator[dict[str, Any]]:
        """Subscribe to widget updates on a canvas.

        Yields decoded JSON updates one at a time. ``widget_type`` may be
        any of the type paths (``notes``, ``images``, ...) or omitted for
        the generic ``/widgets`` endpoint.
        """
        path = (
            f"canvases/{canvas_id}/{widget_type}"
            if widget_type is not None
            else f"canvases/{canvas_id}/widgets"
        )
        query = dict(params) if params else {}
        query["subscribe"] = "true"
        async for line in self._transport.stream_lines("GET", path, params=query):
            try:
                yield json.loads(line)
            except json.JSONDecodeError:
                continue

    def subscribe_one(
        self,
        canvas_id: str,
        widget_id: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> AsyncIterator[Widget]:
        """Subscribe to one widget via the generic ``/widgets/{id}`` endpoint.

        Phase 4b §4.2 #13.
        """
        return self._typed_subscribe(
            Widget, f"canvases/{canvas_id}/widgets/{widget_id}", params=params
        )

    def subscribe_uploads_folder(
        self,
        canvas_id: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> AsyncIterator[UploadItem]:
        """Subscribe to a canvas's uploads-folder (Phase 4b §4.2 #14)."""
        return self._typed_subscribe(
            UploadItem,
            f"canvases/{canvas_id}/uploads-folder",
            params=params,
        )


__all__ = [
    "CLONE_WIDGET_PATHS",
    "AnchorsResource",
    "BrowsersResource",
    "ConnectorsResource",
    "IPVideosResource",
    "ImagesResource",
    "NotesResource",
    "PDFsResource",
    "RDPConnectionsResource",
    "TablesResource",
    "VideoInputsResource",
    "VideosResource",
    "WidgetsResource",
]
