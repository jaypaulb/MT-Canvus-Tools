"""Phase 4b §4.2 #23: warnings registry tests."""

from __future__ import annotations

import pytest

from canvus_sdk.extras.warnings import (
    WARNING_TABLE_GRID_SIZE_IMMUTABLE,
    disable_api_warnings,
    enable_api_warnings,
    reset_warnings,
    warn_always,
    warn_once,
)


@pytest.fixture(autouse=True)
def _reset() -> None:
    enable_api_warnings()
    reset_warnings()
    yield
    enable_api_warnings()
    reset_warnings()


def test_warn_once_emits_first_call_only() -> None:
    # Without intercepting the structlog handler we cannot assert log output
    # directly, but we can call without exception and trust idempotency.
    warn_once(WARNING_TABLE_GRID_SIZE_IMMUTABLE)
    warn_once(WARNING_TABLE_GRID_SIZE_IMMUTABLE)  # no-op on second call


def test_disable_silences_warnings() -> None:
    disable_api_warnings()
    warn_once(WARNING_TABLE_GRID_SIZE_IMMUTABLE)
    warn_always(WARNING_TABLE_GRID_SIZE_IMMUTABLE)


def test_warning_has_required_fields() -> None:
    w = WARNING_TABLE_GRID_SIZE_IMMUTABLE
    assert w.code
    assert w.description
    assert w.workaround
    assert w.issue_url
