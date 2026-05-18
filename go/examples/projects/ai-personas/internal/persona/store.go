// Package persona owns the persona generation workflow: extracting business
// context, calling Gemini, drawing persona notes inside the Personas anchor,
// and uploading DALL-E portraits.
package persona

import (
	"sync"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/atom"
)

// Store records the persona-note IDs created for each Qnote so subsequent
// Q&A workflows can fetch them back into atom.Persona values.
type Store struct {
	mu    sync.Mutex
	byQID map[string][]string
}

// NewStore returns an empty Store.
func NewStore() *Store {
	return &Store{byQID: make(map[string][]string)}
}

// Set records the persona-note IDs for qnoteID, replacing any prior value.
func (s *Store) Set(qnoteID string, noteIDs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byQID[qnoteID] = append([]string(nil), noteIDs...)
}

// Get returns the persona note IDs for a qnoteID and whether any were stored.
func (s *Store) Get(qnoteID string) ([]string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids, ok := s.byQID[qnoteID]
	if !ok {
		return nil, false
	}
	return append([]string(nil), ids...), true
}

// Has reports whether a persona-id list exists for the qnoteID.
func (s *Store) Has(qnoteID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.byQID[qnoteID]
	return ok
}

// FormatPersona is re-exported so callers can avoid pulling in the atom
// package directly when they only need the canonical note-text renderer.
func FormatPersona(p atom.Persona) string { return atom.FormatPersonaNote(p) }
