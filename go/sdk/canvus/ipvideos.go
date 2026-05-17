package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// ListIPVideos retrieves all IP-video widgets for a canvas.
//
// Per changelog §2, IP Video widgets cannot be created via the API — only
// from the Canvus desktop client. This SDK file therefore exposes only
// GET / PATCH / DELETE; CreateWidget("ip_video", ...) returns ErrWidgetTypeNotCreatable.
func (s *Session) ListIPVideos(ctx context.Context, canvasID string) ([]IPVideo, error) {
	var widgets []IPVideo
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/ip-videos", canvasID), nil, &widgets, nil, false); err != nil {
		return nil, fmt.Errorf("ListIPVideos: %w", err)
	}
	return widgets, nil
}

// GetIPVideo retrieves a single IP-video widget by ID.
func (s *Session) GetIPVideo(ctx context.Context, canvasID, widgetID string) (*IPVideo, error) {
	var widget IPVideo
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/ip-videos/%s", canvasID, widgetID), nil, &widget, nil, false); err != nil {
		return nil, fmt.Errorf("GetIPVideo: %w", err)
	}
	return &widget, nil
}

// UpdateIPVideo updates an IP-video widget.
func (s *Session) UpdateIPVideo(ctx context.Context, canvasID, widgetID string, req any) (*IPVideo, error) {
	var widget IPVideo
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvases/%s/ip-videos/%s", canvasID, widgetID), req, &widget, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateIPVideo: %w", err)
	}
	return &widget, nil
}

// DeleteIPVideo deletes an IP-video widget.
func (s *Session) DeleteIPVideo(ctx context.Context, canvasID, widgetID string) error {
	return s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("canvases/%s/ip-videos/%s", canvasID, widgetID), nil, nil, nil, false)
}
