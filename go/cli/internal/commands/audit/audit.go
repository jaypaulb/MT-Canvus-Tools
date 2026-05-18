// Package audit provides commands for accessing audit logs.
package audit

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

// AuditCmd is the parent command for audit log operations.
var AuditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Access audit logs",
}

func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
