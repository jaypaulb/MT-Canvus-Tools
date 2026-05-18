package canvas

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move <canvas-id>",
	Short: "Move a canvas to a different folder",
	Long: `Move a canvas to a different folder in your Canvus workspace.

The --folder flag is required to specify the destination folder.

Examples:
  # Move canvas to a specific folder
  canvus canvas move canvas-123 --folder folder-456

  # Move canvas to root (no folder)
  canvus canvas move canvas-123 --folder ""`,
	Args: cobra.ExactArgs(1),
	RunE: runMove,
}

var (
	moveFolderID string
)

func init() {
	moveCmd.Flags().StringVar(&moveFolderID, "folder", "", "Destination folder ID (required)")
	moveCmd.MarkFlagRequired("folder")
	CanvasCmd.AddCommand(moveCmd)
}

func runMove(cmd *cobra.Command, args []string) error {
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

	// Move canvas using SDK MoveCanvas method
	req := canvus.MoveOrCopyCanvasRequest{
		FolderID: moveFolderID,
	}

	ctx := context.Background()
	canvas, err := sess.MoveCanvas(ctx, canvasID, req)
	if err != nil {
		return fmt.Errorf("failed to move canvas: %w", err)
	}

	// Format and output the result
	return output.OutputSingle(cmd.Context(), canvas, "")
}
