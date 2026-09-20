package canvus_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func TestWorkspaceSelectionNeverDefaultsOrChoosesFirstMatch(t *testing.T) {
	name := "wall"
	for _, tt := range []struct {
		name, body string
		selector   canvus.WorkspaceSelector
	}{
		{"unspecified", `[]`, canvus.WorkspaceSelector{}},
		{"ambiguous", `[{"index":0,"workspace_name":"wall"},{"index":1,"workspace_name":"wall"}]`, canvus.WorkspaceSelector{Name: &name}},
		{"missing_index", `[{"workspace_name":"wall"}]`, canvus.WorkspaceSelector{Name: &name}},
		{"null_index", `[{"index":null,"workspace_name":"wall"}]`, canvus.WorkspaceSelector{Name: &name}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var writes atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPatch {
					writes.Add(1)
					fmt.Fprint(w, `{"index":0}`)
					return
				}
				fmt.Fprint(w, tt.body)
			}))
			defer srv.Close()
			_, err := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}).UpdateWorkspace(context.Background(), "c", tt.selector, map[string]any{"pinned": true})
			require.Error(t, err)
			require.Zero(t, writes.Load())
		})
	}
}

func TestWorkspaceGetRejectsChangedOrMissingIdentity(t *testing.T) {
	for _, body := range []string{`{"index":1}`, `{"canvas_id":"c"}`} {
		t.Run(body, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) }))
			defer srv.Close()
			idx := 0
			_, err := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}).GetWorkspace(context.Background(), "client", canvus.WorkspaceSelector{Index: &idx})
			require.Error(t, err)
		})
	}
}

func TestClientAcceptsNumericAndStringAssociationWithoutInferringOperator(t *testing.T) {
	for _, raw := range []string{`7`, `"7"`} {
		var c canvus.ClientInfo
		require.NoError(t, json.Unmarshal([]byte(`{"id":"client","user_id":`+raw+`}`), &c))
		require.Equal(t, "7", c.UserID)
	}
}

func TestWorkspaceMissingAndZeroRemainDistinct(t *testing.T) {
	var missing, zero canvus.Workspace
	require.NoError(t, json.Unmarshal([]byte(`{"user":"actor@example.invalid"}`), &missing))
	require.NoError(t, json.Unmarshal([]byte(`{"index":0,"user":"actor@example.invalid"}`), &zero))
	require.NotEqual(t, missing, zero)
}
