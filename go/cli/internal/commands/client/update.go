package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

// updateCmd — see create.go for rationale. No /clients/{id} update endpoint
// exists in the public Canvus API; preserve the command-tree shape, surface
// the gap clearly to anyone scripting against the CLI.
var updateCmd = &cobra.Command{
	Use:    "update <client-id>",
	Short:  "(unsupported) Update a client",
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("client update is not supported by the Canvus API (Phase 4d gap)")
	},
}

func init() {
	updateCmd.Flags().String("name", "", "(unused) new client name")
	ClientCmd.AddCommand(updateCmd)
}
