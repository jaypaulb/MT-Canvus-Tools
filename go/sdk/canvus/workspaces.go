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
	GetWidget(ctx context.Context, clientID, widgetID string) (*Widget, error)
}

func (s *Session) resolveWorkspaceIndex(ctx context.Context, clientID string, selector WorkspaceSelector) (int, error) {
	if selector.Index != nil {
		return *selector.Index, nil
	}
	workspaces, err := s.ListWorkspaces(ctx, clientID)
	if err != nil {
		return 0, err
	}
	if selector.Name != nil {
		for _, ws := range workspaces {
			if ws.WorkspaceName == *selector.Name {
				return ws.Index, nil
			}
		}
		return 0, fmt.Errorf("workspace with name %q not found", *selector.Name)
	}
	if selector.User != nil {
		for _, ws := range workspaces {
			if ws.User == *selector.User {
				return ws.Index, nil
			}
		}
		return 0, fmt.Errorf("workspace for user %q not found", *selector.User)
	}
	return 0, nil
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

// SetWorkspaceViewport sets the workspace viewport either explicitly or by
// centering on a widget.
func SetWorkspaceViewport(ctx context.Context, getter WorkspaceWidgetGetter, apiClient *Session, clientID string, selector WorkspaceSelector, opts SetViewportOptions) error {
	var rect *Rectangle
	switch {
	case opts.WidgetID != nil:
		widget, err := getter.GetWidget(ctx, clientID, *opts.WidgetID)
		if err != nil {
			return err
		}
		margin := opts.Margin
		if margin == 0 {
			margin = 20
		}
		rect = &Rectangle{
			X:      widget.Location.X - margin,
			Y:      widget.Location.Y - margin,
			Width:  widget.Size.Width + 2*margin,
			Height: widget.Size.Height + 2*margin,
		}
	case opts.X != nil && opts.Y != nil && opts.Width != nil && opts.Height != nil:
		rect = &Rectangle{X: *opts.X, Y: *opts.Y, Width: *opts.Width, Height: *opts.Height}
	default:
		return errors.New("must provide either WidgetID or all of X, Y, Width, Height")
	}
	_, err := apiClient.UpdateWorkspace(ctx, clientID, selector, UpdateWorkspaceRequest{ViewRectangle: rect})
	return err
}

// OpenCanvasOnWorkspace opens a canvas on a client workspace and optionally
// centers the viewport.
func (s *Session) OpenCanvasOnWorkspace(ctx context.Context, clientID string, selector WorkspaceSelector, opts OpenCanvasOptions) error {
	idx, err := s.resolveWorkspaceIndex(ctx, clientID, selector)
	if err != nil {
		return err
	}
	payload := map[string]any{"canvas_id": opts.CanvasID}
	if opts.ServerID != "" {
		payload["server_id"] = opts.ServerID
	}
	if opts.UserEmail != "" {
		payload["user_email"] = opts.UserEmail
	}
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("clients/%s/workspaces/%d/open-canvas", clientID, idx), payload, nil, nil, false); err != nil {
		return fmt.Errorf("OpenCanvasOnWorkspace: %w", err)
	}

	timeout := 10 * time.Second
	interval := 200 * time.Millisecond
	if opts.PollTimeout > 0 {
		timeout = opts.PollTimeout
	}
	if opts.PollInterval > 0 {
		interval = opts.PollInterval
	}
	var ws *Workspace
	start := time.Now()
	for {
		ws, err = s.GetWorkspace(ctx, clientID, selector)
		if err != nil {
			return fmt.Errorf("OpenCanvasOnWorkspace: polling GetWorkspace failed: %w", err)
		}
		if ws.CanvasID == opts.CanvasID {
			break
		}
		if time.Since(start) > timeout {
			return fmt.Errorf("OpenCanvasOnWorkspace: timed out waiting for canvas ID %s (last seen: %s)", opts.CanvasID, ws.CanvasID)
		}
		time.Sleep(interval)
	}

	if opts.CenterX != nil && opts.CenterY != nil {
		rect := &Rectangle{X: *opts.CenterX, Y: *opts.CenterY, Width: ws.Size.Width, Height: ws.Size.Height}
		if _, err = s.UpdateWorkspace(ctx, clientID, selector, UpdateWorkspaceRequest{ViewRectangle: rect}); err != nil {
			return fmt.Errorf("OpenCanvasOnWorkspace: failed to set viewport: %w", err)
		}
	} else if opts.WidgetID != nil {
		if err := SetWorkspaceViewport(ctx, s, s, clientID, selector, SetViewportOptions{WidgetID: opts.WidgetID}); err != nil {
			return fmt.Errorf("OpenCanvasOnWorkspace: failed to center on widget: %w", err)
		}
	}
	return nil
}
