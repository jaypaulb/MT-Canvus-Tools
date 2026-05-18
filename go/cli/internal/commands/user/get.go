package user

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <user-id>",
	Short: "Get user details",
	Long: `Get detailed information about a specific user.

Requires administrative privileges. Retrieves and displays all information
about the specified user account.

Examples:
  # Get user details
  canvus user get 123

  # Get user details in JSON format
  canvus user get 123 --output json`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

func init() {
	UserCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	userIDStr := args[0]

	// Parse user ID as int64
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user ID %q: must be a number", userIDStr)
	}

	// Get SDK session
	sess, err := getSession(cmd)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Get user via SDK
	ctx := context.Background()
	user, err := sess.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Output result
	return output.PrintOutput(cmd.Context(), user, "")
}
