# Users & Groups Endpoints

## Users Management

### `GET /api/v1/users`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all users on the server.

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| subscribe | boolean | no | false | Enable streaming updates |

**Response (200):** JSON array of user objects

```json
[
  {
    "user-id": "uuid",
    "email": "user@example.com",
    "full-name": "John Doe",
    "is-admin": false,
    "is-blocked": false,
    "groups": ["group-uuid-1", "group-uuid-2"],
    "avatar-color": "#FF0000",
    "created": "2025-01-15T10:30:00Z",
    "last-login": "2025-05-17T14:22:15Z"
  }
]
```

**Errors:**
- 401: Unauthorized
- 500: Server error

---

### `GET /api/v1/users/{user-id}`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single user by ID.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| user-id | string (UUID) | User identifier |

**Response (200):** Single user object

**Errors:**
- 404: User not found
- 500: Server error

---

### `POST /api/v1/users`

**Auth:** login-token (admin required) | api-key (admin required)  
**Streaming:** no  
**Status:** implemented

Create a new user (admin only). Unlike `/users/register`, this endpoint can create users without email confirmation requirements.

**Request body:**

```json
{
  "email": "newuser@example.com",
  "full-name": "Jane Doe",
  "password": "initial-password-123",
  "is-admin": false,
  "avatar-color": "#00FF00"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| email | string | yes | Email address (must be unique) |
| full-name | string | yes | User's full name |
| password | string | yes | Initial password (plaintext; TLS required) |
| is-admin | boolean | no | Grant admin privileges (default: false) |
| avatar-color | string (hex) | no | Avatar color for UI (default: auto-assigned) |

**Response (201):** Created user object

**Errors:**
- 400: Bad request (invalid email or missing fields)
- 401: Unauthorized or insufficient privileges
- 409: Conflict (email already exists)
- 500: Server error

---

### `PATCH /api/v1/users/{user-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update user information (name, avatar color, etc.). Users can update their own profile; admins can update any user.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| user-id | string (UUID) | User identifier |

**Request body:**

```json
{
  "full-name": "Jane Smith",
  "avatar-color": "#0000FF"
}
```

All fields are optional; only specified fields are updated.

**Response (200):** Updated user object

**Errors:**
- 401: Unauthorized
- 404: User not found
- 500: Server error

---

### `POST /api/v1/users/{user-id}/password`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Set or change a user's password. Users can change their own password (with old password verification); admins can reset any user's password.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| user-id | string (UUID) | User identifier |

**Request body (user changing own password):**

```json
{
  "old-password": "current-password",
  "new-password": "new-password-123"
}
```

**Request body (admin resetting password):**

```json
{
  "new-password": "new-password-123"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| old-password | string | conditional | Required if user is changing own password (plaintext; TLS required) |
| new-password | string | yes | New password (plaintext; TLS required) |

**Response (200):** Password update success

**Errors:**
- 400: Bad request (weak password or missing fields)
- 401: Unauthorized or incorrect old password
- 404: User not found
- 500: Server error

---

### `POST /api/v1/users/{user-id}/change-email`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Request an email change. Server sends confirmation email to new address; user must confirm before change takes effect.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| user-id | string (UUID) | User identifier |

**Request body:**

```json
{
  "new-email": "newemail@example.com"
}
```

**Response (200):** Email change request initiated

```json
{
  "msg": "Confirmation email sent to newemail@example.com"
}
```

**Errors:**
- 400: Bad request (invalid email format)
- 401: Unauthorized
- 404: User not found
- 409: Conflict (email already in use)
- 500: Server error

---

### `POST /api/v1/users/{user-id}/block`

**Auth:** login-token (admin required) | api-key (admin required)  
**Streaming:** no  
**Status:** implemented

Block a user account. User cannot log in, but account data is retained.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| user-id | string (UUID) | User identifier |

**Request body:** empty JSON object

```json
{}
```

**Response (200):** User blocked

**Errors:**
- 401: Unauthorized or insufficient privileges
- 404: User not found
- 500: Server error

---

### `POST /api/v1/users/{user-id}/unblock`

**Auth:** login-token (admin required) | api-key (admin required)  
**Streaming:** no  
**Status:** implemented

Unblock a previously blocked user account.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| user-id | string (UUID) | User identifier |

**Request body:** empty JSON object

```json
{}
```

**Response (200):** User unblocked

**Errors:**
- 401: Unauthorized
- 404: User not found
- 500: Server error

---

### `POST /api/v1/users/{user-id}/approve`

**Auth:** login-token (admin required) | api-key (admin required)  
**Streaming:** no  
**Status:** implemented

Approve a pending user registration (if admin approval is required).

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| user-id | string (UUID) | User identifier |

**Request body:** empty JSON object

```json
{}
```

**Response (200):** User approved and can now log in

**Errors:**
- 401: Unauthorized
- 404: User not found or not pending approval
- 500: Server error

---

### `POST /api/v1/users/{user-id}/reset-password`

**Auth:** login-token (admin required) | api-key (admin required)  
**Streaming:** no  
**Status:** implemented

Force password reset for a user (admin only). Nulls the user's password, revokes all login sessions, and sends a password reset email.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| user-id | string (UUID) | User identifier |

**Request body:** empty JSON object

```json
{}
```

**Response (200):** Password reset initiated

```json
{
  "msg": "Password reset initiated. Email sent to user (if configured)."
}
```

**Note:** All existing session tokens for the user are immediately revoked.

**Errors:**
- 401: Unauthorized
- 404: User not found
- 500: Server error

---

### `DELETE /api/v1/users/{user-id}`

**Auth:** login-token (admin required) | api-key (admin required)  
**Streaming:** no  
**Status:** implemented

Delete a user account permanently. All user data, canvases, and content owned by the user are removed.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| user-id | string (UUID) | User identifier |

**Response (204):** No content (user deleted)

**Errors:**
- 401: Unauthorized
- 404: User not found
- 500: Server error

---

## Groups Management

### `GET /api/v1/groups`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all groups on the server.

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| subscribe | boolean | no | false | Enable streaming updates |

**Response (200):** JSON array of group objects

```json
[
  {
    "group-id": "uuid",
    "group-name": "Engineering Team",
    "description": "All engineers and developers",
    "created": "2025-01-15T10:30:00Z",
    "member-count": 12
  }
]
```

**Errors:**
- 401: Unauthorized
- 500: Server error

---

### `GET /api/v1/groups/{group-id}`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

Retrieve a single group by ID.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| group-id | string (UUID) | Group identifier |

**Response (200):** Single group object

**Errors:**
- 404: Group not found
- 500: Server error

---

### `POST /api/v1/groups`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Create a new group.

**Request body:**

```json
{
  "group-name": "Marketing Team",
  "description": "All marketing staff and contractors"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| group-name | string | yes | Name of the group |
| description | string | no | Group description |

**Response (201):** Created group object

**Errors:**
- 400: Bad request (missing group-name)
- 401: Unauthorized
- 409: Conflict (group name already exists)
- 500: Server error

---

### `PATCH /api/v1/groups/{group-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Update group information.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| group-id | string (UUID) | Group identifier |

**Request body:**

```json
{
  "group-name": "Updated Group Name",
  "description": "Updated description"
}
```

All fields are optional.

**Response (200):** Updated group object

**Errors:**
- 400: Bad request
- 401: Unauthorized
- 404: Group not found
- 500: Server error

---

### `DELETE /api/v1/groups/{group-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Delete a group. Users remain on the server but are removed from the group.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| group-id | string (UUID) | Group identifier |

**Response (204):** No content (group deleted)

**Errors:**
- 401: Unauthorized
- 404: Group not found
- 500: Server error

---

## Group Membership

### `GET /api/v1/groups/{group-id}/members`

**Auth:** login-token | api-key  
**Streaming:** yes (subscribe=true)  
**Status:** implemented

List all members of a group.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| group-id | string (UUID) | Group identifier |

**Query parameters:**
| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| subscribe | boolean | no | false | Enable streaming updates |

**Response (200):** JSON array of user objects (members of the group)

```json
[
  {
    "user-id": "uuid",
    "email": "user@example.com",
    "full-name": "John Doe",
    "is-admin": false,
    "avatar-color": "#FF0000",
    "joined-group": "2025-01-15T10:30:00Z"
  }
]
```

**Errors:**
- 404: Group not found
- 500: Server error

---

### `POST /api/v1/groups/{group-id}/members`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Add a user to a group.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| group-id | string (UUID) | Group identifier |

**Request body:**

```json
{
  "user-id": "user-uuid"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| user-id | string (UUID) | yes | User to add to the group |

**Response (201):** Group membership created

**Errors:**
- 400: Bad request (missing user-id)
- 401: Unauthorized
- 404: Group or user not found
- 409: Conflict (user already in group)
- 500: Server error

---

### `DELETE /api/v1/groups/{group-id}/members/{user-id}`

**Auth:** login-token | api-key  
**Streaming:** no  
**Status:** implemented

Remove a user from a group.

**Path parameters:**
| Name | Type | Description |
|---|---|---|
| group-id | string (UUID) | Group identifier |
| user-id | string (UUID) | User identifier |

**Response (204):** No content (membership removed)

**Errors:**
- 401: Unauthorized
- 404: Group or user not found, or user not in group
- 500: Server error
