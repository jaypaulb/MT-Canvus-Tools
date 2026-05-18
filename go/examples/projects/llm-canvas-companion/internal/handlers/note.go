package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/llm-canvas-companion/internal/ai"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/llm-canvas-companion/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// HandleNote processes a Note widget that contains a {{ }} trigger.
// It inspects the text, marks the note as "processing", calls the LLM, and
// creates a response note (text or image).
func HandleNote(ctx context.Context, s *canvus.Session, cfg *config.Config, canvasID string, w canvus.Widget) {
	start := time.Now()

	note, err := s.GetNote(ctx, canvasID, w.ID)
	if err != nil {
		slog.Warn("HandleNote: GetNote failed", "id", w.ID, "err", err)
		incrementErrors()
		return
	}

	text := note.Text
	if !strings.Contains(text, "{{") || !strings.Contains(text, "}}") {
		return // not a trigger
	}

	// Strip delimiters to get the plain prompt.
	baseText := strings.ReplaceAll(strings.ReplaceAll(text, "{{", ""), "}}", "")

	// Mark note as processing immediately.
	if err := updateNoteWithRetry(ctx, s, canvasID, note.ID,
		baseText+"\n\nStarting AI Processing...",
		cfg.MaxRetries, cfg.RetryDelay); err != nil {
		slog.Warn("HandleNote: mark-processing failed", "id", note.ID, "err", err)
		incrementErrors()
		return
	}

	slog.Info("HandleNote: processing trigger",
		"note_id", note.ID,
		"prompt", truncate(baseText, 60))

	aiCtx, cancel := context.WithTimeout(ctx, cfg.AITimeout)
	defer cancel()

	_ = updateNoteWithRetry(ctx, s, canvasID, note.ID,
		baseText+"\n\nAnalysing request...",
		cfg.MaxRetries, cfg.RetryDelay)

	systemMsg := "You are an assistant capable of interpreting structured text triggers from a Note widget. " +
		"Evaluate whether the content in the Note is better suited for generating text or creating an image. " +
		"If generating text, respond with a JSON object like: {\"type\": \"text\", \"content\": \"...\"}. " +
		"If creating an image, respond with a JSON object like: {\"type\": \"image\", \"content\": \"...\"}. " +
		"For image requests, craft a vivid, detailed prompt for an AI image generator. " +
		"Do NOT simply repeat the user's input. The response should be rich and meaningful."

	aiResp, err := ai.ChatJSON(aiCtx, cfg, cfg.OpenAINoteModel, systemMsg, baseText, int(cfg.NoteResponseTokens))
	if err != nil {
		slog.Error("HandleNote: AI error", "note_id", note.ID, "err", err)
		_ = createResponseNote(ctx, s, canvasID, note.ID, w,
			fmt.Sprintf("# AI Processing Error\n\nFailed to process your request.\n\n*Error: %v*", err),
			"", true)
		_ = updateNoteWithRetry(ctx, s, canvasID, note.ID, baseText, cfg.MaxRetries, cfg.RetryDelay)
		incrementErrors()
		return
	}

	var processErr error
	switch aiResp.Type {
	case "text":
		content := strings.ReplaceAll(aiResp.Content, "\\n", "\n")
		_ = updateNoteWithRetry(ctx, s, canvasID, note.ID,
			baseText+"\n\nGenerating text response...", cfg.MaxRetries, cfg.RetryDelay)
		processErr = createResponseNote(ctx, s, canvasID, note.ID, w, content, "", false)

	case "image":
		_ = updateNoteWithRetry(ctx, s, canvasID, note.ID,
			baseText+"\n\nGenerating image... This may take up to 30 seconds.", cfg.MaxRetries, cfg.RetryDelay)
		processErr = handleImageGeneration(ctx, s, cfg, canvasID, w, aiResp.Content)

	default:
		processErr = fmt.Errorf("unexpected AI response type: %q", aiResp.Type)
	}

	if processErr != nil {
		slog.Error("HandleNote: response creation failed", "note_id", note.ID, "err", processErr)
		_ = createResponseNote(ctx, s, canvasID, note.ID, w,
			fmt.Sprintf("# AI Image Generation Error\n\nFailed to generate response.\n\n*Error: %v*", processErr),
			"", true)
		_ = updateNoteWithRetry(ctx, s, canvasID, note.ID, baseText, cfg.MaxRetries, cfg.RetryDelay)
		incrementErrors()
		return
	}

	// Restore note to its clean base text.
	_ = updateNoteWithRetry(ctx, s, canvasID, note.ID, baseText, cfg.MaxRetries, cfg.RetryDelay)

	atomic.AddInt64(&Metrics.ProcessedNotes, 1)
	addDuration(time.Since(start))
	slog.Info("HandleNote: completed", "note_id", note.ID, "elapsed", time.Since(start))
}

// truncate shortens s to n bytes, appending "…" if truncated.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
