package browser

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

var (
	createURL    string
	createTitle  string
	createX      float64
	createY      float64
	createWidth  float64
	createHeight float64
)

var createCmd = &cobra.Command{
	Use:   "create <canvas-id>",
	Short: "Create a browser widget",
	Long: `Create a browser widget that displays a web page.

Examples:
  # Create a browser showing a website
  canvus widget browser create <canvas-id> --url "https://miro.com" --title "Miro" --x 1000 --y 1000 --width 800 --height 600`,
	Args: cobra.ExactArgs(1),
	RunE: runCreate,
}

func init() {
	createCmd.Flags().StringVar(&createURL, "url", "", "URL to display (required)")
	createCmd.Flags().StringVar(&createTitle, "title", "", "Browser title")
	createCmd.Flags().Float64Var(&createX, "x", 0, "X coordinate")
	createCmd.Flags().Float64Var(&createY, "y", 0, "Y coordinate")
	createCmd.Flags().Float64Var(&createWidth, "width", 800, "Width")
	createCmd.Flags().Float64Var(&createHeight, "height", 600, "Height")
	createCmd.MarkFlagRequired("url")
	BrowserCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	canvasID := args[0]

	sess, err := session.GetSession(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	req := map[string]interface{}{
		"url": createURL,
	}

	if createTitle != "" {
		req["title"] = createTitle
	}

	if createX != 0 || createY != 0 {
		req["location"] = map[string]interface{}{
			"x": createX,
			"y": createY,
		}
	}

	if createWidth != 0 || createHeight != 0 {
		req["size"] = map[string]interface{}{
			"width":  createWidth,
			"height": createHeight,
		}
	}

	ctx := context.Background()
	browser, err := sess.CreateBrowser(ctx, canvasID, req)
	if err != nil {
		return fmt.Errorf("failed to create browser: %w", err)
	}

	return output.OutputSingle(cmd.Context(), browser, "")
}
