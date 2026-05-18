package qa

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/ai"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/atom"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/business"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/canvasx"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/persona"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// Workflow bundles the dependencies needed to answer one trigger note.
type Workflow struct {
	Session         *canvus.Session
	CanvasID        string
	Gemini          *ai.Gemini
	OpenAI          *ai.OpenAI
	PersonaStore    *persona.Store
	PersonaWorkflow *persona.Workflow
	Helpers         *HelperTracker
	ChatTokenLimit  int
	QuestionTimeout time.Duration
	processing      sync.Map // qnoteID -> struct{}
	answered        sync.Map // qnoteID -> struct{}
}

// NewWorkflow constructs a Workflow.
func NewWorkflow(s *canvus.Session, canvasID string, gem *ai.Gemini, oa *ai.OpenAI,
	pw *persona.Workflow, store *persona.Store, helpers *HelperTracker,
	tokenLimit int, questionTimeout time.Duration) *Workflow {
	return &Workflow{
		Session:         s,
		CanvasID:        canvasID,
		Gemini:          gem,
		OpenAI:          oa,
		PersonaWorkflow: pw,
		PersonaStore:    store,
		Helpers:         helpers,
		ChatTokenLimit:  tokenLimit,
		QuestionTimeout: questionTimeout,
	}
}

// HandleAIQuestion runs the end-to-end Q&A workflow for a New_AI_Question
// trigger note: ensures personas exist, waits for the question to be typed,
// generates answers + meta-answers, lays them out, links them with connectors,
// and groups them in an anchor.
//
// Concurrency: a given Qnote is processed at most once. Concurrent triggers
// for the same Qnote return immediately.
func (w *Workflow) HandleAIQuestion(ctx context.Context, qnoteID string) {
	if _, busy := w.processing.LoadOrStore(qnoteID, struct{}{}); busy {
		slog.Debug("qa: already processing", "qnote_id", qnoteID)
		return
	}
	defer w.processing.Delete(qnoteID)

	slog.Info("qa: HandleAIQuestion start", "qnote_id", qnoteID)

	if err := w.ensurePersonas(ctx, qnoteID); err != nil {
		slog.Error("qa: ensure personas failed", "qnote_id", qnoteID, "err", err)
		return
	}

	if !CheckQuestionPresent(ctx, w.Session, w.CanvasID, qnoteID) {
		if !w.waitForQuestionWithHelper(ctx, qnoteID) {
			return
		}
	}

	if err := w.runAnswerCycle(ctx, qnoteID); err != nil {
		slog.Error("qa: runAnswerCycle failed", "qnote_id", qnoteID, "err", err)
		return
	}

	w.answered.Store(qnoteID, struct{}{})
	slog.Info("qa: HandleAIQuestion complete", "qnote_id", qnoteID)
}

// ensurePersonas runs the persona-creation workflow if personas are missing.
func (w *Workflow) ensurePersonas(ctx context.Context, qnoteID string) error {
	count, err := persona.CountFromCanvas(ctx, w.Session, w.CanvasID)
	if err != nil {
		return fmt.Errorf("count personas: %w", err)
	}
	if count >= persona.MinRequired && w.PersonaStore.Has(qnoteID) {
		return nil
	}
	// Drop the persona-waiting helper before kicking off creation so the
	// dashboard reflects what we're doing.
	qnote, err := w.Session.GetNote(ctx, w.CanvasID, qnoteID)
	if err != nil {
		return fmt.Errorf("get qnote: %w", err)
	}
	helperID, err := createOrUpdateHelper(ctx, w.Session, w.CanvasID,
		HelperTitlePersonas,
		"Personas are being generated. Please wait before proceeding.",
		canvasx.ColorGreyHelper, qnote)
	if err != nil {
		slog.Warn("qa: persona helper create failed", "err", err)
	} else {
		w.Helpers.Set(qnoteID, helperID)
		if _, err := w.Session.UpdateNote(ctx, w.CanvasID, qnoteID, map[string]any{
			"background_color": canvasx.ColorAmberProcessing,
		}); err != nil {
			slog.Warn("qa: set qnote amber failed", "err", err)
		}
	}

	defer DeleteHelper(ctx, w.Session, w.CanvasID, w.Helpers, qnoteID)

	if err := w.PersonaWorkflow.Create(ctx, qnoteID); err != nil {
		return fmt.Errorf("create personas: %w", err)
	}
	return nil
}

// waitForQuestionWithHelper drops a helper next to the qnote, waits for the
// user to type a "?"-terminated question, and cleans up. Returns true when
// a question is detected, false on timeout.
func (w *Workflow) waitForQuestionWithHelper(ctx context.Context, qnoteID string) bool {
	qnote, err := w.Session.GetNote(ctx, w.CanvasID, qnoteID)
	if err != nil {
		slog.Error("qa: GetNote for question helper failed", "err", err)
		return false
	}
	helperID, err := createOrUpdateHelper(ctx, w.Session, w.CanvasID,
		HelperTitleQuestion,
		"Please enter a question in the main note to begin the Q&A process.",
		canvasx.ColorGreyHelper, qnote)
	if err == nil {
		w.Helpers.Set(qnoteID, helperID)
	}
	if _, err := w.Session.UpdateNote(ctx, w.CanvasID, qnoteID, map[string]any{
		"background_color": canvasx.ColorAmberProcessing,
	}); err != nil {
		slog.Warn("qa: set qnote amber failed", "err", err)
	}

	if !WaitForQuestion(ctx, w.Session, w.CanvasID, qnoteID, w.QuestionTimeout) {
		CreateTimeoutHelper(ctx, w.Session, w.CanvasID, qnoteID, w.QuestionTimeout)
		DeleteHelper(ctx, w.Session, w.CanvasID, w.Helpers, qnoteID)
		return false
	}
	return true
}

// runAnswerCycle is the core: load personas+context, generate answers,
// generate meta-answers, lay everything out, wire connectors, group anchor,
// then close out the helper and recolor the qnote to "done".
func (w *Workflow) runAnswerCycle(ctx context.Context, qnoteID string) error {
	qnote, err := w.Session.GetNote(ctx, w.CanvasID, qnoteID)
	if err != nil {
		return fmt.Errorf("get qnote: %w", err)
	}

	// Tell the user we're working on it.
	helperID, err := createOrUpdateHelper(ctx, w.Session, w.CanvasID,
		HelperTitleQuestion,
		generationWaitMessage(w.Gemini.ChatModel()),
		canvasx.ColorGreyHelper, qnote)
	if err == nil {
		w.Helpers.Set(qnoteID, helperID)
	}

	personas, err := persona.FetchPersonas(ctx, w.Session, w.CanvasID, w.PersonaStore, qnoteID)
	if err != nil || len(personas) < persona.MinRequired {
		// Try regenerating personas in case the store is out of sync.
		if err := w.PersonaWorkflow.Create(ctx, qnoteID); err != nil {
			return fmt.Errorf("recreate personas: %w", err)
		}
		personas, err = persona.FetchPersonas(ctx, w.Session, w.CanvasID, w.PersonaStore, qnoteID)
		if err != nil || len(personas) < persona.MinRequired {
			return fmt.Errorf("not enough personas after retry: %w", err)
		}
	}
	slog.Info("qa.cycle: working with personas", "count", len(personas))

	bc, err := business.Extract(ctx, w.Session, w.CanvasID)
	if err != nil {
		return fmt.Errorf("business context: %w", err)
	}

	question := extractQuestion(qnote.Text)
	if question == "" {
		return fmt.Errorf("empty question after extraction")
	}

	spec, err := newGridSpec(qnote)
	if err != nil {
		return err
	}

	sm := w.Gemini.NewSessionManager()

	start := time.Now()
	answers, answerErrs := generateAnswers(ctx, sm, personas, question, bc.Body, w.ChatTokenLimit)
	logAnswerStats("answers", answers, answerErrs)
	if countSuccess(answers, answerErrs) < MinAnswers {
		return fmt.Errorf("not enough answers generated")
	}
	slog.Info("qa.cycle: answers generated", "elapsed", time.Since(start))

	answerIDs := createGridNotes(ctx, w.Session, w.CanvasID, personas, answers, answerErrs,
		spec, spec.answers, " Answer")

	metaStart := time.Now()
	metaAnswers, metaErrs := generateMetaAnswers(ctx, sm, personas, answers, answerErrs, bc.Body, w.ChatTokenLimit)
	logAnswerStats("meta", metaAnswers, metaErrs)
	slog.Info("qa.cycle: meta-answers generated", "elapsed", time.Since(metaStart))

	metaIDs := createGridNotes(ctx, w.Session, w.CanvasID, personas, metaAnswers, metaErrs,
		spec, spec.metaOffs, " Meta Answer")

	n := createConnectors(ctx, w.Session, w.CanvasID, qnoteID, answerIDs, metaIDs)
	slog.Info("qa.cycle: connectors created", "count", n)

	allNoteIDs := append(append([]string{}, answerIDs...), metaIDs...)
	createGroupAnchor(ctx, w.Session, w.CanvasID, question+" (Script Made)", filterNonEmpty(allNoteIDs))

	// Restore the qnote: green background, only the original question text.
	if _, err := w.Session.UpdateNote(ctx, w.CanvasID, qnoteID, map[string]any{
		"background_color": canvasx.ColorGreenComplete,
		"text":             question,
	}); err != nil {
		slog.Warn("qa.cycle: restore qnote failed", "err", err)
	}
	DeleteHelper(ctx, w.Session, w.CanvasID, w.Helpers, qnoteID)
	return nil
}

// HandleFollowupConnector handles the "connector drawn from a persona-answer
// note to a fresh question note" case: it generates a single follow-up
// answer in that persona's voice and places it opposite the source note.
func (w *Workflow) HandleFollowupConnector(ctx context.Context, conn canvus.Connector) {
	if conn.Src == nil || conn.Dst == nil {
		return
	}
	srcID, dstID := conn.Src.ID, conn.Dst.ID
	if srcID == "" || dstID == "" {
		return
	}
	src, err := w.Session.GetNote(ctx, w.CanvasID, srcID)
	if err != nil {
		slog.Debug("qa.followup: src is not a note", "id", srcID)
		return
	}
	dst, err := w.Session.GetNote(ctx, w.CanvasID, dstID)
	if err != nil {
		slog.Debug("qa.followup: dst is not a note", "id", dstID)
		return
	}

	personaColors := canvasx.PersonaColorSet()
	if !strings.HasSuffix(src.Title, " Answer") ||
		!personaColors[strings.ToLower(src.BackgroundColor)] {
		return
	}
	if !hasQuestion(dst.Text) {
		return
	}
	personaName := strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(src.Title,
		" Followup Answer"), " Meta Answer"), " Answer"))

	// Make sure personas exist; reuse if so.
	if err := w.PersonaWorkflow.Create(ctx, dstID); err != nil {
		slog.Warn("qa.followup: persona create failed", "err", err)
		return
	}
	personas, err := persona.FetchPersonas(ctx, w.Session, w.CanvasID, w.PersonaStore, dstID)
	if err != nil {
		slog.Warn("qa.followup: fetch personas failed", "err", err)
		return
	}
	var p *atom.Persona
	for i := range personas {
		if personas[i].Name == personaName {
			p = &personas[i]
			break
		}
	}
	if p == nil {
		slog.Warn("qa.followup: persona not found", "name", personaName)
		return
	}

	bc, err := business.Extract(ctx, w.Session, w.CanvasID)
	if err != nil {
		slog.Warn("qa.followup: business context failed", "err", err)
		return
	}

	sm := w.Gemini.NewSessionManager()
	ans, err := sm.AnswerAs(ctx, *p, bc.Body, dst.Text)
	if err != nil {
		slog.Warn("qa.followup: answer failed", "err", err)
		return
	}
	if len(ans) > w.ChatTokenLimit {
		ans2, err := sm.AnswerAs(ctx, *p, bc.Body,
			fmt.Sprintf("Please rephrase your answer in a much more succinct, short, and verbal way. Limit your response to %d characters.", w.ChatTokenLimit))
		if err == nil {
			ans = ans2
		}
	}
	if src.Location == nil || dst.Location == nil || dst.Size == nil {
		return
	}
	dx := src.Location.X - dst.Location.X
	dy := src.Location.Y - dst.Location.Y
	fX, fY := dst.Location.X+dx, dst.Location.Y+dy
	scale := dst.Scale
	if scale == 0 {
		scale = 1
	}
	fup, err := w.Session.CreateNote(ctx, w.CanvasID, map[string]any{
		"title":            p.Name + " Followup Answer",
		"text":             ans,
		"location":         map[string]any{"x": fX, "y": fY},
		"size":             map[string]any{"width": dst.Size.Width, "height": dst.Size.Height},
		"background_color": src.BackgroundColor,
		"scale":            scale,
	})
	if err != nil {
		slog.Warn("qa.followup: create followup note failed", "err", err)
		return
	}
	if _, err := w.Session.CreateConnector(ctx, w.CanvasID,
		canvasx.BuildConnectorPayload(dstID, fup.ID)); err != nil {
		slog.Warn("qa.followup: create connector failed", "err", err)
	}
	slog.Info("qa.followup: follow-up created", "persona", p.Name, "note_id", fup.ID)
}

// filterNonEmpty returns ids with empty strings removed.
func filterNonEmpty(ids []string) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != "" {
			out = append(out, id)
		}
	}
	return out
}

// generationWaitMessage returns a model-specific user-facing wait string so
// users know how long to expect (flash-lite is fast, pro is slow).
func generationWaitMessage(chatModel string) string {
	m := strings.ToLower(chatModel)
	switch {
	case strings.Contains(m, "flash-lite"):
		return "Generating answers, please wait... This can take up to 30 seconds."
	case strings.Contains(m, "flash"):
		return "Generating answers, please wait... This can take up to 60 seconds."
	case strings.Contains(m, "pro"):
		return "Generating answers, please wait... This could take a few minutes as the model thinks about its answers."
	}
	return "Generating answers, please wait..."
}
