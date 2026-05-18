"""Canvas-level MCP tools.

Wraps ``client.canvases.*`` from the new Python SDK. Decomposed out of the
legacy 3,169-LOC ``mcp_tools/canvas.py`` per the Phase 4c audit; this file
holds only canvas-resource tools (CRUD + copy/move).
"""

from __future__ import annotations

from typing import Any

from canvus_sdk import Client

from ._helpers import (
    dump_model,
    optional_str,
    require_str,
    run_with_api_error_translation,
)
from .base import BaseMCPTool


class CanvasGetTool(BaseMCPTool):
    """Retrieve a single canvas by ID."""

    def __init__(self, client: Client) -> None:
        super().__init__(
            name="canvas_get",
            description="Retrieve detailed information about a specific canvas by ID.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        self.logger.info("canvas_get", canvas_id=canvas_id)
        canvas = await run_with_api_error_translation(
            self.name,
            "Retrieve canvas",
            lambda: self.client.canvases.get(canvas_id),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            "canvas_data": dump_model(canvas),
            "status": "success",
        }


class CanvasListTool(BaseMCPTool):
    """List all canvases visible to the caller."""

    def __init__(self, client: Client) -> None:
        super().__init__(
            name="canvas_list",
            description="Retrieve a list of all available canvases.",
        )
        self.client = client

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        self.logger.info("canvas_list")
        canvases = await run_with_api_error_translation(
            self.name,
            "List canvases",
            lambda: self.client.canvases.list(),
        )
        canvas_dicts = dump_model(canvases)
        return {
            "success": True,
            "canvases": canvas_dicts,
            "count": len(canvases),
            "status": "success",
        }


class CanvasCreateTool(BaseMCPTool):
    """Create a new canvas."""

    def __init__(self, client: Client) -> None:
        super().__init__(
            name="canvas_create",
            description="Create a new canvas with the specified name and optional description.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "name", self.name)
        optional_str(kwargs, "description", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        name = require_str(kwargs, "name", self.name)
        description = optional_str(kwargs, "description", self.name)
        self.logger.info("canvas_create", name=name)
        payload: dict[str, Any] = {"name": name}
        if description is not None:
            payload["description"] = description
        created = await run_with_api_error_translation(
            self.name,
            "Create canvas",
            lambda: self.client.canvases.create(payload),
        )
        return {
            "success": True,
            "canvas_name": name,
            "canvas_id": created.id,
            "canvas_description": description,
            "create_result": dump_model(created),
            "status": "success",
        }


class CanvasUpdateTool(BaseMCPTool):
    """Update canvas metadata (name, description)."""

    def __init__(self, client: Client) -> None:
        super().__init__(
            name="canvas_update",
            description="Update canvas metadata (name, description).",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        name = optional_str(kwargs, "name", self.name)
        description = optional_str(kwargs, "description", self.name)
        if name is None and description is None:
            from .base import MCPToolValidationError

            raise MCPToolValidationError(
                "At least one of 'name' or 'description' must be provided",
                self.name,
            )
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        name = optional_str(kwargs, "name", self.name)
        description = optional_str(kwargs, "description", self.name)
        payload: dict[str, Any] = {}
        if name is not None:
            payload["name"] = name
        if description is not None:
            payload["description"] = description
        self.logger.info("canvas_update", canvas_id=canvas_id, fields=list(payload))
        updated = await run_with_api_error_translation(
            self.name,
            "Update canvas",
            lambda: self.client.canvases.update(canvas_id, payload),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            "updated_fields": payload,
            "update_result": dump_model(updated),
            "status": "success",
        }


class CanvasCopyTool(BaseMCPTool):
    """Copy an existing canvas."""

    def __init__(self, client: Client) -> None:
        super().__init__(
            name="canvas_copy",
            description="Create a copy of an existing canvas.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        optional_str(kwargs, "new_name", self.name)
        optional_str(kwargs, "folder_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        new_name = optional_str(kwargs, "new_name", self.name)
        folder_id = optional_str(kwargs, "folder_id", self.name)
        payload: dict[str, Any] = {}
        if new_name is not None:
            payload["name"] = new_name
        if folder_id is not None:
            payload["folder_id"] = folder_id
        self.logger.info("canvas_copy", canvas_id=canvas_id, new_name=new_name)
        copied = await run_with_api_error_translation(
            self.name,
            "Copy canvas",
            lambda: self.client.canvases.copy(canvas_id, payload),
        )
        return {
            "success": True,
            "original_canvas_id": canvas_id,
            "new_canvas_id": copied.id,
            "new_canvas_name": copied.name,
            "copy_result": dump_model(copied),
            "status": "success",
        }


class CanvasMoveTool(BaseMCPTool):
    """Move a canvas to a different folder."""

    def __init__(self, client: Client) -> None:
        super().__init__(
            name="canvas_move",
            description="Move a canvas to a different folder.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        # 'destination' is the legacy name; SDK uses folder_id.
        if not (kwargs.get("destination") or kwargs.get("folder_id")):
            from .base import MCPToolValidationError

            raise MCPToolValidationError(
                "destination (folder_id) parameter is required", self.name
            )
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        destination = optional_str(kwargs, "folder_id", self.name) or optional_str(
            kwargs, "destination", self.name
        )
        assert destination is not None  # validated above
        self.logger.info("canvas_move", canvas_id=canvas_id, dest=destination)
        moved = await run_with_api_error_translation(
            self.name,
            "Move canvas",
            lambda: self.client.canvases.move(canvas_id, destination),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            "destination": destination,
            "move_result": dump_model(moved),
            "status": "success",
        }


__all__ = [
    "CanvasCopyTool",
    "CanvasCreateTool",
    "CanvasGetTool",
    "CanvasListTool",
    "CanvasMoveTool",
    "CanvasUpdateTool",
]
