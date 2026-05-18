package canvas

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var trashCmd = &cobra.Command{
	Use:   "trash <canvas-id>",
	Short: "Move canvas to trash",
	Args:  cobra.ExactArgs(1),
	RunE:  runTrash,
}

func init() {
	CanvasCmd.AddCommand(trashCmd)
}

func runTrash(cmd *cobra.Command, args []string) error {
	canvasID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	canvas, err := sess.TrashCanvas(context.Background(), canvasID, "")
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), canvas, "")
}
