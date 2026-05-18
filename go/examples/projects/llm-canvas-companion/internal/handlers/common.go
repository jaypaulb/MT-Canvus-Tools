// Package handlers contains the per-trigger-type processing logic for
// llm-canvas-companion. Each file covers one trigger category:
//
//   - common.go    — shared helpers (note sizing, update-with-retry, cleanup)
//   - note.go      — {{ }} text/image triggers in Note widgets
//   - snapshot.go  — "Snapshot at …" Image widgets → OCR via Google Vision
//   - pdf.go       — AI_Icon_PDF_Precis → PDF download + chunk + summarise
//   - canvas.go    — AI_Icon_Canvus_Precis → canvas-wide summary
//   - image.go     — AI image generation (OpenAI / Azure routing)
package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// processingNoteTitle is the standard title for in-progress status notes.
const processingNoteTitle = "AI Processing"

// processingNoteColor is the dark-red background used for in-progress notes.
const processingNoteColor = "#8B0000FF"

// processingNoteTextColor is white text on the dark-red processing note.
const processingNoteTextColor = "#FFFFFFFF"

// Metrics tracks aggregate counts for the lifetime of the companion process.
var Metrics struct {
	ProcessedNotes  int64
	ProcessedImages int64
	ProcessedPDFs   int64
	Errors          int64

	mu                 sync.Mutex
	ProcessingDuration time.Duration
}

// addDuration safely adds d to Metrics.ProcessingDuration.
func addDuration(d time.Duration) {
	Metrics.mu.Lock()
	Metrics.ProcessingDuration += d
	Metrics.mu.Unlock()
}

// updateNoteWithRetry patches a note text with up to maxRetries attempts.
func updateNoteWithRetry(ctx context.Context, s *canvus.Session, canvasID, noteID, text string, maxRetries int, delay time.Duration) error {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		_, err := s.UpdateNote(ctx, canvasID, noteID, map[string]any{"text": text})
		if err == nil {
			return nil
		}
		lastErr = err
		slog.Warn("updateNoteWithRetry: attempt failed",
			"note_id", noteID, "attempt", attempt, "max", maxRetries, "err", err)
		time.Sleep(delay)
	}
	return fmt.Errorf("updateNoteWithRetry: %w", lastErr)
}

// deleteWidget deletes a widget by type string ("note", "image", "pdf").
// Errors are logged but not propagated — callers treat cleanup as best-effort.
func deleteWidget(ctx context.Context, s *canvus.Session, canvasID, widgetType, widgetID string) {
	var err error
	switch strings.ToLower(widgetType) {
	case "note":
		err = s.DeleteNote(ctx, canvasID, widgetID)
	case "image":
		err = s.DeleteImage(ctx, canvasID, widgetID)
	case "pdf":
		err = s.DeletePDF(ctx, canvasID, widgetID)
	default:
		slog.Warn("deleteWidget: unsupported type", "type", widgetType, "id", widgetID)
		return
	}
	if err != nil {
		slog.Warn("deleteWidget: delete failed", "type", widgetType, "id", widgetID, "err", err)
	}
}

// createProcessingNote creates a status note near the triggering widget and
// returns its ID.  The caller owns deletion of the processing note.
func createProcessingNote(ctx context.Context, s *canvus.Session, canvasID string, w canvus.Widget) (string, error) {
	x, y := 0.0, 0.0
	if w.Location != nil {
		x = w.Location.X
		y = w.Location.Y
	}
	note, err := s.CreateNote(ctx, canvasID, map[string]any{
		"title":            processingNoteTitle,
		"text":             "Processing...",
		"location":         map[string]any{"x": x, "y": y},
		"size":             map[string]any{"width": 300.0, "height": 300.0},
		"depth":            w.Depth + 10,
		"scale":            w.Scale,
		"background_color": processingNoteColor,
		"text_color":       processingNoteTextColor,
		"auto_text_color":  false,
		"pinned":           true,
	})
	if err != nil {
		return "", fmt.Errorf("createProcessingNote: %w", err)
	}
	return note.ID, nil
}

// createResponseNote creates a new note adjacent to the triggering widget with
// automatically sized dimensions based on content length.
// bgColor is the background color to use (pass "" for the default white).
func createResponseNote(ctx context.Context, s *canvus.Session, canvasID, triggerID string, w canvus.Widget, content, bgColor string, isError bool) error {
	size, scale := sizeForContent(content, w, isError)

	if bgColor == "" {
		bgColor = "#FFFFFFBF"
	}

	payload := map[string]any{
		"title":            fmt.Sprintf("Response to %s", triggerID),
		"text":             content,
		"location":         w.Location,
		"size":             size,
		"depth":            w.Depth + 200,
		"scale":            scale,
		"background_color": bgColor,
		"auto_text_color":  false,
		"text_color":       "#000000FF",
	}

	_, err := s.CreateNote(ctx, canvasID, payload)
	if err != nil {
		return fmt.Errorf("createResponseNote: %w", err)
	}
	return nil
}

// sizeForContent calculates width/height/scale for the response note.
// Mirrors the sizing heuristics from the original handlers.go.
func sizeForContent(content string, w canvus.Widget, isError bool) (map[string]any, float64) {
	origWidth := 400.0
	origHeight := 300.0
	origScale := 1.0
	if w.Size != nil {
		origWidth = w.Size.Width
		origHeight = w.Size.Height
	}
	if w.Scale > 0 {
		origScale = w.Scale
	}

	if isError {
		return map[string]any{"width": origWidth, "height": origHeight}, origScale
	}

	contentTokens := float64(len(content)) / 4.0
	if contentTokens < 150 {
		return map[string]any{"width": origWidth, "height": origHeight}, origScale
	}

	const (
		targetWidth    = 830.0
		charsPerLine   = 100.0
		linesPerHeight = 40.0
		baseScale      = 0.37
		minWidth       = 300.0
		maxScale       = 1.0
	)

	contentLines := float64(strings.Count(content, "\n") + 1)
	maxLineLen := 0.0
	totalChars := 0.0
	for _, line := range strings.Split(content, "\n") {
		l := float64(len(line))
		if l > maxLineLen {
			maxLineLen = l
		}
		totalChars += l
	}
	avgLineLen := totalChars / contentLines
	isFormatted := avgLineLen < charsPerLine*0.5 && contentLines > 5

	width := targetWidth
	if isFormatted {
		width = math.Max(minWidth, (maxLineLen/charsPerLine)*targetWidth)
	}

	totalLines := contentLines
	if !isFormatted {
		totalLines += (maxLineLen / charsPerLine)
	}
	height := (totalLines / linesPerHeight) * 1200.0

	contentRatio := math.Min(1.0, math.Max(width/targetWidth, height/1200.0))
	scale := math.Min(maxScale, baseScale*(1.0+(1.0-contentRatio))*2)

	return map[string]any{"width": width, "height": height}, scale
}

// incrementErrors atomically increments the Metrics.Errors counter.
func incrementErrors() {
	atomic.AddInt64(&Metrics.Errors, 1)
}

// removeFile removes path ignoring "not found" errors.
func removeFile(path string) {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		slog.Warn("removeFile: failed", "path", path, "err", err)
	}
}
