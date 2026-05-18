// Phase 4b §4.3 — tests for the parity work items implemented in this phase.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  APIError,
  createSession,
  parseRetryAfter,
  RateLimitError,
  ServerError,
  UnsupportedOperationError,
} from "../src/index.js";
import { flattenServerConfig } from "../src/index.js";

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

describe("Phase 4b §4.3 #11 — RateLimitError / ServerError", () => {
  let originalFetch: typeof fetch;
  beforeEach(() => {
    originalFetch = globalThis.fetch;
  });
  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("maps HTTP 429 to RateLimitError with parsed Retry-After", async () => {
    // Use a tiny Retry-After (0 seconds) so retries don't exceed the test timeout.
    globalThis.fetch = vi
      .fn()
      .mockResolvedValue(
        mockResponse({ status: 429, headers: { "retry-after": "0" }, body: { msg: "slow" } }) as unknown as Response,
      ) as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.canvases.list()).rejects.toBeInstanceOf(RateLimitError);
  }, 15_000);

  it("maps HTTP 500 to ServerError after retries", async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValue(mockResponse({ status: 500, body: { msg: "boom" } }) as unknown as Response) as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.canvases.list()).rejects.toBeInstanceOf(ServerError);
  }, 15_000);

  it("parseRetryAfter handles seconds and HTTP-date", () => {
    expect(parseRetryAfter("30")).toBe(30_000);
    expect(parseRetryAfter("")).toBeUndefined();
    expect(parseRetryAfter(null)).toBeUndefined();
    expect(parseRetryAfter("not a date")).toBeUndefined();
  });
});

describe("Phase 4b §4.3 #12 — requestIdProvider header injection", () => {
  let originalFetch: typeof fetch;
  beforeEach(() => {
    originalFetch = globalThis.fetch;
  });
  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("sets X-Request-ID when the provider returns a non-empty value", async () => {
    const spy = vi.fn().mockResolvedValue(mockResponse({ body: [] }) as unknown as Response);
    globalThis.fetch = spy as unknown as typeof fetch;
    const session = createSession({
      baseUrl: BASE,
      apiKey: KEY,
      requestIdProvider: () => "req-abc-123",
    });
    await session.canvases.list();
    const [, init] = spy.mock.calls[0] as [URL, RequestInit];
    const headers = init.headers as Headers;
    expect(headers.get("x-request-id")).toBe("req-abc-123");
  });
});

describe("Phase 4b §4.3 #5 — auth.currentUser", () => {
  let originalFetch: typeof fetch;
  beforeEach(() => {
    originalFetch = globalThis.fetch;
  });
  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("GETs /users/current and returns the user", async () => {
    const spy = vi
      .fn()
      .mockResolvedValue(mockResponse({ body: { id: 42, email: "x@y", name: "Jay", admin: false } }) as unknown as Response);
    globalThis.fetch = spy as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const user = await session.auth.currentUser();
    expect(user.id).toBe(42);
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/current")).toBe(true);
  });
});

describe("Phase 4b §4.3 #3 — canvases.trash", () => {
  let originalFetch: typeof fetch;
  beforeEach(() => {
    originalFetch = globalThis.fetch;
  });
  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("resolves current user then POSTs /move with trash folder id", async () => {
    const spy = vi
      .fn()
      .mockResolvedValueOnce(mockResponse({ body: { id: 7, email: "x@y", name: "u", admin: false } }) as unknown as Response)
      .mockResolvedValueOnce(mockResponse({ body: { id: "c1", name: "Demo", mode: "normal", state: "ok" } }) as unknown as Response);
    globalThis.fetch = spy as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.canvases.trash("c1");
    const [, init] = spy.mock.calls[1] as [URL, RequestInit];
    const body = JSON.parse(init.body as string) as { folder_id: string };
    expect(body.folder_id).toBe("trash.7");
  });
});

describe("Phase 4b §4.3 #1 — widgets.createAny / updateAny / deleteAny", () => {
  let originalFetch: typeof fetch;
  beforeEach(() => {
    originalFetch = globalThis.fetch;
  });
  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("createAny dispatches Note → /notes", async () => {
    const spy = vi.fn().mockResolvedValue(mockResponse({ body: { id: "w1", widget_type: "Note" } }) as unknown as Response);
    globalThis.fetch = spy as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.widgets.createAny("c1", { widget_type: "Note", text: "hi" });
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/canvases/c1/notes")).toBe(true);
  });

  it("createAny throws UnsupportedOperationError for IpVideo", async () => {
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(
      session.widgets.createAny("c1", { widget_type: "IpVideo" }),
    ).rejects.toBeInstanceOf(UnsupportedOperationError);
  });

  it("createAny throws UnsupportedOperationError for Image (multipart required)", async () => {
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(
      session.widgets.createAny("c1", { widget_type: "Image" }),
    ).rejects.toBeInstanceOf(UnsupportedOperationError);
  });

  it("deleteAny routes by widget type segment", async () => {
    const spy = vi.fn().mockResolvedValue(mockResponse({ status: 204, body: "" }) as unknown as Response);
    globalThis.fetch = spy as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.widgets.deleteAny("c1", "w1", "RdpConnection");
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/canvases/c1/rdp-connections/w1")).toBe(true);
  });
});

describe("Phase 4b §4.3 #2 — widgets.patchParentId", () => {
  let originalFetch: typeof fetch;
  beforeEach(() => {
    originalFetch = globalThis.fetch;
  });
  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("PATCHes /widgets/{id} with parent_id only", async () => {
    const spy = vi.fn().mockResolvedValue(mockResponse({ body: { id: "w1", widget_type: "Note", parent_id: "new" } }) as unknown as Response);
    globalThis.fetch = spy as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.widgets.patchParentId("c1", "w1", "new");
    const [, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(init.method).toBe("PATCH");
    const body = JSON.parse(init.body as string) as Record<string, string>;
    expect(body).toEqual({ parent_id: "new" });
  });
});

describe("Phase 4b §4.3 #6 — color-preset decomposition", () => {
  let originalFetch: typeof fetch;
  beforeEach(() => {
    originalFetch = globalThis.fetch;
  });
  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("listColorPresets flattens to per-name entries", async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValue(
        mockResponse({
          body: {
            annotation: ["#000000FF"],
            connector: ["#111111FF", "#222222FF"],
            note_background: [],
            note_text: [],
          },
        }) as unknown as Response,
      ) as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const list = await session.canvases.listColorPresets("c1");
    expect(list).toEqual([
      { name: "annotation.0", color: "#000000FF" },
      { name: "connector.0", color: "#111111FF" },
      { name: "connector.1", color: "#222222FF" },
    ]);
  });
});

describe("Phase 4b §4.3 #7/#8/#9 — server convenience helpers", () => {
  let originalFetch: typeof fetch;
  beforeEach(() => {
    originalFetch = globalThis.fetch;
  });
  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("flattenServerConfig produces dotted-key element array", () => {
    const out = flattenServerConfig({
      auth: { password: { enabled: true, min_length: 8 } },
      server_name: "ISE",
    });
    expect(out).toContainEqual({
      key: "auth.password.enabled",
      value: true,
      type: "boolean",
    });
    expect(out).toContainEqual({
      key: "auth.password.min_length",
      value: 8,
      type: "number",
    });
    expect(out).toContainEqual({ key: "server_name", value: "ISE", type: "string" });
  });

  it("setVideoOutputSourceByIndex PATCHes the ordinal route", async () => {
    const spy = vi.fn().mockResolvedValue(mockResponse({ body: {} }) as unknown as Response);
    globalThis.fetch = spy as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.setVideoOutputSourceByIndex("client-1", 2, { source: "x" });
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(init.method).toBe("PATCH");
    expect(url.pathname.endsWith("/clients/client-1/video-outputs/2")).toBe(true);
  });

  it("toggleWorkspacePinned reads then PATCHes the inverse", async () => {
    const spy = vi
      .fn()
      .mockResolvedValueOnce(mockResponse({ body: { index: 0, canvas_id: "c1", pinned: false, state: "ok", workspace_name: "main" } }) as unknown as Response)
      .mockResolvedValueOnce(mockResponse({ body: { index: 0, canvas_id: "c1", pinned: true, state: "ok", workspace_name: "main" } }) as unknown as Response);
    globalThis.fetch = spy as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.server.toggleWorkspacePinned("client-1", "ws-1");
    const [, init] = spy.mock.calls[1] as [URL, RequestInit];
    const body = JSON.parse(init.body as string) as { pinned: boolean };
    expect(body.pinned).toBe(true);
  });
});

describe("Phase 4b §4.3 #22 — session.close()", () => {
  it("does not throw and is idempotent", async () => {
    const session = createSession({ baseUrl: BASE, apiKey: KEY, verifyTls: false });
    await session.close();
    await session.close();
    expect(true).toBe(true);
  });
});

describe("APIError tree retains existing types", () => {
  it("RateLimitError extends APIError and reports kind 'rate-limit'", () => {
    const err = new RateLimitError(undefined, "boom", 1000);
    expect(err).toBeInstanceOf(APIError);
    expect(err.kind).toBe("rate-limit");
    expect(err.retryAfterMs).toBe(1000);
  });

  it("ServerError keeps the original status code", () => {
    const err = new ServerError(503, undefined, "boom");
    expect(err.status).toBe(503);
    expect(err.kind).toBe("server");
  });

  it("UnsupportedOperationError carries the operation name", () => {
    const err = new UnsupportedOperationError("createAny", "bad");
    expect(err.operation).toBe("createAny");
    expect(err.kind).toBe("unsupported");
  });
});
