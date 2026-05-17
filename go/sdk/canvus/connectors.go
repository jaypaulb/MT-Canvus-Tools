package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// ListConnectors retrieves all connectors for a canvas.
func (s *Session) ListConnectors(ctx context.Context, canvasID string) ([]Connector, error) {
	var connectors []Connector
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/connectors", canvasID), nil, &connectors, nil, false); err != nil {
		return nil, fmt.Errorf("ListConnectors: %w", err)
	}
	return connectors, nil
}

// GetConnector retrieves a connector by ID.
func (s *Session) GetConnector(ctx context.Context, canvasID, connectorID string) (*Connector, error) {
	var connector Connector
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/connectors/%s", canvasID, connectorID), nil, &connector, nil, false); err != nil {
		return nil, fmt.Errorf("GetConnector: %w", err)
	}
	return &connector, nil
}

// CreateConnector creates a new connector on a canvas.
//
// If req["src"] or req["dst"] is a widget-JSON map rather than a string ID,
// the SDK creates that widget first and uses its ID. This preserves the
// convenience helper from the legacy SDK.
func (s *Session) CreateConnector(ctx context.Context, canvasID string, req any) (*Connector, error) {
	m, ok := req.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("CreateConnector: req must be a map[string]interface{}")
	}
	resolveEnd := func(key string) (string, error) {
		v, ok := m[key]
		if !ok {
			return "", fmt.Errorf("CreateConnector: missing %s", key)
		}
		if id, ok := v.(string); ok {
			return id, nil
		}
		if widgetData, ok := v.(map[string]any); ok {
			widget, err := s.CreateWidget(ctx, canvasID, widgetData)
			if err != nil {
				return "", fmt.Errorf("CreateConnector: failed to create widget for %s: %w", key, err)
			}
			return widget.ID, nil
		}
		return "", fmt.Errorf("CreateConnector: %s must be string or widget JSON", key)
	}
	srcID, err := resolveEnd("src")
	if err != nil {
		return nil, err
	}
	dstID, err := resolveEnd("dst")
	if err != nil {
		return nil, err
	}
	m["src"] = map[string]any{"id": srcID}
	m["dst"] = map[string]any{"id": dstID}
	var connector Connector
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/connectors", canvasID), m, &connector, nil, false); err != nil {
		return nil, fmt.Errorf("CreateConnector: %w", err)
	}
	return &connector, nil
}

// UpdateConnector updates a connector by ID.
func (s *Session) UpdateConnector(ctx context.Context, canvasID, connectorID string, req any) (*Connector, error) {
	var connector Connector
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvases/%s/connectors/%s", canvasID, connectorID), req, &connector, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateConnector: %w", err)
	}
	return &connector, nil
}

// DeleteConnector deletes a connector by ID.
func (s *Session) DeleteConnector(ctx context.Context, canvasID, connectorID string) error {
	return s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("canvases/%s/connectors/%s", canvasID, connectorID), nil, nil, nil, false)
}
