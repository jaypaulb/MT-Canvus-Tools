package canvus

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

// CameraUpdateError identifies the stage of a non-atomic camera operation.
// ZoomAcknowledged=false means no acknowledgment, NOT proof nothing happened.
// Never automatically replay or undo an uncertain/partial camera operation.
type CameraUpdateError struct {
	Stage            string // "zoom", "settle", or "pan"
	ZoomAcknowledged bool
	Err              error
}

// Error describes the failed camera stage without discarding its outcome.
func (e *CameraUpdateError) Error() string {
	return fmt.Sprintf("camera %s (zoom acknowledged=%t): %v", e.Stage, e.ZoomAcknowledged, e.Err)
}

// Unwrap preserves API status, accepted-response and cancellation errors.
func (e *CameraUpdateError) Unwrap() error { return e.Err }

// FrameWorkspaceRegion fits a canvas-pixel region using the documented two-step
// zoom-then-pan protocol. canvasID is required and checked against the selected
// workspace. The initial actor, resolved index, canvas/server and pixel size are
// bound across the operation. Metadata is rechecked before pan, but no server
// atomic compare-and-swap or physical-operator lock exists: this is not a lock.
func (s *Session) FrameWorkspaceRegion(ctx context.Context, clientID string, selector WorkspaceSelector, canvasID string, region Rectangle) error {
	if canvasID == "" {
		return fmt.Errorf("FrameWorkspaceRegion: %w: canvas ID required", ErrInvalidRequest)
	}
	ctx = s.freezeAuthority(ctx)
	ws, err := s.GetWorkspace(ctx, clientID, selector)
	if err != nil {
		return fmt.Errorf("FrameWorkspaceRegion: %w", err)
	}
	if ws.CanvasID != canvasID {
		return fmt.Errorf("FrameWorkspaceRegion: %w", ErrWorkspaceChanged)
	}
	return s.frameWorkspaceRegion(ctx, clientID, *ws, region)
}

func (s *Session) frameWorkspaceRegion(ctx context.Context, clientID string, before Workspace, region Rectangle) error {
	if !before.validIndex() || before.CanvasID == "" || before.ServerID == "" || before.WorkspaceState != "open" || before.Size == nil || !positiveSize(*before.Size) {
		return fmt.Errorf("frame workspace: %w", ErrWorkspaceMetadata)
	}
	raw, err := ViewRectangleForRegion(region, *before.Size)
	if err != nil {
		return fmt.Errorf("frame workspace geometry: %w", err)
	}
	selector := WorkspaceSelector{Index: &before.Index}
	current, err := s.GetWorkspace(ctx, clientID, selector)
	if err != nil {
		return fmt.Errorf("frame workspace recheck: %w", err)
	}
	if !sameWorkspaceTarget(before, *current) {
		return fmt.Errorf("frame workspace: %w", ErrWorkspaceChanged)
	}
	zoom := Rectangle{Width: raw.Width, Height: raw.Height}
	if _, err = s.UpdateWorkspace(ctx, clientID, selector, UpdateWorkspaceRequest{ViewRectangle: &zoom}); err != nil {
		var accepted *AcceptedResponseError
		return &CameraUpdateError{Stage: "zoom", ZoomAcknowledged: errors.As(err, &accepted), Err: err}
	}
	// Existing canvus-stress focusView waits 100 ms between zoom and pan.
	if err = waitForRetry(ctx, 100*time.Millisecond); err != nil {
		return &CameraUpdateError{Stage: "settle", ZoomAcknowledged: true, Err: err}
	}
	settle, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	for {
		current, err := s.GetWorkspace(settle, clientID, selector)
		if err != nil {
			return &CameraUpdateError{Stage: "settle", ZoomAcknowledged: true, Err: err}
		}
		if !sameWorkspaceTarget(before, *current) {
			return &CameraUpdateError{Stage: "settle", ZoomAcknowledged: true, Err: ErrWorkspaceChanged}
		}
		if current.ViewRectangle != nil && nearCamera(current.ViewRectangle.Width, raw.Width) && nearCamera(current.ViewRectangle.Height, raw.Height) {
			break
		}
		if err = waitForRetry(settle, 25*time.Millisecond); err != nil {
			return &CameraUpdateError{Stage: "settle", ZoomAcknowledged: true, Err: err}
		}
	}
	if _, err = s.UpdateWorkspace(ctx, clientID, selector, UpdateWorkspaceRequest{ViewRectangle: &raw}); err != nil {
		return &CameraUpdateError{Stage: "pan", ZoomAcknowledged: true, Err: err}
	}
	return nil
}

func sameWorkspaceTarget(a, b Workspace) bool {
	return a.Index == b.Index && a.CanvasID == b.CanvasID && a.ServerID == b.ServerID && a.User == b.User && b.WorkspaceState == "open" && a.Size != nil && b.Size != nil && *a.Size == *b.Size
}

func nearCamera(a, b float64) bool {
	return finite(a, b) && math.Abs(a-b) <= 1e-5*math.Max(1, math.Abs(b))
}

func widgetAncestry(ctx context.Context, getter WorkspaceWidgetGetter, canvasID, id string) ([]Widget, error) {
	if getter == nil || id == "" {
		return nil, fmt.Errorf("widget ancestry: %w", ErrInvalidRequest)
	}
	var nodes []Widget
	seen := map[string]bool{}
	for id != "" {
		if seen[id] || len(nodes) >= 128 {
			return nil, fmt.Errorf("widget ancestry: %w: cycle/depth limit", ErrInvalidGeometry)
		}
		seen[id] = true
		w, err := getter.GetWidget(ctx, canvasID, id)
		if err != nil {
			return nil, fmt.Errorf("widget ancestry: %w", err)
		}
		if w == nil || w.ID != id {
			return nil, fmt.Errorf("widget ancestry: %w: missing/changed identity", ErrInvalidGeometry)
		}
		nodes = append(nodes, *w)
		if strings.EqualFold(w.WidgetType, "SharedCanvas") {
			return nodes, nil
		}
		id = w.ParentID
	}
	return nil, fmt.Errorf("widget ancestry: %w: SharedCanvas root missing", ErrInvalidGeometry)
}
