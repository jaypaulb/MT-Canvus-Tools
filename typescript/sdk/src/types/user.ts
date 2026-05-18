import type { IsoDateTime } from "./common.js";

/**
 * A user account as returned by the API.
 *
 * Live-server verification (2026-05-18): `id` is an INTEGER (not a UUID);
 * field names are UNDERSCORED. `name` (not `full-name`), `admin` (not
 * `is-admin`), `blocked` (not `is-blocked`).
 */
export interface User {
  readonly id: number;
  readonly email: string;
  readonly name: string;
  readonly admin: boolean;
  readonly approved?: boolean;
  readonly blocked?: boolean;
  readonly created_at?: IsoDateTime;
  readonly last_login?: IsoDateTime | null;
  readonly state?: string;
}

/** Login request body. */
export interface LoginRequest {
  readonly email: string;
  readonly password: string;
  readonly remember?: boolean;
}

/** Login response shape. */
export interface LoginResponse {
  readonly token: string;
  readonly user: User;
}

/** SAML login request body. */
export interface SamlLoginRequest {
  readonly inResponseTo: string;
  readonly responseXml: string;
  readonly remember?: boolean;
}

/** Registration request body. */
export interface RegisterUserRequest {
  readonly email: string;
  readonly name: string;
  readonly password: string;
}

/** Registration response. */
export interface RegisterUserResponse {
  readonly msg: string;
  readonly user: User;
}

/** Admin-side user creation request body. */
export interface CreateUserRequest {
  readonly email: string;
  readonly name: string;
  readonly password: string;
  readonly admin?: boolean;
  readonly approved?: boolean;
  readonly blocked?: boolean;
}

/** Patch shape for `PATCH /users/{id}`. */
export interface UpdateUserRequest {
  readonly email?: string;
  readonly name?: string;
  readonly password?: string;
  readonly admin?: boolean;
  readonly approved?: boolean;
  readonly blocked?: boolean;
}

/**
 * Password change request body.
 *
 * The Canvus spec uses hyphenated keys here (`old-password`, `new-password`);
 * this is one of the few PATCH/POST bodies where hyphens are required.
 */
export interface ChangePasswordRequest {
  readonly "old-password"?: string;
  readonly "new-password": string;
}

/**
 * Email change request body.
 *
 * The Canvus spec uses `new-email` (hyphenated).
 */
export interface ChangeEmailRequest {
  readonly "new-email": string;
}

/** Password reset token request. */
export interface CreateResetTokenRequest {
  readonly email: string;
}

/** Password reset request. */
export interface ResetPasswordRequest {
  readonly token: string;
  readonly password: string;
}

/** Email-confirmation request body. */
export interface ConfirmEmailRequest {
  readonly token: string;
}

/**
 * API access token (metadata; the secret is only returned on create).
 *
 * Live-server verification: `id` is an OPAQUE STRING token (not a UUID
 * integer), `description` is the user-facing label, and timestamps use
 * `created_at` (underscored).
 */
export interface AccessToken {
  readonly id: string;
  readonly description: string;
  readonly created_at: IsoDateTime;
  readonly name?: string;
  readonly expires?: IsoDateTime | null;
  readonly scopes?: readonly string[];
}

/** {@link AccessToken} returned on creation, including the secret token. */
export interface AccessTokenWithSecret extends AccessToken {
  readonly plain_token: string;
}

/** Create-token request body. */
export interface CreateAccessTokenRequest {
  readonly description?: string;
  readonly name?: string;
  readonly expires?: IsoDateTime | null;
  readonly scopes?: readonly string[];
}

/** Patch-token request body. */
export interface UpdateAccessTokenRequest {
  readonly description?: string;
  readonly name?: string;
  readonly expires?: IsoDateTime | null;
  readonly scopes?: readonly string[];
}

/**
 * A group as returned by the API.
 *
 * Live-server verification: `id` is an INTEGER, fields are UNDERSCORED.
 */
export interface Group {
  readonly id: number;
  readonly name: string;
  readonly description?: string;
}

/** Create-group request body. */
export interface CreateGroupRequest {
  readonly name: string;
  readonly description?: string;
}

/** Update-group request body. */
export interface UpdateGroupRequest {
  readonly name?: string;
  readonly description?: string;
}

/** Add-member request body. The server expects `{ id: <integer-user-id> }`. */
export interface AddGroupMemberRequest {
  readonly id: number;
}
