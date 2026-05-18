import type { AssetHash, Uuid } from "./common.js";

/**
 * Mipmap metadata returned by `GET /api/v1/mipmaps/{hash}`.
 *
 * Shape varies by server version. Known fields are underscored.
 */
export interface MipmapInfo {
  readonly hash?: AssetHash;
  readonly public_hash_hex?: string;
  readonly canvas_id?: Uuid;
  readonly levels?: readonly number[] | number;
  readonly max_level?: number;
  readonly pages?: number;
  readonly format?: string;
  readonly width?: number;
  readonly height?: number;
  readonly resolution?: { readonly width: number; readonly height: number };
  readonly tile_size?: number;
  readonly tiles_per_page?: number;
}

/**
 * Asset metadata returned alongside upload widget responses.
 *
 * Field names are underscored on the wire.
 */
export interface AssetMeta {
  readonly hash: AssetHash;
  readonly mime_type?: string;
  readonly file_size?: number;
  readonly original_filename?: string;
}
