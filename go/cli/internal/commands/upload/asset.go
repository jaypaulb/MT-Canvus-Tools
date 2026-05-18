package upload

import (
	"fmt"
	"github.com/spf13/cobra"
)

var assetCmd = &cobra.Command{
	Use:   "asset <canvas-id>",
	Short: "Upload an asset",
	Long: `Upload an asset to a canvas using multipart form data.

Note: This command requires proper multipart formatting which is complex.
Consider using the SDK directly for uploads.`,
	Args: cobra.ExactArgs(1),
	RunE: runAsset,
}

func init() {
	UploadCmd.AddCommand(assetCmd)
}

func runAsset(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("asset upload requires multipart form data - please use the SDK directly for this operation")
}
