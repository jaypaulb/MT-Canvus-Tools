package canvus_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func TestCameraUsesKnownZoomThenPanProtocol(t *testing.T) {
	var mu sync.Mutex
	var writes []canvus.Rectangle
	current := canvus.Rectangle{Width: 1000, Height: 600}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.Method == http.MethodPatch {
			var body struct {
				View canvus.Rectangle `json:"view_rectangle"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			writes = append(writes, body.View)
			current = body.View
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"index": 0, "canvas_id": "canvas", "server_id": "server", "workspace_state": "open", "size": canvus.Size{Width: 1000, Height: 600}, "view_rectangle": current})
	}))
	defer srv.Close()
	idx := 0
	x, y, width, height := 100.0, 200.0, 400.0, 200.0
	canvasID := "canvas"
	s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL})
	require.NoError(t, canvus.SetWorkspaceViewport(context.Background(), s, s, "client", canvus.WorkspaceSelector{Index: &idx}, canvus.SetViewportOptions{CanvasID: &canvasID, X: &x, Y: &y, Width: &width, Height: &height}))
	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, []canvus.Rectangle{{Width: 2500, Height: 1500}, {X: -250, Y: -450, Width: 2500, Height: 1500}}, writes)
}

func TestCameraStopsAfterTargetChangeAndPreservesNative500(t *testing.T) {
	for _, name := range []string{"changed_after_zoom", "native_500"} {
		t.Run(name, func(t *testing.T) {
			var mu sync.Mutex
			writes := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				if r.Method == http.MethodPatch {
					writes++
					if name == "native_500" {
						w.WriteHeader(500)
						fmt.Fprint(w, `{"code":"not_owner","status_code":403}`)
						return
					}
				}
				canvas := "canvas"
				if writes > 0 && name == "changed_after_zoom" {
					canvas = "other"
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"index": 0, "canvas_id": canvas, "server_id": "server", "workspace_state": "open", "size": canvus.Size{Width: 1000, Height: 600}})
			}))
			defer srv.Close()
			idx := 0
			err := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}).FrameWorkspaceRegion(context.Background(), "client", canvus.WorkspaceSelector{Index: &idx}, "canvas", canvus.Rectangle{Width: 400, Height: 200})
			var partial *canvus.CameraUpdateError
			require.ErrorAs(t, err, &partial)
			require.False(t, canvus.IsRetryableError(err), "camera operations must not be replayed as safe reads")
			if name == "changed_after_zoom" {
				require.ErrorIs(t, err, canvus.ErrWorkspaceChanged)
				require.True(t, partial.ZoomAcknowledged)
			} else {
				var apiErr *canvus.APIError
				require.ErrorAs(t, err, &apiErr)
				require.Equal(t, 500, apiErr.StatusCode)
				require.False(t, errors.Is(err, canvus.ErrForbidden))
			}
			mu.Lock()
			defer mu.Unlock()
			require.Equal(t, 1, writes)
		})
	}
}

func TestCameraKeepsOriginalActorAcrossExplicitLogin(t *testing.T) {
	var mu sync.Mutex
	var headers [][]string
	current := canvus.Rectangle{Width: 1000, Height: 600}
	zoomed := make(chan struct{})
	actorInstalled := make(chan struct{})
	var readyOnce sync.Once
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/login" {
			fmt.Fprint(w, `{"token":"synthetic-B","user":{"id":8}}`)
			return
		}
		mu.Lock()
		hasZoom := len(headers) > 0
		mu.Unlock()
		if r.Method == http.MethodGet && hasZoom {
			<-actorInstalled
		}
		mu.Lock()
		defer mu.Unlock()
		if r.Method == http.MethodPatch {
			var body struct {
				View canvus.Rectangle `json:"view_rectangle"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			current = body.View
			headers = append(headers, r.Header.Values("Private-Token"))
			if len(headers) == 1 {
				close(zoomed)
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"index": 0, "canvas_id": "canvas", "server_id": "server", "workspace_state": "open", "size": canvus.Size{Width: 1000, Height: 600}, "view_rectangle": current})
	}))
	defer srv.Close()
	defer readyOnce.Do(func() { close(actorInstalled) })
	s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}, canvus.WithToken("synthetic-A"))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	idx := 0
	finished := make(chan error, 1)
	go func() {
		finished <- s.FrameWorkspaceRegion(ctx, "client", canvus.WorkspaceSelector{Index: &idx}, "canvas", canvus.Rectangle{Width: 400, Height: 200})
	}()
	select {
	case <-zoomed:
	case <-ctx.Done():
		t.Fatal("zoom not reached")
	}
	err := s.Login(ctx, "actor@example.invalid", "synthetic")
	readyOnce.Do(func() { close(actorInstalled) })
	require.NoError(t, err)
	require.NoError(t, <-finished)
	require.EqualValues(t, 8, s.UserID())
	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, [][]string{{"synthetic-A"}, {"synthetic-A"}}, headers)
}

func TestOpenCanvasCentersInCanvasSpaceAndRejectsMissingSize(t *testing.T) {
	for _, missing := range []bool{false, true} {
		t.Run(fmt.Sprint(missing), func(t *testing.T) {
			var mu sync.Mutex
			var writes []canvus.Rectangle
			current := canvus.Rectangle{X: -1280, Y: -720, Width: 512, Height: 288}
			var size *canvus.Size
			if !missing {
				size = &canvus.Size{Width: 1280, Height: 720}
			}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				if r.Method == http.MethodPatch {
					var body struct {
						View canvus.Rectangle `json:"view_rectangle"`
					}
					_ = json.NewDecoder(r.Body).Decode(&body)
					current = body.View
					writes = append(writes, current)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"index": 0, "canvas_id": "canvas", "server_id": "server", "workspace_state": "open", "size": size, "view_rectangle": current})
			}))
			defer srv.Close()
			idx := 0
			x, y := 1000.0, 1000.0
			var err error
			require.NotPanics(t, func() {
				err = canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}).OpenCanvasOnWorkspace(context.Background(), "client", canvus.WorkspaceSelector{Index: &idx}, canvus.OpenCanvasOptions{CanvasID: "canvas", CenterX: &x, CenterY: &y})
			})
			mu.Lock()
			defer mu.Unlock()
			if missing {
				require.ErrorIs(t, err, canvus.ErrWorkspaceMetadata)
				require.Empty(t, writes)
			} else {
				require.NoError(t, err)
				require.Equal(t, []canvus.Rectangle{{Width: 512, Height: 288}, {X: 240, Y: -40, Width: 512, Height: 288}}, writes)
			}
		})
	}
}

func TestAcceptedOpenCannotBeReplayedAfterPollingFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(204)
			return
		}
		w.WriteHeader(500)
	}))
	defer srv.Close()
	idx := 0
	err := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}).OpenCanvasOnWorkspace(context.Background(), "client", canvus.WorkspaceSelector{Index: &idx}, canvus.OpenCanvasOptions{CanvasID: "canvas"})
	require.Error(t, err)
	require.False(t, canvus.IsRetryableError(err))
	var outcome *canvus.OpenCanvasError
	require.ErrorAs(t, err, &outcome)
	require.True(t, outcome.CommandAcknowledged)
	require.Equal(t, "readiness", outcome.Stage)
}

func TestCameraPreservesAcceptedZoomWithMalformedResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			fmt.Fprint(w, `{`)
			return
		}
		fmt.Fprint(w, `{"index":0,"canvas_id":"canvas","server_id":"server","workspace_state":"open","size":{"width":1000,"height":600}}`)
	}))
	defer srv.Close()
	idx := 0
	err := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}).FrameWorkspaceRegion(context.Background(), "client", canvus.WorkspaceSelector{Index: &idx}, "canvas", canvus.Rectangle{Width: 400, Height: 200})
	var partial *canvus.CameraUpdateError
	require.ErrorAs(t, err, &partial)
	require.True(t, partial.ZoomAcknowledged)
	var accepted *canvus.AcceptedResponseError
	require.ErrorAs(t, err, &accepted)
}

func TestWidgetFramingUsesObservedSharedCanvasRoot(t *testing.T) {
	var mu sync.Mutex
	var writes []canvus.Rectangle
	view := canvus.Rectangle{Width: 600, Height: 600}
	widgets := map[string]canvus.Widget{
		"child":       {ID: "child", WidgetType: "Note", ParentID: "parent", Location: &canvus.Point{X: 20, Y: 30}, Size: &canvus.Size{Width: 90, Height: 90}, Scale: .5},
		"parent":      {ID: "parent", WidgetType: "Note", ParentID: "shared-root", Location: &canvus.Point{X: 600, Y: 100}, Size: &canvus.Size{Width: 200, Height: 120}, Scale: 2},
		"shared-root": {ID: "shared-root", WidgetType: "SharedCanvas", ParentID: "", Location: &canvus.Point{}, Size: &canvus.Size{Width: 9600, Height: 5400}, Scale: 1},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for id, widget := range widgets {
			if r.URL.Path == "/canvases/canvas-resource/widgets/"+id {
				_ = json.NewEncoder(w).Encode(widget)
				return
			}
		}
		mu.Lock()
		defer mu.Unlock()
		if r.Method == http.MethodPatch {
			var body struct {
				View canvus.Rectangle `json:"view_rectangle"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			view = body.View
			writes = append(writes, view)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"index": 0, "canvas_id": "canvas-resource", "server_id": "server", "workspace_state": "open", "size": canvus.Size{Width: 600, Height: 600}, "view_rectangle": view})
	}))
	defer srv.Close()
	s := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL})
	index := 0
	canvas, id, padding := "canvas-resource", "child", 30.0
	err := canvus.SetWorkspaceViewport(context.Background(), s, s, "client", canvus.WorkspaceSelector{Index: &index}, canvus.SetViewportOptions{CanvasID: &canvas, WidgetID: &id, NotePadding: &padding, Margin: 5})
	require.NoError(t, err)
	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, []canvus.Rectangle{{Width: 3600, Height: 3600}, {X: -4170, Y: -1290, Width: 3600, Height: 3600}}, writes)
}

func TestCameraDoesNotPanAtUnconfirmedOrClampedZoom(t *testing.T) {
	for _, missing := range []bool{false, true} {
		t.Run(fmt.Sprint(missing), func(t *testing.T) {
			var mu sync.Mutex
			writes := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				if r.Method == http.MethodPatch {
					writes++
				}
				body := map[string]any{"index": 0, "canvas_id": "canvas", "server_id": "server", "workspace_state": "open", "size": canvus.Size{Width: 1000, Height: 600}}
				if !missing {
					body["view_rectangle"] = canvus.Rectangle{Width: 1000, Height: 600}
				}
				_ = json.NewEncoder(w).Encode(body)
			}))
			defer srv.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			defer cancel()
			idx := 0
			err := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}).FrameWorkspaceRegion(ctx, "client", canvus.WorkspaceSelector{Index: &idx}, "canvas", canvus.Rectangle{Width: 400, Height: 200})
			var partial *canvus.CameraUpdateError
			require.ErrorAs(t, err, &partial)
			require.Equal(t, "settle", partial.Stage)
			require.True(t, partial.ZoomAcknowledged)
			require.False(t, canvus.IsRetryableError(err))
			mu.Lock()
			defer mu.Unlock()
			require.Equal(t, 1, writes)
		})
	}
}

func TestOpenCanvasDoesNotReportLoadingAsReady(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"index":0,"canvas_id":"canvas","server_id":"server","workspace_state":"loading"}`)
	}))
	defer srv.Close()
	idx := 0
	err := canvus.NewSession(&canvus.SessionConfig{BaseURL: srv.URL}).OpenCanvasOnWorkspace(context.Background(), "client", canvus.WorkspaceSelector{Index: &idx}, canvus.OpenCanvasOptions{CanvasID: "canvas", PollTimeout: 30 * time.Millisecond, PollInterval: 5 * time.Millisecond})
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
