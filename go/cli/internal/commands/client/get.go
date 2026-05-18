package client

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <client-id>",
	Short: "Get client details",
	Long: `Get details of a specific client by ID.

Examples:
  # Get client details
  canvus client get client-123

  # Get client as JSON
  canvus client get client-123 --output json`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

func init() {
	ClientCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	clientID := args[0]

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	client, err := sess.GetClient(context.Background(), clientID)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), client, "")
}
