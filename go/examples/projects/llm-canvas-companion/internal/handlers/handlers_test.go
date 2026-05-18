package handlers

import (
	"strings"
	"testing"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// --- splitIntoChunks ---

func TestSplitIntoChunks_SmallText(t *testing.T) {
	text := "Hello world\n\nThis is a paragraph."
	chunks := splitIntoChunks(text, 10000)
	if len(chunks) != 1 {
		t.Errorf("expected 1 chunk for small text, got %d", len(chunks))
	}
}

func TestSplitIntoChunks_LargeText(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < 100; i++ {
		sb.WriteString(strings.Repeat("a", 200))
		sb.WriteString("\n\n")
	}
	chunks := splitIntoChunks(sb.String(), 500)
	if len(chunks) < 2 {
		t.Errorf("expected multiple chunks for large text, got %d", len(chunks))
	}
}

func TestSplitIntoChunks_EmptyText(t *testing.T) {
	chunks := splitIntoChunks("", 1000)
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for empty text, got %d", len(chunks))
	}
}

// --- extractJSONContent ---

func TestExtractJSONContent_Valid(t *testing.T) {
	content, err := extractJSONContent(`{"type":"text","content":"summary here"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "summary here" {
		t.Errorf("content: got %q", content)
	}
}

func TestExtractJSONContent_WithPreamble(t *testing.T) {
	raw := `Sure! Here is the analysis: {"type":"text","content":"overview"}`
	content, err := extractJSONContent(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "overview" {
		t.Errorf("content: got %q", content)
	}
}

func TestExtractJSONContent_NoJSON(t *testing.T) {
	_, err := extractJSONContent("plain text no JSON")
	if err == nil {
		t.Fatal("expected error for non-JSON input")
	}
}

// --- sizeForContent ---

func TestSizeForContent_ShortContent(t *testing.T) {
	w := canvus.Widget{Scale: 1.0}
	_, scale := sizeForContent("short", w, false)
	if scale <= 0 {
		t.Errorf("expected positive scale, got %v", scale)
	}
}

func TestSizeForContent_ErrorNote(t *testing.T) {
	w := canvus.Widget{
		Scale: 0.5,
		Size:  &canvus.Size{Width: 300, Height: 200},
	}
	size, scale := sizeForContent("error content", w, true)
	// For error notes, size and scale should mirror the original.
	if size["width"] != 300.0 {
		t.Errorf("width: got %v, want 300", size["width"])
	}
	if size["height"] != 200.0 {
		t.Errorf("height: got %v, want 200", size["height"])
	}
	if scale != 0.5 {
		t.Errorf("scale: got %v, want 0.5", scale)
	}
}
