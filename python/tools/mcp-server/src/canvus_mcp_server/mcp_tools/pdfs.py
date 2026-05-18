"""PDF widget MCP tools (see :mod:`_asset_tools` for the shared shape).

Source had no ``PdfDeleteTool`` in the public registry; we mirror that by
omitting it here. If/when the audit approves adding it the wiring is one
line.
"""

from __future__ import annotations

from canvus_sdk import Client

from ._asset_tools import (
    _AssetCreateTool,
    _AssetGetTool,
    _AssetListTool,
    _AssetUpdateTool,
)


class PdfCreateTool(_AssetCreateTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            client,
            name="pdf_create",
            description="Create a new PDF element on a canvas with file upload.",
            kind="pdfs",
            id_key="pdf_id",
        )


class PdfGetTool(_AssetGetTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            client,
            name="pdf_get",
            description="Retrieve detailed information about a specific PDF by ID.",
            kind="pdfs",
            id_param="pdf_id",
        )


class PdfListTool(_AssetListTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            client,
            name="pdf_list",
            description="Retrieve a list of all PDFs on a specific canvas.",
            kind="pdfs",
            list_key="pdfs",
        )


class PdfUpdateTool(_AssetUpdateTool):
    def __init__(self, client: Client) -> None:
        super().__init__(
            client,
            name="pdf_update",
            description="Update PDF positioning or metadata.",
            kind="pdfs",
            id_param="pdf_id",
        )


__all__ = [
    "PdfCreateTool",
    "PdfGetTool",
    "PdfListTool",
    "PdfUpdateTool",
]
