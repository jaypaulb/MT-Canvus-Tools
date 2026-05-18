// Package canvus wraps the new MT-Canvus-Tools SDK for note-mapper's needs.
//
// The source project had two layers: a hand-rolled HTTP client
// (internal/canvusapi) and an MCS adapter (internal/mcs) that partially used
// the old client and partially rebuilt raw HTTP calls. Both are replaced here
// by a thin wrapper around canvus.Session that exposes exactly the operations
// note-mapper needs.
//
// SDK methods used:
//   - session.ListCanvases    → replaces raw GET /canvases
//   - session.ListAnchors     → replaces raw GET /canvases/{id}/anchors
//   - session.GetAnchor       → replaces raw GET /canvases/{id}/anchors/{id}
//   - session.CreateNote      → replaces client.CreateNote
//
// GetCanvasSize is dropped: the source tried to find a "SharedCanvas" widget;
// in practice the anchor zone dimensions are all that is needed for note
// placement. Callers should use GetAnchor to obtain zone bounds.
package canvus

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// CanvasInfo is a lightweight canvas summary for the UI selection list.
type CanvasInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AnchorInfo carries the fields note-mapper needs from an anchor widget.
type AnchorInfo struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Scale  float64 `json:"scale"`
}

// Client wraps a canvus.Session to provide note-mapper–specific operations.
type Client struct {
	session *canvus.Session
}

// NewClient constructs a Client using the provided Canvus API URL and API key.
func NewClient(canvusURL, apiKey string) *Client {
	cfg := &canvus.SessionConfig{BaseURL: canvusURL}
	s := canvus.NewSession(cfg, canvus.WithAPIKey(apiKey))
	return &Client{session: s}
}

// ListCanvases returns a summary list of all accessible canvases.
func (c *Client) ListCanvases(ctx context.Context) ([]CanvasInfo, error) {
	canvases, err := c.session.ListCanvases(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("ListCanvases: %w", err)
	}
	result := make([]CanvasInfo, 0, len(canvases))
	for _, cv := range canvases {
		result = append(result, CanvasInfo{ID: cv.ID, Name: cv.Name})
	}
	return result, nil
}

// ListAnchors returns all anchors for canvasID.
func (c *Client) ListAnchors(ctx context.Context, canvasID string) ([]AnchorInfo, error) {
	anchors, err := c.session.ListAnchors(ctx, canvasID)
	if err != nil {
		return nil, fmt.Errorf("ListAnchors: %w", err)
	}
	result := make([]AnchorInfo, 0, len(anchors))
	for _, a := range anchors {
		result = append(result, anchorToInfo(a))
	}
	return result, nil
}

// GetAnchor retrieves a specific anchor by ID.
func (c *Client) GetAnchor(ctx context.Context, canvasID, anchorID string) (*AnchorInfo, error) {
	a, err := c.session.GetAnchor(ctx, canvasID, anchorID)
	if err != nil {
		return nil, fmt.Errorf("GetAnchor: %w", err)
	}
	info := anchorToInfo(*a)
	return &info, nil
}

// CreateNote creates a note on canvasID with the given payload.
// payload must be JSON-serialisable (e.g. mapping.NoteForAPI or map[string]any).
func (c *Client) CreateNote(ctx context.Context, canvasID string, payload any) (*canvus.Note, error) {
	note, err := c.session.CreateNote(ctx, canvasID, payload)
	if err != nil {
		return nil, fmt.Errorf("CreateNote: %w", err)
	}
	return note, nil
}

// anchorToInfo converts a SDK Anchor to AnchorInfo, extracting the flat
// fields needed for spatial mapping.
func anchorToInfo(a canvus.Anchor) AnchorInfo {
	info := AnchorInfo{
		ID:    a.ID,
		Name:  a.AnchorName,
		Scale: a.Scale,
	}
	if a.Location != nil {
		info.X = a.Location.X
		info.Y = a.Location.Y
	}
	if a.Size != nil {
		info.Width = a.Size.Width
		info.Height = a.Size.Height
	}
	return info
}
