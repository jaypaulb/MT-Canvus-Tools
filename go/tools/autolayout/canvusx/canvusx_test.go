package canvusx

import (
	"math"
	"testing"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/autolayout/layout"
)

func w(id, typ, parent string, x, y, wd, h, scale float64) canvus.Widget {
	return canvus.Widget{ID: id, WidgetType: typ, ParentID: parent,
		Location: &canvus.Point{X: x, Y: y}, Size: &canvus.Size{Width: wd, Height: h}, Scale: scale}
}

func snapshot() []canvus.Widget {
	root := canvus.Widget{ID: "root", WidgetType: "SharedCanvas", Location: &canvus.Point{}, Size: &canvus.Size{Width: 1, Height: 1}, Scale: 1}
	pinned := w("pin", "Image", "root", 10, 10, 100, 100, 1)
	pinned.Pinned = true
	return []canvus.Widget{
		root,
		w("z", "Anchor", "root", 0, 0, 1920, 1080, 1),
		w("other-anchor", "Anchor", "root", 5000, 0, 100, 100, 1),
		w("img", "Image", "root", 100, 100, 1920, 1080, 0.25),
		w("note", "Note", "root", 600, 100, 300, 300, 1),
		w("share", "VideoInput", "root", 900, 600, 1280, 720, 0.5),
		w("child", "Note", "img", 5, 5, 50, 50, 1),
		pinned,
	}
}

func TestBuild(t *testing.T) {
	snap, err := Build(snapshot(), map[string]string{"z": "Auto Layout: large left", "other-anchor": "Lobby"}, 30)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Zones) != 1 || snap.Zones[0].ID != "z" || snap.Zones[0].Mode != layout.LargeLeft {
		t.Fatalf("zones %+v", snap.Zones)
	}
	got := map[string]bool{}
	for _, m := range snap.Members {
		got[m.ID] = true
		if m.ID == "share" && !m.Live {
			t.Error("video input should be live")
		}
		if m.ID == "img" && (m.Rect.W != 480 || m.Rect.H != 270) {
			t.Errorf("img rendered rect %v", m.Rect)
		}
	}
	if !got["img"] || !got["note"] || !got["share"] || got["child"] || got["pin"] || got["z"] {
		t.Fatalf("members %v", got)
	}
}

func TestPatchRoundTrip(t *testing.T) {
	widgets := snapshot()
	pad := 30.0
	model := canvus.GeometryModel{RootWidgetID: "root", NotePadding: &pad}
	to := layout.Rect{X: 250, Y: 400, W: 600, H: 600}
	for _, idx := range []int{3, 4} { // image and note
		target := widgets[idx]
		body, err := Patch(target, to, pad)
		if err != nil {
			t.Fatal(err)
		}
		loc := body["location"].(map[string]float64)
		moved := target
		moved.Location = &canvus.Point{X: loc["x"], Y: loc["y"]}
		moved.Scale = body["scale"].(float64)
		ws := append([]canvus.Widget{}, widgets...)
		ws[idx] = moved
		r, err := canvus.WidgetCanvasBounds(target.ID, ws, model)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(r.X-to.X) > 1e-9 || math.Abs(r.Y-to.Y) > 1e-9 || r.Width > to.W+1e-9 || r.Height > to.H+1e-9 {
			t.Errorf("%s: rendered %v, wanted inside %v at its origin", target.ID, r, to)
		}
	}
}
