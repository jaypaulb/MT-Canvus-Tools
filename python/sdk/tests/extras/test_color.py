"""Phase 4b §4.2 #22: color utilities tests."""

from __future__ import annotations

import pytest

from canvus_sdk.errors import ValidationError
from canvus_sdk.extras.color import (
    COLOR_BLACK,
    COLOR_WHITE,
    color_to_rgb,
    color_to_rgba,
    color_with_alpha,
    normalize_color,
    rgba_to_color,
    validate_color,
)


def test_validate_color_accepts_canonical() -> None:
    validate_color("FF00CCFF")


@pytest.mark.parametrize(
    "value", ["", "FF", "FF00CC", "GG00CCFF", "ff00ccff", "FF00CCFFFF"]
)
def test_validate_color_rejects(value: str) -> None:
    with pytest.raises(ValidationError):
        validate_color(value)


def test_normalize_color_handles_rgb_and_hash() -> None:
    assert normalize_color("#ff00cc") == "FF00CCFF"
    assert normalize_color("ff00cc") == "FF00CCFF"
    assert normalize_color("FF00CCAB") == "FF00CCAB"
    assert normalize_color("#FF00CCAB") == "FF00CCAB"


def test_normalize_color_rejects_garbage() -> None:
    with pytest.raises(ValidationError):
        normalize_color("xyz")


def test_color_to_rgba_and_back() -> None:
    r, g, b, a = color_to_rgba("FF00CC80")
    assert (r, g, b, a) == (255, 0, 204, 128)
    assert rgba_to_color(r, g, b, a) == "FF00CC80"


def test_color_to_rgb_strips_alpha() -> None:
    assert color_to_rgb("FF00CC80") == "#FF00CC"


def test_color_with_alpha() -> None:
    assert color_with_alpha(COLOR_WHITE, 0) == "FFFFFF00"
    assert color_with_alpha(COLOR_BLACK, 255) == COLOR_BLACK


def test_rgba_to_color_bounds_check() -> None:
    with pytest.raises(ValidationError):
        rgba_to_color(300, 0, 0, 0)
