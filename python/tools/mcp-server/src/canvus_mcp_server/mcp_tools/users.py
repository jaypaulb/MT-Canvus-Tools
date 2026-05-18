"""User management MCP tools.

Wraps ``client.auth.*``, ``client.users.*``, and ``client.canvases.get_permissions``.

Note on auth-tool semantics: ``user_login`` and ``user_logout`` invoke the
SDK's session-scoped auth endpoints — they mutate the *current session's*
authentication state. In an MCP-server context where the SDK client is a
process-wide singleton, calling these tools changes the credentials of
*every subsequent tool execution*. The legacy server had the same behaviour;
the audit's Phase 4d should consider per-request session scoping.
"""

from __future__ import annotations

from typing import Any

from canvus_sdk import Client

from ._helpers import (
    dump_model,
    optional_dict,
    require_str,
    run_with_api_error_translation,
)
from .base import BaseMCPTool, MCPToolValidationError


class UserLoginTool(BaseMCPTool):
    """Authenticate as a specific user (email + password)."""

    def __init__(self, client: Client) -> None:
        super().__init__(
            name="user_login",
            description="Authenticate and login as a specific user (email + password).",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        # Accept legacy "username" as an alias for the canonical "email"
        # since the Canvus API only accepts {email, password} per
        # VERIFIED-CORRECTIONS §4.
        email_or_username = kwargs.get("email") or kwargs.get("username")
        if not isinstance(email_or_username, str) or not email_or_username.strip():
            raise MCPToolValidationError(
                "email (or legacy username) parameter is required", self.name
            )
        require_str(kwargs, "password", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        # Per VERIFIED-CORRECTIONS §4 the server rejects requests carrying
        # both `email` and `username`; we always normalise to `email`.
        email_value = kwargs.get("email") or kwargs["username"]
        if not isinstance(email_value, str):
            raise MCPToolValidationError(
                "email/username must be a string", self.name
            )
        password = require_str(kwargs, "password", self.name)
        self.logger.info("user_login", email=email_value)
        login_result = await run_with_api_error_translation(
            self.name,
            "Login",
            lambda: self.client.auth.login(email=email_value, password=password),
        )
        return {
            "success": True,
            "message": f"Successfully logged in as {email_value}",
            "login_result": dump_model(login_result),
        }


class UserLogoutTool(BaseMCPTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            name="user_logout",
            description="Logout the current user session.",
        )
        self.client = client

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        self.logger.info("user_logout")
        result = await run_with_api_error_translation(
            self.name,
            "Logout",
            lambda: self.client.auth.logout(),
        )
        return {
            "success": True,
            "message": "Successfully logged out",
            "logout_result": dump_model(result),
        }


class GetCurrentUserTool(BaseMCPTool):
    """Return the user record for the credentials in use.

    Backed by ``GET /users/current``; the SDK does not yet expose a typed
    helper (parity-matrix §1.1.6 — current state ❌). Use the raw transport.
    """

    def __init__(self, client: Client) -> None:
        super().__init__(
            name="get_current_user",
            description="Get information about the currently authenticated user.",
        )
        self.client = client

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        self.logger.info("get_current_user")
        # The transport is a private attribute on the Client — using it
        # here is a documented escape hatch; the audit calls this out as a
        # Phase 4d SDK gap ("Get current user (`GET /users/current`)").
        raw = await run_with_api_error_translation(
            self.name,
            "Get current user",
            lambda: self.client._transport.request("GET", "users/current"),
        )
        return {"success": True, "user": dump_model(raw)}


class UserListTool(BaseMCPTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            name="user_list",
            description="List all users in the system (requires admin privileges).",
        )
        self.client = client

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        self.logger.info("user_list")
        users = await run_with_api_error_translation(
            self.name,
            "List users",
            lambda: self.client.users.list(),
        )
        return {
            "success": True,
            "users": dump_model(users),
            "count": len(users),
        }


class UserCreateTool(BaseMCPTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            name="user_create",
            description="Create a new user account.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        user_data = optional_dict(kwargs, "user_data", self.name)
        if user_data is None:
            raise MCPToolValidationError(
                "user_data parameter is required", self.name
            )
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        user_data = optional_dict(kwargs, "user_data", self.name) or {}
        username = user_data.get("username") or user_data.get("email")
        self.logger.info("user_create", username=username)
        created = await run_with_api_error_translation(
            self.name,
            "Create user",
            lambda: self.client.users.create(user_data),
        )
        return {
            "success": True,
            "message": f"Successfully created user: {username}",
            "user": dump_model(created),
        }


class GetCanvasPermissionsTool(BaseMCPTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            name="get_canvas_permissions",
            description="Get permissions for a specific canvas.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        self.logger.info("get_canvas_permissions", canvas_id=canvas_id)
        perms = await run_with_api_error_translation(
            self.name,
            "Get canvas permissions",
            lambda: self.client.canvases.get_permissions(canvas_id),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            "permissions": dump_model(perms),
        }


__all__ = [
    "GetCanvasPermissionsTool",
    "GetCurrentUserTool",
    "UserCreateTool",
    "UserListTool",
    "UserLoginTool",
    "UserLogoutTool",
]
