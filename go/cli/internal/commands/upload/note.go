package upload

import (
	"fmt"
	"github.com/spf13/cobra"
)

var noteCmd = &cobra.Command{
	Use:   "note <canvas-id>",
	Short: "Upload a note",
	Long: `Upload a note to a canvas using multipart form data.

Note: This command requires proper multipart formatting which is complex.
Consider using the SDK directly for uploads.`,
	Args: cobra.ExactArgs(1),
	RunE: runNote,
}

func init() {
	UploadCmd.AddCommand(noteCmd)
}

func runNote(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("note upload requires multipart form data - please use the SDK directly for this operation")
}
