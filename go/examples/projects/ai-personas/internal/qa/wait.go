package qa

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/canvasx"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// pollQuestionInterval is the gap between successive GetNote checks while
// waiting for a user to finish typing a question.
const pollQuestionInterval = 500 * time.Millisecond

// pollErrorTolerance is the max number of consecutive GetNote failures
// WaitForQuestion will absorb before aborting the wait. At the default
// 500ms tick this gives ~10s of grace for transient server errors.
const pollErrorTolerance = 20

// hasQuestion returns true when the trimmed note text ends with "?".
func hasQuestion(text string) bool {
	return strings.HasSuffix(strings.TrimSpace(text), "?")
}

// CheckQuestionPresent returns true when the qnote currently ends with "?".
func CheckQuestionPresent(ctx context.Context, s *canvus.Session, canvasID, qnoteID string) bool {
	note, err := s.GetNote(ctx, canvasID, qnoteID)
	if err != nil {
		return false
	}
	return hasQuestion(note.Text)
}

// WaitForQuestion polls qnoteID until its text ends with "?" or the timeout
// expires. Returns true when a question is detected.
//
// Polling errors (e.g. the qnote was deleted, auth dropped, server is 5xx-ing)
// are tolerated for `pollErrorTolerance` consecutive ticks — the first error
// is logged at Warn, subsequent errors at Debug, and a recovered tick logs
// the recovery. Beyond the tolerance window the wait short-circuits and
// returns false so callers don't sit indefinitely on a broken note.
func WaitForQuestion(ctx context.Context, s *canvus.Session, canvasID, qnoteID string, timeout time.Duration) bool {
	slog.Info("qa.wait: waiting for question", "qnote_id", qnoteID, "timeout", timeout)
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(pollQuestionInterval)
	defer ticker.Stop()
	var consecutiveErrors int
	for {
		select {
		case <-tctx.Done():
			slog.Warn("qa.wait: timed out", "qnote_id", qnoteID, "timeout", timeout)
			return false
		case <-ticker.C:
			note, err := s.GetNote(tctx, canvasID, qnoteID)
			if err != nil {
				consecutiveErrors++
				if consecutiveErrors == 1 {
					slog.Warn("qa.wait: GetNote failed; will retry",
						"qnote_id", qnoteID, "err", err)
				} else {
					slog.Debug("qa.wait: GetNote still failing",
						"qnote_id", qnoteID, "consecutive_errors", consecutiveErrors, "err", err)
				}
				if consecutiveErrors >= pollErrorTolerance {
					slog.Warn("qa.wait: aborting wait after sustained GetNote failures",
						"qnote_id", qnoteID, "consecutive_errors", consecutiveErrors)
					return false
				}
				continue
			}
			if consecutiveErrors > 0 {
				slog.Info("qa.wait: GetNote recovered",
					"qnote_id", qnoteID, "after_errors", consecutiveErrors)
				consecutiveErrors = 0
			}
			if hasQuestion(note.Text) {
				slog.Info("qa.wait: question detected", "qnote_id", qnoteID)
				return true
			}
		}
	}
}

// CreateTimeoutHelper drops an amber helper note next to the qnote informing
// the user that the wait expired.
func CreateTimeoutHelper(ctx context.Context, s *canvus.Session, canvasID, qnoteID string, timeout time.Duration) {
	qnote, err := s.GetNote(ctx, canvasID, qnoteID)
	if err != nil {
		slog.Warn("qa.wait: timeout helper: GetNote failed", "err", err)
		return
	}
	if qnote.Location == nil || qnote.Size == nil {
		return
	}
	hX, hY, hW, hH := canvasx.HelperNotePosition(qnote.Location.X, qnote.Location.Y,
		qnote.Size.Width, qnote.Size.Height)
	text := fmt.Sprintf("The system waited %v for a question to be entered, but none was detected.\n\nPlease enter your question (ending with ?) in the note, then create a new 'New_AI_Question' trigger note to restart the Q&A process.", timeout)
	helper, err := s.CreateNote(ctx, canvasID, map[string]any{
		"title":            HelperTitleTimeout,
		"text":             text,
		"location":         map[string]any{"x": hX, "y": hY},
		"size":             map[string]any{"width": hW, "height": hH},
		"background_color": canvasx.ColorAmberTimeoutHelper,
	})
	if err != nil {
		slog.Warn("qa.wait: timeout helper create failed", "err", err)
		return
	}
	if _, err := s.CreateConnector(ctx, canvasID,
		canvasx.BuildConnectorPayload(helper.ID, qnoteID)); err != nil {
		slog.Warn("qa.wait: timeout helper connector failed", "err", err)
	}
	slog.Info("qa.wait: created timeout helper", "helper_id", helper.ID, "qnote_id", qnoteID)
}
