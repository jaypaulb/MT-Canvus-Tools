package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// ListCanvases retrieves all canvases. If filter is non-nil, results are
// filtered client-side.
func (s *Session) ListCanvases(ctx context.Context, filter *Filter) ([]Canvas, error) {
	var canvases []Canvas
	if err := s.doRequest(ctx, http.MethodGet, "canvases", nil, &canvases, nil, false); err != nil {
		return nil, fmt.Errorf("ListCanvases: %w", err)
	}
	if filter != nil {
		canvases = FilterSlice(canvases, filter)
	}
	return canvases, nil
}

// GetCanvas retrieves a single canvas by ID.
func (s *Session) GetCanvas(ctx context.Context, id string) (*Canvas, error) {
	var canvas Canvas
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s", id), nil, &canvas, nil, false); err != nil {
		return nil, fmt.Errorf("GetCanvas: %w", err)
	}
	return &canvas, nil
}

// CreateCanvas creates a new canvas.
// req can be CreateCanvasRequest or map[string]any.
func (s *Session) CreateCanvas(ctx context.Context, req any) (*Canvas, error) {
	var canvas Canvas
	if err := s.doRequest(ctx, http.MethodPost, "canvases", req, &canvas, nil, false); err != nil {
		return nil, fmt.Errorf("CreateCanvas: %w", err)
	}
	return &canvas, nil
}

// UpdateCanvas renames or changes the mode of a canvas.
// req can be UpdateCanvasRequest or map[string]any.
func (s *Session) UpdateCanvas(ctx context.Context, id string, req any) (*Canvas, error) {
	var canvas Canvas
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvases/%s", id), req, &canvas, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateCanvas: %w", err)
	}
	return &canvas, nil
}

// DeleteCanvas permanently deletes a canvas.
func (s *Session) DeleteCanvas(ctx context.Context, id string) error {
	return s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("canvases/%s", id), nil, nil, nil, false)
}

// GetCanvasPreview downloads the preview image bytes of a canvas, if available.
func (s *Session) GetCanvasPreview(ctx context.Context, id string) ([]byte, error) {
	var preview []byte
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/preview", id), nil, &preview, nil, true); err != nil {
		return nil, fmt.Errorf("GetCanvasPreview: %w", err)
	}
	return preview, nil
}

// RestoreDemoCanvas restores a demo canvas to its last saved state.
func (s *Session) RestoreDemoCanvas(ctx context.Context, id string) error {
	return s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/restore", id), nil, nil, nil, false)
}

// SaveDemoState updates the saved demo canvas state with current changes.
func (s *Session) SaveDemoState(ctx context.Context, id string) error {
	return s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/save", id), nil, nil, nil, false)
}

// MoveCanvas moves a canvas to another folder.
func (s *Session) MoveCanvas(ctx context.Context, id string, req MoveOrCopyCanvasRequest) (*Canvas, error) {
	var canvas Canvas
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/move", id), req, &canvas, nil, false); err != nil {
		return nil, fmt.Errorf("MoveCanvas: %w", err)
	}
	return &canvas, nil
}

// CopyCanvas copies a canvas to another folder.
func (s *Session) CopyCanvas(ctx context.Context, id string, req MoveOrCopyCanvasRequest) (*Canvas, error) {
	var canvas Canvas
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/copy", id), req, &canvas, nil, false); err != nil {
		return nil, fmt.Errorf("CopyCanvas: %w", err)
	}
	return &canvas, nil
}

// TrashCanvas moves a canvas to the current user's trash folder.
// Requires a logged-in session; API-key sessions cannot infer a user ID.
func (s *Session) TrashCanvas(ctx context.Context, id string, _ string) (*Canvas, error) {
	userID := s.UserID()
	if userID == 0 {
		return nil, fmt.Errorf("TrashCanvas: user ID not set; must login first")
	}
	return s.MoveCanvas(ctx, id, MoveOrCopyCanvasRequest{FolderID: fmt.Sprintf("trash.%d", userID)})
}

// GetCanvasPermissions gets the permission overrides on a canvas.
func (s *Session) GetCanvasPermissions(ctx context.Context, id string) (*CanvasPermissions, error) {
	var perms CanvasPermissions
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/permissions", id), nil, &perms, nil, false); err != nil {
		return nil, fmt.Errorf("GetCanvasPermissions: %w", err)
	}
	return &perms, nil
}

// SetCanvasPermissions sets permission overrides on a canvas.
func (s *Session) SetCanvasPermissions(ctx context.Context, id string, perms CanvasPermissions) (*CanvasPermissions, error) {
	var updated CanvasPermissions
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/permissions", id), perms, &updated, nil, false); err != nil {
		return nil, fmt.Errorf("SetCanvasPermissions: %w", err)
	}
	return &updated, nil
}
