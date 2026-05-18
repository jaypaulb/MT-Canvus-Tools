package commands

import (
	"crypto/tls"
	"net/http"
)

// insecureHTTPClient returns an *http.Client that skips TLS certificate
// verification. This is intentional for Canvus server deployments that use
// self-signed certificates and is gated on the InsecureTLS config flag.
func insecureHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			//nolint:gosec // Self-signed certs are common on Canvus server installs; this is opt-in via config.
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}
