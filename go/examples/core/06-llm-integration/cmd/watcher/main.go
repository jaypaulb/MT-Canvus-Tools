// Command watcher subscribes to notes on a Canvus canvas, watches for new
// "question notes" (text starting with `?`), forwards each question to a
// local Ollama LLM, and posts the answer back to the canvas as a sibling
// note positioned to the right of the question.
//
// Deduplication: the server emits the full current note list on every
// frame (snapshots, not deltas). We track every note ID we have seen and
// answer only IDs that are new and start with `?`. The initial snapshot is
// recorded but its existing notes are NOT answered — only notes that
// appear after the watcher started are eligible.
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
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

const (
	defaultOllamaURL   = "http://localhost:11434"
	defaultOllamaModel = "llama3.2"
	answerColor        = "#1D71B8FF" // MT blue, full alpha.
	answerOffsetX      = 400.0       // pixels right of the question note.
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
		session:   s,
		canvasID:  canvasID,
		ollamaURL: ollamaURL,
		model:     model,
		seen:      make(map[string]struct{}),
	}
	slog.Info("watcher starting",
		"canvas_id", canvasID,
		"ollama_url", ollamaURL,
		"model", model,
		"answer_color", answerColor)
	return w.subscribe(ctx)
}

type watcher struct {
	session   *canvus.Session
	canvasID  string
	ollamaURL string
	model     string

	mu              sync.Mutex
	seen            map[string]struct{}
	initialSnapshot bool
}

// subscribe opens the streaming notes endpoint and processes each frame.
func (w *watcher) subscribe(ctx context.Context) error {
	u, err := url.Parse(w.session.BaseURL)
	if err != nil {
		return fmt.Errorf("parse base URL: %w", err)
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("canvases/%s/notes", w.canvasID))
	q := u.Query()
	q.Set("subscribe", "true")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Accept", "application/x-ndjson, application/json")

	resp, err := w.session.HTTPClient.Do(req)
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
		if err := w.handleFrame(ctx, line); err != nil {
			slog.Warn("handleFrame error", "err", err)
		}
	}
	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			slog.Info("watcher cancelled cleanly")
			return nil
		}
		return fmt.Errorf("scan: %w", err)
	}
	return nil
}

// handleFrame decodes one NDJSON frame into a []Note, then handles any
// newly observed question notes.
func (w *watcher) handleFrame(ctx context.Context, line []byte) error {
	var notes []canvus.Note
	if err := json.Unmarshal(line, &notes); err != nil {
		return fmt.Errorf("decode notes: %w", err)
	}

	w.mu.Lock()
	firstSnapshot := !w.initialSnapshot
	var newQuestions []canvus.Note
	for _, n := range notes {
		if _, seen := w.seen[n.ID]; seen {
			continue
		}
		w.seen[n.ID] = struct{}{}
		// Skip the initial snapshot — only react to notes that appear
		// after the watcher started.
		if firstSnapshot {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(n.Text), "?") {
			newQuestions = append(newQuestions, n)
		}
	}
	w.initialSnapshot = true
	w.mu.Unlock()

	if firstSnapshot {
		slog.Info("initial snapshot recorded",
			"existing_note_count", len(notes),
			"existing_questions_skipped", countQuestions(notes))
		return nil
	}

	for _, q := range newQuestions {
		if err := w.answer(ctx, q); err != nil {
			slog.Warn("answer failed",
				"question_id", q.ID,
				"err", err)
		}
	}
	return nil
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

	loc := q.Location
	answerReq := map[string]any{
		"widget_type":      "note",
		"text":             answerText,
		"background_color": answerColor,
	}
	if loc != nil {
		answerReq["location"] = map[string]any{
			"x": loc.X + answerOffsetX,
			"y": loc.Y,
		}
	}
	created, err := w.session.CreateNote(ctx, w.canvasID, answerReq)
	if err != nil {
		return fmt.Errorf("CreateNote (answer): %w", err)
	}

	// Mark the answer note as seen so the next snapshot frame doesn't
	// trigger another round (defensive — the answer doesn't start with `?`
	// but a bug here would be a runaway loop, so be explicit).
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
		return "", fmt.Errorf("ollama returned empty response")
	}
	return strings.TrimSpace(gr.Response), nil
}

func countQuestions(notes []canvus.Note) int {
	n := 0
	for _, note := range notes {
		if strings.HasPrefix(strings.TrimSpace(note.Text), "?") {
			n++
		}
	}
	return n
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
