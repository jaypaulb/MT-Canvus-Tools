package webui_test

import (
	"testing"

	webui "github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/webui"
)

func TestNewAPIClient_SecureTLS(t *testing.T) {
	c, err := webui.NewAPIClient("https://example.com/api/v1", "tok", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil APIClient")
	}
}

func TestNewAPIClient_InsecureTLS(t *testing.T) {
	c, err := webui.NewAPIClient("https://example.com/api/v1", "tok", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil APIClient")
	}
}

func TestGetWidgetPatchEndpoint(t *testing.T) {
	cases := []struct {
		widgetType string
		want       string
	}{
		{"note", "/notes"},
		{"NOTE", "/notes"},
		{"image", "/images"},
		{"pdf", "/pdfs"},
		{"video", "/videos"},
		{"unknown", "/notes"},
	}
	for _, tc := range cases {
		got := webui.GetWidgetPatchEndpoint(tc.widgetType)
		if got != tc.want {
			t.Errorf("GetWidgetPatchEndpoint(%q) = %q, want %q", tc.widgetType, got, tc.want)
		}
	}
}
