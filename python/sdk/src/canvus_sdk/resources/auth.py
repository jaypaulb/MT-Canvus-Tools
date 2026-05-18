"""Authentication, password-reset, registration, and access-token endpoints.

Covers ``POST /users/login``, ``POST /users/login/saml``, ``POST /users/logout``,
``POST /users/password/*``, ``POST /users/register``, ``POST /users/confirm-email``,
and the per-user ``/access-tokens`` CRUD set.
"""

from __future__ import annotations

from collections.abc import AsyncIterator
from typing import Any

from ..models import (
    AccessToken,
    AccessTokenWithSecret,
    LoginResponse,
    User,
)
from ._base import Resource


class AuthResource(Resource):
    """All authentication / session / token operations."""

    # ---- current user ------------------------------------------------------

    async def get_current_user(self) -> User:
        """Get the user attached to the current session token.

        Phase 4b §4.2 #7: ``GET /users/current``. Used by the trash helpers
        on :class:`~canvus_sdk.resources.canvases.CanvasesResource` and
        :class:`~canvus_sdk.resources.canvases.FoldersResource` to discover
        the integer user ID needed to build the canonical ``trash.{user_id}``
        folder ID.

        Returns:
            :class:`~canvus_sdk.models.User` for the authenticated session.
        """
        data = await self._transport.request("GET", "users/current")
        return self._parse(User, data)

    # ---- session -----------------------------------------------------------

    async def login(
        self,
        *,
        email: str | None = None,
        username: str | None = None,
        password: str,
    ) -> LoginResponse:
        """Log a user in.

        Per spec the body field is ``email``. Some legacy server builds also
        accept ``username``; if a caller passes both, both are sent.

        Args:
            email: User's email (preferred per spec).
            username: Legacy username key, sent only when supplied.
            password: User's password.

        Returns:
            :class:`LoginResponse` with the issued session token.
        """
        body: dict[str, Any] = {"password": password}
        if email is not None:
            body["email"] = email
        if username is not None:
            body["username"] = username
        data = await self._transport.request("POST", "users/login", json_body=body)
        return self._parse(LoginResponse, data)

    async def login_saml(
        self,
        *,
        in_response_to: str,
        response_xml: str,
        remember: bool = False,
    ) -> LoginResponse:
        """SAML login (Phase 3 Python work item #20).

        Spec body keys: ``inResponseTo`` (camelCase), ``responseXml``,
        ``remember``.
        """
        body: dict[str, Any] = {
            "inResponseTo": in_response_to,
            "responseXml": response_xml,
            "remember": remember,
        }
        data = await self._transport.request(
            "POST", "users/login/saml", json_body=body
        )
        return self._parse(LoginResponse, data)

    async def logout(self, *, token: str | None = None) -> None:
        """Log the current session out. Pass ``token`` to invalidate a specific session token."""
        headers = {"Authorization": f"Bearer {token}"} if token else None
        await self._transport.request("POST", "users/logout", headers=headers)

    # ---- password / registration ------------------------------------------

    async def request_password_reset(self, email: str) -> dict[str, Any]:
        """Request a password-reset token by email."""
        data = await self._transport.request(
            "POST",
            "users/password/create-reset-token",
            json_body={"email": email},
        )
        return data if isinstance(data, dict) else {}

    async def validate_reset_token(self, token: str) -> dict[str, Any]:
        """Check whether a password-reset token is still valid."""
        data = await self._transport.request(
            "GET",
            "users/password/validate-reset-token",
            params={"token": token},
        )
        return data if isinstance(data, dict) else {}

    async def reset_password(self, *, token: str, new_password: str) -> dict[str, Any]:
        """Complete a password reset using a previously issued token."""
        data = await self._transport.request(
            "POST",
            "users/password/reset",
            json_body={"token": token, "new-password": new_password},
        )
        return data if isinstance(data, dict) else {}

    async def register(self, payload: dict[str, Any]) -> User:
        """Register a new user (self-service signup)."""
        data = await self._transport.request("POST", "users/register", json_body=payload)
        return self._parse(User, data)

    async def confirm_email(self, token: str) -> dict[str, Any]:
        """Confirm an email-verification token."""
        data = await self._transport.request(
            "POST", "users/confirm-email", json_body={"token": token}
        )
        return data if isinstance(data, dict) else {}

    # ---- access tokens (per user) -----------------------------------------

    async def list_tokens(self, user_id: str) -> list[AccessToken]:
        """List API tokens belonging to a user.

        Note:
            ``user_id`` is typed as ``str`` (UUID) — see migration note #21.
        """
        data = await self._transport.request(
            "GET", f"users/{user_id}/access-tokens"
        )
        return self._parse_list(AccessToken, data)

    async def get_token(self, user_id: str, token_id: str) -> AccessToken:
        """Get one access token."""
        data = await self._transport.request(
            "GET", f"users/{user_id}/access-tokens/{token_id}"
        )
        return self._parse(AccessToken, data)

    async def create_token(
        self,
        user_id: str,
        *,
        name: str,
        expires: str | None = None,
        scopes: list[str] | None = None,
    ) -> AccessTokenWithSecret:
        """Create a new access token (Phase 3 Python work item #19).

        Per spec the body is ``{name, expires?, scopes?}`` — legacy SDK
        passed a single ``description`` argument; this method matches the
        documented contract.

        Args:
            user_id: User's UUID.
            name: Human-readable token name.
            expires: Optional ISO-8601 expiry timestamp.
            scopes: Optional list of permission scope strings.

        Returns:
            :class:`AccessTokenWithSecret` — the only response that includes
            ``plain_token``. Record it immediately; it cannot be retrieved
            again.
        """
        body: dict[str, Any] = {"name": name}
        if expires is not None:
            body["expires"] = expires
        if scopes is not None:
            body["scopes"] = scopes
        data = await self._transport.request(
            "POST", f"users/{user_id}/access-tokens", json_body=body
        )
        return self._parse(AccessTokenWithSecret, data)

    async def update_token(
        self, user_id: str, token_id: str, payload: dict[str, Any]
    ) -> AccessToken:
        """Update an access token's metadata."""
        data = await self._transport.request(
            "PATCH",
            f"users/{user_id}/access-tokens/{token_id}",
            json_body=payload,
        )
        return self._parse(AccessToken, data)

    async def delete_token(self, user_id: str, token_id: str) -> None:
        """Revoke an access token."""
        await self._transport.request(
            "DELETE", f"users/{user_id}/access-tokens/{token_id}"
        )

    # ---- subscribe helpers (Phase 4b §4.2 #13) -----------------------------

    def subscribe_tokens(
        self,
        user_id: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> AsyncIterator[AccessToken]:
        """Subscribe to a user's access-token list."""
        return self._typed_subscribe(
            AccessToken,
            f"users/{user_id}/access-tokens",
            params=params,
        )

    def subscribe_token(
        self,
        user_id: str,
        token_id: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> AsyncIterator[AccessToken]:
        """Subscribe to one access token."""
        return self._typed_subscribe(
            AccessToken,
            f"users/{user_id}/access-tokens/{token_id}",
            params=params,
        )


__all__ = ["AuthResource"]
