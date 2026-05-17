import type { AssetHash } from "./common.js";

/** Mipmap metadata returned by `GET /api/v1/mipmaps/{hash}`. */
export interface MipmapInfo {
  readonly hash: AssetHash;
  readonly "tile-size": number;
  readonly levels: number;
  readonly width: number;
  readonly height: number;
  readonly pages: number;
  readonly "tiles-per-page": number;
}

/** Asset metadata returned alongside upload widget responses. */
export interface AssetMeta {
  readonly "asset-hash": AssetHash;
  readonly "asset-hash-private"?: AssetHash;
  readonly "mime-type": string;
  readonly "file-size": number;
  readonly "original-filename": string;
}
