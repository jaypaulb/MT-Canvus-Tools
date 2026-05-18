"""Note widget MCP tools.

Wraps ``client.widgets.notes.*`` from the new Python SDK. Decomposed out of
the legacy 3,169-LOC ``mcp_tools/canvas.py`` per the Phase 4c audit.
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


class NoteCreateTool(BaseMCPTool):
    """Create a note widget on a canvas."""

    def __init__(self, client: Client) -> None:
        super().__init__(
            name="note_create",
            description="Create a new note element on a canvas with text content and styling.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        payload = optional_dict(kwargs, "payload", self.name)
        if payload is None:
            raise MCPToolValidationError("payload parameter is required", self.name)
        if "text" not in payload:
            raise MCPToolValidationError(
                "payload must contain 'text' field", self.name
            )
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        payload = optional_dict(kwargs, "payload", self.name) or {}
        self.logger.info("note_create", canvas_id=canvas_id)
        created = await run_with_api_error_translation(
            self.name,
            "Create note",
            lambda: self.client.widgets.notes.create(canvas_id, payload),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            "note_id": created.id,
            "payload": payload,
            "note_data": dump_model(created),
            "status": "success",
        }


class NoteGetTool(BaseMCPTool):
    """Retrieve a single note."""

    def __init__(self, client: Client) -> None:
        super().__init__(
            name="note_get",
            description="Retrieve detailed information about a specific note by ID.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        require_str(kwargs, "note_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        note_id = require_str(kwargs, "note_id", self.name)
        self.logger.info("note_get", canvas_id=canvas_id, note_id=note_id)
        note = await run_with_api_error_translation(
            self.name,
            "Retrieve note",
            lambda: self.client.widgets.notes.get(canvas_id, note_id),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            "note_id": note_id,
            "note_data": dump_model(note),
            "status": "success",
        }


class NoteListTool(BaseMCPTool):
    """List all notes on a canvas."""

    def __init__(self, client: Client) -> None:
        super().__init__(
            name="note_list",
            description="Retrieve a list of all notes on a specific canvas.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        self.logger.info("note_list", canvas_id=canvas_id)
        notes = await run_with_api_error_translation(
            self.name,
            "List notes",
            lambda: self.client.widgets.notes.list(canvas_id),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            "notes": dump_model(notes),
            "count": len(notes),
            "status": "success",
        }


class NoteUpdateTool(BaseMCPTool):
    """Update a note widget."""

    def __init__(self, client: Client) -> None:
        super().__init__(
            name="note_update",
            description="Update note content, color, or positioning.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        require_str(kwargs, "note_id", self.name)
        payload = optional_dict(kwargs, "payload", self.name)
        if not payload:
            raise MCPToolValidationError(
                "payload parameter is required and must be non-empty", self.name
            )
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        note_id = require_str(kwargs, "note_id", self.name)
        payload = optional_dict(kwargs, "payload", self.name) or {}
        self.logger.info("note_update", canvas_id=canvas_id, note_id=note_id)
        updated = await run_with_api_error_translation(
            self.name,
            "Update note",
            lambda: self.client.widgets.notes.update(canvas_id, note_id, payload),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            "note_id": note_id,
            "payload": payload,
            "note_data": dump_model(updated),
            "status": "success",
        }


class NoteDeleteTool(BaseMCPTool):
    """Delete a note widget."""

    def __init__(self, client: Client) -> None:
        super().__init__(
            name="note_delete",
            description="Delete a specific note from a canvas.",
        )
        self.client = client

    def validate_input(self, **kwargs: Any) -> bool:
        require_str(kwargs, "canvas_id", self.name)
        require_str(kwargs, "note_id", self.name)
        return True

    async def execute(self, **kwargs: Any) -> dict[str, Any]:
        canvas_id = require_str(kwargs, "canvas_id", self.name)
        note_id = require_str(kwargs, "note_id", self.name)
        self.logger.info("note_delete", canvas_id=canvas_id, note_id=note_id)
        await run_with_api_error_translation(
            self.name,
            "Delete note",
            lambda: self.client.widgets.notes.delete(canvas_id, note_id),
        )
        return {
            "success": True,
            "canvas_id": canvas_id,
            "note_id": note_id,
            "status": "deleted",
        }


__all__ = [
    "NoteCreateTool",
    "NoteDeleteTool",
    "NoteGetTool",
    "NoteListTool",
    "NoteUpdateTool",
]
