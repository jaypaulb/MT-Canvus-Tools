package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// User represents a user in the Canvus system.
type User struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Admin     bool   `json:"admin"`
	Approved  bool   `json:"approved"`
	Blocked   bool   `json:"blocked"`
	CreatedAt string `json:"created_at"`
	LastLogin string `json:"last_login"`
	State     string `json:"state"`
}

// CreateUserRequest is the payload for creating a user.
type CreateUserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password,omitempty"`
	Admin    *bool  `json:"admin,omitempty"`
	Approved *bool  `json:"approved,omitempty"`
	Blocked  *bool  `json:"blocked,omitempty"`
}

// UpdateUserRequest is the payload for updating a user.
type UpdateUserRequest struct {
	Email    *string `json:"email,omitempty"`
	Name     *string `json:"name,omitempty"`
	Password *string `json:"password,omitempty"`
	Admin    *bool   `json:"admin,omitempty"`
	Approved *bool   `json:"approved,omitempty"`
	Blocked  *bool   `json:"blocked,omitempty"`
}

// ListUsers retrieves all users.
func (s *Session) ListUsers(ctx context.Context) ([]User, error) {
	var users []User
	if err := s.doRequest(ctx, http.MethodGet, "users", nil, &users, nil, false); err != nil {
		return nil, fmt.Errorf("ListUsers: %w", err)
	}
	return users, nil
}

// GetCurrentUser uses GET users/{id} after login. For a bootstrap API token
// with no known ID it performs ONE explicit protocol token exchange (POST
// users/login), installing the resulting authenticated session. It does not
// guess users/current or infer identity from a workspace email. SAML tokens
// may reject re-exchange; that error is returned without authority fallback.
func (s *Session) GetCurrentUser(ctx context.Context) (*User, error) {
	if err := s.beginAuthChange(ctx); err != nil {
		return nil, err
	}
	defer s.endAuthChange()
	if frozen, ok := ctx.Value(authorityKey{}).(requestAuthority); ok && frozen.auth != s.requestAuthenticator() {
		return nil, fmt.Errorf("GetCurrentUser: %w: captured actor is no longer current", ErrIdentityUnavailable)
	}
	if id := s.UserID(); id != 0 {
		user, err := s.GetUser(ctx, id)
		if err != nil {
			return nil, err
		}
		if user.ID != id {
			return nil, fmt.Errorf("GetCurrentUser: %w: identity mismatch", ErrIdentityUnavailable)
		}
		return user, nil
	}
	var token string
	switch auth := s.requestAuthenticator().(type) {
	case *TokenAuthenticator:
		token = auth.Token
	case *APIKeyAuthenticator:
		token = auth.APIKey
	}
	if token == "" {
		return nil, fmt.Errorf("GetCurrentUser: %w", ErrIdentityUnavailable)
	}
	return s.loginAuthenticated(ctx, "users/login", map[string]any{"token": token, "remember": false})
}

// GetUser retrieves a user by ID.
func (s *Session) GetUser(ctx context.Context, id int64) (*User, error) {
	var user User
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("users/%d", id), nil, &user, nil, false); err != nil {
		return nil, fmt.Errorf("GetUser: %w", err)
	}
	return &user, nil
}

// CreateUser creates a new user. req can be CreateUserRequest or map[string]any.
func (s *Session) CreateUser(ctx context.Context, req any) (*User, error) {
	var user User
	if err := s.doRequest(ctx, http.MethodPost, "users", req, &user, nil, false); err != nil {
		return nil, fmt.Errorf("CreateUser: %w", err)
	}
	return &user, nil
}

// UpdateUser updates a user by ID.
func (s *Session) UpdateUser(ctx context.Context, id int64, req any) (*User, error) {
	var user User
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("users/%d", id), req, &user, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateUser: %w", err)
	}
	return &user, nil
}

// DeleteUser deletes a user by ID.
func (s *Session) DeleteUser(ctx context.Context, id int64) error {
	if err := s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("users/%d", id), nil, nil, nil, false); err != nil {
		return fmt.Errorf("DeleteUser: %w", err)
	}
	return nil
}

// SamlLoginRequest represents the payload for SAML login.
type SamlLoginRequest struct {
	InResponseTo string `json:"inResponseTo"`
	ResponseXML  string `json:"responseXml"`
	Remember     bool   `json:"remember"`
}

// SamlLogin validates a SAML assertion with the server and installs the
// authenticated user/token response. It does not validate assertions locally.
func (s *Session) SamlLogin(ctx context.Context, req SamlLoginRequest) error {
	if err := s.beginAuthChange(ctx); err != nil {
		return err
	}
	defer s.endAuthChange()
	_, err := s.loginAuthenticated(ctx, "users/login/saml", req)
	return err
}

// ValidateResetToken checks if a password reset token is valid.
func (s *Session) ValidateResetToken(ctx context.Context, token string) error {
	return s.doRequest(ctx, http.MethodGet, "users/password/validate-reset-token", nil, nil, map[string]string{"token": token}, false)
}

// RegisterUser registers a new user via the public registration flow.
func (s *Session) RegisterUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	var user User
	if err := s.doRequest(ctx, http.MethodPost, "users/register", req, &user, nil, false); err != nil {
		return nil, fmt.Errorf("RegisterUser: %w", err)
	}
	return &user, nil
}

// ConfirmEmail confirms a user's email address using a token.
func (s *Session) ConfirmEmail(ctx context.Context, token string) error {
	return s.doRequest(ctx, http.MethodPost, "users/confirm-email", map[string]string{"token": token}, nil, nil, false)
}

// CreateResetToken creates a password reset token for a user.
func (s *Session) CreateResetToken(ctx context.Context, email string) error {
	return s.doRequest(ctx, http.MethodPost, "users/password/create-reset-token", map[string]string{"email": email}, nil, nil, false)
}

// ResetUserPassword resets a user's password using a token.
func (s *Session) ResetUserPassword(ctx context.Context, token, newPassword string) error {
	req := map[string]string{"token": token, "password": newPassword}
	return s.doRequest(ctx, http.MethodPost, "users/password/reset", req, nil, nil, false)
}

// ChangeUserEmail changes a user's email address.
func (s *Session) ChangeUserEmail(ctx context.Context, userID int64, newEmail string) error {
	// Per Phase 3 work item #9, the spec uses `new-email` for the body field
	// while older Canvus servers accepted `email`. Send both for defensiveness.
	req := map[string]string{"email": newEmail, "new-email": newEmail}
	return s.doRequest(ctx, http.MethodPost, fmt.Sprintf("users/%d/change-email", userID), req, nil, nil, false)
}

// SetUserPassword requests an administrative password change using the observed
// new_password wire field. Server authorization still applies. For self-change,
// use ChangeUserPassword with the current password; no fallback fields are sent.
func (s *Session) SetUserPassword(ctx context.Context, userID int64, newPassword string) error {
	req := map[string]string{"new_password": newPassword}
	return s.doRequest(ctx, http.MethodPost, fmt.Sprintf("users/%d/password", userID), req, nil, nil, false)
}

// ChangeUserPassword changes a password using the current_password/new_password
// contract observed in the disposable trial and existing web-client API helper.
func (s *Session) ChangeUserPassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	req := map[string]string{"current_password": currentPassword, "new_password": newPassword}
	return s.doRequest(ctx, http.MethodPost, fmt.Sprintf("users/%d/password", userID), req, nil, nil, false)
}

// BlockUser blocks a user.
func (s *Session) BlockUser(ctx context.Context, userID int64) error {
	return s.doRequest(ctx, http.MethodPost, fmt.Sprintf("users/%d/block", userID), nil, nil, nil, false)
}

// UnblockUser unblocks a user.
func (s *Session) UnblockUser(ctx context.Context, userID int64) error {
	return s.doRequest(ctx, http.MethodPost, fmt.Sprintf("users/%d/unblock", userID), nil, nil, nil, false)
}

// ApproveUser approves a user.
func (s *Session) ApproveUser(ctx context.Context, userID int64) error {
	return s.doRequest(ctx, http.MethodPost, fmt.Sprintf("users/%d/approve", userID), nil, nil, nil, false)
}

// ForcePasswordResetUser forces a password reset for a user (admin action).
func (s *Session) ForcePasswordResetUser(ctx context.Context, userID int64) error {
	return s.doRequest(ctx, http.MethodPost, fmt.Sprintf("users/%d/reset-password", userID), nil, nil, nil, false)
}
