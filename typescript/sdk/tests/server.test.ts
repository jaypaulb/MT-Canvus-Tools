// Phase 4d Round C: focused coverage for ServerResource. Exercises every
// method's happy path through a mocked fetch, plus a representative slice
// of the error-mapping branches (401, 404, 429, 5xx).

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  AuthError,
  NotFoundError,
  RateLimitError,
  ServerError,
  createSession,
  flattenServerConfig,
} from "../src/index.js";

const BASE = "https://canvus.example.com/api/v1/";
const KEY = "cv_test_token_abcdef0123456789";

interface MockResponseInit {
  status?: number;
  body?: unknown;
  headers?: Record<string, string>;
}

function mockResponse(init: MockResponseInit): {
  ok: boolean;
  status: number;
  headers: Headers;
  json: () => Promise<unknown>;
  text: () => Promise<string>;
  blob: () => Promise<Blob>;
} {
  const status = init.status ?? 200;
  const headers = new Headers(init.headers ?? { "content-type": "application/json" });
  const bodyText =
    init.body === undefined
      ? ""
      : typeof init.body === "string"
        ? init.body
        : JSON.stringify(init.body);
  return {
    ok: status >= 200 && status < 300,
    status,
    headers,
    json: async () => JSON.parse(bodyText) as unknown,
    text: async () => bodyText,
    blob: async () => new Blob([bodyText]),
  };
}

function setFetchOnce(init: MockResponseInit): ReturnType<typeof vi.fn> {
  const spy = vi.fn().mockResolvedValue(mockResponse(init) as unknown as Response);
  globalThis.fetch = spy as unknown as typeof fetch;
  return spy;
}

const CID = "00000000-0000-4000-8000-000000000001";
const WID = "00000000-0000-4000-8000-000000000002";
const OID = "00000000-0000-4000-8000-000000000003";
const IID = "00000000-0000-4000-8000-000000000004";
const CANVAS_ID = "00000000-0000-4000-8000-0000000000aa";
const WIDGET_ID = "00000000-0000-4000-8000-0000000000bb";

describe("ServerResource — server info & config", () => {
  let originalFetch: typeof fetch;
  beforeEach(() => { originalFetch = globalThis.fetch; });
  afterEach(() => { globalThis.fetch = originalFetch; vi.restoreAllMocks(); });

  it("info() GETs /server-info", async () => {
    const spy = setFetchOnce({ body: { version: "v1.2.3" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const out = await session.server.info();
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/server-info")).toBe(true);
    expect(out.version).toBe("v1.2.3");
  });

  it("config() GETs /server-config", async () => {
    const spy = setFetchOnce({ body: [{ key: "k", value: "v", type: "string" }] });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const out = await session.server.config();
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/server-config")).toBe(true);
    expect(out.length).toBe(1);
  });

  it("configRaw() flattens the nested server-config response", async () => {
    setFetchOnce({
      body: {
        authentication: { password: { enabled: true } },
        port: 8080,
      },
    });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const out = await session.server.configRaw();
    const keys = out.map((e) => e.key);
    expect(keys).toContain("authentication.password.enabled");
    expect(keys).toContain("port");
  });

  it("updateConfig() PATCHes /server-config", async () => {
    const spy = setFetchOnce({ body: [] });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.updateConfig({ port: 9090 });
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/server-config")).toBe(true);
    expect(init.method).toBe("PATCH");
  });

  it("reloadCerts() POSTs /server-config/reload-certs", async () => {
    const spy = setFetchOnce({ body: { msg: "ok" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.reloadCerts();
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/server-config/reload-certs")).toBe(true);
    expect(init.method).toBe("POST");
  });

  it("sendTestEmail() POSTs /server-config/send-test-email", async () => {
    const spy = setFetchOnce({ body: { msg: "sent" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.sendTestEmail({ "recipient-email": "x@y" });
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/server-config/send-test-email")).toBe(true);
    expect(init.method).toBe("POST");
  });

  it("flattenServerConfig() handles null/undefined/array leaves", () => {
    const out = flattenServerConfig({
      a: null,
      b: undefined,
      c: [1, 2, 3],
      d: { e: "string", f: 7 },
    });
    const byKey = new Map(out.map((e) => [e.key, e]));
    expect(byKey.get("a")?.type).toBe("object"); // typeof null === "object"
    expect(byKey.get("b")?.type).toBe("undefined");
    expect(byKey.get("c")?.type).toBe("array");
    expect(byKey.get("d.e")?.value).toBe("string");
    expect(byKey.get("d.f")?.type).toBe("number");
  });
});

describe("ServerResource — license", () => {
  let originalFetch: typeof fetch;
  beforeEach(() => { originalFetch = globalThis.fetch; });
  afterEach(() => { globalThis.fetch = originalFetch; vi.restoreAllMocks(); });

  it("license() GETs /license", async () => {
    const spy = setFetchOnce({
      body: { edition: "pro", has_expired: false, is_valid: true, max_clients: 10, seat_model: "named", type: "perpetual" },
    });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const out = await session.server.license();
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/license")).toBe(true);
    expect(out.is_valid).toBe(true);
  });

  it("licenseRequest() GETs /license/request", async () => {
    const spy = setFetchOnce({ body: { request: "payload" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.licenseRequest();
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/license/request")).toBe(true);
  });

  it("installLicense() POSTs /license", async () => {
    const spy = setFetchOnce({ body: { msg: "installed" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.installLicense({ license: "LICENSE-KEY" });
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/license")).toBe(true);
    expect(init.method).toBe("POST");
  });

  it("activateLicense() POSTs /license/activate", async () => {
    const spy = setFetchOnce({ body: { msg: "ok" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.activateLicense();
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/license/activate")).toBe(true);
    expect(init.method).toBe("POST");
  });
});

describe("ServerResource — audit log", () => {
  let originalFetch: typeof fetch;
  beforeEach(() => { originalFetch = globalThis.fetch; });
  afterEach(() => { globalThis.fetch = originalFetch; vi.restoreAllMocks(); });

  it("auditLog() GETs /audit-log with query encoded", async () => {
    const spy = setFetchOnce({ body: { events: [], "total-count": 0, page: 0, "per-page": 100 } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.auditLog({ page: 1, "per-page": 50 });
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.searchParams.get("page")).toBe("1");
    expect(url.searchParams.get("per-page")).toBe("50");
  });

  it("auditLog() uses default empty query when no args provided", async () => {
    const spy = setFetchOnce({ body: { events: [], "total-count": 0, page: 0, "per-page": 100 } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.auditLog();
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/audit-log")).toBe(true);
  });

  it("exportAuditCsv() GETs /audit-log/export-csv and returns a Blob", async () => {
    const spy = setFetchOnce({
      body: "id,action\n1,create",
      headers: { "content-type": "text/csv" },
    });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const blob = await session.server.exportAuditCsv({ action: "create_canvas" });
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/audit-log/export-csv")).toBe(true);
    const headers = init.headers as Headers;
    expect(headers.get("accept")).toBe("text/csv");
    expect(blob).toBeInstanceOf(Blob);
  });
});

describe("ServerResource — clients & workspaces", () => {
  let originalFetch: typeof fetch;
  beforeEach(() => { originalFetch = globalThis.fetch; });
  afterEach(() => { globalThis.fetch = originalFetch; vi.restoreAllMocks(); });

  it("clients() GETs /clients", async () => {
    const spy = setFetchOnce({ body: [{ id: CID }] });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.clients();
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/clients")).toBe(true);
  });

  it("client() GETs /clients/{cid}", async () => {
    const spy = setFetchOnce({ body: { id: CID } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.client(CID);
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith(`/clients/${CID}`)).toBe(true);
  });

  it("workspaces() GETs /clients/{cid}/workspaces", async () => {
    const spy = setFetchOnce({ body: [] });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.workspaces(CID);
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith(`/clients/${CID}/workspaces`)).toBe(true);
  });

  it("workspace() GETs /clients/{cid}/workspaces/{wid}", async () => {
    const spy = setFetchOnce({
      body: { index: 0, canvas_id: CANVAS_ID, pinned: false, state: "ok", workspace_name: "main" },
    });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.workspace(CID, WID);
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith(`/clients/${CID}/workspaces/${WID}`)).toBe(true);
  });

  it("updateWorkspace() PATCHes /clients/{cid}/workspaces/{wid}", async () => {
    const spy = setFetchOnce({
      body: { index: 0, canvas_id: CANVAS_ID, pinned: true, state: "ok", workspace_name: "main" },
    });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.updateWorkspace(CID, WID, { pinned: true });
    const [, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(init.method).toBe("PATCH");
  });

  it("openCanvasInWorkspace() POSTs /clients/{cid}/workspaces/{wid}/open-canvas", async () => {
    const spy = setFetchOnce({
      body: { index: 0, canvas_id: CANVAS_ID, pinned: false, state: "ok", workspace_name: "main" },
    });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.openCanvasInWorkspace(CID, WID, { canvas_id: CANVAS_ID });
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith(`/clients/${CID}/workspaces/${WID}/open-canvas`)).toBe(true);
    expect(init.method).toBe("POST");
  });

  it("toggleWorkspaceInfoPanel() does GET-then-PATCH, flipping the bit", async () => {
    const spy = vi
      .fn()
      .mockResolvedValueOnce(
        mockResponse({
          body: {
            index: 0, canvas_id: CANVAS_ID, pinned: false, state: "ok",
            workspace_name: "main", info_panel_visible: false,
          },
        }) as unknown as Response,
      )
      .mockResolvedValueOnce(
        mockResponse({
          body: {
            index: 0, canvas_id: CANVAS_ID, pinned: false, state: "ok",
            workspace_name: "main", info_panel_visible: true,
          },
        }) as unknown as Response,
      );
    globalThis.fetch = spy as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const out = await session.server.toggleWorkspaceInfoPanel(CID, WID);
    expect(out.info_panel_visible).toBe(true);
    const [, patchInit] = spy.mock.calls[1] as [URL, RequestInit];
    expect(patchInit.method).toBe("PATCH");
    const body = JSON.parse(patchInit.body as string) as Record<string, unknown>;
    expect(body.info_panel_visible).toBe(true);
  });

  it("toggleWorkspacePinned() does GET-then-PATCH, flipping pinned", async () => {
    const spy = vi
      .fn()
      .mockResolvedValueOnce(
        mockResponse({
          body: {
            index: 0, canvas_id: CANVAS_ID, pinned: true, state: "ok", workspace_name: "main",
          },
        }) as unknown as Response,
      )
      .mockResolvedValueOnce(
        mockResponse({
          body: {
            index: 0, canvas_id: CANVAS_ID, pinned: false, state: "ok", workspace_name: "main",
          },
        }) as unknown as Response,
      );
    globalThis.fetch = spy as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const out = await session.server.toggleWorkspacePinned(CID, WID);
    expect(out.pinned).toBe(false);
  });

  it("setWorkspaceViewport({ rect }) PATCHes with the explicit rect", async () => {
    const spy = setFetchOnce({
      body: { index: 0, canvas_id: CANVAS_ID, pinned: false, state: "ok", workspace_name: "main" },
    });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.setWorkspaceViewport(CID, WID, {
      rect: { x: 0, y: 0, width: 100, height: 100 },
    });
    const [, init] = spy.mock.calls[0] as [URL, RequestInit];
    const body = JSON.parse(init.body as string) as Record<string, unknown>;
    expect(body.view_rectangle).toEqual({ x: 0, y: 0, width: 100, height: 100 });
  });

  it("setWorkspaceViewport({ widgetId }) fetches widget then PATCHes with computed rect", async () => {
    const spy = vi
      .fn()
      .mockResolvedValueOnce(
        mockResponse({
          body: { location: { x: 10, y: 20 }, size: { width: 100, height: 200 } },
        }) as unknown as Response,
      )
      .mockResolvedValueOnce(
        mockResponse({
          body: { index: 0, canvas_id: CANVAS_ID, pinned: false, state: "ok", workspace_name: "main" },
        }) as unknown as Response,
      );
    globalThis.fetch = spy as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.setWorkspaceViewport(CID, WID, {
      canvasId: CANVAS_ID,
      widgetId: WIDGET_ID,
      margin: 5,
    });
    const [, patchInit] = spy.mock.calls[1] as [URL, RequestInit];
    const body = JSON.parse(patchInit.body as string) as { view_rectangle: { x: number; y: number; width: number; height: number } };
    expect(body.view_rectangle.x).toBe(5); // 10 - 5
    expect(body.view_rectangle.y).toBe(15); // 20 - 5
    expect(body.view_rectangle.width).toBe(110); // 100 + 2*5
    expect(body.view_rectangle.height).toBe(210);
  });

  it("setWorkspaceViewport({ widgetId }) uses default margin=20 when omitted", async () => {
    const spy = vi
      .fn()
      .mockResolvedValueOnce(
        mockResponse({
          body: { location: { x: 100, y: 100 }, size: { width: 50, height: 50 } },
        }) as unknown as Response,
      )
      .mockResolvedValueOnce(
        mockResponse({
          body: { index: 0, canvas_id: CANVAS_ID, pinned: false, state: "ok", workspace_name: "main" },
        }) as unknown as Response,
      );
    globalThis.fetch = spy as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.setWorkspaceViewport(CID, WID, {
      canvasId: CANVAS_ID,
      widgetId: WIDGET_ID,
    });
    const [, patchInit] = spy.mock.calls[1] as [URL, RequestInit];
    const body = JSON.parse(patchInit.body as string) as { view_rectangle: { x: number; y: number; width: number; height: number } };
    expect(body.view_rectangle.x).toBe(80);
    expect(body.view_rectangle.width).toBe(90);
  });
});

describe("ServerResource — client video I/O", () => {
  let originalFetch: typeof fetch;
  beforeEach(() => { originalFetch = globalThis.fetch; });
  afterEach(() => { globalThis.fetch = originalFetch; vi.restoreAllMocks(); });

  it("videoOutputs() GETs /clients/{cid}/video-outputs", async () => {
    const spy = setFetchOnce({ body: [] });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.videoOutputs(CID);
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith(`/clients/${CID}/video-outputs`)).toBe(true);
  });

  it("videoOutput() GETs /clients/{cid}/video-outputs/{oid}", async () => {
    const spy = setFetchOnce({ body: { id: OID } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.videoOutput(CID, OID);
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith(`/clients/${CID}/video-outputs/${OID}`)).toBe(true);
  });

  it("updateVideoOutput() PATCHes /clients/{cid}/video-outputs/{oid}", async () => {
    const spy = setFetchOnce({ body: { id: OID } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.updateVideoOutput(CID, OID, { suspended: true });
    const [, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(init.method).toBe("PATCH");
  });

  it("setVideoOutputSourceByIndex() PATCHes using ordinal index", async () => {
    const spy = setFetchOnce({ body: { id: OID } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.setVideoOutputSourceByIndex(CID, 0, { source: "src" });
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith(`/clients/${CID}/video-outputs/0`)).toBe(true);
    expect(init.method).toBe("PATCH");
  });

  it("videoInputs() GETs /clients/{cid}/video-inputs", async () => {
    const spy = setFetchOnce({ body: [] });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.videoInputs(CID);
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith(`/clients/${CID}/video-inputs`)).toBe(true);
  });

  it("videoInput() GETs /clients/{cid}/video-inputs/{iid}", async () => {
    const spy = setFetchOnce({ body: { id: IID, name: "in" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.videoInput(CID, IID);
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith(`/clients/${CID}/video-inputs/${IID}`)).toBe(true);
  });
});

describe("ServerResource — subscribe entrypoints", () => {
  it("subscribe*() methods return async iterables", () => {
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    expect(typeof session.server.subscribeConfig()[Symbol.asyncIterator]).toBe("function");
    expect(typeof session.server.subscribeLicense()[Symbol.asyncIterator]).toBe("function");
    expect(typeof session.server.subscribeClients()[Symbol.asyncIterator]).toBe("function");
    expect(typeof session.server.subscribeClient(CID)[Symbol.asyncIterator]).toBe("function");
    expect(typeof session.server.subscribeWorkspaces(CID)[Symbol.asyncIterator]).toBe("function");
    expect(typeof session.server.subscribeWorkspace(CID, WID)[Symbol.asyncIterator]).toBe("function");
    expect(typeof session.server.subscribeVideoOutputs(CID)[Symbol.asyncIterator]).toBe("function");
    expect(typeof session.server.subscribeVideoOutput(CID, OID)[Symbol.asyncIterator]).toBe("function");
    expect(typeof session.server.subscribeVideoInputs(CID)[Symbol.asyncIterator]).toBe("function");
    expect(typeof session.server.subscribeVideoInput(CID, IID)[Symbol.asyncIterator]).toBe("function");
  });
});

describe("ServerResource — error mapping", () => {
  let originalFetch: typeof fetch;
  beforeEach(() => { originalFetch = globalThis.fetch; });
  afterEach(() => { globalThis.fetch = originalFetch; vi.restoreAllMocks(); });

  it("maps 401 to AuthError on config()", async () => {
    setFetchOnce({ status: 401, body: { msg: "no" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.server.config()).rejects.toBeInstanceOf(AuthError);
  });

  it("maps 404 to NotFoundError on client()", async () => {
    setFetchOnce({ status: 404, body: { msg: "missing" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.server.client(CID)).rejects.toBeInstanceOf(NotFoundError);
  });

  it("maps 429 to RateLimitError on installLicense()", { timeout: 15_000 }, async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValue(
        mockResponse({ status: 429, headers: { "retry-after": "0" }, body: { msg: "slow" } }) as unknown as Response,
      ) as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.server.installLicense({ license: "x" })).rejects.toBeInstanceOf(
      RateLimitError,
    );
  });

  it("maps 500 to ServerError on sendTestEmail()", { timeout: 15_000 }, async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValue(mockResponse({ status: 500, body: { msg: "boom" } }) as unknown as Response) as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(
      session.server.sendTestEmail({ "recipient-email": "x@y" }),
    ).rejects.toBeInstanceOf(ServerError);
  });
});
