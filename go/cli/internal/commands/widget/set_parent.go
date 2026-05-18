package widget

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var setParentCmd = &cobra.Command{
	Use:   "set-parent <canvas-id> <widget-id> <parent-id>",
	Short: "Set widget parent",
	Args:  cobra.ExactArgs(3),
	RunE:  runSetParent,
}

func init() {
	WidgetCmd.AddCommand(setParentCmd)
}

func runSetParent(cmd *cobra.Command, args []string) error {
	canvasID := args[0]
	widgetID := args[1]
	parentID := args[2]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	widget, err := sess.PatchParentID(context.Background(), canvasID, widgetID, parentID)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), widget, "")
}
