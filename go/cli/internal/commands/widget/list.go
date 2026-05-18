package widget

import (
	"context"
	"fmt"
	"strings"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <canvas-id>",
	Short: "List widgets in a canvas",
	Long: `List all widgets in the specified canvas.

You can optionally filter widgets by type.

Examples:
  # List all widgets in a canvas
  canvus widget list canvas-123

  # List only note widgets
  canvus widget list canvas-123 --type note

  # List widgets as JSON
  canvus widget list canvas-123 --output json`,
	Args: cobra.ExactArgs(1),
	RunE: runList,
}

var (
	listType string
)

func init() {
	listCmd.Flags().StringVar(&listType, "type", "", "Filter by widget type")
	WidgetCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	canvasID := args[0]

	// Validate canvas ID
	if err := validateCanvasID(canvasID); err != nil {
		return err
	}

	// Validate widget type if provided
	if listType != "" {
		if err := validateWidgetType(listType); err != nil {
			return err
		}
	}

	// Get SDK session from context
	sess, err := session.GetSession(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Build filter for SDK if type is specified
	var filter *canvus.Filter
	if listType != "" {
		filter = &canvus.Filter{
			Criteria: map[string]interface{}{
				"widget_type": listType,
			},
		}
	}

	// List widgets using SDK
	ctx := context.Background()
	widgets, err := sess.ListWidgets(ctx, canvasID, filter)
	if err != nil {
		return fmt.Errorf("failed to list widgets: %w", err)
	}

	// Apply additional client-side type filtering if needed (as fallback)
	var filteredWidgets []canvus.Widget
	if listType != "" {
		for _, widget := range widgets {
			if strings.EqualFold(widget.WidgetType, listType) {
				filteredWidgets = append(filteredWidgets, widget)
			}
		}
	} else {
		filteredWidgets = widgets
	}

	// Convert to []interface{} for output formatting
	items := make([]interface{}, len(filteredWidgets))
	for i, widget := range filteredWidgets {
		items[i] = widget
	}

	// Format and output the results
	return output.OutputList(cmd.Context(), items, "")
}
