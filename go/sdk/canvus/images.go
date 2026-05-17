package canvus

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// ListImages retrieves all images for a canvas.
func (s *Session) ListImages(ctx context.Context, canvasID string) ([]Image, error) {
	var images []Image
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/images", canvasID), nil, &images, nil, false); err != nil {
		return nil, fmt.Errorf("ListImages: %w", err)
	}
	return images, nil
}

// GetImage retrieves an image by ID.
func (s *Session) GetImage(ctx context.Context, canvasID, imageID string) (*Image, error) {
	var image Image
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/images/%s", canvasID, imageID), nil, &image, nil, false); err != nil {
		return nil, fmt.Errorf("GetImage: %w", err)
	}
	return &image, nil
}

// CreateImage creates a new image on a canvas. The body must be a multipart
// POST with a 'json' part (metadata) and a 'data' part (binary).
func (s *Session) CreateImage(ctx context.Context, canvasID string, multipartBody io.Reader, contentType string) (*Image, error) {
	var image Image
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/images", canvasID), multipartBody, &image, nil, false, contentType); err != nil {
		return nil, fmt.Errorf("CreateImage: %w", err)
	}
	return &image, nil
}

// UpdateImage updates an image by ID.
//
// API limitation: size changes via PATCH do not preserve aspect ratio. See
// WarningImageAspectRatioNotPreserved.
func (s *Session) UpdateImage(ctx context.Context, canvasID, imageID string, req any) (*Image, error) {
	warnOnce(WarningImageAspectRatioNotPreserved)
	var image Image
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvases/%s/images/%s", canvasID, imageID), req, &image, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateImage: %w", err)
	}
	return &image, nil
}

// DeleteImage deletes an image by ID.
func (s *Session) DeleteImage(ctx context.Context, canvasID, imageID string) error {
	return s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("canvases/%s/images/%s", canvasID, imageID), nil, nil, nil, false)
}

// DownloadImage downloads an image file by ID.
func (s *Session) DownloadImage(ctx context.Context, canvasID, imageID string) ([]byte, error) {
	var data []byte
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/images/%s/download", canvasID, imageID), nil, &data, nil, true); err != nil {
		return nil, fmt.Errorf("DownloadImage: %w", err)
	}
	return data, nil
}
