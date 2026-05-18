// Package image provides commands for managing image widgets.
package image

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

// ImageCmd is the parent command for image widget operations.
var ImageCmd = &cobra.Command{
	Use:   "image",
	Short: "Manage image widgets",
}

func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
