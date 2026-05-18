// Package session provides SDK session management for the Canvus CLI.
// It creates and configures SDK sessions based on CLI configuration.
package session

import (
	"context"
	"fmt"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// contextKey is a type for context keys to avoid collisions
type contextKey int

const (
	sessionKey contextKey = iota
)

// NewSession creates a new Canvus SDK session from the provided configuration.
//
// Authentication is selected in this order:
//  1. APIKey (preferred)
//  2. Username + Password (uses Login to acquire a temporary token)
//
// TLS verification is on by default. When cfg.Insecure is true (typically from
// --insecure or CANVUS_INSECURE) the session is created with WithVerifyTLS(false).
// This is opt-in only.
func NewSession(cfg *config.Config) (*canvus.Session, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	sessionCfg := canvus.DefaultSessionConfig()
	sessionCfg.BaseURL = cfg.URL
	sessionCfg.RequestTimeout = time.Duration(cfg.Timeout) * time.Second

	opts := []canvus.SessionConfigOption{}
	if cfg.Insecure {
		opts = append(opts, canvus.WithVerifyTLS(false))
	}

	switch {
	case cfg.APIKey != "":
		return canvus.NewSession(sessionCfg, append(opts, canvus.WithAPIKey(cfg.APIKey))...), nil
	case cfg.Username != "" && cfg.Password != "":
		sess := canvus.NewSession(sessionCfg, opts...)
		ctx := context.Background()
		if err := sess.Login(ctx, cfg.Username, cfg.Password); err != nil {
			return nil, fmt.Errorf("login with username/password: %w", err)
		}
		return sess, nil
	default:
		return nil, fmt.Errorf("authentication is required: either API key or username/password must be provided")
	}
}

// WithSession returns a new context with the session attached.
func WithSession(ctx context.Context, session *canvus.Session) context.Context {
	return context.WithValue(ctx, sessionKey, session)
}

// GetSession retrieves the session from the context.
func GetSession(ctx context.Context) (*canvus.Session, error) {
	session, ok := ctx.Value(sessionKey).(*canvus.Session)
	if !ok || session == nil {
		return nil, fmt.Errorf("session not found in context")
	}
	return session, nil
}

// MustGetSession retrieves the session from context or panics.
// Use only in command handlers where session presence is guaranteed.
func MustGetSession(ctx context.Context) *canvus.Session {
	session, err := GetSession(ctx)
	if err != nil {
		panic(fmt.Sprintf("session not found in context: %v", err))
	}
	return session
}
