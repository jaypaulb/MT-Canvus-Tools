"""Smoke tests for the FastAPI app factory and registry wiring."""

from __future__ import annotations

import httpx
import respx
from canvus_sdk import Client
from fastapi.testclient import TestClient

from canvus_mcp_server.app import build_registry, create_app
from canvus_mcp_server.config import Settings


def _make_settings() -> Settings:
    return Settings(  # type: ignore[call-arg]
        CANVUS_API_URL="https://canvus.test.invalid/api/v1",
        CANVUS_API_KEY="ck_test",
        CANVUS_VERIFY_SSL=False,
        DATABASE_PATH="/tmp/canvus-mcp-server-test/pdf_cache.db",
        OLLAMA_ENABLED=False,
    )


def test_build_registry_registers_all_known_tools() -> None:
    settings = _make_settings()
    client = Client(
        settings.api_url, settings.api_key, verify_ssl=False, max_retries=0
    )
    registry = build_registry(client)
    # The legacy server exposed 41+ tools; we register all of them (each
    # user-visible MCP tool is preserved across the port — see audit
    # constraint).
    assert len(registry) >= 41
    # Spot-check a representative set.
    for name in (
        "canvas_get",
        "note_create",
        "image_create",
        "pdf_create",
        "connector_create",
        "user_login",
        "llm_health_check",
        "brainstorming_note_retrieval",
        "brainstorming_summary_report",
    ):
        assert name in registry, name


def test_http_health_endpoint() -> None:
    settings = _make_settings()
    client = Client(
        settings.api_url, settings.api_key, verify_ssl=False, max_retries=0
    )
    registry = build_registry(client)
    app = create_app(settings=settings, client=client, registry=registry)
    with TestClient(app) as http:
        response = http.get("/health")
        assert response.status_code == 200
        body = response.json()
        assert body["status"] == "ok"
        assert body["registered_tools"] == len(registry)


def test_http_tools_listing_and_404() -> None:
    settings = _make_settings()
    client = Client(
        settings.api_url, settings.api_key, verify_ssl=False, max_retries=0
    )
    registry = build_registry(client)
    app = create_app(settings=settings, client=client, registry=registry)
    with TestClient(app) as http:
        response = http.get("/tools")
        assert response.status_code == 200
        names = {t["name"] for t in response.json()["tools"]}
        assert "canvas_list" in names

        missing = http.get("/tools/nope")
        assert missing.status_code == 404


def test_http_execute_canvas_list_with_mocked_backend() -> None:
    settings = _make_settings()
    client = Client(
        settings.api_url, settings.api_key, verify_ssl=False, max_retries=0
    )
    registry = build_registry(client)
    app = create_app(settings=settings, client=client, registry=registry)
    with respx.mock(
        base_url=settings.api_url, assert_all_called=False
    ) as router:
        router.get("/canvases").mock(
            return_value=httpx.Response(200, json=[{"id": "x", "name": "X"}])
        )
        with TestClient(app) as http:
            response = http.post("/tools/canvas_list/execute", json={})
            assert response.status_code == 200
            data = response.json()
            assert data["data"]["count"] == 1
