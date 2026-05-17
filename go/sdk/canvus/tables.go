package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// CreateTableRequest is the body for CreateTable. Per changelog §4 the SDK
// deliberately omits `column_widths` and `row_heights`; the server never
// serializes them. Per changelog §5 grid_size is immutable post-create.
type CreateTableRequest struct {
	Title    string    `json:"title,omitempty"`
	GridSize *GridSize `json:"grid_size,omitempty"`
	Location *Point    `json:"location,omitempty"`
	Size     *Size     `json:"size,omitempty"`
	ParentID string    `json:"parent_id,omitempty"`
}

// ListTables retrieves all table widgets for a canvas.
func (s *Session) ListTables(ctx context.Context, canvasID string) ([]Table, error) {
	var tables []Table
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/tables", canvasID), nil, &tables, nil, false); err != nil {
		return nil, fmt.Errorf("ListTables: %w", err)
	}
	return tables, nil
}

// GetTable retrieves a single table widget by ID.
func (s *Session) GetTable(ctx context.Context, canvasID, tableID string) (*Table, error) {
	var table Table
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/tables/%s", canvasID, tableID), nil, &table, nil, false); err != nil {
		return nil, fmt.Errorf("GetTable: %w", err)
	}
	return &table, nil
}

// CreateTable creates a new table widget. req may be a CreateTableRequest or
// a map[string]any. To clone an existing table from another canvas, supply
// `source_canvas_id` and `source_widget_id` (see CloneWidget for a helper).
func (s *Session) CreateTable(ctx context.Context, canvasID string, req any) (*Table, error) {
	var table Table
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/tables", canvasID), req, &table, nil, false); err != nil {
		return nil, fmt.Errorf("CreateTable: %w", err)
	}
	return &table, nil
}

// UpdateTable updates a table widget.
//
// Per changelog §5, `grid_size` is silently ignored by the server on PATCH.
// We strip it from map payloads before sending and emit a single-shot warning
// so callers find out about it once during local development.
func (s *Session) UpdateTable(ctx context.Context, canvasID, tableID string, req any) (*Table, error) {
	if m, ok := req.(map[string]any); ok {
		if _, present := m["grid_size"]; present {
			warnOnce(WarningTableGridSizeImmutable)
			delete(m, "grid_size")
		}
	}
	var table Table
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvases/%s/tables/%s", canvasID, tableID), req, &table, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateTable: %w", err)
	}
	return &table, nil
}

// DeleteTable deletes a table widget.
func (s *Session) DeleteTable(ctx context.Context, canvasID, tableID string) error {
	return s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("canvases/%s/tables/%s", canvasID, tableID), nil, nil, nil, false)
}

// ListTableCells lists all cells in a table widget.
func (s *Session) ListTableCells(ctx context.Context, canvasID, tableID string) ([]TableCell, error) {
	var cells []TableCell
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/tables/%s/cells", canvasID, tableID), nil, &cells, nil, false); err != nil {
		return nil, fmt.Errorf("ListTableCells: %w", err)
	}
	return cells, nil
}
