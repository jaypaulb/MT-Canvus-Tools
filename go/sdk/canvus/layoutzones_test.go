package canvus

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLayoutZoneTransport(t *testing.T) {
	mode := LayoutGrid
	order := []string{}
	response := `{"id":"z","widget_type":"LayoutZone","layout_mode":"grid","slot_order":[],"main_promoted":false,"gap_fraction":0.015,"layout_version":1,"size":{"width":960,"height":540}}`
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		require.Contains(t, []string{"/api/v1/canvases/c/layout-zones", "/api/v1/canvases/c/layout-zones/z"}, r.URL.Path)
		if r.Method == "POST" || r.Method == "PATCH" {
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			require.Equal(t, "grid", body["layout_mode"])
			require.Equal(t, []any{}, body["slot_order"])
		}
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "DELETE" {
			w.WriteHeader(204)
			return
		}
		if r.Method == "GET" && r.URL.Path == "/api/v1/canvases/c/layout-zones" {
			response = "[" + response + "]"
		}
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()
	cfg := DefaultSessionConfig()
	cfg.BaseURL = server.URL + "/api/v1"
	cfg.MaxRetries = 0
	s := NewSession(cfg)
	ctx := context.Background()
	z, err := s.CreateLayoutZone(ctx, "c", LayoutZoneRequest{LayoutMode: &mode, SlotOrder: &order})
	require.NoError(t, err)
	require.Equal(t, LayoutGrid, z.LayoutMode)
	require.NotNil(t, z.SlotOrder)
	z, err = s.UpdateWidget(ctx, "c", "z", map[string]any{"widget_type": "LayoutZone", "layout_mode": "grid", "slot_order": []string{}})
	require.NoError(t, err)
	require.Equal(t, 1, z.LayoutVersion)
	_, err = s.GetLayoutZone(ctx, "c", "z")
	require.NoError(t, err)
	require.NoError(t, s.DeleteWidget(ctx, "c", "z", "LayoutZone"))
	zones, err := s.ListLayoutZones(ctx, "c")
	require.NoError(t, err)
	require.Len(t, zones, 1)
	require.Equal(t, []string{"POST", "PATCH", "GET", "DELETE", "GET"}, methods)
}
