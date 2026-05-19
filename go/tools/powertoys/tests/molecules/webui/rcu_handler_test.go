package webui_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	webuimol "github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/molecules/webui"
)

func TestRCUHandler_ConfigRoundTrip(t *testing.T) {
	h := webuimol.NewRCUHandler()

	// GET should return default config
	req := httptest.NewRequest("GET", "/api/v1/canvases/c1/rcu/config", nil)
	w := httptest.NewRecorder()
	h.HandleConfig(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET config returned %d", w.Code)
	}

	// POST should update config
	body := `{"enabled":true,"port":9090,"timeout":60}`
	req = httptest.NewRequest("POST", "/api/v1/canvases/c1/rcu/config", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	h.HandleConfig(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("POST config returned %d", w.Code)
	}

	// GET again should reflect posted config
	req = httptest.NewRequest("GET", "/api/v1/canvases/c1/rcu/config", nil)
	w = httptest.NewRecorder()
	h.HandleConfig(w, req)
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["port"] != float64(9090) {
		t.Errorf("port = %v, want 9090", resp["port"])
	}
}
