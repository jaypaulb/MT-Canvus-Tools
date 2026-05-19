package screenxml

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// CellEditor handles editing individual cell properties.
type CellEditor struct {
	grid            *GridWidget
	resolutionHandler *ResolutionHandler
}

// NewCellEditor creates a new cell editor.
func NewCellEditor(grid *GridWidget, resolutionHandler *ResolutionHandler) *CellEditor {
	return &CellEditor{
		grid:              grid,
		resolutionHandler: resolutionHandler,
	}
}

// ShowCellEditor shows a dialog to edit cell properties.
func (ce *CellEditor) ShowCellEditor(window fyne.Window, row, col int) {
	cell := ce.grid.GetCell(row, col)
	if cell == nil {
		return
	}

	// Parse GPU output (format: 1:2, 1:3, etc. - convert to 0-based for input)
	gpuStr := ""
	outputStr := ""
	if cell.GPUOutput != "" {
		parts := strings.Split(cell.GPUOutput, ":")
		if len(parts) == 2 {
			var gpu, output int
			fmt.Sscanf(parts[0], "%d", &gpu)
			fmt.Sscanf(parts[1], "%d", &output)
			// Convert from 1-based to 0-based for input fields
			gpuStr = fmt.Sprintf("%d", gpu-1)
			outputStr = fmt.Sprintf("%d", output-1)
		}
	}

	// Create form fields
	gpuEntry := widget.NewEntry()
	gpuEntry.SetText(gpuStr)
	gpuEntry.SetPlaceHolder("0")

	outputEntry := widget.NewEntry()
	outputEntry.SetText(outputStr)
	outputEntry.SetPlaceHolder("0")

	// Resolution dropdown
	resOptions := make([]string, len(CommonResolutions))
	for i, res := range CommonResolutions {
		resOptions[i] = res.Name
	}
	resSelect := widget.NewSelect(resOptions, nil)
	// Find current resolution in list
	currentResName := cell.Resolution.Name
	if currentResName == "" {
		currentResName = "1920x1080 (Full HD)"
	}
	resSelect.SetSelected(currentResName)

	// Index entry
	indexEntry := widget.NewEntry()
	indexEntry.SetText(cell.Index)
	indexEntry.SetPlaceHolder("Empty")

	// Create form
	form := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Edit Cell (%d, %d)", row, col)),
		widget.NewSeparator(),
		widget.NewLabel("GPU:"),
		gpuEntry,
		widget.NewLabel("Output:"),
		outputEntry,
		widget.NewLabel("Resolution:"),
		resSelect,
		widget.NewLabel("Index:"),
		indexEntry,
	)

	// Create a variable to hold the dialog reference
	var d dialog.Dialog

	// Save button
	saveBtn := widget.NewButton("Save", func() {
		// Validate and save
		gpuNum, err1 := strconv.Atoi(strings.TrimSpace(gpuEntry.Text))
		outputNum, err2 := strconv.Atoi(strings.TrimSpace(outputEntry.Text))

		if gpuEntry.Text != "" && (err1 != nil || gpuNum < 0) {
			dialog.ShowError(fmt.Errorf("Invalid GPU number"), window)
			return
		}
		if outputEntry.Text != "" && (err2 != nil || outputNum < 0) {
			dialog.ShowError(fmt.Errorf("Invalid output number"), window)
			return
		}

		// Set GPU output (convert to 1-based format with colon)
		if gpuEntry.Text != "" && outputEntry.Text != "" {
			gpuOutput := fmt.Sprintf("%d:%d", gpuNum+1, outputNum+1)
			ce.grid.SetCellGPUOutput(row, col, gpuOutput)
		} else {
			ce.grid.ClearCellGPUOutput(row, col)
		}

		// Set resolution
		selectedRes := resSelect.Selected
		for _, res := range CommonResolutions {
			if res.Name == selectedRes {
				ce.grid.SetCellResolution(row, col, res)
				break
			}
		}

		// Set index
		index := strings.TrimSpace(indexEntry.Text)
		ce.grid.SetCellIndex(row, col, index)

		d.Hide()
	})

	cancelBtn := widget.NewButton("Cancel", func() {
		d.Hide()
	})

	buttons := container.NewHBox(saveBtn, cancelBtn)
	content := container.NewVBox(form, buttons)

	d = dialog.NewCustom("Edit Cell", "Close", content, window)
	d.Resize(fyne.NewSize(400, 300))
	d.Show()
}

