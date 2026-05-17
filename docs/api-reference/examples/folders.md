# Canvas Folders Examples

## Folder Lifecycle

### Example: Create a folder

**Source:** mt-restapi-tests/tests/canvas_folders_test.go:13 (TestCanvasFolderLifecycle)

**Request:**
```http
POST /api/v1/canvas-folders HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "folder-name": "My Projects"
}
```

**Response (200-299):**
```json
{
  "id": "folder-abc123",
  "folder-name": "My Projects",
  "parent-id": null,
  "created-at": "2024-03-16T16:30:00Z",
  "updated-at": "2024-03-16T16:30:00Z"
}
```

**Notes:** 
- Minimal request requires only `folder-name`
- `parent-id` defaults to root if omitted
- Returns generated folder ID

---

### Example: List all folders

**Source:** mt-restapi-tests/tests/canvas_folders_test.go:31 (TestCanvasFolderLifecycle)

**Request:**
```http
GET /api/v1/canvas-folders HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```json
[
  {
    "id": "folder-abc123",
    "folder-name": "My Projects",
    "parent-id": null
  },
  {
    "id": "folder-def456",
    "folder-name": "Client Work",
    "parent-id": null
  },
  {
    "id": "folder-xyz789",
    "folder-name": "Q1 Deliverables",
    "parent-id": "folder-abc123"
  }
]
```

**Notes:** 
- Returns all accessible folders (user's own and shared)
- Hierarchical structure shown via `parent-id` field

---

### Example: Get folder details

**Source:** mt-restapi-tests/tests/canvas_folders_test.go:48 (TestCanvasFolderLifecycle)

**Request:**
```http
GET /api/v1/canvas-folders/folder-abc123 HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```json
{
  "id": "folder-abc123",
  "folder-name": "My Projects",
  "parent-id": null,
  "children": [
    {
      "id": "folder-xyz789",
      "folder-name": "Q1 Deliverables",
      "type": "folder"
    },
    {
      "id": "canvas-proj-001",
      "canvas-name": "Website Redesign",
      "type": "canvas"
    }
  ],
  "created-at": "2024-03-16T16:30:00Z",
  "updated-at": "2024-03-16T16:30:00Z"
}
```

**Notes:** 
- Includes nested structure if applicable
- Children array contains both subfolders and canvases

---

### Example: Update folder name

**Source:** mt-restapi-tests/tests/canvas_folders_test.go:56 (TestCanvasFolderLifecycle)

**Request:**
```http
PATCH /api/v1/canvas-folders/folder-abc123 HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "folder-name": "Active Projects 2024"
}
```

**Response (200):**
```json
{
  "id": "folder-abc123",
  "folder-name": "Active Projects 2024",
  "parent-id": null,
  "updated-at": "2024-03-16T16:35:00Z"
}
```

**Notes:** PATCH allows partial updates.

---

## Folder Permissions

### Example: Get folder permissions

**Source:** mt-restapi-tests/tests/canvas_folders_test.go:65 (TestCanvasFolderLifecycle)

**Request:**
```http
GET /api/v1/canvas-folders/folder-abc123/permissions HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```json
{
  "owner": "user-abc123",
  "shared-with": [
    {
      "user-id": "user-def456",
      "permission": "view"
    },
    {
      "user-id": "user-xyz789",
      "permission": "edit"
    }
  ],
  "link-permission": "none",
  "public": false
}
```

**Notes:** Returns sharing settings and access levels for the folder.

---

### Example: Set folder permissions

**Source:** mt-restapi-tests/tests/canvas_folders_test.go:73 (TestCanvasFolderLifecycle)

**Request:**
```http
POST /api/v1/canvas-folders/folder-abc123/permissions HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{}
```

**Response (200 or other 2xx):**
```
(status varies; may return updated permissions or empty body)
```

**Notes:** 
- Endpoint exists but payload handling is server-specific
- Typically used to update sharing settings

---

## Folder Organization

### Example: Move folder to a parent folder

**Source:** mt-restapi-tests/tests/canvas_folders_test.go:111 (TestCanvasFolderMoveAndCopy)

**Request:**
```http
POST /api/v1/canvas-folders/folder-abc123/move HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "folder-id": "folder-parent-001"
}
```

**Response (200 or other 2xx):**
```
(empty or status confirmation)
```

**Notes:** 
- Moves folder to a new parent location
- Omitting or passing empty `folder-id` moves to root

---

### Example: Move folder using PATCH (alternative)

**Source:** mt-restapi-tests/tests/canvas_folders_test.go:125 (TestCanvasFolderMoveAndCopy)

**Request:**
```http
PATCH /api/v1/canvas-folders/folder-abc123/move HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "folder-id": "folder-parent-001"
}
```

**Response (2xx):**
```
(status varies)
```

**Notes:** 
- Both POST and PATCH are accepted for move operations
- Behavior is identical between the two methods

---

### Example: Copy a folder

**Source:** mt-restapi-tests/tests/canvas_folders_test.go:134 (TestCanvasFolderMoveAndCopy)

**Request:**
```http
POST /api/v1/canvas-folders/folder-abc123/copy HTTP/1.1
Authorization: Bearer <token>
```

**Response (2xx):**
```json
{
  "id": "folder-copy-001",
  "folder-name": "My Projects (Copy)",
  "created-at": "2024-03-16T16:40:00Z"
}
```

**Notes:** 
- Creates a deep copy of the folder and its contents
- Copy has a new ID and is owned by the requesting user

---

### Example: Copy folder using PATCH (alternative)

**Source:** mt-restapi-tests/tests/canvas_folders_test.go:148 (TestCanvasFolderMoveAndCopy)

**Request:**
```http
PATCH /api/v1/canvas-folders/folder-abc123/copy HTTP/1.1
Authorization: Bearer <token>
```

**Response (2xx):**
```json
{
  "id": "folder-copy-002",
  "folder-name": "My Projects (Copy)",
  "created-at": "2024-03-16T16:45:00Z"
}
```

**Notes:** 
- Both POST and PATCH are accepted for copy operations
- Each invocation creates a new copy

---

## Folder Deletion

### Example: Delete folder children

**Source:** mt-restapi-tests/tests/canvas_folders_test.go:191 (TestCanvasFolderDeleteChildren)

**Request:**
```http
DELETE /api/v1/canvas-folders/folder-abc123/children HTTP/1.1
Authorization: Bearer <token>
```

**Response (2xx):**
```
(status varies)
```

**Notes:** 
- Deletes all contents (subfolders and canvases) within the folder
- Use with caution: this is a destructive operation
- Some servers may refuse this operation for safety

---

### Example: Delete an empty folder

**Source:** mt-restapi-tests/tests/canvas_folders_test.go:198 (TestCanvasFolderDeleteChildren)

**Request:**
```http
DELETE /api/v1/canvas-folders/folder-abc123 HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```
(empty body)
```

**Verification (subsequent GET):**
```http
GET /api/v1/canvas-folders/folder-abc123 HTTP/1.1
Authorization: Bearer <token>
```

**Response (404):**
```json
{
  "error": "folder not found"
}
```

**Notes:** 
- Folder must be empty (or deletion cascades based on server policy)
- Deletion is permanent

---

## Folder with Nested Structure

### Example: Create a subfolder

**Source:** mt-restapi-tests/tests/canvas_folders_test.go:180 (TestCanvasFolderDeleteChildren)

**Request:**
```http
POST /api/v1/canvas-folders HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "folder-name": "Q1 Deliverables",
  "parent-id": "folder-abc123"
}
```

**Response (200-299):**
```json
{
  "id": "folder-xyz789",
  "folder-name": "Q1 Deliverables",
  "parent-id": "folder-abc123",
  "created-at": "2024-03-16T16:50:00Z"
}
```

**Notes:** 
- `parent-id` specifies the parent folder
- Creates a hierarchical folder structure
- Subfolders inherit some permissions from parent (server-specific)
