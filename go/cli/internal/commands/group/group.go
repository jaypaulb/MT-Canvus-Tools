// Package group provides commands for managing groups in Canvus.
package group

import (
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

// GroupCmd is the parent command for all group-related operations.
var GroupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage groups",
	Long: `Manage groups in your Canvus workspace.

Groups are collections of users that can be used to manage permissions
and access control. Group management operations require administrative
privileges.

Examples:
  # Create a new group
  canvus group create --name "Development Team"

  # List all groups
  canvus group list

  # Get group details
  canvus group get <group-id>

  # Add a user to a group
  canvus group add-user <group-id> <user-id>

  # Remove a user from a group
  canvus group remove-user <group-id> <user-id>

  # Delete a group
  canvus group delete <group-id>`,
}

// getSession retrieves the SDK session from the command context.
// This is a helper function used by all group commands.
func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}

// validateGroupID validates that a group ID is provided and not empty.
func validateGroupID(id string) error {
	if id == "" {
		return fmt.Errorf("group ID is required")
	}
	return nil
}

// validateGroupName validates that a group name is provided and not empty.
func validateGroupName(name string) error {
	if name == "" {
		return fmt.Errorf("group name is required")
	}
	return nil
}
