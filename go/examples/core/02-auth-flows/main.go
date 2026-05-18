// Command 02-auth-flows demonstrates the three authentication paths the
// Canvus SDK supports:
//
//  1. Static API key (Private-Token header) — long-lived, suitable for daemons.
//  2. Email + password login — exchanges credentials for a short-lived bearer
//     token, suitable for interactive tools.
//  3. Access-token CRUD — creates a programmatic token on behalf of the
//     logged-in user, lists it back, and deletes it.
//
// Each path runs to completion regardless of whether the previous path
// succeeded, so partial credential setups still produce useful output.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func main() {
	setupLogging()
	if err := run(); err != nil {
		slog.Error("auth-flows failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	baseURL, err := mustEnv("CANVUS_BASE_URL")
	if err != nil {
		return err
	}
	ctx := context.Background()

	slog.Info("=== auth path 1: API key ===")
	if err := runAPIKey(ctx, baseURL); err != nil {
		slog.Warn("api-key path failed", "err", err)
	}

	slog.Info("=== auth path 2: email + password login ===")
	loginSession, err := runLogin(ctx, baseURL)
	if err != nil {
		slog.Warn("login path failed — skipping access-token lifecycle", "err", err)
		return nil
	}

	slog.Info("=== auth path 3: access-token CRUD ===")
	if err := runTokenLifecycle(ctx, loginSession); err != nil {
		slog.Warn("token lifecycle failed", "err", err)
	}
	return nil
}

// runAPIKey demonstrates the static API-key flow. The Private-Token header is
// injected by an http.RoundTripper installed by WithAPIKey.
func runAPIKey(ctx context.Context, baseURL string) error {
	apiKey := os.Getenv("CANVUS_API_KEY")
	if apiKey == "" {
		return errors.New("CANVUS_API_KEY not set; skipping API-key path")
	}
	cfg := &canvus.SessionConfig{BaseURL: baseURL}
	s := canvus.NewSession(cfg, canvus.WithAPIKey(apiKey))

	canvases, err := s.ListCanvases(ctx, nil)
	if err != nil {
		return fmt.Errorf("ListCanvases via API key: %w", err)
	}
	slog.Info("api-key auth succeeded", "canvas_count", len(canvases))
	return nil
}

// runLogin demonstrates the email + password flow. Per
// VERIFIED-CORRECTIONS.md §4, the server rejects login bodies that include
// any field other than {email, password, remember} — the SDK sends only
// {email, password}.
//
// Returns the authenticated session so callers can demonstrate dependent
// flows (e.g. access-token CRUD).
func runLogin(ctx context.Context, baseURL string) (*canvus.Session, error) {
	email := os.Getenv("CANVUS_EMAIL")
	password := os.Getenv("CANVUS_PASSWORD")
	if email == "" || password == "" {
		return nil, errors.New("CANVUS_EMAIL and/or CANVUS_PASSWORD not set; skipping login path")
	}

	cfg := &canvus.SessionConfig{BaseURL: baseURL}
	// No WithAPIKey: login itself is anonymous.
	s := canvus.NewSession(cfg)
	if err := s.Login(ctx, email, password); err != nil {
		return nil, fmt.Errorf("Login: %w", err)
	}
	userID := s.UserID()
	slog.Info("login succeeded", "user_id", userID)

	// Trivial authenticated call to prove the token works.
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return s, fmt.Errorf("GetUser(self) post-login: %w", err)
	}
	slog.Info("fetched self", "user_id", user.ID, "email", user.Email, "name", user.Name)
	return s, nil
}

// runTokenLifecycle creates a programmatic access token tied to the
// logged-in user, lists tokens back to confirm it appears, then deletes it
// so we leave the campsite clean.
func runTokenLifecycle(ctx context.Context, s *canvus.Session) error {
	userID := s.UserID()
	if userID == 0 {
		return errors.New("session has no user ID; was Login called?")
	}

	// Create.
	createReq := canvus.CreateAccessTokenRequest{
		Name:        fmt.Sprintf("example-02-auth-flows %s", time.Now().UTC().Format(time.RFC3339)),
		Description: "Temporary token created by mt-canvus-tools example 02-auth-flows.",
	}
	token, err := s.CreateAccessToken(ctx, userID, createReq)
	if err != nil {
		return fmt.Errorf("CreateAccessToken: %w", err)
	}
	slog.Info("access token created",
		"token_id", token.ID,
		"name", token.Name,
		"has_plain_token", token.PlainToken != "")

	// List.
	tokens, err := s.ListAccessTokens(ctx, userID)
	if err != nil {
		return fmt.Errorf("ListAccessTokens: %w", err)
	}
	found := false
	for _, t := range tokens {
		if t.ID == token.ID {
			found = true
			break
		}
	}
	slog.Info("listed access tokens",
		"count", len(tokens),
		"found_new_token", found)

	// Delete.
	if err := s.DeleteAccessToken(ctx, userID, token.ID); err != nil {
		return fmt.Errorf("DeleteAccessToken: %w", err)
	}
	slog.Info("access token deleted", "token_id", token.ID)
	return nil
}

// mustEnv returns the value of name or an error if it is empty/unset.
func mustEnv(name string) (string, error) {
	v := os.Getenv(name)
	if v == "" {
		return "", fmt.Errorf("missing required env var %s", name)
	}
	return v, nil
}

// setupLogging configures slog.Default with a text or JSON handler depending
// on LOG_FORMAT.
func setupLogging() {
	var h slog.Handler
	switch os.Getenv("LOG_FORMAT") {
	case "json":
		h = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	default:
		h = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	slog.SetDefault(slog.New(h))
}
