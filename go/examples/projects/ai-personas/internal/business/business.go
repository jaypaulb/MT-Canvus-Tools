// Package business extracts the Business Model Canvas notes from a fetched
// widget list and locates the "Personas" anchor where personas will be drawn.
package business

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/canvasx"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// RequiredNoteTitles is the canonical Business Model Canvas note set the
// persona generator expects to see on the canvas.
var RequiredNoteTitles = []string{
	"KEY PARTNERS",
	"KEY ACTIVITIES",
	"VALUE PROPOSITIONS",
	"CUSTOMER RELATIONSHIPS",
	"CUSTOMER SEGMENTS",
	"KEY RESOURCES",
	"CHANNELS",
	"COST STRUCTURE",
	"REVENUE STREAMS",
}

// Context bundles the extracted business model.
type Context struct {
	// Body is the concatenated "TITLE: text" pairs for all required notes.
	Body string
	// PersonasAnchor is the anchor whose AnchorName equals "Personas".
	PersonasAnchor *canvus.Anchor
	// Missing lists required titles that were not found on the canvas.
	Missing []string
}

// Extract scans notes+anchors on the canvas and assembles a Context.
// Returns an error (with Missing populated) if any required note is absent
// or the Personas anchor is not found.
func Extract(ctx context.Context, s *canvus.Session, canvasID string) (*Context, error) {
	notes, err := s.ListNotes(ctx, canvasID)
	if err != nil {
		return nil, fmt.Errorf("business.Extract: list notes: %w", err)
	}
	anchors, err := s.ListAnchors(ctx, canvasID)
	if err != nil {
		return nil, fmt.Errorf("business.Extract: list anchors: %w", err)
	}
	return assemble(notes, anchors)
}

// assemble does the pure-data work of pairing notes to required titles and
// locating the Personas anchor. Pure function for unit testing.
func assemble(notes []canvus.Note, anchors []canvus.Anchor) (*Context, error) {
	found := make(map[string]canvus.Note, len(RequiredNoteTitles))
	for _, n := range notes {
		titleUpper := strings.ToUpper(strings.TrimSpace(n.Title))
		for _, req := range RequiredNoteTitles {
			if titleUpper == req {
				if _, already := found[req]; !already {
					found[req] = n
				}
				break
			}
		}
	}

	var missing []string
	for _, req := range RequiredNoteTitles {
		if _, ok := found[req]; !ok {
			missing = append(missing, req)
		}
	}

	var personasAnchor *canvus.Anchor
	for i := range anchors {
		if strings.EqualFold(strings.TrimSpace(anchors[i].AnchorName), "Personas") {
			personasAnchor = &anchors[i]
			break
		}
	}

	if len(missing) > 0 {
		return &Context{Missing: missing, PersonasAnchor: personasAnchor},
			fmt.Errorf("business.Extract: missing required notes: %v", missing)
	}
	if personasAnchor == nil {
		return &Context{}, fmt.Errorf("business.Extract: Personas anchor not found")
	}

	var parts []string
	for _, req := range RequiredNoteTitles {
		n := found[req]
		parts = append(parts, fmt.Sprintf("%s: %s", n.Title, n.Text))
	}
	body := strings.Join(parts, "\n\n")

	const minLen = 100
	if len(strings.TrimSpace(body)) < minLen {
		slog.Warn("business context appears short", "chars", len(strings.TrimSpace(body)))
	}
	slog.Info("extracted business context", "notes", len(found), "chars", len(body))
	return &Context{Body: body, PersonasAnchor: personasAnchor}, nil
}

// CreateMissingNotesHelper drops a red helper note onto the canvas listing
// the required Business Model Canvas notes that are absent. Returns the new
// note's ID, or "" if creation failed.
func CreateMissingNotesHelper(ctx context.Context, s *canvus.Session, canvasID string, missing []string, personasAnchor *canvus.Anchor) string {
	if len(missing) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("The following required Business Model Canvas notes are missing:\n\n")
	for i, m := range missing {
		fmt.Fprintf(&sb, "%d. %s\n", i+1, m)
	}
	sb.WriteString("\nPlease add these notes with the exact titles listed above, then try again.")

	x, y, w, h := 0.0, 0.0, 400.0, 300.0
	if personasAnchor != nil && personasAnchor.Location != nil {
		x = personasAnchor.Location.X - 450
		y = personasAnchor.Location.Y
	}
	if personasAnchor != nil && personasAnchor.Size != nil {
		if pw := personasAnchor.Size.Width * 0.5; pw > 300 {
			w = pw
		}
		if ph := personasAnchor.Size.Height * 0.3; ph > 200 {
			h = ph
		}
	}

	note, err := s.CreateNote(ctx, canvasID, map[string]any{
		"title":            "Missing Required Notes",
		"text":             sb.String(),
		"location":         map[string]any{"x": x, "y": y},
		"size":             map[string]any{"width": w, "height": h},
		"background_color": canvasx.ColorRedFailedPersona,
	})
	if err != nil {
		slog.Error("failed to create missing-notes helper", "err", err)
		return ""
	}
	slog.Info("created missing-notes helper", "note_id", note.ID, "missing", len(missing))
	return note.ID
}
