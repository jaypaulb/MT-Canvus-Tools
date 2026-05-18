"""Phase 4b §4.2 #22: color utilities port.

Canvus colors are 8-character uppercase hex strings in RRGGBBAA format
(alpha last; ``00`` transparent, ``FF`` opaque). Ported from Go's ``color.go``.
"""

from __future__ import annotations

import re

from ..errors import ValidationError

__all__ = [
    "COLOR_BLACK",
    "COLOR_BLUE",
    "COLOR_CYAN",
    "COLOR_DARK_GRAY",
    "COLOR_GRAY",
    "COLOR_GREEN",
    "COLOR_LIGHT_GRAY",
    "COLOR_MAGENTA",
    "COLOR_RED",
    "COLOR_TRANSPARENT",
    "COLOR_WHITE",
    "COLOR_YELLOW",
    "color_to_rgb",
    "color_to_rgba",
    "color_with_alpha",
    "normalize_color",
    "rgba_to_color",
    "validate_color",
]

_RGBA_PATTERN = re.compile(r"^[0-9A-F]{8}$")
_RGB_PATTERN = re.compile(r"^[0-9A-F]{6}$")


def validate_color(color: str) -> None:
    """Raise :class:`ValidationError` if ``color`` is not RRGGBBAA uppercase hex."""
    if len(color) != 8:
        raise ValidationError(
            f"color must be exactly 8 characters (RRGGBBAA), got {len(color)}"
        )
    if not _RGBA_PATTERN.match(color):
        raise ValidationError(
            f"color must be uppercase hex RRGGBBAA format, got {color!r}"
        )


def normalize_color(color: str) -> str:
    """Normalise an input color string to canonical Canvus RRGGBBAA form.

    Accepts: RRGGBBAA, RRGGBB (adds FF alpha), or ``#``-prefixed variants of
    either. Case-insensitive on input; output is always uppercase.
    """
    candidate = color.removeprefix("#").upper()
    if len(candidate) == 6:
        if _RGB_PATTERN.match(candidate):
            return candidate + "FF"
        raise ValidationError(f"invalid 6-character color format: {color!r}")
    validate_color(candidate)
    return candidate


def color_to_rgba(color: str) -> tuple[int, int, int, int]:
    """Convert a RRGGBBAA string to four 0-255 ints (R, G, B, A)."""
    validate_color(color)
    return (
        int(color[0:2], 16),
        int(color[2:4], 16),
        int(color[4:6], 16),
        int(color[6:8], 16),
    )


def rgba_to_color(r: int, g: int, b: int, a: int) -> str:
    """Convert four 0-255 ints to a Canvus RRGGBBAA color string."""
    for name, value in (("r", r), ("g", g), ("b", b), ("a", a)):
        if not 0 <= value <= 255:
            raise ValidationError(
                f"{name} channel must be in 0..255, got {value}"
            )
    return f"{r:02X}{g:02X}{b:02X}{a:02X}"


def color_to_rgb(color: str) -> str:
    """Return ``#RRGGBB`` (no alpha) from a Canvus color."""
    validate_color(color)
    return "#" + color[:6]


def color_with_alpha(color: str, alpha: int) -> str:
    """Return a new color string with ``alpha`` (0-255) replacing the alpha channel."""
    validate_color(color)
    if not 0 <= alpha <= 255:
        raise ValidationError(f"alpha must be in 0..255, got {alpha}")
    return color[:6] + f"{alpha:02X}"


# Common opaque colors — matches Go's exported constants.
COLOR_BLACK = "000000FF"
COLOR_WHITE = "FFFFFFFF"
COLOR_RED = "FF0000FF"
COLOR_GREEN = "00FF00FF"
COLOR_BLUE = "0000FFFF"
COLOR_YELLOW = "FFFF00FF"
COLOR_CYAN = "00FFFFFF"
COLOR_MAGENTA = "FF00FFFF"
COLOR_GRAY = "808080FF"
COLOR_LIGHT_GRAY = "D3D3D3FF"
COLOR_DARK_GRAY = "404040FF"
COLOR_TRANSPARENT = "00000000"
