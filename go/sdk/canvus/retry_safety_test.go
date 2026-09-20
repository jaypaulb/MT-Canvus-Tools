package canvus_test

import (
	"context"
	"errors"
	"fmt"
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
			if tt.budget < 0 && !strings.Contains(err.Error(), "retry") {
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
