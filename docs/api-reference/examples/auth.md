# Authentication Examples

## Login Examples

### Example: Login with valid credentials

**Source:** mt-restapi-tests/tests/auth_test.go:7 (TestLoginValid)

**Request:**
```http
POST /api/v1/users/login HTTP/1.1
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Notes:** The returned `token` is a JWT that can be used in subsequent requests via the `Authorization: Bearer <token>` header or `Private-Token: <token>` header. This is the primary authentication method for all API endpoints.

---

### Example: Login with invalid password

**Source:** mt-restapi-tests/tests/auth_test.go:23 (TestLoginInvalid)

**Request:**
```http
POST /api/v1/users/login HTTP/1.1
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "WrongPassword123!"
}
```

**Response (401 or 403):**
```json
{
  "error": "invalid credentials"
}
```

**Notes:** Both 401 and 403 are accepted as valid error responses for authentication failures.

---

### Example: Login with remember-me flag

**Source:** mt-restapi-tests/tests/auth_test.go:36 (TestLoginRemember)

**Request:**
```http
POST /api/v1/users/login HTTP/1.1
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "remember": true
}
```

**Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Notes:** The `remember` flag typically extends the token expiration time. This is a boolean optional parameter.

---

## Logout Examples

### Example: Logout and invalidate token

**Source:** mt-restapi-tests/tests/auth_test.go:56 (TestLogout)

**Request:**
```http
POST /api/v1/users/logout HTTP/1.1
Authorization: Bearer <token>
```

**Response (200 or 204):**
```
(empty body)
```

**Notes:** 
- The token is invalidated immediately after logout.
- Subsequent API calls with this token will return 401 or 403.
- Both 200 and 204 are accepted as successful logout responses.

---

## SAML Login Examples

### Example: SAML login attempt with invalid assertion

**Source:** mt-restapi-tests/tests/auth_test.go:100 (TestLoginSAML)

**Request:**
```http
POST /api/v1/users/login/saml HTTP/1.1
Content-Type: application/json

{
  "inResponseTo": "test_request_id",
  "responseXml": "<invalid/>"
}
```

**Response (400 or 401):**
```json
{
  "error": "invalid SAML response"
}
```

**Notes:** 
- SAML authentication requires valid assertions from a configured identity provider.
- This endpoint is only functional if SAML is configured on the server.
- Invalid assertions will return 400 (bad request) or 401 (unauthorized).
