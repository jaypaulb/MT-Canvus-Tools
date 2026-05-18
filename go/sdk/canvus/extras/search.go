// Phase 4b §4.1 #15: cross-canvas search port from
// CanvusPythonAPI/canvus_api/search.py:16-577.
//
// SearchResult expands the core canvus.WidgetMatch with MatchScore,
// MatchReason, and a drill-down path (canvasID:widgetID). The search
// engine pulls widgets per canvas via a small interface so callers can
// pass either a live *canvus.Session or a mock for testing.
package extras

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// SearchResult is one hit from a cross-canvas search.
type SearchResult struct {
	CanvasID    string
	CanvasName  string
	WidgetID    string
	WidgetType  string
	Widget      canvus.Widget
	MatchScore  float64
	MatchReason string
}

// DrillDownPath returns "canvasID:widgetID" for callers wanting a single string key.
func (r SearchResult) DrillDownPath() string {
	return fmt.Sprintf("%s:%s", r.CanvasID, r.WidgetID)
}

// CanvasLister is the minimum interface CrossCanvasSearch needs from a Canvus
// session. *canvus.Session satisfies this; mocks can stub it for tests.
type CanvasLister interface {
	ListCanvases(ctx context.Context, filter *canvus.Filter) ([]canvus.Canvas, error)
	GetCanvas(ctx context.Context, id string) (*canvus.Canvas, error)
	ListWidgets(ctx context.Context, canvasID string, filter *canvus.Filter, includeAnnotations ...bool) ([]canvus.Widget, error)
}

// CrossCanvasSearch runs structured queries across one or many canvases.
type CrossCanvasSearch struct {
	Lister CanvasLister
}

// NewCrossCanvasSearch returns a search engine bound to the given lister.
func NewCrossCanvasSearch(l CanvasLister) *CrossCanvasSearch {
	return &CrossCanvasSearch{Lister: l}
}

// SearchOptions tune a FindWidgetsAcrossCanvases call.
type SearchOptions struct {
	CanvasIDs      []string          // limit to specific canvases (default: all)
	WidgetTypes    []string          // limit to widget types (case-insensitive)
	SpatialFilter  *canvus.Rectangle // restrict to widgets intersecting this area
	MaxResults     int               // upper bound on results (default 100)
	IncludeDeleted bool              // include widgets with state == "deleted"
}

// FindWidgetsAcrossCanvases runs a filter-criteria query across the
// requested canvases. query may be a map[string]any of field criteria or a
// plain string (treated as wildcard text search on "text").
func (s *CrossCanvasSearch) FindWidgetsAcrossCanvases(ctx context.Context, query any, opts SearchOptions) ([]SearchResult, error) {
	if opts.MaxResults <= 0 {
		opts.MaxResults = 100
	}
	criteria, err := parseQuery(query)
	if err != nil {
		return nil, err
	}
	canvases, err := s.canvasesToSearch(ctx, opts.CanvasIDs)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, c := range canvases {
		widgets, err := s.Lister.ListWidgets(ctx, c.ID, nil)
		if err != nil {
			// Skip canvases we cannot read; mirror Python behaviour.
			continue
		}
		filtered := applyFilters(widgets, criteria, opts)
		for _, w := range filtered {
			r := SearchResult{
				CanvasID:    c.ID,
				CanvasName:  c.Name,
				WidgetID:    w.ID,
				WidgetType:  w.WidgetType,
				Widget:      w,
				MatchScore:  scoreMatch(w, criteria),
				MatchReason: matchReason(w, criteria),
			}
			results = append(results, r)
			if len(results) >= opts.MaxResults {
				break
			}
		}
		if len(results) >= opts.MaxResults {
			break
		}
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].MatchScore > results[j].MatchScore
	})
	if len(results) > opts.MaxResults {
		results = results[:opts.MaxResults]
	}
	return results, nil
}

// FindWidgetsByText is a thin convenience that wraps the query as a wildcard
// substring search on the "text" field.
func (s *CrossCanvasSearch) FindWidgetsByText(ctx context.Context, text string, opts SearchOptions) ([]SearchResult, error) {
	q := map[string]any{"text": "*" + text + "*"}
	return s.FindWidgetsAcrossCanvases(ctx, q, opts)
}

// FindWidgetsByType wraps the query as a widget_type equality match.
func (s *CrossCanvasSearch) FindWidgetsByType(ctx context.Context, widgetType string, opts SearchOptions) ([]SearchResult, error) {
	q := map[string]any{"widget_type": widgetType}
	return s.FindWidgetsAcrossCanvases(ctx, q, opts)
}

// FindWidgetsInArea finds every widget intersecting the given area across
// canvases. The area is applied via opts.SpatialFilter; query is empty.
func (s *CrossCanvasSearch) FindWidgetsInArea(ctx context.Context, area canvus.Rectangle, opts SearchOptions) ([]SearchResult, error) {
	opts.SpatialFilter = &area
	return s.FindWidgetsAcrossCanvases(ctx, map[string]any{}, opts)
}

// FindWidgetsByProperty finds widgets whose `propertyPath` (dot notation)
// equals the given value.
func (s *CrossCanvasSearch) FindWidgetsByProperty(ctx context.Context, propertyPath string, value any, opts SearchOptions) ([]SearchResult, error) {
	q := map[string]any{propertyPath: value}
	return s.FindWidgetsAcrossCanvases(ctx, q, opts)
}

// --- internals ---

func parseQuery(query any) (map[string]any, error) {
	switch q := query.(type) {
	case nil:
		return map[string]any{}, nil
	case map[string]any:
		return q, nil
	case string:
		// Try JSON first; fall back to wildcard substring text search.
		var m map[string]any
		if err := json.Unmarshal([]byte(q), &m); err == nil {
			return m, nil
		}
		return map[string]any{"text": "*" + q + "*"}, nil
	}
	return nil, fmt.Errorf("parseQuery: unsupported query type %T", query)
}

func (s *CrossCanvasSearch) canvasesToSearch(ctx context.Context, ids []string) ([]canvus.Canvas, error) {
	if len(ids) == 0 {
		return s.Lister.ListCanvases(ctx, nil)
	}
	out := make([]canvus.Canvas, 0, len(ids))
	for _, id := range ids {
		c, err := s.Lister.GetCanvas(ctx, id)
		if err != nil {
			continue
		}
		out = append(out, *c)
	}
	return out, nil
}

func applyFilters(widgets []canvus.Widget, criteria map[string]any, opts SearchOptions) []canvus.Widget {
	out := make([]canvus.Widget, 0, len(widgets))
	allowedTypes := make(map[string]struct{}, len(opts.WidgetTypes))
	for _, t := range opts.WidgetTypes {
		allowedTypes[strings.ToLower(t)] = struct{}{}
	}
	for _, w := range widgets {
		if !opts.IncludeDeleted && w.State == "deleted" {
			continue
		}
		if len(allowedTypes) > 0 {
			if _, ok := allowedTypes[strings.ToLower(w.WidgetType)]; !ok {
				continue
			}
		}
		if opts.SpatialFilter != nil {
			if !Intersects(canvus.WidgetBoundingBox(w), *opts.SpatialFilter) {
				continue
			}
		}
		if len(criteria) > 0 && !matchesCriteria(w, criteria) {
			continue
		}
		out = append(out, w)
	}
	return out
}

func matchesCriteria(w canvus.Widget, criteria map[string]any) bool {
	wm := w.AsMap()
	for k, v := range criteria {
		var actual any
		if strings.Contains(k, ".") {
			actual = getNested(wm, k)
		} else {
			actual = wm[k]
		}
		if actual == nil {
			return false
		}
		if sv, isStr := v.(string); isStr && strings.Contains(sv, "*") {
			if !wildcardMatch(actual, sv) {
				return false
			}
			continue
		}
		if k == "widget_type" {
			if !strings.EqualFold(fmt.Sprintf("%v", actual), fmt.Sprintf("%v", v)) {
				return false
			}
			continue
		}
		if !equal(actual, v) {
			return false
		}
	}
	return true
}

func scoreMatch(w canvus.Widget, criteria map[string]any) float64 {
	if len(criteria) == 0 {
		return 1.0
	}
	if matchesCriteria(w, criteria) {
		return 1.0
	}
	return 0.0
}

func matchReason(w canvus.Widget, criteria map[string]any) string {
	if len(criteria) == 0 {
		return "No filter applied"
	}
	wm := w.AsMap()
	for k, v := range criteria {
		actual := wm[k]
		if actual == nil {
			continue
		}
		if fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", v) {
			return "Exact match on " + k
		}
		if strings.Contains(fmt.Sprintf("%v", actual), fmt.Sprintf("%v", v)) {
			return "Partial match on " + k
		}
	}
	return "Filter criteria matched"
}
