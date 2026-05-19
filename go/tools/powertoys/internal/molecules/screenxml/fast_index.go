package screenxml

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// FastIndexHandler handles fast index assignment by dragging.
type FastIndexHandler struct {
	grid        *GridWidget
	statusLabel *widget.Label
	active      bool
}

// NewFastIndexHandler creates a new fast index handler.
func NewFastIndexHandler(grid *GridWidget) *FastIndexHandler {
	return &FastIndexHandler{
		grid:   grid,
		active: false,
	}
}

// CreateUI creates the UI for fast index assignment.
func (fih *FastIndexHandler) CreateUI() fyne.CanvasObject {
	instructions := widget.NewRichTextFromMarkdown(`
**Fast Index Assignment:**
Drag from one cell to another to mark corners.
All cells in the rectangle will be assigned to the next available index.
`)

	fih.statusLabel = widget.NewLabel("Fast Index: Disabled - Click button to enable")

	toggleBtn := widget.NewButton("Enable Fast Index", func() {
		fih.toggle()
	})

	form := container.NewVBox(
		instructions,
		widget.NewSeparator(),
		toggleBtn,
		fih.statusLabel,
	)

	return form
}

// toggle enables/disables fast index mode.
func (fih *FastIndexHandler) toggle() {
	fih.active = !fih.active
	if fih.active {
		fih.statusLabel.SetText("Fast Index: Disabled - Drag selection has been removed")
		// Drag functionality removed - this feature is deprecated
		// fih.enableFastIndex()
	} else {
		fih.statusLabel.SetText("Fast Index: Disabled")
		// fih.grid.SetOnCellDrag(nil)
	}
}

// enableFastIndex enables drag-to-assign-index mode.
// DEPRECATED: Drag selection has been removed. This function is no longer functional.
func (fih *FastIndexHandler) enableFastIndex() {
	// Drag functionality removed
	/*
	fih.grid.SetOnCellDrag(func(startRow, startCol, endRow, endCol int) {
		// Ensure start is top-left and end is bottom-right
		if startRow > endRow {
			startRow, endRow = endRow, startRow
		}
		if startCol > endCol {
			startCol, endCol = endCol, startCol
		}

		// Find next available index
		nextIndex := fih.findNextIndex()

		// Assign index to all cells in the rectangle
		assignedCount := 0
		gridRows := fih.grid.GetRows()
		gridCols := fih.grid.GetCols()
		for row := startRow; row <= endRow && row < gridRows; row++ {
			for col := startCol; col <= endCol && col < gridCols; col++ {
				fih.grid.SetCellIndex(row, col, nextIndex)
				assignedCount++
			}
		}

		fih.statusLabel.SetText(fmt.Sprintf("Assigned index '%s' to %d cells", nextIndex, assignedCount))
		fih.grid.Refresh()
	})
	*/
}

// findNextIndex finds the next available index string.
func (fih *FastIndexHandler) findNextIndex() string {
	// Find the highest numeric index used
	maxIndex := -1
	rows := fih.grid.GetRows()
	cols := fih.grid.GetCols()
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := fih.grid.GetCell(row, col)
			if cell != nil && cell.Index != "" {
				var idx int
				if _, err := fmt.Sscanf(cell.Index, "%d", &idx); err == nil {
					if idx > maxIndex {
						maxIndex = idx
					}
				}
			}
		}
	}

	// Return next index (starting from 0 if none assigned)
	return fmt.Sprintf("%d", maxIndex+1)
}

