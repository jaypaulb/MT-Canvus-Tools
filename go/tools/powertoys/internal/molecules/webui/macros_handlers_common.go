package webui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/logger"
	webuiatoms "github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/webui"
)

// validateZoneRequest validates a zone request and returns the canvas ID.
func (h *MacrosHandler) validateZoneRequest(w http.ResponseWriter, r *http.Request, method string) (string, bool) {
	if r.Method != method {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return "", false
	}

	canvasID := h.canvasService.GetCanvasID()
	if canvasID == "" {
		logger.Logf("[MacrosHandler] ERROR: Canvas ID is empty - canvas not available yet\n")
		logger.Logf("[MacrosHandler] ClientID: %s, Connected: %v\n", h.canvasService.GetClientID(), h.canvasService.IsConnected())
		sendErrorResponse(w, "Canvas not available", http.StatusServiceUnavailable)
		return "", false
	}

	logger.Logf("[MacrosHandler] Validated request - CanvasID: %s\n", canvasID)
	return canvasID, true
}

// parseZoneIDRequest parses a request body expecting a zoneId field.
func parseZoneIDRequest(r *http.Request) (string, error) {
	var req struct {
		ZoneID string `json:"zoneId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return "", fmt.Errorf("invalid request body: %w", err)
	}
	if req.ZoneID == "" {
		return "", fmt.Errorf("zoneId is required")
	}
	return req.ZoneID, nil
}

// parseZonePairRequest parses a request body expecting sourceZoneId and targetZoneId fields.
func parseZonePairRequest(r *http.Request) (sourceZoneID, targetZoneID string, err error) {
	var req struct {
		SourceZoneID string `json:"sourceZoneId"`
		TargetZoneID string `json:"targetZoneId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return "", "", fmt.Errorf("invalid request body: %w", err)
	}
	if req.SourceZoneID == "" || req.TargetZoneID == "" {
		return "", "", fmt.Errorf("sourceZoneId and targetZoneId are required")
	}
	return req.SourceZoneID, req.TargetZoneID, nil
}

// sendJSONResponse sends a JSON response with CORS headers.
func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// sendErrorResponse sends an error response as JSON.
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

// moveWidgets moves widgets from source zone to target zone.
func (h *MacrosHandler) moveWidgets(canvasID, sourceZoneID, targetZoneID string) (int, error) {
	logger.Logf("[MacrosHandler] moveWidgets - canvasID: %s, sourceZoneID: %s, targetZoneID: %s\n", canvasID, sourceZoneID, targetZoneID)
	ops := NewMacrosOperations(h.apiClient, h.canvasService)

	// Get zone bounding boxes
	sourceBB, err := webuiatoms.GetZoneBoundingBox(h.apiClient, canvasID, sourceZoneID)
	if err != nil {
		logger.Logf("[MacrosHandler] ERROR: Failed to get source zone: %v\n", err)
		return 0, fmt.Errorf("failed to get source zone: %w", err)
	}
	logger.Logf("[MacrosHandler] Source zone BB: X=%.2f, Y=%.2f, W=%.2f, H=%.2f, Scale=%.2f\n", sourceBB.X, sourceBB.Y, sourceBB.Width, sourceBB.Height, sourceBB.Scale)

	targetBB, err := webuiatoms.GetZoneBoundingBox(h.apiClient, canvasID, targetZoneID)
	if err != nil {
		logger.Logf("[MacrosHandler] ERROR: Failed to get target zone: %v\n", err)
		return 0, fmt.Errorf("failed to get target zone: %w", err)
	}
	logger.Logf("[MacrosHandler] Target zone BB: X=%.2f, Y=%.2f, W=%.2f, H=%.2f, Scale=%.2f\n", targetBB.X, targetBB.Y, targetBB.Width, targetBB.Height, targetBB.Scale)

	// Get all widgets
	allWidgets, err := webuiatoms.GetAllWidgets(context.Background(), h.apiClient, canvasID)
	if err != nil {
		logger.Logf("[MacrosHandler] ERROR: Failed to get widgets: %v\n", err)
		return 0, fmt.Errorf("failed to get widgets: %w", err)
	}
	logger.Logf("[MacrosHandler] Retrieved %d total widgets\n", len(allWidgets))

	// Log first few widgets for debugging
	for i, w := range allWidgets {
		if i >= 5 {
			break
		}
		if w.Location != nil {
			logger.Logf("[MacrosHandler] Widget[%d]: ID=%s, Type=%s, Location=(%.2f, %.2f), Size=(%.2f, %.2f), Scale=%.2f\n",
				i, w.ID, w.WidgetType, w.Location.X, w.Location.Y,
				func() float64 {
					if w.Size != nil {
						return w.Size.Width
					} else {
						return 0
					}
				}(),
				func() float64 {
					if w.Size != nil {
						return w.Size.Height
					} else {
						return 0
					}
				}(),
				w.Scale)
		}
	}

	// Filter widgets in source zone
	toMove := FilterWidgetsInZone(allWidgets, sourceBB, "")
	logger.Logf("[MacrosHandler] Found %d widgets in source zone to move\n", len(toMove))

	// Transform and update each widget
	var updates []WidgetUpdate
	for _, widget := range toMove {
		cloned := widget
		webuiatoms.TransformWidgetLocationAndScale(&cloned, sourceBB, targetBB)
		updates = append(updates, WidgetUpdate{
			WidgetID:   widget.ID,
			WidgetType: widget.WidgetType,
			Payload: map[string]interface{}{
				"location": cloned.Location,
				"scale":    cloned.Scale,
			},
		})
		logger.Logf("[MacrosHandler] Prepared update for widget %s (%s): location=(%.2f, %.2f), scale=%.2f\n",
			widget.ID[:8], widget.WidgetType, cloned.Location.X, cloned.Location.Y, cloned.Scale)
	}

	movedCount := ops.BatchUpdateWidgets(canvasID, updates)
	logger.Logf("[MacrosHandler] moveWidgets completed: %d widgets moved\n", movedCount)
	return movedCount, nil
}

// copyWidgets copies widgets from source zone to target zone.
func (h *MacrosHandler) copyWidgets(canvasID, sourceZoneID, targetZoneID string) (int, error) {
	logger.Logf("[MacrosHandler] copyWidgets - canvasID: %s, sourceZoneID: %s, targetZoneID: %s\n", canvasID, sourceZoneID, targetZoneID)

	// Get zone bounding boxes
	sourceBB, err := webuiatoms.GetZoneBoundingBox(h.apiClient, canvasID, sourceZoneID)
	if err != nil {
		return 0, fmt.Errorf("failed to get source zone: %w", err)
	}

	targetBB, err := webuiatoms.GetZoneBoundingBox(h.apiClient, canvasID, targetZoneID)
	if err != nil {
		return 0, fmt.Errorf("failed to get target zone: %w", err)
	}

	// Get all widgets
	allWidgets, err := webuiatoms.GetAllWidgets(context.Background(), h.apiClient, canvasID)
	if err != nil {
		return 0, fmt.Errorf("failed to get widgets: %w", err)
	}

	// Filter widgets in source zone
	toCopy := FilterWidgetsInZone(allWidgets, sourceBB, "")
	logger.Logf("[MacrosHandler] Found %d widgets in source zone to copy\n", len(toCopy))

	// Copy widgets (create new widgets with transformed locations)
	copiedCount := 0
	for _, widget := range toCopy {
		cloned := widget
		webuiatoms.TransformWidgetLocationAndScale(&cloned, sourceBB, targetBB)
		widgetType := strings.ToLower(widget.WidgetType)

		var err error
		switch widgetType {
		case "note":
			err = h.copyNote(canvasID, widget.ID, &cloned)
		case "image":
			err = h.copyImage(canvasID, widget.ID, &cloned)
		case "video":
			err = h.copyVideo(canvasID, widget.ID, &cloned)
		case "pdf":
			err = h.copyPDF(canvasID, widget.ID, &cloned)
		default:
			logger.Logf("[MacrosHandler] Skipping unsupported widget type: %s\n", widget.WidgetType)
			continue
		}

		if err == nil {
			copiedCount++
			logger.Logf("[MacrosHandler] Successfully copied widget %s (%s)\n", widget.ID[:8], widget.WidgetType)
		} else {
			logger.Logf("[MacrosHandler] ERROR: Failed to copy widget %s (%s): %v\n", widget.ID[:8], widget.WidgetType, err)
		}
	}

	logger.Logf("[MacrosHandler] copyWidgets completed: %d widgets copied\n", copiedCount)
	return copiedCount, nil
}

// copyNote copies a note widget, fetching full note data to get text and background_color.
func (h *MacrosHandler) copyNote(canvasID, noteID string, cloned *webuiatoms.Widget) error {
	// Fetch full note data to get text and background_color
	endpoint := fmt.Sprintf("/api/v1/canvases/%s/notes/%s", canvasID, noteID)
	data, err := h.apiClient.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch note: %w", err)
	}

	var noteData map[string]interface{}
	if err := json.Unmarshal(data, &noteData); err != nil {
		return fmt.Errorf("failed to parse note: %w", err)
	}

	// Build payload with all note fields
	payload := map[string]interface{}{
		"location": cloned.Location,
		"scale":    cloned.Scale,
	}
	if cloned.Size != nil {
		payload["size"] = cloned.Size
	}
	if title, ok := noteData["title"].(string); ok {
		payload["title"] = title
	}
	if text, ok := noteData["text"].(string); ok {
		payload["text"] = text
	}
	if bgColor, ok := noteData["background_color"].(string); ok {
		payload["background_color"] = bgColor
	}

	// Create new note
	createEndpoint := fmt.Sprintf("/api/v1/canvases/%s/notes", canvasID)
	_, err = h.apiClient.Post(createEndpoint, payload)
	return err
}

// copyImage copies an image widget by downloading and re-uploading the file.
func (h *MacrosHandler) copyImage(canvasID, imageID string, cloned *webuiatoms.Widget) error {
	// Download image file
	downloadEndpoint := fmt.Sprintf("/api/v1/canvases/%s/images/%s/download", canvasID, imageID)
	fileData, err := h.apiClient.Get(downloadEndpoint)
	if err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}

	// Get image metadata
	metaEndpoint := fmt.Sprintf("/api/v1/canvases/%s/images/%s", canvasID, imageID)
	metaData, err := h.apiClient.Get(metaEndpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch image metadata: %w", err)
	}

	var imageMeta map[string]interface{}
	if err := json.Unmarshal(metaData, &imageMeta); err != nil {
		return fmt.Errorf("failed to parse image metadata: %w", err)
	}

	// Build JSON payload
	jsonPayload := map[string]interface{}{
		"location": cloned.Location,
		"scale":    cloned.Scale,
	}
	if cloned.Size != nil {
		jsonPayload["size"] = cloned.Size
	}
	if title, ok := imageMeta["title"].(string); ok {
		jsonPayload["title"] = title
	}
	if origFilename, ok := imageMeta["original_filename"].(string); ok {
		jsonPayload["original_filename"] = origFilename
	}

	// Determine filename
	fileName := "image.jpg"
	if origFilename, ok := imageMeta["original_filename"].(string); ok && origFilename != "" {
		fileName = origFilename
	}

	// Upload image using multipart
	createEndpoint := fmt.Sprintf("/api/v1/canvases/%s/images", canvasID)
	fileReader := bytes.NewReader(fileData)
	_, err = h.apiClient.PostMultipart(createEndpoint, jsonPayload, fileReader, fileName)
	return err
}

// copyVideo copies a video widget by downloading and re-uploading the file.
func (h *MacrosHandler) copyVideo(canvasID, videoID string, cloned *webuiatoms.Widget) error {
	// Download video file
	downloadEndpoint := fmt.Sprintf("/api/v1/canvases/%s/videos/%s/download", canvasID, videoID)
	fileData, err := h.apiClient.Get(downloadEndpoint)
	if err != nil {
		return fmt.Errorf("failed to download video: %w", err)
	}

	// Get video metadata
	metaEndpoint := fmt.Sprintf("/api/v1/canvases/%s/videos/%s", canvasID, videoID)
	metaData, err := h.apiClient.Get(metaEndpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch video metadata: %w", err)
	}

	var videoMeta map[string]interface{}
	if err := json.Unmarshal(metaData, &videoMeta); err != nil {
		return fmt.Errorf("failed to parse video metadata: %w", err)
	}

	// Build JSON payload
	jsonPayload := map[string]interface{}{
		"location": cloned.Location,
		"scale":    cloned.Scale,
	}
	if cloned.Size != nil {
		jsonPayload["size"] = cloned.Size
	}
	if title, ok := videoMeta["title"].(string); ok {
		jsonPayload["title"] = title
	}
	if origFilename, ok := videoMeta["original_filename"].(string); ok {
		jsonPayload["original_filename"] = origFilename
	}

	// Determine filename
	fileName := "video.mp4"
	if origFilename, ok := videoMeta["original_filename"].(string); ok && origFilename != "" {
		fileName = origFilename
	}

	// Upload video using multipart
	createEndpoint := fmt.Sprintf("/api/v1/canvases/%s/videos", canvasID)
	fileReader := bytes.NewReader(fileData)
	_, err = h.apiClient.PostMultipart(createEndpoint, jsonPayload, fileReader, fileName)
	return err
}

// copyPDF copies a PDF widget by downloading and re-uploading the file.
func (h *MacrosHandler) copyPDF(canvasID, pdfID string, cloned *webuiatoms.Widget) error {
	// Download PDF file
	downloadEndpoint := fmt.Sprintf("/api/v1/canvases/%s/pdfs/%s/download", canvasID, pdfID)
	fileData, err := h.apiClient.Get(downloadEndpoint)
	if err != nil {
		return fmt.Errorf("failed to download PDF: %w", err)
	}

	// Get PDF metadata
	metaEndpoint := fmt.Sprintf("/api/v1/canvases/%s/pdfs/%s", canvasID, pdfID)
	metaData, err := h.apiClient.Get(metaEndpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch PDF metadata: %w", err)
	}

	var pdfMeta map[string]interface{}
	if err := json.Unmarshal(metaData, &pdfMeta); err != nil {
		return fmt.Errorf("failed to parse PDF metadata: %w", err)
	}

	// Build JSON payload
	jsonPayload := map[string]interface{}{
		"location": cloned.Location,
		"scale":    cloned.Scale,
	}
	if cloned.Size != nil {
		jsonPayload["size"] = cloned.Size
	}
	if title, ok := pdfMeta["title"].(string); ok {
		jsonPayload["title"] = title
	}
	if origFilename, ok := pdfMeta["original_filename"].(string); ok {
		jsonPayload["original_filename"] = origFilename
	}
	if index, ok := pdfMeta["index"].(float64); ok {
		jsonPayload["index"] = int(index)
	}

	// Determine filename
	fileName := "document.pdf"
	if origFilename, ok := pdfMeta["original_filename"].(string); ok && origFilename != "" {
		fileName = origFilename
	}

	// Upload PDF using multipart
	createEndpoint := fmt.Sprintf("/api/v1/canvases/%s/pdfs", canvasID)
	fileReader := bytes.NewReader(fileData)
	_, err = h.apiClient.PostMultipart(createEndpoint, jsonPayload, fileReader, fileName)
	return err
}

// pinWidgetsInZone pins or unpins all widgets in a zone.
func (h *MacrosHandler) pinWidgetsInZone(canvasID, zoneID string, pinned bool) (int, error) {
	logger.Logf("[MacrosHandler] pinWidgetsInZone - canvasID: %s, zoneID: %s, pinned: %v\n", canvasID, zoneID, pinned)
	ops := NewMacrosOperations(h.apiClient, h.canvasService)

	zoneBB, allWidgets, err := ops.GetZoneAndWidgets(zoneID)
	if err != nil {
		logger.Logf("[MacrosHandler] ERROR: pinWidgetsInZone failed to get zone/widgets: %v\n", err)
		return 0, err
	}

	// Filter widgets in zone
	inZone := FilterWidgetsInZone(allWidgets, zoneBB, zoneID)
	logger.Logf("[MacrosHandler] pinWidgetsInZone: Found %d widgets in zone\n", len(inZone))

	// Update widgets
	var updates []WidgetUpdate
	for _, widget := range inZone {
		updates = append(updates, WidgetUpdate{
			WidgetID:   widget.ID,
			WidgetType: widget.WidgetType,
			Payload:    map[string]interface{}{"pinned": pinned},
		})
		logger.Logf("[MacrosHandler] pinWidgetsInZone: Prepared update for widget %s (%s) - pinned: %v\n",
			widget.ID[:8], widget.WidgetType, pinned)
	}

	logger.Logf("[MacrosHandler] pinWidgetsInZone: Prepared %d updates, calling BatchUpdateWidgets\n", len(updates))
	pinnedCount := ops.BatchUpdateWidgets(canvasID, updates)
	logger.Logf("[MacrosHandler] pinWidgetsInZone completed: %d widgets updated\n", pinnedCount)
	return pinnedCount, nil
}

// organizeWidgetsInGrid organizes widgets in a grid within a zone.
func (h *MacrosHandler) organizeWidgetsInGrid(canvasID, zoneID string) (int, error) {
	logger.Logf("[MacrosHandler] organizeWidgetsInGrid - canvasID: %s, zoneID: %s\n", canvasID, zoneID)
	ops := NewMacrosOperations(h.apiClient, h.canvasService)

	zoneBB, allWidgets, err := ops.GetZoneAndWidgets(zoneID)
	if err != nil {
		logger.Logf("[MacrosHandler] ERROR: organizeWidgetsInGrid failed to get zone/widgets: %v\n", err)
		return 0, err
	}

	// Filter widgets in zone
	inZone := FilterWidgetsInZone(allWidgets, zoneBB, zoneID)
	logger.Logf("[MacrosHandler] organizeWidgetsInGrid: Found %d widgets in zone\n", len(inZone))

	if len(inZone) == 0 {
		logger.Logf("[MacrosHandler] organizeWidgetsInGrid: No widgets to organize\n")
		return 0, nil
	}

	// Determine optimal grid size
	bestRows, bestCols := CalculateOptimalGrid(len(inZone), zoneBB)
	cellWidth, cellHeight := CalculateCellDimensions(zoneBB, bestRows, bestCols)
	logger.Logf("[MacrosHandler] organizeWidgetsInGrid: Grid layout: %d rows x %d cols, cell size: %.2f x %.2f\n", bestRows, bestCols, cellWidth, cellHeight)

	// Position widgets in grid
	var updates []WidgetUpdate
	buffer := 100.0
	for i, widget := range inZone {
		row := i / bestCols
		col := i % bestCols
		x := zoneBB.X + buffer + float64(col)*(cellWidth+buffer)
		y := zoneBB.Y + buffer + float64(row)*(cellHeight+buffer)

		updates = append(updates, WidgetUpdate{
			WidgetID:   widget.ID,
			WidgetType: widget.WidgetType,
			Payload: map[string]interface{}{
				"location": map[string]float64{"x": x, "y": y},
			},
		})
		logger.Logf("[MacrosHandler] organizeWidgetsInGrid: Prepared update for widget %s (%s) at (%.2f, %.2f)\n",
			widget.ID[:8], widget.WidgetType, x, y)
	}

	logger.Logf("[MacrosHandler] organizeWidgetsInGrid: Prepared %d updates, calling BatchUpdateWidgets\n", len(updates))
	griddedCount := ops.BatchUpdateWidgets(canvasID, updates)
	logger.Logf("[MacrosHandler] organizeWidgetsInGrid completed: %d widgets organized\n", griddedCount)
	return griddedCount, nil
}

// groupWidgetsByAttribute groups widgets by an attribute (color or title) and positions them.
// Uses bounding boxes to filter widgets within the zone before grouping.
func (h *MacrosHandler) groupWidgetsByAttribute(canvasID, zoneID string, getAttribute func(webuiatoms.Widget) string) (int, error) {
	logger.Logf("[MacrosHandler] groupWidgetsByAttribute - canvasID: %s, zoneID: %s\n", canvasID, zoneID)
	ops := NewMacrosOperations(h.apiClient, h.canvasService)

	zoneBB, allWidgets, err := ops.GetZoneAndWidgets(zoneID)
	if err != nil {
		logger.Logf("[MacrosHandler] ERROR: groupWidgetsByAttribute failed to get zone/widgets: %v\n", err)
		return 0, err
	}

	// Filter widgets in zone using bounding box (excludes anchors/connectors)
	inZone := FilterWidgetsInZone(allWidgets, zoneBB, zoneID)
	logger.Logf("[MacrosHandler] groupWidgetsByAttribute: Found %d widgets in zone (using bounding box)\n", len(inZone))

	if len(inZone) == 0 {
		logger.Logf("[MacrosHandler] groupWidgetsByAttribute: No widgets in zone to group\n")
		return 0, nil
	}

	// Group filtered widgets by attribute
	groups := make(map[string][]webuiatoms.Widget)
	for _, w := range inZone {
		attr := getAttribute(w)
		if attr == "" {
			attr = "default"
		}
		groups[attr] = append(groups[attr], w)
	}

	logger.Logf("[MacrosHandler] groupWidgetsByAttribute: Created %d groups with total %d widgets\n", len(groups), len(inZone))

	groupedCount := ops.PositionWidgetGroups(groups, zoneBB, canvasID)
	logger.Logf("[MacrosHandler] groupWidgetsByAttribute completed: %d widgets grouped\n", groupedCount)
	return groupedCount, nil
}

// groupWidgetsByColor groups Note widgets by their background_color.
// Only includes Note widgets that have a background_color field.
// Skips PDFs, images, videos, and notes without background_color.
func (h *MacrosHandler) groupWidgetsByColor(canvasID, zoneID string) (int, error) {
	logger.Logf("[MacrosHandler] groupWidgetsByColor - canvasID: %s, zoneID: %s\n", canvasID, zoneID)
	ops := NewMacrosOperations(h.apiClient, h.canvasService)

	zoneBB, allWidgets, err := ops.GetZoneAndWidgets(zoneID)
	if err != nil {
		logger.Logf("[MacrosHandler] ERROR: groupWidgetsByColor failed to get zone/widgets: %v\n", err)
		return 0, err
	}

	// Filter widgets in zone using bounding box
	inZone := FilterWidgetsInZone(allWidgets, zoneBB, zoneID)
	logger.Logf("[MacrosHandler] groupWidgetsByColor: Found %d widgets in zone\n", len(inZone))

	// Filter to only Note widgets and fetch their background_color
	type noteWithColor struct {
		widget          webuiatoms.Widget
		backgroundColor string
	}
	var notesWithColor []noteWithColor

	for _, widget := range inZone {
		// Only process Note widgets
		if strings.ToLower(widget.WidgetType) != "note" {
			logger.Logf("[MacrosHandler] groupWidgetsByColor: Skipping non-note widget %s (%s)\n", widget.ID[:8], widget.WidgetType)
			continue
		}

		// Fetch full note data to get background_color
		noteEndpoint := fmt.Sprintf("/api/v1/canvases/%s/notes/%s", canvasID, widget.ID)
		noteData, err := h.apiClient.Get(noteEndpoint)
		if err != nil {
			logger.Logf("[MacrosHandler] groupWidgetsByColor: ERROR - Failed to fetch note %s: %v\n", widget.ID[:8], err)
			continue
		}

		var note map[string]interface{}
		if err := json.Unmarshal(noteData, &note); err != nil {
			logger.Logf("[MacrosHandler] groupWidgetsByColor: ERROR - Failed to parse note %s: %v\n", widget.ID[:8], err)
			continue
		}

		// Get background_color if it exists
		bgColor, ok := note["background_color"].(string)
		if !ok || bgColor == "" {
			logger.Logf("[MacrosHandler] groupWidgetsByColor: Skipping note %s (no background_color)\n", widget.ID[:8])
			continue
		}

		notesWithColor = append(notesWithColor, noteWithColor{
			widget:          widget,
			backgroundColor: bgColor,
		})
		logger.Logf("[MacrosHandler] groupWidgetsByColor: Note %s has background_color: %s\n", widget.ID[:8], bgColor)
	}

	logger.Logf("[MacrosHandler] groupWidgetsByColor: Found %d notes with background_color\n", len(notesWithColor))

	if len(notesWithColor) == 0 {
		logger.Logf("[MacrosHandler] groupWidgetsByColor: No notes with background_color to group\n")
		return 0, nil
	}

	// Group by background_color
	groups := make(map[string][]webuiatoms.Widget)
	for _, nwc := range notesWithColor {
		groups[nwc.backgroundColor] = append(groups[nwc.backgroundColor], nwc.widget)
	}

	logger.Logf("[MacrosHandler] groupWidgetsByColor: Created %d color groups\n", len(groups))

	// Position groups
	groupedCount := ops.PositionWidgetGroups(groups, zoneBB, canvasID)
	logger.Logf("[MacrosHandler] groupWidgetsByColor completed: %d notes grouped by color\n", groupedCount)
	return groupedCount, nil
}
