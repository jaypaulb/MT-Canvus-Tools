// Package videooutput provides commands for managing video outputs.
package videooutput

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

// VideoOutputCmd is the parent command for video output operations.
var VideoOutputCmd = &cobra.Command{
	Use:   "videooutput",
	Short: "Manage video outputs",
}

func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
