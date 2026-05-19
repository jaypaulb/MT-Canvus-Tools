"""Canvas + folder + canvas-meta endpoints.

Covers every endpoint under:

- ``/canvases`` (CRUD, move, copy, save/restore demo, preview)
- ``/canvases/{id}/background``
- ``/canvases/{id}/color-presets``
- ``/canvases/{id}/permissions``
- ``/canvas-folders`` (CRUD, move, copy, permissions)
"""

from __future__ import annotations

import builtins
from collections.abc import AsyncIterator
from typing import Any

from ..errors import ValidationError
from ..models import (
    Canvas,
    CanvasBackground,
    CanvasFolder,
    CanvasPermissions,
    ColorPresets,
    FolderPermissions,
    User,
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

    async def trash(self, canvas_id: str) -> Canvas:
        """Move a canvas to the current user's trash folder.

        Phase 4b §4.2 #5: mirrors Go's ``canvases.go:95 TrashCanvas``. The
        Canvus server represents each user's trash as the synthetic folder ID
        ``trash.{user_id}`` (integer user ID per VERIFIED-CORRECTIONS §6).
        This helper looks up the current user via ``GET /users/current`` and
        then PATCHes the canvas's ``folder_id``.

        Raises:
            ValidationError: ``GET /users/current`` returned a user with no
                integer ID — the session is not associated with a real user
                (e.g. expired token).
        """
        current = await self._transport.request("GET", "users/current")
        user = User.model_validate(current)
        if user.id is None:
            raise ValidationError(
                "trash: GET /users/current did not return a numeric user id; "
                "cannot derive the trash folder id.",
            )
        # Per Go's TrashCanvas implementation, the move is performed via the
        # POST /move endpoint (not a bare PATCH) — keeps server-side hooks /
        # audit-log entries consistent across SDKs.
        return await self.move(canvas_id, f"trash.{user.id}")

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

    # ---- per-preset decomposition (Phase 4b §4.2 #8) -----------------------
    # The Canvus v1.2 server does NOT expose ``/color-presets/{name}``
    # endpoints. These five helpers decompose / recompose the bulk presets
    # object client-side so callers can ergonomically work with one preset at
    # a time. They all round-trip through GET + PATCH of the bulk endpoint.

    async def list_color_presets(self, canvas_id: str) -> builtins.list[str]:
        """Return the names of every color preset defined on the canvas."""
        presets = await self.get_color_presets(canvas_id)
        return sorted(presets.presets.keys())

    async def get_color_preset(
        self, canvas_id: str, name: str
    ) -> dict[str, Any]:
        """Return one color preset as the raw wire dict."""
        presets = await self.get_color_presets(canvas_id)
        if name not in presets.presets:
            raise KeyError(
                f"color preset {name!r} not found on canvas {canvas_id!r}; "
                f"defined presets: {sorted(presets.presets.keys())}",
            )
        value = presets.presets[name]
        if not isinstance(value, dict):
            return {"value": value}
        return dict(value)

    async def create_color_preset(
        self,
        canvas_id: str,
        name: str,
        preset: dict[str, Any],
    ) -> dict[str, Any]:
        """Add a new color preset. Raises if ``name`` already exists."""
        presets = await self.get_color_presets(canvas_id)
        if name in presets.presets:
            raise ValidationError(
                f"color preset {name!r} already exists on canvas {canvas_id!r}; "
                "use update_color_preset() to replace it.",
            )
        merged = dict(presets.presets)
        merged[name] = preset
        await self.update_color_presets(canvas_id, merged)
        return dict(preset)

    async def update_color_preset(
        self,
        canvas_id: str,
        name: str,
        preset: dict[str, Any],
    ) -> dict[str, Any]:
        """Replace one preset's value (raises if absent)."""
        presets = await self.get_color_presets(canvas_id)
        if name not in presets.presets:
            raise KeyError(
                f"color preset {name!r} not found on canvas {canvas_id!r}",
            )
        merged = dict(presets.presets)
        merged[name] = preset
        await self.update_color_presets(canvas_id, merged)
        return dict(preset)

    async def delete_color_preset(self, canvas_id: str, name: str) -> None:
        """Remove one preset by name (no-op if absent)."""
        presets = await self.get_color_presets(canvas_id)
        if name not in presets.presets:
            return
        merged = dict(presets.presets)
        del merged[name]
        await self.update_color_presets(canvas_id, merged)

    # ---- permissions --------------------------------------------------------

    async def get_permissions(self, canvas_id: str) -> CanvasPermissions:
        """Get the canvas permissions block."""
        data = await self._transport.request(
            "GET", f"canvases/{canvas_id}/permissions"
        )
        return self._parse(CanvasPermissions, data)

    # ---- subscribe helpers (Phase 4b §4.2 #13) -----------------------------

    def subscribe(
        self,
        *,
        params: dict[str, Any] | None = None,
    ) -> AsyncIterator[Canvas]:
        """Subscribe to ``/canvases?subscribe=true``."""
        return self._typed_subscribe(Canvas, "canvases", params=params)

    def subscribe_one(
        self,
        canvas_id: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> AsyncIterator[Canvas]:
        """Subscribe to a single canvas."""
        return self._typed_subscribe(
            Canvas, f"canvases/{canvas_id}", params=params
        )

    def subscribe_permissions(
        self,
        canvas_id: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> AsyncIterator[CanvasPermissions]:
        """Subscribe to a canvas's permissions block.

        Live-verified 2026-05-19 against ``dev-mtcs.multitaction.com``
        (parity-matrix §5.5): the server emits an initial snapshot
        followed by a change event each time the canvas permissions are
        POSTed.
        """
        return self._typed_subscribe(
            CanvasPermissions,
            f"canvases/{canvas_id}/permissions",
            params=params,
        )

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

    async def trash(self, folder_id: str) -> CanvasFolder:
        """Move a folder to the current user's trash folder.

        Phase 4b §4.2 #6: mirrors Go's ``folders.go:135 TrashFolder``. The
        synthetic destination folder ID is ``trash.{user_id}`` (integer user
        id per VERIFIED-CORRECTIONS §6).

        Raises:
            ValidationError: ``GET /users/current`` returned a user with no
                integer ID.
        """
        current = await self._transport.request("GET", "users/current")
        user = User.model_validate(current)
        if user.id is None:
            raise ValidationError(
                "trash: GET /users/current did not return a numeric user id; "
                "cannot derive the trash folder id.",
            )
        return await self.move(folder_id, f"trash.{user.id}")

    # ---- subscribe helpers (Phase 4b §4.2 #13) -----------------------------

    def subscribe(
        self,
        *,
        params: dict[str, Any] | None = None,
    ) -> AsyncIterator[CanvasFolder]:
        """Subscribe to ``/canvas-folders?subscribe=true``."""
        return self._typed_subscribe(CanvasFolder, "canvas-folders", params=params)

    def subscribe_one(
        self,
        folder_id: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> AsyncIterator[CanvasFolder]:
        """Subscribe to a single folder."""
        return self._typed_subscribe(
            CanvasFolder, f"canvas-folders/{folder_id}", params=params
        )

    def subscribe_permissions(
        self,
        folder_id: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> AsyncIterator[FolderPermissions]:
        """Subscribe to a folder's permissions block."""
        return self._typed_subscribe(
            FolderPermissions,
            f"canvas-folders/{folder_id}/permissions",
            params=params,
        )

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
