// Phase 4b §4.3 #18: export port (Node-only).
//
// Widget + asset export to an on-disk folder. Matches the canonical Go +
// Python on-disk schema for cross-runtime portability:
//
//   <export folder>/
//     export.json          # {widgets, assets, region}
//     image_<id>.jpg       # one per Image widget
//     pdf_<id>.pdf         # one per PDF widget
//     video_<id>.mp4       # one per Video widget
//
// The `assets` map in `export.json` is `{widget_id: filename}`. Non-asset
// widgets (notes, browsers, connectors, anchors, tables, …) are present
// in the `widgets` array but absent from `assets`.

import { Buffer } from "node:buffer";
import * as fs from "node:fs/promises";
import * as path from "node:path";
import type { Session } from "../session.js";
import type { Uuid } from "../types/common.js";
import type { Widget } from "../types/widget.js";

/** Configuration for {@link WidgetExporter}. */
export interface ExportConfig {
  readonly includeAssets?: boolean;
  /** Base directory for the export. Defaults to `process.cwd()`. */
  readonly exportPath?: string;
  /** Allow overwriting an existing export folder. */
  readonly overwriteExisting?: boolean;
}

/** Rectangle region the export was scoped to, written to `export.json`. */
export interface ExportRegion {
  readonly x: number;
  readonly y: number;
  readonly width: number;
  readonly height: number;
}

/**
 * Top-level wire shape of `export.json`.
 *
 * Identical to Go's `ExportedWidgetSet` (`go/sdk/canvus/export.go:85-89`)
 * and Python's `ExportedWidgetSet.to_dict()`
 * (`python/sdk/src/canvus_sdk/extras/export.py:50-60`).
 */
export interface ExportManifest {
  readonly widgets: readonly Record<string, unknown>[];
  readonly assets: Readonly<Record<Uuid, string>>;
  readonly region: ExportRegion | null;
}

/**
 * Export widgets + (optionally) asset binaries from a canvas region to a
 * local folder. Returns the absolute path to the created folder.
 */
export class WidgetExporter {
  constructor(
    private readonly session: Session,
    private readonly config: ExportConfig = {},
  ) {}

  /**
   * @param canvasId — source canvas
   * @param widgetIds — when provided, only these widget IDs are exported;
   *                    otherwise all widgets on the canvas are included.
   * @param region — optional spatial bounds recorded in `export.json`.
   * @param folderPath — output folder; defaults to
   *   `<config.exportPath || cwd>/canvus_export_<canvasId>_<YYYYMMDD_HHMMSS>`.
   */
  async exportWidgetsToFolder(
    canvasId: Uuid,
    widgetIds?: readonly Uuid[],
    region?: ExportRegion,
    folderPath?: string,
  ): Promise<string> {
    const basePath = folderPath ?? this.config.exportPath ?? process.cwd();
    const exportFolder =
      folderPath ??
      path.join(basePath, `canvus_export_${canvasId}_${formatTimestamp(new Date())}`);
    await this.createFolder(exportFolder);

    let widgets: readonly Widget[];
    try {
      widgets = await this.session.widgets.list(canvasId);
    } catch (err) {
      await fs.rm(exportFolder, { recursive: true, force: true });
      throw err;
    }

    if (widgetIds && widgetIds.length > 0) {
      const set = new Set(widgetIds);
      widgets = widgets.filter((w) => set.has(w.id));
    }

    const widgetsOut: Record<string, unknown>[] = [];
    const assets: Record<Uuid, string> = {};

    for (const w of widgets) {
      widgetsOut.push(w as unknown as Record<string, unknown>);
      if (this.config.includeAssets === false) continue;
      const exportedFilename = await this.maybeWriteAsset(w, canvasId, exportFolder);
      if (exportedFilename !== undefined) {
        assets[w.id] = exportedFilename;
      }
    }

    const manifest: ExportManifest = {
      widgets: widgetsOut,
      assets,
      region: region ?? null,
    };
    await fs.writeFile(
      path.join(exportFolder, "export.json"),
      JSON.stringify(manifest, null, 2),
      "utf8",
    );
    return exportFolder;
  }

  private async createFolder(folder: string): Promise<void> {
    try {
      const stat = await fs.stat(folder);
      if (stat.isDirectory() && this.config.overwriteExisting !== true) {
        throw new Error(`export folder already exists: ${folder}`);
      }
    } catch (err) {
      if ((err as NodeJS.ErrnoException).code !== "ENOENT") throw err;
    }
    await fs.mkdir(folder, { recursive: true });
  }

  /**
   * Downloads an Image/Video/PDF widget's binary and writes it as a flat
   * sibling of `export.json`. Returns the filename written, or undefined
   * if this widget is not an asset type or download fails.
   */
  private async maybeWriteAsset(
    widget: Widget,
    canvasId: Uuid,
    folder: string,
  ): Promise<string | undefined> {
    const wType = widget.widget_type.toLowerCase();
    if (wType !== "image" && wType !== "video" && wType !== "pdf") return undefined;

    const filename = assetFilename(wType, widget.id);
    let blob: Blob;
    try {
      blob = await this.downloadAsset(widget, canvasId);
    } catch {
      // Asset download failures are non-fatal — the widget still ships
      // in `widgets[]` but without an `assets` entry. Matches Go behaviour
      // (Go bails out of the entire export on download failure; we follow
      // Python's softer behaviour for cross-runtime stability).
      return undefined;
    }
    const buf = Buffer.from(await blob.arrayBuffer());
    await fs.writeFile(path.join(folder, filename), buf);
    return filename;
  }

  private async downloadAsset(widget: Widget, canvasId: Uuid): Promise<Blob> {
    if (widget.widget_type === "Image") {
      return this.session.widgets.images.download(canvasId, widget.id);
    }
    if (widget.widget_type === "Video") {
      return this.session.widgets.videos.download(canvasId, widget.id);
    }
    if (widget.widget_type === "Pdf") {
      return this.session.widgets.pdfs.download(canvasId, widget.id);
    }
    throw new Error(`downloadAsset: unsupported widget type ${widget.widget_type}`);
  }
}

/**
 * Canonical asset filename per widget type. Matches Go's `export.go:57/67/77`
 * and Python's `extras/export.py:166/169/172`.
 */
export function assetFilename(widgetTypeLower: string, widgetId: Uuid): string {
  switch (widgetTypeLower) {
    case "image":
      return `image_${widgetId}.jpg`;
    case "pdf":
      return `pdf_${widgetId}.pdf`;
    case "video":
      return `video_${widgetId}.mp4`;
    default:
      throw new Error(`assetFilename: not an asset type: ${widgetTypeLower}`);
  }
}

function formatTimestamp(d: Date): string {
  const pad = (n: number): string => n.toString().padStart(2, "0");
  return `${d.getUTCFullYear().toString()}${pad(d.getUTCMonth() + 1)}${pad(d.getUTCDate())}_${pad(d.getUTCHours())}${pad(d.getUTCMinutes())}${pad(d.getUTCSeconds())}`;
}
