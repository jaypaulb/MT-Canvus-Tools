// go/tools/powertoys/internal/atoms/webui/widget.go
package webui

import (
	"context"
	"fmt"
	"strings"

	canvus "github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// WidgetLocation is a type alias for canvus.Point.
// All existing code using WidgetLocation.X / .Y compiles unchanged.
type WidgetLocation = canvus.Point

// WidgetSize is a type alias for canvus.Size.
// All existing code using WidgetSize.Width / .Height compiles unchanged.
type WidgetSize = canvus.Size

// Widget represents a canvas widget for local processing.
type Widget struct {
	ID         string          `json:"id"`
	WidgetType string          `json:"widget_type"`
	Location   *WidgetLocation `json:"location,omitempty"`
	Size       *WidgetSize     `json:"size,omitempty"`
	Scale      float64         `json:"scale,omitempty"`
	Pinned     bool            `json:"pinned,omitempty"`
	Title      string          `json:"title,omitempty"`
	Color      string          `json:"color,omitempty"`
}

// GetAllWidgets fetches all widgets for canvasID via the SDK.
func GetAllWidgets(ctx context.Context, apiClient *APIClient, canvasID string) ([]Widget, error) {
	if canvasID == "" {
		return nil, fmt.Errorf("GetAllWidgets: canvas ID is required")
	}
	sdkWidgets, err := apiClient.session.ListWidgets(ctx, canvasID, nil)
	if err != nil {
		return nil, fmt.Errorf("GetAllWidgets: %w", err)
	}
	out := make([]Widget, 0, len(sdkWidgets))
	for _, w := range sdkWidgets {
		out = append(out, Widget{
			ID:         w.ID,
			WidgetType: w.WidgetType,
			Location:   w.Location,
			Size:       w.Size,
			Scale:      w.Scale,
			Pinned:     w.Pinned,
		})
	}
	return out, nil
}

// TransformWidgetLocationAndScale transforms a widget's location and scale
// from the source bounding box into the target's.
func TransformWidgetLocationAndScale(widget *Widget, sourceBB, targetBB *ZoneBoundingBox) {
	if widget.Location == nil {
		return
	}
	scaleFactor := targetBB.Width / sourceBB.Width
	deltaX := widget.Location.X - sourceBB.X
	deltaY := widget.Location.Y - sourceBB.Y
	widget.Location.X = targetBB.X + deltaX*scaleFactor
	widget.Location.Y = targetBB.Y + deltaY*scaleFactor
	oldScale := widget.Scale
	if oldScale == 0 {
		oldScale = 1
	}
	widget.Scale = oldScale * scaleFactor
}

// GetWidgetPatchEndpoint returns the API path segment for a given widget_type.
func GetWidgetPatchEndpoint(widgetType string) string {
	switch strings.ToLower(widgetType) {
	case "note":
		return "/notes"
	case "browser":
		return "/browsers"
	case "image":
		return "/images"
	case "pdf":
		return "/pdfs"
	case "video":
		return "/videos"
	case "connector":
		return "/connectors"
	case "anchor":
		return "/anchors"
	default:
		return "/notes"
	}
}
