// Package zone turns a rectangle on the canvas into a live auto-layout
// widget: sources dropped into it are arranged by the layout package, and
// the controller reacts to people moving things the way they would expect.
//
// Gestures (all are ordinary Canvus drags, no new UI needed):
//
//   - Drop a widget into the zone       -> it joins (as a secondary, at the end).
//   - Drop it onto the large/main slot  -> it joins as the main source.
//   - Drag a member onto the main slot  -> it becomes main; the old main takes its place.
//   - Drag a member onto another tile   -> the two swap places.
//   - Drag a member out of the zone     -> it leaves; the rest re-flow in order.
//   - Drag a member and drop it nowhere -> it snaps back.
//   - Rename the zone's anchor          -> the mode changes (e.g. "Auto Layout: grid").
//
// The controller is a pure state machine fed with observations (Step). It
// never moves a widget while it is still moving: an observation must be
// unchanged for Settle before the controller acts on it, so it never fights
// a person's drag. It also recognises its own writes, so a widget it just
// placed is not mistaken for a user move.
package zone

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/autolayout/layout"
)

// Member is an observed widget near or in the zone, in canvas pixels.
type Member struct {
	ID   string
	Type string
	Rect layout.Rect
	// Aspect of the content; 0 means use Rect's aspect.
	Aspect float64
	// Live marks a screen share or other live source (video input, IP video,
	// RDP, browser). With exactly one live source and no promoted main, Auto
	// gives that source the large spot.
	Live bool
}

func (m Member) aspect() float64 {
	if m.Aspect > 0 {
		return m.Aspect
	}
	return m.Rect.Aspect()
}

// Spec describes the zone itself (normally a named Canvus anchor).
type Spec struct {
	ID   string
	Rect layout.Rect
	Mode layout.Mode
}

// Move is one widget the caller should place at Rect (canvas pixels).
type Move struct {
	ID   string
	Type string
	From layout.Rect
	To   layout.Rect
}

// Config tunes the controller. Zero values take the defaults.
type Config struct {
	Settle    time.Duration // how long a rect must be still before acting (default 600ms)
	Tolerance float64       // px; rects closer than this are equal (default 2)
	Layout    layout.Options
}

type obs struct {
	rect  layout.Rect
	since time.Time
}

// Controller holds one zone's state. It is not safe for concurrent use.
type Controller struct {
	cfg      Config
	slots    []string // slots[0] is main; the rest are secondaries in order
	promoted bool     // a person put slots[0] in the main spot on purpose
	expected map[string]layout.Rect
	seen     map[string]obs
	last     layout.Plan
	lastSpec Spec
}

// New returns a controller with no members.
func New(cfg Config) *Controller {
	if cfg.Settle <= 0 {
		cfg.Settle = 600 * time.Millisecond
	}
	if cfg.Tolerance <= 0 {
		cfg.Tolerance = 2
	}
	return &Controller{cfg: cfg, expected: map[string]layout.Rect{}, seen: map[string]obs{}}
}

// Members returns the current slot order (main first).
func (c *Controller) Members() []string { return append([]string(nil), c.slots...) }

// Promoted reports whether the main source was chosen by a person.
func (c *Controller) Promoted() bool { return c.promoted }

// Plan returns the most recent plan.
func (c *Controller) Plan() layout.Plan { return c.last }

func (c *Controller) near(a, b layout.Rect) bool {
	t := c.cfg.Tolerance
	return math.Abs(a.X-b.X) <= t && math.Abs(a.Y-b.Y) <= t &&
		math.Abs(a.W-b.W) <= t && math.Abs(a.H-b.H) <= t
}

func (c *Controller) index(id string) int {
	for i, s := range c.slots {
		if s == id {
			return i
		}
	}
	return -1
}

// slotAt returns the slot whose cell contains (x, y) in the last plan, or -1.
func (c *Controller) slotAt(x, y float64) int {
	for _, p := range c.last.Placements {
		if p.Cell.ContainsPoint(x, y) {
			return c.index(p.ID)
		}
	}
	return -1
}

func (c *Controller) remove(i int) {
	c.slots = append(c.slots[:i], c.slots[i+1:]...)
	if i == 0 {
		c.promoted = false
	}
}

// Step feeds one observation of the canvas and returns the moves to apply.
// members should include every candidate widget (inside the zone or not);
// widgets absent from members are treated as deleted.
func (c *Controller) Step(now time.Time, spec Spec, members []Member) []Move {
	byID := make(map[string]Member, len(members))
	for _, m := range members {
		byID[m.ID] = m
	}

	// Track stillness per widget.
	stable := map[string]bool{}
	for _, m := range members {
		o, ok := c.seen[m.ID]
		if !ok || !c.near(o.rect, m.Rect) {
			o = obs{rect: m.Rect, since: now}
			c.seen[m.ID] = o
		}
		stable[m.ID] = now.Sub(o.since) >= c.cfg.Settle
	}
	for id := range c.seen {
		if _, ok := byID[id]; !ok {
			delete(c.seen, id)
		}
	}

	dirty := spec != c.lastSpec

	// 1. Deleted members leave at once.
	for i := len(c.slots) - 1; i >= 0; i-- {
		if _, ok := byID[c.slots[i]]; !ok {
			delete(c.expected, c.slots[i])
			c.remove(i)
			dirty = true
		}
	}

	// 2. Settled user moves of existing members: leave, promote, swap or snap back.
	for _, id := range append([]string(nil), c.slots...) {
		m := byID[id]
		exp, placed := c.expected[id]
		if !stable[id] || (placed && c.near(exp, m.Rect)) {
			continue
		}
		cx, cy := m.Rect.Center()
		i := c.index(id)
		switch {
		case !spec.Rect.ContainsPoint(cx, cy):
			delete(c.expected, id)
			c.remove(i)
		default:
			if j := c.slotAt(cx, cy); j >= 0 && j != i {
				c.slots[i], c.slots[j] = c.slots[j], c.slots[i]
				if i == 0 || j == 0 {
					c.promoted = true
				}
			}
			// Otherwise: dropped on a gap or its own cell -> snap back.
		}
		dirty = true
	}

	// 3. Newcomers: settled widgets whose centre is in the zone.
	var newcomers []Member
	for _, m := range members {
		if c.index(m.ID) >= 0 || !stable[m.ID] {
			continue
		}
		cx, cy := m.Rect.Center()
		if spec.Rect.ContainsPoint(cx, cy) {
			newcomers = append(newcomers, m)
		}
	}
	// Reading order, so a batch dropped together joins predictably.
	sort.SliceStable(newcomers, func(a, b int) bool {
		ra, rb := newcomers[a].Rect, newcomers[b].Rect
		if math.Abs(ra.Y-rb.Y) > c.cfg.Tolerance {
			return ra.Y < rb.Y
		}
		return ra.X < rb.X
	})
	if len(c.slots) == 0 && len(newcomers) > 1 {
		// Cold start (e.g. the tool restarted over an arranged zone): the
		// largest widget is taken as main, and as promoted if it is clearly
		// bigger than the rest, so a large-left/right/centre zone survives.
		big := 0
		for i, m := range newcomers {
			if m.Rect.Area() > newcomers[big].Rect.Area() {
				big = i
			}
		}
		second := 0.0
		for i, m := range newcomers {
			if i != big && m.Rect.Area() > second {
				second = m.Rect.Area()
			}
		}
		lead := newcomers[big]
		newcomers = append(append([]Member{lead}, newcomers[:big]...), newcomers[big+1:]...)
		c.slots = append(c.slots, lead.ID)
		newcomers = newcomers[1:]
		c.promoted = lead.Rect.Area() >= 1.5*second
		dirty = true
	}
	for _, m := range newcomers {
		cx, cy := m.Rect.Center()
		// A drop only claims the main slot when there is a visibly large spot.
		// In a grid (or a single full-zone source) newcomers queue at the end;
		// promoting there is a deliberate drag of an existing member.
		onMain := isLarge(c.last.Resolved) && c.slotAt(cx, cy) == 0
		if onMain {
			c.slots = append([]string{m.ID}, c.slots...)
			c.promoted = true
		} else {
			c.slots = append(c.slots, m.ID)
		}
		dirty = true
	}

	// Exactly one live share among static content takes the main slot until
	// a person promotes something else.
	if !c.promoted && len(c.slots) > 1 {
		live := -1
		for i, id := range c.slots {
			if byID[id].Live {
				if live >= 0 {
					live = -2
					break
				}
				live = i
			}
		}
		if live > 0 {
			id := c.slots[live]
			c.slots = append([]string{id}, append(c.slots[:live:live], c.slots[live+1:]...)...)
			dirty = true
		}
	}

	if !dirty && len(c.last.Placements) == len(c.slots) {
		return nil
	}
	c.lastSpec = spec
	return c.relayout(spec, byID, stable)
}

func (c *Controller) relayout(spec Spec, byID map[string]Member, stable map[string]bool) []Move {
	order := c.slots
	emphasis := c.promoted || (len(order) > 1 && c.singleLive(byID))

	in := layout.Input{Zone: spec.Rect, Emphasis: emphasis}
	for i, id := range order {
		it := layout.Item{ID: id, Aspect: byID[id].aspect()}
		if i == 0 {
			in.Main = &it
		} else {
			in.Secondary = append(in.Secondary, it)
		}
	}
	c.last = layout.Compute(spec.Mode, in, c.cfg.Layout)

	var moves []Move
	for _, p := range c.last.Placements {
		m := byID[p.ID]
		c.expected[p.ID] = p.Content
		if !stable[p.ID] || c.near(m.Rect, p.Content) {
			continue
		}
		moves = append(moves, Move{ID: p.ID, Type: m.Type, From: m.Rect, To: p.Content})
	}
	return moves
}

func isLarge(m layout.Mode) bool {
	return m == layout.LargeLeft || m == layout.LargeRight || m == layout.LargeCenter
}

// singleLive reports whether exactly one member is live and it holds slot 0.
func (c *Controller) singleLive(byID map[string]Member) bool {
	n := 0
	for _, id := range c.slots {
		if byID[id].Live {
			n++
		}
	}
	return n == 1 && byID[c.slots[0]].Live
}

var anchorRe = regexp.MustCompile(`(?i)^\s*auto[\s_-]*layout\b\s*[:\-–]?\s*(.*)$`)

// ParseAnchorName reports whether an anchor name marks an auto-layout zone
// and which mode it asks for: "Auto Layout", "Auto Layout: large left",
// "autolayout - grid". A zone with no or an unknown mode uses Auto.
func ParseAnchorName(name string) (layout.Mode, bool) {
	m := anchorRe.FindStringSubmatch(name)
	if m == nil {
		return layout.Auto, false
	}
	mode, _ := layout.ParseMode(strings.TrimSpace(m[1]))
	return mode, true
}

// IsLiveType reports whether a Canvus widget type is a live source.
func IsLiveType(widgetType string) bool {
	switch normType(widgetType) {
	case "videoinput", "ipvideo", "rdpconnection", "browser", "videostream":
		return true
	}
	return false
}

// IsArrangeable reports whether a widget type can be a member of a zone.
func IsArrangeable(widgetType string) bool {
	switch normType(widgetType) {
	case "", "anchor", "connector", "sharedcanvas", "annotation":
		return false
	}
	return true
}

func normType(t string) string {
	return strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.ToLower(t))
}
