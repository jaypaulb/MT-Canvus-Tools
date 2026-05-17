package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// ListRDPConnections retrieves all RDP connection widgets for a canvas.
//
// Per changelog §2, RDP Connection widgets cannot be created via the API —
// only from the Canvus desktop client. This SDK file exposes only
// GET / PATCH / DELETE; CreateWidget("rdp_connection", ...) returns
// ErrWidgetTypeNotCreatable.
//
// Per changelog §3 (unresolved), the C++ source serializes field names with
// hyphens (`host-id`, `content-id`, `connection-name`, `host-site`); we use
// that form in the RDPConnection struct since the C++ code is the runtime
// source of truth. Verify via:
//
//	curl -s -H "Private-Token: $TOKEN" \
//	  https://<server>/api/v1/canvases/<id>/rdp-connections | jq
func (s *Session) ListRDPConnections(ctx context.Context, canvasID string) ([]RDPConnection, error) {
	var widgets []RDPConnection
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/rdp-connections", canvasID), nil, &widgets, nil, false); err != nil {
		return nil, fmt.Errorf("ListRDPConnections: %w", err)
	}
	return widgets, nil
}

// GetRDPConnection retrieves a single RDP connection widget by ID.
func (s *Session) GetRDPConnection(ctx context.Context, canvasID, widgetID string) (*RDPConnection, error) {
	var widget RDPConnection
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/rdp-connections/%s", canvasID, widgetID), nil, &widget, nil, false); err != nil {
		return nil, fmt.Errorf("GetRDPConnection: %w", err)
	}
	return &widget, nil
}

// UpdateRDPConnection updates an RDP connection widget.
func (s *Session) UpdateRDPConnection(ctx context.Context, canvasID, widgetID string, req any) (*RDPConnection, error) {
	var widget RDPConnection
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvases/%s/rdp-connections/%s", canvasID, widgetID), req, &widget, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateRDPConnection: %w", err)
	}
	return &widget, nil
}

// DeleteRDPConnection deletes an RDP connection widget.
func (s *Session) DeleteRDPConnection(ctx context.Context, canvasID, widgetID string) error {
	return s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("canvases/%s/rdp-connections/%s", canvasID, widgetID), nil, nil, nil, false)
}
