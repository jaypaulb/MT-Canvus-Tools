// Command ai-personas hosts a persona-driven Q&A workflow against a Canvus
// canvas. It monitors the canvas for trigger notes ("Create_Personas" and
// "New_AI_Question") plus connector creation events, generates customer
// personas using Google Gemini, draws answer/meta-answer grids, and serves
// a small QR-coded web dashboard so remote attendees can submit questions.
//
// Required environment variables: CANVUS_API_URL, CANVUS_API_KEY, CANVAS_ID,
// GEMINI_API_KEY, OPENAI_API_KEY. See internal/config/config.go for the full
// list and defaults.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/ai"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/atom"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/monitor"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/persona"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/qa"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/web"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func main() {
	setupLogging()
	if err := run(); err != nil {
		slog.Error("ai-personas failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	slog.Info("ai-personas starting",
		"canvus_url", cfg.CanvusAPIURL,
		"canvas_id", cfg.CanvasID,
		"personas_model", cfg.PersonasModel,
		"chat_model", cfg.ChatModel,
		"gemini_key", atom.MaskKey(cfg.GeminiAPIKey),
		"openai_key", atom.MaskKey(cfg.OpenAIAPIKey),
		"canvus_key", atom.MaskKey(cfg.CanvusAPIKey),
	)

	rootCtx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// SDK session — long timeout because the widget subscribe is long-lived.
	session := canvus.NewSession(&canvus.SessionConfig{
		BaseURL:        cfg.CanvusAPIURL,
		RequestTimeout: 10 * time.Minute,
	}, canvus.WithAPIKey(cfg.CanvusAPIKey))

	// Validate API access before we start the monitor.
	if err := validateStartup(rootCtx, cfg, session); err != nil {
		return fmt.Errorf("startup validation: %w", err)
	}

	gemini, err := ai.NewGemini(rootCtx, cfg.GeminiAPIKey,
		cfg.PersonasModel, cfg.ChatModel, cfg.LLMTemp)
	if err != nil {
		return err
	}
	openai, err := ai.NewOpenAI(cfg.OpenAIAPIKey)
	if err != nil {
		return err
	}

	personaStore := persona.NewStore()
	helpers := qa.NewHelperTracker()

	personaWorkflow := &persona.Workflow{
		Session:  session,
		CanvasID: cfg.CanvasID,
		Gemini:   gemini,
		OpenAI:   openai,
		Store:    personaStore,
	}
	qaWorkflow := qa.NewWorkflow(session, cfg.CanvasID, gemini, openai,
		personaWorkflow, personaStore, helpers,
		cfg.ChatTokenLimit, cfg.QuestionTimeout)

	// Web dashboard (QR + question form + /health).
	srv := web.New(session, cfg.CanvasID, cfg.WebPort, cfg.PublicWebURL)
	srv.Start(rootCtx)

	// Monitor + graceful shutdown wait group.
	var handlerWG sync.WaitGroup
	mon := monitor.New(session, cfg.CanvasID, cfg.SnapshotDrainSeconds,
		personaWorkflow, qaWorkflow, &handlerWG)
	monErr := mon.Run(rootCtx)

	// Wait for in-flight handlers to complete, bounded by ShutdownTimeout.
	slog.Info("waiting for in-flight handlers", "timeout", cfg.ShutdownTimeout)
	done := make(chan struct{})
	go func() { handlerWG.Wait(); close(done) }()
	select {
	case <-done:
		slog.Info("all handlers completed")
	case <-time.After(cfg.ShutdownTimeout):
		slog.Warn("shutdown timeout reached; some handlers may still be running")
	}
	if monErr != nil && !errors.Is(monErr, context.Canceled) {
		return fmt.Errorf("monitor: %w", monErr)
	}
	slog.Info("ai-personas stopped cleanly")
	return nil
}

// validateStartup proves we can talk to Canvus (the SDK call also exercises
// the API key) and to OpenAI before launching the long-running monitor. The
// Gemini key is validated implicitly the first time GenerateContent runs.
func validateStartup(ctx context.Context, cfg *config.Config, s *canvus.Session) error {
	probeCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if _, err := s.GetCanvas(probeCtx, cfg.CanvasID); err != nil {
		return fmt.Errorf("Canvus probe: %w", err)
	}
	oa, err := ai.NewOpenAI(cfg.OpenAIAPIKey)
	if err != nil {
		return err
	}
	if err := oa.ValidateKey(probeCtx); err != nil {
		return fmt.Errorf("OpenAI probe: %w", err)
	}
	return nil
}

func setupLogging() {
	var h slog.Handler
	switch os.Getenv("LOG_FORMAT") {
	case "json":
		h = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel()})
	default:
		h = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel()})
	}
	slog.SetDefault(slog.New(h))
}

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
