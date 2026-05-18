package canvus

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithVerifyTLS_FalseInstallsInsecureTransport(t *testing.T) {
	cfg := &SessionConfig{BaseURL: "https://example.invalid/api/v1"}
	s := NewSession(cfg, WithVerifyTLS(false))
	require.NotNil(t, s)
	require.NotNil(t, s.HTTPClient)

	transport, ok := s.HTTPClient.Transport.(*http.Transport)
	require.True(t, ok, "expected *http.Transport, got %T", s.HTTPClient.Transport)
	require.NotNil(t, transport.TLSClientConfig, "TLSClientConfig should be set")
	assert.True(t, transport.TLSClientConfig.InsecureSkipVerify, "InsecureSkipVerify should be true")
}

func TestWithVerifyTLS_TrueKeepsSecureTransport(t *testing.T) {
	cfg := &SessionConfig{BaseURL: "https://example.invalid/api/v1"}
	s := NewSession(cfg, WithVerifyTLS(true))
	require.NotNil(t, s)
	require.NotNil(t, s.HTTPClient)

	// Default transport is nil (uses http.DefaultTransport) — no insecure config.
	if s.HTTPClient.Transport != nil {
		transport, ok := s.HTTPClient.Transport.(*http.Transport)
		if ok && transport.TLSClientConfig != nil {
			assert.False(t, transport.TLSClientConfig.InsecureSkipVerify,
				"InsecureSkipVerify should not be set when WithVerifyTLS(true)")
		}
	}
}

func TestWithVerifyTLS_DefaultIsSecure(t *testing.T) {
	// No WithVerifyTLS call — default behaviour must be verify=true.
	cfg := &SessionConfig{BaseURL: "https://example.invalid/api/v1"}
	s := NewSession(cfg)
	require.NotNil(t, s)
	require.NotNil(t, s.HTTPClient)

	// Transport should be nil (uses http.DefaultTransport) — no insecure override.
	if s.HTTPClient.Transport != nil {
		transport, ok := s.HTTPClient.Transport.(*http.Transport)
		if ok && transport.TLSClientConfig != nil {
			assert.False(t, transport.TLSClientConfig.InsecureSkipVerify,
				"default session must not skip TLS verification")
		}
	}
}

func TestWithVerifyTLS_IgnoredWhenHTTPClientSupplied(t *testing.T) {
	// When WithHTTPClient supplies a client, WithVerifyTLS(false) must not
	// overwrite it. The caller's transport takes precedence.
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false}, //nolint:gosec // test fixture.
	}
	customClient := &http.Client{Transport: customTransport}

	cfg := &SessionConfig{BaseURL: "https://example.invalid/api/v1"}
	// Apply WithHTTPClient first, then WithVerifyTLS(false); the custom client wins.
	s := NewSession(cfg, WithHTTPClient(customClient), WithVerifyTLS(false))
	require.NotNil(t, s)

	// The session must use the caller's client.
	assert.Same(t, customClient, s.HTTPClient,
		"WithVerifyTLS(false) must not replace a caller-supplied HTTPClient")
	// The transport must be the caller's transport (verify=false is ignored).
	assert.False(t, s.HTTPClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify,
		"caller's TLS config must not be mutated by WithVerifyTLS")
}
