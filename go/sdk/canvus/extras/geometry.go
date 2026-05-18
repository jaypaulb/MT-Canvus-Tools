// Phase 4b §4.1 #11: geometry port from CanvusPythonAPI/canvus_api/geometry.py.
//
// The core SDK already ships a small set of geometry primitives (Contains,
// Touches, WidgetBoundingBox, WidgetContains, WidgetsTouch — see
// canvus/geometry.go). This package extends that surface with the remaining
// helpers from the legacy Python module: Intersects (strict overlap), union /
// intersection rectangles, distance between widgets, area / point lookups,
// canvas-bounds. All helpers operate on canvus.Widget and canvus.Rectangle
// so callers can mix-and-match with the core SDK types.
package extras

import (
	"math"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// Intersects reports whether two rectangles have a strictly overlapping
// area (touching edges alone do not count, unlike canvus.Touches).
func Intersects(a, b canvus.Rectangle) bool {
	return !(a.X+a.Width <= b.X ||
		a.X >= b.X+b.Width ||
		a.Y+a.Height <= b.Y ||
		a.Y >= b.Y+b.Height)
}

// GetIntersection returns the overlap rectangle of a and b, or nil if they
// do not overlap.
func GetIntersection(a, b canvus.Rectangle) *canvus.Rectangle {
	if !Intersects(a, b) {
		return nil
	}
	left := math.Max(a.X, b.X)
	top := math.Max(a.Y, b.Y)
	right := math.Min(a.X+a.Width, b.X+b.Width)
	bottom := math.Min(a.Y+a.Height, b.Y+b.Height)
	return &canvus.Rectangle{X: left, Y: top, Width: right - left, Height: bottom - top}
}

// GetUnion returns the smallest rectangle that contains both a and b.
func GetUnion(a, b canvus.Rectangle) canvus.Rectangle {
	left := math.Min(a.X, b.X)
	top := math.Min(a.Y, b.Y)
	right := math.Max(a.X+a.Width, b.X+b.Width)
	bottom := math.Max(a.Y+a.Height, b.Y+b.Height)
	return canvus.Rectangle{X: left, Y: top, Width: right - left, Height: bottom - top}
}

// WidgetsIntersect reports whether two widgets' bounding boxes overlap.
func WidgetsIntersect(a, b canvus.Widget) bool {
	return Intersects(canvus.WidgetBoundingBox(a), canvus.WidgetBoundingBox(b))
}

// GetWidgetIntersection returns the overlap rectangle of two widgets'
// bounding boxes, or nil if they do not overlap.
func GetWidgetIntersection(a, b canvus.Widget) *canvus.Rectangle {
	return GetIntersection(canvus.WidgetBoundingBox(a), canvus.WidgetBoundingBox(b))
}

// GetWidgetUnion returns the smallest rectangle that contains both widgets.
func GetWidgetUnion(a, b canvus.Widget) canvus.Rectangle {
	return GetUnion(canvus.WidgetBoundingBox(a), canvus.WidgetBoundingBox(b))
}

// DistanceBetweenWidgets returns the cartesian gap distance between two
// widgets' bounding rectangles (0 if they overlap).
//
// When the rectangles are separated on both axes, returns the euclidean
// distance between the nearest corners (math.Hypot(dx, dy)). When
// separated on a single axis only, returns that axis's gap directly.
// Symmetric across Go / Python / TS SDK extras packages.
//
// Note: this deliberately differs from the legacy CanvusPythonAPI helper,
// which returned min(dx, dy) in the both-positive case — the legacy answer
// was a single-axis projection, not a cartesian distance.
func DistanceBetweenWidgets(a, b canvus.Widget) float64 {
	r1 := canvus.WidgetBoundingBox(a)
	r2 := canvus.WidgetBoundingBox(b)
	if Intersects(r1, r2) {
		return 0
	}
	dx := math.Max(0, math.Max(r1.X-(r2.X+r2.Width), r2.X-(r1.X+r1.Width)))
	dy := math.Max(0, math.Max(r1.Y-(r2.Y+r2.Height), r2.Y-(r1.Y+r1.Height)))
	if dx > 0 && dy > 0 {
		return math.Hypot(dx, dy)
	}
	return math.Max(dx, dy)
}

// FindWidgetsInArea returns the subset of widgets whose bounding boxes
// intersect (strictly overlap) the given area.
func FindWidgetsInArea(widgets []canvus.Widget, area canvus.Rectangle) []canvus.Widget {
	out := make([]canvus.Widget, 0, len(widgets))
	for _, w := range widgets {
		if Intersects(canvus.WidgetBoundingBox(w), area) {
			out = append(out, w)
		}
	}
	return out
}

// FindWidgetsContainingPoint returns widgets whose bounding boxes contain
// (inclusive) the given point.
func FindWidgetsContainingPoint(widgets []canvus.Widget, p canvus.Point) []canvus.Widget {
	out := make([]canvus.Widget, 0, len(widgets))
	for _, w := range widgets {
		r := canvus.WidgetBoundingBox(w)
		if p.X >= r.X && p.X <= r.X+r.Width && p.Y >= r.Y && p.Y <= r.Y+r.Height {
			out = append(out, w)
		}
	}
	return out
}

// GetCanvasBounds returns the smallest rectangle that contains every widget
// in the slice, or nil if the slice is empty.
func GetCanvasBounds(widgets []canvus.Widget) *canvus.Rectangle {
	if len(widgets) == 0 {
		return nil
	}
	bounds := canvus.WidgetBoundingBox(widgets[0])
	for _, w := range widgets[1:] {
		bounds = GetUnion(bounds, canvus.WidgetBoundingBox(w))
	}
	return &bounds
}
