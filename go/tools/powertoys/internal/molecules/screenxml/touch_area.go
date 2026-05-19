package screenxml

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// TouchAreaHandler manages touch area assignment.
type TouchAreaHandler struct {
	grid         *GridWidget
	areaIndexInput *widget.Entry
	statusLabel   *widget.Label
	currentAreaIndex int
}

// NewTouchAreaHandler creates a new touch area handler.
func NewTouchAreaHandler(grid *GridWidget) *TouchAreaHandler {
	return &TouchAreaHandler{
		grid:            grid,
		currentAreaIndex: 1,
	}
}

// CreateUI creates the UI for touch area assignment.
func (tah *TouchAreaHandler) CreateUI() fyne.CanvasObject {
	instructions := widget.NewRichTextFromMarkdown(`
**Touch Area Assignment:**
1. Enter touch area index (starting from 1)
2. Click and drag over grid cells to assign them to the touch area
3. All cells under the drag rectangle will be assigned to the same index
4. Cells will show pink frame when part of a layout area
`)

	areaLabel := widget.NewLabel("Touch Area Index:")
	tah.areaIndexInput = widget.NewEntry()
	tah.areaIndexInput.SetText("1")
	tah.areaIndexInput.OnChanged = func(text string) {
		// Validate and update current area index
		if idx := tah.parseAreaIndex(text); idx > 0 {
			tah.currentAreaIndex = idx
		}
	}

	clearBtn := widget.NewButton("Clear All Touch Areas", func() {
		tah.clearAllTouchAreas()
	})

	tah.statusLabel = widget.NewLabel("Touch Area Assignment - Drag selection has been removed")

	// DEPRECATED: Drag selection has been removed. Touch area assignment via drag is no longer functional.
	// Set up drag handler
	/*
	tah.grid.SetOnCellDrag(func(startRow, startCol, endRow, endCol int) {
		tah.assignTouchArea(startRow, startCol, endRow, endCol)
	})
	*/

	form := container.NewVBox(
		instructions,
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			areaLabel, tah.areaIndexInput,
		),
		clearBtn,
		tah.statusLabel,
	)

	return form
}

// assignTouchArea assigns cells in the drag rectangle to the current touch area index.
func (tah *TouchAreaHandler) assignTouchArea(startRow, startCol, endRow, endCol int) {
	// Ensure start is top-left and end is bottom-right
	if startRow > endRow {
		startRow, endRow = endRow, startRow
	}
	if startCol > endCol {
		startCol, endCol = endCol, startCol
	}

	assignedCount := 0
	gridRows := tah.grid.GetRows()
	gridCols := tah.grid.GetCols()
	for row := startRow; row <= endRow && row < gridRows; row++ {
		for col := startCol; col <= endCol && col < gridCols; col++ {
			cell := tah.grid.GetCell(row, col)
			if cell != nil {
				cell.TouchArea = tah.currentAreaIndex
				cell.IsLayoutArea = true
				assignedCount++
			}
		}
	}

	tah.grid.Refresh()
	tah.statusLabel.SetText(fmt.Sprintf("Assigned %d cells to touch area %d", assignedCount, tah.currentAreaIndex))
}

// clearAllTouchAreas clears all touch area assignments.
func (tah *TouchAreaHandler) clearAllTouchAreas() {
	rows := tah.grid.GetRows()
	cols := tah.grid.GetCols()
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := tah.grid.GetCell(row, col)
			if cell != nil {
				cell.TouchArea = -1
				cell.IsLayoutArea = false
			}
		}
	}
	tah.grid.Refresh()
	tah.statusLabel.SetText("Cleared all touch area assignments")
}

// parseAreaIndex parses and validates the area index input.
func (tah *TouchAreaHandler) parseAreaIndex(text string) int {
	if text == "" {
		return 0
	}
	var idx int
	if _, err := fmt.Sscanf(text, "%d", &idx); err != nil {
		return 0
	}
	if idx < 1 {
		return 0
	}
	return idx
}

// GetTouchAreas returns a map of touch area indices to their cell coordinates.
func (tah *TouchAreaHandler) GetTouchAreas() map[int][]string {
	areas := make(map[int][]string)
	rows := tah.grid.GetRows()
	cols := tah.grid.GetCols()
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := tah.grid.GetCell(row, col)
			if cell != nil && cell.TouchArea > 0 {
				key := fmt.Sprintf("%d,%d", row, col)
				areas[cell.TouchArea] = append(areas[cell.TouchArea], key)
			}
		}
	}
	return areas
}

