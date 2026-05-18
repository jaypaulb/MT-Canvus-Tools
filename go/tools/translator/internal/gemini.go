package internal

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"google.golang.org/genai"
)

const (
	// defaultModel is the Gemini model used for translation.
	// gemini-2.0-flash is fast and cost-effective for this batch workload.
	defaultModel = "gemini-2.0-flash"
)

// GeminiClient wraps the google.golang.org/genai client for translation tasks.
// It uses the unified Gemini/Vertex SDK that replaced the deprecated
// github.com/google/generative-ai-go/genai package.
type GeminiClient struct {
	client *genai.Client
	model  string
}

// NewGeminiClient creates a GeminiClient authenticated with apiKey.
// The GEMINI_API_KEY environment variable is also accepted by the SDK
// directly, but we require the caller to pass the key explicitly so
// config ownership stays in LoadConfig.
func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("NewGeminiClient: %w", err)
	}
	return &GeminiClient{client: client, model: defaultModel}, nil
}

// TranslateText translates text into targetLanguage using Gemini.
// It returns the translated string prefixed with the target language name
// (e.g., "Spanish\nHola mundo") — this preserves the format of the original
// service so existing canvas layouts are not broken.
//
// Empty input returns an error; callers should filter empty notes before
// calling.
func (g *GeminiClient) TranslateText(ctx context.Context, text, targetLanguage string) (string, error) {
	if text == "" {
		return "", fmt.Errorf("TranslateText: text must not be empty")
	}
	if targetLanguage == "" {
		return "", fmt.Errorf("TranslateText: targetLanguage must not be empty")
	}

	prompt := buildTranslatePrompt(text, targetLanguage)

	slog.Debug("gemini translate", "target_language", targetLanguage, "text_len", len(text))

	result, err := g.client.Models.GenerateContent(ctx, g.model, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("TranslateText: GenerateContent: %w", err)
	}

	raw := result.Text()
	if raw == "" {
		return "", fmt.Errorf("TranslateText: empty response from Gemini")
	}

	translated := formatTranslation(targetLanguage, raw)
	slog.Debug("gemini translate complete", "target_language", targetLanguage, "result_len", len(translated))
	return translated, nil
}

// buildTranslatePrompt assembles the prompt sent to Gemini for a single
// translation. Extracted as a package-private helper so the prompt shape can
// be unit-tested without invoking the SDK.
func buildTranslatePrompt(text, targetLanguage string) string {
	return fmt.Sprintf(
		"Translate the following text from English to %s. "+
			"Provide a DIRECT, non-literal translation. "+
			"Respond ONLY with the translated text and nothing else. "+
			"Input text: %s",
		targetLanguage, text,
	)
}

// formatTranslation prefixes Gemini's raw response with the target language
// name separated by a newline — matches the original translator's output
// format so existing canvas note styling (which may rely on the first line
// being the language label) is preserved.
func formatTranslation(targetLanguage, raw string) string {
	return fmt.Sprintf("%s\n%s", targetLanguage, strings.TrimSpace(raw))
}
