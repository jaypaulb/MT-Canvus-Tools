"""Example 05: subscribe to a canvas's notes endpoint for a fixed duration.

Opens the SDK's NDJSON streaming subscription, prints a one-line summary per
event, and exits cleanly after ``STREAM_DURATION_SECONDS`` (default 30) or on
``SIGINT``.

The Canvus streaming protocol sends:
    - a snapshot of all currently-existing items first;
    - then incremental updates as JSON objects (each newline-terminated);
    - and occasional empty keep-alive lines that the SDK silently drops.
"""

from __future__ import annotations

import asyncio
import contextlib
import os
import signal
import sys
from typing import Any

import structlog
from pydantic import Field, ValidationError
from pydantic_settings import BaseSettings, SettingsConfigDict

from canvus_sdk import APIError, Client, configure_logging

logger = structlog.get_logger(__name__)


class StreamSettings(BaseSettings):
    """Environment configuration for the streaming example."""

    model_config = SettingsConfigDict(
        env_prefix="CANVUS_",
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    api_url: str = Field(...)
    api_key: str = Field(...)
    canvas_id: str = Field(...)


def _stream_duration() -> float:
    raw = os.environ.get("STREAM_DURATION_SECONDS", "30")
    try:
        return float(raw)
    except ValueError:
        logger.warning("STREAM_DURATION_SECONDS invalid; using 30", raw=raw)
        return 30.0


async def _consume(
    client: Client,
    canvas_id: str,
    stop_event: asyncio.Event,
) -> tuple[int, int]:
    """Drain the subscription until ``stop_event`` is set.

    Returns (event_count, batch_count).
    """
    events_total = 0
    batches = 0
    sub = client.widgets.subscribe(canvas_id, widget_type="notes")
    # `subscribe()` returns an async generator directly (it `yield`s); we
    # iterate it under a race with the stop event so SIGINT or the duration
    # timeout cuts in promptly.
    aiter = sub.__aiter__()
    while not stop_event.is_set():
        next_task = asyncio.create_task(aiter.__anext__())
        stop_task = asyncio.create_task(stop_event.wait())
        done, pending = await asyncio.wait(
            {next_task, stop_task},
            return_when=asyncio.FIRST_COMPLETED,
        )
        for p in pending:
            p.cancel()
        if next_task in done:
            try:
                payload: Any = next_task.result()
            except StopAsyncIteration:
                logger.info("stream ended by server")
                break
            except APIError as e:
                logger.error(
                    "stream api error",
                    status_code=e.status_code,
                    response_body=e.response_body,
                )
                break
            batches += 1
            if isinstance(payload, list):
                events_total += len(payload)
                first = payload[0] if payload else None
                logger.info(
                    "batch",
                    batch=batches,
                    count=len(payload),
                    first_id=first.get("id") if isinstance(first, dict) else None,
                    first_text=(
                        first.get("text", "")[:60] if isinstance(first, dict) else None
                    ),
                )
            elif isinstance(payload, dict):
                events_total += 1
                logger.info(
                    "event",
                    batch=batches,
                    id=payload.get("id"),
                    text=payload.get("text", "")[:60],
                )
            else:
                logger.debug("unknown payload type", payload_type=type(payload).__name__)
        else:
            # stop_event won the race
            break
    return events_total, batches


async def main() -> int:
    """Entry point."""
    configure_logging()
    try:
        settings = StreamSettings()  # type: ignore[call-arg]
    except ValidationError as e:
        logger.error("config invalid", errors=e.errors())
        return 2

    duration = _stream_duration()
    stop_event = asyncio.Event()
    loop = asyncio.get_running_loop()

    def _signal_handler() -> None:
        logger.info("sigint received; draining stream")
        stop_event.set()

    # Windows event loop doesn't support add_signal_handler.
    for sig in (signal.SIGINT, signal.SIGTERM):
        with contextlib.suppress(NotImplementedError):
            loop.add_signal_handler(sig, _signal_handler)

    async with Client(settings.api_url, settings.api_key) as client:
        logger.info(
            "subscribing",
            canvas_id=settings.canvas_id,
            duration_seconds=duration,
        )
        consumer = asyncio.create_task(
            _consume(client, settings.canvas_id, stop_event),
        )
        try:
            await asyncio.wait_for(asyncio.shield(consumer), timeout=duration)
        except TimeoutError:
            logger.info("duration elapsed; signalling stop")
            stop_event.set()
            events_total, batches = await consumer
        else:
            events_total, batches = consumer.result()

    logger.info(
        "stream finished",
        batches=batches,
        events_total=events_total,
        duration_seconds=duration,
    )
    print(f"received {events_total} events across {batches} batches in {duration}s")
    return 0


if __name__ == "__main__":
    sys.exit(asyncio.run(main()))
