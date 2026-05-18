"""Entrypoint for ``python -m canvus_mcp_server``.

Launches the FastAPI application via uvicorn. Configuration comes from the
shared :class:`Settings` (``CANVUS_API_URL`` / ``CANVUS_API_KEY`` /
``CANVUS_MCP_SERVER_HOST`` / ``CANVUS_MCP_SERVER_PORT``).
"""

from __future__ import annotations

import uvicorn

from .app import create_app
from .config import Settings


def main() -> None:
    """Run the MCP server."""
    settings = Settings()  # type: ignore[call-arg]
    app = create_app(settings=settings)
    uvicorn.run(
        app,
        host=settings.host,
        port=settings.port,
        log_level=settings.log_level.lower(),
    )


if __name__ == "__main__":
    main()
