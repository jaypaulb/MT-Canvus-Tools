// Command watcher subscribes to notes on a Canvus canvas via the typed
// Session.SubscribeNotes helper, watches for new "question notes" (text
// starting with `?`), forwards each to a local Ollama LLM, and posts the
// answer back to the canvas as a sibling note positioned to the right.
//
// Dedup strategy: the typed Subscribe channel yields each note as the
// server emits it. The server re-sends the full snapshot every time
// something changes, so we keep a set of seen IDs and skip duplicates.
// To avoid answering pre-existing questions on startup we apply a
// snapshot-drain window: any ID we see during the first
// SNAPSHOT_DRAIN_SECONDS (default 2s) is marked seen but not answered.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

const (
	defaultOllamaURL         = "http://localhost:11434"
	defaultOllamaModel       = "llama3.2"
	defaultSnapshotDrainSecs = 2
	answerColor              = "#1D71B8FF" // MT blue, full alpha.
	answerOffsetX            = 400.0       // pixels right of the question note.
)

type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type generateResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func main() {
	setupLogging()
	if err := run(); err != nil {
		slog.Error("watcher failed", "err", err)
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
	ollamaURL := strings.TrimRight(envDefault("OLLAMA_URL", defaultOllamaURL), "/")
	model := envDefault("OLLAMA_MODEL", defaultOllamaModel)

	cfg := &canvus.SessionConfig{
		BaseURL:        baseURL,
		RequestTimeout: 10 * time.Minute, // long-lived subscription.
	}
	s := canvus.NewSession(cfg, canvus.WithAPIKey(apiKey))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	w := &watcher{
		session:       s,
		canvasID:      canvasID,
		ollamaURL:     ollamaURL,
		model:         model,
		seen:          make(map[string]struct{}),
		drainDuration: snapshotDrainDuration(),
	}
	slog.Info("watcher starting",
		"canvas_id", canvasID,
		"ollama_url", ollamaURL,
		"model", model,
		"answer_color", answerColor,
		"snapshot_drain_seconds", int(w.drainDuration.Seconds()))
	return w.subscribe(ctx)
}

type watcher struct {
	session       *canvus.Session
	canvasID      string
	ollamaURL     string
	model         string
	drainDuration time.Duration

	mu              sync.Mutex
	seen            map[string]struct{}
	snapshotDrained bool
}

// subscribe opens the typed notes channel and processes each yield.
func (w *watcher) subscribe(ctx context.Context) error {
	events, err := w.session.SubscribeNotes(ctx, w.canvasID)
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	// After drainDuration elapses, every previously-unseen note that arrives
	// is treated as fresh and may be answered. Pre-drain arrivals are simply
	// marked seen.
	drainTimer := time.NewTimer(w.drainDuration)
	defer drainTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("watcher cancelled cleanly")
			return nil
		case <-drainTimer.C:
			w.mu.Lock()
			snapshotSize := len(w.seen)
			w.snapshotDrained = true
			w.mu.Unlock()
			slog.Info("snapshot drain window elapsed", "existing_notes_marked_seen", snapshotSize)
		case note, ok := <-events:
			if !ok {
				return nil
			}
			if err := w.handle(ctx, note); err != nil {
				slog.Warn("handle error", "id", note.ID, "err", err)
			}
		}
	}
}

// handle records the note id and answers it if it's new, post-drain, and
// starts with `?`.
func (w *watcher) handle(ctx context.Context, n canvus.Note) error {
	w.mu.Lock()
	_, alreadySeen := w.seen[n.ID]
	w.seen[n.ID] = struct{}{}
	drained := w.snapshotDrained
	w.mu.Unlock()

	if alreadySeen {
		return nil
	}
	if !drained {
		return nil
	}
	if !strings.HasPrefix(strings.TrimSpace(n.Text), "?") {
		return nil
	}
	return w.answer(ctx, n)
}

// answer runs one question through Ollama and posts the response back to
// the canvas as a sibling note.
func (w *watcher) answer(ctx context.Context, q canvus.Note) error {
	prompt := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(q.Text), "?"))
	slog.Info("question received",
		"question_id", q.ID,
		"prompt", truncate(prompt, 80))

	answerText, err := w.askOllama(ctx, prompt)
	if err != nil {
		return fmt.Errorf("ask ollama: %w", err)
	}
	slog.Info("ollama responded",
		"question_id", q.ID,
		"response_len", len(answerText))

	answerReq := map[string]any{
		"widget_type":      "note",
		"text":             answerText,
		"background_color": answerColor,
	}
	if loc := q.Location; loc != nil {
		answerReq["location"] = map[string]any{
			"x": loc.X + answerOffsetX,
			"y": loc.Y,
		}
	}
	created, err := w.session.CreateNote(ctx, w.canvasID, answerReq)
	if err != nil {
		return fmt.Errorf("CreateNote (answer): %w", err)
	}

	// Mark the answer note as seen so it can't trigger a self-answer loop.
	w.mu.Lock()
	w.seen[created.ID] = struct{}{}
	w.mu.Unlock()

	slog.Info("answer posted",
		"question_id", q.ID,
		"answer_id", created.ID)
	return nil
}

// askOllama posts the prompt to /api/generate and returns the response text.
func (w *watcher) askOllama(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(generateRequest{
		Model:  w.model,
		Prompt: prompt,
		Stream: false,
	})
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	// Generous timeout: small LLMs on CPU can take 30s+ for a short answer.
	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost,
		w.ollamaURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("POST /api/generate: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("ollama HTTP %d: %s", resp.StatusCode, b)
	}

	var gr generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	if gr.Response == "" {
		return "", errors.New("ollama returned empty response")
	}
	return strings.TrimSpace(gr.Response), nil
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func mustEnv(name string) (string, error) {
	v := os.Getenv(name)
	if v == "" {
		return "", fmt.Errorf("missing required env var %s", name)
	}
	return v, nil
}

func envDefault(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
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
