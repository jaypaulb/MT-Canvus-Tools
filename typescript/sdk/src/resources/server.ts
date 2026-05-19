import type { Transport } from "../transport.js";
import { streamNdjson, type StreamOptions } from "../streaming.js";
import type { Uuid } from "../types/common.js";
import type {
  AuditLogPage,
  AuditLogQuery,
  ClientVideoInput,
  ClientVideoOutput,
  ClientWorkspace,
  ConnectedClient,
  InstallLicenseRequest,
  License,
  LicenseRequestPayload,
  OpenCanvasInWorkspaceRequest,
  SendTestEmailRequest,
  ServerConfigEntry,
  ServerInfo,
  UpdateClientVideoOutputRequest,
  UpdateServerConfigRequest,
  UpdateWorkspaceRequest,
} from "../types/server.js";

/**
 * Server administration endpoints — server info, config, license, audit
 * log, connected clients, workspaces, and client-side video I/O.
 *
 * Implements the 22 endpoints documented in
 * `docs/api-reference/endpoints/server.md`.
 */
export class ServerResource {
  constructor(private readonly transport: Transport) {}

  // ---- Server info & config ----------------------------------------------

  /** `GET /api/v1/server-info` — no auth required. */
  async info(): Promise<ServerInfo> {
    return this.transport.request<ServerInfo>("GET", "server-info");
  }

  /** `GET /api/v1/server-config` — list of settings. */
  async config(): Promise<readonly ServerConfigEntry[]> {
    return this.transport.request<readonly ServerConfigEntry[]>("GET", "server-config");
  }

  /**
   * Phase 4b §4.3 #7: `GET /api/v1/server-config` flattened into the spec's
   * documented element-array form `[{key, value, type}, ...]`.
   *
   * The live server returns a nested config object (see VERIFIED-CORRECTIONS §3);
   * this helper converts it to the flat shape the spec describes, by depth-first
   * walking nested keys joined with `.`.
   */
  async configRaw(): Promise<readonly ServerConfigEntry[]> {
    const nested = await this.transport.request<Record<string, unknown>>("GET", "server-config");
    return flattenServerConfig(nested);
  }

  /** Subscribe to server-config updates. */
  subscribeConfig(opts?: StreamOptions): AsyncGenerator<ServerConfigEntry, void, void> {
    return streamNdjson<ServerConfigEntry>(this.transport, "server-config", opts);
  }

  /** `PATCH /api/v1/server-config` — admin only. */
  async updateConfig(body: UpdateServerConfigRequest): Promise<readonly ServerConfigEntry[]> {
    return this.transport.request<readonly ServerConfigEntry[]>(
      "PATCH",
      "server-config",
      body,
    );
  }

  /** `POST /api/v1/server-config/reload-certs`. */
  async reloadCerts(): Promise<{ readonly msg: string }> {
    return this.transport.request<{ readonly msg: string }>(
      "POST",
      "server-config/reload-certs",
      {},
    );
  }

  /** `POST /api/v1/server-config/send-test-email`. */
  async sendTestEmail(body: SendTestEmailRequest): Promise<{ readonly msg: string }> {
    return this.transport.request<{ readonly msg: string }>(
      "POST",
      "server-config/send-test-email",
      body,
    );
  }

  // ---- License -----------------------------------------------------------

  /** `GET /api/v1/license`. */
  async license(): Promise<License> {
    return this.transport.request<License>("GET", "license");
  }

  /** Subscribe to license-state updates. */
  subscribeLicense(opts?: StreamOptions): AsyncGenerator<License, void, void> {
    return streamNdjson<License>(this.transport, "license", opts);
  }

  /** `GET /api/v1/license/request` — offline activation payload. */
  async licenseRequest(): Promise<LicenseRequestPayload> {
    return this.transport.request<LicenseRequestPayload>("GET", "license/request");
  }

  /** `POST /api/v1/license` — install a license (typically offline). */
  async installLicense(body: InstallLicenseRequest): Promise<{ readonly msg: string }> {
    return this.transport.request<{ readonly msg: string }>("POST", "license", body);
  }

  /** `POST /api/v1/license/activate` — online activation. */
  async activateLicense(): Promise<{ readonly msg: string }> {
    return this.transport.request<{ readonly msg: string }>("POST", "license/activate", {});
  }

  // ---- Audit log ---------------------------------------------------------

  /**
   * `GET /api/v1/audit-log` — paginated audit events.
   *
   * Subscribe is not supported by the server; use {@link exportAuditCsv}
   * for offline analysis.
   */
  async auditLog(query: AuditLogQuery = {}): Promise<AuditLogPage> {
    return this.transport.request<AuditLogPage>("GET", "audit-log", undefined, {
      query: query as Record<string, string | number | boolean | undefined>,
    });
  }

  /** `GET /api/v1/audit-log/export-csv` — returns the CSV body as a `Blob`. */
  async exportAuditCsv(query: Omit<AuditLogQuery, "page" | "per-page"> = {}): Promise<Blob> {
    const response = await this.transport.rawRequest("GET", "audit-log/export-csv", undefined, {
      query: query,
      accept: "text/csv",
    });
    return response.blob();
  }

  // ---- Connected clients & workspaces ------------------------------------

  /** `GET /api/v1/clients` — list connected Canvus clients. */
  async clients(): Promise<readonly ConnectedClient[]> {
    return this.transport.request<readonly ConnectedClient[]>("GET", "clients");
  }

  /** Subscribe to the connected-clients list. */
  subscribeClients(opts?: StreamOptions): AsyncGenerator<ConnectedClient, void, void> {
    return streamNdjson<ConnectedClient>(this.transport, "clients", opts);
  }

  /** `GET /api/v1/clients/{cid}`. */
  async client(clientId: Uuid): Promise<ConnectedClient> {
    return this.transport.request<ConnectedClient>("GET", `clients/${clientId}`);
  }

  /** Subscribe to a single connected client. */
  subscribeClient(
    clientId: Uuid,
    opts?: StreamOptions,
  ): AsyncGenerator<ConnectedClient, void, void> {
    return streamNdjson<ConnectedClient>(this.transport, `clients/${clientId}`, opts);
  }

  /** `GET /api/v1/clients/{cid}/workspaces`. */
  async workspaces(clientId: Uuid): Promise<readonly ClientWorkspace[]> {
    return this.transport.request<readonly ClientWorkspace[]>(
      "GET",
      `clients/${clientId}/workspaces`,
    );
  }

  /** Subscribe to a client's workspace list. */
  subscribeWorkspaces(
    clientId: Uuid,
    opts?: StreamOptions,
  ): AsyncGenerator<ClientWorkspace, void, void> {
    return streamNdjson<ClientWorkspace>(
      this.transport,
      `clients/${clientId}/workspaces`,
      opts,
    );
  }

  /** `GET /api/v1/clients/{cid}/workspaces/{wid}`. */
  async workspace(clientId: Uuid, workspaceId: Uuid): Promise<ClientWorkspace> {
    return this.transport.request<ClientWorkspace>(
      "GET",
      `clients/${clientId}/workspaces/${workspaceId}`,
    );
  }

  /** Subscribe to a single workspace. */
  subscribeWorkspace(
    clientId: Uuid,
    workspaceId: Uuid,
    opts?: StreamOptions,
  ): AsyncGenerator<ClientWorkspace, void, void> {
    return streamNdjson<ClientWorkspace>(
      this.transport,
      `clients/${clientId}/workspaces/${workspaceId}`,
      opts,
    );
  }

  /** `PATCH /api/v1/clients/{cid}/workspaces/{wid}`. */
  async updateWorkspace(
    clientId: Uuid,
    workspaceId: Uuid,
    body: UpdateWorkspaceRequest,
  ): Promise<ClientWorkspace> {
    return this.transport.request<ClientWorkspace>(
      "PATCH",
      `clients/${clientId}/workspaces/${workspaceId}`,
      body,
    );
  }

  /** `POST /api/v1/clients/{cid}/workspaces/{wid}/open-canvas`. */
  async openCanvasInWorkspace(
    clientId: Uuid,
    workspaceId: Uuid,
    body: OpenCanvasInWorkspaceRequest,
  ): Promise<ClientWorkspace> {
    return this.transport.request<ClientWorkspace>(
      "POST",
      `clients/${clientId}/workspaces/${workspaceId}/open-canvas`,
      body,
    );
  }

  // ---- Client video I/O --------------------------------------------------

  /** `GET /api/v1/clients/{cid}/video-outputs`. */
  async videoOutputs(clientId: Uuid): Promise<readonly ClientVideoOutput[]> {
    return this.transport.request<readonly ClientVideoOutput[]>(
      "GET",
      `clients/${clientId}/video-outputs`,
    );
  }

  /** Subscribe to a client's video-output list. */
  subscribeVideoOutputs(
    clientId: Uuid,
    opts?: StreamOptions,
  ): AsyncGenerator<ClientVideoOutput, void, void> {
    return streamNdjson<ClientVideoOutput>(
      this.transport,
      `clients/${clientId}/video-outputs`,
      opts,
    );
  }

  /** `GET /api/v1/clients/{cid}/video-outputs/{oid}`. */
  async videoOutput(clientId: Uuid, outputId: Uuid): Promise<ClientVideoOutput> {
    return this.transport.request<ClientVideoOutput>(
      "GET",
      `clients/${clientId}/video-outputs/${outputId}`,
    );
  }

  /** Subscribe to a single video output. */
  subscribeVideoOutput(
    clientId: Uuid,
    outputId: Uuid,
    opts?: StreamOptions,
  ): AsyncGenerator<ClientVideoOutput, void, void> {
    return streamNdjson<ClientVideoOutput>(
      this.transport,
      `clients/${clientId}/video-outputs/${outputId}`,
      opts,
    );
  }

  /** `PATCH /api/v1/clients/{cid}/video-outputs/{oid}`. */
  async updateVideoOutput(
    clientId: Uuid,
    outputId: Uuid,
    body: UpdateClientVideoOutputRequest,
  ): Promise<ClientVideoOutput> {
    return this.transport.request<ClientVideoOutput>(
      "PATCH",
      `clients/${clientId}/video-outputs/${outputId}`,
      body,
    );
  }

  /**
   * Phase 4b §4.3 #8: set the source for the Nth video output on a client.
   *
   * Mirrors Go's `videooutputs.go:28 SetVideoOutputSource`. Uses the ordinal
   * index rather than the opaque output ID — convenient when the operator
   * knows "the first output" but not the server-assigned UUID.
   */
  async setVideoOutputSourceByIndex(
    clientId: Uuid,
    index: number,
    body: UpdateClientVideoOutputRequest,
  ): Promise<ClientVideoOutput> {
    return this.transport.request<ClientVideoOutput>(
      "PATCH",
      `clients/${clientId}/video-outputs/${index.toString()}`,
      body,
    );
  }

  // ---- Phase 4b §4.3 #9 — workspace orchestration helpers ----------------

  /**
   * Toggle the `info_panel_visible` flag on a workspace.
   *
   * Fetches current state, flips the bit, PATCHes back. Mirrors Go's
   * `workspaces.go:79 ToggleWorkspaceInfoPanel`.
   */
  async toggleWorkspaceInfoPanel(
    clientId: Uuid,
    workspaceId: Uuid,
  ): Promise<ClientWorkspace> {
    const ws = await this.workspace(clientId, workspaceId);
    return this.updateWorkspace(clientId, workspaceId, {
      info_panel_visible: !(ws.info_panel_visible ?? false),
    });
  }

  /**
   * Toggle the `pinned` flag on a workspace. Mirrors Go's
   * `workspaces.go:90 ToggleWorkspacePinned`.
   */
  async toggleWorkspacePinned(
    clientId: Uuid,
    workspaceId: Uuid,
  ): Promise<ClientWorkspace> {
    const ws = await this.workspace(clientId, workspaceId);
    return this.updateWorkspace(clientId, workspaceId, { pinned: !ws.pinned });
  }

  /**
   * Set the workspace viewport.
   *
   * Two modes — pass `{ widgetId, margin? }` to center on a widget
   * (the helper fetches the widget to compute its bounding rectangle plus
   * margin), or pass `{ rect }` to set an explicit view rectangle. Mirrors
   * Go's `workspaces.go:102 SetWorkspaceViewport`.
   */
  async setWorkspaceViewport(
    clientId: Uuid,
    workspaceId: Uuid,
    opts: SetWorkspaceViewportOptions,
  ): Promise<ClientWorkspace> {
    let rect: NonNullable<UpdateWorkspaceRequest["view_rectangle"]>;
    if ("rect" in opts) {
      rect = opts.rect;
    } else {
      const widget = await this.transport.request<{
        location: { x: number; y: number };
        size: { width: number; height: number };
      }>(
        "GET",
        `canvases/${opts.canvasId}/widgets/${opts.widgetId}`,
      );
      const margin = opts.margin ?? 20;
      rect = {
        x: widget.location.x - margin,
        y: widget.location.y - margin,
        width: widget.size.width + 2 * margin,
        height: widget.size.height + 2 * margin,
      };
    }
    return this.updateWorkspace(clientId, workspaceId, { view_rectangle: rect });
  }

  /** `GET /api/v1/clients/{cid}/video-inputs`. */
  async videoInputs(clientId: Uuid): Promise<readonly ClientVideoInput[]> {
    return this.transport.request<readonly ClientVideoInput[]>(
      "GET",
      `clients/${clientId}/video-inputs`,
    );
  }

  /** Subscribe to a client's video-input list. */
  subscribeVideoInputs(
    clientId: Uuid,
    opts?: StreamOptions,
  ): AsyncGenerator<ClientVideoInput, void, void> {
    return streamNdjson<ClientVideoInput>(
      this.transport,
      `clients/${clientId}/video-inputs`,
      opts,
    );
  }

  /** `GET /api/v1/clients/{cid}/video-inputs/{iid}`. */
  async videoInput(clientId: Uuid, inputId: Uuid): Promise<ClientVideoInput> {
    return this.transport.request<ClientVideoInput>(
      "GET",
      `clients/${clientId}/video-inputs/${inputId}`,
    );
  }

  /** Subscribe to a single client-scoped video input. */
  subscribeVideoInput(
    clientId: Uuid,
    inputId: Uuid,
    opts?: StreamOptions,
  ): AsyncGenerator<ClientVideoInput, void, void> {
    return streamNdjson<ClientVideoInput>(
      this.transport,
      `clients/${clientId}/video-inputs/${inputId}`,
      opts,
    );
  }
}

/**
 * Phase 4b §4.3 #9: options for {@link ServerResource.setWorkspaceViewport}.
 *
 * Pass `{ rect }` to set an explicit view rectangle. Pass
 * `{ canvasId, widgetId, margin? }` to center the viewport on a widget;
 * the helper looks up the widget to compute the bounding rectangle.
 */
export type SetWorkspaceViewportOptions =
  | { readonly rect: { readonly x: number; readonly y: number; readonly width: number; readonly height: number } }
  | { readonly canvasId: Uuid; readonly widgetId: Uuid; readonly margin?: number };

/**
 * Phase 4b §4.3 #7: flatten a nested server-config object into the spec's
 * documented `[{key, value, type}, ...]` element-array form.
 *
 * Nested objects produce dotted keys (e.g. `authentication.password.enabled`).
 * Leaf values (primitives, arrays) become entries directly.
 */
export function flattenServerConfig(nested: Record<string, unknown>): readonly ServerConfigEntry[] {
  const out: ServerConfigEntry[] = [];
  const walk = (prefix: string, value: unknown): void => {
    if (value === null || value === undefined) {
      out.push({ key: prefix, value, type: typeof value });
      return;
    }
    if (Array.isArray(value)) {
      out.push({ key: prefix, value, type: "array" });
      return;
    }
    if (typeof value === "object") {
      for (const [k, v] of Object.entries(value as Record<string, unknown>)) {
        const nextKey = prefix === "" ? k : `${prefix}.${k}`;
        walk(nextKey, v);
      }
      return;
    }
    out.push({ key: prefix, value, type: typeof value });
  };
  walk("", nested);
  return out;
}
