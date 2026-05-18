// Package note provides commands for managing note widgets.
package note

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

// NoteCmd is the parent command for note widget operations.
var NoteCmd = &cobra.Command{
	Use:   "note",
	Short: "Manage note widgets",
}

func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
