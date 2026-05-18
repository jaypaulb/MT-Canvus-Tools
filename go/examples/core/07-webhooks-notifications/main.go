// Command 07-webhooks-notifications subscribes to the widgets endpoint of
// a Canvus canvas using Session.SubscribeWidgets and POSTs a small JSON
// event to WEBHOOK_URL for every widget that appears AFTER the watcher
// starts.
//
// The Canvus API does not have native outbound webhooks; this example
// demonstrates the canonical pattern for synthesising them on top of the
// streaming subscribe endpoint.
//
// Dedup strategy: the typed channel yields each widget as the server
// emits it; snapshot frames are re-sent on every change. We track seen
// widget IDs and apply a SNAPSHOT_DRAIN_SECONDS (default 2s) window at
// startup during which existing widgets are merely recorded, not
// forwarded — so a restart doesn't re-fire webhooks for everything on
// the canvas.
//
// On non-2xx webhook responses, the POST is retried with exponential
// backoff (2s, 4s, 8s; up to 3 attempts total).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

const (
	maxRetries               = 3
	defaultSnapshotDrainSecs = 2
)

var retryBackoff = []time.Duration{
	2 * time.Second,
	4 * time.Second,
	8 * time.Second,
}

type webhookEvent struct {
	Event      string    `json:"event"`
	CanvasID   string    `json:"canvas_id"`
	WidgetID   string    `json:"widget_id"`
	WidgetType string    `json:"widget_type"`
	Timestamp  time.Time `json:"timestamp"`
}

func main() {
	setupLogging()
	if err := run(); err != nil {
		slog.Error("webhooks-notifications failed", "err", err)
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
	webhookURL, err := mustEnv("WEBHOOK_URL")
	if err != nil {
		return err
	}

	cfg := &canvus.SessionConfig{
		BaseURL:        baseURL,
		RequestTimeout: 10 * time.Minute,
	}
	s := canvus.NewSession(cfg, canvus.WithAPIKey(apiKey))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	b := &bridge{
		session:       s,
		canvasID:      canvasID,
		webhookURL:    webhookURL,
		seen:          make(map[string]struct{}),
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		drainDuration: snapshotDrainDuration(),
	}
	slog.Info("webhook bridge starting",
		"canvas_id", canvasID,
		"webhook_url", webhookURL,
		"max_retries", maxRetries,
		"snapshot_drain_seconds", int(b.drainDuration.Seconds()))
	return b.subscribe(ctx)
}

type bridge struct {
	session       *canvus.Session
	canvasID      string
	webhookURL    string
	httpClient    *http.Client
	drainDuration time.Duration

	mu              sync.Mutex
	seen            map[string]struct{}
	snapshotDrained bool
}

// subscribe opens the typed widgets channel and dispatches per arrival.
func (b *bridge) subscribe(ctx context.Context) error {
	events, err := b.session.SubscribeWidgets(ctx, b.canvasID)
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	drainTimer := time.NewTimer(b.drainDuration)
	defer drainTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("bridge cancelled cleanly")
			return nil
		case <-drainTimer.C:
			b.mu.Lock()
			snapshotSize := len(b.seen)
			b.snapshotDrained = true
			b.mu.Unlock()
			slog.Info("snapshot drain window elapsed", "existing_widgets_marked_seen", snapshotSize)
		case w, ok := <-events:
			if !ok {
				return nil
			}
			b.handle(ctx, w)
		}
	}
}

// handle records the widget id and, if it's a fresh post-drain arrival,
// dispatches a webhook for it.
func (b *bridge) handle(ctx context.Context, w canvus.Widget) {
	b.mu.Lock()
	_, alreadySeen := b.seen[w.ID]
	b.seen[w.ID] = struct{}{}
	drained := b.snapshotDrained
	b.mu.Unlock()

	if alreadySeen || !drained {
		return
	}
	evt := webhookEvent{
		Event:      "widget.created",
		CanvasID:   b.canvasID,
		WidgetID:   w.ID,
		WidgetType: w.WidgetType,
		Timestamp:  time.Now().UTC(),
	}
	if err := b.deliver(ctx, evt); err != nil {
		slog.Warn("webhook delivery failed",
			"widget_id", w.ID,
			"widget_type", w.WidgetType,
			"err", err)
	}
}

// deliver POSTs the event to WEBHOOK_URL with exponential-backoff retry on
// non-2xx and on transport errors.
func (b *bridge) deliver(ctx context.Context, evt webhookEvent) error {
	body, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			wait := retryBackoff[attempt-1]
			slog.Info("retrying webhook",
				"attempt", attempt+1,
				"wait", wait.String(),
				"widget_id", evt.WidgetID)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}

		start := time.Now()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost,
			b.webhookURL, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("new request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := b.httpClient.Do(req)
		latency := time.Since(start)
		if err != nil {
			lastErr = err
			slog.Warn("webhook transport error",
				"attempt", attempt+1,
				"err", err,
				"latency_ms", latency.Milliseconds())
			continue
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			slog.Info("webhook delivered",
				"widget_id", evt.WidgetID,
				"widget_type", evt.WidgetType,
				"status", resp.StatusCode,
				"attempt", attempt+1,
				"latency_ms", latency.Milliseconds())
			return nil
		}

		lastErr = fmt.Errorf("non-2xx status %d", resp.StatusCode)
		slog.Warn("webhook non-2xx",
			"attempt", attempt+1,
			"status", resp.StatusCode,
			"latency_ms", latency.Milliseconds())
	}
	return fmt.Errorf("gave up after %d attempts: %w", maxRetries, lastErr)
}

func snapshotDrainDuration() time.Duration {
	if v := os.Getenv("SNAPSHOT_DRAIN_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return time.Duration(n) * time.Second
		}
		slog.Warn("SNAPSHOT_DRAIN_SECONDS could not be parsed; using default",
			"value", v, "default_seconds", defaultSnapshotDrainSecs)
	}
	return defaultSnapshotDrainSecs * time.Second
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
