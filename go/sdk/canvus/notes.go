package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// ListNotes retrieves all notes for a canvas.
//
// API limitation: the `title` field is not exposed by the Canvus API; see
// WarningNoteTitleNotExposed.
func (s *Session) ListNotes(ctx context.Context, canvasID string) ([]Note, error) {
	warnOnce(WarningNoteTitleNotExposed)
	var notes []Note
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/notes", canvasID), nil, &notes, nil, false); err != nil {
		return nil, fmt.Errorf("ListNotes: %w", err)
	}
	return notes, nil
}

// GetNote retrieves a note by ID.
func (s *Session) GetNote(ctx context.Context, canvasID, noteID string) (*Note, error) {
	var note Note
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/notes/%s", canvasID, noteID), nil, &note, nil, false); err != nil {
		return nil, fmt.Errorf("GetNote: %w", err)
	}
	return &note, nil
}

// CreateNote creates a new note on a canvas. To clone a note from another
// canvas pass `source_canvas_id` and `source_widget_id` in req — see
// CloneWidget for the convenience wrapper.
func (s *Session) CreateNote(ctx context.Context, canvasID string, req any) (*Note, error) {
	var note Note
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/notes", canvasID), req, &note, nil, false); err != nil {
		return nil, fmt.Errorf("CreateNote: %w", err)
	}
	return &note, nil
}

// UpdateNote updates a note by ID.
func (s *Session) UpdateNote(ctx context.Context, canvasID, noteID string, req any) (*Note, error) {
	var note Note
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvases/%s/notes/%s", canvasID, noteID), req, &note, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateNote: %w", err)
	}
	return &note, nil
}

// DeleteNote deletes a note by ID.
func (s *Session) DeleteNote(ctx context.Context, canvasID, noteID string) error {
	return s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("canvases/%s/notes/%s", canvasID, noteID), nil, nil, nil, false)
}
