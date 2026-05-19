package webui

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type rcuConfig struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
	Timeout int  `json:"timeout"`
}

type rcuStatus struct {
	Connected  bool        `json:"connected"`
	LastUpdate any `json:"last_update"`
}

// RCUHandler serves the WebUI's own admin RCU surface.
// /api/v1/canvases/{id}/rcu/* are NOT Canvus-server endpoints — they are this
// WebUI server's admin API. State is in-memory and resets on server restart.
type RCUHandler struct {
	mu     sync.RWMutex
	config rcuConfig
	status rcuStatus
}

func NewRCUHandler() *RCUHandler {
	return &RCUHandler{
		config: rcuConfig{Enabled: false, Port: 8080, Timeout: 30},
		status: rcuStatus{Connected: false},
	}
}

// HandleConfig handles GET and POST for /api/v1/canvases/{id}/rcu/config.
func (h *RCUHandler) HandleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	switch r.Method {
	case http.MethodGet:
		h.mu.RLock()
		defer h.mu.RUnlock()
		json.NewEncoder(w).Encode(h.config)

	case http.MethodPost:
		var req rcuConfig
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		h.mu.Lock()
		h.config = req
		cfg := h.config
		h.mu.Unlock()
		json.NewEncoder(w).Encode(cfg)

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// HandleStatus handles GET /api/v1/canvases/{id}/rcu/status.
func (h *RCUHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	h.mu.RLock()
	defer h.mu.RUnlock()
	json.NewEncoder(w).Encode(h.status)
}

// HandleTest handles POST /api/v1/canvases/{id}/rcu/test.
func (h *RCUHandler) HandleTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	h.mu.Lock()
	h.status.Connected = true
	h.status.LastUpdate = time.Now()
	h.mu.Unlock()
	json.NewEncoder(w).Encode(map[string]any{"success": true})
}
