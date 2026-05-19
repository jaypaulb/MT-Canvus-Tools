"""Canvas, folder, and canvas-permission models."""

from __future__ import annotations

from typing import Any

from pydantic import Field

from ._base import CanvusModel


class Canvas(CanvusModel):
    """A Canvus canvas.

    Reflects the response shape documented in
    ``docs/api-reference/endpoints/canvases.md``.
    """

    id: str
    name: str
    folder_id: str | None = None
    access: str | None = None
    asset_size: int = 0
    in_trash: bool = False
    mode: str = "normal"
    state: str | None = None
    preview_hash: str | None = None
    description: str | None = None
    link_permission: str | None = None
    owner_id: str | None = None
    created_at: str | None = None
    modified_at: str | None = None


class CanvasFolder(CanvusModel):
    """A folder containing canvases."""

    id: str
    name: str
    folder_id: str | None = None
    access: str | None = None
    in_trash: bool = False
    state: str | None = None
    created_at: str | None = None
    modified_at: str | None = None


class CanvasPermissionOverride(CanvusModel):
    """A single subject-permission override on a canvas.

    Per spec, ``subject_type`` is one of ``user`` or ``group``.
    """

    subject_type: str = Field(..., description="`user` or `group`.")
    subject_id: str
    permission: str = Field(..., description="`view`, `edit`, or `owner`.")


class CanvasPermissions(CanvusModel):
    """Permission set for a canvas.

    Per spec the wire body shape is ``{link-permission, permission-overrides[]}``.
    Both are exposed here in snake_case Python form.
    """

    link_permission: str = Field(
        default="none",
        description="`none`, `view`, or `edit`.",
    )
    permission_overrides: list[CanvasPermissionOverride] = Field(default_factory=list)


class CanvasBackground(CanvusModel):
    """Canvas background metadata."""

    background_type: str | None = None
    background_color: str | None = None
    hash: str | None = None


class ColorPresets(CanvusModel):
    """The set of color presets defined on a canvas."""

    presets: dict[str, Any] = Field(default_factory=dict)


class FolderUserPermission(CanvusModel):
    """A single user permission entry on a canvas folder."""

    id: int
    permission: str
    inherited: bool = False


class FolderGroupPermission(CanvusModel):
    """A single group permission entry on a canvas folder."""

    id: int
    permission: str
    inherited: bool = False


class FolderPermissions(CanvusModel):
    """Permission set for a canvas folder."""

    editors_can_share: bool = False
    users: list[FolderUserPermission] = Field(default_factory=list)
    groups: list[FolderGroupPermission] = Field(default_factory=list)


__all__ = [
    "Canvas",
    "CanvasBackground",
    "CanvasFolder",
    "CanvasPermissionOverride",
    "CanvasPermissions",
    "ColorPresets",
    "FolderGroupPermission",
    "FolderPermissions",
    "FolderUserPermission",
]
