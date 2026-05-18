import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import * as fs from "node:fs/promises";
import * as os from "node:os";
import * as path from "node:path";
import { createSession } from "../../src/index.js";
import { WidgetExporter } from "../../src/extras/export.js";

const BASE = "https://canvus.example.com/api/v1/";
const KEY = "cv_test_token_abcdef0123456789";

function mockResponse(body: unknown): {
  ok: boolean;
  status: number;
  headers: Headers;
  json: () => Promise<unknown>;
  text: () => Promise<string>;
  blob: () => Promise<Blob>;
} {
  return {
    ok: true,
    status: 200,
    headers: new Headers({ "content-type": "application/json" }),
    json: async () => body,
    text: async () => JSON.stringify(body),
    blob: async () => new Blob([JSON.stringify(body)]),
  };
}

describe("WidgetExporter", () => {
  let originalFetch: typeof fetch;
  let tmpDir: string;

  beforeEach(async () => {
    originalFetch = globalThis.fetch;
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), "canvus-export-test-"));
  });

  afterEach(async () => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  it("writes manifest + widget files to disk", async () => {
    const canvas = { id: "c1", name: "Demo", mode: "normal", state: "ok" };
    const widgets = [
      {
        id: "w1",
        widget_type: "Note",
        parent_id: "c1",
        location: { x: 1, y: 2 },
        size: { width: 3, height: 4 },
        depth: 0,
        scale: 1,
        pinned: false,
        state: "ok",
        text: "hi",
        background_color: "FFFFFFFF",
        text_color: "000000FF",
        auto_text_color: false,
      },
    ];

    globalThis.fetch = vi
      .fn()
      .mockResolvedValueOnce(mockResponse(canvas) as unknown as Response)
      .mockResolvedValueOnce(mockResponse(widgets) as unknown as Response) as unknown as typeof fetch;

    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const exporter = new WidgetExporter(session, { exportPath: tmpDir, includeAssets: false });
    const out = await exporter.exportWidgetsToFolder("c1");

    const manifestText = await fs.readFile(path.join(out, "manifest.json"), "utf8");
    const manifest = JSON.parse(manifestText) as { widgets: { id: string }[]; canvases: Record<string, { name: string }> };
    expect(manifest.widgets).toHaveLength(1);
    expect(manifest.widgets[0]?.id).toBe("w1");
    expect(manifest.canvases.c1?.name).toBe("Demo");
  });
});
