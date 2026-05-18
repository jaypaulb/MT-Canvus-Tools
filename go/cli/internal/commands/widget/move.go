package widget

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

// moveCmd performs a cross-canvas widget move. The Canvus API exposes no
// dedicated widget-move endpoint, so the operation is implemented as
// CloneWidget (into the destination) followed by DeleteWidget (on the
// source). If the clone fails the source is left untouched; if the delete
// fails the new clone is left in place and the original is returned in the
// error context.
//
// Behaviour note: the legacy CLI accepted (widget-id, target-canvas-id) and
// relied on a /widgets/{id}/move endpoint that does not exist in the real
// Canvus API. The new signature requires the source canvas-id so the SDK
// can resolve the widget's type for the clone+delete round-trip.
var moveCmd = &cobra.Command{
	Use:   "move <source-canvas-id> <widget-id> <target-canvas-id>",
	Short: "Move a widget to another canvas (clone then delete original)",
	Args:  cobra.ExactArgs(3),
	RunE:  runMove,
}

func init() {
	WidgetCmd.AddCommand(moveCmd)
}

func runMove(cmd *cobra.Command, args []string) error {
	sourceCanvasID, widgetID, targetCanvasID := args[0], args[1], args[2]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	ctx := context.Background()
	src, err := sess.GetWidget(ctx, sourceCanvasID, widgetID)
	if err != nil {
		return fmt.Errorf("look up source widget: %w", err)
	}
	wtPath, err := widgetTypePath(src.WidgetType)
	if err != nil {
		return err
	}

	clone, err := sess.CloneWidget(ctx, targetCanvasID, sourceCanvasID, widgetID, wtPath, nil)
	if err != nil {
		return fmt.Errorf("clone into target canvas: %w", err)
	}

	if err := sess.DeleteWidget(ctx, sourceCanvasID, widgetID, src.WidgetType); err != nil {
		return fmt.Errorf("clone succeeded (new widget id %s in canvas %s) but delete of original failed: %w",
			clone.ID, targetCanvasID, err)
	}

	return output.OutputSingle(cmd.Context(), clone, "Widget moved successfully")
}
