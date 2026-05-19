package webui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// HandleCreateZones handles POST /create-zones - Create zones or subzones
func (h *PagesHandler) HandleCreateZones(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		GridSize    interface{} `json:"gridSize"`    // Can be string or int
		GridPattern string      `json:"gridPattern"`
		SubZoneID   string      `json:"subZoneId"`
		SubZoneArray string     `json:"subZoneArray"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	canvasID := h.canvasService.GetCanvasID()
	if canvasID == "" {
		sendErrorResponse(w, "Canvas not available", http.StatusServiceUnavailable)
		return
	}

	// Handle subzone creation
	if req.SubZoneID != "" && req.SubZoneArray != "" {
		result, err := h.createSubZones(canvasID, req.SubZoneID, req.SubZoneArray)
		if err != nil {
			sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sendJSONResponse(w, result, http.StatusOK)
		return
	}

	// Handle main zone creation
	gridSize, err := parseGridSize(req.GridSize)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !isValidGridSize(gridSize) {
		sendErrorResponse(w, "Invalid grid size. Choose 1, 3, 4, or 5.", http.StatusBadRequest)
		return
	}

	gridPattern := req.GridPattern
	if gridPattern == "" {
		gridPattern = "Z"
	}

	if !isValidGridPattern(gridPattern) {
		sendErrorResponse(w, fmt.Sprintf("Invalid grid pattern. Choose one of: Z, Snake, Spiral"), http.StatusBadRequest)
		return
	}

	result, err := h.createInitialZones(canvasID, gridSize, gridPattern)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, result, http.StatusOK)
}

// HandleDeleteZones handles DELETE /delete-zones - Delete script-created zones
func (h *PagesHandler) HandleDeleteZones(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	canvasID := h.canvasService.GetCanvasID()
	if canvasID == "" {
		sendErrorResponse(w, "Canvas not available", http.StatusServiceUnavailable)
		return
	}

	// Fetch all anchors
	endpoint := fmt.Sprintf("/api/v1/canvases/%s/anchors", canvasID)
	data, err := h.apiClient.Get(endpoint)
	if err != nil {
		sendErrorResponse(w, fmt.Sprintf("Failed to fetch zones: %v", err), http.StatusInternalServerError)
		return
	}

	var anchors []map[string]interface{}
	if err := json.Unmarshal(data, &anchors); err != nil {
		sendErrorResponse(w, fmt.Sprintf("Failed to parse zones: %v", err), http.StatusInternalServerError)
		return
	}

	// Filter script-created anchors
	var scriptAnchors []map[string]interface{}
	for _, anchor := range anchors {
		if anchorName, ok := anchor["anchor_name"].(string); ok {
			if strings.HasSuffix(anchorName, "(Script Made)") {
				scriptAnchors = append(scriptAnchors, anchor)
			}
		}
	}

	if len(scriptAnchors) == 0 {
		sendJSONResponse(w, map[string]interface{}{
			"success": true,
			"message": "No script-created zones to delete.",
		}, http.StatusOK)
		return
	}

	deletedCount := 0
	failedCount := 0

	for _, anchor := range scriptAnchors {
		anchorID, ok := anchor["id"].(string)
		if !ok {
			failedCount++
			continue
		}

		deleteEndpoint := fmt.Sprintf("/api/v1/canvases/%s/anchors/%s", canvasID, anchorID)
		if err := h.apiClient.Delete(deleteEndpoint); err != nil {
			failedCount++
			continue
		}

		deletedCount++
	}

	sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("%d zones deleted successfully, %d failed.", deletedCount, failedCount),
	}, http.StatusOK)
}

// Helper functions

func parseGridSize(v interface{}) (int, error) {
	switch val := v.(type) {
	case int:
		return val, nil
	case float64:
		return int(val), nil
	case string:
		return strconv.Atoi(val)
	default:
		return 0, fmt.Errorf("invalid grid size type")
	}
}

func isValidGridSize(size int) bool {
	return size == 1 || size == 3 || size == 4 || size == 5
}

func isValidGridPattern(pattern string) bool {
	validPatterns := []string{"Z", "Snake", "Spiral"}
	for _, p := range validPatterns {
		if pattern == p {
			return true
		}
	}
	return false
}

func (h *PagesHandler) createSubZones(canvasID, subZoneID, subZoneArray string) (map[string]interface{}, error) {
	// Fetch selected zone
	endpoint := fmt.Sprintf("/api/v1/canvases/%s/anchors/%s", canvasID, subZoneID)
	data, err := h.apiClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch zone: %w", err)
	}

	var selectedZone map[string]interface{}
	if err := json.Unmarshal(data, &selectedZone); err != nil {
		return nil, fmt.Errorf("failed to parse zone: %w", err)
	}

	// Parse subzone array (e.g., "2x2", "3x3")
	parts := strings.Split(subZoneArray, "x")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid SubZone Array format")
	}

	cols, err := strconv.Atoi(parts[0])
	if err != nil || cols <= 0 {
		return nil, fmt.Errorf("invalid SubZone Array format")
	}

	rows, err := strconv.Atoi(parts[1])
	if err != nil || rows <= 0 {
		return nil, fmt.Errorf("invalid SubZone Array format")
	}

	// Get zone location and size
	location, _ := selectedZone["location"].(map[string]interface{})
	size, _ := selectedZone["size"].(map[string]interface{})
	anchorName, _ := selectedZone["anchor_name"].(string)

	zoneX := getFloat(location, "x")
	zoneY := getFloat(location, "y")
	zoneWidth := getFloat(size, "width")
	zoneHeight := getFloat(size, "height")

	subZoneWidth := zoneWidth / float64(cols)
	subZoneHeight := zoneHeight / float64(rows)

	// Extract zone number from anchor name for subzone naming
	zoneNumber := "1"
	if matches := strings.Split(anchorName, " "); len(matches) > 0 {
		for i, part := range matches {
			if part == "Zone" && i+1 < len(matches) {
				zoneNumber = matches[i+1]
				break
			}
		}
	}

	createdCount := 0
	failedCount := 0

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			subZoneName := fmt.Sprintf("SubZone %s.%d (Script Made)", zoneNumber, row*cols+col+1)
			payload := map[string]interface{}{
				"anchor_name": subZoneName,
				"location": map[string]interface{}{
					"x": zoneX + float64(col)*subZoneWidth,
					"y": zoneY + float64(row)*subZoneHeight,
				},
				"size": map[string]interface{}{
					"width":  subZoneWidth,
					"height": subZoneHeight,
				},
				"pinned": true,
				"scale":  1,
				"depth":  1,
			}

			createEndpoint := fmt.Sprintf("/api/v1/canvases/%s/anchors", canvasID)
			if _, err := h.apiClient.Post(createEndpoint, payload); err != nil {
				failedCount++
				continue
			}

			createdCount++
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("%d subzones created successfully, %d failed.", createdCount, failedCount),
	}, nil
}

func (h *PagesHandler) createInitialZones(canvasID string, gridSize int, gridPattern string) (map[string]interface{}, error) {
	// Fetch widgets to find SharedCanvas
	widgetsEndpoint := fmt.Sprintf("/api/v1/canvases/%s/widgets", canvasID)
	data, err := h.apiClient.Get(widgetsEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch widgets: %w", err)
	}

	var widgets []map[string]interface{}
	if err := json.Unmarshal(data, &widgets); err != nil {
		return nil, fmt.Errorf("failed to parse widgets: %w", err)
	}

	// Find SharedCanvas
	var sharedCanvas map[string]interface{}
	for _, widget := range widgets {
		if widgetType, _ := widget["widget_type"].(string); widgetType == "SharedCanvas" {
			sharedCanvas = widget
			break
		}
	}

	if sharedCanvas == nil {
		return nil, fmt.Errorf("SharedCanvas widget not found")
	}

	size, _ := sharedCanvas["size"].(map[string]interface{})
	canvasWidth := getFloat(size, "width")
	canvasHeight := getFloat(size, "height")

	if canvasWidth == 0 || canvasHeight == 0 {
		return nil, fmt.Errorf("invalid canvas size")
	}

	// Generate coordinates based on pattern
	coordinates := generateZoneCoordinates(gridSize, gridPattern, canvasWidth, canvasHeight)

	zoneWidth := canvasWidth / float64(gridSize)
	zoneHeight := canvasHeight / float64(gridSize)

	createdCount := 0
	failedCount := 0
	zoneNumber := 1

	for _, coord := range coordinates {
		anchorName := fmt.Sprintf("%dx%d Zone %d (Script Made)", gridSize, gridSize, zoneNumber)
		payload := map[string]interface{}{
			"anchor_name": anchorName,
			"location": map[string]interface{}{
				"x": coord.x,
				"y": coord.y,
			},
			"size": map[string]interface{}{
				"width":  zoneWidth,
				"height": zoneHeight,
			},
			"pinned": true,
			"scale":  1,
			"depth":  0,
		}

		createEndpoint := fmt.Sprintf("/api/v1/canvases/%s/anchors", canvasID)
		if _, err := h.apiClient.Post(createEndpoint, payload); err != nil {
			failedCount++
		} else {
			createdCount++
		}

		zoneNumber++
	}

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("%d zones created successfully, %d failed.", createdCount, failedCount),
	}, nil
}

type coord struct {
	x, y float64
}

func generateZoneCoordinates(gridSize int, pattern string, canvasWidth, canvasHeight float64) []coord {
	var gridCoords []gridCoord

	switch pattern {
	case "Z":
		gridCoords = generateZOrder(gridSize)
	case "Snake":
		gridCoords = generateSnakeOrder(gridSize)
	case "Spiral":
		gridCoords = generateSpiralOrder(gridSize)
	default:
		gridCoords = generateZOrder(gridSize)
	}

	result := make([]coord, len(gridCoords))
	for i, c := range gridCoords {
		result[i] = coord{
			x: float64(c.col) * (canvasWidth / float64(gridSize)),
			y: float64(c.row) * (canvasHeight / float64(gridSize)),
		}
	}

	return result
}

type gridCoord struct {
	row, col int
}

func generateZOrder(gridSize int) []gridCoord {
	var coords []gridCoord
	for row := 0; row < gridSize; row++ {
		for col := 0; col < gridSize; col++ {
			coords = append(coords, gridCoord{row: row, col: col})
		}
	}
	return coords
}

func generateSnakeOrder(gridSize int) []gridCoord {
	var coords []gridCoord
	for row := 0; row < gridSize; row++ {
		cols := make([]int, gridSize)
		for col := 0; col < gridSize; col++ {
			cols[col] = col
		}
		if row%2 != 0 {
			// Reverse for odd rows
			for i, j := 0, len(cols)-1; i < j; i, j = i+1, j-1 {
				cols[i], cols[j] = cols[j], cols[i]
			}
		}
		for _, col := range cols {
			coords = append(coords, gridCoord{row: row, col: col})
		}
	}
	return coords
}

func generateSpiralOrder(gridSize int) []gridCoord {
	var coords []gridCoord
	x := gridSize / 2
	y := gridSize / 2

	coords = append(coords, gridCoord{row: y, col: x})
	step := 1

	for len(coords) < gridSize*gridSize {
		// Move right
		for i := 0; i < step; i++ {
			x++
			if x >= 0 && x < gridSize && y >= 0 && y < gridSize {
				coords = append(coords, gridCoord{row: y, col: x})
			}
		}
		// Move down
		for i := 0; i < step; i++ {
			y++
			if x >= 0 && x < gridSize && y >= 0 && y < gridSize {
				coords = append(coords, gridCoord{row: y, col: x})
			}
		}

		step++

		// Move left
		for i := 0; i < step; i++ {
			x--
			if x >= 0 && x < gridSize && y >= 0 && y < gridSize {
				coords = append(coords, gridCoord{row: y, col: x})
			}
		}
		// Move up
		for i := 0; i < step; i++ {
			y--
			if x >= 0 && x < gridSize && y >= 0 && y < gridSize {
				coords = append(coords, gridCoord{row: y, col: x})
			}
		}

		step++
	}

	return coords
}

func getFloat(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}

