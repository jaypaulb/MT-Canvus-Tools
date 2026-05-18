"""Phase 4b §4.2 #23: APIWarning registry port.

Catalogues known Canvus API quirks the SDK is aware of. Each entry has a
machine-readable ``code``, a human description, a recommended workaround,
and a link to the upstream tracker. Ports Go's ``warnings.go``.

Usage::

    from canvus_sdk.extras import warnings as api_warnings

    api_warnings.warn_once(api_warnings.WARNING_TABLE_GRID_SIZE_IMMUTABLE)

Set ``CANVUS_SDK_DISABLE_WARNINGS=1`` to silence at process startup, or call
:func:`disable_api_warnings` / :func:`enable_api_warnings` at runtime.
"""

from __future__ import annotations

import os
import threading
from dataclasses import dataclass

import structlog

__all__ = [
    "WARNING_IMAGE_ASPECT_RATIO_NOT_PRESERVED",
    "WARNING_NOTE_TITLE_NOT_EXPOSED",
    "WARNING_PDF_SIZE_BUG",
    "WARNING_TABLE_GRID_SIZE_IMMUTABLE",
    "WARNING_VIDEO_ASPECT_RATIO_NOT_PRESERVED",
    "WARNING_VIDEO_INPUT_TITLE_NOT_EXPOSED",
    "APIWarning",
    "disable_api_warnings",
    "enable_api_warnings",
    "reset_warnings",
    "warn_always",
    "warn_once",
]

_logger = structlog.get_logger(__name__)


@dataclass(frozen=True, slots=True)
class APIWarning:
    """One catalogued API quirk the SDK can surface to callers."""

    code: str
    description: str
    workaround: str
    issue_url: str


WARNING_NOTE_TITLE_NOT_EXPOSED = APIWarning(
    code="NOTE_TITLE_NOT_EXPOSED",
    description=(
        "Note widget 'title' field is not exposed by the Canvus API. Title "
        "values in requests are ignored and responses will not include the title."
    ),
    workaround="Use the 'name' field instead for identifying notes.",
    issue_url="https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/38",
)

WARNING_VIDEO_INPUT_TITLE_NOT_EXPOSED = APIWarning(
    code="VIDEOINPUT_TITLE_NOT_EXPOSED",
    description="VideoInput widget 'title' field is not exposed by the Canvus API.",
    workaround="No workaround available. Await Canvus API fix.",
    issue_url="https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/13",
)

WARNING_PDF_SIZE_BUG = APIWarning(
    code="PDF_SIZE_BUG",
    description=(
        "PDF widget size changes via PATCH update the bounding box but the "
        "actual PDF content stays at its original size."
    ),
    workaround="Avoid resizing PDFs via API. Delete and recreate if different size is needed.",
    issue_url="https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/15",
)

WARNING_IMAGE_ASPECT_RATIO_NOT_PRESERVED = APIWarning(
    code="IMAGE_ASPECT_RATIO_NOT_PRESERVED",
    description="Image widget size changes via PATCH do not preserve aspect ratio.",
    workaround="Calculate correct aspect-ratio-preserving dimensions before update.",
    issue_url="https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/39",
)

WARNING_VIDEO_ASPECT_RATIO_NOT_PRESERVED = APIWarning(
    code="VIDEO_ASPECT_RATIO_NOT_PRESERVED",
    description="Video widget size changes via PATCH do not preserve aspect ratio.",
    workaround="Calculate correct aspect-ratio-preserving dimensions before update.",
    issue_url="https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/39",
)

WARNING_TABLE_GRID_SIZE_IMMUTABLE = APIWarning(
    code="TABLE_GRID_SIZE_IMMUTABLE",
    description=(
        "Table widget 'grid_size' is set at creation time and silently "
        "ignored on PATCH."
    ),
    workaround="Omit grid_size from PATCH requests; recreate the table to change dimensions.",
    issue_url="https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/-",
)


_lock = threading.Lock()
_issued: set[str] = set()
_enabled = os.environ.get("CANVUS_SDK_DISABLE_WARNINGS", "0") != "1"


def disable_api_warnings() -> None:
    """Silence every subsequent :func:`warn_once` / :func:`warn_always` call."""
    global _enabled
    with _lock:
        _enabled = False


def enable_api_warnings() -> None:
    """Re-enable warnings if previously disabled."""
    global _enabled
    with _lock:
        _enabled = True


def reset_warnings() -> None:
    """Clear the once-only-emission record (mainly for tests)."""
    with _lock:
        _issued.clear()


def warn_once(warning: APIWarning) -> None:
    """Emit a warning the first time it appears in this process."""
    with _lock:
        if not _enabled or warning.code in _issued:
            return
        _issued.add(warning.code)
    _logger.warning(
        "canvus_sdk api warning",
        code=warning.code,
        description=warning.description,
        workaround=warning.workaround,
        issue_url=warning.issue_url,
    )


def warn_always(warning: APIWarning) -> None:
    """Emit a warning on every invocation (use sparingly)."""
    with _lock:
        if not _enabled:
            return
    _logger.warning(
        "canvus_sdk api warning",
        code=warning.code,
        description=warning.description,
    )
