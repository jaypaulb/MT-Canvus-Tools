package canvus_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

type identityStore struct {
	mu       sync.Mutex
	token    string
	fail     bool
	readFail bool
}

func (s *identityStore) GetToken() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.readFail {
		return "", errors.New("synthetic store read failure")
	}
	return s.token, nil
}
func (s *identityStore) StoreToken(token string, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fail {
		return errors.New("synthetic storage failure")
	}
	s.token = token
	return nil
}
func (s *identityStore) ClearToken() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = ""
	return nil
}

func TestTokenStoreReadFailureCannotFallBackToServiceAuthority(t *testing.T) {
	requests := make(chan struct{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests <- struct{}{}; fmt.Fprint(w, `{"id":"n"}`) }))
	defer srv.Close()
	s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL, APIKey: "synthetic-service", TokenStore: &identityStore{readFail: true}})
	_, err := s.GetNote(context.Background(), "c", "n")
	require.ErrorIs(t, err, canvus.ErrTokenPersistence)
	select {
	case <-requests:
		t.Fatal("sent request despite unknown stored actor")
	default:
	}
}

func TestAuthenticationPersistsWithoutRestoringOldAuthorityOnStoreFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			headers := make(chan []string, 1)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/users/login" {
					fmt.Fprint(w, `{"token":"synthetic-actor","user":{"id":7}}`)
					return
				}
				headers <- r.Header.Values("Private-Token")
				fmt.Fprint(w, `{"id":"n"}`)
			}))
			defer srv.Close()
			store := &identityStore{fail: fail}
			cfg := &canvus.SessionConfig{BaseURL: srv.URL, TokenStore: store, APIKey: "synthetic-service"}
			s := canvus.NewSession(cfg)
			err := s.Login(context.Background(), "actor@example.invalid", "synthetic")
			if fail {
				require.ErrorIs(t, err, canvus.ErrTokenPersistence)
				require.False(t, canvus.IsRetryableError(err))
			} else {
				require.NoError(t, err)
				s = canvus.NewSession(cfg)
			}
			_, err = s.GetNote(context.Background(), "c", "n")
			require.NoError(t, err)
			require.Equal(t, []string{"synthetic-actor"}, <-headers)
		})
	}
}

func TestWaitingForAuthenticationHonorsCancellation(t *testing.T) {
	started := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		close(started)
		<-r.Context().Done()
	}))
	defer srv.Close()
	s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}, canvus.WithToken("synthetic"))
	first, cancelFirst := context.WithTimeout(context.Background(), time.Second)
	defer cancelFirst()
	finished := make(chan error, 1)
	go func() { finished <- s.Login(first, "actor@example.invalid", "synthetic") }()
	<-started
	second, cancelSecond := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancelSecond()
	start := time.Now()
	_, err := s.GetCurrentUser(second)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Less(t, time.Since(start), 300*time.Millisecond)
	cancelFirst()
	<-finished
}

func TestPasswordHelperUsesObservedWireFields(t *testing.T) {
	payloads := make(chan map[string]string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		payloads <- body
		w.WriteHeader(204)
	}))
	defer srv.Close()
	s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL})
	require.NoError(t, s.SetUserPassword(context.Background(), 7, "synthetic-new"))
	require.Equal(t, map[string]string{"new_password": "synthetic-new"}, <-payloads)
	require.NoError(t, s.ChangeUserPassword(context.Background(), 7, "synthetic-current", "synthetic-next"))
	require.Equal(t, map[string]string{"current_password": "synthetic-current", "new_password": "synthetic-next"}, <-payloads)
}

func TestSAMLInstallsAuthenticatedIdentity(t *testing.T) {
	headers := make(chan []string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/login/saml" {
			fmt.Fprint(w, `{"user":{"id":7,"email":"actor@example.invalid"},"token":"synthetic-saml"}`)
			return
		}
		headers <- r.Header.Values("Private-Token")
		fmt.Fprint(w, `{"id":"n"}`)
	}))
	defer srv.Close()
	s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL})
	require.NoError(t, s.SamlLogin(context.Background(), canvus.SamlLoginRequest{ResponseXML: "synthetic-assertion"}))
	require.EqualValues(t, 7, s.UserID())
	_, err := s.GetNote(context.Background(), "c", "n")
	require.NoError(t, err)
	require.Equal(t, []string{"synthetic-saml"}, <-headers)
}

func TestTokenIdentityUsesVerifiedExchangeNotCurrentAlias(t *testing.T) {
	payloads := make(chan map[string]any, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/users/login" {
			w.WriteHeader(400)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		payloads <- body
		fmt.Fprint(w, `{"user":{"id":7,"email":"actor@example.invalid"},"token":"synthetic-session"}`)
	}))
	defer srv.Close()
	s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}, canvus.WithToken("synthetic-api-token"))
	user, err := s.GetCurrentUser(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, 7, user.ID)
	require.EqualValues(t, 7, s.UserID())
	require.Equal(t, "synthetic-api-token", (<-payloads)["token"])
}
