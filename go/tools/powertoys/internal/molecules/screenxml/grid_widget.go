package screenxml

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

const (
	InitialGridCols = 4
	InitialGridRows = 3
)

// Orientation represents the display orientation.
type Orientation string

const (
	OrientationLandscape Orientation = "landscape"
	OrientationPortrait  Orientation = "portrait"
)

// CellState represents the state of a grid cell.
type CellState struct {
	GPUOutput    string      // Format: "gpu#:output#" (e.g., "1:2")
	TouchArea    int         // Touch area index (-1 if not assigned)
	IsLayoutArea bool        // True if part of layout area (pink frame)
	Resolution   Resolution  // Resolution for this cell (default: 1920x1080)
	Index        string      // Index string (empty by default)
	Orientation  Orientation // Display orientation (landscape or portrait)
}

// GridWidget is a custom widget for displaying a dynamic grid.
type GridWidget struct {
	widget.BaseWidget
	cells         [][]*CellState // Dynamic 2D slice
	onCellClick   func(row, col int)
	onSizeChanged func() // Callback when grid size changes
}

// NewGridWidget creates a new grid widget with initial 4x3 size.
func NewGridWidget() *GridWidget {
	grid := &GridWidget{
		cells: make([][]*CellState, InitialGridRows),
	}

	// Initialize all cells
	for row := 0; row < InitialGridRows; row++ {
		grid.cells[row] = make([]*CellState, InitialGridCols)
		for col := 0; col < InitialGridCols; col++ {
			grid.cells[row][col] = &CellState{
				TouchArea:    -1,
				IsLayoutArea: false,
				Resolution:   Resolution{Width: 1920, Height: 1080, Name: "1920x1080 (Full HD)"},
				Index:        "",
				Orientation:  OrientationLandscape,
			}
		}
	}

	grid.ExtendBaseWidget(grid)
	return grid
}

// GetCols returns the current number of columns.
func (g *GridWidget) GetCols() int {
	if len(g.cells) == 0 {
		return 0
	}
	return len(g.cells[0])
}

// GetRows returns the current number of rows.
func (g *GridWidget) GetRows() int {
	return len(g.cells)
}

// SetOnSizeChanged sets a callback that's called when the grid size changes.
func (g *GridWidget) SetOnSizeChanged(fn func()) {
	g.onSizeChanged = fn
}

// AddColumn adds a new column to the grid.
func (g *GridWidget) AddColumn() {
	rows := g.GetRows()
	for row := 0; row < rows; row++ {
		g.cells[row] = append(g.cells[row], &CellState{
			TouchArea:    -1,
			IsLayoutArea: false,
			Resolution:   Resolution{Width: 1920, Height: 1080, Name: "1920x1080 (Full HD)"},
			Index:        "",
			Orientation:  OrientationLandscape,
		})
	}
	g.Refresh()
	if g.onSizeChanged != nil {
		g.onSizeChanged()
	}
}

// AddRow adds a new row to the grid.
func (g *GridWidget) AddRow() {
	cols := g.GetCols()
	newRow := make([]*CellState, cols)
	for col := 0; col < cols; col++ {
		newRow[col] = &CellState{
			TouchArea:    -1,
			IsLayoutArea: false,
			Resolution:   Resolution{Width: 1920, Height: 1080, Name: "1920x1080 (Full HD)"},
			Index:        "",
			Orientation:  OrientationLandscape,
		}
	}
	g.cells = append(g.cells, newRow)
	g.Refresh()
	if g.onSizeChanged != nil {
		g.onSizeChanged()
	}
}

// SetOnCellClick sets the callback for cell click events.
func (g *GridWidget) SetOnCellClick(fn func(row, col int)) {
	g.onCellClick = fn
}

// GetCell returns the state of a cell.
func (g *GridWidget) GetCell(row, col int) *CellState {
	rows := g.GetRows()
	cols := g.GetCols()
	if row < 0 || row >= rows || col < 0 || col >= cols {
		return nil
	}
	return g.cells[row][col]
}

// SetCellGPUOutput sets the GPU output for a cell.
func (g *GridWidget) SetCellGPUOutput(row, col int, gpuOutput string) {
	rows := g.GetRows()
	cols := g.GetCols()
	if row < 0 || row >= rows || col < 0 || col >= cols {
		return
	}
	g.cells[row][col].GPUOutput = gpuOutput
	g.Refresh()
}

// ClearCellGPUOutput clears the GPU output for a cell.
func (g *GridWidget) ClearCellGPUOutput(row, col int) {
	rows := g.GetRows()
	cols := g.GetCols()
	if row < 0 || row >= rows || col < 0 || col >= cols {
		return
	}
	g.cells[row][col].GPUOutput = ""
	g.Refresh()
}

// SetCellResolution sets the resolution for a cell.
func (g *GridWidget) SetCellResolution(row, col int, res Resolution) {
	rows := g.GetRows()
	cols := g.GetCols()
	if row < 0 || row >= rows || col < 0 || col >= cols {
		return
	}
	g.cells[row][col].Resolution = res
	g.Refresh()
}

// SetCellIndex sets the index for a cell.
func (g *GridWidget) SetCellIndex(row, col int, index string) {
	rows := g.GetRows()
	cols := g.GetCols()
	if row < 0 || row >= rows || col < 0 || col >= cols {
		return
	}
	g.cells[row][col].Index = index
	g.Refresh()
}

// SetCellOrientation sets the orientation for a cell.
func (g *GridWidget) SetCellOrientation(row, col int, orientation Orientation) {
	rows := g.GetRows()
	cols := g.GetCols()
	if row < 0 || row >= rows || col < 0 || col >= cols {
		return
	}
	g.cells[row][col].Orientation = orientation
	g.Refresh()
}

// HasCellsWithData checks if there are any cells with GPU output assigned.
func (g *GridWidget) HasCellsWithData() bool {
	rows := g.GetRows()
	cols := g.GetCols()
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			if g.cells[row][col].GPUOutput != "" {
				return true
			}
		}
	}
	return false
}

// Tapped handles tap/click events on the grid.
func (g *GridWidget) Tapped(e *fyne.PointEvent) {
	row, col := g.getCellFromPosition(e.Position)
	if row >= 0 && col >= 0 && g.onCellClick != nil {
		g.onCellClick(row, col)
	}
}

// getCellFromPosition converts a screen position to grid cell coordinates.
func (g *GridWidget) getCellFromPosition(pos fyne.Position) (row, col int) {
	size := g.Size()
	cols := float32(g.GetCols())
	rows := float32(g.GetRows())
	if cols == 0 || rows == 0 {
		return -1, -1
	}
	cellWidth := size.Width / cols
	cellHeight := size.Height / rows

	col = int(pos.X / cellWidth)
	row = int(pos.Y / cellHeight)

	actualRows := g.GetRows()
	actualCols := g.GetCols()
	if row < 0 || row >= actualRows || col < 0 || col >= actualCols {
		return -1, -1
	}
	return row, col
}

// CreateRenderer creates the renderer for the grid widget.
func (g *GridWidget) CreateRenderer() fyne.WidgetRenderer {
	rows := g.GetRows()
	return &gridRenderer{
		grid:   g,
		cells:  make([][]fyne.CanvasObject, rows),
		border: canvas.NewRectangle(color.RGBA{R: 200, G: 200, B: 200, A: 255}),
	}
}

// gridRenderer renders the grid widget.
type gridRenderer struct {
	grid   *GridWidget
	cells  [][]fyne.CanvasObject
	border *canvas.Rectangle
}

func (r *gridRenderer) Layout(size fyne.Size) {
	// Add spacing between cells (2px gap)
	spacing := float32(2)
	cols := float32(r.grid.GetCols())
	rows := float32(r.grid.GetRows())
	if cols == 0 || rows == 0 {
		return
	}
	cellWidth := (size.Width - float32(cols-1)*spacing) / cols
	cellHeight := (size.Height - float32(rows-1)*spacing) / rows

	actualRows := r.grid.GetRows()
	actualCols := r.grid.GetCols()

	// Ensure r.cells has enough rows
	if len(r.cells) < actualRows {
		oldLen := len(r.cells)
		r.cells = append(r.cells, make([][]fyne.CanvasObject, actualRows-oldLen)...)
	}

	for row := 0; row < actualRows; row++ {
		if r.cells[row] == nil {
			r.cells[row] = make([]fyne.CanvasObject, actualCols)
		}
		// Ensure row slice has enough capacity
		if len(r.cells[row]) < actualCols {
			oldLen := len(r.cells[row])
			r.cells[row] = append(r.cells[row], make([]fyne.CanvasObject, actualCols-oldLen)...)
		}
		for col := 0; col < actualCols; col++ {
			if r.cells[row][col] == nil {
				rect := canvas.NewRectangle(color.RGBA{R: 255, G: 255, B: 255, A: 255})
				rect.StrokeColor = color.RGBA{R: 150, G: 150, B: 150, A: 255}
				rect.StrokeWidth = 1
				r.cells[row][col] = rect
			}
			// Calculate position with spacing
			x := float32(col) * (cellWidth + spacing)
			y := float32(row) * (cellHeight + spacing)
			r.cells[row][col].Resize(fyne.NewSize(cellWidth, cellHeight))
			r.cells[row][col].Move(fyne.NewPos(x, y))
		}
	}

	r.border.Resize(size)
	r.border.Move(fyne.NewPos(0, 0))
}

func (r *gridRenderer) MinSize() fyne.Size {
	return fyne.NewSize(600, 300) // Minimum size for 10x5 grid
}

func (r *gridRenderer) Objects() []fyne.CanvasObject {
	objects := []fyne.CanvasObject{r.border}
	rows := r.grid.GetRows()
	cols := r.grid.GetCols()

	// Ensure r.cells has enough rows
	if len(r.cells) < rows {
		oldLen := len(r.cells)
		r.cells = append(r.cells, make([][]fyne.CanvasObject, rows-oldLen)...)
	}

	for row := 0; row < rows; row++ {
		if row >= len(r.cells) || r.cells[row] == nil {
			continue
		}
		for col := 0; col < cols; col++ {
			if col < len(r.cells[row]) && r.cells[row][col] != nil {
				objects = append(objects, r.cells[row][col])
			}
		}
	}
	return objects
}

func (r *gridRenderer) Refresh() {
	cellState := r.grid.cells
	rows := r.grid.GetRows()
	cols := r.grid.GetCols()

	// Ensure r.cells has enough rows
	if len(r.cells) < rows {
		oldLen := len(r.cells)
		r.cells = append(r.cells, make([][]fyne.CanvasObject, rows-oldLen)...)
	}

	for row := 0; row < rows; row++ {
		// Ensure row slice is initialized
		if r.cells[row] == nil {
			r.cells[row] = make([]fyne.CanvasObject, cols)
		}
		// Ensure row slice has enough capacity
		if len(r.cells[row]) < cols {
			oldLen := len(r.cells[row])
			r.cells[row] = append(r.cells[row], make([]fyne.CanvasObject, cols-oldLen)...)
		}
		for col := 0; col < cols; col++ {
			// Ensure cell is initialized
			if r.cells[row][col] == nil {
				rect := canvas.NewRectangle(color.RGBA{R: 255, G: 255, B: 255, A: 255})
				rect.StrokeColor = color.RGBA{R: 150, G: 150, B: 150, A: 255}
				rect.StrokeWidth = 1
				r.cells[row][col] = rect
			}
			if r.cells[row][col] != nil {
				rect := r.cells[row][col].(*canvas.Rectangle)
				// Ensure cellState has valid data for this cell
				var state *CellState
				if row < len(cellState) && col < len(cellState[row]) {
					state = cellState[row][col]
				} else {
					// Create default state if out of bounds
					state = &CellState{
						TouchArea:    -1,
						IsLayoutArea: false,
						Resolution:   Resolution{Width: 1920, Height: 1080, Name: "1920x1080 (Full HD)"},
						Index:        "",
						Orientation:  OrientationLandscape,
					}
				}

				// Set cell color based on state
				if state.GPUOutput != "" {
					// Cell has GPU output assigned - light blue
					rect.FillColor = color.RGBA{R: 173, G: 216, B: 230, A: 255}
				} else if state.IsLayoutArea {
					// Layout area - light pink border
					rect.FillColor = color.RGBA{R: 255, G: 240, B: 245, A: 255}
					rect.StrokeColor = color.RGBA{R: 255, G: 192, B: 203, A: 255}
					rect.StrokeWidth = 2
				} else {
					// Default - white
					rect.FillColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
					rect.StrokeColor = color.RGBA{R: 150, G: 150, B: 150, A: 255}
					rect.StrokeWidth = 1
				}
			}
		}
	}
	canvas.Refresh(r.grid)
}

func (r *gridRenderer) Destroy() {
}
