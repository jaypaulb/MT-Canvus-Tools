package canvas

import (
	"fmt"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export <canvas-id> <folder-path>",
	Short: "Export widgets to folder",
	Args:  cobra.ExactArgs(2),
	RunE:  runExport,
}

func init() {
	CanvasCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("export requires additional parameters - please use the SDK directly for this operation")
}
