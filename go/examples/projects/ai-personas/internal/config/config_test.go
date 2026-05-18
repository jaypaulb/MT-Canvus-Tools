package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setEnv sets a key to value for the duration of the test and clears it after.
func setEnv(t *testing.T, key, value string) {
	t.Helper()
	t.Setenv(key, value)
}

func setRequired(t *testing.T) {
	t.Helper()
	t.Setenv("CANVUS_API_URL", "https://example/api/v1/")
	t.Setenv("CANVUS_API_KEY", "k1")
	t.Setenv("CANVAS_ID", "c1")
	t.Setenv("GEMINI_API_KEY", "g1")
	t.Setenv("OPENAI_API_KEY", "o1")
}

func TestLoad_RequiresAllKeys(t *testing.T) {
	required := []string{"CANVUS_API_URL", "CANVUS_API_KEY", "CANVAS_ID", "GEMINI_API_KEY", "OPENAI_API_KEY"}
	for _, missing := range required {
		t.Run("missing "+missing, func(t *testing.T) {
			setRequired(t)
			t.Setenv(missing, "")
			_, err := Load()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), missing)
		})
	}
}

func TestLoad_Defaults(t *testing.T) {
	setRequired(t)
	t.Setenv("PORT", "")
	t.Setenv("WEB_PORT", "")
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "https://example/api/v1/", cfg.CanvusAPIURL)
	assert.Equal(t, "gemini-2.5-flash", cfg.PersonasModel)
	assert.Equal(t, "gemini-2.5-flash", cfg.ChatModel)
	assert.InDelta(t, 0.7, cfg.LLMTemp, 1e-6)
	assert.Equal(t, 256, cfg.ChatTokenLimit)
	assert.Equal(t, 5*time.Minute, cfg.QuestionTimeout)
	assert.Equal(t, 3, cfg.SnapshotDrainSeconds)
	assert.Equal(t, 30*time.Second, cfg.ShutdownTimeout)
	assert.Equal(t, "8080", cfg.WebPort)
}

func TestLoad_OverridesAndDurationFormats(t *testing.T) {
	setRequired(t)
	setEnv(t, "GEMINI_MODEL_PERSONAS", "gemini-2.5-pro")
	setEnv(t, "GEMINI_MODEL_CHAT", "gemini-2.5-flash-lite")
	setEnv(t, "LLM_TEMP", "0.2")
	setEnv(t, "CHAT_TOKEN_LIMIT", "512")
	setEnv(t, "QUESTION_TIMEOUT", "120") // bare seconds
	setEnv(t, "SNAPSHOT_DRAIN_SECONDS", "10")
	setEnv(t, "SHUTDOWN_TIMEOUT", "2m") // duration string
	setEnv(t, "WEB_PORT", "9999")
	setEnv(t, "PUBLIC_WEB_URL", "https://public.example/")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "gemini-2.5-pro", cfg.PersonasModel)
	assert.Equal(t, "gemini-2.5-flash-lite", cfg.ChatModel)
	assert.InDelta(t, 0.2, cfg.LLMTemp, 1e-6)
	assert.Equal(t, 512, cfg.ChatTokenLimit)
	assert.Equal(t, 120*time.Second, cfg.QuestionTimeout)
	assert.Equal(t, 10, cfg.SnapshotDrainSeconds)
	assert.Equal(t, 2*time.Minute, cfg.ShutdownTimeout)
	assert.Equal(t, "9999", cfg.WebPort)
	assert.Equal(t, "https://public.example/", cfg.PublicWebURL)
}

func TestLoad_PortPrefersPORTOverWEB_PORT(t *testing.T) {
	setRequired(t)
	setEnv(t, "PORT", "7777")
	setEnv(t, "WEB_PORT", "9999")
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "7777", cfg.WebPort)
}
