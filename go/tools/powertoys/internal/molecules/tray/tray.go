//go:build windows
// +build windows

package tray

import (
	"bytes"
	"image/png"

	"fyne.io/fyne/v2"
	"github.com/getlantern/systray"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/assets"
)

// Manager handles system tray integration.
type Manager struct {
	window fyne.Window
	app    fyne.App
}

// NewManager creates a new system tray manager.
func NewManager(window fyne.Window, app fyne.App) *Manager {
	return &Manager{
		window: window,
		app:    app,
	}
}

// Setup initializes the system tray.
func (m *Manager) Setup() {
	// Hide window when close button (X) is clicked - this hides to tray
	// Note: Minimize button minimizes to taskbar (standard Windows behavior)
	m.window.SetCloseIntercept(func() {
		m.window.Hide()
	})

	// Start systray in background
	go func() {
		systray.Run(m.onReady, m.onExit)
	}()
}

func (m *Manager) onReady() {
	systray.SetTitle("Canvus PowerToys")
	systray.SetTooltip("Canvus PowerToys")

	// Load and set tray icon from embedded assets
	// Windows system tray expects ICO format, so we embed a pre-generated ICO file
	iconData, err := assets.Icons.ReadFile("CanvusPowerToysIcon.ico")
	if err != nil || len(iconData) == 0 {
		iconData, err = assets.Icons.ReadFile("icons/CanvusPowerToysIcon.ico")
	}
	if err == nil && len(iconData) > 0 {
		systray.SetIcon(iconData)
	} else {
		// Fallback to PNG in case ICO not found (some environments accept PNG)
		if pngData, err := assets.Icons.ReadFile("CanvusPowerToysIcon.png"); err == nil && len(pngData) > 0 {
			if _, err := png.Decode(bytes.NewReader(pngData)); err == nil {
				systray.SetIcon(pngData)
			}
		}
	}

	showItem := systray.AddMenuItem("Show", "Show window")
	systray.AddSeparator()
	quitItem := systray.AddMenuItem("Quit", "Exit application")

	go func() {
		for {
			select {
			case <-showItem.ClickedCh:
				fyne.Do(func() {
					m.window.Show()
					m.window.RequestFocus()
				})
			case <-quitItem.ClickedCh:
				systray.Quit()
				fyne.Do(func() {
					m.app.Quit()
				})
			}
		}
	}()
}

func (m *Manager) onExit() {
}
