import type { Transport } from "../transport.js";
import { streamNdjson, type StreamOptions } from "../streaming.js";
import type { Uuid } from "../types/common.js";
import type {
  AccessToken,
  AccessTokenWithSecret,
  ChangePasswordRequest,
  ConfirmEmailRequest,
  CreateAccessTokenRequest,
  CreateResetTokenRequest,
  LoginRequest,
  LoginResponse,
  RegisterUserRequest,
  RegisterUserResponse,
  ResetPasswordRequest,
  SamlLoginRequest,
  UpdateAccessTokenRequest,
  User,
} from "../types/user.js";

/**
 * Authentication endpoints: login, logout, password reset, registration,
 * email confirmation, and API-access-token management.
 */
export class AuthResource {
  constructor(private readonly transport: Transport) {}

  // ---- Session lifecycle --------------------------------------------------

  /** `POST /api/v1/users/login` — email + password login. */
  async login(body: LoginRequest): Promise<LoginResponse> {
    return this.transport.request<LoginResponse>("POST", "users/login", body);
  }

  /** `POST /api/v1/users/login/saml` — SAML SSO login. */
  async loginSaml(body: SamlLoginRequest): Promise<LoginResponse> {
    return this.transport.request<LoginResponse>("POST", "users/login/saml", body);
  }

  /** `POST /api/v1/users/logout` — invalidate the current session token. */
  async logout(): Promise<{ readonly msg: string }> {
    return this.transport.request<{ readonly msg: string }>("POST", "users/logout", {});
  }

  // ---- Password reset (unauthenticated) -----------------------------------

  /** `POST /api/v1/users/password/create-reset-token`. */
  async createResetToken(body: CreateResetTokenRequest): Promise<{ readonly msg: string }> {
    return this.transport.request<{ readonly msg: string }>(
      "POST",
      "users/password/create-reset-token",
      body,
    );
  }

  /** `GET /api/v1/users/password/validate-reset-token`. */
  async validateResetToken(token: string): Promise<{ readonly valid: boolean; readonly msg?: string }> {
    return this.transport.request<{ readonly valid: boolean; readonly msg?: string }>(
      "GET",
      "users/password/validate-reset-token",
      undefined,
      { query: { token } },
    );
  }

  /** `POST /api/v1/users/password/reset`. */
  async resetPassword(body: ResetPasswordRequest): Promise<{ readonly msg: string }> {
    return this.transport.request<{ readonly msg: string }>(
      "POST",
      "users/password/reset",
      body,
    );
  }

  // ---- Self-registration --------------------------------------------------

  /** `POST /api/v1/users/register` (only if enabled server-side). */
  async register(body: RegisterUserRequest): Promise<RegisterUserResponse> {
    return this.transport.request<RegisterUserResponse>("POST", "users/register", body);
  }

  /** `POST /api/v1/users/confirm-email`. */
  async confirmEmail(body: ConfirmEmailRequest): Promise<{ readonly msg: string }> {
    return this.transport.request<{ readonly msg: string }>(
      "POST",
      "users/confirm-email",
      body,
    );
  }

  // ---- Per-user password (for an authenticated session) -------------------

  /** `POST /api/v1/users/{uid}/password` — change a user's password. */
  async changePassword(userId: Uuid, body: ChangePasswordRequest): Promise<User> {
    return this.transport.request<User>("POST", `users/${userId}/password`, body);
  }

  // ---- API access tokens --------------------------------------------------

  /** `GET /api/v1/users/{uid}/access-tokens`. */
  async listAccessTokens(userId: Uuid): Promise<readonly AccessToken[]> {
    return this.transport.request<readonly AccessToken[]>(
      "GET",
      `users/${userId}/access-tokens`,
    );
  }

  /** Subscribe to a user's access-token list. */
  subscribeAccessTokens(
    userId: Uuid,
    opts?: StreamOptions,
  ): AsyncGenerator<AccessToken, void, void> {
    return streamNdjson<AccessToken>(
      this.transport,
      `users/${userId}/access-tokens`,
      opts,
    );
  }

  /** `GET /api/v1/users/{uid}/access-tokens/{tid}`. */
  async getAccessToken(userId: Uuid, tokenId: Uuid): Promise<AccessToken> {
    return this.transport.request<AccessToken>(
      "GET",
      `users/${userId}/access-tokens/${tokenId}`,
    );
  }

  /** Subscribe to a single access-token. */
  subscribeAccessToken(
    userId: Uuid,
    tokenId: Uuid,
    opts?: StreamOptions,
  ): AsyncGenerator<AccessToken, void, void> {
    return streamNdjson<AccessToken>(
      this.transport,
      `users/${userId}/access-tokens/${tokenId}`,
      opts,
    );
  }

  /**
   * `POST /api/v1/users/{uid}/access-tokens`.
   *
   * The `token` value in the response is the only time the secret is
   * returned. Store it securely; the SDK does not persist it.
   */
  async createAccessToken(
    userId: Uuid,
    body: CreateAccessTokenRequest,
  ): Promise<AccessTokenWithSecret> {
    return this.transport.request<AccessTokenWithSecret>(
      "POST",
      `users/${userId}/access-tokens`,
      body,
    );
  }

  /** `PATCH /api/v1/users/{uid}/access-tokens/{tid}`. */
  async updateAccessToken(
    userId: Uuid,
    tokenId: Uuid,
    body: UpdateAccessTokenRequest,
  ): Promise<AccessToken> {
    return this.transport.request<AccessToken>(
      "PATCH",
      `users/${userId}/access-tokens/${tokenId}`,
      body,
    );
  }

  /** `DELETE /api/v1/users/{uid}/access-tokens/{tid}` — revoke a token. */
  async deleteAccessToken(userId: Uuid, tokenId: Uuid): Promise<void> {
    await this.transport.request<void>(
      "DELETE",
      `users/${userId}/access-tokens/${tokenId}`,
    );
  }
}
