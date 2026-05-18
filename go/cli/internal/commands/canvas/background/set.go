package background

import (
	"fmt"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <canvas-id>",
	Short: "Set canvas background (image)",
	Long:  "Set canvas background to an image. Requires multipart data.",
	Args:  cobra.ExactArgs(1),
	RunE:  runSet,
}

func init() {
	BackgroundCmd.AddCommand(setCmd)
}

func runSet(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("setting background image requires multipart form data - please use the SDK directly")
}
