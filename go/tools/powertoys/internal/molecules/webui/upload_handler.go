package webui

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	webuiatoms "github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/webui"
)

// UploadHandler handles remote content upload API endpoints.
type UploadHandler struct {
	apiClient     *webuiatoms.APIClient
	canvasService *CanvasService
	uploadDir     string
}

// NewUploadHandler creates a new upload handler.
func NewUploadHandler(apiClient *webuiatoms.APIClient, canvasService *CanvasService, uploadDir string) *UploadHandler {
	// Ensure upload directory exists
	os.MkdirAll(uploadDir, 0755)

	return &UploadHandler{
		apiClient:     apiClient,
		canvasService: canvasService,
		uploadDir:     uploadDir,
	}
}

// HandleUpload handles POST /api/remote-upload - Upload files.
func (h *UploadHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (32MB max)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		sendErrorResponse(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		sendErrorResponse(w, "No files provided", http.StatusBadRequest)
		return
	}

	uploadPath := r.FormValue("path")
	if uploadPath == "" {
		uploadPath = "/"
	}

	canvasID := h.canvasService.GetCanvasID()
	if canvasID == "" {
		sendErrorResponse(w, "Canvas not available", http.StatusServiceUnavailable)
		return
	}

	var uploadedFiles []map[string]interface{}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}

		// Save file locally first
		filename := filepath.Join(h.uploadDir, fileHeader.Filename)
		dst, err := os.Create(filename)
		if err != nil {
			file.Close()
			continue
		}

		io.Copy(dst, file)
		file.Close()
		dst.Close()

		// Upload to Canvus API
		// TODO: Actually upload file to Canvus API
		// This would require multipart form upload to the API
		uploadedFiles = append(uploadedFiles, map[string]interface{}{
			"name": fileHeader.Filename,
			"size": fileHeader.Size,
			"path": uploadPath,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"files":   uploadedFiles,
	})
}

// HandleHistory handles GET /api/remote-upload/history - Get upload history.
func (h *UploadHandler) HandleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read upload history from local storage or API
	// For now, return empty list
	history := []map[string]interface{}{}

	// TODO: Implement actual history tracking
	// Could read from a local database or fetch from Canvus API

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(history)
}

// UploadRecord represents an upload history record.
type UploadRecord struct {
	Filename   string    `json:"filename"`
	Size       int64     `json:"size"`
	Path       string    `json:"path"`
	UploadedAt time.Time `json:"uploaded_at"`
}

