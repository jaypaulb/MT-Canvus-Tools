package canvus_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func TestActorAuthoritySurvivesUnauthorized(t *testing.T) {
	for _, key := range []string{"", "synthetic-service"} {
		t.Run("bootstrap_"+key, func(t *testing.T) {
			var received [][]string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/users/login" {
					fmt.Fprint(w, `{"token":"synthetic-actor","user":{"id":7}}`)
					return
				}
				received = append(received, r.Header.Values("Private-Token"))
				w.WriteHeader(http.StatusUnauthorized)
			}))
			defer srv.Close()
			cfg := canvus.DefaultSessionConfig()
			cfg.BaseURL = srv.URL
			s := canvus.NewSession(cfg, canvus.WithAPIKey(key))
			if err := s.Login(context.Background(), "actor@example.invalid", "synthetic-password"); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 2; i++ {
				if _, err := s.GetNote(context.Background(), "c", "n"); err == nil {
					t.Fatal("expected 401")
				}
			}
			want := [][]string{{"synthetic-actor"}, {"synthetic-actor"}}
			if !reflect.DeepEqual(received, want) {
				t.Fatalf("credentials=%v, want %v", received, want)
			}
		})
	}
}

func TestLogoutDoesNotRestoreBootstrapService(t *testing.T) {
	var headers []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/login":
			fmt.Fprint(w, `{"token":"synthetic-actor","user":{"id":7}}`)
		case "/users/logout":
			w.WriteHeader(http.StatusNoContent)
		default:
			headers = r.Header.Values("Private-Token")
			fmt.Fprint(w, `{"id":"n"}`)
		}
	}))
	defer srv.Close()
	s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}, canvus.WithAPIKey("synthetic-service"))
	if err := s.Login(context.Background(), "actor@example.invalid", "synthetic-password"); err != nil {
		t.Fatal(err)
	}
	if err := s.Logout(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetNote(context.Background(), "c", "n"); err != nil {
		t.Fatal(err)
	}
	if len(headers) != 0 || s.UserID() != 0 {
		t.Fatalf("logout restored authority: headers=%v id=%d", headers, s.UserID())
	}
}

func TestDefaultRedirectsDoNotReplayOrForwardAuthority(t *testing.T) {
	for _, write := range []bool{false, true} {
		t.Run(fmt.Sprint(write), func(t *testing.T) {
			forwarded := false
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { forwarded = true; fmt.Fprint(w, `{"id":"n"}`) }))
			defer target.Close()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
			}))
			defer srv.Close()
			s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}, canvus.WithToken("synthetic-actor"))
			var err error
			if write {
				_, err = s.CreateNote(context.Background(), "c", map[string]any{"text": "hello"})
			} else {
				_, err = s.GetNote(context.Background(), "c", "n")
			}
			if err == nil || forwarded {
				t.Fatalf("redirect followed=%v error=%v", forwarded, err)
			}
		})
	}
}

func TestCustomClientAndConfigRemainIsolated(t *testing.T) {
	var received [][]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = append(received, r.Header.Values("Private-Token"))
		fmt.Fprint(w, `{"id":"n"}`)
	}))
	defer srv.Close()
	original := &http.Client{Timeout: time.Second}
	cfg := canvus.DefaultSessionConfig()
	cfg.BaseURL = srv.URL
	cfg.HTTPClient = original
	first := canvus.NewSession(cfg, canvus.WithAPIKey("synthetic-first"))
	second := canvus.NewSession(cfg, canvus.WithAPIKey("synthetic-second"))
	for _, s := range []*canvus.Session{first, second, first} {
		if _, err := s.GetNote(context.Background(), "c", "n"); err != nil {
			t.Fatal(err)
		}
	}
	want := [][]string{{"synthetic-first"}, {"synthetic-second"}, {"synthetic-first"}}
	if !reflect.DeepEqual(received, want) {
		t.Fatalf("credentials=%v", received)
	}
	if original.Transport != nil || original.Timeout != time.Second || cfg.APIKey != "" {
		t.Fatal("caller-owned configuration mutated")
	}
}
