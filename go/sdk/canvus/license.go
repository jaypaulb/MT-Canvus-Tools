package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// LicenseInfo represents the license information for the server.
type LicenseInfo struct {
	Edition    string `json:"edition,omitempty"`
	HasExpired bool   `json:"has_expired,omitempty"`
	IsValid    bool   `json:"is_valid,omitempty"`
	MaxClients int    `json:"max_clients,omitempty"`
	SeatModel  string `json:"seat_model,omitempty"`
	Type       string `json:"type,omitempty"`
}

// GetLicenseInfo retrieves the current license information.
func (s *Session) GetLicenseInfo(ctx context.Context) (*LicenseInfo, error) {
	var info LicenseInfo
	if err := s.doRequest(ctx, http.MethodGet, "license", nil, &info, nil, false); err != nil {
		return nil, fmt.Errorf("GetLicenseInfo: %w", err)
	}
	return &info, nil
}

// GetActivationRequest retrieves the offline activation request token.
func (s *Session) GetActivationRequest(ctx context.Context) (string, error) {
	var resp map[string]any
	if err := s.doRequest(ctx, http.MethodGet, "license/request", nil, &resp, nil, false); err != nil {
		return "", fmt.Errorf("GetActivationRequest: %w", err)
	}
	if token, ok := resp["request"].(string); ok {
		return token, nil
	}
	return "", fmt.Errorf("GetActivationRequest: response did not contain 'request' field")
}

// InstallLicense installs a new license key.
func (s *Session) InstallLicense(ctx context.Context, key string) error {
	req := map[string]string{"license": key}
	return s.doRequest(ctx, http.MethodPost, "license", req, nil, nil, false)
}

// ActivateLicense activates a license either online (empty key) or via an
// offline activation key.
//
// Per Phase 3 work item #9 the spec says POST /license/activate takes an
// empty body. The legacy SDK sent `{"key": activationKey}`. We send `key`
// only if non-empty.
func (s *Session) ActivateLicense(ctx context.Context, activationKey string) error {
	var body any
	if activationKey != "" {
		body = map[string]string{"key": activationKey}
	}
	return s.doRequest(ctx, http.MethodPost, "license/activate", body, nil, nil, false)
}
