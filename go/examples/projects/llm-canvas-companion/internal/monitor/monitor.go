// Package monitor subscribes to the Canvus widget stream and dispatches
// incoming widgets to the appropriate handler.
//
// # Trigger conventions
//
// Note widgets: if the text contains {{ and }}, route to handlers.HandleNote.
// Image widgets: if the title begins with "Snapshot at", route to
// handlers.HandleSnapshot.
// Image widgets: if the title begins with "AI_Icon_PDF_Precis", route to
// handlers.HandlePDFPrecis (parent_id identifies the PDF).
// Image widgets: if the title begins with "AI_Icon_Canvus_Precis", route to
// handlers.HandleCanvasPrecis.
//
// # Snapshot drain
//
// On startup the server sends a full snapshot of all existing widgets.
// To avoid answering pre-existing triggers, any widget ID seen during
// the first SnapshotDrainSeconds is recorded as "already seen" without
// dispatching.  After the drain window, new IDs are dispatched.
package monitor

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/llm-canvas-companion/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/llm-canvas-companion/internal/handlers"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// Monitor subscribes to a Canvus canvas and dispatches trigger events.
type Monitor struct {
	session       *canvus.Session
	cfg           *config.Config
	canvasID      string
	drainDuration time.Duration

	mu              sync.Mutex
	seen            map[string]struct{}
	snapshotDrained bool
	sharedCanvasID  string // ID of the SharedCanvas widget, if discovered
}

// New creates a Monitor ready to run.
func New(s *canvus.Session, cfg *config.Config) *Monitor {
	return &Monitor{
		session:       s,
		cfg:           cfg,
		canvasID:      cfg.CanvasID,
		drainDuration: time.Duration(cfg.SnapshotDrainSeconds) * time.Second,
		seen:          make(map[string]struct{}),
	}
}

// Run subscribes to the widget stream and blocks until ctx is cancelled.
func (m *Monitor) Run(ctx context.Context) error {
	events, err := m.session.SubscribeWidgets(ctx, m.canvasID)
	if err != nil {
		return err
	}

	drainTimer := time.NewTimer(m.drainDuration)
	defer drainTimer.Stop()

	slog.Info("monitor: subscribed to canvas",
		"canvas_id", m.canvasID,
		"drain_seconds", m.drainDuration.Seconds())

	for {
		select {
		case <-ctx.Done():
			slog.Info("monitor: context cancelled, stopping")
			return nil

		case <-drainTimer.C:
			m.mu.Lock()
			m.snapshotDrained = true
			n := len(m.seen)
			m.mu.Unlock()
			slog.Info("monitor: drain window elapsed",
				"pre_existing_seen", n)

		case widget, ok := <-events:
			if !ok {
				slog.Info("monitor: event channel closed")
				return nil
			}
			if err := m.dispatch(ctx, widget); err != nil {
				slog.Warn("monitor: dispatch error",
					"widget_id", widget.ID,
					"type", widget.WidgetType,
					"err", err)
			}
		}
	}
}

// dispatch handles a single widget event.
func (m *Monitor) dispatch(ctx context.Context, w canvus.Widget) error {
	if w.State != "normal" {
		return nil
	}

	m.mu.Lock()
	_, alreadySeen := m.seen[w.ID]
	m.seen[w.ID] = struct{}{}
	drained := m.snapshotDrained
	sharedID := m.sharedCanvasID
	m.mu.Unlock()

	// Track SharedCanvas widget.
	if w.WidgetType == "SharedCanvas" {
		m.mu.Lock()
		if m.sharedCanvasID == "" {
			m.sharedCanvasID = w.ID
		}
		m.mu.Unlock()
		return nil
	}

	if alreadySeen {
		return nil
	}
	if !drained {
		return nil
	}

	switch w.WidgetType {
	case "Note":
		// Note triggers: {{ }} in text — we need to fetch the Note to read text.
		// The Widget envelope doesn't carry text, so we fetch it.
		note, err := m.session.GetNote(ctx, m.canvasID, w.ID)
		if err != nil {
			return err
		}
		if !strings.Contains(note.Text, "{{") || !strings.Contains(note.Text, "}}") {
			return nil
		}
		go handlers.HandleNote(ctx, m.session, m.cfg, m.canvasID, w)

	case "Image":
		title := m.getImageTitle(ctx, w.ID)
		switch {
		case strings.HasPrefix(title, "Snapshot at"):
			go handlers.HandleSnapshot(ctx, m.session, m.cfg, m.canvasID, w)

		case strings.HasPrefix(title, "AI_Icon_PDF_Precis"):
			// Parent must not be the shared canvas.
			if w.ParentID == "" || w.ParentID == sharedID {
				slog.Debug("monitor: PDF precis icon ignored (on shared canvas)")
				return nil
			}
			go handlers.HandlePDFPrecis(ctx, m.session, m.cfg, m.canvasID, w.ParentID, w)

		case strings.HasPrefix(title, "AI_Icon_Canvus_Precis"):
			// Parent must be the shared canvas.
			if w.ParentID != sharedID {
				slog.Warn("monitor: AI_Icon_Canvus_Precis parent is not shared canvas",
					"parent_id", w.ParentID, "shared_canvas_id", sharedID)
				return nil
			}
			go handlers.HandleCanvasPrecis(ctx, m.session, m.cfg, m.canvasID, w)
		}
	}

	return nil
}

// getImageTitle fetches the title field of an Image widget.
// Returns "" on error (non-fatal; the dispatcher skips the widget).
func (m *Monitor) getImageTitle(ctx context.Context, imageID string) string {
	img, err := m.session.GetImage(ctx, m.canvasID, imageID)
	if err != nil {
		slog.Debug("monitor: GetImage failed", "id", imageID, "err", err)
		return ""
	}
	return img.Title
}
