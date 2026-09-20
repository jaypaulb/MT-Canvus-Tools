package canvus

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// SessionConfig holds configuration for the API session.
type SessionConfig struct {
	// BaseURL is the base URL for all API requests.
	BaseURL string
	// HTTPClient is the HTTP client to use for requests.
	// If nil, a fresh client is built. Supplied clients are copied, not mutated.
	HTTPClient *http.Client
	// MaxRetries bounds automatic retries of bodyless GET/HEAD requests only.
	// Zero disables retries; negative values cause a pre-send configuration error.
	// DefaultSessionConfig sets 3. Mutations are never automatically retried.
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
	// ConnectTimeout caps the time spent dialing a TCP connection. When
	// non-zero and HTTPClient is nil (or its transport is the default), the
	// session installs a custom dialer with this connect deadline.
	// Phase 4b §4.1 #16.
	ConnectTimeout time.Duration
	// RequestIDFunc, when non-nil, is invoked for every outbound request and
	// its return value is used as the X-Request-ID header. The value is also
	// echoed into APIError.RequestID for failed requests. Phase 4b §4.1 #3.
	RequestIDFunc func() string
	// SkipTLSVerify disables TLS certificate verification when true.
	// Set via WithVerifyTLS(false). Ignored when HTTPClient is supplied
	// by the caller via WithHTTPClient — the caller's client takes precedence.
	// Phase 4d Round B.
	SkipTLSVerify bool
	// APIKey selects the initial Private-Token authority when no token is supplied.
	// Explicit Login replaces it; a 401 never falls back to it.
	APIKey string
	// SubscribeBuffer is the channel capacity used by subscribeStream for every
	// Subscribe* call on this session. Default: 4. Set via WithSubscribeBuffer.
	// Phase 4d Round B.
	SubscribeBuffer int

	bootstrapAuth Authenticator
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
		SubscribeBuffer: 4,
	}
}

// SessionConfigOption is a function that configures a SessionConfig.
type SessionConfigOption func(*SessionConfig)

// WithHTTPClient sets the HTTP client for the session.
func WithHTTPClient(client *http.Client) SessionConfigOption {
	return func(c *SessionConfig) { c.HTTPClient = client }
}

// WithMaxRetries bounds retries of safe reads. Zero disables retries.
// Negative budgets are rejected before sending; writes are never retried.
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

// WithConnectTimeout splits out the TCP connect timeout from RequestTimeout.
// When this option is set and the caller has not supplied a custom
// HTTPClient, NewSession installs a transport with the given dial timeout.
// Phase 4b §4.1 #16.
func WithConnectTimeout(d time.Duration) SessionConfigOption {
	return func(c *SessionConfig) { c.ConnectTimeout = d }
}

// WithRequestIDFunc registers a callback the SDK invokes for each outbound
// request; the returned string is sent as the X-Request-ID header and is
// surfaced on APIError.RequestID when the request fails. Phase 4b §4.1 #3.
func WithRequestIDFunc(fn func() string) SessionConfigOption {
	return func(c *SessionConfig) { c.RequestIDFunc = fn }
}

// WithVerifyTLS controls TLS certificate verification for the SDK's internal
// HTTP client. Pass verify=true (the default) to enforce verification;
// pass verify=false to skip it — suitable for Canvus servers that use
// self-signed certificates.
//
// Precedence: if the caller has already supplied a custom *http.Client via
// WithHTTPClient, this option has no effect — the caller's transport is used
// as-is. Apply WithVerifyTLS before WithHTTPClient, or configure TLS
// directly on your own transport. Phase 4d Round B.
func WithVerifyTLS(verify bool) SessionConfigOption {
	return func(c *SessionConfig) {
		c.SkipTLSVerify = !verify
	}
}

// FromEnv builds a Session from environment variables, applying any
// additional options on top. Recognised variables (all optional except
// CANVUS_API_URL):
//
//	CANVUS_API_URL     — Base URL (required).
//	CANVUS_API_KEY     — Private-Token value; installs WithAPIKey if set.
//	CANVUS_TIMEOUT_MS  — RequestTimeout, parsed as milliseconds.
//	CANVUS_VERIFY_TLS  — "0"/"false" disables TLS verification (default: verify).
//
// Phase 4b §4.1 #1: parity with python.from_env and ts.fromEnv.
func FromEnv(opts ...SessionConfigOption) (*Session, error) {
	baseURL := strings.TrimSpace(os.Getenv("CANVUS_API_URL"))
	if baseURL == "" {
		return nil, errors.New("FromEnv: CANVUS_API_URL is required")
	}
	cfg := DefaultSessionConfig()
	cfg.BaseURL = baseURL

	if v := strings.TrimSpace(os.Getenv("CANVUS_TIMEOUT_MS")); v != "" {
		ms, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("FromEnv: invalid CANVUS_TIMEOUT_MS %q: %w", v, err)
		}
		cfg.RequestTimeout = time.Duration(ms) * time.Millisecond
	}

	verify := true
	if v := strings.TrimSpace(os.Getenv("CANVUS_VERIFY_TLS")); v != "" {
		// Accept 1/0/true/false (case-insensitive).
		switch strings.ToLower(v) {
		case "0", "false", "no", "off":
			verify = false
		case "1", "true", "yes", "on":
			verify = true
		default:
			return nil, fmt.Errorf("FromEnv: invalid CANVUS_VERIFY_TLS %q (expected 1/0/true/false)", v)
		}
	}
	prepend := []SessionConfigOption{}
	if !verify {
		prepend = append(prepend, WithVerifyTLS(false))
	}
	if key := strings.TrimSpace(os.Getenv("CANVUS_API_KEY")); key != "" {
		prepend = append(prepend, WithAPIKey(key))
	}
	prepend = append(prepend, opts...)
	return NewSession(cfg, prepend...), nil
}

// dialTimeoutTransport wraps an http.Transport with a dial timeout.
// Returned by buildConnectTimeoutTransport when ConnectTimeout > 0 and no
// custom HTTPClient is supplied.
func buildConnectTimeoutTransport(connect time.Duration, skipTLS bool) http.RoundTripper {
	t := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   connect,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: connect,
	}
	if skipTLS {
		//nolint:gosec // explicit opt-out via WithVerifyTLS(false).
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return t
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
