/**
 * Users route tests. The Users module touches the userStore but not the
 * Canvus session, so we can feed an empty session and focus on the
 * color-assignment semantics.
 */

import { describe, expect, it } from "vitest";
import { buildApp } from "../src/index.js";
import { buildFakeDeps } from "./helpers.js";

describe("users routes", () => {
  it("/identify-user assigns a new color per (team, name) pair", async () => {
    const deps = buildFakeDeps({ session: {} });
    const app = buildApp(deps);

    const res = await app.request("/identify-user", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ team: 3, name: "Alice" }),
    });
    expect(res.status).toBe(200);
    const body = (await res.json()) as { success: boolean; color: string };
    expect(body.success).toBe(true);
    expect(body.color).toMatch(/^#[0-9A-Fa-f]{8}$/);
  });

  it("/identify-user returns the same color on repeat call", async () => {
    const deps = buildFakeDeps({ session: {}, users: { "3": { Alice: "#11223344" } } });
    const app = buildApp(deps);

    const res = await app.request("/identify-user", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ team: 3, name: "Alice" }),
    });
    const body = (await res.json()) as { success: boolean; color: string };
    expect(body.color).toBe("#11223344");
  });

  it("/identify-user rejects invalid team via zod", async () => {
    const deps = buildFakeDeps({ session: {} });
    const app = buildApp(deps);

    const res = await app.request("/identify-user", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ team: 9, name: "Bob" }),
    });
    expect(res.status).toBe(400);
  });

  it("/list-users flattens the team→name map", async () => {
    const deps = buildFakeDeps({
      session: {},
      users: { "1": { Alice: "#aaaaaaff" }, "2": { Bob: "#bbbbbbff" } },
    });
    const app = buildApp(deps);
    const res = await app.request("/list-users");
    const body = (await res.json()) as {
      success: boolean;
      users: Array<{ team: number; name: string; color: string }>;
    };
    expect(body.users).toHaveLength(2);
    expect(body.users.find((u) => u.name === "Alice")?.team).toBe(1);
  });

  it("/delete-users with all:true wipes the store", async () => {
    const deps = buildFakeDeps({ session: {}, users: { "1": { Alice: "#aaaaaaff" } } });
    const app = buildApp(deps);
    const res = await app.request("/delete-users", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ all: true }),
    });
    expect(res.status).toBe(200);
    const after = await deps.variables.userStore.read();
    expect(Object.keys(after)).toHaveLength(0);
  });

  it("/delete-users removes specific (team, name) tuples", async () => {
    const deps = buildFakeDeps({
      session: {},
      users: { "1": { Alice: "#aaaaaaff", Bob: "#bbbbbbff" } },
    });
    const app = buildApp(deps);
    const res = await app.request("/delete-users", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ users: [{ team: 1, name: "Alice" }] }),
    });
    expect(res.status).toBe(200);
    const after = await deps.variables.userStore.read();
    expect(after["1"]).toEqual({ Bob: "#bbbbbbff" });
  });
});
