"""Phase 4d Round D Task D1 — live verification of canvas permissions subscribe.

Runs against the dev/test Canvus server (dev-mtcs.multitaction.com). The
test is marked ``live`` and is skipped when the required env vars are not
set, so ``pytest`` without ``-m live`` remains a green run.

Run:

    source /path/to/.secrets && \\
      CANVUS_DEV_BASE_URL=$CANVUS_DEV_BASE_URL \\
      CANVUS_DEV_API_KEY=$CANVUS_DEV_API_KEY \\
      uv run pytest sdk/tests/test_permissions_live.py -m live -v

Notes:

- ``FoldersResource`` does not expose a ``subscribe_permissions`` helper,
  so only the canvas-permissions stream is covered by the Python live
  test. The Go SDK does expose both helpers and is verified in
  ``go/sdk/canvus/permissions_subscribe_live_test.go``.
- We mutate permissions via raw ``httpx`` POST to avoid the SDK's response
  validation, which is unrelated to the streaming surface under test.
"""

from __future__ import annotations

import asyncio
import contextlib
import os
import time

import httpx
import pytest

from canvus_sdk import Client

pytestmark = pytest.mark.live

SENTINEL_GROUP_ID = 1  # "All Users" — present on every Canvus server.


def _require_env() -> tuple[str, str]:
    base_url = os.environ.get("CANVUS_DEV_BASE_URL", "").strip()
    api_key = os.environ.get("CANVUS_DEV_API_KEY", "").strip()
    if not base_url or not api_key:
        pytest.skip(
            "live test: CANVUS_DEV_BASE_URL and CANVUS_DEV_API_KEY must both be set"
        )
    return base_url, api_key


async def _consume_into_queue(stream, queue: asyncio.Queue) -> None:
    """Background task that drains ``stream`` into ``queue``.

    We can't ``asyncio.wait_for`` on ``stream.__anext__()`` directly,
    because cancelling that coroutine closes the async generator and
    tears down the underlying HTTP stream. Instead, this task owns the
    iterator and pushes events into a queue we control.
    """
    try:
        async for event in stream:
            await queue.put(("event", event))
    except Exception as exc:
        await queue.put(("error", exc))
    finally:
        await queue.put(("eof", None))


async def _drain_until_quiet(
    queue: asyncio.Queue, settle_seconds: float = 2.0
) -> list:
    """Drain ``queue`` until no event arrives for ``settle_seconds``.

    Returns the list of events drained. ``error`` and ``eof`` markers are
    re-queued so they reach the next reader.
    """
    drained: list = []
    while True:
        try:
            kind, payload = await asyncio.wait_for(
                queue.get(), timeout=settle_seconds
            )
        except TimeoutError:
            return drained
        if kind == "event":
            drained.append(payload)
            continue
        # Surface error/eof to the next reader.
        await queue.put((kind, payload))
        return drained


def _raw_post(base_url: str, api_key: str, endpoint: str, body: dict) -> None:
    """POST to the permissions endpoint with raw httpx, bypassing SDK validation."""
    url = base_url.rstrip("/") + "/" + endpoint
    with httpx.Client(verify=False, timeout=10.0) as client:
        resp = client.post(
            url,
            headers={"Private-Token": api_key, "Content-Type": "application/json"},
            json=body,
        )
        if resp.status_code >= 300:
            raise AssertionError(
                f"raw POST {endpoint} returned {resp.status_code}: {resp.text}"
            )


def _best_effort_post(base_url: str, api_key: str, endpoint: str, body: dict) -> None:
    """Non-fatal POST used for revert/cleanup."""
    try:
        _raw_post(base_url, api_key, endpoint, body)
    except Exception as exc:
        print(f"best-effort revert {endpoint}: {exc}")


async def test_subscribe_canvas_permissions_live() -> None:
    """Verify ``CanvasesResource.subscribe_permissions`` emits an event on PATCH."""
    base_url, api_key = _require_env()

    sub_client = Client(base_url, api_key, verify_ssl=False, max_retries=0)
    mut_client = Client(base_url, api_key, verify_ssl=False, max_retries=0)
    try:
        # Create a temporary canvas.
        canvas = await sub_client.canvases.create(
            {"name": f"d1-perms-canvas-py-{time.time_ns()}"}
        )
        canvas_id = canvas.id
        assert canvas_id, "create canvas returned no id"
        print(f"created canvas {canvas_id}")

        stream = sub_client.canvases.subscribe_permissions(canvas_id)
        event_queue: asyncio.Queue = asyncio.Queue()
        consumer = asyncio.create_task(_consume_into_queue(stream, event_queue))
        try:
            drained = await _drain_until_quiet(event_queue, settle_seconds=2.0)
            print(f"snapshot: drained {len(drained)} initial event(s)")

            # PATCH permissions via raw HTTP.
            _raw_post(
                base_url,
                api_key,
                f"canvases/{canvas_id}/permissions",
                {
                    "editors_can_share": True,
                    "link_permission": "view",
                    "users": [],
                    "groups": [{"id": SENTINEL_GROUP_ID, "permission": "view"}],
                },
            )
            print(
                f"PATCHed canvas {canvas_id} permissions: "
                f"group {SENTINEL_GROUP_ID} → view"
            )

            # Wait for the change event (10s timeout).
            try:
                kind, payload = await asyncio.wait_for(
                    event_queue.get(), timeout=10.0
                )
            except TimeoutError:
                pytest.fail(
                    "FAIL: no change event on subscribe stream within 10s "
                    "after permissions PATCH"
                )
            if kind == "error":
                pytest.fail(f"FAIL: subscribe stream errored: {payload!r}")
            if kind == "eof":
                pytest.fail(
                    "FAIL: subscribe stream closed before change event arrived"
                )
            event = payload
            print(f"RECEIVED CHANGE EVENT: {event!r}")

            # Best-effort payload check.
            assert (
                getattr(event, "editors_can_share", False)
                or getattr(event, "link_permission", "")
                or getattr(event, "groups", [])
            ), f"change event payload looked empty: {event!r}"

            # Revert (best-effort).
            _best_effort_post(
                base_url,
                api_key,
                f"canvases/{canvas_id}/permissions",
                {
                    "editors_can_share": False,
                    "link_permission": "",
                    "users": [],
                    "groups": [],
                },
            )
        finally:
            # Stop the consumer (will tear down the underlying stream).
            consumer.cancel()
            with contextlib.suppress(asyncio.CancelledError, Exception):
                await consumer
            # Cleanup canvas.
            try:
                await mut_client.canvases.delete(canvas_id)
            except Exception as exc:
                print(f"cleanup: delete canvas {canvas_id} failed: {exc}")
    finally:
        await sub_client.aclose()
        await mut_client.aclose()
