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

// SectionGroup represents a UI group for a configuration section with expand/collapse.
type SectionGroup struct {
	section      *ConfigSection
	iniFile      *ini.File
	window       fyne.Window
	formControls map[string]*FormControl
	expanded     bool
	content      fyne.CanvasObject
	onValueChange func(section, key, value string)
}

// NewSectionGroup creates a new section group UI component.
func NewSectionGroup(section *ConfigSection, iniFile *ini.File, window fyne.Window, onValueChange func(section, key, value string)) *SectionGroup {
	return &SectionGroup{
		section:       section,
		iniFile:       iniFile,
		window:        window,
		formControls:  make(map[string]*FormControl),
		expanded:      true,
		onValueChange: onValueChange,
	}
}

// CreateUI creates the UI for the section group.
func (sg *SectionGroup) CreateUI() *widget.AccordionItem {
	form := container.NewVBox()

	// Add all options for this section
	for _, option := range sg.section.Options {
		if option.IsCompound {
			// Skip compound options here - they'll be handled separately
			continue
		}

		// Get current value from INI file or use default
		currentValue := sg.getCurrentValue(option)

		// Create form control
		formControl := CreateFormControl(option, sg.window, currentValue)
		sg.formControls[option.Key] = formControl

		// Create label with clickable name for tooltip
		labelText := widget.NewLabel(option.Key)
		labelText.TextStyle.Bold = true

		// Create tooltip with full description and default
		tooltipText := sg.buildTooltip(option)

		// Make label clickable to show tooltip in a popup (for better visibility)
		infoBtn := widget.NewButton("ℹ", func() {
			// Show popup with full information
			dialog.ShowInformation(option.Key, tooltipText, sg.window)
		})
		infoBtn.Importance = widget.LowImportance

		// Create horizontal layout: [tab] setting name | info button | input
		// Tab spacing (16px default Fyne padding) for sub-section indentation
		labelContainer := container.NewHBox(
			labelText,
			infoBtn,
		)

		// Layout: [tab spacer] | [label + info] | [input]
		// Use Border layout with left padded spacer for tab indentation (16px)
		// Then use GridWithColumns for proper alignment: col1=tab+label+info, col2=input
		tabSpacer := widget.NewLabel("") // Empty label as spacer
		tabSpacerPadded := container.NewPadded(tabSpacer) // Apply padding for 16px spacing
		leftSideWithTab := container.NewBorder(
			nil, nil,
			tabSpacerPadded, // Left: tab spacer (16px via padding)
			nil,
			labelContainer, // Center: label + info button
		)
		row := container.NewGridWithColumns(2,
			leftSideWithTab,     // Left: tab + label + info button
			formControl.Control, // Right: input control
		)

		// Add row to form (no separator between items, only at end if needed)
		form.Add(row)

		// Set up change handler
		sg.setupChangeHandler(option, formControl)
	}

	title := sg.section.Name
	if title == "" {
		title = "General"
	}
	if sg.section.Description != "" {
		title = fmt.Sprintf("%s - %s", title, sg.section.Description)
	}

	// Never use scroll - each section should be exactly the size of its items
	// The parent accordion container will handle scrolling
	return &widget.AccordionItem{
		Title:  title,
		Detail: form,
		Open:   true,
	}
}

// buildTooltip builds a comprehensive tooltip with full description and default value.
func (sg *SectionGroup) buildTooltip(option *ConfigOption) string {
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

// getCurrentValue gets the current value for an option from the INI file.
// Returns the default value if the key is missing, null, or empty (after trimming whitespace).
func (sg *SectionGroup) getCurrentValue(option *ConfigOption) string {
	if sg.iniFile == nil {
		return option.Default
	}

	section := sg.iniFile.Section(option.Section)
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

// setupChangeHandler sets up the change handler for a form control.
func (sg *SectionGroup) setupChangeHandler(option *ConfigOption, formControl *FormControl) {
	// Wire up change handlers based on control type
	switch ctrl := formControl.Control.(type) {
	case *widget.Entry:
		ctrl.OnChanged = func(text string) {
			if sg.onValueChange != nil {
				sg.onValueChange(option.Section, option.Key, formControl.GetValue())
			}
		}
	case *widget.Check:
		ctrl.OnChanged = func(checked bool) {
			if sg.onValueChange != nil {
				sg.onValueChange(option.Section, option.Key, formControl.GetValue())
			}
		}
	case *widget.Select:
		ctrl.OnChanged = func(selected string) {
			if sg.onValueChange != nil {
				sg.onValueChange(option.Section, option.Key, formControl.GetValue())
			}
		}
	case *fyne.Container:
		// For file path controls with browse button, find the entry
		for _, obj := range ctrl.Objects {
			if entry, ok := obj.(*widget.Entry); ok {
				entry.OnChanged = func(text string) {
					if sg.onValueChange != nil {
						sg.onValueChange(option.Section, option.Key, formControl.GetValue())
					}
				}
			}
		}
	}
}

// GetValues gets all current values from the form controls.
func (sg *SectionGroup) GetValues() map[string]string {
	values := make(map[string]string)
	for key, formControl := range sg.formControls {
		values[key] = formControl.GetValue()
	}
	return values
}

// SetValues sets values in the form controls.
func (sg *SectionGroup) SetValues(values map[string]string) {
	for key, formControl := range sg.formControls {
		if value, ok := values[key]; ok {
			formControl.SetValue(value)
		}
	}
}

