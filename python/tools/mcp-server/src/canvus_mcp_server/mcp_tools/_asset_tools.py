"""Shared CRUD MCP tools for asset widgets (images, videos, PDFs).

The legacy ``canvus-mcp-server`` had near-identical 100-line classes per
asset type (ImageCreateTool, VideoCreateTool, PdfCreateTool, etc.). The new
SDK exposes a uniform ``client.widgets.<asset>.upload(...)`` /
``.get/list/update/delete`` shape across these three types, so we factor out
the common implementation here and parameterise on the SDK resource name.

This is one place where the "rule of three" is satisfied — three near-
identical implementations existed in the source. The factoring is an
intentional Phase 4c cleanup; per-asset modules thinly subclass and set
the metadata.
"""

from __future__ import annotations

import mimetypes
from pathlib import Path
from typing import Any

from canvus_sdk import Client

from ._helpers import (
    dump_model,
    optional_dict,
    optional_number,
    require_str,
    run_with_api_error_translation,
)
from .base import BaseMCPTool, MCPToolValidationError


def _resource(client: Client, kind: str) -> Any:
    """Return the SDK sub-resource for ``kind`` (``"images"``, etc.)."""
    return getattr(client.widgets, kind)


class _AssetCreateTool(BaseMCPTool):
    """Create an asset widget by reading bytes from a local file path.

    The legacy tools accepted ``file_path`` directly; we preserve that
    surface for the existing MCP clients but call the new SDK's
    ``upload(bytes, filename)`` rather than the legacy ``create_image(path)``.
    """

    _kind: str
    _id_key: str

    def __init__(
        self,
        client: Client,
        *,
        name: str,
        description: str,
        kind: str,
        id_key: str,
    ) -> None:
        super().__init__(name=name, description=description)
        self.client = client
        self._kind = kind
        self._id_key = id_key

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        require_str(kwargs, "file_path", self.name)
        optional_number(kwargs, "x", self.name)
        optional_number(kwargs, "y", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        file_path = require_str(kwargs, "file_path", self.name)
        x = optional_number(kwargs, "x", self.name)
        y = optional_number(kwargs, "y", self.name)
        path = Path(file_path)
        if not path.is_file():
            raise MCPToolValidationError(
                f"file_path does not exist or is not a file: {file_path}",
                self.name,
            )
        content_type = (
            mimetypes.guess_type(path.name)[0] or "application/octet-stream"
        )
        file_bytes = path.read_bytes()
        metadata: dict[str, Any] | None = None
        if x is not None and y is not None:
            metadata = {"location": {"x": x, "y": y}}
        self.logger.info(
            f"{self._kind}_create",
            canvas_id=canvas_id,
            file=path.name,
            bytes=len(file_bytes),
        )
        resource = _resource(self.client, self._kind)
        created = await run_with_api_error_translation(
            self.name,
            f"Create {self._kind} widget",
            lambda: resource.upload(
                canvas_id,
                file_bytes,
                path.name,
                content_type=content_type,
                metadata=metadata,
            ),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            self._id_key: created.id,
            "file_path": file_path,
            "position": {"x": x, "y": y} if x is not None and y is not None else None,
            "create_result": dump_model(created),
            "status": "success",
        }


class _AssetGetTool(BaseMCPTool):
    _kind: str
    _id_param: str

    def __init__(
        self,
        client: Client,
        *,
        name: str,
        description: str,
        kind: str,
        id_param: str,
    ) -> None:
        super().__init__(name=name, description=description)
        self.client = client
        self._kind = kind
        self._id_param = id_param

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        require_str(kwargs, self._id_param, self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        widget_id = require_str(kwargs, self._id_param, self.name)
        resource = _resource(self.client, self._kind)
        widget = await run_with_api_error_translation(
            self.name,
            f"Retrieve {self._kind} widget",
            lambda: resource.get(canvas_id, widget_id),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            self._id_param: widget_id,
            f"{self._kind.rstrip('s')}_data": dump_model(widget),
            "status": "success",
        }


class _AssetListTool(BaseMCPTool):
    _kind: str
    _list_key: str

    def __init__(
        self,
        client: Client,
        *,
        name: str,
        description: str,
        kind: str,
        list_key: str,
    ) -> None:
        super().__init__(name=name, description=description)
        self.client = client
        self._kind = kind
        self._list_key = list_key

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        resource = _resource(self.client, self._kind)
        items = await run_with_api_error_translation(
            self.name,
            f"List {self._kind} widgets",
            lambda: resource.list(canvas_id),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            self._list_key: dump_model(items),
            "count": len(items),
            "status": "success",
        }


class _AssetUpdateTool(BaseMCPTool):
    _kind: str
    _id_param: str

    def __init__(
        self,
        client: Client,
        *,
        name: str,
        description: str,
        kind: str,
        id_param: str,
    ) -> None:
        super().__init__(name=name, description=description)
        self.client = client
        self._kind = kind
        self._id_param = id_param

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        require_str(kwargs, self._id_param, self.name)
        payload = optional_dict(kwargs, "payload", self.name)
        # Allow legacy x/y as shorthand
        x = optional_number(kwargs, "x", self.name)
        y = optional_number(kwargs, "y", self.name)
        if not payload and x is None and y is None:
            raise MCPToolValidationError(
                "payload (or x/y) parameter is required and must be non-empty",
                self.name,
            )
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        widget_id = require_str(kwargs, self._id_param, self.name)
        payload: dict[str, Any] = dict(
            optional_dict(kwargs, "payload", self.name) or {}
        )
        x = optional_number(kwargs, "x", self.name)
        y = optional_number(kwargs, "y", self.name)
        if x is not None or y is not None:
            location = payload.setdefault("location", {})
            if x is not None:
                location["x"] = x
            if y is not None:
                location["y"] = y
        resource = _resource(self.client, self._kind)
        updated = await run_with_api_error_translation(
            self.name,
            f"Update {self._kind} widget",
            lambda: resource.update(canvas_id, widget_id, payload),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            self._id_param: widget_id,
            "payload": payload,
            f"{self._kind.rstrip('s')}_data": dump_model(updated),
            "status": "success",
        }


class _AssetDeleteTool(BaseMCPTool):
    _kind: str
    _id_param: str

    def __init__(
        self,
        client: Client,
        *,
        name: str,
        description: str,
        kind: str,
        id_param: str,
    ) -> None:
        super().__init__(name=name, description=description)
        self.client = client
        self._kind = kind
        self._id_param = id_param

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        require_str(kwargs, self._id_param, self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        widget_id = require_str(kwargs, self._id_param, self.name)
        resource = _resource(self.client, self._kind)
        await run_with_api_error_translation(
            self.name,
            f"Delete {self._kind} widget",
            lambda: resource.delete(canvas_id, widget_id),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            self._id_param: widget_id,
            "status": "deleted",
        }


__all__ = [
    "_AssetCreateTool",
    "_AssetDeleteTool",
    "_AssetGetTool",
    "_AssetListTool",
    "_AssetUpdateTool",
]
