// Package canvas provides commands for managing canvases in Canvus.
package canvas

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

// CanvasCmd is the parent command for all canvas-related operations.
var CanvasCmd = &cobra.Command{
	Use:   "canvas",
	Short: "Manage canvases",
	Long: `Manage canvases in your Canvus workspace.

Canvases are the primary containers for collaborative content in Canvus.
Each canvas can contain multiple widgets (notes, images, PDFs, videos, etc.)
and can be organized into folders.

Examples:
  # Create a new canvas
  canvus canvas create "My Canvas"

  # List all canvases
  canvus canvas list

  # Get canvas details
  canvus canvas get <canvas-id>

  # Update canvas name
  canvus canvas update <canvas-id> --name "New Name"

  # Delete a canvas
  canvus canvas delete <canvas-id>`,
}

// getSession retrieves the SDK session from the command context.
// This is a helper function used by all canvas commands.
func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}

// confirmAction prompts the user for confirmation of a destructive action.
// Returns true if the user confirms, false otherwise.
// If force is true, the confirmation is skipped and true is returned.
func confirmAction(message string, force bool) (bool, error) {
	if force {
		return true, nil
	}

	fmt.Printf("%s [y/N]: ", message)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

// validateCanvasID validates that a canvas ID is provided and not empty.
func validateCanvasID(id string) error {
	if id == "" {
		return fmt.Errorf("canvas ID is required")
	}
	return nil
}

// validateCanvasName validates that a canvas name is provided and not empty.
func validateCanvasName(name string) error {
	if name == "" {
		return fmt.Errorf("canvas name is required")
	}
	return nil
}
