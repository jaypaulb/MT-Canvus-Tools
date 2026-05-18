package qa

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeSession creates a Session pointed at the given httptest server.
func makeSession(t *testing.T, srv *httptest.Server) *canvus.Session {
	t.Helper()
	cfg := canvus.DefaultSessionConfig()
	cfg.BaseURL = srv.URL + "/api/v1"
	cfg.MaxRetries = 0
	return canvus.NewSession(cfg)
}

// noteJSON serialises a note's text field as the server stream format.
func noteJSON(t *testing.T, text string) []byte {
	t.Helper()
	b, err := json.Marshal(map[string]any{"id": "note-1", "text": text})
	require.NoError(t, err)
	return append(b, '\n')
}

// writeFlush writes bytes to a ResponseWriter and flushes.
func writeFlush(t *testing.T, w http.ResponseWriter, data []byte) {
	t.Helper()
	flusher, ok := w.(http.Flusher)
	require.True(t, ok, "ResponseWriter must support Flusher")
	_, err := w.Write(data)
	require.NoError(t, err)
	flusher.Flush()
}

// TestWaitForQuestion_SettlesOnQuestion verifies that a note text ending with
// "?" that remains stable for the settle window causes WaitForQuestion to
// return true.
func TestWaitForQuestion_SettlesOnQuestion(t *testing.T) {
	t.Setenv("QNOTE_SETTLE_MS", "50")

	// serverDone is closed by the handler once it has written the event, so the
	// test can verify the result without racing against the server goroutine.
	serverDone := make(chan struct{})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(serverDone)
		writeFlush(t, w, noteJSON(t, "Is this a question?"))
		// Hold the connection open until the subscribe context closes so the
		// stream close happens after the settle timer fires.
		<-r.Context().Done()
	}))
	defer srv.Close()

	s := makeSession(t, srv)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := WaitForQuestion(ctx, s, "canvas-1", "note-1", 5*time.Second)

	assert.True(t, result, "expected true: question text arrived and settled")
	<-serverDone
}

// TestWaitForQuestion_ResetOnTextChange verifies that when the note text
// changes mid-settle (even to another "?"-ending text), the settle timer
// resets, and WaitForQuestion still returns true once the text is stable.
func TestWaitForQuestion_ResetOnTextChange(t *testing.T) {
	t.Setenv("QNOTE_SETTLE_MS", "50")

	// Gate signals that WaitForQuestion has received the first event and the
	// handler may proceed to send the second event.
	gate := make(chan struct{})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First event: qualifying question text.
		writeFlush(t, w, noteJSON(t, "First question?"))

		// Wait until the context on the subscribe side acknowledges the first
		// event. We use a brief pause here because we cannot hook into
		// WaitForQuestion's internal state; the pause ensures the settle timer
		// is running before we send the next event.
		select {
		case <-gate:
		case <-r.Context().Done():
			return
		}

		// Second event: different qualifying text — settle timer should reset.
		writeFlush(t, w, noteJSON(t, "Second question?"))

		// Hold connection open until the settle timer fires and WaitForQuestion
		// returns (which cancels tctx, which cancels r.Context).
		<-r.Context().Done()
	}))
	defer srv.Close()

	s := makeSession(t, srv)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Run WaitForQuestion in a goroutine so we can signal the gate.
	resultCh := make(chan bool, 1)
	go func() {
		resultCh <- WaitForQuestion(ctx, s, "canvas-1", "note-1", 5*time.Second)
	}()

	// Give WaitForQuestion a moment to receive and process the first event, then
	// release the gate so the server sends the second event.
	time.Sleep(20 * time.Millisecond)
	close(gate)

	result := <-resultCh
	assert.True(t, result, "expected true: stable question detected after mid-settle text change")
}

// TestWaitForQuestion_StreamCloseBeforeQuestion verifies that if the
// subscription stream closes without any "?"-ending event, WaitForQuestion
// returns false.
func TestWaitForQuestion_StreamCloseBeforeQuestion(t *testing.T) {
	t.Setenv("QNOTE_SETTLE_MS", "50")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send a non-question event, then close the stream immediately.
		writeFlush(t, w, noteJSON(t, "Not a question."))
		// Return without blocking; HTTP server closes the response body.
	}))
	defer srv.Close()

	s := makeSession(t, srv)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := WaitForQuestion(ctx, s, "canvas-1", "note-1", 5*time.Second)

	assert.False(t, result, "expected false: stream closed before any question was detected")
}

// TestWaitForQuestion_ContextCancellation verifies that cancelling the parent
// context causes WaitForQuestion to return false promptly.
func TestWaitForQuestion_ContextCancellation(t *testing.T) {
	t.Setenv("QNOTE_SETTLE_MS", "50")

	// ready is closed once the server is holding the connection open (i.e.
	// WaitForQuestion is blocked in its select loop).
	ready := make(chan struct{})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do not send any events; just keep the stream open.
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		close(ready)
		<-r.Context().Done()
	}))
	defer srv.Close()

	s := makeSession(t, srv)
	ctx, cancel := context.WithCancel(context.Background())

	resultCh := make(chan bool, 1)
	go func() {
		resultCh <- WaitForQuestion(ctx, s, "canvas-1", "note-1", 10*time.Second)
	}()

	// Wait until the server is holding the connection, then cancel the context.
	<-ready
	cancel()

	select {
	case result := <-resultCh:
		assert.False(t, result, "expected false: context was cancelled")
	case <-time.After(3 * time.Second):
		t.Fatal("WaitForQuestion did not return within 3 seconds after context cancellation")
	}
}
