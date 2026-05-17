"""Shared pytest fixtures."""

from __future__ import annotations

from collections.abc import AsyncIterator

import pytest

from canvus_sdk import Client


@pytest.fixture
async def client() -> AsyncIterator[Client]:
    """A throwaway :class:`Client` against a non-routable test URL.

    Tests that need to make HTTP calls should mock the transport with
    ``respx`` (see ``test_client.py`` for the pattern).
    """
    c = Client(
        base_url="https://canvus.test.invalid",
        api_key="test-key",
        max_retries=0,
        request_timeout_seconds=1.0,
        connect_timeout_seconds=1.0,
    )
    try:
        yield c
    finally:
        await c.aclose()
