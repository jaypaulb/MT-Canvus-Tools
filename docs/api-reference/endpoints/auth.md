# Authentication Endpoints

## Login & Session Management

### `POST /api/v1/users/login`

> Verified against dev-mtcs.multitaction.com 2026-05-18

**Auth:** none  
**Streaming:** no  
**Status:** implemented

Authenticate with email and password to obtain a session token.

**Request body:**

```json
{
  "email": "user@example.com",
  "password": "password123",
  "remember": true
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| email | string | yes | User email address |
| password | string | yes | User password (plaintext; TLS required) |
| remember | boolean | no | Extend session lifetime (persistent login) |

**Important:** The server performs **strict field validation** on this body. Sending any unknown field (e.g. `username`) causes the server to reject the request with `{"msg":"Login request must have either email and password or token"}` rather than attempting authentication. Send only the documented fields above.

**Response (200):**

```json
{
  "token": "session-token-string",
  "user": {
    "id": 1000,
    "email": "user@example.com",
    "name": "John Doe",
    "admin": false,
    "blocked": false
  }
}
```

| Field | Type | Description |
|---|---|---|
| token | string | Session token for subsequent API requests (use in `Private-Token` header or `CanvusSession` cookie) |
| user | object | Authenticated user details |

**User object fields** — keys use **underscores**; `id` is an **integer** (not a UUID):

| Field | Type | Description |
|---|---|---|
| id | integer | User identifier (integer, not UUID) |
| email | string | User email address |
| name | string | User's display name |
| admin | boolean | Whether the user has admin privileges |
| blocked | boolean | Whether the user account is blocked |

**Note:** The fields `user-id`, `full-name`, `is-admin`, `is-blocked`, and `avatar-color` shown in earlier spec drafts do not exist on the v1.2 wire shape.

**Errors:**
- 400: Bad request (missing email or password, or unexpected fields in body)
- 401: Unauthorized (invalid credentials or user blocked)
- 402: Payment Required (license seat limit reached)
- 500: Server error

**Note:** On successful login, the server returns a token. This token must be sent with all subsequent authenticated requests via the `Private-Token` HTTP header or stored as a `CanvusSession` cookie.

---

### `POST /api/v1/users/login/saml`

**Auth:** none  
**Streaming:** no  
**Status:** implemented

Authenticate with SAML assertion (if SAML is enabled in server config).

**Request body:**

```json
{
  "inResponseTo": "request-id",
  "responseXml": "<samlp:Response xmlns:samlp='urn:oasis:names:tc:SAML:2.0:protocol'>...",
  "remember": true
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| inResponseTo | string | yes | SAML request ID this response is answering |
| responseXml | string | yes | Base64-encoded or raw SAML assertion XML |
| remember | boolean | no | Extend session lifetime |

**Response (200):** Session token and user object (same format as `/login`)

**Errors:**
- 400: Bad request (missing or invalid SAML response)
- 401: Unauthorized (SAML assertion validation failed)
- 402: Payment Required (license seat limit)
- 500: Server error

**Note:** SAML authentication must be enabled in server configuration via the SAML provider settings.

---

### `POST /api/v1/users/logout`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Invalidate the current session token. User must log in again to regain access.

**Request body:** empty JSON object

```json
{}
```

**Response (200):** Logout confirmation

```json
{
  "msg": "Logged out successfully"
}
```

**Errors:**
- 401: Unauthorized (invalid or missing token)
- 500: Server error

**Note:** Logout revokes the token immediately. Any in-flight subscriptions using this token will be terminated by the server.

---

## Password Management (Unauthenticated)

Users can reset their password without authentication using a token sent via email.

### `POST /api/v1/users/password/create-reset-token`

**Auth:** none  
**Streaming:** no  
**Status:** implemented

Request a password reset. Server sends an email with a reset token to the user's registered email address.

**Request body:**

```json
{
  "email": "user@example.com"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| email | string | yes | Email address of account to reset |

**Response (200):** Confirmation message

```json
{
  "msg": "Password reset email sent (if account exists)"
}
```

**Note:** Response is always 200 regardless of whether the email exists (prevents email enumeration). Email is only sent if:
- Account exists
- User is not blocked
- SMTP is configured on the server

**Errors:**
- 400: Bad request (missing email)
- 500: Server error (SMTP failure)

---

### `GET /api/v1/users/password/validate-reset-token`

**Auth:** none  
**Streaming:** no  
**Status:** implemented

Validate that a password reset token is still valid before allowing password change.

**Query parameters:**
| Name | Type | Required | Description |
|---|---|---|---|
| token | string | yes | Password reset token from email link |

**Response (200):** Token validation success

```json
{
  "valid": true
}
```

**Response (400):** Token expired or invalid

```json
{
  "valid": false,
  "msg": "Token expired or invalid"
}
```

**Errors:**
- 400: Bad request or invalid token
- 500: Server error

---

### `POST /api/v1/users/password/reset`

**Auth:** none  
**Streaming:** no  
**Status:** implemented

Reset user password using a valid reset token from `/users/password/create-reset-token`.

**Request body:**

```json
{
  "token": "reset-token-from-email",
  "password": "new-password-123"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| token | string | yes | Password reset token |
| password | string | yes | New password (plaintext; TLS required) |

**Response (200):** Password reset success

```json
{
  "msg": "Password reset successfully. You can now log in with your new password."
}
```

**Errors:**
- 400: Bad request (missing fields or token expired)
- 500: Server error

**Note:** After password reset, all existing session tokens for that user are revoked. User must log in again with new password.

---

## User Registration (If Enabled)

Self-registration is available if enabled in server config.

### `POST /api/v1/users/register`

**Auth:** none  
**Streaming:** no  
**Status:** implemented

Register a new user account (requires self-registration to be enabled).

**Request body:**

```json
{
  "email": "newuser@example.com",
  "full-name": "Jane Doe",
  "password": "password123"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| email | string | yes | Email address (must be unique) |
| full-name | string | yes | User's full name |
| password | string | yes | Initial password (plaintext; TLS required) |

**Response (201):** Registration success

```json
{
  "msg": "Account created. Check email for confirmation link.",
  "user": {
    "user-id": "uuid",
    "email": "newuser@example.com",
    "full-name": "Jane Doe",
    "is-admin": false,
    "is-blocked": false
  }
}
```

**Errors:**
- 400: Bad request (invalid email format or missing fields)
- 402: Registration disabled on server
- 409: Conflict (email already registered)
- 500: Server error

**Note:** After registration, user must confirm their email address before they can log in (unless admin approval is required first).

---

### `POST /api/v1/users/confirm-email`

**Auth:** none  
**Streaming:** no  
**Status:** implemented

Confirm email address using token sent in registration email.

**Request body:**

```json
{
  "token": "email-confirmation-token-from-email"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| token | string | yes | Email confirmation token |

**Response (200):** Email confirmed

```json
{
  "msg": "Email confirmed successfully"
}
```

**Errors:**
- 400: Bad request (token missing or expired)
- 500: Server error

**Note:** After email confirmation, user can log in with their credentials.

---

## API Access Tokens

API tokens are long-lived credentials for programmatic access, distinct from login session tokens.

### `GET /api/v1/users/{user-id}/access-tokens`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all API access tokens for a user.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| user-id | string (UUID) | User identifier |

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| subscribe | boolean | no | false | Enable streaming updates |

**Response (200):** JSON array of API token objects

```json
[
  {
    "token-id": "uuid",
    "name": "CI/CD Token",
    "created": "2025-01-15T10:30:00Z",
    "last-used": "2025-05-17T14:22:15Z",
    "expires": "2025-12-31T23:59:59Z",
    "scopes": ["read", "write"],
    "prefix": "cv_abc123..."
  }
]
```

| Field | Type | Description |
|---|---|---|
| token-id | string | Unique identifier for this token |
| name | string | User-assigned name |
| created | string (ISO 8601) | Creation timestamp |
| last-used | string (ISO 8601) | Last API request using this token |
| expires | string (ISO 8601) | Expiration date (may be null for no expiry) |
| scopes | array | List of permission scopes (read, write, admin) |
| prefix | string | First 16 characters of token (for identification; full token only returned on creation) |

**Errors:**
- 401: Unauthorized
- 404: User not found
- 500: Server error

---

### `GET /api/v1/users/{user-id}/access-tokens/{token-id}`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single API access token by ID.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| user-id | string (UUID) | User identifier |
| token-id | string (UUID) | Token identifier |

**Response (200):** Single API token object (same format as GET list)

**Errors:**
- 401: Unauthorized
- 404: User or token not found
- 500: Server error

---

### `POST /api/v1/users/{user-id}/access-tokens`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a new API access token for a user.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| user-id | string (UUID) | User identifier |

**Request body:**

```json
{
  "name": "CI/CD Pipeline",
  "expires": "2025-12-31T23:59:59Z",
  "scopes": ["read", "write"]
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| name | string | yes | Descriptive name for the token |
| expires | string (ISO 8601) | no | Expiration date (null = never expire) |
| scopes | array | no | Permission scopes: read, write, admin (default: read, write) |

**Response (201):** Created token with full token value (only returned once)

```json
{
  "token-id": "uuid",
  "name": "CI/CD Pipeline",
  "token": "cv_1234567890abcdefghijklmnopqrstuvwxyz",
  "created": "2025-05-17T14:22:15Z",
  "expires": "2025-12-31T23:59:59Z",
  "scopes": ["read", "write"],
  "prefix": "cv_1234567890ab"
}
```

**Important:** The full `token` value is only returned during creation. Store it securely. If lost, create a new token.

**Errors:**
- 400: Bad request (invalid expiry format)
- 401: Unauthorized
- 404: User not found
- 500: Server error

---

### `PATCH /api/v1/users/{user-id}/access-tokens/{token-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update an API access token's metadata (name, expiry, scopes).

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| user-id | string (UUID) | User identifier |
| token-id | string (UUID) | Token identifier |

**Request body:**

```json
{
  "name": "Updated Name",
  "expires": "2026-12-31T23:59:59Z",
  "scopes": ["read"]
}
```

All fields are optional; only specified fields are updated.

**Response (200):** Updated token object (token value not included in updates)

**Errors:**
- 400: Bad request
- 401: Unauthorized
- 404: User or token not found
- 500: Server error

---

### `DELETE /api/v1/users/{user-id}/access-tokens/{token-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Revoke an API access token. Token becomes invalid immediately.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| user-id | string (UUID) | User identifier |
| token-id | string (UUID) | Token identifier |

**Response (204):** No content (token deleted)

**Errors:**
- 401: Unauthorized
- 404: User or token not found
- 500: Server error
