// Command 03-widget-crud exercises the full lifecycle of a single sticky
// note on a Canvus canvas: create, update, delete, then verify the delete
// took effect.
//
// It exists to demonstrate the idiomatic shape of widget operations through
// the SDK and to give SDK consumers a copy-pasteable template for their own
// widget code.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func main() {
	setupLogging()
	if err := run(); err != nil {
		slog.Error("widget-crud failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	baseURL, err := mustEnv("CANVUS_BASE_URL")
	if err != nil {
		return err
	}
	apiKey, err := mustEnv("CANVUS_API_KEY")
	if err != nil {
		return err
	}
	canvasID, err := mustEnv("CANVUS_CANVAS_ID")
	if err != nil {
		return err
	}

	cfg := &canvus.SessionConfig{BaseURL: baseURL}
	s := canvus.NewSession(cfg, canvus.WithAPIKey(apiKey))
	ctx := context.Background()

	// Step 1 — create.
	createText := fmt.Sprintf("hello from Go @ %s", time.Now().UTC().Format(time.RFC3339))
	createReq := map[string]any{
		"widget_type": "note",
		"text":        createText,
	}
	note, err := s.CreateNote(ctx, canvasID, createReq)
	if err != nil {
		return fmt.Errorf("step 1 CreateNote: %w", err)
	}
	slog.Info("step 1 — note created",
		"widget_id", note.ID,
		"text", note.Text,
		"background_color", note.BackgroundColor)

	// Step 2 — print id (already logged; emit to stdout for piping).
	fmt.Printf("created note id: %s\n", note.ID)

	// Step 3 — update.
	updateText := fmt.Sprintf("updated from Go @ %s", time.Now().UTC().Format(time.RFC3339))
	updateReq := map[string]any{
		"widget_type":      "note",
		"text":             updateText,
		"background_color": "#3AAA34FF",
	}
	updated, err := s.UpdateNote(ctx, canvasID, note.ID, updateReq)
	if err != nil {
		return fmt.Errorf("step 3 UpdateNote: %w", err)
	}
	slog.Info("step 3 — note updated",
		"widget_id", updated.ID,
		"text", updated.Text,
		"background_color", updated.BackgroundColor)

	// Step 4 — print updated widget (already logged; emit to stdout).
	fmt.Printf("updated note id: %s text: %q\n", updated.ID, updated.Text)

	// Step 5 — delete.
	if err := s.DeleteNote(ctx, canvasID, note.ID); err != nil {
		return fmt.Errorf("step 5 DeleteNote: %w", err)
	}
	slog.Info("step 5 — note deleted", "widget_id", note.ID)

	// Step 6 — verify deletion. The server should return 404; the SDK wraps
	// that as canvus.ErrNotFound via APIError.Unwrap.
	_, err = s.GetNote(ctx, canvasID, note.ID)
	switch {
	case err == nil:
		return fmt.Errorf("step 6 expected NotFound after delete, got nil error")
	case errors.Is(err, canvus.ErrNotFound):
		slog.Info("step 6 — verified deletion (got ErrNotFound as expected)")
		return nil
	default:
		return fmt.Errorf("step 6 GetNote returned unexpected error: %w", err)
	}
}

// mustEnv returns the value of name or an error if it is empty/unset.
func mustEnv(name string) (string, error) {
	v := os.Getenv(name)
	if v == "" {
		return "", fmt.Errorf("missing required env var %s", name)
	}
	return v, nil
}

// setupLogging configures slog.Default with a text or JSON handler.
func setupLogging() {
	var h slog.Handler
	switch os.Getenv("LOG_FORMAT") {
	case "json":
		h = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	default:
		h = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	slog.SetDefault(slog.New(h))
}
