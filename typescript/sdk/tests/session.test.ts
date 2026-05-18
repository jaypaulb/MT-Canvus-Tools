import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  APIError,
  AuthError,
  NotFoundError,
  ValidationError,
  createSession,
  loadConfig,
} from "../src/index.js";

const BASE = "https://canvus.example.com/api/v1/";
const KEY = "cv_test_token_abcdef0123456789";

interface MockResponseInit {
  status?: number;
  body?: unknown;
  headers?: Record<string, string>;
}

function mockFetchOnce(init: MockResponseInit): ReturnType<typeof vi.fn> {
  const status = init.status ?? 200;
  const headers = new Headers(init.headers ?? { "content-type": "application/json" });
  const bodyText =
    init.body === undefined
      ? ""
      : typeof init.body === "string"
        ? init.body
        : JSON.stringify(init.body);
  const fakeResp = {
    ok: status >= 200 && status < 300,
    status,
    headers,
    json: async (): Promise<unknown> => JSON.parse(bodyText) as unknown,
    text: async (): Promise<string> => bodyText,
    blob: async (): Promise<Blob> => new Blob([bodyText]),
  };
  const spy = vi.fn().mockResolvedValueOnce(fakeResp as unknown as Response);
  globalThis.fetch = spy as unknown as typeof fetch;
  return spy;
}

describe("createSession", () => {
  it("constructs a Session with all resource namespaces", () => {
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    expect(session.canvases).toBeDefined();
    expect(session.widgets).toBeDefined();
    expect(session.widgets.notes).toBeDefined();
    expect(session.widgets.ipVideos).toBeDefined();
    expect(session.widgets.rdpConnections).toBeDefined();
    expect(session.auth).toBeDefined();
    expect(session.users).toBeDefined();
    expect(session.folders).toBeDefined();
    expect(session.assets).toBeDefined();
    expect(session.server).toBeDefined();
  });

  it("rejects an invalid base URL via ValidationError", () => {
    expect(() => createSession({ baseUrl: "not-a-url" })).toThrow(ValidationError);
  });

  it("normalises the base URL to include a trailing slash", () => {
    const session = createSession({ baseUrl: "https://canvus.example.com/api/v1" });
    expect(session.config.apiBaseUrl).toBe("https://canvus.example.com/api/v1/");
  });
});

describe("loadConfig", () => {
  it("parses CANVUS_* env vars", () => {
    const cfg = loadConfig({
      CANVUS_API_URL: BASE,
      CANVUS_API_KEY: KEY,
      CANVUS_TIMEOUT_MS: "5000",
      CANVUS_VERIFY_TLS: "false",
    });
    expect(cfg.apiBaseUrl).toBe(BASE);
    expect(cfg.apiKey).toBe(KEY);
    expect(cfg.timeoutMs).toBe(5000);
    expect(cfg.verifyTls).toBe(false);
  });

  it("throws ValidationError on a missing apiBaseUrl", () => {
    expect(() => loadConfig({})).toThrow(ValidationError);
  });
});

describe("Transport (via Session)", () => {
  let originalFetch: typeof fetch;

  beforeEach(() => {
    originalFetch = globalThis.fetch;
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("sets the private-token header on every request", async () => {
    const spy = mockFetchOnce({ body: [] });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.canvases.list();
    const [, init] = spy.mock.calls[0] as [URL, RequestInit];
    const headers = init.headers as Headers;
    expect(headers.get("private-token")).toBe(KEY);
    expect(headers.get("accept")).toBe("application/json");
  });

  it("builds URLs against the configured base", async () => {
    const spy = mockFetchOnce({ body: { "canvas-id": "abc", "canvas-name": "x" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.canvases.get("abc");
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.toString()).toBe("https://canvus.example.com/api/v1/canvases/abc");
  });

  it("encodes query parameters", async () => {
    const spy = mockFetchOnce({ body: { events: [], "total-count": 0, page: 0, "per-page": 100 } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.auditLog({ page: 2, "per-page": 50, action: "create_canvas" });
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.searchParams.get("page")).toBe("2");
    expect(url.searchParams.get("per-page")).toBe("50");
    expect(url.searchParams.get("action")).toBe("create_canvas");
  });

  it("maps HTTP 401 to AuthError", async () => {
    mockFetchOnce({ status: 401, body: { msg: "bad token" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.canvases.list()).rejects.toBeInstanceOf(AuthError);
  });

  it("maps HTTP 404 to NotFoundError", async () => {
    mockFetchOnce({ status: 404, body: { msg: "missing" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.canvases.get("does-not-exist")).rejects.toBeInstanceOf(NotFoundError);
  });

  it("maps generic HTTP 500 to APIError", async () => {
    mockFetchOnce({ status: 500, body: { msg: "boom" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.canvases.list()).rejects.toSatisfy(
      (e: unknown) => e instanceof APIError && (e as APIError).status === 500,
    );
  });

  it("returns undefined on 204 No Content", async () => {
    mockFetchOnce({ status: 204, body: "", headers: {} });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.canvases.delete("abc")).resolves.toBeUndefined();
  });
});

describe("WidgetsResource cross-cutting behaviour", () => {
  let originalFetch: typeof fetch;

  beforeEach(() => {
    originalFetch = globalThis.fetch;
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("clone() POSTs to the destination resource segment", async () => {
    const spy = mockFetchOnce({ status: 201, body: { "widget-id": "new" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.widgets.clone({
      destCanvasId: "dest",
      sourceCanvasId: "src",
      sourceWidgetId: "wid",
      widgetType: "note",
    });
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/canvases/dest/notes")).toBe(true);
    expect(init.method).toBe("POST");
    const body = JSON.parse(init.body as string) as Record<string, unknown>;
    expect(body.source_canvas_id).toBe("src");
    expect(body.source_widget_id).toBe("wid");
  });

  it("ipVideos has no create method (changelog §2)", () => {
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    expect("create" in session.widgets.ipVideos).toBe(false);
  });

  it("rdpConnections has no create method (changelog §2)", () => {
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    expect("create" in session.widgets.rdpConnections).toBe(false);
  });

  it("tables.update strips grid-size before sending (changelog §4)", async () => {
    const spy = mockFetchOnce({ body: { "table-id": "t1", "grid-size": { columns: 2, rows: 2 } } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.widgets.tables.update("c1", "t1", {
      title: "updated",
      // @ts-expect-error grid-size intentionally not in UpdateTableRequest
      "grid-size": { columns: 99, rows: 99 },
    });
    const [, init] = spy.mock.calls[0] as [URL, RequestInit];
    const sentBody = JSON.parse(init.body as string) as Record<string, unknown>;
    expect("grid-size" in sentBody).toBe(false);
    expect(sentBody.title).toBe("updated");
  });
});

describe("AssetsResource", () => {
  let originalFetch: typeof fetch;

  beforeEach(() => {
    originalFetch = globalThis.fetch;
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("injects the canvas-id header on asset downloads", async () => {
    const spy = mockFetchOnce({ body: "binary-blob", headers: { "content-type": "image/jpeg" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.assets.download("deadbeef", "canvas-uuid");
    const [, init] = spy.mock.calls[0] as [URL, RequestInit];
    const headers = init.headers as Headers;
    expect(headers.get("canvas-id")).toBe("canvas-uuid");
  });
});
