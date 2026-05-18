package canvasx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPersonaColors(t *testing.T) {
	c := PersonaColors()
	assert.Len(t, c, 4)
	for _, hex := range c {
		assert.True(t, PersonaColorSet()[hex],
			"PersonaColors entry %q missing from PersonaColorSet", hex)
	}
}

func TestBuildConnectorPayload(t *testing.T) {
	p := BuildConnectorPayload("src-1", "dst-2")
	assert.Equal(t, "Connector", p["widget_type"])

	src, ok := p["src"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "src-1", src["id"])
	assert.Equal(t, true, src["auto_location"])
	assert.Equal(t, "none", src["tip"])

	dst, ok := p["dst"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "dst-2", dst["id"])
	assert.Equal(t, "solid-equilateral-triangle", dst["tip"])

	assert.Equal(t, "#e7e7f2ff", p["line_color"])
	assert.Equal(t, 5, p["line_width"])
}
