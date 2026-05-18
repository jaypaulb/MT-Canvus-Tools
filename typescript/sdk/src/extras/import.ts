// Phase 4b §4.3 #18: import port (Node-only).
//
// Restore widgets exported by {@link WidgetExporter} into a target canvas.
// Asset re-upload happens via the per-type upload helpers.

import { Buffer } from "node:buffer";
import * as fs from "node:fs/promises";
import * as path from "node:path";
import type { Session } from "../session.js";
import type { Uuid } from "../types/common.js";
import type { ExportManifest, ExportedAsset, ExportedWidget } from "./export.js";

/** Configuration for {@link WidgetImporter}. */
export interface ImportConfig {
  readonly importAssets?: boolean;
  readonly restoreSpatialData?: boolean;
  readonly restoreMetadata?: boolean;
  /** Target canvas. If undefined, uses the canvas ID embedded in the export. */
  readonly targetCanvasId?: Uuid;
  readonly spatialOffset?: { readonly x: number; readonly y: number };
  /** Currently informational only — server always assigns new IDs. */
  readonly preserveIds?: boolean;
}

/** Summary returned by {@link WidgetImporter.importWidgetsFromFolder}. */
export interface ImportSummary {
  readonly imported: number;
  readonly skipped: number;
  readonly errors: readonly { widgetId: Uuid; reason: string }[];
  readonly idMapping: Readonly<Record<Uuid, Uuid>>;
}

/**
 * Import a folder previously produced by {@link WidgetExporter}.
 *
 * Reads `manifest.json`, restores each widget via the matching create
 * helper, and re-uploads asset binaries when `importAssets !== false`.
 */
export class WidgetImporter {
  constructor(
    private readonly session: Session,
    private readonly config: ImportConfig = {},
  ) {}

  async importWidgetsFromFolder(
    folderPath: string,
    targetCanvasId?: Uuid,
  ): Promise<ImportSummary> {
    const manifestRaw = await fs.readFile(path.join(folderPath, "manifest.json"), "utf8");
    const manifest = JSON.parse(manifestRaw) as ExportManifest;
    const targetCanvas =
      targetCanvasId ?? this.config.targetCanvasId ?? Object.keys(manifest.canvases)[0];
    if (targetCanvas === undefined) {
      throw new Error("import: no target canvas resolved (manifest has no canvases?)");
    }

    const idMapping: Record<Uuid, Uuid> = {};
    const errors: { widgetId: Uuid; reason: string }[] = [];
    let imported = 0;
    let skipped = 0;

    for (const entry of manifest.widgets) {
      try {
        const newId = await this.importOne(entry, folderPath, targetCanvas);
        if (newId === undefined) {
          skipped++;
        } else {
          imported++;
          idMapping[entry.id] = newId;
        }
      } catch (err) {
        errors.push({ widgetId: entry.id, reason: (err as Error).message });
      }
    }
    return { imported, skipped, errors, idMapping };
  }

  private async importOne(
    entry: ExportedWidget,
    folder: string,
    targetCanvas: Uuid,
  ): Promise<Uuid | undefined> {
    const payload = this.preparePayload(entry);
    const wType = entry.widget_type.toLowerCase();

    // Asset-bearing types: re-upload the binary if available.
    if (this.config.importAssets !== false && entry.assets && entry.assets.length > 0) {
      const asset = entry.assets[0];
      if (asset === undefined) return undefined;
      const filename = asset.filename;
      const binPath = await this.resolveAssetPath(folder, asset);
      const buf = await fs.readFile(binPath);
      const blob = new Blob([Buffer.from(buf)]);
      if (wType === "image") {
        const created = await this.session.widgets.images.upload(targetCanvas, blob, filename);
        return created.id;
      }
      if (wType === "video") {
        const created = await this.session.widgets.videos.upload(targetCanvas, blob, filename);
        return created.id;
      }
      if (wType === "pdf") {
        const created = await this.session.widgets.pdfs.upload(targetCanvas, blob, filename);
        return created.id;
      }
    }

    // Non-asset types use generic create.
    if (
      wType === "image" ||
      wType === "video" ||
      wType === "pdf" ||
      wType === "ipvideo" ||
      wType === "rdpconnection"
    ) {
      // Cannot recreate without binary / forbidden.
      return undefined;
    }
    const created = await this.session.widgets.createAny(targetCanvas, payload);
    return created.id;
  }

  private preparePayload(entry: ExportedWidget): Record<string, unknown> {
    const out: Record<string, unknown> = { ...entry.data, widget_type: entry.widget_type };
    if (this.config.restoreSpatialData === false) {
      delete out.location;
      delete out.size;
    } else if (this.config.spatialOffset !== undefined) {
      const loc = out.location as { x?: number; y?: number } | undefined;
      if (loc !== undefined && typeof loc.x === "number" && typeof loc.y === "number") {
        out.location = {
          x: loc.x + this.config.spatialOffset.x,
          y: loc.y + this.config.spatialOffset.y,
        };
      }
    }
    // Strip server-controlled fields the create endpoint will reject.
    delete out.id;
    delete out.state;
    delete out.depth;
    return out;
  }

  private async resolveAssetPath(folder: string, asset: ExportedAsset): Promise<string> {
    // Manifest stores absolute path; fall back to <folder>/assets/<filename>.
    try {
      await fs.access(asset.local_path);
      return asset.local_path;
    } catch {
      return path.join(folder, "assets", asset.filename);
    }
  }
}
