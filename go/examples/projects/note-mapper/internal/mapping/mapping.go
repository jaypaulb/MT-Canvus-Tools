// Package mapping converts Post-it note pixel coordinates from image space into
// Canvus canvas anchor space.
//
// The Canvus API uses absolute pixel coordinates (origin at the top-left of the
// canvas, not normalised 0–1). Anchors are positioned the same way. When
// placing notes inside an anchor zone, coordinates must be scaled and offset to
// account for the zone size and position.
package mapping

import (
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/note-mapper/internal/llm"
)

// ZoneParams describes the target anchor zone in canvas pixels.
type ZoneParams struct {
	// X is the left edge of the zone in canvas pixels.
	X float64
	// Y is the top edge of the zone in canvas pixels.
	Y float64
	// Width is the zone width in canvas pixels.
	Width float64
	// Height is the zone height in canvas pixels.
	Height float64
	// Scale is the anchor's scale factor.
	Scale float64
}

// NoteForAPI is the Canvus API payload shape for note creation.
type NoteForAPI struct {
	BackgroundColor string             `json:"background_color"`
	Text            string             `json:"text"`
	Location        map[string]float64 `json:"location"`
	Size            map[string]float64 `json:"size"`
	Scale           float64            `json:"scale"`
	WidgetType      string             `json:"widget_type"`
	State           string             `json:"state"`
}

// MapToZone maps notes from image pixel coordinates into the target anchor
// zone. The image dimensions (imageW × imageH) define the source coordinate
// space. The zone is fit-letterboxed: the image is scaled uniformly to fill as
// much of the zone as possible while preserving aspect ratio.
//
// A 10-pixel inset is applied to each edge so notes don't sit flush against
// zone borders (matching the source behaviour).
func MapToZone(notes []llm.Note, imageW, imageH int, zone ZoneParams) []NoteForAPI {
	const inset = 10.0

	if imageW == 0 || imageH == 0 || zone.Width == 0 || zone.Height == 0 {
		// Cannot scale; return notes at zone origin.
		return toAPIFormat(notes, zone.X+inset, zone.Y+inset, 1.0, 1.0)
	}

	// Uniform scale to fit the image aspect ratio inside the zone.
	scale := min(zone.Width/float64(imageW), zone.Height/float64(imageH))
	// Centering offset within the zone.
	offsetX := (zone.Width - float64(imageW)*scale) / 2
	offsetY := (zone.Height - float64(imageH)*scale) / 2

	result := make([]NoteForAPI, 0, len(notes))
	for _, n := range notes {
		x := zone.X + inset + float64(n.X)*scale + offsetX
		y := zone.Y + inset + float64(n.Y)*scale + offsetY
		w := float64(n.Width) * scale
		h := float64(n.Height) * scale

		result = append(result, NoteForAPI{
			BackgroundColor: n.Color,
			Text:            n.Content,
			Location:        map[string]float64{"x": x, "y": y},
			Size:            map[string]float64{"width": w, "height": h},
			Scale:           1.0, // all scaling absorbed into position/size math
			WidgetType:      "Note",
			State:           "normal",
		})
	}
	return result
}

// MapWithAnchorScale scales notes from pre-mapped image-zone coordinates into
// final canvas coordinates, accounting for the anchor's own scale factor.
// This mirrors the two-step logic in the source: first MapToZone to get
// zone-relative positions, then this function to apply the anchor scale and
// place notes absolutely on the canvas.
func MapWithAnchorScale(notes []llm.Note, imageW, imageH int, zone ZoneParams) []NoteForAPI {
	const inset = 10.0

	if imageW == 0 || imageH == 0 {
		return toAPIFormat(notes, zone.X, zone.Y, zone.Scale, zone.Scale)
	}

	scaleFactor := 1.0
	if zone.Width > 0 && zone.Height > 0 {
		scaleFactor = min(zone.Width/float64(imageW), zone.Height/float64(imageH))
	}
	finalScale := scaleFactor * zone.Scale

	result := make([]NoteForAPI, 0, len(notes))
	for _, n := range notes {
		x := zone.X + inset + float64(n.X)*finalScale
		y := zone.Y + inset + float64(n.Y)*finalScale
		w := float64(n.Width) * finalScale
		h := float64(n.Height) * finalScale

		result = append(result, NoteForAPI{
			BackgroundColor: n.Color,
			Text:            n.Content,
			Location:        map[string]float64{"x": x, "y": y},
			Size:            map[string]float64{"width": w, "height": h},
			Scale:           1.0,
			WidgetType:      "Note",
			State:           "normal",
		})
	}
	return result
}

// toAPIFormat is a simple helper that converts notes without spatial mapping.
func toAPIFormat(notes []llm.Note, baseX, baseY, scaleX, scaleY float64) []NoteForAPI {
	result := make([]NoteForAPI, 0, len(notes))
	for _, n := range notes {
		result = append(result, NoteForAPI{
			BackgroundColor: n.Color,
			Text:            n.Content,
			Location:        map[string]float64{"x": baseX + float64(n.X)*scaleX, "y": baseY + float64(n.Y)*scaleY},
			Size:            map[string]float64{"width": float64(n.Width) * scaleX, "height": float64(n.Height) * scaleY},
			Scale:           1.0,
			WidgetType:      "Note",
			State:           "normal",
		})
	}
	return result
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
