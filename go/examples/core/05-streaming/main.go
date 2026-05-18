// Command 05-streaming subscribes to the notes endpoint of a Canvus canvas
// using the typed Session.SubscribeNotes helper and prints one summary line
// per note received, exiting cleanly on SIGINT or after
// STREAM_DURATION_SECONDS elapses.
//
// The SDK's typed Subscribe helpers (Phase 4b §4.1 #5-#10) handle the wire
// details — long-lived HTTPS subscription, NDJSON line decoding, batched
// snapshot frames vs. single-object deltas — and yield individual typed
// values on a channel. Authentication and the API base URL come from the
// Session, so the example needs no HTTP machinery of its own.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

const defaultDurationSeconds = 30

func main() {
	setupLogging()
	if err := run(); err != nil {
		slog.Error("streaming failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	baseURL, err := mustEnv("CANVUS_API_URL")
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

	cfg := &canvus.SessionConfig{
		BaseURL: baseURL,
		// Long timeout — the stream itself is long-lived. The SDK's default
		// 30s timeout would kill the request after 30s; we want minutes.
		RequestTimeout: 10 * time.Minute,
	}
	s := canvus.NewSession(cfg, canvus.WithAPIKey(apiKey))

	duration := streamDuration()
	slog.Info("starting subscribe",
		"canvas_id", canvasID,
		"duration_seconds", int(duration.Seconds()))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	deadline, cancelDeadline := context.WithTimeout(ctx, duration)
	defer cancelDeadline()

	received, err := stream(deadline, s, canvasID)
	if err != nil {
		if ctxErr := deadline.Err(); ctxErr != nil {
			slog.Info("stream ended on context",
				"reason", ctxErr,
				"notes_received", received)
			return nil
		}
		return fmt.Errorf("stream: %w", err)
	}
	slog.Info("stream completed", "notes_received", received)
	return nil
}

// stream opens a typed subscription and prints one line per note. The
// server re-emits the full snapshot whenever something changes, so the
// same note id is expected to appear multiple times — that is normal.
func stream(ctx context.Context, s *canvus.Session, canvasID string) (int, error) {
	events, err := s.SubscribeNotes(ctx, canvasID)
	if err != nil {
		return 0, fmt.Errorf("subscribe: %w", err)
	}
	var received int
	for note := range events {
		received++
		slog.Info("note",
			"ts", time.Now().UTC().Format(time.RFC3339Nano),
			"id", note.ID,
			"text", truncate(note.Text, 60))
	}
	return received, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func streamDuration() time.Duration {
	if v := os.Getenv("STREAM_DURATION_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
		slog.Warn("STREAM_DURATION_SECONDS could not be parsed; using default",
			"value", v, "default_seconds", defaultDurationSeconds)
	}
	return defaultDurationSeconds * time.Second
}

func mustEnv(name string) (string, error) {
	v := os.Getenv(name)
	if v == "" {
		return "", fmt.Errorf("missing required env var %s", name)
	}
	return v, nil
}

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
