// Command llm-canvas-companion monitors a Canvus canvas for AI trigger widgets
// and routes them to the appropriate LLM workflow.
//
// # Trigger types
//
//   - Note with {{content}}: routes to text or image generation (auto-decided by LLM)
//   - Image titled "Snapshot at …": OCR via Google Vision
//   - Image titled "AI_Icon_PDF_Precis" placed on a PDF: chunk + summarise
//   - Image titled "AI_Icon_Canvus_Precis" on the shared canvas: full canvas summary
//
// # Configuration
//
// All settings are environment variables; see internal/config/config.go for the
// full list. Required: CANVUS_API_URL, CANVUS_API_KEY, CANVAS_ID.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/llm-canvas-companion/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/llm-canvas-companion/internal/monitor"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func main() {
	setupLogging()
	if err := run(); err != nil {
		slog.Error("llm-canvas-companion failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	slog.Info("llm-canvas-companion starting",
		"canvas_id", cfg.CanvasID,
		"canvas_name", cfg.CanvasName,
		"canvus_url", cfg.CanvusAPIURL,
		"base_llm_url", cfg.BaseLLMURL,
		"drain_seconds", cfg.SnapshotDrainSeconds)

	// Create downloads directory.
	if err := os.MkdirAll(cfg.DownloadsDir, 0o755); err != nil {
		return fmt.Errorf("mkdir downloads: %w", err)
	}

	// Build SDK session.
	sessionCfg := &canvus.SessionConfig{
		BaseURL:        cfg.CanvusAPIURL,
		RequestTimeout: 10 * time.Minute, // long-lived subscription
	}
	s := canvus.NewSession(sessionCfg, canvus.WithAPIKey(cfg.CanvusAPIKey))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	m := monitor.New(s, cfg)
	if err := m.Run(ctx); err != nil {
		return fmt.Errorf("monitor: %w", err)
	}

	slog.Info("llm-canvas-companion stopped cleanly")
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
