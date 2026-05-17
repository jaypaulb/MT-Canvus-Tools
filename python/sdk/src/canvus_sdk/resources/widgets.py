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

import json
import warnings
from collections.abc import AsyncIterator
from typing import Any, ClassVar

from ..errors import UnsupportedOperationError
from ..models import (
    Anchor,
    Browser,
    Connector,
    IPVideo,
    Image,
    Note,
    PDF,
    RDPConnection,
    Table,
    TableCell,
    UploadItem,
    Video,
    VideoInput,
    Widget,
)
from ._base import Resource

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


class _TypedSubResource(Resource):
    """Mixin providing CRUD helpers for a single widget type.

    Subclasses set :cvar:`_path` (URL segment) and :cvar:`_model` (Pydantic
    class). The standard CRUD methods are defined once here.
    """

    _path: ClassVar[str] = ""
    _model: ClassVar[type[Any]] = dict

    async def _list_impl(
        self, canvas_id: str, *, params: dict[str, Any] | None = None
    ) -> list[Any]:
        data = await self._transport.request(
            "GET", f"canvases/{canvas_id}/{self._path}", params=params
        )
        return self._parse_list(self._model, data)

    async def _get_impl(self, canvas_id: str, widget_id: str) -> Any:
        data = await self._transport.request(
            "GET", f"canvases/{canvas_id}/{self._path}/{widget_id}"
        )
        return self._parse(self._model, data)

    async def _create_impl(
        self, canvas_id: str, payload: dict[str, Any]
    ) -> Any:
        data = await self._transport.request(
            "POST", f"canvases/{canvas_id}/{self._path}", json_body=payload
        )
        return self._parse(self._model, data)

    async def _patch_impl(
        self, canvas_id: str, widget_id: str, payload: dict[str, Any]
    ) -> Any:
        data = await self._transport.request(
            "PATCH",
            f"canvases/{canvas_id}/{self._path}/{widget_id}",
            json_body=payload,
        )
        return self._parse(self._model, data)

    async def _delete_impl(self, canvas_id: str, widget_id: str) -> None:
        await self._transport.request(
            "DELETE", f"canvases/{canvas_id}/{self._path}/{widget_id}"
        )

    async def _download_impl(self, canvas_id: str, widget_id: str) -> bytes:
        return await self._transport.request_bytes(
            "GET", f"canvases/{canvas_id}/{self._path}/{widget_id}/download"
        )


# ---- notes -----------------------------------------------------------------


class NotesResource(_TypedSubResource):
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


# ---- images / videos / pdfs (multipart-capable) ----------------------------


class _AssetWidgetMixin(_TypedSubResource):
    """Helper for asset-backed widgets that share an upload + download surface."""

    async def upload(
        self,
        canvas_id: str,
        file_bytes: bytes,
        filename: str,
        *,
        content_type: str = "application/octet-stream",
        metadata: dict[str, Any] | None = None,
    ) -> Any:
        """Create a widget by uploading a file as multipart/form-data."""
        files = {"data": (filename, file_bytes, content_type)}
        if metadata is not None:
            files["json"] = (
                None,
                json.dumps(metadata),
                "application/json",
            )
        data = await self._transport.request(
            "POST", f"canvases/{canvas_id}/{self._path}", files=files
        )
        return self._parse(self._model, data)


class ImagesResource(_AssetWidgetMixin):
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


class VideosResource(_AssetWidgetMixin):
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


class PDFsResource(_AssetWidgetMixin):
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


# ---- browsers / anchors / connectors --------------------------------------


class BrowsersResource(_TypedSubResource):
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


class AnchorsResource(_TypedSubResource):
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


class ConnectorsResource(_TypedSubResource):
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


# ---- tables ---------------------------------------------------------------


class TablesResource(_TypedSubResource):
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

    async def list_cells(self, canvas_id: str, table_id: str) -> list[TableCell]:
        """Return the cells inside a table."""
        data = await self._transport.request(
            "GET", f"canvases/{canvas_id}/tables/{table_id}/cells"
        )
        return self._parse_list(TableCell, data)


# ---- video inputs (canvas-scoped) ------------------------------------------


class VideoInputsResource(_TypedSubResource):
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


# ---- ip-videos -------------------------------------------------------------


class IPVideosResource(_TypedSubResource):
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


# ---- rdp-connections -------------------------------------------------------


class RDPConnectionsResource(_TypedSubResource):
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

    async def list_uploads_folder(self, canvas_id: str) -> list[UploadItem]:
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
        files = {"data": (filename, file_bytes, content_type)}
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


__all__ = [
    "AnchorsResource",
    "BrowsersResource",
    "CLONE_WIDGET_PATHS",
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
