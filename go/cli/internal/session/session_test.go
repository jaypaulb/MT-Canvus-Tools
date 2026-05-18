package session

import (
	"context"
	"testing"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/config"
)

func TestNewSession(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
		skipReason string
	}{
		{
			name: "valid session with API key",
			config: &config.Config{
				URL:     "https://canvus.example.com",
				APIKey:  "test-api-key",
				Timeout: 30,
			},
			wantErr: false,
		},
		{
			name: "valid session with username/password",
			config: &config.Config{
				URL:      "https://canvus.example.com",
				Username: "testuser",
				Password: "testpass",
				Timeout:  30,
			},
			wantErr: false,
			skipReason: "requires actual server connection for Login()",
		},
		{
			name: "valid session with insecure mode",
			config: &config.Config{
				URL:      "https://canvus.example.com",
				APIKey:   "test-api-key",
				Timeout:  30,
				Insecure: true,
			},
			wantErr: false,
		},
		{
			name: "prefer API key over username/password",
			config: &config.Config{
				URL:      "https://canvus.example.com",
				APIKey:   "test-api-key",
				Username: "testuser",
				Password: "testpass",
				Timeout:  30,
			},
			wantErr: false,
		},
		{
			name: "missing authentication",
			config: &config.Config{
				URL:     "https://canvus.example.com",
				Timeout: 30,
			},
			wantErr: true,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			session, err := NewSession(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && session == nil {
				t.Error("NewSession() returned nil session without error")
			}

			// For successful cases, verify session has expected configuration
			if !tt.wantErr && session != nil {
				// Session should be created - we can't check internal fields
				// but we can verify it's not nil
				if session.BaseURL != tt.config.URL {
					t.Errorf("Session BaseURL = %s, want %s", session.BaseURL, tt.config.URL)
				}
			}
		})
	}
}

func TestWithSessionAndGetSession(t *testing.T) {
	cfg := &config.Config{
		URL:     "https://context.example.com",
		APIKey:  "context-api-key",
		Timeout: 30,
	}

	session, err := NewSession(cfg)
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	ctx := context.Background()
	ctx = WithSession(ctx, session)

	retrievedSession, err := GetSession(ctx)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}

	if retrievedSession != session {
		t.Error("GetSession() returned different session than stored")
	}
}

func TestGetSessionMissing(t *testing.T) {
	ctx := context.Background()
	_, err := GetSession(ctx)
	if err == nil {
		t.Error("expected error for missing session in context, got nil")
	}
}

func TestMustGetSession(t *testing.T) {
	cfg := &config.Config{
		URL:     "https://must.example.com",
		APIKey:  "must-api-key",
		Timeout: 30,
	}

	session, err := NewSession(cfg)
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	ctx := context.Background()
	ctx = WithSession(ctx, session)

	// Should not panic
	retrievedSession := MustGetSession(ctx)
	if retrievedSession != session {
		t.Error("MustGetSession() returned different session than stored")
	}
}

func TestMustGetSessionPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGetSession() did not panic with missing session")
		}
	}()

	ctx := context.Background()
	MustGetSession(ctx)
}

func TestInsecureMode(t *testing.T) {
	cfg := &config.Config{
		URL:      "https://insecure.example.com",
		APIKey:   "insecure-api-key",
		Timeout:  30,
		Insecure: true,
	}

	session, err := NewSession(cfg)
	if err != nil {
		t.Fatalf("NewSession() with insecure mode error = %v", err)
	}

	if session == nil {
		t.Error("NewSession() returned nil session for insecure mode")
	}

	// Session should be created with custom HTTP client that skips TLS verification
	// We can't directly test the TLS config without making actual requests,
	// but we can verify the session was created successfully
	if session.HTTPClient == nil {
		t.Error("Expected custom HTTP client for insecure mode, got nil")
	}
}

func TestAuthenticationPreference(t *testing.T) {
	// Test that API key is preferred when both API key and username/password are provided
	cfg := &config.Config{
		URL:      "https://preference.example.com",
		APIKey:   "preferred-api-key",
		Username: "testuser",
		Password: "testpass",
		Timeout:  30,
	}

	session, err := NewSession(cfg)
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	if session == nil {
		t.Error("NewSession() returned nil session")
	}

	// Session should be created using API key (no login call should have been made)
	// Since we're using API key, the session creation should succeed immediately
}
