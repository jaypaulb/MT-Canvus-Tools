package webui

import (
	"encoding/json"
	"fmt"
	"net/http"

	webuiatoms "github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/webui"
)

// PagesHandler handles pages management API endpoints.
type PagesHandler struct {
	apiClient     *webuiatoms.APIClient
	canvasService *CanvasService
}

// NewPagesHandler creates a new pages handler.
func NewPagesHandler(apiClient *webuiatoms.APIClient, canvasService *CanvasService) *PagesHandler {
	return &PagesHandler{
		apiClient:     apiClient,
		canvasService: canvasService,
	}
}

// HandleList handles GET /api/pages - List all pages.
func (h *PagesHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	canvasID := h.canvasService.GetCanvasID()
	if canvasID == "" {
		sendErrorResponse(w, "Canvas not available", http.StatusServiceUnavailable)
		return
	}

	// Fetch anchors (pages) from Canvus API
	// Note: "Pages" in WebUI are "Anchors" in Canvus API
	endpoint := fmt.Sprintf("/api/v1/canvases/%s/anchors", canvasID)
	data, err := h.apiClient.Get(endpoint)
	if err != nil {
		sendErrorResponse(w, fmt.Sprintf("Failed to fetch pages: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// HandleCreate handles POST /api/pages - Create a new anchor (page).
// Note: Anchors are created as widgets with widget_type="Anchor" in the Canvus API.
func (h *PagesHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		sendErrorResponse(w, "Page name is required", http.StatusBadRequest)
		return
	}

	canvasID := h.canvasService.GetCanvasID()
	if canvasID == "" {
		sendErrorResponse(w, "Canvas not available", http.StatusServiceUnavailable)
		return
	}

	// Create anchor (page) via Canvus API
	// Note: Use /anchors endpoint, not /widgets (widgets is read-only)
	payload := map[string]interface{}{
		"anchor_name": req.Name,
	}
	endpoint := fmt.Sprintf("/api/v1/canvases/%s/anchors", canvasID)
	data, err := h.apiClient.Post(endpoint, payload)
	if err != nil {
		sendErrorResponse(w, fmt.Sprintf("Failed to create page: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

// HandleGetZones handles GET /get-zones - Returns zones (anchors) in format expected by pages.js
func (h *PagesHandler) HandleGetZones(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	canvasID := h.canvasService.GetCanvasID()
	if canvasID == "" {
		// Return empty zones array with success=true when canvas not available (graceful degradation)
		response := map[string]interface{}{
			"success": true,
			"zones":   []interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Fetch anchors (zones) from Canvus API
	endpoint := fmt.Sprintf("/api/v1/canvases/%s/anchors", canvasID)
	data, err := h.apiClient.Get(endpoint)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to fetch zones: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse anchors array
	var zones []map[string]interface{}
	if err := json.Unmarshal(data, &zones); err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to parse zones: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Return in expected format
	response := map[string]interface{}{
		"success": true,
		"zones":   zones,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
