package canvus

import (
	"fmt"
	"math"
	"strings"
)

// GeometryModel selects an explicit registration model for a canvas snapshot.
// CanvasID is required; ancestry must terminate at that explicit root. Missing
// parent IDs are unknown metadata, not permission to assume a root transform.
// NotePadding is required for Notes: 30 reproduces the documented legacy
// padding/border model, 0 selects normalized model origins. Do not guess the
// deployed model or silently rewrite the stored location. Non-Notes use zero
// registration padding. Only unrotated, uniform positive scales are supported.
type GeometryModel struct {
	CanvasID    string
	NotePadding *float64
}

func finite(values ...float64) bool {
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	return true
}

func positiveSize(s Size) bool { return finite(s.Width, s.Height) && s.Width > 0 && s.Height > 0 }
func positiveRect(r Rectangle) bool {
	return finite(r.X, r.Y, r.Width, r.Height, r.X+r.Width, r.Y+r.Height) && r.Width > 0 && r.Height > 0
}

// VisibleCanvasRegion decodes the established scaled view_rectangle contract:
// scale=raw.width/workspace.width, origin=-raw.position/scale, size=workspace/scale.
// Reference: CanvusConsole computeViewportMetrics and canvus-stress ViewRectFor.
func VisibleCanvasRegion(raw Rectangle, workspace Size) (Rectangle, error) {
	if !positiveRect(raw) || !positiveSize(workspace) {
		return Rectangle{}, fmt.Errorf("VisibleCanvasRegion: %w", ErrInvalidGeometry)
	}
	scale := raw.Width / workspace.Width
	if !finite(scale) || scale <= 0 || math.Abs(raw.Height-workspace.Height*scale) > 1e-5*math.Max(1, math.Abs(raw.Height)) {
		return Rectangle{}, fmt.Errorf("VisibleCanvasRegion: %w: inconsistent scale/aspect", ErrInvalidGeometry)
	}
	visible := Rectangle{X: -raw.X / scale, Y: -raw.Y / scale, Width: workspace.Width / scale, Height: workspace.Height / scale}
	if !positiveRect(visible) {
		return Rectangle{}, fmt.Errorf("VisibleCanvasRegion: %w: overflow", ErrInvalidGeometry)
	}
	return visible, nil
}

// ViewRectangleForRegion encodes a canvas-pixel region using the established
// best-fit/centred camera transform. Apply zoom first, then pan; a single combined
// PATCH is not equivalent on affected clients. No fallback workspace size is used.
func ViewRectangleForRegion(region Rectangle, workspace Size) (Rectangle, error) {
	if !positiveRect(region) || !positiveSize(workspace) {
		return Rectangle{}, fmt.Errorf("ViewRectangleForRegion: %w", ErrInvalidGeometry)
	}
	scale := math.Min(workspace.Width/region.Width, workspace.Height/region.Height)
	x := region.X + region.Width/2 - workspace.Width/scale/2
	y := region.Y + region.Height/2 - workspace.Height/scale/2
	raw := Rectangle{X: -x * scale, Y: -y * scale, Width: workspace.Width * scale, Height: workspace.Height * scale}
	if !positiveRect(raw) {
		return Rectangle{}, fmt.Errorf("ViewRectangleForRegion: %w: overflow", ErrInvalidGeometry)
	}
	return raw, nil
}

// WidgetCanvasBounds returns rendered border-box bounds in canvas pixels using
// a complete ancestry snapshot. It incorporates parent scale and the explicit
// Note registration model (issue #47), rejecting cycles/missing transforms.
// Sizes must be current server/client observations, not assumed final write echoes:
// native text fitting can change them asynchronously. Legacy WidgetBoundingBox
// remains a raw model-space helper and is not interchangeable with this function.
func WidgetCanvasBounds(id string, widgets []Widget, model GeometryModel) (Rectangle, error) {
	if model.CanvasID == "" {
		return Rectangle{}, fmt.Errorf("WidgetCanvasBounds: %w: canvas root required", ErrGeometryModelRequired)
	}
	if id == "" || id == model.CanvasID {
		return Rectangle{}, fmt.Errorf("WidgetCanvasBounds: %w: widget ID required", ErrInvalidGeometry)
	}
	if model.NotePadding != nil && (!finite(*model.NotePadding) || *model.NotePadding < 0) {
		return Rectangle{}, fmt.Errorf("WidgetCanvasBounds: %w: padding", ErrInvalidGeometry)
	}
	byID := make(map[string]Widget, len(widgets))
	for _, w := range widgets {
		if w.ID == "" {
			return Rectangle{}, fmt.Errorf("WidgetCanvasBounds: %w: missing ID", ErrInvalidGeometry)
		}
		if _, exists := byID[w.ID]; exists {
			return Rectangle{}, fmt.Errorf("WidgetCanvasBounds: %w: duplicate ID", ErrInvalidGeometry)
		}
		byID[w.ID] = w
	}
	var chain []Widget
	seen := map[string]bool{}
	for current := id; current != "" && current != model.CanvasID; {
		if seen[current] {
			return Rectangle{}, fmt.Errorf("WidgetCanvasBounds: %w: parent cycle", ErrInvalidGeometry)
		}
		seen[current] = true
		w, ok := byID[current]
		if !ok {
			return Rectangle{}, fmt.Errorf("WidgetCanvasBounds: %w: missing ancestor", ErrInvalidGeometry)
		}
		if w.Location == nil || !finite(w.Location.X, w.Location.Y, w.Scale) || w.Scale <= 0 || w.WidgetType == "" || w.ParentID == "" {
			return Rectangle{}, fmt.Errorf("WidgetCanvasBounds: %w: missing/invalid transform", ErrInvalidGeometry)
		}
		chain = append(chain, w)
		current = w.ParentID
	}
	target := chain[0]
	if target.Size == nil || !positiveSize(*target.Size) {
		return Rectangle{}, fmt.Errorf("WidgetCanvasBounds: %w: size", ErrInvalidGeometry)
	}
	x, y, scale, padding := 0.0, 0.0, 1.0, 0.0
	for i := len(chain) - 1; i >= 0; i-- {
		w := chain[i]
		padding = 0
		if strings.EqualFold(w.WidgetType, "Note") {
			if model.NotePadding == nil {
				return Rectangle{}, fmt.Errorf("WidgetCanvasBounds: %w", ErrGeometryModelRequired)
			}
			padding = *model.NotePadding
		}
		x += scale * (w.Location.X + padding)
		y += scale * (w.Location.Y + padding)
		scale *= w.Scale
		if !finite(x, y, scale) || scale <= 0 {
			return Rectangle{}, fmt.Errorf("WidgetCanvasBounds: %w: transform overflow", ErrInvalidGeometry)
		}
	}
	bounds := Rectangle{X: x - scale*padding, Y: y - scale*padding, Width: target.Size.Width * scale, Height: target.Size.Height * scale}
	if !positiveRect(bounds) {
		return Rectangle{}, fmt.Errorf("WidgetCanvasBounds: %w: bounds overflow", ErrInvalidGeometry)
	}
	return bounds, nil
}
