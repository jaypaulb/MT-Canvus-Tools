package extras

import (
	"testing"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/stretchr/testify/assert"
)

func rect(x, y, w, h float64) canvus.Rectangle {
	return canvus.Rectangle{X: x, Y: y, Width: w, Height: h}
}

func widgetAt(id string, x, y, w, h float64) canvus.Widget {
	return canvus.Widget{
		ID:       id,
		Location: &canvus.Point{X: x, Y: y},
		Size:     &canvus.Size{Width: w, Height: h},
	}
}

func TestIntersectsAndUnion(t *testing.T) {
	a := rect(0, 0, 10, 10)
	b := rect(5, 5, 10, 10)
	c := rect(20, 20, 5, 5)

	assert.True(t, Intersects(a, b))
	assert.False(t, Intersects(a, c))

	inter := GetIntersection(a, b)
	if assert.NotNil(t, inter) {
		assert.Equal(t, rect(5, 5, 5, 5), *inter)
	}
	assert.Nil(t, GetIntersection(a, c))

	union := GetUnion(a, c)
	assert.Equal(t, rect(0, 0, 25, 25), union)
}

func TestDistanceBetweenWidgets(t *testing.T) {
	w1 := widgetAt("a", 0, 0, 10, 10)
	w2 := widgetAt("b", 5, 5, 10, 10)
	w3 := widgetAt("c", 20, 0, 5, 5)
	w4 := widgetAt("d", 30, 30, 5, 5)

	assert.Equal(t, 0.0, DistanceBetweenWidgets(w1, w2))           // overlap
	assert.Equal(t, 10.0, DistanceBetweenWidgets(w1, w3))          // pure horizontal
	assert.InDelta(t, 28.28, DistanceBetweenWidgets(w1, w4), 0.01) // diagonal
}

func TestFindWidgetsHelpers(t *testing.T) {
	widgets := []canvus.Widget{
		widgetAt("a", 0, 0, 10, 10),
		widgetAt("b", 100, 100, 10, 10),
		widgetAt("c", 50, 50, 10, 10),
	}
	inArea := FindWidgetsInArea(widgets, rect(0, 0, 60, 60))
	assert.Len(t, inArea, 2)

	atPoint := FindWidgetsContainingPoint(widgets, canvus.Point{X: 105, Y: 105})
	assert.Len(t, atPoint, 1)
	assert.Equal(t, "b", atPoint[0].ID)

	bounds := GetCanvasBounds(widgets)
	if assert.NotNil(t, bounds) {
		assert.Equal(t, 0.0, bounds.X)
		assert.Equal(t, 110.0, bounds.Width)
	}

	assert.Nil(t, GetCanvasBounds(nil))
}
