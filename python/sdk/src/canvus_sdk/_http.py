"""Internal HTTP transport for the Canvus SDK.

This module is private. Use :class:`canvus_sdk.Client` instead. It centralises
URL composition, authentication header injection, error classification, retry
behaviour, and JSON / binary / streaming response handling on top of
``httpx.AsyncClient``.
"""

from __future__ import annotations

import asyncio
import json
from collections.abc import AsyncIterator, Mapping
from typing import Any

import httpx
import structlog

from .errors import (
    APIError,
    AuthError,
    NotFoundError,
    RateLimitError,
    ServerError,
    TransportError,
)

logger = structlog.get_logger(__name__)


def normalise_base_url(base_url: str) -> str:
    """Normalise a Canvus base URL to end with ``/api/v1/``.

    Accepts either the bare server URL (``https://canvus.example.com``) or
    a URL that already contains the API version path. Always returns a URL
    that ends with a single trailing slash so that ``urljoin`` semantics are
    predictable.

    Args:
        base_url: Server URL as provided by the user.

    Returns:
        Normalised base URL ending in ``/api/v1/``.
    """
    url = base_url.rstrip("/")
    if url.endswith("/api/v1"):
        return url + "/"
    return url + "/api/v1/"


def classify_error(
    status_code: int,
    response_body: str,
    request_id: str | None,
) -> APIError:
    """Map an HTTP status code to the most specific :class:`APIError` subclass."""
    message = f"Canvus API error ({status_code}): {response_body[:200]}"
    kwargs: dict[str, Any] = {
        "status_code": status_code,
        "response_body": response_body,
        "request_id": request_id,
    }
    if status_code in (401, 403):
        return AuthError(message, **kwargs)
    if status_code == 404:
        return NotFoundError(message, **kwargs)
    if status_code == 429:
        return RateLimitError(message, **kwargs)
    if 500 <= status_code < 600:
        return ServerError(message, **kwargs)
    return APIError(message, **kwargs)


def is_retryable(status_code: int) -> bool:
    """Return True if a non-2xx response should trigger an automatic retry."""
    if status_code in (408, 429):
        return True
    # 501 Not Implemented is not retryable; everything else 5xx is.
    return 500 <= status_code < 600 and status_code != 501


class Transport:
    """Thin async HTTP transport with retry + error classification.

    The :class:`Transport` owns one long-lived ``httpx.AsyncClient`` per
    instance. Resource classes call its :meth:`request`, :meth:`request_bytes`,
    or :meth:`stream_lines` methods rather than constructing HTTP requests
    directly.
    """

    def __init__(
        self,
        base_url: str,
        api_key: str,
        *,
        verify_ssl: bool = True,
        connect_timeout_seconds: float = 5.0,
        request_timeout_seconds: float = 30.0,
        max_retries: int = 3,
        retry_initial_delay_seconds: float = 1.0,
        retry_backoff_factor: float = 2.0,
        subscribe_buffer: int = 4,
    ) -> None:
        self._base_url = normalise_base_url(base_url)
        self._api_key = api_key
        self._max_retries = max_retries
        self._retry_initial_delay = retry_initial_delay_seconds
        self._retry_backoff = retry_backoff_factor
        if subscribe_buffer < 1:
            raise ValueError(f"subscribe_buffer must be >= 1, got {subscribe_buffer}")
        self.subscribe_buffer: int = subscribe_buffer

        timeout = httpx.Timeout(
            connect=connect_timeout_seconds,
            read=request_timeout_seconds,
            write=request_timeout_seconds,
            pool=connect_timeout_seconds,
        )
        limits = httpx.Limits(max_connections=100, max_keepalive_connections=20)
        self._client = httpx.AsyncClient(
            base_url=self._base_url,
            headers={
                "Private-Token": api_key,
                "Accept": "application/json",
            },
            timeout=timeout,
            limits=limits,
            verify=verify_ssl,
        )

    @property
    def base_url(self) -> str:
        """The fully-normalised base URL used for all requests."""
        return self._base_url

    async def aclose(self) -> None:
        """Close the underlying ``httpx.AsyncClient``. Idempotent."""
        await self._client.aclose()

    async def __aenter__(self) -> Transport:
        return self

    async def __aexit__(self, *exc: object) -> None:
        await self.aclose()

    # ------------------------------------------------------------------ core

    async def request(
        self,
        method: str,
        path: str,
        *,
        params: Mapping[str, Any] | None = None,
        json_body: Any | None = None,
        data: Any | None = None,
        files: Any | None = None,
        headers: Mapping[str, str] | None = None,
    ) -> Any:
        """Perform a JSON request and return the decoded body (``dict``, ``list``, or ``None``).

        Args:
            method: HTTP verb (``GET``, ``POST``, ``PATCH``, ``DELETE``, ...).
            path: API path relative to the base URL, without a leading ``/``.
            params: Query string parameters.
            json_body: Object to JSON-encode as the request body.
            data: Form-encoded body. Mutually exclusive with ``json_body``.
            files: Multipart files dict for upload endpoints.
            headers: Extra request headers (merged with defaults).

        Returns:
            Decoded JSON body, or ``None`` for empty responses.

        Raises:
            APIError: For any non-2xx response after retries are exhausted.
            TransportError: For unrecoverable network failures.
        """
        response = await self._send(
            method,
            path,
            params=params,
            json_body=json_body,
            data=data,
            files=files,
            headers=headers,
        )
        if response.status_code == 204 or not response.content:
            return None
        try:
            return response.json()
        except json.JSONDecodeError:
            # Some endpoints (e.g. CSV export) decline to return JSON even on
            # success. Callers needing bytes should use request_bytes() instead.
            return response.text

    async def request_bytes(
        self,
        method: str,
        path: str,
        *,
        params: Mapping[str, Any] | None = None,
        headers: Mapping[str, str] | None = None,
    ) -> bytes:
        """Perform a request and return the raw response body as bytes."""
        response = await self._send(
            method,
            path,
            params=params,
            headers=headers,
            accept="*/*",
        )
        return response.content

    async def stream_lines(
        self,
        method: str,
        path: str,
        *,
        params: Mapping[str, Any] | None = None,
        headers: Mapping[str, str] | None = None,
    ) -> AsyncIterator[str]:
        """Yield newline-delimited response chunks from a streaming endpoint.

        The Canvus streaming endpoints (``?subscribe=true``) return
        newline-delimited JSON. Decoding into Python objects is the caller's
        responsibility.

        Yields:
            One stripped line of response body per iteration.
        """
        request_headers = self._merge_headers(headers, accept="application/json")
        async with self._client.stream(
            method,
            path,
            params=dict(params) if params else None,
            headers=request_headers,
        ) as response:
            if response.status_code >= 400:
                body = await response.aread()
                raise classify_error(
                    response.status_code,
                    body.decode("utf-8", errors="replace"),
                    response.headers.get("x-request-id"),
                )
            async for line in response.aiter_lines():
                if line:
                    yield line

    # ----------------------------------------------------------------- internals

    def _merge_headers(
        self,
        extra: Mapping[str, str] | None,
        *,
        accept: str = "application/json",
    ) -> dict[str, str]:
        headers: dict[str, str] = {"Accept": accept}
        if extra:
            headers.update(extra)
        return headers

    async def _send(
        self,
        method: str,
        path: str,
        *,
        params: Mapping[str, Any] | None = None,
        json_body: Any | None = None,
        data: Any | None = None,
        files: Any | None = None,
        headers: Mapping[str, str] | None = None,
        accept: str = "application/json",
    ) -> httpx.Response:
        request_headers = self._merge_headers(headers, accept=accept)
        attempt = 0
        delay = self._retry_initial_delay
        last_error: BaseException | None = None
        while True:
            attempt += 1
            try:
                logger.debug(
                    "canvus request",
                    method=method,
                    path=path,
                    attempt=attempt,
                )
                response = await self._client.request(
                    method,
                    path,
                    params=dict(params) if params else None,
                    json=json_body,
                    data=data,
                    files=files,
                    headers=request_headers,
                )
            except (httpx.TimeoutException, httpx.TransportError) as exc:
                last_error = exc
                if attempt > self._max_retries:
                    raise TransportError(
                        f"transport failure after {attempt} attempt(s): {exc}",
                    ) from exc
                logger.warning(
                    "canvus transport error, retrying",
                    method=method,
                    path=path,
                    attempt=attempt,
                    error=str(exc),
                    delay=delay,
                )
                await asyncio.sleep(delay)
                delay *= self._retry_backoff
                continue

            if response.status_code < 400:
                return response

            if attempt <= self._max_retries and is_retryable(response.status_code):
                logger.warning(
                    "canvus retryable error",
                    method=method,
                    path=path,
                    attempt=attempt,
                    status=response.status_code,
                    delay=delay,
                )
                await asyncio.sleep(delay)
                delay *= self._retry_backoff
                continue

            body = response.text
            request_id = response.headers.get("x-request-id")
            if response.status_code == 429:
                retry_after_raw = response.headers.get("retry-after")
                retry_after: float | None = None
                if retry_after_raw is not None:
                    try:
                        retry_after = float(retry_after_raw)
                    except ValueError:
                        retry_after = None
                raise RateLimitError(
                    f"Canvus API rate limit hit: {body[:200]}",
                    status_code=response.status_code,
                    response_body=body,
                    request_id=request_id,
                    retry_after=retry_after,
                )
            raise classify_error(response.status_code, body, request_id)

        # Unreachable; included for type-checker exhaustiveness.
        raise TransportError(
            f"request failed after {self._max_retries + 1} attempts: {last_error!r}",
        )


__all__ = ["Transport", "classify_error", "is_retryable", "normalise_base_url"]
