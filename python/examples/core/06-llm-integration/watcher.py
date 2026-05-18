"""Example 06 watcher: turn Canvus "question" notes into Ollama-backed answers.

Subscribes to a canvas's notes endpoint. When a note appears whose text
starts with ``?``, the watcher:

    1. Strips the leading ``?`` and uses the rest as the prompt.
    2. POSTs ``/api/generate`` to the configured Ollama instance with
       ``stream=false``.
    3. Creates a sibling note 400 px to the right of the question note, with
       background colour ``#1d71b8ff``, containing the model's response.

The watcher tracks seen note ids in an in-memory set so the initial snapshot
does not double-answer questions that were already on the canvas at startup.
"""

from __future__ import annotations

import asyncio
import contextlib
import os
import signal
import sys
from typing import Any

import httpx
import structlog
from pydantic import Field, ValidationError
from pydantic_settings import BaseSettings, SettingsConfigDict

from canvus_sdk import APIError, Client, configure_logging

logger = structlog.get_logger(__name__)

ANSWER_BG = "#1d71b8ff"
ANSWER_OFFSET_PX = 400.0


class WatcherSettings(BaseSettings):
    """Environment configuration for the watcher."""

    model_config = SettingsConfigDict(
        env_prefix="CANVUS_",
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    api_url: str = Field(...)
    api_key: str = Field(...)
    canvas_id: str = Field(...)


def _ollama_url() -> str:
    return os.environ.get("OLLAMA_URL", "http://localhost:11434").rstrip("/")


def _ollama_model() -> str:
    return os.environ.get("OLLAMA_MODEL", "llama3.2")


async def _ask_ollama(
    http: httpx.AsyncClient,
    base_url: str,
    model: str,
    prompt: str,
) -> str:
    """POST a single prompt to Ollama. Returns the response text or an error string."""
    try:
        resp = await http.post(
            f"{base_url}/api/generate",
            json={"model": model, "prompt": prompt, "stream": False},
            timeout=httpx.Timeout(connect=5.0, read=120.0, write=30.0, pool=5.0),
        )
        resp.raise_for_status()
    except httpx.HTTPError as e:
        logger.error("ollama request failed", error=str(e))
        return f"(ollama error: {e})"
    body = resp.json()
    return body.get("response", "(ollama returned no `response` field)").strip()


async def _post_answer(
    client: Client,
    canvas_id: str,
    question: dict[str, Any],
    answer_text: str,
) -> None:
    """Create an answer note adjacent to the question."""
    q_loc = question.get("location", {}) or {}
    q_size = question.get("size", {}) or {}
    q_x = float(q_loc.get("x", 0.0))
    q_y = float(q_loc.get("y", 0.0))
    q_w = float(q_size.get("width", 250.0))
    payload = {
        "text": answer_text[:4000],  # avoid sending wall-of-text payloads
        "background_color": ANSWER_BG,
        "location": {"x": q_x + q_w + ANSWER_OFFSET_PX, "y": q_y},
        "size": {"width": 400.0, "height": 300.0},
    }
    try:
        created = await client.widgets.notes.create(canvas_id, payload)
    except APIError as e:
        logger.error(
            "answer note create failed",
            status_code=e.status_code,
            response_body=e.response_body,
        )
        return
    logger.info("answered", question_id=question.get("id"), answer_id=created.id)


def _is_question_note(item: dict[str, Any]) -> bool:
    """Return True if the payload looks like a question note we should answer."""
    if not isinstance(item, dict):
        return False
    if item.get("widget_type") not in (None, "Note"):
        return False
    text = item.get("text") or ""
    if not isinstance(text, str):
        return False
    return text.strip().startswith("?")


async def _consume(
    client: Client,
    http: httpx.AsyncClient,
    settings: WatcherSettings,
    stop_event: asyncio.Event,
) -> None:
    """Drain the subscription, answering question notes as they appear."""
    seen: set[str] = set()
    # The first frame the server emits is the snapshot of pre-existing
    # notes. Mark all snapshot IDs as `seen` without answering them so a
    # restart doesn't flood the canvas with duplicate replies to old
    # questions. Only notes that appear in frame 2+ trigger the LLM.
    snapshot_drained = False
    sub = client.widgets.subscribe(settings.canvas_id, widget_type="notes")
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

        items: list[dict[str, Any]]
        if isinstance(payload, list):
            items = [i for i in payload if isinstance(i, dict)]
        elif isinstance(payload, dict):
            items = [payload]
        else:
            continue

        for item in items:
            note_id = item.get("id")
            if not isinstance(note_id, str):
                continue
            if note_id in seen:
                continue
            seen.add(note_id)
            if not snapshot_drained:
                # Initial snapshot — mark as seen but do not answer.
                continue
            if not _is_question_note(item):
                continue
            text = (item.get("text") or "").strip()
            prompt = text.lstrip("?").strip()
            if not prompt:
                continue
            logger.info("question detected", note_id=note_id, prompt=prompt[:80])
            answer = await _ask_ollama(http, _ollama_url(), _ollama_model(), prompt)
            await _post_answer(client, settings.canvas_id, item, answer)

        # After processing the first batch (whatever shape it had), enable
        # answering for subsequent frames.
        snapshot_drained = True


async def main() -> int:
    """Entry point."""
    configure_logging()
    try:
        settings = WatcherSettings()  # type: ignore[call-arg]
    except ValidationError as e:
        logger.error("config invalid", errors=e.errors())
        return 2

    stop_event = asyncio.Event()
    loop = asyncio.get_running_loop()

    def _signal_handler() -> None:
        logger.info("signal received; stopping watcher")
        stop_event.set()

    for sig in (signal.SIGINT, signal.SIGTERM):
        with contextlib.suppress(NotImplementedError):
            loop.add_signal_handler(sig, _signal_handler)

    logger.info(
        "watcher starting",
        canvas_id=settings.canvas_id,
        ollama_url=_ollama_url(),
        ollama_model=_ollama_model(),
    )

    async with (
        Client(settings.api_url, settings.api_key) as client,
        httpx.AsyncClient() as http,
    ):
        await _consume(client, http, settings, stop_event)

    logger.info("watcher stopped")
    return 0


if __name__ == "__main__":
    sys.exit(asyncio.run(main()))
