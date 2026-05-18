package user

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	Long: `List all users in the Canvus workspace.

Requires administrative privileges. Returns a list of all user accounts
with their details.

Examples:
  # List all users in table format
  canvus user list

  # List all users in JSON format
  canvus user list --output json`,
	RunE: runList,
}

func init() {
	UserCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	// Get SDK session
	sess, err := getSession(cmd)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// List users via SDK
	ctx := context.Background()
	users, err := sess.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	// Output results
	return output.PrintOutput(cmd.Context(), users, "")
}
