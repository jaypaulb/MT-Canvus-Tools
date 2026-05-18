// Package ai wraps the Google Gemini and OpenAI clients used by ai-personas.
//
// The Gemini client uses google.golang.org/genai (the unified Gemini + Vertex
// SDK), not the deprecated github.com/google/generative-ai-go/genai package.
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"google.golang.org/genai"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/atom"
)

// Gemini retry configuration.
const (
	geminiMaxRetries     = 5
	geminiInitialBackoff = 1 * time.Second
	geminiMaxBackoff     = 32 * time.Second
)

// Gemini wraps a *genai.Client with persona-aware helpers.
type Gemini struct {
	client        *genai.Client
	personasModel string
	chatModel     string
	temperature   float32
}

// NewGemini constructs a Gemini client with the given API key and model defaults.
func NewGemini(ctx context.Context, apiKey, personasModel, chatModel string, temperature float32) (*Gemini, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("ai.NewGemini: GEMINI_API_KEY is required")
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("ai.NewGemini: %w", err)
	}
	return &Gemini{
		client:        client,
		personasModel: personasModel,
		chatModel:     chatModel,
		temperature:   temperature,
	}, nil
}

// Client returns the underlying genai client for callers that need direct access
// (e.g. building chat sessions).
func (g *Gemini) Client() *genai.Client { return g.client }

// ChatModel returns the configured chat model id.
func (g *Gemini) ChatModel() string { return g.chatModel }

// Temperature returns the configured sampling temperature.
func (g *Gemini) Temperature() float32 { return g.temperature }

// GeneratePersonas calls Gemini to produce exactly 4 personas from the given
// business-model context, parsing the JSON-array response.
func (g *Gemini) GeneratePersonas(ctx context.Context, businessContext string) ([]atom.Persona, error) {
	prompt := `Given the following business model context, generate exactly 4 diverse personas as a JSON array. These personas should represent POTENTIAL CLIENTS from 4 DIFFERENT MARKET SECTORS who would be interested in the products/services described. They should NOT be employees of the company, but rather external customers, buyers, or decision-makers from different industries or market segments.

Each persona should have the following fields: name, role, description, background, goals, age, sex, race. The "goals" field should be an array of strings representing their key objectives related to the business context.

Respond ONLY with the JSON array, no extra text.

Business Context:
` + businessContext

	cfg := &genai.GenerateContentConfig{Temperature: genai.Ptr(g.temperature)}
	model := g.personasModel

	resp, err := g.generateWithRetry(ctx, model, prompt, cfg, true)
	if err != nil {
		return nil, fmt.Errorf("GeneratePersonas: %w", err)
	}
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("GeneratePersonas: no response from Gemini")
	}
	jsonText := atom.StripMarkdownCodeBlock(resp.Candidates[0].Content.Parts[0].Text)

	var personas []atom.Persona
	if err := json.Unmarshal([]byte(jsonText), &personas); err != nil {
		return nil, fmt.Errorf("GeneratePersonas: parse JSON: %w\nraw: %s", err, jsonText)
	}
	return personas, nil
}

// generateWithRetry calls Models.GenerateContent with exponential backoff. If
// allowModelFallback is true and the first attempt fails with a NOT_FOUND, it
// switches to gemini-2.5-flash-lite for subsequent attempts.
func (g *Gemini) generateWithRetry(ctx context.Context, model, prompt string, cfg *genai.GenerateContentConfig, allowModelFallback bool) (*genai.GenerateContentResponse, error) {
	contents := []*genai.Content{{Parts: []*genai.Part{{Text: prompt}}}}
	var resp *genai.GenerateContentResponse
	var lastErr error

	for attempt := 1; attempt <= geminiMaxRetries; attempt++ {
		resp, lastErr = g.client.Models.GenerateContent(ctx, model, contents, cfg)
		if lastErr == nil {
			return resp, nil
		}
		if allowModelFallback && attempt == 1 && isModelNotFound(lastErr) {
			slog.Warn("gemini model not found; falling back",
				"model", model, "fallback", "gemini-2.5-flash-lite")
			model = "gemini-2.5-flash-lite"
			continue
		}
		if !isRetryable(lastErr) {
			return nil, lastErr
		}
		if attempt == geminiMaxRetries {
			break
		}
		backoff := atom.CalculateBackoff(attempt, geminiInitialBackoff, geminiMaxBackoff, 0.1)
		slog.Warn("gemini retry",
			"attempt", attempt, "max", geminiMaxRetries, "err", lastErr, "sleep", backoff)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}
	return nil, lastErr
}

// SessionManager maintains one Gemini chat session per persona name so a
// persona can answer multiple questions while remembering its own previous
// answers within a single Q&A workflow.
type SessionManager struct {
	mu       sync.Mutex
	client   *genai.Client
	model    string
	cfg      *genai.GenerateContentConfig
	sessions map[string]*genai.Chat
}

// NewSessionManager creates an empty session manager keyed on persona name.
func (g *Gemini) NewSessionManager() *SessionManager {
	return &SessionManager{
		client:   g.client,
		model:    g.chatModel,
		cfg:      &genai.GenerateContentConfig{Temperature: genai.Ptr(g.temperature)},
		sessions: make(map[string]*genai.Chat),
	}
}

// AnswerAs sends question through the persona's chat session, lazily creating
// the session and injecting the persona system prompt on first use.
func (sm *SessionManager) AnswerAs(ctx context.Context, persona atom.Persona, businessContext, question string) (string, error) {
	chat, err := sm.getOrCreateChat(ctx, persona, businessContext)
	if err != nil {
		return "", err
	}
	resp, err := sendWithRetry(ctx, chat, question, persona.Name)
	if err != nil {
		return "", fmt.Errorf("AnswerAs %s: %w", persona.Name, err)
	}
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("AnswerAs %s: no response", persona.Name)
	}
	return resp.Candidates[0].Content.Parts[0].Text, nil
}

func (sm *SessionManager) getOrCreateChat(ctx context.Context, persona atom.Persona, businessContext string) (*genai.Chat, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if chat, ok := sm.sessions[persona.Name]; ok {
		return chat, nil
	}
	chat, err := createChatWithRetry(ctx, sm.client, sm.model, sm.cfg)
	if err != nil {
		return nil, fmt.Errorf("create chat for %s: %w", persona.Name, err)
	}
	if _, err := chat.Send(ctx, &genai.Part{Text: atom.GenerateSystemPrompt(persona, businessContext)}); err != nil {
		return nil, fmt.Errorf("inject system prompt for %s: %w", persona.Name, err)
	}
	sm.sessions[persona.Name] = chat
	return chat, nil
}

func createChatWithRetry(ctx context.Context, client *genai.Client, model string, cfg *genai.GenerateContentConfig) (*genai.Chat, error) {
	var lastErr error
	for attempt := 1; attempt <= geminiMaxRetries; attempt++ {
		chat, err := client.Chats.Create(ctx, model, cfg, nil)
		if err == nil {
			return chat, nil
		}
		lastErr = err
		if attempt == 1 && isModelNotFound(err) {
			slog.Warn("chat model not found; falling back",
				"model", model, "fallback", "gemini-2.5-flash-lite")
			model = "gemini-2.5-flash-lite"
			continue
		}
		if !isRetryable(err) {
			return nil, err
		}
		if attempt == geminiMaxRetries {
			break
		}
		backoff := atom.CalculateBackoff(attempt, geminiInitialBackoff, geminiMaxBackoff, 0.1)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}
	return nil, lastErr
}

func sendWithRetry(ctx context.Context, chat *genai.Chat, question, personaName string) (*genai.GenerateContentResponse, error) {
	var lastErr error
	for attempt := 1; attempt <= geminiMaxRetries; attempt++ {
		resp, err := chat.Send(ctx, &genai.Part{Text: question})
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !isRetryable(err) {
			return nil, err
		}
		if attempt == geminiMaxRetries {
			break
		}
		backoff := atom.CalculateBackoff(attempt, geminiInitialBackoff, geminiMaxBackoff, 0.1)
		slog.Warn("gemini chat retry",
			"persona", personaName, "attempt", attempt, "err", err, "sleep", backoff)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}
	return nil, lastErr
}

// isModelNotFound matches Gemini error strings indicating an unknown model id.
func isModelNotFound(err error) bool {
	if err == nil {
		return false
	}
	es := err.Error()
	return strings.Contains(es, "not found") || strings.Contains(es, "NOT_FOUND")
}

// isRetryable matches Gemini error strings indicating a transient failure
// (rate limit, 5xx, internal). Returning true causes a backoff and retry.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	es := err.Error()
	switch {
	case strings.Contains(es, "429"),
		strings.Contains(es, "RESOURCE_EXHAUSTED"),
		strings.Contains(es, "rate limit"),
		strings.Contains(es, "quota exceeded"),
		strings.Contains(es, "Too Many Requests"),
		strings.Contains(es, "500"),
		strings.Contains(es, "502"),
		strings.Contains(es, "503"),
		strings.Contains(es, "504"),
		strings.Contains(es, "INTERNAL"),
		strings.Contains(es, "UNAVAILABLE"):
		return true
	}
	return false
}
