package user

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <user-id>",
	Short: "Delete a user",
	Long: `Delete a user account from the Canvus workspace.

Requires administrative privileges. This is a destructive operation that
permanently removes the user account. By default, a confirmation prompt
will be displayed unless --force is used.

Examples:
  # Delete a user (with confirmation)
  canvus user delete 123

  # Delete a user without confirmation
  canvus user delete 123 --force`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

var deleteForce bool

func init() {
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "Skip confirmation prompt")

	UserCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	userIDStr := args[0]

	// Parse user ID as int64
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user ID %q: must be a number", userIDStr)
	}

	// Confirm deletion unless --force is used
	if !deleteForce {
		confirmed, err := confirmDeletion(userIDStr)
		if err != nil {
			return err
		}
		if !confirmed {
			return fmt.Errorf("user deletion cancelled")
		}
	}

	// Get SDK session
	sess, err := getSession(cmd)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Delete user via SDK
	ctx := context.Background()
	if err := sess.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Output success message
	output.OutputSuccess(cmd.Context(), fmt.Sprintf("User %s deleted successfully", userIDStr))
	return nil
}

// confirmDeletion prompts the user to confirm deletion of a user.
func confirmDeletion(userID string) (bool, error) {
	fmt.Printf("Are you sure you want to delete user %s? This action cannot be undone. [y/N]: ", userID)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}
