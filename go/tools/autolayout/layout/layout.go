// Package layout computes split-screen arrangements for a set of sources
// (screen shares, videos, images, notes...) inside a rectangular zone.
//
// It is pure geometry: no Canvus calls, no clocks, no randomness. The same
// input always yields the same plan, which keeps live re-layouts stable and
// makes every mode testable.
//
// Five modes are supported:
//
//   - LargeLeft:   main source on the left, the others tiled in a strip on the right.
//   - LargeRight:  mirror of LargeLeft.
//   - LargeCenter: main source in the middle, the others split across flanking strips.
//   - Grid:        every source gets an equal cell; main takes the first (top-left) cell.
//   - Auto:        picks one of the above from the source count, whether the main
//     source should be emphasised, the zone's shape and the sources' aspect ratios.
//
// Sources are never cropped or stretched: each is fitted inside its cell with
// its aspect ratio preserved and centred (letterboxed), because Canvus widgets
// scale uniformly.
package layout

import (
	"fmt"
	"math"
	"strings"
)

// Rect is an axis-aligned rectangle in canvas pixels.
type Rect struct {
	X, Y, W, H float64
}

// Area returns W*H (0 for degenerate rects).
func (r Rect) Area() float64 {
	if r.W <= 0 || r.H <= 0 {
		return 0
	}
	return r.W * r.H
}

// Center returns the centre point of r.
func (r Rect) Center() (float64, float64) { return r.X + r.W/2, r.Y + r.H/2 }

// ContainsPoint reports whether (x, y) lies inside r (edges inclusive).
func (r Rect) ContainsPoint(x, y float64) bool {
	return x >= r.X && x <= r.X+r.W && y >= r.Y && y <= r.Y+r.H
}

// Inset shrinks r by d on every side (never below zero size).
func (r Rect) Inset(d float64) Rect {
	out := Rect{X: r.X + d, Y: r.Y + d, W: r.W - 2*d, H: r.H - 2*d}
	if out.W < 0 {
		out.W = 0
	}
	if out.H < 0 {
		out.H = 0
	}
	return out
}

// Aspect returns W/H, or 0 for a degenerate rect.
func (r Rect) Aspect() float64 {
	if r.H <= 0 {
		return 0
	}
	return r.W / r.H
}

// Mode selects an arrangement.
type Mode int

const (
	Auto Mode = iota
	LargeLeft
	LargeRight
	LargeCenter
	Grid
	// Fill is not user-selectable: it is what every mode does with one source.
	Fill
)

var modeNames = map[Mode]string{
	Auto: "auto", LargeLeft: "large-left", LargeRight: "large-right",
	LargeCenter: "large-center", Grid: "grid", Fill: "fill",
}

func (m Mode) String() string {
	if s, ok := modeNames[m]; ok {
		return s
	}
	return fmt.Sprintf("Mode(%d)", int(m))
}

// ParseMode accepts the names users are likely to type on a canvas:
// "large left", "Large-Left", "left", "large centre", "center", "even grid",
// "grid", "auto". Unknown or empty text returns (Auto, false).
func ParseMode(s string) (Mode, bool) {
	t := strings.ToLower(strings.TrimSpace(s))
	t = strings.NewReplacer("_", " ", "-", " ", ":", " ").Replace(t)
	t = strings.Join(strings.Fields(t), " ")
	switch t {
	case "auto", "automatic", "smart":
		return Auto, true
	case "large left", "left", "main left", "big left":
		return LargeLeft, true
	case "large right", "right", "main right", "big right":
		return LargeRight, true
	case "large center", "large centre", "center", "centre", "main center", "main centre", "big center", "big centre":
		return LargeCenter, true
	case "grid", "even grid", "even", "equal", "tile", "tiles":
		return Grid, true
	}
	return Auto, false
}

// Item is one source to place. Aspect is width/height of its content; 0 means
// "unknown", in which case it fills whatever cell it gets.
type Item struct {
	ID     string
	Aspect float64
}

// Input is what a layout is computed from. Main is placed in the large spot;
// Secondary keeps the caller's order (arrival order), which is what keeps
// tiles from shuffling when one source is added or removed.
type Input struct {
	Zone      Rect
	Main      *Item
	Secondary []Item
	// Emphasis tells Auto that the main source deserves the large spot (a
	// user promoted it, or it is the only live screen share). Without it,
	// Auto treats every source as equal and uses the grid.
	Emphasis bool
}

// Options tune spacing and proportions. Zero values take the defaults.
type Options struct {
	// GapFrac is the gap between tiles as a fraction of the zone's shorter side.
	GapFrac float64
	// PaddingFrac is the inner margin of the zone, same unit as GapFrac.
	PaddingFrac float64
	// MinMainFrac is the smallest share of the split axis the main region may
	// take in the large modes (default 0.5 for left/right, 0.4 for centre).
	MinMainFrac float64
	// MinStripFrac is the smallest share of the split axis one strip may take
	// (default 0.18), so secondaries never collapse into slivers.
	MinStripFrac float64
	// MinTileFrac is Auto's legibility floor: if an emphasised layout would
	// leave a secondary tile smaller than this fraction of the zone area, Auto
	// falls back to the grid (default 0.02).
	MinTileFrac float64
	// WideZone is the zone aspect at or above which Auto prefers LargeCenter
	// once there are enough secondaries to flank both sides (default 2.0, e.g.
	// a 2x1 or 3x1 video wall).
	WideZone float64
}

func (o Options) withDefaults() Options {
	if o.GapFrac <= 0 {
		o.GapFrac = 0.015
	}
	if o.PaddingFrac < 0 {
		o.PaddingFrac = 0
	} else if o.PaddingFrac == 0 {
		o.PaddingFrac = 0.015
	}
	if o.MinStripFrac <= 0 {
		o.MinStripFrac = 0.18
	}
	if o.MinTileFrac <= 0 {
		o.MinTileFrac = 0.02
	}
	if o.WideZone <= 0 {
		o.WideZone = 2.0
	}
	return o
}

// Role tells a placement apart.
type Role int

const (
	RoleMain Role = iota
	RoleSecondary
)

func (r Role) String() string {
	if r == RoleMain {
		return "main"
	}
	return "secondary"
}

// Placement is one source's slot. Cell is the slot it owns (used for drop
// targeting); Content is where the source itself goes after aspect fitting.
type Placement struct {
	ID      string
	Role    Role
	Slot    int // 0 = main, 1.. = secondaries in input order
	Cell    Rect
	Content Rect
}

// Plan is the result of a layout.
type Plan struct {
	Requested  Mode
	Resolved   Mode // what was actually drawn (Auto resolves; one source is Fill)
	Placements []Placement
	// Coverage is the fraction of the inner zone covered by content (0..1),
	// a quality measure: higher means less letterboxing and wasted space.
	Coverage float64
}

// Compute arranges in.Main and in.Secondary in in.Zone using mode.
func Compute(mode Mode, in Input, opts Options) Plan {
	o := opts.withDefaults()
	plan := Plan{Requested: mode}
	if in.Zone.W <= 0 || in.Zone.H <= 0 {
		plan.Resolved = mode
		return plan
	}
	items := all(in)
	short := math.Min(in.Zone.W, in.Zone.H)
	gap := o.GapFrac * short
	inner := in.Zone.Inset(o.PaddingFrac * short)

	resolved := mode
	switch {
	case len(items) == 0:
		if resolved == Auto {
			resolved = Grid
		}
	case len(items) == 1:
		resolved = Fill
	case mode == Auto:
		resolved = chooseAuto(in, inner, gap, o)
	}
	plan.Resolved = resolved

	cells := cellsFor(resolved, inner, gap, items, o)
	for i, it := range items {
		role := RoleSecondary
		if i == 0 && in.Main != nil {
			role = RoleMain
		}
		plan.Placements = append(plan.Placements, Placement{
			ID: it.ID, Role: role, Slot: slotIndex(i, in.Main != nil), Cell: cells[i],
			Content: Fit(it.Aspect, cells[i]),
		})
	}
	plan.Coverage = coverage(plan.Placements, inner)
	return plan
}

func slotIndex(i int, hasMain bool) int {
	if hasMain {
		return i
	}
	return i + 1
}

func all(in Input) []Item {
	var out []Item
	if in.Main != nil {
		out = append(out, *in.Main)
	}
	return append(out, in.Secondary...)
}

func coverage(ps []Placement, inner Rect) float64 {
	if inner.Area() == 0 {
		return 0
	}
	var a float64
	for _, p := range ps {
		a += p.Content.Area()
	}
	return a / inner.Area()
}

// Fit returns the largest rect of the given aspect (w/h) centred in cell.
// Aspect <= 0 fills the cell.
func Fit(aspect float64, cell Rect) Rect {
	if aspect <= 0 || cell.W <= 0 || cell.H <= 0 {
		return cell
	}
	w, h := cell.W, cell.W/aspect
	if h > cell.H {
		h = cell.H
		w = h * aspect
	}
	return Rect{X: cell.X + (cell.W-w)/2, Y: cell.Y + (cell.H-h)/2, W: w, H: h}
}

// cellsFor returns one cell per item, in item order (main first if present).
func cellsFor(mode Mode, inner Rect, gap float64, items []Item, o Options) []Rect {
	n := len(items)
	switch {
	case n == 0:
		return nil
	case n == 1 || mode == Fill:
		return []Rect{inner}
	}
	switch mode {
	case LargeLeft, LargeRight:
		return largeSide(mode, inner, gap, items, o)
	case LargeCenter:
		return largeCenter(inner, gap, items, o)
	default:
		return BestGrid(inner, gap, aspects(items))
	}
}

func aspects(items []Item) []float64 {
	out := make([]float64, len(items))
	for i, it := range items {
		out[i] = it.Aspect
	}
	return out
}

// BestGrid tiles len(asp) equal cells in region, choosing the column count
// that shows the most content after aspect fitting. Cells fill row by row;
// a short last row is centred rather than left-aligned, so odd counts stay
// balanced instead of drifting into a diagonal.
func BestGrid(region Rect, gap float64, asp []float64) []Rect {
	n := len(asp)
	if n == 0 {
		return nil
	}
	bestCols, bestScore := 1, -1.0
	for cols := 1; cols <= n; cols++ {
		rows := (n + cols - 1) / cols
		cw := (region.W - gap*float64(cols-1)) / float64(cols)
		ch := (region.H - gap*float64(rows-1)) / float64(rows)
		if cw <= 0 || ch <= 0 {
			continue
		}
		var score float64
		for _, a := range asp {
			score += Fit(a, Rect{W: cw, H: ch}).Area()
		}
		// Prefer fewer empty cells on a tie (e.g. 2x2 over 4x1 for squares).
		empty := float64(rows*cols - n)
		score -= empty * 1e-9
		if score > bestScore+1e-6 {
			bestScore, bestCols = score, cols
		}
	}
	cols := bestCols
	rows := (n + cols - 1) / cols
	cw := (region.W - gap*float64(cols-1)) / float64(cols)
	ch := (region.H - gap*float64(rows-1)) / float64(rows)
	cells := make([]Rect, 0, n)
	for i := 0; i < n; i++ {
		r, c := i/cols, i%cols
		inRow := cols
		if r == rows-1 && n%cols != 0 {
			inRow = n % cols
		}
		offset := (float64(cols-inRow) * (cw + gap)) / 2
		cells = append(cells, Rect{
			X: region.X + offset + float64(c)*(cw+gap),
			Y: region.Y + float64(r)*(ch+gap),
			W: cw, H: ch,
		})
	}
	return cells
}

// mainWidth gives the main region its natural width at full zone height,
// clamped so the main region is never smaller than minMain of the axis and
// the strips never thinner than minStrip each.
func mainWidth(aspect float64, inner Rect, gap float64, strips int, minMain, minStrip float64) float64 {
	usable := inner.W - gap*float64(strips)
	natural := usable * minMain
	if aspect > 0 {
		natural = inner.H * aspect
	}
	lo := usable * minMain
	hi := usable * (1 - float64(strips)*minStrip)
	if hi < lo {
		hi = lo
	}
	return math.Max(lo, math.Min(hi, natural))
}

func largeSide(mode Mode, inner Rect, gap float64, items []Item, o Options) []Rect {
	minMain := o.MinMainFrac
	if minMain <= 0 {
		minMain = 0.5
	}
	mw := mainWidth(items[0].Aspect, inner, gap, 1, minMain, o.MinStripFrac)
	sw := inner.W - mw - gap
	main := Rect{X: inner.X, Y: inner.Y, W: mw, H: inner.H}
	strip := Rect{X: inner.X + mw + gap, Y: inner.Y, W: sw, H: inner.H}
	if mode == LargeRight {
		strip.X = inner.X
		main.X = inner.X + sw + gap
	}
	return append([]Rect{main}, BestGrid(strip, gap, aspects(items[1:]))...)
}

func largeCenter(inner Rect, gap float64, items []Item, o Options) []Rect {
	minMain := o.MinMainFrac
	if minMain <= 0 {
		minMain = 0.4
	}
	mw := mainWidth(items[0].Aspect, inner, gap, 2, minMain, o.MinStripFrac)
	sw := (inner.W - mw - 2*gap) / 2
	left := Rect{X: inner.X, Y: inner.Y, W: sw, H: inner.H}
	main := Rect{X: inner.X + sw + gap, Y: inner.Y, W: mw, H: inner.H}
	right := Rect{X: main.X + mw + gap, Y: inner.Y, W: sw, H: inner.H}

	// Alternate secondaries left, right, left... so both flanks fill evenly
	// and arrival order reads outward from the main source.
	var li, ri []int
	for i := 1; i < len(items); i++ {
		if (i-1)%2 == 0 {
			li = append(li, i)
		} else {
			ri = append(ri, i)
		}
	}
	cells := make([]Rect, len(items))
	cells[0] = main
	place := func(region Rect, idx []int) {
		asp := make([]float64, len(idx))
		for k, i := range idx {
			asp[k] = items[i].Aspect
		}
		for k, c := range BestGrid(region, gap, asp) {
			cells[idx[k]] = c
		}
	}
	place(left, li)
	place(right, ri)
	return cells
}

// chooseAuto picks a concrete mode for two or more sources.
//
//  1. No emphasis (nobody promoted a main source, no single live share): Grid.
//     Everyone is equal, as in Solstice's default grid or a T1V quad.
//  2. Emphasis: LargeLeft, or LargeCenter when the zone is wide (WideZone)
//     and there are at least four secondaries to flank both sides.
//  3. Legibility floor: if the emphasised layout would shrink any secondary
//     below MinTileFrac of the zone, use Grid instead.
func chooseAuto(in Input, inner Rect, gap float64, o Options) Mode {
	if !in.Emphasis || in.Main == nil {
		return Grid
	}
	items := all(in)
	pick := LargeLeft
	if inner.Aspect() >= o.WideZone && len(in.Secondary) >= 4 {
		pick = LargeCenter
	}
	cells := cellsFor(pick, inner, gap, items, o)
	floor := o.MinTileFrac * inner.Area()
	for i := 1; i < len(items); i++ {
		if Fit(items[i].Aspect, cells[i]).Area() < floor {
			return Grid
		}
	}
	return pick
}

// Lerp interpolates between two rects; t is clamped to [0,1] and eased
// (cubic in-out) so moves start and land gently.
func Lerp(a, b Rect, t float64) Rect {
	t = math.Max(0, math.Min(1, t))
	if t < 0.5 {
		t = 4 * t * t * t
	} else {
		t = 1 - math.Pow(-2*t+2, 3)/2
	}
	return Rect{
		X: a.X + (b.X-a.X)*t, Y: a.Y + (b.Y-a.Y)*t,
		W: a.W + (b.W-a.W)*t, H: a.H + (b.H-a.H)*t,
	}
}
