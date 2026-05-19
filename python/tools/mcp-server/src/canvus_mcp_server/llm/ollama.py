"""Async Ollama client built on ``httpx``.

This is the Phase 4d port of the legacy ``llm_client.py`` (541 LOC,
``aiohttp``-based, module-scoped global, silent offline-mode fallbacks).
The port follows the monorepo Python conventions:

* :mod:`httpx` async client per instance (not per-call construction).
* ``structlog`` for logging — never ``print()`` or ``logging.getLogger()``.
* No silent fallbacks — failures raise :class:`OllamaError`.
* No module-level mutable state — configuration is injected as
  :class:`OllamaConfig`.
"""

from __future__ import annotations

import asyncio
import json
from dataclasses import dataclass
from typing import Any, TypedDict

import httpx
import structlog

logger = structlog.get_logger(__name__)


class OllamaError(Exception):
    """Base class for Ollama client errors."""


class OllamaConnectionError(OllamaError):
    """Raised when the Ollama HTTP API is unreachable."""


class OllamaInferenceError(OllamaError):
    """Raised when an inference request returns a non-2xx response."""


class Message(TypedDict):
    """Single chat message — Ollama ``/api/chat`` schema."""

    role: str
    content: str


@dataclass(frozen=True)
class OllamaConfig:
    """Immutable runtime configuration for :class:`OllamaClient`.

    Attributes:
        base_url: Ollama HTTP endpoint, e.g. ``http://localhost:11434``.
        model: Default model name (``gemma3:2b`` etc.).
        timeout_seconds: Per-request timeout in seconds.
        max_retries: Retry budget on connection/inference errors.
        temperature: Default sampling temperature.
        max_tokens: Default ``num_predict`` (max tokens to generate).
        top_p: Default nucleus-sampling threshold.
    """

    base_url: str
    model: str
    timeout_seconds: float = 30.0
    max_retries: int = 3
    temperature: float = 0.7
    max_tokens: int = 2048
    top_p: float = 0.9


class OllamaClient:
    """Lightweight Ollama API client.

    The client owns a single :class:`httpx.AsyncClient` for its lifetime.
    Call :meth:`aclose` when done, or use the instance as an async context
    manager.
    """

    def __init__(self, config: OllamaConfig) -> None:
        self._config = config
        self._http = httpx.AsyncClient(
            base_url=config.base_url.rstrip("/"),
            timeout=config.timeout_seconds,
        )
        self._available_models: list[str] | None = None
        self._log = logger.bind(component="ollama_client", base_url=config.base_url)

    @property
    def config(self) -> OllamaConfig:
        """Return the immutable config."""
        return self._config

    async def __aenter__(self) -> OllamaClient:
        return self

    async def __aexit__(self, *exc_info: Any) -> None:
        await self.aclose()

    async def aclose(self) -> None:
        """Close the underlying HTTP client."""
        await self._http.aclose()

    # ---- Health & catalogue -------------------------------------------------

    async def health_check(self) -> bool:
        """Return ``True`` iff ``/api/tags`` responds with HTTP 200.

        Raises:
            OllamaConnectionError: if the HTTP transport itself fails.
        """
        try:
            response = await self._http.get("/api/tags")
        except httpx.HTTPError as exc:
            self._log.warning("ollama health-check transport failure", error=str(exc))
            raise OllamaConnectionError(f"Ollama unreachable: {exc}") from exc
        return response.status_code == 200

    async def get_available_models(self) -> list[str]:
        """List installed model names (cached for the client's lifetime).

        Raises:
            OllamaConnectionError: transport failure.
            OllamaInferenceError: non-200 response from Ollama.
        """
        if self._available_models is not None:
            return self._available_models
        try:
            response = await self._http.get("/api/tags")
        except httpx.HTTPError as exc:
            raise OllamaConnectionError(f"Ollama unreachable: {exc}") from exc
        if response.status_code != 200:
            raise OllamaInferenceError(
                f"/api/tags returned HTTP {response.status_code}"
            )
        data = response.json()
        models = [str(m["name"]) for m in data.get("models", []) if "name" in m]
        self._available_models = models
        return models

    async def is_model_available(self, model_name: str) -> bool:
        """Return ``True`` iff ``model_name`` is in the catalogue."""
        try:
            models = await self.get_available_models()
        except OllamaError:
            return False
        return model_name in models

    # ---- Inference ----------------------------------------------------------

    async def generate(
        self,
        prompt: str,
        *,
        model: str | None = None,
        system: str | None = None,
        temperature: float | None = None,
        max_tokens: int | None = None,
        top_p: float | None = None,
    ) -> str:
        """Run an ``/api/generate`` completion and return the response text.

        Args:
            prompt: User prompt body.
            model: Override the default model.
            system: Optional system prompt.
            temperature: Override default sampling temperature.
            max_tokens: Override default ``num_predict``.
            top_p: Override default top-p.

        Returns:
            The model's textual response (the ``response`` field).

        Raises:
            OllamaConnectionError: transport failure.
            OllamaInferenceError: non-200 response after retries exhausted.
        """
        payload: dict[str, Any] = {
            "model": model or self._config.model,
            "prompt": prompt,
            "stream": False,
            "options": {
                "temperature": (
                    temperature if temperature is not None else self._config.temperature
                ),
                "num_predict": (
                    max_tokens if max_tokens is not None else self._config.max_tokens
                ),
                "top_p": top_p if top_p is not None else self._config.top_p,
            },
        }
        if system is not None:
            payload["system"] = system

        data = await self._post_with_retry("/api/generate", payload)
        return str(data.get("response", ""))

    async def chat(
        self,
        messages: list[Message],
        *,
        model: str | None = None,
        temperature: float | None = None,
        max_tokens: int | None = None,
        top_p: float | None = None,
    ) -> str:
        """Run an ``/api/chat`` completion and return the assistant message.

        Raises:
            OllamaConnectionError: transport failure.
            OllamaInferenceError: non-200 response after retries exhausted.
        """
        payload: dict[str, Any] = {
            "model": model or self._config.model,
            "messages": list(messages),
            "stream": False,
            "options": {
                "temperature": (
                    temperature if temperature is not None else self._config.temperature
                ),
                "num_predict": (
                    max_tokens if max_tokens is not None else self._config.max_tokens
                ),
                "top_p": top_p if top_p is not None else self._config.top_p,
            },
        }
        data = await self._post_with_retry("/api/chat", payload)
        message = data.get("message")
        if isinstance(message, dict):
            return str(message.get("content", ""))
        return ""

    async def generate_json(
        self,
        prompt: str,
        *,
        model: str | None = None,
        system: str | None = None,
        temperature: float | None = None,
    ) -> Any:
        """Run :meth:`generate` and parse the result as JSON.

        Raises:
            OllamaInferenceError: if the response cannot be parsed as JSON
                — this is *not* swallowed silently per python.md error rules.
        """
        text = await self.generate(
            prompt, model=model, system=system, temperature=temperature
        )
        text = text.strip()
        # Strip markdown ```json fences if present (common in instruct models).
        if text.startswith("```"):
            lines = text.splitlines()
            if lines[0].startswith("```"):
                lines = lines[1:]
            if lines and lines[-1].startswith("```"):
                lines = lines[:-1]
            text = "\n".join(lines).strip()
        try:
            return json.loads(text)
        except json.JSONDecodeError as exc:
            raise OllamaInferenceError(
                f"LLM response was not valid JSON: {exc.msg}"
            ) from exc

    # ---- Internals ----------------------------------------------------------

    async def _post_with_retry(
        self, path: str, payload: dict[str, Any]
    ) -> dict[str, Any]:
        last_error: Exception | None = None
        for attempt in range(self._config.max_retries):
            try:
                response = await self._http.post(path, json=payload)
            except httpx.HTTPError as exc:
                last_error = exc
                self._log.warning(
                    "ollama transport error",
                    path=path,
                    attempt=attempt + 1,
                    max_retries=self._config.max_retries,
                    error=str(exc),
                )
                await asyncio.sleep(1.0 * (attempt + 1))
                continue

            if response.status_code != 200:
                last_error = OllamaInferenceError(
                    f"{path} returned HTTP {response.status_code}"
                )
                self._log.warning(
                    "ollama non-200 response",
                    path=path,
                    status=response.status_code,
                    attempt=attempt + 1,
                )
                await asyncio.sleep(1.0 * (attempt + 1))
                continue

            try:
                data = response.json()
            except json.JSONDecodeError as exc:
                raise OllamaInferenceError(
                    f"{path} returned non-JSON body: {exc.msg}"
                ) from exc
            if not isinstance(data, dict):
                raise OllamaInferenceError(
                    f"{path} returned a non-object JSON payload"
                )
            return data

        assert last_error is not None
        if isinstance(last_error, OllamaError):
            raise last_error
        raise OllamaConnectionError(
            f"Ollama {path} failed after {self._config.max_retries} attempts: "
            f"{last_error}"
        ) from last_error


__all__ = [
    "Message",
    "OllamaClient",
    "OllamaConfig",
    "OllamaConnectionError",
    "OllamaError",
    "OllamaInferenceError",
]
