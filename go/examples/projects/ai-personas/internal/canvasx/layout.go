package canvasx

// GridOffset is a (col,row) integer offset in the Q&A grid where (0,0) is
// the question note in the centre, x grows right, y grows down.
type GridOffset [2]int

// AnswerGridOffsets returns the cardinal offsets for the 4 persona answers
// (top, right, bottom, left).
func AnswerGridOffsets() []GridOffset {
	return []GridOffset{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
}

// MetaAnswerGridOffsets returns the diagonal offsets for the 4 meta-answers
// (top-right, bottom-right, bottom-left, top-left).
func MetaAnswerGridOffsets() []GridOffset {
	return []GridOffset{{1, -1}, {1, 1}, {-1, 1}, {-1, -1}}
}

// Spacing returns the gap between adjacent grid cells, calibrated to one
// fifth of the (scaled) question width.
func Spacing(qWidth, scale float64) float64 {
	return (qWidth * scale) / 5.0
}

// GridPosition returns the absolute (x, y) for a grid cell relative to the
// centre at (centreX, centreY) for cells of size (cellW, cellH) at scale and
// spacing.
func GridPosition(centreX, centreY float64, offset GridOffset, cellW, cellH, scale, spacing float64) (x, y float64) {
	x = centreX + float64(offset[0])*((cellW*scale)+spacing)
	y = centreY + float64(offset[1])*((cellH*scale)+spacing)
	return x, y
}

// PersonaColumnLayout computes the x/y/width/height for the i-th persona note
// inside the Personas anchor box. The anchor is divided into 4 vertical columns
// with a small border and gap; the persona image sits at the top of the
// column, the note below it.
func PersonaColumnLayout(index int, anchorX, anchorY, anchorW, anchorH float64) (x, imgY, noteY, colW, imgH, noteH float64) {
	const (
		border     = 0.02
		colWidth   = 0.23
		gap        = 0.01
		imageH     = 0.10
		noteHeight = 0.40
		noteYStart = 0.34
	)
	colW = anchorW * colWidth
	imgH = anchorH * imageH
	noteH = noteHeight * anchorH

	x = anchorX + anchorW*border + float64(index)*(anchorW*colWidth+anchorW*gap)
	imgY = anchorY + anchorH*border
	noteY = anchorY + (anchorH * noteYStart)
	return x, imgY, noteY, colW, imgH, noteH
}

// HelperNotePosition returns the position+size for a question helper note
// placed to the left of the question with a small vertical offset.
func HelperNotePosition(qX, qY, qW, qH float64) (hX, hY, hW, hH float64) {
	return qX - 1.2*qW, qY - 0.33*qH, qW, qH * 0.7
}

// BoundingBox describes a min/max rectangle.
type BoundingBox struct{ MinX, MinY, MaxX, MaxY float64 }

// Width returns MaxX - MinX.
func (b BoundingBox) Width() float64 { return b.MaxX - b.MinX }

// Height returns MaxY - MinY.
func (b BoundingBox) Height() float64 { return b.MaxY - b.MinY }
