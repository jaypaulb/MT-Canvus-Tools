"""Shared base class for all Canvus SDK Pydantic models."""

from __future__ import annotations

from pydantic import BaseModel, ConfigDict


class CanvusModel(BaseModel):
    """Common configuration for Canvus wire-format models.

    - Allows population by either field name or alias (so callers can
      use Python-native names regardless of whether the wire format uses
      hyphens or underscores).
    - Ignores extra fields rather than rejecting them — the Canvus API
      occasionally introduces new fields, and the SDK should not break on
      forward-compatible additions.
    """

    model_config = ConfigDict(
        populate_by_name=True,
        extra="allow",
        validate_assignment=True,
    )


__all__ = ["CanvusModel"]
