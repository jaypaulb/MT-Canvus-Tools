package widget

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWidgetTypePath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"note lowercase", "note", "notes", false},
		{"Note titlecase", "Note", "notes", false},
		{"image", "image", "images", false},
		{"video", "video", "videos", false},
		{"pdf", "pdf", "pdfs", false},
		{"browser", "browser", "browsers", false},
		{"anchor", "anchor", "anchors", false},
		{"table", "table", "tables", false},
		{"connector", "connector", "connectors", false},
		{"unknown", "frobnicator", "", true},
		{"empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := widgetTypePath(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPinCmd_Registration(t *testing.T) {
	found := false
	for _, c := range WidgetCmd.Commands() {
		if c.Name() == "pin" {
			found = true
			assert.Equal(t, "pin <canvas-id> <widget-id>", c.Use,
				"pin command should take both canvas and widget ids")
			break
		}
	}
	assert.True(t, found, "pin command should be registered")
}

func TestUnpinCmd_Registration(t *testing.T) {
	found := false
	for _, c := range WidgetCmd.Commands() {
		if c.Name() == "unpin" {
			found = true
			assert.Equal(t, "unpin <canvas-id> <widget-id>", c.Use,
				"unpin command should take both canvas and widget ids")
			break
		}
	}
	assert.True(t, found, "unpin command should be registered")
}

func TestCopyCmd_Registration(t *testing.T) {
	found := false
	for _, c := range WidgetCmd.Commands() {
		if c.Name() == "copy" {
			found = true
			assert.Equal(t, "copy <source-canvas-id> <widget-id> <target-canvas-id>", c.Use,
				"copy command should take source-canvas, widget, target-canvas")
			break
		}
	}
	assert.True(t, found, "copy command should be registered")
}

func TestMoveCmd_Registration(t *testing.T) {
	found := false
	for _, c := range WidgetCmd.Commands() {
		if c.Name() == "move" {
			found = true
			assert.Equal(t, "move <source-canvas-id> <widget-id> <target-canvas-id>", c.Use,
				"move command should take source-canvas, widget, target-canvas")
			break
		}
	}
	assert.True(t, found, "move command should be registered")
}
