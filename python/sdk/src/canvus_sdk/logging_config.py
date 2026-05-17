"""structlog configuration for the Canvus SDK.

The SDK does not call :func:`configure_logging` by default — it is the
embedding application's responsibility to decide on log format and level.
Call :func:`configure_logging` once during application startup if you want
the SDK's structured log lines to be emitted with sensible defaults.
"""

from __future__ import annotations

import logging
import os
import sys

import structlog
from structlog.types import Processor


def configure_logging() -> None:
    """Configure structlog once for the process. Idempotent.

    Reads ``LOG_LEVEL`` (default ``INFO``) and ``LOG_FORMAT``
    (``json`` | ``console``; default ``console`` if attached to a TTY,
    otherwise ``json``) from the environment.
    """
    log_level = os.environ.get("LOG_LEVEL", "INFO").upper()
    log_format = os.environ.get(
        "LOG_FORMAT",
        "console" if sys.stderr.isatty() else "json",
    )

    timestamper = structlog.processors.TimeStamper(fmt="iso", utc=True)
    shared_processors: list[Processor] = [
        structlog.contextvars.merge_contextvars,
        structlog.processors.add_log_level,
        structlog.processors.StackInfoRenderer(),
        timestamper,
    ]

    if log_format == "json":
        renderer: Processor = structlog.processors.JSONRenderer()
    else:
        renderer = structlog.dev.ConsoleRenderer(colors=True)

    structlog.configure(
        processors=[
            *shared_processors,
            structlog.processors.format_exc_info,
            renderer,
        ],
        wrapper_class=structlog.make_filtering_bound_logger(
            getattr(logging, log_level, logging.INFO),
        ),
        context_class=dict,
        logger_factory=structlog.PrintLoggerFactory(file=sys.stderr),
        cache_logger_on_first_use=True,
    )


__all__ = ["configure_logging"]
