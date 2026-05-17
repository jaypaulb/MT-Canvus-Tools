package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// ListVideos retrieves all videos for a canvas.
func (s *Session) ListVideos(ctx context.Context, canvasID string) ([]Video, error) {
	var videos []Video
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/videos", canvasID), nil, &videos, nil, false); err != nil {
		return nil, fmt.Errorf("ListVideos: %w", err)
	}
	return videos, nil
}

// GetVideo retrieves a single video by ID.
func (s *Session) GetVideo(ctx context.Context, canvasID, videoID string) (*Video, error) {
	var video Video
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/videos/%s", canvasID, videoID), nil, &video, nil, false); err != nil {
		return nil, fmt.Errorf("GetVideo: %w", err)
	}
	return &video, nil
}

// DownloadVideo downloads a video file by ID.
func (s *Session) DownloadVideo(ctx context.Context, canvasID, videoID string) ([]byte, error) {
	var data []byte
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/videos/%s/download", canvasID, videoID), nil, &data, nil, true); err != nil {
		return nil, fmt.Errorf("DownloadVideo: %w", err)
	}
	return data, nil
}

// CreateVideo creates a new video on a canvas. The body must be a multipart
// POST with a 'json' part and a 'data' part.
func (s *Session) CreateVideo(ctx context.Context, canvasID string, multipartBody any, contentType string) (*Video, error) {
	var video Video
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/videos", canvasID), multipartBody, &video, nil, false, contentType); err != nil {
		return nil, fmt.Errorf("CreateVideo: %w", err)
	}
	return &video, nil
}

// UpdateVideo updates a video by ID.
//
// API limitation: size changes via PATCH do not preserve aspect ratio. See
// WarningVideoAspectRatioNotPreserved.
func (s *Session) UpdateVideo(ctx context.Context, canvasID, videoID string, req any) (*Video, error) {
	warnOnce(WarningVideoAspectRatioNotPreserved)
	var video Video
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvases/%s/videos/%s", canvasID, videoID), req, &video, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateVideo: %w", err)
	}
	return &video, nil
}

// DeleteVideo deletes a video by ID.
func (s *Session) DeleteVideo(ctx context.Context, canvasID, videoID string) error {
	return s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("canvases/%s/videos/%s", canvasID, videoID), nil, nil, nil, false)
}
