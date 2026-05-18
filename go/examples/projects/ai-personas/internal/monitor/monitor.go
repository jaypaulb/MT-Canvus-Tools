// Package monitor subscribes to the Canvus widget stream for the configured
// canvas and dispatches the three persona-trigger events: Create_Personas
// notes, New_AI_Question notes, and Connector creation.
//
// # Triggers
//
//   - Note titled "Create_Personas" → invokes persona.Workflow.Create
//   - Note titled "New_AI_Question" with white background → invokes qa.Workflow.HandleAIQuestion
//   - Connector creation → invokes qa.Workflow.HandleFollowupConnector
//
// # Snapshot drain
//
// On startup the Canvus server replays every existing widget through the
// subscribe stream. The monitor records all widgets seen during
// SnapshotDrainSeconds without dispatching them, then begins dispatching
// genuinely new triggers.
package monitor

import (
	"context"
	"log/slog"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/persona"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/qa"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// Monitor subscribes to a Canvus canvas and dispatches trigger events.
type Monitor struct {
	session         *canvus.Session
	canvasID        string
	drainDuration   time.Duration
	personaWorkflow *persona.Workflow
	qaWorkflow      *qa.Workflow
	wg              *sync.WaitGroup // tracks outstanding handler goroutines for graceful shutdown

	mu      sync.Mutex
	seen    map[string]struct{}
	drained bool
}

// New constructs a Monitor.
func New(s *canvus.Session, canvasID string, drainSeconds int,
	pw *persona.Workflow, qw *qa.Workflow, wg *sync.WaitGroup) *Monitor {
	return &Monitor{
		session:         s,
		canvasID:        canvasID,
		drainDuration:   time.Duration(drainSeconds) * time.Second,
		personaWorkflow: pw,
		qaWorkflow:      qw,
		wg:              wg,
		seen:            make(map[string]struct{}),
	}
}

// Run subscribes to the canvas widget stream and dispatches events until ctx
// is cancelled.
func (m *Monitor) Run(ctx context.Context) error {
	events, err := m.session.SubscribeWidgets(ctx, m.canvasID)
	if err != nil {
		return err
	}
	drainTimer := time.NewTimer(m.drainDuration)
	defer drainTimer.Stop()

	slog.Info("monitor: subscribed", "canvas_id", m.canvasID, "drain_seconds", m.drainDuration.Seconds())

	for {
		select {
		case <-ctx.Done():
			slog.Info("monitor: context cancelled")
			return nil
		case <-drainTimer.C:
			m.mu.Lock()
			m.drained = true
			n := len(m.seen)
			m.mu.Unlock()
			slog.Info("monitor: drain window elapsed", "pre_existing", n)
		case widget, ok := <-events:
			if !ok {
				slog.Info("monitor: event channel closed")
				return nil
			}
			m.dispatch(ctx, widget)
		}
	}
}

// dispatch routes a widget event to the appropriate handler. Only widgets in
// the "normal" state that have not been seen before during the drain window
// (and after drain) are passed to a handler.
func (m *Monitor) dispatch(ctx context.Context, w canvus.Widget) {
	if w.State != "normal" {
		return
	}
	m.mu.Lock()
	_, already := m.seen[w.ID]
	m.seen[w.ID] = struct{}{}
	drained := m.drained
	m.mu.Unlock()
	if already || !drained {
		return
	}

	switch w.WidgetType {
	case "Note":
		m.handleNoteEvent(ctx, w)
	case "Connector":
		m.handleConnectorEvent(ctx, w)
	}
}

// handleNoteEvent fetches the full note (the Widget envelope omits title/text)
// and dispatches Create_Personas or New_AI_Question if applicable.
func (m *Monitor) handleNoteEvent(ctx context.Context, w canvus.Widget) {
	note, err := m.session.GetNote(ctx, m.canvasID, w.ID)
	if err != nil {
		slog.Debug("monitor: GetNote failed", "id", w.ID, "err", err)
		return
	}
	title := strings.TrimSpace(note.Title)
	switch {
	case title == "Create_Personas":
		slog.Info("monitor: Create_Personas trigger", "note_id", note.ID)
		m.runWithRecover("CreatePersonas", func() {
			if err := m.personaWorkflow.Create(ctx, note.ID); err != nil {
				slog.Error("monitor: persona create failed", "err", err)
				return
			}
			if err := m.session.DeleteNote(ctx, m.canvasID, note.ID); err != nil {
				slog.Warn("monitor: delete Create_Personas note failed", "err", err)
			} else {
				slog.Info("monitor: Create_Personas note deleted", "note_id", note.ID)
			}
		})
	case strings.EqualFold(title, "New_AI_Question") && isWhiteBackground(note.BackgroundColor):
		slog.Info("monitor: New_AI_Question trigger", "note_id", note.ID)
		m.runWithRecover("HandleAIQuestion", func() {
			m.qaWorkflow.HandleAIQuestion(ctx, note.ID)
		})
	}
}

// handleConnectorEvent treats any new Connector as a follow-up trigger; the
// workflow itself decides whether the source/destination shapes are eligible.
func (m *Monitor) handleConnectorEvent(ctx context.Context, w canvus.Widget) {
	conn, err := m.session.GetConnector(ctx, m.canvasID, w.ID)
	if err != nil {
		slog.Debug("monitor: GetConnector failed", "id", w.ID, "err", err)
		return
	}
	slog.Info("monitor: Connector trigger", "connector_id", conn.ID)
	m.runWithRecover("HandleFollowupConnector", func() {
		m.qaWorkflow.HandleFollowupConnector(ctx, *conn)
	})
}

// runWithRecover launches fn in a tracked goroutine, recovering from any
// panic so a single failing handler can't take down the monitor.
func (m *Monitor) runWithRecover(name string, fn func()) {
	if m.wg != nil {
		m.wg.Add(1)
	}
	go func() {
		if m.wg != nil {
			defer m.wg.Done()
		}
		defer func() {
			if r := recover(); r != nil {
				slog.Error("monitor: handler panic",
					"handler", name, "panic", r, "stack", string(debug.Stack()))
			}
		}()
		fn()
	}()
}

// isWhiteBackground accepts the two RGBA representations of white that the
// New_AI_Question trigger requires (the legacy app stored both).
func isWhiteBackground(c string) bool {
	c = strings.ToLower(strings.TrimSpace(c))
	return c == "#ffffffff" || c == "#ffffff"
}
