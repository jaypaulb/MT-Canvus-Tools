package extras

import (
	"context"
	"testing"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeLister implements CanvasLister for tests with no network IO.
type fakeLister struct {
	canvases []canvus.Canvas
	widgets  map[string][]canvus.Widget
}

func (f *fakeLister) ListCanvases(_ context.Context, _ *canvus.Filter) ([]canvus.Canvas, error) {
	return f.canvases, nil
}

func (f *fakeLister) GetCanvas(_ context.Context, id string) (*canvus.Canvas, error) {
	for i := range f.canvases {
		if f.canvases[i].ID == id {
			return &f.canvases[i], nil
		}
	}
	return nil, assert.AnError
}

func (f *fakeLister) ListWidgets(_ context.Context, canvasID string, _ *canvus.Filter, _ ...bool) ([]canvus.Widget, error) {
	return f.widgets[canvasID], nil
}

func TestCrossCanvasSearch_ByType(t *testing.T) {
	lister := &fakeLister{
		canvases: []canvus.Canvas{
			{ID: "c1", Name: "first"},
			{ID: "c2", Name: "second"},
		},
		widgets: map[string][]canvus.Widget{
			"c1": {
				{ID: "w1", WidgetType: "note"},
				{ID: "w2", WidgetType: "image"},
			},
			"c2": {
				{ID: "w3", WidgetType: "note"},
			},
		},
	}
	se := NewCrossCanvasSearch(lister)
	res, err := se.FindWidgetsByType(context.Background(), "note", SearchOptions{})
	require.NoError(t, err)
	require.Len(t, res, 2)
	for _, r := range res {
		assert.Equal(t, "note", r.WidgetType)
		assert.NotEmpty(t, r.DrillDownPath())
	}
}

func TestCrossCanvasSearch_InArea(t *testing.T) {
	lister := &fakeLister{
		canvases: []canvus.Canvas{{ID: "c1", Name: "x"}},
		widgets: map[string][]canvus.Widget{
			"c1": {
				widgetAt("inside", 10, 10, 5, 5),
				widgetAt("outside", 1000, 1000, 5, 5),
			},
		},
	}
	se := NewCrossCanvasSearch(lister)
	res, err := se.FindWidgetsInArea(context.Background(), canvus.Rectangle{X: 0, Y: 0, Width: 100, Height: 100}, SearchOptions{})
	require.NoError(t, err)
	require.Len(t, res, 1)
	assert.Equal(t, "inside", res[0].WidgetID)
}

func TestCrossCanvasSearch_MaxResults(t *testing.T) {
	lister := &fakeLister{
		canvases: []canvus.Canvas{{ID: "c1"}},
		widgets: map[string][]canvus.Widget{
			"c1": {
				{ID: "a", WidgetType: "note"},
				{ID: "b", WidgetType: "note"},
				{ID: "c", WidgetType: "note"},
			},
		},
	}
	se := NewCrossCanvasSearch(lister)
	res, err := se.FindWidgetsByType(context.Background(), "note", SearchOptions{MaxResults: 2})
	require.NoError(t, err)
	assert.Len(t, res, 2)
}
