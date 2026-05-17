# Streaming and Real-time Updates

## Overview

The Canvus REST API supports server-sent streaming for real-time updates. By adding the `subscribe=true` query parameter to certain endpoints, clients receive continuous newline-delimited JSON updates as changes occur on the server.

---

## Real-time Canvas Updates

### Example: Subscribe to canvas widget changes

**Documentation Reference:** Canvus Go SDK documentation mentions HTTPS streaming with `subscribe=true`

**Request:**
```http
GET /api/v1/canvases/canvas-abc123?subscribe=true HTTP/1.1
Authorization: Bearer <token>
Accept: application/x-ndjson
```

**Response (200):**
```
Connection: keep-alive
Content-Type: application/x-ndjson
Transfer-Encoding: chunked

{"id":"canvas-abc123","name":"My Canvas","updated-at":"2024-03-16T18:00:00Z"}
{"type":"widget-added","widget":{"id":"note-xyz789","type":"note","text":"New note"}}
{"type":"widget-updated","widget":{"id":"note-xyz789","text":"Updated note"}}
{"type":"widget-deleted","widget":{"id":"note-xyz789"}}
```

**Notes:** 
- Each line is a separate JSON object (newline-delimited JSON / NDJSON)
- Connection remains open, streaming updates as they occur
- Client must handle newline-delimited parsing
- Connection may timeout if no activity for extended period

---

### Example: Subscribe to specific widget changes

**Request:**
```http
GET /api/v1/canvases/canvas-abc123/notes/note-xyz789?subscribe=true HTTP/1.1
Authorization: Bearer <token>
Accept: application/x-ndjson
```

**Response (200):**
```
{"id":"note-xyz789","type":"note","text":"Current content"}
{"type":"updated","text":"Modified content","updated-at":"2024-03-16T18:01:00Z"}
{"type":"updated","background-color":"#FF0000","updated-at":"2024-03-16T18:02:00Z"}
```

**Notes:** 
- Streams updates specific to the widget
- Initial response may include current state
- Subsequent messages are incremental updates

---

## Polling as Alternative to Streaming

### Example: Poll for canvas changes (alternative to subscribe)

**Request:**
```http
GET /api/v1/canvases/canvas-abc123?last-modified=2024-03-16T17:59:00Z HTTP/1.1
Authorization: Bearer <token>
```

**Response (200):**
```json
{
  "id": "canvas-abc123",
  "name": "My Canvas",
  "updated-at": "2024-03-16T18:01:00Z",
  "changes": [
    {
      "type": "widget-added",
      "widget": {
        "id": "note-xyz789",
        "type": "note",
        "text": "New note"
      }
    }
  ]
}
```

**Notes:** 
- Use `last-modified` query parameter to request only changes since a timestamp
- More efficient than re-fetching entire canvas
- Returns empty `changes` array if no updates since specified time
- Client implements polling loop to check periodically

---

## Connection Management

### Example: Graceful subscription termination

**Request:**
```http
GET /api/v1/canvases/canvas-abc123?subscribe=true HTTP/1.1
Authorization: Bearer <token>
Accept: application/x-ndjson
Connection: close
```

**Response:**
```
[streams updates until client closes connection]
```

**Client-side handling:**
```
1. Open connection with subscribe=true
2. Parse incoming NDJSON stream
3. Process each update as it arrives
4. On error or user action, close connection
5. Optionally reconnect with last-modified timestamp for resumption
```

**Notes:** 
- Connection is closed when client or server terminates it
- Server may enforce idle timeouts (typically 30-60 seconds)
- Reconnect logic recommended for robust implementations

---

## Batch Operations with Streaming

### Example: Upload multiple assets with progress streaming

**Request:**
```http
POST /api/v1/canvases/canvas-abc123/batch-upload?subscribe=true HTTP/1.1
Authorization: Bearer <token>
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="items[0].data"; filename="file1.jpg"
Content-Type: image/jpeg

[binary data]
------WebKitFormBoundary--
```

**Response (200 with streaming):**
```
{"status":"processing","total":3,"completed":0}
{"status":"uploading","item":0,"filename":"file1.jpg","progress":50}
{"status":"completed","item":0,"id":"image-001"}
{"status":"uploading","item":1,"filename":"file2.jpg","progress":75}
{"status":"completed","item":1,"id":"image-002"}
{"status":"uploading","item":2,"filename":"file3.jpg","progress":100}
{"status":"completed","item":2,"id":"image-003"}
{"status":"finished","total":3,"success":3}
```

**Notes:** 
- Streaming updates during long-running operations
- Useful for progress indication in client UIs
- Final message indicates completion status

---

## Error Handling in Streams

### Example: Handle mid-stream errors

**Request:**
```http
GET /api/v1/canvases/canvas-abc123?subscribe=true HTTP/1.1
Authorization: Bearer <token>
```

**Response (initial 200, then error):**
```
{"id":"canvas-abc123","name":"My Canvas"}
{"type":"widget-added","widget":{"id":"note-1"}}
{"error":"permission-denied","message":"User access revoked","code":403}
```

**Client-side handling:**
```
1. Monitor incoming JSON for "error" field
2. On error, stop processing further updates
3. Log error and reconnect if appropriate
4. Use last known state if recovery not possible
```

**Notes:** 
- HTTP 200 is sent immediately, but stream may contain errors later
- Errors in stream indicate mid-operation failures
- Client must handle both initial HTTP errors (401, 403) and stream-level errors

---

## Performance Considerations

### Example: Filtering subscriptions by change type (proposed)

**Request (if supported):**
```http
GET /api/v1/canvases/canvas-abc123?subscribe=true&filter=widget-added,widget-updated HTTP/1.1
Authorization: Bearer <token>
```

**Response:**
```
[streams only specified event types]
```

**Notes:** 
- Filtering may reduce bandwidth and client processing
- Check server documentation for supported filters
- Common filters: `widget-added`, `widget-updated`, `widget-deleted`, `permission-changed`

---

### Example: Heartbeat and keep-alive

**Request:**
```http
GET /api/v1/canvases/canvas-abc123?subscribe=true HTTP/1.1
Authorization: Bearer <token>
```

**Response:**
```
{"id":"canvas-abc123"}
{"type":"update","data":{...}}
{"type":"heartbeat","timestamp":"2024-03-16T18:05:00Z"}
{"type":"update","data":{...}}
{"type":"heartbeat","timestamp":"2024-03-16T18:05:30Z"}
```

**Notes:** 
- Server may send heartbeat messages to keep connection alive
- Heartbeats prevent timeout on idle connections
- Client can use heartbeat timestamps to detect connection staleness

---

## Supported Streaming Endpoints

Based on API structure, these endpoints likely support streaming:

| Endpoint | Subscribe Parameter | Use Case |
|----------|-------------------|----------|
| `GET /canvases/{id}` | `?subscribe=true` | Real-time canvas changes |
| `GET /canvases/{id}/widgets` | `?subscribe=true` | Widget list updates |
| `GET /canvases/{id}/notes` | `?subscribe=true` | Note changes |
| `GET /canvases/{id}/images` | `?subscribe=true` | Image changes |
| `GET /canvases/{id}/videos` | `?subscribe=true` | Video metadata changes |
| `GET /canvases/{id}/pdfs` | `?subscribe=true` | PDF metadata changes |

**Notes:** 
- Check server documentation for complete list
- Most endpoints that support GET are candidates for streaming
- POST, PATCH, DELETE operations typically do not support streaming

---

## Implementation Example (Pseudocode)

```python
# Python example of streaming subscription
import httpx
import json

def subscribe_to_canvas(canvas_id, token):
    url = f"https://api.server.com/api/v1/canvases/{canvas_id}?subscribe=true"
    headers = {"Authorization": f"Bearer {token}"}
    
    with httpx.stream("GET", url, headers=headers) as response:
        for line in response.iter_lines():
            if line:
                update = json.loads(line)
                if "error" in update:
                    print(f"Error: {update['message']}")
                    break
                else:
                    handle_update(update)

def handle_update(update):
    if update.get("type") == "widget-added":
        print(f"Widget added: {update['widget']['id']}")
    elif update.get("type") == "widget-updated":
        print(f"Widget updated: {update['widget']['id']}")
    elif update.get("type") == "heartbeat":
        print(f"Keep-alive: {update['timestamp']}")
```

---

## Reconnection Strategy

### Recommended client approach:

```
1. Open initial subscription with subscribe=true
2. Record current timestamp as last_update
3. Parse incoming NDJSON stream
4. If connection closes unexpectedly:
   a. Wait exponential backoff (1s, 2s, 4s, ...)
   b. Reconnect with ?subscribe=true&since=last_update
   c. Resume from where you left off
5. On permission errors (403):
   a. Stop reconnection attempts
   b. Refresh auth token and retry
6. On server errors (5xx):
   a. Implement exponential backoff
   b. Eventually give up after max retries
```

**Benefits:**
- Seamless reconnection without data loss
- Resilience to transient network failures
- Minimal overhead by resuming from last known state
