# Widgets Examples

## Note Widget Examples

### Example: Create a sticky note

**Source:** mt-restapi-tests/tests/notes_test.go:7 (TestNoteLifecycle)

**Request:**
```http
POST /api/v1/canvases/canvas-abc123/notes HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "text": "E2E test note content",
  "background-color": "#FFFF00"
}
```

**Response (200-299):**
```json
{
  "id": "note-xyz789",
  "type": "note",
  "text": "E2E test note content",
  "background-color": "#FFFF00",
  "position": {"x": 100, "y": 100},
  "size": {"width": 200, "height": 100},
  "created-at": "2024-03-16T15:45:00Z",
  "updated-at": "2024-03-16T15:45:00Z"
}
```

**Notes:** 
- `background-color` is optional and defaults to a server-configured color
- Coordinates and dimensions may be auto-assigned by the server
- Both `position` and `size` objects contain x/y and width/height fields

---

### Example: List all widgets on a canvas

**Source:** mt-restapi-tests/tests/widgets_test.go:7 (TestWidgetsList)

**Request:**
```http
GET /api/v1/canvases/canvas-abc123/widgets HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```json
[
  {
    "id": "note-xyz789",
    "type": "note",
    "text": "First note",
    "background-color": "#FFFF00"
  },
  {
    "id": "image-abc123",
    "type": "image",
    "title": "Product Screenshot"
  },
  {
    "id": "video-def456",
    "type": "video",
    "title": "Demo Video"
  }
]
```

**Notes:** Returns all widgets regardless of type. Each widget has a `type` field indicating its resource-specific endpoint.

---

### Example: Get a specific widget

**Source:** mt-restapi-tests/tests/widgets_test.go:30 (TestWidgetsList)

**Request:**
```http
GET /api/v1/canvases/canvas-abc123/widgets/note-xyz789 HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```json
{
  "id": "note-xyz789",
  "type": "note",
  "text": "E2E test note content",
  "background-color": "#FFFF00",
  "text-color": "#000000",
  "auto-text-color": false,
  "position": {"x": 100, "y": 100},
  "size": {"width": 200, "height": 100},
  "created-at": "2024-03-16T15:45:00Z",
  "updated-at": "2024-03-16T15:45:00Z"
}
```

**Notes:** Fetching by ID returns full widget details.

---

### Example: List all notes on a canvas

**Source:** mt-restapi-tests/tests/notes_test.go:25 (TestNoteLifecycle)

**Request:**
```http
GET /api/v1/canvases/canvas-abc123/notes HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```json
[
  {
    "id": "note-xyz789",
    "type": "note",
    "text": "First note",
    "background-color": "#FFFF00"
  },
  {
    "id": "note-abc123",
    "type": "note",
    "text": "Second note",
    "background-color": "#FF0000"
  }
]
```

**Notes:** Returns only note widgets from the canvas.

---

### Example: Get a specific note

**Source:** mt-restapi-tests/tests/notes_test.go:42 (TestNoteLifecycle)

**Request:**
```http
GET /api/v1/canvases/canvas-abc123/notes/note-xyz789 HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```json
{
  "id": "note-xyz789",
  "type": "note",
  "text": "E2E test note content",
  "background-color": "#FFFF00",
  "text-color": "#000000",
  "auto-text-color": false,
  "position": {"x": 100, "y": 100},
  "size": {"width": 200, "height": 100},
  "created-at": "2024-03-16T15:45:00Z",
  "updated-at": "2024-03-16T15:45:00Z"
}
```

**Notes:** Resource-specific GET returns full details.

---

### Example: Update a note

**Source:** mt-restapi-tests/tests/notes_test.go:50 (TestNoteLifecycle)

**Request:**
```http
PATCH /api/v1/canvases/canvas-abc123/notes/note-xyz789 HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "text": "E2E updated note"
}
```

**Response (200):**
```json
{
  "id": "note-xyz789",
  "type": "note",
  "text": "E2E updated note",
  "background-color": "#FFFF00",
  "updated-at": "2024-03-16T15:50:00Z"
}
```

**Notes:** PATCH allows partial updates. Only provided fields are modified.

---

### Example: Delete a note

**Source:** mt-restapi-tests/tests/notes_test.go:59 (TestNoteLifecycle)

**Request:**
```http
DELETE /api/v1/canvases/canvas-abc123/notes/note-xyz789 HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```
(empty body)
```

**Verification (subsequent GET):**
```http
GET /api/v1/canvases/canvas-abc123/notes/note-xyz789 HTTP/1.1
Authorization: Bearer <token>
```

**Response (404):**
```json
{
  "error": "widget not found"
}
```

**Notes:** Deletion is permanent.

---

## Widget Error Handling

### Example: Text color conflict - auto-text-color enabled

**Source:** mt-restapi-tests/tests/notes_test.go:76 (TestNoteAutoTextColorConflict)

**Precondition:** Note created with `auto-text-color: true`

**Request:**
```http
PATCH /api/v1/canvases/canvas-abc123/notes/note-xyz789 HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "text-color": "#FF0000"
}
```

**Response (409 or other 4xx):**
```json
{
  "error": "cannot set text-color while auto-text-color is enabled"
}
```

**Notes:** Setting explicit text color conflicts with automatic color determination. The server typically rejects this with 409 Conflict.

---

### Example: Clone endpoint (not implemented)

**Source:** mt-restapi-tests/tests/widgets_test.go:40 (TestWidgetsClone501)

**Request:**
```http
POST /api/v1/canvases/canvas-abc123/widgets/clone HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{}
```

**Response (501):**
```json
{
  "error": "not implemented"
}
```

**Notes:** The generic `/widgets/clone` endpoint returns 501 Not Implemented. Cloning is available at resource-specific levels (e.g., canvases can be copied, but individual widgets cannot be cloned via this endpoint).
