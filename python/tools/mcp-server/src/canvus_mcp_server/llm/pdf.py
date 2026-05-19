"""PDF processing pipeline.

Fetches a PDF over HTTP, extracts text with ``pdfplumber`` (matching the
legacy ``pdf_processing.py`` choice), splits the result into chunks,
and produces an LLM-driven summary via :class:`OllamaClient`.

Intentionally simpler than the 714 LOC legacy implementation — the
legacy module conflated cache management and SDK quirks. Here we
keep the PDF pipeline focused on the in-memory transformation
``URL -> bytes -> text -> chunks -> summary`` and let the caller wire
caching (via :class:`LLMCache`) and SDK fetches.
"""

from __future__ import annotations

import io
from dataclasses import dataclass, field
from typing import Any

import httpx
import pdfplumber
import structlog

from .ollama import OllamaClient, OllamaError

logger = structlog.get_logger(__name__)


class PDFExtractError(Exception):
    """Raised when PDF acquisition or extraction fails."""


@dataclass(frozen=True)
class PDFProcessorConfig:
    """Runtime config for :class:`PDFProcessor`.

    Attributes:
        max_bytes: Refuse PDFs larger than this (defence against DoS).
        chunk_chars: Approximate target size per text chunk.
        fetch_timeout_seconds: HTTP timeout when downloading the PDF.
    """

    max_bytes: int = 50 * 1024 * 1024  # 50 MiB
    chunk_chars: int = 4_000
    fetch_timeout_seconds: float = 60.0


@dataclass
class PDFSummary:
    """Structured result of a PDF summarisation."""

    page_count: int
    total_characters: int
    chunks: list[str] = field(default_factory=list)
    chunk_summaries: list[str] = field(default_factory=list)
    overall_summary: str = ""

    def to_dict(self) -> dict[str, Any]:
        """Return a JSON-serialisable representation."""
        return {
            "page_count": self.page_count,
            "total_characters": self.total_characters,
            "chunks": self.chunks,
            "chunk_summaries": self.chunk_summaries,
            "overall_summary": self.overall_summary,
        }


class PDFProcessor:
    """Async PDF fetch + extract + summarise pipeline."""

    def __init__(
        self, ollama: OllamaClient, config: PDFProcessorConfig | None = None
    ) -> None:
        self._ollama = ollama
        self._config = config or PDFProcessorConfig()
        self._log = logger.bind(component="pdf_processor")

    async def fetch(self, url: str) -> bytes:
        """Download the PDF at ``url`` and return its bytes.

        Raises:
            PDFExtractError: HTTP transport failure or size-cap exceeded.
        """
        try:
            async with httpx.AsyncClient(
                timeout=self._config.fetch_timeout_seconds, follow_redirects=True
            ) as http:
                response = await http.get(url)
                response.raise_for_status()
        except httpx.HTTPError as exc:
            raise PDFExtractError(f"failed to fetch PDF from {url}: {exc}") from exc

        content = response.content
        if len(content) > self._config.max_bytes:
            raise PDFExtractError(
                f"PDF exceeds size cap: {len(content)} > "
                f"{self._config.max_bytes} bytes"
            )
        return content

    def extract_text(self, pdf_bytes: bytes) -> tuple[str, int]:
        """Return ``(full_text, page_count)`` extracted from ``pdf_bytes``.

        Raises:
            PDFExtractError: parser failure.
        """
        try:
            with pdfplumber.open(io.BytesIO(pdf_bytes)) as pdf:
                pages = pdf.pages
                page_texts = [page.extract_text() or "" for page in pages]
                return "\n\n".join(page_texts), len(pages)
        except Exception as exc:  # pdfplumber raises bare Exception subclasses
            raise PDFExtractError(f"PDF text extraction failed: {exc}") from exc

    def chunk(self, text: str) -> list[str]:
        """Split ``text`` into roughly ``chunk_chars``-sized blocks.

        Splits prefer paragraph boundaries; the final chunk may be smaller.
        """
        size = self._config.chunk_chars
        if size <= 0 or len(text) <= size:
            return [text] if text else []

        chunks: list[str] = []
        remaining = text
        while len(remaining) > size:
            # Find the last paragraph break within the window.
            window = remaining[:size]
            split_at = window.rfind("\n\n")
            if split_at < size // 2:
                # Fall back to the last newline or hard cut.
                split_at = window.rfind("\n")
            if split_at < size // 2:
                split_at = size
            chunks.append(remaining[:split_at].strip())
            remaining = remaining[split_at:].lstrip()
        if remaining:
            chunks.append(remaining.strip())
        return [c for c in chunks if c]

    async def summarise(self, url: str) -> PDFSummary:
        """Run the full pipeline and return a structured summary.

        Raises:
            PDFExtractError: download or extraction failure.
            OllamaError: LLM call failure.
        """
        pdf_bytes = await self.fetch(url)
        text, page_count = self.extract_text(pdf_bytes)
        chunks = self.chunk(text)
        chunk_summaries: list[str] = []
        for index, chunk in enumerate(chunks):
            prompt = (
                "Summarise the following PDF section in 3-5 sentences, "
                "preserving key facts and figures. Do not invent details.\n\n"
                f"=== Section {index + 1}/{len(chunks)} ===\n{chunk}"
            )
            try:
                summary = await self._ollama.generate(
                    prompt,
                    system="You are a careful, factual summariser.",
                    temperature=0.2,
                )
            except OllamaError:
                self._log.warning(
                    "chunk summary failed; re-raising", chunk_index=index
                )
                raise
            chunk_summaries.append(summary.strip())

        overall = ""
        if chunk_summaries:
            joined = "\n\n".join(
                f"Section {i + 1}: {s}" for i, s in enumerate(chunk_summaries)
            )
            overall_prompt = (
                "Combine the following per-section summaries into one "
                "overall summary of 5-10 sentences.\n\n" + joined
            )
            overall = (
                await self._ollama.generate(
                    overall_prompt,
                    system="You are a careful, factual summariser.",
                    temperature=0.2,
                )
            ).strip()

        return PDFSummary(
            page_count=page_count,
            total_characters=len(text),
            chunks=chunks,
            chunk_summaries=chunk_summaries,
            overall_summary=overall,
        )


__all__ = [
    "PDFExtractError",
    "PDFProcessor",
    "PDFProcessorConfig",
    "PDFSummary",
]
