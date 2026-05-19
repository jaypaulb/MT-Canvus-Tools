package screenxml

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// GridContainer is a container that displays a grid of cell widgets.
type GridContainer struct {
	grid        *GridWidget
	cellWidgets [][]*CellWidget // Dynamic 2D slice
	container   *fyne.Container
	lastRows    int // Track last row count to detect size changes
	lastCols    int // Track last column count to detect size changes
}

// NewGridContainer creates a new grid container with cell widgets.
func NewGridContainer(grid *GridWidget) *GridContainer {
	gc := &GridContainer{
		grid:        grid,
		cellWidgets: make([][]*CellWidget, 0),
	}

	// Set up callback for when grid size changes
	grid.SetOnSizeChanged(func() {
		gc.rebuildContainer()
	})

	gc.rebuildContainer()
	return gc
}

// rebuildContainer rebuilds the container with current grid size.
func (gc *GridContainer) rebuildContainer() {
	rows := gc.grid.GetRows()
	cols := gc.grid.GetCols()

	// Check if grid size actually changed (or if this is first build)
	sizeChanged := gc.container == nil || rows != gc.lastRows || cols != gc.lastCols

	// Remove all widgets from old container if it exists and size changed
	if gc.container != nil && sizeChanged {
		gc.container.RemoveAll()
		// Set to nil to ensure we don't try to reuse it
		gc.container = nil
	}

	// Ensure cellWidgets has correct size
	if len(gc.cellWidgets) < rows {
		oldLen := len(gc.cellWidgets)
		gc.cellWidgets = append(gc.cellWidgets, make([][]*CellWidget, rows-oldLen)...)
	}

	// Create new container with correct column count
	cellContainer := container.NewGridWithColumns(cols)

	// Create or update cell widgets
	for row := 0; row < rows; row++ {
		// Ensure row slice exists
		if gc.cellWidgets[row] == nil {
			gc.cellWidgets[row] = make([]*CellWidget, cols)
		}
		// Ensure row slice has correct length
		if len(gc.cellWidgets[row]) < cols {
			oldLen := len(gc.cellWidgets[row])
			gc.cellWidgets[row] = append(gc.cellWidgets[row], make([]*CellWidget, cols-oldLen)...)
		}
		for col := 0; col < cols; col++ {
			// If size changed, always create new widgets to avoid container conflicts
			// Otherwise, reuse existing widgets if they exist
			if sizeChanged || gc.cellWidgets[row][col] == nil {
				// Create new cell widget
				cellWidget := NewCellWidget(gc.grid, row, col)
				gc.cellWidgets[row][col] = cellWidget
			}
			// Add widget to container
			cellContainer.Add(gc.cellWidgets[row][col])
		}
	}

	// Update tracked size
	gc.lastRows = rows
	gc.lastCols = cols

	gc.container = cellContainer
}

// GetContainer returns the container for this grid.
func (gc *GridContainer) GetContainer() *fyne.Container {
	return gc.container
}

