// Package client provides commands for managing clients in Canvus.
package client

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

// ClientCmd is the parent command for all client-related operations.
var ClientCmd = &cobra.Command{
	Use:   "client",
	Short: "Manage clients",
	Long: `Manage clients in your Canvus system.

Clients represent connected devices or browser sessions in Canvus.

Examples:
  # List all clients
  canvus client list

  # Get client details
  canvus client get <client-id>

  # Create a new client
  canvus client create --name "My Client" --user-id <user-id>

  # Update a client
  canvus client update <client-id> --name "New Name"

  # Delete a client
  canvus client delete <client-id>`,
}

// getSession retrieves the SDK session from the command context.
func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
