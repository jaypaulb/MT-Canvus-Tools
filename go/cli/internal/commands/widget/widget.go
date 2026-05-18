// Package widget provides commands for managing widgets in Canvus.
package widget

import (
	"fmt"
	"strconv"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
)

// WidgetCmd is the parent command for all widget-related operations.
var WidgetCmd = &cobra.Command{
	Use:   "widget",
	Short: "Manage widgets",
	Long: `Manage widgets in your Canvus canvases.

Widgets are the content elements within a canvas. They can be:
- Notes: Text content with formatting
- Images: Image files from URL or local file
- PDFs: PDF documents
- Videos: Video content
- Browser: Embedded web content
- Anchors: Connection points
- Connectors: Lines connecting widgets

Examples:
  # Create a note widget
  canvus widget create <canvas-id> --type note --text "My note"

  # List widgets in a canvas
  canvus widget list <canvas-id>

  # Get widget details
  canvus widget get <canvas-id> <widget-id>

  # Update widget position
  canvus widget update <canvas-id> <widget-id> --x 100 --y 200

  # Delete a widget
  canvus widget delete <canvas-id> <widget-id>`,
}

// getSession retrieves the SDK session from the command context.
func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}

// validateWidgetType validates that a widget type is one of the supported types.
func validateWidgetType(widgetType string) error {
	validTypes := map[string]bool{
		"note":      true,
		"image":     true,
		"pdf":       true,
		"video":     true,
		"browser":   true,
		"anchor":    true,
		"connector": true,
	}

	if !validTypes[widgetType] {
		return fmt.Errorf("invalid widget type '%s'. Valid types: note, image, pdf, video, browser, anchor, connector", widgetType)
	}

	return nil
}

// validatePosition validates widget position coordinates.
func validatePosition(x, y float64) error {
	// Basic validation - coordinates should be reasonable
	// No strict limits as canvas can be large
	return nil
}

// validateSize validates widget size dimensions.
func validateSize(width, height float64) error {
	if width < 0 || height < 0 {
		return fmt.Errorf("width and height must be non-negative")
	}
	return nil
}

// validateCanvasID validates that a canvas ID is provided and not empty.
func validateCanvasID(id string) error {
	if id == "" {
		return fmt.Errorf("canvas ID is required")
	}
	return nil
}

// validateWidgetID validates that a widget ID is provided and not empty.
func validateWidgetID(id string) error {
	if id == "" {
		return fmt.Errorf("widget ID is required")
	}
	return nil
}

// parseFloat is a helper to parse float values from flags.
func parseFloat(s string, defaultValue float64) (float64, error) {
	if s == "" {
		return defaultValue, nil
	}
	return strconv.ParseFloat(s, 64)
}
