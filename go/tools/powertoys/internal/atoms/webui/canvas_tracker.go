package webui

import (
	"sync"
)

// CanvasTracker manages the current canvas_id state.
type CanvasTracker struct {
	mu         sync.RWMutex
	canvasID   string
	canvasName string
}

// NewCanvasTracker creates a new canvas tracker.
func NewCanvasTracker() *CanvasTracker {
	return &CanvasTracker{}
}

// UpdateCanvas updates the current canvas_id and canvas_name.
func (ct *CanvasTracker) UpdateCanvas(canvasID, canvasName string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.canvasID = canvasID
	ct.canvasName = canvasName
}

// GetCanvas returns the current canvas_id and canvas_name.
func (ct *CanvasTracker) GetCanvas() (string, string) {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.canvasID, ct.canvasName
}

// GetCanvasID returns the current canvas_id.
func (ct *CanvasTracker) GetCanvasID() string {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.canvasID
}

// GetCanvasName returns the current canvas_name.
func (ct *CanvasTracker) GetCanvasName() string {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.canvasName
}
