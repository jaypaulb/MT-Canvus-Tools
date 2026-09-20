package canvus

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"
)

// Authenticator applies authentication to an HTTP request.
type Authenticator interface {
	Authenticate(req *http.Request)
}

// APIKeyAuthenticator authenticates using a static API key and header.
type APIKeyAuthenticator struct {
	Header string
	APIKey string
}

// Authenticate sets the API key header on the request.
func (a *APIKeyAuthenticator) Authenticate(req *http.Request) {
	if a.Header != "" && a.APIKey != "" {
		req.Header.Set(a.Header, a.APIKey)
	}
}

// WithAPIKey configures the session to authenticate using a static API key.
//
// This option selects the bootstrap authority unless a token is supplied.
// Explicit Login replaces this authority; a 401 never restores it.
// TLS verification is controlled separately by WithVerifyTLS — pass
// WithVerifyTLS(false) when connecting to servers that use self-signed
// certificates (common for Canvus dev/test servers).
func WithAPIKey(apiKey string) SessionConfigOption {
	return func(cfg *SessionConfig) {
		cfg.APIKey = apiKey
	}
}

// TokenAuthenticator authenticates using a bearer token (Private-Token header).
type TokenAuthenticator struct {
	Token string
}

// Authenticate sets the Private-Token header on the request.
func (a *TokenAuthenticator) Authenticate(req *http.Request) {
	if a.Token != "" {
		req.Header.Set("Private-Token", a.Token)
	}
}

// WithToken configures the session to use a bearer token from the outset.
// Useful when a token has already been obtained via Login on another Session.
func WithToken(token string) SessionConfigOption {
	return func(cfg *SessionConfig) {
		registerBootstrapAuth(cfg, &TokenAuthenticator{Token: token})
	}
}

// circuitState represents the state of the circuit breaker.
type circuitState int

const (
	circuitStateClosed circuitState = iota
	circuitStateOpen
	circuitStateHalfOpen
)

// circuitBreaker implements a simple circuit breaker pattern.
type circuitBreaker struct {
	state        circuitState
	failures     int
	maxFailures  int
	resetTimeout time.Duration
	lastFailure  time.Time
	mutex        sync.RWMutex
}

func newCircuitBreaker(maxFailures int, resetTimeout time.Duration) *circuitBreaker {
	if maxFailures <= 0 {
		maxFailures = 5
	}
	if resetTimeout <= 0 {
		resetTimeout = 30 * time.Second
	}
	return &circuitBreaker{
		state:        circuitStateClosed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

func (cb *circuitBreaker) allow() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if cb.state == circuitStateClosed {
		return true
	}
	if cb.state == circuitStateOpen && time.Since(cb.lastFailure) > cb.resetTimeout {
		cb.state = circuitStateHalfOpen
		return true
	}
	return cb.state == circuitStateHalfOpen
}

func (cb *circuitBreaker) success() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	switch cb.state {
	case circuitStateHalfOpen:
		cb.state = circuitStateClosed
		cb.failures = 0
	case circuitStateClosed:
		cb.failures = 0
	}
}

func (cb *circuitBreaker) failure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	switch cb.state {
	case circuitStateClosed:
		cb.failures++
		if cb.failures >= cb.maxFailures {
			cb.state = circuitStateOpen
			cb.lastFailure = time.Now()
		}
	case circuitStateHalfOpen:
		cb.state = circuitStateOpen
		cb.lastFailure = time.Now()
	}
}

// tokenManager retains the caller's optional token store for explicit logout.
// Automatic refresh is not supported; it must never switch actor authority.
type tokenManager struct {
	tokenStore TokenStore
	mu         sync.Mutex
}

func newTokenManager(config *SessionConfig) *tokenManager {
	return &tokenManager{tokenStore: config.TokenStore}
}

func (tm *tokenManager) clearToken() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if tm.tokenStore != nil {
		_ = tm.tokenStore.ClearToken()
	}
}

// Session is the main entry point for interacting with the Canvus API.
type Session struct {
	// BaseURL is fixed at construction. Create a new session to change origin.
	BaseURL string
	// HTTPClient supports direct requests; configure its client/transport via
	// WithHTTPClient, not by replacing them after construction.
	HTTPClient     *http.Client
	config         *SessionConfig
	authenticator  Authenticator
	authMu         sync.RWMutex
	tokenManager   *tokenManager
	circuitBreaker *circuitBreaker
	userID         int64
	logger         *slog.Logger
	transport      *authTransport
	initErr        error
}

func registerBootstrapAuth(cfg *SessionConfig, a Authenticator) {
	cfg.bootstrapAuth = a
}

// requestAuthenticator snapshots the selected identity for a logical request.
// Installed authenticators are immutable; explicit login affects new requests.
func (s *Session) requestAuthenticator() Authenticator {
	s.authMu.RLock()
	defer s.authMu.RUnlock()
	return s.authenticator
}

// NewSession creates a new Canvus API session.
//
// Configuration and supplied HTTP client values are copied. Use
// DefaultSessionConfig for default read retries, or MaxRetries=0 for none.
func NewSession(cfg *SessionConfig, opts ...SessionConfigOption) *Session {
	if cfg == nil {
		cfg = DefaultSessionConfig()
	}
	copyConfig := *cfg
	cfg = &copyConfig
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 30 * time.Second
	}
	if cfg.HTTPClient == nil {
		// Build a fresh http.Client rather than mutating http.DefaultClient,
		// which would leak our timeout into unrelated code that uses the
		// default client. Phase 4b §4.1 #16: honour ConnectTimeout when set.
		// Phase 4d Round B: honour SkipTLSVerify when set via WithVerifyTLS(false).
		// Both may now be combined without one silently dropping the other.
		client := &http.Client{Timeout: cfg.RequestTimeout}
		if cfg.ConnectTimeout > 0 {
			client.Transport = buildConnectTimeoutTransport(cfg.ConnectTimeout, cfg.SkipTLSVerify)
		} else if cfg.SkipTLSVerify {
			//nolint:gosec // explicit opt-out via WithVerifyTLS(false).
			client.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}
		cfg.HTTPClient = client
	} else {
		copyClient := *cfg.HTTPClient
		cfg.HTTPClient = &copyClient
		if cfg.HTTPClient.Timeout == 0 {
			cfg.HTTPClient.Timeout = cfg.RequestTimeout
		}
	}
	if cfg.HTTPClient.CheckRedirect == nil {
		cfg.HTTPClient.CheckRedirect = safeRedirect
	}
	if cfg.RetryWaitMin == 0 {
		cfg.RetryWaitMin = 100 * time.Millisecond
	}
	if cfg.RetryWaitMax == 0 {
		cfg.RetryWaitMax = time.Second
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "mt-canvus-tools-go-sdk/v0.1.0"
	}

	s := &Session{
		BaseURL:        cfg.BaseURL,
		HTTPClient:     cfg.HTTPClient,
		config:         cfg,
		tokenManager:   newTokenManager(cfg),
		circuitBreaker: newCircuitBreaker(cfg.CircuitBreaker.MaxFailures, cfg.CircuitBreaker.ResetTimeout),
		logger:         slog.Default().With("component", "canvus-sdk"),
	}

	if boot := cfg.bootstrapAuth; boot != nil {
		s.authenticator = boot
	} else if s.tokenManager.tokenStore != nil {
		if token, err := s.tokenManager.tokenStore.GetToken(); err == nil && token != "" {
			s.authenticator = &TokenAuthenticator{Token: token}
		}
	}

	if s.authenticator == nil && cfg.APIKey != "" {
		s.authenticator = &APIKeyAuthenticator{Header: "Private-Token", APIKey: cfg.APIKey}
	}

	base := s.HTTPClient.Transport
	// Reusing an SDK client must not nest another session's auth injection.
	for {
		wrapped, ok := base.(*authTransport)
		if !ok {
			break
		}
		base = wrapped.base
	}
	if base == nil {
		base = http.DefaultTransport
	}
	origin, err := url.Parse(cfg.BaseURL)
	if err != nil || origin == nil || (origin.Scheme != "http" && origin.Scheme != "https") || origin.Hostname() == "" || origin.User != nil {
		s.initErr = fmt.Errorf("%w: BaseURL must be an absolute HTTP(S) URL without userinfo", ErrInvalidRequest)
	}
	if cfg.MaxRetries < 0 {
		s.initErr = ErrInvalidRetryBudget
	}
	s.transport = &authTransport{base: base, origin: origin, selected: s.requestAuthenticator, validate: s.validateRequestConfig}
	s.HTTPClient.Transport = s.transport

	s.logger.Debug("session created")
	return s
}

// SetLogger overrides the slog.Logger used internally by the SDK.
// Useful if the caller wants SDK logs to carry extra context (request_id, etc.).
func (s *Session) SetLogger(l *slog.Logger) {
	if l != nil {
		s.logger = l
	}
}

func (s *Session) validateRequestConfig() error {
	if s.initErr != nil {
		return s.initErr
	}
	if s.BaseURL != s.config.BaseURL || s.HTTPClient != s.config.HTTPClient || s.HTTPClient.Transport != s.transport {
		return fmt.Errorf("%w: session URL/client/transport changed; construct a new session with options", ErrInvalidRequest)
	}
	return nil
}

// doRequest issues an HTTP request with safe-read retries and circuit breaking.
// It never changes authority after rejection. queryParams may be nil for none.
func (s *Session) doRequest(ctx context.Context, method, endpoint string, body any, out any, queryParams map[string]string, rawResponse bool, contentType ...string) error {
	var lastErr error
	var resp *http.Response
	var respBody []byte

	if err := s.validateRequestConfig(); err != nil {
		return err
	}
	maxRetries := 0
	if (method == http.MethodGet || method == http.MethodHead) && body == nil {
		maxRetries = s.config.MaxRetries
	}

	if !s.circuitBreaker.allow() {
		return &APIError{
			StatusCode: http.StatusServiceUnavailable,
			Code:       "circuit_breaker_open",
			Message:    "service unavailable due to circuit breaker being open",
		}
	}

	u, err := url.Parse(s.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	u.Path = path.Join(u.Path, endpoint)

	if len(queryParams) > 0 {
		q := u.Query()
		for k, v := range queryParams {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	var ct string
	if len(contentType) > 0 {
		ct = contentType[0]
	} else if body != nil || (method != http.MethodGet && method != http.MethodHead) {
		ct = "application/json"
	}

	auth := s.requestAuthenticator()
	for attempt := 0; attempt <= maxRetries; attempt++ {
		reqBody, err := s.prepareRequestBody(body)
		if err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		if ct != "" {
			req.Header.Set("Content-Type", ct)
		}
		req.Header.Set("User-Agent", s.config.UserAgent)
		req = withRequestAuthority(req, auth)
		// Phase 4b §4.1 #3: inject X-Request-ID if configured.
		var requestID string
		if s.config.RequestIDFunc != nil {
			requestID = s.config.RequestIDFunc()
			if requestID != "" {
				req.Header.Set("X-Request-ID", requestID)
			}
		}

		resp, err = s.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			s.logger.Debug("request transport error", "method", method, "url", u.String(), "attempt", attempt, "err", err)
			if !IsRetryableError(err) || attempt == maxRetries {
				s.circuitBreaker.failure()
				return lastErr
			}
			if err := waitForRetry(ctx, calculateBackoff(attempt, s.config)); err != nil {
				return fmt.Errorf("retry wait: %w", errors.Join(err, lastErr))
			}
			continue
		}

		respBody, err = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = responseError(method, resp, respBody, err)
			s.circuitBreaker.failure()
			return lastErr
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = s.handleErrorResponse(resp, respBody, attempt)
			var apiErr *APIError
			if errors.As(lastErr, &apiErr) {
				if requestID != "" && apiErr.RequestID == "" {
					apiErr.RequestID = requestID
				}
				if IsRetryableError(apiErr) && attempt < maxRetries {
					if err := waitForRetry(ctx, calculateBackoff(attempt, s.config)); err != nil {
						return fmt.Errorf("retry wait: %w", errors.Join(err, lastErr))
					}
					continue
				}
			}
			s.circuitBreaker.failure()
			s.logger.Debug("request error response",
				"method", method, "url", u.String(),
				"status", resp.StatusCode, "attempt", attempt)
			return lastErr
		}

		s.circuitBreaker.success()
		if rawResponse {
			if ptr, ok := out.(*[]byte); ok {
				*ptr = respBody
				return nil
			}
			return errors.New("out must be *[]byte when rawResponse is true")
		}
		if out != nil && len(respBody) > 0 {
			if err := json.Unmarshal(respBody, out); err != nil {
				return responseError(method, resp, respBody, err)
			}
		}
		return nil
	}

	s.circuitBreaker.failure()
	if lastErr != nil {
		return fmt.Errorf("request failed after %d attempts: %w", maxRetries+1, lastErr)
	}
	return errors.New("request failed: unknown error")
}

func (s *Session) prepareRequestBody(body any) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}
	if rdr, ok := body.(io.Reader); ok {
		if seeker, ok := rdr.(io.ReadSeeker); ok {
			if _, err := seeker.Seek(0, io.SeekStart); err != nil {
				return nil, fmt.Errorf("failed to rewind request body: %w", err)
			}
		}
		return rdr, nil
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return bytes.NewReader(b), nil
}

func (s *Session) handleErrorResponse(resp *http.Response, body []byte, _ int) error {
	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return &APIError{StatusCode: resp.StatusCode, Code: CodeRedirectRefused, Message: "redirect not followed; configure the canonical API URL or an explicit safe redirect policy"}
	}
	var apiErr *APIError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr != nil && apiErr.Code != "" {
		apiErr.StatusCode = resp.StatusCode
		return apiErr
	}
	return &APIError{
		StatusCode: resp.StatusCode,
		Code:       fmt.Sprintf("http_%d", resp.StatusCode),
		Message:    string(body),
	}
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func calculateBackoff(attempt int, config *SessionConfig) time.Duration {
	minWait := float64(config.RetryWaitMin)
	maxWait := float64(config.RetryWaitMax)
	if minWait >= maxWait {
		return config.RetryWaitMax
	}
	backoff := minWait * math.Pow(2, float64(attempt))
	if backoff > maxWait {
		backoff = maxWait
	}
	randVal, _ := rand.Int(rand.Reader, big.NewInt(1000))
	jitter := (float64(randVal.Int64()) / 1000.0) * (backoff / 2)
	duration := time.Duration(backoff + jitter)
	if duration > config.RetryWaitMax {
		duration = config.RetryWaitMax
	}
	return duration
}

// doRequestWithHeaders is like doRequest but supports custom headers and
// flexible query params (string/int values).
func (s *Session) doRequestWithHeaders(ctx context.Context, method, endpoint string, body any, out any, queryParams any, headers map[string]string, rawResponse bool) error {
	if err := s.validateRequestConfig(); err != nil {
		return err
	}
	u, err := url.Parse(s.BaseURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, endpoint)

	qp := make(map[string]string)
	switch params := queryParams.(type) {
	case map[string]string:
		qp = params
	case map[string]any:
		for k, v := range params {
			qp[k] = toString(v)
		}
	case nil:
	default:
		return errors.New("queryParams must be map[string]string or map[string]interface{} or nil")
	}

	if len(qp) > 0 {
		q := u.Query()
		for k, v := range qp {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return err
	}
	req = withRequestAuthority(req, s.requestAuthenticator())
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return responseError(method, resp, respBody, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return s.handleErrorResponse(resp, respBody, 0)
	}

	if out != nil {
		if rawResponse {
			if ptr, ok := out.(*[]byte); ok {
				*ptr = respBody
			} else {
				return errors.New("out must be *[]byte when rawResponse is true")
			}
		} else if len(respBody) > 0 {
			if err := json.Unmarshal(respBody, out); err != nil {
				return responseError(method, resp, respBody, err)
			}
		}
	}
	return nil
}

func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		return fmt.Sprintf("%t", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Login authenticates a user and stores the returned token + user ID.
func (s *Session) Login(ctx context.Context, emailOrUser, password string) error {
	loginReq := map[string]string{
		"email":    emailOrUser,
		"password": password,
	}
	var loginResp struct {
		Token string `json:"token"`
		User  struct {
			ID int64 `json:"id"`
		} `json:"user"`
	}
	if err := s.doRequest(ctx, http.MethodPost, "users/login", loginReq, &loginResp, nil, false); err != nil {
		return err
	}
	if loginResp.Token == "" {
		return errors.New("login: no token returned")
	}
	s.authMu.Lock()
	s.authenticator = &TokenAuthenticator{Token: loginResp.Token}
	s.userID = loginResp.User.ID
	s.authMu.Unlock()
	s.logger.Debug("login succeeded", "user_id", loginResp.User.ID)
	return nil
}

// Logout invalidates the current token and clears authentication.
func (s *Session) Logout(ctx context.Context) error {
	if err := s.doRequest(ctx, http.MethodPost, "users/logout", map[string]string{}, nil, nil, false); err != nil {
		return err
	}
	s.authMu.Lock()
	s.authenticator = nil
	s.userID = 0
	s.authMu.Unlock()
	s.tokenManager.clearToken()
	s.logger.Debug("logout succeeded")
	return nil
}

// Users returns the Session itself; provided for chaining read-flows like
// `session.Users().ListUsers(ctx)` that mirror the older SDK shape.
func (s *Session) Users() *Session { return s }

// UserID returns the authenticated user's ID, or 0 if not logged in.
func (s *Session) UserID() int64 {
	s.authMu.RLock()
	defer s.authMu.RUnlock()
	return s.userID
}
