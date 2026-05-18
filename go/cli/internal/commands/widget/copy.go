package widget

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

// copyCmd performs a cross-canvas widget clone via the SDK's CloneWidget
// helper, which POSTs to the destination canvas's type-specific endpoint
// (see SDK widgets.go:CloneWidget).
//
// Behaviour note: the legacy CLI accepted (widget-id, target-canvas-id) and
// relied on a /widgets/{id}/copy endpoint that does not exist in the real
// Canvus API. The new signature requires the source canvas-id (so the SDK
// can look up the widget's type) and uses the documented clone endpoint.
var copyCmd = &cobra.Command{
	Use:   "copy <source-canvas-id> <widget-id> <target-canvas-id>",
	Short: "Clone a widget into another canvas",
	Args:  cobra.ExactArgs(3),
	RunE:  runCopy,
}

func init() {
	WidgetCmd.AddCommand(copyCmd)
}

func runCopy(cmd *cobra.Command, args []string) error {
	sourceCanvasID, widgetID, targetCanvasID := args[0], args[1], args[2]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	ctx := context.Background()
	src, err := sess.GetWidget(ctx, sourceCanvasID, widgetID)
	if err != nil {
		return fmt.Errorf("look up source widget: %w", err)
	}
	wtPath, err := widgetTypePath(src.WidgetType)
	if err != nil {
		return err
	}

	cloned, err := sess.CloneWidget(ctx, targetCanvasID, sourceCanvasID, widgetID, wtPath, nil)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), cloned, "Widget cloned successfully")
}

// widgetTypePath maps a widget's `widget_type` field to the path segment used
// by the type-specific create/clone endpoint (see SDK widgets.go:CloneWidget).
func widgetTypePath(widgetType string) (string, error) {
	switch widgetType {
	case "note", "Note":
		return "notes", nil
	case "image", "Image":
		return "images", nil
	case "video", "Video":
		return "videos", nil
	case "pdf", "PDF":
		return "pdfs", nil
	case "browser", "Browser":
		return "browsers", nil
	case "anchor", "Anchor":
		return "anchors", nil
	case "table", "Table":
		return "tables", nil
	case "connector", "Connector":
		return "connectors", nil
	}
	return "", fmt.Errorf("unsupported widget type for clone: %q", widgetType)
}
