// Package webui provides the WebUI molecule layer including the Manager, which
// wires together the HTTP server lifecycle, Fyne UI, and canvas service.
// This file is a compile stub — Task 9 will replace it with the full implementation.
package webui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/organisms/services"
)

// Manager handles WebUI integration and local server lifecycle.
// It is initialised from a Fyne window and exposes a single tab widget.
type Manager struct {
	fileService *services.FileService
}

// NewManager creates a new WebUI Manager.
// Full implementation will be provided in Task 9 (manager.go decomposition).
func NewManager(fileService *services.FileService) (*Manager, error) {
	return &Manager{fileService: fileService}, nil
}

// CreateUI returns the Fyne widget for the WebUI tab.
// Full implementation will be provided in Task 9.
func (m *Manager) CreateUI(_ fyne.Window) fyne.CanvasObject {
	return widget.NewLabel("WebUI — initialising…")
}
