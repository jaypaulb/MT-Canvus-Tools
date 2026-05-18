package group

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var removeUserCmd = &cobra.Command{
	Use:   "remove-user <group-id> <user-id>",
	Short: "Remove a user from a group",
	Long: `Remove a user from a group in the Canvus workspace.

Requires administrative privileges. This removes the specified user from
the specified group, revoking any permissions associated with the group.

Examples:
  # Remove a user from a group
  canvus group remove-user 123 456

  # Remove a user with JSON output
  canvus group remove-user 123 456 --output json`,
	Args: cobra.ExactArgs(2),
	RunE: runRemoveUser,
}

func init() {
	GroupCmd.AddCommand(removeUserCmd)
}

func runRemoveUser(cmd *cobra.Command, args []string) error {
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

	// Remove user from group via SDK
	ctx := context.Background()
	if err := sess.RemoveUserFromGroup(ctx, groupID, userID); err != nil {
		return fmt.Errorf("failed to remove user from group: %w", err)
	}

	// Output success message
	output.OutputSuccess(cmd.Context(), fmt.Sprintf("User %s removed from group %s successfully", userIDStr, groupIDStr))
	return nil
}
