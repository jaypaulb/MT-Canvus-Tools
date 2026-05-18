package colorpreset

import (
	"context"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <canvas-id> <name>",
	Short: "Delete color preset",
	Args:  cobra.ExactArgs(2),
	RunE:  runDelete,
}

func init() {
	ColorPresetCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	canvasID, name := args[0], args[1]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	err = sess.DeleteColorPreset(context.Background(), canvasID, name)
	if err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Color preset deleted successfully")
	return nil
}
