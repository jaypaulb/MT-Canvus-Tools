import type { IsoDateTime, Size, Uuid } from "./common.js";

/** Server info as returned by `GET /api/v1/server-info`. */
export interface ServerInfo {
  readonly version: string;
  readonly commit?: string;
  readonly "build-date"?: IsoDateTime;
  readonly "go-version"?: string;
  readonly platform?: string;
}

/** A single server-config element. */
export interface ServerConfigEntry {
  readonly "setting-key": string;
  readonly "setting-value": string;
  readonly "setting-type"?: "string" | "boolean" | "integer" | "number";
}

/** PATCH body for `PATCH /api/v1/server-config`. */
export interface UpdateServerConfigRequest {
  readonly settings: ReadonlyArray<Pick<ServerConfigEntry, "setting-key" | "setting-value">>;
}

/** Send-test-email request body. */
export interface SendTestEmailRequest {
  readonly "recipient-email": string;
}

/** License information. */
export interface License {
  readonly status: "valid" | "expired" | "invalid";
  readonly clients: number;
  readonly "max-clients": number;
  readonly valid: boolean;
  readonly message?: string;
  readonly "expiry-date"?: IsoDateTime;
  readonly "seat-model"?: "usage_reported" | "fixed_seats" | "none";
  readonly "activation-required"?: boolean;
}

/** License activation request payload. */
export interface LicenseRequestPayload {
  readonly request: string;
}

/** Install-license body. */
export interface InstallLicenseRequest {
  readonly "license-data": string;
}

/** Audit log event. */
export interface AuditEntry {
  readonly "event-id": Uuid;
  readonly timestamp: IsoDateTime;
  readonly "user-id"?: Uuid;
  readonly "user-email"?: string;
  readonly action: string;
  readonly "resource-type"?: string;
  readonly "resource-id"?: string;
  readonly changes?: Readonly<Record<string, unknown>>;
  readonly "ip-address"?: string;
  readonly "user-agent"?: string;
}

/** Audit log paged response. */
export interface AuditLogPage {
  readonly events: readonly AuditEntry[];
  readonly "total-count": number;
  readonly page: number;
  readonly "per-page": number;
}

/** Audit log query parameters. */
export interface AuditLogQuery {
  readonly page?: number;
  readonly "per-page"?: number;
  readonly filter?: string;
  readonly "start-time"?: IsoDateTime;
  readonly "end-time"?: IsoDateTime;
  readonly "user-id"?: Uuid;
  readonly action?: string;
}

/** Connected client. */
export interface ConnectedClient {
  readonly "client-id": Uuid;
  readonly "app-name": string;
  readonly "app-version": string;
  readonly "connected-at": IsoDateTime;
  readonly "last-activity": IsoDateTime;
  readonly "ip-address"?: string;
  readonly "user-email"?: string;
  readonly "workspace-count"?: number;
}

/** Client workspace (an open canvas inside a Canvus app). */
export interface ClientWorkspace {
  readonly "workspace-id": Uuid;
  readonly "workspace-name": string;
  readonly "canvas-id": Uuid;
  readonly "canvas-name": string;
  readonly size: Size;
  readonly "canvas-size": Size;
  readonly "workspace-state": "active" | "minimized" | "archived";
  readonly "view-location": { readonly x: number; readonly y: number };
  readonly "view-scale": number;
  readonly pinned: boolean;
  readonly "info-panel-visible"?: boolean;
  readonly owner?: string;
  readonly "server-id"?: Uuid;
}

/** PATCH body for `PATCH /api/v1/clients/{cid}/workspaces/{wid}`. */
export interface UpdateWorkspaceRequest {
  readonly "view-location"?: { readonly x: number; readonly y: number };
  readonly "view-scale"?: number;
  readonly pinned?: boolean;
  readonly "info-panel-visible"?: boolean;
}

/** Open-canvas-in-workspace body. */
export interface OpenCanvasInWorkspaceRequest {
  readonly "canvas-id": Uuid;
}

/** Video output device (client-scoped). */
export interface ClientVideoOutput {
  readonly "output-id": Uuid;
  readonly name?: string;
  readonly resolution?: string;
  readonly "canvas-id"?: Uuid;
  readonly [key: string]: unknown;
}

/** PATCH body for video output. */
export interface UpdateClientVideoOutputRequest {
  readonly [key: string]: unknown;
}

/** Video input device (client-scoped). */
export interface ClientVideoInput {
  readonly "input-id": Uuid;
  readonly name: string;
  readonly source: string;
  readonly resolution?: string;
  readonly fps?: number;
}
