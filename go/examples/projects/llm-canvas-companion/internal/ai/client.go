// Package ai provides a thin wrapper around sashabaranov/go-openai to support
// OpenAI, Azure OpenAI, and local LLM endpoints (LMStudio / Ollama).
//
// Endpoint routing:
//   - Text completions use TextLLMURL when set, otherwise BaseLLMURL.
//   - Image generation uses ImageLLMURL when set. If the endpoint contains
//     "openai.azure.com" or "cognitiveservices.azure.com" the Azure path is
//     taken. Local endpoints (127.0.0.1 / localhost) reject image requests.
//
// All clients are constructed fresh per call; no global state.
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/llm-canvas-companion/internal/config"
	"github.com/sashabaranov/go-openai"
)

// AIResponse is the structured JSON the LLM must return for {{ }} triggers.
type AIResponse struct {
	Type    string `json:"type"`    // "text" or "image"
	Content string `json:"content"` // text body or image-generation prompt
}

// ChatJSON calls the chat completions endpoint and extracts a JSON AIResponse
// from the raw reply. The system message enforces the {"type","content"} format.
func ChatJSON(ctx context.Context, cfg *config.Config, model, systemMsg, userMsg string, maxTokens int) (*AIResponse, error) {
	client := textClient(cfg)

	enhancedSystem := "You are a JSON-only response generator. " +
		"CRITICAL: Respond with ONLY valid JSON. No other text allowed. " +
		"No explanations. No XML tags. No thinking out loud. " +
		systemMsg + "\n" +
		"RESPONSE FORMAT:\n" +
		"For text: {\"type\": \"text\", \"content\": \"your response\"}\n" +
		"For image: {\"type\": \"image\", \"content\": \"your prompt\"}"

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: enhancedSystem},
			{Role: openai.ChatMessageRoleUser, Content: userMsg},
		},
		MaxTokens:   maxTokens,
		Temperature: 0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("ChatJSON: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("ChatJSON: no choices returned")
	}

	raw := resp.Choices[0].Message.Content
	return extractAIResponse(raw)
}

// ChatMessages sends a pre-built message slice and returns the raw reply text.
// Used by the PDF multi-chunk protocol.
func ChatMessages(ctx context.Context, cfg *config.Config, model string, messages []openai.ChatCompletionMessage, maxTokens int) (string, error) {
	client := textClient(cfg)
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: 0.3,
	})
	if err != nil {
		return "", fmt.Errorf("ChatMessages: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("ChatMessages: no choices returned")
	}
	return resp.Choices[0].Message.Content, nil
}

// GenerateImageURL calls the image generation API and returns the resulting
// image URL. Selects OpenAI or Azure backend based on endpoint heuristics.
func GenerateImageURL(ctx context.Context, cfg *config.Config, prompt string) (string, error) {
	endpoint, isAzure := resolveImageEndpoint(cfg)

	if isLocal(endpoint) {
		return "", fmt.Errorf("GenerateImageURL: local endpoint %q does not support image generation; set IMAGE_LLM_URL", endpoint)
	}

	client := imageClientFor(cfg, endpoint)

	model := cfg.OpenAIImageModel
	if model == "" {
		model = "dall-e-3"
	}

	req := openai.ImageRequest{
		Prompt:         prompt,
		Model:          model,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
	}

	// dall-e-3 style param; skip for DALL-E 2 and most Azure deployments.
	if model == "dall-e-3" || (!isAzure && strings.Contains(strings.ToLower(model), "dall-e")) {
		req.Style = openai.CreateImageStyleVivid
	}
	if isAzure {
		if strings.Contains(strings.ToLower(model), "dalle3") || strings.Contains(strings.ToLower(model), "dall-e") {
			req.Style = openai.CreateImageStyleVivid
		}
	}

	img, err := client.CreateImage(ctx, req)
	if err != nil {
		return "", fmt.Errorf("GenerateImageURL: %w", err)
	}
	if len(img.Data) == 0 || img.Data[0].URL == "" {
		return "", fmt.Errorf("GenerateImageURL: API returned no image URL")
	}
	return img.Data[0].URL, nil
}

// --- internal helpers ---

// textClient builds an OpenAI client for text completions.
func textClient(cfg *config.Config) *openai.Client {
	c := openai.DefaultConfig(cfg.OpenAIAPIKey)
	if cfg.TextLLMURL != "" {
		c.BaseURL = cfg.TextLLMURL
	} else if cfg.BaseLLMURL != "" {
		c.BaseURL = cfg.BaseLLMURL
	}
	return openai.NewClientWithConfig(c)
}

// imageClientFor builds an OpenAI client for the resolved image endpoint.
func imageClientFor(cfg *config.Config, endpoint string) *openai.Client {
	c := openai.DefaultConfig(cfg.OpenAIAPIKey)
	c.BaseURL = endpoint
	return openai.NewClientWithConfig(c)
}

// resolveImageEndpoint picks the image endpoint and whether it is Azure.
func resolveImageEndpoint(cfg *config.Config) (endpoint string, isAzure bool) {
	switch {
	case cfg.ImageLLMURL != "":
		endpoint = cfg.ImageLLMURL
	case cfg.AzureOpenAIEndpoint != "":
		endpoint = cfg.AzureOpenAIEndpoint
		isAzure = true
	case cfg.BaseLLMURL != "":
		endpoint = cfg.BaseLLMURL
	default:
		endpoint = "https://api.openai.com/v1"
	}
	if !isAzure {
		isAzure = isAzureEndpoint(endpoint)
	}
	return endpoint, isAzure
}

// isAzureEndpoint returns true if the URL looks like an Azure OpenAI endpoint.
func isAzureEndpoint(u string) bool {
	lower := strings.ToLower(u)
	return strings.Contains(lower, "openai.azure.com") ||
		strings.Contains(lower, "cognitiveservices.azure.com")
}

// isLocal returns true for localhost / 127.0.0.1 endpoints.
func isLocal(u string) bool {
	lower := strings.ToLower(u)
	return strings.Contains(lower, "127.0.0.1") ||
		strings.Contains(lower, "localhost")
}


// extractAIResponse extracts the first {...} JSON object from raw and parses it.
func extractAIResponse(raw string) (*AIResponse, error) {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start == -1 || end == -1 || start > end {
		return nil, fmt.Errorf("extractAIResponse: no JSON object in response: %q", raw)
	}
	var r AIResponse
	if err := json.Unmarshal([]byte(raw[start:end+1]), &r); err != nil {
		return nil, fmt.Errorf("extractAIResponse: unmarshal: %w", err)
	}
	if r.Type == "" {
		return nil, fmt.Errorf("extractAIResponse: missing 'type' field")
	}
	if r.Content == "" {
		return nil, fmt.Errorf("extractAIResponse: missing 'content' field")
	}
	return &r, nil
}
