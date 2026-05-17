package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// ListUploads lists items in a canvas's uploads-folder.
// Added per Phase 3 Go work item #7 — the spec defines both GET and POST on
// this endpoint, but the legacy SDK only implemented POST.
func (s *Session) ListUploads(ctx context.Context, canvasID string) ([]UploadItem, error) {
	var items []UploadItem
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/uploads-folder", canvasID), nil, &items, nil, false); err != nil {
		return nil, fmt.Errorf("ListUploads: %w", err)
	}
	return items, nil
}

// UploadNote uploads a note to the uploads-folder of a canvas via multipart POST.
func (s *Session) UploadNote(ctx context.Context, canvasID string, multipartBody any, contentType ...string) (*Note, error) {
	var note Note
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/uploads-folder", canvasID), multipartBody, &note, nil, false, contentType...); err != nil {
		return nil, fmt.Errorf("UploadNote: %w", err)
	}
	return &note, nil
}

// UploadAsset uploads a file asset to the uploads-folder of a canvas via multipart POST.
func (s *Session) UploadAsset(ctx context.Context, canvasID string, multipartBody any, contentType ...string) (*Asset, error) {
	var asset Asset
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/uploads-folder", canvasID), multipartBody, &asset, nil, false, contentType...); err != nil {
		return nil, fmt.Errorf("UploadAsset: %w", err)
	}
	return &asset, nil
}
