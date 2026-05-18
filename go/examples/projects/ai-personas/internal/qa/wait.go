package qa

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/canvasx"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// defaultSettleSeconds is how long the note text must remain unchanged (and
// end with "?") before WaitForQuestion declares a question ready. This guards
// against intermediate edit states where the "?" is typed mid-word.
// Override with QNOTE_SETTLE_SECONDS in the environment.
const defaultSettleSeconds = 2

// settleWindow returns the configured settle duration, reading
// QNOTE_SETTLE_MS (milliseconds, takes precedence) or QNOTE_SETTLE_SECONDS
// from the environment when set.
func settleWindow() time.Duration {
	if v := os.Getenv("QNOTE_SETTLE_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return time.Duration(n) * time.Millisecond
		}
		slog.Warn("qa.wait: QNOTE_SETTLE_MS could not be parsed; ignoring",
			"value", v)
	}
	if v := os.Getenv("QNOTE_SETTLE_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return time.Duration(n) * time.Second
		}
		slog.Warn("qa.wait: QNOTE_SETTLE_SECONDS could not be parsed; using default",
			"value", v, "default_seconds", defaultSettleSeconds)
	}
	return defaultSettleSeconds * time.Second
}

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

// WaitForQuestion subscribes to qnoteID's event stream and returns true when
// the note text ends with "?" and has not changed for settleWindow() seconds.
//
// The subscription replaces the previous 500 ms GetNote poll loop. On startup
// the server replays the current note state; WaitForQuestion evaluates it
// immediately rather than suppressing it, because the question might already
// be present.
//
// If the subscription stream closes before a question is detected (e.g. the
// note was deleted or the server dropped the connection), the function returns
// false so callers don't wait indefinitely.
//
// The function respects ctx cancellation and the timeout parameter.
func WaitForQuestion(ctx context.Context, s *canvus.Session, canvasID, qnoteID string, timeout time.Duration) bool {
	slog.Info("qa.wait: waiting for question", "qnote_id", qnoteID, "timeout", timeout)

	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	events, err := s.SubscribeNote(tctx, canvasID, qnoteID)
	if err != nil {
		slog.Warn("qa.wait: SubscribeNote failed; aborting wait",
			"qnote_id", qnoteID, "err", err)
		return false
	}

	settle := settleWindow()

	// settleTimer fires when the most recently seen question text has been
	// stable for the settle window. timerC is nil (blocking in select) when no
	// qualifying text is pending — a nil channel in a select case never fires.
	var settleTimer *time.Timer
	var timerC <-chan time.Time // nil until a qualifying event is received

	defer func() {
		if settleTimer != nil {
			settleTimer.Stop()
		}
	}()

	// lastText tracks the previous event's text so duplicate server events
	// (same text replayed) do not reset the settle timer.
	var lastText string
	firstEvent := true

	for {
		select {
		case <-tctx.Done():
			slog.Warn("qa.wait: timed out", "qnote_id", qnoteID, "timeout", timeout)
			return false

		case note, ok := <-events:
			if !ok {
				slog.Warn("qa.wait: subscription stream closed before question detected",
					"qnote_id", qnoteID)
				return false
			}

			text := note.Text

			// Suppress duplicate events: same text as last delivery is a no-op.
			if !firstEvent && text == lastText {
				continue
			}
			firstEvent = false
			lastText = text

			if hasQuestion(text) {
				if settleTimer == nil {
					// First qualifying text: start the settle timer.
					settleTimer = time.NewTimer(settle)
					timerC = settleTimer.C
					slog.Debug("qa.wait: question candidate; waiting to settle",
						"qnote_id", qnoteID, "settle", settle)
				} else {
					// Text changed but still qualifies: restart the settle window.
					if !settleTimer.Stop() {
						// Drain the channel if the timer already fired but we
						// haven't processed it yet (rare race on busy loop).
						select {
						case <-settleTimer.C:
						default:
						}
					}
					settleTimer.Reset(settle)
					slog.Debug("qa.wait: question text changed; settle timer reset",
						"qnote_id", qnoteID)
				}
			} else {
				// Text no longer qualifies: cancel the settle timer.
				if settleTimer != nil {
					if !settleTimer.Stop() {
						select {
						case <-settleTimer.C:
						default:
						}
					}
					settleTimer = nil
					timerC = nil
					slog.Debug("qa.wait: question condition lost; settle timer cancelled",
						"qnote_id", qnoteID)
				}
			}

		case <-timerC:
			// timerC is only non-nil when settleTimer is running after a
			// qualifying event — the question text has been stable for the
			// settle window.
			timerC = nil
			slog.Info("qa.wait: question detected and settled", "qnote_id", qnoteID)
			return true
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
