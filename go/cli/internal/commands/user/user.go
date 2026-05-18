// Package user provides commands for managing users in Canvus.
package user

import (
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

// UserCmd is the parent command for all user-related operations.
var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long: `Manage users in your Canvus workspace.

Users are accounts that can access the Canvus system. User management
operations require administrative privileges.

Examples:
  # Create a new user
  canvus user create --name "John Doe" --username "jdoe" --email "john@example.com"

  # List all users
  canvus user list

  # Get user details
  canvus user get <user-id>

  # Update user information
  canvus user update <user-id> --name "Jane Doe"

  # Delete a user
  canvus user delete <user-id>

  # Create API token for a user
  canvus user token create <user-id> --name "CI Token"

  # List user's API tokens
  canvus user token list <user-id>`,
}

// getSession retrieves the SDK session from the command context.
// This is a helper function used by all user commands.
func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}

// validateUserID validates that a user ID is provided and not empty.
func validateUserID(id string) error {
	if id == "" {
		return fmt.Errorf("user ID is required")
	}
	return nil
}

// validateEmail validates that an email address is provided and not empty.
func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	// Basic email validation - just check for @ symbol
	// The server will do more thorough validation
	if len(email) < 3 || email[0] == '@' || email[len(email)-1] == '@' {
		return fmt.Errorf("invalid email format")
	}
	hasAt := false
	for _, c := range email {
		if c == '@' {
			hasAt = true
			break
		}
	}
	if !hasAt {
		return fmt.Errorf("email must contain @ symbol")
	}
	return nil
}

// validateUsername validates that a username is provided and not empty.
func validateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username is required")
	}
	return nil
}

// validateName validates that a name is provided and not empty.
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}
