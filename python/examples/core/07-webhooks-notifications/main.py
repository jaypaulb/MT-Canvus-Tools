"""Example 07: bridge Canvus subscribe events to an outbound webhook.

The Canvus API does not ship a built-in webhook system; this example shows
the canonical *pattern* for building one on top of the streaming subscribe
endpoint.

For every NEW widget event observed on the canvas, the bridge POSTs a small
JSON envelope to ``WEBHOOK_URL`` with retry + exponential backoff. The
initial snapshot batch is treated as "already existed" and is not forwarded.
"""

from __future__ import annotations

import asyncio
import contextlib
import os
import signal
import sys
import time
from datetime import UTC, datetime
from typing import Any

import httpx
import structlog
from pydantic import Field, ValidationError
from pydantic_settings import BaseSettings, SettingsConfigDict

from canvus_sdk import APIError, Client, configure_logging

logger = structlog.get_logger(__name__)

MAX_ATTEMPTS = 3
BACKOFF_SECONDS = (2.0, 4.0, 8.0)


class WebhookSettings(BaseSettings):
    """Environment configuration for the webhook bridge."""

    model_config = SettingsConfigDict(
        env_prefix="CANVUS_",
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    api_url: str = Field(...)
    api_key: str = Field(...)
    canvas_id: str = Field(...)


def _webhook_url() -> str:
    url = os.environ.get("WEBHOOK_URL", "").strip()
    if not url:
        raise RuntimeError("WEBHOOK_URL not set")
    return url


async def _post_with_retry(
    http: httpx.AsyncClient,
    url: str,
    body: dict[str, Any],
) -> bool:
    """POST ``body`` to ``url`` with up to ``MAX_ATTEMPTS`` tries.

    Returns True on first 2xx response, False if all attempts fail.
    """
    for attempt in range(1, MAX_ATTEMPTS + 1):
        started = time.perf_counter()
        try:
            resp = await http.post(url, json=body, timeout=10.0)
        except httpx.HTTPError as e:
            elapsed_ms = (time.perf_counter() - started) * 1000
            logger.warning(
                "webhook transport error",
                attempt=attempt,
                elapsed_ms=round(elapsed_ms, 1),
                error=str(e),
            )
        else:
            elapsed_ms = (time.perf_counter() - started) * 1000
            if 200 <= resp.status_code < 300:
                logger.info(
                    "webhook delivered",
                    attempt=attempt,
                    status_code=resp.status_code,
                    elapsed_ms=round(elapsed_ms, 1),
                )
                return True
            logger.warning(
                "webhook non-2xx",
                attempt=attempt,
                status_code=resp.status_code,
                elapsed_ms=round(elapsed_ms, 1),
                body_snippet=resp.text[:200],
            )
        if attempt < MAX_ATTEMPTS:
            delay = BACKOFF_SECONDS[min(attempt - 1, len(BACKOFF_SECONDS) - 1)]
            await asyncio.sleep(delay)
    logger.error("webhook gave up", attempts=MAX_ATTEMPTS)
    return False


async def _consume(
    client: Client,
    http: httpx.AsyncClient,
    settings: WebhookSettings,
    webhook_url: str,
    stop_event: asyncio.Event,
) -> None:
    """Forward newly-observed widget events to the webhook."""
    seen: set[str] = set()
    snapshot_drained = False
    sub = client.widgets.subscribe(settings.canvas_id)
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
        if stop_task in done:
            break
        try:
            payload = next_task.result()
        except StopAsyncIteration:
            logger.info("stream ended")
            break
        except APIError as e:
            logger.error("stream api error", status_code=e.status_code)
            break

        items: list[dict[str, Any]] = []
        if isinstance(payload, list):
            items = [i for i in payload if isinstance(i, dict)]
        elif isinstance(payload, dict):
            items = [payload]

        if not snapshot_drained:
            # The first batch is a snapshot of pre-existing widgets; treat
            # them as "already there" so we only forward truly NEW events.
            for item in items:
                wid = item.get("id")
                if isinstance(wid, str):
                    seen.add(wid)
            snapshot_drained = True
            logger.info("snapshot drained", count=len(seen))
            continue

        for item in items:
            wid = item.get("id")
            if not isinstance(wid, str):
                continue
            if wid in seen:
                continue
            seen.add(wid)
            envelope = {
                "event": "widget.created",
                "canvas_id": settings.canvas_id,
                "widget_id": wid,
                "widget_type": item.get("widget_type"),
                "timestamp": datetime.now(UTC).isoformat(),
            }
            await _post_with_retry(http, webhook_url, envelope)


async def main() -> int:
    """Entry point."""
    configure_logging()
    try:
        settings = WebhookSettings()  # type: ignore[call-arg]
    except ValidationError as e:
        logger.error("config invalid", errors=e.errors())
        return 2
    try:
        webhook_url = _webhook_url()
    except RuntimeError as e:
        logger.error("config invalid", error=str(e))
        return 2

    stop_event = asyncio.Event()
    loop = asyncio.get_running_loop()

    def _signal_handler() -> None:
        logger.info("signal received; stopping bridge")
        stop_event.set()

    for sig in (signal.SIGINT, signal.SIGTERM):
        with contextlib.suppress(NotImplementedError):
            loop.add_signal_handler(sig, _signal_handler)

    logger.info(
        "bridge starting",
        canvas_id=settings.canvas_id,
        webhook_url=webhook_url,
    )
    async with (
        Client(settings.api_url, settings.api_key) as client,
        httpx.AsyncClient() as http,
    ):
        await _consume(client, http, settings, webhook_url, stop_event)
    logger.info("bridge stopped")
    return 0


if __name__ == "__main__":
    sys.exit(asyncio.run(main()))
