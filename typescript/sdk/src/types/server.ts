import type { IsoDateTime, Size, Uuid } from "./common.js";

/**
 * Server info as returned by `GET /api/v1/server-info`.
 *
 * Live-server verification: `{api, go, server_id, version}` — UNDERSCORED.
 */
export interface ServerInfo {
  readonly version: string;
  readonly api?: readonly string[];
  readonly server_id?: string;
  readonly go?: string;
}

/**
 * A single server-config element.
 *
 * The spec describes a flat `[]ConfigElement{key, value, type}` shape.
 * Field names are simple lowercase (no hyphens or underscores needed).
 */
export interface ServerConfigEntry {
  readonly key: string;
  readonly value: unknown;
  readonly type?: string;
}

/**
 * PATCH body for `PATCH /api/v1/server-config`.
 *
 * The server accepts an arbitrary partial config object; the wire shape
 * varies by deployment. Typed as a permissive map.
 */
export type UpdateServerConfigRequest = Readonly<Record<string, unknown>>;

/**
 * Send-test-email request body.
 *
 * Spec uses `recipient-email` (hyphenated). The Go SDK sends both
 * `recipient-email` and `recipient_email` defensively.
 */
export interface SendTestEmailRequest {
  readonly "recipient-email": string;
}

/**
 * License information.
 *
 * Live-server verification (2026-05-18): the wire shape is
 * `{edition, has_expired, is_valid, max_clients, seat_model, type}`,
 * all UNDERSCORED. Legacy fields (`status`, `expiry-date`, etc.) are
 * not returned.
 */
export interface License {
  readonly edition: string;
  readonly has_expired: boolean;
  readonly is_valid: boolean;
  readonly max_clients: number;
  readonly seat_model: string;
  readonly type: string;
}

/** License activation request payload. */
export interface LicenseRequestPayload {
  readonly request: string;
}

/** Install-license body — the server expects `{ license: <key-string> }`. */
export interface InstallLicenseRequest {
  readonly license: string;
}

/**
 * Audit log event.
 *
 * Live-server verification (2026-05-18): all fields UNDERSCORED.
 * `id` is an INTEGER, `author_id` and `target_id` may be null.
 */
export interface AuditEntry {
  readonly id: number;
  readonly action: string;
  readonly author_id: number | null;
  readonly target_id: string | null;
  readonly target_type: string;
  readonly ip_address: string;
  readonly created_at: IsoDateTime;
  readonly details?: string;
}

/**
 * Audit log page.
 *
 * The list endpoint returns a flat array. This typed page exists for
 * callers that want pagination metadata; field names use hyphens
 * (`total-count`, `per-page`) per the spec.
 */
export interface AuditLogPage {
  readonly events: readonly AuditEntry[];
  readonly "total-count": number;
  readonly page: number;
  readonly "per-page": number;
}

/**
 * Audit log query parameters.
 *
 * Server expects hyphenated keys (`per-page`, `start-time`, `end-time`,
 * `user-id`) per the spec. The Go SDK sends both hyphen and underscore
 * forms defensively, but TS callers should use the hyphenated form.
 */
export interface AuditLogQuery {
  readonly page?: number;
  readonly "per-page"?: number;
  readonly filter?: string;
  readonly "start-time"?: IsoDateTime;
  readonly "end-time"?: IsoDateTime;
  readonly "user-id"?: number;
  readonly action?: string;
}

/**
 * Connected client.
 *
 * Live-server shape: `{id, name, user_id, created_at}` — UNDERSCORED.
 * Additional fields (state, version, etc.) appear on some server builds.
 */
export interface ConnectedClient {
  readonly id: Uuid;
  readonly name?: string;
  readonly user_id?: string;
  readonly state?: string;
  readonly version?: string;
  readonly address?: string;
  readonly created_at?: IsoDateTime;
  readonly last_seen?: IsoDateTime;
}

/**
 * Client workspace (an open canvas inside a Canvus app).
 *
 * Live-server shape: every field UNDERSCORED — `canvas_id`, `canvas_size`,
 * `info_panel_visible`, `server_id`, `view_rectangle`, `workspace_name`,
 * `workspace_state`.
 */
export interface ClientWorkspace {
  readonly index: number;
  readonly canvas_id: Uuid;
  readonly canvas_size?: Size;
  readonly info_panel_visible?: boolean;
  readonly location?: { readonly x: number; readonly y: number };
  readonly pinned: boolean;
  readonly server_id?: Uuid;
  readonly size?: Size;
  readonly state: string;
  readonly user?: string;
  readonly view_rectangle?: { readonly x: number; readonly y: number; readonly width: number; readonly height: number };
  readonly workspace_name: string;
  readonly workspace_state?: string;
}

/** PATCH body for `PATCH /api/v1/clients/{cid}/workspaces/{wid}`. */
export interface UpdateWorkspaceRequest {
  readonly info_panel_visible?: boolean;
  readonly pinned?: boolean;
  readonly view_rectangle?: { readonly x: number; readonly y: number; readonly width: number; readonly height: number };
}

/** Open-canvas-in-workspace body. */
export interface OpenCanvasInWorkspaceRequest {
  readonly canvas_id: Uuid;
  readonly server_id?: Uuid;
  readonly user_email?: string;
}

/**
 * Video output device (client-scoped).
 *
 * Shape varies by server version — typed defensively with an index signature.
 * Known field names are underscored.
 */
export interface ClientVideoOutput {
  readonly id?: string;
  readonly index?: number;
  readonly name?: string;
  readonly label?: string;
  readonly source?: string;
  readonly suspended?: boolean;
  readonly resolution?: Size;
  readonly state?: string;
  readonly [key: string]: unknown;
}

/** PATCH body for video output. Typed as open map; concrete fields vary. */
export interface UpdateClientVideoOutputRequest {
  readonly source?: string;
  readonly suspended?: boolean;
  readonly [key: string]: unknown;
}

/** Video input device (client-scoped). */
export interface ClientVideoInput {
  readonly id: Uuid;
  readonly name: string;
  readonly source?: string;
  readonly resolution?: string;
  readonly fps?: number;
}
