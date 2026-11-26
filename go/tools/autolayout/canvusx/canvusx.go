// Package canvusx adapts Canvus REST snapshots to the zone controller and
// turns its moves back into widget PATCH bodies.
//
// Only widgets that sit directly on the canvas (parent = SharedCanvas root)
// are arranged; children of other widgets keep their parent's layout. Pinned
// widgets are left alone: a pin is a person saying "don't move this".
package canvusx

import (
	"context"
	"fmt"
	"strings"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/autolayout/layout"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/autolayout/zone"
)

// Snapshot is one observation of a canvas.
type Snapshot struct {
	Root    string
	Zones   []zone.Spec
	Members []zone.Member
	ByID    map[string]canvus.Widget
}

func toRect(r canvus.Rectangle) layout.Rect {
	return layout.Rect{X: r.X, Y: r.Y, W: r.Width, H: r.Height}
}

// Build classifies widgets into zones (named anchors) and candidate members.
// anchorNames maps anchor widget ID to its name (from ListAnchors).
func Build(widgets []canvus.Widget, anchorNames map[string]string, notePadding float64) (Snapshot, error) {
	snap := Snapshot{ByID: map[string]canvus.Widget{}}
	for _, w := range widgets {
		snap.ByID[w.ID] = w
		if strings.EqualFold(w.WidgetType, "SharedCanvas") {
			snap.Root = w.ID
		}
	}
	if snap.Root == "" {
		return snap, fmt.Errorf("canvas snapshot has no SharedCanvas root")
	}
	model := canvus.GeometryModel{RootWidgetID: snap.Root, NotePadding: &notePadding}
	for _, w := range widgets {
		if w.ID == snap.Root || w.State == "deleted" {
			continue
		}
		if strings.EqualFold(w.WidgetType, "Anchor") {
			mode, ok := zone.ParseAnchorName(anchorNames[w.ID])
			if !ok {
				continue
			}
			r, err := canvus.WidgetCanvasBounds(w.ID, widgets, model)
			if err != nil {
				continue
			}
			snap.Zones = append(snap.Zones, zone.Spec{ID: w.ID, Rect: toRect(r), Mode: mode})
			continue
		}
		if !zone.IsArrangeable(w.WidgetType) || w.Pinned || w.ParentID != snap.Root {
			continue
		}
		r, err := canvus.WidgetCanvasBounds(w.ID, widgets, model)
		if err != nil {
			continue
		}
		snap.Members = append(snap.Members, zone.Member{
			ID: w.ID, Type: w.WidgetType, Rect: toRect(r), Live: zone.IsLiveType(w.WidgetType),
		})
	}
	return snap, nil
}

// Patch returns the PATCH body that places a root-level widget's rendered
// border box at to. Size is kept (images and videos keep their native pixel
// size); the visual size changes through scale, which Canvus applies
// uniformly. Notes carry registration padding, so their stored location is
// offset by scale*pad - pad (see canvus.WidgetCanvasBounds).
func Patch(w canvus.Widget, to layout.Rect, notePadding float64) (map[string]any, error) {
	if w.Size == nil || w.Size.Width <= 0 || w.Size.Height <= 0 {
		return nil, fmt.Errorf("widget %s has no size", w.ID)
	}
	scale := to.W / w.Size.Width
	if sy := to.H / w.Size.Height; sy < scale {
		scale = sy // never overflow the target, even if the stored aspect drifted
	}
	pad := 0.0
	if strings.EqualFold(w.WidgetType, "Note") {
		pad = notePadding
	}
	return map[string]any{
		"location": map[string]float64{"x": to.X + scale*pad - pad, "y": to.Y + scale*pad - pad},
		"scale":    scale,
	}, nil
}

// Updater is the subset of the SDK session the runner needs to move widgets.
type Updater interface {
	UpdateNote(ctx context.Context, canvasID, id string, req any) (*canvus.Note, error)
	UpdateImage(ctx context.Context, canvasID, id string, req any) (*canvus.Image, error)
	UpdateBrowser(ctx context.Context, canvasID, id string, req any) (*canvus.Browser, error)
	UpdatePDF(ctx context.Context, canvasID, id string, req any) (*canvus.PDF, error)
	UpdateVideo(ctx context.Context, canvasID, id string, req any) (*canvus.Video, error)
	UpdateTable(ctx context.Context, canvasID, id string, req any) (*canvus.Table, error)
	UpdateIPVideo(ctx context.Context, canvasID, id string, req any) (*canvus.IPVideo, error)
	UpdateRDPConnection(ctx context.Context, canvasID, id string, req any) (*canvus.RDPConnection, error)
	UpdateVideoInput(ctx context.Context, canvasID, id string, req map[string]any) (*canvus.VideoInput, error)
}

// Apply sends one PATCH to the type-specific endpoint.
func Apply(ctx context.Context, u Updater, canvasID string, w canvus.Widget, body map[string]any) error {
	var err error
	switch strings.NewReplacer("_", "", "-", "").Replace(strings.ToLower(w.WidgetType)) {
	case "note":
		_, err = u.UpdateNote(ctx, canvasID, w.ID, body)
	case "image":
		_, err = u.UpdateImage(ctx, canvasID, w.ID, body)
	case "browser":
		_, err = u.UpdateBrowser(ctx, canvasID, w.ID, body)
	case "pdf":
		_, err = u.UpdatePDF(ctx, canvasID, w.ID, body)
	case "video":
		_, err = u.UpdateVideo(ctx, canvasID, w.ID, body)
	case "table":
		_, err = u.UpdateTable(ctx, canvasID, w.ID, body)
	case "ipvideo":
		_, err = u.UpdateIPVideo(ctx, canvasID, w.ID, body)
	case "rdpconnection":
		_, err = u.UpdateRDPConnection(ctx, canvasID, w.ID, body)
	case "videoinput":
		_, err = u.UpdateVideoInput(ctx, canvasID, w.ID, body)
	default:
		return fmt.Errorf("widget type %q cannot be moved through the REST API", w.WidgetType)
	}
	return err
}
