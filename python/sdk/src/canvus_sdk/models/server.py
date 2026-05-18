"""Server-info, configuration, license, client/workspace, and asset models."""

from __future__ import annotations

from typing import Any

from pydantic import Field

from ._base import CanvusModel
from .common import Location, Size, ViewRectangle


class ServerInfo(CanvusModel):
    """Response from ``GET /server-info``."""

    version: str | None = None
    api: list[str] = Field(default_factory=list)
    server_id: str | None = None
    go: str | None = None


class ServerConfig(CanvusModel):
    """Response from ``GET /server-config``.

    The Canvus server has historically returned this as a nested object;
    the API docs describe it as a flat ``[]ConfigElement`` array. The legacy
    SDK and this migrated SDK decode the nested form. Use
    :attr:`extra_fields` or ``model.model_dump()`` to access any
    forward-compatible additions.
    """

    server_name: str | None = None
    features: dict[str, Any] = Field(default_factory=dict)
    auth: dict[str, Any] = Field(default_factory=dict)
    access: str | None = None
    external_url: str | None = None
    email: dict[str, Any] = Field(default_factory=dict)
    authentication: dict[str, Any] = Field(default_factory=dict)


class LicenseInfo(CanvusModel):
    """Response from ``GET /license``.

    Per live-server verification (v1.2), the response contains:
    - ``edition``, ``has_expired``, ``is_valid``, ``max_clients``,
      ``seat_model``, ``type`` (all present).
    - Legacy fields ``status``, ``expiry_date``, ``features``, ``license_key``,
      ``max_canvases``, ``max_users`` are not returned.
    """

    edition: str = ""
    has_expired: bool = False
    is_valid: bool = False
    max_clients: int = -1
    seat_model: str = ""
    type: str = ""


class LicenseActivationRequest(CanvusModel):
    """Response from ``GET /license/request`` — an offline activation blob."""

    request: str | None = None


class MipmapInfo(CanvusModel):
    """Response from ``GET /mipmaps/{hash}``."""

    public_hash_hex: str | None = None
    canvas_id: str | None = None
    levels: list[int] = Field(default_factory=list)
    format: str | None = None
    width: int | None = None
    height: int | None = None


class ClientInfo(CanvusModel):
    """A Canvus desktop client registered against the server."""

    id: str
    name: str | None = None
    state: str | None = None
    version: str | None = None
    address: str | None = None
    created_at: str | None = None
    last_seen: str | None = None


class Workspace(CanvusModel):
    """A workspace on a Canvus desktop client.

    Note:
        Per spec, ``id`` is a UUID string. Some legacy code paths key
        workspaces by integer ``index`` — :attr:`index` is preserved for
        backwards compatibility but new code should use :attr:`id`.
    """

    id: str | None = None
    index: int | None = None
    canvas_id: str | None = None
    info_panel_visible: bool = True
    location: Location | None = None
    pinned: bool = False
    server_id: str | None = None
    size: Size | None = None
    state: str = "normal"
    user: str | None = None
    view_rectangle: ViewRectangle | None = None
    workspace_name: str | None = None
    workspace_state: str | None = None


class VideoOutput(CanvusModel):
    """A client-scoped video output."""

    id: str
    name: str | None = None
    source: str | None = None
    enabled: bool = True
    resolution: str | None = None
    refresh_rate: int | None = None
    config: dict[str, Any] = Field(default_factory=dict)


__all__ = [
    "ClientInfo",
    "LicenseActivationRequest",
    "LicenseInfo",
    "MipmapInfo",
    "ServerConfig",
    "ServerInfo",
    "VideoOutput",
    "Workspace",
]
