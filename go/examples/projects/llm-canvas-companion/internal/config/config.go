// Package config loads llm-canvas-companion settings from environment variables.
//
// Environment variables (all optional unless noted):
//
//	CANVUS_API_URL          (required) Base URL for the Canvus API, e.g. https://server/api/v1/
//	CANVUS_API_KEY          (required) Private-Token API key
//	CANVAS_ID               (required) Canvas to monitor
//	CANVAS_NAME             Human-readable name shown on startup
//	OPENAI_API_KEY          API key for OpenAI / compatible endpoints
//	GOOGLE_VISION_API_KEY   API key for Google Vision OCR
//	BASE_LLM_URL            Base URL for text LLM (default: http://127.0.0.1:1234/v1)
//	TEXT_LLM_URL            Override for text generation only
//	IMAGE_LLM_URL           Override for image generation (default: https://api.openai.com/v1)
//	AZURE_OPENAI_ENDPOINT   Azure OpenAI endpoint URL
//	AZURE_OPENAI_DEPLOYMENT Azure OpenAI deployment name
//	AZURE_OPENAI_API_VERSION Azure API version (default: 2024-02-15-preview)
//	OPENAI_NOTE_MODEL       Chat model for note triggers (default: gpt-4)
//	OPENAI_CANVAS_MODEL     Chat model for canvas precis (default: gpt-4)
//	OPENAI_PDF_MODEL        Chat model for PDF precis (default: gpt-4)
//	IMAGE_GEN_MODEL         Image generation model (default: dall-e-3)
//	OPENAI_PDF_PRECIS_TOKENS    Max tokens for PDF summary (default: 1000)
//	OPENAI_CANVAS_PRECIS_TOKENS Max tokens for canvas summary (default: 600)
//	OPENAI_NOTE_RESPONSE_TOKENS Max tokens for note response (default: 400)
//	OPENAI_IMAGE_ANALYSIS_TOKENS Max tokens for image analysis (default: 16384)
//	OPENAI_ERROR_RESPONSE_TOKENS Max tokens for error notes (default: 200)
//	OPENAI_PDF_CHUNK_SIZE_TOKENS PDF chunk size in tokens (default: 20000)
//	OPENAI_PDF_MAX_CHUNKS_TOKENS Max PDF chunks (default: 10)
//	OPENAI_PDF_SUMMARY_RATIO     Summary ratio (default: 0.3)
//	MAX_RETRIES           API retry count (default: 3)
//	RETRY_DELAY           Seconds between retries (default: 1)
//	AI_TIMEOUT            Seconds for AI calls (default: 60)
//	PROCESSING_TIMEOUT    Seconds for long operations (default: 300)
//	MAX_CONCURRENT        Max concurrent AI jobs (default: 5)
//	MAX_FILE_SIZE         Max upload size in bytes (default: 52428800 = 50MB)
//	DOWNLOADS_DIR         Temp download directory (default: ./downloads)
//	SNAPSHOT_DRAIN_SECONDS Seconds to suppress pre-existing widget triggers (default: 3)
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime settings for llm-canvas-companion.
type Config struct {
	// Canvus connection
	CanvusAPIURL string
	CanvusAPIKey string
	CanvasID     string
	CanvasName   string

	// External AI keys
	OpenAIAPIKey    string
	GoogleVisionKey string

	// LLM endpoint routing
	BaseLLMURL  string
	TextLLMURL  string
	ImageLLMURL string

	// Azure OpenAI
	AzureOpenAIEndpoint   string
	AzureOpenAIDeployment string
	AzureOpenAIAPIVersion string

	// Model selection
	OpenAINoteModel   string
	OpenAICanvasModel string
	OpenAIPDFModel    string
	OpenAIImageModel  string

	// Token budgets
	PDFPrecisTokens       int64
	CanvasPrecisTokens    int64
	NoteResponseTokens    int64
	ImageAnalysisTokens   int64
	ErrorResponseTokens   int64
	PDFChunkSizeTokens    int64
	PDFMaxChunksTokens    int64
	PDFSummaryRatioTokens float64

	// Operational limits
	MaxRetries        int
	RetryDelay        time.Duration
	AITimeout         time.Duration
	ProcessingTimeout time.Duration
	MaxConcurrent     int
	MaxFileSize       int64
	DownloadsDir      string

	// Subscribe drain
	SnapshotDrainSeconds int
}

// Load reads all settings from environment variables.
// Returns an error if any required variable is missing.
func Load() (*Config, error) {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		openAIKey = os.Getenv("OPENAI_KEY") // legacy fallback
	}

	required := []string{"CANVUS_API_URL", "CANVUS_API_KEY", "CANVAS_ID"}
	for _, k := range required {
		if os.Getenv(k) == "" {
			return nil, fmt.Errorf("missing required environment variable %s", k)
		}
	}

	return &Config{
		// Canvus connection
		CanvusAPIURL: os.Getenv("CANVUS_API_URL"),
		CanvusAPIKey: os.Getenv("CANVUS_API_KEY"),
		CanvasID:     os.Getenv("CANVAS_ID"),
		CanvasName:   envOr("CANVAS_NAME", "(unnamed)"),

		// External AI keys
		OpenAIAPIKey:    openAIKey,
		GoogleVisionKey: os.Getenv("GOOGLE_VISION_API_KEY"),

		// LLM endpoint routing
		BaseLLMURL:  envOr("BASE_LLM_URL", "http://127.0.0.1:1234/v1"),
		TextLLMURL:  os.Getenv("TEXT_LLM_URL"),
		ImageLLMURL: envOr("IMAGE_LLM_URL", "https://api.openai.com/v1"),

		// Azure OpenAI
		AzureOpenAIEndpoint:   os.Getenv("AZURE_OPENAI_ENDPOINT"),
		AzureOpenAIDeployment: os.Getenv("AZURE_OPENAI_DEPLOYMENT"),
		AzureOpenAIAPIVersion: envOr("AZURE_OPENAI_API_VERSION", "2024-02-15-preview"),

		// Model selection
		OpenAINoteModel:   envOr("OPENAI_NOTE_MODEL", "gpt-4"),
		OpenAICanvasModel: envOr("OPENAI_CANVAS_MODEL", "gpt-4"),
		OpenAIPDFModel:    envOr("OPENAI_PDF_MODEL", "gpt-4"),
		OpenAIImageModel:  envOr("IMAGE_GEN_MODEL", "dall-e-3"),

		// Token budgets
		PDFPrecisTokens:       envInt64("OPENAI_PDF_PRECIS_TOKENS", 1000),
		CanvasPrecisTokens:    envInt64("OPENAI_CANVAS_PRECIS_TOKENS", 600),
		NoteResponseTokens:    envInt64("OPENAI_NOTE_RESPONSE_TOKENS", 400),
		ImageAnalysisTokens:   envInt64("OPENAI_IMAGE_ANALYSIS_TOKENS", 16384),
		ErrorResponseTokens:   envInt64("OPENAI_ERROR_RESPONSE_TOKENS", 200),
		PDFChunkSizeTokens:    envInt64("OPENAI_PDF_CHUNK_SIZE_TOKENS", 20000),
		PDFMaxChunksTokens:    envInt64("OPENAI_PDF_MAX_CHUNKS_TOKENS", 10),
		PDFSummaryRatioTokens: envFloat64("OPENAI_PDF_SUMMARY_RATIO", 0.3),

		// Operational limits
		MaxRetries:        envInt("MAX_RETRIES", 3),
		RetryDelay:        time.Duration(envInt("RETRY_DELAY", 1)) * time.Second,
		AITimeout:         time.Duration(envInt("AI_TIMEOUT", 60)) * time.Second,
		ProcessingTimeout: time.Duration(envInt("PROCESSING_TIMEOUT", 300)) * time.Second,
		MaxConcurrent:     envInt("MAX_CONCURRENT", 5),
		MaxFileSize:       envInt64("MAX_FILE_SIZE", 52428800),
		DownloadsDir:      envOr("DOWNLOADS_DIR", "./downloads"),

		// Subscribe drain
		SnapshotDrainSeconds: envInt("SNAPSHOT_DRAIN_SECONDS", 3),
	}, nil
}

// envOr returns the environment variable value or a default.
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// envInt parses an integer environment variable with a default.
func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// envInt64 parses an int64 environment variable with a default.
func envInt64(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}

// envFloat64 parses a float64 environment variable with a default.
func envFloat64(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}
