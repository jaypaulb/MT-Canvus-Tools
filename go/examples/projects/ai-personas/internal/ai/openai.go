package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/atom"
)

// OpenAI DALL-E retry configuration.
const (
	openAIMaxRetries     = 5
	openAIInitialBackoff = 1 * time.Second
	openAIMaxBackoff     = 32 * time.Second
	openAIHTTPTimeout    = 30 * time.Second
)

// OpenAI calls the OpenAI image generation endpoint (DALL-E) for persona
// portrait avatars.
type OpenAI struct {
	apiKey string
	http   *http.Client
}

// NewOpenAI constructs an OpenAI image client. apiKey is required.
func NewOpenAI(apiKey string) (*OpenAI, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("ai.NewOpenAI: OPENAI_API_KEY is required")
	}
	return &OpenAI{
		apiKey: apiKey,
		http:   &http.Client{Timeout: openAIHTTPTimeout},
	}, nil
}

// GeneratePersonaImage requests a 512x512 portrait for the persona and returns
// the downloaded image bytes. Retries on 429 and 5xx with exponential backoff
// (and respects the Retry-After header when present).
func (o *OpenAI) GeneratePersonaImage(ctx context.Context, persona atom.Persona) ([]byte, error) {
	prompt := fmt.Sprintf(
		"Business Appropriate Headshot of %s, a %s. %s, %s, %s. The headshot should be tightly cropped, centered on the face, with the full head visible and minimal chest.",
		persona.Name, persona.Role, string(persona.Age), persona.Sex, persona.Race,
	)

	body, _ := json.Marshal(map[string]any{
		"prompt": prompt,
		"n":      1,
		"size":   "512x512",
	})

	var lastErr *openaiErr
	for attempt := 1; attempt <= openAIMaxRetries; attempt++ {
		imgURL, retryAfter, err := o.requestImageURL(ctx, body)
		if err == nil {
			return o.downloadImage(ctx, imgURL)
		}
		lastErr = err
		if !err.retryable {
			return nil, err.err
		}
		if attempt == openAIMaxRetries {
			break
		}
		var sleep time.Duration
		if retryAfter > 0 {
			sleep = retryAfter
		} else {
			sleep = atom.CalculateBackoff(attempt, openAIInitialBackoff, openAIMaxBackoff, 0.1)
		}
		slog.Warn("openai dalle retry",
			"attempt", attempt, "max", openAIMaxRetries, "err", err.err, "sleep", sleep)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(sleep):
		}
	}
	if lastErr == nil {
		return nil, fmt.Errorf("openai dalle: max retries exceeded")
	}
	return nil, fmt.Errorf("openai dalle: max retries exceeded: %w", lastErr.err)
}

// openaiErr carries the retry classification with the wrapped error.
type openaiErr struct {
	err       error
	retryable bool
}

// Error implements error so we can pass *openaiErr around.
func (e *openaiErr) Error() string { return e.err.Error() }

// requestImageURL POSTs the DALL-E request and returns the image URL on
// success, along with any Retry-After hint on retryable failures.
func (o *OpenAI) requestImageURL(ctx context.Context, body []byte) (string, time.Duration, *openaiErr) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.openai.com/v1/images/generations", bytes.NewReader(body))
	if err != nil {
		return "", 0, &openaiErr{err: err, retryable: false}
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.http.Do(req)
	if err != nil {
		return "", 0, &openaiErr{err: fmt.Errorf("openai request: %w", err), retryable: true}
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		return "", atom.ParseRetryAfter(resp), &openaiErr{
			err:       fmt.Errorf("openai rate limit: %s", respBody),
			retryable: true,
		}
	case resp.StatusCode >= 500 && resp.StatusCode < 600:
		return "", 0, &openaiErr{
			err:       fmt.Errorf("openai %d: %s", resp.StatusCode, respBody),
			retryable: true,
		}
	case resp.StatusCode != http.StatusOK:
		return "", 0, &openaiErr{
			err:       fmt.Errorf("openai %d: %s", resp.StatusCode, respBody),
			retryable: bytes.Contains(respBody, []byte("server_error")),
		}
	}

	var parsed struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", 0, &openaiErr{err: fmt.Errorf("openai parse response: %w", err), retryable: false}
	}
	if len(parsed.Data) == 0 || parsed.Data[0].URL == "" {
		return "", 0, &openaiErr{err: fmt.Errorf("openai response missing URL"), retryable: false}
	}
	return parsed.Data[0].URL, 0, nil
}

// downloadImage GETs the URL returned by DALL-E and returns the raw bytes.
func (o *OpenAI) downloadImage(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("openai download new request: %w", err)
	}
	resp, err := o.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai download: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openai download read: %w", err)
	}
	return data, nil
}

// ValidateKey performs a lightweight GET /v1/models call to confirm the API
// key is accepted by OpenAI. Used at startup.
func (o *OpenAI) ValidateKey(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.openai.com/v1/models", nil)
	if err != nil {
		return fmt.Errorf("openai validate: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	resp, err := o.http.Do(req)
	if err != nil {
		return fmt.Errorf("openai validate: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("openai validate: status %d", resp.StatusCode)
	}
	return nil
}
