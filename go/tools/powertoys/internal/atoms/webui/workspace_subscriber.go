// go/tools/powertoys/internal/atoms/webui/workspace_subscriber.go
package webui

import (
	"context"
	"fmt"
	"time"

	canvus "github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// CanvasEvent carries a canvas_id update from a workspace subscription.
type CanvasEvent struct {
	CanvasID   string
	CanvasName string
	Timestamp  time.Time
}

// CanvasEventFromWorkspace translates an SDK Workspace into a CanvasEvent.
func CanvasEventFromWorkspace(ws canvus.Workspace) CanvasEvent {
	return CanvasEvent{
		CanvasID:   ws.CanvasID,
		CanvasName: ws.WorkspaceName,
		Timestamp:  time.Now(),
	}
}

// WorkspaceSubscriber streams canvas updates for a given client using the SDK.
type WorkspaceSubscriber struct {
	session  *canvus.Session
	clientID string
}

// NewWorkspaceSubscriber creates a subscriber for the given client.
func NewWorkspaceSubscriber(session *canvus.Session, clientID string) *WorkspaceSubscriber {
	return &WorkspaceSubscriber{session: session, clientID: clientID}
}

// Subscribe opens an SDK workspace subscription and emits CanvasEvents.
// Reconnects automatically on stream errors (5-second back-off).
func (ws *WorkspaceSubscriber) Subscribe(ctx context.Context) (<-chan CanvasEvent, <-chan error) {
	eventChan := make(chan CanvasEvent, 10)
	errChan := make(chan error, 1)

	go func() {
		defer close(eventChan)
		defer close(errChan)

		for {
			ch, err := ws.session.SubscribeClientWorkspaces(ctx, ws.clientID)
			if err != nil {
				select {
				case errChan <- fmt.Errorf("WorkspaceSubscriber: open: %w", err):
				case <-ctx.Done():
					return
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			for workspace := range ch {
				select {
				case <-ctx.Done():
					return
				case eventChan <- CanvasEventFromWorkspace(workspace):
				}
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
		}
	}()

	return eventChan, errChan
}
