package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// Per Phase 3 Go work item #5, paths use `color-presets` (hyphenated) — the
// canonical spec form. The legacy SDK's no-hyphen path is no longer exposed.
//
// The per-name CRUD methods (GetColorPreset, CreateColorPreset, …) target
// `canvases/{id}/color-presets/{name}` which is NOT in the spec; they are
// retained for backwards compat but will be removed once a downstream audit
// confirms zero callers (see MIGRATION-NOTES).

// GetColorPresets retrieves the color presets for a canvas (typed struct view).
func (s *Session) GetColorPresets(ctx context.Context, canvasID string) (*ColorPresets, error) {
	var presets ColorPresets
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/color-presets", canvasID), nil, &presets, nil, false); err != nil {
		return nil, fmt.Errorf("GetColorPresets: %w", err)
	}
	return &presets, nil
}

// PatchColorPresets updates the color presets for a canvas.
// req can be *ColorPresets or map[string]any.
func (s *Session) PatchColorPresets(ctx context.Context, canvasID string, req any) (*ColorPresets, error) {
	var updated ColorPresets
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvases/%s/color-presets", canvasID), req, &updated, nil, false); err != nil {
		return nil, fmt.Errorf("PatchColorPresets: %w", err)
	}
	return &updated, nil
}

// ListColorPresets returns the color-preset names for a canvas. Legacy shape
// expected to disappear once consumers migrate to GetColorPresets.
func (s *Session) ListColorPresets(ctx context.Context, canvasID string) ([]ColorPreset, error) {
	var presets []ColorPreset
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/color-presets", canvasID), nil, &presets, nil, false); err != nil {
		return nil, fmt.Errorf("ListColorPresets: %w", err)
	}
	return presets, nil
}

// GetColorPreset retrieves a single named preset. Non-spec path; retained for
// backwards compat.
func (s *Session) GetColorPreset(ctx context.Context, canvasID, name string) (*ColorPreset, error) {
	var preset ColorPreset
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/color-presets/%s", canvasID, name), nil, &preset, nil, false); err != nil {
		return nil, fmt.Errorf("GetColorPreset: %w", err)
	}
	return &preset, nil
}

// CreateColorPreset creates a single named preset (non-spec path).
func (s *Session) CreateColorPreset(ctx context.Context, canvasID string, req any) (*ColorPreset, error) {
	var preset ColorPreset
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("canvases/%s/color-presets", canvasID), req, &preset, nil, false); err != nil {
		return nil, fmt.Errorf("CreateColorPreset: %w", err)
	}
	return &preset, nil
}

// UpdateColorPreset updates a single named preset (non-spec path).
func (s *Session) UpdateColorPreset(ctx context.Context, canvasID, name string, req any) (*ColorPreset, error) {
	var preset ColorPreset
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvases/%s/color-presets/%s", canvasID, name), req, &preset, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateColorPreset: %w", err)
	}
	return &preset, nil
}

// DeleteColorPreset deletes a single named preset (non-spec path).
func (s *Session) DeleteColorPreset(ctx context.Context, canvasID, name string) error {
	return s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("canvases/%s/color-presets/%s", canvasID, name), nil, nil, nil, false)
}
