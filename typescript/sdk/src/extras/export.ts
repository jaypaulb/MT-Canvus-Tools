// Phase 4b §4.3 #18: export port (Node-only).
//
// Widget + asset export to an on-disk folder. Mirrors the structure of
// CanvusPythonAPI/canvus_api/export.py and aligns with Go's `export.go`
// manifest schema (`manifest.json` at the export root, `widgets/<id>.json`,
// `assets/<id>.<ext>`).

import { Buffer } from "node:buffer";
import * as fs from "node:fs/promises";
import * as path from "node:path";
import type { Session } from "../session.js";
import type { Canvas } from "../types/canvas.js";
import type { Uuid } from "../types/common.js";
import type { Widget } from "../types/widget.js";

/** Configuration for {@link WidgetExporter}. */
export interface ExportConfig {
  readonly includeAssets?: boolean;
  readonly includeSpatialData?: boolean;
  readonly includeMetadata?: boolean;
  /** Base directory for the export. Defaults to `process.cwd()`. */
  readonly exportPath?: string;
  /** Allow overwriting an existing export folder. */
  readonly overwriteExisting?: boolean;
}

/** Per-widget entry in the export manifest. */
export interface ExportedWidget {
  readonly id: Uuid;
  readonly widget_type: string;
  readonly canvas_id: Uuid;
  readonly exported_at: string;
  readonly data: Record<string, unknown>;
  readonly assets?: readonly ExportedAsset[];
}

/** Per-asset entry. */
export interface ExportedAsset {
  readonly type: "image" | "video" | "pdf";
  readonly local_path: string;
  readonly filename: string;
}

/** Top-level manifest written to `manifest.json`. */
export interface ExportManifest {
  readonly version: "1.0";
  readonly exported_at: string;
  readonly config: ExportConfig;
  readonly canvases: Record<Uuid, { name: string }>;
  readonly widgets: readonly ExportedWidget[];
}

/**
 * Export widgets + (optionally) asset binaries from one or more canvases
 * to a local folder. Returns the absolute path to the created folder.
 *
 * Folder layout:
 * ```
 * canvus_export_<canvasId>_<YYYYMMDD_HHMMSS>/
 *   manifest.json
 *   widgets/<widgetId>.json
 *   assets/<filename>
 * ```
 */
export class WidgetExporter {
  constructor(
    private readonly session: Session,
    private readonly config: ExportConfig = {},
  ) {}

  async exportWidgetsToFolder(
    canvasId: Uuid,
    widgetIds?: readonly Uuid[],
    folderPath?: string,
  ): Promise<string> {
    const basePath = folderPath ?? this.config.exportPath ?? process.cwd();
    const stamp = formatTimestamp(new Date());
    const exportFolder = path.join(basePath, `canvus_export_${canvasId}_${stamp}`);
    await this.createStructure(exportFolder);

    let canvas: Canvas;
    try {
      canvas = await this.session.canvases.get(canvasId);
    } catch (err) {
      await fs.rm(exportFolder, { recursive: true, force: true });
      throw err;
    }

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

    const exported: ExportedWidget[] = [];
    for (const w of widgets) {
      const entry = await this.exportWidget(w, canvasId, exportFolder);
      exported.push(entry);
    }

    const manifest: ExportManifest = {
      version: "1.0",
      exported_at: new Date().toISOString(),
      config: this.config,
      canvases: { [canvasId]: { name: canvas.name } },
      widgets: exported,
    };
    await fs.writeFile(
      path.join(exportFolder, "manifest.json"),
      JSON.stringify(manifest, null, 2),
      "utf8",
    );
    return exportFolder;
  }

  private async createStructure(folder: string): Promise<void> {
    try {
      const stat = await fs.stat(folder);
      if (stat.isDirectory() && this.config.overwriteExisting !== true) {
        throw new Error(`export folder already exists: ${folder}`);
      }
    } catch (err) {
      if ((err as NodeJS.ErrnoException).code !== "ENOENT") throw err;
    }
    await fs.mkdir(path.join(folder, "widgets"), { recursive: true });
    await fs.mkdir(path.join(folder, "assets"), { recursive: true });
  }

  private async exportWidget(
    widget: Widget,
    canvasId: Uuid,
    folder: string,
  ): Promise<ExportedWidget> {
    const entry: ExportedWidget = {
      id: widget.id,
      widget_type: widget.widget_type,
      canvas_id: canvasId,
      exported_at: new Date().toISOString(),
      data: widget as unknown as Record<string, unknown>,
      ...(this.config.includeAssets !== false &&
        (await this.maybeExportAssets(widget, folder))),
    };
    await fs.writeFile(
      path.join(folder, "widgets", `${widget.id}.json`),
      JSON.stringify(entry, null, 2),
      "utf8",
    );
    return entry;
  }

  private async maybeExportAssets(
    widget: Widget,
    folder: string,
  ): Promise<{ assets?: readonly ExportedAsset[] }> {
    const wType = widget.widget_type.toLowerCase();
    if (wType !== "image" && wType !== "video" && wType !== "pdf") return {};
    const ext = wType === "pdf" ? "pdf" : wType === "video" ? "bin" : "bin";
    const filename = `${wType}_${widget.id}.${ext}`;
    const local = path.join(folder, "assets", filename);
    try {
      const blob = await this.downloadAsset(widget);
      const buf = Buffer.from(await blob.arrayBuffer());
      await fs.writeFile(local, buf);
      return {
        assets: [
          {
            type: wType,
            local_path: local,
            filename,
          },
        ],
      };
    } catch {
      return {};
    }
  }

  private async downloadAsset(widget: Widget): Promise<Blob> {
    if (widget.widget_type === "Image") {
      return this.session.widgets.images.download(widget.parent_id, widget.id);
    }
    if (widget.widget_type === "Video") {
      return this.session.widgets.videos.download(widget.parent_id, widget.id);
    }
    if (widget.widget_type === "Pdf") {
      return this.session.widgets.pdfs.download(widget.parent_id, widget.id);
    }
    throw new Error(`downloadAsset: unsupported widget type ${widget.widget_type}`);
  }
}

function formatTimestamp(d: Date): string {
  const pad = (n: number): string => n.toString().padStart(2, "0");
  return `${d.getUTCFullYear().toString()}${pad(d.getUTCMonth() + 1)}${pad(d.getUTCDate())}_${pad(d.getUTCHours())}${pad(d.getUTCMinutes())}${pad(d.getUTCSeconds())}`;
}
