package canvus

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

type authorityKey struct{}
type requestAuthority struct{ auth Authenticator }

// withRequestAuthority freezes identity across retries and redirects, including
// explicitly unauthenticated requests. A nil identity must not select a fallback.
func withRequestAuthority(req *http.Request, auth Authenticator) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), authorityKey{}, requestAuthority{auth: auth}))
}

// authTransport preserves authentication for legacy direct HTTPClient callers,
// while restricting SDK credentials to the configured API origin.
type authTransport struct {
	base     http.RoundTripper
	origin   *url.URL
	selected func() Authenticator
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	for key := range req.Header {
		if strings.EqualFold(key, "Private-Token") {
			delete(req.Header, key)
		}
	}
	if sameOrigin(req.URL, t.origin) {
		auth := t.selected()
		if frozen, ok := req.Context().Value(authorityKey{}).(requestAuthority); ok {
			auth = frozen.auth
		}
		if auth != nil {
			auth.Authenticate(req)
		}
	}
	return t.base.RoundTrip(req)
}

// CloseIdleConnections preserves http.Client's optional transport cleanup hook.
func (t *authTransport) CloseIdleConnections() {
	if closer, ok := t.base.(interface{ CloseIdleConnections() }); ok {
		closer.CloseIdleConnections()
	}
}

func sameOrigin(a, b *url.URL) bool {
	return a != nil && b != nil && a.Scheme == b.Scheme && a.Host == b.Host
}

// safeRedirect permits only same-origin reads; mutations must never be replayed
// by the HTTP client's redirect machinery. Cross-origin reads need an explicit
// caller policy, and authTransport still strips the SDK credential there.
func safeRedirect(req *http.Request, via []*http.Request) error {
	if len(via) == 0 || len(via) >= 10 {
		return http.ErrUseLastResponse
	}
	read := func(method string) bool { return method == http.MethodGet || method == http.MethodHead }
	if !read(via[0].Method) || !read(req.Method) || !sameOrigin(req.URL, via[0].URL) {
		return http.ErrUseLastResponse
	}
	return nil
}
