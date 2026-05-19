package webui

import (
	"context"
	"fmt"
	"strings"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/logger"
	webuiatoms "github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/webui"
)

// MacrosOperations provides reusable operations for macros functionality.
type MacrosOperations struct {
	apiClient     *webuiatoms.APIClient
	canvasService *CanvasService
}

// NewMacrosOperations creates a new macros operations helper.
func NewMacrosOperations(apiClient *webuiatoms.APIClient, canvasService *CanvasService) *MacrosOperations {
	return &MacrosOperations{
		apiClient:     apiClient,
		canvasService: canvasService,
	}
}

// GetZoneAndWidgets gets zone bounding box and all widgets in that zone.
func (mo *MacrosOperations) GetZoneAndWidgets(zoneID string) (*webuiatoms.ZoneBoundingBox, []webuiatoms.Widget, error) {
	canvasID := mo.canvasService.GetCanvasID()
	if canvasID == "" {
		return nil, nil, fmt.Errorf("canvas not available")
	}

	zoneBB, err := webuiatoms.GetZoneBoundingBox(mo.apiClient, canvasID, zoneID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get zone: %w", err)
	}

	allWidgets, err := webuiatoms.GetAllWidgets(context.Background(), mo.apiClient, canvasID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get widgets: %w", err)
	}

	return zoneBB, allWidgets, nil
}

// FilterWidgetsInZone filters widgets that are within a zone, excluding anchors and connectors.
func FilterWidgetsInZone(widgets []webuiatoms.Widget, zoneBB *webuiatoms.ZoneBoundingBox, excludeZoneID string) []webuiatoms.Widget {
	logger.Logf("[FilterWidgetsInZone] Filtering %d widgets, zoneBB: X=%.2f, Y=%.2f, W=%.2f, H=%.2f\n",
		len(widgets), zoneBB.X, zoneBB.Y, zoneBB.Width, zoneBB.Height)

	var filtered []webuiatoms.Widget
	checkedCount := 0
	skippedCount := 0

	for _, w := range widgets {
		if w.ID == excludeZoneID {
			skippedCount++
			continue
		}
		if w.Location == nil {
			skippedCount++
			continue
		}
		wt := strings.ToLower(w.WidgetType)
		if wt == "connector" || wt == "anchor" {
			skippedCount++
			continue
		}
		checkedCount++
		if webuiatoms.WidgetIsInZone(&w, zoneBB) {
			filtered = append(filtered, w)
			logger.Logf("[FilterWidgetsInZone] Widget %s (%s) is IN zone\n", w.ID[:8], w.WidgetType)
		}
	}

	logger.Logf("[FilterWidgetsInZone] Result: %d widgets in zone (checked %d, skipped %d)\n", len(filtered), checkedCount, skippedCount)
	return filtered
}

// UpdateWidgetWithRetry updates a widget with retry logic (up to 3 attempts).
// Uses type-specific endpoints (not /widgets which is read-only).
func (mo *MacrosOperations) UpdateWidgetWithRetry(canvasID, widgetID, widgetType string, payload map[string]interface{}) error {
	// Get type-specific endpoint
	baseEndpoint := webuiatoms.GetWidgetPatchEndpoint(widgetType)
	endpoint := fmt.Sprintf("/api/v1/canvases/%s%s/%s", canvasID, baseEndpoint, widgetID)

	logger.Logf("[UpdateWidgetWithRetry] Updating widget %s (type: %s) via %s\n", widgetID[:8], widgetType, endpoint)

	var lastErr error
	for tries := 0; tries < 3; tries++ {
		_, err := mo.apiClient.Patch(endpoint, payload)
		if err == nil {
			logger.Logf("[UpdateWidgetWithRetry] Successfully updated widget %s\n", widgetID[:8])
			return nil
		}
		logger.Logf("[UpdateWidgetWithRetry] Attempt %d failed for widget %s: %v\n", tries+1, widgetID[:8], err)
		lastErr = err
	}
	logger.Logf("[UpdateWidgetWithRetry] ERROR: Failed to update widget %s after 3 attempts: %v\n", widgetID[:8], lastErr)
	return fmt.Errorf("failed after 3 attempts: %w", lastErr)
}

// BatchUpdateWidgets updates multiple widgets and returns count of successful updates.
func (mo *MacrosOperations) BatchUpdateWidgets(canvasID string, updates []WidgetUpdate) int {
	logger.Logf("[BatchUpdateWidgets] Updating %d widgets\n", len(updates))
	successCount := 0
	for i, update := range updates {
		if err := mo.UpdateWidgetWithRetry(canvasID, update.WidgetID, update.WidgetType, update.Payload); err == nil {
			successCount++
		} else {
			logger.Logf("[BatchUpdateWidgets] Failed to update widget %d/%d (ID: %s, Type: %s): %v\n",
				i+1, len(updates), update.WidgetID[:8], update.WidgetType, err)
		}
	}
	logger.Logf("[BatchUpdateWidgets] Successfully updated %d/%d widgets\n", successCount, len(updates))
	return successCount
}

// WidgetUpdate represents a widget update operation.
type WidgetUpdate struct {
	WidgetID   string
	WidgetType string
	Payload    map[string]interface{}
}

// CalculateOptimalGrid calculates the optimal grid dimensions for n widgets in a zone.
func CalculateOptimalGrid(n int, zoneBB *webuiatoms.ZoneBoundingBox) (rows, cols int) {
	aspectRatio := zoneBB.Width / zoneBB.Height

	bestRows := 1
	bestCols := n
	minEmptySpace := 1e10

	for r := 1; r <= n; r++ {
		c := (n + r - 1) / r // Ceiling division
		gridAspectRatio := float64(c) / float64(r)
		emptySpace := abs(aspectRatio - gridAspectRatio)

		if emptySpace < minEmptySpace {
			minEmptySpace = emptySpace
			bestRows = r
			bestCols = c
		}
	}

	return bestRows, bestCols
}

// CalculateCellDimensions calculates cell width and height for a grid.
func CalculateCellDimensions(zoneBB *webuiatoms.ZoneBoundingBox, rows, cols int) (width, height float64) {
	buffer := 100.0
	width = (zoneBB.Width - buffer*float64(cols+1)) / float64(cols)
	height = (zoneBB.Height - buffer*float64(rows+1)) / float64(rows)
	return width, height
}

// PositionWidgetGroups positions widget groups horizontally with vertical stacking within groups.
func (mo *MacrosOperations) PositionWidgetGroups(groups map[string][]webuiatoms.Widget, zoneBB *webuiatoms.ZoneBoundingBox, canvasID string) int {
	logger.Logf("[PositionWidgetGroups] Positioning %d groups in zone\n", len(groups))
	groupedCount := 0
	xOffset := zoneBB.X + 100
	for groupKey, widgets := range groups {
		logger.Logf("[PositionWidgetGroups] Group '%s': %d widgets\n", groupKey, len(widgets))
		yOffset := zoneBB.Y + 100
		for _, widget := range widgets {
			// Use type-specific endpoint (not /widgets which is read-only)
			baseEndpoint := webuiatoms.GetWidgetPatchEndpoint(widget.WidgetType)
			endpoint := fmt.Sprintf("/api/v1/canvases/%s%s/%s", canvasID, baseEndpoint, widget.ID)
			payload := map[string]interface{}{
				"location": map[string]float64{"x": xOffset, "y": yOffset},
			}
			logger.Logf("[PositionWidgetGroups] Patching widget %s (%s) to (%.2f, %.2f) via %s\n",
				widget.ID[:8], widget.WidgetType, xOffset, yOffset, endpoint)
			_, err := mo.apiClient.Patch(endpoint, payload)
			if err == nil {
				groupedCount++
				logger.Logf("[PositionWidgetGroups] Successfully positioned widget %s\n", widget.ID[:8])
			} else {
				logger.Logf("[PositionWidgetGroups] ERROR: Failed to position widget %s: %v\n", widget.ID[:8], err)
			}
			yOffset += 200
		}
		xOffset += 300
	}
	logger.Logf("[PositionWidgetGroups] Completed: %d widgets positioned\n", groupedCount)
	return groupedCount
}

// abs returns absolute value of a float64.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

