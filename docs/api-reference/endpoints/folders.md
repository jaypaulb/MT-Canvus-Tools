# Canvas Folders Endpoints

Canvas folders organize canvases into a hierarchical folder tree. The root folder is the server index, and canvases can be nested in folders.

## Folder Operations

### `GET /api/v1/canvas-folders`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all canvas folders in the server hierarchy.

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| subscribe | boolean | no | false | Enable streaming updates |

**Response (200):** JSON array of folder objects

```json
[
  {
    "folder-id": "uuid",
    "folder-name": "Projects",
    "parent-folder-id": "root-uuid",
    "owner": "user@example.com",
    "created": "2025-01-15T10:30:00Z",
    "modified": "2025-05-17T14:22:15Z",
    "permission-overrides": [
      {
        "user-id": "user-uuid",
        "permission": "view|edit|own"
      }
    ],
    "children": [
      {
        "type": "folder|canvas",
        "id": "item-uuid",
        "name": "Subfolder or Canvas"
      }
    ]
  }
]
```

**Errors:**
- 401: Unauthorized
- 500: Server error

---

### `GET /api/v1/canvas-folders/{folder-id}`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single folder by ID.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| folder-id | string (UUID) | Folder identifier |

**Response (200):** Single folder object

**Errors:**
- 404: Folder not found
- 500: Server error

---

### `POST /api/v1/canvas-folders`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a new canvas folder.

**Request body:**

```json
{
  "folder-name": "Q2 2025 Projects",
  "parent-folder-id": "parent-uuid"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| folder-name | string | yes | Name for the new folder |
| parent-folder-id | string (UUID) | no | Parent folder ID (default: root folder) |

**Response (201):** Created folder object

**Errors:**
- 400: Bad request (missing folder-name or invalid parent)
- 401: Unauthorized
- 404: Parent folder not found
- 500: Server error

---

### `PATCH /api/v1/canvas-folders/{folder-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Rename a canvas folder.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| folder-id | string (UUID) | Folder identifier |

**Request body:**

```json
{
  "folder-name": "Q3 2025 Projects"
}
```

**Response (200):** Updated folder object

**Errors:**
- 400: Bad request
- 401: Unauthorized
- 404: Folder not found
- 500: Server error

---

### `DELETE /api/v1/canvas-folders/{folder-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Delete an empty folder. Folder must have no children (canvases or subfolders).

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| folder-id | string (UUID) | Folder identifier |

**Response (204):** No content (folder deleted)

**Errors:**
- 401: Unauthorized
- 404: Folder not found
- 409: Conflict (folder not empty; has children)
- 500: Server error

---

### `DELETE /api/v1/canvas-folders/{folder-id}/children`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Delete all children of a folder (all nested canvases and subfolders).

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| folder-id | string (UUID) | Folder identifier |

**Response (204):** No content (all children deleted)

**Errors:**
- 401: Unauthorized
- 404: Folder not found
- 500: Server error

---

## Folder Movement

### `POST /api/v1/canvas-folders/{folder-id}/move`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Move a folder to a different parent folder (POST method).

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| folder-id | string (UUID) | Folder identifier |

**Request body:**

```json
{
  "parent-folder-id": "new-parent-uuid"
}
```

**Response (200):** Updated folder object with new parent reference

**Errors:**
- 400: Bad request (invalid parent)
- 401: Unauthorized
- 404: Folder or parent not found
- 409: Conflict (cannot move folder to its own child)
- 500: Server error

---

### `PATCH /api/v1/canvas-folders/{folder-id}/move`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Move a folder to a different parent folder (PATCH method, alternative to POST).

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| folder-id | string (UUID) | Folder identifier |

**Request body:**

```json
{
  "parent-folder-id": "new-parent-uuid"
}
```

**Response (200):** Updated folder object

**Errors:** (same as POST /move)

---

## Folder Copy

### `POST /api/v1/canvas-folders/{folder-id}/copy`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a copy of a folder with all nested canvases and subfolders (POST method).

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| folder-id | string (UUID) | Folder identifier |

**Request body:**

```json
{
  "folder-name": "Copy of Projects",
  "parent-folder-id": "destination-parent-uuid"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| folder-name | string | yes | Name for the copied folder |
| parent-folder-id | string (UUID) | no | Destination parent folder (default: root) |

**Response (201):** Newly created folder copy with new folder-id

**Errors:**
- 400: Bad request (missing folder-name)
- 401: Unauthorized
- 404: Source folder or parent not found
- 500: Server error

---

### `PATCH /api/v1/canvas-folders/{folder-id}/copy`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a copy of a folder (PATCH method, alternative to POST).

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| folder-id | string (UUID) | Folder identifier |

**Request body:**

```json
{
  "folder-name": "Copy of Projects",
  "parent-folder-id": "destination-parent-uuid"
}
```

**Response (201):** Newly created folder copy

**Errors:** (same as POST /copy)

---

## Folder Permissions

### `GET /api/v1/canvas-folders/{folder-id}/permissions`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve permission settings for a folder including per-user/group access levels.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| folder-id | string (UUID) | Folder identifier |

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| subscribe | boolean | no | false | Enable streaming updates |

**Response (200):**

```json
{
  "folder-id": "uuid",
  "default-permission": "view|edit|own",
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

| Field | Type | Description |
|---|---|---|
| default-permission | string | Default permission level for all users not in overrides |
| permission-overrides | array | Per-user and per-group permission overrides |

**Errors:**
- 401: Unauthorized
- 404: Folder not found
- 500: Server error

---

### `POST /api/v1/canvas-folders/{folder-id}/permissions`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Set or update folder permissions including default level and per-user/group overrides.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| folder-id | string (UUID) | Folder identifier |

**Request body:**

```json
{
  "default-permission": "view",
  "permission-overrides": [
    {
      "user-id": "user-uuid",
      "permission": "edit"
    },
    {
      "group-id": "group-uuid",
      "permission": "view"
    }
  ]
}
```

| Field | Type | Description |
|---|---|---|
| default-permission | string | Default permission: view, edit, own, or none |
| permission-overrides | array | List of user/group-specific overrides |

**Response (200):** Updated permissions object

**Errors:**
- 400: Bad request (invalid permission values)
- 401: Unauthorized
- 404: Folder not found
- 500: Server error

---

## Permission Levels

Permission levels control what users can do with a folder and its contents:

| Level | Description | Can View | Can Edit | Can Delete | Can Change Permissions |
|-------|-------------|----------|----------|------------|----------------------|
| **none** | No access | No | No | No | No |
| **view** | Read-only access | Yes | No | No | No |
| **edit** | Read-write access | Yes | Yes | Yes (own items) | No |
| **own** | Full control | Yes | Yes | Yes (all) | Yes |

Permissions are inherited: child canvases and folders inherit permissions from parent folders unless overridden.
