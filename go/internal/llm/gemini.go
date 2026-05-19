// Package llm provides a small, shared Gemini client wrapper used by the
// translator, note-mapper, and ai-personas projects. It owns only the
// boilerplate of constructing a google.golang.org/genai client and exposing
// a Complete convenience for plain prompt->text use cases. Callers that need
// structured output, multi-part content, chat sessions, or retry logic should
// use Raw() to access the underlying *genai.Client and build on top.
package llm

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/genai"
)

// Config carries the construction parameters for a Gemini client.
//
// APIKey and Model are required; Temperature is optional and only applied
// by Complete (callers using Raw() are responsible for their own
// GenerateContentConfig).
type Config struct {
	APIKey      string
	Model       string
	Temperature float32
}

// Client is a thin wrapper over *genai.Client that pins a default model and
// temperature for the Complete convenience method. The underlying client is
// exposed via Raw() for callers needing the full genai surface.
type Client struct {
	raw         *genai.Client
	model       string
	temperature float32
}

// NewClient constructs a Gemini client using the unified
// google.golang.org/genai SDK. It validates that APIKey and Model are
// non-empty so callers fail loudly at construction rather than on first use.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("llm.NewClient: Config.APIKey must not be empty")
	}
	if cfg.Model == "" {
		return nil, errors.New("llm.NewClient: Config.Model must not be empty")
	}
	raw, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("llm.NewClient: %w", err)
	}
	return &Client{raw: raw, model: cfg.Model, temperature: cfg.Temperature}, nil
}

// Raw returns the underlying *genai.Client for callers that need the full
// SDK surface (chat sessions, structured output, multi-part content, etc.).
func (c *Client) Raw() *genai.Client { return c.raw }

// Model returns the configured default model id.
func (c *Client) Model() string { return c.model }

// Temperature returns the configured default sampling temperature.
func (c *Client) Temperature() float32 { return c.temperature }

// Complete sends prompt to the configured model and returns the response's
// concatenated text. It is the convenience path for plain prompt->text use
// cases. Callers needing structured output or multi-part content should use
// Raw() and call Models.GenerateContent directly.
//
// An empty prompt is an error; an empty response is also an error so callers
// don't silently get back "". Temperature is applied when non-zero.
func (c *Client) Complete(ctx context.Context, prompt string) (string, error) {
	if prompt == "" {
		return "", errors.New("llm.Complete: prompt must not be empty")
	}
	var cfg *genai.GenerateContentConfig
	if c.temperature != 0 {
		cfg = &genai.GenerateContentConfig{Temperature: genai.Ptr(c.temperature)}
	}
	resp, err := c.raw.Models.GenerateContent(ctx, c.model, genai.Text(prompt), cfg)
	if err != nil {
		return "", fmt.Errorf("llm.Complete: GenerateContent: %w", err)
	}
	text := resp.Text()
	if text == "" {
		return "", errors.New("llm.Complete: empty response from Gemini")
	}
	return text, nil
}

// Close releases any resources held by the underlying client. The current
// google.golang.org/genai client has no explicit Close, so this is a no-op
// today; callers should still call it to leave a sensible upgrade path.
func (c *Client) Close() error { return nil }
