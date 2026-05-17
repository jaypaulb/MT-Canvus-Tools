package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// ListBrowsers retrieves all browser widgets for a canvas.
func (s *Session) ListBrowsers(ctx context.Context, canvasID string) ([]Browser, error) {
	var browsers []Browser
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/browsers", canvasID), nil, &browsers, nil, false); err != nil {
		return nil, fmt.Errorf("ListBrowsers: %w", err)
	}
	return browsers, nil
}

// GetBrowser retrieves a browser by ID.
func (s *Session) GetBrowser(ctx context.Context, canvasID, browserID string) (*Browser, error) {
	var browser Browser
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/browsers/%s", canvasID, browserID), nil, &browser, nil, false); err != nil {
		return nil, fmt.Errorf("GetBrowser: %w", err)
	}
	return &browser, nil
}

// CreateBrowser creates a new browser widget.
func (s *Session) CreateBrowser(ctx context.Context, canvasID string, req any) (*Browser, error) {
	var browser Browser
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/browsers", canvasID), req, &browser, nil, false); err != nil {
		return nil, fmt.Errorf("CreateBrowser: %w", err)
	}
	return &browser, nil
}

// UpdateBrowser updates a browser by ID.
func (s *Session) UpdateBrowser(ctx context.Context, canvasID, browserID string, req any) (*Browser, error) {
	var browser Browser
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvases/%s/browsers/%s", canvasID, browserID), req, &browser, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateBrowser: %w", err)
	}
	return &browser, nil
}

// DeleteBrowser deletes a browser by ID.
func (s *Session) DeleteBrowser(ctx context.Context, canvasID, browserID string) error {
	if err := s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("canvases/%s/browsers/%s", canvasID, browserID), nil, nil, nil, false); err != nil {
		return fmt.Errorf("DeleteBrowser: %w", err)
	}
	return nil
}
