package screenxml

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// GPUAssignment handles GPU output assignment workflow.
type GPUAssignment struct {
	grid               *GridWidget
	gpuInput           *widget.Entry
	outputInput        *widget.Entry
	statusLabel        *widget.Label
	quickModeBtn       *widget.Button
	quickModeActive    bool
	nextGPUOutput      string // Next GPU output for quick mode (1:1, 1:2, 1:3, etc.)
	onQuickModeChanged func()  // Callback when quick mode is toggled
}

// IsQuickModeActive returns whether quick mode is currently active.
func (ga *GPUAssignment) IsQuickModeActive() bool {
	return ga.quickModeActive
}

// SetOnQuickModeChanged sets a callback for when quick mode is toggled.
func (ga *GPUAssignment) SetOnQuickModeChanged(fn func()) {
	ga.onQuickModeChanged = fn
}

// NewGPUAssignment creates a new GPU assignment handler.
func NewGPUAssignment(grid *GridWidget) *GPUAssignment {
	return &GPUAssignment{
		grid:            grid,
		quickModeActive: false,
		nextGPUOutput:   "1:1",
	}
}

// CreateUI creates the UI for GPU assignment.
func (ga *GPUAssignment) CreateUI() fyne.CanvasObject {
	// Instructions label
	instructions := widget.NewRichTextFromMarkdown(`
**GPU Output Assignment Instructions:**
1. Draw a large number on each physical screen
2. Enter the GPU number (e.g., 0, 1, 2)
3. Enter the output number (e.g., 1, 2, 3)
4. Click the corresponding grid cell to assign
5. Click again to remove assignment
`)

	// GPU and output input fields
	gpuLabel := widget.NewLabel("GPU Number:")
	ga.gpuInput = widget.NewEntry()
	ga.gpuInput.SetPlaceHolder("0")

	outputLabel := widget.NewLabel("Output Number:")
	ga.outputInput = widget.NewEntry()
	ga.outputInput.SetPlaceHolder("1")

	// Assign button
	assignBtn := widget.NewButton("Assign to Clicked Cell", func() {
		ga.enableAssignmentMode()
	})

	// Quick Mode button
	ga.quickModeBtn = widget.NewButton("Quick Mode: OFF", func() {
		ga.toggleQuickMode()
	})

	// Status label
	ga.statusLabel = widget.NewLabel("Ready - Click a cell to assign GPU output")

	// Layout
	form := container.NewVBox(
		instructions,
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			gpuLabel, ga.gpuInput,
			outputLabel, ga.outputInput,
		),
		assignBtn,
		widget.NewSeparator(),
		ga.quickModeBtn,
		widget.NewSeparator(),
		ga.statusLabel,
	)

	return form
}

// enableAssignmentMode enables click-to-assign mode.
func (ga *GPUAssignment) enableAssignmentMode() {
	gpuStr := strings.TrimSpace(ga.gpuInput.Text)
	outputStr := strings.TrimSpace(ga.outputInput.Text)

	if gpuStr == "" || outputStr == "" {
		ga.statusLabel.SetText("Error: Please enter both GPU and output numbers")
		return
	}

	gpuNum, err := strconv.Atoi(gpuStr)
	if err != nil || gpuNum < 0 {
		ga.statusLabel.SetText("Error: Invalid GPU number")
		return
	}

	outputNum, err := strconv.Atoi(outputStr)
	if err != nil || outputNum < 0 {
		ga.statusLabel.SetText("Error: Invalid output number")
		return
	}

	// Convert to 1-based format with colon: 1:1, 1:2, etc.
	gpuOutput := fmt.Sprintf("%d:%d", gpuNum+1, outputNum+1)
	ga.statusLabel.SetText(fmt.Sprintf("Assignment mode: %s - Click a cell to assign or remove", gpuOutput))

	// Set up click handler
	ga.grid.SetOnCellClick(func(row, col int) {
		cell := ga.grid.GetCell(row, col)
		if cell == nil {
			return
		}

		// Toggle assignment: if already assigned to this GPU output, remove it
		if cell.GPUOutput == gpuOutput {
			ga.grid.ClearCellGPUOutput(row, col)
			ga.statusLabel.SetText(fmt.Sprintf("Removed %s from cell (%d,%d)", gpuOutput, row, col))
		} else {
			ga.grid.SetCellGPUOutput(row, col, gpuOutput)
			ga.statusLabel.SetText(fmt.Sprintf("Assigned %s to cell (%d,%d)", gpuOutput, row, col))
		}
	})
}

// toggleQuickMode toggles quick assignment mode.
func (ga *GPUAssignment) toggleQuickMode() {
	ga.quickModeActive = !ga.quickModeActive
	if ga.quickModeActive {
		ga.quickModeBtn.SetText("Quick Mode: ON")
		ga.statusLabel.SetText("Quick Mode: Click cells to auto-assign GPU outputs (1:1, 1:2, 1:3, etc.)")
		ga.enableQuickMode()
	} else {
		ga.quickModeBtn.SetText("Quick Mode: OFF")
		ga.statusLabel.SetText("Quick Mode disabled")
		// Notify that quick mode was disabled (so default handler can be restored)
		if ga.onQuickModeChanged != nil {
			ga.onQuickModeChanged()
		}
	}
}

// enableQuickMode enables quick assignment mode where clicking cells auto-assigns GPU outputs.
func (ga *GPUAssignment) enableQuickMode() {
	// Find the next available GPU output
	ga.nextGPUOutput = ga.findNextGPUOutput()

	ga.grid.SetOnCellClick(func(row, col int) {
		cell := ga.grid.GetCell(row, col)
		if cell == nil {
			return
		}

		// If already assigned to this GPU output, remove it
		if cell.GPUOutput == ga.nextGPUOutput {
			ga.grid.ClearCellGPUOutput(row, col)
			ga.statusLabel.SetText(fmt.Sprintf("Removed %s from cell (%d,%d)", ga.nextGPUOutput, row, col))
			// Don't advance - allow re-assigning the same output
		} else {
			// Assign the next GPU output
			ga.grid.SetCellGPUOutput(row, col, ga.nextGPUOutput)
			ga.statusLabel.SetText(fmt.Sprintf("Assigned %s to cell (%d,%d) - Next: %s", ga.nextGPUOutput, row, col, ga.getNextGPUOutput(ga.nextGPUOutput)))
			// Advance to next GPU output
			ga.nextGPUOutput = ga.getNextGPUOutput(ga.nextGPUOutput)
		}
	})
}

// findNextGPUOutput finds the next available GPU output starting from 1:1.
func (ga *GPUAssignment) findNextGPUOutput() string {
	// Check all cells to find the highest assigned GPU output
	maxGpu := 0
	maxOutput := 0

	rows := ga.grid.GetRows()
	cols := ga.grid.GetCols()
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := ga.grid.GetCell(row, col)
			if cell != nil && cell.GPUOutput != "" {
				parts := strings.Split(cell.GPUOutput, ":")
				if len(parts) == 2 {
					var gpu, output int
					fmt.Sscanf(parts[0], "%d", &gpu)
					fmt.Sscanf(parts[1], "%d", &output)
					if gpu > maxGpu || (gpu == maxGpu && output > maxOutput) {
						maxGpu = gpu
						maxOutput = output
					}
				}
			}
		}
	}

	// Start from 1:1 if nothing assigned, otherwise continue from next
	if maxGpu == 0 {
		return "1:1"
	}

	// Increment output (1:1 -> 1:2 -> 1:3 -> 1:4 -> 2:1 -> 2:2, etc.)
	if maxOutput < 4 {
		return fmt.Sprintf("%d:%d", maxGpu, maxOutput+1)
	}
	return fmt.Sprintf("%d:1", maxGpu+1)
}

// getNextGPUOutput gets the next GPU output in sequence.
func (ga *GPUAssignment) getNextGPUOutput(current string) string {
	parts := strings.Split(current, ":")
	if len(parts) != 2 {
		return "1:1"
	}

	var gpu, output int
	fmt.Sscanf(parts[0], "%d", &gpu)
	fmt.Sscanf(parts[1], "%d", &output)

	// Increment: 1:1 -> 1:2 -> 1:3 -> 1:4 -> 2:1 -> 2:2, etc.
	if output < 4 {
		return fmt.Sprintf("%d:%d", gpu, output+1)
	}
	return fmt.Sprintf("%d:1", gpu+1)
}

// GetGPUOutputs returns all assigned GPU outputs as a map of cell coordinates to GPU output strings.
func (ga *GPUAssignment) GetGPUOutputs() map[string]string {
	outputs := make(map[string]string)
	rows := ga.grid.GetRows()
	cols := ga.grid.GetCols()
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := ga.grid.GetCell(row, col)
			if cell != nil && cell.GPUOutput != "" {
				key := fmt.Sprintf("%d,%d", row, col)
				outputs[key] = cell.GPUOutput
			}
		}
	}
	return outputs
}

