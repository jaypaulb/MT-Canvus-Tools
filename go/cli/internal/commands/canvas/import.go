package canvas

import (
	"fmt"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <canvas-id> <folder-path>",
	Short: "Import widgets from folder",
	Args:  cobra.ExactArgs(2),
	RunE:  runImport,
}

func init() {
	CanvasCmd.AddCommand(importCmd)
}

func runImport(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("import requires additional parameters - please use the SDK directly for this operation")
}
