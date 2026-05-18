package widget

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <canvas-id> <widget-id>",
	Short: "Delete a widget",
	Long: `Delete a widget from a canvas.

This command removes the widget from the canvas permanently.
You must specify the widget type with the --type flag.

Examples:
  # Delete a note widget
  canvus widget delete canvas-123 widget-456 --type note

  # Delete an image widget
  canvus widget delete canvas-123 widget-456 --type image

  # Delete widget with confirmation skipped (for scripts)
  canvus widget delete canvas-123 widget-456 --type note --force`,
	Args: cobra.ExactArgs(2),
	RunE: runDelete,
}

var (
	deleteForce bool
	deleteType  string
)

func init() {
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "Skip confirmation prompt")
	deleteCmd.Flags().StringVar(&deleteType, "type", "", "Widget type (required): note, image, pdf, video, anchor, connector")
	deleteCmd.MarkFlagRequired("type")
	WidgetCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	canvasID := args[0]
	widgetID := args[1]

	// Validate IDs
	if err := validateCanvasID(canvasID); err != nil {
		return err
	}
	if err := validateWidgetID(widgetID); err != nil {
		return err
	}

	// Validate widget type
	if err := validateWidgetType(deleteType); err != nil {
		return err
	}

	// Get SDK session from context
	sess, err := session.GetSession(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Delete widget using SDK (requires widget type)
	ctx := context.Background()
	err = sess.DeleteWidget(ctx, canvasID, widgetID, deleteType)
	if err != nil {
		return fmt.Errorf("failed to delete widget: %w", err)
	}

	// Output success message
	successMsg := fmt.Sprintf("Successfully deleted %s widget %s from canvas %s", deleteType, widgetID, canvasID)
	output.OutputSuccess(cmd.Context(), successMsg)
	return nil
}
