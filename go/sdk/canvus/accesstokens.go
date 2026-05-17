package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// AccessToken represents an API access token for a user.
type AccessToken struct {
	ID          string   `json:"id"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description"`
	CreatedAt   string   `json:"created_at"`
	Expires     string   `json:"expires,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
	PlainToken  string   `json:"plain_token,omitempty"`
}

// CreateAccessTokenRequest is the payload for creating an access token.
// The spec uses {name, expires, scopes}; the legacy SDK used {description}.
// We expose both Name and Description so callers can pick the convention
// matching their server.
type CreateAccessTokenRequest struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Expires     string   `json:"expires,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
}

// ListAccessTokens retrieves all access tokens for a user.
func (s *Session) ListAccessTokens(ctx context.Context, userID int64) ([]AccessToken, error) {
	var tokens []AccessToken
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("users/%d/access-tokens", userID), nil, &tokens, nil, false); err != nil {
		return nil, fmt.Errorf("ListAccessTokens: %w", err)
	}
	return tokens, nil
}

// GetAccessToken retrieves an access token by ID.
func (s *Session) GetAccessToken(ctx context.Context, userID int64, tokenID string) (*AccessToken, error) {
	if tokenID == "" {
		return nil, fmt.Errorf("GetAccessToken: tokenID is required")
	}
	var token AccessToken
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("users/%d/access-tokens/%s", userID, tokenID), nil, &token, nil, false); err != nil {
		return nil, fmt.Errorf("GetAccessToken: %w", err)
	}
	return &token, nil
}

// CreateAccessToken creates a new access token. req can be a
// CreateAccessTokenRequest or map[string]any.
func (s *Session) CreateAccessToken(ctx context.Context, userID int64, req any) (*AccessToken, error) {
	var token AccessToken
	if err := s.doRequest(ctx, http.MethodPost, fmt.Sprintf("users/%d/access-tokens", userID), req, &token, nil, false); err != nil {
		return nil, fmt.Errorf("CreateAccessToken: %w", err)
	}
	return &token, nil
}

// UpdateAccessToken updates an access token.
func (s *Session) UpdateAccessToken(ctx context.Context, userID int64, tokenID string, req any) (*AccessToken, error) {
	if tokenID == "" {
		return nil, fmt.Errorf("UpdateAccessToken: tokenID is required")
	}
	var token AccessToken
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("users/%d/access-tokens/%s", userID, tokenID), req, &token, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateAccessToken: %w", err)
	}
	return &token, nil
}

// DeleteAccessToken deletes an access token by ID.
func (s *Session) DeleteAccessToken(ctx context.Context, userID int64, tokenID string) error {
	if tokenID == "" {
		return fmt.Errorf("DeleteAccessToken: tokenID is required")
	}
	if err := s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("users/%d/access-tokens/%s", userID, tokenID), nil, nil, nil, false); err != nil {
		return fmt.Errorf("DeleteAccessToken: %w", err)
	}
	return nil
}
