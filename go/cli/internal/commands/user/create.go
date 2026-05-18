package user

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user",
	Long: `Create a new user account in the Canvus workspace.

Requires administrative privileges. You must provide a name, username,
and email address. Optionally, you can specify a password.

Examples:
  # Create a user with all required fields
  canvus user create --name "John Doe" --email "john@example.com"

  # Create a user with a password
  canvus user create --name "Jane Smith" --email "jane@example.com" --password "SecurePass123"`,
	RunE: runCreate,
}

var (
	createName     string
	createEmail    string
	createPassword string
)

func init() {
	createCmd.Flags().StringVar(&createName, "name", "", "Full name of the user (required)")
	createCmd.Flags().StringVar(&createEmail, "email", "", "Email address (required)")
	createCmd.Flags().StringVar(&createPassword, "password", "", "Password (optional)")

	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("email")

	UserCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Validate inputs
	if err := validateName(createName); err != nil {
		return err
	}
	if err := validateEmail(createEmail); err != nil {
		return err
	}

	// Get SDK session
	sess, err := getSession(cmd)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Build user request
	req := canvus.CreateUserRequest{
		Name:  createName,
		Email: createEmail,
	}

	// Add password if provided
	if createPassword != "" {
		req.Password = createPassword
	}

	// Create user via SDK
	ctx := context.Background()
	user, err := sess.CreateUser(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Output result
	return output.PrintOutput(cmd.Context(), user, "")
}
