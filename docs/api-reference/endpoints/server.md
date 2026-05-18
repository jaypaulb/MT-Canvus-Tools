# Server Administration Endpoints

## Server Information

### `GET /api/v1/server-info`

**Auth:** none  
**Streaming:** no  
**Status:** implemented

Retrieve server version and runtime information. No authentication required.

**Response (200):**

```json
{
  "version": "26.4.0",
  "commit": "abc1234567890def",
  "build-date": "2025-05-17T12:00:00Z",
  "go-version": "go1.22.3",
  "platform": "linux/x86_64"
}
```

| Field | Type | Description |
|---|---|---|
| version | string | Canvus server version (semver) |
| commit | string | Git commit hash |
| build-date | string (ISO 8601) | Build timestamp |
| go-version | string | Go runtime version |
| platform | string | OS/architecture (linux/x86_64, windows/x86_64, etc.) |

**Errors:**
- 500: Server error

---

## Server Configuration

### `GET /api/v1/server-config`

> Verified against dev-mtcs.multitaction.com 2026-05-18

**Auth:** login-token | api-key | none (guest access for reading)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve server configuration settings. Guest access allowed (without auth).

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| subscribe | boolean | no | false | Enable streaming updates |

**Response (200):** Deeply nested JSON object with underscore-separated keys. The flat element-array form (`[{setting-key, setting-value, setting-type}]`) documented in earlier spec drafts is **not** served by v1.2.

```json
{
  "access": "rw",
  "authentication": {
    "domain_allow_list": ["*"],
    "password": { "enabled": true, "min_length": 8 },
    "saml": { "acs_url": "...", "enabled": true }
  },
  "email": { "smtp_host": "", "smtp_port": 25 },
  "external_url": "https://your-server.example.com",
  "server_name": "My Canvus Server"
}
```

Keys use **underscores** throughout the GET response.

**Errors:**
- 500: Server error

---

### `PATCH /api/v1/server-config`

**Auth:** login-token (admin required) | api-key (admin required)  
**Streaming:** no  
**Status:** implemented

Update server configuration settings (admin only).

**Request body:**

```json
{
  "settings": [
    {
      "setting-key": "smtp.enabled",
      "setting-value": "true"
    },
    {
      "setting-key": "max-users",
      "setting-value": "200"
    }
  ]
}
```

**Response (200):** Updated configuration settings

**Errors:**
- 400: Bad request (invalid setting key or value)
- 401: Unauthorized or insufficient privileges
- 500: Server error

---

### `POST /api/v1/server-config/reload-certs`

**Auth:** login-token (admin required) | api-key (admin required)  
**Streaming:** no  
**Status:** implemented

Reload TLS certificates from disk without restarting the server.

**Request body:** empty JSON object

```json
{}
```

**Response (200):** Certificates reloaded

```json
{
  "msg": "TLS certificates reloaded successfully"
}
```

**Errors:**
- 400: Bad request
- 401: Unauthorized
- 500: Server error (certificate format issue)

---

### `POST /api/v1/server-config/send-test-email`

**Auth:** login-token (admin required) | api-key (admin required)  
**Streaming:** no  
**Status:** implemented

Send a test email to verify SMTP configuration.

**Request body:**

```json
{
  "recipient-email": "admin@example.com"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| recipient-email | string | yes | Email address to send test to |

**Response (200):** Test email sent

```json
{
  "msg": "Test email sent to admin@example.com"
}
```

**Errors:**
- 400: Bad request (invalid email)
- 401: Unauthorized
- 500: Server error (SMTP not configured or connection failed)

---

## License Management

### `GET /api/v1/license`

> Verified against dev-mtcs.multitaction.com 2026-05-18

**Auth:** login-token (admin required) | api-key (admin required)  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve current license information and usage statistics.

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| subscribe | boolean | no | false | Enable streaming updates |

**Response (200):**

```json
{
  "edition": "",
  "has_expired": false,
  "is_valid": true,
  "max_clients": -1,
  "seat_model": "fixed_seats",
  "type": "lifetime"
}
```

| Field | Type | Description |
|---|---|---|
| edition | string | License edition string |
| has_expired | boolean | Whether the license has expired |
| is_valid | boolean | Whether the license is currently valid |
| max_clients | integer | Maximum allowed clients (-1 = unlimited) |
| seat_model | string | Seat model: usage_reported, fixed_seats, or none |
| type | string | License type (e.g. "lifetime") |

**Note:** Keys use **underscores**. The fields `status`, `clients`, `valid`, `message`, `expiry-date`, `max-clients`, `activation-required`, and `seat-model` (hyphenated) documented in earlier spec drafts do not appear in the v1.2 response. `expiry-date` and `activation-required` may appear on non-lifetime license types.

**Errors:**
- 401: Unauthorized
- 402: Payment Required (license invalid or expired)
- 500: Server error

---

### `GET /api/v1/license/request`

**Auth:** login-token (admin required) | api-key (admin required)  
**Streaming:** no  
**Status:** implemented

Get the activation request payload for offline license activation.

**Response (200):**

```json
{
  "request": "base64-encoded-request-payload"
}
```

**Usage:** Send this request to the license server for offline activation and receive a license file back.

**Errors:**
- 401: Unauthorized
- 500: Server error

---

### `POST /api/v1/license`

> Verified against dev-mtcs.multitaction.com 2026-05-18

**Auth:** login-token (admin required) | api-key (admin required)  
**Streaming:** no  
**Status:** implemented

Install a new license file (typically from offline activation).

**Request body:**

```json
{
  "license": "<license-key-string>"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| license | string | yes | License key or file content |

**Note:** The field is `"license"` (not `"license-data"` or `"key"` as shown in earlier spec drafts). Sending `"key"` returns `{"msg":"license parameter is missing"}`; sending `"license-data"` returns `{"msg":"Unknown action ..."}`. The endpoint is `POST /api/v1/license` — the path `/api/v1/license/install` returns `{"msg":"Unknown action install"}`.

**Response (200):** License installed

```json
{
  "msg": "License installed successfully"
}
```

**Errors:**
- 400: Bad request (invalid license data)
- 401: Unauthorized
- 500: Server error

---

### `POST /api/v1/license/activate`

**Auth:** login-token (admin required) | api-key (admin required)  
**Streaming:** no  
**Status:** implemented

Activate the license online (contacts license server).

**Request body:** empty JSON object

```json
{}
```

**Response (200):** License activated

```json
{
  "msg": "License activated successfully"
}
```

**Errors:**
- 401: Unauthorized
- 500: Server error (network failure or license server unreachable)

---

## Audit Log

### `GET /api/v1/audit-log`

**Auth:** login-token (admin required) | api-key (admin required)  
**Streaming:** no (does NOT support `?subscribe`)  
**Status:** implemented

Retrieve paginated audit log events. Returns events in reverse chronological order (newest first).

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| page | integer | no | 0 | Page number (0-based) |
| per-page | integer | no | 100 | Events per page (max 1000) |
| filter | string | no | — | Query filter (e.g., `"action=create_canvas"`) |
| start-time | string (ISO 8601) | no | — | Filter: events after this timestamp |
| end-time | string (ISO 8601) | no | — | Filter: events before this timestamp |
| user-id | string (UUID) | no | — | Filter: events by user |
| action | string | no | — | Filter: events by action type |

**Response (200):** JSON array of audit log events with pagination

```json
{
  "events": [
    {
      "event-id": "uuid",
      "timestamp": "2025-05-17T14:22:15Z",
      "user-id": "user-uuid",
      "user-email": "user@example.com",
      "action": "create_canvas",
      "resource-type": "canvas",
      "resource-id": "canvas-uuid",
      "changes": {
        "canvas-name": "New Canvas"
      },
      "ip-address": "192.168.1.100",
      "user-agent": "CanvusApp/26.4.0"
    }
  ],
  "total-count": 5432,
  "page": 0,
  "per-page": 100
}
```

**Link Header:** For pagination, responses include a `Link` header:

```
Link: </api/v1/audit-log?page=1&per-page=100>; rel="next"
```

**Common Actions:**
- `create_canvas`, `update_canvas`, `delete_canvas`
- `create_user`, `update_user`, `delete_user`
- `create_group`, `update_group`, `delete_group`
- `login`, `logout`
- `update_settings`
- `create_api_token`, `revoke_api_token`

**Errors:**
- 401: Unauthorized
- 500: Server error

**Note:** Audit log does NOT support `?subscribe` for streaming. Use CSV export instead.

---

### `GET /api/v1/audit-log/export-csv`

**Auth:** login-token (admin required) | api-key (admin required)  
**Streaming:** no  
**Status:** implemented

Export the full audit log as a CSV file download.

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| filter | string | no | — | Query filter (same as `/audit-log`) |
| start-time | string (ISO 8601) | no | — | Filter: events after this timestamp |
| end-time | string (ISO 8601) | no | — | Filter: events before this timestamp |
| user-id | string (UUID) | no | — | Filter: events by user |

**Response (200):** CSV file with `content-disposition: attachment` header

```
event-id,timestamp,user-id,user-email,action,resource-type,resource-id,changes,ip-address
uuid1,2025-05-17T14:22:15Z,user-uuid,user@example.com,create_canvas,canvas,canvas-uuid,...
uuid2,2025-05-17T14:20:00Z,user-uuid,user@example.com,update_canvas,canvas,canvas-uuid,...
```

**Errors:**
- 401: Unauthorized
- 500: Server error (export failure)

---

## Connected Clients

### `GET /api/v1/clients`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all connected Canvus client applications (desktop apps, web browsers, etc.).

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| subscribe | boolean | no | false | Enable streaming updates |

**Response (200):** JSON array of client objects

```json
[
  {
    "client-id": "uuid",
    "app-name": "Canvus Desktop",
    "app-version": "26.4.0",
    "connected-at": "2025-05-17T10:00:00Z",
    "last-activity": "2025-05-17T14:22:15Z",
    "ip-address": "192.168.1.100",
    "user-email": "user@example.com",
    "workspace-count": 3
  }
]
```

**Errors:**
- 401: Unauthorized
- 500: Server error

---

### `GET /api/v1/clients/{client-id}`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve details for a single connected client.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| client-id | string (UUID) | Client identifier |

**Response (200):** Single client object

**Errors:**
- 404: Client not found
- 500: Server error

---

## Client Workspaces & Video I/O

### `GET /api/v1/clients/{client-id}/workspaces`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all workspaces (open canvases) for a connected client.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| client-id | string (UUID) | Client identifier |

**Response (200):** JSON array of workspace objects

```json
[
  {
    "workspace-id": "uuid",
    "workspace-name": "Main Display",
    "canvas-id": "canvas-uuid",
    "canvas-name": "Project A",
    "size": {"width": 3840, "height": 2160},
    "canvas-size": {"width": 3840, "height": 2160},
    "workspace-state": "active|minimized|archived",
    "view-location": {"x": 0, "y": 0},
    "view-scale": 1.0,
    "pinned": true,
    "info-panel-visible": true,
    "owner": "user@example.com",
    "server-id": "server-uuid"
  }
]
```

---

### `GET /api/v1/clients/{client-id}/workspaces/{workspace-id}`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single workspace.

---

### `PATCH /api/v1/clients/{client-id}/workspaces/{workspace-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update workspace settings (view location, scale, pinned state, etc.).

**Request body:**

```json
{
  "view-location": {"x": 100, "y": 200},
  "view-scale": 0.5,
  "pinned": true
}
```

---

### `POST /api/v1/clients/{client-id}/workspaces/{workspace-id}/open-canvas`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Open a different canvas in a specific workspace.

**Request body:**

```json
{
  "canvas-id": "new-canvas-uuid"
}
```

---

### `GET /api/v1/clients/{client-id}/video-outputs`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all video output devices/sinks available on the client.

---

### `GET /api/v1/clients/{client-id}/video-outputs/{output-id}`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single video output.

---

### `PATCH /api/v1/clients/{client-id}/video-outputs/{output-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update a video output configuration.

---

### `GET /api/v1/clients/{client-id}/video-inputs`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all video input devices available on the client (webcams, capture cards, etc.).

**Response (200):** JSON array of video input objects

```json
[
  {
    "input-id": "uuid",
    "name": "Webcam",
    "source": "device-id",
    "resolution": "1920x1080",
    "fps": 30
  }
]
```

---

### `GET /api/v1/clients/{client-id}/video-inputs/{input-id}`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single video input device.
