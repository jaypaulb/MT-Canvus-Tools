// Package video provides commands for managing video widgets.
package video

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

// VideoCmd is the parent command for video widget operations.
var VideoCmd = &cobra.Command{
	Use:   "video",
	Short: "Manage video widgets",
}

func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
