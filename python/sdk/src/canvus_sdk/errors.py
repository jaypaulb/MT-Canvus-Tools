"""Exception hierarchy for the Canvus SDK.

All exceptions raised by the SDK derive from :class:`CanvusError`. HTTP-level
errors (i.e. anything the Canvus server returns with a non-2xx status code)
derive from :class:`APIError`. Local validation and transport errors derive
directly from :class:`CanvusError`.
"""

from __future__ import annotations

from collections.abc import Mapping, Sequence
from typing import Any


class CanvusError(Exception):
    """Base class for every error raised by ``canvus_sdk``."""


class APIError(CanvusError):
    """HTTP-level error returned by the Canvus API.

    Attributes:
        status_code: HTTP status code returned by the server.
        response_body: Raw response body, when available.
        request_id: Optional server-supplied correlation identifier.
    """

    def __init__(
        self,
        message: str,
        *,
        status_code: int,
        response_body: str | None = None,
        request_id: str | None = None,
    ) -> None:
        super().__init__(message)
        self.status_code = status_code
        self.response_body = response_body
        self.request_id = request_id


class AuthError(APIError):
    """401 / 403 from the API — authentication or authorisation failure."""


class NotFoundError(APIError):
    """404 from the API."""


class RateLimitError(APIError):
    """429 from the API. See :attr:`retry_after`."""

    def __init__(
        self,
        message: str,
        *,
        status_code: int,
        response_body: str | None = None,
        request_id: str | None = None,
        retry_after: float | None = None,
    ) -> None:
        super().__init__(
            message,
            status_code=status_code,
            response_body=response_body,
            request_id=request_id,
        )
        self.retry_after = retry_after


class ServerError(APIError):
    """5xx from the API."""


class ValidationError(CanvusError):
    """Local payload validation failed before the request was sent.

    Phase 4b §4.2 #15: carries an optional ``issues`` list — a structured
    field-level breakdown of every validation problem detected, mirroring the
    TypeScript SDK's ``ValidationError.issues`` shape. Each issue is a small
    mapping with the keys ``path`` (dotted field path), ``message`` (human
    readable), and optionally ``code`` (machine readable identifier).
    """

    def __init__(
        self,
        message: str,
        *,
        issues: Sequence[Mapping[str, Any]] | None = None,
    ) -> None:
        super().__init__(message)
        # Copy into a stable list-of-dicts so callers cannot mutate our state.
        self.issues: list[dict[str, Any]] = (
            [dict(issue) for issue in issues] if issues else []
        )


class TransportError(CanvusError):
    """Networking layer failure (DNS, TLS, connection reset, timeout)."""


class UnsupportedOperationError(CanvusError):
    """The requested operation is not supported by the Canvus server.

    Raised, for example, when attempting to create an IP Video or RDP
    Connection widget via the REST API — these widget types can only be
    created from the Canvus desktop client.
    """


__all__ = [
    "APIError",
    "AuthError",
    "CanvusError",
    "NotFoundError",
    "RateLimitError",
    "ServerError",
    "TransportError",
    "UnsupportedOperationError",
    "ValidationError",
]
