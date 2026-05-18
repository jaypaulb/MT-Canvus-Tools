package canvas

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <canvas-id>",
	Short: "Update canvas properties",
	Long: `Update properties of an existing canvas.

You can update the name or mode of the canvas.
Only the properties you specify will be updated.

Examples:
  # Update canvas name
  canvus canvas update canvas-123 --name "New Canvas Name"

  # Update canvas mode
  canvus canvas update canvas-123 --mode "edit"`,
	Args: cobra.ExactArgs(1),
	RunE: runUpdate,
}

var (
	updateName string
	updateMode string
)

func init() {
	updateCmd.Flags().StringVar(&updateName, "name", "", "New name for the canvas")
	updateCmd.Flags().StringVar(&updateMode, "mode", "", "New mode for the canvas")
	CanvasCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	canvasID := args[0]

	// Validate canvas ID
	if err := validateCanvasID(canvasID); err != nil {
		return err
	}

	// Build update request with only provided fields
	req := canvus.UpdateCanvasRequest{}
	hasUpdates := false

	if updateName != "" {
		req.Name = updateName
		hasUpdates = true
	}
	if updateMode != "" {
		req.Mode = updateMode
		hasUpdates = true
	}

	// Check if any fields were provided
	if !hasUpdates {
		return fmt.Errorf("at least one field must be provided for update (--name, --mode)")
	}

	// Get SDK session from context
	sess, err := session.GetSession(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Update canvas using SDK
	ctx := context.Background()
	canvas, err := sess.UpdateCanvas(ctx, canvasID, req)
	if err != nil {
		return fmt.Errorf("failed to update canvas: %w", err)
	}

	// Format and output the result
	return output.OutputSingle(cmd.Context(), canvas, "")
}
