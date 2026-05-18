package group

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var addUserCmd = &cobra.Command{
	Use:   "add-user <group-id> <user-id>",
	Short: "Add a user to a group",
	Long: `Add a user to a group in the Canvus workspace.

Requires administrative privileges. This adds the specified user to the
specified group, granting them any permissions associated with the group.

Examples:
  # Add a user to a group
  canvus group add-user 123 456

  # Add a user with JSON output
  canvus group add-user 123 456 --output json`,
	Args: cobra.ExactArgs(2),
	RunE: runAddUser,
}

func init() {
	GroupCmd.AddCommand(addUserCmd)
}

func runAddUser(cmd *cobra.Command, args []string) error {
	groupIDStr := args[0]
	userIDStr := args[1]

	// Parse group ID as int
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		return fmt.Errorf("invalid group ID %q: must be a number", groupIDStr)
	}

	// Parse user ID as int64 (SDK signature).
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user ID %q: must be a number", userIDStr)
	}

	// Get SDK session
	sess, err := getSession(cmd)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Add user to group via SDK
	ctx := context.Background()
	if err := sess.AddUserToGroup(ctx, groupID, userID); err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	// Output success message
	output.OutputSuccess(cmd.Context(), fmt.Sprintf("User %s added to group %s successfully", userIDStr, groupIDStr))
	return nil
}
