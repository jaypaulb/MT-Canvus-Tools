package webui

import (
	"encoding/json"
	"fmt"
)

// ZoneBoundingBox represents a zone's bounding box from an anchor.
type ZoneBoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Scale  float64 `json:"scale"`
}

// GetZoneBoundingBox gets the bounding box for a zone (anchor) from the Canvus API.
func GetZoneBoundingBox(apiClient *APIClient, canvasID, zoneID string) (*ZoneBoundingBox, error) {
	if canvasID == "" {
		return nil, fmt.Errorf("canvas not available")
	}

	endpoint := fmt.Sprintf("/api/v1/canvases/%s/anchors/%s", canvasID, zoneID)
	fmt.Printf("[GetZoneBoundingBox] Fetching zone %s from %s\n", zoneID, endpoint)
	data, err := apiClient.Get(endpoint)
	if err != nil {
		fmt.Printf("[GetZoneBoundingBox] ERROR: Failed to get anchor: %v\n", err)
		return nil, fmt.Errorf("failed to get anchor: %w", err)
	}

	var anchor struct {
		Location *struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
		} `json:"location"`
		Size *struct {
			Width  float64 `json:"width"`
			Height float64 `json:"height"`
		} `json:"size"`
		Scale float64 `json:"scale"`
	}
	if err := json.Unmarshal(data, &anchor); err != nil {
		fmt.Printf("[GetZoneBoundingBox] ERROR: Failed to parse anchor JSON: %v\n", err)
		fmt.Printf("[GetZoneBoundingBox] Raw response: %s\n", string(data))
		return nil, fmt.Errorf("failed to parse anchor: %w", err)
	}

	if anchor.Location == nil || anchor.Size == nil {
		fmt.Printf("[GetZoneBoundingBox] ERROR: Invalid anchor data - Location=%v, Size=%v\n", anchor.Location, anchor.Size)
		return nil, fmt.Errorf("invalid anchor data for zone ID: %s", zoneID)
	}

	bb := &ZoneBoundingBox{
		X:      anchor.Location.X,
		Y:      anchor.Location.Y,
		Width:  anchor.Size.Width,
		Height: anchor.Size.Height,
		Scale:  anchor.Scale,
	}

	fmt.Printf("[GetZoneBoundingBox] Zone %s: X=%.2f, Y=%.2f, W=%.2f, H=%.2f, Scale=%.2f\n",
		zoneID, bb.X, bb.Y, bb.Width, bb.Height, bb.Scale)

	return bb, nil
}

// WidgetIsInZone checks if a widget is within a zone bounding box.
// Checks if widget's location point is within the zone bounds (with 2px margin).
// Note: This checks the widget's location point, not the widget's bounding box.
func WidgetIsInZone(widget *Widget, zoneBB *ZoneBoundingBox) bool {
	if widget.Location == nil {
		return false
	}
	wx := widget.Location.X
	wy := widget.Location.Y

	// Zone bounds (with 2px margin to avoid edge cases)
	zoneMinX := zoneBB.X + 2
	zoneMaxX := zoneBB.X + zoneBB.Width - 2
	zoneMinY := zoneBB.Y + 2
	zoneMaxY := zoneBB.Y + zoneBB.Height - 2

	withinX := wx >= zoneMinX && wx <= zoneMaxX
	withinY := wy >= zoneMinY && wy <= zoneMaxY

	result := withinX && withinY

	// Debug logging for widgets that should be in zone but aren't
	// Only log first 10 widgets to avoid spam
	// This will help identify why widgets aren't being detected
	if !result && widget.ID != "" {
		// Check if widget is close to zone (within 100 units) - might indicate coordinate system mismatch
		distX := wx - zoneBB.X
		distY := wy - zoneBB.Y
		if distX < 100 && distX > -100 && distY < 100 && distY > -100 {
			fmt.Printf("[WidgetIsInZone] Widget %s (%s) near zone but not in: location=(%.2f, %.2f), zone=(%.2f-%.2f, %.2f-%.2f), dist=(%.2f, %.2f)\n",
				widget.ID[:8], widget.WidgetType, wx, wy, zoneMinX, zoneMaxX, zoneMinY, zoneMaxY, distX, distY)
		}
	}

	return result
}

