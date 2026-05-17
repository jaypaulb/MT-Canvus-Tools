package canvus

// Contains returns true if rectangle a fully contains rectangle b.
func Contains(a, b Rectangle) bool {
	return b.X >= a.X && b.Y >= a.Y &&
		b.X+b.Width <= a.X+a.Width &&
		b.Y+b.Height <= a.Y+a.Height
}

// Touches returns true if rectangles a and b overlap or touch at any edge/corner.
func Touches(a, b Rectangle) bool {
	return a.X < b.X+b.Width && a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height && a.Y+a.Height > b.Y
}

// WidgetBoundingBox returns the bounding box (Rectangle) for a Widget.
func WidgetBoundingBox(w Widget) Rectangle {
	x, y := 0.0, 0.0
	wVal, hVal := 0.0, 0.0
	if w.Location != nil {
		x = w.Location.X
		y = w.Location.Y
	}
	if w.Size != nil {
		wVal = w.Size.Width
		hVal = w.Size.Height
	}
	return Rectangle{X: x, Y: y, Width: wVal, Height: hVal}
}

// WidgetContains returns true if widget a fully contains widget b.
func WidgetContains(a, b Widget) bool {
	return Contains(WidgetBoundingBox(a), WidgetBoundingBox(b))
}

// WidgetsTouch returns true if widgets a and b touch or overlap.
func WidgetsTouch(a, b Widget) bool {
	return Touches(WidgetBoundingBox(a), WidgetBoundingBox(b))
}
