"""User and group management endpoints.

Covers ``/users`` CRUD + lifecycle methods (block/unblock/approve/delete/
change-email/force-reset-password) and ``/groups`` CRUD + members.

Note:
    All ``user_id`` and ``group_id`` parameters are typed as ``str`` (UUID)
    — see ``MIGRATION-NOTES.md`` work item #21 for the rationale.
"""

from __future__ import annotations

import builtins
from collections.abc import AsyncIterator
from typing import Any

from ..models import Group, GroupMember, User
from ._base import Resource


class UsersResource(Resource):
    """Operations on ``/users``."""

    async def list(self, *, params: dict[str, Any] | None = None) -> list[User]:
        """List all users."""
        data = await self._transport.request("GET", "users", params=params)
        return self._parse_list(User, data)

    async def get(self, user_id: str) -> User:
        """Get one user."""
        data = await self._transport.request("GET", f"users/{user_id}")
        return self._parse(User, data)

    async def create(self, payload: dict[str, Any]) -> User:
        """Create a user (admin operation)."""
        data = await self._transport.request("POST", "users", json_body=payload)
        return self._parse(User, data)

    async def update(self, user_id: str, payload: dict[str, Any]) -> User:
        """Update a user."""
        data = await self._transport.request(
            "PATCH", f"users/{user_id}", json_body=payload
        )
        return self._parse(User, data)

    async def delete(self, user_id: str) -> None:
        """Delete a user."""
        await self._transport.request("DELETE", f"users/{user_id}")

    # ---- lifecycle ---------------------------------------------------------

    async def block(self, user_id: str) -> User:
        """Block a user."""
        data = await self._transport.request("POST", f"users/{user_id}/block")
        return self._parse(User, data)

    async def unblock(self, user_id: str) -> User:
        """Unblock a user."""
        data = await self._transport.request("POST", f"users/{user_id}/unblock")
        return self._parse(User, data)

    async def approve(self, user_id: str) -> User:
        """Approve a pending user."""
        data = await self._transport.request("POST", f"users/{user_id}/approve")
        return self._parse(User, data)

    # ---- credentials -------------------------------------------------------

    async def change_password(
        self,
        user_id: str,
        *,
        new_password: str,
        old_password: str | None = None,
    ) -> dict[str, Any]:
        """Change a user's password.

        If ``old_password`` is supplied, the request matches the spec's
        self-change body shape (``{old-password, new-password}``); otherwise
        the legacy admin-set form (``{password}``) is used.
        """
        body: dict[str, Any]
        if old_password is not None:
            body = {"old-password": old_password, "new-password": new_password}
        else:
            body = {"password": new_password}
        data = await self._transport.request(
            "POST", f"users/{user_id}/password", json_body=body
        )
        return data if isinstance(data, dict) else {}

    async def change_email(self, user_id: str, new_email: str) -> dict[str, Any]:
        """Change a user's email address (Phase 3 Python work item #7).

        POSTs ``{new-email}`` to ``/users/{user_id}/change-email``.
        """
        data = await self._transport.request(
            "POST",
            f"users/{user_id}/change-email",
            json_body={"new-email": new_email},
        )
        return data if isinstance(data, dict) else {}

    def subscribe(
        self, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[User]:
        """Subscribe to ``/users?subscribe=true`` (Phase 4b §4.2 #13)."""
        return self._typed_subscribe(User, "users", params=params)

    def subscribe_one(
        self, user_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[User]:
        """Subscribe to a single user."""
        return self._typed_subscribe(User, f"users/{user_id}", params=params)

    async def force_reset_password(self, user_id: str) -> dict[str, Any]:
        """Force-reset a user's password (admin; Phase 3 Python work item #8).

        POSTs an empty body to ``/users/{user_id}/reset-password``. Do not
        confuse this with the self-service password-reset flow under
        :class:`AuthResource`.
        """
        data = await self._transport.request(
            "POST", f"users/{user_id}/reset-password", json_body={}
        )
        return data if isinstance(data, dict) else {}

    async def current(self) -> User:
        """Return the user record for the active session credentials.

        Calls ``GET /users/current``.
        """
        data = await self._transport.request("GET", "users/current")
        return self._parse(User, data)


class GroupsResource(Resource):
    """Operations on ``/groups``."""

    async def list(self, *, params: dict[str, Any] | None = None) -> list[Group]:
        """List groups."""
        data = await self._transport.request("GET", "groups", params=params)
        return self._parse_list(Group, data)

    async def get(self, group_id: str) -> Group:
        """Get a group."""
        data = await self._transport.request("GET", f"groups/{group_id}")
        return self._parse(Group, data)

    async def create(self, payload: dict[str, Any]) -> Group:
        """Create a new group."""
        data = await self._transport.request("POST", "groups", json_body=payload)
        return self._parse(Group, data)

    async def update(self, group_id: str, payload: dict[str, Any]) -> Group:
        """Update a group (Phase 3 Python work item #9).

        The legacy SDK was missing the PATCH method; it is now exposed.
        """
        data = await self._transport.request(
            "PATCH", f"groups/{group_id}", json_body=payload
        )
        return self._parse(Group, data)

    async def delete(self, group_id: str) -> None:
        """Delete a group."""
        await self._transport.request("DELETE", f"groups/{group_id}")

    # ---- members -----------------------------------------------------------

    async def list_members(self, group_id: str) -> builtins.list[GroupMember]:
        """List members of a group."""
        data = await self._transport.request("GET", f"groups/{group_id}/members")
        return self._parse_list(GroupMember, data)

    async def add_member(self, group_id: str, user_id: str) -> dict[str, Any]:
        """Add a user to a group."""
        data = await self._transport.request(
            "POST",
            f"groups/{group_id}/members",
            json_body={"user_id": user_id},
        )
        return data if isinstance(data, dict) else {}

    async def remove_member(self, group_id: str, user_id: str) -> None:
        """Remove a user from a group."""
        await self._transport.request(
            "DELETE", f"groups/{group_id}/members/{user_id}"
        )

    # ---- subscribe helpers (Phase 4b §4.2 #13) -----------------------------

    def subscribe(
        self, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[Group]:
        """Subscribe to ``/groups?subscribe=true``."""
        return self._typed_subscribe(Group, "groups", params=params)

    def subscribe_one(
        self, group_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[Group]:
        """Subscribe to a single group."""
        return self._typed_subscribe(Group, f"groups/{group_id}", params=params)

    def subscribe_members(
        self, group_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[GroupMember]:
        """Subscribe to a group's member list."""
        return self._typed_subscribe(
            GroupMember, f"groups/{group_id}/members", params=params
        )


__all__ = ["GroupsResource", "UsersResource"]
