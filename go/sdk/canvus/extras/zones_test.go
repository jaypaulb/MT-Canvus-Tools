package extras

import (
	"testing"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWidgetZoneManager_Create(t *testing.T) {
	m := NewWidgetZoneManager()
	widgets := []canvus.Widget{
		widgetAt("a", 0, 0, 10, 10),
		widgetAt("b", 50, 50, 10, 10),
	}
	zone, err := m.CreateZoneFromWidgets(widgets, "zone1", "test", 5)
	require.NoError(t, err)
	assert.Equal(t, -5.0, zone.Location.X)
	assert.Equal(t, -5.0, zone.Location.Y)
	assert.Equal(t, 70.0, zone.Size.Width)
	assert.Equal(t, 70.0, zone.Size.Height)
}

func TestWidgetZoneManager_Empty(t *testing.T) {
	m := NewWidgetZoneManager()
	_, err := m.CreateZoneFromWidgets(nil, "z", "", 0)
	assert.Error(t, err)
}

func TestWidgetsInAndTouchingZone(t *testing.T) {
	m := NewWidgetZoneManager()
	zone := WidgetZone{
		Location: canvus.Point{X: 0, Y: 0},
		Size:     canvus.Size{Width: 100, Height: 100},
	}
	widgets := []canvus.Widget{
		widgetAt("inside", 10, 10, 10, 10),
		widgetAt("partial", 95, 95, 20, 20),
		widgetAt("far", 200, 200, 5, 5),
	}
	in := m.WidgetsInZone(widgets, zone)
	require.Len(t, in, 1)
	assert.Equal(t, "inside", in[0].ID)

	touching := m.WidgetsTouchingZone(widgets, zone)
	require.Len(t, touching, 2)
}
