// Package qa drives the persona Q&A workflow: detecting a new question,
// asking each persona, laying out the responses, drawing connectors, and
// grouping everything in a fresh anchor.
package qa

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/canvasx"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// Helper note titles that the workflow recognizes when scanning the canvas.
const (
	HelperTitleQuestion = "Helper: Please enter a question for this note"
	HelperTitlePersonas = "Helper: Generating personas, please wait..."
	HelperTitleTimeout  = "Question Wait Timed Out"
)

// HelperTracker remembers the helper note ID associated with each Qnote so
// it can be deleted once processing finishes.
type HelperTracker struct {
	mu    sync.Mutex
	byQID map[string]string
}

// NewHelperTracker returns an empty tracker.
func NewHelperTracker() *HelperTracker {
	return &HelperTracker{byQID: make(map[string]string)}
}

// Set records the helper-note ID for qnoteID.
func (h *HelperTracker) Set(qnoteID, helperID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.byQID[qnoteID] = helperID
}

// Take returns the helper-note ID for qnoteID and removes it from the tracker.
func (h *HelperTracker) Take(qnoteID string) (string, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	id, ok := h.byQID[qnoteID]
	if ok {
		delete(h.byQID, qnoteID)
	}
	return id, ok
}

// findHelperNote scans the canvas for an existing helper note matching one
// of the well-known titles. Returns ("", nil) when none is found.
func findHelperNote(ctx context.Context, s *canvus.Session, canvasID, title string) (string, error) {
	notes, err := s.ListNotes(ctx, canvasID)
	if err != nil {
		return "", err
	}
	for _, n := range notes {
		if n.Title == title {
			return n.ID, nil
		}
	}
	return "", nil
}

// createOrUpdateHelper ensures a helper note with the given title exists next
// to qnote, updating it (if found) or creating it (if not). Returns the
// helper-note ID.
func createOrUpdateHelper(ctx context.Context, s *canvus.Session, canvasID, title, text, color string, qnote *canvus.Note) (string, error) {
	id, err := findHelperNote(ctx, s, canvasID, title)
	if err != nil {
		return "", err
	}
	if id != "" {
		if _, err := s.UpdateNote(ctx, canvasID, id, map[string]any{"text": text}); err != nil {
			slog.Warn("qa.helper: update existing helper failed", "id", id, "err", err)
		}
		return id, nil
	}

	if qnote.Location == nil || qnote.Size == nil {
		return "", fmt.Errorf("createOrUpdateHelper: qnote missing location/size")
	}
	hX, hY, hW, hH := canvasx.HelperNotePosition(qnote.Location.X, qnote.Location.Y,
		qnote.Size.Width, qnote.Size.Height)
	helper, err := s.CreateNote(ctx, canvasID, map[string]any{
		"title":            title,
		"text":             text,
		"location":         map[string]any{"x": hX, "y": hY},
		"size":             map[string]any{"width": hW, "height": hH},
		"background_color": color,
	})
	if err != nil {
		return "", fmt.Errorf("createOrUpdateHelper: create note: %w", err)
	}
	if _, err := s.CreateConnector(ctx, canvasID,
		canvasx.BuildConnectorPayload(helper.ID, qnote.ID)); err != nil {
		slog.Warn("qa.helper: create connector failed", "err", err)
	}
	return helper.ID, nil
}

// DeleteHelper removes the tracked helper note for qnoteID.
func DeleteHelper(ctx context.Context, s *canvus.Session, canvasID string, tracker *HelperTracker, qnoteID string) {
	id, ok := tracker.Take(qnoteID)
	if !ok {
		return
	}
	if err := s.DeleteNote(ctx, canvasID, id); err != nil {
		slog.Warn("qa.helper: delete failed", "id", id, "err", err)
		return
	}
	slog.Info("qa.helper: deleted", "helper_id", id, "qnote_id", qnoteID)
}

// extractQuestion strips the helper prefix (everything before "-->") and
// trailing "Please wait" notice from the raw note text.
func extractQuestion(text string) string {
	q := text
	if idx := strings.Index(q, "-->"); idx != -1 {
		q = q[idx+3:]
	}
	q = strings.TrimSpace(strings.Split(q, "Please wait")[0])
	return q
}
