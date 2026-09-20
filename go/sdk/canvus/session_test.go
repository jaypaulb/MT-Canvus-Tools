package canvus

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSession_AppliesDefaults(t *testing.T) {
	cfg := &SessionConfig{BaseURL: "https://example.invalid/api/v1"}
	s := NewSession(cfg)
	require.NotNil(t, s)
	assert.Equal(t, "https://example.invalid/api/v1", s.BaseURL)
	assert.NotNil(t, s.HTTPClient)
	assert.Zero(t, s.config.MaxRetries, "zero-value budget explicitly disables retries")
	assert.Equal(t, 30*time.Second, s.config.RequestTimeout)
}

func TestNewSession_WithToken_InstallsAuthenticator(t *testing.T) {
	cfg := &SessionConfig{BaseURL: "https://example.invalid/api/v1"}
	s := NewSession(cfg, WithToken("abc"))
	require.NotNil(t, s.authenticator)
	tok, ok := s.authenticator.(*TokenAuthenticator)
	require.True(t, ok)
	assert.Equal(t, "abc", tok.Token)
}

func TestWithAPIKey_AppliesSingleHeader(t *testing.T) {
	for _, key := range []string{"synthetic-key", ""} {
		t.Run(key, func(t *testing.T) {
			var headers []string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				headers = r.Header.Values("Private-Token")
				_, _ = w.Write([]byte(`{"id":"n"}`))
			}))
			defer srv.Close()
			s := NewSession(&SessionConfig{BaseURL: srv.URL}, WithAPIKey(key))
			_, err := s.GetNote(context.Background(), "c", "n")
			require.NoError(t, err)
			if key == "" {
				assert.Empty(t, headers)
			} else {
				assert.Equal(t, []string{key}, headers)
			}
		})
	}
}

func TestAPIError_ErrorIncludesFields(t *testing.T) {
	e := &APIError{
		StatusCode: 404,
		Code:       "not_found",
		Message:    "no such canvas",
		RequestID:  "req-123",
	}
	s := e.Error()
	assert.Contains(t, s, "404")
	assert.Contains(t, s, "not_found")
	assert.Contains(t, s, "no such canvas")
	assert.Contains(t, s, "req-123")
}

func TestAPIError_IsSentinelMatch(t *testing.T) {
	e := &APIError{StatusCode: 404}
	assert.True(t, errors.Is(e, ErrNotFound))
	assert.False(t, errors.Is(e, ErrConflict))

	e2 := &APIError{StatusCode: 401}
	assert.True(t, errors.Is(e2, ErrUnauthorized))
}

func TestColor_NormalizeColor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"already valid", "FFFFFFFF", "FFFFFFFF", false},
		{"lowercase", "ffffffff", "FFFFFFFF", false},
		{"6char rgb", "FF0000", "FF0000FF", false},
		{"6char with hash", "#FF0000", "FF0000FF", false},
		{"bad length", "FF", "", true},
		{"bad chars", "ZZZZZZZZ", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeColor(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFilter_MatchWildcards(t *testing.T) {
	obj := map[string]any{
		"id":   "abc-123",
		"name": "Hello World",
		"location": map[string]any{
			"x": 10.0, "y": 20.0,
		},
	}
	tests := []struct {
		name string
		f    Filter
		want bool
	}{
		{"exact match", Filter{Criteria: map[string]any{"id": "abc-123"}}, true},
		{"wildcard star", Filter{Criteria: map[string]any{"id": "*"}}, true},
		{"prefix", Filter{Criteria: map[string]any{"name": "Hello*"}}, true},
		{"suffix", Filter{Criteria: map[string]any{"name": "*World"}}, true},
		{"contains", Filter{Criteria: map[string]any{"name": "*lo Wo*"}}, true},
		{"miss", Filter{Criteria: map[string]any{"id": "no"}}, false},
		{"jsonpath nested", Filter{Criteria: map[string]any{"$.location.x": 10.0}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.f.Match(obj))
		})
	}
}

func TestDoRequest_GETandPATCH_Smoke(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/canvases"):
			_, _ = w.Write([]byte(`[{"id":"c1","name":"hello"}]`))
		case r.Method == http.MethodPatch && strings.HasSuffix(r.URL.Path, "/canvases/c1"):
			var req map[string]any
			_ = json.NewDecoder(r.Body).Decode(&req)
			resp := map[string]any{"id": "c1", "name": req["name"]}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	cfg := DefaultSessionConfig()
	cfg.BaseURL = srv.URL + "/api/v1"
	s := NewSession(cfg)

	canvases, err := s.ListCanvases(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, canvases, 1)
	assert.Equal(t, "c1", canvases[0].ID)

	updated, err := s.UpdateCanvas(context.Background(), "c1", map[string]any{"name": "renamed"})
	require.NoError(t, err)
	assert.Equal(t, "renamed", updated.Name)
}

func TestCloneWidget_BuildsRequest(t *testing.T) {
	var got struct {
		path    string
		body    map[string]any
		gotBody bool
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.path = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&got.body)
		got.gotBody = true
		_, _ = w.Write([]byte(`{"id":"new-widget","widget_type":"note"}`))
	}))
	defer srv.Close()

	cfg := DefaultSessionConfig()
	cfg.BaseURL = srv.URL + "/api/v1"
	s := NewSession(cfg)

	loc := &Point{X: 5, Y: 6}
	w, err := s.CloneWidget(context.Background(), "destC", "srcC", "srcW", "notes", loc)
	require.NoError(t, err)
	assert.Equal(t, "new-widget", w.ID)
	assert.True(t, got.gotBody)
	assert.Contains(t, got.path, "/canvases/destC/notes")
	assert.Equal(t, "srcC", got.body["source_canvas_id"])
	assert.Equal(t, "srcW", got.body["source_widget_id"])
	assert.NotNil(t, got.body["location"])
}

func TestCreateWidget_RejectsIPVideoAndRDP(t *testing.T) {
	cfg := DefaultSessionConfig()
	cfg.BaseURL = "https://example.invalid/api/v1"
	s := NewSession(cfg)

	tests := []string{"ip_video", "rdp_connection", "IP_VIDEO", "rdp-connection"}
	for _, wt := range tests {
		t.Run(wt, func(t *testing.T) {
			_, err := s.CreateWidget(context.Background(), "c", map[string]any{"widget_type": wt})
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrWidgetTypeNotCreatable)
		})
	}
}

func TestUpdateTable_StripsGridSize(t *testing.T) {
	var got struct {
		body map[string]any
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got.body)
		_, _ = w.Write([]byte(`{"id":"t1","widget_type":"table","title":"x"}`))
	}))
	defer srv.Close()

	cfg := DefaultSessionConfig()
	cfg.BaseURL = srv.URL + "/api/v1"
	s := NewSession(cfg)

	_, err := s.UpdateTable(context.Background(), "c", "t1", map[string]any{
		"title":     "x",
		"grid_size": map[string]any{"columns": 5, "rows": 5},
	})
	require.NoError(t, err)
	_, has := got.body["grid_size"]
	assert.False(t, has, "grid_size should have been stripped before sending")
	assert.Equal(t, "x", got.body["title"])
}

func TestGeometry_ContainsTouches(t *testing.T) {
	a := Rectangle{X: 0, Y: 0, Width: 10, Height: 10}
	b := Rectangle{X: 2, Y: 2, Width: 5, Height: 5}
	c := Rectangle{X: 9, Y: 9, Width: 5, Height: 5}
	assert.True(t, Contains(a, b))
	assert.False(t, Contains(a, c))
	assert.True(t, Touches(a, c))
}
