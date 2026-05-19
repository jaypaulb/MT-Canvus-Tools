"""LLM integration subpackage.

Houses the Ollama client, SQLite response cache, and PDF processing
pipeline. Tools depend on instances of these classes via constructor
injection — no module-level globals.
"""

from __future__ import annotations

from .cache import CacheError, LLMCache, LLMCacheConfig
from .ollama import (
    Message,
    OllamaClient,
    OllamaConfig,
    OllamaConnectionError,
    OllamaError,
    OllamaInferenceError,
)

__all__ = [
    "CacheError",
    "LLMCache",
    "LLMCacheConfig",
    "Message",
    "OllamaClient",
    "OllamaConfig",
    "OllamaConnectionError",
    "OllamaError",
    "OllamaInferenceError",
]
