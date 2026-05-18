package canvas

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <canvas-id>",
	Short: "Delete a canvas",
	Long: `Delete a canvas from your Canvus workspace.

By default, this command will prompt for confirmation before deleting.
Use the --force flag to skip the confirmation prompt.

Examples:
  # Delete a canvas (with confirmation)
  canvus canvas delete canvas-123

  # Delete a canvas without confirmation
  canvus canvas delete canvas-123 --force

  # Permanently delete a canvas
  canvus canvas delete canvas-123 --permanent --force`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

var (
	deleteForce     bool
	deletePermanent bool
)

func init() {
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "Skip confirmation prompt")
	deleteCmd.Flags().BoolVar(&deletePermanent, "permanent", false, "Permanently delete the canvas")
	CanvasCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	canvasID := args[0]

	// Validate canvas ID
	if err := validateCanvasID(canvasID); err != nil {
		return err
	}

	// Confirm deletion unless --force is provided
	message := fmt.Sprintf("Are you sure you want to delete canvas %s?", canvasID)
	if deletePermanent {
		message = fmt.Sprintf("Are you sure you want to PERMANENTLY delete canvas %s? This cannot be undone!", canvasID)
	}

	confirmed, err := confirmAction(message, deleteForce)
	if err != nil {
		return fmt.Errorf("failed to get confirmation: %w", err)
	}

	if !confirmed {
		return fmt.Errorf("delete operation cancelled")
	}

	// Get SDK session from context
	sess, err := session.GetSession(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Delete canvas using SDK
	ctx := context.Background()
	err = sess.DeleteCanvas(ctx, canvasID)
	if err != nil {
		return fmt.Errorf("failed to delete canvas: %w", err)
	}

	// Output success message
	successMsg := fmt.Sprintf("Successfully deleted canvas %s", canvasID)
	output.OutputSuccess(cmd.Context(), successMsg)
	return nil
}
