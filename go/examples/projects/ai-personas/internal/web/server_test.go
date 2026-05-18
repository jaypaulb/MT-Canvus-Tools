package web

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func anchorAt(x, y, w, h float64) *canvus.Anchor {
	return &canvus.Anchor{
		Location: &canvus.Point{X: x, Y: y},
		Size:     &canvus.Size{Width: w, Height: h},
	}
}

func TestFindFreeSegment_EmptyAnchor(t *testing.T) {
	// With an empty 5x4 anchor, cell 0 is reserved for QR; first free is cell 1.
	a := anchorAt(0, 0, 500, 400)
	x, y, w, h, scale, err := findFreeSegment(nil, nil, a)
	require.NoError(t, err)
	// segW=100, segH=100; cell 1 = (col=1, row=0); centre (150, 50).
	assert.InDelta(t, 150.0, x, 1e-9)
	assert.InDelta(t, 50.0, y, 1e-9)
	assert.InDelta(t, 100.0*(2.0/3.0), w, 1e-9)
	assert.InDelta(t, 100.0*(2.0/3.0), h, 1e-9)
	assert.InDelta(t, 1.5/3.5, scale, 1e-9)
}

func TestFindFreeSegment_AllOccupied(t *testing.T) {
	a := anchorAt(0, 0, 500, 400)
	// Cover the whole anchor with one giant note → no free cells.
	notes := []canvus.Note{
		{Location: &canvus.Point{X: 0, Y: 0}, Size: &canvus.Size{Width: 500, Height: 400}},
	}
	_, _, _, _, _, err := findFreeSegment(notes, nil, a)
	assert.Error(t, err)
}

func TestFindFreeSegment_PartiallyOccupied(t *testing.T) {
	a := anchorAt(0, 0, 500, 400)
	// Occupy cells 1 and 2 (row 0, cols 1 and 2).
	notes := []canvus.Note{
		{Location: &canvus.Point{X: 105, Y: 5}, Size: &canvus.Size{Width: 90, Height: 90}},
		{Location: &canvus.Point{X: 205, Y: 5}, Size: &canvus.Size{Width: 90, Height: 90}},
	}
	x, y, _, _, _, err := findFreeSegment(notes, nil, a)
	require.NoError(t, err)
	// Should pick cell 3 = (col=3, row=0); centre (350, 50).
	assert.InDelta(t, 350.0, x, 1e-9)
	assert.InDelta(t, 50.0, y, 1e-9)
}

func TestFindFreeSegment_MissingGeometry(t *testing.T) {
	_, _, _, _, _, err := findFreeSegment(nil, nil, &canvus.Anchor{})
	assert.Error(t, err)
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{45 * time.Second, "45s"},
		{90 * time.Second, "1m 30s"},
		{2*time.Hour + 5*time.Minute + 3*time.Second, "2h 5m 3s"},
		{49 * time.Hour, "2d 1h 0m 0s"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, formatUptime(tt.d), "duration=%v", tt.d)
	}
}
