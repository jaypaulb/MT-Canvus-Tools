// Command note-mapper is a web application that transforms a photograph of
// physical Post-it notes into digital notes on a Canvus canvas.
//
// # Usage
//
//	CANVUS_API_URL=https://server/api/v1 \
//	CANVUS_API_KEY=your-key \
//	GOOGLE_GENAI_API_KEY=your-gemini-key \
//	note-mapper
//
// The server listens on PORT (default 8080) and serves the web UI from ./web/.
//
// # Environment variables
//
//	CANVUS_API_URL      Canvus API base URL (https://host/api/v1)
//	CANVUS_API_KEY      Canvus API key (Private-Token header value)
//	GOOGLE_GENAI_API_KEY  Google Gemini API key
//	PORT                HTTP listen port (default: 8080)
//	LOG_FORMAT          "json" for structured JSON logs (default: text)
//	LOG_LEVEL           "debug", "warn", "error" (default: info)
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/note-mapper/internal/api"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/note-mapper/internal/config"
)

func main() {
	setupLogging()

	// Load Canvus credentials from environment on startup.
	// They can also be updated at runtime via POST /api/set-credentials.
	cfg := config.Load()
	slog.Info("note-mapper starting",
		"canvus_url_set", cfg.CanvusURL != "",
		"api_key_set", cfg.APIKey != "",
		"gemini_key_set", os.Getenv("GOOGLE_GENAI_API_KEY") != "",
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// API routes.
	mux.HandleFunc("/api/upload-image", api.UploadImageHandler)
	mux.HandleFunc("/api/scan-notes", api.ScanNotesHandler)
	mux.HandleFunc("/api/create-notes", api.CreateNotesHandler)
	mux.HandleFunc("/api/set-credentials", api.SetCredentialsHandler)
	mux.HandleFunc("/api/get-canvases", api.GetCanvasesHandler)
	mux.HandleFunc("/api/get-anchors", api.GetAnchorsOnlyHandler)
	mux.HandleFunc("/api/get-anchor-info", api.GetAnchorInfoHandler)

	// Static web UI.
	mux.Handle("/", http.FileServer(http.Dir("web")))

	slog.Info("listening", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

// setupLogging configures slog.Default based on LOG_FORMAT and LOG_LEVEL env vars.
func setupLogging() {
	var h slog.Handler
	opts := &slog.HandlerOptions{Level: logLevel()}
	switch os.Getenv("LOG_FORMAT") {
	case "json":
		h = slog.NewJSONHandler(os.Stderr, opts)
	default:
		h = slog.NewTextHandler(os.Stderr, opts)
	}
	slog.SetDefault(slog.New(h))
}

// logLevel returns the slog.Level indicated by LOG_LEVEL env var.
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
