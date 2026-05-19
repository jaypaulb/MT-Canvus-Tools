package webui

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/logger"
	webuiatoms "github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/webui"
)

// APIRoutes handles registration of API routes for the WebUI server.
type APIRoutes struct {
	canvasService *CanvasService
	sseHandler    *SSEHandler
	apiClient     *webuiatoms.APIClient
	pagesHandler  *PagesHandler
	macrosHandler *MacrosHandler
	uploadHandler *UploadHandler
	rcuHandler    *RCUHandler
	adminHandler  *AdminHandler
}

// NewAPIRoutes creates a new API routes handler.
func NewAPIRoutes(canvasService *CanvasService, apiClient *webuiatoms.APIClient, uploadDir string) *APIRoutes {
	sseHandler := NewSSEHandler(canvasService)
	pagesHandler := NewPagesHandler(apiClient, canvasService)
	macrosHandler := NewMacrosHandler(apiClient, canvasService)
	uploadHandler := NewUploadHandler(apiClient, canvasService, uploadDir)
	rcuHandler := NewRCUHandlerWithDeps(apiClient, canvasService)
	adminHandler := NewAdminHandler(apiClient, canvasService, rcuHandler)

	return &APIRoutes{
		canvasService: canvasService,
		sseHandler:    sseHandler,
		apiClient:     apiClient,
		pagesHandler:  pagesHandler,
		macrosHandler: macrosHandler,
		uploadHandler: uploadHandler,
		rcuHandler:    rcuHandler,
		adminHandler:  adminHandler,
	}
}

// RegisterRoutes registers all API routes with the given mux.
func (ar *APIRoutes) RegisterRoutes(mux *http.ServeMux) {
	// SSE endpoint for canvas_id updates
	mux.HandleFunc("/api/subscribe-workspace", ar.sseHandler.HandleSubscribe)

	// Canvas info endpoint (current canvas_id and canvas_name)
	mux.HandleFunc("/api/canvas/info", ar.handleCanvasInfo)

	// Installation info endpoint
	mux.HandleFunc("/api/installation/info", ar.handleInstallationInfo)

	// Health check endpoint
	mux.HandleFunc("/api/health", ar.handleHealth)

	// Server info endpoint (for IP detection)
	mux.HandleFunc("/api/server-info", ar.handleServerInfo)

	// Pages endpoints
	mux.HandleFunc("/api/pages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			ar.pagesHandler.HandleList(w, r)
		} else if r.Method == http.MethodPost {
			ar.pagesHandler.HandleCreate(w, r)
		} else {
			sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Zones endpoints (for pages.js compatibility)
	mux.HandleFunc("/get-zones", ar.pagesHandler.HandleGetZones)
	mux.HandleFunc("/create-zones", ar.pagesHandler.HandleCreateZones)
	mux.HandleFunc("/delete-zones", ar.pagesHandler.HandleDeleteZones)

	// Macros endpoints
	mux.HandleFunc("/api/macros/groups", ar.macrosHandler.HandleGroups)
	mux.HandleFunc("/api/macros/pinned", ar.macrosHandler.HandlePinned)
	mux.HandleFunc("/api/macros/move", ar.macrosHandler.HandleMove)
	mux.HandleFunc("/api/macros/copy", ar.macrosHandler.HandleCopy)
	mux.HandleFunc("/api/macros/pin-all", ar.macrosHandler.HandlePinAll)
	mux.HandleFunc("/api/macros/unpin-all", ar.macrosHandler.HandleUnpin)
	mux.HandleFunc("/api/macros/auto-grid", ar.macrosHandler.HandleAutoGrid)
	mux.HandleFunc("/api/macros/group-color", ar.macrosHandler.HandleGroupColor)
	mux.HandleFunc("/api/macros/group-title", ar.macrosHandler.HandleGroupTitle)

	// Remote upload endpoints
	mux.HandleFunc("/api/remote-upload", ar.uploadHandler.HandleUpload)
	mux.HandleFunc("/api/remote-upload/history", ar.uploadHandler.HandleHistory)

	// RCU endpoints
	mux.HandleFunc("/api/rcu/config", ar.rcuHandler.HandleConfig)
	mux.HandleFunc("/api/rcu/status", ar.rcuHandler.HandleStatus)
	mux.HandleFunc("/api/rcu/test", ar.rcuHandler.HandleTest)
	mux.HandleFunc("/identify-user", ar.rcuHandler.HandleIdentifyUser)
	mux.HandleFunc("/create-note", ar.rcuHandler.HandleCreateNote)
	mux.HandleFunc("/upload-item", ar.rcuHandler.HandleUploadItem)

	// Admin endpoints
	mux.HandleFunc("/api/admin/create-targets", ar.adminHandler.HandleCreateTargets)
	mux.HandleFunc("/api/admin/delete-targets", ar.adminHandler.HandleDeleteTargets)
	mux.HandleFunc("/api/admin/test-team", ar.adminHandler.HandleTestTeam)
	mux.HandleFunc("/api/admin/list-users", ar.adminHandler.HandleListUsers)
	mux.HandleFunc("/api/admin/delete-users", ar.adminHandler.HandleDeleteUsers)

	// Client override endpoint
	mux.HandleFunc("/api/client/override", ar.handleClientOverride)

	// Client list endpoint
	mux.HandleFunc("/api/clients", ar.handleClientList)

	// Restart canvas service endpoint
	mux.HandleFunc("/api/canvas/restart", ar.handleCanvasRestart)
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// handleCanvasInfo returns current canvas information.
func (ar *APIRoutes) handleCanvasInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	response := map[string]interface{}{
		"canvas_id":         ar.canvasService.GetCanvasID(),
		"canvas_name":       ar.canvasService.GetCanvasName(),
		"client_id":         ar.canvasService.GetClientID(),
		"client_name":       ar.canvasService.GetClientName(),
		"installation_name": ar.canvasService.GetInstallationName(),
		"connected":         ar.canvasService.IsConnected(),
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

// handleInstallationInfo returns installation information.
func (ar *APIRoutes) handleInstallationInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get canvas info (this will trigger fetch if name is missing)
	canvasID := ar.canvasService.GetCanvasID()
	canvasName := ar.canvasService.GetCanvasName()

	response := map[string]interface{}{
		"installation_name": ar.canvasService.GetInstallationName(),
		"client_id":         ar.canvasService.GetClientID(),
		"client_name":       ar.canvasService.GetClientName(),
		"canvas_id":         canvasID,
		"canvas_name":       canvasName,
		"connected":         ar.canvasService.IsConnected(),
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

// handleHealth returns server health status.
func (ar *APIRoutes) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	response := map[string]interface{}{
		"status":    "ok",
		"connected": ar.canvasService.IsConnected(),
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

// handleClientOverride handles manual client override requests.
func (ar *APIRoutes) handleClientOverride(w http.ResponseWriter, r *http.Request) {
	logger.Logf("[API] handleClientOverride called\n")
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Parse request body
	var request struct {
		ClientName string `json:"client_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Logf("[API] ERROR: Failed to decode request body: %v\n", err)
		response := map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	logger.Logf("[API] Override request - client_name: '%s'\n", request.ClientName)

	// Override client
	if err := ar.canvasService.OverrideClient(request.ClientName); err != nil {
		logger.Logf("[API] ERROR: OverrideClient failed: %v\n", err)
		response := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	logger.Logf("[API] OverrideClient succeeded\n")

	// Success response - use actual client name from service (may differ from request if matched by installation_name)
	actualClientName := ar.canvasService.GetClientName()
	if actualClientName == "" {
		// Fallback to requested name if service hasn't fetched it yet
		actualClientName = request.ClientName
	}

	response := map[string]interface{}{
		"success":     true,
		"client_name": actualClientName,
		"client_id":   ar.canvasService.GetClientID(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleClientList returns the list of available clients.
func (ar *APIRoutes) handleClientList(w http.ResponseWriter, r *http.Request) {
	logger.Logf("[API] handleClientList called\n")
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get clients from Canvus API
	logger.Logf("[API] Fetching clients from Canvus API...\n")
	clients, err := ar.apiClient.GetClients(r.Context())
	if err != nil {
		logger.Logf("[API] ERROR: Failed to fetch clients: %v\n", err)
		sendErrorResponse(w, fmt.Sprintf("Failed to fetch clients: %v", err), http.StatusInternalServerError)
		return
	}
	logger.Logf("[API] Successfully fetched %d clients\n", len(clients))

	// Return all clients with their installation_names
	validClients := make([]map[string]interface{}, 0, len(clients))
	for _, client := range clients {
		logger.Logf("[API] Client: ID=%s, InstallationName='%s'\n", client.ID, client.InstallationName)
		validClients = append(validClients, map[string]interface{}{
			"id":   client.ID,
			"name": client.InstallationName,
		})
	}

	logger.Logf("[API] Returning %d valid clients (with names)\n", len(validClients))
	response := map[string]interface{}{
		"success": true,
		"clients": validClients,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleServerInfo returns server information including actual IP address.
func (ar *APIRoutes) handleServerInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get actual server IP address from network interfaces
	serverIP := getServerIP()
	if serverIP == "" {
		// Fallback: try to get from request
		serverIP = r.Host
		if idx := strings.Index(serverIP, ":"); idx != -1 {
			serverIP = serverIP[:idx]
		}
		if serverIP == "" || serverIP == "localhost" || serverIP == "127.0.0.1" {
			serverIP = "localhost"
		}
	}

	// Get port from request or default
	port := "8080"
	host := r.Host
	if idx := strings.Index(host, ":"); idx != -1 {
		port = host[idx+1:]
	}

	// Build URL with actual IP
	protocol := getScheme(r)
	url := fmt.Sprintf("%s://%s:%s", protocol, serverIP, port)

	response := map[string]interface{}{
		"host":     serverIP,
		"hostname": serverIP,
		"port":     port,
		"url":      url,
		"ip":       serverIP, // Explicit IP field
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func getScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if scheme := r.Header.Get("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	return "http"
}

// getServerIP gets the actual IP address of the server from network interfaces.
// Returns the first non-loopback IPv4 address found.
func getServerIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			// Prefer IPv4
			if ip.To4() != nil {
				return ip.String()
			}
		}
	}

	return ""
}

// handleCanvasRestart handles requests to restart the canvas service.
func (ar *APIRoutes) handleCanvasRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	logger.Logf("[API] handleCanvasRestart called\n")

	// Restart canvas service
	if err := ar.canvasService.Restart(); err != nil {
		logger.Logf("[API] ERROR: Restart failed: %v\n", err)
		response := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	logger.Logf("[API] Restart succeeded\n")

	response := map[string]interface{}{
		"success": true,
		"message": "Canvas service restarted",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
