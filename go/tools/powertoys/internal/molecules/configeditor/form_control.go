package configeditor

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// FormControl represents a form control for a configuration option.
type FormControl struct {
	Option      *ConfigOption
	Control     fyne.CanvasObject
	GetValue    func() string
	SetValue    func(string)
	Validate    func() error
}

// CreateFormControl creates an appropriate form control for a configuration option.
func CreateFormControl(option *ConfigOption, window fyne.Window, currentValue string) *FormControl {
	fc := &FormControl{
		Option: option,
	}

	switch option.Type {
	case ValueTypeBoolean:
		check := widget.NewCheck("", nil)
		check.SetChecked(currentValue == "true" || (currentValue == "" && option.Default == "true"))
		fc.Control = check
		fc.GetValue = func() string {
			if check.Checked {
				return "true"
			}
			return "false"
		}
		fc.SetValue = func(v string) {
			check.SetChecked(v == "true")
		}

	case ValueTypeEnum:
		selectWidget := widget.NewSelect(option.EnumValues, nil)
		if currentValue != "" {
			selectWidget.SetSelected(currentValue)
		} else if option.Default != "" {
			selectWidget.SetSelected(option.Default)
		} else if len(option.EnumValues) > 0 {
			selectWidget.SetSelected(option.EnumValues[0])
		}
		fc.Control = selectWidget
		fc.GetValue = func() string {
			return selectWidget.Selected
		}
		fc.SetValue = func(v string) {
			selectWidget.SetSelected(v)
		}

	case ValueTypeFilePath:
		entry := widget.NewEntry()
		entry.SetText(currentValue)
		// Auto-validate: replace \ with /
		entry.OnChanged = func(text string) {
			normalized := strings.ReplaceAll(text, "\\", "/")
			if normalized != text {
				entry.SetText(normalized)
			}
		}

		browseBtn := widget.NewButton("Browse...", func() {
			dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil || reader == nil {
					return
				}
				defer reader.Close()
				uri := reader.URI()
				var path string
				if uri.Scheme() == "file" {
					path = uri.Path()
				} else {
					path = uri.String()
				}
				// Normalize path separators
				path = strings.ReplaceAll(path, "\\", "/")
				entry.SetText(path)
			}, window)
		})

		// NOTE: Entry widgets in Fyne capture scroll events when hovered.
		// See ValueTypeCommaList case for details on attempted fixes.
		box := container.NewBorder(nil, nil, nil, browseBtn, entry)
		fc.Control = box
		fc.GetValue = func() string {
			text := entry.Text
			return strings.ReplaceAll(text, "\\", "/")
		}
		fc.SetValue = func(v string) {
			entry.SetText(strings.ReplaceAll(v, "\\", "/"))
		}

	case ValueTypeCommaList:
		entry := widget.NewEntry()
		entry.SetText(currentValue)
		// NOTE: Entry widgets in Fyne capture scroll events when hovered.
		// This is a known Fyne limitation. Previous attempts to fix:
		// 1. Removed scroll containers from compound entries
		// 2. Let parent accordion handle scrolling
		// 3. Wrapped entries in containers (didn't work - entries still capture scroll)
		// Next approach would require custom Entry widget or framework changes.
		fc.Control = entry
		fc.GetValue = func() string {
			return entry.Text
		}
		fc.SetValue = func(v string) {
			entry.SetText(v)
		}

	case ValueTypeNumber:
		entry := widget.NewEntry()
		entry.SetText(currentValue)
		// Validate numeric input
		entry.Validator = func(text string) error {
			if text == "" {
				return nil
			}
			// Allow numbers, decimals, and negative
			for _, r := range text {
				if (r < '0' || r > '9') && r != '.' && r != '-' {
					return fmt.Errorf("must be a number")
				}
			}
			return nil
		}
		// NOTE: Entry widgets in Fyne capture scroll events when hovered.
		// See ValueTypeCommaList case for details on attempted fixes.
		fc.Control = entry
		fc.GetValue = func() string {
			return entry.Text
		}
		fc.SetValue = func(v string) {
			entry.SetText(v)
		}

	default: // ValueTypeString
		entry := widget.NewEntry()
		entry.SetText(currentValue)
		// NOTE: Entry widgets in Fyne capture scroll events when hovered.
		// See ValueTypeCommaList case for details on attempted fixes.
		fc.Control = entry
		fc.GetValue = func() string {
			return entry.Text
		}
		fc.SetValue = func(v string) {
			entry.SetText(v)
		}
	}

	// Default validation
	fc.Validate = func() error {
		value := fc.GetValue()
		if option.Required && value == "" {
			return fmt.Errorf("this field is required")
		}
		return nil
	}

	return fc
}

