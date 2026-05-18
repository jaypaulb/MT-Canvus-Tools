"""Phase 4b §4.2 #17: filters tests."""

from __future__ import annotations

from canvus_sdk.extras.filters import (
    Filter,
    FilterOperator,
    combine_filters,
    create_spatial_filter,
    create_text_filter,
    create_widget_type_filter,
    create_wildcard_filter,
)
from canvus_sdk.extras.geometry import Rectangle


def test_equals_and_dotted_path() -> None:
    f = Filter().add_condition("location.x", FilterOperator.EQUALS, 10)
    assert f.matches({"location": {"x": 10, "y": 0}})
    assert not f.matches({"location": {"x": 11}})


def test_contains_starts_ends_with() -> None:
    f = Filter().add_condition("title", FilterOperator.CONTAINS, "hello")
    assert f.matches({"title": "say hello world"})
    assert not f.matches({"title": "hi"})

    assert Filter().add_condition("title", FilterOperator.STARTS_WITH, "say").matches(
        {"title": "say hello"}
    )
    assert Filter().add_condition("title", FilterOperator.ENDS_WITH, "world").matches(
        {"title": "hello world"}
    )


def test_widget_type_filter_in() -> None:
    f = create_widget_type_filter(["note", "image"])
    assert f.matches({"widget_type": "note"})
    assert not f.matches({"widget_type": "pdf"})


def test_wildcard_filter() -> None:
    f = create_wildcard_filter("hel*", "title")
    assert f.matches({"title": "hello world"})
    assert not f.matches({"title": "goodbye"})


def test_spatial_filter_intersects() -> None:
    area = Rectangle(0, 0, 100, 100)
    f = create_spatial_filter(area, "intersects")
    assert f.matches({"location": {"x": 10, "y": 10}, "size": {"width": 20, "height": 20}})
    assert not f.matches(
        {"location": {"x": 200, "y": 200}, "size": {"width": 10, "height": 10}}
    )


def test_text_filter_and_combination() -> None:
    f = create_text_filter("hi", fields=["title", "text"])
    # AND semantics: both must contain
    assert f.matches({"title": "hi mom", "text": "hi dad"})
    assert not f.matches({"title": "hi mom", "text": "hello"})


def test_combine_filters_concatenates_conditions() -> None:
    f1 = Filter().add_condition("widget_type", FilterOperator.EQUALS, "note")
    f2 = Filter().add_condition("title", FilterOperator.CONTAINS, "hello")
    merged = combine_filters(f1, f2)
    assert len(merged.conditions) == 2


def test_to_from_dict_roundtrip() -> None:
    original = Filter().add_condition("title", FilterOperator.EQUALS, "foo")
    data = original.to_dict()
    rebuilt = Filter.from_dict(data)
    assert rebuilt.matches({"title": "foo"})
    assert not rebuilt.matches({"title": "bar"})


def test_filter_iterable() -> None:
    f = Filter().add_condition("widget_type", FilterOperator.EQUALS, "note")
    items = [
        {"widget_type": "note", "id": 1},
        {"widget_type": "image", "id": 2},
        {"widget_type": "note", "id": 3},
    ]
    assert [item["id"] for item in f.filter(items)] == [1, 3]
