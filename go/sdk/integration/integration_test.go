//go:build integration

// Package integration contains live-server integration tests for the SDK.
// Run with:
//
//	CANVUS_API_KEY=... CANVUS_BASE_URL=https://server/api/v1/ \
//	  go test -tags=integration ./integration/...
//
// Tests are skipped if either env var is missing.
package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func newLiveSession(t *testing.T) *canvus.Session {
	t.Helper()
	baseURL := os.Getenv("CANVUS_BASE_URL")
	apiKey := os.Getenv("CANVUS_API_KEY")
	if baseURL == "" || apiKey == "" {
		t.Skip("CANVUS_BASE_URL and CANVUS_API_KEY must be set to run integration tests")
	}
	cfg := canvus.DefaultSessionConfig()
	cfg.BaseURL = baseURL
	return canvus.NewSession(cfg, canvus.WithAPIKey(apiKey))
}

// TestLive_ServerInfo is the canonical smoke test: it exercises auth and a
// simple read endpoint. If this passes, the SDK and credentials are wired up
// correctly.
func TestLive_ServerInfo(t *testing.T) {
	s := newLiveSession(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := s.GetServerInfo(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, info.Version, "server returned an empty version string")
	t.Logf("server: %s, version: %s, id: %s", info.Version, info.Go, info.ServerID)
}

// TestLive_ListCanvases verifies a non-trivial GET round-trip.
func TestLive_ListCanvases(t *testing.T) {
	s := newLiveSession(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	canvases, err := s.ListCanvases(ctx, nil)
	require.NoError(t, err)
	t.Logf("server reports %d canvases", len(canvases))
}
