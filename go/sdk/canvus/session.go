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
	"net"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"
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

// transportWithAPIKey is an http.RoundTripper that adds an API key to requests.
type transportWithAPIKey struct {
	transport http.RoundTripper
	header    string
	apiKey    string
}

// RoundTrip implements http.RoundTripper.
func (t *transportWithAPIKey) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Add(t.header, t.apiKey)
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return t.transport.RoundTrip(req)
}

// WithAPIKey configures the session to use a static API key.
//
// Note: this option installs a round-tripper that adds the Private-Token
// header to every outgoing request. The default round-tripper used by this
// option also disables TLS verification — the Canvus dev/test servers are
// frequently self-signed. Pass WithHTTPClient first if you want a different
// verification policy.
func WithAPIKey(apiKey string) SessionConfigOption {
	return func(cfg *SessionConfig) {
		if cfg.HTTPClient == nil {
			cfg.HTTPClient = &http.Client{
				Transport: &http.Transport{
					//nolint:gosec // Canvus dev servers commonly use self-signed certs; document and accept.
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}
		}
		transport := cfg.HTTPClient.Transport
		if transport == nil {
			transport = http.DefaultTransport
		}
		cfg.HTTPClient.Transport = &transportWithAPIKey{
			transport: transport,
			header:    "Private-Token",
			apiKey:    apiKey,
		}
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

// tokenManager handles token storage and refresh.
type tokenManager struct {
	tokenStore   TokenStore
	currentToken string
	tokenExpiry  time.Time
	refreshMutex sync.Mutex
	config       *SessionConfig
}

func newTokenManager(config *SessionConfig) *tokenManager {
	tm := &tokenManager{config: config}
	if config.TokenStore != nil {
		tm.tokenStore = config.TokenStore
		token, _ := tm.tokenStore.GetToken()
		tm.currentToken = token
	}
	return tm
}

func (tm *tokenManager) getToken() string {
	tm.refreshMutex.Lock()
	defer tm.refreshMutex.Unlock()
	if !tm.tokenExpiry.IsZero() && time.Until(tm.tokenExpiry) < tm.config.TokenRefreshThreshold {
		_ = tm.refreshToken()
	}
	return tm.currentToken
}

func (tm *tokenManager) setToken(token string, expiresIn time.Duration) {
	tm.refreshMutex.Lock()
	defer tm.refreshMutex.Unlock()
	tm.currentToken = token
	if expiresIn > 0 {
		tm.tokenExpiry = time.Now().Add(expiresIn)
	}
	if tm.tokenStore != nil && token != "" {
		_ = tm.tokenStore.StoreToken(token, tm.tokenExpiry)
	}
}

func (tm *tokenManager) clearToken() {
	tm.refreshMutex.Lock()
	defer tm.refreshMutex.Unlock()
	tm.currentToken = ""
	tm.tokenExpiry = time.Time{}
	if tm.tokenStore != nil {
		_ = tm.tokenStore.ClearToken()
	}
}

func (tm *tokenManager) refreshToken() error {
	// Placeholder: token refresh is dependent on the auth flow; consumers
	// extending the SDK supply this. Returning nil keeps the lazy-get behavior.
	return nil
}

// Session is the main entry point for interacting with the Canvus API.
type Session struct {
	BaseURL        string
	HTTPClient     *http.Client
	config         *SessionConfig
	authenticator  Authenticator
	tokenManager   *tokenManager
	circuitBreaker *circuitBreaker
	userID         int64
	logger         *slog.Logger
}

// bootstraps stores per-config initial authenticators registered by options
// (e.g. WithToken) for installation when NewSession runs. Keyed by pointer so
// each config carries its own setup without polluting SessionConfig's public
// surface.
var (
	bootstrapMu sync.Mutex
	bootstraps  = map[*SessionConfig]Authenticator{}
)

// Helper used by WithToken (and any other options) to register an authenticator
// to be installed when NewSession is called with this config.
func registerBootstrapAuth(cfg *SessionConfig, a Authenticator) {
	bootstrapMu.Lock()
	defer bootstrapMu.Unlock()
	bootstraps[cfg] = a
}

func takeBootstrapAuth(cfg *SessionConfig) Authenticator {
	bootstrapMu.Lock()
	defer bootstrapMu.Unlock()
	a := bootstraps[cfg]
	delete(bootstraps, cfg)
	return a
}

// NewSession creates a new Canvus API session.
//
// Drift remediation #4 from go.md: NewSession logs an Info-level lifecycle
// event so embedders can confirm SDK boot ordering. The SDK uses slog.Default
// per the conventions document; embedders configure the slog handler.
func NewSession(cfg *SessionConfig, opts ...SessionConfigOption) *Session {
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
		client := &http.Client{Timeout: cfg.RequestTimeout}
		if cfg.ConnectTimeout > 0 {
			client.Transport = buildConnectTimeoutTransport(cfg.ConnectTimeout)
		}
		cfg.HTTPClient = client
	} else if cfg.HTTPClient.Timeout == 0 {
		cfg.HTTPClient.Timeout = cfg.RequestTimeout
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
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

	if boot := takeBootstrapAuth(cfg); boot != nil {
		s.authenticator = boot
	} else if s.tokenManager.tokenStore != nil {
		if token, err := s.tokenManager.tokenStore.GetToken(); err == nil && token != "" {
			s.authenticator = &TokenAuthenticator{Token: token}
		}
	}

	s.logger.Debug("session created", "base_url", cfg.BaseURL)
	return s
}

// SetLogger overrides the slog.Logger used internally by the SDK.
// Useful if the caller wants SDK logs to carry extra context (request_id, etc.).
func (s *Session) SetLogger(l *slog.Logger) {
	if l != nil {
		s.logger = l
	}
}

// doRequest issues an HTTP request with retry, circuit breaking, and token
// refresh. queryParams may be nil for none.
func (s *Session) doRequest(ctx context.Context, method, endpoint string, body any, out any, queryParams map[string]string, rawResponse bool, contentType ...string) error {
	var lastErr error
	var resp *http.Response
	var respBody []byte

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
	} else if body != nil {
		ct = "application/json"
	}

	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		reqBody, retryable, err := s.prepareRequestBody(body, ct)
		if err != nil {
			if !retryable || attempt == s.config.MaxRetries {
				return err
			}
			continue
		}

		req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		if ct != "" {
			req.Header.Set("Content-Type", ct)
		}
		req.Header.Set("User-Agent", s.config.UserAgent)
		if s.authenticator != nil {
			s.authenticator.Authenticate(req)
		}
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
			if !isRetryableError(err) || attempt == s.config.MaxRetries {
				s.circuitBreaker.failure()
				return lastErr
			}
			if shouldRetry(err, attempt, s.config) {
				time.Sleep(calculateBackoff(attempt, s.config))
				continue
			}
			return lastErr
		}

		respBody, err = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("read response: %w", err)
			s.circuitBreaker.failure()
			return lastErr
		}

		if resp.StatusCode >= 400 {
			lastErr = s.handleErrorResponse(resp, respBody, attempt)
			var apiErr *APIError
			if errors.As(lastErr, &apiErr) {
				if requestID != "" && apiErr.RequestID == "" {
					apiErr.RequestID = requestID
				}
				if apiErr.StatusCode == http.StatusUnauthorized && attempt == 0 {
					if refreshErr := s.refreshAuthToken(ctx); refreshErr == nil {
						continue
					}
				}
				if isRetryableError(apiErr) && attempt < s.config.MaxRetries {
					time.Sleep(calculateBackoff(attempt, s.config))
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
				return fmt.Errorf("failed to decode response: %w", err)
			}
			if err := validateResponse(out, body, method); err != nil {
				return fmt.Errorf("response validation failed: %w", err)
			}
		}
		return nil
	}

	s.circuitBreaker.failure()
	if lastErr != nil {
		return fmt.Errorf("request failed after %d attempts: %w", s.config.MaxRetries, lastErr)
	}
	return errors.New("request failed: unknown error")
}

func (s *Session) prepareRequestBody(body any, _ string) (io.Reader, bool, error) {
	if body == nil {
		return nil, true, nil
	}
	if rdr, ok := body.(io.Reader); ok {
		if seeker, ok := rdr.(io.ReadSeeker); ok {
			_, _ = seeker.Seek(0, io.SeekStart)
		}
		return rdr, true, nil
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, false, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return bytes.NewReader(b), true, nil
}

func (s *Session) handleErrorResponse(resp *http.Response, body []byte, _ int) error {
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

func (s *Session) refreshAuthToken(ctx context.Context) error {
	if tokenAuth, ok := s.authenticator.(*TokenAuthenticator); ok {
		newToken := s.tokenManager.getToken()
		if newToken != "" && newToken != tokenAuth.Token {
			tokenAuth.Token = newToken
			return nil
		}
		s.authenticator = nil
	}
	_ = ctx
	return errors.New("unable to refresh authentication token")
}

func isRetryableError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch {
		case apiErr.StatusCode >= 500:
			return true
		case apiErr.StatusCode == 429:
			return true
		case apiErr.StatusCode == 408:
			return true
		case apiErr.StatusCode == 0:
			return true
		}
		return false
	}
	return false
}

func shouldRetry(err error, attempt int, config *SessionConfig) bool {
	if attempt >= config.MaxRetries {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	return true
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

// validateResponse performs lightweight cross-checking that the server echoed
// fields that were specified in a PATCH/POST/PUT body. See the original
// SDK for the rationale.
func validateResponse(obj any, reqBody any, method string) error {
	if obj == nil {
		return errors.New("response is nil")
	}
	if method == http.MethodDelete {
		respMap := map[string]any{}
		b, err := json.Marshal(obj)
		if err != nil {
			return nil
		}
		if err := json.Unmarshal(b, &respMap); err != nil {
			return nil
		}
		var reqID any
		switch v := reqBody.(type) {
		case map[string]any:
			reqID = v["id"]
		case nil:
		default:
			rb, err := json.Marshal(reqBody)
			if err == nil {
				rm := map[string]any{}
				if err := json.Unmarshal(rb, &rm); err == nil {
					reqID = rm["id"]
				}
			}
		}
		if reqID != nil {
			if respID, ok := respMap["id"]; ok {
				if !reflect.DeepEqual(respID, reqID) {
					return fmt.Errorf("response id mismatch: got %v, want %v", respID, reqID)
				}
			}
		}
		if status, ok := respMap["status"]; ok {
			if v, ok := status.(string); ok && !strings.EqualFold(v, "deleted") {
				return fmt.Errorf("response status is not 'deleted': got %v", v)
			}
		} else if state, ok := respMap["state"]; ok {
			if v, ok := state.(string); ok && !strings.EqualFold(v, "deleted") {
				return fmt.Errorf("response state is not 'deleted': got %v", v)
			}
		}
		return nil
	}

	if method != http.MethodPatch && method != http.MethodPost && method != http.MethodPut {
		return nil
	}

	serverGeneratedFields := map[string]struct{}{
		"id": {}, "created_at": {}, "modified_at": {}, "last_login": {},
		"state": {}, "access": {}, "preview_hash": {}, "asset_size": {},
		"folder_id": {}, "parent_id": {}, "location": {}, "size": {},
	}
	writeOnlyFields := map[string]struct{}{"password": {}}

	var reqMap map[string]any
	switch v := reqBody.(type) {
	case map[string]any:
		reqMap = v
	case nil:
		return nil
	default:
		b, err := json.Marshal(reqBody)
		if err != nil {
			return nil
		}
		if err := json.Unmarshal(b, &reqMap); err != nil {
			return nil
		}
	}
	if len(reqMap) == 0 {
		return nil
	}

	respMap := map[string]any{}
	b, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	if err := json.Unmarshal(b, &respMap); err != nil {
		return nil
	}

	for k, reqVal := range reqMap {
		if _, skip := writeOnlyFields[k]; skip {
			continue
		}
		respVal, ok := respMap[k]
		if !ok {
			continue
		}
		if _, skip := serverGeneratedFields[k]; skip {
			continue
		}
		if k == "widget_type" {
			if s1, ok1 := reqVal.(string); ok1 {
				if s2, ok2 := respVal.(string); ok2 {
					if !strings.EqualFold(s1, s2) {
						return fmt.Errorf("response field %q mismatch (case-insensitive): got %v, want %v", k, respVal, reqVal)
					}
					continue
				}
			}
		}
		if isNumeric(reqVal) && isNumeric(respVal) {
			if !numericEqual(reqVal, respVal) {
				return fmt.Errorf("response field %q mismatch (numeric): got %v, want %v", k, respVal, reqVal)
			}
			continue
		}
		if !reflect.DeepEqual(respVal, reqVal) {
			return fmt.Errorf("response field %q mismatch: got %v, want %v", k, respVal, reqVal)
		}
	}
	return nil
}

// doRequestWithHeaders is like doRequest but supports custom headers and
// flexible query params (string/int values).
func (s *Session) doRequestWithHeaders(ctx context.Context, method, endpoint string, body any, out any, queryParams any, headers map[string]string, rawResponse bool) error {
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
	if s.authenticator != nil {
		s.authenticator.Authenticate(req)
	}
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

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	if out != nil {
		if rawResponse {
			if ptr, ok := out.(*[]byte); ok {
				*ptr = respBody
			} else {
				return errors.New("out must be *[]byte when rawResponse is true")
			}
		} else {
			if err := json.Unmarshal(respBody, out); err != nil {
				return err
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
	s.authenticator = &TokenAuthenticator{Token: loginResp.Token}
	s.userID = loginResp.User.ID
	s.logger.Info("login succeeded", "user_id", loginResp.User.ID)
	return nil
}

// Logout invalidates the current token and clears authentication.
func (s *Session) Logout(ctx context.Context) error {
	if err := s.doRequest(ctx, http.MethodPost, "users/logout", map[string]string{}, nil, nil, false); err != nil {
		return err
	}
	s.authenticator = nil
	s.tokenManager.clearToken()
	s.logger.Info("logout succeeded")
	return nil
}

// Users returns the Session itself; provided for chaining read-flows like
// `session.Users().ListUsers(ctx)` that mirror the older SDK shape.
func (s *Session) Users() *Session { return s }

// UserID returns the authenticated user's ID, or 0 if not logged in.
func (s *Session) UserID() int64 { return s.userID }

func isNumeric(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	}
	return false
}

func numericEqual(a, b any) bool {
	af, aok := toFloat64(a)
	bf, bok := toFloat64(b)
	if aok && bok {
		return af == bf
	}
	return false
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	}
	return 0, false
}
