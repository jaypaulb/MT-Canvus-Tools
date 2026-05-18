// Package api implements the HTTP handler layer for note-mapper.
//
// Routes (all under /api/):
//
//	POST   /api/upload-image       — Upload photo; trigger LLM extraction; return detected notes.
//	POST   /api/scan-notes         — Re-scan the last uploaded image with updated zone params.
//	POST   /api/create-notes       — Place notes into a Canvus anchor zone.
//	POST   /api/set-credentials    — Update Canvus server URL and API key in-session.
//	GET    /api/get-canvases       — List accessible Canvus canvases.
//	GET    /api/get-anchors        — List anchors for a canvas.
//	GET    /api/get-anchor-info    — Get geometry of a specific anchor.
//
// Global image state:
// The handler stores the last uploaded (and preprocessed) image in process
// memory. The scan-notes endpoint re-uses this data so the client can request
// repeated LLM passes without re-uploading the file. This is a single-tenant
// design — concurrent users will conflict.
package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/note-mapper/internal/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/note-mapper/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/note-mapper/internal/image"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/note-mapper/internal/llm"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/note-mapper/internal/mapping"
)

// lastImage holds the most recently preprocessed image in memory.
// Concurrency note: this is unprotected and assumes single-user operation.
var lastImage []byte
var lastImageMIME string

// UploadImageHandler handles POST /api/upload-image.
// Expects a multipart form with an "image" file field and optional zone parameters.
// Returns the detected notes immediately after LLM extraction.
func UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	file, fh, err := r.FormFile("image")
	if err != nil {
		slog.Warn("upload-image: missing file field", "err", err)
		writeError(w, http.StatusBadRequest, "image file required")
		return
	}
	defer func() { _ = file.Close() }()

	raw := make([]byte, fh.Size)
	if _, err := file.Read(raw); err != nil {
		slog.Error("upload-image: read file", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to read uploaded file")
		return
	}
	slog.Info("upload-image: received", "filename", fh.Filename, "bytes", len(raw))

	processed, mime, err := image.ProcessImage(raw)
	if err != nil {
		slog.Error("upload-image: process image", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to process image: "+err.Error())
		return
	}
	lastImage = processed
	lastImageMIME = mime

	zone := parseZoneParams(r)
	notes, err := llm.ExtractPostitNotes(llm.ExtractInput{ImageData: processed, MimeType: mime})
	if err != nil {
		slog.Error("upload-image: LLM extraction", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to extract notes: "+err.Error())
		return
	}
	slog.Info("upload-image: extraction complete", "note_count", len(notes))

	const defaultImgW, defaultImgH = 1280, 720
	mcsNotes := mapping.MapToZone(notes, defaultImgW, defaultImgH, zone)

	writeJSON(w, http.StatusOK, map[string]any{
		"status":      "complete",
		"message":     "Image processed successfully. Notes extracted.",
		"notes":       mcsNotes,
		"imageWidth":  defaultImgW,
		"imageHeight": defaultImgH,
	})
}

// ScanNotesHandler handles POST /api/scan-notes.
// Re-runs LLM extraction on the last uploaded image with updated zone params.
func ScanNotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if len(lastImage) == 0 {
		writeError(w, http.StatusBadRequest, "no image available for scanning")
		return
	}

	zone := parseZoneParams(r)
	notes, err := llm.ExtractPostitNotes(llm.ExtractInput{ImageData: lastImage, MimeType: lastImageMIME})
	if err != nil {
		slog.Error("scan-notes: LLM extraction", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to extract notes: "+err.Error())
		return
	}
	slog.Info("scan-notes: extraction complete", "note_count", len(notes))

	const defaultImgW, defaultImgH = 1280, 720
	mcsNotes := mapping.MapToZone(notes, defaultImgW, defaultImgH, zone)

	writeJSON(w, http.StatusOK, map[string]any{
		"status":      "complete",
		"message":     "LLM processing complete. Notes extracted.",
		"notes":       mcsNotes,
		"imageWidth":  defaultImgW,
		"imageHeight": defaultImgH,
	})
}

// createNotesRequest is the request body for POST /api/create-notes.
type createNotesRequest struct {
	CanvasID    string           `json:"canvasID"`
	Notes       []map[string]any `json:"notes"`
	ZoneID      string           `json:"zoneID"`
	ImageWidth  float64          `json:"imageWidth"`
	ImageHeight float64          `json:"imageHeight"`
}

// CreateNotesHandler handles POST /api/create-notes.
// Fetches anchor geometry and places the supplied notes into Canvus.
func CreateNotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req createNotesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.ImageWidth == 0 || req.ImageHeight == 0 {
		writeError(w, http.StatusBadRequest, "imageWidth and imageHeight required")
		return
	}

	cfg := config.Get()
	if cfg.CanvusURL == "" || cfg.APIKey == "" {
		writeError(w, http.StatusBadRequest, "Canvus credentials not set")
		return
	}

	ctx := context.Background()
	c := canvus.NewClient(cfg.CanvusURL, cfg.APIKey)

	anchor, err := c.GetAnchor(ctx, req.CanvasID, req.ZoneID)
	if err != nil {
		slog.Error("create-notes: get anchor", "canvas_id", req.CanvasID, "anchor_id", req.ZoneID, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch anchor info: "+err.Error())
		return
	}
	slog.Info("create-notes: anchor fetched", "anchor_id", anchor.ID, "x", anchor.X, "y", anchor.Y, "w", anchor.Width, "h", anchor.Height)

	// Apply the same two-stage scaling as the source: fit image into anchor,
	// then multiply by anchor.Scale.
	scaleFactor := 1.0
	if anchor.Width > 0 && anchor.Height > 0 {
		scaleFactor = minFloat(anchor.Width/req.ImageWidth, anchor.Height/req.ImageHeight)
	}
	finalScale := scaleFactor * anchor.Scale
	slog.Debug("create-notes: scale", "scale_factor", scaleFactor, "anchor_scale", anchor.Scale, "final", finalScale)

	created := 0
	for i, noteMap := range req.Notes {
		placed := applyAnchorScale(noteMap, anchor.X, anchor.Y, finalScale)
		note, err := c.CreateNote(ctx, req.CanvasID, placed)
		if err != nil {
			slog.Error("create-notes: create note", "index", i, "err", err)
			writeError(w, http.StatusInternalServerError, "failed to create note: "+err.Error())
			return
		}
		slog.Debug("create-notes: note created", "index", i, "note_id", note.ID)
		created++
	}

	slog.Info("create-notes: complete", "created", created)
	writeJSON(w, http.StatusOK, map[string]any{"status": "notes created", "count": created})
}

// SetCredentialsHandler handles POST /api/set-credentials.
// Accepts {mcsServer, apiKey} and persists them in-process.
func SetCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MCSServer string `json:"mcsServer"`
		APIKey    string `json:"apiKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	config.Set(&config.Config{
		CanvusURL: req.MCSServer,
		APIKey:    req.APIKey,
	})
	slog.Info("credentials updated", "server", req.MCSServer)
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

// GetCanvasesHandler handles GET /api/get-canvases.
func GetCanvasesHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	if cfg.CanvusURL == "" || cfg.APIKey == "" {
		writeError(w, http.StatusBadRequest, "Canvus credentials not set")
		return
	}
	ctx := context.Background()
	c := canvus.NewClient(cfg.CanvusURL, cfg.APIKey)
	canvases, err := c.ListCanvases(ctx)
	if err != nil {
		slog.Error("get-canvases: list", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch canvases: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"canvases": canvases})
}

// GetAnchorsOnlyHandler handles GET /api/get-anchors?canvasID=...
func GetAnchorsOnlyHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	if cfg.CanvusURL == "" || cfg.APIKey == "" {
		writeError(w, http.StatusBadRequest, "Canvus credentials not set")
		return
	}
	canvasID := r.URL.Query().Get("canvasID")
	if canvasID == "" {
		writeError(w, http.StatusBadRequest, "canvasID required")
		return
	}
	ctx := context.Background()
	c := canvus.NewClient(cfg.CanvusURL, cfg.APIKey)
	anchors, err := c.ListAnchors(ctx, canvasID)
	if err != nil {
		slog.Error("get-anchors: list", "canvas_id", canvasID, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch anchors: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"anchors": anchors})
}

// GetAnchorInfoHandler handles GET /api/get-anchor-info?canvasID=...&anchorID=...
func GetAnchorInfoHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	if cfg.CanvusURL == "" || cfg.APIKey == "" {
		writeError(w, http.StatusBadRequest, "Canvus credentials not set")
		return
	}
	canvasID := r.URL.Query().Get("canvasID")
	anchorID := r.URL.Query().Get("anchorID")
	if canvasID == "" || anchorID == "" {
		writeError(w, http.StatusBadRequest, "canvasID and anchorID required")
		return
	}
	ctx := context.Background()
	c := canvus.NewClient(cfg.CanvusURL, cfg.APIKey)
	anchor, err := c.GetAnchor(ctx, canvasID, anchorID)
	if err != nil {
		slog.Error("get-anchor-info", "canvas_id", canvasID, "anchor_id", anchorID, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch anchor info: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, anchor)
}

// --- helpers ---

// parseZoneParams reads zone dimensions, location, and scale from form values.
// Missing values fall back to sensible defaults (640×480 at origin, scale 1.0).
func parseZoneParams(r *http.Request) mapping.ZoneParams {
	zone := mapping.ZoneParams{Width: 640, Height: 480, Scale: 1.0}
	if v := r.FormValue("zoneDimensions"); v != "" {
		var dims [2]float64
		if err := json.Unmarshal([]byte(v), &dims); err == nil {
			zone.Width = dims[0]
			zone.Height = dims[1]
		}
	}
	if v := r.FormValue("zoneLocation"); v != "" {
		var loc [2]float64
		if err := json.Unmarshal([]byte(v), &loc); err == nil {
			zone.X = loc[0]
			zone.Y = loc[1]
		}
	}
	if v := r.FormValue("zoneScale"); v != "" {
		var s float64
		if err := json.Unmarshal([]byte(v), &s); err == nil {
			zone.Scale = s
		}
	}
	return zone
}

// applyAnchorScale scales and offsets a note map by the anchor position and
// combined scale factor, then sets scale=1 (all scaling absorbed into coords).
func applyAnchorScale(note map[string]any, anchorX, anchorY, finalScale float64) map[string]any {
	result := make(map[string]any, len(note))
	for k, v := range note {
		result[k] = v
	}
	if loc, ok := note["location"].(map[string]any); ok {
		x, _ := loc["x"].(float64)
		y, _ := loc["y"].(float64)
		result["location"] = map[string]any{
			"x": anchorX + x*finalScale,
			"y": anchorY + y*finalScale,
		}
	}
	if sz, ok := note["size"].(map[string]any); ok {
		w, _ := sz["width"].(float64)
		h, _ := sz["height"].(float64)
		result["size"] = map[string]any{
			"width":  w * finalScale,
			"height": h * finalScale,
		}
	}
	result["scale"] = 1.0
	return result
}

// writeJSON encodes v as JSON to w with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("writeJSON: encode", "err", err)
	}
}

// writeError sends a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
