package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// GetCanvasBackground retrieves the background settings for a canvas.
func (s *Session) GetCanvasBackground(ctx context.Context, canvasID string) (*CanvasBackground, error) {
	var bg CanvasBackground
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/background", canvasID), nil, &bg, nil, false); err != nil {
		return nil, fmt.Errorf("GetCanvasBackground: %w", err)
	}
	return &bg, nil
}

// PatchCanvasBackground sets the background settings for a canvas (color or haze).
func (s *Session) PatchCanvasBackground(ctx context.Context, canvasID string, req any) error {
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvases/%s/background", canvasID), req, nil, nil, false); err != nil {
		return fmt.Errorf("PatchCanvasBackground: %w", err)
	}
	return nil
}

// PostCanvasBackground sets the background to an image. The body must be a
// multipart POST with a 'data' part and optional 'json' part.
func (s *Session) PostCanvasBackground(ctx context.Context, canvasID string, multipartBody any, contentType string) error {
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/background", canvasID), multipartBody, nil, nil, false, contentType); err != nil {
		return fmt.Errorf("PostCanvasBackground: %w", err)
	}
	return nil
}
