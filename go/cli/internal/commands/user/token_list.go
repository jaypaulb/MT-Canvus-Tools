package user

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var listTokensCmd = &cobra.Command{
	Use:   "list <user-id>",
	Short: "List all API tokens for a user",
	Long: `List all API tokens for the specified user.

Requires administrative privileges. Returns a list of all tokens created
for the user, including their names and creation dates.

Examples:
  # List all tokens for a user
  canvus user token list 123

  # List tokens in JSON format
  canvus user token list 123 --output json`,
	Args: cobra.ExactArgs(1),
	RunE: runListTokens,
}

func init() {
	tokenCmd.AddCommand(listTokensCmd)
}

func runListTokens(cmd *cobra.Command, args []string) error {
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

	// List tokens via SDK
	ctx := context.Background()
	tokens, err := sess.ListAccessTokens(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to list access tokens: %w", err)
	}

	// Output results
	return output.PrintOutput(cmd.Context(), tokens, "")
}
