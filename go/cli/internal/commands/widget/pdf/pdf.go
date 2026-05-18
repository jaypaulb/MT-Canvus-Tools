// Package pdf provides commands for managing pdf widgets.
package pdf

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

// PDFCmd is the parent command for pdf widget operations.
var PDFCmd = &cobra.Command{
	Use:   "pdf",
	Short: "Manage pdf widgets",
}

func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
