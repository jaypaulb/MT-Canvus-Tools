package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// LayoutMode selects the shared server layout algorithm.
type LayoutMode string

const (
	LayoutAuto        LayoutMode = "auto"
	LayoutLargeLeft   LayoutMode = "large_left"
	LayoutLargeRight  LayoutMode = "large_right"
	LayoutLargeCenter LayoutMode = "large_center"
	LayoutGrid        LayoutMode = "grid"
)

// LayoutZone retains the generic geometry and the five layout fields.
type LayoutZone = Widget

// LayoutZoneRequest uses pointers so omitted values differ from false, zero and [].
// An empty zone defaults to auto, [], false, .015, version 1 and 960x540.
type LayoutZoneRequest struct {
	Title         *string     `json:"title,omitempty"`
	ParentID      *string     `json:"parent_id,omitempty"`
	Location      *Point      `json:"location,omitempty"`
	Size          *Size       `json:"size,omitempty"`
	Scale         *float64    `json:"scale,omitempty"`
	Depth         *float64    `json:"depth,omitempty"`
	Pinned        *bool       `json:"pinned,omitempty"`
	LayoutMode    *LayoutMode `json:"layout_mode,omitempty"`
	SlotOrder     *[]string   `json:"slot_order,omitempty"`
	MainPromoted  *bool       `json:"main_promoted,omitempty"`
	GapFraction   *float64    `json:"gap_fraction,omitempty"`
	LayoutVersion *int        `json:"layout_version,omitempty"`
}

// ListLayoutZones retrieves all layoutZone widgets for a canvas.
func (s *Session) ListLayoutZones(ctx context.Context, canvasID string) ([]LayoutZone, error) {
	var layoutZones []LayoutZone
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/layout-zones", canvasID), nil, &layoutZones, nil, false); err != nil {
		return nil, fmt.Errorf("ListLayoutZones: %w", err)
	}
	return layoutZones, nil
}

// GetLayoutZone retrieves a single layoutZone widget by ID.
func (s *Session) GetLayoutZone(ctx context.Context, canvasID, layoutZoneID string) (*LayoutZone, error) {
	var layoutZone LayoutZone
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/layout-zones/%s", canvasID, layoutZoneID), nil, &layoutZone, nil, false); err != nil {
		return nil, fmt.Errorf("GetLayoutZone: %w", err)
	}
	return &layoutZone, nil
}

// CreateLayoutZone creates a new layoutZone widget. req may be a LayoutZoneRequest or
// a map[string]any. To clone an existing layoutZone from another canvas, supply
// `source_canvas_id` and `source_widget_id` (see CloneWidget for a helper).
func (s *Session) CreateLayoutZone(ctx context.Context, canvasID string, req any) (*LayoutZone, error) {
	var layoutZone LayoutZone
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/layout-zones", canvasID), req, &layoutZone, nil, false); err != nil {
		return nil, fmt.Errorf("CreateLayoutZone: %w", err)
	}
	return &layoutZone, nil
}

// UpdateLayoutZone updates a layoutZone widget.
func (s *Session) UpdateLayoutZone(ctx context.Context, canvasID, layoutZoneID string, req any) (*LayoutZone, error) {
	var layoutZone LayoutZone
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvases/%s/layout-zones/%s", canvasID, layoutZoneID), req, &layoutZone, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateLayoutZone: %w", err)
	}
	return &layoutZone, nil
}

// DeleteLayoutZone deletes only the zone, releasing children with their world poses intact.
func (s *Session) DeleteLayoutZone(ctx context.Context, canvasID, layoutZoneID string) error {
	return s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("canvases/%s/layout-zones/%s", canvasID, layoutZoneID), nil, nil, nil, false)
}

// SubscribeLayoutZones streams zone updates using the shared authenticated transport.
func (s *Session) SubscribeLayoutZones(ctx context.Context, canvasID string) (<-chan LayoutZone, error) {
	return subscribeStream[LayoutZone](ctx, s, fmt.Sprintf("canvases/%s/layout-zones", canvasID))
}

func (s *Session) SubscribeLayoutZone(ctx context.Context, canvasID, zoneID string) (<-chan LayoutZone, error) {
	return subscribeStream[LayoutZone](ctx, s, fmt.Sprintf("canvases/%s/layout-zones/%s", canvasID, zoneID))
}
