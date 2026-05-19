package webui_test

import (
	"testing"
	"time"

	canvus "github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	webui "github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/webui"
)

func TestCanvasEventFromWorkspace(t *testing.T) {
	ws := canvus.Workspace{
		CanvasID:      "canvas-123",
		WorkspaceName: "workspace-0",
	}
	ev := webui.CanvasEventFromWorkspace(ws)
	if ev.CanvasID != "canvas-123" {
		t.Errorf("CanvasID = %q, want %q", ev.CanvasID, "canvas-123")
	}
	if ev.CanvasName != "workspace-0" {
		t.Errorf("CanvasName = %q, want %q", ev.CanvasName, "workspace-0")
	}
	if ev.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
	if ev.Timestamp.After(time.Now().Add(time.Second)) {
		t.Error("Timestamp is in the future")
	}
}

func TestNewWorkspaceSubscriber(t *testing.T) {
	// Just verify construction doesn't panic with nil session.
	// (Cannot test stream without a live server.)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewWorkspaceSubscriber panicked: %v", r)
		}
	}()
	sub := webui.NewWorkspaceSubscriber(nil, "client-1")
	if sub == nil {
		t.Fatal("expected non-nil WorkspaceSubscriber")
	}
}
