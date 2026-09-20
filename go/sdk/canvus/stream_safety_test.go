package canvus_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func TestSubscriptionPreservesFrameAndFieldPresence(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "[]\n[{\"index\":0},{\"view_rectangle\":null}]")
	}))
	defer srv.Close()
	s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL})
	stream, err := canvus.Subscribe[canvus.Workspace](context.Background(), s, "clients/client/workspaces")
	require.NoError(t, err)
	defer stream.Close()
	first := <-stream.Frames
	require.Empty(t, first.Items)
	require.JSONEq(t, "[]", string(first.Raw))
	second := <-stream.Frames
	require.Len(t, second.Items, 2)
	require.JSONEq(t, `{"index":0}`, string(second.Items[0].Raw))
	require.JSONEq(t, `{"view_rectangle":null}`, string(second.Items[1].Raw))
	_, ok := <-stream.Frames
	require.False(t, ok)
	require.ErrorIs(t, stream.Wait(context.Background()), io.EOF)
}

func TestSubscriptionReportsTerminalFailures(t *testing.T) {
	for _, tt := range []struct {
		name, body string
		want       error
	}{
		{"malformed", `{"index":`, canvus.ErrInvalidStreamFrame},
		{"non_object", `[null]`, canvus.ErrInvalidStreamFrame},
		{"oversize", strings.Repeat("x", canvus.MaxSubscriptionFrameBytes+1), canvus.ErrStreamFrameTooLarge},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, tt.body) }))
			defer srv.Close()
			stream, err := canvus.Subscribe[canvus.Workspace](context.Background(), canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}), "clients/c/workspaces")
			require.NoError(t, err)
			defer stream.Close()
			for range stream.Frames {
			}
			require.ErrorIs(t, stream.Wait(context.Background()), tt.want)
		})
	}
}

func TestSubscriptionEstablishmentRemainsBounded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer srv.Close()
	cfg := canvus.DefaultSessionConfig()
	cfg.BaseURL = srv.URL
	cfg.RequestTimeout = 50 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := canvus.Subscribe[canvus.Workspace](ctx, canvus.NewSession(cfg), "clients/c/workspaces")
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.NoError(t, ctx.Err(), "establishment should use the ordinary timeout, not wait for lifetime cancellation")
}

func TestSubscriptionCancelReleasesBackpressure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for i := 0; i < 20; i++ {
			fmt.Fprintln(w, `{"index":0}`)
		}
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stream, err := canvus.Subscribe[canvus.Workspace](ctx, canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}, canvus.WithSubscribeBuffer(1)), "clients/c/workspaces")
	require.NoError(t, err)
	stream.Close()
	require.ErrorIs(t, stream.Wait(ctx), context.Canceled)
}

func TestSubscriptionOutlivesOrdinaryRequestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"index":3,"canvas_id":"c"}`)
		w.(http.Flusher).Flush()
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				fmt.Fprintln(w)
				w.(http.Flusher).Flush()
			}
		}
	}))
	defer srv.Close()
	cfg := canvus.DefaultSessionConfig()
	cfg.BaseURL = srv.URL
	cfg.RequestTimeout = 100 * time.Millisecond
	s := canvus.NewSession(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, err := s.SubscribeClientWorkspaces(ctx, "client")
	require.NoError(t, err)
	_, ok := <-ch
	require.True(t, ok)
	select {
	case <-ch:
		t.Fatal("heartbeat stream ended at ordinary request timeout")
	case <-time.After(300 * time.Millisecond):
	}
	cancel()
	select {
	case _, ok := <-ch:
		require.False(t, ok)
	case <-time.After(time.Second):
		t.Fatal("cancellation did not close stream")
	}
}
