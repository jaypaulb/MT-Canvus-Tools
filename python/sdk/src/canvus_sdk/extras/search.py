"""Phase 4b §4.2 #19: search port.

Cross-canvas widget search built on top of the new SDK's typed resources.

Ported from ``CanvusPythonAPI/canvus_api/search.py`` with these changes:

- Talks to :class:`canvus_sdk.Client` instead of the legacy ``CanvusClient``.
- Uses pydantic ``model_dump()`` for matching (the legacy version called
  ``dict()`` and ``model_dump`` interchangeably; here it is always model_dump).
- Returns immutable :class:`SearchResult` frozen dataclasses.
- ``_parse_query`` only treats strings as JSON when they look like JSON;
  bare strings become a text wildcard (legacy did the same but raised on
  non-string non-dict input).
- Reraises errors from per-canvas widget listing as :class:`SearchError`
  with the offending canvas id rather than silently logging to stdout.
"""

from __future__ import annotations

import json
import re
from collections.abc import Mapping, Sequence
from dataclasses import dataclass
from typing import TYPE_CHECKING, Any

from ..errors import CanvusError
from .geometry import Rectangle, intersects, widget_bounding_box

if TYPE_CHECKING:
    from ..client import Client
    from ..models import Canvas

__all__ = [
    "CrossCanvasSearch",
    "SearchError",
    "SearchResult",
    "find_widgets_across_canvases",
    "find_widgets_by_property",
    "find_widgets_by_text",
    "find_widgets_by_type",
    "find_widgets_in_area",
]


class SearchError(CanvusError):
    """Raised when search fails on a specific canvas the caller asked about."""


@dataclass(frozen=True, slots=True)
class SearchResult:
    """One hit from a cross-canvas widget search."""

    canvas_id: str
    canvas_name: str
    widget_id: str
    widget_type: str
    widget: Any
    match_score: float = 1.0
    match_reason: str = ""

    @property
    def drill_down_path(self) -> str:
        """``canvas_id:widget_id`` shorthand suitable for hyperlinks."""
        return f"{self.canvas_id}:{self.widget_id}"

    def to_dict(self) -> dict[str, Any]:
        """Round-trippable plain dict."""
        widget_dump = (
            self.widget.model_dump()
            if hasattr(self.widget, "model_dump")
            else dict(self.widget)
            if isinstance(self.widget, Mapping)
            else {}
        )
        return {
            "canvas_id": self.canvas_id,
            "canvas_name": self.canvas_name,
            "widget_id": self.widget_id,
            "widget_type": self.widget_type,
            "drill_down_path": self.drill_down_path,
            "match_score": self.match_score,
            "match_reason": self.match_reason,
            "widget_data": widget_dump,
        }


class CrossCanvasSearch:
    """Search widgets across many canvases via a :class:`canvus_sdk.Client`."""

    def __init__(self, client: Client) -> None:
        self.client = client

    async def find_widgets_across_canvases(
        self,
        query: str | dict[str, Any],
        *,
        canvas_ids: Sequence[str] | None = None,
        widget_types: Sequence[str] | None = None,
        spatial_filter: Rectangle | None = None,
        max_results: int = 100,
        include_deleted: bool = False,
    ) -> list[SearchResult]:
        """Return widgets matching ``query`` across the requested canvases.

        Args:
            query: Either a dict of ``{field: value-or-pattern}`` or a free-form
                string (wildcards via ``*``).
            canvas_ids: Restrict search to these canvas IDs (``None`` = all
                accessible canvases).
            widget_types: Only return widgets whose ``widget_type`` is in this
                list (case-insensitive).
            spatial_filter: Only return widgets intersecting this rectangle.
            max_results: Hard cap.
            include_deleted: If False (default), widgets with ``state ==
                'deleted'`` are skipped.
        """
        criteria = self._parse_query(query)
        canvases = await self._get_canvases_to_search(canvas_ids)
        results: list[SearchResult] = []
        for canvas in canvases:
            try:
                widgets = await self.client.widgets.list(canvas.id)
            except Exception as exc:
                raise SearchError(
                    f"search: failed to list widgets on canvas {canvas.id!r}: {exc}"
                ) from exc
            filtered = self._apply_filters(
                widgets, criteria, widget_types, spatial_filter, include_deleted
            )
            for widget in filtered:
                result = SearchResult(
                    canvas_id=canvas.id,
                    canvas_name=canvas.name,
                    widget_id=str(getattr(widget, "id", "") or ""),
                    widget_type=str(getattr(widget, "widget_type", "") or ""),
                    widget=widget,
                    match_score=self._calculate_match_score(widget, criteria),
                    match_reason=self._get_match_reason(widget, criteria),
                )
                results.append(result)
                if len(results) >= max_results:
                    break
            if len(results) >= max_results:
                break
        results.sort(key=lambda r: r.match_score, reverse=True)
        return results[:max_results]

    async def find_widgets_by_text(
        self,
        text: str,
        *,
        canvas_ids: Sequence[str] | None = None,
        case_sensitive: bool = False,
        max_results: int = 100,
    ) -> list[SearchResult]:
        """Search for widgets containing ``text`` (wildcard-wrapped by default)."""
        criteria: dict[str, Any] = (
            {"text": text} if case_sensitive else {"text": f"*{text}*"}
        )
        return await self.find_widgets_across_canvases(
            criteria, canvas_ids=canvas_ids, max_results=max_results
        )

    async def find_widgets_by_type(
        self,
        widget_type: str,
        *,
        canvas_ids: Sequence[str] | None = None,
        max_results: int = 100,
    ) -> list[SearchResult]:
        """Search for widgets whose ``widget_type`` matches (case-insensitive)."""
        return await self.find_widgets_across_canvases(
            {"widget_type": widget_type},
            canvas_ids=canvas_ids,
            max_results=max_results,
        )

    async def find_widgets_in_area(
        self,
        area: Rectangle,
        *,
        canvas_ids: Sequence[str] | None = None,
        widget_types: Sequence[str] | None = None,
        max_results: int = 100,
    ) -> list[SearchResult]:
        """Search for widgets whose bounding box intersects ``area``."""
        return await self.find_widgets_across_canvases(
            {},
            canvas_ids=canvas_ids,
            widget_types=widget_types,
            spatial_filter=area,
            max_results=max_results,
        )

    async def find_widgets_by_property(
        self,
        property_path: str,
        value: Any,
        *,
        canvas_ids: Sequence[str] | None = None,
        max_results: int = 100,
    ) -> list[SearchResult]:
        """Find widgets whose dotted-path property equals ``value``."""
        return await self.find_widgets_across_canvases(
            {property_path: value},
            canvas_ids=canvas_ids,
            max_results=max_results,
        )

    # ---- internals ---------------------------------------------------------

    @staticmethod
    def _parse_query(query: str | dict[str, Any]) -> dict[str, Any]:
        if isinstance(query, dict):
            return query
        stripped = query.strip()
        if stripped.startswith("{") and stripped.endswith("}"):
            try:
                parsed = json.loads(stripped)
            except json.JSONDecodeError:
                return {"text": f"*{query}*"}
            if isinstance(parsed, dict):
                return parsed
        return {"text": f"*{query}*"}

    async def _get_canvases_to_search(
        self,
        canvas_ids: Sequence[str] | None,
    ) -> list[Canvas]:
        if not canvas_ids:
            return await self.client.canvases.list()
        out: list[Canvas] = []
        for canvas_id in canvas_ids:
            try:
                out.append(await self.client.canvases.get(canvas_id))
            except Exception as exc:
                raise SearchError(
                    f"search: failed to fetch canvas {canvas_id!r}: {exc}"
                ) from exc
        return out

    def _apply_filters(
        self,
        widgets: Sequence[Any],
        criteria: dict[str, Any],
        widget_types: Sequence[str] | None,
        spatial_filter: Rectangle | None,
        include_deleted: bool,
    ) -> list[Any]:
        type_set = (
            {wt.lower() for wt in widget_types} if widget_types is not None else None
        )
        out: list[Any] = []
        for widget in widgets:
            if not include_deleted and str(getattr(widget, "state", "normal")) == "deleted":
                continue
            if (
                type_set is not None
                and str(getattr(widget, "widget_type", "")).lower() not in type_set
            ):
                continue
            if spatial_filter is not None:
                try:
                    if not intersects(widget_bounding_box(widget), spatial_filter):
                        continue
                except (ValueError, AttributeError):
                    continue
            if criteria and not self._matches_criteria(widget, criteria):
                continue
            out.append(widget)
        return out

    def _matches_criteria(self, widget: Any, criteria: dict[str, Any]) -> bool:
        widget_dict = (
            widget.model_dump()
            if hasattr(widget, "model_dump")
            else dict(widget)
            if isinstance(widget, Mapping)
            else {}
        )
        for key, target in criteria.items():
            current: Any = widget_dict
            for part in key.split("."):
                if isinstance(current, Mapping) and part in current:
                    current = current[part]
                else:
                    return False
            if isinstance(target, str) and "*" in target:
                if not self._wildcard_match(str(current), target):
                    return False
                continue
            if key == "widget_type":
                if str(current).lower() != str(target).lower():
                    return False
                continue
            if str(current) == str(target):
                continue
            try:
                if float(current) == float(target):
                    continue
            except (TypeError, ValueError):
                pass
            return False
        return True

    @staticmethod
    def _wildcard_match(text: str, pattern: str) -> bool:
        regex = re.escape(pattern).replace(r"\*", ".*")
        return re.search(regex, text, re.IGNORECASE) is not None

    def _calculate_match_score(
        self,
        widget: Any,
        criteria: dict[str, Any],
    ) -> float:
        if not criteria:
            return 1.0
        return 1.0 if self._matches_criteria(widget, criteria) else 0.0

    def _get_match_reason(self, widget: Any, criteria: dict[str, Any]) -> str:
        if not criteria:
            return "No filter applied"
        widget_dict = (
            widget.model_dump()
            if hasattr(widget, "model_dump")
            else dict(widget)
            if isinstance(widget, Mapping)
            else {}
        )
        for key, target in criteria.items():
            if key in widget_dict:
                if str(widget_dict[key]) == str(target):
                    return f"Exact match on {key}"
                if str(target).strip("*") in str(widget_dict[key]):
                    return f"Partial match on {key}"
        return "Filter criteria matched"


# ---- module-level shortcuts -----------------------------------------------


async def find_widgets_across_canvases(
    client: Client,
    query: str | dict[str, Any],
    *,
    canvas_ids: Sequence[str] | None = None,
    widget_types: Sequence[str] | None = None,
    spatial_filter: Rectangle | None = None,
    max_results: int = 100,
    include_deleted: bool = False,
) -> list[SearchResult]:
    """One-shot wrapper around :class:`CrossCanvasSearch`."""
    return await CrossCanvasSearch(client).find_widgets_across_canvases(
        query,
        canvas_ids=canvas_ids,
        widget_types=widget_types,
        spatial_filter=spatial_filter,
        max_results=max_results,
        include_deleted=include_deleted,
    )


async def find_widgets_by_text(
    client: Client,
    text: str,
    *,
    canvas_ids: Sequence[str] | None = None,
    case_sensitive: bool = False,
    max_results: int = 100,
) -> list[SearchResult]:
    """Convenience wrapper for text-based search."""
    return await CrossCanvasSearch(client).find_widgets_by_text(
        text,
        canvas_ids=canvas_ids,
        case_sensitive=case_sensitive,
        max_results=max_results,
    )


async def find_widgets_by_type(
    client: Client,
    widget_type: str,
    *,
    canvas_ids: Sequence[str] | None = None,
    max_results: int = 100,
) -> list[SearchResult]:
    """Convenience wrapper for type-based search."""
    return await CrossCanvasSearch(client).find_widgets_by_type(
        widget_type,
        canvas_ids=canvas_ids,
        max_results=max_results,
    )


async def find_widgets_in_area(
    client: Client,
    area: Rectangle,
    *,
    canvas_ids: Sequence[str] | None = None,
    widget_types: Sequence[str] | None = None,
    max_results: int = 100,
) -> list[SearchResult]:
    """Convenience wrapper for spatial search."""
    return await CrossCanvasSearch(client).find_widgets_in_area(
        area,
        canvas_ids=canvas_ids,
        widget_types=widget_types,
        max_results=max_results,
    )


async def find_widgets_by_property(
    client: Client,
    property_path: str,
    value: Any,
    *,
    canvas_ids: Sequence[str] | None = None,
    max_results: int = 100,
) -> list[SearchResult]:
    """Convenience wrapper for property-based search."""
    return await CrossCanvasSearch(client).find_widgets_by_property(
        property_path,
        value,
        canvas_ids=canvas_ids,
        max_results=max_results,
    )
