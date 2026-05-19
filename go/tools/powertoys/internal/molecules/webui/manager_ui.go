// Package webui — Manager UI: struct definition, NewManager, CreateUI,
// and all Fyne widget event-handler helpers.
package webui

import (
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/organisms/services"
)

// Manager handles WebUI integration and local server.
type Manager struct {
	fileService       *services.FileService
	iniParser         *config.INIParser
	insecureTLS       bool
	server            *http.Server
	serverURL         *widget.Entry
	serverSelect      *widget.Select
	authToken         *widget.Entry
	serverPort        *widget.Entry
	serverStatus      *widget.Label
	enabledPages      map[string]*widget.Check
	selectAllPage     *widget.Check
	suppressSelectAll bool
	startStopBtn      *widget.Button
	canvasService     *CanvasService
	apiRoutes         *APIRoutes
	tokenInstructions *fyne.Container
	tokenLinkButton   *widget.Button
}

// NewManager creates a new WebUI Manager.
func NewManager(fileService *services.FileService) (*Manager, error) {
	return &Manager{
		fileService:       fileService,
		iniParser:         config.NewINIParser(),
		enabledPages:      make(map[string]*widget.Check),
		suppressSelectAll: false,
	}, nil
}

// CreateUI creates the UI for the WebUI tab.
func (m *Manager) CreateUI(window fyne.Window) fyne.CanvasObject {
	title := widget.NewLabel("WebUI Integration")
	title.TextStyle = fyne.TextStyle{Bold: true}

	savedConfig := m.loadSavedConfiguration()

	instructions := widget.NewRichTextFromMarkdown(`
**WebUI Integration**

Enable Canvus PowerToys to act as a web server for remote access and control.

**Configuration:**
- Canvus Server URL: Your Canvus server address
- User Auth Token: Access token from Canvus server profile
- WebUI Server Port: Port number for the local WebUI server (default: 8080)
- Enabled Pages: Select which WebUI pages to enable
`)

	serverURLLabel, authTokenLabel, serverPortLabel := m.initConfigFields(savedConfig)

	m.tokenInstructions = m.createTokenInstructions()
	m.serverStatus = widget.NewLabel("Server: Stopped")
	m.serverStatus.Importance = widget.LowImportance
	m.startStopBtn = widget.NewButton("Start Server", func() { m.toggleServer(window) })

	rightColumn := m.buildPageChecksPanel(window)

	titleBar := container.NewBorder(nil, nil, title, m.startStopBtn, nil)

	leftColumn := container.NewVBox(
		instructions,
		widget.NewSeparator(),
		container.NewGridWithColumns(2, serverURLLabel, container.NewVBox(m.serverSelect, m.serverURL)),
		container.NewGridWithColumns(2, authTokenLabel, m.authToken),
		container.NewGridWithColumns(2, serverPortLabel, m.serverPort),
		m.tokenInstructions,
		widget.NewSeparator(),
		m.serverStatus,
	)

	// Two column layout (no individual scrolling - they scroll together)
	twoColumnLayout := container.NewGridWithColumns(2,
		leftColumn,
		rightColumn,
	)

	// Main layout: Title bar on top, two columns below
	mainLayout := container.NewVBox(
		titleBar,
		widget.NewSeparator(),
		twoColumnLayout,
	)

	// Single scroll container for the entire layout
	return container.NewScroll(mainLayout)
}

// ensureHTTPS ensures the URL has https:// prefix if it's missing a protocol.
func ensureHTTPS(url string) string {
	url = strings.TrimSpace(url)
	if url == "" {
		return url
	}
	// Check if URL already has a protocol
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	// Add https:// prefix
	return "https://" + url
}

// initConfigFields initialises the server URL dropdown/entry, auth token entry,
// and port entry from the saved configuration and mt-canvus.ini. It populates
// the manager's widget fields as a side effect and returns the three label widgets
// needed for form layout in CreateUI.
func (m *Manager) initConfigFields(savedConfig *webUIConfiguration) (serverURLLabel, authTokenLabel, serverPortLabel *widget.Label) {
	serverURLLabel = widget.NewLabel("Canvus Server URL:")
	serverNames, serverURLs := m.loadServerNames()
	m.serverSelect = widget.NewSelect(serverNames, func(selected string) {
		if url, ok := serverURLs[selected]; ok {
			m.serverURL.SetText(ensureHTTPS(url))
			m.updateTokenInstructions()
		}
	})
	m.serverSelect.PlaceHolder = "Select server or type URL..."
	m.serverURL = widget.NewEntry()
	m.serverURL.SetPlaceHolder("https://your-canvus-server.com or select from dropdown")
	m.loadServerURL()
	if savedConfig != nil && savedConfig.ServerURL != "" {
		m.serverURL.SetText(savedConfig.ServerURL)
	}
	m.serverURL.OnChanged = func(_ string) { m.updateTokenInstructions() }
	if len(serverNames) > 0 {
		m.serverSelect.SetSelected(serverNames[0])
	}

	authTokenLabel = widget.NewLabel("User Auth Token:")
	m.authToken = widget.NewEntry()
	m.authToken.SetPlaceHolder("Paste your access token here")
	m.authToken.Password = true
	if savedConfig != nil && savedConfig.AuthToken != "" {
		m.authToken.SetText(savedConfig.AuthToken)
	}

	serverPortLabel = widget.NewLabel("WebUI Server Port:")
	m.serverPort = widget.NewEntry()
	m.serverPort.SetPlaceHolder("8080")
	m.serverPort.SetText("8080")
	if savedConfig != nil && savedConfig.ServerPort != "" {
		m.serverPort.SetText(savedConfig.ServerPort)
	}
	return
}

// loadServerNames loads server names from [server:<name>] sections in mt-canvus.ini.
func (m *Manager) loadServerNames() ([]string, map[string]string) {
	serverNames := []string{}
	serverURLs := make(map[string]string)

	iniPath := m.fileService.DetectMtCanvusIni()
	if iniPath == "" {
		return serverNames, serverURLs
	}

	iniFile, err := m.iniParser.Read(iniPath)
	if err != nil {
		return serverNames, serverURLs
	}

	// Find all [server:<name>] sections
	for _, section := range iniFile.Sections() {
		sectionName := section.Name()
		if strings.HasPrefix(sectionName, "server:") {
			serverName := strings.TrimPrefix(sectionName, "server:")
			if serverName != "" {
				// Get server URL from this section
				serverKey := section.Key("server")
				if serverKey != nil {
					serverURL := serverKey.String()
					if serverURL != "" {
						serverURL = ensureHTTPS(serverURL)
						serverNames = append(serverNames, serverName)
						serverURLs[serverName] = serverURL
					}
				}
			}
		}
	}

	return serverNames, serverURLs
}

// loadServerURL loads the server URL from mt-canvus.ini.
func (m *Manager) loadServerURL() {
	iniPath := m.fileService.DetectMtCanvusIni()
	if iniPath == "" {
		return
	}

	iniFile, err := m.iniParser.Read(iniPath)
	if err != nil {
		return
	}

	// Try to get server URL from [server] or [canvas] section
	sections := []string{"server", "canvas", ""}
	for _, sectionName := range sections {
		section, err := iniFile.GetSection(sectionName)
		if err != nil {
			continue
		}

		// Try common server URL keys
		keys := []string{"server-url", "url", "canvus-server", "server"}
		for _, key := range keys {
			if keyObj := section.Key(key); keyObj != nil {
				url := keyObj.String()
				if url != "" {
					// Ensure URL has https:// prefix
					url = ensureHTTPS(url)
					m.serverURL.SetText(url)
					return
				}
			}
		}
	}
}

// toggleServer starts or stops the local web server.
func (m *Manager) toggleServer(window fyne.Window) {
	if m.server == nil {
		// Start server
		m.startServer(window)
	} else {
		// Stop server
		m.stopServer()
	}
}

func (m *Manager) syncSelectAllFromChecks() {
	if m.selectAllPage == nil {
		return
	}

	allSelected := true
	for _, check := range m.enabledPages {
		if check == nil || !check.Checked {
			allSelected = false
			break
		}
	}
	m.setSelectAllState(allSelected)
}

func (m *Manager) setSelectAllState(checked bool) {
	if m.selectAllPage == nil {
		return
	}
	m.suppressSelectAll = true
	m.selectAllPage.SetChecked(checked)
	m.suppressSelectAll = false
}

// createTokenInstructions creates the token instructions widget with dynamic server URL.
func (m *Manager) createTokenInstructions() *fyne.Container {
	title := widget.NewRichTextFromMarkdown("**How to get your Auth Token:**")

	step1 := widget.NewLabel("1. Log in to your Canvus server")
	step2Label := widget.NewLabel("2. Navigate to:")
	step3 := widget.NewLabel("3. Create a new access token")
	step4 := widget.NewLabel("4. Copy and paste it below")

	// Create clickable link button (will be updated with actual URL)
	linkButton := widget.NewButton("", func() {
		serverURL := m.getServerURLForTokenLink()
		if serverURL != "" {
			m.openURL(serverURL)
		}
	})
	linkButton.Importance = widget.LowImportance

	// Container for step 2 with label and link
	step2Container := container.NewHBox(step2Label, linkButton)

	// Store link button reference for updates
	m.tokenLinkButton = linkButton

	// Update link initially
	m.updateTokenInstructions()

	return container.NewVBox(
		title,
		step1,
		step2Container,
		step3,
		step4,
	)
}

// updateTokenInstructions updates the token instructions link with the current server URL.
func (m *Manager) updateTokenInstructions() {
	if m.tokenLinkButton == nil {
		return
	}

	serverURL := m.getServerURLForTokenLink()
	if serverURL != "" {
		m.tokenLinkButton.SetText(serverURL)
		m.tokenLinkButton.OnTapped = func() {
			m.openURL(serverURL)
		}
	} else {
		m.tokenLinkButton.SetText("(Enter server URL above)")
		m.tokenLinkButton.OnTapped = nil
	}
}

// getServerURLForTokenLink gets the server URL for the token link, ensuring it has the profile path.
func (m *Manager) getServerURLForTokenLink() string {
	serverURL := strings.TrimSpace(m.serverURL.Text)
	if serverURL == "" {
		return ""
	}

	// Ensure URL has https:// prefix
	serverURL = ensureHTTPS(serverURL)

	// Remove trailing slash
	serverURL = strings.TrimSuffix(serverURL, "/")

	// Add profile/access-tokens path
	return serverURL + "/profile/access-tokens"
}

// openURL opens a URL in the default browser (cross-platform).
func (m *Manager) openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // Linux and others
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Run() // Ignore errors - if browser doesn't open, user can copy URL
}

// buildPageChecksPanel builds the right-column panel containing the pages checkboxes
// and the Save/Test buttons. It populates m.enabledPages and m.selectAllPage as a
// side effect so that the manager can read checkbox state when persisting config.
func (m *Manager) buildPageChecksPanel(window fyne.Window) *fyne.Container {
	pagesLabel := widget.NewLabel("Enabled Pages:")
	pagesLabel.Importance = widget.LowImportance
	comingSoonNote := widget.NewLabel("! Page selection is coming soon")
	comingSoonNote.Importance = widget.DangerImportance
	comingSoonNote.Wrapping = fyne.TextWrapWord

	// All page checkboxes are locked to checked; page-selection is not yet implemented.
	pageOptions := []string{"Main", "Pages", "Macros", "Remote Upload", "RCU"}
	m.selectAllPage = widget.NewCheck("Select All", func(_ bool) { m.selectAllPage.SetChecked(true) })
	m.selectAllPage.SetChecked(true)
	pageChecks := []fyne.CanvasObject{m.selectAllPage, widget.NewSeparator()}
	for _, page := range pageOptions {
		pageName := page
		check := widget.NewCheck(page, func(_ bool) {
			if storedCheck, ok := m.enabledPages[pageName]; ok {
				storedCheck.SetChecked(true)
			}
		})
		m.enabledPages[page] = check
		check.SetChecked(true)
		pageChecks = append(pageChecks, check)
	}
	m.syncSelectAllFromChecks()

	saveConfigBtn := widget.NewButton("Save Configuration", func() { m.saveConfiguration(window) })
	testConnectionBtn := widget.NewButton("Test Connection", func() { m.testConnection(window) })
	return container.NewVBox(
		comingSoonNote,
		pagesLabel,
		container.NewVBox(pageChecks...),
		widget.NewSeparator(),
		container.NewVBox(saveConfigBtn, testConnectionBtn),
	)
}
