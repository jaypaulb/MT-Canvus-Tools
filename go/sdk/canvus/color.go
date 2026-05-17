package canvus

import (
	"fmt"
	"regexp"
	"strings"
)

// Canvus API color helpers. All colors are 8-character uppercase hex strings
// in RRGGBBAA format (alpha last; 00 transparent, FF opaque).

var colorRGBAPattern = regexp.MustCompile(`^[0-9A-F]{8}$`)

// ValidateColor returns nil if the input matches the RRGGBBAA uppercase format.
func ValidateColor(color string) error {
	if len(color) != 8 {
		return fmt.Errorf("color must be exactly 8 characters (RRGGBBAA), got %d", len(color))
	}
	if !colorRGBAPattern.MatchString(color) {
		return fmt.Errorf("color must be uppercase hex RRGGBBAA format, got %q", color)
	}
	return nil
}

// NormalizeColor converts an input color string to the Canvus uppercase
// RRGGBBAA format. Accepts: RRGGBBAA (already valid), RRGGBB (adds FF alpha),
// and #-prefixed variants.
func NormalizeColor(color string) (string, error) {
	color = strings.TrimPrefix(color, "#")
	color = strings.ToUpper(color)
	if len(color) == 6 {
		if matched, _ := regexp.MatchString(`^[0-9A-F]{6}$`, color); matched {
			return color + "FF", nil
		}
		return "", fmt.Errorf("invalid 6-character color format: %q", color)
	}
	if err := ValidateColor(color); err != nil {
		return "", err
	}
	return color, nil
}

// ColorToRGBA converts a Canvus color (RRGGBBAA) to separate R, G, B, A bytes.
func ColorToRGBA(color string) (r, g, b, a byte, err error) {
	if err := ValidateColor(color); err != nil {
		return 0, 0, 0, 0, err
	}
	var rr, gg, bb, aa uint
	if _, err := fmt.Sscanf(color, "%02X%02X%02X%02X", &rr, &gg, &bb, &aa); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to parse color %q: %w", color, err)
	}
	return byte(rr), byte(gg), byte(bb), byte(aa), nil
}

// RGBAToColor converts separate RGBA bytes to a Canvus color (RRGGBBAA).
func RGBAToColor(r, g, b, a byte) string {
	return fmt.Sprintf("%02X%02X%02X%02X", r, g, b, a)
}

// ColorToRGB converts a Canvus color (RRGGBBAA) to standard #RRGGBB (no alpha).
func ColorToRGB(color string) (string, error) {
	if err := ValidateColor(color); err != nil {
		return "", err
	}
	return "#" + color[:6], nil
}

// ColorWithAlpha returns a new color string with the alpha channel replaced.
func ColorWithAlpha(color string, alpha byte) (string, error) {
	if err := ValidateColor(color); err != nil {
		return "", err
	}
	return color[:6] + fmt.Sprintf("%02X", alpha), nil
}

// Common opaque colors for convenience.
const (
	ColorBlack       = "000000FF"
	ColorWhite       = "FFFFFFFF"
	ColorRed         = "FF0000FF"
	ColorGreen       = "00FF00FF"
	ColorBlue        = "0000FFFF"
	ColorYellow      = "FFFF00FF"
	ColorCyan        = "00FFFFFF"
	ColorMagenta     = "FF00FFFF"
	ColorGray        = "808080FF"
	ColorLightGray   = "D3D3D3FF"
	ColorDarkGray    = "404040FF"
	ColorTransparent = "00000000"
)
