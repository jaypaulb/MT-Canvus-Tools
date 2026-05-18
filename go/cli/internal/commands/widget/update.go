package widget

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <canvas-id> <widget-id>",
	Short: "Update widget properties",
	Long: `Update properties of an existing widget.

You can update text, position, size, or other widget properties.
Only the properties you specify will be updated.

Examples:
  # Update widget text
  canvus widget update canvas-123 widget-456 --text "Updated text"

  # Update widget position
  canvus widget update canvas-123 widget-456 --x 100 --y 200

  # Update widget size
  canvus widget update canvas-123 widget-456 --width 300 --height 200

  # Update multiple properties
  canvus widget update canvas-123 widget-456 --text "New text" --x 50 --y 75`,
	Args: cobra.ExactArgs(2),
	RunE: runUpdate,
}

var (
	updateText   string
	updateParent string
	updateX      *float64
	updateY      *float64
	updateWidth  *float64
	updateHeight *float64
)

func init() {
	updateCmd.Flags().StringVar(&updateText, "text", "", "New text content")
	updateCmd.Flags().StringVar(&updateParent, "parent", "", "Parent widget ID")

	// Use float64 pointers to distinguish between "not set" and "set to 0"
	var x, y, w, h float64
	updateCmd.Flags().Float64Var(&x, "x", 0, "New X coordinate")
	updateCmd.Flags().Float64Var(&y, "y", 0, "New Y coordinate")
	updateCmd.Flags().Float64Var(&w, "width", 0, "New width")
	updateCmd.Flags().Float64Var(&h, "height", 0, "New height")

	// Store pointers if flags were actually set
	updateCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed("x") {
			updateX = &x
		}
		if cmd.Flags().Changed("y") {
			updateY = &y
		}
		if cmd.Flags().Changed("width") {
			updateWidth = &w
		}
		if cmd.Flags().Changed("height") {
			updateHeight = &h
		}
	}

	WidgetCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	canvasID := args[0]
	widgetID := args[1]

	// Validate IDs
	if err := validateCanvasID(canvasID); err != nil {
		return err
	}
	if err := validateWidgetID(widgetID); err != nil {
		return err
	}

	// Get SDK session from context
	sess, err := session.GetSession(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// First, get the widget to determine its type
	ctx := context.Background()
	widget, err := sess.GetWidget(ctx, canvasID, widgetID)
	if err != nil {
		return fmt.Errorf("failed to get widget: %w", err)
	}

	// Build update data with only provided fields
	updateData := make(map[string]interface{})
	updateData["widget_type"] = widget.WidgetType

	if updateText != "" {
		updateData["text"] = updateText
	}

	if updateParent != "" {
		updateData["parent"] = updateParent
	}

	// Update location if either x or y is set
	if updateX != nil || updateY != nil {
		location := make(map[string]interface{})
		if updateX != nil {
			location["x"] = *updateX
		}
		if updateY != nil {
			location["y"] = *updateY
		}
		updateData["location"] = location
	}

	// Update size if either width or height is set
	if updateWidth != nil || updateHeight != nil {
		size := make(map[string]interface{})
		if updateWidth != nil {
			size["width"] = *updateWidth
		}
		if updateHeight != nil {
			size["height"] = *updateHeight
		}
		updateData["size"] = size
	}

	// Check if any fields were provided (besides widget_type which we always add)
	if len(updateData) == 1 {
		return fmt.Errorf("at least one field must be provided for update (--text, --parent, --x, --y, --width, --height)")
	}

	// Update widget using SDK
	updatedWidget, err := sess.UpdateWidget(ctx, canvasID, widgetID, updateData)
	if err != nil {
		return fmt.Errorf("failed to update widget: %w", err)
	}

	// Format and output the result
	return output.OutputSingle(cmd.Context(), updatedWidget, "")
}
