package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deleteCmd — see create.go for rationale. The Canvus API has no
// client-delete endpoint; clients disconnect on their own.
var deleteCmd = &cobra.Command{
	Use:    "delete <client-id>",
	Short:  "(unsupported) Delete a client",
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("client delete is not supported by the Canvus API (Phase 4d gap)")
	},
}

func init() {
	ClientCmd.AddCommand(deleteCmd)
}
