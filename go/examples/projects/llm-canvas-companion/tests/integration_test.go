//go:build integration

// Package tests — integration tests requiring a live Canvus server and LLM.
//
// Run with:
//
//	CANVUS_API_URL=https://... CANVUS_API_KEY=... CANVAS_ID=... \
//	  go test -tags=integration ./tests/...
//
// The tests create and delete real widgets on the target canvas.
// Do not run against production canvases.
package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/llm-canvas-companion/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

var (
	integSession  *canvus.Session
	integCfg      *config.Config
	integCanvasID string
)

func TestMain(m *testing.M) {
	cfg, err := config.Load()
	if err != nil {
		// Integration tests require full config; skip gracefully if not set.
		os.Exit(0)
	}
	integCfg = cfg
	integCanvasID = cfg.CanvasID

	sessionCfg := &canvus.SessionConfig{
		BaseURL:        cfg.CanvusAPIURL,
		RequestTimeout: 30 * time.Second,
	}
	integSession = canvus.NewSession(sessionCfg, canvus.WithAPIKey(cfg.CanvusAPIKey))

	os.Exit(m.Run())
}

func TestIntegration_NoteCreateUpdateDelete(t *testing.T) {
	ctx := context.Background()

	note, err := integSession.CreateNote(ctx, integCanvasID, map[string]any{
		"text": "integration test note",
		"location": map[string]any{"x": 500.0, "y": 500.0},
		"size": map[string]any{"width": 200.0, "height": 150.0},
		"state": "normal",
		"widget_type": "Note",
	})
	if err != nil {
		t.Fatalf("CreateNote: %v", err)
	}
	t.Logf("created note %s", note.ID)

	_, err = integSession.UpdateNote(ctx, integCanvasID, note.ID, map[string]any{
		"text": "updated integration note",
	})
	if err != nil {
		t.Errorf("UpdateNote: %v", err)
	}

	if err := integSession.DeleteNote(ctx, integCanvasID, note.ID); err != nil {
		t.Errorf("DeleteNote: %v", err)
	}
	t.Log("note lifecycle passed")
}

func TestIntegration_ListWidgets(t *testing.T) {
	ctx := context.Background()
	widgets, err := integSession.ListWidgets(ctx, integCanvasID, nil)
	if err != nil {
		t.Fatalf("ListWidgets: %v", err)
	}
	t.Logf("canvas has %d widgets", len(widgets))
}

func TestIntegration_LLMConnection(t *testing.T) {
	if integCfg.OpenAIAPIKey == "" && integCfg.BaseLLMURL == "" {
		t.Skip("no LLM configured")
	}
	// Just verify config parsed; actual LLM call tested manually.
	t.Logf("LLM base URL: %s", integCfg.BaseLLMURL)
	t.Logf("text model: %s", integCfg.OpenAINoteModel)
}
