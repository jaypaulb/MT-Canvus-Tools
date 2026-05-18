package mapping

import (
	"math"
	"testing"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/note-mapper/internal/llm"
)

func TestMapToZone_empty(t *testing.T) {
	result := MapToZone(nil, 1280, 720, ZoneParams{X: 0, Y: 0, Width: 640, Height: 480, Scale: 1})
	if len(result) != 0 {
		t.Fatalf("expected 0 notes, got %d", len(result))
	}
}

func TestMapToZone_zeroDimensions(t *testing.T) {
	notes := []llm.Note{
		{Content: "test", Color: "#fff", X: 100, Y: 100, Width: 50, Height: 50, Scale: 1},
	}
	// Zero image dimensions — should not panic, returns notes at zone origin.
	result := MapToZone(notes, 0, 0, ZoneParams{X: 10, Y: 20, Width: 640, Height: 480, Scale: 1})
	if len(result) != 1 {
		t.Fatalf("expected 1 note, got %d", len(result))
	}
}

func TestMapToZone_topLeftNote(t *testing.T) {
	// A note at (0,0) in a 1280×720 image mapped into a 640×480 zone at (100,200).
	// Scale = min(640/1280, 480/720) = min(0.5, 0.666) = 0.5
	// offsetX = (640 - 1280*0.5)/2 = (640-640)/2 = 0
	// offsetY = (480 - 720*0.5)/2 = (480-360)/2 = 60
	// Note: x = 100 + 10 + 0*0.5 + 0 = 110, y = 200 + 10 + 0*0.5 + 60 = 270
	notes := []llm.Note{
		{Content: "top-left", Color: "#ffeb3b", X: 0, Y: 0, Width: 200, Height: 200},
	}
	zone := ZoneParams{X: 100, Y: 200, Width: 640, Height: 480, Scale: 1}
	result := MapToZone(notes, 1280, 720, zone)
	if len(result) != 1 {
		t.Fatalf("expected 1 note, got %d", len(result))
	}
	n := result[0]

	wantX := 110.0 // 100 (zone.X) + 10 (inset) + 0*0.5 (note.X*scale) + 0 (offsetX)
	wantY := 270.0 // 200 (zone.Y) + 10 (inset) + 0*0.5 (note.Y*scale) + 60 (offsetY)
	wantW := 100.0 // 200*0.5
	wantH := 100.0 // 200*0.5

	if !near(n.Location["x"], wantX) {
		t.Errorf("location.x: got %v, want %v", n.Location["x"], wantX)
	}
	if !near(n.Location["y"], wantY) {
		t.Errorf("location.y: got %v, want %v", n.Location["y"], wantY)
	}
	if !near(n.Size["width"], wantW) {
		t.Errorf("size.width: got %v, want %v", n.Size["width"], wantW)
	}
	if !near(n.Size["height"], wantH) {
		t.Errorf("size.height: got %v, want %v", n.Size["height"], wantH)
	}
	if n.Scale != 1.0 {
		t.Errorf("scale: got %v, want 1.0", n.Scale)
	}
	if n.WidgetType != "Note" {
		t.Errorf("widget_type: got %v, want Note", n.WidgetType)
	}
}

func TestMapToZone_preservesContent(t *testing.T) {
	notes := []llm.Note{
		{Content: "hello", Color: "#ff0000", X: 0, Y: 0, Width: 10, Height: 10},
		{Content: "world", Color: "#00ff00", X: 640, Y: 360, Width: 10, Height: 10},
	}
	result := MapToZone(notes, 1280, 720, ZoneParams{Width: 640, Height: 480, Scale: 1})
	if len(result) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(result))
	}
	if result[0].Text != "hello" {
		t.Errorf("text[0]: got %q, want %q", result[0].Text, "hello")
	}
	if result[1].Text != "world" {
		t.Errorf("text[1]: got %q, want %q", result[1].Text, "world")
	}
	if result[0].BackgroundColor != "#ff0000" {
		t.Errorf("color[0]: got %q, want %q", result[0].BackgroundColor, "#ff0000")
	}
}

func TestMapWithAnchorScale_scalesCorrectly(t *testing.T) {
	// Image 1280×720; anchor 640×360 scale=2.0 at (0,0).
	// scaleFactor = min(640/1280, 360/720) = min(0.5, 0.5) = 0.5
	// finalScale = 0.5 * 2.0 = 1.0
	// note at (100,50): x=0+10+100*1.0=110, y=0+10+50*1.0=60
	notes := []llm.Note{
		{Content: "scale-test", Color: "#123456", X: 100, Y: 50, Width: 80, Height: 40},
	}
	zone := ZoneParams{X: 0, Y: 0, Width: 640, Height: 360, Scale: 2.0}
	result := MapWithAnchorScale(notes, 1280, 720, zone)
	if len(result) != 1 {
		t.Fatalf("expected 1 note, got %d", len(result))
	}
	n := result[0]
	if !near(n.Location["x"], 110.0) {
		t.Errorf("location.x: got %v, want 110", n.Location["x"])
	}
	if !near(n.Location["y"], 60.0) {
		t.Errorf("location.y: got %v, want 60", n.Location["y"])
	}
	if !near(n.Size["width"], 80.0) {
		t.Errorf("size.width: got %v, want 80", n.Size["width"])
	}
	if !near(n.Size["height"], 40.0) {
		t.Errorf("size.height: got %v, want 40", n.Size["height"])
	}
}

func TestMinHelper(t *testing.T) {
	tests := []struct {
		a, b float64
		want float64
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{-1, 1, -1},
	}
	for _, tt := range tests {
		got := min(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("min(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

// near returns true if |a-b| < 0.01 (floating-point tolerance).
func near(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}
