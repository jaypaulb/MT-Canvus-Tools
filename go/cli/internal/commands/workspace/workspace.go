// Package workspace provides commands for managing workspaces in Canvus.
package workspace

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
)

// WorkspaceCmd is the parent command for all workspace-related operations.
var WorkspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage workspaces",
	Long: `Manage workspaces in your Canvus system.

Workspaces represent viewing sessions on client devices.

Examples:
  # List workspaces for a client
  canvus workspace list <client-id>

  # Get workspace details
  canvus workspace get <client-id> --index 0

  # Update workspace
  canvus workspace update <client-id> --index 0 --canvas-id canvas-123

  # Open canvas on workspace
  canvus workspace open <client-id> --index 0 --canvas-id canvas-456`,
}

// getSession retrieves the SDK session from the command context.
func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
