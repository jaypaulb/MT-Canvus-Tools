// Phase 4d Round C: focused coverage for AuthResource. Exercises every
// method's happy path through a mocked fetch, plus the error-mapping branches
// (401→AuthError, 404→NotFoundError, 429→RateLimitError, 5xx→ServerError).

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  AuthError,
  NotFoundError,
  RateLimitError,
  ServerError,
  createSession,
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

describe("AuthResource — happy paths", () => {
  let originalFetch: typeof fetch;

  beforeEach(() => {
    originalFetch = globalThis.fetch;
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("login() POSTs users/login and returns the response body", async () => {
    const spy = setFetchOnce({
      body: { token: "abc", user: { id: 1, email: "x@y", name: "x", admin: false } },
    });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const out = await session.auth.login({ email: "x@y", password: "pw" });
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/login")).toBe(true);
    expect(init.method).toBe("POST");
    expect(out.token).toBe("abc");
  });

  it("loginSaml() POSTs users/login/saml", async () => {
    const spy = setFetchOnce({
      body: { token: "abc", user: { id: 1, email: "x@y", name: "x", admin: false } },
    });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.auth.loginSaml({ inResponseTo: "x", responseXml: "<xml/>" });
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/login/saml")).toBe(true);
  });

  it("logout() POSTs users/logout", async () => {
    const spy = setFetchOnce({ body: { msg: "ok" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const out = await session.auth.logout();
    const [, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(init.method).toBe("POST");
    expect(out.msg).toBe("ok");
  });

  it("currentUser() GETs users/current", async () => {
    const spy = setFetchOnce({ body: { id: 7, email: "j@y", name: "Jay", admin: true } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const u = await session.auth.currentUser();
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/current")).toBe(true);
    expect(init.method).toBe("GET");
    expect(u.id).toBe(7);
  });

  it("createResetToken() POSTs users/password/create-reset-token", async () => {
    const spy = setFetchOnce({ body: { msg: "sent" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.auth.createResetToken({ email: "x@y" });
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/password/create-reset-token")).toBe(true);
  });

  it("validateResetToken() GETs users/password/validate-reset-token with token query", async () => {
    const spy = setFetchOnce({ body: { valid: true } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const out = await session.auth.validateResetToken("tok-123");
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/password/validate-reset-token")).toBe(true);
    expect(url.searchParams.get("token")).toBe("tok-123");
    expect(out.valid).toBe(true);
  });

  it("resetPassword() POSTs users/password/reset", async () => {
    const spy = setFetchOnce({ body: { msg: "reset" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.auth.resetPassword({ token: "t", password: "new" });
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/password/reset")).toBe(true);
  });

  it("register() POSTs users/register", async () => {
    const spy = setFetchOnce({
      body: { user: { id: 9, email: "n@y", name: "n", admin: false }, msg: "ok" },
    });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.auth.register({ email: "n@y", password: "pw", name: "n" });
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/register")).toBe(true);
  });

  it("confirmEmail() POSTs users/confirm-email", async () => {
    const spy = setFetchOnce({ body: { msg: "ok" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.auth.confirmEmail({ token: "t" });
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/confirm-email")).toBe(true);
  });

  it("changePassword() POSTs users/{uid}/password", async () => {
    const spy = setFetchOnce({ body: { id: 7, email: "x@y", name: "Jay", admin: false } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.auth.changePassword(7, { "old-password": "old", "new-password": "new" });
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/7/password")).toBe(true);
    expect(init.method).toBe("POST");
  });

  it("listAccessTokens() GETs users/{uid}/access-tokens", async () => {
    const spy = setFetchOnce({ body: [] });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const out = await session.auth.listAccessTokens(7);
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/7/access-tokens")).toBe(true);
    expect(init.method).toBe("GET");
    expect(out).toEqual([]);
  });

  it("getAccessToken() GETs users/{uid}/access-tokens/{tid}", async () => {
    const spy = setFetchOnce({ body: { id: "tok", description: "d", created_at: "2026-01-01T00:00:00Z" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.auth.getAccessToken(7, "tok");
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/7/access-tokens/tok")).toBe(true);
  });

  it("createAccessToken() POSTs users/{uid}/access-tokens", async () => {
    const spy = setFetchOnce({
      body: { id: "tok", description: "d", created_at: "2026-01-01T00:00:00Z", plain_token: "secret" },
    });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const out = await session.auth.createAccessToken(7, { description: "d" });
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/7/access-tokens")).toBe(true);
    expect(init.method).toBe("POST");
    expect(out.plain_token).toBe("secret");
  });

  it("updateAccessToken() PATCHes users/{uid}/access-tokens/{tid}", async () => {
    const spy = setFetchOnce({ body: { id: "tok", description: "new", created_at: "2026-01-01T00:00:00Z" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.auth.updateAccessToken(7, "tok", { description: "new" });
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/7/access-tokens/tok")).toBe(true);
    expect(init.method).toBe("PATCH");
  });

  it("deleteAccessToken() DELETEs users/{uid}/access-tokens/{tid}", async () => {
    const spy = setFetchOnce({ status: 204, body: "", headers: {} });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.auth.deleteAccessToken(7, "tok")).resolves.toBeUndefined();
    const [, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(init.method).toBe("DELETE");
  });

  it("subscribeAccessTokens() and subscribeAccessToken() return async iterables", () => {
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const iterA = session.auth.subscribeAccessTokens(7);
    const iterB = session.auth.subscribeAccessToken(7, "tok");
    // We only check the call-site is reached and a generator is returned;
    // actual streaming behaviour is exercised by streaming-specific tests.
    expect(typeof iterA[Symbol.asyncIterator]).toBe("function");
    expect(typeof iterB[Symbol.asyncIterator]).toBe("function");
  });
});

describe("AuthResource — error mapping", () => {
  let originalFetch: typeof fetch;

  beforeEach(() => {
    originalFetch = globalThis.fetch;
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("maps 401 to AuthError on login()", async () => {
    setFetchOnce({ status: 401, body: { msg: "bad creds" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.auth.login({ email: "x@y", password: "pw" })).rejects.toBeInstanceOf(
      AuthError,
    );
  });

  it("maps 404 to NotFoundError on getAccessToken()", async () => {
    setFetchOnce({ status: 404, body: { msg: "missing" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.auth.getAccessToken(7, "nope")).rejects.toBeInstanceOf(NotFoundError);
  });

  it("maps 429 to RateLimitError on listAccessTokens()", { timeout: 15_000 }, async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValue(
        mockResponse({
          status: 429,
          headers: { "retry-after": "0", "content-type": "application/json" },
          body: { msg: "slow" },
        }) as unknown as Response,
      ) as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.auth.listAccessTokens(7)).rejects.toBeInstanceOf(RateLimitError);
  });

  it("maps 500 to ServerError on currentUser()", { timeout: 15_000 }, async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValue(mockResponse({ status: 500, body: { msg: "boom" } }) as unknown as Response) as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.auth.currentUser()).rejects.toBeInstanceOf(ServerError);
  });
});
