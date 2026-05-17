# Widget Types & Data Models

Canvus supports a rich set of widget types for displaying and organizing content on the canvas. All widgets share common base fields and follow standard CRUD patterns.

---

## Common Widget Fields

All canvas widgets inherit from the same base model and share these fields:

### Base Widget Properties

```json
{
  "widget-id": "uuid",
  "widget-type": "note|image|video|pdf|browser|anchor|connector|table|video-input|ip-video|rdp-connection",
  "location": {"x": 100.5, "y": 200.75},
  "depth": 5.0,
  "size": {"width": 400, "height": 300},
  "scale": 1.0,
  "is-pinned": false,
  "created": "2025-01-15T10:30:00Z",
  "modified": "2025-05-17T14:22:15Z"
}
```

| Field | Type | Description |
|-------|------|-------------|
| widget-id | string (UUID) | Unique identifier for the widget |
| widget-type | string | Type of widget (immutable after creation) |
| location | object | Position {x, y} in pixels (absolute canvas coordinates) |
| depth | number | Z-order (0 = background, higher = foreground); float for fine control |
| size | object | Dimensions {width, height} in pixels |
| scale | number | Scale factor (1.0 = 100%, 0.5 = 50% size) |
| is-pinned | boolean | If true, widget cannot be moved by users (visual lock) |
| created | string (ISO 8601) | Widget creation timestamp |
| modified | string (ISO 8601) | Last modification timestamp |

### Coordinate System

**CRITICAL:** All coordinates are in **pixels**, not normalized (0-1). A canvas with 3840×2160 resolution has pixel coordinates ranging from 0 to 3840 (x) and 0 to 2160 (y).

- Do NOT multiply canvas size by coordinates
- Location (100, 200) means 100 pixels from left, 200 pixels from top
- Use coordinate utilities ONLY for viewport zoom/pan conversions, not widget placement

---

## Note

A note is a simple text widget with customizable color and styling.

### Data Model

```json
{
  "note-id": "uuid",
  "widget-type": "note",
  "text": "Important reminder",
  "title": "Optional title",
  "background-color": "#FFFF00",
  "text-color": "#000000",
  "auto-text-color": true,
  "location": {"x": 100, "y": 200},
  "depth": 5.0,
  "size": {"width": 300, "height": 200},
  "scale": 1.0,
  "is-pinned": false
}
```

| Field | Type | Writable | Description |
|-------|------|----------|-------------|
| note-id | string (UUID) | no | Unique identifier (generated on creation) |
| text | string | yes | Note text content (markdown supported in some versions) |
| title | string | yes | Optional title displayed above text |
| background-color | string (hex) | yes | Background color (e.g., "#FFFF00") |
| text-color | string (hex) | yes | Text color |
| auto-text-color | boolean | yes | Auto-select text color (black/white) based on background |

### Lifecycle

1. **Create**: POST /api/v1/canvases/{canvas-id}/notes
2. **Read**: GET /api/v1/canvases/{canvas-id}/notes/{note-id}
3. **Update**: PATCH /api/v1/canvases/{canvas-id}/notes/{note-id}
4. **Delete**: DELETE /api/v1/canvases/{canvas-id}/notes/{note-id}
5. **List**: GET /api/v1/canvases/{canvas-id}/notes
6. **Subscribe**: GET /api/v1/canvases/{canvas-id}/notes?subscribe (streaming updates)

---

## Image

An image widget displays a raster image file (PNG, JPEG, WebP, etc.) with support for tiled access.

### Data Model

```json
{
  "image-id": "uuid",
  "widget-type": "image",
  "original-filename": "photo.jpg",
  "title": "My Photo",
  "asset-hash": "a1b2c3d4e5f6",
  "asset-hash-private": "private-hash",
  "mime-type": "image/jpeg",
  "file-size": 1024000,
  "location": {"x": 100, "y": 200},
  "depth": 5.0,
  "size": {"width": 800, "height": 600},
  "scale": 1.0,
  "is-pinned": false
}
```

| Field | Type | Writable | Description |
|-------|------|----------|-------------|
| image-id | string (UUID) | no | Unique identifier |
| original-filename | string | no | Original filename from upload |
| title | string | yes | User-assigned title |
| asset-hash | string | no | Public hash for download (use in /api/v1/assets/{hash}) |
| asset-hash-private | string | no | Private hash (internal use) |
| mime-type | string | no | MIME type (image/jpeg, image/png, etc.) |
| file-size | integer | no | File size in bytes |

### Lifecycle

1. **Create**: POST /api/v1/canvases/{canvas-id}/images (multipart/form-data with file)
2. **Read**: GET /api/v1/canvases/{canvas-id}/images/{image-id}
3. **Update**: PATCH /api/v1/canvases/{canvas-id}/images/{image-id} (metadata only, NOT the file)
4. **Download**: GET /api/v1/canvases/{canvas-id}/images/{image-id}/download
5. **Delete**: DELETE /api/v1/canvases/{canvas-id}/images/{image-id}
6. **List**: GET /api/v1/canvases/{canvas-id}/images
7. **Subscribe**: GET /api/v1/canvases/{canvas-id}/images?subscribe

### Asset Access

- Download via widget: `GET /api/v1/canvases/{canvas-id}/images/{image-id}/download`
- Download via hash: `GET /api/v1/assets/{asset-hash}` (requires canvas-id header)
- Mipmaps for large images: `GET /api/v1/mipmaps/{asset-hash}/{level}`

---

## Video

A video widget displays a video file with playback controls.

### Data Model

```json
{
  "video-id": "uuid",
  "widget-type": "video",
  "original-filename": "video.mp4",
  "title": "My Video",
  "asset-hash": "v1v2v3v4v5v6",
  "mime-type": "video/mp4",
  "file-size": 52428800,
  "seek-position": 0.5,
  "seek-timestamp": "00:01:30",
  "playback-state": "playing|paused|stopped",
  "muted": false,
  "duration": "00:05:00",
  "location": {"x": 100, "y": 200},
  "depth": 5.0,
  "size": {"width": 1280, "height": 720},
  "scale": 1.0,
  "is-pinned": false
}
```

| Field | Type | Writable | Description |
|-------|------|----------|-------------|
| video-id | string (UUID) | no | Unique identifier |
| original-filename | string | no | Original filename |
| title | string | yes | User-assigned title |
| asset-hash | string | no | Public hash for download |
| mime-type | string | no | MIME type (video/mp4, video/webm, etc.) |
| file-size | integer | no | File size in bytes |
| seek-position | number | yes | Playback position (0.0 = start, 1.0 = end) |
| seek-timestamp | string | yes | Playback position as time code (HH:MM:SS) |
| playback-state | string | yes | playing, paused, or stopped |
| muted | boolean | yes | Mute audio |
| duration | string | no | Total video duration (HH:MM:SS) |

### Lifecycle

1. **Create**: POST /api/v1/canvases/{canvas-id}/videos (multipart/form-data)
2. **Read**: GET /api/v1/canvases/{canvas-id}/videos/{video-id}
3. **Update**: PATCH /api/v1/canvases/{canvas-id}/videos/{video-id} (playback state, etc.)
4. **Download**: GET /api/v1/canvases/{canvas-id}/videos/{video-id}/download
5. **Delete**: DELETE /api/v1/canvases/{canvas-id}/videos/{video-id}
6. **List**: GET /api/v1/canvases/{canvas-id}/videos
7. **Subscribe**: GET /api/v1/canvases/{canvas-id}/videos?subscribe

---

## PDF

A PDF widget displays a PDF document with page navigation.

### Data Model

```json
{
  "pdf-id": "uuid",
  "widget-type": "pdf",
  "original-filename": "document.pdf",
  "title": "Report",
  "asset-hash": "p1p2p3p4p5p6",
  "mime-type": "application/pdf",
  "file-size": 2097152,
  "index": 1,
  "page-count": 42,
  "location": {"x": 100, "y": 200},
  "depth": 5.0,
  "size": {"width": 600, "height": 800},
  "scale": 1.0,
  "is-pinned": false
}
```

| Field | Type | Writable | Description |
|-------|------|----------|-------------|
| pdf-id | string (UUID) | no | Unique identifier |
| original-filename | string | no | Original filename |
| title | string | yes | User-assigned title |
| asset-hash | string | no | Public hash for download |
| mime-type | string | no | application/pdf |
| file-size | integer | no | File size in bytes |
| index | integer | yes | Current page number (1-based) |
| page-count | integer | no | Total number of pages |

### Lifecycle

1. **Create**: POST /api/v1/canvases/{canvas-id}/pdfs (multipart/form-data)
2. **Read**: GET /api/v1/canvases/{canvas-id}/pdfs/{pdf-id}
3. **Update**: PATCH /api/v1/canvases/{canvas-id}/pdfs/{pdf-id} (page index, etc.)
4. **Download**: GET /api/v1/canvases/{canvas-id}/pdfs/{pdf-id}/download
5. **Delete**: DELETE /api/v1/canvases/{canvas-id}/pdfs/{pdf-id}
6. **List**: GET /api/v1/canvases/{canvas-id}/pdfs
7. **Subscribe**: GET /api/v1/canvases/{canvas-id}/pdfs?subscribe

---

## Browser

A browser widget embeds web content (HTML, external websites).

### Data Model

```json
{
  "browser-id": "uuid",
  "widget-type": "browser",
  "source": "https://example.com",
  "title": "Example Site",
  "transparent-mode": false,
  "main-frame-scroll-offset": {"x": 0, "y": 0},
  "location": {"x": 100, "y": 200},
  "depth": 5.0,
  "size": {"width": 1024, "height": 768},
  "scale": 1.0,
  "is-pinned": false
}
```

| Field | Type | Writable | Description |
|-------|------|----------|-------------|
| browser-id | string (UUID) | no | Unique identifier |
| source | string (URL) | yes | Web URL to display |
| title | string | yes | User-assigned title |
| transparent-mode | boolean | yes | Render background as transparent |
| main-frame-scroll-offset | object | yes | Scroll position {x, y} |

### Lifecycle

1. **Create**: POST /api/v1/canvases/{canvas-id}/browsers
2. **Read**: GET /api/v1/canvases/{canvas-id}/browsers/{browser-id}
3. **Update**: PATCH /api/v1/canvases/{canvas-id}/browsers/{browser-id}
4. **Delete**: DELETE /api/v1/canvases/{canvas-id}/browsers/{browser-id}
5. **List**: GET /api/v1/canvases/{canvas-id}/browsers
6. **Subscribe**: GET /api/v1/canvases/{canvas-id}/browsers?subscribe

---

## Anchor

An anchor is a named navigation point on the canvas (used for creating shortcuts or camera positions).

### Data Model

```json
{
  "anchor-id": "uuid",
  "widget-type": "anchor",
  "anchor-name": "Anchor 1",
  "location": {"x": 100, "y": 200},
  "depth": 5.0,
  "size": {"width": 0, "height": 0},
  "scale": 1.0,
  "is-pinned": false
}
```

| Field | Type | Writable | Description |
|-------|------|----------|-------------|
| anchor-id | string (UUID) | no | Unique identifier |
| anchor-name | string | yes | Name of the anchor |

### Lifecycle

1. **Create**: POST /api/v1/canvases/{canvas-id}/anchors
2. **Read**: GET /api/v1/canvases/{canvas-id}/anchors/{anchor-id}
3. **Update**: PATCH /api/v1/canvases/{canvas-id}/anchors/{anchor-id}
4. **Delete**: DELETE /api/v1/canvases/{canvas-id}/anchors/{anchor-id}
5. **List**: GET /api/v1/canvases/{canvas-id}/anchors
6. **Subscribe**: GET /api/v1/canvases/{canvas-id}/anchors?subscribe

---

## Connector

A connector is a line or curve connecting two widgets with optional arrow tips.

### Data Model

```json
{
  "connector-id": "uuid",
  "widget-type": "connector",
  "connector-type": "line|curve|arrow",
  "src": "source-widget-id",
  "src-rel-location": {"x": 0.5, "y": 0.5},
  "src-auto-location": true,
  "src-tip": "none|arrow|circle",
  "dst": "dest-widget-id",
  "dst-rel-location": {"x": 0.5, "y": 0.5},
  "dst-auto-location": true,
  "dst-tip": "none|arrow|circle",
  "line-color": "#FF0000",
  "line-width": 2.0,
  "depth": 1.0
}
```

| Field | Type | Writable | Description |
|-------|------|----------|-------------|
| connector-id | string (UUID) | no | Unique identifier |
| connector-type | string | yes | line, curve, or arrow |
| src | string (widget-id) | no | Source widget ID |
| src-rel-location | object | yes | Relative attachment point {x: 0-1, y: 0-1} |
| src-auto-location | boolean | yes | Auto-calculate attachment point |
| src-tip | string | yes | none, arrow, or circle |
| dst | string (widget-id) | no | Destination widget ID |
| dst-rel-location | object | yes | Relative attachment point |
| dst-auto-location | boolean | yes | Auto-calculate attachment point |
| dst-tip | string | yes | none, arrow, or circle |
| line-color | string (hex) | yes | Line color |
| line-width | number | yes | Line width in pixels |
| depth | number | yes | Z-order |

### Lifecycle

1. **Create**: POST /api/v1/canvases/{canvas-id}/connectors
2. **Read**: GET /api/v1/canvases/{canvas-id}/connectors/{connector-id}
3. **Update**: PATCH /api/v1/canvases/{canvas-id}/connectors/{connector-id}
4. **Delete**: DELETE /api/v1/canvases/{canvas-id}/connectors/{connector-id}
5. **List**: GET /api/v1/canvases/{canvas-id}/connectors
6. **Subscribe**: GET /api/v1/canvases/{canvas-id}/connectors?subscribe

---

## Table

A table widget is a grid of cells for organizing data.

### Data Model

```json
{
  "table-id": "uuid",
  "widget-type": "table",
  "title": "Data Table",
  "grid-size": {"columns": 5, "rows": 10},
  "column-widths": [100, 150, 200, 100, 100],
  "row-heights": [20, 20, 20, 20, 20],
  "location": {"x": 100, "y": 200},
  "depth": 5.0,
  "size": {"width": 650, "height": 400},
  "scale": 1.0,
  "is-pinned": false
}
```

| Field | Type | Writable | Description |
|-------|------|----------|-------------|
| table-id | string (UUID) | no | Unique identifier |
| title | string | yes | Table title |
| grid-size | object | no | Grid dimensions {columns, rows} |
| column-widths | array | yes | Width of each column in pixels |
| row-heights | array | yes | Height of each row in pixels |

### Lifecycle

1. **Create**: POST /api/v1/canvases/{canvas-id}/tables
2. **Read**: GET /api/v1/canvases/{canvas-id}/tables/{table-id}
3. **Update**: PATCH /api/v1/canvases/{canvas-id}/tables/{table-id}
4. **Get Cells**: GET /api/v1/canvases/{canvas-id}/tables/{table-id}/cells
5. **Delete**: DELETE /api/v1/canvases/{canvas-id}/tables/{table-id}
6. **List**: GET /api/v1/canvases/{canvas-id}/tables
7. **Subscribe**: GET /api/v1/canvases/{canvas-id}/tables?subscribe

---

## Video Input

A video input widget displays a live video stream from a camera or capture device.

### Data Model

```json
{
  "widget-id": "uuid",
  "widget-type": "video-input",
  "source": "camera-device-id",
  "name": "Webcam",
  "resolution": "1920x1080",
  "location": {"x": 100, "y": 200},
  "depth": 5.0,
  "size": {"width": 960, "height": 540},
  "scale": 1.0,
  "is-pinned": false
}
```

| Field | Type | Writable | Description |
|-------|------|----------|-------------|
| source | string | yes | Camera device ID |
| name | string | yes | Device name |
| resolution | string | yes | Resolution (e.g., "1920x1080") |

### Lifecycle

1. **Create**: POST /api/v1/canvases/{canvas-id}/video-inputs
2. **Read**: GET /api/v1/canvases/{canvas-id}/video-inputs/{widget-id}
3. **Update**: PATCH /api/v1/canvases/{canvas-id}/video-inputs/{widget-id}
4. **Delete**: DELETE /api/v1/canvases/{canvas-id}/video-inputs/{widget-id}
5. **List**: GET /api/v1/canvases/{canvas-id}/video-inputs
6. **Subscribe**: GET /api/v1/canvases/{canvas-id}/video-inputs?subscribe

---

## IP Video

An IP video widget displays a stream from an IP camera or streaming source.

### Data Model

Similar to Video Input, but with a network source (RTSP, MJPEG stream, etc.).

---

## RDP Connection

An RDP connection widget displays a remote desktop session.

### Data Model

```json
{
  "widget-id": "uuid",
  "widget-type": "rdp-connection",
  "connection-name": "Server 1",
  "host-site": "192.168.1.100",
  "content-id": "session-id",
  "title": "RDP Display",
  "location": {"x": 100, "y": 200},
  "depth": 5.0,
  "size": {"width": 1280, "height": 720},
  "scale": 1.0,
  "is-pinned": false
}
```

---

## Widget Creation Patterns

### Minimal Create (Notes)

```bash
curl -X POST https://canvus-server/api/v1/canvases/{canvas-id}/notes \
  -H "Private-Token: token" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Note content",
    "location": {"x": 100, "y": 200}
  }'
```

### Full Create (Image with Multipart)

```bash
curl -X POST https://canvus-server/api/v1/canvases/{canvas-id}/images \
  -H "Private-Token: token" \
  -F "data=@photo.jpg" \
  -F 'json={"title":"My Photo","location":{"x":100,"y":200}}'
```

### Update (PATCH)

```bash
curl -X PATCH https://canvus-server/api/v1/canvases/{canvas-id}/notes/{note-id} \
  -H "Private-Token: token" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Updated text",
    "background-color": "#FF0000"
  }'
```

---

## Widget Deletion

All widgets support standard HTTP DELETE:

```bash
curl -X DELETE https://canvus-server/api/v1/canvases/{canvas-id}/notes/{note-id} \
  -H "Private-Token: token"
# Returns 204 No Content
```

Deletion is permanent and cannot be undone.
