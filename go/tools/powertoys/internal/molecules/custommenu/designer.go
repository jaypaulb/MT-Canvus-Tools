package custommenu

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/backup"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/logger"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/paths"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/organisms/services"
)

// MenuItem represents a menu item in the YAML structure.
type MenuItem struct {
	Tooltip  string     `yaml:"tooltip"`
	Icon     string     `yaml:"icon,omitempty"`
	IconBack string     `yaml:"icon-back,omitempty"`
	Items    []MenuItem `yaml:"items,omitempty"`
	Actions  []Action   `yaml:"actions,omitempty"`
}

// Action represents a menu item action.
type Action struct {
	Name       string                 `yaml:"name"` // "create" or "open-folder"
	Parameters map[string]interface{} `yaml:"parameters,omitempty"`
}

// Designer handles custom menu design and YAML generation.
type Designer struct {
	fileService   *services.FileService
	yamlHandler   *config.YAMLHandler
	iniParser     *config.INIParser
	backupManager *backup.Manager
	menuTree      *widget.Tree
	rootItem      MenuItem   // The root menu item
	menuData      []MenuItem // Deprecated: kept for compatibility
	formContainer *fyne.Container
	selectedID    widget.TreeNodeID
	window        fyne.Window
	iconPicker    *IconPicker
}

// NewDesigner creates a new Custom Menu Designer.
func NewDesigner(fileService *services.FileService) (*Designer, error) {
	iconPicker, err := NewIconPicker()
	if err != nil {
		logger.Logf("Warning: Failed to initialize icon picker: %v", err)
	}

	return &Designer{
		fileService:   fileService,
		yamlHandler:   config.NewYAMLHandler(),
		iniParser:     config.NewINIParser(),
		backupManager: backup.NewManager(""),
		rootItem: MenuItem{
			Tooltip: "Custom Menu",
			Icon:    "icons/custom-menu.png",
			Items:   []MenuItem{},
		},
		menuData:   []MenuItem{}, // Deprecated
		iconPicker: iconPicker,
	}, nil
}

// CreateUI creates the UI for the Custom Menu Designer tab.
func (d *Designer) CreateUI(window fyne.Window) fyne.CanvasObject {
	d.window = window
	title := widget.NewLabel("Custom Menu Designer")
	title.TextStyle = fyne.TextStyle{Bold: true}

	instructions := widget.NewRichTextFromMarkdown(`
**Custom Menu Designer**

Create and edit custom menus for Canvus. Menus are saved as menu.yml files.

**Features:**
- Hierarchical menu structure with unlimited nesting
- Icon support (937x937 PNG, automatically converted)
- Actions: create (note, pdf, video, image, browser) and open-folder
- Full parameter support for all content types
`)

	// Menu tree view
	d.menuTree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			return d.getChildNodes(id)
		},
		func(id widget.TreeNodeID) bool {
			return d.hasChildren(id)
		},
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Menu Item")
		},
		func(id widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			item := d.getItemByID(id)
			if item != nil {
				label.SetText(item.Tooltip)
			}
		},
	)

	d.menuTree.OnSelected = func(id widget.TreeNodeID) {
		d.selectedID = id
		d.showItemForm(id)
	}

	// Form container
	d.formContainer = container.NewVBox(widget.NewLabel("Select a menu item to edit"))

	formScroll := container.NewScroll(d.formContainer)
	formScroll.SetMinSize(fyne.NewSize(500, 0))

	// Buttons
	addItemBtn := widget.NewButton("Add Menu Item", func() {
		d.addMenuItem()
	})

	addSubItemBtn := widget.NewButton("Add Sub-Item", func() {
		d.addSubItem()
	})

	removeItemBtn := widget.NewButton("Remove Item", func() {
		d.removeItem()
	})

	moveUpBtn := widget.NewButton("Move Up", func() {
		d.moveItem(-1)
	})

	moveDownBtn := widget.NewButton("Move Down", func() {
		d.moveItem(1)
	})

	importBtn := widget.NewButton("Import menu.yml", func() {
		d.importMenu(window)
	})

	saveBtn := widget.NewButton("Save menu.yml", func() {
		d.saveMenu(window)
	})

	validateBtn := widget.NewButton("Validate", func() {
		d.validateMenu(window)
	})

	// Layout
	leftPanel := container.NewBorder(
		container.NewVBox(
			title,
			instructions,
			widget.NewSeparator(),
			container.NewGridWithColumns(2, addItemBtn, addSubItemBtn),
			container.NewGridWithColumns(2, removeItemBtn, importBtn),
			container.NewGridWithColumns(2, moveUpBtn, moveDownBtn),
		),
		nil, nil, nil,
		d.menuTree,
	)

	split := container.NewHSplit(leftPanel, formScroll)
	split.SetOffset(0.4)

	return container.NewBorder(
		container.NewHBox(saveBtn, validateBtn),
		nil, nil, nil,
		split,
	)
}

// getChildNodes returns child node IDs for a given parent.
func (d *Designer) getChildNodes(id widget.TreeNodeID) []widget.TreeNodeID {
	if id == "" {
		// Tree root - return "root" node
		return []widget.TreeNodeID{"root"}
	}

	if id == "root" {
		// Root menu item - return its children
		var ids []widget.TreeNodeID
		for i := range d.rootItem.Items {
			ids = append(ids, fmt.Sprintf("root.%d", i))
		}
		return ids
	}

	// Get item and return its children
	item := d.getItemByID(id)
	if item != nil && len(item.Items) > 0 {
		var ids []widget.TreeNodeID
		for i := range item.Items {
			ids = append(ids, fmt.Sprintf("%s.%d", id, i))
		}
		return ids
	}
	return nil
}

// hasChildren checks if a node has children.
func (d *Designer) hasChildren(id widget.TreeNodeID) bool {
	item := d.getItemByID(id)
	return item != nil && len(item.Items) > 0
}

// getItemByID gets a menu item by its ID.
func (d *Designer) getItemByID(id widget.TreeNodeID) *MenuItem {
	if id == "" || id == "root" {
		return &d.rootItem
	}

	// Remove "root." prefix if present
	idStr := string(id)
	if strings.HasPrefix(idStr, "root.") {
		idStr = idStr[5:] // Remove "root."
	}

	// Parse ID (format: "0", "0.1", "0.1.2", etc.)
	parts := d.parseID(widget.TreeNodeID(idStr))
	if len(parts) == 0 {
		return nil
	}

	if parts[0] >= len(d.rootItem.Items) {
		return nil
	}

	item := &d.rootItem.Items[parts[0]]
	for i := 1; i < len(parts); i++ {
		if parts[i] >= len(item.Items) {
			return nil
		}
		item = &item.Items[parts[i]]
	}
	return item
}

// parseID parses a node ID into indices.
func (d *Designer) parseID(id widget.TreeNodeID) []int {
	if id == "" {
		return nil
	}
	parts := strings.Split(string(id), ".")
	indices := make([]int, len(parts))
	for i, part := range parts {
		idx, err := strconv.Atoi(part)
		if err != nil {
			return nil
		}
		indices[i] = idx
	}
	return indices
}

// showItemForm shows the form for editing a menu item.
func (d *Designer) showItemForm(id widget.TreeNodeID) {
	item := d.getItemByID(id)
	if item == nil {
		return
	}

	form := d.createItemForm(item, id)
	d.formContainer.Objects = []fyne.CanvasObject{form}
	d.formContainer.Refresh()
}

// createItemForm creates a form for editing a menu item.
func (d *Designer) createItemForm(item *MenuItem, id widget.TreeNodeID) fyne.CanvasObject {
	// Basic fields
	tooltipEntry := widget.NewEntry()
	tooltipEntry.SetText(item.Tooltip)
	tooltipEntry.SetPlaceHolder("Menu item tooltip (required)")

	iconEntry := widget.NewEntry()
	iconEntry.SetText(item.Icon)
	iconEntry.SetPlaceHolder("Path to icon file")

	iconBrowseBtn := widget.NewButton("Browse...", func() {
		d.browseIcon(iconEntry)
	})

	iconPickerBtn := widget.NewButton("Icon Picker", func() {
		if d.iconPicker != nil {
			d.iconPicker.Show(d.window, func(iconPath string) {
				iconEntry.SetText(iconPath)
			})
		}
	})

	iconBackEntry := widget.NewEntry()
	iconBackEntry.SetText(item.IconBack)
	iconBackEntry.SetPlaceHolder("Optional back button icon")

	// Item type selector
	itemTypeGroup := widget.NewRadioGroup([]string{"Submenu", "Actions"}, func(selected string) {
		// Switch between submenu and actions modes
		if selected == "Actions" {
			// Clear sub-items to allow actions
			item.Items = nil
			if len(item.Actions) == 0 {
				// Add a default action if none exist
				item.Actions = []Action{{
					Name:       "create",
					Parameters: map[string]interface{}{"type": "note"},
				}}
			}
		} else if selected == "Submenu" {
			// Clear actions to allow sub-items
			item.Actions = nil
		}
		// Refresh form to show appropriate sections
		d.showItemForm(id)
	})

	if len(item.Items) > 0 {
		itemTypeGroup.SetSelected("Submenu")
	} else if len(item.Actions) > 0 {
		itemTypeGroup.SetSelected("Actions")
	}

	saveBtn := widget.NewButton("Save Changes", func() {
		item.Tooltip = tooltipEntry.Text
		item.Icon = iconEntry.Text
		item.IconBack = iconBackEntry.Text
		d.menuTree.Refresh()
		dialog.ShowInformation("Saved", "Item changes saved", d.window)
	})

	// Build form based on type
	formContent := container.NewVBox(
		widget.NewLabel("Item Editor"),
		widget.NewSeparator(),
		widget.NewLabel("Tooltip:"),
		tooltipEntry,
		widget.NewLabel("Icon:"),
		container.NewBorder(nil, nil, nil, container.NewHBox(iconBrowseBtn, iconPickerBtn), iconEntry),
		widget.NewLabel("Back Icon (optional):"),
		iconBackEntry,
		widget.NewSeparator(),
		widget.NewLabel("Item Type:"),
		itemTypeGroup,
		widget.NewSeparator(),
	)

	if itemTypeGroup.Selected == "Actions" {
		// Show actions editor
		actionEditor := d.createActionEditor(item)
		formContent.Add(actionEditor)
	} else {
		formContent.Add(widget.NewLabel("This item has sub-items. Use tree to manage children."))
	}

	formContent.Add(widget.NewSeparator())
	formContent.Add(saveBtn)

	return formContent
}

// createActionEditor creates the action editor UI.
func (d *Designer) createActionEditor(item *MenuItem) fyne.CanvasObject {
	actionList := container.NewVBox()

	// Display existing actions
	for i := range item.Actions {
		actionIndex := i
		actionCard := d.createActionCard(&item.Actions[actionIndex], func() {
			// Remove action
			item.Actions = append(item.Actions[:actionIndex], item.Actions[actionIndex+1:]...)
			d.showItemForm(d.selectedID)
		})
		actionList.Add(actionCard)
	}

	// Add action button
	addActionBtn := widget.NewButton("Add Action", func() {
		item.Actions = append(item.Actions, Action{
			Name:       "create",
			Parameters: map[string]interface{}{"type": "note"},
		})
		d.showItemForm(d.selectedID)
	})

	return container.NewVBox(
		widget.NewLabel("Actions:"),
		actionList,
		addActionBtn,
	)
}

// createActionCard creates a card for a single action.
func (d *Designer) createActionCard(action *Action, onRemove func()) fyne.CanvasObject {
	// Action type selector
	actionTypeSelect := widget.NewSelect([]string{"create", "open-folder"}, func(selected string) {
		action.Name = selected
		if action.Parameters == nil {
			action.Parameters = make(map[string]interface{})
		}
		d.showItemForm(d.selectedID)
	})
	actionTypeSelect.SetSelected(action.Name)

	removeBtn := widget.NewButton("Remove", onRemove)

	header := container.NewBorder(nil, nil, nil, removeBtn,
		container.NewHBox(widget.NewLabel("Action Type:"), actionTypeSelect))

	// Parameters form based on action type
	var paramsForm fyne.CanvasObject
	if action.Name == "create" {
		paramsForm = d.createCreateActionForm(action)
	} else if action.Name == "open-folder" {
		paramsForm = d.createOpenFolderForm(action)
	}

	return container.NewVBox(
		widget.NewCard("", "", container.NewVBox(header, paramsForm)),
	)
}

// addMenuItem adds a new top-level menu item.
func (d *Designer) addMenuItem() {
	newItem := MenuItem{
		Tooltip: "New Menu Item",
	}
	d.rootItem.Items = append(d.rootItem.Items, newItem)

	newID := fmt.Sprintf("root.%d", len(d.rootItem.Items)-1)

	// Refresh the tree and then open/select after refresh completes
	d.menuTree.Refresh()

	// Use goroutine to ensure operations happen after refresh completes
	go func() {
		d.menuTree.OpenBranch("root")
		d.menuTree.Select(newID)
	}()
}

// addSubItem adds a sub-item to the currently selected item.
func (d *Designer) addSubItem() {
	if d.selectedID == "" {
		dialog.ShowInformation("No Selection", "Please select a menu item first", d.window)
		return
	}

	item := d.getItemByID(d.selectedID)
	if item == nil {
		return
	}

	// Clear actions when adding sub-items
	item.Actions = nil
	newSubItem := MenuItem{
		Tooltip: "New Sub-Item",
	}
	item.Items = append(item.Items, newSubItem)

	newSubID := fmt.Sprintf("%s.%d", d.selectedID, len(item.Items)-1)
	parentID := d.selectedID

	// Refresh the tree and then open/select after refresh completes
	d.menuTree.Refresh()

	// Use goroutine to ensure operations happen after refresh completes
	go func() {
		d.menuTree.OpenBranch(parentID)
		d.menuTree.Select(newSubID)
	}()
}

// removeItem removes the currently selected item.
func (d *Designer) removeItem() {
	if d.selectedID == "" {
		dialog.ShowInformation("No Selection", "Please select a menu item first", d.window)
		return
	}

	dialog.ShowConfirm("Remove Item", "Are you sure you want to remove this item?", func(confirmed bool) {
		if !confirmed {
			return
		}

		parts := d.parseID(d.selectedID)
		if len(parts) == 0 {
			return
		}

		// Remove from parent
		if len(parts) == 1 {
			// Top level
			d.menuData = append(d.menuData[:parts[0]], d.menuData[parts[0]+1:]...)
		} else {
			// Navigate to parent
			parent := &d.menuData[parts[0]]
			for i := 1; i < len(parts)-1; i++ {
				parent = &parent.Items[parts[i]]
			}
			lastIdx := parts[len(parts)-1]
			parent.Items = append(parent.Items[:lastIdx], parent.Items[lastIdx+1:]...)
		}

		d.selectedID = ""
		d.menuTree.Refresh()
		d.formContainer.Objects = []fyne.CanvasObject{widget.NewLabel("Select a menu item to edit")}
		d.formContainer.Refresh()
	}, d.window)
}

// moveItem moves the selected item up or down.
func (d *Designer) moveItem(direction int) {
	if d.selectedID == "" {
		dialog.ShowInformation("No Selection", "Please select a menu item first", d.window)
		return
	}

	parts := d.parseID(d.selectedID)
	if len(parts) == 0 {
		return
	}

	// Get the slice to reorder
	var slice *[]MenuItem
	idx := parts[len(parts)-1]

	if len(parts) == 1 {
		// Top level
		slice = &d.menuData
	} else {
		// Navigate to parent
		parent := &d.menuData[parts[0]]
		for i := 1; i < len(parts)-1; i++ {
			parent = &parent.Items[parts[i]]
		}
		slice = &parent.Items
	}

	newIdx := idx + direction
	if newIdx < 0 || newIdx >= len(*slice) {
		return // Can't move beyond bounds
	}

	// Swap
	(*slice)[idx], (*slice)[newIdx] = (*slice)[newIdx], (*slice)[idx]

	// Update selected ID
	parts[len(parts)-1] = newIdx
	var newID string
	for i, p := range parts {
		if i > 0 {
			newID += "."
		}
		newID += strconv.Itoa(p)
	}
	d.selectedID = widget.TreeNodeID(newID)

	d.menuTree.Refresh()
	d.menuTree.Select(d.selectedID)
}

// browseIcon opens a file browser for icon selection.
func (d *Designer) browseIcon(iconEntry *widget.Entry) {
	if d.window == nil {
		return
	}

	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		selectedPath := reader.URI().Path()

		// Check if it's a PNG and if conversion is needed
		if filepath.Ext(selectedPath) == ".png" {
			// Optionally convert to 937x937 format
			dialog.ShowConfirm("Convert Icon", "Convert this icon to 937x937 PNG format?", func(convert bool) {
				if convert {
					converter := NewIconConverter()
					outputPath := filepath.Join(filepath.Dir(selectedPath), "converted_"+filepath.Base(selectedPath))
					if err := converter.ConvertIcon(selectedPath, outputPath); err != nil {
						dialog.ShowError(err, d.window)
					} else {
						iconEntry.SetText(outputPath)
						dialog.ShowInformation("Success", "Icon converted successfully", d.window)
					}
				} else {
					iconEntry.SetText(selectedPath)
				}
			}, d.window)
		} else {
			iconEntry.SetText(selectedPath)
		}
	}, d.window)
}

// importMenu imports an existing menu.yml file.
func (d *Designer) importMenu(window fyne.Window) {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		// Read YAML file - structure is root item with items array
		var rootItem MenuItem
		if err := d.yamlHandler.Read(reader.URI().Path(), &rootItem); err != nil {
			dialog.ShowError(fmt.Errorf("failed to import menu.yml: %w", err), window)
			return
		}

		d.rootItem = rootItem
		d.menuTree.Refresh()
		// Open all branches to show imported structure
		d.menuTree.OpenAllBranches()

		// Select root node
		d.menuTree.Select("root")

		dialog.ShowInformation("Imported", fmt.Sprintf("menu.yml imported successfully\nRoot: %s\n%d top-level items", rootItem.Tooltip, len(rootItem.Items)), window)
	}, window)

	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".yml", ".yaml"}))
	d.setDefaultMenuDialogLocation(fileDialog)
	fileDialog.Show()
}

// saveMenu saves the menu to menu.yml and updates mt-canvus.ini.
func (d *Designer) saveMenu(window fyne.Window) {
	customMenuDir := filepath.Join(d.fileService.GetUserConfigPath(), "CustomMenu")
	configDir := customMenuDir
	menuPath := filepath.Join(configDir, "menu.yml")

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		dialog.ShowError(fmt.Errorf("failed to create directory: %w", err), window)
		return
	}

	// Create backup if file exists
	if _, err := os.Stat(menuPath); err == nil {
		if _, err := d.backupManager.CreateBackup(menuPath); err != nil {
			logger.Logf("Warning: Failed to create backup: %v", err)
		}
	}

	// Save YAML using the rootItem
	if err := d.yamlHandler.Write(menuPath, d.rootItem); err != nil {
		dialog.ShowError(fmt.Errorf("failed to save menu.yml: %w", err), window)
		return
	}

	// Update mt-canvus.ini
	iniPath := d.fileService.DetectMtCanvusIni()
	if iniPath != "" {
		if err := d.updateCustomMenuIni(iniPath, menuPath); err != nil {
			dialog.ShowError(fmt.Errorf("failed to update mt-canvus.ini: %w", err), window)
			return
		}
	}

	dialog.ShowInformation("Saved", fmt.Sprintf("menu.yml saved to:\n%s", menuPath), window)
}

// validateMenu validates the menu structure.
func (d *Designer) validateMenu(window fyne.Window) {
	issues := []string{}

	// Validate menu data
	for i, item := range d.menuData {
		itemIssues := d.validateItem(&item, fmt.Sprintf("Item %d", i+1))
		issues = append(issues, itemIssues...)
	}

	if len(issues) == 0 {
		dialog.ShowInformation("Validation", "Menu validation passed! No issues found.", window)
	} else {
		msg := "Validation issues found:\n\n" + strings.Join(issues, "\n")
		dialog.ShowInformation("Validation", msg, window)
	}
}

// validateItem validates a single menu item recursively.
func (d *Designer) validateItem(item *MenuItem, path string) []string {
	issues := []string{}

	if item.Tooltip == "" {
		issues = append(issues, fmt.Sprintf("%s: Missing tooltip", path))
	}

	if item.Icon != "" && !paths.FileExists(item.Icon) {
		issues = append(issues, fmt.Sprintf("%s: Icon file not found: %s", path, item.Icon))
	}

	if len(item.Items) == 0 && len(item.Actions) == 0 {
		issues = append(issues, fmt.Sprintf("%s: Item has no sub-items or actions", path))
	}

	// Validate children
	for _, child := range item.Items {
		childPath := fmt.Sprintf("%s > %s", path, child.Tooltip)
		childIssues := d.validateItem(&child, childPath)
		issues = append(issues, childIssues...)
	}

	return issues
}

// updateCustomMenuIni updates the custom-menu entry in mt-canvus.ini.
func (d *Designer) updateCustomMenuIni(iniPath, menuPath string) error {
	iniFile, err := d.iniParser.Read(iniPath)
	if err != nil {
		return fmt.Errorf("failed to read mt-canvus.ini: %w", err)
	}

	// Get or create [canvas] section
	section, err := iniFile.GetSection("canvas")
	if err != nil {
		section, _ = iniFile.NewSection("canvas")
	}

	// Set custom-menu to relative path
	relPath, err := filepath.Rel(filepath.Dir(iniPath), menuPath)
	if err != nil {
		relPath = menuPath // Fallback to absolute path
	}

	section.Key("custom-menu").SetValue(relPath)

	// Create backup before updating
	if _, err := os.Stat(iniPath); err == nil {
		if _, err := d.backupManager.CreateBackup(iniPath); err != nil {
			logger.Logf("Warning: Failed to create backup: %v", err)
		}
	}

	return d.iniParser.Write(iniFile, iniPath)
}

// setDefaultMenuDialogLocation sets the file dialog location to the default custom menu directory.
func (d *Designer) setDefaultMenuDialogLocation(fileDialog *dialog.FileDialog) {
	if fileDialog == nil || d.fileService == nil {
		return
	}

	defaultMenuPath := d.defaultMenuFilePath()
	if paths.FileExists(defaultMenuPath) {
		if d.setDialogLocation(fileDialog, filepath.Dir(defaultMenuPath)) {
			return
		}
	}

	// Fallback to base Canvus config directory if menu.yml isn't present
	d.setDialogLocation(fileDialog, d.fileService.GetUserConfigPath())
}

// setDialogLocation attempts to point the dialog at a specific path.
func (d *Designer) setDialogLocation(fileDialog *dialog.FileDialog, targetPath string) bool {
	if targetPath == "" {
		return false
	}

	lister, err := storage.ListerForURI(storage.NewFileURI(targetPath))
	if err != nil {
		return false
	}

	fileDialog.SetLocation(lister)
	return true
}

// defaultMenuFilePath returns the expected default location for menu.yml.
func (d *Designer) defaultMenuFilePath() string {
	if d.fileService == nil {
		return ""
	}
	base := d.fileService.GetUserConfigPath()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "CustomMenu", "menu.yml")
}
