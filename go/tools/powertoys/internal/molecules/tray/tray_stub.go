//go:build !windows
// +build !windows

package tray

import "fyne.io/fyne/v2"

// Manager handles system tray integration (stub for non-Windows platforms).
type Manager struct {
	window fyne.Window
	app    fyne.App
}

// NewManager creates a new system tray manager (stub).
func NewManager(window fyne.Window, app fyne.App) *Manager {
	return &Manager{
		window: window,
		app:    app,
	}
}

// Setup initializes the system tray (stub - no-op on non-Windows).
func (m *Manager) Setup() {
	// System tray not supported on this platform
	// Just handle window close to quit
	m.window.SetCloseIntercept(func() {
		m.app.Quit()
	})
}

