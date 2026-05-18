package atom

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateBackoff_NoJitter(t *testing.T) {
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{1, 1 * time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{4, 8 * time.Second},
		{5, 16 * time.Second},
		{6, 32 * time.Second}, // capped
		{10, 32 * time.Second},
	}
	for _, tt := range tests {
		got := CalculateBackoff(tt.attempt, time.Second, 32*time.Second, 0)
		assert.Equal(t, tt.want, got, "attempt=%d", tt.attempt)
	}
}

func TestCalculateBackoff_WithJitter(t *testing.T) {
	// Jitter should keep delay within +/- (delay * jitterFactor) of base.
	base := 4 * time.Second
	jitter := 0.5
	for i := 0; i < 50; i++ {
		got := CalculateBackoff(3, time.Second, 32*time.Second, jitter)
		// base = 4s, +/- 2s = [2s, 6s]
		assert.GreaterOrEqual(t, got, base-2*time.Second)
		assert.LessOrEqual(t, got, base+2*time.Second)
	}
}

func TestCalculateBackoff_NeverNegative(t *testing.T) {
	for i := 0; i < 100; i++ {
		got := CalculateBackoff(1, time.Second, 32*time.Second, 0.99)
		assert.GreaterOrEqual(t, got, time.Duration(0))
	}
}

func TestParseRetryAfter(t *testing.T) {
	t.Run("nil response", func(t *testing.T) {
		assert.Equal(t, time.Duration(0), ParseRetryAfter(nil))
	})
	t.Run("no header", func(t *testing.T) {
		resp := &http.Response{Header: http.Header{}}
		assert.Equal(t, time.Duration(0), ParseRetryAfter(resp))
	})
	t.Run("seconds form", func(t *testing.T) {
		resp := &http.Response{Header: http.Header{"Retry-After": []string{"30"}}}
		assert.Equal(t, 30*time.Second, ParseRetryAfter(resp))
	})
	t.Run("garbage", func(t *testing.T) {
		resp := &http.Response{Header: http.Header{"Retry-After": []string{"not-a-date"}}}
		assert.Equal(t, time.Duration(0), ParseRetryAfter(resp))
	})
}
