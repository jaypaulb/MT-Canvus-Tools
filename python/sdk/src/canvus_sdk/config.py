"""Pydantic-settings configuration for the Canvus SDK.

Settings are sourced from environment variables prefixed with ``CANVUS_``
and (optionally) a ``.env`` file in the working directory. The :class:`Settings`
object is validated once at construction and then treated as frozen.
"""

from __future__ import annotations

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Canvus SDK runtime configuration.

    Environment variable names are derived from field names by prefixing with
    ``CANVUS_`` and upper-casing — e.g. ``api_key`` becomes ``CANVUS_API_KEY``.
    """

    model_config = SettingsConfigDict(
        env_prefix="CANVUS_",
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
        case_sensitive=False,
    )

    api_url: str = Field(
        ...,
        description=(
            "Base URL for the Canvus server, with or without the /api/v1/ suffix. "
            "Example: https://canvus.example.com or https://canvus.example.com/api/v1/"
        ),
    )
    api_key: str = Field(
        ...,
        description="Long-lived API token sent in the Private-Token header.",
    )
    verify_ssl: bool = Field(
        True,
        description="Whether to verify the server's TLS certificate.",
    )
    request_timeout_seconds: float = Field(
        30.0,
        description="Per-request read timeout, in seconds.",
    )
    connect_timeout_seconds: float = Field(
        5.0,
        description="TCP connect timeout, in seconds.",
    )
    max_retries: int = Field(
        3,
        description="Maximum number of automatic retries for transient errors.",
    )
    retry_initial_delay_seconds: float = Field(
        1.0,
        description="Initial backoff delay between retries.",
    )
    retry_backoff_factor: float = Field(
        2.0,
        description="Multiplier applied to the backoff delay after each retry.",
    )
    subscribe_buffer: int = Field(
        4,
        ge=1,
        description=(
            "asyncio.Queue capacity for buffered subscribe helpers. "
            "Raise this value for high-throughput consumers (live dashboards, ai-personas) "
            "to absorb bursts without blocking the streaming coroutine. "
            "Must be >= 1. Phase 4d Round B."
        ),
    )


__all__ = ["Settings"]
