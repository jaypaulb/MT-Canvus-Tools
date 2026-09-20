package canvus

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// LoginWithToken exchanges an API/one-time token for an authenticated session
// through POST users/login. It installs the returned token AND user ID. SAML
// tokens may not support re-exchange: use SamlLogin's authenticated response,
// not a guessed current-user alias or an alternate authority fallback.
func (s *Session) LoginWithToken(ctx context.Context, token string) error {
	if err := s.beginAuthChange(ctx); err != nil {
		return err
	}
	defer s.endAuthChange()
	if token == "" {
		return fmt.Errorf("LoginWithToken: %w", ErrIdentityUnavailable)
	}
	_, err := s.loginAuthenticated(ctx, "users/login", map[string]any{"token": token, "remember": false})
	return err
}

func (s *Session) beginAuthChange(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("authentication wait: %w", err)
	}
	if s.authChanges == nil {
		return fmt.Errorf("authentication wait: %w: construct with NewSession", ErrInvalidRequest)
	}
	select {
	case s.authChanges <- struct{}{}:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("authentication wait: %w", ctx.Err())
	}
}

func (s *Session) endAuthChange() { <-s.authChanges }

// loginAuthenticated requires an authentication-change lease. A valid response installs identity
// before persistence: a store failure cannot restore the old service authority.
func (s *Session) loginAuthenticated(ctx context.Context, endpoint string, payload any) (*User, error) {
	var result struct {
		Token string `json:"token"`
		User  *User  `json:"user"`
	}
	if err := s.doRequest(ctx, http.MethodPost, endpoint, payload, &result, nil, false); err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	if result.Token == "" || result.User == nil || result.User.ID <= 0 {
		return nil, fmt.Errorf("authenticate: %w", ErrInvalidLoginResponse)
	}
	s.authMu.Lock()
	s.authenticator = &TokenAuthenticator{Token: result.Token}
	s.userID = result.User.ID
	s.authMu.Unlock()
	if err := s.tokenManager.storeToken(result.Token); err != nil {
		return result.User, fmt.Errorf("authentication succeeded; persist token: %w", errors.Join(ErrTokenPersistence, err))
	}
	return result.User, nil
}
