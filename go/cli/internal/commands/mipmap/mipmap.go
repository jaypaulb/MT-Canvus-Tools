// Package mipmap provides commands for accessing mipmap data in Canvus.
package mipmap

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

// MipmapCmd is the parent command for all mipmap-related operations.
var MipmapCmd = &cobra.Command{
	Use:   "mipmap",
	Short: "Access mipmap data",
	Long: `Access mipmap data for assets in Canvus.

Mipmaps are pre-processed multi-resolution image representations.

Examples:
  # Get mipmap info
  canvus mipmap info --canvas-id canvas-123 --hash abc123

  # Get mipmap level
  canvus mipmap level --canvas-id canvas-123 --hash abc123 --level 0

  # Get asset by hash
  canvus mipmap asset --canvas-id canvas-123 --hash abc123`,
}

// getSession retrieves the SDK session from the command context.
func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
