package llm

import (
	"context"
	"strings"
	"testing"
)

// TestNewClientValidation pins the Config validation surface so callers fail
// loudly at construction rather than on first use. We don't test the live
// network path here — that requires a real API key and is covered by the
// consumers' integration tests.
func TestNewClientValidation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		cfg     Config
		wantSub string
	}{
		{
			name:    "empty api key rejected",
			cfg:     Config{APIKey: "", Model: "gemini-2.0-flash"},
			wantSub: "APIKey must not be empty",
		},
		{
			name:    "empty model rejected",
			cfg:     Config{APIKey: "key", Model: ""},
			wantSub: "Model must not be empty",
		},
		{
			name:    "both empty rejected (api key checked first)",
			cfg:     Config{APIKey: "", Model: ""},
			wantSub: "APIKey must not be empty",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := NewClient(context.Background(), tc.cfg)
			if err == nil {
				t.Fatalf("expected error, got nil (client=%v)", c)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantSub)
			}
			if c != nil {
				t.Fatalf("expected nil client on error, got %v", c)
			}
		})
	}
}

// TestCompleteRejectsEmptyPrompt covers the cheap input-validation path of
// Complete without hitting the network. Construction is bypassed by building
// a Client directly with a nil raw — Complete must error before touching it.
func TestCompleteRejectsEmptyPrompt(t *testing.T) {
	t.Parallel()
	c := &Client{model: "gemini-2.0-flash"}
	_, err := c.Complete(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty prompt, got nil")
	}
	if !strings.Contains(err.Error(), "prompt must not be empty") {
		t.Fatalf("error %q does not contain %q", err.Error(), "prompt must not be empty")
	}
}

// TestCloseIsNoop ensures Close() can be called on a zero-value Client without
// panicking. The current genai SDK has no explicit close, but callers should
// be able to defer c.Close() safely.
func TestCloseIsNoop(t *testing.T) {
	t.Parallel()
	c := &Client{}
	if err := c.Close(); err != nil {
		t.Fatalf("Close() unexpected error: %v", err)
	}
}
