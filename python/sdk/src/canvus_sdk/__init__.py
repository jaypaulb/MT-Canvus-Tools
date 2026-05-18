"""Canvus SDK — async-first Python client for the Canvus REST API.

Quick start::

    from canvus_sdk import Client

    async with Client(base_url="https://canvus.example.com", api_key="ck_...") as c:
        canvases = await c.canvases.list()

Or use environment-backed configuration (``CANVUS_API_URL``, ``CANVUS_API_KEY``)::

    async with Client.from_env() as c:
        ...

For non-async callers, every resource is mirrored under ``client.sync``::

    client = Client.from_env()
    canvases = client.sync.canvases.list()
    client.sync.close()
"""

from __future__ import annotations

from .client import Client, SyncClient
from .config import Settings
from .errors import (
    APIError,
    AuthError,
    CanvusError,
    NotFoundError,
    RateLimitError,
    ServerError,
    TransportError,
    UnsupportedOperationError,
    ValidationError,
)
from .logging_config import configure_logging
from .models import (
    PDF,
    AccessToken,
    AccessTokenWithSecret,
    Anchor,
    AuditLogEntry,
    AuditLogPage,
    BaseWidget,
    Browser,
    Canvas,
    CanvasFolder,
    CanvasPermissionOverride,
    CanvasPermissions,
    ClientInfo,
    Connector,
    ConnectorEndpoint,
    Group,
    GroupMember,
    Image,
    IPVideo,
    LicenseInfo,
    Location,
    Note,
    RDPConnection,
    ServerConfig,
    ServerInfo,
    Size,
    Table,
    TableCell,
    UploadItem,
    User,
    Video,
    VideoInput,
    VideoOutput,
    Widget,
    Workspace,
)

__version__ = "0.1.0"

__all__ = [
    "PDF",
    "APIError",
    "AccessToken",
    "AccessTokenWithSecret",
    "Anchor",
    "AuditLogEntry",
    "AuditLogPage",
    "AuthError",
    "BaseWidget",
    "Browser",
    "Canvas",
    "CanvasFolder",
    "CanvasPermissionOverride",
    "CanvasPermissions",
    "CanvusError",
    "Client",
    "ClientInfo",
    "Connector",
    "ConnectorEndpoint",
    "Group",
    "GroupMember",
    "IPVideo",
    "Image",
    "LicenseInfo",
    "Location",
    "NotFoundError",
    "Note",
    "RDPConnection",
    "RateLimitError",
    "ServerConfig",
    "ServerError",
    "ServerInfo",
    "Settings",
    "Size",
    "SyncClient",
    "Table",
    "TableCell",
    "TransportError",
    "UnsupportedOperationError",
    "UploadItem",
    "User",
    "ValidationError",
    "Video",
    "VideoInput",
    "VideoOutput",
    "Widget",
    "Workspace",
    "__version__",
    "configure_logging",
]
