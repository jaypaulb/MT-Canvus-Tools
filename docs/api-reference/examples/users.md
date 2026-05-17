# User Management Examples

## User List and Details

### Example: List all users (admin only)

**Source:** mt-restapi-tests/tests/users_test.go:11 (TestUsersList)

**Request:**
```http
GET /api/v1/users HTTP/1.1
Authorization: Bearer <admin-token>
```

**Response (200):**
```json
[
  {
    "id": "user-abc123",
    "email": "alice@example.com",
    "full-name": "Alice Smith",
    "name": "Alice Smith",
    "created-at": "2024-03-01T10:00:00Z"
  },
  {
    "id": "user-def456",
    "email": "bob@example.com",
    "full-name": "Bob Johnson",
    "name": "Bob Johnson",
    "created-at": "2024-03-05T14:30:00Z"
  }
]
```

**Notes:** 
- Admin token required
- Response includes both `full-name` and `name` fields (may vary by server version)

---

### Example: Get user details by ID (admin only)

**Source:** mt-restapi-tests/tests/users_test.go:32 (TestUsersGet)

**Request:**
```http
GET /api/v1/users/user-abc123 HTTP/1.1
Authorization: Bearer <admin-token>
```

**Response (200):**
```json
{
  "id": "user-abc123",
  "email": "alice@example.com",
  "full-name": "Alice Smith",
  "name": "Alice Smith",
  "created-at": "2024-03-01T10:00:00Z",
  "updated-at": "2024-03-15T12:00:00Z",
  "last-login": "2024-03-16T09:30:00Z"
}
```

**Notes:** Admin can view any user's details. Full user metadata is returned.

---

## User Creation and Registration

### Example: Create a user (admin only)

**Source:** mt-restapi-tests/tests/users_test.go:58 (TestUsersCreate)

**Request:**
```http
POST /api/v1/users HTTP/1.1
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "email": "newuser@example.com",
  "name": "New User Name",
  "password": "SecurePassword123!"
}
```

**Response (200-299):**
```json
{
  "id": "user-xyz789",
  "email": "newuser@example.com",
  "name": "New User Name",
  "full-name": "New User Name",
  "created-at": "2024-03-16T16:00:00Z"
}
```

**Notes:** 
- Admin token required
- Password is mandatory
- Email should be unique
- Returns generated user ID

---

### Example: User self-registration

**Source:** mt-restapi-tests/tests/users_test.go:85 (TestUsersRegister)

**Request:**
```http
POST /api/v1/users/register HTTP/1.1
Content-Type: application/json

{
  "email": "selfregistered@example.com",
  "name": "Self Registered User",
  "password": "SecurePassword123!"
}
```

**Response (200-299 or 403 or 501):**
```json
{
  "id": "user-newreg123",
  "email": "selfregistered@example.com",
  "name": "Self Registered User",
  "created-at": "2024-03-16T16:05:00Z"
}
```

**Notes:** 
- No authentication required
- May return 403 (disabled) or 501 (not implemented) if self-registration is disabled
- Server configuration determines if this endpoint is available

---

## User Modification

### Example: Update user information (admin only)

**Source:** mt-restapi-tests/tests/users_test.go:136 (TestUsersUpdate)

**Request:**
```http
PATCH /api/v1/users/user-abc123 HTTP/1.1
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "name": "New User Name"
}
```

**Response (200):**
```json
{
  "id": "user-abc123",
  "email": "alice@example.com",
  "name": "New User Name",
  "full-name": "New User Name",
  "updated-at": "2024-03-16T16:10:00Z"
}
```

**Notes:** PATCH allows partial updates. Only provided fields are modified.

---

### Example: Change user password

**Source:** mt-restapi-tests/tests/users_test.go:186 (TestUsersPassword)

**Request:**
```http
POST /api/v1/users/user-abc123/password HTTP/1.1
Authorization: Bearer <user-token>
Content-Type: application/json

{
  "current_password": "OldPassword123!",
  "new_password": "NewPassword456!"
}
```

**Response (200):**
```json
{
  "success": true
}
```

**Notes:** 
- User must provide current password for verification
- New password must meet complexity requirements
- Typically requires authentication as the user whose password is being changed

---

### Example: Change user email

**Source:** mt-restapi-tests/tests/users_test.go:222 (TestUsersChangeEmail)

**Request:**
```http
POST /api/v1/users/user-abc123/change-email HTTP/1.1
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "email": "newemail@example.com"
}
```

**Response (200 or other 2xx):**
```json
{
  "success": true
}
```

**Notes:** 
- Admin token required
- New email should be unique
- Server may send verification email to new address

---

## User Access Control

### Example: Block a user

**Source:** mt-restapi-tests/tests/users_test.go:253 (TestUsersBlockUnblock)

**Request:**
```http
POST /api/v1/users/user-abc123/block HTTP/1.1
Authorization: Bearer <admin-token>
```

**Response (200):**
```json
{
  "id": "user-abc123",
  "email": "alice@example.com",
  "blocked": true,
  "blocked-at": "2024-03-16T16:15:00Z"
}
```

**Notes:** 
- Admin token required
- Blocked users cannot login or use API
- Blocking is reversible

---

### Example: Unblock a user

**Source:** mt-restapi-tests/tests/users_test.go:288 (TestUsersBlockUnblock)

**Request:**
```http
POST /api/v1/users/user-abc123/unblock HTTP/1.1
Authorization: Bearer <admin-token>
```

**Response (200):**
```json
{
  "id": "user-abc123",
  "email": "alice@example.com",
  "blocked": false
}
```

**Notes:** Restores access for blocked users.

---

### Example: Approve a user (if approval workflow enabled)

**Source:** mt-restapi-tests/tests/users_test.go:297 (TestUsersApprove)

**Request:**
```http
POST /api/v1/users/user-abc123/approve HTTP/1.1
Authorization: Bearer <admin-token>
```

**Response (200 or 409):**
```json
{
  "id": "user-abc123",
  "email": "alice@example.com",
  "approved": true,
  "approved-at": "2024-03-16T16:20:00Z"
}
```

**Notes:** 
- May return 409 Conflict if user is already approved
- Only applicable if approval workflow is enabled

---

## User Password Management

### Example: Reset user password (admin only)

**Source:** mt-restapi-tests/tests/users_test.go:325 (TestUsersAdminResetPassword)

**Request:**
```http
POST /api/v1/users/user-abc123/reset-password HTTP/1.1
Authorization: Bearer <admin-token>
```

**Response (200):**
```json
{
  "success": true,
  "temporary-password": "TempPass123!SecureToken"
}
```

**Notes:** 
- Admin token required
- Typically generates a temporary password that the user must change on next login
- Server implementation may vary

---

### Example: Confirm email address

**Source:** mt-restapi-tests/tests/users_test.go:121 (TestUsersConfirmEmail)

**Request (with invalid token):**
```http
POST /api/v1/users/confirm-email HTTP/1.1
Content-Type: application/json

{
  "token": "invalid-token-12345"
}
```

**Response (4xx):**
```json
{
  "error": "invalid or expired token"
}
```

**Notes:** 
- No authentication required
- Token must be valid and not expired
- Token is typically sent to user's email address

---

## User Deletion

### Example: Delete a user (admin only)

**Source:** mt-restapi-tests/tests/users_test.go:352 (TestUsersDelete)

**Request:**
```http
DELETE /api/v1/users/user-abc123 HTTP/1.1
Authorization: Bearer <admin-token>
```

**Response (200):**
```
(empty body)
```

**Verification (subsequent GET):**
```http
GET /api/v1/users/user-abc123 HTTP/1.1
Authorization: Bearer <admin-token>
```

**Response (404):**
```json
{
  "error": "user not found"
}
```

**Notes:** 
- Admin token required
- Deletion is permanent
- User's data (canvases, widgets, etc.) may be handled according to server policy
