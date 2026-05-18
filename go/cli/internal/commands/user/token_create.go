package user

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var createTokenCmd = &cobra.Command{
	Use:   "create <user-id>",
	Short: "Create a new API token for a user",
	Long: `Create a new API token for the specified user.

Requires administrative privileges. The token can be used for authentication
instead of username/password. You must provide a name for the token to
identify its purpose.

Examples:
  # Create a token for CI/CD
  canvus user token create 123 --name "CI Token"

  # Create a token for automation
  canvus user token create 456 --name "Automation Script"`,
	Args: cobra.ExactArgs(1),
	RunE: runCreateToken,
}

var tokenName string

func init() {
	createTokenCmd.Flags().StringVar(&tokenName, "name", "", "Name/description for the token (required)")
	createTokenCmd.MarkFlagRequired("name")

	tokenCmd.AddCommand(createTokenCmd)
}

func runCreateToken(cmd *cobra.Command, args []string) error {
	userIDStr := args[0]

	// Parse user ID as int64
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user ID %q: must be a number", userIDStr)
	}

	// Validate token name
	if tokenName == "" {
		return fmt.Errorf("token name is required")
	}

	// Get SDK session
	sess, err := getSession(cmd)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Create token request
	req := canvus.CreateAccessTokenRequest{
		Description: tokenName,
	}

	// Create token via SDK
	ctx := context.Background()
	token, err := sess.CreateAccessToken(ctx, userID, req)
	if err != nil {
		return fmt.Errorf("failed to create access token: %w", err)
	}

	// Output result
	return output.PrintOutput(cmd.Context(), token, "")
}
