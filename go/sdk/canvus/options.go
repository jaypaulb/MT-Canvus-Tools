package canvus

import (
	"net/http"
	"time"
)

// SessionConfig holds configuration for the API session.
type SessionConfig struct {
	// BaseURL is the base URL for all API requests.
	BaseURL string
	// HTTPClient is the HTTP client to use for requests.
	// If nil, http.DefaultClient is used.
	HTTPClient *http.Client
	// MaxRetries is the maximum number of retries for failed requests. Default: 3.
	MaxRetries int
	// RetryWaitMin is the minimum time to wait between retries. Default: 100ms.
	RetryWaitMin time.Duration
	// RetryWaitMax is the maximum time to wait between retries. Default: 1s.
	RetryWaitMax time.Duration
	// RequestTimeout is the timeout for each HTTP request. Default: 30s.
	RequestTimeout time.Duration
	// UserAgent is the User-Agent header to send with requests.
	UserAgent string
	// TokenRefreshThreshold controls when token refresh fires. Default: 5 minutes.
	TokenRefreshThreshold time.Duration
	// CircuitBreaker configures the circuit breaker behavior.
	CircuitBreaker CircuitBreakerConfig
	// TokenStore is used to store and retrieve authentication tokens.
	// If nil, tokens are not persisted between sessions.
	TokenStore TokenStore
}

// CircuitBreakerConfig holds configuration for the circuit breaker.
type CircuitBreakerConfig struct {
	// MaxFailures is the number of consecutive failures before opening the circuit. Default: 5.
	MaxFailures int
	// ResetTimeout is the time after which an open circuit will attempt to close. Default: 30s.
	ResetTimeout time.Duration
}

// TokenStore defines the interface for storing and retrieving authentication tokens.
type TokenStore interface {
	// GetToken returns the stored token or an error if not found.
	GetToken() (string, error)
	// StoreToken stores the token.
	StoreToken(token string, expiresAt time.Time) error
	// ClearToken removes the stored token.
	ClearToken() error
}

// DefaultSessionConfig returns a default session configuration.
func DefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		MaxRetries:            3,
		RetryWaitMin:          100 * time.Millisecond,
		RetryWaitMax:          time.Second,
		RequestTimeout:        30 * time.Second,
		UserAgent:             "mt-canvus-tools-go-sdk/v0.1.0",
		TokenRefreshThreshold: 5 * time.Minute,
		CircuitBreaker: CircuitBreakerConfig{
			MaxFailures:  5,
			ResetTimeout: 30 * time.Second,
		},
	}
}

// SessionConfigOption is a function that configures a SessionConfig.
type SessionConfigOption func(*SessionConfig)

// WithHTTPClient sets the HTTP client for the session.
func WithHTTPClient(client *http.Client) SessionConfigOption {
	return func(c *SessionConfig) { c.HTTPClient = client }
}

// WithMaxRetries sets the maximum number of retries for failed requests.
func WithMaxRetries(maxRetries int) SessionConfigOption {
	return func(c *SessionConfig) { c.MaxRetries = maxRetries }
}

// WithRetryWait sets the minimum and maximum wait time between retries.
func WithRetryWait(minWait, maxWait time.Duration) SessionConfigOption {
	return func(c *SessionConfig) {
		c.RetryWaitMin = minWait
		c.RetryWaitMax = maxWait
	}
}

// WithRequestTimeout sets the timeout for each HTTP request.
func WithRequestTimeout(timeout time.Duration) SessionConfigOption {
	return func(c *SessionConfig) { c.RequestTimeout = timeout }
}

// WithUserAgent sets the User-Agent header for requests.
func WithUserAgent(ua string) SessionConfigOption {
	return func(c *SessionConfig) { c.UserAgent = ua }
}

// WithTokenStore sets the token store for the session.
func WithTokenStore(store TokenStore) SessionConfigOption {
	return func(c *SessionConfig) { c.TokenStore = store }
}

// WithCircuitBreaker sets the circuit breaker configuration.
func WithCircuitBreaker(maxFailures int, resetTimeout time.Duration) SessionConfigOption {
	return func(c *SessionConfig) {
		c.CircuitBreaker = CircuitBreakerConfig{
			MaxFailures:  maxFailures,
			ResetTimeout: resetTimeout,
		}
	}
}

// WithTokenRefreshThreshold sets the token refresh threshold.
func WithTokenRefreshThreshold(threshold time.Duration) SessionConfigOption {
	return func(c *SessionConfig) { c.TokenRefreshThreshold = threshold }
}

// ListOptions specifies options for list endpoints (pagination, filtering, etc.).
type ListOptions struct {
	Limit  int    // Maximum number of items to return
	Offset int    // Offset for pagination
	Filter string // Optional filter string
}

// GetOptions specifies options for get endpoints (e.g., subscribe to updates).
type GetOptions struct {
	Subscribe bool // Whether to subscribe to updates (if supported)
}

// SubscribeOptions specifies options for streaming/subscription endpoints.
type SubscribeOptions struct {
	Annotations bool // Whether to include annotations
}

// AuditLogOptions specifies query options for the audit log endpoint.
//
// Per coverage-matrix work item #6, the SDK now exposes the full spec'd
// filter set instead of only PerPage.
type AuditLogOptions struct {
	Page      int    // Page number (1-based)
	PerPage   int    // Items per page
	Filter    string // Free-text filter
	StartTime string // ISO 8601 start time (spec key: start-time)
	EndTime   string // ISO 8601 end time   (spec key: end-time)
	UserID    string // Filter by user UUID (spec key: user-id)
	Action    string // Filter by action name
}
