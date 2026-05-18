// Package upload provides commands for uploading assets in Canvus.
package upload

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

// UploadCmd is the parent command for all upload-related operations.
var UploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload assets",
	Long: `Upload notes and assets to canvases in Canvus.

Note: These commands require multipart form data.

Examples:
  # Upload a note (requires multipart data)
  canvus upload note <canvas-id> --file note.txt

  # Upload an asset (requires multipart data)
  canvus upload asset <canvas-id> --file image.png`,
}

// getSession retrieves the SDK session from the command context.
func getSession(cmd *cobra.Command) (*canvus.Session, error) {
	return session.GetSession(cmd.Context())
}
