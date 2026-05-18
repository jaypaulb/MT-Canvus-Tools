package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/llm-canvas-companion/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/llm-canvas-companion/internal/ocr"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// HandleSnapshot processes an Image widget whose title begins with
// "Snapshot at". It performs OCR via Google Vision and creates a response note
// with the extracted text. On success the snapshot image is deleted.
func HandleSnapshot(ctx context.Context, s *canvus.Session, cfg *config.Config, canvasID string, w canvus.Widget) {
	start := time.Now()

	slog.Info("HandleSnapshot: starting OCR", "image_id", w.ID)

	// Create a processing note near the snapshot.
	processingID, err := createProcessingNote(ctx, s, canvasID, w)
	if err != nil {
		slog.Warn("HandleSnapshot: createProcessingNote failed", "id", w.ID, "err", err)
		incrementErrors()
		return
	}

	opCtx, cancel := context.WithTimeout(ctx, cfg.ProcessingTimeout)
	defer cancel()

	// Download the snapshot image with retries.
	var imageData []byte
	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		_ = updateNoteWithRetry(ctx, s, canvasID, processingID,
			fmt.Sprintf("Downloading snapshot... (attempt %d/%d)", attempt, cfg.MaxRetries),
			1, 0)

		imageData, err = s.DownloadImage(opCtx, canvasID, w.ID)
		if err == nil && len(imageData) > 0 {
			break
		}
		slog.Warn("HandleSnapshot: download attempt failed",
			"attempt", attempt, "err", err)
		if attempt < cfg.MaxRetries {
			time.Sleep(cfg.RetryDelay)
		}
	}
	if len(imageData) == 0 {
		_ = updateNoteWithRetry(ctx, s, canvasID, processingID,
			"Failed to download image after multiple attempts. Click the snapshot again to retry.",
			1, 0)
		incrementErrors()
		return
	}

	_ = updateNoteWithRetry(ctx, s, canvasID, processingID,
		"Processing image through OCR... Please wait.", 1, 0)

	ocrText, err := ocr.PerformOCR(opCtx, cfg.GoogleVisionKey, imageData)
	if err != nil {
		slog.Error("HandleSnapshot: OCR failed", "id", w.ID, "err", err)
		_ = updateNoteWithRetry(ctx, s, canvasID, processingID,
			fmt.Sprintf("Failed to process image: %v\n\nClick the snapshot again to retry.", err),
			1, 0)
		incrementErrors()
		return
	}

	_ = updateNoteWithRetry(ctx, s, canvasID, processingID, "Creating response note...", 1, 0)

	if err := createResponseNote(ctx, s, canvasID, w.ID, w, ocrText, "", false); err != nil {
		slog.Error("HandleSnapshot: createResponseNote failed", "id", w.ID, "err", err)
		_ = updateNoteWithRetry(ctx, s, canvasID, processingID,
			"Failed to create response note. Click the snapshot again to retry.", 1, 0)
		incrementErrors()
		return
	}

	// All succeeded — clean up.
	deleteWidget(ctx, s, canvasID, "note", processingID)
	deleteWidget(ctx, s, canvasID, "image", w.ID)

	atomic.AddInt64(&Metrics.ProcessedImages, 1)
	addDuration(time.Since(start))
	slog.Info("HandleSnapshot: completed", "image_id", w.ID, "elapsed", time.Since(start))
}
