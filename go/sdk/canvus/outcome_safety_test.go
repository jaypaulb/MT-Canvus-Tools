package canvus_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func TestAcceptedResponseFailuresRetainOutcome(t *testing.T) {
	for _, tt := range []struct {
		name, body, id string
		truncated      bool
	}{
		{"malformed", `{"id":`, "", false},
		{"wrong_field_type", `{"id":"created","text":42}`, "created", false},
		{"numeric_resource_id", `{"id":42,"text":"hello"}`, "42", false},
		{"truncated", `{"id":"created"}`, "created", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.Header().Set("X-Request-ID", "synthetic-request")
				if tt.truncated {
					w.Header().Set("Content-Length", "1000")
				}
				w.WriteHeader(201)
				fmt.Fprint(w, tt.body)
			}))
			defer srv.Close()
			cfg := canvus.DefaultSessionConfig()
			cfg.BaseURL = srv.URL
			_, err := canvus.NewSession(cfg).CreateNote(context.Background(), "c", map[string]any{"text": "hello"})
			var accepted *canvus.AcceptedResponseError
			if !errors.As(err, &accepted) {
				t.Fatalf("not an accepted response error: %v", err)
			}
			if accepted.StatusCode != 201 || accepted.ResourceID != tt.id || accepted.RequestID != "synthetic-request" || calls.Load() != 1 {
				t.Fatalf("outcome=%+v attempts=%d", accepted, calls.Load())
			}
			if tt.truncated && !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("lost read error: %v", err)
			}
			if tt.name == "wrong_field_type" {
				var te *json.UnmarshalTypeError
				if !errors.As(err, &te) {
					t.Fatalf("lost decode error: %v", err)
				}
			}
			if strings.Contains(err.Error(), "synthetic-request") {
				t.Fatal("raw metadata should not be interpolated into error message")
			}
		})
	}
}

func TestEmptySuccessfulWritesRemainSuccessful(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusNoContent} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) }))
			defer srv.Close()
			s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL})
			if _, err := s.UpdateNote(context.Background(), "c", "n", map[string]any{"text": "hello"}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestReadDecodeFailureDoesNotImplyAcceptedMutation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{`) }))
	defer srv.Close()
	_, err := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}).GetNote(context.Background(), "c", "n")
	var accepted *canvus.AcceptedResponseError
	if err == nil || errors.As(err, &accepted) {
		t.Fatalf("read result: %v", err)
	}
}

func TestAcceptedWriteDoesNotRequireEchoEquality(t *testing.T) {
	for _, tt := range []struct {
		name, body string
		request    map[string]any
	}{
		{"omitted_title", `{"id":"created","text":"hello"}`, map[string]any{"text": "hello", "title": "requested"}},
		{"normalized_colour", `{"id":"created","background_color":"#ffffffbf"}`, map[string]any{"background_color": "#FFFFFFBF"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(201); fmt.Fprint(w, tt.body) }))
			defer srv.Close()
			cfg := canvus.DefaultSessionConfig()
			cfg.BaseURL = srv.URL
			note, err := canvus.NewSession(cfg).CreateNote(context.Background(), "c", tt.request)
			if err != nil {
				t.Fatalf("accepted write reported as failure: %v", err)
			}
			if note == nil || note.ID != "created" {
				t.Fatalf("lost created ID: %+v", note)
			}
		})
	}
}
