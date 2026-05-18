"""Connector widget MCP tools.

Wraps ``client.widgets.connectors.*``. Connectors are non-asset widgets
(no multipart upload), so they share the note-style CRUD shape rather than
the asset helpers.
"""

from __future__ import annotations

from typing import Any

from canvus_sdk import Client

from ._helpers import (
    dump_model,
    optional_dict,
    optional_str,
    require_str,
    run_with_api_error_translation,
)
from .base import BaseMCPTool, MCPToolValidationError


def _build_endpoint(widget_id: str) -> dict[str, Any]:
    """Construct the per-endpoint dict the SDK expects.

    The Canvus API connector schema uses ``src`` / ``dst`` objects each with
    an ``id`` field. We preserve the legacy ``source_id``/``target_id`` MCP
    surface and translate here.
    """
    return {"id": widget_id, "auto_location": True}


class ConnectorCreateTool(BaseMCPTool):
    """Create a connector between two widgets."""

    def __init__(self, client: Client) -> None:
        super().__init__(
            name="connector_create",
            description="Create a visual connector between two canvas elements.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        require_str(kwargs, "source_id", self.name)
        require_str(kwargs, "target_id", self.name)
        optional_str(kwargs, "color", self.name)
        optional_str(kwargs, "style", self.name)
        optional_dict(kwargs, "payload", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        source_id = require_str(kwargs, "source_id", self.name)
        target_id = require_str(kwargs, "target_id", self.name)
        color = optional_str(kwargs, "color", self.name)
        style = optional_str(kwargs, "style", self.name)
        payload: dict[str, Any] = dict(
            optional_dict(kwargs, "payload", self.name) or {}
        )
        payload.setdefault("src", _build_endpoint(source_id))
        payload.setdefault("dst", _build_endpoint(target_id))
        if color is not None:
            payload["line_color"] = color
        if style is not None:
            payload["type"] = style
        self.logger.info(
            "connector_create",
            canvas_id=canvas_id,
            src=source_id,
            dst=target_id,
        )
        created = await run_with_api_error_translation(
            self.name,
            "Create connector",
            lambda: self.client.widgets.connectors.create(canvas_id, payload),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            "connector_id": created.id,
            "source_id": source_id,
            "target_id": target_id,
            "color": color,
            "style": style,
            "create_result": dump_model(created),
            "status": "success",
        }


class ConnectorGetTool(BaseMCPTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            name="connector_get",
            description="Retrieve detailed information about a specific connector by ID.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        require_str(kwargs, "connector_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        connector_id = require_str(kwargs, "connector_id", self.name)
        connector = await run_with_api_error_translation(
            self.name,
            "Retrieve connector",
            lambda: self.client.widgets.connectors.get(canvas_id, connector_id),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            "connector_id": connector_id,
            "connector_data": dump_model(connector),
            "status": "success",
        }


class ConnectorListTool(BaseMCPTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            name="connector_list",
            description="Retrieve a list of all connectors on a specific canvas.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        connectors = await run_with_api_error_translation(
            self.name,
            "List connectors",
            lambda: self.client.widgets.connectors.list(canvas_id),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            "connectors": dump_model(connectors),
            "count": len(connectors),
            "status": "success",
        }


class ConnectorUpdateTool(BaseMCPTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            name="connector_update",
            description="Update connector properties (color, style, endpoints).",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        require_str(kwargs, "connector_id", self.name)
        payload = optional_dict(kwargs, "payload", self.name)
        color = optional_str(kwargs, "color", self.name)
        style = optional_str(kwargs, "style", self.name)
        if not payload and color is None and style is None:
            raise MCPToolValidationError(
                "payload (or color/style) parameter is required and must be non-empty",
                self.name,
            )
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        connector_id = require_str(kwargs, "connector_id", self.name)
        payload: dict[str, Any] = dict(
            optional_dict(kwargs, "payload", self.name) or {}
        )
        color = optional_str(kwargs, "color", self.name)
        style = optional_str(kwargs, "style", self.name)
        if color is not None:
            payload["line_color"] = color
        if style is not None:
            payload["type"] = style
        updated = await run_with_api_error_translation(
            self.name,
            "Update connector",
            lambda: self.client.widgets.connectors.update(
                canvas_id, connector_id, payload
            ),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            "connector_id": connector_id,
            "payload": payload,
            "connector_data": dump_model(updated),
            "status": "success",
        }


class ConnectorDeleteTool(BaseMCPTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            name="connector_delete",
            description="Delete a specific connector from a canvas.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        require_str(kwargs, "connector_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        connector_id = require_str(kwargs, "connector_id", self.name)
        await run_with_api_error_translation(
            self.name,
            "Delete connector",
            lambda: self.client.widgets.connectors.delete(canvas_id, connector_id),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            "connector_id": connector_id,
            "status": "deleted",
        }


__all__ = [
    "ConnectorCreateTool",
    "ConnectorDeleteTool",
    "ConnectorGetTool",
    "ConnectorListTool",
    "ConnectorUpdateTool",
]
