"""Example 06 responder: tiny diagnostic for the Ollama side of the pair.

Polls ``OLLAMA_URL/api/tags`` and prints the available models. Useful for
verifying that Ollama is reachable and that the model the watcher will ask
for is actually installed.
"""

from __future__ import annotations

import asyncio
import os
import sys

import httpx
import structlog

from canvus_sdk import configure_logging

logger = structlog.get_logger(__name__)


def _ollama_url() -> str:
    return os.environ.get("OLLAMA_URL", "http://localhost:11434").rstrip("/")


async def main() -> int:
    """Entry point. Returns 0 if Ollama responded, 1 otherwise."""
    configure_logging()
    url = f"{_ollama_url()}/api/tags"
    logger.info("polling ollama", url=url)
    try:
        async with httpx.AsyncClient(timeout=10.0) as http:
            resp = await http.get(url)
            resp.raise_for_status()
    except httpx.HTTPError as e:
        logger.error("ollama unreachable", url=url, error=str(e))
        return 1

    payload = resp.json()
    models = payload.get("models", [])
    if not models:
        print("(ollama is reachable but no models are installed)")
        return 0

    print(f"models available at {_ollama_url()}:")
    for model in models:
        if not isinstance(model, dict):
            continue
        name = model.get("name", "(unknown)")
        size = model.get("size", 0)
        try:
            size_gb = float(size) / (1024**3)
            size_str = f"{size_gb:.2f} GiB"
        except (TypeError, ValueError):
            size_str = "?"
        modified = model.get("modified_at", "")
        print(f"  {name:<32} {size_str:>12}   {modified}")
    return 0


if __name__ == "__main__":
    sys.exit(asyncio.run(main()))
