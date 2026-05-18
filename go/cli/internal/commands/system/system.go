// Package system provides commands for system information and administration.
package system

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
)

// SystemCmd is the parent command for all system-related operations.
var SystemCmd = &cobra.Command{
	Use:   "system",
	Short: "System information and administration",
	Long: `System information and administration commands.

These commands provide access to system-level information such as
server version, API version, connectivity status, and license details.

Examples:
  # Get system information
  canvus system info

  # Get license information
  canvus system license`,
}

// getSession retrieves the SDK session from the command context.
// This is a helper function used by all system commands.
func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
