package connector

import (
	"context"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <canvas-id> <connector-id>",
	Short: "Delete connector",
	Args:  cobra.ExactArgs(2),
	RunE:  runDelete,
}

func init() {
	ConnectorCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	canvasID, itemID := args[0], args[1]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	err = sess.DeleteConnector(context.Background(), canvasID, itemID)
	if err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Connector deleted successfully")
	return nil
}
