// Package colorpreset provides commands for managing color presets.
package colorpreset

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
)

// ColorPresetCmd is the parent command for color preset operations.
var ColorPresetCmd = &cobra.Command{
	Use:   "colorpreset",
	Short: "Manage color presets",
}

func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
