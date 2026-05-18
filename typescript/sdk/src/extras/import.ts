// Phase 4b §4.3 #18: import port (Node-only).
//
// Restore widgets exported by {@link WidgetExporter} into a target canvas.
// Reads the canonical `export.json` schema shared with Go and Python SDKs:
//
//   <folder>/
//     export.json       # {widgets, assets, region}
//     image_<id>.jpg    # flat sibling
//     pdf_<id>.pdf
//     video_<id>.mp4

import { Buffer } from "node:buffer";
import * as fs from "node:fs/promises";
import * as path from "node:path";
import type { Session } from "../session.js";
import type { Uuid } from "../types/common.js";
import type { ExportManifest } from "./export.js";

/** Configuration for {@link WidgetImporter}. */
export interface ImportConfig {
  readonly importAssets?: boolean;
  readonly restoreSpatialData?: boolean;
  readonly spatialOffset?: { readonly x: number; readonly y: number };
}

/** Summary returned by {@link WidgetImporter.importWidgetsFromFolder}. */
export interface ImportSummary {
  readonly imported: number;
  readonly skipped: number;
  readonly errors: readonly { widgetId: Uuid; reason: string }[];
  readonly idMapping: Readonly<Record<Uuid, Uuid>>;
}

/**
 * Import a folder previously produced by {@link WidgetExporter} (or by the
 * Go / Python SDKs — the wire shape is shared).
 *
 * Reads `export.json`, restores each widget via the matching create
 * helper, and re-uploads asset binaries when `importAssets !== false`.
 */
export class WidgetImporter {
  constructor(
    private readonly session: Session,
    private readonly config: ImportConfig = {},
  ) {}

  async importWidgetsFromFolder(
    folderPath: string,
    targetCanvasId: Uuid,
  ): Promise<ImportSummary> {
    const manifestRaw = await fs.readFile(path.join(folderPath, "export.json"), "utf8");
    const manifest = JSON.parse(manifestRaw) as ExportManifest;

    const idMapping: Record<Uuid, Uuid> = {};
    const errors: { widgetId: Uuid; reason: string }[] = [];
    let imported = 0;
    let skipped = 0;

    for (const widgetRaw of manifest.widgets) {
      const widgetId = typeof widgetRaw.id === "string" ? widgetRaw.id : "<unknown>";
      try {
        const newId = await this.importOne(
          widgetRaw,
          manifest.assets,
          folderPath,
          targetCanvasId,
        );
        if (newId === undefined) {
          skipped++;
        } else {
          imported++;
          idMapping[widgetId] = newId;
        }
      } catch (err) {
        errors.push({ widgetId, reason: (err as Error).message });
      }
    }
    return { imported, skipped, errors, idMapping };
  }

  private async importOne(
    widgetRaw: Record<string, unknown>,
    assetsMap: Readonly<Record<Uuid, string>>,
    folder: string,
    targetCanvas: Uuid,
  ): Promise<Uuid | undefined> {
    const widgetId = typeof widgetRaw.id === "string" ? widgetRaw.id : undefined;
    const widgetType =
      typeof widgetRaw.widget_type === "string" ? widgetRaw.widget_type : undefined;
    if (widgetId === undefined || widgetType === undefined) return undefined;
    const wType = widgetType.toLowerCase();

    const assetFilename =
      assetsMap[widgetId] !== undefined && this.config.importAssets !== false
        ? assetsMap[widgetId]
        : undefined;

    if (assetFilename !== undefined) {
      const binPath = path.join(folder, assetFilename);
      const buf = await fs.readFile(binPath);
      const blob = new Blob([Buffer.from(buf)]);
      switch (wType) {
        case "image": {
          const created = await this.session.widgets.images.upload(
            targetCanvas,
            blob,
            assetFilename,
          );
          return created.id;
        }
        case "video": {
          const created = await this.session.widgets.videos.upload(
            targetCanvas,
            blob,
            assetFilename,
          );
          return created.id;
        }
        case "pdf": {
          const created = await this.session.widgets.pdfs.upload(
            targetCanvas,
            blob,
            assetFilename,
          );
          return created.id;
        }
      }
    }

    // Asset-required types without a binary cannot be recreated.
    if (wType === "image" || wType === "video" || wType === "pdf") return undefined;
    // Server rejects IPVideo/RDP creation.
    if (wType === "ipvideo" || wType === "rdpconnection") return undefined;

    const payload = this.preparePayload(widgetRaw);
    const created = await this.session.widgets.createAny(targetCanvas, payload);
    return created.id;
  }

  private preparePayload(widgetRaw: Record<string, unknown>): Record<string, unknown> {
    const out: Record<string, unknown> = { ...widgetRaw };
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
}
