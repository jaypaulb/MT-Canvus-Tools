"""Integration smoke tests.

Skipped unless ``CANVUS_API_KEY`` and ``CANVUS_API_URL`` are set in the
environment. Run explicitly with::

    uv run pytest -m integration sdk/tests/integration/
"""

from __future__ import annotations

import os

import pytest

from canvus_sdk import Client

pytestmark = pytest.mark.integration


@pytest.fixture(autouse=True)
def _skip_if_no_creds() -> None:
    if not os.environ.get("CANVUS_API_KEY"):
        pytest.skip("CANVUS_API_KEY not set; skipping integration test")
    if not os.environ.get("CANVUS_API_URL"):
        pytest.skip("CANVUS_API_URL not set; skipping integration test")


@pytest.mark.asyncio
async def test_server_info_round_trip() -> None:
    """Hit a live server and read ``/server-info``.

    This is the cheapest integration test we can run — it requires no
    canvases, users, or write access. Useful for smoke-testing CI
    credentials.
    """
    async with Client.from_env() as client:
        info = await client.server.get_info()
        assert info.version is not None
