package webui

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	webuiatoms "github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/webui"
)

// MacrosHandler handles macros management API endpoints.
// NOTE: Macros are server-side operations that operate on widgets via the widgets API.
// They are NOT a Canvus API resource - they're abstractions built on top of widgets/anchors.
type MacrosHandler struct {
	apiClient     *webuiatoms.APIClient
	canvasService *CanvasService
}

// NewMacrosHandler creates a new macros handler.
func NewMacrosHandler(apiClient *webuiatoms.APIClient, canvasService *CanvasService) *MacrosHandler {
	return &MacrosHandler{
		apiClient:     apiClient,
		canvasService: canvasService,
	}
}

// HandleMove handles POST /api/macros/move - Move widgets from source zone to target zone.
func (h *MacrosHandler) HandleMove(w http.ResponseWriter, r *http.Request) {
	canvasID, ok := h.validateZoneRequest(w, r, http.MethodPost)
	if !ok {
		return
	}

	sourceZoneID, targetZoneID, err := parseZonePairRequest(r)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	movedCount, err := h.moveWidgets(canvasID, sourceZoneID, targetZoneID)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("%d widgets moved", movedCount),
	}, http.StatusOK)
}

// HandleCopy handles POST /api/macros/copy - Copy widgets from source zone to target zone.
func (h *MacrosHandler) HandleCopy(w http.ResponseWriter, r *http.Request) {
	canvasID, ok := h.validateZoneRequest(w, r, http.MethodPost)
	if !ok {
		return
	}

	sourceZoneID, targetZoneID, err := parseZonePairRequest(r)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	copiedCount, err := h.copyWidgets(canvasID, sourceZoneID, targetZoneID)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("%d widgets copied", copiedCount),
	}, http.StatusOK)
}

// HandleGroups handles GET /api/macros/groups - List widget groups (computed from widgets).
func (h *MacrosHandler) HandleGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Groups are computed from widgets - for now return empty array
	// Full implementation would group widgets by color, title, etc.
	sendJSONResponse(w, []interface{}{}, http.StatusOK)
}

// HandlePinned handles GET /api/macros/pinned - List pinned widgets.
func (h *MacrosHandler) HandlePinned(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	canvasID, ok := h.validateZoneRequest(w, r, http.MethodGet)
	if !ok {
		return
	}

	// Get all widgets and filter pinned ones
	allWidgets, err := webuiatoms.GetAllWidgets(r.Context(), h.apiClient, canvasID)
	if err != nil {
		sendErrorResponse(w, fmt.Sprintf("Failed to get widgets: %v", err), http.StatusInternalServerError)
		return
	}

	var pinned []webuiatoms.Widget
	for _, w := range allWidgets {
		if w.Pinned {
			pinned = append(pinned, w)
		}
	}

	sendJSONResponse(w, pinned, http.StatusOK)
}

// HandleUnpin handles POST /api/macros/unpin-all - Unpin all widgets in a zone.
func (h *MacrosHandler) HandleUnpin(w http.ResponseWriter, r *http.Request) {
	canvasID, ok := h.validateZoneRequest(w, r, http.MethodPost)
	if !ok {
		return
	}

	zoneID, err := parseZoneIDRequest(r)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	unpinnedCount, err := h.pinWidgetsInZone(canvasID, zoneID, false)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("%d widgets unpinned", unpinnedCount),
	}, http.StatusOK)
}

// HandlePinAll handles POST /api/macros/pin-all - Pin all widgets in a zone.
func (h *MacrosHandler) HandlePinAll(w http.ResponseWriter, r *http.Request) {
	canvasID, ok := h.validateZoneRequest(w, r, http.MethodPost)
	if !ok {
		return
	}

	zoneID, err := parseZoneIDRequest(r)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	pinnedCount, err := h.pinWidgetsInZone(canvasID, zoneID, true)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("%d widgets pinned", pinnedCount),
	}, http.StatusOK)
}

// HandleAutoGrid handles POST /api/macros/auto-grid - Organize widgets in a grid within a zone.
func (h *MacrosHandler) HandleAutoGrid(w http.ResponseWriter, r *http.Request) {
	canvasID, ok := h.validateZoneRequest(w, r, http.MethodPost)
	if !ok {
		return
	}

	zoneID, err := parseZoneIDRequest(r)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	griddedCount, err := h.organizeWidgetsInGrid(canvasID, zoneID)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if griddedCount == 0 {
		sendJSONResponse(w, map[string]interface{}{
			"success": true,
			"message": "No widgets found to auto-grid",
		}, http.StatusOK)
		return
	}

	sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("%d widgets organized in grid", griddedCount),
	}, http.StatusOK)
}

// HandleGroupColor handles POST /api/macros/group-color - Group widgets by color.
func (h *MacrosHandler) HandleGroupColor(w http.ResponseWriter, r *http.Request) {
	canvasID, ok := h.validateZoneRequest(w, r, http.MethodPost)
	if !ok {
		return
	}

	zoneID, err := parseZoneIDRequest(r)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Group by background_color - only for Note widgets that have background_color
	groupedCount, err := h.groupWidgetsByColor(canvasID, zoneID)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("%d widgets grouped by color", groupedCount),
	}, http.StatusOK)
}

// HandleGroupTitle handles POST /api/macros/group-title - Group widgets by title.
func (h *MacrosHandler) HandleGroupTitle(w http.ResponseWriter, r *http.Request) {
	canvasID, ok := h.validateZoneRequest(w, r, http.MethodPost)
	if !ok {
		return
	}

	zoneID, err := parseZoneIDRequest(r)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get widgets for sorting
	ops := NewMacrosOperations(h.apiClient, h.canvasService)
	zoneBB, allWidgets, err := ops.GetZoneAndWidgets(zoneID)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter widgets in zone
	inZone := FilterWidgetsInZone(allWidgets, zoneBB, zoneID)

	if len(inZone) == 0 {
		sendJSONResponse(w, map[string]interface{}{
			"success": true,
			"message": "No widgets found to group by title",
		}, http.StatusOK)
		return
	}

	// Sort by title
	sort.Slice(inZone, func(i, j int) bool {
		return strings.ToLower(inZone[i].Title) < strings.ToLower(inZone[j].Title)
	})

	// Group by title
	titleGroups := make(map[string][]webuiatoms.Widget)
	for _, w := range inZone {
		title := w.Title
		if title == "" {
			title = "untitled"
		}
		titleGroups[title] = append(titleGroups[title], w)
	}

	// Position widgets by title groups
	groupedCount := ops.PositionWidgetGroups(titleGroups, zoneBB, canvasID)

	sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("%d widgets grouped by title", groupedCount),
	}, http.StatusOK)
}
