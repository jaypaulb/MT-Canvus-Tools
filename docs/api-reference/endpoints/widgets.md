# Widget Endpoints

All canvas widgets share common base fields: `location` (x, y), `depth` (z-order), `size` (width, height), `scale`, and `is-pinned`. Widget-specific fields are documented per type.

## Generic Widgets

### `GET /api/v1/canvases/{canvas-id}/widgets`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all widgets on the canvas, regardless of type. Returns mixed widget types in a single response.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| subscribe | boolean | no | false | Enable streaming updates |

**Response (200):** JSON array of widget objects (mixed types)

```json
[
  {
    "location": {"x": 100, "y": 200},
    "depth": 5.0,
    "size": {"width": 400, "height": 300},
    "scale": 1.0,
    "is-pinned": false,
    "widget-type": "note|image|video|pdf|browser|anchor|connector|table",
    "widget-id": "uuid",
    ...type-specific fields...
  }
]
```

**Errors:**
- 401: Unauthorized (if canvas requires authentication)
- 404: Canvas not found
- 500: Server error

---

### `GET /api/v1/canvases/{canvas-id}/widgets/{widget-id}`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single widget by ID (any type).

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |
| widget-id | string (UUID) | Widget identifier |

**Response (200):** Single widget object with all fields

**Errors:**
- 404: Canvas or widget not found
- 500: Server error

---

### `POST /api/v1/canvases/{canvas-id}/widgets/clone`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** experimental (returns 501 Not Implemented)

Cross-canvas widget cloning (planned feature, see Issue #45).

---

## Notes

### `GET /api/v1/canvases/{canvas-id}/notes`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all notes on the canvas.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| subscribe | boolean | no | false | Enable streaming updates |

**Response (200):** JSON array of note objects

```json
[
  {
    "note-id": "uuid",
    "text": "Note content",
    "title": "Note title",
    "background-color": "#FFFF00",
    "text-color": "#000000",
    "auto-text-color": true,
    "location": {"x": 100, "y": 200},
    "depth": 5.0,
    "size": {"width": 300, "height": 200},
    "scale": 1.0,
    "is-pinned": false
  }
]
```

**Errors:**
- 404: Canvas not found
- 500: Server error

---

### `GET /api/v1/canvases/{canvas-id}/notes/{note-id}`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single note by ID.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |
| note-id | string (UUID) | Note identifier |

**Response (200):** Single note object

**Errors:**
- 404: Canvas or note not found
- 500: Server error

---

### `POST /api/v1/canvases/{canvas-id}/notes`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a new note on the canvas.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Request body:**

```json
{
  "text": "Note content",
  "title": "Optional title",
  "background-color": "#FFFF00",
  "text-color": "#000000",
  "auto-text-color": true,
  "location": {"x": 100, "y": 200},
  "size": {"width": 300, "height": 200},
  "depth": 5.0,
  "scale": 1.0,
  "is-pinned": false
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| text | string | no | Note text content |
| title | string | no | Optional note title |
| background-color | string (hex) | no | Background color (e.g., "#FFFF00") |
| text-color | string (hex) | no | Text color |
| auto-text-color | boolean | no | Auto-select text color based on background |
| location | object | no | Position {x, y} in pixels |
| size | object | no | Dimensions {width, height} |
| depth | number | no | Z-order (0-based) |
| scale | number | no | Scale factor (default 1.0) |
| is-pinned | boolean | no | Pin to canvas (prevent moving) |

**Response (201):** Created note object with note-id

**Errors:**
- 400: Bad request (invalid color or location format)
- 401: Unauthorized
- 404: Canvas not found
- 500: Server error

---

### `PATCH /api/v1/canvases/{canvas-id}/notes/{note-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update a note. All fields are optional.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |
| note-id | string (UUID) | Note identifier |

**Request body:**

```json
{
  "text": "Updated content",
  "background-color": "#FF0000"
}
```

**Response (200):** Updated note object

**Errors:**
- 400: Bad request
- 401: Unauthorized
- 404: Canvas or note not found
- 500: Server error

---

### `DELETE /api/v1/canvases/{canvas-id}/notes/{note-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Delete a note.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |
| note-id | string (UUID) | Note identifier |

**Response (204):** No content (success)

**Errors:**
- 401: Unauthorized
- 404: Canvas or note not found
- 500: Server error

---

## Images

### `GET /api/v1/canvases/{canvas-id}/images`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all images on the canvas.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Response (200):** JSON array of image objects

```json
[
  {
    "image-id": "uuid",
    "original-filename": "photo.jpg",
    "title": "My Photo",
    "asset-hash": "public-hash-hex",
    "asset-hash-private": "private-hash-hex",
    "mime-type": "image/jpeg",
    "file-size": 1024000,
    "location": {"x": 100, "y": 200},
    "depth": 5.0,
    "size": {"width": 800, "height": 600},
    "scale": 1.0,
    "is-pinned": false
  }
]
```

**Errors:**
- 404: Canvas not found
- 500: Server error

---

### `GET /api/v1/canvases/{canvas-id}/images/{image-id}`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single image by ID.

**Response (200):** Single image object

**Errors:**
- 404: Canvas or image not found
- 500: Server error

---

### `POST /api/v1/canvases/{canvas-id}/images`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a new image widget by uploading a file.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Request body:** `multipart/form-data`:
- `data` (required): Image file (PNG, JPG, WebP, etc.)
- `json` (optional): Metadata object as JSON string

```json
{
  "title": "My Photo",
  "location": {"x": 100, "y": 200},
  "size": {"width": 800, "height": 600},
  "depth": 5.0,
  "is-pinned": false
}
```

**Response (201):** Created image object with image-id and asset-hash

**Errors:**
- 400: Bad request (missing or invalid file)
- 401: Unauthorized
- 404: Canvas not found
- 500: Server error (image processing)

---

### `PATCH /api/v1/canvases/{canvas-id}/images/{image-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update image metadata (location, size, title, etc.). Does NOT replace the image file.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |
| image-id | string (UUID) | Image identifier |

**Request body:**

```json
{
  "title": "Updated title",
  "location": {"x": 150, "y": 250}
}
```

**Response (200):** Updated image object

**Errors:**
- 400: Bad request
- 401: Unauthorized
- 404: Canvas or image not found
- 500: Server error

---

### `GET /api/v1/canvases/{canvas-id}/images/{image-id}/download`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** no  
**Status:** implemented

Download the original image file.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |
| image-id | string (UUID) | Image identifier |

**Response (200):** Binary file with `content-disposition: attachment; filename="..."` header

**Cache Control:** Immutable assets are cached as `cache-control: private, max-age=157680000, immutable`

**Errors:**
- 404: Canvas or image not found
- 500: Server error

---

### `DELETE /api/v1/canvases/{canvas-id}/images/{image-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Delete an image widget and its asset.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |
| image-id | string (UUID) | Image identifier |

**Response (204):** No content (success)

**Errors:**
- 401: Unauthorized
- 404: Canvas or image not found
- 500: Server error

---

## Videos

### `GET /api/v1/canvases/{canvas-id}/videos`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all video widgets on the canvas.

**Response (200):** JSON array of video objects

```json
[
  {
    "video-id": "uuid",
    "original-filename": "video.mp4",
    "title": "My Video",
    "asset-hash": "public-hash",
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
]
```

**Errors:**
- 404: Canvas not found
- 500: Server error

---

### `GET /api/v1/canvases/{canvas-id}/videos/{video-id}`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single video widget.

**Response (200):** Single video object

---

### `POST /api/v1/canvases/{canvas-id}/videos`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a new video widget by uploading a file.

**Request body:** `multipart/form-data`:
- `data` (required): Video file (MP4, WebM, etc.)
- `json` (optional): Metadata (title, location, size, etc.)

**Response (201):** Created video object with video-id

---

### `PATCH /api/v1/canvases/{canvas-id}/videos/{video-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update video metadata (seek position, playback state, location, etc.).

**Request body:**

```json
{
  "seek-position": 0.75,
  "playback-state": "paused",
  "muted": true
}
```

**Response (200):** Updated video object

---

### `GET /api/v1/canvases/{canvas-id}/videos/{video-id}/download`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** no  
**Status:** implemented

Download the original video file.

**Response (200):** Binary file with attachment header and immutable cache control

---

### `DELETE /api/v1/canvases/{canvas-id}/videos/{video-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Delete a video widget.

**Response (204):** No content (success)

---

## PDFs

### `GET /api/v1/canvases/{canvas-id}/pdfs`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all PDF widgets on the canvas.

**Response (200):** JSON array of PDF objects

```json
[
  {
    "pdf-id": "uuid",
    "original-filename": "document.pdf",
    "title": "My Document",
    "asset-hash": "public-hash",
    "index": 1,
    "location": {"x": 100, "y": 200},
    "depth": 5.0,
    "size": {"width": 600, "height": 800},
    "scale": 1.0,
    "is-pinned": false
  }
]
```

| Field | Type | Description |
|---|---|---|
| index | integer | Current page number (1-based) |

---

### `GET /api/v1/canvases/{canvas-id}/pdfs/{pdf-id}`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single PDF widget.

---

### `POST /api/v1/canvases/{canvas-id}/pdfs`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a new PDF widget by uploading a file.

**Request body:** `multipart/form-data`:
- `data` (required): PDF file
- `json` (optional): Metadata

**Response (201):** Created PDF object with pdf-id

---

### `PATCH /api/v1/canvases/{canvas-id}/pdfs/{pdf-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update PDF metadata (page index, location, size, etc.).

**Request body:**

```json
{
  "index": 5,
  "location": {"x": 200, "y": 300}
}
```

---

### `GET /api/v1/canvases/{canvas-id}/pdfs/{pdf-id}/download`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** no  
**Status:** implemented

Download the original PDF file.

---

### `DELETE /api/v1/canvases/{canvas-id}/pdfs/{pdf-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Delete a PDF widget.

---

## Browsers

### `GET /api/v1/canvases/{canvas-id}/browsers`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all browser widgets (embedded web content).

**Response (200):** JSON array of browser objects

```json
[
  {
    "browser-id": "uuid",
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
]
```

---

### `GET /api/v1/canvases/{canvas-id}/browsers/{browser-id}`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single browser widget.

---

### `POST /api/v1/canvases/{canvas-id}/browsers`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a new browser widget.

**Request body:**

```json
{
  "source": "https://example.com",
  "title": "Example",
  "transparent-mode": false,
  "location": {"x": 100, "y": 200},
  "size": {"width": 1024, "height": 768}
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| source | string (URL) | yes | Web URL to embed |
| title | string | no | Widget title |
| transparent-mode | boolean | no | Render background as transparent |
| location | object | no | Position {x, y} |
| size | object | no | Dimensions {width, height} |

**Response (201):** Created browser object

---

### `PATCH /api/v1/canvases/{canvas-id}/browsers/{browser-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update browser widget properties.

**Request body:**

```json
{
  "source": "https://newsite.com",
  "main-frame-scroll-offset": {"x": 100, "y": 200}
}
```

---

### `DELETE /api/v1/canvases/{canvas-id}/browsers/{browser-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Delete a browser widget.

---

## Anchors

### `GET /api/v1/canvases/{canvas-id}/anchors`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all anchors (navigation points) on the canvas.

**Response (200):** JSON array of anchor objects

```json
[
  {
    "anchor-id": "uuid",
    "anchor-name": "Anchor 1",
    "location": {"x": 100, "y": 200},
    "depth": 5.0,
    "size": {"width": 0, "height": 0},
    "scale": 1.0,
    "is-pinned": false
  }
]
```

---

### `GET /api/v1/canvases/{canvas-id}/anchors/{anchor-id}`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single anchor.

---

### `POST /api/v1/canvases/{canvas-id}/anchors`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a new anchor.

**Request body:**

```json
{
  "anchor-name": "Anchor 1",
  "location": {"x": 100, "y": 200}
}
```

**Response (201):** Created anchor object

---

### `PATCH /api/v1/canvases/{canvas-id}/anchors/{anchor-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update anchor properties.

---

### `DELETE /api/v1/canvases/{canvas-id}/anchors/{anchor-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Delete an anchor.

---

## Connectors

### `GET /api/v1/canvases/{canvas-id}/connectors`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all connector lines on the canvas.

**Response (200):** JSON array of connector objects

```json
[
  {
    "connector-id": "uuid",
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
]
```

---

### `GET /api/v1/canvases/{canvas-id}/connectors/{connector-id}`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single connector.

---

### `POST /api/v1/canvases/{canvas-id}/connectors`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a new connector between two widgets.

**Request body:**

```json
{
  "connector-type": "arrow",
  "src": "widget-id-1",
  "src-auto-location": true,
  "src-tip": "none",
  "dst": "widget-id-2",
  "dst-auto-location": true,
  "dst-tip": "arrow",
  "line-color": "#FF0000",
  "line-width": 2.0
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| src | string (widget-id) | yes | Source widget ID |
| dst | string (widget-id) | yes | Destination widget ID |
| connector-type | string | no | line, curve, or arrow |
| src-tip | string | no | none, arrow, or circle |
| dst-tip | string | no | none, arrow, or circle |

**Response (201):** Created connector object

---

### `PATCH /api/v1/canvases/{canvas-id}/connectors/{connector-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update connector properties.

---

### `DELETE /api/v1/canvases/{canvas-id}/connectors/{connector-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Delete a connector.

---

## Tables

### `GET /api/v1/canvases/{canvas-id}/tables`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all tables on the canvas.

**Response (200):** JSON array of table objects

```json
[
  {
    "table-id": "uuid",
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
]
```

---

### `GET /api/v1/canvases/{canvas-id}/tables/{table-id}`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single table.

---

### `POST /api/v1/canvases/{canvas-id}/tables`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a new table.

**Request body:**

```json
{
  "title": "Data Table",
  "grid-size": {"columns": 5, "rows": 10},
  "location": {"x": 100, "y": 200}
}
```

---

### `PATCH /api/v1/canvases/{canvas-id}/tables/{table-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update table properties.

---

### `GET /api/v1/canvases/{canvas-id}/tables/{table-id}/cells`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all cells in the table.

**Response (200):** JSON array of cell objects with 2D indices

```json
[
  {
    "cell-id": "uuid",
    "index": [0, 0],
    "content": "Cell content"
  }
]
```

---

### `DELETE /api/v1/canvases/{canvas-id}/tables/{table-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Delete a table.

---

## Video Inputs (Canvas Widgets)

### `GET /api/v1/canvases/{canvas-id}/video-inputs`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all video input widgets (capture sources).

**Response (200):** JSON array of video input objects

```json
[
  {
    "widget-id": "uuid",
    "source": "video-device-id",
    "name": "Webcam",
    "resolution": "1920x1080",
    "location": {"x": 100, "y": 200},
    "depth": 5.0,
    "size": {"width": 960, "height": 540},
    "scale": 1.0,
    "is-pinned": false
  }
]
```

---

### `GET /api/v1/canvases/{canvas-id}/video-inputs/{widget-id}`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single video input widget.

---

### `POST /api/v1/canvases/{canvas-id}/video-inputs`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a new video input widget.

**Request body:**

```json
{
  "source": "video-device-id",
  "name": "Webcam",
  "location": {"x": 100, "y": 200}
}
```

---

### `PATCH /api/v1/canvases/{canvas-id}/video-inputs/{widget-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update video input properties.

---

### `DELETE /api/v1/canvases/{canvas-id}/video-inputs/{widget-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Delete a video input widget.

---

## IP Videos (Canvas Widgets)

### `GET /api/v1/canvases/{canvas-id}/ip-videos`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all IP video stream widgets.

---

### `GET /api/v1/canvases/{canvas-id}/ip-videos/{widget-id}`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single IP video widget.

---

### `POST /api/v1/canvases/{canvas-id}/ip-videos`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a new IP video widget.

---

### `PATCH /api/v1/canvases/{canvas-id}/ip-videos/{widget-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update IP video properties.

---

### `DELETE /api/v1/canvases/{canvas-id}/ip-videos/{widget-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Delete an IP video widget.

---

## RDP Connections (Canvas Widgets)

### `GET /api/v1/canvases/{canvas-id}/rdp-connections`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all RDP connection widgets.

**Response (200):** JSON array of RDP connection objects

```json
[
  {
    "widget-id": "uuid",
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
]
```

---

### `GET /api/v1/canvases/{canvas-id}/rdp-connections/{widget-id}`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single RDP connection widget.

---

### `POST /api/v1/canvases/{canvas-id}/rdp-connections`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a new RDP connection widget.

**Request body:**

```json
{
  "connection-name": "Server 1",
  "host-site": "192.168.1.100",
  "title": "RDP Display",
  "location": {"x": 100, "y": 200}
}
```

---

### `PATCH /api/v1/canvases/{canvas-id}/rdp-connections/{widget-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update RDP connection properties.

---

### `DELETE /api/v1/canvases/{canvas-id}/rdp-connections/{widget-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Delete an RDP connection widget.

---

## Uploads Folder

The uploads folder is a special widget container used to temporarily stage uploaded files before conversion to canvas widgets.

### `GET /api/v1/canvases/{canvas-id}/uploads-folder`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all items in the canvas uploads folder.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Response (200):** JSON array of upload items

---

### `POST /api/v1/canvases/{canvas-id}/uploads-folder`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Upload a file to the canvas uploads folder, optionally with conversion to a canvas widget.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Request body:** `multipart/form-data`:
- `data` (optional): File to upload (omit for notes-only)
- `json` (optional): Metadata object with `upload_type` field

```json
{
  "upload_type": "Note|Image|Video|PDF",
  "title": "Item Title",
  "location": {"x": 100, "y": 200}
}
```

**Response (201):** Created upload item or converted widget object

**Errors:**
- 400: Bad request (invalid upload_type or file)
- 401: Unauthorized
- 404: Canvas not found
- 500: Server error
