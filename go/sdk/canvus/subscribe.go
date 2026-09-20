// Package canvus subscribe helpers.
//
// Phase 4b §4.1 #5-#10: typed NDJSON subscribe helpers for every streamable
// endpoint, parity with the Python/TS streaming surface. All helpers share
// the generic subscribeStream primitive below.
//
// Wire format: the Canvus server returns a long-lived HTTPS response with
// `subscribe=true` query param; each newline-delimited chunk is a JSON
// document representing either a full resource object or a sparse update.
// The generic primitive scans line-by-line, decodes each non-empty line
// into a fresh T, and pushes it to a buffered channel. The channel is
// closed when the context is cancelled, the stream ends, or a decode/IO
// error occurs.
package canvus

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

// subscribeStream opens a long-lived GET on the given endpoint with
// `subscribe=true` and decodes each newline-delimited JSON document into a
// fresh T, pushing it onto the returned channel. The channel is closed when
// ctx is cancelled, the stream EOFs, or a decode/IO error occurs. A
// `keepalive` blank line (empty string after trim) is silently skipped.
//
// Callers SHOULD drain the channel or cancel ctx; leaking the channel will
// leak the underlying response body. The channel capacity is controlled by
// SessionConfig.SubscribeBuffer (default 4, set via WithSubscribeBuffer).
// Phase 4d Round B: configurable via SessionConfig.SubscribeBuffer.
func subscribeStream[T any](ctx context.Context, s *Session, endpoint string) (<-chan T, error) {
	if s == nil {
		return nil, fmt.Errorf("subscribeStream: nil session")
	}
	if err := s.validateRetryBudget(); err != nil {
		return nil, fmt.Errorf("subscribeStream: %w", err)
	}
	u, err := url.Parse(s.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("subscribeStream: invalid base URL: %w", err)
	}
	u.Path = path.Join(u.Path, endpoint)
	q := u.Query()
	q.Set("subscribe", "true")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("subscribeStream: new request: %w", err)
	}
	req.Header.Set("Accept", "application/x-ndjson, application/json")
	req.Header.Set("User-Agent", s.config.UserAgent)
	req = withRequestAuthority(req, s.requestAuthenticator())
	if s.config.RequestIDFunc != nil {
		if id := s.config.RequestIDFunc(); id != "" {
			req.Header.Set("X-Request-ID", id)
		}
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("subscribeStream: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, &APIError{StatusCode: resp.StatusCode, Message: string(body)}
	}

	bufSize := s.config.SubscribeBuffer
	if bufSize < 1 {
		bufSize = 4 // defensive default if config was not built via DefaultSessionConfig.
	}
	ch := make(chan T, bufSize)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		// Allow large frames; the default 64 KiB is too small for full canvas snapshots.
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 8*1024*1024)
		for scanner.Scan() {
			line := bytes.TrimSpace(scanner.Bytes())
			if len(line) == 0 {
				continue
			}
			// Server batches updates: snapshot is `[obj, obj, …]`, deltas may
			// be a single `{obj}` or `[obj, …]`. Peek at the first byte to
			// pick a decode shape. Mirrors Python `_typed_subscribe` behaviour.
			if line[0] == '[' {
				var batch []T
				if err := json.Unmarshal(line, &batch); err != nil {
					s.logger.Debug("subscribeStream: batch decode error, skipping line", "endpoint", endpoint, "err", err)
					continue
				}
				for _, item := range batch {
					select {
					case <-ctx.Done():
						return
					case ch <- item:
					}
				}
				continue
			}
			var item T
			if err := json.Unmarshal(line, &item); err != nil {
				s.logger.Debug("subscribeStream: decode error, skipping line", "endpoint", endpoint, "err", err)
				continue
			}
			select {
			case <-ctx.Done():
				return
			case ch <- item:
			}
		}
	}()
	return ch, nil
}
