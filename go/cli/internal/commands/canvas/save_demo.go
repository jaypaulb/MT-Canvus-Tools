package canvas

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var saveDemoCmd = &cobra.Command{
	Use:   "save-demo <canvas-id>",
	Short: "Save demo canvas state",
	Args:  cobra.ExactArgs(1),
	RunE:  runSaveDemo,
}

func init() {
	CanvasCmd.AddCommand(saveDemoCmd)
}

func runSaveDemo(cmd *cobra.Command, args []string) error {
	canvasID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	err = sess.SaveDemoState(context.Background(), canvasID)
	if err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Demo state saved successfully")
	return nil
}
