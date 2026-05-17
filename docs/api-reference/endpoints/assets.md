# Assets & File Access Endpoints

Assets are files (images, videos, PDFs) uploaded to the Canvus server. They can be accessed directly by hash or through widget endpoints.

## Asset Download by Hash

### `GET /api/v1/assets/{hash}`

**Auth:** login-token | api-key (both `Private-Token` header AND `canvas-id` header required)  
**Streaming:** no  
**Status:** implemented

Download an asset file directly by its public hash. Requires both authentication and canvas context.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| hash | string (hex) | Asset public hash (16-64 character hex string) |

**Headers (Required):**
| Header | Type | Description |
|---|---|---|
| Private-Token | string | API key or session token |
| canvas-id | string (UUID) | UUID of a canvas that references this asset |

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| canvasId | string | no | — | NOT SUPPORTED; use `canvas-id` header instead (returns 404 if used) |

**Response (200):** Binary file with `content-disposition: attachment` header

```
HTTP/1.1 200 OK
Content-Type: image/jpeg
Content-Disposition: attachment; filename="photo.jpg"
Cache-Control: private, max-age=157680000, immutable

<binary file data>
```

**Cache:** Immutable assets are cached as `cache-control: private, max-age=157680000, immutable` (5 years).

**Errors:**
- 400: Bad request (invalid hash format)
- 401: Unauthorized (missing or invalid token)
- 404: Asset not found or canvas-id header missing
- 500: Server error

**Important:** The `canvas-id` **header** is mandatory. The `canvasId` query parameter does NOT work as a substitute and will return 404.

---

## Mipmaps (Tiled Image Access)

Mipmaps are tiled pyramid representations of large images used for efficient rendering. Access requires authentication and canvas context.

### `GET /api/v1/mipmaps/{hash}`

**Auth:** login-token | api-key (both `Private-Token` header AND `canvas-id` header required)  
**Streaming:** no  
**Status:** implemented

Retrieve mipmap metadata (tile pyramid info) for an asset.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| hash | string (hex) | Asset public hash |

**Headers (Required):**
| Header | Type | Description |
|---|---|---|
| Private-Token | string | API key or session token |
| canvas-id | string (UUID) | UUID of a canvas that references this asset |

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| page | integer | no | 0 | Page number (for multi-page assets like PDFs) |

**Response (200):**

```json
{
  "hash": "asset-hash",
  "tile-size": 256,
  "levels": 8,
  "width": 16384,
  "height": 12288,
  "pages": 1,
  "tiles-per-page": 2048
}
```

**Cache:** Immutable assets are cached as `cache-control: private, max-age=157680000, immutable`.

**Errors:**
- 400: Bad request (invalid hash)
- 401: Unauthorized
- 404: Asset not found or canvas-id header missing
- 500: Server error; mipmaps not yet generated (async processing, try again later)

**Note:** If mipmaps have not yet been generated for an asset, the server returns 500. Mipmap generation is asynchronous; retry after a delay.

---

### `GET /api/v1/mipmaps/{hash}/{level}`

**Auth:** login-token | api-key (both `Private-Token` header AND `canvas-id` header required)  
**Streaming:** no  
**Status:** implemented

Download a specific mipmap level (tile data) for an asset.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| hash | string (hex) | Asset public hash |
| level | integer | Mipmap level (0 = highest resolution, increases toward lowest) |

**Headers (Required):**
| Header | Type | Description |
|---|---|---|
| Private-Token | string | API key or session token |
| canvas-id | string (UUID) | UUID of a canvas that references this asset |

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| page | integer | no | 0 | Page number (for multi-page assets) |

**Response (200):** Binary file (mipmap tiles)

```
HTTP/1.1 200 OK
Content-Type: application/octet-stream
Cache-Control: private, max-age=157680000, immutable

<binary tile data>
```

**Errors:**
- 400: Bad request (invalid level or hash)
- 401: Unauthorized
- 404: Asset or level not found
- 500: Server error

---

## Asset Metadata via Widgets

Assets are normally accessed through their parent widget endpoints:

- Images: `GET /api/v1/canvases/{canvas-id}/images/{image-id}/download`
- Videos: `GET /api/v1/canvases/{canvas-id}/videos/{video-id}/download`
- PDFs: `GET /api/v1/canvases/{canvas-id}/pdfs/{pdf-id}/download`

These endpoints automatically handle authentication and canvas context.

---

## Asset Creation (Upload)

Assets are created when uploading widgets:

### Image Upload

```
POST /api/v1/canvases/{canvas-id}/images
Content-Type: multipart/form-data

--boundary
Content-Disposition: form-data; name="data"; filename="photo.jpg"
Content-Type: image/jpeg

<binary image data>
--boundary
Content-Disposition: form-data; name="json"
Content-Type: application/json

{"title": "My Photo", "location": {"x": 100, "y": 50}}
--boundary--
```

### Video Upload

```
POST /api/v1/canvases/{canvas-id}/videos
Content-Type: multipart/form-data

--boundary
Content-Disposition: form-data; name="data"; filename="video.mp4"
Content-Type: video/mp4

<binary video data>
--boundary
Content-Disposition: form-data; name="json"

{"title": "My Video"}
--boundary--
```

### PDF Upload

```
POST /api/v1/canvases/{canvas-id}/pdfs
Content-Type: multipart/form-data

--boundary
Content-Disposition: form-data; name="data"; filename="document.pdf"
Content-Type: application/pdf

<binary PDF data>
--boundary--
```

### Background Image Upload

```
POST /api/v1/canvases/{canvas-id}/background
Content-Type: multipart/form-data

--boundary
Content-Disposition: form-data; name="data"; filename="bg.jpg"
Content-Type: image/jpeg

<binary image data>
--boundary--
```

---

## Asset File Information

When assets are created or updated, the response includes:

```json
{
  "asset-hash": "public-hash-hex",
  "asset-hash-private": "private-hash-hex",
  "mime-type": "image/jpeg",
  "file-size": 1024000,
  "original-filename": "photo.jpg"
}
```

| Field | Type | Description |
|---|---|---|
| asset-hash | string | Public hash for download via `/api/v1/assets/{hash}` |
| asset-hash-private | string | Private hash (internal use only) |
| mime-type | string | MIME type (image/jpeg, video/mp4, application/pdf, etc.) |
| file-size | integer | File size in bytes |
| original-filename | string | Original filename from upload |

---

## Immutable Asset Caching

Assets in the server are immutable once created. They are cached with a 5-year expiry:

```
Cache-Control: private, max-age=157680000, immutable
```

If you need to replace an asset, create a new widget with the new file and delete the old widget.

---

## File Size & Format Limits

**Images:**
- Formats: PNG, JPEG, WebP, GIF, BMP, TIFF
- Max size: 2 GB per file (enforced by server)
- Mipmaps auto-generated for tiles rendering

**Videos:**
- Formats: MP4, WebM, MOV, AVI, MKV (depends on server codec support)
- Max size: Limited by disk space
- Streaming support depends on codec

**PDFs:**
- Format: PDF only
- Max size: Limited by disk space
- Page indexing supported

---

## Authentication Caveat

Direct asset access (`/api/v1/assets/{hash}` and `/api/v1/mipmaps/{hash}`) **requires both**:

1. **`Private-Token` header** — API key or session token
2. **`canvas-id` header** — UUID of a canvas that references the asset

This prevents unauthorized access to assets on other users' canvases, even with a valid token. The canvas context is mandatory.
