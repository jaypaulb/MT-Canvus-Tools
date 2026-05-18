// Package folder provides commands for managing folders in Canvus.
package folder

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

// FolderCmd is the parent command for all folder-related operations.
var FolderCmd = &cobra.Command{
	Use:   "folder",
	Short: "Manage folders",
	Long: `Manage canvas folders in your Canvus system.

Folders organize canvases into a hierarchical structure.

Examples:
  # List all folders
  canvus folder list

  # Create a new folder
  canvus folder create --name "My Folder"

  # Get folder details
  canvus folder get <folder-id>

  # Move a folder
  canvus folder move <folder-id> <parent-folder-id>

  # Delete a folder
  canvus folder delete <folder-id>`,
}

// getSession retrieves the SDK session from the command context.
func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
