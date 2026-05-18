"""Example 02: three Canvus authentication flows side-by-side.

Runs each path in sequence and reports per-path success/failure. A failure in
one path does not abort the next, so a partial setup (e.g. API key present
but login credentials absent) still produces useful output.

Flows:
    1. **API key** — `Private-Token` header (long-lived).
    2. **Email + password login** — `POST /users/login`, returns a session
       token. Sends ``email`` ONLY; sending ``username`` causes the live
       server to reject the request (see VERIFIED-CORRECTIONS.md §4).
    3. **Programmatic access-token lifecycle** — create / list / delete a
       per-user token under the user authenticated by flow 1.
"""

from __future__ import annotations

import asyncio
import sys
from datetime import UTC, datetime

import structlog
from pydantic import Field, ValidationError
from pydantic_settings import BaseSettings, SettingsConfigDict

from canvus_sdk import (
    APIError,
    AuthError,
    CanvusError,
    Client,
    configure_logging,
)

logger = structlog.get_logger(__name__)


class AuthSettings(BaseSettings):
    """Environment-backed configuration for the auth-flows example."""

    model_config = SettingsConfigDict(
        env_prefix="CANVUS_",
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    api_url: str = Field(..., description="Base Canvus URL.")
    api_key: str | None = Field(default=None, description="API key for flow 1 and flow 3.")
    email: str | None = Field(default=None, description="Login email for flow 2.")
    password: str | None = Field(default=None, description="Login password for flow 2.")


async def run_api_key(settings: AuthSettings) -> str | None:
    """Flow 1: authenticate with an API key.

    Returns the authenticated user id on success, ``None`` otherwise.
    """
    if not settings.api_key:
        logger.warning("flow_api_key skipped", reason="CANVUS_API_KEY not set")
        return None
    async with Client(settings.api_url, settings.api_key) as client:
        try:
            canvases = await client.canvases.list()
        except AuthError as e:
            logger.error("flow_api_key failed", status_code=e.status_code)
            return None
        except APIError as e:
            logger.error("flow_api_key api error", status_code=e.status_code)
            return None
        logger.info("flow_api_key ok", canvas_count=len(canvases))
        # No /users/me on v1.2, so we just report success via canvas count.
    return "api-key-authenticated"


async def run_login(settings: AuthSettings) -> str | None:
    """Flow 2: log in with email + password.

    The server rejects requests that double-key email and username; we send
    only ``email`` per VERIFIED-CORRECTIONS.md §4.
    """
    if not (settings.email and settings.password):
        logger.warning(
            "flow_login skipped",
            reason="CANVUS_EMAIL or CANVUS_PASSWORD missing",
        )
        return None
    # The Client needs *some* api key to construct; we'll use an empty one
    # because /users/login does not require auth, then swap in the issued
    # token if we want to make a follow-up call.
    async with Client(settings.api_url, api_key="bootstrap") as client:
        try:
            login = await client.auth.login(
                email=settings.email,
                password=settings.password,
            )
        except AuthError as e:
            logger.error("flow_login failed", status_code=e.status_code)
            return None
        except APIError as e:
            logger.error("flow_login api error", status_code=e.status_code)
            return None
    user_id = login.user.id if login.user else None
    logger.info("flow_login ok", user_id=user_id, has_token=bool(login.token))
    return user_id


async def run_token_lifecycle(settings: AuthSettings, user_id: str) -> None:
    """Flow 3: create + list + delete a programmatic access token.

    Requires the API key path (flow 1) to have produced a usable user id
    via a side channel — the caller passes ``user_id`` explicitly so we
    don't depend on a removed ``/users/me`` endpoint.
    """
    if not settings.api_key:
        logger.warning("flow_token_lifecycle skipped", reason="CANVUS_API_KEY not set")
        return
    token_name = f"example-02 {datetime.now(UTC).isoformat()}"
    async with Client(settings.api_url, settings.api_key) as client:
        try:
            created = await client.auth.create_token(user_id, name=token_name)
        except APIError as e:
            logger.error(
                "flow_token_lifecycle create failed",
                status_code=e.status_code,
            )
            return
        logger.info(
            "flow_token_lifecycle created",
            token_id=created.id,
            plain_token_prefix=(created.plain_token or "")[:6] + "…",
        )

        try:
            tokens = await client.auth.list_tokens(user_id)
            logger.info("flow_token_lifecycle listed", count=len(tokens))
        except APIError as e:
            logger.error("flow_token_lifecycle list failed", status_code=e.status_code)

        try:
            await client.auth.delete_token(user_id, created.id)
            logger.info("flow_token_lifecycle deleted", token_id=created.id)
        except APIError as e:
            logger.error(
                "flow_token_lifecycle delete failed",
                status_code=e.status_code,
                token_id=created.id,
            )


async def main() -> int:
    """Entry point."""
    configure_logging()
    try:
        settings = AuthSettings()  # type: ignore[call-arg]
    except ValidationError as e:
        logger.error("config invalid", errors=e.errors())
        return 2

    await run_api_key(settings)
    login_user_id = await run_login(settings)

    if login_user_id is not None:
        try:
            await run_token_lifecycle(settings, login_user_id)
        except CanvusError as e:
            logger.error("flow_token_lifecycle unexpected error", error=str(e))
    else:
        logger.warning(
            "flow_token_lifecycle skipped",
            reason="no user id available — set CANVUS_EMAIL/PASSWORD to drive flow 3",
        )
    return 0


if __name__ == "__main__":
    sys.exit(asyncio.run(main()))
