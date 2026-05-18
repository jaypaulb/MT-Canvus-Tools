// Package background provides commands for managing canvas backgrounds.
package background

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

// BackgroundCmd is the parent command for canvas background operations.
var BackgroundCmd = &cobra.Command{
	Use:   "background",
	Short: "Manage canvas backgrounds",
}

// getSession retrieves the SDK session from the command context.
func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
