import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createSession } from "../../src/index.js";
import { CrossCanvasSearch } from "../../src/extras/search.js";

const BASE = "https://canvus.example.com/api/v1/";
const KEY = "cv_test_token_abcdef0123456789";

function mockResponse(body: unknown): { ok: boolean; status: number; headers: Headers; json: () => Promise<unknown>; text: () => Promise<string> } {
  return {
    ok: true,
    status: 200,
    headers: new Headers({ "content-type": "application/json" }),
    json: async () => body,
    text: async () => JSON.stringify(body),
  };
}

describe("CrossCanvasSearch", () => {
  let originalFetch: typeof fetch;

  beforeEach(() => {
    originalFetch = globalThis.fetch;
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("finds widgets by text across canvases", async () => {
    const canvases = [{ id: "c1", name: "First", mode: "normal", state: "ok" }];
    const widgets = [
      { id: "w1", widget_type: "Note", parent_id: "c1", location: { x: 0, y: 0 }, size: { width: 1, height: 1 }, depth: 0, scale: 1, pinned: false, state: "ok", text: "hello world", background_color: "FFFFFFFF", text_color: "000000FF", auto_text_color: false },
    ];
    globalThis.fetch = vi
      .fn()
      .mockResolvedValueOnce(mockResponse(canvases) as unknown as Response)
      .mockResolvedValueOnce(mockResponse(widgets) as unknown as Response) as unknown as typeof fetch;

    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const search = new CrossCanvasSearch(session);
    const results = await search.findWidgetsByText("hello");
    expect(results).toHaveLength(1);
    expect(results[0]?.widgetId).toBe("w1");
    expect(results[0]?.canvasName).toBe("First");
    expect(results[0]?.drillDownPath).toBe("c1:w1");
  });
});
