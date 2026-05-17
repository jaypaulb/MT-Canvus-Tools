package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// ListVideoOutputs retrieves all video outputs for a client.
func (s *Session) ListVideoOutputs(ctx context.Context, clientID string) ([]VideoOutput, error) {
	var outputs []VideoOutput
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("clients/%s/video-outputs", clientID), nil, &outputs, nil, false); err != nil {
		return nil, fmt.Errorf("ListVideoOutputs: %w", err)
	}
	return outputs, nil
}

// GetVideoOutput retrieves a single video output by ID for a client.
func (s *Session) GetVideoOutput(ctx context.Context, clientID, outputID string) (*VideoOutput, error) {
	var output VideoOutput
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("clients/%s/video-outputs/%s", clientID, outputID), nil, &output, nil, false); err != nil {
		return nil, fmt.Errorf("GetVideoOutput: %w", err)
	}
	return &output, nil
}

// SetVideoOutputSource sets the source or suspends a video output for a client.
func (s *Session) SetVideoOutputSource(ctx context.Context, clientID string, index int, req any) error {
	return s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("clients/%s/video-outputs/%d", clientID, index), req, nil, nil, false)
}

// SetVideoOutputSourceByID is the UUID variant of SetVideoOutputSource. Use
// when the server returns string IDs rather than indices for video outputs.
func (s *Session) SetVideoOutputSourceByID(ctx context.Context, clientID, outputID string, req any) error {
	return s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("clients/%s/video-outputs/%s", clientID, outputID), req, nil, nil, false)
}
