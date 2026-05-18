package videoinput

import (
	"context"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var listClientCmd = &cobra.Command{
	Use:   "list-client <client-id>",
	Short: "List video input sources for a client",
	Args:  cobra.ExactArgs(1),
	RunE:  runListClient,
}

func init() {
	VideoInputCmd.AddCommand(listClientCmd)
}

func runListClient(cmd *cobra.Command, args []string) error {
	clientID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	sources, err := sess.ListClientVideoInputs(context.Background(), clientID)
	if err != nil {
		return err
	}

	items := make([]interface{}, len(sources))
	for i, source := range sources {
		items[i] = source
	}
	return output.OutputList(cmd.Context(), items, "")
}
