# Asset Management Examples

## Image Assets

### Example: Upload an image

**Source:** mt-restapi-tests/tests/images_test.go:8 (TestImageLifecycle)

**Request:**
```http
POST /api/v1/canvases/canvas-abc123/images HTTP/1.1
Authorization: Bearer <token>
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="data"; filename="product.jpg"
Content-Type: image/jpeg

[binary image data]
------WebKitFormBoundary
Content-Disposition: form-data; name="title"

Product Screenshot
------WebKitFormBoundary--
```

**Response (200-299):**
```json
{
  "id": "image-xyz789",
  "type": "image",
  "title": "Product Screenshot",
  "filename": "product.jpg",
  "size": 245120,
  "width": 1920,
  "height": 1080,
  "mime-type": "image/jpeg",
  "created-at": "2024-03-16T17:00:00Z",
  "updated-at": "2024-03-16T17:00:00Z"
}
```

**Notes:** 
- Multipart form data with file in `data` field
- `title` is optional metadata
- Server automatically detects image dimensions

---

### Example: List images on a canvas

**Source:** mt-restapi-tests/tests/images_test.go:32 (TestImageLifecycle)

**Request:**
```http
GET /api/v1/canvases/canvas-abc123/images HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```json
[
  {
    "id": "image-xyz789",
    "type": "image",
    "title": "Product Screenshot",
    "width": 1920,
    "height": 1080
  },
  {
    "id": "image-abc123",
    "type": "image",
    "title": "Architecture Diagram",
    "width": 2560,
    "height": 1440
  }
]
```

**Notes:** Returns all images on the canvas with basic metadata.

---

### Example: Get image details

**Source:** mt-restapi-tests/tests/images_test.go:49 (TestImageLifecycle)

**Request:**
```http
GET /api/v1/canvases/canvas-abc123/images/image-xyz789 HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```json
{
  "id": "image-xyz789",
  "type": "image",
  "title": "Product Screenshot",
  "filename": "product.jpg",
  "size": 245120,
  "width": 1920,
  "height": 1080,
  "mime-type": "image/jpeg",
  "position": {"x": 100, "y": 100},
  "size": {"width": 1920, "height": 1080},
  "created-at": "2024-03-16T17:00:00Z",
  "updated-at": "2024-03-16T17:00:00Z"
}
```

**Notes:** Full image metadata including position on canvas.

---

### Example: Update image metadata

**Source:** mt-restapi-tests/tests/images_test.go:57 (TestImageLifecycle)

**Request:**
```http
PATCH /api/v1/canvases/canvas-abc123/images/image-xyz789 HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Updated Screenshot Title"
}
```

**Response (200):**
```json
{
  "id": "image-xyz789",
  "type": "image",
  "title": "Updated Screenshot Title",
  "updated-at": "2024-03-16T17:05:00Z"
}
```

**Notes:** PATCH allows metadata updates without re-uploading the file.

---

### Example: Download image

**Source:** mt-restapi-tests/tests/images_test.go:66 (TestImageLifecycle)

**Request:**
```http
GET /api/v1/canvases/canvas-abc123/images/image-xyz789/download HTTP/1.1
Authorization: Bearer <token>
```

**Response (200 or 404):**
```
[binary image data]
```

**Headers:**
```
Content-Type: image/jpeg
Content-Disposition: attachment; filename="product.jpg"
Cache-Control: public, max-age=3600
```

**Notes:** 
- May return 404 if image requires async processing
- Binary response, not JSON
- Content-Disposition header provides filename for downloads

---

### Example: Delete image

**Source:** mt-restapi-tests/tests/images_test.go:93 (TestImageLifecycle)

**Request:**
```http
DELETE /api/v1/canvases/canvas-abc123/images/image-xyz789 HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```
(empty body)
```

**Notes:** Deletion is permanent.

---

## Video Assets

### Example: Upload a video

**Source:** mt-restapi-tests/tests/videos_test.go:12 (TestVideoLifecycle)

**Request:**
```http
POST /api/v1/canvases/canvas-abc123/videos HTTP/1.1
Authorization: Bearer <token>
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="data"; filename="demo.mp4"
Content-Type: video/mp4

[binary video data]
------WebKitFormBoundary--
```

**Response (200-299):**
```json
{
  "id": "video-xyz789",
  "type": "video",
  "filename": "demo.mp4",
  "size": 5242880,
  "duration": 120,
  "mime-type": "video/mp4",
  "playback-state": "STOPPED",
  "muted": false,
  "created-at": "2024-03-16T17:10:00Z"
}
```

**Notes:** 
- No title parameter required for videos
- Server automatically extracts duration
- `playback-state` may be uppercase (STOPPED, PLAYING, PAUSED)

---

### Example: List videos on a canvas

**Source:** mt-restapi-tests/tests/videos_test.go:28 (TestVideoLifecycle)

**Request:**
```http
GET /api/v1/canvases/canvas-abc123/videos HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```json
[
  {
    "id": "video-xyz789",
    "type": "video",
    "filename": "demo.mp4",
    "duration": 120,
    "playback-state": "STOPPED"
  }
]
```

**Notes:** Returns all videos on the canvas.

---

### Example: Get video details

**Source:** mt-restapi-tests/tests/videos_test.go:36 (TestVideoLifecycle)

**Request:**
```http
GET /api/v1/canvases/canvas-abc123/videos/video-xyz789 HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```json
{
  "id": "video-xyz789",
  "type": "video",
  "filename": "demo.mp4",
  "size": 5242880,
  "duration": 120,
  "mime-type": "video/mp4",
  "playback-state": "STOPPED",
  "muted": false,
  "position": {"x": 50, "y": 50},
  "size": {"width": 1280, "height": 720},
  "created-at": "2024-03-16T17:10:00Z"
}
```

**Notes:** Includes playback metadata and canvas position.

---

### Example: Update video properties

**Source:** mt-restapi-tests/tests/videos_test.go:51 (TestVideoLifecycle)

**Request:**
```http
PATCH /api/v1/canvases/canvas-abc123/videos/video-xyz789 HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "muted": true
}
```

**Response (200):**
```json
{
  "id": "video-xyz789",
  "type": "video",
  "muted": true,
  "updated-at": "2024-03-16T17:15:00Z"
}
```

**Notes:** Can update mute status and other properties without re-uploading.

---

### Example: Download video

**Source:** mt-restapi-tests/tests/videos_test.go:60 (TestVideoLifecycle)

**Request:**
```http
GET /api/v1/canvases/canvas-abc123/videos/video-xyz789/download HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```
[binary video data]
```

**Notes:** 
- Returns complete video file
- Binary response, not JSON
- Suitable for streaming

---

### Example: Delete video

**Source:** mt-restapi-tests/tests/videos_test.go:68 (TestVideoLifecycle)

**Request:**
```http
DELETE /api/v1/canvases/canvas-abc123/videos/video-xyz789 HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```
(empty body)
```

**Notes:** Deletion is permanent.

---

## PDF Assets

### Example: Upload a PDF

**Source:** mt-restapi-tests/tests/pdfs_test.go:12 (TestPDFLifecycle)

**Request:**
```http
POST /api/v1/canvases/canvas-abc123/pdfs HTTP/1.1
Authorization: Bearer <token>
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="data"; filename="document.pdf"
Content-Type: application/pdf

[binary PDF data]
------WebKitFormBoundary--
```

**Response (200-299):**
```json
{
  "id": "pdf-xyz789",
  "type": "pdf",
  "filename": "document.pdf",
  "size": 1048576,
  "pages": 24,
  "mime-type": "application/pdf",
  "created-at": "2024-03-16T17:20:00Z"
}
```

**Notes:** 
- Server automatically extracts page count
- No metadata fields required for upload

---

### Example: List PDFs on a canvas

**Source:** mt-restapi-tests/tests/pdfs_test.go:28 (TestPDFLifecycle)

**Request:**
```http
GET /api/v1/canvases/canvas-abc123/pdfs HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```json
[
  {
    "id": "pdf-xyz789",
    "type": "pdf",
    "filename": "document.pdf",
    "pages": 24
  }
]
```

**Notes:** Returns all PDFs on the canvas.

---

### Example: Get PDF details

**Source:** mt-restapi-tests/tests/pdfs_test.go:36 (TestPDFLifecycle)

**Request:**
```http
GET /api/v1/canvases/canvas-abc123/pdfs/pdf-xyz789 HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```json
{
  "id": "pdf-xyz789",
  "type": "pdf",
  "filename": "document.pdf",
  "size": 1048576,
  "pages": 24,
  "mime-type": "application/pdf",
  "position": {"x": 200, "y": 200},
  "size": {"width": 800, "height": 600},
  "created-at": "2024-03-16T17:20:00Z"
}
```

**Notes:** Includes canvas position and page count.

---

### Example: Update PDF metadata

**Source:** mt-restapi-tests/tests/pdfs_test.go:44 (TestPDFLifecycle)

**Request:**
```http
PATCH /api/v1/canvases/canvas-abc123/pdfs/pdf-xyz789 HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Updated Document Title"
}
```

**Response (200):**
```json
{
  "id": "pdf-xyz789",
  "type": "pdf",
  "title": "Updated Document Title",
  "updated-at": "2024-03-16T17:25:00Z"
}
```

**Notes:** Supports title and other metadata updates.

---

### Example: Download PDF

**Source:** mt-restapi-tests/tests/pdfs_test.go:53 (TestPDFLifecycle)

**Request:**
```http
GET /api/v1/canvases/canvas-abc123/pdfs/pdf-xyz789/download HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```
[binary PDF data]
```

**Notes:** 
- Returns complete PDF file
- Binary response, not JSON

---

### Example: Delete PDF

**Source:** mt-restapi-tests/tests/pdfs_test.go:61 (TestPDFLifecycle)

**Request:**
```http
DELETE /api/v1/canvases/canvas-abc123/pdfs/pdf-xyz789 HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```
(empty body)
```

**Notes:** Deletion is permanent.

---

## Uploads Folder (Generic Upload)

### Example: Upload via uploads-folder endpoint (image)

**Source:** mt-restapi-tests/tests/uploads_test.go:23 (TestUploadsFolderImage)

**Request:**
```http
POST /api/v1/canvases/canvas-abc123/uploads-folder HTTP/1.1
Authorization: Bearer <token>
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="title"

E2E Upload Image
------WebKitFormBoundary
Content-Disposition: form-data; name="data"; filename="upload.jpg"
Content-Type: image/jpeg

[binary image data]
------WebKitFormBoundary--
```

**Response (200-299):**
```json
{
  "id": "image-upload-001",
  "type": "image",
  "title": "E2E Upload Image",
  "created-at": "2024-03-16T17:30:00Z"
}
```

**Notes:** 
- Generic upload endpoint that auto-detects file type
- Creates appropriate widget (image, video, PDF) based on mime type

---

### Example: Upload note via uploads-folder (without file)

**Source:** mt-restapi-tests/tests/uploads_test.go:44 (TestUploadsFolderNoteNoData)

**Request:**
```http
POST /api/v1/canvases/canvas-abc123/uploads-folder HTTP/1.1
Authorization: Bearer <token>
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="upload_type"

Note
------WebKitFormBoundary
Content-Disposition: form-data; name="title"

E2E Upload Note
------WebKitFormBoundary--
```

**Response (2xx):**
```json
{
  "id": "note-upload-001",
  "type": "note",
  "title": "E2E Upload Note",
  "created-at": "2024-03-16T17:35:00Z"
}
```

**Notes:** 
- Notes can be created via uploads endpoint without file data
- `upload_type` parameter allows explicit type specification
- Behavior varies by server configuration
