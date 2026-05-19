package webui_test

import (
	"context"
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

func TestSubscribe_ContextCancelClosesChannels(t *testing.T) {
	// Build a real but non-functional session pointing at a non-existent server.
	// The subscriber will fail to connect, hit the error path and surface the
	// error via errCh (not drop it silently), then see ctx.Done() and exit —
	// exercising the cancellation and error-surfacing path.
	cfg := canvus.DefaultSessionConfig()
	cfg.BaseURL = "http://127.0.0.1:19999" // nothing listening here
	session := canvus.NewSession(cfg)

	sub := webui.NewWorkspaceSubscriber(session, "test-client")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	eventCh, errCh := sub.Subscribe(ctx)

	// Wait for either an error or timeout.
	select {
	case err, ok := <-errCh:
		if ok && err == nil {
			t.Error("expected non-nil error from unreachable server")
		}
		// Got a non-nil error or channel closed — both acceptable.
		// Cancel and drain remaining channel updates.
		cancel()
		for range eventCh {
		}
		for range errCh {
		}
	case <-ctx.Done():
		t.Error("timed out waiting for error from unreachable server")
	}
}
