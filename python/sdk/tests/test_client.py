"""Unit tests for :class:`canvus_sdk.Client` and the transport layer."""

from __future__ import annotations

import pytest
import respx
from httpx import Response

from canvus_sdk import (
    APIError,
    AuthError,
    Client,
    NotFoundError,
    ServerError,
    Settings,
    UnsupportedOperationError,
    User,
)
from canvus_sdk._http import classify_error, is_retryable, normalise_base_url

# ---- URL normalisation -----------------------------------------------------


@pytest.mark.parametrize(
    ("given", "expected"),
    [
        ("https://canvus.example.com", "https://canvus.example.com/api/v1/"),
        ("https://canvus.example.com/", "https://canvus.example.com/api/v1/"),
        (
            "https://canvus.example.com/api/v1",
            "https://canvus.example.com/api/v1/",
        ),
        (
            "https://canvus.example.com/api/v1/",
            "https://canvus.example.com/api/v1/",
        ),
    ],
)
def test_normalise_base_url(given: str, expected: str) -> None:
    """The transport always sees a URL ending in ``/api/v1/``."""
    assert normalise_base_url(given) == expected


# ---- error classification --------------------------------------------------


@pytest.mark.parametrize(
    ("status", "expected_cls"),
    [
        (401, AuthError),
        (403, AuthError),
        (404, NotFoundError),
        (418, APIError),
        (502, ServerError),
    ],
)
def test_classify_error_maps_status_codes(
    status: int, expected_cls: type[APIError]
) -> None:
    err = classify_error(status, "body", "req-1")
    assert isinstance(err, expected_cls)
    assert err.status_code == status
    assert err.response_body == "body"
    assert err.request_id == "req-1"


@pytest.mark.parametrize(
    ("status", "retryable"),
    [
        (408, True),
        (429, True),
        (500, True),
        (501, False),  # 501 Not Implemented should not be retried.
        (502, True),
        (400, False),
        (404, False),
    ],
)
def test_is_retryable(status: int, retryable: bool) -> None:
    assert is_retryable(status) is retryable


# ---- Client construction --------------------------------------------------


def test_client_exposes_all_resources() -> None:
    c = Client(base_url="https://x.invalid", api_key="k")
    assert c.canvases is not None
    assert c.folders is not None
    assert c.widgets is not None
    assert c.users is not None
    assert c.groups is not None
    assert c.auth is not None
    assert c.server is not None
    assert c.assets is not None
    # Nested widget resources should be present too.
    assert c.widgets.notes is not None
    assert c.widgets.tables is not None
    assert c.widgets.ip_videos is not None
    assert c.widgets.rdp_connections is not None


def test_client_base_url_normalised() -> None:
    c = Client(base_url="https://canvus.test.invalid", api_key="k")
    assert c.base_url == "https://canvus.test.invalid/api/v1/"


def test_from_env_uses_settings(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("CANVUS_API_URL", "https://env.test.invalid")
    monkeypatch.setenv("CANVUS_API_KEY", "env-key")
    monkeypatch.setenv("CANVUS_VERIFY_SSL", "false")
    c = Client.from_env()
    assert c.base_url == "https://env.test.invalid/api/v1/"


def test_settings_validation_requires_url_and_key(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    from pydantic import ValidationError

    monkeypatch.delenv("CANVUS_API_URL", raising=False)
    monkeypatch.delenv("CANVUS_API_KEY", raising=False)
    with pytest.raises(ValidationError):
        Settings()


# ---- happy-path HTTP via respx --------------------------------------------


@pytest.mark.asyncio
async def test_canvases_list_returns_typed_models(client: Client) -> None:
    """A 200 OK with JSON array deserialises into ``list[Canvas]``."""
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("canvases").mock(
            return_value=Response(
                200,
                json=[
                    {
                        "id": "c1",
                        "name": "First",
                        "folder_id": "root",
                        "access": "rw",
                        "asset_size": 0,
                        "in_trash": False,
                        "mode": "normal",
                        "state": "normal",
                    },
                    {
                        "id": "c2",
                        "name": "Second",
                        "folder_id": "root",
                        "access": "rw",
                        "asset_size": 0,
                        "in_trash": False,
                        "mode": "normal",
                        "state": "normal",
                    },
                ],
            )
        )

        canvases = await client.canvases.list()

    assert len(canvases) == 2
    assert canvases[0].id == "c1"
    assert canvases[1].name == "Second"


@pytest.mark.asyncio
async def test_canvases_get_404_raises_not_found(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("canvases/missing").mock(
            return_value=Response(404, json={"error": "no such canvas"})
        )
        with pytest.raises(NotFoundError) as excinfo:
            await client.canvases.get("missing")
    assert excinfo.value.status_code == 404


@pytest.mark.asyncio
async def test_create_token_uses_spec_body(client: Client) -> None:
    """Phase 3 #19: body shape is ``{name, expires?, scopes?}``."""
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        route = mock.post("users/u1/access-tokens").mock(
            return_value=Response(
                200,
                json={
                    "id": "t1",
                    "name": "ci",
                    "created_at": "2026-05-17T00:00:00Z",
                    "plain_token": "secret-token",
                },
            )
        )

        token = await client.auth.create_token(
            "u1", name="ci", scopes=["read"]
        )

    assert token.plain_token == "secret-token"
    assert route.calls[0].request.content
    import json as _json

    body = _json.loads(route.calls[0].request.content)
    assert body == {"name": "ci", "scopes": ["read"]}


@pytest.mark.asyncio
async def test_send_test_email_sends_recipient_body(client: Client) -> None:
    """Phase 3 #15: body shape is ``{recipient-email: ...}``."""
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        route = mock.post("server-config/send-test-email").mock(
            return_value=Response(200, json={"status": "ok"})
        )
        await client.server.send_test_email("dev@example.com")

    import json as _json

    body = _json.loads(route.calls[0].request.content)
    assert body == {"recipient-email": "dev@example.com"}


@pytest.mark.asyncio
async def test_install_offline_license_uses_license_key(client: Client) -> None:
    """Body shape is ``{"license": ...}`` per VERIFIED-CORRECTIONS §1 — the
    spec's ``license-data`` field name was rejected by the live server; the
    server accepts ``license``."""
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        route = mock.post("license").mock(
            return_value=Response(200, json={"status": "valid", "is_valid": True})
        )
        await client.server.install_offline_license("license-blob")

    import json as _json

    body = _json.loads(route.calls[0].request.content)
    assert body == {"license": "license-blob"}


@pytest.mark.asyncio
async def test_audit_log_translates_filter_keys(client: Client) -> None:
    """Phase 3 #18: filter keys become ``start-time``, ``end-time``, ``user-id``."""
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        route = mock.get("audit-log").mock(
            return_value=Response(
                200,
                json={
                    "events": [],
                    "total-count": 0,
                    "page": 1,
                    "per-page": 50,
                },
            )
        )
        page = await client.server.get_audit_log(
            page=1,
            per_page=50,
            start_time="2026-01-01T00:00:00Z",
            end_time="2026-02-01T00:00:00Z",
            user_id="u-123",
            action="login",
        )

    params = dict(route.calls[0].request.url.params)
    assert params["start-time"] == "2026-01-01T00:00:00Z"
    assert params["end-time"] == "2026-02-01T00:00:00Z"
    assert params["user-id"] == "u-123"
    assert params["action"] == "login"
    assert page.per_page == 50


@pytest.mark.asyncio
async def test_ip_video_create_raises_unsupported(client: Client) -> None:
    """Changelog §2: IP video POST is not supported by the server."""
    with pytest.raises(UnsupportedOperationError):
        await client.widgets.ip_videos.create("canvas-1", {})


@pytest.mark.asyncio
async def test_rdp_connection_create_raises_unsupported(client: Client) -> None:
    """Changelog §2: RDP connection POST is not supported by the server."""
    with pytest.raises(UnsupportedOperationError):
        await client.widgets.rdp_connections.create("canvas-1", {})


@pytest.mark.asyncio
async def test_clone_widget_rejects_unknown_type(client: Client) -> None:
    with pytest.raises(ValueError, match="not cloneable"):
        await client.widgets.clone(
            dest_canvas_id="c2",
            source_canvas_id="c1",
            source_widget_id="w1",
            widget_type="connector",
        )


@pytest.mark.asyncio
async def test_clone_widget_posts_to_dest_canvas_type_endpoint(
    client: Client,
) -> None:
    """Changelog §1: clone goes via standard create endpoint on destination."""
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        route = mock.post("canvases/c2/notes").mock(
            return_value=Response(200, json={"id": "new-note"})
        )
        result = await client.widgets.clone(
            dest_canvas_id="c2",
            source_canvas_id="c1",
            source_widget_id="w1",
            widget_type="note",
            location={"x": 10, "y": 20},
        )

    import json as _json

    body = _json.loads(route.calls[0].request.content)
    assert body == {
        "source_canvas_id": "c1",
        "source_widget_id": "w1",
        "location": {"x": 10, "y": 20},
    }
    assert result == {"id": "new-note"}


@pytest.mark.asyncio
async def test_table_update_warns_on_grid_size(client: Client) -> None:
    """Changelog §5: PATCH with grid_size is silently ignored — warn callers."""
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.patch("canvases/c1/tables/t1").mock(
            return_value=Response(200, json={"id": "t1"})
        )
        with pytest.warns(UserWarning, match="grid_size"):
            await client.widgets.tables.update(
                "c1", "t1", {"grid_size": {"columns": 4, "rows": 4}}
            )


@pytest.mark.asyncio
async def test_request_id_propagated_to_exception(client: Client) -> None:
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("canvases/c1").mock(
            return_value=Response(
                401,
                json={"error": "bad token"},
                headers={"x-request-id": "req-42"},
            )
        )
        with pytest.raises(AuthError) as excinfo:
            await client.canvases.get("c1")
    assert excinfo.value.request_id == "req-42"


@pytest.mark.asyncio
async def test_aclose_is_idempotent(client: Client) -> None:
    await client.aclose()
    await client.aclose()  # second call must not raise


@pytest.mark.asyncio
async def test_users_current_returns_user(client: Client) -> None:
    """``client.users.current()`` calls ``GET /users/current`` and returns a User."""
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.get("users/current").mock(
            return_value=Response(
                200,
                json={
                    "id": 42,
                    "email": "alice@example.com",
                    "name": "Alice",
                    "admin": False,
                    "approved": True,
                    "blocked": False,
                },
            )
        )
        user = await client.users.current()

    assert isinstance(user, User)
    assert user.id == 42
    assert user.email == "alice@example.com"


# ---- subscribe_buffer option (Phase 4d Round B) ----------------------------


def test_subscribe_buffer_default_is_4() -> None:
    """Client with no explicit subscribe_buffer stores 4 on the transport."""
    c = Client(base_url="https://x.invalid", api_key="k")
    assert c._transport.subscribe_buffer == 4


def test_subscribe_buffer_custom_value() -> None:
    """Client constructed with subscribe_buffer=16 propagates it to transport."""
    c = Client(base_url="https://x.invalid", api_key="k", subscribe_buffer=16)
    assert c._transport.subscribe_buffer == 16


def test_subscribe_buffer_rejects_zero() -> None:
    """subscribe_buffer < 1 must raise ValueError."""
    with pytest.raises(ValueError, match="subscribe_buffer must be >= 1"):
        Client(base_url="https://x.invalid", api_key="k", subscribe_buffer=0)


def test_subscribe_buffer_rejects_negative() -> None:
    """Negative subscribe_buffer must raise ValueError."""
    with pytest.raises(ValueError, match="subscribe_buffer must be >= 1"):
        Client(base_url="https://x.invalid", api_key="k", subscribe_buffer=-5)


def test_settings_subscribe_buffer_default() -> None:
    """Settings.subscribe_buffer defaults to 4 per Phase 4d Round B spec."""
    import os

    from canvus_sdk.config import Settings

    # Patch env to satisfy required fields without touching the real env.
    env_backup = {
        "CANVUS_API_URL": os.environ.get("CANVUS_API_URL"),
        "CANVUS_API_KEY": os.environ.get("CANVUS_API_KEY"),
    }
    os.environ["CANVUS_API_URL"] = "https://settings.test.invalid"
    os.environ["CANVUS_API_KEY"] = "test-key"
    try:
        s = Settings()  # type: ignore[call-arg]
        assert s.subscribe_buffer == 4
    finally:
        for k, v in env_backup.items():
            if v is None:
                os.environ.pop(k, None)
            else:
                os.environ[k] = v
