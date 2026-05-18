// Package atom contains small reusable helpers with no inter-package dependencies.
package atom

import (
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// CalculateBackoff returns the exponential backoff delay for a retry attempt with
// optional jitter. attempt is 1-indexed (first retry is attempt 1).
func CalculateBackoff(attempt int, initialDelay, maxDelay time.Duration, jitterFactor float64) time.Duration {
	multiplier := math.Pow(2, float64(attempt-1))
	delay := time.Duration(float64(initialDelay) * multiplier)
	if delay > maxDelay {
		delay = maxDelay
	}
	if jitterFactor > 0 {
		jitter := float64(delay) * jitterFactor
		randomJitter := (rand.Float64()*2 - 1) * jitter
		delay = time.Duration(float64(delay) + randomJitter)
	}
	if delay < 0 {
		delay = 0
	}
	return delay
}

// ParseRetryAfter parses the Retry-After header from an HTTP response.
// Returns the duration to wait, or 0 if the header is absent or malformed.
func ParseRetryAfter(resp *http.Response) time.Duration {
	if resp == nil {
		return 0
	}
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		return time.Duration(seconds) * time.Second
	}
	if t, err := time.Parse(time.RFC1123, retryAfter); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}
