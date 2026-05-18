// Command 07-webhooks-notifications subscribes to the widgets endpoint of
// a Canvus canvas and POSTs a small JSON event to WEBHOOK_URL for every
// widget that appears AFTER the watcher starts (initial snapshot ignored).
//
// The Canvus API does not have native outbound webhooks; this example
// demonstrates the canonical pattern for synthesising them on top of the
// streaming subscribe endpoint.
//
// On non-2xx webhook responses, the POST is retried with exponential
// backoff (2s, 4s, 8s; up to 3 attempts total).
package main

import (
	"bufio"
	"bytes"
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
	"sync"
	"syscall"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

const (
	maxRetries = 3
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
		session:    s,
		canvasID:   canvasID,
		webhookURL: webhookURL,
		seen:       make(map[string]struct{}),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	slog.Info("webhook bridge starting",
		"canvas_id", canvasID,
		"webhook_url", webhookURL,
		"max_retries", maxRetries)
	return b.subscribe(ctx)
}

type bridge struct {
	session    *canvus.Session
	canvasID   string
	webhookURL string
	httpClient *http.Client

	mu              sync.Mutex
	seen            map[string]struct{}
	initialSnapshot bool
}

// subscribe reads NDJSON frames from /canvases/{id}/widgets?subscribe=true.
func (b *bridge) subscribe(ctx context.Context) error {
	u, err := url.Parse(b.session.BaseURL)
	if err != nil {
		return fmt.Errorf("parse base URL: %w", err)
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("canvases/%s/widgets", b.canvasID))
	q := u.Query()
	q.Set("subscribe", "true")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Accept", "application/x-ndjson, application/json")

	resp, err := b.session.HTTPClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("subscribe: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		if err := b.handleFrame(ctx, line); err != nil {
			slog.Warn("handleFrame error", "err", err)
		}
	}
	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			slog.Info("bridge cancelled cleanly")
			return nil
		}
		return fmt.Errorf("scan: %w", err)
	}
	return nil
}

// handleFrame decodes one frame, identifies new widgets, and dispatches a
// webhook for each.
func (b *bridge) handleFrame(ctx context.Context, line []byte) error {
	var widgets []canvus.Widget
	if err := json.Unmarshal(line, &widgets); err != nil {
		return fmt.Errorf("decode widgets: %w", err)
	}

	b.mu.Lock()
	firstSnapshot := !b.initialSnapshot
	var fresh []canvus.Widget
	for _, w := range widgets {
		if _, seen := b.seen[w.ID]; seen {
			continue
		}
		b.seen[w.ID] = struct{}{}
		if firstSnapshot {
			continue
		}
		fresh = append(fresh, w)
	}
	b.initialSnapshot = true
	b.mu.Unlock()

	if firstSnapshot {
		slog.Info("initial snapshot recorded",
			"existing_widget_count", len(widgets))
		return nil
	}

	for _, w := range fresh {
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
	return nil
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
