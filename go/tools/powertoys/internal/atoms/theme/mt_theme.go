package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// MTTheme is a custom theme with MT dark blue and pink colors for entries.
type MTTheme struct {
	fyne.Theme
}

// NewMTTheme creates a new MT theme based on the default theme.
func NewMTTheme() fyne.Theme {
	return &MTTheme{
		Theme: theme.DefaultTheme(),
	}
}

// Color returns custom colors for specific color names.
func (m *MTTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameInputBackground:
		// MT Dark Blue: #1a1a2e for Entry background
		return &color.NRGBA{R: 0x1a, G: 0x1a, B: 0x2e, A: 0xff}
	case theme.ColorNameInputBorder:
		// Slightly lighter blue for Entry border
		return &color.NRGBA{R: 0x2a, G: 0x2a, B: 0x4e, A: 0xff}
	case theme.ColorNameForeground:
		// MT Pink/Magenta: #E6007E for text
		// Note: This affects all text including Entry widgets
		// Entry widgets use Foreground color for their text
		return &color.NRGBA{R: 0xE6, G: 0x00, B: 0x7E, A: 0xff}
	case theme.ColorNamePlaceHolder:
		// Lighter pink for placeholder text
		return &color.NRGBA{R: 0xE6, G: 0x00, B: 0x7E, A: 0x80}
	default:
		return m.Theme.Color(name, variant)
	}
}

// Font returns the default font.
func (m *MTTheme) Font(style fyne.TextStyle) fyne.Resource {
	return m.Theme.Font(style)
}

// Icon returns the default icon.
func (m *MTTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return m.Theme.Icon(name)
}

// Size returns custom sizes for specific size names.
func (m *MTTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameScrollBar, theme.SizeNameScrollBarSmall:
		// Make scrollbar larger (default is typically 12px, increase to 20px for easier grabbing)
		return 20
	default:
		return m.Theme.Size(name)
	}
}

