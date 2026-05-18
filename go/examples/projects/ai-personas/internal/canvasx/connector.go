package canvasx

// PersonaColors returns the standard 4-color palette used to distinguish
// personas on the canvas (blue, green, orange, purple).
func PersonaColors() []string {
	return []string{"#2196f3ff", "#4caf50ff", "#ff9800ff", "#9c27b0ff"}
}

// PersonaColorSet returns the same palette as a lookup set for case-insensitive
// membership tests on a note's background_color value.
func PersonaColorSet() map[string]bool {
	return map[string]bool{
		"#2196f3ff": true,
		"#4caf50ff": true,
		"#ff9800ff": true,
		"#9c27b0ff": true,
	}
}

// Workflow colors.
const (
	ColorAmberProcessing    = "#ffe4b3"
	ColorGreenComplete      = "#ccffcc"
	ColorWhiteAIQuestion    = "#ffffffff"
	ColorGreyHelper         = "#e0e0e0"
	ColorRedFailedPersona   = "#f44336ff"
	ColorAmberTimeoutHelper = "#ff9800ff"
)

// BuildConnectorPayload returns the Canvus connector create payload for an
// arrow from srcID to dstID using the standard line styling.
func BuildConnectorPayload(srcID, dstID string) map[string]any {
	return map[string]any{
		"src": map[string]any{
			"id":            srcID,
			"auto_location": true,
			"tip":           "none",
		},
		"dst": map[string]any{
			"id":            dstID,
			"auto_location": true,
			"tip":           "solid-equilateral-triangle",
		},
		"line_color":  "#e7e7f2ff",
		"line_width":  5,
		"state":       "normal",
		"type":        "curve",
		"widget_type": "Connector",
	}
}
