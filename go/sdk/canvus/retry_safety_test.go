package canvus_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func TestRetryPolicyAtHTTPBoundary(t *testing.T) {
	for _, tt := range []struct {
		name   string
		budget int
		write  bool
		want   int
	}{
		{"zero_read", 0, false, 1}, {"default_read", 3, false, 4}, {"zero_write", 0, true, 1}, {"configured_write", 3, true, 1}, {"negative", -1, false, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(503)
				fmt.Fprint(w, `{"msg":"synthetic unavailable"}`)
			}))
			defer srv.Close()
			cfg := canvus.DefaultSessionConfig()
			cfg.BaseURL = srv.URL
			cfg.MaxRetries = tt.budget
			cfg.RetryWaitMin = time.Millisecond
			cfg.RetryWaitMax = time.Millisecond
			s := canvus.NewSession(cfg)
			var err error
			if tt.write {
				_, err = s.CreateNote(context.Background(), "c", map[string]any{"text": "synthetic"})
			} else {
				_, err = s.GetNote(context.Background(), "c", "n")
			}
			if err == nil {
				t.Fatal("expected error")
			}
			if got := int(calls.Load()); got != tt.want {
				t.Fatalf("attempts=%d want %d", got, tt.want)
			}
			if tt.budget < 0 && !errors.Is(err, canvus.ErrInvalidRetryBudget) {
				t.Fatalf("not a configuration error: %v", err)
			}
		})
	}
}

func TestReadRetryWaitHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { cancel(); w.WriteHeader(503) }))
	defer srv.Close()
	cfg := canvus.DefaultSessionConfig()
	cfg.BaseURL = srv.URL
	cfg.RetryWaitMin = 2 * time.Second
	cfg.RetryWaitMax = 2 * time.Second
	start := time.Now()
	_, err := canvus.NewSession(cfg).GetNote(ctx, "c", "n")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("retry sleep ignored cancellation")
	}
}

type safetyRoundTripper func(*http.Request) (*http.Response, error)

func (f safetyRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCancelledBackoffRetainsServerError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := &http.Client{Transport: safetyRoundTripper(func(r *http.Request) (*http.Response, error) {
		cancel()
		return &http.Response{StatusCode: 503, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"msg":"synthetic"}`)), Request: r}, nil
	})}
	cfg := canvus.DefaultSessionConfig()
	cfg.BaseURL = "https://example.invalid"
	_, err := canvus.NewSession(cfg, canvus.WithHTTPClient(client)).GetNote(ctx, "c", "n")
	var apiErr *canvus.APIError
	if !errors.Is(err, context.Canceled) || !errors.As(err, &apiErr) || apiErr.StatusCode != 503 {
		t.Fatalf("lost cancellation or server status: %v", err)
	}
}

func TestRetryClassificationRejectsAcceptedAndLocalFailures(t *testing.T) {
	for _, err := range []error{&canvus.AcceptedResponseError{StatusCode: 201, Err: io.ErrUnexpectedEOF}, canvus.ErrInvalidRetryBudget, canvus.ErrInvalidRequest, errors.New("local failure"), context.Canceled} {
		if canvus.IsRetryableError(err) {
			t.Fatalf("unsafe retry hint for %v", err)
		}
	}
	if !canvus.IsRetryableError(&canvus.APIError{StatusCode: 503}) {
		t.Fatal("lost transient read classification")
	}
}

func TestBodylessWriteRetainsJSONContentType(t *testing.T) {
	contentType := make(chan string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType <- r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}, canvus.WithAPIKey("synthetic-key"))
	if err := s.ApproveUser(context.Background(), 7); err != nil {
		t.Fatal(err)
	}
	if got := <-contentType; got != "application/json" {
		t.Fatalf("content type=%q", got)
	}
}

func TestUploadIsNotAutomaticallyRetried(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); w.WriteHeader(503) }))
	defer srv.Close()
	cfg := canvus.DefaultSessionConfig()
	cfg.BaseURL = srv.URL
	_, err := canvus.NewSession(cfg).UploadNote(context.Background(), "c", strings.NewReader("synthetic upload"), "text/plain")
	if err == nil || calls.Load() != 1 {
		t.Fatalf("attempts=%d error=%v", calls.Load(), err)
	}
}

func TestLostWriteResponseIsNotReplayed(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		_ = conn.Close()
	}))
	defer srv.Close()
	cfg := canvus.DefaultSessionConfig()
	cfg.BaseURL = srv.URL
	_, err := canvus.NewSession(cfg).CreateNote(context.Background(), "c", map[string]any{"text": "hello"})
	if err == nil || calls.Load() != 1 {
		t.Fatalf("attempts=%d error=%v", calls.Load(), err)
	}
	var accepted *canvus.AcceptedResponseError
	if errors.As(err, &accepted) {
		t.Fatal("no response is not proof of HTTP acceptance")
	}
}

func TestBatchDoesNotReplayCopy(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); w.WriteHeader(503) }))
	defer srv.Close()
	cfg := canvus.DefaultSessionConfig()
	cfg.BaseURL = srv.URL
	cfg.RetryWaitMin = time.Millisecond
	cfg.RetryWaitMax = time.Millisecond
	bc := canvus.DefaultBatchConfig()
	bc.RetryAttempts = 2
	bc.RetryDelay = time.Millisecond
	bp := canvus.NewBatchProcessor(canvus.NewSession(cfg), bc)
	ops := canvus.NewBatchOperationBuilder().Copy("copy", &canvus.Canvas{ID: "c"}, "folder").Build()
	results, err := bp.ExecuteBatch(context.Background(), ops)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Success || results[0].Error == nil {
		t.Fatalf("results=%+v", results)
	}
	if calls.Load() != 1 || results[0].Retries != 0 {
		t.Fatalf("copy attempts=%d retries=%d", calls.Load(), results[0].Retries)
	}
}
