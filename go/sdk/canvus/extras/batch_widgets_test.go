package extras

import (
	"testing"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchOps_Move(t *testing.T) {
	b := NewBatchWidgetOperations()
	widgets := []canvus.Widget{
		widgetAt("a", 0, 0, 10, 10),
		widgetAt("b", 5, 5, 10, 10),
		{ID: "c"}, // no Location → should be skipped
	}
	ops := b.MoveWidgets(widgets, 10, 20)
	require.Len(t, ops, 2)
	loc := ops[0].Payload["location"].(map[string]any)
	assert.Equal(t, 10.0, loc["x"])
	assert.Equal(t, 20.0, loc["y"])
}

func TestBatchOps_Resize(t *testing.T) {
	b := NewBatchWidgetOperations()
	widgets := []canvus.Widget{
		widgetAt("a", 0, 0, 10, 20),
		{ID: "c"}, // no Size → skipped
	}
	ops := b.ResizeWidgets(widgets, 2)
	require.Len(t, ops, 1)
	size := ops[0].Payload["size"].(map[string]any)
	assert.Equal(t, 20.0, size["width"])
	assert.Equal(t, 40.0, size["height"])
}

func TestBatchOps_ContainAndTouchID(t *testing.T) {
	b := NewBatchWidgetOperations()
	widgets := []canvus.Widget{
		widgetAt("outer", 0, 0, 100, 100),
		widgetAt("inner", 10, 10, 5, 5),
		widgetAt("touchy", 95, 95, 10, 10),
		widgetAt("far", 500, 500, 5, 5),
	}
	containers := b.WidgetsContainID(widgets, "inner")
	require.Len(t, containers, 1)
	assert.Equal(t, "outer", containers[0].ID)

	touching := b.WidgetsTouchID(widgets, "outer")
	// inner is inside (touches), touchy overlaps, far is not touching.
	require.Len(t, touching, 2)
}
