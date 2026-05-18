package videooutput

import (
	"context"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <client-id>",
	Short: "List video outputs for a client",
	Args:  cobra.ExactArgs(1),
	RunE:  runList,
}

func init() {
	VideoOutputCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	clientID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	outputs, err := sess.ListVideoOutputs(context.Background(), clientID)
	if err != nil {
		return err
	}

	items := make([]interface{}, len(outputs))
	for i, out := range outputs {
		items[i] = out
	}
	return output.OutputList(cmd.Context(), items, "")
}
