"""Video widget MCP tools (see :mod:`_asset_tools` for the shared shape)."""

from __future__ import annotations

from canvus_sdk import Client

from ._asset_tools import (
    _AssetCreateTool,
    _AssetDeleteTool,
    _AssetGetTool,
    _AssetListTool,
    _AssetUpdateTool,
)


class VideoCreateTool(_AssetCreateTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            client,
            name="video_create",
            description="Create a new video element on a canvas with file upload.",
            kind="videos",
            id_key="video_id",
        )


class VideoGetTool(_AssetGetTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            client,
            name="video_get",
            description="Retrieve detailed information about a specific video by ID.",
            kind="videos",
            id_param="video_id",
        )


class VideoListTool(_AssetListTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            client,
            name="video_list",
            description="Retrieve a list of all videos on a specific canvas.",
            kind="videos",
            list_key="videos",
        )


class VideoUpdateTool(_AssetUpdateTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            client,
            name="video_update",
            description="Update video positioning or metadata.",
            kind="videos",
            id_param="video_id",
        )


class VideoDeleteTool(_AssetDeleteTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            client,
            name="video_delete",
            description="Delete a specific video from a canvas.",
            kind="videos",
            id_param="video_id",
        )


__all__ = [
    "VideoCreateTool",
    "VideoDeleteTool",
    "VideoGetTool",
    "VideoListTool",
    "VideoUpdateTool",
]
