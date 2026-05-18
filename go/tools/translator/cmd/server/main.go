// Command server is the CanvusTranslator web service.
//
// It serves a small single-page frontend that lets a user pick a target
// language; on POST /translate the server fetches all notes from a configured
// Canvus canvas, translates each one via Google Gemini, and patches them back
// in place using the MT-Canvus-Tools Go SDK.
//
// Configuration is entirely through environment variables — see the README
// for the full list. The service is stateless and safe to run as a single
// instance; in-flight translations are bounded by a concurrency semaphore.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/translator/internal"
)

func main() {
	setupLogging()
	if err := run(); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := internal.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Build the Canvus session and Gemini client once at startup so they are
	// shared across all translation requests.
	cs, err := internal.NewCanvusSession(cfg)
	if err != nil {
		return fmt.Errorf("create canvus session: %w", err)
	}

	gs, err := internal.NewGeminiClient(context.Background(), cfg.GeminiAPIKey)
	if err != nil {
		return fmt.Errorf("create gemini client: %w", err)
	}

	mux := http.NewServeMux()

	// Health-check endpoint — used by orchestration to verify the service is up.
	mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			slog.Error("ping: encode response", "err", err)
		}
	})

	mux.HandleFunc("/translate", makeTranslateHandler(cs, gs, cfg.CanvasID))

	// Static frontend — index.html, style.css, script.js, flags/*.svg.
	// The server's working directory must include the web/static tree; see README.
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/", fs)

	addr := ":" + cfg.Port
	slog.Info("canvus-translator starting", "addr", addr, "canvas_id", cfg.CanvasID)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 5 * time.Minute, // translation of a large canvas can take time.
		IdleTimeout:  60 * time.Second,
	}
	return srv.ListenAndServe()
}

// translateRequest is the POST body for /translate.
type translateRequest struct {
	Language string `json:"language"`
}

// makeTranslateHandler returns the /translate handler closed over the shared
// Canvus session, Gemini client, and canvas ID.
func makeTranslateHandler(cs *internal.CanvusSession, gs *internal.GeminiClient, canvasID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "only POST requests are accepted", http.StatusMethodNotAllowed)
			return
		}

		var req translateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.Language == "" {
			http.Error(w, "language field is required", http.StatusBadRequest)
			return
		}

		slog.Info("translation requested", "language", req.Language, "canvas_id", canvasID)

		notes, err := cs.ListNotes(r.Context(), canvasID)
		if err != nil {
			slog.Error("list notes failed", "err", err)
			http.Error(w, fmt.Sprintf("failed to list notes: %v", err), http.StatusInternalServerError)
			return
		}

		slog.Info("notes retrieved", "count", len(notes))

		// maxConcurrency bounds the number of simultaneous Gemini calls.
		// Gemini Flash is fast but rate-limited; 20 is a conservative ceiling.
		const maxConcurrency = 20
		sem := make(chan struct{}, maxConcurrency)
		var wg sync.WaitGroup

		for _, note := range notes {
			sem <- struct{}{}
			wg.Add(1)

			go func(noteID, originalText string) {
				defer wg.Done()
				defer func() { <-sem }()

				if originalText == "" {
					slog.Debug("skipping empty note", "note_id", noteID)
					return
				}

				translated, err := gs.TranslateText(r.Context(), originalText, req.Language)
				if err != nil {
					// Log and continue — one note failure should not abort the batch.
					slog.Error("translate note failed", "note_id", noteID, "err", err)
					return
				}

				if err := cs.UpdateNoteText(r.Context(), canvasID, noteID, translated); err != nil {
					slog.Error("update note failed", "note_id", noteID, "err", err)
					return
				}

				slog.Info("note translated", "note_id", noteID)
			}(note.ID, note.Text)
		}

		wg.Wait()
		slog.Info("translation batch complete", "language", req.Language, "notes", len(notes))

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Translation to %s complete for %d notes.", req.Language, len(notes))
	}
}

// setupLogging configures slog per the monorepo convention (go.md §5):
// LOG_FORMAT=json → JSONHandler, otherwise TextHandler; output to stderr.
func setupLogging() {
	var h slog.Handler
	level := logLevel()
	switch os.Getenv("LOG_FORMAT") {
	case "json":
		h = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	default:
		h = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	}
	slog.SetDefault(slog.New(h))
}

// logLevel maps the LOG_LEVEL env var to a slog.Level, defaulting to Info.
func logLevel() slog.Level {
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
