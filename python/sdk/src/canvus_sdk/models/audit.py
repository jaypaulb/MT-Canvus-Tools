"""Audit log models."""

from __future__ import annotations

from pydantic import Field

from ._base import CanvusModel


class AuditLogEntry(CanvusModel):
    """A single audit-log event.

    Per live-server verification (v1.2), the wire format returns:
    - ``id`` (int), ``action`` (str), ``author_id`` (int | null), ``created_at`` (str),
      ``details`` (str — JSON-encoded), ``ip_address`` (str), ``target_id`` (str | null),
      ``target_type`` (str).
    Legacy fields ``resource_type``, ``resource_id``, ``user_email``, ``user_agent``,
    ``timestamp`` are not present; use ``created_at`` instead.
    """

    id: int | None = None
    action: str | None = None
    author_id: int | None = None
    created_at: str | None = None
    details: str = ""
    ip_address: str | None = None
    target_id: str | None = None
    target_type: str | None = None


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
