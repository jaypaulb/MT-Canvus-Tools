"""Audit log models."""

from __future__ import annotations

from typing import Any

from pydantic import Field

from ._base import CanvusModel


class AuditLogEntry(CanvusModel):
    """A single audit-log event.

    Field names mirror the wire format documented in
    ``docs/api-reference/endpoints/server.md``.
    """

    id: str | None = None
    timestamp: str | None = None
    user_id: str | None = None
    user_email: str | None = None
    action: str | None = None
    resource_type: str | None = None
    resource_id: str | None = None
    details: dict[str, Any] = Field(default_factory=dict)
    ip_address: str | None = None
    user_agent: str | None = None


class AuditLogPage(CanvusModel):
    """A page of audit log results.

    Per spec the wire format is ``{events, total-count, page, per-page}``;
    these are exposed as ``entries``, ``total_count``, ``page``, ``per_page``
    in Python.
    """

    entries: list[AuditLogEntry] = Field(default_factory=list, alias="events")
    total_count: int = Field(default=0, alias="total-count")
    page: int = 1
    per_page: int = Field(default=100, alias="per-page")


__all__ = ["AuditLogEntry", "AuditLogPage"]
