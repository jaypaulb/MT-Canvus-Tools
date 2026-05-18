"""User and group models."""

from __future__ import annotations

from ._base import CanvusModel


class User(CanvusModel):
    """A Canvus user.

    Note:
        Server v1.2 uses INTEGER user IDs, not UUID strings. Live testing
        against dev-mtcs.multitaction.com (``audit_log.author_id`` and
        ``/users/me`` rejection) verified this; see
        ``docs/api-reference/VERIFIED-CORRECTIONS.md`` §6.
    """

    id: int | None = None
    email: str
    name: str
    password: str | None = None
    admin: bool = False
    approved: bool = True
    blocked: bool = False
    created_at: str | None = None
    last_login: str | None = None
    state: str | None = None


class Group(CanvusModel):
    """A user group. Group IDs are integers."""

    id: int
    name: str
    description: str | None = None
    created_at: str | None = None
    modified_at: str | None = None
    member_count: int | None = None


class GroupMember(CanvusModel):
    """A member of a user group. Member IDs are integer user IDs."""

    id: int
    name: str
    email: str
    admin: bool = False
    approved: bool = True
    blocked: bool = False
    created_at: str | None = None
    last_login: str | None = None
    state: str | None = None


__all__ = ["Group", "GroupMember", "User"]
