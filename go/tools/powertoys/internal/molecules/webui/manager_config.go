// Package webui — Manager Config: webUIConfiguration type, load/save helpers,
// testConnection, performFullConnectionTest, and connection status helpers.
package webui

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	webuiatoms "github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/webui"
)

// webUIConfiguration holds persisted WebUI settings.
type webUIConfiguration struct {
	ServerURL    string          `json:"server_url"`
	AuthToken    string          `json:"auth_token"`
	ServerPort   string          `json:"server_port"`
	EnabledPages map[string]bool `json:"enabled_pages"`
}

// saveConfiguration saves the WebUI configuration and shows a success dialog.
func (m *Manager) saveConfiguration(window fyne.Window) {
	if err := m.persistConfiguration(); err != nil {
		dialog.ShowError(err, window)
		return
	}

	dialog.ShowInformation("Saved", "Configuration saved successfully", window)
}

// testConnection tests the full API flow: clients → client → workspaces → canvas.
func (m *Manager) testConnection(window fyne.Window) {
	port := m.serverPort.Text
	if port == "" {
		port = "8080"
	}

	serverURL := m.serverURL.Text
	authToken := m.authToken.Text

	// Validate inputs
	if port == "" {
		dialog.ShowError(fmt.Errorf("Port cannot be empty"), window)
		return
	}

	if serverURL == "" {
		dialog.ShowError(fmt.Errorf("Canvus Server URL cannot be empty"), window)
		return
	}

	if authToken == "" {
		dialog.ShowError(fmt.Errorf("Auth token cannot be empty"), window)
		return
	}

	// Perform the full connection test flow
	testResult, testSuccess, clientName, canvasName := m.performFullConnectionTest(serverURL, authToken)

	if testSuccess {
		m.serverStatus.SetText(fmt.Sprintf("Status: Ready to start server using %s : %s", clientName, canvasName))
		m.serverStatus.Importance = widget.SuccessImportance
		dialog.ShowInformation("Connection Test Results",
			fmt.Sprintf("✅ Connection successful!\n\n%s\n\nStatus: Ready to start server using %s : %s", testResult, clientName, canvasName),
			window)
	} else {
		m.serverStatus.SetText("Connection test failed")
		m.serverStatus.Importance = widget.DangerImportance
		dialog.ShowError(fmt.Errorf("❌ Connection test failed\n\n%s", testResult), window)
	}
}

// performFullConnectionTest performs the full API flow test:
// 1. Get clients list
// 2. Find the relevant client by installation_name
// 3. Get workspaces/0/ to find canvas ID
// 4. Get /canvases/canvasID/ to get canvas name
// Returns: (result message, success, client name, canvas name)
func (m *Manager) performFullConnectionTest(serverURL, authToken string) (string, bool, string, string) {
	var resultMessages []string

	// Normalize server URL
	baseURL := strings.TrimSuffix(serverURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/api/v1")
	baseURL = strings.TrimSuffix(baseURL, "/api")

	// Create API client
	apiClient, err := webuiatoms.NewAPIClient(baseURL, authToken, m.insecureTLS)
	if err != nil {
		return fmt.Sprintf("❌ Failed to create API client\n   Error: %v", err), false, "", ""
	}

	ctx := context.Background()

	// Step 1: Get clients list
	m.serverStatus.SetText("Step 1/4: Getting clients list...")
	m.serverStatus.Importance = widget.MediumImportance

	clients, err := apiClient.GetClients(ctx)
	if err != nil {
		return fmt.Sprintf("❌ Failed to get clients list\n   Error: %v", err), false, "", ""
	}
	resultMessages = append(resultMessages, fmt.Sprintf("✅ Step 1: Retrieved %d clients", len(clients)))

	if len(clients) == 0 {
		return fmt.Sprintf("❌ No clients found on server\n   %s", strings.Join(resultMessages, "\n")), false, "", ""
	}

	// Step 2: Find the relevant client by installation_name
	m.serverStatus.SetText("Step 2/4: Finding client by installation name...")
	m.serverStatus.Importance = widget.MediumImportance

	// Get installation name from mt-canvus.ini or device name
	clientResolver := webuiatoms.NewClientResolver(m.fileService)
	installationName, err := clientResolver.GetInstallationName()
	if err != nil {
		return fmt.Sprintf("❌ Failed to get installation name\n   Error: %v\n   %s", err, strings.Join(resultMessages, "\n")), false, "", ""
	}

	// Find matching client
	var matchedClientID string
	var matchedClientName string
	for i := range clients {
		if clients[i].InstallationName == installationName {
			matchedClientID = clients[i].ID
			matchedClientName = clients[i].InstallationName
			break
		}
	}

	if matchedClientID == "" {
		// List available clients for debugging
		var clientList []string
		for _, c := range clients {
			clientList = append(clientList, fmt.Sprintf("  - %s", c.InstallationName))
		}
		return fmt.Sprintf("❌ Client not found with installation_name: '%s'\n   Available clients:\n%s\n   %s",
			installationName, strings.Join(clientList, "\n"), strings.Join(resultMessages, "\n")), false, "", ""
	}

	clientName := matchedClientName
	clientID := matchedClientID
	resultMessages = append(resultMessages, fmt.Sprintf("✅ Step 2: Found client '%s' (ID: %s)", clientName, clientID))

	// Step 3: Get workspaces/0/ to find canvas ID
	m.serverStatus.SetText("Step 3/4: Getting workspace canvas...")
	m.serverStatus.Importance = widget.MediumImportance

	workspaceData, err := apiClient.Get(fmt.Sprintf("/api/v1/clients/%s/workspaces/0", clientID))
	if err != nil {
		return fmt.Sprintf("❌ Failed to get workspace\n   Error: %v\n   %s", err, strings.Join(resultMessages, "\n")), false, clientName, ""
	}

	// Parse workspace response to get canvas_id
	var workspace struct {
		CanvasID string `json:"canvas_id"`
	}
	if err := json.Unmarshal(workspaceData, &workspace); err != nil {
		return fmt.Sprintf("❌ Failed to parse workspace response\n   Error: %v\n   %s", err, strings.Join(resultMessages, "\n")), false, clientName, ""
	}

	if workspace.CanvasID == "" {
		return fmt.Sprintf("❌ No canvas_id found in workspace\n   %s", strings.Join(resultMessages, "\n")), false, clientName, ""
	}

	canvasID := workspace.CanvasID
	resultMessages = append(resultMessages, fmt.Sprintf("✅ Step 3: Found canvas ID: %s", canvasID))

	// Step 4: Get /canvases/canvasID/ to get canvas name
	m.serverStatus.SetText("Step 4/4: Getting canvas details...")
	m.serverStatus.Importance = widget.MediumImportance

	canvasData, err := apiClient.Get(fmt.Sprintf("/api/v1/canvases/%s", canvasID))
	if err != nil {
		return fmt.Sprintf("❌ Failed to get canvas details\n   Error: %v\n   %s", err, strings.Join(resultMessages, "\n")), false, clientName, ""
	}

	// Parse canvas response to get name
	var canvas struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(canvasData, &canvas); err != nil {
		return fmt.Sprintf("❌ Failed to parse canvas response\n   Error: %v\n   %s", err, strings.Join(resultMessages, "\n")), false, clientName, ""
	}

	canvasName := canvas.Name
	if canvasName == "" {
		canvasName = canvasID // Fallback to ID if name is empty
	}
	resultMessages = append(resultMessages, fmt.Sprintf("✅ Step 4: Canvas name: '%s'", canvasName))

	// All steps successful
	return strings.Join(resultMessages, "\n"), true, clientName, canvasName
}

func (m *Manager) getWebUIConfigPath() string {
	if m.fileService == nil {
		return ""
	}
	return filepath.Join(m.fileService.GetUserConfigPath(), "CanvusPowerToys", "webui_config.json")
}

func (m *Manager) loadSavedConfiguration() *webUIConfiguration {
	configPath := m.getWebUIConfigPath()
	if configPath == "" {
		return nil
	}

	var cfg webUIConfiguration
	if err := m.fileService.ReadJSONFile(configPath, &cfg); err != nil {
		return nil
	}

	if cfg.ServerPort == "" {
		cfg.ServerPort = "8080"
	}

	if cfg.EnabledPages == nil {
		cfg.EnabledPages = make(map[string]bool)
	}

	return &cfg
}

func (m *Manager) persistConfiguration() error {
	if m.serverURL == nil || m.authToken == nil || m.serverPort == nil {
		return fmt.Errorf("configuration inputs are not initialized")
	}

	serverURL := strings.TrimSpace(m.serverURL.Text)
	if serverURL == "" {
		return fmt.Errorf("Server URL cannot be empty")
	}

	authToken := strings.TrimSpace(m.authToken.Text)
	if authToken == "" {
		return fmt.Errorf("Auth token cannot be empty")
	}

	port := strings.TrimSpace(m.serverPort.Text)
	if port == "" {
		port = "8080"
	}

	if m.fileService == nil {
		return fmt.Errorf("file service not available")
	}

	configPath := m.getWebUIConfigPath()
	if configPath == "" {
		return fmt.Errorf("unable to determine configuration path")
	}

	cfg := &webUIConfiguration{
		ServerURL:    ensureHTTPS(serverURL),
		AuthToken:    authToken,
		ServerPort:   port,
		EnabledPages: make(map[string]bool),
	}

	for page, check := range m.enabledPages {
		if check != nil {
			cfg.EnabledPages[page] = check.Checked
		}
	}

	return m.fileService.WriteJSONFile(configPath, cfg)
}

// performConnectionTests checks the local WebUI server health endpoint and the remote
// Canvus server clients endpoint, returning a summary string and success flags for each.
func (m *Manager) performConnectionTests(port, serverURL, authToken string) (string, string, bool, bool) {
	// Create HTTP client with TLS verification disabled for self-signed certs
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	var localTestResult, remoteTestResult string
	var localTestSuccess, remoteTestSuccess bool

	m.serverStatus.SetText("Testing local WebUI server...")
	m.serverStatus.Importance = widget.MediumImportance

	m.serverMu.Lock()
	serverNil := m.server == nil
	m.serverMu.Unlock()

	if serverNil {
		localTestResult = "❌ Local WebUI server is not running\n   Please start the server first."
		localTestSuccess = false
	} else {
		localTestURL := fmt.Sprintf("http://localhost:%s/health", port)
		req, err := http.NewRequest("GET", localTestURL, nil)
		if err != nil {
			localTestResult = fmt.Sprintf("❌ Failed to create request: %v\n   URL: %s", err, localTestURL)
			localTestSuccess = false
		} else {
			resp, err := client.Do(req)
			if err != nil {
				localTestResult = fmt.Sprintf("❌ Cannot connect to local server on port %s\n   Error: %v\n   URL: %s\n   Make sure the server is running and the port is correct.", port, err, localTestURL)
				localTestSuccess = false
			} else {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					localTestResult = fmt.Sprintf("✅ Local WebUI server responding\n   URL: %s\n   Status: OK", localTestURL)
					localTestSuccess = true
				} else {
					localTestResult = fmt.Sprintf("❌ Server returned HTTP %d\n   URL: %s\n   Server may be running but not responding correctly.", resp.StatusCode, localTestURL)
					localTestSuccess = false
				}
			}
		}
	}

	m.serverStatus.SetText("Testing connection to Canvus server...")
	m.serverStatus.Importance = widget.MediumImportance

	baseURL := strings.TrimSuffix(serverURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/api/v1")
	baseURL = strings.TrimSuffix(baseURL, "/api")

	remoteTestURL := fmt.Sprintf("%s/api/v1/clients", baseURL)
	req, err := http.NewRequest("GET", remoteTestURL, nil)
	if err != nil {
		remoteTestResult = fmt.Sprintf("❌ Failed to create request: %v\n   URL: %s", err, remoteTestURL)
		remoteTestSuccess = false
	} else {
		req.Header.Set("Private-Token", authToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			remoteTestResult = fmt.Sprintf("❌ Cannot connect to Canvus server\n   Error: %v\n   URL: %s\n   Check your server URL and network connection.", err, remoteTestURL)
			remoteTestSuccess = false
		} else {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				remoteTestResult = fmt.Sprintf("✅ Canvus server connection successful\n   URL: %s\n   Status: OK", remoteTestURL)
				remoteTestSuccess = true
			} else if resp.StatusCode == http.StatusUnauthorized {
				remoteTestResult = fmt.Sprintf("⚠️  Server reachable but authentication failed\n   URL: %s\n   HTTP Status: %d\n   Please check your auth token.", remoteTestURL, resp.StatusCode)
				remoteTestSuccess = false
			} else {
				remoteTestResult = fmt.Sprintf("❌ Server returned HTTP %d\n   URL: %s\n   Server may be reachable but endpoint not available.", resp.StatusCode, remoteTestURL)
				remoteTestSuccess = false
			}
		}
	}

	return localTestResult, remoteTestResult, localTestSuccess, remoteTestSuccess
}

// updateStatusFromTestResults updates the server status label based on test results.
func (m *Manager) updateStatusFromTestResults(localSuccess, remoteSuccess bool) {
	if localSuccess && remoteSuccess {
		m.serverStatus.SetText("All tests passed")
		m.serverStatus.Importance = widget.SuccessImportance
		return
	}

	if localSuccess || remoteSuccess {
		m.serverStatus.SetText("Partial success - see details")
		m.serverStatus.Importance = widget.WarningImportance
		return
	}

	m.serverStatus.SetText("All tests failed")
	m.serverStatus.Importance = widget.DangerImportance
}
