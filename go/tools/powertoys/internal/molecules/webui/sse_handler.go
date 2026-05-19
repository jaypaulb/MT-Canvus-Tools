package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/logger"
)

// SSEHandler handles Server-Sent Events for canvas_id updates.
type SSEHandler struct {
	canvasService *CanvasService
}

// NewSSEHandler creates a new SSE handler.
func NewSSEHandler(canvasService *CanvasService) *SSEHandler {
	return &SSEHandler{
		canvasService: canvasService,
	}
}

// HandleSubscribe handles the SSE subscription endpoint.
func (h *SSEHandler) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Get request context - this will be cancelled when server shuts down
	// Explicitly type the context to ensure the import is recognized
	var ctx context.Context = r.Context()

	// Send initial canvas state
	h.sendCanvasUpdate(w, h.canvasService.GetCanvasID(), h.canvasService.GetCanvasName())

	// Use a shorter ticker interval (1 second) for more responsive shutdown detection
	// This ensures we check for context cancellation more frequently
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Track last sent canvas_id and canvas_name to detect changes
	lastCanvasID := h.canvasService.GetCanvasID()
	lastCanvasName := h.canvasService.GetCanvasName()

	for {
		// Use a context-aware select that prioritizes context cancellation
		// Context cancellation has highest priority - will exit immediately when server shuts down
		select {
		case <-ctx.Done():
			// Client disconnected or server shutting down
			logger.Logf("[SSEHandler] Connection closed (client disconnected or server shutdown)\n")
			return
		case <-ticker.C:
			// Check context before processing (server might have shut down during tick)
			if ctx.Err() != nil {
				logger.Logf("[SSEHandler] Context cancelled during tick, closing connection\n")
				return
			}

			// Periodic update check - send update if canvas_id OR canvas_name changed
			currentCanvasID := h.canvasService.GetCanvasID()
			currentCanvasName := h.canvasService.GetCanvasName()

			if currentCanvasID != lastCanvasID || currentCanvasName != lastCanvasName {
				// Check context again before sending (write might block)
				if ctx.Err() != nil {
					logger.Logf("[SSEHandler] Context cancelled before sending update, closing connection\n")
					return
				}
				h.sendCanvasUpdate(w, currentCanvasID, currentCanvasName)
				lastCanvasID = currentCanvasID
				lastCanvasName = currentCanvasName
			} else {
				// Check context before sending keepalive
				if ctx.Err() != nil {
					logger.Logf("[SSEHandler] Context cancelled before sending keepalive, closing connection\n")
					return
				}
				// Send keepalive
				h.sendKeepalive(w)
			}
		}
	}
}

// sendCanvasUpdate sends a canvas update event.
func (h *SSEHandler) sendCanvasUpdate(w http.ResponseWriter, canvasID, canvasName string) {
	event := map[string]interface{}{
		"canvas_id":   canvasID,
		"canvas_name": canvasName,
		"client_name": h.canvasService.GetClientName(),
		"client_id":   h.canvasService.GetClientID(),
		"timestamp":   time.Now().Unix(),
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		logger.Logf("Error marshaling canvas event: %v\n", err)
		return
	}

	// Send SSE formatted event - check for write errors
	if _, err := fmt.Fprintf(w, "event: canvas_update\n"); err != nil {
		logger.Logf("Error writing SSE event header: %v\n", err)
		return
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(eventJSON)); err != nil {
		logger.Logf("Error writing SSE event data: %v\n", err)
		return
	}

	// Flush to ensure data is sent immediately
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// sendKeepalive sends a keepalive comment to maintain connection.
func (h *SSEHandler) sendKeepalive(w http.ResponseWriter) {
	if _, err := fmt.Fprintf(w, ": keepalive\n\n"); err != nil {
		// Connection likely closed, will be detected by clientGone channel
		return
	}
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

