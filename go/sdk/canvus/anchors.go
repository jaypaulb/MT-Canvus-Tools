package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// ListAnchors retrieves all anchors for a canvas.
func (s *Session) ListAnchors(ctx context.Context, canvasID string) ([]Anchor, error) {
	var anchors []Anchor
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/anchors", canvasID), nil, &anchors, nil, false); err != nil {
		return nil, fmt.Errorf("ListAnchors: %w", err)
	}
	return anchors, nil
}

// GetAnchor retrieves an anchor by ID.
func (s *Session) GetAnchor(ctx context.Context, canvasID, anchorID string) (*Anchor, error) {
	var anchor Anchor
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/anchors/%s", canvasID, anchorID), nil, &anchor, nil, false); err != nil {
		return nil, fmt.Errorf("GetAnchor: %w", err)
	}
	return &anchor, nil
}

// CreateAnchor creates a new anchor on a canvas.
func (s *Session) CreateAnchor(ctx context.Context, canvasID string, req any) (*Anchor, error) {
	var anchor Anchor
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/anchors", canvasID), req, &anchor, nil, false); err != nil {
		return nil, fmt.Errorf("CreateAnchor: %w", err)
	}
	return &anchor, nil
}

// UpdateAnchor updates an anchor by ID.
func (s *Session) UpdateAnchor(ctx context.Context, canvasID, anchorID string, req any) (*Anchor, error) {
	var anchor Anchor
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvases/%s/anchors/%s", canvasID, anchorID), req, &anchor, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateAnchor: %w", err)
	}
	return &anchor, nil
}

// DeleteAnchor deletes an anchor by ID.
func (s *Session) DeleteAnchor(ctx context.Context, canvasID, anchorID string) error {
	return s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("canvases/%s/anchors/%s", canvasID, anchorID), nil, nil, nil, false)
}
