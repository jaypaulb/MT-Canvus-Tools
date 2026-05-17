# Canvas Examples

## Canvas Lifecycle Examples

### Example: List all canvases

**Source:** mt-restapi-tests/tests/canvases_test.go:11 (TestCanvasesList)

**Request:**
```http
GET /api/v1/canvases HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```json
[
  {
    "id": "canvas-abc123",
    "canvas-name": "My First Canvas",
    "name": "My First Canvas",
    "created-at": "2024-03-15T10:30:00Z",
    "updated-at": "2024-03-15T14:45:00Z"
  },
  {
    "id": "canvas-def456",
    "canvas-name": "Collaboration Board",
    "name": "Collaboration Board",
    "created-at": "2024-03-10T08:00:00Z",
    "updated-at": "2024-03-16T12:15:00Z"
  }
]
```

**Notes:** Returns a list of all canvases accessible to the authenticated user. Both `canvas-name` and `name` fields may be present in the response.

---

### Example: Get canvas details

**Source:** mt-restapi-tests/tests/canvases_test.go:32 (TestCanvasesGet)

**Request:**
```http
GET /api/v1/canvases/canvas-abc123 HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```json
{
  "id": "canvas-abc123",
  "canvas-name": "My First Canvas",
  "name": "My First Canvas",
  "width": 3840,
  "height": 2160,
  "created-at": "2024-03-15T10:30:00Z",
  "updated-at": "2024-03-15T14:45:00Z"
}
```

**Notes:** Response includes canvas dimensions and timestamps.

---

### Example: Create a basic canvas

**Source:** mt-restapi-tests/tests/canvases_test.go:50 (TestCanvasesCreate)

**Request:**
```http
POST /api/v1/canvases HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "canvas-name": "My New Canvas"
}
```

**Response (200-299):**
```json
{
  "id": "canvas-xyz789",
  "canvas-name": "My New Canvas",
  "name": "My New Canvas",
  "created-at": "2024-03-16T15:20:00Z",
  "updated-at": "2024-03-16T15:20:00Z"
}
```

**Notes:** Minimal request requires only `canvas-name`. Response includes generated ID and timestamps.

---

### Example: Create canvas with custom dimensions

**Source:** mt-restapi-tests/tests/canvases_test.go:70 (TestCanvasesCreateWithDimensions)

**Request:**
```http
POST /api/v1/canvases HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "canvas-name": "Custom Canvas",
  "new-width": 3840,
  "new-height": 2160
}
```

**Response (200-299):**
```json
{
  "id": "canvas-custom-001",
  "canvas-name": "Custom Canvas",
  "width": 3840,
  "height": 2160,
  "created-at": "2024-03-16T15:25:00Z",
  "updated-at": "2024-03-16T15:25:00Z"
}
```

**Notes:** Dimensions are in pixels. Omitting these parameters uses server defaults.

---

### Example: Create canvas with password protection

**Source:** mt-restapi-tests/tests/canvases_test.go:92 (TestCanvasesCreateWithPasswords)

**Request:**
```http
POST /api/v1/canvases HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "canvas-name": "Protected Canvas",
  "initial-main-password": "MainPass123",
  "initial-access-password": "AccessPass123"
}
```

**Response (200-299):**
```json
{
  "id": "canvas-protected-001",
  "canvas-name": "Protected Canvas",
  "created-at": "2024-03-16T15:30:00Z",
  "updated-at": "2024-03-16T15:30:00Z"
}
```

**Notes:** Two levels of password protection: main access and view-only access. Both are optional.

---

### Example: Update canvas name

**Source:** mt-restapi-tests/tests/canvases_test.go:114 (TestCanvasesUpdate)

**Request:**
```http
PATCH /api/v1/canvases/canvas-abc123 HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "canvas-name": "Renamed Canvas"
}
```

**Response (200):**
```json
{
  "id": "canvas-abc123",
  "canvas-name": "Renamed Canvas",
  "name": "Renamed Canvas",
  "updated-at": "2024-03-16T15:35:00Z"
}
```

**Notes:** PATCH allows partial updates. Only provided fields are modified.

---

### Example: Delete a canvas

**Source:** mt-restapi-tests/tests/canvases_test.go:139 (TestCanvasesDelete)

**Request:**
```http
DELETE /api/v1/canvases/canvas-abc123 HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```
(empty body)
```

**Verification (subsequent GET):**
```http
GET /api/v1/canvases/canvas-abc123 HTTP/1.1
Authorization: Bearer <token>
```

**Response (404):**
```json
{
  "error": "canvas not found"
}
```

**Notes:** Deletion is permanent. The canvas cannot be accessed after deletion.

---

## Canvas Operations

### Example: Get canvas permissions

**Source:** mt-restapi-tests/tests/canvases_test.go:167 (TestCanvasPermissions)

**Request:**
```http
GET /api/v1/canvases/canvas-abc123/permissions HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```json
{
  "link-permission": "view",
  "password-protected": false,
  "public": false
}
```

**Notes:** Returns sharing and access control settings for the canvas.

---

### Example: Set canvas permissions

**Source:** mt-restapi-tests/tests/canvases_test.go:175 (TestCanvasPermissions)

**Request:**
```http
POST /api/v1/canvases/canvas-abc123/permissions HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "link-permission": "view"
}
```

**Response (200 or 204):**
```
(empty body)
```

**Notes:** 
- `link-permission` values: `"view"`, `"edit"`, `"none"`
- Both 200 and 204 are accepted as successful responses.

---

### Example: Move canvas to folder

**Source:** mt-restapi-tests/tests/canvases_test.go:192 (TestCanvasMove)

**Request:**
```http
POST /api/v1/canvases/canvas-abc123/move HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "folder-id": "folder-xyz789"
}
```

**Response (200 or 204):**
```
(empty body)
```

**Notes:** Moves canvas to a specific folder. Omit `folder-id` or pass empty string to move to root.

---

### Example: Copy a canvas

**Source:** mt-restapi-tests/tests/canvases_test.go:218 (TestCanvasCopy)

**Request:**
```http
POST /api/v1/canvases/canvas-abc123/copy HTTP/1.1
Authorization: Bearer <token>
```

**Response (200-299):**
```json
{
  "id": "canvas-copy-001",
  "canvas-name": "My First Canvas (Copy)",
  "created-at": "2024-03-16T15:40:00Z"
}
```

**Notes:** Creates a deep copy of the canvas with all contents. The copy has a new ID and is owned by the requesting user.

---

### Example: Save canvas state

**Source:** mt-restapi-tests/tests/canvases_test.go:249 (TestCanvasSaveRestore)

**Request:**
```http
POST /api/v1/canvases/canvas-abc123/save HTTP/1.1
Authorization: Bearer <token>
```

**Response (2xx):**
```
(status code varies by server configuration)
```

**Notes:** Creates a checkpoint of the current canvas state. Useful for version control or backup purposes.

---

### Example: Restore canvas from saved state

**Source:** mt-restapi-tests/tests/canvases_test.go:256 (TestCanvasSaveRestore)

**Request:**
```http
POST /api/v1/canvases/canvas-abc123/restore HTTP/1.1
Authorization: Bearer <token>
```

**Response (2xx):**
```
(status code varies by server configuration)
```

**Notes:** Reverts canvas to the most recent saved state.

---

### Example: Get canvas preview

**Source:** mt-restapi-tests/tests/canvases_test.go:265 (TestCanvasPreview)

**Request:**
```http
GET /api/v1/canvases/canvas-abc123/preview HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```
(binary image data)
```

**Notes:** Returns a preview image of the canvas. Response is binary image data, not JSON.
