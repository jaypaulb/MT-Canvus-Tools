"""Server-wide endpoints: server-info, server-config, license, audit log, clients/workspaces.

Covers everything under:

- ``/server-info``
- ``/server-config`` + ``/server-config/send-test-email`` + ``/server-config/reload-certs``
- ``/license`` (info, install, activate, request offline)
- ``/audit-log`` + ``/audit-log/export-csv``
- ``/clients`` (read-only listing per spec; legacy CRUD intentionally omitted)
- ``/clients/{id}/workspaces`` (incl. open-canvas action)
- ``/clients/{id}/video-outputs`` and ``/clients/{id}/video-inputs``
"""

from __future__ import annotations

from typing import Any

from ..models import (
    AuditLogPage,
    ClientInfo,
    LicenseActivationRequest,
    LicenseInfo,
    ServerConfig,
    ServerInfo,
    VideoOutput,
    Workspace,
)
from ._base import Resource


class ServerResource(Resource):
    """Server-wide operations."""

    # ---- info / config -----------------------------------------------------

    async def get_info(self) -> ServerInfo:
        """Get server info (version, API surface, etc.)."""
        data = await self._transport.request("GET", "server-info")
        return self._parse(ServerInfo, data)

    async def get_config(self) -> ServerConfig:
        """Get server configuration."""
        data = await self._transport.request("GET", "server-config")
        # The wire format may be either a nested dict (legacy) or a flat
        # element array (per spec). We accept both and surface what we can.
        if isinstance(data, list):
            flat: dict[str, Any] = {}
            for item in data:
                if isinstance(item, dict) and "key" in item and "value" in item:
                    flat[item["key"]] = item["value"]
            return ServerConfig.model_validate(flat)
        return self._parse(ServerConfig, data or {})

    async def update_config(self, payload: dict[str, Any]) -> ServerConfig:
        """Update server configuration (PATCH)."""
        data = await self._transport.request(
            "PATCH", "server-config", json_body=payload
        )
        return self._parse(ServerConfig, data or {})

    async def send_test_email(self, recipient_email: str) -> dict[str, Any]:
        """Send a test email (Phase 3 Python work item #15).

        Per spec, body shape is ``{recipient-email: str}``; the legacy SDK
        sent no body and would 400.
        """
        data = await self._transport.request(
            "POST",
            "server-config/send-test-email",
            json_body={"recipient-email": recipient_email},
        )
        return data if isinstance(data, dict) else {}

    async def reload_certs(self) -> dict[str, Any]:
        """Reload TLS certificates (Phase 3 Python work item #10)."""
        data = await self._transport.request(
            "POST", "server-config/reload-certs", json_body={}
        )
        return data if isinstance(data, dict) else {}

    # ---- license -----------------------------------------------------------

    async def get_license_info(self) -> LicenseInfo:
        """Get current license information."""
        data = await self._transport.request("GET", "license")
        return self._parse(LicenseInfo, data)

    async def request_offline_activation(self) -> LicenseActivationRequest:
        """Get the offline activation request blob.

        Phase 3 Python work item #17: the legacy SDK sent an erroneous
        ``?key=`` query parameter that is not documented in the spec; it has
        been removed here.
        """
        data = await self._transport.request("GET", "license/request")
        return self._parse(LicenseActivationRequest, data or {})

    async def install_offline_license(self, license_data: str) -> LicenseInfo:
        """Install an offline license blob (Phase 3 Python work item #16).

        Per spec, the body field is ``license-data`` (hyphenated); the legacy
        SDK sent ``license``.
        """
        data = await self._transport.request(
            "POST", "license", json_body={"license-data": license_data}
        )
        return self._parse(LicenseInfo, data)

    async def activate_license(self) -> LicenseInfo:
        """Online license activation (Phase 3 Python work item #11).

        Per spec the body is empty.
        """
        data = await self._transport.request("POST", "license/activate", json_body={})
        return self._parse(LicenseInfo, data)

    # ---- audit log ---------------------------------------------------------

    async def get_audit_log(
        self,
        *,
        page: int | None = None,
        per_page: int | None = None,
        filter: str | None = None,
        start_time: str | None = None,
        end_time: str | None = None,
        user_id: str | None = None,
        action: str | None = None,
    ) -> AuditLogPage:
        """Read the audit log (Phase 3 Python work item #18).

        Args:
            page: 1-based page index.
            per_page: Number of entries per page.
            filter: Free-text filter string.
            start_time: ISO-8601 start of the time window (wire key
                ``start-time``).
            end_time: ISO-8601 end of the time window (wire key ``end-time``).
            user_id: Filter to a single user (wire key ``user-id``).
            action: Filter to a single action verb.

        Returns:
            :class:`AuditLogPage` containing entries, total count, and
            pagination metadata.
        """
        params: dict[str, Any] = {}
        if page is not None:
            params["page"] = page
        if per_page is not None:
            params["per-page"] = per_page
        if filter is not None:
            params["filter"] = filter
        if start_time is not None:
            params["start-time"] = start_time
        if end_time is not None:
            params["end-time"] = end_time
        if user_id is not None:
            params["user-id"] = user_id
        if action is not None:
            params["action"] = action

        data = await self._transport.request("GET", "audit-log", params=params)
        # The spec envelope is `{events, total-count, page, per-page}`.
        # The legacy server returns a flat list; handle both.
        if isinstance(data, list):
            return AuditLogPage(events=data, total_count=len(data))
        return self._parse(AuditLogPage, data or {})

    async def export_audit_log_csv(
        self,
        *,
        start_time: str | None = None,
        end_time: str | None = None,
        user_id: str | None = None,
        action: str | None = None,
        filter: str | None = None,
    ) -> bytes:
        """Export the audit log as a CSV. Accepts the same filters as ``get_audit_log``."""
        params: dict[str, Any] = {}
        if start_time is not None:
            params["start-time"] = start_time
        if end_time is not None:
            params["end-time"] = end_time
        if user_id is not None:
            params["user-id"] = user_id
        if action is not None:
            params["action"] = action
        if filter is not None:
            params["filter"] = filter
        return await self._transport.request_bytes(
            "GET", "audit-log/export-csv", params=params
        )

    # ---- clients -----------------------------------------------------------

    async def list_clients(self) -> list[ClientInfo]:
        """List Canvus desktop clients connected to the server."""
        data = await self._transport.request("GET", "clients")
        return self._parse_list(ClientInfo, data)

    async def get_client(self, client_id: str) -> ClientInfo:
        """Get a single Canvus desktop client."""
        data = await self._transport.request("GET", f"clients/{client_id}")
        return self._parse(ClientInfo, data)

    # ---- workspaces --------------------------------------------------------

    async def list_workspaces(self, client_id: str) -> list[Workspace]:
        """List workspaces on a Canvus client."""
        data = await self._transport.request("GET", f"clients/{client_id}/workspaces")
        return self._parse_list(Workspace, data)

    async def get_workspace(self, client_id: str, workspace_id: str) -> Workspace:
        """Get a single workspace.

        Note:
            ``workspace_id`` is typed as ``str`` (UUID) per spec —
            see migration note #21.
        """
        data = await self._transport.request(
            "GET", f"clients/{client_id}/workspaces/{workspace_id}"
        )
        return self._parse(Workspace, data)

    async def update_workspace(
        self,
        client_id: str,
        workspace_id: str,
        payload: dict[str, Any],
    ) -> Workspace:
        """Update a workspace."""
        data = await self._transport.request(
            "PATCH",
            f"clients/{client_id}/workspaces/{workspace_id}",
            json_body=payload,
        )
        return self._parse(Workspace, data)

    async def open_canvas_in_workspace(
        self,
        client_id: str,
        workspace_id: str,
        canvas_id: str,
    ) -> dict[str, Any]:
        """Open a canvas on a specific workspace (Phase 3 Python work item #12).

        POSTs ``{canvas-id: canvas_id}`` to
        ``/clients/{client_id}/workspaces/{workspace_id}/open-canvas``.
        """
        data = await self._transport.request(
            "POST",
            f"clients/{client_id}/workspaces/{workspace_id}/open-canvas",
            json_body={"canvas-id": canvas_id},
        )
        return data if isinstance(data, dict) else {}

    # ---- video outputs (client-scoped) -------------------------------------

    async def list_client_video_outputs(self, client_id: str) -> list[VideoOutput]:
        """List a client's video outputs."""
        data = await self._transport.request(
            "GET", f"clients/{client_id}/video-outputs"
        )
        return self._parse_list(VideoOutput, data)

    async def get_client_video_output(
        self, client_id: str, output_id: str
    ) -> VideoOutput:
        """Get one video output (Phase 3 Python work item #13)."""
        data = await self._transport.request(
            "GET", f"clients/{client_id}/video-outputs/{output_id}"
        )
        return self._parse(VideoOutput, data)

    async def set_video_output_source(
        self,
        client_id: str,
        output_id: str,
        payload: dict[str, Any],
    ) -> VideoOutput:
        """Set the source of a client video output (Phase 3 Python work item #14).

        Correctly targets ``PATCH /clients/{cid}/video-outputs/{oid}``. The
        legacy ``update_video_output`` method hit a non-existent
        ``/canvases/.../video-outputs/...`` path; that bug is fixed here by
        omitting the broken method entirely.
        """
        data = await self._transport.request(
            "PATCH",
            f"clients/{client_id}/video-outputs/{output_id}",
            json_body=payload,
        )
        return self._parse(VideoOutput, data)

    # ---- video inputs (client-scoped) --------------------------------------

    async def list_client_video_inputs(self, client_id: str) -> list[dict[str, Any]]:
        """List a client's video inputs (raw dicts; no typed model in spec)."""
        data = await self._transport.request(
            "GET", f"clients/{client_id}/video-inputs"
        )
        return list(data) if isinstance(data, list) else []

    async def get_client_video_input(
        self, client_id: str, input_id: str
    ) -> dict[str, Any]:
        """Get one client video input (Phase 3 Python work item #13)."""
        data = await self._transport.request(
            "GET", f"clients/{client_id}/video-inputs/{input_id}"
        )
        return data if isinstance(data, dict) else {}


__all__ = ["ServerResource"]
