package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

// createCmd is preserved as a user-visible command for command-tree
// compatibility, but the Canvus REST API does not expose a client-creation
// endpoint — clients are connected processes the server discovers, not
// records callers can POST. The legacy CLI shipped this command against a
// legacy-SDK helper that hit a non-existent endpoint and would have
// returned 404. Routing to a clear "not supported" error stops users from
// believing the operation could ever succeed.
//
// Tracked as a Phase 4d SDK gap if a real endpoint is later confirmed.
var createCmd = &cobra.Command{
	Use:    "create",
	Short:  "(unsupported) Create a new client",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("client create is not supported by the Canvus API (Phase 4d gap)")
	},
}

func init() {
	// Flags are retained on the parent so scripts that pass --name / --user-id
	// don't fail flag-parsing before the command body returns its error.
	createCmd.Flags().String("name", "", "(unused) client name")
	createCmd.Flags().String("user-id", "", "(unused) user ID")
	ClientCmd.AddCommand(createCmd)
}
