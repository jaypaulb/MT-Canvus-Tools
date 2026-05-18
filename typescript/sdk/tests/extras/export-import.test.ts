import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Buffer } from "node:buffer";
import * as fs from "node:fs/promises";
import * as os from "node:os";
import * as path from "node:path";
import { createSession } from "../../src/index.js";
import { WidgetExporter, assetFilename } from "../../src/extras/export.js";
import { WidgetImporter } from "../../src/extras/import.js";

const BASE = "https://canvus.example.com/api/v1/";
const KEY = "cv_test_token_abcdef0123456789";

function jsonResponse(body: unknown): Response {
  return {
    ok: true,
    status: 200,
    headers: new Headers({ "content-type": "application/json" }),
    json: async () => body,
    text: async () => JSON.stringify(body),
    blob: async () => new Blob([JSON.stringify(body)]),
  } as unknown as Response;
}

function blobResponse(bytes: Uint8Array, contentType: string): Response {
  return {
    ok: true,
    status: 200,
    headers: new Headers({ "content-type": contentType }),
    blob: async () => new Blob([bytes], { type: contentType }),
    json: async () => {
      throw new Error("blobResponse: not JSON");
    },
    text: async () => "",
  } as unknown as Response;
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

  it("writes export.json with canonical {widgets, assets, region} schema", async () => {
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
      .mockResolvedValueOnce(jsonResponse(widgets)) as unknown as typeof fetch;

    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const exporter = new WidgetExporter(session, {
      exportPath: tmpDir,
      includeAssets: false,
    });
    const out = await exporter.exportWidgetsToFolder("c1");

    const manifestText = await fs.readFile(path.join(out, "export.json"), "utf8");
    const manifest = JSON.parse(manifestText) as {
      widgets: { id: string }[];
      assets: Record<string, string>;
      region: unknown;
    };
    expect(manifest.widgets).toHaveLength(1);
    expect(manifest.widgets[0]?.id).toBe("w1");
    expect(manifest.assets).toEqual({});
    expect(manifest.region).toBeNull();
  });

  it("writes image asset as flat sibling image_<id>.jpg + records in assets map", async () => {
    const widgets = [
      {
        id: "img1",
        widget_type: "Image",
        parent_id: "c1",
        location: { x: 0, y: 0 },
        size: { width: 100, height: 100 },
        depth: 0,
        scale: 1,
        pinned: false,
        state: "ok",
      },
    ];
    const imageBytes = new Uint8Array([0xff, 0xd8, 0xff, 0xe0]); // JPEG magic

    globalThis.fetch = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse(widgets))
      .mockResolvedValueOnce(blobResponse(imageBytes, "image/jpeg")) as unknown as typeof fetch;

    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const exporter = new WidgetExporter(session, {
      exportPath: tmpDir,
      includeAssets: true,
    });
    const out = await exporter.exportWidgetsToFolder("c1", undefined, {
      x: 0,
      y: 0,
      width: 200,
      height: 200,
    });

    const manifest = JSON.parse(
      await fs.readFile(path.join(out, "export.json"), "utf8"),
    ) as { assets: Record<string, string>; region: { width: number } | null };

    expect(manifest.assets["img1"]).toBe("image_img1.jpg");
    expect(manifest.region?.width).toBe(200);
    const written = await fs.readFile(path.join(out, "image_img1.jpg"));
    expect(written.equals(Buffer.from(imageBytes))).toBe(true);
  });
});

describe("WidgetImporter", () => {
  let originalFetch: typeof fetch;
  let tmpDir: string;

  beforeEach(async () => {
    originalFetch = globalThis.fetch;
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), "canvus-import-test-"));
  });

  afterEach(async () => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  it("reads canonical export.json schema and recreates a Note via createAny", async () => {
    const manifest = {
      widgets: [
        {
          id: "w1",
          widget_type: "Note",
          parent_id: "old-canvas",
          location: { x: 10, y: 20 },
          size: { width: 100, height: 100 },
          depth: 0,
          scale: 1,
          state: "ok",
          text: "ported",
          background_color: "FFFFFFFF",
          text_color: "000000FF",
          auto_text_color: false,
        },
      ],
      assets: {},
      region: null,
    };
    await fs.writeFile(
      path.join(tmpDir, "export.json"),
      JSON.stringify(manifest),
      "utf8",
    );

    const createdNote = { ...manifest.widgets[0], id: "w1-new", parent_id: "target" };
    globalThis.fetch = vi.fn().mockResolvedValueOnce(jsonResponse(createdNote)) as unknown as typeof fetch;

    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const importer = new WidgetImporter(session);
    const summary = await importer.importWidgetsFromFolder(tmpDir, "target");

    expect(summary.imported).toBe(1);
    expect(summary.skipped).toBe(0);
    expect(summary.errors).toEqual([]);
    expect(summary.idMapping["w1"]).toBe("w1-new");
  });

  it("skips Image widgets when their binary is absent from the assets map", async () => {
    const manifest = {
      widgets: [
        {
          id: "img1",
          widget_type: "Image",
          parent_id: "old",
          location: { x: 0, y: 0 },
          size: { width: 100, height: 100 },
          depth: 0,
          scale: 1,
          state: "ok",
        },
      ],
      assets: {}, // No binary for the image.
      region: null,
    };
    await fs.writeFile(
      path.join(tmpDir, "export.json"),
      JSON.stringify(manifest),
      "utf8",
    );

    globalThis.fetch = vi.fn() as unknown as typeof fetch;

    const session = createSession({ baseUrl: BASE, apiKey: KEY });
    const importer = new WidgetImporter(session);
    const summary = await importer.importWidgetsFromFolder(tmpDir, "target");

    expect(summary.imported).toBe(0);
    expect(summary.skipped).toBe(1);
    expect(summary.errors).toEqual([]);
  });
});

describe("assetFilename", () => {
  it("produces canonical Go/Python-compatible filenames", () => {
    expect(assetFilename("image", "abc")).toBe("image_abc.jpg");
    expect(assetFilename("pdf", "abc")).toBe("pdf_abc.pdf");
    expect(assetFilename("video", "abc")).toBe("video_abc.mp4");
  });

  it("throws for non-asset widget types", () => {
    expect(() => assetFilename("note", "abc")).toThrow();
  });
});
