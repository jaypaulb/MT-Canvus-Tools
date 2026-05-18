package qa

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/atom"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/canvasx"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// gridSpec captures the geometry derived from the question note for laying
// out the answer/meta grid.
type gridSpec struct {
	qx, qy   float64
	qw, qh   float64
	scale    float64
	spacing  float64
	colors   []string
	answers  []canvasx.GridOffset
	metaOffs []canvasx.GridOffset
}

// newGridSpec builds a gridSpec from the question note.
func newGridSpec(q *canvus.Note) (gridSpec, error) {
	if q.Location == nil || q.Size == nil {
		return gridSpec{}, fmt.Errorf("newGridSpec: question note missing location/size")
	}
	scale := q.Scale
	if scale == 0 {
		scale = 1
	}
	return gridSpec{
		qx:       q.Location.X,
		qy:       q.Location.Y,
		qw:       q.Size.Width,
		qh:       q.Size.Height,
		scale:    scale,
		spacing:  canvasx.Spacing(q.Size.Width, scale),
		colors:   canvasx.PersonaColors(),
		answers:  canvasx.AnswerGridOffsets(),
		metaOffs: canvasx.MetaAnswerGridOffsets(),
	}, nil
}

// createGridNotes writes the answer/meta notes in parallel using offsets from
// spec. Indices with an empty answer are skipped, returning "" in the slot.
func createGridNotes(ctx context.Context, s *canvus.Session, canvasID string,
	personas []atom.Persona, answers []string, errs []error,
	spec gridSpec, offsets []canvasx.GridOffset, titleSuffix string) []string {

	ids := make([]string, len(personas))
	var wg sync.WaitGroup
	wg.Add(len(personas))
	for i, p := range personas {
		go func(i int, p atom.Persona) {
			defer wg.Done()
			if answers[i] == "" || errs[i] != nil {
				return
			}
			off := offsets[i%len(offsets)]
			x, y := canvasx.GridPosition(spec.qx, spec.qy, off, spec.qw, spec.qh, spec.scale, spec.spacing)
			note, err := s.CreateNote(ctx, canvasID, map[string]any{
				"title":            p.Name + titleSuffix,
				"text":             answers[i],
				"location":         map[string]any{"x": x, "y": y},
				"size":             map[string]any{"width": spec.qw, "height": spec.qh},
				"background_color": spec.colors[i%len(spec.colors)],
				"scale":            spec.scale,
			})
			if err != nil {
				slog.Error("qa.grid: create note failed",
					"persona", p.Name, "suffix", titleSuffix, "err", err)
				return
			}
			ids[i] = note.ID
		}(i, p)
	}
	wg.Wait()
	return ids
}

// createConnectors creates question->answer and (where present)
// answer->meta connectors in parallel.
func createConnectors(ctx context.Context, s *canvus.Session, canvasID, qnoteID string, answerIDs, metaIDs []string) int {
	var (
		wg sync.WaitGroup
		mu sync.Mutex
		n  int
	)
	for i := range answerIDs {
		if answerIDs[i] == "" {
			continue
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, err := s.CreateConnector(ctx, canvasID,
				canvasx.BuildConnectorPayload(qnoteID, answerIDs[i])); err != nil {
				slog.Error("qa.grid: question->answer connector failed", "index", i, "err", err)
				return
			}
			mu.Lock()
			n++
			mu.Unlock()
			if metaIDs[i] == "" {
				return
			}
			if _, err := s.CreateConnector(ctx, canvasID,
				canvasx.BuildConnectorPayload(answerIDs[i], metaIDs[i])); err != nil {
				slog.Error("qa.grid: answer->meta connector failed", "index", i, "err", err)
				return
			}
			mu.Lock()
			n++
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	return n
}

// createGroupAnchor draws a Canvus anchor around all created notes so the
// user can collapse the Q&A into a single navigable group.
func createGroupAnchor(ctx context.Context, s *canvus.Session, canvasID, anchorName string, noteIDs []string) {
	if len(noteIDs) == 0 {
		return
	}
	bb, n := computeBounds(ctx, s, canvasID, noteIDs)
	if n == 0 {
		return
	}
	payload := map[string]any{
		"anchor_name": anchorName,
		"location":    map[string]any{"x": bb.MinX, "y": bb.MinY},
		"size":        map[string]any{"width": bb.Width(), "height": bb.Height()},
		"notes":       noteIDs,
	}
	if _, err := s.CreateAnchor(ctx, canvasID, payload); err != nil {
		slog.Warn("qa.grid: create anchor failed", "err", err)
		return
	}
	slog.Info("qa.grid: anchor created", "name", anchorName, "notes", len(noteIDs))
}

// computeBounds fetches the just-created notes and computes their union
// bounding box. Returns the box and the count of notes that contributed.
func computeBounds(ctx context.Context, s *canvus.Session, canvasID string, noteIDs []string) (canvasx.BoundingBox, int) {
	notes, err := s.ListNotes(ctx, canvasID)
	if err != nil {
		return canvasx.BoundingBox{}, 0
	}
	want := make(map[string]struct{}, len(noteIDs))
	for _, id := range noteIDs {
		if id != "" {
			want[id] = struct{}{}
		}
	}
	bb := canvasx.BoundingBox{MinX: 1e18, MinY: 1e18, MaxX: -1e18, MaxY: -1e18}
	n := 0
	for _, note := range notes {
		if _, ok := want[note.ID]; !ok {
			continue
		}
		if note.Location == nil || note.Size == nil {
			continue
		}
		x, y := note.Location.X, note.Location.Y
		if x < bb.MinX {
			bb.MinX = x
		}
		if y < bb.MinY {
			bb.MinY = y
		}
		if x+note.Size.Width > bb.MaxX {
			bb.MaxX = x + note.Size.Width
		}
		if y+note.Size.Height > bb.MaxY {
			bb.MaxY = y + note.Size.Height
		}
		n++
	}
	return bb, n
}
