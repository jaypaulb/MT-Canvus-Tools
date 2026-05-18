package internal

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTranslateTextPromptFormat verifies the prompt structure used by
// TranslateText. We do not test actual Gemini calls here (that requires
// a live key and network). The prompt format directly controls translation
// quality, so its shape is worth pinning.
func TestTranslateTextPromptFormat(t *testing.T) {
	text := "Hello world"
	lang := "Spanish"
	prompt := buildTranslatePrompt(text, lang)

	assert.True(t, strings.Contains(prompt, lang), "prompt should contain the target language")
	assert.True(t, strings.Contains(prompt, text), "prompt should contain the source text")
	assert.True(t, strings.Contains(prompt, "ONLY with the translated text"), "prompt should instruct Gemini to return only translated text")
}

// TestTranslateTextOutputFormat verifies the "Language\nTranslation" prefix
// format that formatTranslation applies. This format is a user-visible
// behaviour and must not change silently.
func TestTranslateTextOutputFormat(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		raw      string
		wantPfx  string
	}{
		{
			name:    "standard translation",
			lang:    "Spanish",
			raw:     "Hola mundo",
			wantPfx: "Spanish\nHola mundo",
		},
		{
			name:    "raw with surrounding whitespace stripped",
			lang:    "French",
			raw:     "  Bonjour monde  ",
			wantPfx: "French\nBonjour monde",
		},
		{
			name:    "multiline raw preserved after trim",
			lang:    "German",
			raw:     "Hallo\nWelt",
			wantPfx: "German\nHallo\nWelt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTranslation(tt.lang, tt.raw)
			assert.Equal(t, tt.wantPfx, got)
		})
	}
}

