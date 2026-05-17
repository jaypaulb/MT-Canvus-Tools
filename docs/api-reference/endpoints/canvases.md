# Canvas Endpoints

## Canvas Document Management

### `GET /api/v1/canvases`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all canvases visible to the authenticated user. Returns canvas list with full metadata including permissions, size, ownership, and folder hierarchy.

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| subscribe | boolean | no | false | Enable streaming updates (newline-delimited JSON) |

**Response (200):** JSON array of canvas objects

```json
[
  {
    "canvas-id": "uuid",
    "canvas-name": "Project A",
    "owner": "user@example.com",
    "demo-canvas": false,
    "has-main-password": false,
    "has-access-password": false,
    "size": {"width": 3840, "height": 2160},
    "created": "2025-01-15T10:30:00Z",
    "modified": "2025-05-17T14:22:15Z",
    "link-permission": "view|edit|none",
    "permission-overrides": [...]
  }
]
```

**Errors:**
- 401: Unauthorized (missing/invalid token)
- 500: Server error

---

### `GET /api/v1/canvases/{canvas-id}`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single canvas by ID. Accessible without authentication if canvas has link sharing enabled.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| subscribe | boolean | no | false | Enable streaming updates |

**Response (200):** Single canvas object (see GET /canvases format)

**Errors:**
- 401: Unauthorized (canvas requires auth, link sharing disabled)
- 403: Forbidden (user lacks permission)
- 404: Canvas not found
- 500: Server error

---

### `POST /api/v1/canvases`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a new canvas with optional initial dimensions and password protection.

**Request body:**

```json
{
  "canvas-name": "New Project",
  "new-width": 3840,
  "new-height": 2160,
  "initial-main-password": "secret123",
  "initial-access-password": "viewonly456"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| canvas-name | string | yes | Name for the new canvas |
| new-width | integer | no | Canvas width in pixels (default: 3840) |
| new-height | integer | no | Canvas height in pixels (default: 2160) |
| initial-main-password | string | no | Main canvas password (requires password to edit) |
| initial-access-password | string | no | View-only password (requires password to view) |

**Response (201):** Created canvas object with new canvas-id

**Errors:**
- 400: Bad request (missing canvas-name or invalid dimensions)
- 401: Unauthorized
- 500: Server error

---

### `PATCH /api/v1/canvases/{canvas-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Rename a canvas. Only the canvas-name field is currently supported for patching.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Request body:**

```json
{
  "canvas-name": "Updated Name"
}
```

**Response (200):** Updated canvas object

**Errors:**
- 401: Unauthorized
- 403: Forbidden (user lacks write permission)
- 404: Canvas not found
- 500: Server error

---

### `DELETE /api/v1/canvases/{canvas-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Delete a canvas permanently. All widgets and data are removed.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Response (204):** No content (success)

**Errors:**
- 401: Unauthorized
- 403: Forbidden (user lacks permission)
- 404: Canvas not found
- 500: Server error

---

## Canvas Actions

### `POST /api/v1/canvases/{canvas-id}/move`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Move a canvas to a different folder.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Request body:**

```json
{
  "parent-folder-id": "folder-uuid"
}
```

**Response (200):** Updated canvas object with new folder reference

**Errors:**
- 401: Unauthorized
- 403: Forbidden
- 404: Canvas or folder not found
- 500: Server error

---

### `POST /api/v1/canvases/{canvas-id}/copy`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a copy of a canvas with all widgets and assets duplicated.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Request body:**

```json
{
  "canvas-name": "Copy of Original",
  "parent-folder-id": "folder-uuid"
}
```

**Response (201):** Newly created canvas copy with new canvas-id

**Errors:**
- 400: Bad request (missing canvas-name)
- 401: Unauthorized
- 403: Forbidden
- 404: Source canvas or folder not found
- 500: Server error

---

### `POST /api/v1/canvases/{canvas-id}/save`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Save the current canvas state as a demo snapshot. Users can later restore to this state.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Response (200):** Canvas object with demo-snapshot flag updated

**Errors:**
- 401: Unauthorized
- 403: Forbidden (user lacks write permission)
- 404: Canvas not found
- 500: Server error

---

### `POST /api/v1/canvases/{canvas-id}/restore`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Restore a canvas to its previously saved demo state. Overwrites all current widget state.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Response (200):** Canvas object after restore

**Errors:**
- 401: Unauthorized
- 403: Forbidden
- 404: Canvas not found or no demo state saved
- 500: Server error

---

## Canvas Appearance

### `GET /api/v1/canvases/{canvas-id}/background`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** no  
**Status:** implemented

Retrieve canvas background configuration including type, color, image, grid, and haze settings.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Response (200):**

```json
{
  "background-type": "color|image|none",
  "background-color": "#ffffff",
  "background-image": "asset-hash",
  "image-fit": "fill|contain|cover",
  "grid-visible": true,
  "grid-size": 10,
  "haze-visible": true,
  "haze-color": "#000000",
  "haze-opacity": 0.5
}
```

**Errors:**
- 401: Unauthorized (if canvas requires authentication)
- 404: Canvas not found
- 500: Server error

---

### `PATCH /api/v1/canvases/{canvas-id}/background`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update canvas background settings without uploading a new image file.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Request body:**

```json
{
  "background-type": "color",
  "background-color": "#f0f0f0",
  "image-fit": "contain",
  "grid-visible": true,
  "grid-size": 20,
  "haze-visible": false
}
```

All fields are optional; only specified fields are updated.

**Response (200):** Updated background configuration object

**Errors:**
- 400: Bad request (invalid color format or background-type)
- 401: Unauthorized
- 403: Forbidden
- 404: Canvas not found
- 500: Server error

---

### `POST /api/v1/canvases/{canvas-id}/background`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Upload a new background image for the canvas.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Request body:** `multipart/form-data` with file part:
- `data` (required): Background image file (PNG, JPG, WebP, etc.)

**Response (201):** Updated background configuration with new image asset hash

**Errors:**
- 400: Bad request (missing or invalid file)
- 401: Unauthorized
- 403: Forbidden
- 404: Canvas not found
- 500: Server error (image processing failure)

---

### `GET /api/v1/canvases/{canvas-id}/color-presets`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** no  
**Status:** implemented

Retrieve the canvas color presets used for annotations, note backgrounds, and connectors.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Response (200):**

```json
{
  "annotation-colors": ["#FF0000", "#00FF00", "#0000FF"],
  "note-background-colors": ["#FFFF00", "#FF00FF"],
  "note-text-colors": ["#000000", "#FFFFFF"],
  "connector-colors": ["#FF0000", "#00FF00"]
}
```

**Errors:**
- 404: Canvas not found
- 500: Server error

---

### `PATCH /api/v1/canvases/{canvas-id}/color-presets`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update the canvas color presets.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Request body:**

```json
{
  "annotation-colors": ["#FF0000", "#00FF00", "#0000FF"],
  "note-background-colors": ["#FFFF00"],
  "note-text-colors": ["#000000"]
}
```

All fields are optional.

**Response (200):** Updated color presets object

**Errors:**
- 400: Bad request (invalid color format)
- 401: Unauthorized
- 403: Forbidden
- 404: Canvas not found
- 500: Server error

---

## Canvas Preview & Media

### `GET /api/v1/canvases/{canvas-id}/preview`

**Auth:** login-token | api-key | none (if link sharing enabled)  
**Streaming:** no  
**Status:** implemented

Download a preview image of the canvas (generated thumbnail or snapshot).

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Response (200):** Binary file (PNG or JPG) with `content-disposition: attachment` header

**Errors:**
- 401: Unauthorized (if canvas requires authentication)
- 404: Canvas not found or preview not ready
- 500: Server error

---

## Permissions

### `GET /api/v1/canvases/{canvas-id}/permissions`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve permission settings for a canvas including link permission and per-user/group overrides.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Response (200):**

```json
{
  "link-permission": "view|edit|none",
  "permission-overrides": [
    {
      "user-id": "user-uuid",
      "permission": "view|edit|own|none"
    },
    {
      "group-id": "group-uuid",
      "permission": "view|edit|own|none"
    }
  ]
}
```

**Errors:**
- 401: Unauthorized
- 404: Canvas not found
- 500: Server error

---

### `POST /api/v1/canvases/{canvas-id}/permissions`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Set or update canvas permissions including link sharing and per-user/group access levels.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| canvas-id | string (UUID) | Canvas identifier |

**Request body:**

```json
{
  "link-permission": "view",
  "permission-overrides": [
    {
      "user-id": "user-uuid",
      "permission": "edit"
    }
  ]
}
```

**Response (200):** Updated permissions object

**Errors:**
- 400: Bad request (invalid permission values)
- 401: Unauthorized
- 403: Forbidden (user lacks admin permission)
- 404: Canvas not found
- 500: Server error
