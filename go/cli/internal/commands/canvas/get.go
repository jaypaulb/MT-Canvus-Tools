package canvas

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <canvas-id>",
	Short: "Get canvas details",
	Long: `Retrieve detailed information about a specific canvas.

This command displays all properties of a canvas including its ID, name,
folder, creation date, modification date, and other metadata.

Examples:
  # Get canvas details
  canvus canvas get canvas-123

  # Get canvas details as JSON
  canvus canvas get canvas-123 --output json

  # Get canvas details as YAML
  canvus canvas get canvas-123 --output yaml`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

func init() {
	CanvasCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	canvasID := args[0]

	// Validate canvas ID
	if err := validateCanvasID(canvasID); err != nil {
		return err
	}

	// Get SDK session from context
	sess, err := session.GetSession(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Get canvas using SDK
	ctx := context.Background()
	canvas, err := sess.GetCanvas(ctx, canvasID)
	if err != nil {
		return fmt.Errorf("failed to get canvas: %w", err)
	}

	// Format and output the result
	return output.OutputSingle(cmd.Context(), canvas, "")
}
