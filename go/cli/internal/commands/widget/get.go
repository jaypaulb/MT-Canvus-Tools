package widget

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <canvas-id> <widget-id>",
	Short: "Get widget details",
	Long: `Retrieve detailed information about a specific widget.

This command displays all properties of a widget including its type, content,
location, size, and other metadata.

Examples:
  # Get widget details
  canvus widget get canvas-123 widget-456

  # Get widget details as JSON
  canvus widget get canvas-123 widget-456 --output json`,
	Args: cobra.ExactArgs(2),
	RunE: runGet,
}

func init() {
	WidgetCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
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

	// Get widget using SDK
	ctx := context.Background()
	widget, err := sess.GetWidget(ctx, canvasID, widgetID)
	if err != nil {
		return fmt.Errorf("failed to get widget: %w", err)
	}

	// Format and output the result
	return output.OutputSingle(cmd.Context(), widget, "")
}
