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

// GetCurrentUser returns the currently-authenticated user via
// GET /users/{id} for s.UserID(). The Canvus server does not expose a
// /users/me alias (see VERIFIED-CORRECTIONS §5), so this helper enforces
// integer-ID lookup after a successful Login(). Returns an error if the
// session has not authenticated. Phase 4b §4.1 #2.
func (s *Session) GetCurrentUser(ctx context.Context) (*User, error) {
	userID := s.UserID()
	if userID == 0 {
		return nil, fmt.Errorf("GetCurrentUser: session is not logged in (call Login first)")
	}
	return s.GetUser(ctx, userID)
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

// SamlLogin performs a SAML login.
func (s *Session) SamlLogin(ctx context.Context, req SamlLoginRequest) error {
	return s.doRequest(ctx, http.MethodPost, "users/login/saml", req, nil, nil, false)
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

// SetUserPassword sets a user's password as an admin action.
//
// Field-name reconciliation per work item #9: the spec documents
// `{old-password, new-password}` for self-change; the C++ canonical client
// accepts a simple `{password}` body when invoked as admin. We send both
// `password` and `new-password` so either server interpretation works.
func (s *Session) SetUserPassword(ctx context.Context, userID int64, newPassword string) error {
	req := map[string]string{"password": newPassword, "new-password": newPassword}
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
