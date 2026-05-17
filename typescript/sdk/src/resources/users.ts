import type { Transport } from "../transport.js";
import { streamNdjson, type StreamOptions } from "../streaming.js";
import type { Uuid } from "../types/common.js";
import type {
  AddGroupMemberRequest,
  ChangeEmailRequest,
  CreateGroupRequest,
  CreateUserRequest,
  Group,
  UpdateGroupRequest,
  UpdateUserRequest,
  User,
} from "../types/user.js";

/**
 * User and group management endpoints.
 *
 * Excludes login/registration/password-reset/access-tokens which live on
 * {@link AuthResource}.
 */
export class UsersResource {
  constructor(private readonly transport: Transport) {}

  // ---- Users --------------------------------------------------------------

  /** `GET /api/v1/users` — list users. */
  async list(): Promise<readonly User[]> {
    return this.transport.request<readonly User[]>("GET", "users");
  }

  /** Subscribe to the user list. */
  subscribe(opts?: StreamOptions): AsyncGenerator<User, void, void> {
    return streamNdjson<User>(this.transport, "users", opts);
  }

  /** `GET /api/v1/users/{uid}`. */
  async get(userId: Uuid): Promise<User> {
    return this.transport.request<User>("GET", `users/${userId}`);
  }

  /** Subscribe to a single user. */
  subscribeOne(userId: Uuid, opts?: StreamOptions): AsyncGenerator<User, void, void> {
    return streamNdjson<User>(this.transport, `users/${userId}`, opts);
  }

  /** `POST /api/v1/users` — admin-only user creation. */
  async create(body: CreateUserRequest): Promise<User> {
    return this.transport.request<User>("POST", "users", body);
  }

  /** `PATCH /api/v1/users/{uid}` — update profile fields. */
  async update(userId: Uuid, body: UpdateUserRequest): Promise<User> {
    return this.transport.request<User>("PATCH", `users/${userId}`, body);
  }

  /** `POST /api/v1/users/{uid}/change-email` — request an email change. */
  async changeEmail(userId: Uuid, body: ChangeEmailRequest): Promise<{ readonly msg: string }> {
    return this.transport.request<{ readonly msg: string }>(
      "POST",
      `users/${userId}/change-email`,
      body,
    );
  }

  /** `POST /api/v1/users/{uid}/block` — block (admin only). */
  async block(userId: Uuid): Promise<User> {
    return this.transport.request<User>("POST", `users/${userId}/block`, {});
  }

  /** `POST /api/v1/users/{uid}/unblock` — unblock (admin only). */
  async unblock(userId: Uuid): Promise<User> {
    return this.transport.request<User>("POST", `users/${userId}/unblock`, {});
  }

  /** `POST /api/v1/users/{uid}/approve` — approve a pending registration. */
  async approve(userId: Uuid): Promise<User> {
    return this.transport.request<User>("POST", `users/${userId}/approve`, {});
  }

  /** `POST /api/v1/users/{uid}/reset-password` — force a reset (admin only). */
  async forcePasswordReset(userId: Uuid): Promise<{ readonly msg: string }> {
    return this.transport.request<{ readonly msg: string }>(
      "POST",
      `users/${userId}/reset-password`,
      {},
    );
  }

  /** `DELETE /api/v1/users/{uid}` — permanently delete (admin only). */
  async delete(userId: Uuid): Promise<void> {
    await this.transport.request<void>("DELETE", `users/${userId}`);
  }

  // ---- Groups -------------------------------------------------------------

  /** `GET /api/v1/groups`. */
  async listGroups(): Promise<readonly Group[]> {
    return this.transport.request<readonly Group[]>("GET", "groups");
  }

  /** Subscribe to the group list. */
  subscribeGroups(opts?: StreamOptions): AsyncGenerator<Group, void, void> {
    return streamNdjson<Group>(this.transport, "groups", opts);
  }

  /** `GET /api/v1/groups/{gid}`. */
  async getGroup(groupId: Uuid): Promise<Group> {
    return this.transport.request<Group>("GET", `groups/${groupId}`);
  }

  /** Subscribe to a single group. */
  subscribeGroup(groupId: Uuid, opts?: StreamOptions): AsyncGenerator<Group, void, void> {
    return streamNdjson<Group>(this.transport, `groups/${groupId}`, opts);
  }

  /** `POST /api/v1/groups`. */
  async createGroup(body: CreateGroupRequest): Promise<Group> {
    return this.transport.request<Group>("POST", "groups", body);
  }

  /** `PATCH /api/v1/groups/{gid}`. */
  async updateGroup(groupId: Uuid, body: UpdateGroupRequest): Promise<Group> {
    return this.transport.request<Group>("PATCH", `groups/${groupId}`, body);
  }

  /** `DELETE /api/v1/groups/{gid}`. */
  async deleteGroup(groupId: Uuid): Promise<void> {
    await this.transport.request<void>("DELETE", `groups/${groupId}`);
  }

  /** `GET /api/v1/groups/{gid}/members` — list members. */
  async listGroupMembers(groupId: Uuid): Promise<readonly User[]> {
    return this.transport.request<readonly User[]>("GET", `groups/${groupId}/members`);
  }

  /** Subscribe to a group's member list. */
  subscribeGroupMembers(
    groupId: Uuid,
    opts?: StreamOptions,
  ): AsyncGenerator<User, void, void> {
    return streamNdjson<User>(this.transport, `groups/${groupId}/members`, opts);
  }

  /** `POST /api/v1/groups/{gid}/members` — add a user. */
  async addGroupMember(groupId: Uuid, body: AddGroupMemberRequest): Promise<void> {
    await this.transport.request<void>("POST", `groups/${groupId}/members`, body);
  }

  /** `DELETE /api/v1/groups/{gid}/members/{uid}` — remove a user. */
  async removeGroupMember(groupId: Uuid, userId: Uuid): Promise<void> {
    await this.transport.request<void>("DELETE", `groups/${groupId}/members/${userId}`);
  }
}
