package image

import (
	"context"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <canvas-id> <image-id>",
	Short: "Delete image",
	Args:  cobra.ExactArgs(2),
	RunE:  runDelete,
}

func init() {
	ImageCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	canvasID, itemID := args[0], args[1]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	err = sess.DeleteImage(context.Background(), canvasID, itemID)
	if err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Image deleted successfully")
	return nil
}
