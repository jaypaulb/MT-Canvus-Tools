# Canvus REST API Reference

This folder contains the complete canonical REST API reference for the Canvus Server, extracted from the C++ REST API client implementation (`mt-restapi-client`) and validated against the Go integration test suite (`mt-restapi-tests`).

**Spec Freeze Date:** 2026-05-17  
**Source Commits:** See [SOURCE.md](./SOURCE.md)

## Organization

### Endpoint Specifications

- **[canvases.md](./endpoints/canvases.md)** — Canvas documents, folders, background, color presets
- **[widgets.md](./endpoints/widgets.md)** — Canvas widget CRUD (notes, images, videos, PDFs, browsers, anchors, connectors, tables, video-inputs, etc.)
- **[auth.md](./endpoints/auth.md)** — Login, logout, SAML, password reset flows
- **[users.md](./endpoints/users.md)** — User management, groups, group membership, API tokens
- **[folders.md](./endpoints/folders.md)** — Canvas folder organization and hierarchy
- **[assets.md](./endpoints/assets.md)** — File upload, asset retrieval, mipmaps, hash-based access
- **[server.md](./endpoints/server.md)** — Server info, configuration, license, audit log, clients
- **[permissions.md](./endpoints/permissions.md)** — Permission models for canvases and folders

### Conceptual Documentation

- **[authentication.md](./authentication.md)** — Auth mechanisms, token types, header formats, scopes, expiry
- **[widget-types.md](./widget-types.md)** — Data model for each widget type, serialization format, lifecycle
- **[streaming.md](./streaming.md)** — Subscribe pattern, server-sent updates, keepalive, newline-delimited JSON

### Examples & Test Fixtures

The `examples/` folder contains request/response pairs extracted from the integration test suite (`mt-restapi-tests`).

---

## Quick Reference

All endpoints are prefixed with `/api/v1/` and follow these conventions:

| Aspect | Details |
|--------|---------|
| **Base URL** | `https://canvus-server/api/v1/` |
| **Authentication** | `Private-Token` header (API key) or `CanvusSession` cookie (base64 JSON) |
| **Streaming** | Add `?subscribe` to any GET endpoint to receive live updates as newline-delimited JSON |
| **Content-Type** | `application/json` (default); `multipart/form-data` for asset upload |
| **Errors** | JSON `{"msg": "error description"}` with HTTP status (400, 401, 403, 404, 409, 500, etc.) |

---

## Key Patterns

### Subscription (Long-Polling)

GET endpoints that return lists or documents can be subscribed to for live updates:

```bash
curl -H "Private-Token: abc123" \
  "https://canvus-server/api/v1/canvases?subscribe"
```

Server responds with HTTP 200 and streams updates as newline-delimited JSON, one per line:

```json
{"canvas-id": "uuid", "canvas-name": "Project X", ...}
{"canvas-id": "uuid2", "canvas-name": "Project Y", ...}
```

The connection stays open; server sends keepalive pings every 15 seconds (configurable). Client must unsubscribe by calling the corresponding unsubscribe endpoint or closing the connection.

### Asset Upload

Image, video, and PDF widgets are created via `multipart/form-data`:

```
POST /api/v1/canvases/{canvas-id}/images
Content-Type: multipart/form-data

--boundary
Content-Disposition: form-data; name="data"; filename="photo.jpg"
Content-Type: image/jpeg

<binary file data>
--boundary
Content-Disposition: form-data; name="json"
Content-Type: application/json

{"title": "My Photo", "location": {"x": 100, "y": 50}}
--boundary--
```

### Link Sharing

Canvases can be accessed without authentication if link sharing is enabled. Check the `link-permission` field (view/edit/none).

### Permission Scopes

- **Admin-only**: User must have `is-admin: true`
- **Content scope**: User must have permission to the specific canvas/folder
- **Public/Guest**: Accessible without authentication if link sharing or self-registration is enabled

---

## Common Response Fields

### Canvas Object

```json
{
  "canvas-id": "uuid",
  "canvas-name": "Project Name",
  "owner": "user@example.com",
  "demo-canvas": false,
  "has-main-password": false,
  "has-access-password": false,
  "size": {"width": 3840, "height": 2160},
  "created": "2025-01-15T10:30:00Z",
  "modified": "2025-05-17T14:22:15Z",
  "link-permission": "view",
  "permission-overrides": [...]
}
```

### Widget Object (Common Fields)

```json
{
  "location": {"x": 100, "y": 200},
  "depth": 5.0,
  "size": {"width": 400, "height": 300},
  "scale": 1.0,
  "is-pinned": false
}
```

Note fields (text, background-color, title, text-color) and image fields (original-filename, title) extend this base.

### Error Object

```json
{
  "msg": "Canvas not found"
}
```

---

## Authentication & Authorization

See [authentication.md](./authentication.md) for detailed auth mechanics, token lifecycle, and scope rules.

---

## Integration Notes

- **Coordinate System:** All `location`, `size`, and viewport values are in **pixels** (not normalized 0-1). Use only coordinate utilities for screen↔canvas conversions (zoom/pan).
- **Immutable Assets:** Once uploaded, asset files are immutable and cached with `cache-control: private, max-age=157680000, immutable`.
- **Password Hashing:** User passwords are sent in request bodies; TLS is mandatory.
- **Email Notifications:** Password resets, registration approvals, and email confirmations trigger SMTP (if configured).

---

## Further Reading

- **API Roadmap:** See `canvus-server/docs/canvus-server-26.04-api-roadmap.md` for planned features.
- **Architecture:** See `canvus-server/docs/ARCHITECTURE-REPORT.md` for system design.
- **Migration Notes:** See `mt-restapi-client/migration-to-26.04-concerns.md` for upgrade guidance from earlier versions.

---

*For live API testing, use the [Canvus API client library](https://github.com/jaypaul/Canvus-Go-API) or [MCP server](https://github.com/jaypaul/CanvusMCP) with [Claude Desktop](https://claude.ai/download).*
