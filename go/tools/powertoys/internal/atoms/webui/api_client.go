// go/tools/powertoys/internal/atoms/webui/api_client.go
package webui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	canvus "github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// APIClient wraps the Canvus SDK session.
type APIClient struct {
	session *canvus.Session
	baseURL string
}

// NewAPIClient creates an API client backed by the Canvus SDK.
// insecureTLS skips TLS certificate verification; use only for dev servers
// with self-signed certs.
func NewAPIClient(baseURL, authToken string, insecureTLS bool) (*APIClient, error) {
	cfg := &canvus.SessionConfig{BaseURL: baseURL}
	opts := []canvus.SessionConfigOption{canvus.WithAPIKey(authToken)}
	if insecureTLS {
		opts = append(opts, canvus.WithVerifyTLS(false))
	}
	s := canvus.NewSession(cfg, opts...)
	return &APIClient{session: s, baseURL: baseURL}, nil
}

// Session returns the underlying SDK session for direct typed API calls.
func (c *APIClient) Session() *canvus.Session { return c.session }

// GetClients returns all client devices from the Canvus server.
func (c *APIClient) GetClients(ctx context.Context) ([]canvus.ClientInfo, error) {
	return c.session.ListClients(ctx)
}

// Get performs a GET request using the SDK's underlying HTTP client.
// Provided for backward compatibility with atoms that use raw endpoints.
func (c *APIClient) Get(endpoint string) ([]byte, error) {
	url := c.baseURL + endpoint
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Get: failed to create request: %w", err)
	}
	return c.doHTTP(req)
}

// Post performs a POST request using the SDK's underlying HTTP client.
// Provided for backward compatibility with atoms that use raw endpoints.
func (c *APIClient) Post(endpoint string, data interface{}) ([]byte, error) {
	return c.doJSON(http.MethodPost, endpoint, data)
}

// Put performs a PUT request using the SDK's underlying HTTP client.
// Provided for backward compatibility with atoms that use raw endpoints.
func (c *APIClient) Put(endpoint string, data interface{}) ([]byte, error) {
	return c.doJSON(http.MethodPut, endpoint, data)
}

// Patch performs a PATCH request using the SDK's underlying HTTP client.
// Provided for backward compatibility with atoms that use raw endpoints.
func (c *APIClient) Patch(endpoint string, data interface{}) ([]byte, error) {
	return c.doJSON(http.MethodPatch, endpoint, data)
}

// Delete performs a DELETE request using the SDK's underlying HTTP client.
// Provided for backward compatibility with atoms that use raw endpoints.
func (c *APIClient) Delete(endpoint string) error {
	url := c.baseURL + endpoint
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("Delete: failed to create request: %w", err)
	}
	_, err = c.doHTTP(req)
	return err
}

// doJSON marshals data, builds a request with method and endpoint, and delegates to doHTTP.
func (c *APIClient) doJSON(method, endpoint string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal data: %w", method, err)
	}
	url := c.baseURL + endpoint
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create request: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doHTTP(req)
}

// doHTTP executes the request using the SDK session's HTTP client.
func (c *APIClient) doHTTP(req *http.Request) ([]byte, error) {
	resp, err := c.session.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}
	return body, nil
}
