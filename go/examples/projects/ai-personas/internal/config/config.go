// Package config loads ai-personas runtime settings from environment
// variables.
//
// Required:
//
//	CANVUS_API_URL  Base URL for the Canvus API, e.g. https://server/api/v1/
//	CANVUS_API_KEY  Private-Token API key
//	CANVAS_ID       Canvas to monitor
//	GEMINI_API_KEY  Google Gemini API key
//	OPENAI_API_KEY  OpenAI API key (DALL-E persona portraits)
//
// Optional:
//
//	GEMINI_MODEL_PERSONAS   Persona generation model (default: gemini-2.5-flash)
//	GEMINI_MODEL_CHAT       Persona Q&A model (default: gemini-2.5-flash)
//	LLM_TEMP                Gemini sampling temperature (default: 0.7)
//	CHAT_TOKEN_LIMIT        Max characters for persona answers (default: 256)
//	QUESTION_TIMEOUT        Question wait timeout, e.g. "5m" (default: 5m)
//	SNAPSHOT_DRAIN_SECONDS  Suppress pre-existing widget triggers (default: 3)
//	PORT or WEB_PORT        HTTP port for the QR / question dashboard (default: 8080)
//	PUBLIC_WEB_URL          Public URL advertised on the QR code (default: auto-detected)
//	SHUTDOWN_TIMEOUT        Graceful shutdown deadline, e.g. "30s" (default: 30s)
//	LOG_LEVEL               debug|info|warn|error (default: info)
//	LOG_FORMAT              json|text (default: text)
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime settings for ai-personas.
type Config struct {
	// Canvus
	CanvusAPIURL string
	CanvusAPIKey string
	CanvasID     string

	// LLM keys
	GeminiAPIKey string
	OpenAIAPIKey string

	// Gemini models
	PersonasModel string
	ChatModel     string
	LLMTemp       float32

	// App behavior
	ChatTokenLimit       int
	QuestionTimeout      time.Duration
	SnapshotDrainSeconds int
	ShutdownTimeout      time.Duration

	// Web server
	WebPort      string
	PublicWebURL string
}

// Load reads environment variables and returns a populated Config.
// Returns an error if any required variable is missing.
func Load() (*Config, error) {
	required := map[string]string{
		"CANVUS_API_URL": os.Getenv("CANVUS_API_URL"),
		"CANVUS_API_KEY": os.Getenv("CANVUS_API_KEY"),
		"CANVAS_ID":      os.Getenv("CANVAS_ID"),
		"GEMINI_API_KEY": os.Getenv("GEMINI_API_KEY"),
		"OPENAI_API_KEY": os.Getenv("OPENAI_API_KEY"),
	}
	for k, v := range required {
		if v == "" {
			return nil, fmt.Errorf("missing required environment variable %s", k)
		}
	}

	port := envOr("PORT", envOr("WEB_PORT", "8080"))

	return &Config{
		CanvusAPIURL: required["CANVUS_API_URL"],
		CanvusAPIKey: required["CANVUS_API_KEY"],
		CanvasID:     required["CANVAS_ID"],
		GeminiAPIKey: required["GEMINI_API_KEY"],
		OpenAIAPIKey: required["OPENAI_API_KEY"],

		PersonasModel: envOr("GEMINI_MODEL_PERSONAS", "gemini-2.5-flash"),
		ChatModel:     envOr("GEMINI_MODEL_CHAT", "gemini-2.5-flash"),
		LLMTemp:       float32(envFloat64("LLM_TEMP", 0.7)),

		ChatTokenLimit:       envInt("CHAT_TOKEN_LIMIT", 256),
		QuestionTimeout:      envDuration("QUESTION_TIMEOUT", 5*time.Minute),
		SnapshotDrainSeconds: envInt("SNAPSHOT_DRAIN_SECONDS", 3),
		ShutdownTimeout:      envDuration("SHUTDOWN_TIMEOUT", 30*time.Second),

		WebPort:      port,
		PublicWebURL: os.Getenv("PUBLIC_WEB_URL"),
	}, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envFloat64(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

// envDuration parses an env var either as seconds (bare integer) or as a
// time.Duration string ("5m", "300s", "1h").
func envDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	if n, err := strconv.Atoi(v); err == nil {
		return time.Duration(n) * time.Second
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	return def
}
