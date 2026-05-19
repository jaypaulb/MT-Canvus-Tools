package custommenu

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// createCreateActionForm creates a form for "create" action parameters.
func (d *Designer) createCreateActionForm(action *Action) fyne.CanvasObject {
	if action.Parameters == nil {
		action.Parameters = make(map[string]interface{})
	}

	// Content type selector
	contentType := getStringParam(action.Parameters, "type", "note")
	contentTypeSelect := widget.NewSelect(
		[]string{"note", "pdf", "video", "image", "browser"},
		func(selected string) {
			action.Parameters["type"] = selected
			d.showItemForm(d.selectedID)
		},
	)
	contentTypeSelect.SetSelected(contentType)

	// Common parameters
	coordSystemSelect := widget.NewSelect(
		[]string{"viewport", "canvas"},
		func(selected string) {
			action.Parameters["coordinate-system"] = selected
		},
	)
	coordSystemSelect.SetSelected(getStringParam(action.Parameters, "coordinate-system", "viewport"))

	coordOffsetSelect := widget.NewSelect(
		[]string{"none", "finger", "menu"},
		func(selected string) {
			action.Parameters["coordinate-offset"] = selected
		},
	)
	coordOffsetSelect.SetSelected(getStringParam(action.Parameters, "coordinate-offset", "menu"))

	originEntry := widget.NewEntry()
	originEntry.SetText(getStringParam(action.Parameters, "origin", "0 0"))
	originEntry.SetPlaceHolder("0 0 (top-left), 0.5 0.5 (center), 1 1 (bottom-right)")
	originEntry.OnChanged = func(value string) {
		action.Parameters["origin"] = value
	}

	locationEntry := widget.NewEntry()
	locationEntry.SetText(getStringParam(action.Parameters, "location", ""))
	locationEntry.SetPlaceHolder("e.g., 50% 50% or 100px 200px")
	locationEntry.OnChanged = func(value string) {
		action.Parameters["location"] = value
	}

	sizeEntry := widget.NewEntry()
	sizeEntry.SetText(getStringParam(action.Parameters, "size", ""))
	sizeEntry.SetPlaceHolder("e.g., 20% 15% or 300px 200px")
	sizeEntry.OnChanged = func(value string) {
		action.Parameters["size"] = value
	}

	scaleEntry := widget.NewEntry()
	scaleEntry.SetText(fmt.Sprintf("%v", getParam(action.Parameters, "scale", 1.0)))
	scaleEntry.SetPlaceHolder("1.0")
	scaleEntry.OnChanged = func(value string) {
		action.Parameters["scale"] = value
	}

	pinnedCheck := widget.NewCheck("Pinned", func(checked bool) {
		action.Parameters["pinned"] = checked
	})
	pinnedCheck.SetChecked(getBoolParam(action.Parameters, "pinned", false))

	commonForm := container.NewVBox(
		widget.NewLabel("Common Parameters:"),
		widget.NewLabel("Coordinate System:"),
		coordSystemSelect,
		widget.NewLabel("Coordinate Offset:"),
		coordOffsetSelect,
		widget.NewLabel("Origin:"),
		originEntry,
		widget.NewLabel("Location:"),
		locationEntry,
		widget.NewLabel("Size:"),
		sizeEntry,
		widget.NewLabel("Scale:"),
		scaleEntry,
		pinnedCheck,
		widget.NewSeparator(),
	)

	// Content-specific parameters
	var contentForm fyne.CanvasObject
	switch contentType {
	case "note":
		contentForm = d.createNoteForm(action)
	case "pdf":
		contentForm = d.createPdfForm(action)
	case "video":
		contentForm = d.createVideoForm(action)
	case "image":
		contentForm = d.createImageForm(action)
	case "browser":
		contentForm = d.createBrowserForm(action)
	}

	return container.NewVBox(
		widget.NewLabel("Content Type:"),
		contentTypeSelect,
		widget.NewSeparator(),
		commonForm,
		contentForm,
	)
}

// createNoteForm creates a form for note-specific parameters.
func (d *Designer) createNoteForm(action *Action) fyne.CanvasObject {
	textEntry := widget.NewMultiLineEntry()
	textEntry.SetText(getStringParam(action.Parameters, "text", ""))
	textEntry.SetPlaceHolder("Note text content")
	textEntry.OnChanged = func(value string) {
		action.Parameters["text"] = value
	}

	titleEntry := widget.NewEntry()
	titleEntry.SetText(getStringParam(action.Parameters, "title", ""))
	titleEntry.SetPlaceHolder("Note title (optional)")
	titleEntry.OnChanged = func(value string) {
		action.Parameters["title"] = value
	}

	colorEntry := widget.NewEntry()
	colorEntry.SetText(getStringParam(action.Parameters, "color", "#ffffff"))
	colorEntry.SetPlaceHolder("#ffffff, red, transparent")
	colorEntry.OnChanged = func(value string) {
		action.Parameters["color"] = value
	}

	colorPickerBtn := widget.NewButton("Pick Color", func() {
		d.showColorPicker(colorEntry)
	})

	return container.NewVBox(
		widget.NewLabel("Note Parameters:"),
		widget.NewLabel("Text:"),
		textEntry,
		widget.NewLabel("Title:"),
		titleEntry,
		widget.NewLabel("Color:"),
		container.NewBorder(nil, nil, nil, colorPickerBtn, colorEntry),
	)
}

// createPdfForm creates a form for PDF-specific parameters.
func (d *Designer) createPdfForm(action *Action) fyne.CanvasObject {
	sourceEntry := widget.NewEntry()
	sourceEntry.SetText(getStringParam(action.Parameters, "source", ""))
	sourceEntry.SetPlaceHolder("Path to PDF file")
	sourceEntry.OnChanged = func(value string) {
		action.Parameters["source"] = value
	}

	browseBtn := widget.NewButton("Browse...", func() {
		d.browseFile(sourceEntry, []string{".pdf"})
	})

	titleEntry := widget.NewEntry()
	titleEntry.SetText(getStringParam(action.Parameters, "title", ""))
	titleEntry.SetPlaceHolder("PDF title (optional)")
	titleEntry.OnChanged = func(value string) {
		action.Parameters["title"] = value
	}

	return container.NewVBox(
		widget.NewLabel("PDF Parameters:"),
		widget.NewLabel("Source:"),
		container.NewBorder(nil, nil, nil, browseBtn, sourceEntry),
		widget.NewLabel("Title:"),
		titleEntry,
	)
}

// createVideoForm creates a form for video-specific parameters.
func (d *Designer) createVideoForm(action *Action) fyne.CanvasObject {
	sourceEntry := widget.NewEntry()
	sourceEntry.SetText(getStringParam(action.Parameters, "source", ""))
	sourceEntry.SetPlaceHolder("Path to video file")
	sourceEntry.OnChanged = func(value string) {
		action.Parameters["source"] = value
	}

	browseBtn := widget.NewButton("Browse...", func() {
		d.browseFile(sourceEntry, []string{".mp4", ".avi", ".mov", ".mkv"})
	})

	titleEntry := widget.NewEntry()
	titleEntry.SetText(getStringParam(action.Parameters, "title", ""))
	titleEntry.SetPlaceHolder("Video title (optional)")
	titleEntry.OnChanged = func(value string) {
		action.Parameters["title"] = value
	}

	return container.NewVBox(
		widget.NewLabel("Video Parameters:"),
		widget.NewLabel("Source:"),
		container.NewBorder(nil, nil, nil, browseBtn, sourceEntry),
		widget.NewLabel("Title:"),
		titleEntry,
	)
}

// createImageForm creates a form for image-specific parameters.
func (d *Designer) createImageForm(action *Action) fyne.CanvasObject {
	sourceEntry := widget.NewEntry()
	sourceEntry.SetText(getStringParam(action.Parameters, "source", ""))
	sourceEntry.SetPlaceHolder("Path to image file")
	sourceEntry.OnChanged = func(value string) {
		action.Parameters["source"] = value
	}

	browseBtn := widget.NewButton("Browse...", func() {
		d.browseFile(sourceEntry, []string{".png", ".jpg", ".jpeg", ".gif", ".bmp"})
	})

	titleEntry := widget.NewEntry()
	titleEntry.SetText(getStringParam(action.Parameters, "title", ""))
	titleEntry.SetPlaceHolder("Image title (optional)")
	titleEntry.OnChanged = func(value string) {
		action.Parameters["title"] = value
	}

	return container.NewVBox(
		widget.NewLabel("Image Parameters:"),
		widget.NewLabel("Source:"),
		container.NewBorder(nil, nil, nil, browseBtn, sourceEntry),
		widget.NewLabel("Title:"),
		titleEntry,
	)
}

// createBrowserForm creates a form for browser-specific parameters.
func (d *Designer) createBrowserForm(action *Action) fyne.CanvasObject {
	urlEntry := widget.NewEntry()
	urlEntry.SetText(getStringParam(action.Parameters, "url", ""))
	urlEntry.SetPlaceHolder("https://example.com")
	urlEntry.OnChanged = func(value string) {
		action.Parameters["url"] = value
	}

	titleEntry := widget.NewEntry()
	titleEntry.SetText(getStringParam(action.Parameters, "title", ""))
	titleEntry.SetPlaceHolder("Browser title (optional)")
	titleEntry.OnChanged = func(value string) {
		action.Parameters["title"] = value
	}

	transparentCheck := widget.NewCheck("Transparent Mode", func(checked bool) {
		action.Parameters["transparent-mode"] = checked
	})
	transparentCheck.SetChecked(getBoolParam(action.Parameters, "transparent-mode", false))

	return container.NewVBox(
		widget.NewLabel("Browser Parameters:"),
		widget.NewLabel("URL:"),
		urlEntry,
		widget.NewLabel("Title:"),
		titleEntry,
		transparentCheck,
	)
}

// createOpenFolderForm creates a form for open-folder action parameters.
func (d *Designer) createOpenFolderForm(action *Action) fyne.CanvasObject {
	if action.Parameters == nil {
		action.Parameters = make(map[string]interface{})
	}

	sourceEntry := widget.NewEntry()
	sourceEntry.SetText(getStringParam(action.Parameters, "source", ""))
	sourceEntry.SetPlaceHolder("Path to folder")
	sourceEntry.OnChanged = func(value string) {
		action.Parameters["source"] = value
	}

	browseBtn := widget.NewButton("Browse...", func() {
		d.browseFolder(sourceEntry)
	})

	titleEntry := widget.NewEntry()
	titleEntry.SetText(getStringParam(action.Parameters, "title", ""))
	titleEntry.SetPlaceHolder("Folder title")
	titleEntry.OnChanged = func(value string) {
		action.Parameters["title"] = value
	}

	return container.NewVBox(
		widget.NewLabel("Open Folder Parameters:"),
		widget.NewLabel("Source:"),
		container.NewBorder(nil, nil, nil, browseBtn, sourceEntry),
		widget.NewLabel("Title:"),
		titleEntry,
	)
}

// browseFile opens a file browser for file selection.
func (d *Designer) browseFile(entry *widget.Entry, extensions []string) {
	if d.window == nil {
		return
	}

	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()
		entry.SetText(reader.URI().Path())
	}, d.window)
}

// browseFolder opens a folder browser.
func (d *Designer) browseFolder(entry *widget.Entry) {
	if d.window == nil {
		return
	}

	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil || uri == nil {
			return
		}
		entry.SetText(uri.Path())
	}, d.window)
}

// showColorPicker shows a simple color picker dialog.
func (d *Designer) showColorPicker(entry *widget.Entry) {
	if d.window == nil {
		return
	}

	// Predefined color options
	colors := map[string]string{
		"White":       "#ffffff",
		"Red":         "#ff0000",
		"Green":       "#00ff00",
		"Blue":        "#0000ff",
		"Yellow":      "#ffff00",
		"Orange":      "#ff9900",
		"Purple":      "#9933ff",
		"Pink":        "#ff99cc",
		"Transparent": "transparent",
	}

	colorList := widget.NewList(
		func() int {
			return len(colors)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			i := 0
			for name := range colors {
				if i == id {
					label.SetText(name)
					break
				}
				i++
			}
		},
	)

	colorList.OnSelected = func(id widget.ListItemID) {
		i := 0
		for _, hex := range colors {
			if i == id {
				entry.SetText(hex)
				break
			}
			i++
		}
	}

	customEntry := widget.NewEntry()
	customEntry.SetPlaceHolder("Custom hex color (e.g., #ff00ff)")

	applyBtn := widget.NewButton("Apply Custom", func() {
		if customEntry.Text != "" {
			entry.SetText(customEntry.Text)
		}
	})

	content := container.NewBorder(
		widget.NewLabel("Select a color:"),
		container.NewVBox(
			widget.NewSeparator(),
			widget.NewLabel("Or enter custom color:"),
			customEntry,
			applyBtn,
		),
		nil, nil,
		colorList,
	)

	dialog.ShowCustom("Color Picker", "Close", content, d.window)
}

// Helper functions to get parameters with defaults.

func getParam(params map[string]interface{}, key string, defaultVal interface{}) interface{} {
	if val, ok := params[key]; ok {
		return val
	}
	return defaultVal
}

func getStringParam(params map[string]interface{}, key string, defaultVal string) string {
	if val, ok := params[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultVal
}

func getBoolParam(params map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := params[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultVal
}

func getFloatParam(params map[string]interface{}, key string, defaultVal float64) float64 {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		}
	}
	return defaultVal
}
