package canvas

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var restoreDemoCmd = &cobra.Command{
	Use:   "restore-demo <canvas-id>",
	Short: "Restore demo canvas state",
	Args:  cobra.ExactArgs(1),
	RunE:  runRestoreDemo,
}

func init() {
	CanvasCmd.AddCommand(restoreDemoCmd)
}

func runRestoreDemo(cmd *cobra.Command, args []string) error {
	canvasID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	err = sess.RestoreDemoCanvas(context.Background(), canvasID)
	if err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Demo canvas restored successfully")
	return nil
}
