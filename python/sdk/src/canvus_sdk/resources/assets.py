"""Asset download endpoints: ``/assets/{hash}`` and the mipmap surface."""

from __future__ import annotations

from ..models import MipmapInfo
from ._base import Resource


class AssetsResource(Resource):
    """Operations on asset hashes (image / video / pdf binaries)."""

    async def download_by_hash(self, asset_hash: str, canvas_id: str) -> bytes:
        """Download an asset by its public hash.

        Per spec the request requires a ``canvas-id`` header naming any
        canvas that contains the asset.
        """
        return await self._transport.request_bytes(
            "GET",
            f"assets/{asset_hash}",
            headers={"canvas-id": canvas_id},
        )

    async def get_mipmap_info(self, asset_hash: str) -> MipmapInfo:
        """Get mipmap metadata for an image asset."""
        data = await self._transport.request("GET", f"mipmaps/{asset_hash}")
        return self._parse(MipmapInfo, data)

    async def download_mipmap_level(self, asset_hash: str, level: int) -> bytes:
        """Download a single mipmap level (binary)."""
        return await self._transport.request_bytes(
            "GET", f"mipmaps/{asset_hash}/{level}"
        )


__all__ = ["AssetsResource"]
