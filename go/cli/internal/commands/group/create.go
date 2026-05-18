package group

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new group",
	Long: `Create a new group in the Canvus workspace.

Requires administrative privileges. You must provide a name for the group.
Groups can be used to organize users and manage permissions.

Examples:
  # Create a new group
  canvus group create --name "Development Team"

  # Create a group with a descriptive name
  canvus group create --name "Marketing Department"`,
	RunE: runCreate,
}

var createName string

func init() {
	createCmd.Flags().StringVar(&createName, "name", "", "Name of the group (required)")
	createCmd.MarkFlagRequired("name")

	GroupCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Validate input
	if err := validateGroupName(createName); err != nil {
		return err
	}

	// Get SDK session
	sess, err := getSession(cmd)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Build group request
	req := canvus.CreateGroupRequest{
		Name: createName,
	}

	// Create group via SDK
	ctx := context.Background()
	group, err := sess.CreateGroup(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	// Output result
	return output.PrintOutput(cmd.Context(), group, "")
}
