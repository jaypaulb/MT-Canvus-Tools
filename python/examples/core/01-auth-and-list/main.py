"""Example 01: authenticate with an API key and list canvases.

Smoke test for the Canvus SDK. Prints a fixed-width table of every canvas the
authenticated user can see. Exits 0 on success, 2 on missing env vars, 1 on
any SDK error.
"""

from __future__ import annotations

import asyncio
import sys

import structlog

from canvus_sdk import APIError, Client, Settings, configure_logging

logger = structlog.get_logger(__name__)


def _print_table(rows: list[tuple[str, str, str, str, str]]) -> None:
    """Print a fixed-width table to stdout."""
    headers = ("ID", "NAME", "MODE", "ACCESS", "STATE")
    widths = [
        max(len(headers[i]), max((len(r[i]) for r in rows), default=0)) for i in range(5)
    ]
    fmt = "  ".join(f"{{:<{w}}}" for w in widths)
    print(fmt.format(*headers))
    print(fmt.format(*("-" * w for w in widths)))
    for row in rows:
        print(fmt.format(*row))


async def main() -> int:
    """Entry point. Returns process exit code."""
    configure_logging()
    try:
        settings = Settings()  # type: ignore[call-arg]
    except Exception as e:
        logger.error("missing required configuration", error=str(e))
        return 2

    async with Client.from_env(settings) as client:
        try:
            canvases = await client.canvases.list()
        except APIError as e:
            logger.error(
                "list canvases failed",
                status_code=e.status_code,
                response_body=e.response_body,
            )
            return 1

    rows: list[tuple[str, str, str, str, str]] = [
        (
            c.id,
            c.name or "",
            c.mode or "",
            c.access or "",
            c.state or "",
        )
        for c in canvases
    ]
    if not rows:
        print("(no canvases visible to this API key)")
        return 0
    _print_table(rows)
    logger.info("listed canvases", count=len(rows))
    return 0


if __name__ == "__main__":
    sys.exit(asyncio.run(main()))
