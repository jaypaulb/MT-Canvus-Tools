package user

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <user-id>",
	Short: "Update user information",
	Long: `Update information for an existing user.

Requires administrative privileges. You can update the user's name and/or
email address. Only the specified fields will be updated.

Examples:
  # Update user's name
  canvus user update 123 --name "Jane Doe"

  # Update user's email
  canvus user update 123 --email "jane.new@example.com"

  # Update both name and email
  canvus user update 123 --name "Jane Smith" --email "jane.smith@example.com"`,
	Args: cobra.ExactArgs(1),
	RunE: runUpdate,
}

var (
	updateName  string
	updateEmail string
)

func init() {
	updateCmd.Flags().StringVar(&updateName, "name", "", "New full name for the user")
	updateCmd.Flags().StringVar(&updateEmail, "email", "", "New email address for the user")

	UserCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	userIDStr := args[0]

	// Parse user ID as int64
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user ID %q: must be a number", userIDStr)
	}

	// Check that at least one field is being updated
	if updateName == "" && updateEmail == "" {
		return fmt.Errorf("at least one field must be specified for update (--name or --email)")
	}

	// Build update request with only provided fields
	req := canvus.UpdateUserRequest{}

	if updateName != "" {
		if err := validateName(updateName); err != nil {
			return err
		}
		req.Name = &updateName
	}

	if updateEmail != "" {
		if err := validateEmail(updateEmail); err != nil {
			return err
		}
		req.Email = &updateEmail
	}

	// Get SDK session
	sess, err := getSession(cmd)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Update user via SDK
	ctx := context.Background()
	user, err := sess.UpdateUser(ctx, userID, req)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Output result
	return output.PrintOutput(cmd.Context(), user, "")
}
