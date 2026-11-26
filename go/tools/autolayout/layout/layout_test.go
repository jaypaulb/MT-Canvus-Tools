package layout

import (
	"fmt"
	"math"
	"testing"
)

const eps = 1e-6

var zone169 = Rect{X: 1000, Y: 500, W: 1920, H: 1080}

func items(n int, aspect float64) (*Item, []Item) {
	m := &Item{ID: "m", Aspect: aspect}
	var s []Item
	for i := 1; i < n; i++ {
		s = append(s, Item{ID: fmt.Sprintf("s%d", i), Aspect: aspect})
	}
	return m, s
}

func inside(outer, inner Rect) bool {
	return inner.X >= outer.X-eps && inner.Y >= outer.Y-eps &&
		inner.X+inner.W <= outer.X+outer.W+eps && inner.Y+inner.H <= outer.Y+outer.H+eps
}

func overlap(a, b Rect) bool {
	return a.X < b.X+b.W-eps && b.X < a.X+a.W-eps && a.Y < b.Y+b.H-eps && b.Y < a.Y+a.H-eps
}

// checkPlan asserts the invariants every plan must hold.
func checkPlan(t *testing.T, name string, p Plan, in Input) {
	t.Helper()
	want := len(in.Secondary)
	if in.Main != nil {
		want++
	}
	if len(p.Placements) != want {
		t.Fatalf("%s: got %d placements, want %d", name, len(p.Placements), want)
	}
	for i, a := range p.Placements {
		if !inside(in.Zone, a.Cell) {
			t.Errorf("%s: cell %d %v escapes zone %v", name, i, a.Cell, in.Zone)
		}
		if !inside(a.Cell, a.Content) {
			t.Errorf("%s: content %d escapes its cell", name, i)
		}
		if a.Cell.W <= 0 || a.Cell.H <= 0 {
			t.Errorf("%s: cell %d degenerate %v", name, i, a.Cell)
		}
		for j := i + 1; j < len(p.Placements); j++ {
			if overlap(a.Cell, p.Placements[j].Cell) {
				t.Errorf("%s: cells %d and %d overlap", name, i, j)
			}
		}
	}
	// Order is preserved: main first, then secondaries in input order.
	k := 0
	if in.Main != nil {
		if p.Placements[0].ID != in.Main.ID || p.Placements[0].Role != RoleMain {
			t.Errorf("%s: main not first", name)
		}
		k = 1
	}
	for i, s := range in.Secondary {
		if p.Placements[k+i].ID != s.ID {
			t.Errorf("%s: order changed at %d", name, i)
		}
	}
}

func TestAllModesInvariants(t *testing.T) {
	zones := []Rect{zone169, {W: 1080, H: 1920}, {W: 5760, H: 1080}, {W: 1000, H: 1000}}
	asps := []float64{16.0 / 9, 4.0 / 3, 9.0 / 16, 0}
	modes := []Mode{Auto, LargeLeft, LargeRight, LargeCenter, Grid}
	for _, z := range zones {
		for _, a := range asps {
			for n := 1; n <= 12; n++ {
				for _, m := range modes {
					for _, emph := range []bool{false, true} {
						main, sec := items(n, a)
						in := Input{Zone: z, Main: main, Secondary: sec, Emphasis: emph}
						name := fmt.Sprintf("zone=%v a=%.2f n=%d mode=%v emph=%v", z, a, n, m, emph)
						checkPlan(t, name, Compute(m, in, Options{}), in)
					}
				}
			}
		}
	}
}

func TestOneSourceFills(t *testing.T) {
	for _, m := range []Mode{Auto, LargeLeft, LargeRight, LargeCenter, Grid} {
		main, _ := items(1, 16.0/9)
		p := Compute(m, Input{Zone: zone169, Main: main}, Options{})
		if p.Resolved != Fill {
			t.Errorf("%v: resolved %v, want fill", m, p.Resolved)
		}
		// 16:9 in a 16:9 zone: only the uniform padding changes the shape slightly.
		if p.Coverage < 0.98 {
			t.Errorf("%v: coverage %.3f, want >= 0.98", m, p.Coverage)
		}
	}
}

func TestAspectPreserved(t *testing.T) {
	asp := []float64{16.0 / 9, 4.0 / 3, 9.0 / 16, 1, 21.0 / 9}
	m := &Item{ID: "m", Aspect: asp[0]}
	var sec []Item
	for i, a := range asp[1:] {
		sec = append(sec, Item{ID: fmt.Sprint(i), Aspect: a})
	}
	for _, mode := range []Mode{LargeLeft, LargeRight, LargeCenter, Grid} {
		p := Compute(mode, Input{Zone: zone169, Main: m, Secondary: sec}, Options{})
		for i, pl := range p.Placements {
			if got := pl.Content.Aspect(); math.Abs(got-asp[i]) > 1e-6 {
				t.Errorf("%v placement %d aspect %.4f want %.4f", mode, i, got, asp[i])
			}
		}
	}
}

func TestMainIsLargestInLargeModes(t *testing.T) {
	for _, mode := range []Mode{LargeLeft, LargeRight, LargeCenter} {
		for n := 2; n <= 10; n++ {
			main, sec := items(n, 16.0/9)
			p := Compute(mode, Input{Zone: zone169, Main: main, Secondary: sec}, Options{})
			ma := p.Placements[0].Content.Area()
			for _, s := range p.Placements[1:] {
				if s.Content.Area() >= ma {
					t.Errorf("%v n=%d: secondary %s (%.0f) not smaller than main (%.0f)", mode, n, s.ID, s.Content.Area(), ma)
				}
			}
		}
	}
}

func TestSidePlacement(t *testing.T) {
	main, sec := items(3, 16.0/9)
	in := Input{Zone: zone169, Main: main, Secondary: sec}
	l := Compute(LargeLeft, in, Options{})
	r := Compute(LargeRight, in, Options{})
	if !(l.Placements[0].Cell.X < l.Placements[1].Cell.X) {
		t.Error("large-left: main should be left of the strip")
	}
	if !(r.Placements[0].Cell.X > r.Placements[1].Cell.X) {
		t.Error("large-right: main should be right of the strip")
	}
	c := Compute(LargeCenter, Input{Zone: zone169, Main: main, Secondary: append(sec, Item{ID: "s3", Aspect: 16.0 / 9})}, Options{})
	mx := c.Placements[0].Cell
	if !(c.Placements[1].Cell.X < mx.X && c.Placements[2].Cell.X > mx.X) {
		t.Error("large-center: first secondary goes left, second right")
	}
}

func TestStripNeverSliver(t *testing.T) {
	// Main wider than the zone would like to take everything; the strip keeps MinStripFrac.
	main := &Item{ID: "m", Aspect: 32.0 / 9}
	p := Compute(LargeLeft, Input{Zone: zone169, Main: main, Secondary: []Item{{ID: "a", Aspect: 16.0 / 9}}}, Options{})
	if w := p.Placements[1].Cell.W; w < 0.17*zone169.W {
		t.Errorf("strip width %.0f below floor", w)
	}
	// Portrait main still gets at least half the width.
	main = &Item{ID: "m", Aspect: 9.0 / 16}
	p = Compute(LargeLeft, Input{Zone: zone169, Main: main, Secondary: []Item{{ID: "a", Aspect: 16.0 / 9}}}, Options{})
	if w := p.Placements[0].Cell.W; w < 0.45*zone169.W {
		t.Errorf("main width %.0f below half", w)
	}
}

func TestGridCentresShortLastRow(t *testing.T) {
	cells := BestGrid(Rect{W: 1920, H: 1080}, 20, []float64{16.0 / 9, 16.0 / 9, 16.0 / 9})
	// 3 x 16:9 in 16:9 -> 2 columns, 2 rows; the lone third tile is centred.
	if len(cells) != 3 {
		t.Fatal("want 3 cells")
	}
	cx, _ := cells[2].Center()
	if math.Abs(cx-960) > 1e-6 {
		t.Errorf("last row not centred: centre x=%.1f", cx)
	}
}

func TestGridPicksShapeFromAspect(t *testing.T) {
	// Four 16:9 sources in 16:9 -> 2x2 (like a T1V quad).
	cells := BestGrid(Rect{W: 1920, H: 1080}, 0, []float64{1.78, 1.78, 1.78, 1.78})
	if cells[1].Y != cells[0].Y || cells[2].Y == cells[0].Y {
		t.Error("expected 2x2 for four 16:9 sources")
	}
	// Three portrait phones on a landscape wall -> one row of three.
	cells = BestGrid(Rect{W: 1920, H: 1080}, 0, []float64{0.5625, 0.5625, 0.5625})
	if cells[0].Y != cells[2].Y {
		t.Error("expected a single row for three portrait sources")
	}
}

func TestAutoDecisions(t *testing.T) {
	main, sec := items(3, 16.0/9)
	if got := Compute(Auto, Input{Zone: zone169, Main: main, Secondary: sec}, Options{}).Resolved; got != Grid {
		t.Errorf("no emphasis: got %v want grid", got)
	}
	if got := Compute(Auto, Input{Zone: zone169, Main: main, Secondary: sec, Emphasis: true}, Options{}).Resolved; got != LargeLeft {
		t.Errorf("emphasis on 16:9: got %v want large-left", got)
	}
	wall := Rect{W: 5760, H: 1080}
	main, sec = items(5, 16.0/9)
	if got := Compute(Auto, Input{Zone: wall, Main: main, Secondary: sec, Emphasis: true}, Options{}).Resolved; got != LargeCenter {
		t.Errorf("wide wall, 4 secondaries: got %v want large-center", got)
	}
	// Too many secondaries for a legible strip: back to the grid.
	main, sec = items(30, 16.0/9)
	if got := Compute(Auto, Input{Zone: zone169, Main: main, Secondary: sec, Emphasis: true}, Options{}).Resolved; got != Grid {
		t.Errorf("30 sources: got %v want grid", got)
	}
}

func TestStableOrderOnRemove(t *testing.T) {
	// Removing one secondary must not reorder the rest.
	main, sec := items(5, 16.0/9)
	before := Compute(LargeLeft, Input{Zone: zone169, Main: main, Secondary: sec}, Options{})
	after := Compute(LargeLeft, Input{Zone: zone169, Main: main, Secondary: append([]Item{}, sec[1:]...)}, Options{})
	for i, p := range after.Placements[1:] {
		if p.ID != before.Placements[i+2].ID {
			t.Errorf("order changed after removal: %s", p.ID)
		}
	}
}

func TestDeterministic(t *testing.T) {
	main, sec := items(7, 4.0/3)
	in := Input{Zone: zone169, Main: main, Secondary: sec, Emphasis: true}
	a := Compute(Auto, in, Options{})
	b := Compute(Auto, in, Options{})
	for i := range a.Placements {
		if a.Placements[i] != b.Placements[i] {
			t.Fatal("non-deterministic plan")
		}
	}
}

func TestParseMode(t *testing.T) {
	cases := map[string]Mode{
		"Large Left": LargeLeft, "large-right": LargeRight, "LARGE_CENTRE": LargeCenter,
		"center": LargeCenter, "even grid": Grid, "Grid": Grid, "auto": Auto, " left ": LargeLeft,
	}
	for in, want := range cases {
		got, ok := ParseMode(in)
		if !ok || got != want {
			t.Errorf("ParseMode(%q) = %v,%v want %v", in, got, ok, want)
		}
	}
	if _, ok := ParseMode("banana"); ok {
		t.Error("unknown mode accepted")
	}
}

func TestFitAndLerp(t *testing.T) {
	f := Fit(1, Rect{W: 200, H: 100})
	if f != (Rect{X: 50, Y: 0, W: 100, H: 100}) {
		t.Errorf("Fit square in 2:1: %v", f)
	}
	a, b := Rect{0, 0, 10, 10}, Rect{100, 100, 20, 20}
	if Lerp(a, b, 0) != a || Lerp(a, b, 1) != b {
		t.Error("Lerp endpoints")
	}
	if m := Lerp(a, b, 0.5); math.Abs(m.X-50) > eps {
		t.Errorf("Lerp midpoint %v", m)
	}
}
