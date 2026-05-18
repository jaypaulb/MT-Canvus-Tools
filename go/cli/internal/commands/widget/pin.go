package widget

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

// pinCmd pins a widget on a canvas. The Canvus API has no dedicated
// /widgets/{id}/pin endpoint — "pinned" is a property of the widget itself
// (see docs/api-reference/endpoints/widgets.md). Pinning is therefore a
// PATCH on the widget that flips `pinned: true`.
//
// Behaviour note: the legacy CLI accepted a single widget-id argument and
// relied on a /widgets/{id}/pin endpoint that does not exist in the real
// Canvus API. The new signature requires the canvas-id because the
// underlying SDK call is UpdateWidget(canvas, widget, ...).
var pinCmd = &cobra.Command{
	Use:   "pin <canvas-id> <widget-id>",
	Short: "Pin a widget (sets pinned=true via PATCH)",
	Args:  cobra.ExactArgs(2),
	RunE:  runPin,
}

func init() {
	WidgetCmd.AddCommand(pinCmd)
}

func runPin(cmd *cobra.Command, args []string) error {
	canvasID, widgetID := args[0], args[1]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	if _, err := sess.UpdateWidget(context.Background(), canvasID, widgetID, map[string]any{
		"pinned": true,
	}); err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Widget pinned successfully")
	return nil
}
