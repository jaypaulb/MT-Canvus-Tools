package canvus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// ClientInfo represents a client device registered with the server.
type ClientInfo struct {
	ID               string `json:"id"`
	InstallationName string `json:"installation_name"`
	Name             string `json:"name"`
	// UserID is client association metadata, not proof of the physical operator.
	// Numeric wire IDs are normalized to strings for source compatibility.
	UserID    string `json:"user_id"`
	UserIDRaw string `json:"-"` // Original scalar JSON, preserving wire representation.
	CreatedAt string `json:"created_at"`
}

// UnmarshalJSON accepts string/integer association metadata without inventing
// an operator identity. Workspace.User remains the separately observed email.
func (c *ClientInfo) UnmarshalJSON(data []byte) error {
	type plain ClientInfo
	var value plain
	aux := struct {
		*plain
		UserID json.RawMessage `json:"user_id"`
	}{plain: &value}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("decode ClientInfo: %w", err)
	}
	raw := bytes.TrimSpace(aux.UserID)
	value.UserIDRaw = string(raw)
	if len(raw) != 0 && !bytes.Equal(raw, []byte("null")) {
		if raw[0] == '"' {
			if err := json.Unmarshal(raw, &value.UserID); err != nil {
				return fmt.Errorf("decode client user: %w", err)
			}
		} else {
			id, err := strconv.ParseInt(string(raw), 10, 64)
			if err != nil {
				return fmt.Errorf("decode client user: %w", err)
			}
			value.UserID = strconv.FormatInt(id, 10)
		}
	}
	*c = ClientInfo(value)
	return nil
}

// ListClients retrieves all clients.
func (s *Session) ListClients(ctx context.Context) ([]ClientInfo, error) {
	var clients []ClientInfo
	if err := s.doRequest(ctx, http.MethodGet, "clients", nil, &clients, nil, false); err != nil {
		return nil, fmt.Errorf("ListClients: %w", err)
	}
	return clients, nil
}

// GetClient retrieves a client by ID.
func (s *Session) GetClient(ctx context.Context, id string) (*ClientInfo, error) {
	if id == "" {
		return nil, fmt.Errorf("GetClient: id is required")
	}
	var client ClientInfo
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("clients/%s", id), nil, &client, nil, false); err != nil {
		return nil, fmt.Errorf("GetClient: %w", err)
	}
	return &client, nil
}

// NewSessionFromConfig is a convenience constructor that builds a Session
// pre-configured for an API key from baseURL and apiKey strings. Useful in
// tools that load credentials from a settings file.
func NewSessionFromConfig(baseURL, apiKey string) *Session {
	cfg := DefaultSessionConfig()
	cfg.BaseURL = baseURL
	session := NewSession(cfg)
	if apiKey != "" {
		session.authenticator = &APIKeyAuthenticator{Header: "Private-Token", APIKey: apiKey}
	}
	return session
}

// NewDefaultSession creates a new session with default configuration.
func NewDefaultSession(baseURL string) *Session {
	cfg := DefaultSessionConfig()
	cfg.BaseURL = baseURL
	return NewSession(cfg)
}

// --- Test helpers retained from legacy SDK ---

// TestClient wraps a Session and manages a temporary test user and token.
type TestClient struct {
	Session     *Session
	userID      int64
	email       string
	password    string
	cleanupFunc func(context.Context) error
}

// UserClient wraps a Session and manages a temporary token for an existing user.
type UserClient struct {
	Session     *Session
	cleanupFunc func(context.Context) error
}

// NewTestClient creates a new test user, logs in as that user, and returns a TestClient.
func NewTestClient(ctx context.Context, adminSession *Session, baseURL, testEmail, testUsername, testPassword string) (*TestClient, error) {
	user, err := adminSession.CreateUser(ctx, CreateUserRequest{
		Name:     testUsername,
		Email:    testEmail,
		Password: testPassword,
	})
	if err != nil {
		return nil, err
	}
	cfg := DefaultSessionConfig()
	cfg.BaseURL = baseURL
	testSession := NewSession(cfg)
	if err := testSession.Login(ctx, testEmail, testPassword); err != nil {
		_ = adminSession.DeleteUser(ctx, user.ID)
		return nil, err
	}
	cleanup := func(ctx context.Context) error {
		_ = testSession.Logout(ctx)
		return adminSession.DeleteUser(ctx, user.ID)
	}
	return &TestClient{
		Session:     testSession,
		userID:      user.ID,
		email:       testEmail,
		password:    testPassword,
		cleanupFunc: cleanup,
	}, nil
}

// Cleanup logs out and deletes the test user.
func (tc *TestClient) Cleanup(ctx context.Context) error {
	if tc.cleanupFunc != nil {
		return tc.cleanupFunc(ctx)
	}
	return nil
}

// NewUserClient logs in as an existing user and returns a UserClient.
func NewUserClient(ctx context.Context, baseURL, email, password string) (*UserClient, error) {
	cfg := DefaultSessionConfig()
	cfg.BaseURL = baseURL
	session := NewSession(cfg)
	if err := session.Login(ctx, email, password); err != nil {
		return nil, err
	}
	cleanup := func(ctx context.Context) error { return session.Logout(ctx) }
	return &UserClient{Session: session, cleanupFunc: cleanup}, nil
}

// Cleanup logs out and invalidates the token.
func (uc *UserClient) Cleanup(ctx context.Context) error {
	if uc.cleanupFunc != nil {
		return uc.cleanupFunc(ctx)
	}
	return nil
}
