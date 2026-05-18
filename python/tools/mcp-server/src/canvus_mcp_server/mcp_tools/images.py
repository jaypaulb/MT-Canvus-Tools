"""Image widget MCP tools.

Per the Phase 4c audit, image / video / PDF create/get/list/update/delete
share an identical shape — the implementation lives in :mod:`_asset_tools`
and this file binds it to image-specific names and descriptions.
"""

from __future__ import annotations

from canvus_sdk import Client

from ._asset_tools import (
    _AssetCreateTool,
    _AssetDeleteTool,
    _AssetGetTool,
    _AssetListTool,
    _AssetUpdateTool,
)


class ImageCreateTool(_AssetCreateTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            client,
            name="image_create",
            description="Create a new image element on a canvas with file upload.",
            kind="images",
            id_key="image_id",
        )


class ImageGetTool(_AssetGetTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            client,
            name="image_get",
            description="Retrieve detailed information about a specific image by ID.",
            kind="images",
            id_param="image_id",
        )


class ImageListTool(_AssetListTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            client,
            name="image_list",
            description="Retrieve a list of all images on a specific canvas.",
            kind="images",
            list_key="images",
        )


class ImageUpdateTool(_AssetUpdateTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            client,
            name="image_update",
            description="Update image positioning or metadata.",
            kind="images",
            id_param="image_id",
        )


class ImageDeleteTool(_AssetDeleteTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            client,
            name="image_delete",
            description="Delete a specific image from a canvas.",
            kind="images",
            id_param="image_id",
        )


__all__ = [
    "ImageCreateTool",
    "ImageDeleteTool",
    "ImageGetTool",
    "ImageListTool",
    "ImageUpdateTool",
]
