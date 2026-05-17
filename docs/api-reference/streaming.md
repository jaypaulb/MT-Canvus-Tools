# Streaming & Real-Time Updates

The Canvus REST API supports real-time updates via a **subscribe pattern**. Clients can request a long-lived HTTP connection that streams updates as newline-delimited JSON whenever data changes on the server.

---

## Subscribe Pattern

### How It Works

1. Client sends GET request with `?subscribe` query parameter
2. Server responds with HTTP 200 and keeps connection open
3. Server sends initial data (the current state)
4. Server streams updates as they occur (newline-delimited JSON)
5. Server sends periodic keepalive pings (empty line or ping frame)
6. Client closes connection or server closes on timeout

### Supported Endpoints

All GET endpoints that return lists or documents support `?subscribe`:

**Canvases & Folders:**
- GET /api/v1/canvases?subscribe
- GET /api/v1/canvases/{canvas-id}?subscribe
- GET /api/v1/canvas-folders?subscribe
- GET /api/v1/canvas-folders/{folder-id}?subscribe

**Widgets:**
- GET /api/v1/canvases/{canvas-id}/widgets?subscribe
- GET /api/v1/canvases/{canvas-id}/notes?subscribe
- GET /api/v1/canvases/{canvas-id}/images?subscribe
- GET /api/v1/canvases/{canvas-id}/videos?subscribe
- GET /api/v1/canvases/{canvas-id}/pdfs?subscribe
- GET /api/v1/canvases/{canvas-id}/browsers?subscribe
- GET /api/v1/canvases/{canvas-id}/anchors?subscribe
- GET /api/v1/canvases/{canvas-id}/connectors?subscribe
- GET /api/v1/canvases/{canvas-id}/tables?subscribe
- (and individual widget types: `/notes/{id}?subscribe`, etc.)

**Users & Groups:**
- GET /api/v1/users?subscribe
- GET /api/v1/users/{user-id}?subscribe
- GET /api/v1/users/{user-id}/access-tokens?subscribe
- GET /api/v1/groups?subscribe
- GET /api/v1/groups/{group-id}?subscribe
- GET /api/v1/groups/{group-id}/members?subscribe

**Server:**
- GET /api/v1/server-config?subscribe
- GET /api/v1/license?subscribe
- GET /api/v1/audit-log (does NOT support subscribe)
- GET /api/v1/clients?subscribe
- GET /api/v1/clients/{client-id}/workspaces?subscribe

**NOT Supported:** POST, PATCH, DELETE, and endpoints that are inherently write operations do not support subscribe.

---

## Example: Subscribing to Canvas List

### Request

```bash
curl -N "https://canvus-server/api/v1/canvases?subscribe" \
  -H "Private-Token: session-token"
```

The `-N` flag tells cURL not to buffer output (required for streaming).

### Response

```
HTTP/1.1 200 OK
Content-Type: application/json
Transfer-Encoding: chunked
Cache-Control: no-cache

{"canvas-id":"uuid-1","canvas-name":"Project A","owner":"user@example.com",...}
{"canvas-id":"uuid-2","canvas-name":"Project B","owner":"user@example.com",...}

{"canvas-id":"uuid-3","canvas-name":"Project C","owner":"user@example.com",...}

{"canvas-id":"uuid-1","canvas-name":"Project A (Renamed)","owner":"user@example.com",...}
```

**Key observations:**

1. **Initial state**: First few lines are the current state (list of canvases)
2. **One JSON object per line**: Each newline indicates a separate object (not an array)
3. **Empty lines**: Periodic keepalive pings (heartbeat, typically every 15 seconds)
4. **Updates**: When a canvas is renamed, the server sends a new object for that canvas
5. **Connection stays open**: The TCP connection remains open; new updates arrive in real time

### Unsubscribe

To stop receiving updates, close the HTTP connection:

```bash
# Press Ctrl+C to close the cURL connection
# Or send a request to explicitly unsubscribe:
curl -X POST https://canvus-server/api/v1/unsubscribe \
  -H "Private-Token: session-token" \
  -d '{"request-id": ...}'  # (if applicable)
```

---

## Data Format

All subscribed responses are **newline-delimited JSON (NDJSON)**.

### Single Object (Streaming)

```json
{"widget-id":"uuid","text":"Note content","location":{"x":100,"y":200}}
{"widget-id":"uuid","text":"Note content updated","location":{"x":100,"y":200}}
```

### Array (Initial State)

When first subscribing to a list endpoint, the server sends the array of current items:

```json
[
  {"widget-id":"uuid-1","text":"Note 1"},
  {"widget-id":"uuid-2","text":"Note 2"}
]
```

Then, subsequent updates for individual items are sent as standalone objects:

```json
{"widget-id":"uuid-1","text":"Note 1 Updated"}
```

### Keepalive Pings

The server sends empty lines (or whitespace) periodically to keep the connection alive and let clients know the server is still responsive.

```
{"canvas-id":"uuid-1",...}

{"canvas-id":"uuid-2",...}
```

The blank line is a keepalive ping. Clients should ignore empty lines.

---

## Keepalive Interval

The server sends keepalive pings every **15 seconds** (configurable in server config via `keepalive-interval`).

Clients should:
1. **Expect empty lines**: Don't treat them as errors
2. **Have a connection timeout**: If no data (including pings) arrive for 30+ seconds, close and reconnect
3. **Filter empty lines**: When parsing NDJSON, skip lines that are only whitespace

---

## Client Implementation Examples

### cURL (Manual)

```bash
# Subscribe to canvases
curl -N "https://canvus-server/api/v1/canvases?subscribe" \
  -H "Private-Token: $TOKEN" \
  | while IFS= read -r line; do
      [ -z "$line" ] && continue  # Skip empty lines
      echo "$line" | jq .  # Parse and pretty-print JSON
    done
```

### Python

```python
import requests
import json

url = "https://canvus-server/api/v1/canvases?subscribe"
headers = {"Private-Token": token}

with requests.get(url, headers=headers, stream=True) as r:
    for line in r.iter_lines():
        if not line:
            continue  # Skip empty lines (keepalive)
        data = json.loads(line)
        print(f"Canvas: {data.get('canvas-name')}")
```

### JavaScript / Node.js

```javascript
const token = 'session-token';
const url = new URL('https://canvus-server/api/v1/canvases?subscribe');

const response = await fetch(url, {
  headers: { 'Private-Token': token }
});

const reader = response.body.getReader();
const decoder = new TextDecoder();
let buffer = '';

while (true) {
  const { done, value } = await reader.read();
  if (done) break;

  buffer += decoder.decode(value, { stream: true });
  const lines = buffer.split('\n');

  for (let i = 0; i < lines.length - 1; i++) {
    const line = lines[i].trim();
    if (line) {
      const data = JSON.parse(line);
      console.log('Canvas:', data['canvas-name']);
    }
  }

  buffer = lines[lines.length - 1];  // Incomplete line
}
```

### Go

```go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
)

func subscribeCanvases(token string) {
	req, _ := http.NewRequest("GET", "https://canvus-server/api/v1/canvases?subscribe", nil)
	req.Header.Set("Private-Token", token)

	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue // Keepalive ping
		}

		var canvas map[string]interface{}
		json.Unmarshal([]byte(line), &canvas)
		fmt.Printf("Canvas: %v\n", canvas["canvas-name"])
	}
}
```

### TypeScript with Fetch + AsyncIterator

```typescript
async function* streamCanvases(token: string) {
  const response = await fetch(
    'https://canvus-server/api/v1/canvases?subscribe',
    { headers: { 'Private-Token': token } }
  );

  const reader = response.body!.getReader();
  const decoder = new TextDecoder();
  let buffer = '';

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });
    const lines = buffer.split('\n');

    for (let i = 0; i < lines.length - 1; i++) {
      const line = lines[i].trim();
      if (line) {
        yield JSON.parse(line);
      }
    }

    buffer = lines[lines.length - 1];
  }
}

// Usage
(async () => {
  for await (const canvas of streamCanvases(token)) {
    console.log(`Canvas: ${canvas['canvas-name']}`);
  }
})();
```

---

## Handling Connection Failures

### Timeout & Reconnect

If the subscription connection times out or breaks:

```python
import time

def subscribe_with_reconnect(url, headers, max_retries=5):
    retries = 0
    while retries < max_retries:
        try:
            with requests.get(url, headers=headers, stream=True, timeout=30) as r:
                for line in r.iter_lines(timeout=30):
                    if not line:
                        continue
                    data = json.loads(line)
                    yield data
                retries = 0  # Reset on success
        except (requests.Timeout, requests.ConnectionError) as e:
            print(f"Subscription failed: {e}")
            retries += 1
            if retries < max_retries:
                wait_time = min(2 ** retries, 30)  # Exponential backoff
                print(f"Reconnecting in {wait_time}s...")
                time.sleep(wait_time)
            else:
                raise

# Usage
for canvas in subscribe_with_reconnect(url, headers):
    print(f"Canvas: {canvas['canvas-name']}")
```

### Handling Partial Frames

Network libraries may not receive complete lines at once. Always:

1. Buffer incomplete lines
2. Process only when a newline is received
3. Handle multiple lines per network read

---

## Performance Considerations

### Bandwidth

Streaming updates are typically smaller than full re-queries:

- **First request** (initial state): Full list (all canvases, all widgets)
- **Subsequent updates**: Only changed items (one canvas renamed, one widget added)

### Latency

Subscription updates typically arrive within **100-500ms** of server-side changes, depending on network latency and server load.

### Connection Lifetime

A subscription connection can remain open **indefinitely**, but:

1. Network intermediaries (proxies, firewalls) may close idle connections after 30-60 minutes
2. Server will close connections after extended inactivity (configurable, typically 1 hour)
3. Client should implement exponential backoff reconnection logic for robustness

### Concurrent Subscriptions

A single client can have multiple concurrent subscriptions (one per tab or window):

```bash
# Terminal 1: Subscribe to canvases
curl -N "https://canvus-server/api/v1/canvases?subscribe" \
  -H "Private-Token: $TOKEN"

# Terminal 2: Subscribe to users
curl -N "https://canvus-server/api/v1/users?subscribe" \
  -H "Private-Token: $TOKEN"

# Both streams run simultaneously
```

---

## Error Handling in Subscriptions

### Connection Errors

```
HTTP/1.1 401 Unauthorized

{"msg": "Unauthorized: invalid token"}
```

If credentials expire mid-subscription, the server closes the connection with an error response.

### Permission Changes

If a user's permissions change while subscribed (e.g., canvas is unshared), the server may:

1. Stop sending updates for that resource
2. Disconnect the subscription
3. Send an error response

Client should handle gracefully and re-authenticate if needed.

### Server Shutdown

If the server shuts down while a client is subscribed:

- TCP connection is dropped by OS
- Client library detects connection loss
- Client should implement reconnection logic

---

## Comparison: Polling vs. Streaming

### Polling (without subscribe)

```bash
# Client polls every 5 seconds
while true; do
  curl "https://canvus-server/api/v1/canvases" \
    -H "Private-Token: $TOKEN" | jq .
  sleep 5
done
```

**Pros:**
- Simple stateless protocol
- No long-lived connections
- Works through HTTP/1.0 proxies

**Cons:**
- High latency (up to 5 seconds)
- High bandwidth (full response each time)
- Server load increases with polling frequency

### Streaming (with subscribe)

```bash
# Client receives updates in real time
curl -N "https://canvus-server/api/v1/canvases?subscribe" \
  -H "Private-Token: $TOKEN"
```

**Pros:**
- Low latency (updates within 100-500ms)
- Low bandwidth (only changes sent)
- Server can push to many clients efficiently

**Cons:**
- Requires long-lived HTTP connections
- More complex error handling
- Some proxies/firewalls may terminate idle connections

---

## Best Practices

1. **Always skip empty lines** when parsing NDJSON
2. **Use reasonable timeouts** (30-60 seconds) to detect dead connections
3. **Implement exponential backoff** for reconnection
4. **Buffer incomplete JSON** at line boundaries
5. **Log keepalive pings** during development (helps debug connection issues)
6. **Monitor connection duration** and reconnect after extended periods (>1 hour)
7. **Gracefully handle auth errors** (re-login and restart subscription)
8. **Test with network failures** (kill -9 server, pull network cable, proxy timeouts)
