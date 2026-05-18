import { describe, expect, it } from "vitest";
import { createSession } from "../../src/index.js";

/**
 * Live-server integration tests.
 *
 * These run only when `CANVUS_API_URL` and `CANVUS_API_KEY` are
 * exported in the environment; otherwise the suite is skipped. Do NOT
 * hard-code credentials here.
 *
 * Run with:
 *   CANVUS_API_URL=... CANVUS_API_KEY=... pnpm test:integration
 */
const BASE = process.env.CANVUS_API_URL;
const KEY = process.env.CANVUS_API_KEY;
const HAVE_CREDS = typeof BASE === "string" && BASE !== "" && typeof KEY === "string" && KEY !== "";

const describeIf = HAVE_CREDS ? describe : describe.skip;

describeIf("Canvus live server", () => {
  const session = createSession({
    baseUrl: BASE!,
    apiKey: KEY!,
  });

  it("returns server info without authentication", async () => {
    const info = await session.server.info();
    expect(typeof info.version).toBe("string");
  });

  it("lists canvases", async () => {
    const canvases = await session.canvases.list();
    expect(Array.isArray(canvases)).toBe(true);
  });

  it("creates and deletes a canvas (round-trip)", async () => {
    const created = await session.canvases.create({
      "canvas-name": `sdk-integration-${Date.now().toString()}`,
    });
    try {
      expect(created["canvas-id"]).toBeTruthy();
      const fetched = await session.canvases.get(created["canvas-id"]);
      expect(fetched["canvas-id"]).toBe(created["canvas-id"]);
    } finally {
      await session.canvases.delete(created["canvas-id"]);
    }
  });
});
