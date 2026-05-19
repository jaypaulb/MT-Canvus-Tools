// Package webui — Manager Lifecycle: startServer, stopServer,
// server-started dialog, and canvas browser widget creation.
package webui

import (
	"context"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/logger"
	webuiatoms "github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/webui"
)

// startServer starts the local web server.
func (m *Manager) startServer(window fyne.Window) {
	port := m.serverPort.Text
	if port == "" {
		port = "8080"
	}

	// Validate port
	if port == "" {
		dialog.ShowError(fmt.Errorf("Port cannot be empty. Please enter a valid port number (e.g., 8080)"), window)
		return
	}

	// Check if port is already in use by checking if server is already running
	if m.server != nil {
		dialog.ShowError(fmt.Errorf("Server is already running. Please stop it first."), window)
		return
	}

	// Check if port is already in use
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Port %s is already in use. Please choose a different port or stop the process using it.", port), window)
		return
	}
	listener.Close() // Close immediately, we'll open it again in the server

	// Get server URL and auth token
	serverURL := m.serverURL.Text
	authToken := m.authToken.Text

	if serverURL == "" {
		dialog.ShowError(fmt.Errorf("Server URL cannot be empty"), window)
		return
	}

	if authToken == "" {
		dialog.ShowError(fmt.Errorf("Auth token cannot be empty"), window)
		return
	}

	if err := m.persistConfiguration(); err != nil {
		dialog.ShowError(fmt.Errorf("failed to save configuration: %w", err), window)
		return
	}

	// Normalize server URL
	apiBaseURL := strings.TrimSuffix(serverURL, "/")
	apiBaseURL = strings.TrimSuffix(apiBaseURL, "/api/v1")
	apiBaseURL = strings.TrimSuffix(apiBaseURL, "/api")

	// Create API client
	apiClient, err := webuiatoms.NewAPIClient(apiBaseURL, authToken, m.insecureTLS)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to create API client: %w", err), window)
		return
	}

	// Create canvas service using the SDK session from the API client
	canvasService := NewCanvasService(m.fileService, apiClient.Session())

	// Try to start canvas service, but don't fail if it doesn't work.
	// User can override client selection in WebUI.
	if err := canvasService.Start(); err != nil {
		logger.Logf("[WebUI] Canvas service auto-start failed: %v (user can override in WebUI)", err)
		// Don't show error dialog - just log it and continue.
		// The WebUI will load and user can manually override.
	}

	// Create API routes (uploadDir can be empty for now)
	apiRoutes := NewAPIRoutes(canvasService, apiClient, "")

	// Store references
	m.canvasService = canvasService
	m.apiRoutes = apiRoutes

	mux := http.NewServeMux()

	// Register API routes first (before static handler).
	// This ensures more specific routes like /api/* take precedence over catch-all /.
	apiRoutes.RegisterRoutes(mux)
	logger.Logf("[WebUI] API routes registered")

	// Use StaticHandler to serve actual WebUI pages (not placeholder pages).
	staticHandler := NewStaticHandler()
	staticHandler.ServeFiles(mux)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Debug endpoint to list embedded files
	mux.HandleFunc("/debug/files", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		dbgHandler := NewStaticHandler()

		var listFiles func(fs.FS, string, int) string
		listFiles = func(fsys fs.FS, path string, depth int) string {
			if depth > 5 {
				return ""
			}
			result := ""
			entries, err := fs.ReadDir(fsys, path)
			if err != nil {
				return fmt.Sprintf("Error reading %s: %v\n", path, err)
			}
			for _, entry := range entries {
				indent := strings.Repeat("  ", depth)
				result += fmt.Sprintf("%s%s\n", indent, entry.Name())
				if entry.IsDir() {
					subFS, _ := fs.Sub(fsys, path)
					result += listFiles(subFS, entry.Name(), depth+1)
				}
			}
			return result
		}

		fileList := "Embedded filesystem contents:\n"
		fileList += listFiles(dbgHandler.fileSystem, ".", 0)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fileList))
	})

	m.server = &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logf("Server error: %v", err)
			// Don't update UI from goroutine - port check already happens before starting.
			// If we get here, it's an unexpected error - just log it.
			// The server will be nil and user can try starting again.
			m.server = nil
		}
	}()

	// Give server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Update UI
	serverURLStr := fmt.Sprintf("http://localhost:%s", port)
	m.serverStatus.SetText(fmt.Sprintf("Server: Running on %s", serverURLStr))
	m.serverStatus.Importance = widget.SuccessImportance
	m.startStopBtn.SetText("Stop Server")

	localTestResult, remoteTestResult, localTestSuccess, remoteTestSuccess := m.performConnectionTests(port, serverURL, authToken)
	m.updateStatusFromTestResults(localTestSuccess, remoteTestSuccess)
	m.showServerStartedDialog(serverURLStr, localTestResult, remoteTestResult, window)
}

// stopServer stops the local web server.
func (m *Manager) stopServer() {
	if m.server == nil {
		return
	}

	// Stop canvas service first to stop workspace subscriptions
	if m.canvasService != nil {
		m.canvasService.Stop()
		m.canvasService = nil
	}

	// Reduced timeout to 5 seconds since SSE handler now checks context every 1 second.
	// This should be sufficient for graceful shutdown of all connections.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.server.Shutdown(ctx); err != nil {
		// Log error but don't fail - server will still stop
		if err == context.DeadlineExceeded {
			logger.Logf("Server shutdown: Some connections did not close within timeout, forcing close")
		} else {
			logger.Logf("Server shutdown error: %v", err)
		}
		// Force close if graceful shutdown failed
		if m.server != nil {
			m.server.Close()
		}
	} else {
		logger.Logf("Server shutdown: All connections closed gracefully")
	}

	m.server = nil
	m.apiRoutes = nil
	m.serverStatus.SetText("Server: Stopped")
	m.serverStatus.Importance = widget.LowImportance
	m.startStopBtn.SetText("Start Server")
}

func (m *Manager) showServerStartedDialog(serverURL string, localResult, remoteResult string, window fyne.Window) {
	// Create browser widget on canvas with specific size and position
	m.createBrowserWidgetOnCanvas(serverURL, 1024, 768, 4800, 2700)

	resultsLabel := widget.NewLabel(fmt.Sprintf("%s\n\n%s", localResult, remoteResult))
	resultsLabel.Wrapping = fyne.TextWrapWord

	noteLabel := widget.NewLabel("Note: Check the center of the canvas for the WebUI window")
	noteLabel.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("WebUI server is running on %s", serverURL)),
		widget.NewSeparator(),
		widget.NewLabel("Connection Test Results:"),
		resultsLabel,
		widget.NewSeparator(),
		noteLabel,
	)

	dialog.NewCustomConfirm(
		"Server Started & Tested",
		"Close to System Tray",
		"Dismiss",
		content,
		func(closeToTray bool) {
			if closeToTray {
				window.Hide()
			}
		},
		window,
	).Show()
}

// createBrowserWidgetOnCanvas creates a browser widget on the Canvus canvas via MTCS API.
// Size: width x height, Position: x, y coordinates
func (m *Manager) createBrowserWidgetOnCanvas(url string, width, height, posX, posY int) {
	if m.canvasService == nil {
		logger.Logf("[createBrowserWidgetOnCanvas] Canvas service not available")
		return
	}

	canvasID := m.canvasService.GetCanvasID()
	if canvasID == "" {
		logger.Logf("[createBrowserWidgetOnCanvas] Canvas ID not available")
		return
	}

	// Get server URL and auth token from manager
	serverURL := m.serverURL.Text
	authToken := m.authToken.Text

	if serverURL == "" || authToken == "" {
		logger.Logf("[createBrowserWidgetOnCanvas] Server URL or auth token not available")
		return
	}

	// Normalize server URL (same as in startServer)
	apiBaseURL := strings.TrimSuffix(serverURL, "/")
	apiBaseURL = strings.TrimSuffix(apiBaseURL, "/api/v1")
	apiBaseURL = strings.TrimSuffix(apiBaseURL, "/api")

	// Create API client
	apiClient, err := webuiatoms.NewAPIClient(apiBaseURL, authToken, m.insecureTLS)
	if err != nil {
		logger.Logf("[createBrowserWidgetOnCanvas] Failed to create API client: %v", err)
		return
	}

	// Create browser widget payload
	payload := map[string]interface{}{
		"widget_type": "Browser",
		"url":         url,
		"location": map[string]float64{
			"x": float64(posX),
			"y": float64(posY),
		},
		"size": map[string]float64{
			"width":  float64(width),
			"height": float64(height),
		},
	}

	// POST to /canvases/:id/browsers to create browser widget (widgets endpoint is read-only)
	endpoint := fmt.Sprintf("/api/v1/canvases/%s/browsers", canvasID)
	logger.Logf("[createBrowserWidgetOnCanvas] Creating browser widget at (%d, %d) with size %dx%d, URL: %s", posX, posY, width, height, url)

	response, err := apiClient.Post(endpoint, payload)
	if err != nil {
		logger.Logf("[createBrowserWidgetOnCanvas] ERROR: Failed to create browser widget: %v", err)
		return
	}

	logger.Logf("[createBrowserWidgetOnCanvas] Successfully created browser widget: %s", string(response))
}
