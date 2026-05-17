"""Endpoint-grouped resource classes for the Canvus SDK.

Each resource class wraps a small slice of the Canvus REST surface area and
is attached to :class:`canvus_sdk.Client` as a public attribute.
"""

from __future__ import annotations

from .assets import AssetsResource
from .auth import AuthResource
from .canvases import CanvasesResource, FoldersResource
from .server import ServerResource
from .users import GroupsResource, UsersResource
from .widgets import WidgetsResource

__all__ = [
    "AssetsResource",
    "AuthResource",
    "CanvasesResource",
    "FoldersResource",
    "GroupsResource",
    "ServerResource",
    "UsersResource",
    "WidgetsResource",
]
