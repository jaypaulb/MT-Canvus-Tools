package colorpreset

import (
	"context"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var getAllCmd = &cobra.Command{
	Use:   "get-all <canvas-id>",
	Short: "Get all color presets",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetAll,
}

func init() {
	ColorPresetCmd.AddCommand(getAllCmd)
}

func runGetAll(cmd *cobra.Command, args []string) error {
	canvasID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	presets, err := sess.GetColorPresets(context.Background(), canvasID)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), presets, "")
}
