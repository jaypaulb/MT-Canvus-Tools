package user

import (
	"github.com/spf13/cobra"
)

// tokenCmd is the parent command for user token operations.
var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage user API tokens",
	Long: `Manage API tokens for users.

API tokens provide an alternative authentication method to username/password.
Tokens can be created and listed for users.

Examples:
  # Create a new API token for a user
  canvus user token create <user-id> --name "CI Token"

  # List all tokens for a user
  canvus user token list <user-id>`,
}

func init() {
	UserCmd.AddCommand(tokenCmd)
}
