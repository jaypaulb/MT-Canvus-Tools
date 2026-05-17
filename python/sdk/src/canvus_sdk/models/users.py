"""User and group models."""

from __future__ import annotations

from ._base import CanvusModel


class User(CanvusModel):
    """A Canvus user.

    Note:
        Per spec, ``id`` is a UUID string (not an integer). The legacy
        ``CanvusPythonAPI`` typed this as ``int``; the migrated SDK uses
        ``str`` everywhere — see ``MIGRATION-NOTES.md`` work item #21.
    """

    id: str | None = None
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
    """A user group."""

    id: str
    name: str
    description: str | None = None
    created_at: str | None = None
    modified_at: str | None = None
    member_count: int | None = None


class GroupMember(CanvusModel):
    """A member of a user group."""

    id: str
    name: str
    email: str
    admin: bool = False
    approved: bool = True
    blocked: bool = False
    created_at: str | None = None
    last_login: str | None = None
    state: str | None = None


__all__ = ["Group", "GroupMember", "User"]
