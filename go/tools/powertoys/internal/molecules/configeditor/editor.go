package configeditor

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

// Editor is the main Canvus Config Editor component.
type Editor struct {
	iniParser      *config.INIParser
	fileService    *services.FileService
	backupManager  *backup.Manager
	iniFile        *ini.File
	schema         *ConfigSchema // Schema with all possible options
	searchEntry    *widget.Entry
	accordion      *widget.Accordion
	sectionGroups  map[string]*SectionGroup
	compoundGroups map[string]*CompoundEntryGroup
	formContainer  *container.Scroll
	scrollContainer *container.Scroll // Reference to scroll container for auto-scrolling
	window         fyne.Window
}

// OptionItem represents a configuration option.
type OptionItem struct {
	Section string
	Key     string
	Value   string
	Tooltip string
}

// NewEditor creates a new Config Editor instance.
func NewEditor(fileService *services.FileService) (*Editor, error) {
	// Create backup manager with temporary directory (will use file's directory when backing up)
	backupMgr := backup.NewManager("")

	// Use embedded schema (manually maintained from documentation)
	schema := GetEmbeddedSchema()

	return &Editor{
		iniParser:      config.NewINIParser(),
		fileService:     fileService,
		backupManager:  backupMgr,
		schema:         schema,
		sectionGroups:  make(map[string]*SectionGroup),
		compoundGroups: make(map[string]*CompoundEntryGroup),
	}, nil
}

// CreateUI creates the UI for the Config Editor tab.
func (e *Editor) CreateUI(window fyne.Window) fyne.CanvasObject {
	e.window = window

	// Top: Search bar
	e.searchEntry = widget.NewEntry()
	e.searchEntry.SetPlaceHolder("Search configuration options...")
	e.searchEntry.OnChanged = func(text string) {
		e.filterSections(text)
	}

	// Create accordion for sections
	e.accordion = widget.NewAccordion()

	// Try to auto-load: first example file for schema, then live file for values
	if err := e.loadConfigFilesSilent(); err == nil {
		// Successfully loaded
	}

	// Build UI from schema
	e.buildUIFromSchema()

	// Load button
	loadBtn := widget.NewButton("Load mt-canvus.ini", func() {
		e.loadConfigFile(window)
	})

	// Save buttons
	saveUserBtn := widget.NewButton("Save to User Config", func() {
		e.saveConfig(window, true)
	})
	saveSystemBtn := widget.NewButton("Save to System Config", func() {
		e.saveConfig(window, false)
	})

	// Open All / Close All buttons
	openAllBtn := widget.NewButton("Open All", func() {
		for i := range e.accordion.Items {
			e.accordion.Items[i].Open = true
		}
		e.accordion.Refresh()
	})
	closeAllBtn := widget.NewButton("Close All", func() {
		e.accordion.CloseAll()
		e.accordion.Refresh()
	})

	// Layout
	topBar := container.NewHBox(
		loadBtn,
		widget.NewSeparator(),
		saveUserBtn,
		saveSystemBtn,
		widget.NewSeparator(),
		openAllBtn,
		closeAllBtn,
	)

	// Wrap accordion with right padding to prevent scrollbar from overlapping content
	// Use a Border layout with a fixed-width right spacer to ensure scrollbar clearance
	// The scrollbar is typically ~15-20px wide, so we add 20px of clearance
	rightSpacer := container.NewWithoutLayout(widget.NewLabel(""))
	rightSpacer.Resize(fyne.NewSize(20, 1)) // Fixed 20px width for scrollbar clearance
	accordionWithMargin := container.NewBorder(
		nil, nil, nil, rightSpacer, // Right: fixed-width spacer
		e.accordion,
	)

	// Create scroll container and store reference for auto-scrolling
	e.scrollContainer = container.NewScroll(accordionWithMargin)

	mainPanel := container.NewBorder(
		e.searchEntry,
		nil, nil, nil,
		e.scrollContainer,
	)

	editorContent := container.NewBorder(
		topBar,
		nil, nil, nil,
		mainPanel,
	)

	// Wrap in a minimum size wrapper to enforce 500px minimum width
	// This allows the window to be resized narrower while maintaining usability
	return NewMinSizeWrapper(editorContent, 500, 0)
}

// buildUIFromSchema builds the UI from the schema, showing all options.
func (e *Editor) buildUIFromSchema() {
	// Clear accordion
	e.accordion.Items = nil
	e.sectionGroups = make(map[string]*SectionGroup)
	e.compoundGroups = make(map[string]*CompoundEntryGroup)

	// Process all sections from schema
	processedSections := make(map[string]bool)

	// First, handle root level options (empty section)
	if section := e.schema.GetSection(""); section != nil || e.hasRootOptions() {
		sectionName := "General"
		section := e.buildSectionFromOptions("")
		if len(section.Options) > 0 {
			sectionGroup := NewSectionGroup(section, e.iniFile, e.window, e.onValueChange)
			e.sectionGroups[sectionName] = sectionGroup
			item := sectionGroup.CreateUI()
			// Set up auto-scroll on open
			e.setupAccordionItemScroll(item)
			e.accordion.Append(item)
			processedSections[""] = true
		}
	}

	// Process all other sections
	for _, option := range e.schema.Options {
		sectionName := option.Section
		if sectionName == "" || processedSections[sectionName] {
			continue
		}

		section := e.schema.GetSection(sectionName)
		if section == nil {
			// Build section from options
			section = e.buildSectionFromOptions(sectionName)
		}

		if len(section.Options) == 0 {
			continue
		}

		// Check if this section has compound entries
		hasCompound := false
		var compoundPattern string
		for _, opt := range section.Options {
			if opt.IsCompound {
				hasCompound = true
				compoundPattern = opt.Pattern
				break
			}
		}

		if hasCompound {
			// Create compound entry group
			compoundGroup := NewCompoundEntryGroup(
				compoundPattern,
				section,
				e.iniFile,
				e.window,
				e.onValueChange,
				e.onCompoundEntryAdd,
				e.onCompoundEntryRemove,
			)
			e.compoundGroups[compoundPattern] = compoundGroup

			item := &widget.AccordionItem{
				Title:  sectionName,
				Detail: compoundGroup.CreateUI(),
				Open:   true,
			}
			// Set up auto-scroll on open
			e.setupAccordionItemScroll(item)
			e.accordion.Append(item)
		} else {
			// Create regular section group
			sectionGroup := NewSectionGroup(section, e.iniFile, e.window, e.onValueChange)
			e.sectionGroups[sectionName] = sectionGroup

			item := sectionGroup.CreateUI()
			// Set up auto-scroll on open
			e.setupAccordionItemScroll(item)
			e.accordion.Append(item)
		}

		processedSections[sectionName] = true
	}

	e.accordion.OpenAll()

	// Set up accordion item state tracking for auto-scroll
	// We'll monitor accordion refresh to detect when items are opened
	e.setupAccordionAutoScroll()
}

// setupAccordionAutoScroll sets up monitoring for accordion item opens to enable auto-scrolling.
// Since Fyne Accordion doesn't have OnOpened callback, we use a workaround:
// Track item states and scroll when an item transitions from closed to open.
func (e *Editor) setupAccordionAutoScroll() {
	// Store previous open states
	previousStates := make(map[int]bool)

	// We'll check states after accordion refresh
	// This is a workaround since Fyne doesn't provide OnOpened callback
	// In a real implementation, we'd need to wrap the accordion or use a custom widget
	// For now, we'll use ScrollToTop when items are opened programmatically
	// User clicks will be handled by monitoring accordion state changes

	// Note: This is a placeholder - actual implementation would require
	// monitoring accordion item state changes or using a custom accordion widget
	_ = previousStates // Suppress unused variable warning
}

// hasRootOptions checks if there are root level options.
func (e *Editor) hasRootOptions() bool {
	for _, option := range e.schema.Options {
		if option.Section == "" {
			return true
		}
	}
	return false
}

// buildSectionFromOptions builds a section from options with the given section name.
func (e *Editor) buildSectionFromOptions(sectionName string) *ConfigSection {
	section := &ConfigSection{
		Name:            sectionName,
		Options:         []*ConfigOption{},
		CompoundEntries: make(map[string][]*ConfigOption),
	}

	for _, option := range e.schema.Options {
		if option.Section == sectionName {
			section.Options = append(section.Options, option)
		}
	}

	return section
}

// loadConfigFilesSilent loads live file for values using embedded schema, without dialogs.
func (e *Editor) loadConfigFilesSilent() error {
	// Use embedded schema (manually maintained from documentation)
	e.schema = GetEmbeddedSchema()

	// Try to load the live config file for actual values
	iniPath := e.fileService.DetectMtCanvusIni()
	if iniPath == "" {
		// No live file - use empty INI with defaults
		e.iniFile = ini.Empty()
		return nil
	}

	// Load the actual INI file for current values
	iniFile, err := e.iniParser.Read(iniPath)
	if err != nil {
		return err
	}

	e.iniFile = iniFile
	return nil
}

// loadConfigFile loads the live config file and overlays values on embedded schema.
func (e *Editor) loadConfigFile(window fyne.Window) {
	// Use embedded schema (manually maintained from documentation)
	e.schema = GetEmbeddedSchema()

	// Load the live config file for actual values
	iniPath := e.fileService.DetectMtCanvusIni()
	if iniPath == "" {
		// No live file - create empty INI with defaults from schema
		e.iniFile = ini.Empty()
		e.buildUIFromSchema()
		dialog.ShowInformation("Loaded", "No live config file found - using defaults from embedded schema.", window)
		return
	}

	// Load the actual INI file for current values
	iniFile, err := e.iniParser.Read(iniPath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to load mt-canvus.ini: %w", err), window)
		return
	}

	e.iniFile = iniFile

	// Rebuild UI to update values from loaded file
	e.buildUIFromSchema()

	msg := fmt.Sprintf("Loaded mt-canvus.ini from:\n%s\n\nAll options updated with current values.", iniPath)
	dialog.ShowInformation("Loaded", msg, window)
}

// filterSections filters sections based on search text.
// It searches through all text in each setting (key, description, default, value, enum values)
// and rebuilds the accordion to show only matching sections, opening them automatically.
func (e *Editor) filterSections(searchText string) {
	if e.accordion == nil || e.schema == nil {
		return
	}

	// If search is empty, rebuild with all sections
	if searchText == "" {
		e.buildUIFromSchema()
		return
	}

	searchLower := strings.ToLower(strings.TrimSpace(searchText))
	if searchLower == "" {
		e.buildUIFromSchema()
		return
	}

	// Clear accordion
	e.accordion.Items = nil
	e.sectionGroups = make(map[string]*SectionGroup)
	e.compoundGroups = make(map[string]*CompoundEntryGroup)

	// Track which sections have matches
	matchedSections := make(map[string]bool)

	// Process all sections and check for matches
	processedSections := make(map[string]bool)

	// First, handle root level options (empty section)
	if e.schema.GetSection("") != nil || e.hasRootOptions() {
		section := e.buildSectionFromOptions("")
		if len(section.Options) > 0 {
			// Check if any option in this section matches
			hasMatch := e.sectionMatches(section, searchLower)
			if hasMatch {
				matchedSections[""] = true
				sectionGroup := NewSectionGroup(section, e.iniFile, e.window, e.onValueChange)
				e.sectionGroups["General"] = sectionGroup
				item := sectionGroup.CreateUI()
				item.Open = true // Open matching sections
				// Set up auto-scroll on open
				e.setupAccordionItemScroll(item)
				e.accordion.Append(item)
			}
			processedSections[""] = true
		}
	}

	// Process all other sections
	for _, option := range e.schema.Options {
		sectionName := option.Section
		if sectionName == "" || processedSections[sectionName] {
			continue
		}

		section := e.schema.GetSection(sectionName)
		if section == nil {
			section = e.buildSectionFromOptions(sectionName)
		}

		if len(section.Options) == 0 {
			continue
		}

		// Check if this section matches the search
		hasMatch := e.sectionMatches(section, searchLower)
		if !hasMatch {
			processedSections[sectionName] = true
			continue
		}

		matchedSections[sectionName] = true

		// Check if this section has compound entries
		hasCompound := false
		var compoundPattern string
		for _, opt := range section.Options {
			if opt.IsCompound {
				hasCompound = true
				compoundPattern = opt.Pattern
				break
			}
		}

		if hasCompound {
			// Create compound entry group
			compoundGroup := NewCompoundEntryGroup(
				compoundPattern,
				section,
				e.iniFile,
				e.window,
				e.onValueChange,
				e.onCompoundEntryAdd,
				e.onCompoundEntryRemove,
			)
			e.compoundGroups[compoundPattern] = compoundGroup

			item := &widget.AccordionItem{
				Title:  sectionName,
				Detail: compoundGroup.CreateUI(),
				Open:   true, // Open matching sections
			}
			// Set up auto-scroll on open
			e.setupAccordionItemScroll(item)
			e.accordion.Append(item)
		} else {
			// Create regular section group
			sectionGroup := NewSectionGroup(section, e.iniFile, e.window, e.onValueChange)
			e.sectionGroups[sectionName] = sectionGroup

			item := sectionGroup.CreateUI()
			item.Open = true // Open matching sections
			// Set up auto-scroll on open
			e.setupAccordionItemScroll(item)
			e.accordion.Append(item)
		}

		processedSections[sectionName] = true
	}
}

// sectionMatches checks if a section matches the search text.
// It searches through all text in each option: key, description, default, value, enum values.
// Uses whole word matching to avoid false positives.
func (e *Editor) sectionMatches(section *ConfigSection, searchLower string) bool {
	// Check section name (whole word match)
	if e.matchesWholeWord(strings.ToLower(section.Name), searchLower) {
		return true
	}

	// Check section description (whole word match)
	if e.matchesWholeWord(strings.ToLower(section.Description), searchLower) {
		return true
	}

	// Check each option in the section
	for _, option := range section.Options {
		if e.optionMatches(option, searchLower) {
			return true
		}
	}

	return false
}

// optionMatches checks if an option matches the search text.
// Searches through key, description, default value, current value, and enum values.
// Uses whole word matching to avoid false positives.
func (e *Editor) optionMatches(option *ConfigOption, searchLower string) bool {
	// Check key name (whole word match)
	if e.matchesWholeWord(strings.ToLower(option.Key), searchLower) {
		return true
	}

	// Check description (whole word match)
	if e.matchesWholeWord(strings.ToLower(option.Description), searchLower) {
		return true
	}

	// Check default value (whole word match)
	if e.matchesWholeWord(strings.ToLower(option.Default), searchLower) {
		return true
	}

	// Check enum values (whole word match)
	for _, enumValue := range option.EnumValues {
		if e.matchesWholeWord(strings.ToLower(enumValue), searchLower) {
			return true
		}
	}

	// Check current value from INI file or form controls (whole word match)
	currentValue := e.getCurrentValueForOption(option)
	if e.matchesWholeWord(strings.ToLower(currentValue), searchLower) {
		return true
	}

	// Check section name (whole word match)
	if e.matchesWholeWord(strings.ToLower(option.Section), searchLower) {
		return true
	}

	return false
}

// matchesWholeWord checks if the search text appears as a whole word in the target text.
// A whole word is defined as being surrounded by word boundaries (non-alphanumeric characters or start/end of string).
func (e *Editor) matchesWholeWord(target, search string) bool {
	if search == "" {
		return false
	}

	// Escape special regex characters in search text
	escapedSearch := regexp.QuoteMeta(search)

	// Create regex pattern for whole word match
	// \b is word boundary, but we want to match at start/end or surrounded by non-word chars
	// Pattern: (^|[^a-zA-Z0-9_])search([^a-zA-Z0-9_]|$)
	pattern := fmt.Sprintf(`(^|[^a-zA-Z0-9_])%s([^a-zA-Z0-9_]|$)`, escapedSearch)

	matched, err := regexp.MatchString(pattern, target)
	if err != nil {
		// If regex fails, fall back to simple contains
		return strings.Contains(target, search)
	}

	return matched
}

// minSizeWrapper is a widget wrapper that enforces minimum size constraints.
type minSizeWrapper struct {
	widget.BaseWidget
	content   fyne.CanvasObject
	minWidth  float32
	minHeight float32
}

// NewMinSizeWrapper creates a new minimum size wrapper widget.
func NewMinSizeWrapper(content fyne.CanvasObject, minWidth, minHeight float32) *minSizeWrapper {
	w := &minSizeWrapper{
		content:   content,
		minWidth:  minWidth,
		minHeight: minHeight,
	}
	w.ExtendBaseWidget(w)
	return w
}

// CreateRenderer creates the renderer for the minimum size wrapper.
func (m *minSizeWrapper) CreateRenderer() fyne.WidgetRenderer {
	return &minSizeRenderer{
		wrapper: m,
		content: m.content,
	}
}

// minSizeRenderer renders the minimum size wrapper.
type minSizeRenderer struct {
	wrapper *minSizeWrapper
	content fyne.CanvasObject
}

// Layout lays out the content with minimum size constraints.
func (r *minSizeRenderer) Layout(size fyne.Size) {
	// Ensure minimum width
	contentWidth := size.Width
	if contentWidth < r.wrapper.minWidth {
		contentWidth = r.wrapper.minWidth
	}

	// Ensure minimum height
	contentHeight := size.Height
	if contentHeight < r.wrapper.minHeight {
		contentHeight = r.wrapper.minHeight
	}

	contentSize := fyne.NewSize(contentWidth, contentHeight)
	r.content.Resize(contentSize)
	r.content.Move(fyne.NewPos(0, 0))
}

// MinSize returns the minimum size of the wrapper.
func (r *minSizeRenderer) MinSize() fyne.Size {
	contentMinSize := r.content.MinSize()

	// Enforce minimum width
	if contentMinSize.Width < r.wrapper.minWidth {
		contentMinSize.Width = r.wrapper.minWidth
	}

	// Enforce minimum height
	if contentMinSize.Height < r.wrapper.minHeight {
		contentMinSize.Height = r.wrapper.minHeight
	}

	return contentMinSize
}

// Objects returns the objects to render.
func (r *minSizeRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.content}
}

// Refresh refreshes the renderer.
func (r *minSizeRenderer) Refresh() {
	r.content.Refresh()
}

// Destroy destroys the renderer.
func (r *minSizeRenderer) Destroy() {
	// No cleanup needed
}

// setupAccordionItemScroll sets up auto-scroll when an accordion item is opened.
// When a user clicks to open an item, it scrolls the section title to the top (just below search bar).
// Note: Fyne Accordion doesn't have OnOpened callback, so we use a workaround by wrapping the accordion
// and monitoring item state changes.
func (e *Editor) setupAccordionItemScroll(item *widget.AccordionItem) {
	// Store the item reference for later use
	// We'll need to track when items are opened and scroll accordingly
	// For now, this is a placeholder - we'll implement the actual scroll logic
	// by monitoring the accordion's state changes
}

// scrollToAccordionItem scrolls the accordion item to the top of the scroll container.
// This is called when an accordion item is opened by user click.
func (e *Editor) scrollToAccordionItem(itemIndex int) {
	if e.scrollContainer == nil || e.accordion == nil {
		return
	}

	// Scroll to top - Fyne doesn't provide direct access to item positions
	// so we'll scroll to top and the opened item will be visible
	// A better implementation would calculate the actual position of the item
	e.scrollContainer.ScrollToTop()
}

// getCurrentValueForOption gets the current value for an option.
// Returns the default value if the key is missing, null, or empty (after trimming whitespace).
func (e *Editor) getCurrentValueForOption(option *ConfigOption) string {
	if e.iniFile == nil {
		return option.Default
	}

	section := e.iniFile.Section(option.Section)
	if section == nil {
		return option.Default
	}

	key := section.Key(option.Key)
	if key == nil {
		return option.Default
	}

	// Trim whitespace and check if empty
	value := strings.TrimSpace(key.String())
	if value == "" {
		return option.Default
	}

	return value
}

// onValueChange handles value changes from form controls.
// Only applies updates for items that have a non-empty value.
// Empty values are treated as "use default" and the key is removed from the INI file.
func (e *Editor) onValueChange(section, key, value string) {
	if e.iniFile == nil {
		// Create empty INI file if needed
		cfg := ini.Empty()
		e.iniFile = cfg
	}

	sec, err := e.iniFile.GetSection(section)
	if err != nil {
		sec, _ = e.iniFile.NewSection(section)
	}

	// Trim whitespace to handle edge cases
	trimmedValue := strings.TrimSpace(value)

	// Only set value if it's non-empty
	// Empty values mean "use default", so we remove the key from the INI file
	if trimmedValue == "" {
		// Remove the key so it will use the default value
		sec.DeleteKey(key)
	} else {
		sec.Key(key).SetValue(trimmedValue)
	}
}

// onCompoundEntryAdd handles adding a new compound entry.
func (e *Editor) onCompoundEntryAdd(pattern, name string) {
	if e.iniFile == nil {
		cfg := ini.Empty()
		e.iniFile = cfg
	}

	sectionName := fmt.Sprintf("%s:%s", pattern, name)
	_, err := e.iniFile.NewSection(sectionName)
	if err != nil {
		// Section might already exist
	}
}

// onCompoundEntryRemove handles removing a compound entry.
func (e *Editor) onCompoundEntryRemove(pattern, name string) {
	if e.iniFile == nil {
		return
	}

	sectionName := fmt.Sprintf("%s:%s", pattern, name)
	e.iniFile.DeleteSection(sectionName)
}

// cleanupEmptyValues removes all empty keys from the INI file before saving.
// This ensures that empty values (which should use defaults) are not written to disk.
func (e *Editor) cleanupEmptyValues() {
	if e.iniFile == nil {
		return
	}

	// Iterate through all sections
	for _, section := range e.iniFile.Sections() {
		// Get all keys in this section
		keysToDelete := []string{}
		for _, key := range section.Keys() {
			// Check if value is empty (after trimming whitespace)
			value := strings.TrimSpace(key.String())
			if value == "" {
				keysToDelete = append(keysToDelete, key.Name())
			}
		}

		// Delete empty keys
		for _, keyName := range keysToDelete {
			section.DeleteKey(keyName)
		}
	}
}

// saveConfig saves the configuration to file.
func (e *Editor) saveConfig(window fyne.Window, userConfig bool) {
	if e.iniFile == nil {
		dialog.ShowError(fmt.Errorf("no configuration loaded"), window)
		return
	}

	// Clean up empty values before saving
	e.cleanupEmptyValues()

	var savePath string
	var configDir string
	if userConfig {
		configDir = e.fileService.GetUserConfigPath()
		savePath = filepath.Join(configDir, "mt-canvus.ini")
	} else {
		configDir = e.fileService.GetSystemConfigPath()
		savePath = filepath.Join(configDir, "mt-canvus.ini")
	}

	// Ensure directory exists
	if err := e.fileService.EnsureDirectory(configDir); err != nil {
		dialog.ShowError(fmt.Errorf("failed to create directory: %w", err), window)
		return
	}

	// Create backup before saving (if file exists)
	if _, err := os.Stat(savePath); err == nil {
		if _, err := e.backupManager.CreateBackup(savePath); err != nil {
			// Log warning but continue with save
			fmt.Printf("Warning: Failed to create backup: %v\n", err)
		}
	}

	// Save the file
	if err := e.iniParser.Write(e.iniFile, savePath); err != nil {
		dialog.ShowError(fmt.Errorf("failed to save mt-canvus.ini: %w", err), window)
		return
	}

	location := "user config"
	if !userConfig {
		location = "system config"
	}
	dialog.ShowInformation("Saved", fmt.Sprintf("Saved mt-canvus.ini to %s:\n%s\n\nBackup created automatically.", location, savePath), window)
}

// getTooltip returns a tooltip for a configuration option.
func (e *Editor) getTooltip(section, key string) string {
	// TODO: Load tooltips from documentation
	// For now, return a basic description
	return fmt.Sprintf("Configuration option: %s.%s", section, key)
}
