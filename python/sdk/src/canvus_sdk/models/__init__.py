"""Pydantic v2 models for the Canvus API.

Models intentionally mirror the wire format with one practical concession:
all field names are exposed in their canonical Python ``snake_case`` form,
with hyphenated wire keys (where the API uses them) declared via ``alias``.

To serialise back to the wire format, use ``model.model_dump(by_alias=True)``.
"""

from __future__ import annotations

from .audit import AuditLogEntry, AuditLogPage
from .auth import AccessToken, AccessTokenWithSecret, LoginResponse
from .canvases import (
    Canvas,
    CanvasBackground,
    CanvasFolder,
    CanvasPermissionOverride,
    CanvasPermissions,
    ColorPresets,
    FolderGroupPermission,
    FolderPermissions,
    FolderUserPermission,
)
from .common import GridSize, Location, RelativeLocation, Size, ViewRectangle
from .server import (
    ClientInfo,
    LicenseActivationRequest,
    LicenseInfo,
    MipmapInfo,
    ServerConfig,
    ServerInfo,
    VideoOutput,
    Workspace,
)
from .users import Group, GroupMember, User
from .widgets import (
    PDF,
    Anchor,
    BaseWidget,
    Browser,
    Connector,
    ConnectorEndpoint,
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

__all__ = [
    "PDF",
    "AccessToken",
    "AccessTokenWithSecret",
    "Anchor",
    "AuditLogEntry",
    "AuditLogPage",
    "BaseWidget",
    "Browser",
    "Canvas",
    "CanvasBackground",
    "CanvasFolder",
    "CanvasPermissionOverride",
    "CanvasPermissions",
    "ClientInfo",
    "ColorPresets",
    "Connector",
    "ConnectorEndpoint",
    "FolderGroupPermission",
    "FolderPermissions",
    "FolderUserPermission",
    "GridSize",
    "Group",
    "GroupMember",
    "IPVideo",
    "Image",
    "LicenseActivationRequest",
    "LicenseInfo",
    "Location",
    "LoginResponse",
    "MipmapInfo",
    "Note",
    "RDPConnection",
    "RelativeLocation",
    "ServerConfig",
    "ServerInfo",
    "Size",
    "Table",
    "TableCell",
    "UploadItem",
    "User",
    "Video",
    "VideoInput",
    "VideoOutput",
    "ViewRectangle",
    "Widget",
    "Workspace",
]
