// Package token provides commands for managing access tokens in Canvus.
package token

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

// TokenCmd is the parent command for all access token-related operations.
var TokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage access tokens",
	Long: `Manage API access tokens for users in your Canvus system.

Examples:
  # List access tokens for a user
  canvus token list --user-id 123

  # Create a new access token
  canvus token create --user-id 123 --description "My Token"

  # Delete an access token
  canvus token delete --user-id 123 <token-id>`,
}

// getSession retrieves the SDK session from the command context.
func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
