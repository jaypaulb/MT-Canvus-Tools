package videoinput

import (
	"context"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <canvas-id>",
	Short: "List video inputs for a canvas",
	Args:  cobra.ExactArgs(1),
	RunE:  runList,
}

func init() {
	VideoInputCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	canvasID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	inputs, err := sess.ListVideoInputs(context.Background(), canvasID)
	if err != nil {
		return err
	}

	items := make([]interface{}, len(inputs))
	for i, input := range inputs {
		items[i] = input
	}
	return output.OutputList(cmd.Context(), items, "")
}
