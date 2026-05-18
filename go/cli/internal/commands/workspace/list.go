package workspace

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <client-id>",
	Short: "List all workspaces for a client",
	Args:  cobra.ExactArgs(1),
	RunE:  runList,
}

func init() {
	WorkspaceCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	clientID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	workspaces, err := sess.ListWorkspaces(context.Background(), clientID)
	if err != nil {
		return err
	}

	items := make([]interface{}, len(workspaces))
	for i, ws := range workspaces {
		items[i] = ws
	}
	return output.OutputList(cmd.Context(), items, "")
}
