// Phase 4b §4.1 #14: batch widget ops port from
// CanvusPythonAPI/canvus_api/widget_operations.py:191-388.
//
// These helpers compute update payloads for batches of widgets. They do
// not perform any network IO — orchestration is left to the caller (often
// via canvus.BatchOperationBuilder in the core SDK).
package extras

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// WidgetOperation describes a single batch update. Operation is a free-text
// tag for the caller's bookkeeping (e.g. "move", "resize"); Payload is the
// patch body to send to the per-type PATCH endpoint.
type WidgetOperation struct {
	WidgetID  string
	Operation string
	Payload   map[string]any
}

// BatchWidgetOperations bundles widget-batch helpers behind a SpatialTolerance.
type BatchWidgetOperations struct {
	Tolerance SpatialTolerance
}

// NewBatchWidgetOperations returns a batch helper with default tolerance.
func NewBatchWidgetOperations() *BatchWidgetOperations {
	return &BatchWidgetOperations{Tolerance: DefaultSpatialTolerance()}
}

// MoveWidgets returns a move-operation per widget that translates each
// widget by (offsetX, offsetY). Widgets without a Location are skipped.
func (b *BatchWidgetOperations) MoveWidgets(widgets []canvus.Widget, offsetX, offsetY float64) []WidgetOperation {
	ops := make([]WidgetOperation, 0, len(widgets))
	for _, w := range widgets {
		if w.Location == nil {
			continue
		}
		ops = append(ops, WidgetOperation{
			WidgetID:  w.ID,
			Operation: "move",
			Payload: map[string]any{
				"location": map[string]any{
					"x": w.Location.X + offsetX,
					"y": w.Location.Y + offsetY,
				},
			},
		})
	}
	return ops
}

// ResizeWidgets returns a resize-operation per widget that scales Size by
// scaleFactor. Widgets without a Size are skipped.
func (b *BatchWidgetOperations) ResizeWidgets(widgets []canvus.Widget, scaleFactor float64) []WidgetOperation {
	ops := make([]WidgetOperation, 0, len(widgets))
	for _, w := range widgets {
		if w.Size == nil {
			continue
		}
		ops = append(ops, WidgetOperation{
			WidgetID:  w.ID,
			Operation: "resize",
			Payload: map[string]any{
				"size": map[string]any{
					"width":  w.Size.Width * scaleFactor,
					"height": w.Size.Height * scaleFactor,
				},
			},
		})
	}
	return ops
}

// WidgetsContainID returns the widgets that fully contain the widget with
// ID == targetID (using each widget's bounding box). The target widget
// itself is excluded.
func (b *BatchWidgetOperations) WidgetsContainID(widgets []canvus.Widget, targetID string) []canvus.Widget {
	var target *canvus.Widget
	for i := range widgets {
		if widgets[i].ID == targetID {
			target = &widgets[i]
			break
		}
	}
	if target == nil {
		return nil
	}
	out := make([]canvus.Widget, 0)
	for _, w := range widgets {
		if w.ID == targetID {
			continue
		}
		if canvus.WidgetContains(w, *target) {
			out = append(out, w)
		}
	}
	return out
}

// WidgetsTouchID returns the widgets whose bounding boxes touch or overlap
// the widget with ID == targetID. The target widget itself is excluded.
func (b *BatchWidgetOperations) WidgetsTouchID(widgets []canvus.Widget, targetID string) []canvus.Widget {
	var target *canvus.Widget
	for i := range widgets {
		if widgets[i].ID == targetID {
			target = &widgets[i]
			break
		}
	}
	if target == nil {
		return nil
	}
	out := make([]canvus.Widget, 0)
	for _, w := range widgets {
		if w.ID == targetID {
			continue
		}
		if canvus.WidgetsTouch(w, *target) {
			out = append(out, w)
		}
	}
	return out
}
