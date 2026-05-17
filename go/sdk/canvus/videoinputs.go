package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// VideoInputSource represents a video input source on a client device.
type VideoInputSource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListVideoInputs lists all video-input widgets on a canvas.
func (s *Session) ListVideoInputs(ctx context.Context, canvasID string) ([]VideoInput, error) {
	warnOnce(WarningVideoInputTitleNotExposed)
	var inputs []VideoInput
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/video-inputs", canvasID), nil, &inputs, nil, false); err != nil {
		return nil, fmt.Errorf("ListVideoInputs: %w", err)
	}
	return inputs, nil
}

// GetVideoInput retrieves a single video-input widget by ID on a canvas.
func (s *Session) GetVideoInput(ctx context.Context, canvasID, inputID string) (*VideoInput, error) {
	var input VideoInput
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/video-inputs/%s", canvasID, inputID), nil, &input, nil, false); err != nil {
		return nil, fmt.Errorf("GetVideoInput: %w", err)
	}
	return &input, nil
}

// CreateVideoInput creates a video-input widget on a canvas. Payload must
// include `source` and `host-id`.
func (s *Session) CreateVideoInput(ctx context.Context, canvasID string, req any) (*VideoInput, error) {
	var input VideoInput
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/video-inputs", canvasID), req, &input, nil, false); err != nil {
		return nil, fmt.Errorf("CreateVideoInput: %w", err)
	}
	return &input, nil
}

// UpdateVideoInput updates a video-input widget on a canvas.
func (s *Session) UpdateVideoInput(ctx context.Context, canvasID, inputID string, req map[string]any) (*VideoInput, error) {
	var input VideoInput
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvases/%s/video-inputs/%s", canvasID, inputID), req, &input, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateVideoInput: %w", err)
	}
	return &input, nil
}

// DeleteVideoInput deletes a video-input widget by ID.
func (s *Session) DeleteVideoInput(ctx context.Context, canvasID, inputID string) error {
	return s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("canvases/%s/video-inputs/%s", canvasID, inputID), nil, nil, nil, false)
}

// ListClientVideoInputs lists all video-input sources on a client device.
func (s *Session) ListClientVideoInputs(ctx context.Context, clientID string) ([]VideoInputSource, error) {
	var sources []VideoInputSource
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("clients/%s/video-inputs", clientID), nil, &sources, nil, false); err != nil {
		return nil, fmt.Errorf("ListClientVideoInputs: %w", err)
	}
	return sources, nil
}

// GetClientVideoInput retrieves a single video-input source by ID on a client.
func (s *Session) GetClientVideoInput(ctx context.Context, clientID, inputID string) (*VideoInput, error) {
	var input VideoInput
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("clients/%s/video-inputs/%s", clientID, inputID), nil, &input, nil, false); err != nil {
		return nil, fmt.Errorf("GetClientVideoInput: %w", err)
	}
	return &input, nil
}
