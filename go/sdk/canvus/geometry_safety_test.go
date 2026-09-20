package canvus_test

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func TestKnownCameraConversion(t *testing.T) {
	visible, err := canvus.VisibleCanvasRegion(canvus.Rectangle{X: -1280, Y: -720, Width: 512, Height: 288}, canvus.Size{Width: 1280, Height: 720})
	require.NoError(t, err)
	require.Equal(t, canvus.Rectangle{X: 3200, Y: 1800, Width: 3200, Height: 1800}, visible)
	raw, err := canvus.ViewRectangleForRegion(canvus.Rectangle{X: 100, Y: 200, Width: 400, Height: 200}, canvus.Size{Width: 1000, Height: 600})
	require.NoError(t, err)
	require.Equal(t, canvus.Rectangle{X: -250, Y: -450, Width: 2500, Height: 1500}, raw)
}

func TestKnownNoteRegistrationAndParentScale(t *testing.T) {
	padding := 30.0
	nodes := []canvus.Widget{
		{ID: "parent", WidgetType: "Note", ParentID: "canvas", Location: &canvus.Point{X: 600, Y: 100}, Size: &canvus.Size{Width: 200, Height: 120}, Scale: 2},
		{ID: "child", WidgetType: "Note", ParentID: "parent", Location: &canvus.Point{X: 20, Y: 30}, Size: &canvus.Size{Width: 90, Height: 90}, Scale: .5},
	}
	model := canvus.GeometryModel{CanvasID: "canvas", NotePadding: &padding}
	root, err := canvus.WidgetCanvasBounds("parent", nodes, model)
	require.NoError(t, err)
	require.Equal(t, canvus.Rectangle{X: 570, Y: 70, Width: 400, Height: 240}, root)
	child, err := canvus.WidgetCanvasBounds("child", nodes, model)
	require.NoError(t, err)
	require.Equal(t, canvus.Rectangle{X: 700, Y: 220, Width: 90, Height: 90}, child)
}

func TestGeometryRejectsIncompleteAndUnsupportedInputs(t *testing.T) {
	for _, size := range []canvus.Size{{}, {Width: math.NaN(), Height: 10}, {Width: 10, Height: -1}} {
		_, err := canvus.ViewRectangleForRegion(canvus.Rectangle{Width: 10, Height: 10}, size)
		require.ErrorIs(t, err, canvus.ErrInvalidGeometry)
	}
	node := canvus.Widget{ID: "n", WidgetType: "Note", ParentID: "canvas", Location: &canvus.Point{}, Size: &canvus.Size{Width: 90, Height: 90}, Scale: 1}
	_, err := canvus.WidgetCanvasBounds("n", []canvus.Widget{node}, canvus.GeometryModel{})
	require.ErrorIs(t, err, canvus.ErrGeometryModelRequired)
	_, err = canvus.WidgetCanvasBounds("n", []canvus.Widget{node}, canvus.GeometryModel{CanvasID: "canvas"})
	require.ErrorIs(t, err, canvus.ErrGeometryModelRequired)
	p := 30.0
	model := canvus.GeometryModel{CanvasID: "canvas", NotePadding: &p}
	for _, parent := range []string{"", "missing", "n"} {
		node.ParentID = parent
		_, err = canvus.WidgetCanvasBounds("n", []canvus.Widget{node}, model)
		require.ErrorIs(t, err, canvus.ErrInvalidGeometry)
	}
	node.ParentID = "canvas"
	node.Scale = 0
	_, err = canvus.WidgetCanvasBounds("n", []canvus.Widget{node}, model)
	require.ErrorIs(t, err, canvus.ErrInvalidGeometry)
}

func TestSharedGeometryParityFixtures(t *testing.T) {
	var fixture struct {
		Rectangles []struct {
			Name              string
			A, B              canvus.Rectangle
			Contains, Touches bool
		}
		Widget canvus.Widget    `json:"raw_widget"`
		Bounds canvus.Rectangle `json:"raw_bounds"`
	}
	data, err := os.ReadFile("testdata/geometry-parity.json")
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &fixture))
	for _, item := range fixture.Rectangles {
		t.Run(item.Name, func(t *testing.T) {
			require.Equal(t, item.Contains, canvus.Contains(item.A, item.B))
			require.Equal(t, item.Touches, canvus.Touches(item.A, item.B))
		})
	}
	require.Equal(t, fixture.Bounds, canvus.WidgetBoundingBox(fixture.Widget))
}

func TestTouchesIncludesSharedEdge(t *testing.T) {
	require.True(t, canvus.Touches(canvus.Rectangle{Width: 10, Height: 10}, canvus.Rectangle{X: 10, Width: 4, Height: 4}))
}
