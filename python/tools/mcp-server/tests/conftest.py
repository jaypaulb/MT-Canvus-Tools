"""Shared pytest fixtures for the MCP-server test suite.

Tests exercise the production tool classes via real :class:`canvus_sdk.Client`
instances whose underlying httpx transport is mocked at the HTTP boundary
with :mod:`respx`. Per the Round 1+2 lessons we do not stub the SDK
client itself — that would let tool bugs hide behind a duplicate mock.
"""

from __future__ import annotations

from collections.abc import AsyncIterator

import pytest
import pytest_asyncio
import respx
from canvus_sdk import Client

API_URL = "https://canvus.test.invalid/api/v1"
API_KEY = "test_token"


@pytest.fixture
def base_url() -> str:
    return API_URL


@pytest_asyncio.fixture
async def client() -> AsyncIterator[Client]:
    """A real SDK client whose transport is mocked by ``respx_mock`` below."""
    c = Client(API_URL, API_KEY, verify_ssl=False, max_retries=0)
    try:
        yield c
    finally:
        await c.aclose()


@pytest.fixture
def respx_mock() -> respx.MockRouter:
    """Activate respx for the duration of one test."""
    with respx.mock(
        base_url=API_URL, assert_all_called=False
    ) as router:
        yield router
