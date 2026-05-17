# Authentication

The Canvus Server supports two distinct authentication mechanisms:

1. **Login Token** (session-based) — obtained via `/api/v1/users/login`, for interactive users
2. **API Key** (long-lived) — created via `/api/v1/users/{user-id}/access-tokens`, for programmatic access

Both are interchangeable in the HTTP header. Choose based on your use case.

---

## Login Token (Session-Based)

A login token is a short-to-medium-lived credential obtained when a user logs in with email and password.

### Obtaining a Token

```bash
curl -X POST https://canvus-server/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "remember": false
  }'
```

**Response:**

```json
{
  "token": "session-token-abc1234567890def",
  "user": {
    "user-id": "user-uuid",
    "email": "user@example.com",
    "full-name": "John Doe",
    "is-admin": false,
    "is-blocked": false,
    "avatar-color": "#FF0000"
  }
}
```

### Token Lifetime

- **Standard**: Session token expires after a configurable period (typically 30 days)
- **Remember Me**: If `"remember": true`, the token lifetime extends (typically 1 year)
- **Revocation**: Token is invalidated immediately if:
  - User logs out via `POST /api/v1/users/logout`
  - User password is changed or reset
  - User is blocked by an admin
  - User is deleted

### Sending the Token

Include the token in the `Private-Token` HTTP header:

```bash
curl -H "Private-Token: session-token-abc1234567890def" \
  https://canvus-server/api/v1/canvases
```

Alternatively, send as a cookie (used by web clients):

```bash
curl -b "CanvusSession=base64(json_payload)" \
  https://canvus-server/api/v1/canvases
```

Where `json_payload` is:

```json
{
  "passport": {
    "user": {
      "token": "session-token-abc1234567890def"
    }
  }
}
```

---

## API Key (Long-Lived)

An API key is a long-lived credential for programmatic access, created and managed per user.

### Creating an API Key

```bash
curl -X POST \
  https://canvus-server/api/v1/users/{user-id}/access-tokens \
  -H "Private-Token: existing-login-token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "CI/CD Pipeline",
    "expires": "2025-12-31T23:59:59Z",
    "scopes": ["read", "write"]
  }'
```

**Response (only returned once):**

```json
{
  "token-id": "token-uuid",
  "name": "CI/CD Pipeline",
  "token": "cv_1234567890abcdefghijklmnopqrstuvwxyz",
  "created": "2025-05-17T14:22:15Z",
  "expires": "2025-12-31T23:59:59Z",
  "scopes": ["read", "write"],
  "prefix": "cv_1234567890ab"
}
```

**Important:** Store the full `token` value securely. It is only displayed during creation. If lost, create a new token.

### API Key Format

API keys have the prefix `cv_` followed by alphanumeric characters:

```
cv_1234567890abcdefghijklmnopqrstuvwxyz
```

### Token Lifetime

- **Custom Expiry**: API tokens can have no expiry date (`"expires": null`) or a specific expiration date
- **Scopes**: Currently, scopes are: `read`, `write`, `admin`
- **Revocation**: Token is invalidated immediately when:
  - Explicitly revoked via `DELETE /api/v1/users/{user-id}/access-tokens/{token-id}`
  - User password is reset by an admin
  - User is deleted

### Sending the API Key

Include the key in the `Private-Token` HTTP header:

```bash
curl -H "Private-Token: cv_1234567890abcdefghijklmnopqrstuvwxyz" \
  https://canvus-server/api/v1/canvases
```

---

## Authentication Headers

All authenticated endpoints require one of:

### 1. Private-Token Header (Recommended)

```
GET /api/v1/canvases HTTP/1.1
Host: canvus-server
Private-Token: session-token-or-api-key
```

### 2. CanvusSession Cookie (Web Clients)

```
GET /api/v1/canvases HTTP/1.1
Host: canvus-server
Cookie: CanvusSession=base64-encoded-json
```

### 3. Bearer Token (OAuth, if supported)

```
GET /api/v1/canvases HTTP/1.1
Host: canvus-server
Authorization: Bearer session-token-or-api-key
```

---

## Scopes & Permissions

### Login Token Scopes

A login token inherits the user's permissions:

| Permission | Effect |
|------------|--------|
| **Admin** | Full read-write access to all canvases, users, groups, configuration |
| **Normal User** | Read-write access to owned canvases and canvases shared with them |
| **Blocked User** | No access; login fails |

### API Key Scopes

API keys can be created with limited scopes:

| Scope | Effect |
|-------|--------|
| **read** | GET requests only (list, read canvases, widgets, users, etc.) |
| **write** | POST, PATCH, DELETE requests (create, update, delete canvases, widgets, etc.) |
| **admin** | Admin-only endpoints (configuration, license, users, groups) |

If no scopes are specified during creation, the key defaults to `["read", "write"]`.

**Example:** Create a read-only CI key for downloading canvases:

```bash
curl -X POST \
  https://canvus-server/api/v1/users/{user-id}/access-tokens \
  -H "Private-Token: admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Read-Only CI Key",
    "scopes": ["read"]
  }'
```

---

## TLS / HTTPS

**All authentication must occur over TLS/HTTPS.** HTTP is not secure for transmitting passwords or tokens.

If you connect via HTTP:

```bash
curl -H "Private-Token: abc123" http://canvus-server/api/v1/canvases
# May fail or return 301 redirect to HTTPS
```

Always use `https://`:

```bash
curl -H "Private-Token: abc123" https://canvus-server/api/v1/canvases
```

---

## Content Scope (Permissions)

Even with a valid token, users cannot access canvases or resources they don't have permission for.

### Canvas Access

A user can access a canvas if:

1. User **owns** the canvas (created it)
2. Canvas is **shared** with the user or their group
3. Canvas has **link sharing** enabled (public/view-only link)
4. User is an **admin** (can access all canvases)

Attempting to access a canvas without permission returns 403 Forbidden:

```json
{
  "msg": "You do not have permission to access this canvas"
}
```

### Metadata Scope vs Content Scope

Some endpoints require different permission scopes:

- **Metadata Scope**: Can list/read canvas names, IDs, permissions (but not contents)
- **Content Scope**: Can read/modify canvas widgets and data

For example, `GET /api/v1/canvases` returns only canvases the user has access to (metadata scope).

---

## Token Validation & Renewal

### Validating a Token

Tokens are validated on every request. Invalid or expired tokens return 401 Unauthorized:

```json
{
  "msg": "Unauthorized: invalid token"
}
```

### Renewing a Login Token

Login tokens can be renewed by logging in again. There is no explicit "refresh token" endpoint; simply log in again with credentials.

API keys do not expire during their lifetime; only at the expiration date set during creation.

---

## Error Handling

### Missing Token

```bash
curl https://canvus-server/api/v1/canvases
# Returns 401 Unauthorized
```

### Invalid Token

```bash
curl -H "Private-Token: invalid-token" \
  https://canvus-server/api/v1/canvases
# Returns 401 Unauthorized
```

### Insufficient Permissions

```bash
curl -H "Private-Token: normal-user-token" \
  https://canvus-server/api/v1/users
# Returns 401 or 403 (depending on endpoint)
```

### Token Expired

```bash
curl -H "Private-Token: expired-login-token" \
  https://canvus-server/api/v1/canvases
# Returns 401 Unauthorized
```

---

## Security Best Practices

1. **Use HTTPS Only**: Never send tokens over unencrypted HTTP
2. **Store API Keys Securely**: Treat API keys like passwords; store in environment variables or secrets managers (e.g., `~/.netrc`, AWS Secrets Manager, HashiCorp Vault)
3. **Limit Key Scope**: Create API keys with minimal required scopes (`read`-only if possible)
4. **Rotate Keys**: Periodically create new keys and revoke old ones
5. **Log Out**: Call `POST /api/v1/users/logout` when done with a session
6. **Revoke Suspicious Tokens**: If a token is compromised, revoke it immediately via the API or admin dashboard

---

## Integration Examples

### cURL

```bash
# Login
TOKEN=$(curl -s -X POST https://canvus-server/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"pass"}' \
  | jq -r '.token')

# Use token
curl -H "Private-Token: $TOKEN" \
  https://canvus-server/api/v1/canvases
```

### Python

```python
import requests

# Login
response = requests.post(
    'https://canvus-server/api/v1/users/login',
    json={'email': 'user@example.com', 'password': 'pass'}
)
token = response.json()['token']

# Use token
headers = {'Private-Token': token}
canvases = requests.get(
    'https://canvus-server/api/v1/canvases',
    headers=headers
).json()
```

### Go

```go
// Login
loginResp, _ := http.Post(
    "https://canvus-server/api/v1/users/login",
    "application/json",
    bytes.NewBufferString(`{"email":"user@example.com","password":"pass"}`),
)
var result struct{ Token string }
json.NewDecoder(loginResp.Body).Decode(&result)

// Use token
client := &http.Client{}
req, _ := http.NewRequest("GET", "https://canvus-server/api/v1/canvases", nil)
req.Header.Set("Private-Token", result.Token)
resp, _ := client.Do(req)
```

### TypeScript / Node.js

```typescript
// Login
const loginResp = await fetch('https://canvus-server/api/v1/users/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ email: 'user@example.com', password: 'pass' })
});
const { token } = await loginResp.json();

// Use token
const canvases = await fetch('https://canvus-server/api/v1/canvases', {
  headers: { 'Private-Token': token }
}).then(r => r.json());
```
