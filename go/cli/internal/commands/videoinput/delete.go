package videoinput

import (
	"context"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <canvas-id> <input-id>",
	Short: "Delete video input",
	Args:  cobra.ExactArgs(2),
	RunE:  runDelete,
}

func init() {
	VideoInputCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	canvasID, inputID := args[0], args[1]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	err = sess.DeleteVideoInput(context.Background(), canvasID, inputID)
	if err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Video input deleted successfully")
	return nil
}
