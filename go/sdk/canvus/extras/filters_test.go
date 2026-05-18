package extras

import (
	"testing"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilter_EqualsAndContains(t *testing.T) {
	item := map[string]any{
		"title": "Hello World",
		"size":  map[string]any{"width": 100.0, "height": 50.0},
	}
	f := NewFilter().
		AddCondition("title", OpEquals, "Hello World").
		AddCondition("size.width", OpGreaterEqual, 100)
	assert.True(t, f.Matches(item))

	f2 := NewFilter().AddCondition("title", OpContains, "Nope")
	assert.False(t, f2.Matches(item))
}

func TestFilter_Wildcard(t *testing.T) {
	item := map[string]any{"title": "Hello World"}
	f := NewWildcardFilter("hello*")
	assert.True(t, f.Matches(item))
	f2 := NewWildcardFilter("planet*")
	assert.False(t, f2.Matches(item))
}

func TestFilter_TextAndWidgetType(t *testing.T) {
	item := map[string]any{"text": "important note", "title": "ignore", "widget_type": "note"}
	tf := NewTextFilter("important")
	// AND semantics across fields: contains in "text" passes, but "title" and "description" don't contain.
	assert.False(t, tf.Matches(item))

	wf := NewWidgetTypeFilter("Note", "Image")
	assert.True(t, wf.Matches(item))

	wfMiss := NewWidgetTypeFilter("Pdf")
	assert.False(t, wfMiss.Matches(item))
}

func TestFilter_Spatial(t *testing.T) {
	item := map[string]any{
		"location": map[string]any{"x": 10.0, "y": 10.0},
		"size":     map[string]any{"width": 20.0, "height": 20.0},
	}
	area := canvus.Rectangle{X: 0, Y: 0, Width: 100, Height: 100}
	fInter := NewSpatialFilter(area, "intersects")
	assert.True(t, fInter.Matches(item))
	fWithin := NewSpatialFilter(area, "within")
	assert.True(t, fWithin.Matches(item))
	smallArea := canvus.Rectangle{X: 0, Y: 0, Width: 5, Height: 5}
	fWithinMiss := NewSpatialFilter(smallArea, "within")
	assert.False(t, fWithinMiss.Matches(item))
}

func TestFilter_RoundTripToFromMap(t *testing.T) {
	f := NewFilter().AddCondition("title", OpStartsWith, "Hello")
	m := f.ToMap()
	// Re-parse via simulated JSON-ish shape (conditions as []any of map[string]any).
	conds := m["conditions"].([]map[string]any)
	asAny := make([]any, len(conds))
	for i, c := range conds {
		asAny[i] = c
	}
	roundtrip, err := FromMap(map[string]any{"conditions": asAny})
	require.NoError(t, err)
	assert.Equal(t, f.Conditions, roundtrip.Conditions)
}

func TestCombineFilters(t *testing.T) {
	a := NewFilter().AddCondition("title", OpEquals, "x")
	b := NewFilter().AddCondition("widget_type", OpEquals, "note")
	c := CombineFilters(a, b, nil)
	assert.Len(t, c.Conditions, 2)
}
