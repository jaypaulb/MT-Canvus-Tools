package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// CanvusSession wraps a canvus.Session and provides the small subset of
// Canvus operations the translator needs.
//
// Real SDK surface used:
//   - ListNotes  → to fetch all notes from the canvas (typed, includes text).
//   - UpdateNote → to patch the text of each note in place.
type CanvusSession struct {
	session *canvus.Session
}

// NewCanvusSession creates a Canvus SDK session from the supplied Config.
// The SDK's WithAPIKey option installs the Private-Token header and allows
// self-signed TLS (common on Canvus dev servers).
func NewCanvusSession(cfg *Config) (*CanvusSession, error) {
	scfg := &canvus.SessionConfig{
		BaseURL: cfg.CanvusAPIURL,
		// The translation batch may take several minutes on a large canvas;
		// keep the session alive for individual calls.
		RequestTimeout: 2 * time.Minute,
	}
	s := canvus.NewSession(scfg, canvus.WithAPIKey(cfg.CanvusAPIKey))
	return &CanvusSession{session: s}, nil
}

// NoteRef is a minimal note record containing only the fields the translator
// needs. Using a concrete type avoids the old map[string]interface{} pattern
// and makes field access explicit.
type NoteRef struct {
	ID   string
	Text string
}

// ListNotes fetches all notes from canvasID via the SDK's typed ListNotes
// helper and maps them to NoteRef to avoid leaking SDK types into the handler.
func (cs *CanvusSession) ListNotes(ctx context.Context, canvasID string) ([]NoteRef, error) {
	sdkNotes, err := cs.session.ListNotes(ctx, canvasID)
	if err != nil {
		return nil, fmt.Errorf("ListNotes: %w", err)
	}

	notes := make([]NoteRef, len(sdkNotes))
	for i, n := range sdkNotes {
		notes[i] = NoteRef{ID: n.ID, Text: n.Text}
	}
	return notes, nil
}

// UpdateNoteText patches the text field of a single note.
func (cs *CanvusSession) UpdateNoteText(ctx context.Context, canvasID, noteID, text string) error {
	if _, err := cs.session.UpdateNote(ctx, canvasID, noteID, map[string]any{"text": text}); err != nil {
		return fmt.Errorf("UpdateNoteText(%s): %w", noteID, err)
	}
	return nil
}
