package client

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all clients",
	Long: `List all clients in the Canvus system.

Examples:
  # List all clients
  canvus client list

  # List clients as JSON
  canvus client list --output json`,
	RunE: runList,
}

func init() {
	ClientCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	clients, err := sess.ListClients(context.Background())
	if err != nil {
		return err
	}

	items := make([]interface{}, len(clients))
	for i, client := range clients {
		items[i] = client
	}

	return output.OutputList(cmd.Context(), items, "")
}
