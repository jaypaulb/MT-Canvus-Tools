package configeditor

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gopkg.in/ini.v1"
)

// CompoundEntryGroup represents a UI group for compound entries (e.g., [server:name], [remote-desktop:name]).
type CompoundEntryGroup struct {
	pattern          string
	section          *ConfigSection
	iniFile          *ini.File
	window           fyne.Window
	entries          map[string]*CompoundEntry // Map of entry name to entry
	entriesContainer *fyne.Container           // Container for entries (stored for dynamic addition)
	onValueChange    func(section, key, value string)
	onEntryAdd       func(pattern, name string)
	onEntryRemove    func(pattern, name string)
}

// CompoundEntry represents a single compound entry instance.
type CompoundEntry struct {
	name         string
	pattern      string
	options      []*ConfigOption
	formControls map[string]*FormControl
	container    *fyne.Container
}

// NewCompoundEntryGroup creates a new compound entry group.
func NewCompoundEntryGroup(pattern string, section *ConfigSection, iniFile *ini.File, window fyne.Window, onValueChange func(section, key, value string), onEntryAdd, onEntryRemove func(pattern, name string)) *CompoundEntryGroup {
	return &CompoundEntryGroup{
		pattern:       pattern,
		section:       section,
		iniFile:       iniFile,
		window:        window,
		entries:       make(map[string]*CompoundEntry),
		onValueChange: onValueChange,
		onEntryAdd:    onEntryAdd,
		onEntryRemove: onEntryRemove,
	}
}

// CreateUI creates the UI for the compound entry group.
func (ceg *CompoundEntryGroup) CreateUI() fyne.CanvasObject {
	// Title
	title := widget.NewLabel(fmt.Sprintf("%s Entries", ceg.pattern))
	title.TextStyle.Bold = true

	// Add button
	addBtn := widget.NewButton(fmt.Sprintf("Add New %s", ceg.pattern), func() {
		ceg.showAddEntryDialog()
	})

	// Container for entries - no scroll, let parent accordion handle scrolling
	entriesContainer := container.NewVBox()
	ceg.entriesContainer = entriesContainer // Store reference for dynamic addition

	// Load existing entries from INI file
	if ceg.iniFile != nil {
		for _, section := range ceg.iniFile.Sections() {
			sectionName := section.Name()
			// Check if this section matches the pattern (e.g., "server:name" or "remote-desktop:name")
			if ceg.matchesPattern(sectionName) {
				entryName := ceg.extractEntryName(sectionName)
				if entryName != "" {
					ceg.addEntry(entryName, entriesContainer)
				}
			}
		}
	}

	// Don't create default entries - only show entries that actually exist in the INI file
	// Users can add new entries using the "Add New" button if needed

	// Layout: [tab] Section Name [tab] [Add New Button]
	// Tab spacing (16px default Fyne padding) for sub-section indentation
	tabSpacer := widget.NewLabel("") // Spacer for tab
	tabSpacerPadded := container.NewPadded(tabSpacer) // Apply padding for 16px spacing
	titleWithTab := container.NewBorder(
		nil, nil,
		tabSpacerPadded, // Left: tab spacer (16px via padding)
		nil,
		title, // Center: title
	)
	titleRow := container.NewBorder(
		nil, nil,
		titleWithTab, // Left: tab + title
		nil,
		addBtn, // Right: Add New button
	)

	// No scroll container - let parent accordion handle scrolling
	// This prevents scrollbars and allows scrolling when hovering over entry fields
	return container.NewVBox(
		titleRow,
		entriesContainer,
	)
}

// matchesPattern checks if a section name matches the compound pattern.
func (ceg *CompoundEntryGroup) matchesPattern(sectionName string) bool {
	// Check if section name starts with pattern:
	// "server:name" -> pattern "server"
	// "remote-desktop:name" -> pattern "remote-desktop"
	return len(sectionName) > len(ceg.pattern) &&
		   sectionName[:len(ceg.pattern)] == ceg.pattern &&
		   sectionName[len(ceg.pattern)] == ':'
}

// extractEntryName extracts the entry name from a section name.
func (ceg *CompoundEntryGroup) extractEntryName(sectionName string) string {
	// "server:My Server" -> "My Server"
	// "remote-desktop:Connection 1" -> "Connection 1"
	prefix := ceg.pattern + ":"
	if len(sectionName) > len(prefix) {
		return sectionName[len(prefix):]
	}
	return ""
}

// addEntry adds a new compound entry to the UI.
func (ceg *CompoundEntryGroup) addEntry(name string, parent *fyne.Container) {
	entry := &CompoundEntry{
		name:         name,
		pattern:      ceg.pattern,
		options:      ceg.getCompoundOptions(),
		formControls: make(map[string]*FormControl),
	}

	// Create form for options - using same layout as SectionGroup
	form := container.NewVBox()

	// Add all option fields with horizontal layout (label + info button on left, control on right)
	for _, option := range entry.options {
		// Get current value from INI
		currentValue := ceg.getCurrentValue(option, name)

		formControl := CreateFormControl(option, ceg.window, currentValue)
		entry.formControls[option.Key] = formControl

		// Create label with clickable name for tooltip (same as SectionGroup)
		labelText := widget.NewLabel(option.Key)
		labelText.TextStyle.Bold = true

		// Create tooltip with full description and default
		tooltipText := ceg.buildTooltip(option)

		// Make label clickable to show tooltip in a popup
		infoBtn := widget.NewButton("ℹ", func() {
			dialog.ShowInformation(option.Key, tooltipText, ceg.window)
		})
		infoBtn.Importance = widget.LowImportance

		// Layout: [tab][tab] Setting | Input
		// Double tab spacing (32px total) for entry-level indentation
		labelContainer := container.NewHBox(labelText, infoBtn)

		// Use Border layout with double left padded spacer for double tab indentation
		tabSpacer1 := widget.NewLabel("") // First tab
		tabSpacer2 := widget.NewLabel("") // Second tab
		tabSpacer1Padded := container.NewPadded(tabSpacer1) // Apply padding for 16px spacing
		tabSpacer2Padded := container.NewPadded(tabSpacer2) // Apply padding for 16px spacing
		// Create nested Border layouts for double tab spacing
		innerTab := container.NewBorder(
			nil, nil,
			tabSpacer1Padded, // Left: first tab (16px via padding)
			nil,
			labelContainer, // Center: label + info
		)
		outerTab := container.NewBorder(
			nil, nil,
			tabSpacer2Padded, // Left: second tab (16px via padding)
			nil,
			innerTab, // Center: first tab + label + info
		)
		row := container.NewGridWithColumns(2,
			outerTab,            // Left: double tab + label + info button
			formControl.Control, // Right: input control
		)

		form.Add(row)

		// Set up change handler
		ceg.setupChangeHandler(option, formControl, name)
	}

	// Layout: [tab][tab] Entry Name | Remove Button
	// Double tab spacing (32px total) for entry-level indentation
	entryTitle := widget.NewLabel(fmt.Sprintf("%s: %s", ceg.pattern, name))
	entryTitle.TextStyle.Bold = true

	// Remove button
	removeBtn := widget.NewButton("Remove", func() {
		if ceg.onEntryRemove != nil {
			ceg.onEntryRemove(ceg.pattern, entry.name)
		}
		parent.Remove(entry.container)
		delete(ceg.entries, entry.name)
	})

	// Use Border layout with double left padded spacer for double tab indentation
	tabSpacer1 := widget.NewLabel("") // First tab
	tabSpacer2 := widget.NewLabel("") // Second tab
	tabSpacer1Padded := container.NewPadded(tabSpacer1) // Apply padding for 16px spacing
	tabSpacer2Padded := container.NewPadded(tabSpacer2) // Apply padding for 16px spacing
	// Create nested Border layouts for double tab spacing
	innerTab := container.NewBorder(
		nil, nil,
		tabSpacer1Padded, // Left: first tab (16px via padding)
		nil,
		entryTitle, // Center: entry title
	)
	outerTab := container.NewBorder(
		nil, nil,
		tabSpacer2Padded, // Left: second tab (16px via padding)
		nil,
		innerTab, // Center: first tab + entry title
	)

	// Title bar: [tab][tab] Entry Name | Remove Button
	titleBar := container.NewBorder(
		nil, nil,
		outerTab,  // Left: double tab + entry title
		nil,
		removeBtn, // Right: Remove button
	)

	// Frame for entry with padding: larger left margin, smaller right margin
	// Use Border layout with spacing widgets for asymmetric padding
	// Left spacer (larger) and right spacer (smaller)
	// Note: Fyne doesn't support direct size setting on empty containers,
	// so we use Border with the content and apply padding via the frame
	entryContent := container.NewBorder(
		titleBar,
		nil,
		nil, // Left padding will be handled by outer container
		nil, // Right padding will be handled by outer container
		form,
	)

	// Wrap in padded container for visual frame with asymmetric padding
	// The padding will be applied uniformly, but we can use Border for asymmetric spacing
	// For now, use uniform padding - the visual frame provides the separation
	entryFrame := container.NewPadded(entryContent)

	entry.container = entryFrame
	ceg.entries[name] = entry

	parent.Add(entryFrame)
	parent.Add(widget.NewSeparator())
}

// buildTooltip builds a comprehensive tooltip with full description and default value.
func (ceg *CompoundEntryGroup) buildTooltip(option *ConfigOption) string {
	var tooltip strings.Builder

	// Add full description
	if option.Description != "" {
		tooltip.WriteString(option.Description)
		tooltip.WriteString("\n\n")
	}

	// Add default value
	if option.Default != "" {
		tooltip.WriteString(fmt.Sprintf("Default: %s", option.Default))
	} else {
		tooltip.WriteString("Default: (empty)")
	}

	// Add enum values if applicable
	if len(option.EnumValues) > 0 {
		tooltip.WriteString(fmt.Sprintf("\n\nAvailable values: %s", strings.Join(option.EnumValues, ", ")))
	}

	// Add type information
	tooltip.WriteString(fmt.Sprintf("\n\nType: %s", option.Type))

	return tooltip.String()
}

// setupChangeHandler sets up the change handler for a form control.
func (ceg *CompoundEntryGroup) setupChangeHandler(option *ConfigOption, formControl *FormControl, entryName string) {
	// Wire up change handlers based on control type
	switch ctrl := formControl.Control.(type) {
	case *widget.Entry:
		ctrl.OnChanged = func(text string) {
			if ceg.onValueChange != nil {
				sectionName := fmt.Sprintf("%s:%s", ceg.pattern, entryName)
				ceg.onValueChange(sectionName, option.Key, formControl.GetValue())
			}
		}
	case *widget.Check:
		ctrl.OnChanged = func(checked bool) {
			if ceg.onValueChange != nil {
				sectionName := fmt.Sprintf("%s:%s", ceg.pattern, entryName)
				ceg.onValueChange(sectionName, option.Key, formControl.GetValue())
			}
		}
	case *widget.Select:
		ctrl.OnChanged = func(selected string) {
			if ceg.onValueChange != nil {
				sectionName := fmt.Sprintf("%s:%s", ceg.pattern, entryName)
				ceg.onValueChange(sectionName, option.Key, formControl.GetValue())
			}
		}
	case *fyne.Container:
		// For file path controls with browse button, find the entry
		for _, obj := range ctrl.Objects {
			if entry, ok := obj.(*widget.Entry); ok {
				entry.OnChanged = func(text string) {
					if ceg.onValueChange != nil {
						sectionName := fmt.Sprintf("%s:%s", ceg.pattern, entryName)
						ceg.onValueChange(sectionName, option.Key, formControl.GetValue())
					}
				}
			}
		}
	}
}

// getCompoundOptions gets the options for this compound entry type.
func (ceg *CompoundEntryGroup) getCompoundOptions() []*ConfigOption {
	var options []*ConfigOption
	for _, option := range ceg.section.Options {
		if option.IsCompound && option.Pattern == ceg.pattern {
			options = append(options, option)
		}
	}
	return options
}

// getCurrentValue gets the current value for an option from a compound entry.
// Returns the default value if the key is missing, null, or empty (after trimming whitespace).
func (ceg *CompoundEntryGroup) getCurrentValue(option *ConfigOption, entryName string) string {
	if ceg.iniFile == nil || entryName == "" {
		return option.Default
	}

	sectionName := fmt.Sprintf("%s:%s", ceg.pattern, entryName)
	section := ceg.iniFile.Section(sectionName)
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

// showAddEntryDialog shows a dialog to add a new compound entry.
func (ceg *CompoundEntryGroup) showAddEntryDialog() {
	var entryName string

	// For fixed-workspace, auto-generate the next index number
	if ceg.pattern == "fixed-workspace" {
		// Find the highest existing index from both entries map and INI file
		maxIndex := 0

		// Check entries map
		for name := range ceg.entries {
			var index int
			if _, err := fmt.Sscanf(name, "%d", &index); err == nil {
				if index > maxIndex {
					maxIndex = index
				}
			}
		}

		// Check INI file sections
		if ceg.iniFile != nil {
			for _, section := range ceg.iniFile.Sections() {
				sectionName := section.Name()
				if ceg.matchesPattern(sectionName) {
					extractedName := ceg.extractEntryName(sectionName)
					var index int
					if _, err := fmt.Sscanf(extractedName, "%d", &index); err == nil {
						if index > maxIndex {
							maxIndex = index
						}
					}
				}
			}
		}

		// Use next index
		entryName = fmt.Sprintf("%d", maxIndex+1)
	} else {
		// For other patterns, show dialog to enter name
		nameEntry := widget.NewEntry()
		nameEntry.SetPlaceHolder(fmt.Sprintf("Enter %s name", ceg.pattern))

		content := container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Add New %s", ceg.pattern)),
			widget.NewLabel("Name:"),
			nameEntry,
		)

		dialog.ShowCustomConfirm(
			fmt.Sprintf("Add %s", ceg.pattern),
			"Add",
			"Cancel",
			content,
			func(confirmed bool) {
				if confirmed && nameEntry.Text != "" {
					entryName = nameEntry.Text
					// Add to INI file via callback
					if ceg.onEntryAdd != nil {
						ceg.onEntryAdd(ceg.pattern, entryName)
					}
					// Add to UI immediately
					if ceg.entriesContainer != nil {
						ceg.addEntry(entryName, ceg.entriesContainer)
					}
				}
			},
			ceg.window,
		)
		return
	}

	// For fixed-workspace, add immediately without dialog
	if entryName != "" {
		// Add to INI file via callback
		if ceg.onEntryAdd != nil {
			ceg.onEntryAdd(ceg.pattern, entryName)
		}
		// Add to UI immediately
		if ceg.entriesContainer != nil {
			ceg.addEntry(entryName, ceg.entriesContainer)
		}
	}
}

// GetEntries returns all compound entries.
func (ceg *CompoundEntryGroup) GetEntries() map[string]map[string]string {
	result := make(map[string]map[string]string)
	for name, entry := range ceg.entries {
		values := make(map[string]string)
		for key, formControl := range entry.formControls {
			values[key] = formControl.GetValue()
		}
		result[name] = values
	}
	return result
}

