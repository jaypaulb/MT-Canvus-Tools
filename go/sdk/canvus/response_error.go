package canvus

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// AcceptedResponseError reports an HTTP 2xx whose response could not be read or
// decoded. The server acknowledged the request: do not repeat a mutation solely
// because decoding failed. Reconcile using ResourceID when available.
// This does not guarantee completion of asynchronous server/native-client work.
// Absence of this error does not prove a failed mutation was never dispatched.
type AcceptedResponseError struct {
	StatusCode int
	RequestID  string
	ResourceID string
	Err        error
}

// Error describes the accepted response without retaining raw response content.
func (e *AcceptedResponseError) Error() string {
	return fmt.Sprintf("HTTP %d accepted, response could not be read or decoded: %v", e.StatusCode, e.Err)
}

// Unwrap exposes the read/decode error to errors.Is and errors.As.
func (e *AcceptedResponseError) Unwrap() error { return e.Err }

func responseError(resp *http.Response, body []byte, err error) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("read response (HTTP %d): %w", resp.StatusCode, err)
	}
	var resource struct {
		ID string `json:"id"`
	}
	// IDs are retained only from a complete JSON object; never guess from broken JSON.
	_ = json.Unmarshal(body, &resource)
	return &AcceptedResponseError{StatusCode: resp.StatusCode, RequestID: resp.Header.Get("X-Request-ID"), ResourceID: resource.ID, Err: err}
}
