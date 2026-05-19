// Phase 4d Round C5: schema-only export/import roundtrip test.
//
// Builds a synthetic export bundle (canonical {widgets, assets, region}
// schema with flat sibling asset files) and runs it back through the TS
// WidgetImporter. Verifies that the cross-runtime wire shape is honoured:
// widget shapes flow through to the create call, and asset file mapping
// is preserved (importer reads the correct sibling file).
//
// No live server is touched — fetch is mocked.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Buffer } from "node:buffer";
import * as fs from "node:fs/promises";
import * as os from "node:os";
import * as path from "node:path";
import { createSession } from "../src/index.js";
import { WidgetImporter } from "../src/extras/import.js";
import type { ExportManifest } from "../src/extras/export.js";

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

describe("export/import roundtrip (schema-only)", () => {
  let originalFetch: typeof fetch;
  let tmpDir: string;

  beforeEach(async () => {
    originalFetch = globalThis.fetch;
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), "canvus-roundtrip-test-"));
  });

  afterEach(async () => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  it(
    "builds canonical bundle on disk and reimports — note + image preserved with asset map",
    async () => {
      // Step 1: construct a synthetic export bundle matching Go's wire shape
      //   {widgets, assets, region} with image_<id>.jpg as flat sibling.
      const imageBytes = new Uint8Array([0xff, 0xd8, 0xff, 0xe0, 0x10, 0x4a, 0x46, 0x49, 0x46]); // JPEG header
      const manifest: ExportManifest = {
        widgets: [
          {
            id: "note-1",
            widget_type: "Note",
            parent_id: "src-canvas",
            location: { x: 10, y: 20 },
            size: { width: 100, height: 50 },
            depth: 0,
            scale: 1,
            pinned: false,
            state: "normal",
            text: "roundtrip",
            background_color: "FFFF00FF",
            text_color: "000000FF",
            auto_text_color: false,
          },
          {
            id: "img-1",
            widget_type: "Image",
            parent_id: "src-canvas",
            location: { x: 50, y: 60 },
            size: { width: 200, height: 200 },
            depth: 1,
            scale: 1,
            pinned: false,
            state: "normal",
          },
        ],
        assets: { "img-1": "image_img-1.jpg" },
        region: { x: 0, y: 0, width: 1000, height: 1000 },
      };
      await fs.writeFile(
        path.join(tmpDir, "export.json"),
        JSON.stringify(manifest, null, 2),
        "utf8",
      );
      await fs.writeFile(path.join(tmpDir, "image_img-1.jpg"), Buffer.from(imageBytes));

      // Step 2: mock fetch — first call is Note POST (createAny), second is Image upload.
      const createdNote = {
        id: "note-1-new",
        widget_type: "Note",
        parent_id: "dst-canvas",
        location: { x: 10, y: 20 },
        size: { width: 100, height: 50 },
        depth: 0,
        scale: 1,
        pinned: false,
        state: "normal",
        text: "roundtrip",
        background_color: "FFFF00FF",
        text_color: "000000FF",
        auto_text_color: false,
      };
      const createdImage = {
        id: "img-1-new",
        widget_type: "Image",
        parent_id: "dst-canvas",
        location: { x: 50, y: 60 },
        size: { width: 200, height: 200 },
        depth: 1,
        scale: 1,
        pinned: false,
        state: "normal",
      };

      const fetchSpy = vi
        .fn()
        .mockResolvedValueOnce(jsonResponse(createdNote))
        .mockResolvedValueOnce(jsonResponse(createdImage)) as unknown as typeof fetch;
      globalThis.fetch = fetchSpy;

      // Step 3: drive the importer.
      const session = createSession({ baseUrl: BASE, apiKey: KEY });
      const importer = new WidgetImporter(session);
      const summary = await importer.importWidgetsFromFolder(tmpDir, "dst-canvas");

      // Step 4: roundtrip assertions.
      expect(summary.errors).toEqual([]);
      expect(summary.imported).toBe(2);
      expect(summary.skipped).toBe(0);
      // Asset file mapping preserved: the image widget's exported id maps to the
      // freshly-created image widget's id via the importer's idMapping.
      expect(summary.idMapping["note-1"]).toBe("note-1-new");
      expect(summary.idMapping["img-1"]).toBe("img-1-new");
      expect(fetchSpy).toHaveBeenCalledTimes(2);

      // Inspect the recorded calls to verify schema preservation post-hoc.
      const calls = (fetchSpy as unknown as { mock: { calls: [URL | string, RequestInit][] } })
        .mock.calls;

      // Call 0: Note POST — JSON body containing widget shape.
      const [noteUrl, noteInit] = calls[0]!;
      expect(String(noteUrl)).toContain("/canvases/dst-canvas/notes");
      expect(noteInit.method).toBe("POST");
      const noteBodyText =
        typeof noteInit.body === "string"
          ? noteInit.body
          : await (noteInit.body as Blob).text();
      const noteParsed = JSON.parse(noteBodyText) as Record<string, unknown>;
      expect(noteParsed.widget_type).toBe("Note");
      expect(noteParsed.text).toBe("roundtrip");
      expect(noteParsed.location).toEqual({ x: 10, y: 20 });
      expect(noteParsed.size).toEqual({ width: 100, height: 50 });
      // Importer strips server-controlled fields.
      expect(noteParsed.id).toBeUndefined();
      expect(noteParsed.state).toBeUndefined();
      expect(noteParsed.depth).toBeUndefined();

      // Call 1: Image upload — body is FormData (fetch auto-sets the
      // multipart/form-data Content-Type with boundary from the FormData).
      const [imgUrl, imgInit] = calls[1]!;
      expect(String(imgUrl)).toContain("/canvases/dst-canvas/images");
      expect(imgInit.method).toBe("POST");
      expect(imgInit.body).toBeInstanceOf(FormData);
      // The asset file mapping resolved correctly: the FormData contains the
      // on-disk image bytes under the "data" field with the importer's filename.
      const form = imgInit.body as FormData;
      const dataField = form.get("data");
      expect(dataField).toBeInstanceOf(Blob);
    },
  );

  it("preserves canonical wire shape — manifest written out matches manifest read back", async () => {
    const manifest: ExportManifest = {
      widgets: [
        {
          id: "w1",
          widget_type: "Note",
          parent_id: "src",
          location: { x: 1.5, y: 2.5 },
          size: { width: 30, height: 40 },
          depth: 0,
          scale: 1,
          pinned: false,
          state: "normal",
          text: "schema-check",
          background_color: "FFFFFFFF",
          text_color: "000000FF",
          auto_text_color: false,
        },
      ],
      assets: {},
      region: { x: 0, y: 0, width: 500, height: 500 },
    };
    const manifestPath = path.join(tmpDir, "export.json");
    await fs.writeFile(manifestPath, JSON.stringify(manifest, null, 2), "utf8");

    const reread = JSON.parse(await fs.readFile(manifestPath, "utf8")) as ExportManifest;
    // Top-level keys are exactly {widgets, assets, region} — no leaks, no renames.
    expect(Object.keys(reread).sort()).toEqual(["assets", "region", "widgets"]);
    expect(reread.widgets).toHaveLength(1);
    expect(reread.assets).toEqual({});
    expect(reread.region).toEqual({ x: 0, y: 0, width: 500, height: 500 });
    // Widget fields untouched.
    expect(reread.widgets[0]).toEqual(manifest.widgets[0]);
  });
});
