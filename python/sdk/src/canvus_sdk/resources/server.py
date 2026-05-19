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

from collections.abc import AsyncIterator
from typing import Any

from ..errors import ValidationError
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


def _flatten_config(
    obj: dict[str, Any],
    prefix: str = "",
) -> list[dict[str, Any]]:
    """Walk a nested server-config dict and produce flat element triples.

    Each leaf becomes a ``{setting-key, setting-value, setting-type}`` entry
    where ``setting-key`` is a dotted path. The triple shape matches the
    ``PATCH /server-config`` write form documented in
    ``docs/api-reference/endpoints/server.md``.
    """
    out: list[dict[str, Any]] = []
    for key, value in obj.items():
        composite = f"{prefix}.{key}" if prefix else key
        if isinstance(value, dict):
            out.extend(_flatten_config(value, composite))
        else:
            out.append(
                {
                    "setting-key": composite,
                    "setting-value": value,
                    "setting-type": type(value).__name__,
                }
            )
    return out


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

    async def get_config_raw(self) -> list[dict[str, Any]]:
        """Return the server config as the spec's flat element-array form.

        Phase 4b §4.2 #9: mirrors Go's ``serverconfig.go:89 GetServerConfigRaw``.
        Per VERIFIED-CORRECTIONS §3, the v1.2 server actually returns a
        deeply nested object on ``GET /server-config``; the documented flat
        ``[{setting-key, setting-value, setting-type}, …]`` form is the shape
        the spec'd PATCH writer expects. This helper recursively flattens the
        nested response so callers can hand the result straight back to
        ``update_config({"settings": result})`` after edits.

        Returns:
            A flat list of ``{"setting-key", "setting-value", "setting-type"}``
            mappings.
        """
        data = await self._transport.request("GET", "server-config")
        # Honour either wire shape: pass through if already flat, otherwise
        # walk the nested dict and emit one element per leaf.
        if isinstance(data, list):
            return [dict(item) for item in data if isinstance(item, dict)]
        if not isinstance(data, dict):
            return []
        return _flatten_config(data)

    async def set_video_output_source_by_index(
        self,
        client_id: str,
        index: int,
        source: dict[str, Any] | str,
    ) -> dict[str, Any]:
        """Set a client's video-output source by integer index.

        Phase 4b §4.2 #10: mirrors Go's
        ``videooutputs.go:28 SetVideoOutputSource``. The hyphenated
        ``video-outputs/{index}`` path accepts either a numeric index or a
        UUID — :meth:`set_video_output_source` is the by-ID variant.

        Args:
            client_id: ID of the Canvus desktop client.
            index: Zero-based output index on that client.
            source: Either a full PATCH body dict or a bare source identifier
                string (auto-wrapped into ``{"source": source}``).
        """
        body: dict[str, Any] = source if isinstance(source, dict) else {"source": source}
        data = await self._transport.request(
            "PATCH",
            f"clients/{client_id}/video-outputs/{index}",
            json_body=body,
        )
        return data if isinstance(data, dict) else {}

    # ---- workspace orchestration (Phase 4b §4.2 #11-#12) -------------------

    async def toggle_workspace_info_panel(
        self,
        client_id: str,
        workspace_id: str,
    ) -> Workspace:
        """Flip ``info_panel_visible`` for a workspace.

        Phase 4b §4.2 #11: mirrors Go's ``workspaces.go:79``. GETs the
        workspace to read the current value, then PATCHes the negation.
        """
        ws = await self.get_workspace(client_id, workspace_id)
        new_value = not bool(ws.info_panel_visible)
        return await self.update_workspace(
            client_id, workspace_id, {"info_panel_visible": new_value}
        )

    async def toggle_workspace_pinned(
        self,
        client_id: str,
        workspace_id: str,
    ) -> Workspace:
        """Flip ``pinned`` for a workspace.

        Phase 4b §4.2 #11: mirrors Go's ``workspaces.go:90``.
        """
        ws = await self.get_workspace(client_id, workspace_id)
        new_value = not bool(ws.pinned)
        return await self.update_workspace(
            client_id, workspace_id, {"pinned": new_value}
        )

    async def set_workspace_viewport(
        self,
        client_id: str,
        workspace_id: str,
        *,
        widget_canvas_id: str | None = None,
        widget_id: str | None = None,
        rect: dict[str, float] | None = None,
        margin: float = 20.0,
    ) -> Workspace:
        """Set a workspace's ``view_rectangle`` either by widget or by raw rect.

        Phase 4b §4.2 #12: mirrors Go's ``workspaces.go:102 SetWorkspaceViewport``.

        Exactly one of (``widget_canvas_id`` + ``widget_id``) or ``rect``
        must be provided. When a widget is supplied, the SDK fetches it and
        builds a viewport that includes its bounds plus ``margin`` pixels of
        breathing room on each side.

        Args:
            client_id: Owning Canvus client ID.
            workspace_id: ID of the workspace to update.
            widget_canvas_id: Canvas ID hosting the widget to centre on.
            widget_id: Widget ID to centre on.
            rect: Explicit ``{x, y, width, height}`` viewport rectangle.
            margin: Padding (pixels) when computing the rect from a widget.

        Raises:
            ValidationError: Neither a widget reference nor an explicit rect
                was supplied (or both were).
        """
        if rect is not None and widget_id is not None:
            raise ValidationError(
                "set_workspace_viewport: pass either rect= or (widget_canvas_id+"
                "widget_id), not both.",
            )
        if rect is None:
            if widget_canvas_id is None or widget_id is None:
                raise ValidationError(
                    "set_workspace_viewport: must provide rect= or both "
                    "widget_canvas_id= and widget_id=.",
                )
            widget_data = await self._transport.request(
                "GET",
                f"canvases/{widget_canvas_id}/widgets/{widget_id}",
            )
            if not isinstance(widget_data, dict):
                raise ValidationError(
                    f"set_workspace_viewport: widget {widget_id!r} fetch returned "
                    f"{type(widget_data).__name__}, expected dict.",
                )
            location = widget_data.get("location") or {}
            size = widget_data.get("size") or {}
            x = float(location.get("x", 0.0)) - margin
            y = float(location.get("y", 0.0)) - margin
            width = float(size.get("width", 0.0)) + 2.0 * margin
            height = float(size.get("height", 0.0)) + 2.0 * margin
            rect = {"x": x, "y": y, "width": width, "height": height}
        return await self.update_workspace(
            client_id, workspace_id, {"view_rectangle": rect}
        )

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

        Per live-server verification (v1.2), the body field is ``license`` (not
        ``license-data``); the endpoint is ``POST /license`` (not
        ``/license/install``).
        """
        data = await self._transport.request(
            "POST", "license", json_body={"license": license_data}
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
        user_id: int | None = None,
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
            return AuditLogPage.model_validate({"events": data, "total-count": len(data)})
        return self._parse(AuditLogPage, data or {})

    async def export_audit_log_csv(
        self,
        *,
        start_time: str | None = None,
        end_time: str | None = None,
        user_id: int | None = None,
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

    # ---- subscribe helpers (Phase 4b §4.2 #13) -----------------------------

    def subscribe_config(
        self, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[ServerConfig]:
        """Subscribe to ``/server-config?subscribe=true``."""
        return self._typed_subscribe(ServerConfig, "server-config", params=params)

    def subscribe_license(
        self, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[LicenseInfo]:
        """Subscribe to ``/license?subscribe=true``."""
        return self._typed_subscribe(LicenseInfo, "license", params=params)

    def subscribe_clients(
        self, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[ClientInfo]:
        """Subscribe to ``/clients?subscribe=true``."""
        return self._typed_subscribe(ClientInfo, "clients", params=params)

    def subscribe_client(
        self, client_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[ClientInfo]:
        """Subscribe to a single client."""
        return self._typed_subscribe(
            ClientInfo, f"clients/{client_id}", params=params
        )

    def subscribe_workspaces(
        self, client_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[Workspace]:
        """Subscribe to a client's workspaces."""
        return self._typed_subscribe(
            Workspace, f"clients/{client_id}/workspaces", params=params
        )

    def subscribe_workspace(
        self,
        client_id: str,
        workspace_id: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> AsyncIterator[Workspace]:
        """Subscribe to a single workspace."""
        return self._typed_subscribe(
            Workspace,
            f"clients/{client_id}/workspaces/{workspace_id}",
            params=params,
        )

    def subscribe_video_outputs(
        self, client_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[VideoOutput]:
        """Subscribe to a client's video outputs."""
        return self._typed_subscribe(
            VideoOutput, f"clients/{client_id}/video-outputs", params=params
        )

    def subscribe_video_output(
        self,
        client_id: str,
        output_id: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> AsyncIterator[VideoOutput]:
        """Subscribe to one client video output."""
        return self._typed_subscribe(
            VideoOutput,
            f"clients/{client_id}/video-outputs/{output_id}",
            params=params,
        )

    def subscribe_video_inputs(
        self, client_id: str, *, params: dict[str, Any] | None = None
    ) -> AsyncIterator[dict[str, Any]]:
        """Subscribe to a client's video inputs (raw dicts; no first-class model)."""
        return self._raw_subscribe(
            f"clients/{client_id}/video-inputs", params=params
        )

    def subscribe_video_input(
        self,
        client_id: str,
        input_id: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> AsyncIterator[dict[str, Any]]:
        """Subscribe to one client video input (raw dicts)."""
        return self._raw_subscribe(
            f"clients/{client_id}/video-inputs/{input_id}", params=params
        )


__all__ = ["ServerResource"]
