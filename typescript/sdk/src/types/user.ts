import type { IsoDateTime, Uuid } from "./common.js";

/** A user account as returned by the API. */
export interface User {
  readonly "user-id": number;
  readonly email: string;
  readonly "full-name": string;
  readonly "is-admin": boolean;
  readonly "is-blocked": boolean;
  readonly groups?: readonly Uuid[];
  readonly "avatar-color"?: string;
  readonly created?: IsoDateTime;
  readonly "last-login"?: IsoDateTime;
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
  readonly "full-name": string;
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
  readonly "full-name": string;
  readonly password: string;
  readonly "is-admin"?: boolean;
  readonly "avatar-color"?: string;
}

/** Patch shape for `PATCH /users/{id}`. */
export interface UpdateUserRequest {
  readonly "full-name"?: string;
  readonly "avatar-color"?: string;
  readonly "is-admin"?: boolean;
  readonly "is-blocked"?: boolean;
}

/** Password change request body. */
export interface ChangePasswordRequest {
  readonly "old-password"?: string;
  readonly "new-password": string;
}

/** Email change request body. */
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

/** API access token (metadata; `token` is only returned on create). */
export interface AccessToken {
  readonly "token-id": Uuid;
  readonly name: string;
  readonly created: IsoDateTime;
  readonly "last-used"?: IsoDateTime | null;
  readonly expires?: IsoDateTime | null;
  readonly scopes: readonly string[];
  readonly prefix: string;
}

/** {@link AccessToken} returned on creation, including the secret `token`. */
export interface AccessTokenWithSecret extends AccessToken {
  readonly token: string;
}

/** Create-token request body. */
export interface CreateAccessTokenRequest {
  readonly name: string;
  readonly expires?: IsoDateTime | null;
  readonly scopes?: readonly string[];
}

/** Patch-token request body. */
export interface UpdateAccessTokenRequest {
  readonly name?: string;
  readonly expires?: IsoDateTime | null;
  readonly scopes?: readonly string[];
}

/** A group as returned by the API. */
export interface Group {
  readonly "group-id": Uuid;
  readonly "group-name": string;
  readonly description?: string;
  readonly created?: IsoDateTime;
  readonly "member-count"?: number;
}

/** Create-group request body. */
export interface CreateGroupRequest {
  readonly "group-name": string;
  readonly description?: string;
}

/** Update-group request body. */
export interface UpdateGroupRequest {
  readonly "group-name"?: string;
  readonly description?: string;
}

/** Add-member request body. */
export interface AddGroupMemberRequest {
  readonly "user-id": Uuid;
}
