package colorpreset

import (
	"context"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <canvas-id> <name>",
	Short: "Get color preset by name",
	Args:  cobra.ExactArgs(2),
	RunE:  runGet,
}

func init() {
	ColorPresetCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	canvasID, name := args[0], args[1]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	preset, err := sess.GetColorPreset(context.Background(), canvasID, name)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), preset, "")
}
