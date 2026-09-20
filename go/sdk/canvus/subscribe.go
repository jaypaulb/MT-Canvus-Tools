package canvus

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

// MaxSubscriptionFrameBytes bounds an individual NDJSON frame, including its newline.
const MaxSubscriptionFrameBytes = 8 * 1024 * 1024

// StreamItem retains decoded data and its original object, including absent,
// explicit null, zero, unknown fields and deletion markers. Value alone is not a delta.
type StreamItem[T any] struct {
	Value T
	Raw   json.RawMessage
}

// StreamFrame is one wire frame. Empty arrays and array/object boundaries are
// preserved. No snapshot/delta/merge semantics are inferred from its shape.
type StreamFrame[T any] struct {
	Items []StreamItem[T]
	Raw   json.RawMessage
}

// Subscription owns one connection without automatic reconnect or state merging.
// Drain Frames or call Close; otherwise backpressure intentionally pauses reads.
// After Frames closes, Wait reports EOF, cancellation, decode or read failure.
type Subscription[T any] struct {
	Frames <-chan StreamFrame[T]
	done   chan struct{}
	cancel context.CancelCauseFunc
	err    error // written before done closes; read only after that synchronization
}

// Close cancels the subscription, including blocked reads and backpressure.
// Use Wait to observe worker termination. Close is safe to call repeatedly.
func (s *Subscription[T]) Close() { s.cancel(context.Canceled) }

// Wait waits for termination and returns its cause (including io.EOF on normal
// server closure). Cancelling this wait does not cancel the subscription.
func (s *Subscription[T]) Wait(ctx context.Context) error {
	select {
	case <-s.done:
		return s.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Subscribe opens an authenticated, presence-preserving NDJSON subscription on
// a relative resource endpoint, e.g. "clients/<id>/workspaces". The ordinary
// client timeout bounds establishment/error responses, NOT the stream lifetime;
// the supplied context owns that lifetime. The client/transport/TLS settings are
// retained without mutating the session's HTTP client. Reconnect explicitly and
// re-establish a snapshot; never merge sparse objects into zero-valued DTOs.
func Subscribe[T any](parent context.Context, s *Session, endpoint string) (*Subscription[T], error) {
	if s == nil {
		return nil, fmt.Errorf("Subscribe: %w: nil session", ErrInvalidRequest)
	}
	if err := s.validateRequestConfig(); err != nil {
		return nil, fmt.Errorf("Subscribe: %w", err)
	}
	relative, err := url.Parse(endpoint)
	if err != nil || endpoint == "" || relative.IsAbs() || relative.Host != "" || strings.HasPrefix(relative.Path, "/") || relative.RawQuery != "" || relative.Fragment != "" {
		return nil, fmt.Errorf("Subscribe: %w: relative resource path required", ErrInvalidRequest)
	}
	for _, part := range strings.Split(relative.Path, "/") {
		if part == ".." || part == "." {
			return nil, fmt.Errorf("Subscribe: %w: path traversal", ErrInvalidRequest)
		}
	}
	u, err := url.Parse(s.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("Subscribe URL: %w", err)
	}
	u.Path = path.Join(u.Path, relative.Path)
	u.RawPath = ""
	q := u.Query()
	q.Set("subscribe", "true")
	u.RawQuery = q.Encode()
	ctx, cancel := context.WithCancelCause(parent)
	handedOff := false
	defer func() {
		if !handedOff {
			cancel(context.Canceled)
		}
	}()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Subscribe request: %w", err)
	}
	req.Header.Set("Accept", "application/x-ndjson, application/json")
	req.Header.Set("User-Agent", s.config.UserAgent)
	req = withRequestAuthority(req, s.authorityForContext(ctx))
	if s.config.RequestIDFunc != nil {
		if id := s.config.RequestIDFunc(); id != "" {
			req.Header.Set("X-Request-ID", id)
		}
	}
	budget := s.HTTPClient.Timeout
	if budget <= 0 {
		budget = s.config.RequestTimeout
	}
	if budget > 0 {
		timer := time.AfterFunc(budget, func() { cancel(context.DeadlineExceeded) })
		defer timer.Stop()
	}
	client := *s.HTTPClient
	client.Timeout = 0
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Subscribe open: %w", errors.Join(err, context.Cause(ctx)))
	}
	closeBody := sync.OnceFunc(func() { _ = resp.Body.Close() })
	stopClose := context.AfterFunc(ctx, closeBody)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer stopClose()
		defer closeBody()
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, MaxSubscriptionFrameBytes))
		return nil, fmt.Errorf("Subscribe response: %w", errors.Join(s.handleErrorResponse(resp, body, 0), readErr, context.Cause(ctx)))
	}
	capacity := s.config.SubscribeBuffer
	if capacity < 1 {
		capacity = 4
	}
	frames := make(chan StreamFrame[T], capacity)
	stream := &Subscription[T]{Frames: frames, done: make(chan struct{}), cancel: cancel}
	handedOff = true
	go func() {
		defer close(stream.done)
		defer close(frames)
		defer cancel(context.Canceled)
		defer stopClose()
		defer closeBody()
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), MaxSubscriptionFrameBytes)
		for scanner.Scan() {
			line := bytes.TrimSpace(scanner.Bytes())
			if len(line) == 0 {
				continue
			}
			frame, err := decodeStreamFrame[T](line)
			if err != nil {
				stream.err = fmt.Errorf("Subscribe decode: %w", errors.Join(ErrInvalidStreamFrame, err, scanner.Err(), context.Cause(ctx)))
				return
			}
			select {
			case frames <- frame:
			case <-ctx.Done():
				stream.err = context.Cause(ctx)
				return
			}
		}
		if err := context.Cause(ctx); err != nil {
			stream.err = errors.Join(err, scanner.Err())
		} else if err := scanner.Err(); err != nil {
			if errors.Is(err, bufio.ErrTooLong) {
				err = errors.Join(ErrStreamFrameTooLarge, err)
			}
			stream.err = fmt.Errorf("Subscribe read: %w", err)
		} else {
			stream.err = io.EOF
		}
	}()
	return stream, nil
}

func decodeStreamFrame[T any](line []byte) (StreamFrame[T], error) {
	frame := StreamFrame[T]{Raw: append(json.RawMessage(nil), line...)}
	var objects []json.RawMessage
	if line[0] == '[' {
		if err := json.Unmarshal(line, &objects); err != nil {
			return frame, errors.Join(ErrInvalidStreamFrame, err)
		}
	} else {
		objects = []json.RawMessage{frame.Raw}
	}
	for _, raw := range objects {
		raw = bytes.TrimSpace(raw)
		if len(raw) == 0 || raw[0] != '{' {
			return frame, ErrInvalidStreamFrame
		}
		var value T
		if err := json.Unmarshal(raw, &value); err != nil {
			return frame, errors.Join(ErrInvalidStreamFrame, err)
		}
		frame.Items = append(frame.Items, StreamItem[T]{Value: value, Raw: raw})
	}
	return frame, nil
}

// subscribeStream adapts the legacy value-only helpers. These signatures cannot
// expose asynchronous errors or field presence; use Subscribe for state tracking.
func subscribeStream[T any](ctx context.Context, s *Session, endpoint string) (<-chan T, error) {
	stream, err := Subscribe[T](ctx, s, endpoint)
	if err != nil {
		return nil, err
	}
	capacity := s.config.SubscribeBuffer
	if capacity < 1 {
		capacity = 4
	}
	values := make(chan T, capacity)
	go func() {
		defer close(values)
		defer stream.Close()
		for frame := range stream.Frames {
			for _, item := range frame.Items {
				select {
				case values <- item.Value:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return values, nil
}
