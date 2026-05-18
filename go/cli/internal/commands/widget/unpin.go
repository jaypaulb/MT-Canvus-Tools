package widget

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

// unpinCmd is the mirror of pinCmd — see pin.go for the API rationale.
var unpinCmd = &cobra.Command{
	Use:   "unpin <canvas-id> <widget-id>",
	Short: "Unpin a widget (sets pinned=false via PATCH)",
	Args:  cobra.ExactArgs(2),
	RunE:  runUnpin,
}

func init() {
	WidgetCmd.AddCommand(unpinCmd)
}

func runUnpin(cmd *cobra.Command, args []string) error {
	canvasID, widgetID := args[0], args[1]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	if _, err := sess.UpdateWidget(context.Background(), canvasID, widgetID, map[string]any{
		"pinned": false,
	}); err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Widget unpinned successfully")
	return nil
}
