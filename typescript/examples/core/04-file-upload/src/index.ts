/**
 * Example 04 — file-upload.
 *
 * Uploads a tiny embedded PNG as an image widget, then patches it to a
 * specific position and scale. If `CANVUS_KEEP_WIDGET=1`, leaves the
 * widget on the canvas; otherwise deletes it as cleanup.
 *
 * The example is self-contained: a 16×16 solid-blue PNG is bundled as a
 * base64 constant so users don't have to find an image file to test
 * with. To upload your own image, set `CANVUS_IMAGE_PATH`.
 */

import { readFile } from "node:fs/promises";
import { basename } from "node:path";
import {
  createSession,
  isCanvusError,
} from "@mt-canvus-tools/sdk";
import pino from "pino";
import { z } from "zod";

const isPretty = process.env["LOG_FORMAT"] === "pretty";

const logger = pino({
  level: process.env["LOG_LEVEL"] ?? "info",
  base: { component: "example-file-upload" },
  ...(isPretty
    ? {
        transport: {
          target: "pino-pretty",
          options: { colorize: true, translateTime: "SYS:HH:MM:ss.l" },
        },
      }
    : {}),
});

const envSchema = z.object({
  CANVUS_BASE_URL: z.string().url(),
  CANVUS_API_KEY: z.string().min(1),
  CANVUS_CANVAS_ID: z.string().min(1),
  CANVUS_IMAGE_PATH: z.string().min(1).optional(),
  CANVUS_KEEP_WIDGET: z.enum(["0", "1"]).default("0"),
});

type Env = z.infer<typeof envSchema>;

/**
 * A minimal 16×16 solid-blue PNG, base64-encoded. 82 bytes on the wire.
 * Used when `CANVUS_IMAGE_PATH` is not set so the example is fully
 * self-contained and does not require the user to provide an image.
 */
const SAMPLE_PNG_BASE64 =
  "iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAGUlEQVR4nGPw7/v+nxLMMGrAqAGjBgwXAwC1hdMfddN3mAAAAABJRU5ErkJggg==";

function loadEnv(): Env {
  const parsed = envSchema.safeParse(process.env);
  if (!parsed.success) {
    logger.error(
      { fieldErrors: parsed.error.flatten().fieldErrors },
      "missing or invalid env",
    );
    process.exit(2);
  }
  return parsed.data;
}

/** Load the image bytes from disk if `CANVUS_IMAGE_PATH` is set, else
 * fall back to the bundled sample PNG. */
async function loadImage(env: Env): Promise<{ bytes: Buffer; filename: string }> {
  if (env.CANVUS_IMAGE_PATH) {
    const bytes = await readFile(env.CANVUS_IMAGE_PATH);
    return { bytes, filename: basename(env.CANVUS_IMAGE_PATH) };
  }
  const bytes = Buffer.from(SAMPLE_PNG_BASE64, "base64");
  return { bytes, filename: "sample.png" };
}

async function run(): Promise<void> {
  const env = loadEnv();
  const session = createSession({
    baseUrl: env.CANVUS_BASE_URL,
    apiKey: env.CANVUS_API_KEY,
  });
  const canvasId = env.CANVUS_CANVAS_ID;
  const { bytes, filename } = await loadImage(env);
  logger.info({ filename, byteCount: bytes.length }, "loaded image bytes");

  // Upload — the SDK builds the multipart body internally. The `meta`
  // argument becomes a `json` part with widget defaults; we leave it
  // empty here because we patch position/scale immediately after.
  const image = await session.widgets.images.upload(
    canvasId,
    bytes,
    filename,
  );
  const widgetId = image["widget-id"];
  logger.info(
    {
      widgetId,
      assetHash: image["asset-hash"],
      mimeType: image["mime-type"],
      fileSize: image["file-size"],
    },
    "uploaded image",
  );

  // Patch position + scale. Canvas coordinates are in pixels; do NOT
  // pre-multiply by canvas size.
  const patched = await session.widgets.images.update(canvasId, widgetId, {
    location: { x: 100, y: 200 },
    scale: 0.5,
  });
  logger.info(
    {
      widgetId: patched["widget-id"],
      location: patched.location,
      scale: patched.scale,
    },
    "patched image position and scale",
  );

  // Cleanup unless explicitly suppressed.
  if (env.CANVUS_KEEP_WIDGET === "1") {
    logger.info({ widgetId }, "CANVUS_KEEP_WIDGET=1 — leaving widget on canvas");
    return;
  }
  await session.widgets.images.delete(canvasId, widgetId);
  logger.info({ widgetId }, "deleted image (cleanup)");
}

run().catch((err: unknown) => {
  if (isCanvusError(err)) {
    logger.error({ kind: err.kind, err: err.message }, "file-upload failed");
  } else {
    logger.error({ err }, "file-upload failed (non-SDK error)");
  }
  process.exit(1);
});
