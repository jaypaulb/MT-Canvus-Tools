package canvas

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new canvas",
	Long: `Create a new canvas in your Canvus workspace.

A canvas is a collaborative workspace that can contain multiple widgets
such as notes, images, PDFs, videos, and more.

Examples:
  # Create a canvas with a name
  canvus canvas create "My New Canvas"

  # Create a canvas in a specific folder
  canvus canvas create "Project Planning" --folder folder-123

  # Create a canvas with additional options
  canvus canvas create "Team Workspace" --folder folder-456`,
	Args: cobra.ExactArgs(1),
	RunE: runCreate,
}

var (
	createFolderID string
)

func init() {
	createCmd.Flags().StringVar(&createFolderID, "folder", "", "Folder ID to create the canvas in")
	CanvasCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Validate canvas name
	if err := validateCanvasName(name); err != nil {
		return err
	}

	// Get SDK session provider from context (supports both real and mock sessions)
	sess, err := session.GetSessionProvider(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Build canvas creation request
	req := canvus.CreateCanvasRequest{
		Name: name,
	}

	// Add optional folder if provided
	if createFolderID != "" {
		req.FolderID = createFolderID
	}

	// Create canvas using SDK
	ctx := context.Background()
	canvas, err := sess.CreateCanvas(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create canvas: %w", err)
	}

	// Format and output the result
	return output.PrintOutput(cmd.Context(), canvas, "")
}
