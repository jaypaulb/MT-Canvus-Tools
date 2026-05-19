import type { Transport } from "../transport.js";
import { streamNdjson, type StreamOptions } from "../streaming.js";
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
 * User IDs are INTEGERS on the live server (per VERIFIED-CORRECTIONS.md §6);
 * access-token IDs are OPAQUE STRINGS. The path segments accept either
 * form coerced to string.
 */
type UserId = number | string;
type TokenId = string;

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

  /**
   * Phase 4b §4.3 #5: `GET /api/v1/users/current` — fetch the user the
   * configured credential resolves to.
   *
   * Useful for trash helpers (`canvases.trash`, `folders.trash`) that need
   * the calling user's ID to compose the trash folder path.
   */
  async currentUser(): Promise<User> {
    return this.transport.request<User>("GET", "users/current");
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
  async changePassword(userId: UserId, body: ChangePasswordRequest): Promise<User> {
    return this.transport.request<User>("POST", `users/${userId}/password`, body);
  }

  // ---- API access tokens --------------------------------------------------

  /** `GET /api/v1/users/{uid}/access-tokens`. */
  async listAccessTokens(userId: UserId): Promise<readonly AccessToken[]> {
    return this.transport.request<readonly AccessToken[]>(
      "GET",
      `users/${userId}/access-tokens`,
    );
  }

  /** Subscribe to a user's access-token list. */
  subscribeAccessTokens(
    userId: UserId,
    opts?: StreamOptions,
  ): AsyncGenerator<AccessToken, void, void> {
    return streamNdjson<AccessToken>(
      this.transport,
      `users/${userId}/access-tokens`,
      opts,
    );
  }

  /** `GET /api/v1/users/{uid}/access-tokens/{tid}`. */
  async getAccessToken(userId: UserId, tokenId: TokenId): Promise<AccessToken> {
    return this.transport.request<AccessToken>(
      "GET",
      `users/${userId}/access-tokens/${tokenId}`,
    );
  }

  /** Subscribe to a single access-token. */
  subscribeAccessToken(
    userId: UserId,
    tokenId: TokenId,
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
    userId: UserId,
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
    userId: UserId,
    tokenId: TokenId,
    body: UpdateAccessTokenRequest,
  ): Promise<AccessToken> {
    return this.transport.request<AccessToken>(
      "PATCH",
      `users/${userId}/access-tokens/${tokenId}`,
      body,
    );
  }

  /** `DELETE /api/v1/users/{uid}/access-tokens/{tid}` — revoke a token. */
  async deleteAccessToken(userId: UserId, tokenId: TokenId): Promise<void> {
    await this.transport.request<undefined>(
      "DELETE",
      `users/${userId}/access-tokens/${tokenId}`,
    );
  }
}
