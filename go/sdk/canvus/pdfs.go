package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// ListPDFs retrieves all PDFs for a canvas.
func (s *Session) ListPDFs(ctx context.Context, canvasID string) ([]PDF, error) {
	var pdfs []PDF
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/pdfs", canvasID), nil, &pdfs, nil, false); err != nil {
		return nil, fmt.Errorf("ListPDFs: %w", err)
	}
	return pdfs, nil
}

// GetPDF retrieves a single PDF by ID.
func (s *Session) GetPDF(ctx context.Context, canvasID, pdfID string) (*PDF, error) {
	var pdf PDF
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/pdfs/%s", canvasID, pdfID), nil, &pdf, nil, false); err != nil {
		return nil, fmt.Errorf("GetPDF: %w", err)
	}
	return &pdf, nil
}

// DownloadPDF downloads a PDF file by ID.
func (s *Session) DownloadPDF(ctx context.Context, canvasID, pdfID string) ([]byte, error) {
	var data []byte
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/pdfs/%s/download", canvasID, pdfID), nil, &data, nil, true); err != nil {
		return nil, fmt.Errorf("DownloadPDF: %w", err)
	}
	return data, nil
}

// CreatePDF creates a new PDF on a canvas. The body must be a multipart POST
// with a 'json' part and a 'data' part.
func (s *Session) CreatePDF(ctx context.Context, canvasID string, multipartBody any, contentType string) (*PDF, error) {
	var pdf PDF
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/pdfs", canvasID), multipartBody, &pdf, nil, false, contentType); err != nil {
		return nil, fmt.Errorf("CreatePDF: %w", err)
	}
	return &pdf, nil
}

// UpdatePDF updates a PDF by ID.
//
// API limitation: size changes via PATCH cause a visual disconnect. See
// WarningPDFSizeBug.
func (s *Session) UpdatePDF(ctx context.Context, canvasID, pdfID string, req any) (*PDF, error) {
	warnOnce(WarningPDFSizeBug)
	var pdf PDF
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvases/%s/pdfs/%s", canvasID, pdfID), req, &pdf, nil, false); err != nil {
		return nil, fmt.Errorf("UpdatePDF: %w", err)
	}
	return &pdf, nil
}

// DeletePDF deletes a PDF by ID.
func (s *Session) DeletePDF(ctx context.Context, canvasID, pdfID string) error {
	return s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("canvases/%s/pdfs/%s", canvasID, pdfID), nil, nil, nil, false)
}
