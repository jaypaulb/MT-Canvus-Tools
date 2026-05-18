package persona

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/ai"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/atom"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/business"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/canvasx"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// PersonaCount is the fixed number of personas the workflow generates per
// Business Model Canvas. The grid layout (4 columns, 4-color palette) assumes
// this value.
const PersonaCount = 4

// MinRequired is the minimum number of personas that must be successfully
// created for the workflow to be considered a partial success.
const MinRequired = 1

// Workflow holds shared dependencies for persona creation. A single instance
// is constructed at startup and reused for every Create_Personas trigger.
type Workflow struct {
	Session  *canvus.Session
	CanvasID string
	Gemini   *ai.Gemini
	OpenAI   *ai.OpenAI
	Store    *Store
}

// Create runs the full persona-generation workflow for one Qnote: extracts
// business context, generates personas via Gemini, creates persona notes in
// the Personas anchor, and uploads DALL-E portraits in the background.
//
// Returns an error if the workflow cannot meet MinRequired personas. Partial
// success (1-3 of PersonaCount) is logged but not an error.
func (w *Workflow) Create(ctx context.Context, qnoteID string) error {
	bc, err := business.Extract(ctx, w.Session, w.CanvasID)
	if err != nil {
		if bc != nil && len(bc.Missing) > 0 {
			business.CreateMissingNotesHelper(ctx, w.Session, w.CanvasID, bc.Missing, bc.PersonasAnchor)
		}
		return fmt.Errorf("persona.Create: business context: %w", err)
	}
	slog.Info("persona.Create: starting", "qnote_id", qnoteID, "context_chars", len(bc.Body))

	// Reuse existing persona notes when present (so re-triggering doesn't
	// duplicate work after a partial failure).
	existing, err := w.findExistingPersonas(ctx)
	if err != nil {
		return fmt.Errorf("persona.Create: scan existing: %w", err)
	}
	if len(existing) == PersonaCount {
		ids := make([]string, PersonaCount)
		for i := 0; i < PersonaCount; i++ {
			ids[i] = existing[i].ID
		}
		w.Store.Set(qnoteID, ids)
		slog.Info("persona.Create: reusing all existing persona notes", "qnote_id", qnoteID)
		return nil
	}

	personas, err := w.Gemini.GeneratePersonas(ctx, bc.Body)
	if err != nil {
		return fmt.Errorf("persona.Create: gemini: %w", err)
	}
	slog.Info("persona.Create: gemini returned", "count", len(personas))

	if bc.PersonasAnchor == nil || bc.PersonasAnchor.Location == nil || bc.PersonasAnchor.Size == nil {
		return fmt.Errorf("persona.Create: personas anchor missing location/size")
	}

	colors := canvasx.PersonaColors()
	noteIDs := make([]string, PersonaCount)
	var (
		imgWG    sync.WaitGroup
		successN int
		errs     []error
	)

	for i := 0; i < PersonaCount; i++ {
		if existingNote, ok := existing[i]; ok {
			noteIDs[i] = existingNote.ID
			successN++
			continue
		}
		x, imgY, noteY, colW, imgHpx, noteH := canvasx.PersonaColumnLayout(i,
			bc.PersonasAnchor.Location.X, bc.PersonasAnchor.Location.Y,
			bc.PersonasAnchor.Size.Width, bc.PersonasAnchor.Size.Height)

		if i >= len(personas) {
			id := w.createFailedPersonaNote(ctx, i, "Gemini did not generate enough personas", x, noteY, colW, noteH)
			noteIDs[i] = id
			errs = append(errs, fmt.Errorf("persona %d: no data from Gemini", i+1))
			continue
		}

		p := personas[i]
		title := fmt.Sprintf("Persona %d: %s", i+1, p.Name)
		note, err := w.Session.CreateNote(ctx, w.CanvasID, map[string]any{
			"title":            title,
			"text":             atom.FormatPersonaNote(p),
			"location":         map[string]any{"x": x, "y": noteY},
			"size":             map[string]any{"width": colW, "height": noteH},
			"background_color": colors[i%len(colors)],
		})
		if err != nil {
			slog.Error("persona.Create: create note failed", "index", i+1, "err", err)
			id := w.createFailedPersonaNote(ctx, i, err.Error(), x, noteY, colW, noteH)
			noteIDs[i] = id
			errs = append(errs, fmt.Errorf("persona %d: %w", i+1, err))
			continue
		}
		noteIDs[i] = note.ID
		successN++

		// Generate the portrait in the background.
		imgWG.Add(1)
		go func(p atom.Persona, x, imgY, colW, imgHpx float64, title string) {
			defer imgWG.Done()
			if err := w.uploadPersonaImage(ctx, p, title, x, imgY, colW, imgHpx); err != nil {
				slog.Warn("persona.Create: image upload failed",
					"title", title, "err", err)
			}
		}(p, x, imgY, colW, imgHpx, title)
	}

	if successN < MinRequired {
		return fmt.Errorf("persona.Create: only %d/%d personas (min %d): %v",
			successN, PersonaCount, MinRequired, errs)
	}
	if successN < PersonaCount {
		slog.Warn("persona.Create: partial success",
			"successful", successN, "of", PersonaCount, "errors", errs)
	}

	// Filter the failed-placeholder IDs out of the persisted set so Q&A
	// only iterates real personas.
	valid := make([]string, 0, PersonaCount)
	for _, id := range noteIDs {
		if id != "" && !strings.Contains(id, "FAILED") {
			valid = append(valid, id)
		}
	}
	w.Store.Set(qnoteID, valid)
	slog.Info("persona.Create: done", "qnote_id", qnoteID, "persona_ids", len(valid))
	return nil
}

// findExistingPersonas returns persona notes already present on the canvas,
// keyed by their 1-based "Persona N" index. Failed-marker notes are skipped.
func (w *Workflow) findExistingPersonas(ctx context.Context) (map[int]canvus.Note, error) {
	notes, err := w.Session.ListNotes(ctx, w.CanvasID)
	if err != nil {
		return nil, err
	}
	out := make(map[int]canvus.Note)
	for _, n := range notes {
		for i := 0; i < PersonaCount; i++ {
			prefix := fmt.Sprintf("Persona %d: ", i+1)
			if strings.HasPrefix(strings.TrimSpace(n.Title), prefix) && !strings.Contains(n.Title, "FAILED") {
				if _, dup := out[i]; !dup {
					out[i] = n
				}
				break
			}
		}
	}
	return out, nil
}

// createFailedPersonaNote drops a red marker note where persona i should have
// been. Returns the new note ID, or "" if note creation also failed.
func (w *Workflow) createFailedPersonaNote(ctx context.Context, i int, reason string, x, y, width, height float64) string {
	note, err := w.Session.CreateNote(ctx, w.CanvasID, map[string]any{
		"title":            fmt.Sprintf("Persona %d: FAILED", i+1),
		"text":             fmt.Sprintf("Failed to create persona %d.\n\nReason: %s\n\nThis persona will be skipped in Q&A sessions.", i+1, reason),
		"location":         map[string]any{"x": x, "y": y},
		"size":             map[string]any{"width": width, "height": height},
		"background_color": canvasx.ColorRedFailedPersona,
	})
	if err != nil {
		slog.Error("persona.Create: failed marker creation also failed", "index", i+1, "err", err)
		return ""
	}
	return note.ID
}

// uploadPersonaImage generates a DALL-E portrait for the persona, then uploads
// it to the canvas at the top of the persona's column. The persona note is not
// touched here — image upload is fire-and-forget.
func (w *Workflow) uploadPersonaImage(ctx context.Context, p atom.Persona, title string, x, imgY, colW, imgHpx float64) error {
	imgBytes, err := w.OpenAI.GeneratePersonaImage(ctx, p)
	if err != nil {
		return fmt.Errorf("dalle: %w", err)
	}
	meta := map[string]any{
		"title":    title + " Headshot",
		"location": map[string]any{"x": x, "y": imgY},
		"size":     map[string]any{"width": colW, "height": imgHpx},
	}
	body, contentType, err := canvasx.BuildImageMultipart(meta, "persona.png", imgBytes)
	if err != nil {
		return err
	}
	if _, err := w.Session.CreateImage(ctx, w.CanvasID, body, contentType); err != nil {
		return fmt.Errorf("upload: %w", err)
	}
	return nil
}

// FetchPersonas fetches and parses the persona notes previously stored for the
// given Qnote. Returns an error if no IDs are stored or none of the notes can
// be read.
func FetchPersonas(ctx context.Context, s *canvus.Session, canvasID string, store *Store, qnoteID string) ([]atom.Persona, error) {
	ids, ok := store.Get(qnoteID)
	if !ok || len(ids) == 0 {
		return nil, fmt.Errorf("persona.FetchPersonas: no persona IDs stored for %s", qnoteID)
	}
	personas := make([]atom.Persona, 0, len(ids))
	var errs []string
	for _, id := range ids {
		note, err := s.GetNote(ctx, canvasID, id)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", id, err))
			continue
		}
		personas = append(personas, atom.ParsePersonaNote(note.Text))
	}
	if len(personas) == 0 {
		return nil, fmt.Errorf("persona.FetchPersonas: no readable notes (%d errors): %s",
			len(errs), strings.Join(errs, "; "))
	}
	return personas, nil
}

// CountFromCanvas inspects current canvas notes and returns how many real
// (non-FAILED) "Persona N:" notes exist. Used to decide whether persona
// creation needs to run before answering a question.
func CountFromCanvas(ctx context.Context, s *canvus.Session, canvasID string) (int, error) {
	notes, err := s.ListNotes(ctx, canvasID)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, note := range notes {
		t := strings.TrimSpace(note.Title)
		if strings.HasPrefix(t, "Persona ") && !strings.Contains(t, "FAILED") {
			n++
		}
	}
	return n, nil
}

