package webui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	webuiatoms "github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/webui"
)

// AdminHandler handles admin API endpoints for RCU management.
type AdminHandler struct {
	apiClient     *webuiatoms.APIClient
	canvasService *CanvasService
	rcuHandler    *RCUHandler // Reference to RCU handler for creating notes
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(apiClient *webuiatoms.APIClient, canvasService *CanvasService, rcuHandler *RCUHandler) *AdminHandler {
	return &AdminHandler{
		apiClient:     apiClient,
		canvasService: canvasService,
		rcuHandler:    rcuHandler,
	}
}

// HandleCreateTargets handles POST /api/admin/create-targets - Create target notes for teams 1-7.
func (h *AdminHandler) HandleCreateTargets(w http.ResponseWriter, r *http.Request) {
	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			sendErrorResponse(w, fmt.Sprintf("Internal error: %v", r), http.StatusInternalServerError)
		}
	}()

	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	canvasID := h.canvasService.GetCanvasID()
	if canvasID == "" {
		sendErrorResponse(w, "Canvas not available", http.StatusServiceUnavailable)
		return
	}

	// Get zones (anchors) from Canvus API
	endpoint := fmt.Sprintf("/api/v1/canvases/%s/anchors", canvasID)
	data, err := h.apiClient.Get(endpoint)
	if err != nil {
		sendErrorResponse(w, fmt.Sprintf("Failed to fetch zones: %v", err), http.StatusInternalServerError)
		return
	}

	var zones []map[string]interface{}
	if err := json.Unmarshal(data, &zones); err != nil {
		sendErrorResponse(w, fmt.Sprintf("Failed to parse zones: %v", err), http.StatusInternalServerError)
		return
	}

	if len(zones) < 8 {
		sendErrorResponse(w, "Not enough zones available (need at least 8 zones)", http.StatusBadRequest)
		return
	}

	// Team colors (ROYGBIV)
	teamColors := map[int]string{
		1: "#FF0000FF", // Red
		2: "#FF7F00FF", // Orange
		3: "#FFFF00FF", // Yellow
		4: "#00FF00FF", // Green
		5: "#0000FFFF", // Blue
		6: "#4B0082FF", // Indigo
		7: "#8B00FFFF", // Violet
	}

	// Create notes for teams 1-7 (using zones with anchor_index 1-7, skipping index 0)
	createdTeams := []int{}
	var errors []string
	for i := 1; i <= 7; i++ {
		// Find zone with anchor_index == i
		var zone map[string]interface{}
		for _, z := range zones {
			if z == nil {
				continue
			}
			// Safely extract anchor_index (could be float64, int, or string)
			var anchorIndex int
			switch v := z["anchor_index"].(type) {
			case float64:
				anchorIndex = int(v)
			case int:
				anchorIndex = v
			case int64:
				anchorIndex = int(v)
			default:
				continue
			}
			if anchorIndex == i {
				zone = z
				break
			}
		}

		if zone == nil {
			errors = append(errors, fmt.Sprintf("Team %d: zone with anchor_index %d not found", i, i))
			continue
		}

		location, ok := zone["location"].(map[string]interface{})
		if !ok || location == nil {
			errors = append(errors, fmt.Sprintf("Team %d: zone has no location", i))
			continue
		}

		// Validate location has x and y coordinates
		if _, hasX := location["x"]; !hasX {
			errors = append(errors, fmt.Sprintf("Team %d: location missing x coordinate", i))
			continue
		}
		if _, hasY := location["y"]; !hasY {
			errors = append(errors, fmt.Sprintf("Team %d: location missing y coordinate", i))
			continue
		}

		noteTitle := fmt.Sprintf("Team_%d_Target", i)
		noteText := fmt.Sprintf("Team %d", i)
		noteColor := teamColors[i]

		// Create note widget at zone location
		// Note: Use /notes endpoint, not /widgets (widgets is read-only)
		payload := map[string]interface{}{
			"title":       noteTitle,
			"text":        noteText,
			"background_color": noteColor,
			"location":     location,
			"auto_text_color": true,
			"state":        "normal",
		}

		noteEndpoint := fmt.Sprintf("/api/v1/canvases/%s/notes", canvasID)
		_, err := h.apiClient.Post(noteEndpoint, payload)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Team %d: failed to create widget: %v", i, err))
			continue
		}

		createdTeams = append(createdTeams, i)
	}

	if len(createdTeams) == 0 {
		errorMsg := "Failed to create any team targets"
		if len(errors) > 0 {
			errorMsg += ": " + strings.Join(errors, "; ")
		}
		sendErrorResponse(w, errorMsg, http.StatusInternalServerError)
		return
	}

	message := fmt.Sprintf("Created Team Targets %s", formatTeamList(createdTeams))
	if len(errors) > 0 {
		message += fmt.Sprintf(" (warnings: %d failed)", len(errors))
	}

	response := map[string]interface{}{
		"success": true,
		"message": message,
		"teams":   createdTeams,
	}
	if len(errors) > 0 {
		response["warnings"] = errors
	}
	sendJSONResponse(w, response, http.StatusOK)
}

// HandleDeleteTargets handles POST /api/admin/delete-targets - Delete all target notes.
func (h *AdminHandler) HandleDeleteTargets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	canvasID := h.canvasService.GetCanvasID()
	if canvasID == "" {
		sendErrorResponse(w, "Canvas not available", http.StatusServiceUnavailable)
		return
	}

	// Get all widgets from Canvus API
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

	// Filter target notes - widgets with widget_type="Note" and title matching "Team_\d+_Target"
	targetPattern := regexp.MustCompile(`^Team_\d+_Target$`)
	var targetNotes []map[string]interface{}
	for _, widget := range widgets {
		widgetType, _ := widget["widget_type"].(string)
		title, _ := widget["title"].(string)
		if widgetType == "Note" && targetPattern.MatchString(title) {
			targetNotes = append(targetNotes, widget)
		}
	}

	// Delete each target note
	deletedCount := 0
	for _, note := range targetNotes {
		noteID, ok := note["id"].(string)
		if !ok {
			continue
		}

		deleteEndpoint := fmt.Sprintf("/api/v1/canvases/%s/notes/%s", canvasID, noteID)
		if err := h.apiClient.Delete(deleteEndpoint); err != nil {
			continue
		}

		deletedCount++
	}

	message := fmt.Sprintf("Deleted %d target notes", deletedCount)
	sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": message,
	}, http.StatusOK)
}

// HandleTestTeam handles POST /api/admin/test-team - Send test note to a team.
// Uses HandleCreateNote with user "Admin" instead of separate endpoint.
func (h *AdminHandler) HandleTestTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Team int    `json:"team"`
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Team < 1 || req.Team > 7 {
		sendErrorResponse(w, "Team must be between 1 and 7", http.StatusBadRequest)
		return
	}

	// Use RCU handler's create note functionality with user "Admin"
	// Get team color for Admin's note
	teamColors := map[int]string{
		1: "#FF0000FF", // Red
		2: "#FF7F00FF", // Orange
		3: "#FFFF00FF", // Yellow
		4: "#00FF00FF", // Green
		5: "#0000FFFF", // Blue
		6: "#4B0082FF", // Indigo
		7: "#8B00FFFF", // Violet
	}
	adminColor := teamColors[req.Team]

	noteText := req.Text
	if noteText == "" {
		noteText = fmt.Sprintf("Test note from Admin to Team %d", req.Team)
	}

	// Create a new request body for HandleCreateNote
	createNoteReq := map[string]interface{}{
		"team":  req.Team,
		"name":  "Admin",
		"text":  noteText,
		"color": adminColor,
	}

	// Create a new request with the same context
	createNoteBody, _ := json.Marshal(createNoteReq)
	newReq, _ := http.NewRequestWithContext(r.Context(), "POST", "/create-note", strings.NewReader(string(createNoteBody)))
	newReq.Header.Set("Content-Type", "application/json")

	// Call HandleCreateNote directly
	h.rcuHandler.HandleCreateNote(w, newReq)
}

// HandleListUsers handles GET /api/admin/list-users - List RCU users.
func (h *AdminHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Load users from RCU handler's persistent storage
	users := h.rcuHandler.loadUsers()

	// Convert structure {team: {username: color}} to a list
	userList := []map[string]interface{}{}
	for teamKey, teamUsers := range users {
		teamUsersMap, ok := teamUsers.(map[string]interface{})
		if !ok {
			continue
		}

		team, err := parseInt(teamKey)
		if err != nil {
			continue
		}

		for username, color := range teamUsersMap {
			colorStr, ok := color.(string)
			if !ok {
				continue
			}
			userList = append(userList, map[string]interface{}{
				"team":  team,
				"name":  username,
				"color": colorStr,
			})
		}
	}

	sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"users":   userList,
	}, http.StatusOK)
}

// HandleDeleteUsers handles POST /api/admin/delete-users - Delete RCU users.
func (h *AdminHandler) HandleDeleteUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		All   bool     `json:"all"`
		Users []string `json:"users"` // Format: ["team:name", ...]
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	users := h.rcuHandler.loadUsers()

	if req.All {
		// Delete all users
		users = make(map[string]interface{})
	} else if len(req.Users) > 0 {
		// Delete specific users
		for _, userKey := range req.Users {
			// Parse "team:name" format
			parts := strings.Split(userKey, ":")
			if len(parts) != 2 {
				continue
			}
			teamKey := parts[0]
			userName := parts[1]

			if teamUsers, ok := users[teamKey].(map[string]interface{}); ok {
				delete(teamUsers, userName)
			}
		}
	}

	h.rcuHandler.saveUsers(users)

	message := "Users deleted successfully"
	if !req.All && len(req.Users) == 0 {
		message = "No users to delete"
	}

	sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": message,
	}, http.StatusOK)
}

// Helper functions

func formatTeamList(teams []int) string {
	if len(teams) == 0 {
		return ""
	}
	parts := make([]string, len(teams))
	for i, team := range teams {
		parts[i] = fmt.Sprintf("%d", team)
	}
	return strings.Join(parts, ", ")
}

