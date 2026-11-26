package zone

import (
	"testing"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/autolayout/layout"
)

var t0 = time.Date(2026, 11, 26, 9, 0, 0, 0, time.UTC)

const settle = 500 * time.Millisecond

func spec(mode layout.Mode) Spec {
	return Spec{ID: "zone", Rect: layout.Rect{X: 0, Y: 0, W: 1920, H: 1080}, Mode: mode}
}

// world is a fake canvas: it applies moves the way the Canvus server would.
type world struct {
	t   *testing.T
	c   *Controller
	now time.Time
	ws  map[string]Member
	ord []string
}

func newWorld(t *testing.T) *world {
	return &world{t: t, c: New(Config{Settle: settle}), now: t0, ws: map[string]Member{}}
}

func (w *world) put(id string, r layout.Rect, live bool) {
	if _, ok := w.ws[id]; !ok {
		w.ord = append(w.ord, id)
	}
	w.ws[id] = Member{ID: id, Type: "Image", Rect: r, Aspect: 16.0 / 9, Live: live}
}

func (w *world) del(id string) {
	delete(w.ws, id)
	for i, x := range w.ord {
		if x == id {
			w.ord = append(w.ord[:i], w.ord[i+1:]...)
		}
	}
}

func (w *world) members() []Member {
	var out []Member
	for _, id := range w.ord {
		out = append(out, w.ws[id])
	}
	return out
}

// tick advances time and runs one Step, applying the moves.
func (w *world) tick(s Spec, d time.Duration) []Move {
	w.now = w.now.Add(d)
	moves := w.c.Step(w.now, s, w.members())
	for _, m := range moves {
		mm := w.ws[m.ID]
		mm.Rect = m.To
		w.ws[m.ID] = mm
	}
	return moves
}

// settleAll runs ticks until the controller is quiet.
func (w *world) settleAll(s Spec) {
	for i := 0; i < 10; i++ {
		w.tick(s, settle)
	}
}

func (w *world) slots() []string { return w.c.Members() }

func box(x, y float64) layout.Rect { return layout.Rect{X: x, Y: y, W: 320, H: 180} }

func eq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestDropJoinsAfterSettle(t *testing.T) {
	w := newWorld(t)
	s := spec(layout.Grid)
	w.put("a", box(100, 100), false)
	if moves := w.tick(s, 0); len(moves) != 0 || len(w.slots()) != 0 {
		t.Fatal("a widget still moving must not join or be moved")
	}
	w.tick(s, settle)
	if !eq(w.slots(), []string{"a"}) {
		t.Fatalf("slots %v", w.slots())
	}
	// One member fills the zone.
	if got := w.ws["a"].Rect; got.W < 1800 {
		t.Errorf("single member not filling: %v", got)
	}
}

func TestOutsideWidgetIgnored(t *testing.T) {
	w := newWorld(t)
	w.put("far", box(5000, 5000), false)
	w.settleAll(spec(layout.Grid))
	if len(w.slots()) != 0 || w.ws["far"].Rect != box(5000, 5000) {
		t.Fatal("widget outside the zone was touched")
	}
}

func TestOwnWritesAreNotUserMoves(t *testing.T) {
	w := newWorld(t)
	s := spec(layout.Grid)
	w.put("a", box(100, 100), false)
	w.put("b", box(900, 100), false)
	w.settleAll(s)
	before := w.slots()
	if moves := w.tick(s, settle); len(moves) != 0 {
		t.Fatalf("controller keeps moving its own placements: %v", moves)
	}
	if !eq(before, w.slots()) {
		t.Fatal("order changed with no user action")
	}
}

func TestDragOntoMainPromotes(t *testing.T) {
	w := newWorld(t)
	s := spec(layout.LargeLeft)
	w.put("a", box(100, 100), false)
	w.settleAll(s)
	w.put("b", box(1500, 100), false)
	w.put("c", box(1500, 700), false)
	w.settleAll(s)
	if !eq(w.slots(), []string{"a", "b", "c"}) {
		t.Fatalf("slots %v", w.slots())
	}
	// Drag c onto the main cell.
	main := w.c.Plan().Placements[0].Cell
	cx, cy := main.Center()
	w.put("c", layout.Rect{X: cx - 160, Y: cy - 90, W: 320, H: 180}, false)
	w.settleAll(s)
	if !eq(w.slots(), []string{"c", "b", "a"}) {
		t.Fatalf("after promote slots %v", w.slots())
	}
	if !w.c.Promoted() {
		t.Error("promotion not recorded")
	}
}

func TestDragOntoTileSwaps(t *testing.T) {
	w := newWorld(t)
	s := spec(layout.Grid)
	for i, id := range []string{"a", "b", "c", "d"} {
		w.put(id, box(float64(100+i*400), 100), false)
	}
	w.settleAll(s)
	pl := w.c.Plan().Placements
	cx, cy := pl[3].Cell.Center() // d's cell
	w.put("b", layout.Rect{X: cx - 10, Y: cy - 10, W: 20, H: 20}, false)
	w.settleAll(s)
	if !eq(w.slots(), []string{"a", "d", "c", "b"}) {
		t.Fatalf("swap slots %v", w.slots())
	}
}

func TestDragOutLeavesAndReflows(t *testing.T) {
	w := newWorld(t)
	s := spec(layout.Grid)
	for i, id := range []string{"a", "b", "c"} {
		w.put(id, box(float64(100+i*500), 100), false)
	}
	w.settleAll(s)
	w.put("b", box(4000, 100), false)
	w.settleAll(s)
	if !eq(w.slots(), []string{"a", "c"}) {
		t.Fatalf("slots %v", w.slots())
	}
	if w.ws["b"].Rect != box(4000, 100) {
		t.Error("widget dragged out was moved back")
	}
}

func TestDropOnGapSnapsBack(t *testing.T) {
	w := newWorld(t)
	s := spec(layout.Grid)
	w.put("a", box(100, 100), false)
	w.put("b", box(1200, 100), false)
	w.settleAll(s)
	home := w.ws["a"].Rect
	// Nudge a slightly inside its own cell.
	w.put("a", layout.Rect{X: home.X + 40, Y: home.Y + 40, W: home.W, H: home.H}, false)
	w.settleAll(s)
	if w.ws["a"].Rect != home {
		t.Errorf("not snapped back: %v want %v", w.ws["a"].Rect, home)
	}
}

func TestDeleteMainPromotesNext(t *testing.T) {
	w := newWorld(t)
	s := spec(layout.LargeLeft)
	w.put("a", box(100, 100), false)
	w.settleAll(s)
	w.put("b", box(1500, 100), false)
	w.put("c", box(1500, 700), false)
	w.settleAll(s)
	w.del("a")
	w.tick(s, settle)
	if !eq(w.slots(), []string{"b", "c"}) {
		t.Fatalf("slots %v", w.slots())
	}
}

func TestSingleLiveShareTakesMainInAuto(t *testing.T) {
	w := newWorld(t)
	s := spec(layout.Auto)
	w.put("img1", box(100, 100), false)
	w.put("img2", box(900, 100), false)
	w.settleAll(s)
	if w.c.Plan().Resolved != layout.Grid {
		t.Fatalf("static content should be a grid, got %v", w.c.Plan().Resolved)
	}
	w.put("share", box(500, 700), true)
	w.settleAll(s)
	if w.slots()[0] != "share" || w.c.Plan().Resolved != layout.LargeLeft {
		t.Fatalf("live share should be main in large-left: %v %v", w.slots(), w.c.Plan().Resolved)
	}
	// A second live share means nobody is singled out: back to a grid.
	w.put("share2", box(1700, 900), true) // dropped on the strip, not the main spot
	w.settleAll(s)
	if w.c.Plan().Resolved != layout.Grid {
		t.Fatalf("two live shares should grid, got %v", w.c.Plan().Resolved)
	}
}

func TestModeChangeRelayouts(t *testing.T) {
	w := newWorld(t)
	w.put("a", box(100, 100), false)
	w.put("b", box(900, 100), false)
	w.settleAll(spec(layout.Grid))
	moves := w.tick(spec(layout.LargeRight), settle)
	if len(moves) == 0 || w.c.Plan().Resolved != layout.LargeRight {
		t.Fatal("mode change did not re-layout")
	}
}

func TestColdStartKeepsLargestAsMain(t *testing.T) {
	w := newWorld(t)
	w.put("small", layout.Rect{X: 100, Y: 100, W: 300, H: 169}, false)
	w.put("big", layout.Rect{X: 600, Y: 100, W: 1200, H: 675}, false)
	w.put("small2", layout.Rect{X: 100, Y: 500, W: 300, H: 169}, false)
	w.tick(spec(layout.Auto), 0)
	w.tick(spec(layout.Auto), settle)
	if w.slots()[0] != "big" || !w.c.Promoted() {
		t.Fatalf("cold start lost the main: %v promoted=%v", w.slots(), w.c.Promoted())
	}
}

func TestParseAnchorName(t *testing.T) {
	cases := []struct {
		in   string
		ok   bool
		mode layout.Mode
	}{
		{"Auto Layout", true, layout.Auto},
		{"Auto Layout: large left", true, layout.LargeLeft},
		{"autolayout - grid", true, layout.Grid},
		{"AUTO-LAYOUT large centre", true, layout.LargeCenter},
		{"Auto Layout: nonsense", true, layout.Auto},
		{"Meeting room", false, layout.Auto},
	}
	for _, c := range cases {
		m, ok := ParseAnchorName(c.in)
		if ok != c.ok || m != c.mode {
			t.Errorf("%q -> %v,%v want %v,%v", c.in, m, ok, c.mode, c.ok)
		}
	}
}

func TestTypes(t *testing.T) {
	if !IsLiveType("VideoInput") || !IsLiveType("ip_video") || IsLiveType("Image") {
		t.Error("IsLiveType")
	}
	if IsArrangeable("Anchor") || IsArrangeable("SharedCanvas") || !IsArrangeable("Note") {
		t.Error("IsArrangeable")
	}
}

func TestAutoGridDragToFirstTileEmphasises(t *testing.T) {
	w := newWorld(t)
	s := spec(layout.Auto)
	for i, id := range []string{"a", "b", "c"} {
		w.put(id, box(float64(100+i*500), 100), false)
	}
	w.settleAll(s)
	if w.c.Plan().Resolved != layout.Grid {
		t.Fatalf("want grid, got %v", w.c.Plan().Resolved)
	}
	cx, cy := w.c.Plan().Placements[0].Cell.Center()
	w.put("c", layout.Rect{X: cx - 10, Y: cy - 10, W: 20, H: 20}, false)
	w.settleAll(s)
	if w.slots()[0] != "c" || w.c.Plan().Resolved != layout.LargeLeft {
		t.Fatalf("drag to first tile should emphasise: %v %v", w.slots(), w.c.Plan().Resolved)
	}
}

func TestNewcomerDroppedOnLargeSpotBecomesMain(t *testing.T) {
	w := newWorld(t)
	s := spec(layout.LargeLeft)
	w.put("a", box(100, 100), false)
	w.put("b", box(900, 100), false)
	w.settleAll(s)
	cx, cy := w.c.Plan().Placements[0].Cell.Center()
	w.put("n", layout.Rect{X: cx - 10, Y: cy - 10, W: 20, H: 20}, false)
	w.settleAll(s)
	if !eq(w.slots(), []string{"n", "a", "b"}) {
		t.Fatalf("slots %v", w.slots())
	}
}

func TestNoOverlapAfterChurn(t *testing.T) {
	w := newWorld(t)
	s := spec(layout.Auto)
	for i := 0; i < 7; i++ {
		w.put(string(rune('a'+i)), box(float64(50+i*200), 50), i == 3)
		w.settleAll(s)
	}
	w.del("b")
	w.put("e", box(4000, 4000), true)
	w.settleAll(s)
	ps := w.c.Plan().Placements
	if len(ps) != 5 {
		t.Fatalf("want 5 members, got %d", len(ps))
	}
	for i := range ps {
		got := w.ws[ps[i].ID].Rect
		if got != ps[i].Content {
			t.Errorf("%s not at its placement", ps[i].ID)
		}
		for j := i + 1; j < len(ps); j++ {
			a, b := ps[i].Cell, ps[j].Cell
			if a.X < b.X+b.W && b.X < a.X+a.W && a.Y < b.Y+b.H && b.Y < a.Y+a.H {
				t.Errorf("cells %s and %s overlap", ps[i].ID, ps[j].ID)
			}
		}
	}
}
