"""Example 08: clone a widget from one canvas to another.

Demonstrates the SDK's :meth:`Client.widgets.clone` helper, which implements
the contract added in API changelog §1: cross-canvas cloning is a behaviour
of the *per-type* create endpoints, not a separate ``/widgets/clone`` URL.

Steps:
    1. Fetch the source widget to discover its ``widget_type``.
    2. Call ``client.widgets.clone(dest_canvas_id, src_canvas_id,
       src_widget_id, widget_type)``.
    3. Print the new widget id on the destination canvas.
    4. If ``CANVUS_CLEANUP=1`` is set, pause 2 s then delete the clone.
"""

from __future__ import annotations

import asyncio
import os
import sys

import structlog
from pydantic import Field, ValidationError
from pydantic_settings import BaseSettings, SettingsConfigDict

from canvus_sdk import APIError, Client, configure_logging

logger = structlog.get_logger(__name__)

# Mapping from server-supplied widget_type to the per-type DELETE path used
# for cleanup. Matches the same set the SDK's clone helper supports.
WIDGET_TYPE_TO_PATH: dict[str, str] = {
    "Note": "notes",
    "Image": "images",
    "Video": "videos",
    "Pdf": "pdfs",
    "Browser": "browsers",
    "Anchor": "anchors",
    "Table": "tables",
}


class CloneSettings(BaseSettings):
    """Environment configuration for the cross-canvas clone example."""

    model_config = SettingsConfigDict(
        env_prefix="CANVUS_",
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    api_url: str = Field(...)
    api_key: str = Field(...)
    canvas_id: str = Field(..., description="Source canvas id.")
    dest_canvas_id: str = Field(..., description="Destination canvas id.")
    source_widget_id: str = Field(..., description="Widget to clone from source.")


async def _delete_clone(
    client: Client,
    dest_canvas_id: str,
    new_widget_id: str,
    widget_type: str,
) -> None:
    """Best-effort cleanup of the cloned widget."""
    path = WIDGET_TYPE_TO_PATH.get(widget_type)
    if path is None:
        logger.warning(
            "cleanup skipped — unknown widget_type",
            widget_type=widget_type,
        )
        return
    # Reach into the appropriate per-type resource by attribute name. The
    # SDK exposes them on `client.widgets` (notes, images, videos, …).
    resource = getattr(client.widgets, path, None)
    if resource is None or not hasattr(resource, "delete"):
        logger.warning("cleanup skipped — no resource for path", path=path)
        return
    try:
        await resource.delete(dest_canvas_id, new_widget_id)
        logger.info("cleanup deleted", new_widget_id=new_widget_id)
    except APIError as e:
        logger.warning(
            "cleanup delete failed",
            status_code=e.status_code,
            new_widget_id=new_widget_id,
        )


async def main() -> int:
    """Entry point."""
    configure_logging()
    try:
        settings = CloneSettings()  # type: ignore[call-arg]
    except ValidationError as e:
        logger.error("config invalid", errors=e.errors())
        return 2

    cleanup = os.environ.get("CANVUS_CLEANUP") == "1"

    async with Client(settings.api_url, settings.api_key) as client:
        # Step 1: fetch source widget to discover its type.
        try:
            src = await client.widgets.get(settings.canvas_id, settings.source_widget_id)
        except APIError as e:
            logger.error(
                "fetch source failed",
                status_code=e.status_code,
                response_body=e.response_body,
            )
            return 1
        widget_type = src.widget_type or ""
        if widget_type not in WIDGET_TYPE_TO_PATH:
            logger.error(
                "source widget type is not cloneable",
                widget_type=widget_type,
                allowed=sorted(WIDGET_TYPE_TO_PATH.keys()),
            )
            return 1
        logger.info(
            "source fetched",
            widget_id=src.id,
            widget_type=widget_type,
        )

        # Step 2: clone via the SDK helper.
        try:
            cloned = await client.widgets.clone(
                settings.dest_canvas_id,
                settings.canvas_id,
                settings.source_widget_id,
                widget_type,
            )
        except APIError as e:
            logger.error(
                "clone failed",
                status_code=e.status_code,
                response_body=e.response_body,
            )
            return 1
        new_id = cloned.get("id") if isinstance(cloned, dict) else None
        logger.info(
            "cloned",
            source_canvas_id=settings.canvas_id,
            dest_canvas_id=settings.dest_canvas_id,
            new_widget_id=new_id,
        )
        print(
            f"cloned widget {settings.source_widget_id} ({widget_type}) "
            f"into canvas {settings.dest_canvas_id} as {new_id}",
        )

        # Step 4: optional cleanup
        if cleanup and isinstance(new_id, str):
            await asyncio.sleep(2.0)
            await _delete_clone(client, settings.dest_canvas_id, new_id, widget_type)

    return 0


if __name__ == "__main__":
    sys.exit(asyncio.run(main()))
