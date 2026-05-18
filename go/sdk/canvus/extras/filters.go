// Phase 4b §4.1 #12: filters port from CanvusPythonAPI/canvus_api/filters.py.
//
// This is the rich, structured filter — distinct from the simple wildcard
// canvus.Filter that already lives in the core SDK. Use this when you need
// multiple conditions, spatial operators, dot-path field access, or want
// round-trip JSON serialisation.
package extras

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// FilterOperator enumerates the supported filter operators.
type FilterOperator string

const (
	OpEquals            FilterOperator = "equals"
	OpNotEquals         FilterOperator = "not_equals"
	OpContains          FilterOperator = "contains"
	OpNotContains       FilterOperator = "not_contains"
	OpStartsWith        FilterOperator = "starts_with"
	OpEndsWith          FilterOperator = "ends_with"
	OpGreaterThan       FilterOperator = "greater_than"
	OpLessThan          FilterOperator = "less_than"
	OpGreaterEqual      FilterOperator = "greater_equal"
	OpLessEqual         FilterOperator = "less_equal"
	OpIn                FilterOperator = "in"
	OpNotIn             FilterOperator = "not_in"
	OpExists            FilterOperator = "exists"
	OpNotExists         FilterOperator = "not_exists"
	OpSpatialIntersects FilterOperator = "spatial_intersects"
	OpSpatialContains   FilterOperator = "spatial_contains"
	OpSpatialWithin     FilterOperator = "spatial_within"
	OpWildcardMatch     FilterOperator = "wildcard_match"
)

// FilterCondition is one term in a Filter. Field uses dot-notation for
// nested map access (e.g. "location.x"). Field == "spatial" is reserved
// for the spatial_* operators; in that case Value must be canvus.Rectangle.
type FilterCondition struct {
	Field    string         `json:"field"`
	Operator FilterOperator `json:"operator"`
	Value    any            `json:"value"`
}

// Filter is a chainable AND-only filter. To express OR, build two filters
// and check Matches on each (mirroring the legacy Python behaviour — see
// combine_filters: AND-only in practice).
type Filter struct {
	Conditions []FilterCondition `json:"conditions"`
}

// NewFilter returns an empty filter.
func NewFilter() *Filter { return &Filter{} }

// AddCondition appends an arbitrary condition.
func (f *Filter) AddCondition(field string, op FilterOperator, value any) *Filter {
	f.Conditions = append(f.Conditions, FilterCondition{Field: field, Operator: op, Value: value})
	return f
}

// AddSpatialCondition appends a spatial condition. operator must be one of
// "intersects", "contains", or "within"; it is prefixed with "spatial_".
func (f *Filter) AddSpatialCondition(operator string, area canvus.Rectangle) *Filter {
	f.Conditions = append(f.Conditions, FilterCondition{
		Field:    "spatial",
		Operator: FilterOperator("spatial_" + operator),
		Value:    area,
	})
	return f
}

// AddWildcardCondition appends a wildcard pattern match on a single field.
func (f *Filter) AddWildcardCondition(field, pattern string) *Filter {
	f.Conditions = append(f.Conditions, FilterCondition{
		Field:    field,
		Operator: OpWildcardMatch,
		Value:    pattern,
	})
	return f
}

// Matches returns true if every condition matches the given item.
func (f *Filter) Matches(item map[string]any) bool {
	for _, c := range f.Conditions {
		if !matchesCondition(item, c) {
			return false
		}
	}
	return true
}

// ToMap renders the filter as a serialisable map[string]any (same shape as
// the Python implementation's to_dict).
func (f *Filter) ToMap() map[string]any {
	conds := make([]map[string]any, 0, len(f.Conditions))
	for _, c := range f.Conditions {
		conds = append(conds, map[string]any{
			"field":    c.Field,
			"operator": string(c.Operator),
			"value":    c.Value,
		})
	}
	return map[string]any{"conditions": conds}
}

// FromMap parses a filter previously serialised via ToMap (or by the
// equivalent Python/TS helpers).
func FromMap(data map[string]any) (*Filter, error) {
	raw, ok := data["conditions"].([]any)
	if !ok {
		// Empty filter is fine.
		return NewFilter(), nil
	}
	f := NewFilter()
	for i, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("FromMap: condition %d is not an object", i)
		}
		field, _ := m["field"].(string)
		op, _ := m["operator"].(string)
		f.Conditions = append(f.Conditions, FilterCondition{
			Field:    field,
			Operator: FilterOperator(op),
			Value:    m["value"],
		})
	}
	return f, nil
}

// NewSpatialFilter builds a filter containing only the given spatial
// condition. operator: "intersects" | "contains" | "within".
func NewSpatialFilter(area canvus.Rectangle, operator string) *Filter {
	return NewFilter().AddSpatialCondition(operator, area)
}

// NewWidgetTypeFilter matches any widget whose "type" or "widget_type"
// field is in the given list (case-insensitive).
func NewWidgetTypeFilter(widgetTypes ...string) *Filter {
	// Lower-case the values up-front for cheap comparison.
	lc := make([]any, len(widgetTypes))
	for i, t := range widgetTypes {
		lc[i] = strings.ToLower(t)
	}
	return NewFilter().AddCondition("widget_type", OpIn, lc)
}

// NewTextFilter creates a filter that matches if any of the listed fields
// contains text. Default fields: "title", "text", "description".
func NewTextFilter(text string, fields ...string) *Filter {
	if len(fields) == 0 {
		fields = []string{"title", "text", "description"}
	}
	f := NewFilter()
	for _, fld := range fields {
		f.AddCondition(fld, OpContains, text)
	}
	return f
}

// NewWildcardFilter creates a filter that matches a wildcard pattern on a
// single field. Defaults to "title".
func NewWildcardFilter(pattern string, field ...string) *Filter {
	fld := "title"
	if len(field) > 0 && field[0] != "" {
		fld = field[0]
	}
	return NewFilter().AddWildcardCondition(fld, pattern)
}

// CombineFilters AND-combines multiple filters into a single filter by
// concatenating their conditions. Mirrors the legacy Python behaviour
// (OR is not supported; build per-filter and union match results instead).
func CombineFilters(filters ...*Filter) *Filter {
	out := NewFilter()
	for _, f := range filters {
		if f == nil {
			continue
		}
		out.Conditions = append(out.Conditions, f.Conditions...)
	}
	return out
}

// --- internals ---

func matchesCondition(item map[string]any, c FilterCondition) bool {
	if c.Field == "spatial" && strings.HasPrefix(string(c.Operator), "spatial_") {
		return matchesSpatial(item, c.Operator, c.Value)
	}
	v := getNested(item, c.Field)
	switch c.Operator {
	case OpEquals:
		return equal(v, c.Value)
	case OpNotEquals:
		return !equal(v, c.Value)
	case OpContains:
		return containsAny(v, c.Value)
	case OpNotContains:
		return !containsAny(v, c.Value)
	case OpStartsWith:
		return v != nil && strings.HasPrefix(fmt.Sprintf("%v", v), fmt.Sprintf("%v", c.Value))
	case OpEndsWith:
		return v != nil && strings.HasSuffix(fmt.Sprintf("%v", v), fmt.Sprintf("%v", c.Value))
	case OpGreaterThan:
		return compareNum(v, c.Value, func(a, b float64) bool { return a > b })
	case OpLessThan:
		return compareNum(v, c.Value, func(a, b float64) bool { return a < b })
	case OpGreaterEqual:
		return compareNum(v, c.Value, func(a, b float64) bool { return a >= b })
	case OpLessEqual:
		return compareNum(v, c.Value, func(a, b float64) bool { return a <= b })
	case OpIn:
		return inList(v, c.Value)
	case OpNotIn:
		return !inList(v, c.Value)
	case OpExists:
		return v != nil
	case OpNotExists:
		return v == nil
	case OpWildcardMatch:
		return wildcardMatch(v, fmt.Sprintf("%v", c.Value))
	}
	return false
}

func matchesSpatial(item map[string]any, op FilterOperator, value any) bool {
	area, ok := value.(canvus.Rectangle)
	if !ok {
		return false
	}
	loc, lok := item["location"].(map[string]any)
	size, sok := item["size"].(map[string]any)
	if !lok || !sok {
		return false
	}
	x, _ := toFloat(loc["x"])
	y, _ := toFloat(loc["y"])
	w, _ := toFloat(size["width"])
	h, _ := toFloat(size["height"])
	rect := canvus.Rectangle{X: x, Y: y, Width: w, Height: h}
	switch op {
	case OpSpatialIntersects:
		return Intersects(rect, area)
	case OpSpatialContains:
		return Intersects(area, rect)
	case OpSpatialWithin:
		return canvus.Contains(area, rect)
	}
	return false
}

func getNested(item map[string]any, field string) any {
	if !strings.Contains(field, ".") {
		return item[field]
	}
	cur := any(item)
	for _, part := range strings.Split(field, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = m[part]
		if cur == nil {
			return nil
		}
	}
	return cur
}

func equal(a, b any) bool {
	if a == nil || b == nil {
		return a == b
	}
	if af, aok := toFloat(a); aok {
		if bf, bok := toFloat(b); bok {
			return af == bf
		}
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func containsAny(haystack, needle any) bool {
	if haystack == nil {
		return false
	}
	switch h := haystack.(type) {
	case string:
		return strings.Contains(h, fmt.Sprintf("%v", needle))
	case []any:
		for _, item := range h {
			if equal(item, needle) {
				return true
			}
		}
	}
	return false
}

func compareNum(a, b any, cmp func(float64, float64) bool) bool {
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if !aok || !bok {
		return false
	}
	return cmp(af, bf)
}

func inList(v, list any) bool {
	items, ok := list.([]any)
	if !ok {
		return false
	}
	for _, item := range items {
		if equal(v, item) {
			return true
		}
	}
	return false
}

func wildcardMatch(v any, pattern string) bool {
	if v == nil {
		return false
	}
	// Convert wildcard pattern to a regex; * → .*, ? → .
	regex := regexp.QuoteMeta(pattern)
	regex = strings.ReplaceAll(regex, `\*`, ".*")
	regex = strings.ReplaceAll(regex, `\?`, ".")
	re, err := regexp.Compile("(?i)^" + regex + "$")
	if err != nil {
		return false
	}
	return re.MatchString(fmt.Sprintf("%v", v))
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	}
	return 0, false
}
