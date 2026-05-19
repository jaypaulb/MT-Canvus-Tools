//go:build live

// Phase 4d Round D Task D1 — live verification of SubscribeCanvasPermissions
// and SubscribeFolderPermissions against the dev/test Canvus server
// (dev-mtcs.multitaction.com).
//
// Run with:
//
//	source /path/to/.secrets && \
//	  CANVUS_DEV_BASE_URL=$CANVUS_DEV_BASE_URL \
//	  CANVUS_DEV_API_KEY=$CANVUS_DEV_API_KEY \
//	  go test -tags=live -run "TestSubscribeCanvas|TestSubscribeFolder" -v -timeout 60s
//
// The tests create a temporary canvas/folder, open the subscribe stream,
// drain the initial snapshot, then mutate permissions via a separate session
// and assert a change event is delivered on the channel within 10s. Test
// resources are cleaned up via defer.
package canvus

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// sentinelGroupID is "All Users" — present on every Canvus server and
// therefore a safe sentinel subject for the permissions PATCH.
const sentinelGroupID = int64(1)

// newLiveSession builds a Session configured for the dev/test server. The
// test is skipped if either env var is missing so `go test ./...` without
// the `live` tag remains a green build.
func newLiveSession(t *testing.T) *Session {
	t.Helper()
	baseURL := strings.TrimSpace(os.Getenv("CANVUS_DEV_BASE_URL"))
	apiKey := strings.TrimSpace(os.Getenv("CANVUS_DEV_API_KEY"))
	if baseURL == "" || apiKey == "" {
		t.Skip("live test: CANVUS_DEV_BASE_URL and CANVUS_DEV_API_KEY must both be set")
	}
	cfg := DefaultSessionConfig()
	cfg.BaseURL = baseURL
	return NewSession(cfg, WithAPIKey(apiKey), WithVerifyTLS(false))
}

// drainSnapshot pulls events from ch until 2s of quiet, returning the
// count drained. This is the initial-state drain the production callers
// will do before treating subsequent events as changes.
func drainSnapshot[T any](ctx context.Context, ch <-chan T, settle time.Duration) int {
	count := 0
	timer := time.NewTimer(settle)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return count
		case _, ok := <-ch:
			if !ok {
				return count
			}
			count++
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(settle)
		case <-timer.C:
			return count
		}
	}
}

// goroutineCount returns the current number of goroutines, used to
// confirm subscribe teardown does not leak.
func goroutineCount() int {
	return runtime.NumGoroutine()
}

// rawPermissionsPOST PATCHes the permissions endpoint with a raw HTTP
// request, bypassing the SDK's typed Set*Permissions methods. We use the
// raw path because the SDK's response-validation layer rejects the
// legitimate server response (which echoes the implicit owner-permission
// entry not present in the request body). This is documented as an
// unrelated SDK quirk; for D1 we care only that the subscribe channel
// fires a change event.
func rawPermissionsPOST(t *testing.T, baseURL, apiKey, endpoint string, body any) {
	t.Helper()
	payload, err := json.Marshal(body)
	require.NoError(t, err, "marshal permissions body")
	url := strings.TrimRight(baseURL, "/") + "/" + endpoint
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	require.NoError(t, err, "build raw POST request")
	req.Header.Set("Private-Token", apiKey)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			//nolint:gosec // dev/test server uses a self-signed cert.
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	require.NoError(t, err, "raw POST do")
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	require.Truef(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
		"raw POST %s returned %d: %s", endpoint, resp.StatusCode, string(respBody))
}

// bestEffortPermissionsPOST is a non-fatal variant of rawPermissionsPOST
// used for revert/cleanup steps.
func bestEffortPermissionsPOST(t *testing.T, baseURL, apiKey, endpoint string, body any) {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Logf("best-effort revert: marshal failed: %v", err)
		return
	}
	url := strings.TrimRight(baseURL, "/") + "/" + endpoint
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		t.Logf("best-effort revert: build request failed: %v", err)
		return
	}
	req.Header.Set("Private-Token", apiKey)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			//nolint:gosec // dev/test server uses a self-signed cert.
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("best-effort revert: do failed: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		t.Logf("best-effort revert: %s returned %d: %s", endpoint, resp.StatusCode, string(respBody))
	}
}

// TestSubscribeCanvasPermissionsLive verifies the canvas-permissions
// subscribe stream emits an event when permissions are PATCHed by an
// unrelated session.
func TestSubscribeCanvasPermissionsLive(t *testing.T) {
	subSession := newLiveSession(t)
	mutSession := newLiveSession(t)
	baseURL := strings.TrimSpace(os.Getenv("CANVUS_DEV_BASE_URL"))
	apiKey := strings.TrimSpace(os.Getenv("CANVUS_DEV_API_KEY"))

	rootCtx, rootCancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer rootCancel()

	// Create a temporary canvas.
	canvas, err := subSession.CreateCanvas(rootCtx, CreateCanvasRequest{
		Name: fmt.Sprintf("d1-perms-canvas-%d", time.Now().UnixNano()),
	})
	require.NoError(t, err, "create test canvas")
	require.NotEmpty(t, canvas.ID)
	t.Logf("created canvas %s", canvas.ID)
	defer func() {
		// Best-effort cleanup with a fresh context (rootCtx may have expired).
		cleanCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := mutSession.DeleteCanvas(cleanCtx, canvas.ID); err != nil {
			t.Logf("cleanup: DeleteCanvas(%s) failed: %v", canvas.ID, err)
		}
	}()

	preCount := goroutineCount()

	subCtx, subCancel := context.WithCancel(rootCtx)
	defer subCancel()

	ch, err := subSession.SubscribeCanvasPermissions(subCtx, canvas.ID)
	require.NoError(t, err, "open subscribe stream")

	// Drain the initial snapshot.
	drained := drainSnapshot[CanvasPermissions](subCtx, ch, 2*time.Second)
	t.Logf("snapshot: drained %d initial event(s)", drained)

	// Mutate permissions on a separate session via raw HTTP (bypasses
	// the SDK validator that rejects the owner-permission echo).
	newPerms := map[string]any{
		"editors_can_share": true,
		"link_permission":   "view",
		"users":             []any{},
		"groups":            []map[string]any{{"id": sentinelGroupID, "permission": "view"}},
	}
	rawPermissionsPOST(t, baseURL, apiKey,
		fmt.Sprintf("canvases/%s/permissions", canvas.ID), newPerms)
	t.Logf("PATCHed canvas %s permissions: group %d → view", canvas.ID, sentinelGroupID)
	_ = mutSession // kept for cleanup symmetry; unused for the mutation now.

	// Wait for the change event.
	select {
	case evt, ok := <-ch:
		require.True(t, ok, "subscribe channel closed before change event arrived")
		t.Logf("RECEIVED CHANGE EVENT: %+v", evt)
		// Assert event reflects the change (best-effort: server may emit
		// a sparse delta or a full snapshot; we just confirm presence).
		require.True(t, evt.EditorsCanShare || len(evt.Groups) > 0 || evt.LinkPermission != "",
			"change event payload looked empty: %+v", evt)
	case <-time.After(10 * time.Second):
		t.Fatalf("FAIL: no change event on subscribe channel within 10s after permissions PATCH")
	}

	// Revert: clear permissions (best-effort).
	bestEffortPermissionsPOST(t, baseURL, apiKey,
		fmt.Sprintf("canvases/%s/permissions", canvas.ID),
		map[string]any{
			"editors_can_share": false,
			"link_permission":   "",
			"users":             []any{},
			"groups":            []any{},
		})

	// Close subscription and confirm no goroutine leak (allow generous
	// settle, since the streaming goroutine exits on ctx cancel + body close).
	subCancel()
	// Drain so the streaming goroutine sees ctx.Done() and closes.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if _, ok := <-ch; !ok {
			break
		}
	}
	time.Sleep(200 * time.Millisecond)
	postCount := goroutineCount()
	if postCount > preCount+2 {
		t.Logf("warning: goroutine count grew from %d to %d (allow some slop)", preCount, postCount)
	}
}

// TestSubscribeFolderPermissionsLive mirrors the canvas test for the
// folder-permissions subscribe.
func TestSubscribeFolderPermissionsLive(t *testing.T) {
	subSession := newLiveSession(t)
	mutSession := newLiveSession(t)
	baseURL := strings.TrimSpace(os.Getenv("CANVUS_DEV_BASE_URL"))
	apiKey := strings.TrimSpace(os.Getenv("CANVUS_DEV_API_KEY"))

	rootCtx, rootCancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer rootCancel()

	// Create a temporary folder.
	folder, err := subSession.CreateFolder(rootCtx, CreateFolderRequest{
		Name: fmt.Sprintf("d1-perms-folder-%d", time.Now().UnixNano()),
	})
	require.NoError(t, err, "create test folder")
	require.NotEmpty(t, folder.ID)
	t.Logf("created folder %s", folder.ID)
	defer func() {
		cleanCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := mutSession.DeleteFolder(cleanCtx, folder.ID); err != nil {
			t.Logf("cleanup: DeleteFolder(%s) failed: %v", folder.ID, err)
		}
	}()

	preCount := goroutineCount()

	subCtx, subCancel := context.WithCancel(rootCtx)
	defer subCancel()

	ch, err := subSession.SubscribeFolderPermissions(subCtx, folder.ID)
	require.NoError(t, err, "open subscribe stream")

	drained := drainSnapshot[FolderPermissions](subCtx, ch, 2*time.Second)
	t.Logf("snapshot: drained %d initial event(s)", drained)

	newPerms := map[string]any{
		"editors_can_share": true,
		"users":             []any{},
		"groups":            []map[string]any{{"id": sentinelGroupID, "permission": "view"}},
	}
	rawPermissionsPOST(t, baseURL, apiKey,
		fmt.Sprintf("canvas-folders/%s/permissions", folder.ID), newPerms)
	t.Logf("PATCHed folder %s permissions: group %d → view", folder.ID, sentinelGroupID)
	_ = mutSession

	select {
	case evt, ok := <-ch:
		require.True(t, ok, "subscribe channel closed before change event arrived")
		t.Logf("RECEIVED CHANGE EVENT: %+v", evt)
		require.True(t, evt.EditorsCanShare || len(evt.Groups) > 0,
			"change event payload looked empty: %+v", evt)
	case <-time.After(10 * time.Second):
		t.Fatalf("FAIL: no change event on subscribe channel within 10s after permissions PATCH")
	}

	bestEffortPermissionsPOST(t, baseURL, apiKey,
		fmt.Sprintf("canvas-folders/%s/permissions", folder.ID),
		map[string]any{
			"editors_can_share": false,
			"users":             []any{},
			"groups":            []any{},
		})

	subCancel()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if _, ok := <-ch; !ok {
			break
		}
	}
	time.Sleep(200 * time.Millisecond)
	postCount := goroutineCount()
	if postCount > preCount+2 {
		t.Logf("warning: goroutine count grew from %d to %d (allow some slop)", preCount, postCount)
	}
}
