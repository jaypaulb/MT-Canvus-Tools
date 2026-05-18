package anchor

import (
	"context"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <canvas-id> <anchor-id>",
	Short: "Delete anchor",
	Args:  cobra.ExactArgs(2),
	RunE:  runDelete,
}

func init() {
	AnchorCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	canvasID, itemID := args[0], args[1]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	err = sess.DeleteAnchor(context.Background(), canvasID, itemID)
	if err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Anchor deleted successfully")
	return nil
}
