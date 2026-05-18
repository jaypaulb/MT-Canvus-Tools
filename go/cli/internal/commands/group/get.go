package group

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <group-id>",
	Short: "Get group details",
	Long: `Get detailed information about a specific group.

Requires administrative privileges. Retrieves and displays all information
about the specified group, including its members.

Examples:
  # Get group details
  canvus group get 123

  # Get group details in JSON format
  canvus group get 123 --output json`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

func init() {
	GroupCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	groupIDStr := args[0]

	// Parse group ID as int
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		return fmt.Errorf("invalid group ID %q: must be a number", groupIDStr)
	}

	// Get SDK session
	sess, err := getSession(cmd)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Get group via SDK
	ctx := context.Background()
	group, err := sess.GetGroup(ctx, groupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %w", err)
	}

	// Output result
	return output.PrintOutput(cmd.Context(), group, "")
}
