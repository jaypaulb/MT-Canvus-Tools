package screenxml

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
)

// ScreenXML represents the screen.xml structure.
type ScreenXML struct {
	XMLName            xml.Name       `xml:"multihead"`
	Comment            string         `xml:",comment"`
	Type               string         `xml:"type,attr,omitempty"`
	DPI                *XMLAttr       `xml:"dpi,omitempty"`
	DPIComment         string         `xml:",comment"`
	DPMS               *XMLAttr       `xml:"dpms,omitempty"`
	DPMSComment        string         `xml:",comment"`
	HwColorCorrection  *XMLAttr       `xml:"hw-color-correction,omitempty"`
	Vsync              *XMLAttr       `xml:"vsync,omitempty"`
	VsyncComment       string         `xml:",comment"`
	LayerSize          *XMLAttr       `xml:"layer-size,omitempty"`
	LayerSizeComment   string         `xml:",comment"`
	GlFinish           *XMLAttr       `xml:"gl-finish,omitempty"`
	GlFinishComment    string         `xml:",comment"`
	AsyncTextureUpload *XMLAttr       `xml:"async-texture-upload,omitempty"`
	AsyncTextureUploadComment string  `xml:",comment"`
	Windows            []WindowConfig `xml:"window"`
}

// XMLAttr represents an XML element with type attribute (e.g., <dpi type="">value</dpi>).
type XMLAttr struct {
	Type  string `xml:"type,attr,omitempty"`
	Value string `xml:",chardata"`
}


// WindowConfig represents a window element in screen.xml.
type WindowConfig struct {
	Comment                string       `xml:",comment"`
	Type                   string       `xml:"type,attr"`
	DirectRendering        *XMLAttr     `xml:"direct-rendering,omitempty"`
	DirectRenderingComment string       `xml:",comment"`
	Frameless              *XMLAttr     `xml:"frameless,omitempty"`
	FramelessComment       string       `xml:",comment"`
	FsaaSamples            *XMLAttr     `xml:"fsaa-samples,omitempty"`
	FsaaSamplesComment     string       `xml:",comment"`
	Fullscreen             *XMLAttr     `xml:"fullscreen,omitempty"`
	FullscreenComment      string       `xml:",comment"`
	Location               *XMLAttr     `xml:"location,omitempty"`
	LocationComment        string       `xml:",comment"`
	Resizable              *XMLAttr     `xml:"resizable,omitempty"`
	ResizableComment       string       `xml:",comment"`
	ScreenNumber           *XMLAttr     `xml:"screennumber,omitempty"`
	ScreenNumberComment    string       `xml:",comment"`
	Size                   *XMLAttr     `xml:"size,omitempty"`
	SizeComment            string       `xml:",comment"`
	Areas                  []AreaConfig `xml:"area"`
}

// AreaConfig represents an area element within a window.
type AreaConfig struct {
	Comment                string   `xml:",comment"`
	Type                   string   `xml:"type,attr"`
	GraphicsLocation       *XMLAttr `xml:"graphicslocation,omitempty"`
	GraphicsLocationComment string  `xml:",comment"`
	GraphicsSize           *XMLAttr `xml:"graphicssize,omitempty"`
	GraphicsSizeComment    string   `xml:",comment"`
	Location               *XMLAttr `xml:"location,omitempty"`
	Seams                  *XMLAttr `xml:"seams,omitempty"`
	Size                   *XMLAttr `xml:"size,omitempty"`
}


// XMLGenerator generates screen.xml from grid configuration.
type XMLGenerator struct {
	grid           *GridWidget
	gpuAssignments *GPUAssignment
	touchAreas     *TouchAreaHandler
	resolution     Resolution
	defaultRes     Resolution
	areaPerGPU     bool // If true, create area per window (per GPU); if false, area per GPU output (per screen)
}

// NewXMLGenerator creates a new XML generator.
func NewXMLGenerator(grid *GridWidget, gpuAssignments *GPUAssignment, touchAreas *TouchAreaHandler, resolution Resolution) *XMLGenerator {
	return &XMLGenerator{
		grid:           grid,
		gpuAssignments: gpuAssignments,
		touchAreas:     touchAreas,
		resolution:     resolution,
		defaultRes:     CommonResolutions[0], // 1920x1080
		areaPerGPU:     false,                // Default: area per Screen (per GPU output)
	}
}

// SetAreaPerGPU sets whether to create area per GPU (window) or per Screen (GPU output).
// true = per GPU (one area per window matching all outputs on that GPU)
// false = per Screen (one area per GPU output, each output is one screen)
func (xg *XMLGenerator) SetAreaPerGPU(perGPU bool) {
	xg.areaPerGPU = perGPU
}

// Generate generates screen.xml content from the grid configuration.
func (xg *XMLGenerator) Generate() ([]byte, error) {
	// Calculate total layer size from all assigned cells
	layerSize := xg.calculateLayerSize()

	screenXML := ScreenXML{
		Comment: `The multihead element defines global display options used by MT Canvus.
These settings apply to all windows and areas in the configuration.`,
		Type:              "",
		DPI:               &XMLAttr{Type: "", Value: "40.053"},
		DPIComment:        `DPI: "Dots per inch" setting for converting physical dimensions to pixels in .css files.
For instance, width: 10cm; will match physical dimension of 10 cm if the dpi is correct.
Default: 40.053, which is the DPI for 55" FullHD displays.`,
		DPMS:              &XMLAttr{Type: "", Value: "0 0 0"},
		DPMSComment:       `DPMS: Display Power Management Signaling - controls monitor power states (standby, suspend, off).
Format: "standby suspend off" in seconds. Set to "0 0 0" to disable power management.`,
		HwColorCorrection: &XMLAttr{Type: "", Value: "0"},
		Vsync:             &XMLAttr{Type: "", Value: "0"}, // Default for Windows
		VsyncComment:      `vsync: Should vertical sync be enabled. Enable to remove tearing artifacts.
Default value depends on the platform (1 on Linux, 0 on Windows).`,
		LayerSize:         &XMLAttr{Type: "", Value: layerSize},
		LayerSizeComment: `layer-size: If empty ("0 0"), then the layer-size is calculated automatically from the graphics coordinates.
The layer size defines the size of the overall touch interaction layer.`,
		GlFinish:          &XMLAttr{Type: "", Value: "0"},
		GlFinishComment:  `gl-finish: Enabling this might reduce rendering latency and tearing between GPUs by eliminating
frame buffering but at a cost of reduced performance. On a low level this basically specifies
if glFinish should be called after every rendered frame. Test and see if the new framerate is acceptable.`,
		AsyncTextureUpload: &XMLAttr{Type: "", Value: "0"},
		AsyncTextureUploadComment: `async-texture-upload: If set to 1, some texture data will be uploaded asynchronously to the GPU
using a separate upload threads and OpenGL contexts. This can improve GPU upload throughput and
reduce "Render collect" time and therefore improve performance. This is mostly useful in applications
that upload a lot of content to the GPU, like an app with lots of videos. Enable this if you plan
to have more than 5 videos running per GPU. Otherwise disable.`,
		Windows:           []WindowConfig{},
	}

	// Group cells by GPU output to create windows
	gpuGroups := xg.groupCellsByGPU()

	// Counters for unique window and area names
	windowCounter := 1
	areaCounter := 1

	if xg.areaPerGPU {
		// Area per GPU (window) mode: Create one area per window matching all outputs on that GPU
		// Group by GPU number to create windows
		windowsByGPU := xg.groupByGPU(gpuGroups)

		// Create windows for each GPU
		for gpuNum, outputs := range windowsByGPU {
			window := xg.createWindowForGPU(gpuNum, outputs, &windowCounter, &areaCounter)
			if window != nil {
				screenXML.Windows = append(screenXML.Windows, *window)
			}
		}
	} else {
		// Area per Screen (GPU output) mode: Create one area per GPU output (each output is one screen)
		// Create a window for each GPU (group by GPU number, not GPU output)
		windowsByGPU := xg.groupByGPU(gpuGroups)

		// Create windows for each GPU, with one area per output
		for gpuNum, outputs := range windowsByGPU {
			window := xg.createWindowForScreen(gpuNum, outputs, &windowCounter, &areaCounter)
			if window != nil {
				screenXML.Windows = append(screenXML.Windows, *window)
			}
		}
	}

	// Generate XML with proper encoding and formatting
	xmlData, err := xml.MarshalIndent(screenXML, "", "  ") // Use 2 spaces for indentation
	if err != nil {
		return nil, fmt.Errorf("failed to marshal XML: %w", err)
	}

	// Add XML header with UTF-8 encoding and DOCTYPE
	var buf []byte
	buf = append(buf, []byte(`<?xml version="1.0" encoding="UTF-8"?>`+"\n")...)
	buf = append(buf, []byte(`<!DOCTYPE mtdoc>`+"\n")...)
	buf = append(buf, xmlData...)

	return buf, nil
}

// calculateLayerSize calculates the total layer size from cells with layer > 0.
// Only considers cells that have a layer index > 0 (not layer 0 or empty).
// Uses effective dimensions based on cell orientation.
func (xg *XMLGenerator) calculateLayerSize() string {
	// Find the maximum graphics coordinates from cells with layer > 0
	maxX, maxY := 0, 0

	rows := xg.grid.GetRows()
	cols := xg.grid.GetCols()
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := xg.grid.GetCell(row, col)
			if cell != nil && cell.GPUOutput != "" {
				// Check if cell has a layer index > 0
				hasLayer := false
				if cell.Index != "" {
					var layerIdx int
					if _, err := fmt.Sscanf(cell.Index, "%d", &layerIdx); err == nil && layerIdx > 0 {
						hasLayer = true
					}
				}

				// Only consider cells with layer > 0
				if hasLayer {
					// Calculate cell's graphics coordinates using effective dimensions
					effWidth := cell.Resolution.EffectiveWidth(cell.Orientation)
					effHeight := cell.Resolution.EffectiveHeight(cell.Orientation)
					x := col * effWidth
					y := row * effHeight
					cellMaxX := x + effWidth
					cellMaxY := y + effHeight

					if cellMaxX > maxX {
						maxX = cellMaxX
					}
					if cellMaxY > maxY {
						maxY = cellMaxY
					}
				}
			}
		}
	}

	// If no cells with layer assigned, return "0 0" for auto-calculation
	if maxX == 0 && maxY == 0 {
		return "0 0"
	}

	return fmt.Sprintf("%d %d", maxX, maxY)
}

// groupByGPU groups GPU outputs by GPU number.
func (xg *XMLGenerator) groupByGPU(gpuGroups map[string][]CellCoord) map[int]map[string][]CellCoord {
	windowsByGPU := make(map[int]map[string][]CellCoord)

	for gpuOutput, cells := range gpuGroups {
		parts := strings.Split(gpuOutput, ":")
		if len(parts) != 2 {
			continue
		}
		var gpuNum int
		fmt.Sscanf(parts[0], "%d", &gpuNum)

		if windowsByGPU[gpuNum] == nil {
			windowsByGPU[gpuNum] = make(map[string][]CellCoord)
		}
		windowsByGPU[gpuNum][gpuOutput] = cells
	}

	return windowsByGPU
}

// createWindowForGPU creates a WindowConfig with a single area matching the window size.
// This is used when "area per GPU" mode is enabled (one area per window, matching all outputs on that GPU).
func (xg *XMLGenerator) createWindowForGPU(gpuNum int, outputs map[string][]CellCoord, windowCounter *int, areaCounter *int) *WindowConfig {
	if len(outputs) == 0 {
		return nil
	}

	// Calculate window size from all outputs
	minX, minY, maxX, maxY := xg.calculateWindowBounds(outputs)
	windowWidth := maxX - minX
	windowHeight := maxY - minY

	windowName := fmt.Sprintf("window%d", *windowCounter)
	*windowCounter++

	outputList := formatOutputList(outputs)

	window := &WindowConfig{
		Comment: fmt.Sprintf(`The window element defines a drawable region available for displaying applications.
It extends across the Cells' screen surface and is similar to a computer desktop.
It is also sometimes called the operating system window. For efficient rendering,
we recommend that you define one window element per GPU.

Window %s for GPU %d. Covers outputs %s in graphics bounds (%d,%d) to (%d,%d) (%d x %d px).`,
			windowName, gpuNum, outputList, minX, minY, maxX, maxY, windowWidth, windowHeight),
		Type:            windowName,
		DirectRendering: &XMLAttr{Type: "", Value: "1"},
		DirectRenderingComment: `direct-rendering: Improves performance when not doing color correction or inter-GPU frame locking.
If enabled, the viewports / areas on this window are rendered directly to the window. This gives
the best performance. If disabled, rendering is done using off-screen buffers that adds overhead,
but is required for frame lock and area color correction to work. Recommended to leave on (1).`,
		Frameless:       &XMLAttr{Type: "", Value: "1"},
		FramelessComment: `frameless: Enable or disable frameless window mode. Frameless window doesn't have borders,
title bar, system menu or minimize/maximize/close buttons, can't be moved or resized, and
disables OS touch gestures on top of the window. Frameless mode is the recommended way of
configuring wall installations.

 0: Normal application window with window frames, title bar, close buttons etc,
 1: Window has no frames. Similar to fullscreen-mode, but isn't restricted to
   one screen since the window can be arbitrary size.

 Should the window stay on top of other windows. The default value
 normally is 0 (disabled). If the window is frameless, the default
 for this is 1 (enabled).`,
		FsaaSamples:     &XMLAttr{Type: "", Value: "4"},
		FsaaSamplesComment: ` Full-screen anti-aliasing samples. Typical values are 0, 2 or 4.
 If not defined, a reasonable default value is chosen based on the
 hardware capabilities.`,
		Fullscreen:      &XMLAttr{Type: "", Value: "0"},
		FullscreenComment: ` Create the application window in full-screen mode.  NB this works well for windows using mosaic but will fail otherwise as it MS Windows only supports fullscreen toggle to a single desktop area.`,
		Location:        &XMLAttr{Type: "", Value: fmt.Sprintf("%d %d", minX, minY)},
		LocationComment: `location: Specifies the desktop pixel coordinates of the window's top-left corner.
Origin (0, 0) is in the top-left corner of the primary display. If location is not defined,
the window is located on the center of the screen by default.`,
		Resizable:       &XMLAttr{Type: "", Value: "0"},
		ResizableComment: ` Can the application window be resized.`,
		ScreenNumber:    &XMLAttr{Type: "", Value: "-1"},
		ScreenNumberComment: `screennumber: X screen number starting from 0. The location of the window is relative to the
selected X screen. Only used in Linux. Default: -1 which selects the current X screen.`,
		Size:            &XMLAttr{Type: "", Value: fmt.Sprintf("%d %d", windowWidth, windowHeight)},
		SizeComment: `size: Specifies the window width and height in pixels. Like location, this only affects
the window size, it doesn't affect what is rendered inside the window. If size is not defined,
it is set automatically to 100% of the display in frameless mode or 80% of the display in windowed mode.`,
		Areas:           []AreaConfig{},
	}

	// Create a single area that matches the window size (all outputs on this GPU)
	// Collect all cells from all outputs
	allCells := []CellCoord{}
	for _, cells := range outputs {
		allCells = append(allCells, cells...)
	}

	// Create one area for the entire window
	area := xg.createAreaForWindow(allCells, minX, minY, windowWidth, windowHeight,
		fmt.Sprintf("Area %s spans entire GPU %d window covering outputs %s.", fmt.Sprintf("area%d", *areaCounter), gpuNum, outputList),
		areaCounter)
	if area != nil {
		window.Areas = append(window.Areas, *area)
		*areaCounter++
	}

	return window
}

// calculateWindowBounds calculates the bounding box for a window from its outputs.
// Uses effective dimensions based on cell orientation.
func (xg *XMLGenerator) calculateWindowBounds(outputs map[string][]CellCoord) (minX, minY, maxX, maxY int) {
	first := true
	for _, cells := range outputs {
		for _, cell := range cells {
			cellState := xg.grid.GetCell(cell.Row, cell.Col)
			if cellState == nil {
				continue
			}
			effWidth := cellState.Resolution.EffectiveWidth(cellState.Orientation)
			effHeight := cellState.Resolution.EffectiveHeight(cellState.Orientation)
			x := cell.Col * effWidth
			y := cell.Row * effHeight
			cellMaxX := x + effWidth
			cellMaxY := y + effHeight

			if first {
				minX, minY, maxX, maxY = x, y, cellMaxX, cellMaxY
				first = false
			} else {
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if cellMaxX > maxX {
					maxX = cellMaxX
				}
				if cellMaxY > maxY {
					maxY = cellMaxY
				}
			}
		}
	}
	return
}

// groupCellsByGPU groups grid cells by their assigned GPU output.
func (xg *XMLGenerator) groupCellsByGPU() map[string][]CellCoord {
	groups := make(map[string][]CellCoord)

	rows := xg.grid.GetRows()
	cols := xg.grid.GetCols()
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := xg.grid.GetCell(row, col)
			if cell != nil && cell.GPUOutput != "" {
				groups[cell.GPUOutput] = append(groups[cell.GPUOutput], CellCoord{Row: row, Col: col})
			}
		}
	}

	return groups
}

// CellCoord represents a cell coordinate.
type CellCoord struct {
	Row int
	Col int
}

// createAreaForGPU creates an AreaConfig for a GPU output group.
// Uses effective dimensions based on cell orientation.
func (xg *XMLGenerator) createAreaForGPU(gpuOutput string, cells []CellCoord, areaCounter *int) *AreaConfig {
	if len(cells) == 0 {
		return nil
	}

	// Calculate bounding box for the cells
	minRow, maxRow := cells[0].Row, cells[0].Row
	minCol, maxCol := cells[0].Col, cells[0].Col

	// Get resolution and orientation from first cell
	firstCell := xg.grid.GetCell(cells[0].Row, cells[0].Col)
	if firstCell == nil {
		return nil
	}
	cellWidth := firstCell.Resolution.EffectiveWidth(firstCell.Orientation)
	cellHeight := firstCell.Resolution.EffectiveHeight(firstCell.Orientation)

	for _, cell := range cells {
		if cell.Row < minRow {
			minRow = cell.Row
		}
		if cell.Row > maxRow {
			maxRow = cell.Row
		}
		if cell.Col < minCol {
			minCol = cell.Col
		}
		if cell.Col > maxCol {
			maxCol = cell.Col
		}
	}

	// Calculate graphics coordinates (relative to layer)
	x := minCol * cellWidth
	y := minRow * cellHeight
	width := (maxCol - minCol + 1) * cellWidth
	height := (maxRow - minRow + 1) * cellHeight

	// Parse GPU output to determine type
	// Format: "gpu#:output#" (e.g., "1:2" means GPU 0, output 1)
	parts := strings.Split(gpuOutput, ":")
	if len(parts) != 2 {
		return nil
	}

	var gpuNum, outputNum int
	fmt.Sscanf(parts[0], "%d", &gpuNum)
	fmt.Sscanf(parts[1], "%d", &outputNum)
	// Note: area type is not used in the XML structure - it's determined by the GPU output

	areaName := fmt.Sprintf("area%d", *areaCounter)

	// Include orientation info in comment
	orientStr := "landscape"
	if firstCell.Orientation == OrientationPortrait {
		orientStr = "portrait"
	}

	area := &AreaConfig{
		Comment: fmt.Sprintf(`In the application virtual graphics coordinates, graphicslocation and graphicssize
define the part of the application that is rendered to this area / viewport.

Area %s drives GPU output %s covering rows %d-%d, cols %d-%d (%d x %d px at %dx%d resolution, %s).`,
			areaName, gpuOutput, minRow, maxRow, minCol, maxCol, width, height, cellWidth, cellHeight, orientStr),
		Type:             areaName,
		GraphicsLocation: &XMLAttr{Type: "", Value: fmt.Sprintf("%d %d", x, y)},
		GraphicsLocationComment: `This doesn't affect where on the window the viewport is rendered, but it defines what part of
application is rendered here.`,
		GraphicsSize:     &XMLAttr{Type: "", Value: fmt.Sprintf("%d %d", width, height)},
		GraphicsSizeComment: `The graphics size doesn't need to be the same as area size or even have the same aspect ratio.
The given part of the application is rendered so that it fills the whole area. Different areas
can render arbitrary parts of the application, even if the parts overlap.

However, in a typical use case the graphics size does match the area size so that we have 1:1
pixel mapping from the virtual application graphics coordinates to the window coordinates so
that all the UI elements have correct size.`,
		Location:         &XMLAttr{Type: "", Value: "0 0"},
		Seams:            &XMLAttr{Type: "", Value: "0 0 0 0"},
		Size:             &XMLAttr{Type: "", Value: fmt.Sprintf("%d %d", width, height)},
	}

	return area
}

// createWindowForScreen creates a WindowConfig with one area per GPU output (per screen).
// This is used when "area per Screen" mode is enabled (each GPU output = one screen).
func (xg *XMLGenerator) createWindowForScreen(gpuNum int, outputs map[string][]CellCoord, windowCounter *int, areaCounter *int) *WindowConfig {
	if len(outputs) == 0 {
		return nil
	}

	// Calculate window size from all outputs
	minX, minY, maxX, maxY := xg.calculateWindowBounds(outputs)
	windowWidth := maxX - minX
	windowHeight := maxY - minY

	windowName := fmt.Sprintf("window%d", *windowCounter)
	*windowCounter++
	outputList := formatOutputList(outputs)

	window := &WindowConfig{
		Comment: fmt.Sprintf(`The window element defines a drawable region available for displaying applications.
It extends across the Cells' screen surface and is similar to a computer desktop.
It is also sometimes called the operating system window. For efficient rendering,
we recommend that you define one window element per GPU.

Window %s for GPU %d. Covers outputs %s in graphics bounds (%d,%d) to (%d,%d) (%d x %d px).`,
			windowName, gpuNum, outputList, minX, minY, maxX, maxY, windowWidth, windowHeight),
		Type:            windowName,
		DirectRendering: &XMLAttr{Type: "", Value: "1"},
		DirectRenderingComment: `direct-rendering: Improves performance when not doing color correction or inter-GPU frame locking.
If enabled, the viewports / areas on this window are rendered directly to the window. This gives
the best performance. If disabled, rendering is done using off-screen buffers that adds overhead,
but is required for frame lock and area color correction to work. Recommended to leave on (1).`,
		Frameless:       &XMLAttr{Type: "", Value: "1"},
		FramelessComment: `frameless: Enable or disable frameless window mode. Frameless window doesn't have borders,
title bar, system menu or minimize/maximize/close buttons, can't be moved or resized, and
disables OS touch gestures on top of the window. Frameless mode is the recommended way of
configuring wall installations.

 0: Normal application window with window frames, title bar, close buttons etc,
 1: Window has no frames. Similar to fullscreen-mode, but isn't restricted to
   one screen since the window can be arbitrary size.

 Should the window stay on top of other windows. The default value
 normally is 0 (disabled). If the window is frameless, the default
 for this is 1 (enabled).`,
		FsaaSamples:     &XMLAttr{Type: "", Value: "4"},
		FsaaSamplesComment: ` Full-screen anti-aliasing samples. Typical values are 0, 2 or 4.
 If not defined, a reasonable default value is chosen based on the
 hardware capabilities.`,
		Fullscreen:      &XMLAttr{Type: "", Value: "0"},
		FullscreenComment: ` Create the application window in full-screen mode.  NB this works well for windows using mosaic but will fail otherwise as it MS Windows only supports fullscreen toggle to a single desktop area.`,
		Location:        &XMLAttr{Type: "", Value: fmt.Sprintf("%d %d", minX, minY)},
		LocationComment: `location: Specifies the desktop pixel coordinates of the window's top-left corner.
Origin (0, 0) is in the top-left corner of the primary display. If location is not defined,
the window is located on the center of the screen by default.`,
		Resizable:       &XMLAttr{Type: "", Value: "0"},
		ResizableComment: ` Can the application window be resized.`,
		ScreenNumber:    &XMLAttr{Type: "", Value: "-1"},
		ScreenNumberComment: `screennumber: X screen number starting from 0. The location of the window is relative to the
selected X screen. Only used in Linux. Default: -1 which selects the current X screen.`,
		Size:            &XMLAttr{Type: "", Value: fmt.Sprintf("%d %d", windowWidth, windowHeight)},
		SizeComment: `size: Specifies the window width and height in pixels. Like location, this only affects
the window size, it doesn't affect what is rendered inside the window. If size is not defined,
it is set automatically to 100% of the display in frameless mode or 80% of the display in windowed mode.`,
		Areas:           []AreaConfig{},
	}

	// Create one area for each GPU output (each output = one screen)
	for gpuOutput, cells := range outputs {
		area := xg.createAreaForGPU(gpuOutput, cells, areaCounter)
		if area != nil {
			window.Areas = append(window.Areas, *area)
			*areaCounter++
		}
	}

	return window
}

// createAreaForWindow creates a single AreaConfig that matches the window size.
func (xg *XMLGenerator) createAreaForWindow(cells []CellCoord, x, y, width, height int, comment string, areaCounter *int) *AreaConfig {
	if len(cells) == 0 {
		return nil
	}

	areaName := fmt.Sprintf("area%d", *areaCounter)
	finalComment := comment
	if finalComment == "" {
		finalComment = fmt.Sprintf(`In the application virtual graphics coordinates, graphicslocation and graphicssize
define the part of the application that is rendered to this area / viewport.

Area %s spans window bounds (%d,%d) size %dx%d px for aggregated GPU outputs.`, areaName, x, y, width, height)
	}

	area := &AreaConfig{
		Comment:          finalComment,
		Type:             areaName,
		GraphicsLocation: &XMLAttr{Type: "", Value: fmt.Sprintf("%d %d", x, y)},
		GraphicsLocationComment: `This doesn't affect where on the window the viewport is rendered, but it defines what part of
application is rendered here.`,
		GraphicsSize:     &XMLAttr{Type: "", Value: fmt.Sprintf("%d %d", width, height)},
		GraphicsSizeComment: `The graphics size doesn't need to be the same as area size or even have the same aspect ratio.
The given part of the application is rendered so that it fills the whole area. Different areas
can render arbitrary parts of the application, even if the parts overlap.

However, in a typical use case the graphics size does match the area size so that we have 1:1
pixel mapping from the virtual application graphics coordinates to the window coordinates so
that all the UI elements have correct size.`,
		Location:         &XMLAttr{Type: "", Value: "0 0"},
		Seams:            &XMLAttr{Type: "", Value: "0 0 0 0"},
		Size:             &XMLAttr{Type: "", Value: fmt.Sprintf("%d %d", width, height)},
	}

	return area
}

// Validate validates the generated XML structure.
func (xg *XMLGenerator) Validate(xmlData []byte) error {
	var screen ScreenXML
	if err := xml.Unmarshal(xmlData, &screen); err != nil {
		return fmt.Errorf("invalid XML structure: %w", err)
	}

	// Basic validation
	if len(screen.Windows) == 0 {
		return fmt.Errorf("no windows defined in screen.xml")
	}

	for i, window := range screen.Windows {
		if len(window.Areas) == 0 {
			return fmt.Errorf("window %d has no areas", i)
		}
		for j, area := range window.Areas {
			if area.Type == "" {
				return fmt.Errorf("window %d, area %d missing type attribute", i, j)
			}
			if area.GraphicsSize == nil || area.GraphicsSize.Value == "" {
				return fmt.Errorf("window %d, area %d missing graphics size", i, j)
			}
		}
	}

	return nil
}

// formatOutputList produces a sorted, human-readable list of GPU outputs.
func formatOutputList(outputs map[string][]CellCoord) string {
	if len(outputs) == 0 {
		return ""
	}
	names := make([]string, 0, len(outputs))
	for name := range outputs {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
