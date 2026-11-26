// Command autolayout turns named anchors on a Canvus canvas into live
// auto-layout zones, and previews layouts offline.
//
//	autolayout run     --canvas <id>        watch a canvas and arrange its zones
//	autolayout preview --mode large-left    print a plan as JSON (no server)
//
// A zone is any anchor whose name starts with "Auto Layout", optionally
// followed by a mode: "Auto Layout: large left", "Auto Layout: grid".
// Rename the anchor in Canvus to switch modes.
//
// run reads CANVUS_API_URL and CANVUS_API_KEY (see canvus.FromEnv).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/autolayout/canvusx"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/autolayout/layout"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/autolayout/zone"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "run":
		err = runCmd(os.Args[2:])
	case "preview":
		err = previewCmd(os.Args[2:], os.Stdout)
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		slog.Error("autolayout", "err", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: autolayout run --canvas <id> [flags] | autolayout preview --mode <mode> [flags]")
}

// ---- preview ----

func parseAspect(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if a, b, ok := strings.Cut(s, ":"); ok {
		x, err1 := strconv.ParseFloat(a, 64)
		y, err2 := strconv.ParseFloat(b, 64)
		if err1 != nil || err2 != nil || y == 0 {
			return 0, fmt.Errorf("bad aspect %q", s)
		}
		return x / y, nil
	}
	return strconv.ParseFloat(s, 64)
}

func parseSize(s string) (float64, float64, error) {
	a, b, ok := strings.Cut(strings.ToLower(s), "x")
	if !ok {
		return 0, 0, fmt.Errorf("bad size %q, want WxH", s)
	}
	w, err1 := strconv.ParseFloat(a, 64)
	h, err2 := strconv.ParseFloat(b, 64)
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("bad size %q", s)
	}
	return w, h, nil
}

type previewOut struct {
	Requested  string        `json:"requested"`
	Resolved   string        `json:"resolved"`
	Coverage   float64       `json:"coverage"`
	Zone       layout.Rect   `json:"zone"`
	Placements []previewTile `json:"placements"`
}

type previewTile struct {
	ID      string      `json:"id"`
	Role    string      `json:"role"`
	Cell    layout.Rect `json:"cell"`
	Content layout.Rect `json:"content"`
}

func previewCmd(args []string, out *os.File) error {
	fs := flag.NewFlagSet("preview", flag.ContinueOnError)
	modeS := fs.String("mode", "auto", "auto | large-left | large-right | large-center | grid")
	zoneS := fs.String("zone", "1920x1080", "zone size WxH")
	aspS := fs.String("aspects", "16:9,16:9,16:9", "comma-separated source aspects, main first (16:9, 4:3, 0.5625 ...)")
	emph := fs.Bool("emphasis", false, "main source is promoted (affects auto)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	mode, ok := layout.ParseMode(*modeS)
	if !ok {
		return fmt.Errorf("unknown mode %q", *modeS)
	}
	zw, zh, err := parseSize(*zoneS)
	if err != nil {
		return err
	}
	in := layout.Input{Zone: layout.Rect{W: zw, H: zh}, Emphasis: *emph}
	for i, a := range strings.Split(*aspS, ",") {
		if strings.TrimSpace(a) == "" {
			continue
		}
		v, err := parseAspect(a)
		if err != nil {
			return err
		}
		it := layout.Item{ID: fmt.Sprintf("s%d", i), Aspect: v}
		if in.Main == nil {
			it.ID = "main"
			in.Main = &it
		} else {
			in.Secondary = append(in.Secondary, it)
		}
	}
	p := layout.Compute(mode, in, layout.Options{})
	res := previewOut{Requested: p.Requested.String(), Resolved: p.Resolved.String(), Coverage: p.Coverage, Zone: in.Zone}
	for _, pl := range p.Placements {
		res.Placements = append(res.Placements, previewTile{ID: pl.ID, Role: pl.Role.String(), Cell: pl.Cell, Content: pl.Content})
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}

// ---- run ----

func runCmd(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	canvasID := fs.String("canvas", os.Getenv("CANVUS_CANVAS_ID"), "canvas ID (or CANVUS_CANVAS_ID)")
	interval := fs.Duration("interval", 400*time.Millisecond, "poll interval")
	settle := fs.Duration("settle", 600*time.Millisecond, "how long a widget must be still before it is arranged")
	notePad := fs.Float64("note-padding", 30, "Note registration padding of the deployed server (30 legacy, 0 normalised)")
	dry := fs.Bool("dry-run", false, "log moves without sending them")
	once := fs.Bool("once", false, "arrange once and exit")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *canvasID == "" {
		return errors.New("--canvas is required")
	}
	s, err := canvus.FromEnv()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	r := &runner{s: s, canvas: *canvasID, notePad: *notePad, dry: *dry,
		cfg: zone.Config{Settle: *settle}, zones: map[string]*zone.Controller{}}
	tick := time.NewTicker(*interval)
	defer tick.Stop()
	start := time.Now()
	for {
		if err := r.step(ctx, time.Now()); err != nil {
			slog.Warn("step failed", "err", err)
		}
		if *once && time.Since(start) > *settle+*interval {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
		}
	}
}

type runner struct {
	s       *canvus.Session
	canvas  string
	notePad float64
	dry     bool
	cfg     zone.Config
	zones   map[string]*zone.Controller
}

func (r *runner) step(ctx context.Context, now time.Time) error {
	widgets, err := r.s.ListWidgets(ctx, r.canvas, nil)
	if err != nil {
		return fmt.Errorf("list widgets: %w", err)
	}
	anchors, err := r.s.ListAnchors(ctx, r.canvas)
	if err != nil {
		return fmt.Errorf("list anchors: %w", err)
	}
	names := make(map[string]string, len(anchors))
	for _, a := range anchors {
		names[a.ID] = a.AnchorName
	}
	snap, err := canvusx.Build(widgets, names, r.notePad)
	if err != nil {
		return err
	}

	// A widget belongs to at most one zone: the one that already holds it,
	// otherwise whichever zone sees it first.
	owner := map[string]string{}
	live := map[string]bool{}
	for _, z := range snap.Zones {
		live[z.ID] = true
		if c, ok := r.zones[z.ID]; ok {
			for _, id := range c.Members() {
				owner[id] = z.ID
			}
		}
	}
	for id := range r.zones {
		if !live[id] {
			slog.Info("zone removed", "zone", id)
			delete(r.zones, id)
		}
	}
	for _, z := range snap.Zones {
		c, ok := r.zones[z.ID]
		if !ok {
			c = zone.New(r.cfg)
			r.zones[z.ID] = c
			slog.Info("zone found", "zone", z.ID, "mode", z.Mode.String())
		}
		var ms []zone.Member
		for _, m := range snap.Members {
			if o, taken := owner[m.ID]; !taken || o == z.ID {
				ms = append(ms, m)
			}
		}
		for _, mv := range c.Step(now, z, ms) {
			owner[mv.ID] = z.ID
			w := snap.ByID[mv.ID]
			body, err := canvusx.Patch(w, mv.To, r.notePad)
			if err != nil {
				slog.Warn("skip", "widget", mv.ID, "err", err)
				continue
			}
			slog.Info("move", "zone", z.ID, "widget", mv.ID, "type", mv.Type, "mode", c.Plan().Resolved.String(), "to", mv.To)
			if r.dry {
				continue
			}
			if err := canvusx.Apply(ctx, r.s, r.canvas, w, body); err != nil {
				slog.Warn("move failed", "widget", mv.ID, "err", err)
			}
		}
		for _, id := range c.Members() {
			owner[id] = z.ID
		}
	}
	return nil
}
