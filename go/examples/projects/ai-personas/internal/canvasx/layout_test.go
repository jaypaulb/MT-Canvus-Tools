package canvasx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpacing(t *testing.T) {
	assert.InDelta(t, 100.0, Spacing(500, 1.0), 1e-9)
	assert.InDelta(t, 200.0, Spacing(500, 2.0), 1e-9)
}

func TestGridPosition(t *testing.T) {
	// At centre (100, 200) with offset (1, -1), cell 50x40, scale 1, spacing 10.
	// step.x = (50*1) + 10 = 60; step.y = (40*1) + 10 = 50.
	x, y := GridPosition(100, 200, GridOffset{1, -1}, 50, 40, 1, 10)
	assert.InDelta(t, 160.0, x, 1e-9)
	assert.InDelta(t, 150.0, y, 1e-9)
}

func TestAnswerGridOffsets(t *testing.T) {
	a := AnswerGridOffsets()
	// Cardinal directions: top, right, bottom, left.
	assert.Equal(t, []GridOffset{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}, a)
}

func TestMetaAnswerGridOffsets(t *testing.T) {
	m := MetaAnswerGridOffsets()
	// Diagonals.
	assert.Equal(t, []GridOffset{{1, -1}, {1, 1}, {-1, 1}, {-1, -1}}, m)
}

func TestPersonaColumnLayout_FourColumns(t *testing.T) {
	// Anchor at (0,0), 1000 wide, 500 tall. Each column should be 230 wide,
	// note 200 tall, image 50 tall.
	x0, _, _, colW, imgH, noteH := PersonaColumnLayout(0, 0, 0, 1000, 500)
	x1, _, _, _, _, _ := PersonaColumnLayout(1, 0, 0, 1000, 500)
	assert.InDelta(t, 20.0, x0, 1e-9)    // border 2% * 1000
	assert.InDelta(t, 230.0, colW, 1e-9) // 23% * 1000
	assert.InDelta(t, 50.0, imgH, 1e-9)  // 10% * 500
	assert.InDelta(t, 200.0, noteH, 1e-9)
	// Column 1 is further right than column 0 by (colWidth + gap) * anchorW.
	assert.Greater(t, x1, x0)
	assert.InDelta(t, 240.0, x1-x0, 1e-9)
}

func TestBoundingBox(t *testing.T) {
	b := BoundingBox{MinX: 10, MinY: 20, MaxX: 30, MaxY: 60}
	assert.Equal(t, 20.0, b.Width())
	assert.Equal(t, 40.0, b.Height())
}

func TestHelperNotePosition(t *testing.T) {
	hX, hY, hW, hH := HelperNotePosition(100, 200, 50, 80)
	assert.InDelta(t, 40.0, hX, 1e-9)  // 100 - 1.2*50
	assert.InDelta(t, 173.6, hY, 1e-9) // 200 - 0.33*80
	assert.InDelta(t, 50.0, hW, 1e-9)
	assert.InDelta(t, 56.0, hH, 1e-9) // 80 * 0.7
}
