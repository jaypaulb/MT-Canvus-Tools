package screenxml

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/organisms/services"
)

// Creator is the main Screen.xml Creator component.
type Creator struct {
	grid           *GridWidget
	gridContainer  *GridContainer
	xmlGenerator   *XMLGenerator
	iniIntegration *INIIntegration
	fileService    *services.FileService
	window         fyne.Window
	areaPerScreen  *widget.Check   // Checkbox for "area per screen or per gpu"
	addColBtn      *widget.Button  // Reference to add column button
	addRowBtn      *widget.Button  // Reference to add row button
	topBar         *fyne.Container // Reference to top bar for UI updates
	mainContainer  *fyne.Container // Reference to main container for updates
	centerArea     *fyne.Container  // Reference to center area container
}

// NewCreator creates a new Screen.xml Creator.
func NewCreator(fileService *services.FileService) (*Creator, error) {
	grid := NewGridWidget()
	gridContainer := NewGridContainer(grid)

	creator := &Creator{
		grid:           grid,
		gridContainer:  gridContainer,
		iniIntegration: NewINIIntegration(),
		fileService:    fileService,
	}

	// Initialize XML generator (simplified - no longer need separate handlers)
	defaultRes := Resolution{Width: 1920, Height: 1080, Name: "1920x1080 (Full HD)"}
	creator.xmlGenerator = NewXMLGenerator(grid, nil, nil, defaultRes)

	return creator, nil
}

// CreateUI creates the UI for the Screen.xml Creator tab.
func (c *Creator) CreateUI(window fyne.Window) fyne.CanvasObject {
	c.window = window

	// Top row: Checkbox for area generation mode
	// Per Screen = per GPU output (each output is one screen)
	// Per GPU = per window (one window per GPU, matching all outputs on that GPU)
	c.areaPerScreen = widget.NewCheck("Area per GPU (not per Screen)", func(checked bool) {
		// Update XML generator mode
		// checked = true means "per GPU" (per window)
		// checked = false means "per Screen" (per GPU output)
		c.xmlGenerator.SetAreaPerGPU(checked)
	})
	c.areaPerScreen.SetChecked(false) // Default: area per Screen (per GPU output)

	// Top: Checkbox + 3 buttons
	c.topBar = container.NewHBox(
		c.areaPerScreen,
		widget.NewSeparator(),
		widget.NewButton("Generate Screen.xml", func() {
			c.generateAndPreview(window)
		}),
		widget.NewButton("Save screen.xml", func() {
			c.saveScreenXML(window)
		}),
		widget.NewButton("Update mt-canvus.ini", func() {
			c.updateMtCanvusIni(window)
		}),
	)

	// Main: Grid container with cell widgets
	gridContainer := c.gridContainer.GetContainer()

	// Add Column button (on the right)
	c.addColBtn = widget.NewButton("+ Column", func() {
		c.grid.AddColumn()
		c.updateGridContainer()
	})

	// Add Row button (on the bottom)
	c.addRowBtn = widget.NewButton("+ Row", func() {
		c.grid.AddRow()
		c.updateGridContainer()
	})

	// Center area: grid with add column button on the right
	c.centerArea = container.NewBorder(
		nil, nil, nil, c.addColBtn, // Right: Add Column button
		gridContainer,
	)

	// Full layout: top bar, center area with buttons, bottom button
	c.mainContainer = container.NewBorder(
		c.topBar,     // Top
		c.addRowBtn,  // Bottom: Add Row button
		nil, nil,     // Left, Right
		c.centerArea, // Center
	)

	return c.mainContainer
}

// updateGridContainer updates the grid container in the UI when grid size changes.
func (c *Creator) updateGridContainer() {
	if c.window == nil || c.topBar == nil || c.addColBtn == nil || c.addRowBtn == nil {
		return
	}

	// Get the new container (this will trigger rebuildContainer which removes old widgets)
	newGridContainer := c.gridContainer.GetContainer()

	// Recreate the center area Border container with new grid container
	c.centerArea = container.NewBorder(
		nil, nil, nil, c.addColBtn, // Right: Add Column button
		newGridContainer, // Center: new grid container
	)

	// Recreate the entire main layout with updated grid
	c.mainContainer = container.NewBorder(
		c.topBar,     // Top
		c.addRowBtn,  // Bottom: Add Row button
		nil, nil,     // Left, Right
		c.centerArea, // Center: new center area with updated grid
	)

	// Update the tab content instead of replacing the entire window content
	// This preserves the header with tabs and "close to sys tray" message
	c.updateTabContent()
}

// updateTabContent updates the Screen.xml Creator tab content without affecting the window structure.
func (c *Creator) updateTabContent() {
	if c.window == nil || c.mainContainer == nil {
		return
	}

	// Find the AppTabs in the window content and update our tab
	windowContent := c.window.Content()
	if windowContent == nil {
		return
	}

	// Traverse the content tree to find AppTabs
	tabs := c.findAppTabs(windowContent)
	if tabs != nil {
		// Find the "Screen.xml Creator" tab and update its content
		for _, tab := range tabs.Items {
			if tab.Text == "Screen.xml Creator" {
				// Directly update the tab's content
				tab.Content = c.mainContainer
				// Refresh the tabs to reflect the change
				tabs.Refresh()
				return
			}
		}
	}

	// Fallback: if we can't find the tabs, just refresh the container
	c.mainContainer.Refresh()
}

// findAppTabs recursively searches for AppTabs in the content tree.
func (c *Creator) findAppTabs(obj fyne.CanvasObject) *container.AppTabs {
	// Check if this is an AppTabs container
	if tabs, ok := obj.(*container.AppTabs); ok {
		return tabs
	}

	// If it's a container, check its children
	if cont, ok := obj.(*fyne.Container); ok {
		for _, child := range cont.Objects {
			if tabs := c.findAppTabs(child); tabs != nil {
				return tabs
			}
		}
	}

	return nil
}

// updateMtCanvusIni updates mt-canvus.ini with video-output configuration.
func (c *Creator) updateMtCanvusIni(window fyne.Window) {
	iniPath := c.fileService.DetectMtCanvusIni()
	if iniPath == "" {
		dialog.ShowInformation("Not Found", "mt-canvus.ini not found in standard locations", window)
		return
	}

	outputCells := c.iniIntegration.DetectVideoOutputs(c.grid)
	if len(outputCells) == 0 {
		dialog.ShowInformation("No Outputs", "No cells without layout detected. Assign GPU outputs to cells (without layer) first.", window)
		return
	}

	if err := c.iniIntegration.UpdateMtCanvusIni(iniPath, outputCells); err != nil {
		dialog.ShowError(err, window)
		return
	}

	dialog.ShowInformation("Success", fmt.Sprintf("mt-canvus.ini updated successfully.\n\nCreated %d output section(s) for cells without layout.", len(outputCells)), window)
}

// generateAndPreview generates screen.xml and shows preview.
func (c *Creator) generateAndPreview(window fyne.Window) {
	// Check if there are any cells with data
	if !c.grid.HasCellsWithData() {
		dialog.ShowError(fmt.Errorf("Please add some displays to your array!"), window)
		return
	}

	xmlData, err := c.xmlGenerator.Generate()
	if err != nil {
		dialog.ShowError(err, window)
		return
	}

	// Validate
	if err := c.xmlGenerator.Validate(xmlData); err != nil {
		dialog.ShowError(fmt.Errorf("validation failed: %w", err), window)
		return
	}

	// Use MultiLineEntry for better performance with large XML content
	// MultiLineEntry handles large content much better than canvas.Text
	// which was causing the freezing issue
	previewEntry := widget.NewMultiLineEntry()
	xmlText := string(xmlData)
	previewEntry.SetText(xmlText)
	// Keep enabled to allow text selection, but prevent editing.
	isResetting := false
	previewEntry.OnChanged = func(text string) {
		if isResetting || text == xmlText {
			return
		}
		isResetting = true
		previewEntry.SetText(xmlText)
		isResetting = false
	}
	previewEntry.Wrapping = fyne.TextWrapOff // Don't wrap XML

	// Copy to clipboard button
	copyBtn := widget.NewButton("Copy to Clipboard", func() {
		window.Clipboard().SetContent(previewEntry.Text)
		dialog.ShowInformation("Copied", "XML content copied to clipboard", window)
	})

	// Create a container with the entry and button
	content := container.NewBorder(
		copyBtn, // Top: Copy button
		nil, nil, nil,
		container.NewScroll(previewEntry), // Center: Scrollable text
	)

	previewDialog := dialog.NewCustom("Generated screen.xml Preview", "Close", content, window)
	previewDialog.Resize(fyne.NewSize(800, 600))
	previewDialog.Show()
}

// saveScreenXML saves the generated screen.xml to the default location with backup.
func (c *Creator) saveScreenXML(window fyne.Window) {
	// Check if there are any cells with data
	if !c.grid.HasCellsWithData() {
		dialog.ShowError(fmt.Errorf("Please add some displays to your array!"), window)
		return
	}

	xmlData, err := c.xmlGenerator.Generate()
	if err != nil {
		dialog.ShowError(err, window)
		return
	}

	// Validate
	if err := c.xmlGenerator.Validate(xmlData); err != nil {
		dialog.ShowError(fmt.Errorf("validation failed: %w", err), window)
		return
	}

	// Get default screen.xml path
	defaultPath, err := c.fileService.GetDefaultScreenXmlPath()
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to get default path: %w", err), window)
		return
	}

	// Create backup of existing file if it exists
	if _, err := os.Stat(defaultPath); err == nil {
		// File exists, create backup with date suffix
		dir := filepath.Dir(defaultPath)
		baseName := filepath.Base(defaultPath)
		ext := filepath.Ext(baseName)
		nameWithoutExt := baseName[:len(baseName)-len(ext)]

		// Format date as YYYYMMDD_HHMMSS
		dateStr := time.Now().Format("20060102_150405")
		backupPath := filepath.Join(dir, fmt.Sprintf("%s_%s%s", nameWithoutExt, dateStr, ext))

		if err := os.Rename(defaultPath, backupPath); err != nil {
			dialog.ShowError(fmt.Errorf("failed to create backup: %w", err), window)
			return
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(defaultPath)
	if err := c.fileService.EnsureDirectory(dir); err != nil {
		dialog.ShowError(fmt.Errorf("failed to create directory: %w", err), window)
		return
	}

	// Write the new file
	if err := os.WriteFile(defaultPath, xmlData, 0644); err != nil {
		dialog.ShowError(fmt.Errorf("failed to write file: %w", err), window)
		return
	}

	dialog.ShowInformation("Success", fmt.Sprintf("screen.xml saved successfully to:\n%s", defaultPath), window)
}
