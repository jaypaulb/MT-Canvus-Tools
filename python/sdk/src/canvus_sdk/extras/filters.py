"""Phase 4b §4.2 #17: filters port.

Advanced client-side filtering for canvases and widgets. Ported from
``CanvusPythonAPI/canvus_api/filters.py``; cleaned up for mypy --strict.

The new module's :class:`Filter` is a small predicate composer that builds a
list of conditions and evaluates them all-true against a target dict. It is
explicitly **client-side**: the server's wire-level filtering vocabulary is
limited; use :class:`Filter` after a list call when finer matching is needed.

Supports:

- Comparison operators (equals, contains, starts_with, gt/lt, in, exists).
- Spatial operators against an axis-aligned :class:`Rectangle`.
- Wildcard matching using ``*`` and ``?``.
"""

from __future__ import annotations

import re
from collections.abc import Iterable, Sequence
from dataclasses import dataclass, field
from enum import StrEnum
from typing import Any

from .geometry import Rectangle, contains, intersects

__all__ = [
    "Condition",
    "Filter",
    "FilterOperator",
    "combine_filters",
    "create_filter",
    "create_spatial_filter",
    "create_text_filter",
    "create_widget_type_filter",
    "create_wildcard_filter",
]


class FilterOperator(StrEnum):
    """Supported filter operators."""

    EQUALS = "equals"
    NOT_EQUALS = "not_equals"
    CONTAINS = "contains"
    NOT_CONTAINS = "not_contains"
    STARTS_WITH = "starts_with"
    ENDS_WITH = "ends_with"
    GREATER_THAN = "greater_than"
    LESS_THAN = "less_than"
    GREATER_EQUAL = "greater_equal"
    LESS_EQUAL = "less_equal"
    IN = "in"
    NOT_IN = "not_in"
    EXISTS = "exists"
    NOT_EXISTS = "not_exists"
    SPATIAL_INTERSECTS = "spatial_intersects"
    SPATIAL_CONTAINS = "spatial_contains"
    SPATIAL_WITHIN = "spatial_within"
    WILDCARD_MATCH = "wildcard_match"


@dataclass(frozen=True, slots=True)
class Condition:
    """One predicate in a :class:`Filter`. ``field`` supports dot-notation."""

    field: str
    operator: FilterOperator
    value: Any


@dataclass(slots=True)
class Filter:
    """Composable client-side filter.

    Conditions are evaluated in the order they were added and combined with
    logical AND. ``OR`` semantics can be approximated by running two filters
    and unioning the results, or by use of
    :func:`combine_filters` (which concatenates conditions = AND).
    """

    conditions: list[Condition] = field(default_factory=list)

    def add_condition(
        self,
        field_name: str,
        operator: FilterOperator | str,
        value: Any,
    ) -> Filter:
        """Append a comparison condition and return ``self``."""
        op = FilterOperator(operator) if isinstance(operator, str) else operator
        self.conditions.append(Condition(field=field_name, operator=op, value=value))
        return self

    def add_spatial_condition(
        self,
        operator: str,
        area: Rectangle,
    ) -> Filter:
        """Append a spatial condition. ``operator`` is one of intersects/contains/within."""
        op = FilterOperator(f"spatial_{operator}")
        self.conditions.append(Condition(field="spatial", operator=op, value=area))
        return self

    def add_wildcard_condition(self, field_name: str, pattern: str) -> Filter:
        """Append a wildcard condition (``*`` matches any sequence; ``?`` matches one)."""
        self.conditions.append(
            Condition(field=field_name, operator=FilterOperator.WILDCARD_MATCH, value=pattern)
        )
        return self

    def matches(self, item: dict[str, Any]) -> bool:
        """Return True iff ``item`` satisfies every condition."""
        return all(self._matches_condition(item, c) for c in self.conditions)

    def filter(self, items: Iterable[dict[str, Any]]) -> list[dict[str, Any]]:
        """Filter an iterable of items; returns those matching all conditions."""
        return [item for item in items if self.matches(item)]

    def to_dict(self) -> dict[str, Any]:
        """Serialise to a plain dict (round-trip with :meth:`from_dict`)."""
        return {
            "conditions": [
                {"field": c.field, "operator": c.operator.value, "value": c.value}
                for c in self.conditions
            ],
        }

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> Filter:
        """Rebuild a Filter from the dict shape produced by :meth:`to_dict`."""
        out = cls()
        for raw in data.get("conditions", []):
            out.conditions.append(
                Condition(
                    field=str(raw["field"]),
                    operator=FilterOperator(raw["operator"]),
                    value=raw["value"],
                )
            )
        return out

    # ---- internals ---------------------------------------------------------

    def _matches_condition(self, item: dict[str, Any], cond: Condition) -> bool:
        op = cond.operator
        if cond.field == "spatial" and op.value.startswith("spatial_"):
            return self._spatial(item, op, cond.value)
        field_value = _get_nested(item, cond.field)
        return _evaluate(field_value, op, cond.value)

    @staticmethod
    def _spatial(item: dict[str, Any], op: FilterOperator, area: Rectangle) -> bool:
        location = item.get("location")
        size = item.get("size")
        if not isinstance(location, dict) or not isinstance(size, dict):
            return False
        widget_rect = Rectangle(
            x=float(location.get("x", 0.0)),
            y=float(location.get("y", 0.0)),
            width=float(size.get("width", 0.0)),
            height=float(size.get("height", 0.0)),
        )
        if op is FilterOperator.SPATIAL_INTERSECTS:
            return intersects(widget_rect, area)
        if op is FilterOperator.SPATIAL_CONTAINS:
            # Widget contains the query area
            return contains(widget_rect, area)
        if op is FilterOperator.SPATIAL_WITHIN:
            # Widget within the query area
            return contains(area, widget_rect)
        return False


# ---- evaluation helpers ----------------------------------------------------


def _get_nested(item: dict[str, Any], field_name: str) -> Any:
    if "." not in field_name:
        return item.get(field_name)
    current: Any = item
    for part in field_name.split("."):
        if isinstance(current, dict):
            current = current.get(part)
        else:
            return None
        if current is None:
            return None
    return current


def _evaluate(field_value: Any, op: FilterOperator, target: Any) -> bool:
    if op is FilterOperator.EQUALS:
        return bool(field_value == target)
    if op is FilterOperator.NOT_EQUALS:
        return bool(field_value != target)
    if op is FilterOperator.CONTAINS:
        return _safe_in(target, field_value)
    if op is FilterOperator.NOT_CONTAINS:
        return not _safe_in(target, field_value)
    if op is FilterOperator.STARTS_WITH:
        return field_value is not None and str(field_value).startswith(str(target))
    if op is FilterOperator.ENDS_WITH:
        return field_value is not None and str(field_value).endswith(str(target))
    if op is FilterOperator.GREATER_THAN:
        return field_value is not None and field_value > target
    if op is FilterOperator.LESS_THAN:
        return field_value is not None and field_value < target
    if op is FilterOperator.GREATER_EQUAL:
        return field_value is not None and field_value >= target
    if op is FilterOperator.LESS_EQUAL:
        return field_value is not None and field_value <= target
    if op is FilterOperator.IN:
        return field_value is not None and _safe_in(field_value, target)
    if op is FilterOperator.NOT_IN:
        return field_value is None or not _safe_in(field_value, target)
    if op is FilterOperator.EXISTS:
        return field_value is not None
    if op is FilterOperator.NOT_EXISTS:
        return field_value is None
    if op is FilterOperator.WILDCARD_MATCH:
        return _wildcard_match(field_value, str(target))
    return False


def _safe_in(needle: Any, haystack: Any) -> bool:
    if haystack is None:
        return False
    try:
        return needle in haystack
    except TypeError:
        return False


def _wildcard_match(field_value: Any, pattern: str) -> bool:
    if field_value is None:
        return False
    regex = re.escape(pattern).replace(r"\*", ".*").replace(r"\?", ".")
    try:
        return bool(re.match(f"^{regex}$", str(field_value), re.IGNORECASE))
    except re.error:
        return False


# ---- factories -------------------------------------------------------------


def create_filter() -> Filter:
    """Return a new empty :class:`Filter`."""
    return Filter()


def create_spatial_filter(area: Rectangle, operator: str = "intersects") -> Filter:
    """Build a single-condition spatial filter."""
    return Filter().add_spatial_condition(operator, area)


def create_widget_type_filter(widget_types: str | Sequence[str]) -> Filter:
    """Build a filter matching widgets whose ``widget_type`` is in ``widget_types``."""
    types: Sequence[str] = (widget_types,) if isinstance(widget_types, str) else widget_types
    return Filter().add_condition("widget_type", FilterOperator.IN, list(types))


def create_text_filter(text: str, fields: Sequence[str] | None = None) -> Filter:
    """Build a text-search filter (all fields must contain ``text`` — AND).

    Note:
        The legacy module uses AND semantics across all listed fields. Callers
        wanting OR-style search across multiple text fields should run a
        :class:`Filter` per field and union the results.
    """
    targets: Sequence[str] = ("title", "text", "description") if fields is None else fields
    f = Filter()
    for field_name in targets:
        f.add_condition(field_name, FilterOperator.CONTAINS, text)
    return f


def create_wildcard_filter(pattern: str, field_name: str = "title") -> Filter:
    """Build a single-condition wildcard filter."""
    return Filter().add_wildcard_condition(field_name, pattern)


def combine_filters(*filters: Filter, operator: str = "AND") -> Filter:
    """Concatenate filters' conditions into a single AND-combined filter.

    ``operator="OR"`` is accepted for parity with the legacy signature but
    currently behaves identically to ``"AND"`` — true OR semantics require
    running each filter independently and unioning the result sets.
    """
    # Reference the operator so it appears in the API surface even though
    # only AND is implemented; future expansion may add a parallel path.
    _ = operator
    if not filters:
        return Filter()
    combined = Filter()
    for f in filters:
        combined.conditions.extend(f.conditions)
    return combined
