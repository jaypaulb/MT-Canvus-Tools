package canvus

import "fmt"

// WithSubscribeBuffer sets the channel capacity used by every Subscribe* call
// on this session. High-throughput consumers (live dashboards, ai-personas) can
// raise this value to absorb bursts without blocking the streaming goroutine.
//
// The default is 4; values less than 1 are rejected at construction time.
// Phase 4d Round B.
func WithSubscribeBuffer(size int) SessionConfigOption {
	return func(c *SessionConfig) {
		if size < 1 {
			panic(fmt.Sprintf("WithSubscribeBuffer: size must be >= 1, got %d", size))
		}
		c.SubscribeBuffer = size
	}
}
