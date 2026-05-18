// Command 05-streaming subscribes to the notes endpoint of a Canvus canvas
// and prints summary lines for each NDJSON event the server emits, exiting
// cleanly on SIGINT or after STREAM_DURATION_SECONDS elapses.
//
// The Canvus SDK does not yet expose a typed Subscribe helper, so this
// example builds the HTTP request directly using the Session's exported
// HTTPClient field — the API-key round-tripper installed by WithAPIKey is
// still in effect, so authentication "just works" without re-implementing
// the header logic.
//
// Wire shape: GET /api/v1/canvases/{id}/notes?subscribe=true returns a
// streaming response of newline-delimited JSON. Each frame is the full
// current list of notes, not a delta — clients receive the latest snapshot
// on connect plus a fresh snapshot every time something changes. Empty
// lines are keep-alives and must be ignored.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strconv"
	"sync/atomic"
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

	stats, err := stream(deadline, s, canvasID)
	if err != nil {
		// Cancellation is not a failure — it's our exit condition.
		if ctxErr := deadline.Err(); ctxErr != nil {
			slog.Info("stream ended on context",
				"reason", ctxErr,
				"frames", stats.frames,
				"non_empty_frames", stats.nonEmpty)
			return nil
		}
		return fmt.Errorf("stream: %w", err)
	}
	slog.Info("stream completed", "frames", stats.frames, "non_empty_frames", stats.nonEmpty)
	return nil
}

type streamStats struct {
	frames   int64
	nonEmpty int64
}

// stream issues the subscribe request and prints one summary line per
// non-empty frame.
func stream(ctx context.Context, s *canvus.Session, canvasID string) (streamStats, error) {
	var stats streamStats

	u, err := url.Parse(s.BaseURL)
	if err != nil {
		return stats, fmt.Errorf("parse base URL: %w", err)
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("canvases/%s/notes", canvasID))
	q := u.Query()
	q.Set("subscribe", "true")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return stats, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Accept", "application/x-ndjson, application/json")
	// Authentication is added by the round-tripper installed by WithAPIKey.

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return stats, fmt.Errorf("do: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return stats, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	// Notes responses can be many KB for busy canvases.
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		atomic.AddInt64(&stats.frames, 1)
		if len(line) == 0 {
			// Keep-alive frame per the SDK observation; skip silently.
			continue
		}
		atomic.AddInt64(&stats.nonEmpty, 1)
		printFrame(line)
	}
	if err := scanner.Err(); err != nil {
		return stats, fmt.Errorf("scan: %w", err)
	}
	return stats, nil
}

// printFrame decodes a single NDJSON frame and prints a one-line summary.
// The server emits arrays of notes; if the shape ever diverges we still
// print the raw byte count so the caller sees activity.
func printFrame(line []byte) {
	var notes []canvus.Note
	if err := json.Unmarshal(line, &notes); err != nil {
		slog.Warn("frame did not parse as []Note", "err", err, "bytes", len(line))
		return
	}
	var firstID, firstText string
	if len(notes) > 0 {
		firstID = notes[0].ID
		firstText = truncate(notes[0].Text, 60)
	}
	slog.Info("frame",
		"ts", time.Now().UTC().Format(time.RFC3339Nano),
		"count", len(notes),
		"first_id", firstID,
		"first_text", firstText)
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

// mustEnv returns the value of name or an error if it is empty/unset.
func mustEnv(name string) (string, error) {
	v := os.Getenv(name)
	if v == "" {
		return "", fmt.Errorf("missing required env var %s", name)
	}
	return v, nil
}

// setupLogging configures slog.Default.
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
