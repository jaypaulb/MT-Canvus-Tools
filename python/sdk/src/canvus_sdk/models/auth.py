"""Authentication and access-token models."""

from __future__ import annotations

from ._base import CanvusModel
from .users import User


class AccessToken(CanvusModel):
    """An API access token. ``plain_token`` is only returned on creation."""

    id: str
    name: str | None = None
    description: str | None = None
    created_at: str | None = None
    expires: str | None = None
    scopes: list[str] | None = None
    plain_token: str | None = None


class AccessTokenWithSecret(AccessToken):
    """The variant returned by ``POST /users/{user-id}/access-tokens``.

    Always includes the ``plain_token`` field that gives the caller a single
    chance to record the secret.
    """

    plain_token: str


class LoginResponse(CanvusModel):
    """Response from ``POST /users/login``.

    The server returns the user object plus a short-lived session token.
    """

    user: User | None = None
    token: str | None = None


__all__ = ["AccessToken", "AccessTokenWithSecret", "LoginResponse"]
