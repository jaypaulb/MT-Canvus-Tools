// Package videoinput provides commands for managing video inputs.
package videoinput

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

// VideoInputCmd is the parent command for video input operations.
var VideoInputCmd = &cobra.Command{
	Use:   "videoinput",
	Short: "Manage video inputs",
}

func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
