package canvus

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// WorkspaceWidgetGetter allows fetching a widget by ID for viewport logic.
type WorkspaceWidgetGetter interface {
	GetWidget(ctx context.Context, canvasID, widgetID string) (*Widget, error)
}

func (s *Session) resolveWorkspaceIndex(ctx context.Context, clientID string, selector WorkspaceSelector) (int, error) {
	count := 0
	if selector.Index != nil {
		count++
	}
	if selector.Name != nil {
		count++
	}
	if selector.User != nil {
		count++
	}
	if clientID == "" || count != 1 {
		return 0, fmt.Errorf("resolve workspace: %w: client and exactly one selector required", ErrInvalidRequest)
	}
	if selector.Index != nil {
		if *selector.Index < 0 {
			return 0, fmt.Errorf("resolve workspace: %w: negative index", ErrInvalidRequest)
		}
		return *selector.Index, nil
	}
	if (selector.Name != nil && *selector.Name == "") || (selector.User != nil && *selector.User == "") {
		return 0, fmt.Errorf("resolve workspace: %w: empty selector", ErrInvalidRequest)
	}
	workspaces, err := s.ListWorkspaces(ctx, clientID)
	if err != nil {
		return 0, err
	}
	var matches []Workspace
	for _, ws := range workspaces {
		if (selector.Name != nil && ws.WorkspaceName == *selector.Name) || (selector.User != nil && ws.User == *selector.User) {
			matches = append(matches, ws)
		}
	}
	if len(matches) == 0 {
		return 0, fmt.Errorf("resolve workspace: %w", ErrNotFound)
	}
	if len(matches) != 1 {
		return 0, fmt.Errorf("resolve workspace: %w", ErrAmbiguousWorkspace)
	}
	if !matches[0].validIndex() {
		return 0, fmt.Errorf("resolve workspace: %w", ErrWorkspaceMetadata)
	}
	return matches[0].Index, nil
}

// ListWorkspaces retrieves all workspaces for a client.
func (s *Session) ListWorkspaces(ctx context.Context, clientID string) ([]Workspace, error) {
	var workspaces []Workspace
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("clients/%s/workspaces", clientID), nil, &workspaces, nil, false); err != nil {
		return nil, fmt.Errorf("ListWorkspaces: %w", err)
	}
	return workspaces, nil
}

// GetWorkspace retrieves a single workspace by selector (index/name/user).
func (s *Session) GetWorkspace(ctx context.Context, clientID string, selector WorkspaceSelector) (*Workspace, error) {
	idx, err := s.resolveWorkspaceIndex(ctx, clientID, selector)
	if err != nil {
		return nil, err
	}
	var ws Workspace
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("clients/%s/workspaces/%d", clientID, idx), nil, &ws, nil, false); err != nil {
		return nil, fmt.Errorf("GetWorkspace: %w", err)
	}
	if !ws.validIndex() {
		return nil, fmt.Errorf("GetWorkspace: %w", ErrWorkspaceMetadata)
	}
	if ws.Index != idx || (selector.Name != nil && ws.WorkspaceName != *selector.Name) || (selector.User != nil && ws.User != *selector.User) {
		return nil, fmt.Errorf("GetWorkspace: %w", ErrWorkspaceChanged)
	}
	return &ws, nil
}

// UpdateWorkspace updates workspace parameters.
func (s *Session) UpdateWorkspace(ctx context.Context, clientID string, selector WorkspaceSelector, req any) (*Workspace, error) {
	idx, err := s.resolveWorkspaceIndex(ctx, clientID, selector)
	if err != nil {
		return nil, err
	}
	var ws Workspace
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("clients/%s/workspaces/%d", clientID, idx), req, &ws, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateWorkspace: %w", err)
	}
	return &ws, nil
}

// ToggleWorkspaceInfoPanel toggles the info_panel_visible state.
func (s *Session) ToggleWorkspaceInfoPanel(ctx context.Context, clientID string, selector WorkspaceSelector) error {
	ws, err := s.GetWorkspace(ctx, clientID, selector)
	if err != nil {
		return err
	}
	newVal := !ws.InfoPanelVisible
	_, err = s.UpdateWorkspace(ctx, clientID, selector, UpdateWorkspaceRequest{InfoPanelVisible: &newVal})
	return err
}

// ToggleWorkspacePinned toggles the pinned state.
func (s *Session) ToggleWorkspacePinned(ctx context.Context, clientID string, selector WorkspaceSelector) error {
	ws, err := s.GetWorkspace(ctx, clientID, selector)
	if err != nil {
		return err
	}
	newVal := !ws.Pinned
	_, err = s.UpdateWorkspace(ctx, clientID, selector, UpdateWorkspaceRequest{Pinned: &newVal})
	return err
}

// SetWorkspaceViewport frames an explicit CANVAS-pixel region or a widget's
// rendered bounds. CanvasID is required in both modes. Notes require an explicit
// NotePadding model. For raw wire rectangles use UpdateWorkspace instead.
func SetWorkspaceViewport(ctx context.Context, getter WorkspaceWidgetGetter, apiClient *Session, clientID string, selector WorkspaceSelector, opts SetViewportOptions) error {
	if apiClient == nil || opts.CanvasID == nil || *opts.CanvasID == "" {
		return fmt.Errorf("SetWorkspaceViewport: %w: session/canvas required", ErrInvalidRequest)
	}
	anyRect := opts.X != nil || opts.Y != nil || opts.Width != nil || opts.Height != nil
	fullRect := opts.X != nil && opts.Y != nil && opts.Width != nil && opts.Height != nil
	if (opts.WidgetID != nil && anyRect) || (opts.WidgetID == nil && !fullRect) {
		return fmt.Errorf("SetWorkspaceViewport: %w: choose widget OR complete canvas rectangle", ErrInvalidRequest)
	}
	ctx = apiClient.freezeAuthority(ctx)
	ws, err := apiClient.GetWorkspace(ctx, clientID, selector)
	if err != nil {
		return err
	}
	if ws.CanvasID != *opts.CanvasID {
		return fmt.Errorf("SetWorkspaceViewport: %w", ErrWorkspaceChanged)
	}
	var region Rectangle
	if opts.WidgetID != nil {
		nodes, err := widgetAncestry(ctx, getter, *opts.CanvasID, *opts.WidgetID)
		if err != nil {
			return err
		}
		region, err = WidgetCanvasBounds(*opts.WidgetID, nodes, GeometryModel{CanvasID: *opts.CanvasID, NotePadding: opts.NotePadding})
		if err != nil {
			return err
		}
		margin := opts.Margin
		if margin == 0 {
			margin = 20
		}
		if !finite(margin) || margin < 0 {
			return fmt.Errorf("SetWorkspaceViewport: %w: margin", ErrInvalidGeometry)
		}
		region.X -= margin
		region.Y -= margin
		region.Width += 2 * margin
		region.Height += 2 * margin
	} else {
		region = Rectangle{X: *opts.X, Y: *opts.Y, Width: *opts.Width, Height: *opts.Height}
	}
	return apiClient.frameWorkspaceRegion(ctx, clientID, *ws, region)
}

// OpenCanvasOnWorkspace submits one open command and waits for matching native
// readiness, not merely canvas_id while loading. It does not authenticate the
// native client. Polling/camera work pins the resolved index and initial actor.
func (s *Session) OpenCanvasOnWorkspace(ctx context.Context, clientID string, selector WorkspaceSelector, opts OpenCanvasOptions) (err error) {
	if opts.CanvasID == "" || (opts.CenterX == nil) != (opts.CenterY == nil) || (opts.WidgetID != nil && opts.CenterX != nil) {
		return fmt.Errorf("OpenCanvasOnWorkspace: %w: canvas and unambiguous centre required", ErrInvalidRequest)
	}
	if opts.CenterX != nil && !finite(*opts.CenterX, *opts.CenterY) {
		return fmt.Errorf("OpenCanvasOnWorkspace: %w", ErrInvalidGeometry)
	}
	ctx = s.freezeAuthority(ctx)
	idx, err := s.resolveWorkspaceIndex(ctx, clientID, selector)
	if err != nil {
		return err
	}
	selector = WorkspaceSelector{Index: &idx}
	payload := map[string]any{"canvas_id": opts.CanvasID}
	if opts.ServerID != "" {
		payload["server_id"] = opts.ServerID
	}
	if opts.UserEmail != "" {
		payload["user_email"] = opts.UserEmail
	}
	stage, acknowledged := "open", false
	defer func() {
		if err != nil {
			var accepted *AcceptedResponseError
			err = &OpenCanvasError{Stage: stage, CommandAcknowledged: acknowledged || errors.As(err, &accepted), Err: err}
		}
	}()
	if err = s.doRequest(ctx, http.MethodPost, fmt.Sprintf("clients/%s/workspaces/%d/open-canvas", clientID, idx), payload, nil, nil, false); err != nil {
		return fmt.Errorf("OpenCanvasOnWorkspace: %w", err)
	}
	stage, acknowledged = "readiness", true
	timeout, interval := 10*time.Second, 200*time.Millisecond
	if opts.PollTimeout > 0 {
		timeout = opts.PollTimeout
	}
	if opts.PollInterval > 0 {
		interval = opts.PollInterval
	}
	poll, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var ws *Workspace
	for {
		ws, err = s.GetWorkspace(poll, clientID, selector)
		if err != nil {
			return fmt.Errorf("OpenCanvasOnWorkspace: command accepted, readiness unconfirmed: %w", err)
		}
		if ws.CanvasID == opts.CanvasID && ws.WorkspaceState == "open" && (opts.ServerID == "" || ws.ServerID == opts.ServerID) && (opts.UserEmail == "" || ws.User == opts.UserEmail) {
			break
		}
		if err = waitForRetry(poll, interval); err != nil {
			return fmt.Errorf("OpenCanvasOnWorkspace: command accepted, readiness unconfirmed: %w", err)
		}
	}
	cancel()
	stage = "camera"
	if opts.CenterX != nil {
		if ws.Size == nil || ws.ViewRectangle == nil {
			return fmt.Errorf("OpenCanvasOnWorkspace: open, but %w", ErrWorkspaceMetadata)
		}
		visible, err := VisibleCanvasRegion(*ws.ViewRectangle, *ws.Size)
		if err != nil {
			return err
		}
		region := Rectangle{X: *opts.CenterX - visible.Width/2, Y: *opts.CenterY - visible.Height/2, Width: visible.Width, Height: visible.Height}
		return s.frameWorkspaceRegion(ctx, clientID, *ws, region)
	}
	if opts.WidgetID != nil {
		return SetWorkspaceViewport(ctx, s, s, clientID, selector, SetViewportOptions{CanvasID: &opts.CanvasID, WidgetID: opts.WidgetID, NotePadding: opts.NotePadding})
	}
	return nil
}
