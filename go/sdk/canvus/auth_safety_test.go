package canvus_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func TestActorAuthoritySurvivesUnauthorized(t *testing.T) {
	for _, tt := range []struct{ name, key string }{{"guest_bootstrap", ""}, {"service_bootstrap", "synthetic-service"}} {
		t.Run(tt.name, func(t *testing.T) {
			received := make(chan []string, 8)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/users/login" {
					fmt.Fprint(w, `{"token":"synthetic-actor","user":{"id":7}}`)
					return
				}
				received <- r.Header.Values("Private-Token")
				w.WriteHeader(http.StatusUnauthorized)
			}))
			defer srv.Close()
			cfg := canvus.DefaultSessionConfig()
			cfg.BaseURL = srv.URL
			s := canvus.NewSession(cfg, canvus.WithAPIKey(tt.key))
			require.NoError(t, s.Login(context.Background(), "actor@example.invalid", "synthetic-password"))
			for i := 0; i < 2; i++ {
				_, err := s.GetNote(context.Background(), "c", "n")
				require.Error(t, err)
				assert.Equal(t, []string{"synthetic-actor"}, <-received)
			}
		})
	}
}

func TestLogoutDoesNotRestoreBootstrapService(t *testing.T) {
	headers := make(chan []string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/login":
			fmt.Fprint(w, `{"token":"synthetic-actor","user":{"id":7}}`)
		case "/users/logout":
			w.WriteHeader(http.StatusNoContent)
		default:
			headers <- r.Header.Values("Private-Token")
			fmt.Fprint(w, `{"id":"n"}`)
		}
	}))
	defer srv.Close()
	s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}, canvus.WithAPIKey("synthetic-service"))
	require.NoError(t, s.Login(context.Background(), "actor@example.invalid", "synthetic-password"))
	require.NoError(t, s.Logout(context.Background()))
	_, err := s.GetNote(context.Background(), "c", "n")
	require.NoError(t, err)
	assert.Empty(t, <-headers)
	assert.Zero(t, s.UserID())
}

func TestDefaultRedirectsDoNotReplayOrForwardAuthority(t *testing.T) {
	for _, name := range []string{"read", "write"} {
		t.Run(name, func(t *testing.T) {
			var forwarded atomic.Bool
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { forwarded.Store(true); fmt.Fprint(w, `{"id":"n"}`) }))
			defer target.Close()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
			}))
			defer srv.Close()
			s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}, canvus.WithToken("synthetic-actor"))
			var err error
			if name == "write" {
				_, err = s.CreateNote(context.Background(), "c", map[string]any{"text": "hello"})
			} else {
				_, err = s.GetNote(context.Background(), "c", "n")
			}
			require.Error(t, err)
			assert.False(t, forwarded.Load())
		})
	}
}

func TestSameOriginReadRedirectPreservesAuthentication(t *testing.T) {
	headers := make(chan []string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/canonical" {
			http.Redirect(w, r, "/canonical", http.StatusMovedPermanently)
			return
		}
		headers <- r.Header.Values("Private-Token")
		fmt.Fprint(w, `{"id":"n"}`)
	}))
	defer srv.Close()
	s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}, canvus.WithToken("synthetic-actor"))
	note, err := s.GetNote(context.Background(), "c", "n")
	require.NoError(t, err)
	assert.Equal(t, "n", note.ID)
	assert.Equal(t, []string{"synthetic-actor"}, <-headers)
}

func TestExplicitCrossOriginRedirectDoesNotLeakCredential(t *testing.T) {
	headers := make(chan []string, 1)
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers <- r.Header.Values("Private-Token")
		fmt.Fprint(w, `{"id":"n"}`)
	}))
	defer target.Close()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, http.StatusFound) }))
	defer srv.Close()
	custom := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return nil }}
	s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}, canvus.WithToken("synthetic-actor"), canvus.WithHTTPClient(custom))
	_, err := s.GetNote(context.Background(), "c", "n")
	require.NoError(t, err)
	assert.Empty(t, <-headers)
}

func TestDirectHTTPClientUsesSelectedActor(t *testing.T) {
	headers := make(chan []string, 3)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/login" {
			fmt.Fprint(w, `{"token":"synthetic-actor","user":{"id":7}}`)
			return
		}
		headers <- r.Header.Values("Private-Token")
		fmt.Fprint(w, `{"id":"n"}`)
	}))
	defer srv.Close()
	s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}, canvus.WithAPIKey("synthetic-service"))
	for _, want := range []string{"synthetic-service", "synthetic-actor"} {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
		require.NoError(t, err)
		req.Header.Add("Private-Token", "must-not-override-selected-authority")
		req.Header["private-token"] = []string{"must-not-bypass-case-normalization"}
		resp, err := s.HTTPClient.Do(req)
		require.NoError(t, err)
		_, _ = io.Copy(io.Discard, resp.Body)
		require.NoError(t, resp.Body.Close())
		assert.Equal(t, []string{want}, <-headers)
		if want == "synthetic-service" {
			require.NoError(t, s.Login(context.Background(), "actor@example.invalid", "synthetic-password"))
		}
	}
	// A new session reusing an SDK client must not inherit its auth transport.
	second := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}, canvus.WithHTTPClient(s.HTTPClient), canvus.WithToken("synthetic-second"))
	_, err := second.GetNote(context.Background(), "c", "n")
	require.NoError(t, err)
	assert.Equal(t, []string{"synthetic-second"}, <-headers)
}

func TestCustomClientAndConfigRemainIsolated(t *testing.T) {
	received := make(chan []string, 3)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r.Header.Values("Private-Token")
		fmt.Fprint(w, `{"id":"n"}`)
	}))
	defer srv.Close()
	original := &http.Client{Timeout: time.Second}
	cfg := canvus.DefaultSessionConfig()
	cfg.BaseURL = srv.URL
	cfg.HTTPClient = original
	first := canvus.NewSession(cfg, canvus.WithAPIKey("synthetic-first"))
	second := canvus.NewSession(cfg, canvus.WithAPIKey("synthetic-second"))
	for i, s := range []*canvus.Session{first, second, first} {
		_, err := s.GetNote(context.Background(), "c", "n")
		require.NoError(t, err)
		assert.Equal(t, []string{[]string{"synthetic-first", "synthetic-second", "synthetic-first"}[i]}, <-received)
	}
	assert.Nil(t, original.Transport)
	assert.Equal(t, time.Second, original.Timeout)
	assert.Empty(t, cfg.APIKey)
}
