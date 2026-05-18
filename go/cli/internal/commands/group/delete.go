package group

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
	Use:   "delete <group-id>",
	Short: "Delete a group",
	Long: `Delete a group from the Canvus workspace.

Requires administrative privileges. This is a destructive operation that
permanently removes the group. By default, a confirmation prompt will be
displayed unless --force is used.

Examples:
  # Delete a group (with confirmation)
  canvus group delete 123

  # Delete a group without confirmation
  canvus group delete 123 --force`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

var deleteForce bool

func init() {
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "Skip confirmation prompt")

	GroupCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	groupIDStr := args[0]

	// Parse group ID as int
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		return fmt.Errorf("invalid group ID %q: must be a number", groupIDStr)
	}

	// Confirm deletion unless --force is used
	if !deleteForce {
		confirmed, err := confirmDeletion(groupIDStr)
		if err != nil {
			return err
		}
		if !confirmed {
			return fmt.Errorf("group deletion cancelled")
		}
	}

	// Get SDK session
	sess, err := getSession(cmd)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Delete group via SDK
	ctx := context.Background()
	if err := sess.DeleteGroup(ctx, groupID); err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	// Output success message
	output.OutputSuccess(cmd.Context(), fmt.Sprintf("Group %s deleted successfully", groupIDStr))
	return nil
}

// confirmDeletion prompts the user to confirm deletion of a group.
func confirmDeletion(groupID string) (bool, error) {
	fmt.Printf("Are you sure you want to delete group %s? This action cannot be undone. [y/N]: ", groupID)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}
