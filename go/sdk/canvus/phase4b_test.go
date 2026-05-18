// Phase 4b §4.1 #1-#10, #16 test coverage.
package canvus

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromEnv_RequiresBaseURL(t *testing.T) {
	t.Setenv("CANVUS_API_URL", "")
	_, err := FromEnv()
	require.Error(t, err)
}

func TestFromEnv_HappyPath(t *testing.T) {
	t.Setenv("CANVUS_API_URL", "https://example.invalid/api/v1")
	t.Setenv("CANVUS_API_KEY", "secret")
	t.Setenv("CANVUS_TIMEOUT_MS", "1234")
	t.Setenv("CANVUS_VERIFY_TLS", "false")
	s, err := FromEnv()
	require.NoError(t, err)
	assert.Equal(t, "https://example.invalid/api/v1", s.BaseURL)
	assert.Equal(t, 1234*time.Millisecond, s.config.RequestTimeout)
}

func TestFromEnv_InvalidTimeout(t *testing.T) {
	t.Setenv("CANVUS_API_URL", "https://example.invalid/api/v1")
	t.Setenv("CANVUS_TIMEOUT_MS", "notanumber")
	_, err := FromEnv()
	require.Error(t, err)
}

func TestFromEnv_InvalidVerifyTLS(t *testing.T) {
	t.Setenv("CANVUS_API_URL", "https://example.invalid/api/v1")
	t.Setenv("CANVUS_VERIFY_TLS", "maybe")
	_, err := FromEnv()
	require.Error(t, err)
}

func TestWithConnectTimeout_InstallsTransport(t *testing.T) {
	cfg := &SessionConfig{BaseURL: "https://example.invalid/api/v1"}
	s := NewSession(cfg, WithConnectTimeout(2*time.Second))
	require.NotNil(t, s.HTTPClient.Transport, "expected custom transport when ConnectTimeout set")
}

func TestWithRequestIDFunc_InjectsHeader(t *testing.T) {
	var seen atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.Store(r.Header.Get("X-Request-ID"))
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	cfg := DefaultSessionConfig()
	cfg.BaseURL = srv.URL + "/api/v1"
	s := NewSession(cfg, WithRequestIDFunc(func() string { return "req-42" }))
	_, err := s.ListCanvases(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "req-42", seen.Load())
}

func TestWithRequestIDFunc_SurfacesOnAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"not_found","message":"no"}`, http.StatusNotFound)
	}))
	defer srv.Close()

	cfg := DefaultSessionConfig()
	cfg.BaseURL = srv.URL + "/api/v1"
	cfg.MaxRetries = 0
	s := NewSession(cfg, WithRequestIDFunc(func() string { return "req-77" }))

	_, err := s.GetCanvas(context.Background(), "missing")
	require.Error(t, err)
	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, "req-77", apiErr.RequestID)
}

func TestErrUnsupportedOperation_WrapsCreateRejection(t *testing.T) {
	cfg := DefaultSessionConfig()
	cfg.BaseURL = "https://example.invalid/api/v1"
	s := NewSession(cfg)
	_, err := s.CreateWidget(context.Background(), "c", map[string]any{"widget_type": "ip_video"})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrWidgetTypeNotCreatable)
	assert.ErrorIs(t, err, ErrUnsupportedOperation, "expected ErrUnsupportedOperation sentinel")
}

func TestGetCurrentUser_RequiresLogin(t *testing.T) {
	cfg := DefaultSessionConfig()
	cfg.BaseURL = "https://example.invalid/api/v1"
	s := NewSession(cfg)
	_, err := s.GetCurrentUser(context.Background())
	require.Error(t, err)
}

func TestGetCurrentUser_DispatchesToGetUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/users/42") {
			http.Error(w, "wrong path: "+r.URL.Path, http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 42, "email": "x@y", "name": "me"})
	}))
	defer srv.Close()
	cfg := DefaultSessionConfig()
	cfg.BaseURL = srv.URL + "/api/v1"
	s := NewSession(cfg)
	s.userID = 42
	u, err := s.GetCurrentUser(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(42), u.ID)
	assert.Equal(t, "x@y", u.Email)
}

func TestSubscribeCanvases_StreamsNDJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("subscribe"))
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"c1","name":"first"}` + "\n"))
		flusher.Flush()
		_, _ = w.Write([]byte("\n"))                                 // keepalive blank line
		_, _ = w.Write([]byte(`{"id":"c2","name":"second"}` + "\n")) // second record
		flusher.Flush()
	}))
	defer srv.Close()

	cfg := DefaultSessionConfig()
	cfg.BaseURL = srv.URL + "/api/v1"
	s := NewSession(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ch, err := s.SubscribeCanvases(ctx)
	require.NoError(t, err)

	var got []Canvas
	for c := range ch {
		got = append(got, c)
		if len(got) == 2 {
			cancel()
		}
	}
	require.Len(t, got, 2)
	assert.Equal(t, "c1", got[0].ID)
	assert.Equal(t, "c2", got[1].ID)
}

func TestSubscribeCanvases_HandlesBatchedSnapshot(t *testing.T) {
	// Real server behaviour: first frame is a snapshot array of all resources,
	// then individual delta objects. The primitive must yield each array item.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":"c1","name":"first"},{"id":"c2","name":"second"}]` + "\n"))
		flusher.Flush()
		_, _ = w.Write([]byte(`{"id":"c3","name":"third"}` + "\n"))
		flusher.Flush()
	}))
	defer srv.Close()

	cfg := DefaultSessionConfig()
	cfg.BaseURL = srv.URL + "/api/v1"
	s := NewSession(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ch, err := s.SubscribeCanvases(ctx)
	require.NoError(t, err)

	var got []Canvas
	for c := range ch {
		got = append(got, c)
		if len(got) == 3 {
			cancel()
		}
	}
	require.Len(t, got, 3)
	assert.Equal(t, []string{"c1", "c2", "c3"}, []string{got[0].ID, got[1].ID, got[2].ID})
}

func TestSubscribeStream_PropagatesAPIErrorOnNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"forbidden","message":"nope"}`, http.StatusForbidden)
	}))
	defer srv.Close()
	cfg := DefaultSessionConfig()
	cfg.BaseURL = srv.URL + "/api/v1"
	s := NewSession(cfg)
	_, err := s.SubscribeCanvases(context.Background())
	require.Error(t, err)
	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
}

// TestWithSubscribeBuffer verifies that WithSubscribeBuffer sets the channel
// capacity used by subscribeStream. Phase 4d Round B.
func TestWithSubscribeBuffer_CustomSize(t *testing.T) {
	// A server that sends one canvas and then blocks so we can inspect the
	// channel without draining it.
	ready := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"c1","name":"buf-test"}` + "\n"))
		flusher.Flush()
		<-ready // hold the connection open so the goroutine stays alive
	}))
	defer func() {
		close(ready)
		srv.Close()
	}()

	t.Run("buffer=16", func(t *testing.T) {
		cfg := DefaultSessionConfig()
		cfg.BaseURL = srv.URL + "/api/v1"
		s := NewSession(cfg, WithSubscribeBuffer(16))
		assert.Equal(t, 16, s.config.SubscribeBuffer, "config field should reflect option")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		ch, err := s.SubscribeCanvases(ctx)
		require.NoError(t, err)
		assert.Equal(t, 16, cap(ch), "channel capacity should be 16")
		cancel()
	})

	t.Run("default=4", func(t *testing.T) {
		cfg := DefaultSessionConfig()
		cfg.BaseURL = srv.URL + "/api/v1"
		s := NewSession(cfg)
		assert.Equal(t, 4, s.config.SubscribeBuffer, "default SubscribeBuffer should be 4")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		ch, err := s.SubscribeCanvases(ctx)
		require.NoError(t, err)
		assert.Equal(t, 4, cap(ch), "channel capacity should be 4")
		cancel()
	})
}
