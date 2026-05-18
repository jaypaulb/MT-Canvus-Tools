package canvas

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy <canvas-id>",
	Short: "Copy a canvas",
	Long: `Create a copy of an existing canvas.

You must specify a destination folder for the copied canvas.

Examples:
  # Copy canvas to a specific folder
  canvus canvas copy canvas-123 --folder folder-456

  # Copy canvas to root folder
  canvus canvas copy canvas-123 --folder ""`,
	Args: cobra.ExactArgs(1),
	RunE: runCopy,
}

var (
	copyFolderID string
)

func init() {
	copyCmd.Flags().StringVar(&copyFolderID, "folder", "", "Destination folder ID for the copy (required)")
	copyCmd.MarkFlagRequired("folder")
	CanvasCmd.AddCommand(copyCmd)
}

func runCopy(cmd *cobra.Command, args []string) error {
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

	// Copy canvas using SDK
	req := canvus.MoveOrCopyCanvasRequest{
		FolderID: copyFolderID,
	}

	ctx := context.Background()
	canvas, err := sess.CopyCanvas(ctx, canvasID, req)
	if err != nil {
		return fmt.Errorf("failed to copy canvas: %w", err)
	}

	// Format and output the result
	return output.OutputSingle(cmd.Context(), canvas, "")
}
