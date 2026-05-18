package colorpreset

import (
	"context"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <canvas-id>",
	Short: "List color presets",
	Args:  cobra.ExactArgs(1),
	RunE:  runList,
}

func init() {
	ColorPresetCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	canvasID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	presets, err := sess.ListColorPresets(context.Background(), canvasID)
	if err != nil {
		return err
	}

	items := make([]interface{}, len(presets))
	for i, p := range presets {
		items[i] = p
	}
	return output.OutputList(cmd.Context(), items, "")
}
