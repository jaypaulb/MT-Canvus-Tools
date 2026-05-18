// Package connector provides commands for managing connector widgets.
package connector

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
)

// ConnectorCmd is the parent command for connector widget operations.
var ConnectorCmd = &cobra.Command{
	Use:   "connector",
	Short: "Manage connector widgets",
}

func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
