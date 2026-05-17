// Package canvus contains shared types for the Canvus SDK.
//
// COLOR FORMAT:
// All color values in the Canvus API use the RRGGBBAA format (8 uppercase hex digits).
//   - RR: Red component (00-FF)
//   - GG: Green component (00-FF)
//   - BB: Blue component (00-FF)
//   - AA: Alpha/transparency component (00-FF; 00=transparent, FF=opaque)
//
// Colors are ALWAYS uppercase when sent to and received from the API.
package canvus

import (
	"strings"
	"time"
)

// Filter provides generic, client-side filtering for SDK list/get endpoints.
// It supports arbitrary JSON criteria, wildcards ("*"), and JSONPath-like
// selectors ("$.foo.bar").
type Filter struct {
	Criteria map[string]any
}

// Filterable is implemented by types that can be matched against a Filter.
type Filterable interface {
	// AsMap returns the object as a map for filtering.
	AsMap() map[string]any
}

func getByJSONPath(obj map[string]any, path string) (any, bool) {
	if len(path) < 2 || path[:2] != "$." {
		return nil, false
	}
	parts := splitJSONPath(path[2:])
	cur := obj
	for i, part := range parts {
		if i == len(parts)-1 {
			if v, ok := cur[part]; ok {
				return v, true
			}
			return nil, false
		}
		if next, ok := cur[part].(map[string]any); ok {
			cur = next
		} else {
			return nil, false
		}
	}
	return nil, false
}

func splitJSONPath(path string) []string {
	return strings.Split(path, ".")
}

// Match returns true if obj matches the filter criteria. Supports wildcards
// ("*", "abc*", "*xyz", "*mid*") and JSONPath-style selectors.
func (f *Filter) Match(obj map[string]any) bool {
	for k, v := range f.Criteria {
		var actual any
		var ok bool
		if len(k) > 1 && k[:2] == "$." {
			actual, ok = getByJSONPath(obj, k)
		} else {
			actual, ok = obj[k], true
		}
		if !ok {
			return false
		}
		if v == "*" {
			continue
		}
		vs, vIsStr := v.(string)
		as, aIsStr := actual.(string)
		if vIsStr && aIsStr {
			if strings.HasPrefix(vs, "*") && strings.HasSuffix(vs, "*") {
				needle := vs[1 : len(vs)-1]
				if !strings.Contains(as, needle) {
					return false
				}
				continue
			}
			if strings.HasPrefix(vs, "*") {
				if !strings.HasSuffix(as, vs[1:]) {
					return false
				}
				continue
			}
			if strings.HasSuffix(vs, "*") {
				if !strings.HasPrefix(as, vs[:len(vs)-1]) {
					return false
				}
				continue
			}
		}
		if actual != v {
			return false
		}
	}
	return true
}

// Canvas represents a canvas resource in the Canvus system.
type Canvas struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Access      string `json:"access"`
	AssetSize   int64  `json:"asset_size"`
	CreatedAt   string `json:"created_at"`
	FolderID    string `json:"folder_id"`
	InTrash     bool   `json:"in_trash"`
	Mode        string `json:"mode"`
	ModifiedAt  string `json:"modified_at"`
	PreviewHash string `json:"preview_hash"`
	State       string `json:"state"`
}

// Note represents a note widget in the Canvus system.
type Note struct {
	ID              string  `json:"id"`
	Text            string  `json:"text"`
	Title           string  `json:"title"`
	BackgroundColor string  `json:"background_color"`
	Depth           float64 `json:"depth"`
	Location        *Point  `json:"location,omitempty"`
	ParentID        string  `json:"parent_id"`
	Pinned          bool    `json:"pinned"`
	Scale           float64 `json:"scale"`
	Size            *Size   `json:"size,omitempty"`
	State           string  `json:"state"`
	WidgetType      string  `json:"widget_type"`
}

// Image represents an image widget in the Canvus system.
type Image struct {
	ID               string  `json:"id"`
	Hash             string  `json:"hash"`
	Title            string  `json:"title"`
	OriginalFilename string  `json:"original_filename"`
	ParentID         string  `json:"parent_id"`
	Pinned           bool    `json:"pinned"`
	Scale            float64 `json:"scale"`
	Size             *Size   `json:"size,omitempty"`
	Location         *Point  `json:"location,omitempty"`
	State            string  `json:"state"`
	WidgetType       string  `json:"widget_type"`
	Depth            float64 `json:"depth"`
}

// PDF represents a PDF asset in the Canvus system.
type PDF struct {
	ID               string  `json:"id"`
	Hash             string  `json:"hash"`
	Title            string  `json:"title"`
	OriginalFilename string  `json:"original_filename"`
	ParentID         string  `json:"parent_id"`
	Pinned           bool    `json:"pinned"`
	Scale            float64 `json:"scale"`
	Size             *Size   `json:"size,omitempty"`
	Location         *Point  `json:"location,omitempty"`
	State            string  `json:"state"`
	WidgetType       string  `json:"widget_type"`
	Depth            float64 `json:"depth"`
	Index            int     `json:"index"`
}

// Video represents a video asset in the Canvus system.
type Video struct {
	ID               string  `json:"id"`
	Hash             string  `json:"hash"`
	Title            string  `json:"title"`
	OriginalFilename string  `json:"original_filename"`
	ParentID         string  `json:"parent_id"`
	Pinned           bool    `json:"pinned"`
	Scale            float64 `json:"scale"`
	Size             *Size   `json:"size,omitempty"`
	Location         *Point  `json:"location,omitempty"`
	State            string  `json:"state"`
	WidgetType       string  `json:"widget_type"`
	Depth            float64 `json:"depth"`
	PlaybackPosition float64 `json:"playback_position"`
	PlaybackState    string  `json:"playback_state"`
}

// Widget is the polymorphic envelope returned by the generic widgets endpoints.
type Widget struct {
	ID          string       `json:"id"`
	WidgetType  string       `json:"widget_type"`
	ParentID    string       `json:"parent_id"`
	Location    *Point       `json:"location,omitempty"`
	Size        *Size        `json:"size,omitempty"`
	Pinned      bool         `json:"pinned"`
	Scale       float64      `json:"scale"`
	State       string       `json:"state"`
	Depth       float64      `json:"depth"`
	Annotations []Annotation `json:"annotations,omitempty"`
}

// Annotation represents a drawing annotation (stroke) on a widget.
type Annotation struct {
	ID         string  `json:"id"`
	WidgetType string  `json:"widget_type"`
	ParentID   string  `json:"parent_id,omitempty"`
	State      string  `json:"state,omitempty"`
	Page       int     `json:"page,omitempty"`
	Depth      float64 `json:"depth"`
	LineColor  string  `json:"line_color"`
	Points     string  `json:"points"`
}

// Anchor represents a navigation anchor widget.
type Anchor struct {
	ID          string  `json:"id"`
	AnchorIndex int     `json:"anchor_index"`
	AnchorName  string  `json:"anchor_name"`
	ParentID    string  `json:"parent_id"`
	Pinned      bool    `json:"pinned"`
	Scale       float64 `json:"scale"`
	Size        *Size   `json:"size,omitempty"`
	Location    *Point  `json:"location,omitempty"`
	State       string  `json:"state"`
	WidgetType  string  `json:"widget_type"`
	Depth       float64 `json:"depth"`
}

// Browser represents a browser widget.
type Browser struct {
	ID                    string  `json:"id"`
	WidgetType            string  `json:"widget_type"`
	Depth                 float64 `json:"depth"`
	Location              *Point  `json:"location,omitempty"`
	MainFrameScrollOffset *Point  `json:"main_frame_scroll_offset,omitempty"`
	ParentID              string  `json:"parent_id"`
	Pinned                bool    `json:"pinned"`
	Scale                 float64 `json:"scale"`
	Size                  *Size   `json:"size,omitempty"`
	State                 string  `json:"state"`
	Title                 string  `json:"title"`
	TransparentMode       bool    `json:"transparent_mode"`
	URL                   string  `json:"url"`
}

// Connector represents a connector widget between two endpoints.
type Connector struct {
	ID         string        `json:"id"`
	Src        *ConnectorEnd `json:"src,omitempty"`
	Dst        *ConnectorEnd `json:"dst,omitempty"`
	LineColor  string        `json:"line_color"`
	LineWidth  int           `json:"line_width"`
	State      string        `json:"state"`
	Type       string        `json:"type"`
	WidgetType string        `json:"widget_type"`
}

// ConnectorEnd represents the source or destination of a connector.
type ConnectorEnd struct {
	AutoLocation bool   `json:"auto_location"`
	ID           string `json:"id"`
	RelLocation  *Point `json:"rel_location,omitempty"`
	Tip          string `json:"tip"`
}

// Table represents a table widget in the Canvus system.
//
// Per changelog §4, the server's serializeTableProperties() emits only
// `title` and `grid_size`. Documented but never-returned fields
// `column_widths` and `row_heights` are deliberately omitted here.
type Table struct {
	ID         string    `json:"id"`
	WidgetType string    `json:"widget_type"`
	Title      string    `json:"title"`
	GridSize   *GridSize `json:"grid_size,omitempty"`
	ParentID   string    `json:"parent_id,omitempty"`
	Pinned     bool      `json:"pinned"`
	Scale      float64   `json:"scale"`
	Size       *Size     `json:"size,omitempty"`
	Location   *Point    `json:"location,omitempty"`
	State      string    `json:"state"`
	Depth      float64   `json:"depth"`
}

// GridSize represents the immutable grid dimensions of a Table widget.
// Per changelog §5, GridSize is set at creation time and silently ignored on PATCH.
type GridSize struct {
	Columns int `json:"columns"`
	Rows    int `json:"rows"`
}

// TableCell represents a single cell inside a Table widget.
type TableCell struct {
	CellID  string `json:"cell_id"`
	Index   []int  `json:"index"`
	Content string `json:"content"`
}

// IPVideo represents an IP video stream widget.
//
// Per changelog §2, IP Video widgets cannot be created via the API; the
// SDK exposes GET/PATCH/DELETE only. Field naming uses snake_case based on
// the broader Canvus API convention; if a live-API check reveals hyphen
// keys (per changelog §3 unresolved), update tags in a follow-up.
type IPVideo struct {
	ID         string  `json:"id"`
	WidgetType string  `json:"widget_type"`
	Title      string  `json:"title,omitempty"`
	URL        string  `json:"url,omitempty"`
	ParentID   string  `json:"parent_id,omitempty"`
	Location   *Point  `json:"location,omitempty"`
	Size       *Size   `json:"size,omitempty"`
	Pinned     bool    `json:"pinned"`
	Scale      float64 `json:"scale"`
	State      string  `json:"state"`
	Depth      float64 `json:"depth"`
}

// RDPConnection represents an RDP session widget.
//
// Per changelog §2, RDP Connection widgets cannot be created via the API;
// the SDK exposes GET/PATCH/DELETE only.
//
// Per changelog §3 (unresolved), the C++ source uses hyphenated keys
// (`host-id`, `content-id`, `connection-name`, `host-site`) while the
// published docs use snake_case. We pick the hyphenated form as the more
// defensive choice — the C++ source is the runtime source of truth.
// Once Phase 4 confirms live-API behaviour, update tags in this struct
// (search for marker "ChangelogSection3" below).
type RDPConnection struct {
	ID             string  `json:"id"`
	WidgetType     string  `json:"widget_type"`
	ConnectionName string  `json:"connection-name,omitempty"` // ChangelogSection3
	HostSite       string  `json:"host-site,omitempty"`       // ChangelogSection3
	HostID         string  `json:"host-id,omitempty"`         // ChangelogSection3
	ContentID      string  `json:"content-id,omitempty"`      // ChangelogSection3
	Title          string  `json:"title,omitempty"`
	ParentID       string  `json:"parent_id,omitempty"`
	Location       *Point  `json:"location,omitempty"`
	Size           *Size   `json:"size,omitempty"`
	Pinned         bool    `json:"pinned"`
	Scale          float64 `json:"scale"`
	State          string  `json:"state"`
	Depth          float64 `json:"depth"`
}

// UploadItem represents a single item in the canvas uploads-folder listing.
type UploadItem struct {
	ID               string  `json:"id"`
	WidgetType       string  `json:"widget_type,omitempty"`
	Title            string  `json:"title,omitempty"`
	OriginalFilename string  `json:"original_filename,omitempty"`
	Hash             string  `json:"hash,omitempty"`
	ParentID         string  `json:"parent_id,omitempty"`
	Location         *Point  `json:"location,omitempty"`
	Size             *Size   `json:"size,omitempty"`
	State            string  `json:"state,omitempty"`
	Depth            float64 `json:"depth,omitempty"`
}

// Background is the legacy nested-struct background payload kept for
// backwards-compat with the old SDK. New code should prefer CanvasBackground.
type Background struct {
	Type            string           `json:"type"`
	BackgroundColor string           `json:"background_color,omitempty"`
	Haze            *BackgroundHaze  `json:"haze,omitempty"`
	Grid            *BackgroundGrid  `json:"grid,omitempty"`
	Image           *BackgroundImage `json:"image,omitempty"`
}

// BackgroundHaze represents an animated haze background.
type BackgroundHaze struct {
	Color1 string  `json:"color1"`
	Color2 string  `json:"color2"`
	Speed  float64 `json:"speed"`
	Scale  float64 `json:"scale"`
}

// BackgroundGrid represents an overlay grid on the canvas background.
type BackgroundGrid struct {
	Visible bool   `json:"visible"`
	Color   string `json:"color"`
}

// BackgroundImage represents an image asset used as canvas background.
type BackgroundImage struct {
	Hash string `json:"hash"`
	Fit  string `json:"fit"`
}

// ColorPreset is a single named color preset entry (legacy per-name shape).
type ColorPreset struct {
	Name string `json:"name"`
}

// ColorPresets bundles the four preset color arrays for a canvas.
// All color arrays contain RRGGBBAA format strings.
type ColorPresets struct {
	Annotation     []string `json:"annotation"`
	Connector      []string `json:"connector"`
	NoteBackground []string `json:"note_background"`
	NoteText       []string `json:"note_text"`
}

// MipmapInfo represents mipmap information for an asset.
type MipmapInfo struct {
	Resolution struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"resolution"`
	MaxLevel int `json:"max_level"`
	Pages    int `json:"pages"`
}

// VideoInput represents a video input widget on a canvas.
type VideoInput struct {
	ID         string  `json:"id"`
	WidgetType string  `json:"widget_type"`
	Depth      float64 `json:"depth"`
	HostID     string  `json:"host-id"`
	Location   *Point  `json:"location,omitempty"`
	ParentID   string  `json:"parent_id"`
	Pinned     bool    `json:"pinned"`
	Scale      float64 `json:"scale"`
	Size       *Size   `json:"size,omitempty"`
	Source     string  `json:"source"`
	State      string  `json:"state"`
}

// VideoOutput represents a video output channel on a client device.
type VideoOutput struct {
	Index      int    `json:"index,omitempty"`
	Label      string `json:"label,omitempty"`
	Source     string `json:"source,omitempty"`
	Suspended  bool   `json:"suspended,omitempty"`
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	Resolution *Size  `json:"resolution,omitempty"`
	State      string `json:"state,omitempty"`
	WidgetType string `json:"widget_type,omitempty"`
}

// Workspace represents one of the workspaces on a client device.
type Workspace struct {
	CanvasID         string     `json:"canvas_id"`
	CanvasSize       *Size      `json:"canvas_size,omitempty"`
	Index            int        `json:"index"`
	InfoPanelVisible bool       `json:"info_panel_visible"`
	Location         *Point     `json:"location,omitempty"`
	Pinned           bool       `json:"pinned"`
	ServerID         string     `json:"server_id"`
	Size             *Size      `json:"size,omitempty"`
	State            string     `json:"state"`
	User             string     `json:"user"`
	ViewRectangle    *Rectangle `json:"view_rectangle,omitempty"`
	WorkspaceName    string     `json:"workspace_name"`
	WorkspaceState   string     `json:"workspace_state"`
}

// Size describes a 2-D extent in canvas units (pixels).
type Size struct {
	Height float64 `json:"height"`
	Width  float64 `json:"width"`
}

// Point describes a 2-D position in canvas units (pixels).
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Rectangle describes an axis-aligned rectangle in canvas units (pixels).
type Rectangle struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// WorkspaceSelector selects a workspace by index, name, or user.
type WorkspaceSelector struct {
	Index *int
	Name  *string
	User  *string
}

// Viewport represents a workspace's view rectangle in screen coordinates.
type Viewport struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// SetViewportOptions supports setting a workspace viewport either explicitly
// (X/Y/Width/Height) or by centering on a widget (WidgetID).
type SetViewportOptions struct {
	WidgetID *string
	X        *float64
	Y        *float64
	Width    *float64
	Height   *float64
	Margin   float64
}

// OpenCanvasOptions configures Session.OpenCanvasOnWorkspace.
type OpenCanvasOptions struct {
	CanvasID     string        `json:"canvas_id"`
	ServerID     string        `json:"server_id,omitempty"`
	UserEmail    string        `json:"user_email,omitempty"`
	CenterX      *float64      `json:"center_x,omitempty"`
	CenterY      *float64      `json:"center_y,omitempty"`
	WidgetID     *string       `json:"widget_id,omitempty"`
	PollTimeout  time.Duration `json:"-"`
	PollInterval time.Duration `json:"-"`
}

// UpdateWorkspaceRequest is the body of a workspace PATCH.
type UpdateWorkspaceRequest struct {
	InfoPanelVisible *bool      `json:"info_panel_visible,omitempty"`
	Pinned           *bool      `json:"pinned,omitempty"`
	ViewRectangle    *Rectangle `json:"view_rectangle,omitempty"`
}

// CreateCanvasRequest is the payload for creating a canvas.
type CreateCanvasRequest struct {
	Name     string `json:"name,omitempty"`
	FolderID string `json:"folder_id,omitempty"`
}

// UpdateCanvasRequest is the payload for updating a canvas (rename, mode change).
type UpdateCanvasRequest struct {
	Name string `json:"name,omitempty"`
	Mode string `json:"mode,omitempty"`
}

// MoveOrCopyCanvasRequest is the payload for moving or copying a canvas.
type MoveOrCopyCanvasRequest struct {
	FolderID  string `json:"folder_id"`
	Conflicts string `json:"conflicts,omitempty"`
}

// CanvasPermissions represents permission overrides on a canvas.
type CanvasPermissions struct {
	EditorsCanShare bool                    `json:"editors_can_share"`
	Users           []CanvasUserPermission  `json:"users"`
	Groups          []CanvasGroupPermission `json:"groups"`
	LinkPermission  string                  `json:"link_permission"`
}

// CanvasBackground represents the background settings for a canvas.
type CanvasBackground struct {
	Type            string           `json:"type"`
	Haze            *HazeSettings    `json:"haze,omitempty"`
	Grid            *GridSettings    `json:"grid,omitempty"`
	Image           *BackgroundImage `json:"image,omitempty"`
	BackgroundColor string           `json:"background_color,omitempty"`
}

// HazeSettings represents haze background settings.
type HazeSettings struct {
	Color1 string  `json:"color1"`
	Color2 string  `json:"color2"`
	Speed  float64 `json:"speed"`
	Scale  float64 `json:"scale"`
}

// GridSettings represents grid overlay settings.
type GridSettings struct {
	Visible bool   `json:"visible"`
	Color   string `json:"color"`
}

// CanvasUserPermission grants a specific user a permission on a canvas.
type CanvasUserPermission struct {
	ID         int64  `json:"id"`
	Permission string `json:"permission"`
	Inherited  bool   `json:"inherited"`
}

// CanvasGroupPermission grants a group a permission on a canvas.
type CanvasGroupPermission struct {
	ID         int64  `json:"id"`
	Permission string `json:"permission"`
	Inherited  bool   `json:"inherited"`
}

// Asset represents a generic asset in the Canvus system.
type Asset struct {
	ID string `json:"id"`
}

// AsMap returns the Canvas as a map for filtering.
func (c Canvas) AsMap() map[string]any {
	return map[string]any{
		"id":           c.ID,
		"name":         c.Name,
		"access":       c.Access,
		"asset_size":   c.AssetSize,
		"created_at":   c.CreatedAt,
		"folder_id":    c.FolderID,
		"in_trash":     c.InTrash,
		"mode":         c.Mode,
		"modified_at":  c.ModifiedAt,
		"preview_hash": c.PreviewHash,
		"state":        c.State,
	}
}

// AsMap returns the Widget as a map for filtering.
func (w Widget) AsMap() map[string]any {
	m := map[string]any{
		"id":          w.ID,
		"widget_type": w.WidgetType,
		"parent_id":   w.ParentID,
		"pinned":      w.Pinned,
		"scale":       w.Scale,
		"state":       w.State,
		"depth":       w.Depth,
	}
	if w.Location != nil {
		m["location"] = map[string]any{"x": w.Location.X, "y": w.Location.Y}
	}
	if w.Size != nil {
		m["size"] = map[string]any{"width": w.Size.Width, "height": w.Size.Height}
	}
	return m
}

// FilterSlice returns a new slice containing only the elements that match filter.
//
//	filter := &canvus.Filter{Criteria: map[string]interface{}{"name": "My Canvas"}}
//	filtered := canvus.FilterSlice(canvases, filter)
func FilterSlice[T Filterable](elems []T, filter *Filter) []T {
	if filter == nil {
		return elems
	}
	var out []T
	for _, elem := range elems {
		if filter.Match(elem.AsMap()) {
			out = append(out, elem)
		}
	}
	return out
}
