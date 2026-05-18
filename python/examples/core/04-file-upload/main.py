"""Example 04: upload a PNG and place it as an image widget.

If ``CANVUS_IMAGE_PATH`` is unset (or points at a non-existent path) the
example synthesises a minimal 256x256 solid-colour PNG using only the
standard library, so the example is self-contained.

Steps:
    1. Acquire image bytes (from disk or freshly generated).
    2. Upload via :meth:`Client.widgets.images.upload` (multipart POST).
    3. PATCH the new widget to position (100, 200) at scale 0.5.
    4. Print the widget id and on-screen position.
    5. Unless ``CANVUS_KEEP_WIDGET=1``, delete it (cleanup).
"""

from __future__ import annotations

import asyncio
import os
import struct
import sys
import zlib
from pathlib import Path

import structlog
from pydantic import Field, ValidationError
from pydantic_settings import BaseSettings, SettingsConfigDict

from canvus_sdk import APIError, Client, configure_logging

logger = structlog.get_logger(__name__)


class UploadSettings(BaseSettings):
    """Environment configuration for the file-upload example."""

    model_config = SettingsConfigDict(
        env_prefix="CANVUS_",
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    api_url: str = Field(...)
    api_key: str = Field(...)
    canvas_id: str = Field(..., description="Canvas to upload into.")
    image_path: str | None = Field(default=None, description="Optional local PNG path.")
    keep_widget: bool = Field(
        default=False,
        description="If true, do not delete the widget at the end.",
    )


def _generate_solid_png(
    width: int = 256,
    height: int = 256,
    rgb: tuple[int, int, int] = (40, 130, 200),
) -> bytes:
    """Synthesise a minimal valid solid-colour PNG using only the stdlib.

    Implements PNG signature + IHDR + IDAT (raw, then zlib-compressed) + IEND.
    """
    sig = b"\x89PNG\r\n\x1a\n"

    def chunk(tag: bytes, data: bytes) -> bytes:
        return (
            struct.pack(">I", len(data))
            + tag
            + data
            + struct.pack(">I", zlib.crc32(tag + data) & 0xFFFFFFFF)
        )

    ihdr = struct.pack(
        ">IIBBBBB",
        width,
        height,
        8,  # bit depth
        2,  # colour type: truecolour RGB
        0,  # compression
        0,  # filter
        0,  # interlace
    )
    # PNG scanlines: one filter byte per row, then 3 bytes per pixel.
    row = b"\x00" + bytes(rgb) * width
    raw = row * height
    idat = zlib.compress(raw, level=6)
    return sig + chunk(b"IHDR", ihdr) + chunk(b"IDAT", idat) + chunk(b"IEND", b"")


def _load_image_bytes(image_path: str | None) -> tuple[bytes, str]:
    """Return (bytes, filename) — load from disk or generate."""
    if image_path:
        path = Path(image_path)
        if path.is_file():
            return path.read_bytes(), path.name
        logger.warning(
            "image_path does not exist; falling back to generated sample",
            image_path=image_path,
        )
    return _generate_solid_png(), "sample.png"


async def main() -> int:
    """Entry point."""
    configure_logging()
    try:
        settings = UploadSettings()  # type: ignore[call-arg]
    except ValidationError as e:
        logger.error("config invalid", errors=e.errors())
        return 2

    img_bytes, filename = _load_image_bytes(settings.image_path)
    logger.info("loaded image", filename=filename, byte_count=len(img_bytes))

    async with Client(settings.api_url, settings.api_key) as client:
        canvas_id = settings.canvas_id

        # Step 2: upload
        try:
            widget = await client.widgets.images.upload(
                canvas_id,
                img_bytes,
                filename=filename,
                content_type="image/png",
            )
        except APIError as e:
            logger.error(
                "upload failed",
                status_code=e.status_code,
                response_body=e.response_body,
            )
            return 1
        widget_id = widget.id
        if widget_id is None:
            logger.error("upload returned no widget id", widget=widget.model_dump())
            return 1
        logger.info(
            "uploaded",
            widget_id=widget_id,
            asset_hash=widget.hash,
            file_size=widget.file_size,
        )

        # Step 3: reposition
        try:
            patched = await client.widgets.images.update(
                canvas_id,
                widget_id,
                {
                    "location": {"x": 100.0, "y": 200.0},
                    "scale": 0.5,
                },
            )
        except APIError as e:
            logger.error(
                "reposition failed",
                status_code=e.status_code,
                response_body=e.response_body,
            )
            return 1
        logger.info(
            "repositioned",
            widget_id=patched.id,
            location=patched.location,
            scale=patched.scale,
        )
        print(
            f"image widget {patched.id} at "
            f"({patched.location.get('x')}, {patched.location.get('y')}) "
            f"scale={patched.scale}",
        )

        # Step 5: cleanup
        keep = settings.keep_widget or os.environ.get("CANVUS_KEEP_WIDGET") == "1"
        if keep:
            logger.info("cleanup skipped (CANVUS_KEEP_WIDGET=1)", widget_id=widget_id)
        else:
            try:
                await client.widgets.images.delete(canvas_id, widget_id)
                logger.info("cleanup deleted", widget_id=widget_id)
            except APIError as e:
                logger.warning(
                    "cleanup delete failed (widget left on canvas)",
                    status_code=e.status_code,
                    widget_id=widget_id,
                )

    return 0


if __name__ == "__main__":
    sys.exit(asyncio.run(main()))
