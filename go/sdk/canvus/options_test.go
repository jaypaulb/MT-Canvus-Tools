package canvus

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithSubscribeBuffer_PanicsOnZero(t *testing.T) {
	cfg := DefaultSessionConfig()
	assert.Panics(t, func() { WithSubscribeBuffer(0)(cfg) })
}

func TestWithSubscribeBuffer_PanicsOnNegative(t *testing.T) {
	cfg := DefaultSessionConfig()
	assert.Panics(t, func() { WithSubscribeBuffer(-1)(cfg) })
}

func TestWithVerifyTLS_PoliciesAtHTTPBoundary(t *testing.T) {
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, `{"id":"n"}`) }))
	srv.Config.ErrorLog = log.New(io.Discard, "", 0)
	srv.StartTLS()
	defer srv.Close()
	for _, tt := range []struct {
		name      string
		opts      []SessionConfigOption
		wantError bool
	}{
		{"default", nil, true},
		{"explicit_verify", []SessionConfigOption{WithVerifyTLS(true)}, true},
		{"api_key_does_not_disable_tls", []SessionConfigOption{WithAPIKey("synthetic-key")}, true},
		{"explicit_skip", []SessionConfigOption{WithVerifyTLS(false)}, false},
		{"api_key_and_skip", []SessionConfigOption{WithAPIKey("synthetic-key"), WithVerifyTLS(false)}, false},
		{"custom_secure_client_wins", []SessionConfigOption{WithHTTPClient(&http.Client{}), WithVerifyTLS(false)}, true},
		{"custom_trusted_client", []SessionConfigOption{WithHTTPClient(srv.Client())}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSession(&SessionConfig{BaseURL: srv.URL}, tt.opts...)
			_, err := s.GetNote(context.Background(), "c", "n")
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
