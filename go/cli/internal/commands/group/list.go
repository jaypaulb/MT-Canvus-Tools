package group

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all groups",
	Long: `List all groups in the Canvus workspace.

Requires administrative privileges. Returns a list of all groups
with their details.

Examples:
  # List all groups in table format
  canvus group list

  # List all groups in JSON format
  canvus group list --output json`,
	RunE: runList,
}

func init() {
	GroupCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	// Get SDK session
	sess, err := getSession(cmd)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// List groups via SDK
	ctx := context.Background()
	groups, err := sess.ListGroups(ctx)
	if err != nil {
		return fmt.Errorf("failed to list groups: %w", err)
	}

	// Output results
	return output.PrintOutput(cmd.Context(), groups, "")
}
