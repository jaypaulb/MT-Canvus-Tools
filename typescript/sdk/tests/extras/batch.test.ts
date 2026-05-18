import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createSession } from "../../src/index.js";
import {
  BatchOperationBuilder,
  BatchProcessor,
  summarize,
} from "../../src/extras/batch.js";

const BASE = "https://canvus.example.com/api/v1/";
const KEY = "cv_test_token_abcdef0123456789";

describe("BatchProcessor", () => {
  let originalFetch: typeof fetch;

  beforeEach(() => {
    originalFetch = globalThis.fetch;
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("runs all queued operations", async () => {
    const headers = new Headers({ "content-type": "application/json" });
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      headers,
      json: async () => ({}),
      text: async () => "{}",
    } as unknown as Response) as unknown as typeof fetch;
    const session = createSession({ baseUrl: BASE, apiKey: KEY });

    const canvas = {
      id: "c1",
      name: "x",
      mode: "normal",
      state: "ok",
    } as never;
    const ops = new BatchOperationBuilder()
      .move("m1", canvas, "f1")
      .copy("c1", canvas, "f1")
      .build();
    const proc = new BatchProcessor(session, { maxConcurrency: 1 });
    const results = await proc.execute(ops);

    expect(results).toHaveLength(2);
    expect(results.every((r) => r.success)).toBe(true);
  });

  it("summarize aggregates per-op stats", () => {
    const start = new Date();
    const end = new Date(start.getTime() + 50);
    const summary = summarize([
      { operationId: "a", success: true, startTime: start, endTime: end, durationMs: 50, retries: 0 },
      { operationId: "b", success: false, error: "x", startTime: start, endTime: end, durationMs: 100, retries: 0 },
    ]);
    expect(summary.successful).toBe(1);
    expect(summary.failed).toBe(1);
    expect(summary.totalDurationMs).toBe(150);
    expect(summary.averageDurationMs).toBe(75);
    expect(summary.failedOperations).toHaveLength(1);
  });
});
