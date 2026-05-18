package ai

import (
	"testing"
)

func TestExtractAIResponse_ValidText(t *testing.T) {
	raw := `{"type":"text","content":"hello world"}`
	r, err := extractAIResponse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != "text" {
		t.Errorf("type: got %q", r.Type)
	}
	if r.Content != "hello world" {
		t.Errorf("content: got %q", r.Content)
	}
}

func TestExtractAIResponse_WithPreamble(t *testing.T) {
	raw := `Here is my JSON response: {"type":"image","content":"a vivid sunset"}`
	r, err := extractAIResponse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != "image" {
		t.Errorf("type: got %q", r.Type)
	}
}

func TestExtractAIResponse_NoJSON(t *testing.T) {
	_, err := extractAIResponse("no json here")
	if err == nil {
		t.Fatal("expected error for input with no JSON")
	}
}

func TestExtractAIResponse_MissingType(t *testing.T) {
	_, err := extractAIResponse(`{"content":"hello"}`)
	if err == nil {
		t.Fatal("expected error when 'type' is missing")
	}
}

func TestExtractAIResponse_MissingContent(t *testing.T) {
	_, err := extractAIResponse(`{"type":"text"}`)
	if err == nil {
		t.Fatal("expected error when 'content' is missing")
	}
}
