package browser

import (
	"github.com/spf13/cobra"
)

var BrowserCmd = &cobra.Command{
	Use:   "browser",
	Short: "Manage browser widgets",
	Long: `Manage browser widgets in your Canvus canvases.

Browser widgets display web pages within the canvas.

Examples:
  # Create a browser widget
  canvus widget browser create <canvas-id> --url "https://example.com" --x 100 --y 200

  # List browser widgets
  canvus widget browser list <canvas-id>`,
}
