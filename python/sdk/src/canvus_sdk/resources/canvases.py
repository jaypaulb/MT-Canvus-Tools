"""Canvas + folder + canvas-meta endpoints.

Covers every endpoint under:

- ``/canvases`` (CRUD, move, copy, save/restore demo, preview)
- ``/canvases/{id}/background``
- ``/canvases/{id}/color-presets``
- ``/canvases/{id}/permissions``
- ``/canvas-folders`` (CRUD, move, copy, permissions)
"""

from __future__ import annotations

from typing import Any

from ..models import (
    Canvas,
    CanvasBackground,
    CanvasFolder,
    CanvasPermissions,
    ColorPresets,
)
from ._base import Resource


class CanvasesResource(Resource):
    """Operations on canvases and their immediate sub-resources."""

    # ---- canvas CRUD --------------------------------------------------------

    async def list(self, *, params: dict[str, Any] | None = None) -> list[Canvas]:
        """List canvases. ``params`` is forwarded as query string."""
        data = await self._transport.request("GET", "canvases", params=params)
        return self._parse_list(Canvas, data)

    async def get(self, canvas_id: str) -> Canvas:
        """Get a single canvas."""
        data = await self._transport.request("GET", f"canvases/{canvas_id}")
        return self._parse(Canvas, data)

    async def create(self, payload: dict[str, Any]) -> Canvas:
        """Create a new canvas."""
        data = await self._transport.request("POST", "canvases", json_body=payload)
        return self._parse(Canvas, data)

    async def update(self, canvas_id: str, payload: dict[str, Any]) -> Canvas:
        """Update canvas properties (PATCH)."""
        data = await self._transport.request(
            "PATCH", f"canvases/{canvas_id}", json_body=payload
        )
        return self._parse(Canvas, data)

    async def delete(self, canvas_id: str) -> None:
        """Delete a canvas."""
        await self._transport.request("DELETE", f"canvases/{canvas_id}")

    async def move(self, canvas_id: str, folder_id: str) -> Canvas:
        """Move a canvas to a different folder."""
        data = await self._transport.request(
            "POST",
            f"canvases/{canvas_id}/move",
            json_body={"folder_id": folder_id},
        )
        return self._parse(Canvas, data)

    async def copy(self, canvas_id: str, payload: dict[str, Any]) -> Canvas:
        """Copy a canvas. Payload includes ``name`` and optionally ``folder_id``."""
        data = await self._transport.request(
            "POST",
            f"canvases/{canvas_id}/copy",
            json_body=payload,
        )
        return self._parse(Canvas, data)

    async def save_demo_state(self, canvas_id: str) -> Canvas:
        """Save the current state of a demo canvas."""
        data = await self._transport.request("POST", f"canvases/{canvas_id}/save")
        return self._parse(Canvas, data)

    async def restore_demo_state(self, canvas_id: str) -> Canvas:
        """Restore a demo canvas to its previously saved state."""
        data = await self._transport.request("POST", f"canvases/{canvas_id}/restore")
        return self._parse(Canvas, data)

    async def get_preview(self, canvas_id: str) -> bytes:
        """Get a preview thumbnail of the canvas as raw image bytes."""
        return await self._transport.request_bytes(
            "GET", f"canvases/{canvas_id}/preview"
        )

    # ---- background ---------------------------------------------------------

    async def get_background(self, canvas_id: str) -> CanvasBackground:
        """Get a canvas's background configuration."""
        data = await self._transport.request("GET", f"canvases/{canvas_id}/background")
        return self._parse(CanvasBackground, data)

    async def set_background(
        self, canvas_id: str, payload: dict[str, Any]
    ) -> CanvasBackground:
        """Update a canvas's background configuration (PATCH)."""
        data = await self._transport.request(
            "PATCH",
            f"canvases/{canvas_id}/background",
            json_body=payload,
        )
        return self._parse(CanvasBackground, data)

    async def upload_background_image(
        self,
        canvas_id: str,
        file_bytes: bytes,
        filename: str = "background",
        content_type: str = "application/octet-stream",
    ) -> CanvasBackground:
        """Upload a new background image (multipart POST)."""
        files = {"data": (filename, file_bytes, content_type)}
        data = await self._transport.request(
            "POST",
            f"canvases/{canvas_id}/background",
            files=files,
        )
        return self._parse(CanvasBackground, data)

    # ---- color presets ------------------------------------------------------

    async def get_color_presets(self, canvas_id: str) -> ColorPresets:
        """Get the canvas's color presets (spec path: ``color-presets``)."""
        data = await self._transport.request(
            "GET", f"canvases/{canvas_id}/color-presets"
        )
        # The server returns either an object or a dict; wrap it for callers.
        if isinstance(data, dict):
            return ColorPresets(presets=data)
        return ColorPresets(presets={})

    async def update_color_presets(
        self, canvas_id: str, presets: dict[str, Any]
    ) -> ColorPresets:
        """Update a canvas's color presets (PATCH)."""
        data = await self._transport.request(
            "PATCH",
            f"canvases/{canvas_id}/color-presets",
            json_body=presets,
        )
        if isinstance(data, dict):
            return ColorPresets(presets=data)
        return ColorPresets(presets={})

    # ---- permissions --------------------------------------------------------

    async def get_permissions(self, canvas_id: str) -> CanvasPermissions:
        """Get the canvas permissions block."""
        data = await self._transport.request(
            "GET", f"canvases/{canvas_id}/permissions"
        )
        return self._parse(CanvasPermissions, data)

    async def set_permissions(
        self,
        canvas_id: str,
        permissions: CanvasPermissions | dict[str, Any],
    ) -> CanvasPermissions:
        """Set the canvas permissions block.

        Per spec, the body shape is
        ``{link-permission: str, permission-overrides: [{...}]}``. Accepting a
        typed :class:`CanvusPermissions` (or a dict) avoids the legacy
        ``payload: dict[str, Any]`` drift documented in the coverage matrix.
        """
        if isinstance(permissions, CanvasPermissions):
            body = permissions.model_dump(mode="json")
        else:
            body = permissions
        data = await self._transport.request(
            "POST",
            f"canvases/{canvas_id}/permissions",
            json_body=body,
        )
        return self._parse(CanvasPermissions, data)


class FoldersResource(Resource):
    """Operations on ``/canvas-folders``."""

    async def list(self, *, params: dict[str, Any] | None = None) -> list[CanvasFolder]:
        """List canvas folders."""
        data = await self._transport.request("GET", "canvas-folders", params=params)
        return self._parse_list(CanvasFolder, data)

    async def get(self, folder_id: str) -> CanvasFolder:
        """Get a single folder."""
        data = await self._transport.request("GET", f"canvas-folders/{folder_id}")
        return self._parse(CanvasFolder, data)

    async def create(self, payload: dict[str, Any]) -> CanvasFolder:
        """Create a folder."""
        data = await self._transport.request("POST", "canvas-folders", json_body=payload)
        return self._parse(CanvasFolder, data)

    async def update(self, folder_id: str, payload: dict[str, Any]) -> CanvasFolder:
        """Rename or otherwise update a folder (PATCH)."""
        data = await self._transport.request(
            "PATCH", f"canvas-folders/{folder_id}", json_body=payload
        )
        return self._parse(CanvasFolder, data)

    async def delete(self, folder_id: str) -> None:
        """Delete a folder."""
        await self._transport.request("DELETE", f"canvas-folders/{folder_id}")

    async def delete_children(self, folder_id: str) -> None:
        """Delete every canvas / sub-folder inside a folder."""
        await self._transport.request(
            "DELETE", f"canvas-folders/{folder_id}/children"
        )

    async def move(
        self,
        folder_id: str,
        new_parent_id: str,
        *,
        use_patch: bool = False,
    ) -> CanvasFolder:
        """Move a folder to a new parent.

        Args:
            folder_id: ID of the folder to move.
            new_parent_id: ID of the new parent folder.
            use_patch: If ``True``, use ``PATCH`` rather than the default
                ``POST``. The spec lists both as acceptable.
        """
        method = "PATCH" if use_patch else "POST"
        data = await self._transport.request(
            method,
            f"canvas-folders/{folder_id}/move",
            json_body={"folder_id": new_parent_id},
        )
        return self._parse(CanvasFolder, data)

    async def copy(
        self,
        folder_id: str,
        payload: dict[str, Any],
        *,
        use_patch: bool = False,
    ) -> CanvasFolder:
        """Copy a folder. ``use_patch`` swaps the verb (POST default; PATCH alt)."""
        method = "PATCH" if use_patch else "POST"
        data = await self._transport.request(
            method,
            f"canvas-folders/{folder_id}/copy",
            json_body=payload,
        )
        return self._parse(CanvasFolder, data)

    async def get_permissions(self, folder_id: str) -> dict[str, Any]:
        """Get a folder's permissions (returned as a raw dict)."""
        data = await self._transport.request(
            "GET", f"canvas-folders/{folder_id}/permissions"
        )
        return data if isinstance(data, dict) else {}

    async def set_permissions(
        self, folder_id: str, payload: dict[str, Any]
    ) -> dict[str, Any]:
        """Set a folder's permissions."""
        data = await self._transport.request(
            "POST",
            f"canvas-folders/{folder_id}/permissions",
            json_body=payload,
        )
        return data if isinstance(data, dict) else {}


__all__ = ["CanvasesResource", "FoldersResource"]
