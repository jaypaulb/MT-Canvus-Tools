// Phase 4d Round C: focused coverage for UsersResource. Exercises every
// method's happy path through a mocked fetch, plus a representative slice
// of the error-mapping branches (401, 404, 429, 5xx).

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

const USER = { id: 1, email: "u@y", name: "U", admin: false };
const GROUP = { id: 1, name: "g" };

describe("UsersResource — users CRUD", () => {
  let originalFetch: typeof fetch;
  beforeEach(() => { originalFetch = globalThis.fetch; });
  afterEach(() => { globalThis.fetch = originalFetch; vi.restoreAllMocks(); });

  it("list() GETs /users", async () => {
    const spy = setFetchOnce({ body: [USER] });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const out = await session.users.list();
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users")).toBe(true);
    expect(init.method).toBe("GET");
    expect(out.length).toBe(1);
  });

  it("get() GETs /users/{uid}", async () => {
    const spy = setFetchOnce({ body: USER });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.users.get(42);
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/42")).toBe(true);
  });

  it("create() POSTs /users", async () => {
    const spy = setFetchOnce({ body: USER });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.users.create({ email: "n@y", name: "N", password: "pw" });
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users")).toBe(true);
    expect(init.method).toBe("POST");
  });

  it("update() PATCHes /users/{uid}", async () => {
    const spy = setFetchOnce({ body: USER });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.users.update(42, { name: "Renamed" });
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/42")).toBe(true);
    expect(init.method).toBe("PATCH");
  });

  it("changeEmail() POSTs /users/{uid}/change-email", async () => {
    const spy = setFetchOnce({ body: { msg: "ok" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.users.changeEmail(42, { "new-email": "x@y" });
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/42/change-email")).toBe(true);
  });

  it("block() POSTs /users/{uid}/block", async () => {
    const spy = setFetchOnce({ body: USER });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.users.block(42);
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/42/block")).toBe(true);
  });

  it("unblock() POSTs /users/{uid}/unblock", async () => {
    const spy = setFetchOnce({ body: USER });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.users.unblock(42);
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/42/unblock")).toBe(true);
  });

  it("approve() POSTs /users/{uid}/approve", async () => {
    const spy = setFetchOnce({ body: USER });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.users.approve(42);
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/42/approve")).toBe(true);
  });

  it("forcePasswordReset() POSTs /users/{uid}/reset-password", async () => {
    const spy = setFetchOnce({ body: { msg: "ok" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.users.forcePasswordReset(42);
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/42/reset-password")).toBe(true);
  });

  it("delete() DELETEs /users/{uid}", async () => {
    const spy = setFetchOnce({ status: 204, body: "", headers: {} });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.users.delete(42)).resolves.toBeUndefined();
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/users/42")).toBe(true);
    expect(init.method).toBe("DELETE");
  });
});

describe("UsersResource — groups CRUD", () => {
  let originalFetch: typeof fetch;
  beforeEach(() => { originalFetch = globalThis.fetch; });
  afterEach(() => { globalThis.fetch = originalFetch; vi.restoreAllMocks(); });

  it("listGroups() GETs /groups", async () => {
    const spy = setFetchOnce({ body: [GROUP] });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const out = await session.users.listGroups();
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/groups")).toBe(true);
    expect(out.length).toBe(1);
  });

  it("getGroup() GETs /groups/{gid}", async () => {
    const spy = setFetchOnce({ body: GROUP });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.users.getGroup(9);
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/groups/9")).toBe(true);
  });

  it("createGroup() POSTs /groups", async () => {
    const spy = setFetchOnce({ body: GROUP });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.users.createGroup({ name: "g" });
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/groups")).toBe(true);
    expect(init.method).toBe("POST");
  });

  it("updateGroup() PATCHes /groups/{gid}", async () => {
    const spy = setFetchOnce({ body: GROUP });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.users.updateGroup(9, { name: "g2" });
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/groups/9")).toBe(true);
    expect(init.method).toBe("PATCH");
  });

  it("deleteGroup() DELETEs /groups/{gid}", async () => {
    const spy = setFetchOnce({ status: 204, body: "", headers: {} });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.users.deleteGroup(9)).resolves.toBeUndefined();
    const [, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(init.method).toBe("DELETE");
  });

  it("listGroupMembers() GETs /groups/{gid}/members", async () => {
    const spy = setFetchOnce({ body: [USER] });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.users.listGroupMembers(9);
    const [url] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/groups/9/members")).toBe(true);
  });

  it("addGroupMember() POSTs /groups/{gid}/members", async () => {
    const spy = setFetchOnce({ status: 204, body: "", headers: {} });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.users.addGroupMember(9, { id: 42 });
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/groups/9/members")).toBe(true);
    expect(init.method).toBe("POST");
  });

  it("removeGroupMember() DELETEs /groups/{gid}/members/{uid}", async () => {
    const spy = setFetchOnce({ status: 204, body: "", headers: {} });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await session.users.removeGroupMember(9, 42);
    const [url, init] = spy.mock.calls[0] as [URL, RequestInit];
    expect(url.pathname.endsWith("/groups/9/members/42")).toBe(true);
    expect(init.method).toBe("DELETE");
  });
});

describe("UsersResource — subscribe entrypoints", () => {
  it("subscribe*() methods return async iterables", () => {
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    expect(typeof session.users.subscribe()[Symbol.asyncIterator]).toBe("function");
    expect(typeof session.users.subscribeOne(1)[Symbol.asyncIterator]).toBe("function");
    expect(typeof session.users.subscribeGroups()[Symbol.asyncIterator]).toBe("function");
    expect(typeof session.users.subscribeGroup(1)[Symbol.asyncIterator]).toBe("function");
    expect(typeof session.users.subscribeGroupMembers(1)[Symbol.asyncIterator]).toBe("function");
  });
});

describe("UsersResource — error mapping", () => {
  let originalFetch: typeof fetch;
  beforeEach(() => { originalFetch = globalThis.fetch; });
  afterEach(() => { globalThis.fetch = originalFetch; vi.restoreAllMocks(); });

  it("maps 401 to AuthError on list()", async () => {
    setFetchOnce({ status: 401, body: { msg: "no" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.users.list()).rejects.toBeInstanceOf(AuthError);
  });

  it("maps 404 to NotFoundError on get()", async () => {
    setFetchOnce({ status: 404, body: { msg: "missing" } });
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.users.get(9999)).rejects.toBeInstanceOf(NotFoundError);
  });

  it("maps 429 to RateLimitError on create()", { timeout: 15_000 }, async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValue(
        mockResponse({ status: 429, headers: { "retry-after": "0" }, body: { msg: "slow" } }) as unknown as Response,
      ) as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(
      session.users.create({ email: "x@y", name: "N", password: "pw" }),
    ).rejects.toBeInstanceOf(RateLimitError);
  });

  it("maps 500 to ServerError on listGroups()", { timeout: 15_000 }, async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValue(mockResponse({ status: 500, body: { msg: "boom" } }) as unknown as Response) as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    await expect(session.users.listGroups()).rejects.toBeInstanceOf(ServerError);
  });
});
