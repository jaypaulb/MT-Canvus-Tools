"""Example 03: full CRUD lifecycle for a note widget.

Runs six ordered steps against `CANVUS_CANVAS_ID`. On any step's failure the
program logs the error and STOPS — subsequent steps would only confuse the
debugging story. The lifecycle:

    1. Create a note at (100, 100).
    2. Print the new widget id.
    3. PATCH the note's text and background colour.
    4. Print the updated widget.
    5. DELETE the note.
    6. Verify the delete by GET — expecting :class:`NotFoundError`.
"""

from __future__ import annotations

import asyncio
import sys
from datetime import UTC, datetime

import structlog
from pydantic import Field, ValidationError
from pydantic_settings import BaseSettings, SettingsConfigDict

from canvus_sdk import APIError, Client, NotFoundError, configure_logging

logger = structlog.get_logger(__name__)


class WidgetCRUDSettings(BaseSettings):
    """Environment configuration for the widget CRUD example."""

    model_config = SettingsConfigDict(
        env_prefix="CANVUS_",
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    api_url: str = Field(...)
    api_key: str = Field(...)
    canvas_id: str = Field(..., description="Canvas to operate on.")


async def main() -> int:
    """Entry point."""
    configure_logging()
    try:
        settings = WidgetCRUDSettings()  # type: ignore[call-arg]
    except ValidationError as e:
        logger.error("config invalid", errors=e.errors())
        return 2

    async with Client(settings.api_url, settings.api_key) as client:
        canvas_id = settings.canvas_id

        # Step 1: create
        now_iso = datetime.now(UTC).isoformat()
        try:
            created = await client.widgets.notes.create(
                canvas_id,
                {
                    "text": f"hello from python @ {now_iso}",
                    "location": {"x": 100.0, "y": 100.0},
                    "size": {"width": 250.0, "height": 200.0},
                },
            )
        except APIError as e:
            logger.error(
                "step1_create failed",
                status_code=e.status_code,
                response_body=e.response_body,
            )
            return 1
        note_id = created.id
        if note_id is None:
            logger.error("step1_create returned no id", note=created.model_dump())
            return 1
        logger.info("step1_create ok", note_id=note_id, text=created.text)

        # Step 2: print
        print(f"created note id: {note_id}")

        # Step 3: patch
        try:
            updated = await client.widgets.notes.update(
                canvas_id,
                note_id,
                {
                    "text": f"updated from python @ {datetime.now(UTC).isoformat()}",
                    "background_color": "#3aaa34ff",
                },
            )
        except APIError as e:
            logger.error(
                "step3_patch failed",
                status_code=e.status_code,
                response_body=e.response_body,
            )
            return 1
        logger.info(
            "step3_patch ok",
            note_id=updated.id,
            text=updated.text,
            background_color=updated.background_color,
        )

        # Step 4: print updated
        print(f"updated note: text={updated.text!r} bg={updated.background_color}")

        # Step 5: delete
        try:
            await client.widgets.notes.delete(canvas_id, note_id)
        except APIError as e:
            logger.error(
                "step5_delete failed",
                status_code=e.status_code,
                response_body=e.response_body,
            )
            return 1
        logger.info("step5_delete ok", note_id=note_id)

        # Step 6: verify by GET → expect NotFoundError
        try:
            ghost = await client.widgets.notes.get(canvas_id, note_id)
            logger.error(
                "step6_verify FAILED — note still exists after delete",
                note=ghost.model_dump(),
            )
            return 1
        except NotFoundError:
            logger.info("step6_verify ok", note_id=note_id, status_code=404)
            print(f"verified deletion of note {note_id} (NotFoundError as expected)")
        except APIError as e:
            logger.error(
                "step6_verify got non-404 error",
                status_code=e.status_code,
                response_body=e.response_body,
            )
            return 1

    return 0


if __name__ == "__main__":
    sys.exit(asyncio.run(main()))
