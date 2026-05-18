// Phase 4b §4.1 #13: zones port from CanvusPythonAPI/canvus_api/widget_operations.py:29-189.
//
// A Zone is a named rectangular region you compute client-side from a set
// of widgets. The Canvus server has no first-class "zone" entity; the
// helpers here let you treat groups of widgets as a logical zone for
// spatial queries (which widgets are inside, which widgets touch).
package extras

import (
	"errors"
	"fmt"
	"hash/fnv"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// SpatialTolerance carries fuzz factors for spatial widget operations.
// Mirrors the Python dataclass of the same name.
type SpatialTolerance struct {
	PositionTolerance  float64 // tolerance for position-based operations
	SizeTolerance      float64 // tolerance for size-based operations
	OverlapTolerance   float64 // minimum overlap for intersection operations
	DistanceTolerance  float64 // tolerance for distance-based operations
}

// DefaultSpatialTolerance returns the Python defaults (5 / 2 / 1 / 10).
func DefaultSpatialTolerance() SpatialTolerance {
	return SpatialTolerance{
		PositionTolerance: 5.0,
		SizeTolerance:     2.0,
		OverlapTolerance:  1.0,
		DistanceTolerance: 10.0,
	}
}

// WidgetZone is a named rectangular region computed from a set of widgets.
type WidgetZone struct {
	ID          string
	Name        string
	Description string
	Location    canvus.Point
	Size        canvus.Size
}

// Rectangle returns the zone's bounds as a canvus.Rectangle.
func (z WidgetZone) Rectangle() canvus.Rectangle {
	return canvus.Rectangle{X: z.Location.X, Y: z.Location.Y, Width: z.Size.Width, Height: z.Size.Height}
}

// WidgetZoneManager bundles zone-creation and zone-query helpers, sharing
// a SpatialTolerance config across calls.
type WidgetZoneManager struct {
	Tolerance SpatialTolerance
}

// NewWidgetZoneManager returns a zone manager with default tolerance.
func NewWidgetZoneManager() *WidgetZoneManager {
	return &WidgetZoneManager{Tolerance: DefaultSpatialTolerance()}
}

// CreateZoneFromWidgets builds a WidgetZone that encloses every widget in
// the slice, expanded by padding on every side. Returns an error if the
// slice is empty.
func (m *WidgetZoneManager) CreateZoneFromWidgets(widgets []canvus.Widget, name, description string, padding float64) (WidgetZone, error) {
	if len(widgets) == 0 {
		return WidgetZone{}, errors.New("CreateZoneFromWidgets: empty widget list")
	}
	bounds, err := calculateBounds(widgets)
	if err != nil {
		return WidgetZone{}, err
	}
	bounds.X -= padding
	bounds.Y -= padding
	bounds.Width += 2 * padding
	bounds.Height += 2 * padding
	return WidgetZone{
		ID:          fmt.Sprintf("zone_%d_%d", len(widgets), hashString(name)),
		Name:        name,
		Description: description,
		Location:    canvus.Point{X: bounds.X, Y: bounds.Y},
		Size:        canvus.Size{Width: bounds.Width, Height: bounds.Height},
	}, nil
}

// WidgetsInZone returns the subset of widgets fully contained by zone.
func (m *WidgetZoneManager) WidgetsInZone(widgets []canvus.Widget, zone WidgetZone) []canvus.Widget {
	zr := zone.Rectangle()
	out := make([]canvus.Widget, 0, len(widgets))
	for _, w := range widgets {
		if canvus.Contains(zr, canvus.WidgetBoundingBox(w)) {
			out = append(out, w)
		}
	}
	return out
}

// WidgetsTouchingZone returns the subset of widgets that touch or overlap zone.
func (m *WidgetZoneManager) WidgetsTouchingZone(widgets []canvus.Widget, zone WidgetZone) []canvus.Widget {
	zr := zone.Rectangle()
	out := make([]canvus.Widget, 0, len(widgets))
	for _, w := range widgets {
		if canvus.Touches(zr, canvus.WidgetBoundingBox(w)) {
			out = append(out, w)
		}
	}
	return out
}

// calculateBounds returns the smallest rectangle that contains every widget.
func calculateBounds(widgets []canvus.Widget) (canvus.Rectangle, error) {
	if len(widgets) == 0 {
		return canvus.Rectangle{}, errors.New("calculateBounds: empty widget list")
	}
	first := canvus.WidgetBoundingBox(widgets[0])
	minX, minY := first.X, first.Y
	maxX, maxY := first.X+first.Width, first.Y+first.Height
	for _, w := range widgets[1:] {
		r := canvus.WidgetBoundingBox(w)
		if r.X < minX {
			minX = r.X
		}
		if r.Y < minY {
			minY = r.Y
		}
		if r.X+r.Width > maxX {
			maxX = r.X + r.Width
		}
		if r.Y+r.Height > maxY {
			maxY = r.Y + r.Height
		}
	}
	return canvus.Rectangle{X: minX, Y: minY, Width: maxX - minX, Height: maxY - minY}, nil
}

func hashString(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
