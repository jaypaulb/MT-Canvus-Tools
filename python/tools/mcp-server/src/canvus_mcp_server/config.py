"""Pydantic-settings configuration for the Canvus MCP server.

The MCP server inherits ``CANVUS_API_URL`` / ``CANVUS_API_KEY`` /
``CANVUS_VERIFY_SSL`` from the SDK's settings and adds MCP-server-specific
fields (HTTP host/port, SQLite cache path, Ollama, etc.).

Only env-driven configuration belongs here; runtime values like the active
canvases the server has touched live in module state inside the MCP tools.
"""

from __future__ import annotations

from pathlib import Path

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """MCP server runtime configuration."""

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
        case_sensitive=False,
    )

    # ---- Canvus SDK (mirrors canvus_sdk.Settings field names) ----------------
    api_url: str = Field(
        ...,
        alias="CANVUS_API_URL",
        description="Base URL for the Canvus server.",
    )
    api_key: str = Field(
        ...,
        alias="CANVUS_API_KEY",
        description="Long-lived Private-Token for the Canvus API.",
    )
    verify_ssl: bool = Field(
        True,
        alias="CANVUS_VERIFY_SSL",
        description="Verify the Canvus server's TLS certificate.",
    )

    # ---- MCP server HTTP -----------------------------------------------------
    host: str = Field("0.0.0.0", alias="CANVUS_MCP_SERVER_HOST")
    port: int = Field(8000, alias="CANVUS_MCP_SERVER_PORT")
    debug: bool = Field(False, alias="CANVUS_MCP_SERVER_DEBUG")

    # ---- SQLite cache --------------------------------------------------------
    cache_enabled: bool = Field(True, alias="CACHE_ENABLED")
    database_path: Path = Field(
        Path.home() / ".canvus-mcp-server" / "pdf_cache.db",
        alias="DATABASE_PATH",
    )
    database_timeout: int = Field(30, alias="DATABASE_TIMEOUT")
    pdf_cache_ttl_seconds: int = Field(86_400, alias="PDF_CACHE_TTL")

    # ---- Logging -------------------------------------------------------------
    log_level: str = Field("INFO", alias="LOG_LEVEL")
    log_format: str = Field("console", alias="LOG_FORMAT")

    # ---- Ollama / LLM --------------------------------------------------------
    ollama_enabled: bool = Field(True, alias="OLLAMA_ENABLED")
    ollama_host: str = Field("localhost", alias="OLLAMA_HOST")
    ollama_port: int = Field(11_434, alias="OLLAMA_PORT")
    ollama_base_url: str | None = Field(None, alias="OLLAMA_BASE_URL")
    ollama_timeout: int = Field(30, alias="OLLAMA_TIMEOUT")
    ollama_max_retries: int = Field(3, alias="OLLAMA_MAX_RETRIES")

    default_model: str = Field("gemma3:2b", alias="DEFAULT_MODEL")
    model_temperature: float = Field(0.7, alias="MODEL_TEMPERATURE")
    model_max_tokens: int = Field(2048, alias="MODEL_MAX_TOKENS")
    model_top_p: float = Field(0.9, alias="MODEL_TOP_P")

    # ---- Health checks -------------------------------------------------------
    health_check_timeout: int = Field(10, alias="HEALTH_CHECK_TIMEOUT")

    @property
    def is_production(self) -> bool:
        """True when ``debug`` is False."""
        return not self.debug

    @property
    def effective_ollama_base_url(self) -> str:
        """Compose Ollama URL from host/port if ``ollama_base_url`` is unset."""
        return self.ollama_base_url or f"http://{self.ollama_host}:{self.ollama_port}"

    def ensure_directories(self) -> None:
        """Create directories the server expects to write to."""
        self.database_path.parent.mkdir(parents=True, exist_ok=True)


__all__ = ["Settings"]
