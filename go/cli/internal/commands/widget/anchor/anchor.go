// Package anchor provides commands for managing anchor widgets.
package anchor

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

// AnchorCmd is the parent command for anchor widget operations.
var AnchorCmd = &cobra.Command{
	Use:   "anchor",
	Short: "Manage anchor widgets",
}

func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
