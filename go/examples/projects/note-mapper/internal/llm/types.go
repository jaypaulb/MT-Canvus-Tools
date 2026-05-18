// Package llm provides Google Gemini multimodal integration for note-mapper.
//
// ExtractPostitNotes sends an image to the Gemini API and returns structured
// data describing each Post-it note visible in the image: its text, colour,
// pixel position, and dimensions.
package llm

// ExtractInput carries the image data and MIME type sent to Gemini.
type ExtractInput struct {
	// ImageData is the raw image bytes (JPEG or PNG).
	ImageData []byte
	// MimeType is the MIME type of the image (e.g. "image/jpeg").
	MimeType string
}

// Note is a single Post-it note as extracted by the LLM.
//
// Pixel coordinates are relative to the top-left of the source image.
// The caller is responsible for mapping these to the target Canvus canvas
// anchor coordinate space (see package mapping).
type Note struct {
	// Content is the text written on the note.
	Content string
	// Color is the background hex colour string (e.g. "#FFEB3B").
	Color string
	// X is the left edge of the note in source-image pixels.
	X int
	// Y is the top edge of the note in source-image pixels.
	Y int
	// Width is the note width in source-image pixels.
	Width int
	// Height is the note height in source-image pixels.
	Height int
	// Scale is a note-level scale hint from the LLM (typically 1.0).
	Scale float64
}

// extractOutput is the raw JSON shape returned by Gemini before conversion
// to Note.
type extractOutput struct {
	BackgroundColor string         `json:"background_color"`
	Location        map[string]int `json:"location"`
	Scale           float64        `json:"scale"`
	Size            map[string]int `json:"size"`
	Text            string         `json:"text"`
	WidgetType      string         `json:"widget_type"`
}
