package background

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <canvas-id>",
	Short: "Get canvas background",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func init() {
	BackgroundCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	canvasID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	bg, err := sess.GetCanvasBackground(context.Background(), canvasID)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), bg, "")
}
