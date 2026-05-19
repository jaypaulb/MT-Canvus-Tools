package webui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	webuiatoms "github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/webui"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/organisms/services"
)

type rcuConfig struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
	Timeout int  `json:"timeout"`
}

type rcuStatus struct {
	Connected  bool `json:"connected"`
	LastUpdate any  `json:"last_update"`
}

// RCUHandler handles RCU configuration API endpoints.
// Config/status is maintained in-memory and resets on server restart.
// User data is persisted to users.json via fileService.
type RCUHandler struct {
	apiClient     *webuiatoms.APIClient
	canvasService *CanvasService
	fileService   *services.FileService
	usersPath     string
	mu            sync.RWMutex
	usersMu       sync.RWMutex
	config        rcuConfig
	status        rcuStatus
}

// NewRCUHandler creates a new RCU handler with no external dependencies.
// Suitable for tests and admin-surface-only usage (config/status endpoints).
func NewRCUHandler() *RCUHandler {
	return NewRCUHandlerWithDeps(nil, nil)
}

// NewRCUHandlerWithDeps creates a new RCU handler with API client and canvas service.
// Use this in production where HandleCreateNote, HandleUploadItem, and HandleIdentifyUser are needed.
func NewRCUHandlerWithDeps(apiClient *webuiatoms.APIClient, canvasService *CanvasService) *RCUHandler {
	fileService, _ := services.NewFileService()
	usersPath := ""
	if fileService != nil {
		usersPath = filepath.Join(fileService.GetUserConfigPath(), "users.json")
	}

	return &RCUHandler{
		apiClient:     apiClient,
		canvasService: canvasService,
		fileService:   fileService,
		usersPath:     usersPath,
		config:        rcuConfig{Enabled: false, Port: 8080, Timeout: 30},
		status:        rcuStatus{Connected: false},
	}
}

// HandleConfig handles GET and POST for /api/rcu/config.
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

// HandleStatus handles GET /api/rcu/status.
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

// HandleTest handles POST /api/rcu/test.
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

// HandleIdentifyUser handles POST /identify-user - Identify user and assign color.
func (h *RCUHandler) HandleIdentifyUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Team int    `json:"team"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Team < 1 || req.Team > 7 {
		sendErrorResponse(w, "Team must be between 1 and 7", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		sendErrorResponse(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Team base colors (ROYGBIV)
	teamColors := map[int]string{
		1: "#FF0000FF", // Red
		2: "#FF7F00FF", // Orange
		3: "#FFFF00FF", // Yellow
		4: "#00FF00FF", // Green
		5: "#0000FFFF", // Blue
		6: "#4B0082FF", // Indigo
		7: "#8B00FFFF", // Violet
	}

	users := h.loadUsers()
	teamKey := fmt.Sprintf("%d", req.Team)

	if users[teamKey] == nil {
		users[teamKey] = make(map[string]interface{})
	}

	teamUsers := users[teamKey].(map[string]interface{})
	if existingColor, ok := teamUsers[req.Name].(string); ok {
		sendJSONResponse(w, map[string]interface{}{
			"success": true,
			"color":   existingColor,
		}, http.StatusOK)
		return
	}

	baseColor := teamColors[req.Team]
	userColor := generateColorVariationHSL(baseColor, req.Name)

	teamUsers[req.Name] = userColor
	h.saveUsers(users)

	sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"color":   userColor,
	}, http.StatusOK)
}

// HandleCreateNote handles POST /create-note - Create note near team target.
func (h *RCUHandler) HandleCreateNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Team  int    `json:"team"`
		Name  string `json:"name"`
		Text  string `json:"text"`
		Color string `json:"color"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Team < 1 || req.Team > 7 {
		sendErrorResponse(w, "Team must be between 1 and 7", http.StatusBadRequest)
		return
	}

	if req.Text == "" {
		sendErrorResponse(w, "Text is required", http.StatusBadRequest)
		return
	}

	canvasID := h.canvasService.GetCanvasID()
	if canvasID == "" {
		sendErrorResponse(w, "Canvas not available", http.StatusServiceUnavailable)
		return
	}

	// Find the team's target note
	widgetsEndpoint := fmt.Sprintf("/api/v1/canvases/%s/widgets", canvasID)
	data, err := h.apiClient.Get(widgetsEndpoint)
	if err != nil {
		sendErrorResponse(w, fmt.Sprintf("Failed to fetch widgets: %v", err), http.StatusInternalServerError)
		return
	}

	var widgets []map[string]interface{}
	if err := json.Unmarshal(data, &widgets); err != nil {
		sendErrorResponse(w, fmt.Sprintf("Failed to parse widgets: %v", err), http.StatusInternalServerError)
		return
	}

	targetTitle := fmt.Sprintf("Team_%d_Target", req.Team)
	var targetNote map[string]interface{}
	for _, widget := range widgets {
		widgetType, _ := widget["widget_type"].(string)
		title, _ := widget["title"].(string)
		if widgetType == "Note" && title == targetTitle {
			targetNote = widget
			break
		}
	}

	if targetNote == nil {
		sendErrorResponse(w, fmt.Sprintf("Target note for Team %d not found. Please create targets first.", req.Team), http.StatusNotFound)
		return
	}

	location, ok := targetNote["location"].(map[string]interface{})
	if !ok || location == nil {
		sendErrorResponse(w, "Target note has no location", http.StatusInternalServerError)
		return
	}

	noteLocation := map[string]interface{}{
		"x": getFloat(location, "x") + 100,
		"y": getFloat(location, "y") + 100,
	}

	noteColor := req.Color
	if noteColor == "" {
		noteColor = "#FFFFFF00"
	}

	now := time.Now()
	dateStr := now.Format("06/01/02")
	timeStr := now.Format("15:04")
	title := fmt.Sprintf("%s @ %s - %s", req.Name, dateStr, timeStr)

	payload := map[string]interface{}{
		"title":            title,
		"text":             req.Text,
		"background_color": noteColor,
		"location":         noteLocation,
		"auto_text_color":  true,
		"state":            "normal",
	}

	noteEndpoint := fmt.Sprintf("/api/v1/canvases/%s/notes", canvasID)
	_, err = h.apiClient.Post(noteEndpoint, payload)
	if err != nil {
		sendErrorResponse(w, fmt.Sprintf("Failed to create note: %v", err), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Note posted successfully",
	}, http.StatusOK)
}

// HandleUploadItem handles POST /upload-item - Upload file and create widget.
func (h *RCUHandler) HandleUploadItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		sendErrorResponse(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	teamStr := r.FormValue("team")
	name := r.FormValue("name")
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		sendErrorResponse(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		sendErrorResponse(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	fileReader := bytes.NewReader(fileData)

	team, err := parseInt(teamStr)
	if err != nil || team < 1 || team > 7 {
		sendErrorResponse(w, "Team must be between 1 and 7", http.StatusBadRequest)
		return
	}

	canvasID := h.canvasService.GetCanvasID()
	if canvasID == "" {
		sendErrorResponse(w, "Canvas not available", http.StatusServiceUnavailable)
		return
	}

	widgetsEndpoint := fmt.Sprintf("/api/v1/canvases/%s/widgets", canvasID)
	data, err := h.apiClient.Get(widgetsEndpoint)
	if err != nil {
		sendErrorResponse(w, fmt.Sprintf("Failed to fetch widgets: %v", err), http.StatusInternalServerError)
		return
	}

	var widgets []map[string]interface{}
	if err := json.Unmarshal(data, &widgets); err != nil {
		sendErrorResponse(w, fmt.Sprintf("Failed to parse widgets: %v", err), http.StatusInternalServerError)
		return
	}

	targetTitle := fmt.Sprintf("Team_%d_Target", team)
	var targetNote map[string]interface{}
	for _, widget := range widgets {
		widgetType, _ := widget["widget_type"].(string)
		title, _ := widget["title"].(string)
		if widgetType == "Note" && title == targetTitle {
			targetNote = widget
			break
		}
	}

	if targetNote == nil {
		sendErrorResponse(w, fmt.Sprintf("Target note for Team %d not found. Please create targets first.", team), http.StatusNotFound)
		return
	}

	location, ok := targetNote["location"].(map[string]interface{})
	if !ok || location == nil {
		sendErrorResponse(w, "Target note has no location", http.StatusInternalServerError)
		return
	}

	fileLocation := map[string]interface{}{
		"x": getFloat(location, "x") + 100,
		"y": getFloat(location, "y") + 100,
	}

	fileName := fileHeader.Filename
	ext := getFileExtension(fileName)

	now := time.Now()
	dateStr := now.Format("06/01/02")
	timeStr := now.Format("15:04")
	title := fmt.Sprintf("%s @ %s - %s", name, dateStr, timeStr)

	var endpoint string
	jsonPayload := map[string]interface{}{
		"title":    title,
		"location": fileLocation,
	}

	if isImageFile(ext) {
		endpoint = fmt.Sprintf("/api/v1/canvases/%s/images", canvasID)
	} else if isVideoFile(ext) {
		endpoint = fmt.Sprintf("/api/v1/canvases/%s/videos", canvasID)
	} else {
		endpoint = fmt.Sprintf("/api/v1/canvases/%s/pdfs", canvasID)
	}

	_, err = h.apiClient.PostMultipart(endpoint, jsonPayload, fileReader, fileName)
	if err != nil {
		sendErrorResponse(w, fmt.Sprintf("Failed to upload file: %v", err), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("File uploaded successfully: %s", fileName),
	}, http.StatusOK)
}

// loadUsers loads users from persistent storage.
func (h *RCUHandler) loadUsers() map[string]interface{} {
	h.usersMu.RLock()
	defer h.usersMu.RUnlock()

	users := make(map[string]interface{})
	if h.usersPath == "" || h.fileService == nil {
		return users
	}

	if err := h.fileService.ReadJSONFile(h.usersPath, &users); err != nil {
		return make(map[string]interface{})
	}

	return users
}

// saveUsers saves users to persistent storage.
func (h *RCUHandler) saveUsers(users map[string]interface{}) {
	h.usersMu.Lock()
	defer h.usersMu.Unlock()

	if h.usersPath == "" || h.fileService == nil {
		return
	}

	h.fileService.WriteJSONFile(h.usersPath, users) //nolint:errcheck
}

// Helper functions

// generateColorVariationHSL generates a color variation using HSL to stay in team color range.
func generateColorVariationHSL(baseColor string, name string) string {
	hsl := hexToHSL(baseColor)
	if hsl == nil {
		return baseColor
	}

	hash := 0
	for _, char := range name {
		hash = int(char) + ((hash << 5) - hash)
	}
	hash = hash & 0xFFFFFF

	lightnessVariation := float64(hash%21) - 10.0
	newL := hsl.L + lightnessVariation
	if newL < 0 {
		newL = 0
	}
	if newL > 100 {
		newL = 100
	}

	return hslToHex(hsl.H, hsl.S, newL)
}

// HSL represents HSL color values.
type HSL struct {
	H float64
	S float64
	L float64
}

// hexToHSL converts a hex color (#RRGGBBAA or #RRGGBB) to HSL.
func hexToHSL(hex string) *HSL {
	if len(hex) < 7 || hex[0] != '#' {
		return nil
	}

	r, _ := parseHex(hex[1:3])
	g, _ := parseHex(hex[3:5])
	b, _ := parseHex(hex[5:7])

	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := rf
	if gf > max {
		max = gf
	}
	if bf > max {
		max = bf
	}

	min := rf
	if gf < min {
		min = gf
	}
	if bf < min {
		min = bf
	}

	l := (max + min) / 2.0
	var h, s float64

	if max == min {
		h = 0
		s = 0
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2.0 - max - min)
		} else {
			s = d / (max + min)
		}
		switch {
		case max == rf:
			h = ((gf - bf) / d)
			if gf < bf {
				h += 6
			}
		case max == gf:
			h = ((bf - rf) / d) + 2
		case max == bf:
			h = ((rf - gf) / d) + 4
		}
		h /= 6
	}

	return &HSL{H: h * 360, S: s * 100, L: l * 100}
}

// hslToHex converts HSL to hex color (#RRGGBBFF).
func hslToHex(h, s, l float64) string {
	h = h / 360.0
	s = s / 100.0
	l = l / 100.0

	var r, g, b float64

	if s == 0 {
		r = l
		g = l
		b = l
	} else {
		hue2rgb := func(p, q, t float64) float64 {
			if t < 0 {
				t += 1
			}
			if t > 1 {
				t -= 1
			}
			if t < 1.0/6.0 {
				return p + (q-p)*6*t
			}
			if t < 1.0/2.0 {
				return q
			}
			if t < 2.0/3.0 {
				return p + (q-p)*(2.0/3.0-t)*6
			}
			return p
		}

		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hue2rgb(p, q, h+1.0/3.0)
		g = hue2rgb(p, q, h)
		b = hue2rgb(p, q, h-1.0/3.0)
	}

	ri := int(r * 255)
	gi := int(g * 255)
	bi := int(b * 255)

	return fmt.Sprintf("#%02X%02X%02XFF", ri, gi, bi)
}

func parseHex(s string) (int, error) {
	var result int
	for _, char := range s {
		result *= 16
		if char >= '0' && char <= '9' {
			result += int(char - '0')
		} else if char >= 'A' && char <= 'F' {
			result += int(char-'A') + 10
		} else if char >= 'a' && char <= 'f' {
			result += int(char-'a') + 10
		} else {
			return 0, fmt.Errorf("invalid hex character: %c", char)
		}
	}
	return result, nil
}

func parseInt(s string) (int, error) {
	var result int
	for _, char := range s {
		if char < '0' || char > '9' {
			return 0, fmt.Errorf("invalid number: %s", s)
		}
		result = result*10 + int(char-'0')
	}
	return result, nil
}

func getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return ""
	}
	return strings.ToLower(parts[len(parts)-1])
}

func isImageFile(ext string) bool {
	imageExts := []string{"jpg", "jpeg", "png", "gif", "bmp", "tiff", "webp"}
	for _, e := range imageExts {
		if ext == e {
			return true
		}
	}
	return false
}

func isVideoFile(ext string) bool {
	videoExts := []string{"mp4", "avi", "mov", "wmv", "mkv", "webm"}
	for _, e := range videoExts {
		if ext == e {
			return true
		}
	}
	return false
}

