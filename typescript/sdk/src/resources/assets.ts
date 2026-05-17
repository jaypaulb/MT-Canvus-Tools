import type { Transport } from "../transport.js";
import type { AssetHash, Uuid } from "../types/common.js";
import type { MipmapInfo } from "../types/asset.js";

/** Options for mipmap calls. `page` defaults to 0. */
export interface MipmapOptions {
  readonly page?: number;
}

/**
 * Asset-by-hash endpoints.
 *
 * Every call requires a `canvas-id` header proving the caller has access
 * to *some* canvas that references the asset. The SDK adds this header
 * automatically from the `canvasId` argument.
 *
 * Note: the `canvasId` query parameter does NOT work as a substitute —
 * see endpoints/assets.md.
 */
export class AssetsResource {
  constructor(private readonly transport: Transport) {}

  /**
   * `GET /api/v1/assets/{hash}` — download an asset file directly by hash.
   *
   * @returns Binary `Blob` of the asset contents.
   */
  async download(hash: AssetHash, canvasId: Uuid): Promise<Blob> {
    const response = await this.transport.rawRequest("GET", `assets/${hash}`, undefined, {
      headers: { "canvas-id": canvasId },
      accept: "*/*",
    });
    return response.blob();
  }

  /** `GET /api/v1/mipmaps/{hash}` — mipmap metadata for a tileable asset. */
  async mipmap(hash: AssetHash, canvasId: Uuid, opts?: MipmapOptions): Promise<MipmapInfo> {
    return this.transport.request<MipmapInfo>("GET", `mipmaps/${hash}`, undefined, {
      headers: { "canvas-id": canvasId },
      query: opts?.page === undefined ? undefined : { page: opts.page },
    });
  }

  /**
   * `GET /api/v1/mipmaps/{hash}/{level}` — download a specific mipmap level.
   *
   * Level 0 is the highest resolution. Returns binary tile data as a `Blob`.
   */
  async mipmapLevel(
    hash: AssetHash,
    level: number,
    canvasId: Uuid,
    opts?: MipmapOptions,
  ): Promise<Blob> {
    const response = await this.transport.rawRequest(
      "GET",
      `mipmaps/${hash}/${level.toString()}`,
      undefined,
      {
        headers: { "canvas-id": canvasId },
        accept: "application/octet-stream",
        query: opts?.page === undefined ? undefined : { page: opts.page },
      },
    );
    return response.blob();
  }
}
