package cssoptions

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gopkg.in/ini.v1"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/backup"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/organisms/services"
)

// PluginManifest represents a .canvusplugin manifest file.
type PluginManifest struct {
	APIVersion  string `json:"api-version"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// Manager handles CSS options and plugin generation.
type Manager struct {
	iniParser            *config.INIParser
	fileService          *services.FileService
	backupManager        *backup.Manager
	movingEnabled        *widget.Check
	scalingEnabled       *widget.Check
	rotationEnabled      *widget.Check
	videoLoopEnabled     *widget.Check
	kioskModeEnabled     *widget.Check
	kioskPlusEnabled     *widget.Check
	hideTitleBarsEnabled *widget.Check
	hideResizeHandlesEnabled *widget.Check
	hideSidebarEnabled   *widget.Check
	hideMainMenuEnabled  *widget.Check
	hideFingerMenuEnabled *widget.Check
	hideConnectorBulletsEnabled *widget.Check
	hideConnectorBulletsPresentationEnabled *widget.Check
	standbyImageEnabled   *widget.Check
	standbyImagePath       string
	standbyImageLabel      *widget.Label
	statusLabel            *widget.Label
}

// NewManager creates a new CSS Options Manager.
func NewManager(fileService *services.FileService) (*Manager, error) {
	return &Manager{
		iniParser:     config.NewINIParser(),
		fileService:   fileService,
		backupManager: backup.NewManager(""),
	}, nil
}

// CreateUI creates the UI for the CSS Options Manager tab.
func (m *Manager) CreateUI(window fyne.Window) fyne.CanvasObject {
	title := widget.NewLabel("CSS Options Manager")
	title.TextStyle = fyne.TextStyle{Bold: true}

	instructions := widget.NewRichTextFromMarkdown(`
**CSS Options Manager**

Enable CSS-based features for Canvus. These options create plugins that modify Canvus behavior.

**Requirements:**
- default-canvas must be set in mt-canvus.ini
- auto-pin must be 0 for kiosk modes
`)

	// Widget Options section
	widgetOptionsLabel := widget.NewLabel("Widget Options")
	widgetOptionsLabel.TextStyle = fyne.TextStyle{Bold: true}

	movingTooltip := "Disable moving canvas items (default: enabled in Canvus)"
	m.movingEnabled = widget.NewCheck("", nil)
	m.movingEnabled.SetChecked(false) // Default: unchecked (moving is enabled by default in Canvus)

	scalingTooltip := "Disable resizing canvas items (default: enabled in Canvus)"
	m.scalingEnabled = widget.NewCheck("", nil)
	m.scalingEnabled.SetChecked(false) // Default: unchecked (scaling is enabled by default in Canvus)

	rotationTooltip := "Enable rotating canvas items (default: disabled in Canvus)"
	m.rotationEnabled = widget.NewCheck("", nil)
	m.rotationEnabled.SetChecked(false) // Default: unchecked (rotation is disabled by default in Canvus)

	// Video Loop option
	videoLoopTooltip := "Enable automatic looping for videos (warning: may consume significant memory)"
	videoLoopWarning := widget.NewLabel("⚠ Memory Warning: Large videos may consume significant resources")
	videoLoopWarning.Importance = widget.WarningImportance
	m.videoLoopEnabled = widget.NewCheck("", nil)
	m.videoLoopEnabled.SetChecked(false) // Default: unchecked

	// UI Visibility Options
	uiVisibilityLabel := widget.NewLabel("UI Visibility Options")
	uiVisibilityLabel.TextStyle = fyne.TextStyle{Bold: true}

	hideTitleBarsTooltip := "Hide title bars on canvas items"
	m.hideTitleBarsEnabled = widget.NewCheck("", nil)

	hideResizeHandlesTooltip := "Hide resize handles on canvas items"
	m.hideResizeHandlesEnabled = widget.NewCheck("", nil)

	hideSidebarTooltip := "Hide sidebar menu that appears by widgets"
	m.hideSidebarEnabled = widget.NewCheck("", nil)

	hideMainMenuTooltip := "Hide main menu at side/bottom of screen"
	m.hideMainMenuEnabled = widget.NewCheck("", nil)

	hideFingerMenuTooltip := "Hide finger menu (CanvusCanvasMenu)"
	m.hideFingerMenuEnabled = widget.NewCheck("", nil)

	// Connector Options section
	connectorOptionsLabel := widget.NewLabel("Connector Options")
	connectorOptionsLabel.TextStyle = fyne.TextStyle{Bold: true}

	hideConnectorBulletsTooltip := "Hide connector bullets (blue dots around widgets for drawing connectors)"
	m.hideConnectorBulletsEnabled = widget.NewCheck("", nil)

	hideConnectorBulletsPresentationTooltip := "Hide connector bullets only in presentation or full-screen mode"
	m.hideConnectorBulletsPresentationEnabled = widget.NewCheck("", nil)

	// Standby Image for Video Outputs
	standbyImageTooltip := "Enable custom standby image for video outputs (PNG with transparency supported)"
	m.standbyImageEnabled = widget.NewCheck("", nil)
	m.standbyImageEnabled.SetChecked(false) // Default: unchecked
	standbyImageSectionLabel := widget.NewLabel("Video Output Standby Image")
	standbyImageSectionLabel.TextStyle = fyne.TextStyle{Bold: true}
	standbyImageDescription := widget.NewLabel("Set a custom standby image for video outputs (PNG with transparency supported)")
	m.standbyImageLabel = widget.NewLabel("No image selected")
	uploadImageBtn := widget.NewButton("Upload PNG Image", func() {
		m.uploadStandbyImage(window)
	})
	clearImageBtn := widget.NewButton("Clear", func() {
		m.standbyImagePath = ""
		m.standbyImageLabel.SetText("No image selected")
	})

	// Kiosk Mode option - mutually exclusive with Kiosk Plus
	kioskModeTooltip := "Requires: default-canvas set, auto-pin=0. Hides all UI elements including finger menu"
	m.kioskModeEnabled = widget.NewCheck("", func(checked bool) {
		if checked && m.kioskPlusEnabled.Checked {
			m.kioskPlusEnabled.SetChecked(false)
		}
		// Auto-enable all UI visibility options when kiosk mode is enabled
		if checked {
			m.hideTitleBarsEnabled.SetChecked(true)
			m.hideResizeHandlesEnabled.SetChecked(true)
			m.hideSidebarEnabled.SetChecked(true)
			m.hideMainMenuEnabled.SetChecked(true)
			m.hideFingerMenuEnabled.SetChecked(true)
		} else {
			// Auto-disable all UI visibility options when kiosk mode is disabled
			m.hideTitleBarsEnabled.SetChecked(false)
			m.hideResizeHandlesEnabled.SetChecked(false)
			m.hideSidebarEnabled.SetChecked(false)
			m.hideMainMenuEnabled.SetChecked(false)
			m.hideFingerMenuEnabled.SetChecked(false)
		}
	})

	// Kiosk Plus Mode option - mutually exclusive with Kiosk Mode
	kioskPlusTooltip := "Requires: default-canvas set, auto-pin=0. Hides all UI elements except finger menu"
	m.kioskPlusEnabled = widget.NewCheck("", func(checked bool) {
		if checked && m.kioskModeEnabled.Checked {
			m.kioskModeEnabled.SetChecked(false)
		}
		// Auto-enable all UI visibility options except finger menu when kiosk plus is enabled
		if checked {
			m.hideTitleBarsEnabled.SetChecked(true)
			m.hideResizeHandlesEnabled.SetChecked(true)
			m.hideSidebarEnabled.SetChecked(true)
			m.hideMainMenuEnabled.SetChecked(true)
			m.hideFingerMenuEnabled.SetChecked(false) // Keep finger menu visible
		} else {
			// Auto-disable all UI visibility options when kiosk plus is disabled
			m.hideTitleBarsEnabled.SetChecked(false)
			m.hideResizeHandlesEnabled.SetChecked(false)
			m.hideSidebarEnabled.SetChecked(false)
			m.hideMainMenuEnabled.SetChecked(false)
			m.hideFingerMenuEnabled.SetChecked(false)
		}
	})

	// Status label
	m.statusLabel = widget.NewLabel("Ready")

	// Helper function to create an option row with aligned checkbox
	// The entire row is clickable to toggle the checkbox
	createOptionRow := func(labelText, tooltipText string, checkbox *widget.Check) fyne.CanvasObject {
		labelWithColon := widget.NewLabel(labelText + ":")
		labelWithColon.Truncation = fyne.TextTruncateOff

		tooltip := widget.NewLabel(tooltipText)
		tooltip.Wrapping = fyne.TextWrapWord

		// Use Border layout: indentation + label on left, checkbox on right (with padding), description in center
		// This ensures description has enough space to wrap horizontally
		// Add padding on right so checkbox isn't covered by scrollbar
		indentLabel := container.NewHBox(
			widget.NewLabel("    "), // Indentation
			labelWithColon,
		)

		// Wrap checkbox with padding on the right to avoid scrollbar overlap
		checkboxWithPadding := container.NewHBox(
			checkbox,
			widget.NewLabel("					    "), // Right padding to avoid scrollbar
		)

		rowContent := container.NewBorder(
			nil, nil,
			indentLabel,        // Left: indentation + label
			checkboxWithPadding, // Right: checkbox with padding
			tooltip,           // Center: description (can expand and wrap horizontally)
		)

		// Make the entire row clickable by wrapping in a custom widget that handles taps
		clickableRow := &clickableRowWidget{
			content:  rowContent,
			checkbox: checkbox,
		}
		clickableRow.ExtendBaseWidget(clickableRow)

		return clickableRow
	}

	// Buttons - will be in fixed header
	// TODO: Generate Plugin is coming soon - implement plugin generation with proper file structure and validation
	// This feature should: create .canvusplugin manifest, generate styles.css, update mt-canvus.ini plugin-folders
	// Need to add documentation for plugin format and update UI to show plugin status
	generateBtn := widget.NewButton("Generate Plugin (Coming Soon)", func() {
		dialog.ShowInformation("Coming Soon", "Plugin generation feature is coming in a future release.\n\nFor now, you can:\n1. Preview the CSS using 'Preview CSS'\n2. Manually create plugin files following Canvus plugin documentation\n3. Use 'Launch Canvus with Current Config' for temporary testing", window)
	})
	generateBtn.Disable() // Disable until feature is ready

	validateBtn := widget.NewButton("Validate Requirements", func() {
		m.validateRequirements(window)
	})

	previewBtn := widget.NewButton("Preview CSS", func() {
		m.previewCSS(window)
	})

	launchBtn := widget.NewButton("Launch Canvus with Current Config", func() {
		m.launchCanvusWithConfig(window)
	})

	// Fixed header with buttons
	header := container.NewBorder(
		nil, nil, nil,
		container.NewHBox(validateBtn, previewBtn, generateBtn, launchBtn),
		title,
	)

	// Scrollable content
	form := container.NewVBox(
		instructions,
		widget.NewSeparator(),
		widgetOptionsLabel,
		createOptionRow("Disable Moving", movingTooltip, m.movingEnabled),
		createOptionRow("Disable Scaling", scalingTooltip, m.scalingEnabled),
		createOptionRow("Enable Rotation", rotationTooltip, m.rotationEnabled),
		widget.NewSeparator(),
		createOptionRow("Enable Video Looping", videoLoopTooltip, m.videoLoopEnabled),
		container.NewHBox(
			widget.NewLabel("    "), // Indentation
			videoLoopWarning,
		),
		widget.NewSeparator(),
		uiVisibilityLabel,
		createOptionRow("Enable Kiosk Mode", kioskModeTooltip, m.kioskModeEnabled),
		createOptionRow("Enable Kiosk Plus Mode", kioskPlusTooltip, m.kioskPlusEnabled),
		widget.NewSeparator(),
		createOptionRow("Hide Title Bars", hideTitleBarsTooltip, m.hideTitleBarsEnabled),
		createOptionRow("Hide Resize Handles", hideResizeHandlesTooltip, m.hideResizeHandlesEnabled),
		createOptionRow("Hide Sidebar", hideSidebarTooltip, m.hideSidebarEnabled),
		createOptionRow("Hide Main Menu", hideMainMenuTooltip, m.hideMainMenuEnabled),
		createOptionRow("Hide Finger Menu", hideFingerMenuTooltip, m.hideFingerMenuEnabled),
		widget.NewSeparator(),
		connectorOptionsLabel,
		createOptionRow("Hide Connector Bullets", hideConnectorBulletsTooltip, m.hideConnectorBulletsEnabled),
		createOptionRow("Hide Connector Bullets (Presentation/Full-screen)", hideConnectorBulletsPresentationTooltip, m.hideConnectorBulletsPresentationEnabled),
		widget.NewSeparator(),
		standbyImageSectionLabel,
		createOptionRow("Enable Standby Image", standbyImageTooltip, m.standbyImageEnabled),
		container.NewHBox(
			widget.NewLabel("    "), // Indentation
			standbyImageDescription,
		),
		container.NewHBox(
			widget.NewLabel("    "), // Indentation
			container.NewHBox(m.standbyImageLabel, uploadImageBtn, clearImageBtn),
		),
		widget.NewSeparator(),
		container.NewHBox(
			widget.NewLabel("    "), // Indentation
			m.statusLabel,
		),
	)

	// Use Border layout: header at top (fixed), scrollable content in center
	return container.NewBorder(
		header, // Top: fixed header with title and buttons
		nil,    // Bottom: nothing
		nil,    // Left: nothing
		nil,    // Right: nothing
		container.NewScroll(form), // Center: scrollable form content
	)
}

// validateRequirements validates that requirements are met for enabled options.
func (m *Manager) validateRequirements(window fyne.Window) {
	iniPath := m.fileService.DetectMtCanvusIni()
	if iniPath == "" {
		dialog.ShowError(fmt.Errorf("mt-canvus.ini not found"), window)
		return
	}

	iniFile, err := m.iniParser.Read(iniPath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to read mt-canvus.ini: %w", err), window)
		return
	}

	var errors []string
	var warnings []string

	// Helper function to get default-canvas from INI
	getDefaultCanvas := func() string {
		defaultCanvas := ""
		// Try root section first
		if section, err := iniFile.GetSection(""); err == nil {
			if key := section.Key("default-canvas"); key != nil {
				defaultCanvas = strings.TrimSpace(key.String())
			}
		}
		// Also check [content] section as fallback
		if defaultCanvas == "" {
			if section, err := iniFile.GetSection("content"); err == nil {
				if key := section.Key("default-canvas"); key != nil {
					defaultCanvas = strings.TrimSpace(key.String())
				}
			}
		}
		// Also check [canvas] section as fallback
		if defaultCanvas == "" {
			if section, err := iniFile.GetSection("canvas"); err == nil {
				if key := section.Key("default-canvas"); key != nil {
					defaultCanvas = strings.TrimSpace(key.String())
				}
			}
		}
		return defaultCanvas
	}

	// Helper function to get auto-pin from INI
	getAutoPin := func() string {
		autoPin := ""
		if section, err := iniFile.GetSection(""); err == nil {
			if key := section.Key("auto-pin"); key != nil {
				autoPin = strings.TrimSpace(key.String())
			}
		}
		// Also check [canvas] section for auto-pin
		if autoPin == "" {
			if section, err := iniFile.GetSection("canvas"); err == nil {
				if key := section.Key("auto-pin"); key != nil {
					autoPin = strings.TrimSpace(key.String())
				}
			}
		}
		return autoPin
	}

	// Check default-canvas when main menu is hidden (critical - no way to open canvas otherwise)
	if m.hideMainMenuEnabled.Checked {
		defaultCanvas := getDefaultCanvas()
		if defaultCanvas == "" {
			errors = append(errors, "default-canvas must be set when main menu is hidden (no way to open canvas otherwise)")
		}
	}

	// Check default-canvas when kiosk modes are enabled
	if m.kioskModeEnabled.Checked || m.kioskPlusEnabled.Checked {
		defaultCanvas := getDefaultCanvas()
		if defaultCanvas == "" {
			errors = append(errors, "default-canvas must be set for kiosk modes")
		}
	}

	// Check auto-pin when sidebar is hidden (pin/unpin controls are in sidebar)
	if m.hideSidebarEnabled.Checked {
		autoPin := getAutoPin()
		if autoPin != "0" {
			errors = append(errors, "auto-pin must be 0 when sidebar is hidden (pin/unpin controls are in sidebar, widgets will auto-pin with no way to unpin)")
		}
	}

	// Check auto-pin when kiosk modes are enabled
	if m.kioskModeEnabled.Checked || m.kioskPlusEnabled.Checked {
		autoPin := getAutoPin()
		if autoPin != "0" {
			errors = append(errors, "auto-pin must be 0 for kiosk modes")
		}
	}

	if len(errors) > 0 {
		m.statusLabel.SetText(fmt.Sprintf("Validation failed: %s", errors[0]))
		dialog.ShowError(fmt.Errorf("Validation failed:\n%s", fmt.Sprintf("%s", errors[0])), window)
		return
	}

	if len(warnings) > 0 {
		m.statusLabel.SetText(fmt.Sprintf("Warnings: %s", warnings[0]))
	} else {
		m.statusLabel.SetText("All requirements met")
		dialog.ShowInformation("Valid", "All requirements are met", window)
	}
}

// generatePlugin generates the CSS plugin files.
func (m *Manager) generatePlugin(window fyne.Window) {
	// Validate first
	iniPath := m.fileService.DetectMtCanvusIni()
	if iniPath == "" {
		dialog.ShowError(fmt.Errorf("mt-canvus.ini not found"), window)
		return
	}

	iniFile, err := m.iniParser.Read(iniPath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to read mt-canvus.ini: %w", err), window)
		return
	}

	// Helper function to get default-canvas from INI
	getDefaultCanvas := func() string {
		defaultCanvas := ""
		// Try root section first
		if section, err := iniFile.GetSection(""); err == nil {
			if key := section.Key("default-canvas"); key != nil {
				defaultCanvas = strings.TrimSpace(key.String())
			}
		}
		// Also check [content] section as fallback
		if defaultCanvas == "" {
			if section, err := iniFile.GetSection("content"); err == nil {
				if key := section.Key("default-canvas"); key != nil {
					defaultCanvas = strings.TrimSpace(key.String())
				}
			}
		}
		// Also check [canvas] section as fallback
		if defaultCanvas == "" {
			if section, err := iniFile.GetSection("canvas"); err == nil {
				if key := section.Key("default-canvas"); key != nil {
					defaultCanvas = strings.TrimSpace(key.String())
				}
			}
		}
		return defaultCanvas
	}

	// Helper function to get auto-pin from INI
	getAutoPin := func() string {
		autoPin := ""
		if section, err := iniFile.GetSection(""); err == nil {
			if key := section.Key("auto-pin"); key != nil {
				autoPin = strings.TrimSpace(key.String())
			}
		}
		// Also check [canvas] section for auto-pin
		if autoPin == "" {
			if section, err := iniFile.GetSection("canvas"); err == nil {
				if key := section.Key("auto-pin"); key != nil {
					autoPin = strings.TrimSpace(key.String())
				}
			}
		}
		return autoPin
	}

	// Check requirements
	// Check default-canvas when main menu is hidden (critical - no way to open canvas otherwise)
	if m.hideMainMenuEnabled.Checked {
		defaultCanvas := getDefaultCanvas()
		if defaultCanvas == "" {
			dialog.ShowError(fmt.Errorf("default-canvas must be set when main menu is hidden (no way to open canvas otherwise)"), window)
			return
		}
	}

	// Check default-canvas when kiosk modes are enabled
	if m.kioskModeEnabled.Checked || m.kioskPlusEnabled.Checked {
		defaultCanvas := getDefaultCanvas()
		if defaultCanvas == "" {
			dialog.ShowError(fmt.Errorf("default-canvas must be set for kiosk modes"), window)
			return
		}
	}

	// Check auto-pin when sidebar is hidden (pin/unpin controls are in sidebar)
	if m.hideSidebarEnabled.Checked {
		autoPin := getAutoPin()
		if autoPin != "0" {
			dialog.ShowError(fmt.Errorf("auto-pin must be 0 when sidebar is hidden (pin/unpin controls are in sidebar, widgets will auto-pin with no way to unpin)"), window)
			return
		}
	}

	// Check auto-pin when kiosk modes are enabled
	if m.kioskModeEnabled.Checked || m.kioskPlusEnabled.Checked {
		autoPin := getAutoPin()
		if autoPin != "0" {
			dialog.ShowError(fmt.Errorf("auto-pin must be 0 for kiosk modes"), window)
			return
		}
	}

	// Get plugin directory
	pluginDir := m.getPluginDirectory()
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		dialog.ShowError(fmt.Errorf("failed to create plugin directory: %w", err), window)
		return
	}

	// Generate plugin manifest
	manifest := m.createPluginManifest()
	manifestPath := filepath.Join(pluginDir, ".canvusplugin")
	if err := m.writeManifest(manifestPath, manifest); err != nil {
		dialog.ShowError(fmt.Errorf("failed to write manifest: %w", err), window)
		return
	}

	// Generate CSS
	css := m.generateCSS()
	cssPath := filepath.Join(pluginDir, "styles.css")
	if err := os.WriteFile(cssPath, []byte(css), 0644); err != nil {
		dialog.ShowError(fmt.Errorf("failed to write CSS: %w", err), window)
		return
	}

	// Update mt-canvus.ini with plugin folder
	if err := m.updatePluginFolders(iniFile, iniPath, pluginDir); err != nil {
		dialog.ShowError(fmt.Errorf("failed to update mt-canvus.ini: %w", err), window)
		return
	}

	m.statusLabel.SetText("Plugin generated successfully")
	dialog.ShowInformation("Success", fmt.Sprintf("CSS plugin generated at:\n%s", pluginDir), window)
}

// createPluginManifest creates the plugin manifest.
func (m *Manager) createPluginManifest() *PluginManifest {
	return &PluginManifest{
		APIVersion:  "1.0",
		Name:        "CanvusPowerToys CSS Options",
		Version:     "1.0.0",
		Description: "CSS options generated by Canvus PowerToys",
	}
}

// writeManifest writes the plugin manifest to file.
func (m *Manager) writeManifest(path string, manifest *PluginManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// generateCSS generates the CSS content based on enabled options.
func (m *Manager) generateCSS() string {
	var css strings.Builder

	// Widget Options: Control moving, scaling, and rotation
	// Canvus defaults: moving=enabled, scaling=enabled, rotation=disabled
	// Only generate CSS if user wants to override defaults
	needsWidgetCSS := false
	var widgetRules []string

	if m.movingEnabled.Checked {
		// User wants to disable moving (default is enabled)
		widgetRules = append(widgetRules, "input-translate: false !important;")
		needsWidgetCSS = true
	}
	if m.scalingEnabled.Checked {
		// User wants to disable scaling (default is enabled)
		widgetRules = append(widgetRules, "input-scale: false !important;")
		needsWidgetCSS = true
	}
	if m.rotationEnabled.Checked {
		// User wants to enable rotation (default is disabled)
		widgetRules = append(widgetRules, "input-rotation: true !important;")
		needsWidgetCSS = true
	}

	if needsWidgetCSS {
		css.WriteString("/* Widget Options CSS - Override Canvus defaults */\n")
		css.WriteString("/* Note: These are temporary and revert when canvas is closed */\n")
		css.WriteString("/* Canvus defaults: moving=enabled, scaling=enabled, rotation=disabled */\n")
		css.WriteString("CanvusCanvasWidget > * {\n")
		for _, rule := range widgetRules {
			css.WriteString("  " + rule + "\n")
		}
		css.WriteString("}\n\n")
	}

	// Video Looping: Enable looping for video widgets
	if m.videoLoopEnabled.Checked {
		css.WriteString("/* Video Looping CSS - Enables automatic looping for videos */\n")
		css.WriteString("/* Warning: May consume significant memory with large videos */\n")
		css.WriteString("CanvusCanvasVideo {\n")
		css.WriteString("  loop: true !important;\n")
		css.WriteString("}\n\n")
	}

	// UI Visibility Options - Individual controls
	if m.hideTitleBarsEnabled.Checked {
		css.WriteString("/* Hide Title Bars CSS - Hides title bars on canvas items */\n")
		css.WriteString("TitleBarWidget {\n")
		css.WriteString("  display: none !important;\n")
		css.WriteString("}\n\n")
	}

	if m.hideResizeHandlesEnabled.Checked {
		css.WriteString("/* Hide Resize Handles CSS - Hides resize handles on canvas items */\n")
		css.WriteString("CanvusResizeHandleWidget {\n")
		css.WriteString("  display: none !important;\n")
		css.WriteString("}\n\n")
	}

	if m.hideSidebarEnabled.Checked {
		css.WriteString("/* Hide Sidebar CSS - Hides sidebar menu that appears by widgets */\n")
		css.WriteString("SidebarWidget {\n")
		css.WriteString("  display: none !important;\n")
		css.WriteString("}\n\n")
	}

	if m.hideMainMenuEnabled.Checked {
		css.WriteString("/* Hide Main Menu CSS - Hides main menu at side/bottom of screen */\n")
		css.WriteString("MainMenu {\n")
		css.WriteString("  display: none !important;\n")
		css.WriteString("}\n\n")
	}

	if m.hideFingerMenuEnabled.Checked {
		css.WriteString("/* Hide Finger Menu CSS - Hides finger menu (CanvusCanvasMenu) */\n")
		css.WriteString("CanvusCanvasMenu {\n")
		css.WriteString("  display: none !important;\n")
		css.WriteString("}\n\n")
	}

	// Connector Options
	if m.hideConnectorBulletsEnabled.Checked {
		css.WriteString("/* Hide Connector Bullets - hides blue dots around widgets for drawing connectors */\n")
		css.WriteString("CanvasItemOverlayWidget:bullets > Widget {\n")
		css.WriteString("  display: none !important;\n")
		css.WriteString("}\n\n")
	}

	if m.hideConnectorBulletsPresentationEnabled.Checked {
		css.WriteString("/* Hide Connector Bullets in Presentation/Full-screen Mode */\n")
		css.WriteString("Workspace.presentation-mode CanvasItemOverlayWidget:bullets > Widget {\n")
		css.WriteString("  display: none !important;\n")
		css.WriteString("}\n")
		css.WriteString(".presenting CanvasItemOverlayWidget:bullets > Widget {\n")
		css.WriteString("  display: none !important;\n")
		css.WriteString("}\n\n")
	}

	// Note: Third-party touch menus are not affected by these rules
	css.WriteString("/* Note: Third-party touch menus are not affected by these rules */\n\n")

	// Standby Image for Video Outputs
	if m.standbyImageEnabled.Checked && m.standbyImagePath != "" {
		css.WriteString("/* Standby Image CSS - Sets custom standby image for video outputs */\n")
		css.WriteString("VideoOutputWidget > .output-info {\n")
		css.WriteString("  size: 100% !important;\n")
		// Use full absolute path with forward slashes
		imagePath := m.getStandbyImagePathForCSS()
		css.WriteString(fmt.Sprintf("  source: \"%s\" !important;\n", imagePath))
		css.WriteString("}\n\n")
	}

	return css.String()
}

// previewCSS shows a preview dialog with the generated CSS content.
func (m *Manager) previewCSS(window fyne.Window) {
	// Generate CSS content
	css := m.generateCSS()

	if css == "" {
		dialog.ShowInformation("Preview CSS", "No CSS will be generated with current options.\n\nPlease enable at least one option.", window)
		return
	}

	// Use MultiLineEntry for better performance with large CSS content
	previewEntry := widget.NewMultiLineEntry()
	previewEntry.SetText(css)
	previewEntry.Wrapping = fyne.TextWrapOff // Don't wrap CSS
	previewEntry.OnChanged = func(s string) {} // Make it read-only by ignoring changes - can still select/copy

	// Copy to clipboard button
	copyBtn := widget.NewButton("Copy to Clipboard", func() {
		window.Clipboard().SetContent(previewEntry.Text)
		dialog.ShowInformation("Copied", "CSS content copied to clipboard", window)
	})

	// Show image path info if standby image is enabled
	var infoText string
	if m.standbyImageEnabled.Checked && m.standbyImagePath != "" {
		imagePath := m.getStandbyImagePathForCSS()
		infoText = fmt.Sprintf("Standby Image Path: %s\nOriginal Path: %s\nCSS Directory: %s\n\n",
			imagePath, m.standbyImagePath, m.getCSSDirectory())
	}

	infoLabel := widget.NewLabel(infoText)
	infoLabel.Wrapping = fyne.TextWrapWord

	// Create a scrollable container for the text entry with minimum height
	// This provides a spacious preview area instead of a cramped 1-line box
	scrollContainer := container.NewScroll(previewEntry)
	scrollContainer.SetMinSize(fyne.NewSize(750, 400)) // Large enough for comfortable viewing

	// Create a container with info, entry and button
	content := container.NewVBox(
		infoLabel,
		container.NewBorder(
			copyBtn,           // Top: Copy button
			nil, nil, nil,
			scrollContainer,   // Center: Scrollable text with proper sizing
		),
	)

	previewDialog := dialog.NewCustom("Generated CSS Preview", "Close", content, window)
	previewDialog.Resize(fyne.NewSize(900, 700)) // Larger dialog for better visibility
	previewDialog.Show()
}

// getPluginDirectory returns the plugin directory path.
func (m *Manager) getPluginDirectory() string {
	configDir := m.fileService.GetUserConfigPath()
	return filepath.Join(configDir, "plugins", "canvus-powertoys-css")
}

// updatePluginFolders updates the plugin-folders setting in mt-canvus.ini.
// Plugin folders are registered in the [content] section per mt-canvus-plugin-api.md
func (m *Manager) updatePluginFolders(iniFile *ini.File, iniPath, pluginDir string) error {
	section, err := iniFile.GetSection("content")
	if err != nil {
		section, _ = iniFile.NewSection("content")
	}

	// Normalize path for INI file: replace backslashes with forward slashes
	// Windows paths need to use / or \\ in INI files, forward slashes are safer
	normalizedPluginDir := strings.ReplaceAll(pluginDir, "\\", "/")

	// Get existing plugin-folders
	existingFolders := ""
	if key := section.Key("plugin-folders"); key != nil {
		existingFolders = key.String()
	}

	// Normalize existing folders too (in case they have backslashes)
	normalizedExisting := strings.ReplaceAll(existingFolders, "\\", "/")

	// Add our plugin directory if not already present
	if !strings.Contains(normalizedExisting, normalizedPluginDir) {
		if normalizedExisting != "" {
			normalizedExisting += "," + normalizedPluginDir
		} else {
			normalizedExisting = normalizedPluginDir
		}
		section.Key("plugin-folders").SetValue(normalizedExisting)
	}

	// Create backup before updating
	if _, err := os.Stat(iniPath); err == nil {
		if _, err := m.backupManager.CreateBackup(iniPath); err != nil {
			fmt.Printf("Warning: Failed to create backup: %v\n", err)
		}
	}

	return m.iniParser.Write(iniFile, iniPath)
}

// launchCanvusWithConfig launches Canvus with the current CSS configuration.
func (m *Manager) launchCanvusWithConfig(window fyne.Window) {
	// Check if any options are enabled (widget options, video looping, UI visibility, or standby image)
	hasWidgetOptions := m.movingEnabled.Checked || m.scalingEnabled.Checked || m.rotationEnabled.Checked
	hasVideoOptions := m.videoLoopEnabled.Checked
	hasUIOptions := m.hideTitleBarsEnabled.Checked || m.hideResizeHandlesEnabled.Checked ||
		m.hideSidebarEnabled.Checked || m.hideMainMenuEnabled.Checked || m.hideFingerMenuEnabled.Checked
	hasStandbyImage := m.standbyImageEnabled.Checked && m.standbyImagePath != ""

	if !hasWidgetOptions && !hasVideoOptions && !hasUIOptions && !hasStandbyImage {
		dialog.ShowError(fmt.Errorf("please enable at least one CSS option before launching"), window)
		return
	}

	// Find Canvus executable
	canvusExe := m.findCanvusExecutable()
	if canvusExe == "" {
		// Ask user to select Canvus executable
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()
			uri := reader.URI()
			var exePath string
			if uri.Scheme() == "file" {
				exePath = uri.Path()
			} else {
				exePath = uri.String()
			}
			m.launchCanvusWithPath(window, exePath)
		}, window)
		return
	}

	m.launchCanvusWithPath(window, canvusExe)
}

// launchCanvusWithPath launches Canvus with the specified executable path.
func (m *Manager) launchCanvusWithPath(window fyne.Window, canvusExe string) {
	// Generate CSS file for current options
	cssDir := m.getCSSDirectory()
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		dialog.ShowError(fmt.Errorf("failed to create CSS directory: %w", err), window)
		return
	}

	// Generate unique filename based on enabled options
	cssFileName := m.generateCSSFileName()
	cssPath := filepath.Join(cssDir, cssFileName)

	// Generate CSS content
	css := m.generateCSS()
	if err := os.WriteFile(cssPath, []byte(css), 0644); err != nil {
		dialog.ShowError(fmt.Errorf("failed to write CSS file: %w", err), window)
		return
	}

	// Launch Canvus with CSS file
	workingDir := filepath.Dir(canvusExe)
	cmd := exec.Command(canvusExe, "--css", cssPath)
	cmd.Dir = workingDir

	if err := cmd.Start(); err != nil {
		dialog.ShowError(fmt.Errorf("failed to launch Canvus: %w", err), window)
		return
	}

	m.statusLabel.SetText(fmt.Sprintf("Canvus launched with CSS: %s", cssFileName))
	dialog.ShowInformation("Launched", fmt.Sprintf("Canvus launched with CSS configuration:\n%s", cssPath), window)
}

// findCanvusExecutable attempts to find the Canvus executable in common locations.
func (m *Manager) findCanvusExecutable() string {
	// Common Windows installation locations
	// Canvus executable is: %PROGRAMFILES%\MT Canvus\bin\mt-canvus-app.exe
	possiblePaths := []string{
		`C:\Program Files\MT Canvus\bin\mt-canvus-app.exe`,
		`C:\Program Files (x86)\MT Canvus\bin\mt-canvus-app.exe`,
	}

	// Check if mt-canvus-app.exe exists in any of these locations
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Check environment variables
	if programFiles := os.Getenv("ProgramFiles"); programFiles != "" {
		path := filepath.Join(programFiles, "MT Canvus", "bin", "mt-canvus-app.exe")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	if programFilesX86 := os.Getenv("ProgramFiles(x86)"); programFilesX86 != "" {
		path := filepath.Join(programFilesX86, "MT Canvus", "bin", "mt-canvus-app.exe")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// getCSSDirectory returns the directory for CSS files.
func (m *Manager) getCSSDirectory() string {
	configDir := m.fileService.GetUserConfigPath()
	return filepath.Join(configDir, "css")
}

// generateCSSFileName generates a unique filename based on enabled options.
func (m *Manager) generateCSSFileName() string {
	var parts []string

	if m.movingEnabled.Checked {
		parts = append(parts, "nomove")
	}
	if m.scalingEnabled.Checked {
		parts = append(parts, "noscale")
	}
	if m.rotationEnabled.Checked {
		parts = append(parts, "rotate")
	}
	if m.videoLoopEnabled.Checked {
		parts = append(parts, "loop")
	}
	if m.kioskModeEnabled.Checked {
		parts = append(parts, "kiosk")
	}
	if m.kioskPlusEnabled.Checked {
		parts = append(parts, "kioskplus")
	}
	if m.hideTitleBarsEnabled.Checked {
		parts = append(parts, "notitle")
	}
	if m.hideResizeHandlesEnabled.Checked {
		parts = append(parts, "noresize")
	}
	if m.hideSidebarEnabled.Checked {
		parts = append(parts, "nosidebar")
	}
	if m.hideMainMenuEnabled.Checked {
		parts = append(parts, "nomainmenu")
	}
	if m.hideFingerMenuEnabled.Checked {
		parts = append(parts, "nofinger")
	}
	if m.hideConnectorBulletsEnabled.Checked {
		parts = append(parts, "nobullets")
	}
	if m.hideConnectorBulletsPresentationEnabled.Checked {
		parts = append(parts, "nobulletspres")
	}
	if m.standbyImageEnabled.Checked && m.standbyImagePath != "" {
		parts = append(parts, "standby")
	}

	if len(parts) == 0 {
		parts = append(parts, "default")
	}

	filename := "canvus-" + strings.Join(parts, "-") + ".css"
	return filename
}

// generateShortcutName generates a unique shortcut name based on enabled options.
func (m *Manager) generateShortcutName() string {
	var parts []string

	if m.movingEnabled.Checked {
		parts = append(parts, "NoMove")
	}
	if m.scalingEnabled.Checked {
		parts = append(parts, "NoScale")
	}
	if m.rotationEnabled.Checked {
		parts = append(parts, "Rotate")
	}
	if m.videoLoopEnabled.Checked {
		parts = append(parts, "Loop")
	}
	if m.kioskModeEnabled.Checked {
		parts = append(parts, "Kiosk")
	}
	if m.kioskPlusEnabled.Checked {
		parts = append(parts, "KioskPlus")
	}
	if m.hideTitleBarsEnabled.Checked {
		parts = append(parts, "NoTitle")
	}
	if m.hideResizeHandlesEnabled.Checked {
		parts = append(parts, "NoResize")
	}
	if m.hideSidebarEnabled.Checked {
		parts = append(parts, "NoSidebar")
	}
	if m.hideMainMenuEnabled.Checked {
		parts = append(parts, "NoMainMenu")
	}
	if m.hideFingerMenuEnabled.Checked {
		parts = append(parts, "NoFinger")
	}
	if m.hideConnectorBulletsEnabled.Checked {
		parts = append(parts, "NoBullets")
	}
	if m.hideConnectorBulletsPresentationEnabled.Checked {
		parts = append(parts, "NoBulletsPres")
	}
	if m.standbyImageEnabled.Checked && m.standbyImagePath != "" {
		parts = append(parts, "Standby")
	}

	if len(parts) == 0 {
		return "Canvus Custom CSS"
	}

	return "Canvus " + strings.Join(parts, " ")
}

// uploadStandbyImage handles uploading a PNG image for video output standby.
func (m *Manager) uploadStandbyImage(window fyne.Window) {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		uri := reader.URI()
		var sourcePath string
		if uri.Scheme() == "file" {
			sourcePath = uri.Path()
		} else {
			sourcePath = uri.String()
		}

		// Validate it's a PNG file
		if !strings.HasSuffix(strings.ToLower(sourcePath), ".png") {
			dialog.ShowError(fmt.Errorf("only PNG files are supported for standby images"), window)
			return
		}

		// Copy image to CSS directory
		imagesDir := m.getImagesDirectory()
		if err := os.MkdirAll(imagesDir, 0755); err != nil {
			dialog.ShowError(fmt.Errorf("failed to create images directory: %w", err), window)
			return
		}

		// Get filename from source path
		filename := filepath.Base(sourcePath)
		destPath := filepath.Join(imagesDir, filename)

		// Read source file
		sourceData, err := os.ReadFile(sourcePath)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to read image file: %w", err), window)
			return
		}

		// Write to destination
		if err := os.WriteFile(destPath, sourceData, 0644); err != nil {
			dialog.ShowError(fmt.Errorf("failed to save image: %w", err), window)
			return
		}

		// Store path and update label
		m.standbyImagePath = destPath
		m.standbyImageLabel.SetText(fmt.Sprintf("Image: %s", filename))
		m.statusLabel.SetText(fmt.Sprintf("Standby image uploaded: %s", filename))
	}, window)
}

// getImagesDirectory returns the directory for storing standby images.
func (m *Manager) getImagesDirectory() string {
	configDir := m.fileService.GetUserConfigPath()
	return filepath.Join(configDir, "css", "images")
}

// clickableRowWidget is a widget that makes an entire row clickable to toggle a checkbox.
type clickableRowWidget struct {
	widget.BaseWidget
	content  fyne.CanvasObject
	checkbox *widget.Check
}

// CreateRenderer creates the renderer for the clickable row widget.
func (c *clickableRowWidget) CreateRenderer() fyne.WidgetRenderer {
	return &clickableRowRenderer{
		widget:  c,
		content: c.content,
	}
}

// Tapped handles tap events on the row, toggling the checkbox.
func (c *clickableRowWidget) Tapped(*fyne.PointEvent) {
	c.checkbox.SetChecked(!c.checkbox.Checked)
}

// clickableRowRenderer renders the clickable row widget.
type clickableRowRenderer struct {
	widget  *clickableRowWidget
	content fyne.CanvasObject
}

// Layout lays out the content.
func (r *clickableRowRenderer) Layout(size fyne.Size) {
	r.content.Resize(size)
	r.content.Move(fyne.NewPos(0, 0))
}

// MinSize returns the minimum size of the widget.
func (r *clickableRowRenderer) MinSize() fyne.Size {
	return r.content.MinSize()
}

// Objects returns the objects to render.
func (r *clickableRowRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.content}
}

// Refresh refreshes the renderer.
func (r *clickableRowRenderer) Refresh() {
	r.content.Refresh()
}

// Destroy destroys the renderer.
func (r *clickableRowRenderer) Destroy() {
	// No cleanup needed
}

// getStandbyImagePathForCSS returns the full absolute path to use in CSS with forward slashes.
func (m *Manager) getStandbyImagePathForCSS() string {
	if m.standbyImagePath == "" {
		return ""
	}

	// Get absolute path (in case it's relative)
	absPath, err := filepath.Abs(m.standbyImagePath)
	if err != nil {
		// If absolute path fails, use the original path
		absPath = m.standbyImagePath
	}

	// Normalize path separators for CSS (use forward slashes instead of backslashes)
	return strings.ReplaceAll(absPath, "\\", "/")
}
