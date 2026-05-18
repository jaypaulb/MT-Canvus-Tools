/**
 * Example 08 — cross-canvas-clone.
 *
 * Clones a widget from one canvas into another using the SDK's
 * `session.widgets.clone({ destCanvasId, sourceCanvasId, sourceWidgetId,
 * widgetType })` helper (changelog §1). This is the canonical way to
 * copy widgets across canvases. The legacy
 * `POST /api/v1/canvases/{id}/widgets/clone` endpoint returns 501 and
 * is intentionally not used.
 *
 * Steps:
 *   1. Fetch the source widget to discover its type.
 *   2. Call `widgets.clone(...)` against the destination canvas.
 *   3. Print the new widget id.
 *   4. If `CANVUS_CLEANUP=1`, wait 2s and delete the cloned widget.
 */

import {
  createSession,
  isCanvusError,
  type CloneWidgetArgs,
} from "@mt-canvus-tools/sdk";
import pino from "pino";
import { setTimeout as sleep } from "node:timers/promises";
import { z } from "zod";

const isPretty = process.env["LOG_FORMAT"] === "pretty";

const logger = pino({
  level: process.env["LOG_LEVEL"] ?? "info",
  base: { component: "example-cross-canvas-clone" },
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
  CANVUS_SOURCE_WIDGET_ID: z.string().min(1),
  CANVUS_DEST_CANVAS_ID: z.string().min(1),
  CANVUS_CLEANUP: z.enum(["0", "1"]).default("0"),
});

type Env = z.infer<typeof envSchema>;

/** Widget kinds that the clone helper supports. Other kinds
 * (connector, video-input, ip-video, rdp-connection) cannot be cloned
 * via this endpoint per changelog §1 + §2. */
type CloneableWidgetType = CloneWidgetArgs["widgetType"];

const CLONEABLE: ReadonlySet<string> = new Set<CloneableWidgetType>([
  "note",
  "image",
  "video",
  "pdf",
  "browser",
  "anchor",
  "table",
]);

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

/** Narrow a string to a `CloneableWidgetType`, else return undefined. */
function asCloneableType(kind: string): CloneableWidgetType | undefined {
  return CLONEABLE.has(kind) ? (kind as CloneableWidgetType) : undefined;
}

/** Resolve the appropriate sub-resource `delete` helper for cleanup.
 * Returns undefined if the widget type has no SDK-exposed delete (which
 * should not happen for any cloneable type). */
function deleterFor(
  session: ReturnType<typeof createSession>,
  kind: CloneableWidgetType,
): (canvasId: string, widgetId: string) => Promise<void> {
  switch (kind) {
    case "note":
      return session.widgets.notes.delete;
    case "image":
      return session.widgets.images.delete;
    case "video":
      return session.widgets.videos.delete;
    case "pdf":
      return session.widgets.pdfs.delete;
    case "browser":
      return session.widgets.browsers.delete;
    case "anchor":
      return session.widgets.anchors.delete;
    case "table":
      return session.widgets.tables.delete;
  }
}

async function run(): Promise<void> {
  const env = loadEnv();
  const session = createSession({
    baseUrl: env.CANVUS_BASE_URL,
    apiKey: env.CANVUS_API_KEY,
  });

  // Step 1: fetch the source widget and determine its type.
  const source = await session.widgets.get(
    env.CANVUS_CANVAS_ID,
    env.CANVUS_SOURCE_WIDGET_ID,
  );
  const kindRaw = source["widget-type"];
  logger.info(
    {
      sourceCanvasId: env.CANVUS_CANVAS_ID,
      sourceWidgetId: env.CANVUS_SOURCE_WIDGET_ID,
      widgetType: kindRaw,
    },
    "fetched source widget",
  );

  const kind = asCloneableType(kindRaw);
  if (kind === undefined) {
    logger.error(
      { widgetType: kindRaw, cloneable: [...CLONEABLE] },
      "widget type is not cloneable via the clone endpoint",
    );
    process.exit(1);
  }

  // Step 2: clone.
  const cloned = await session.widgets.clone({
    destCanvasId: env.CANVUS_DEST_CANVAS_ID,
    sourceCanvasId: env.CANVUS_CANVAS_ID,
    sourceWidgetId: env.CANVUS_SOURCE_WIDGET_ID,
    widgetType: kind,
  });
  const clonedId = (cloned as { "widget-id"?: string })["widget-id"];
  logger.info(
    {
      destCanvasId: env.CANVUS_DEST_CANVAS_ID,
      sourceWidgetId: env.CANVUS_SOURCE_WIDGET_ID,
      clonedWidgetId: clonedId,
      widgetType: kind,
    },
    "cloned widget",
  );

  if (clonedId === undefined) {
    logger.error({ cloned }, "clone response missing widget-id — cannot proceed");
    process.exit(1);
  }

  // Step 4: optional cleanup.
  if (env.CANVUS_CLEANUP !== "1") {
    logger.info(
      { clonedWidgetId: clonedId },
      "CANVUS_CLEANUP=0 — leaving cloned widget on destination canvas",
    );
    return;
  }
  logger.info({ delayMs: 2000 }, "waiting before cleanup");
  await sleep(2000);
  const del = deleterFor(session, kind);
  await del(env.CANVUS_DEST_CANVAS_ID, clonedId);
  logger.info({ clonedWidgetId: clonedId }, "deleted cloned widget");
}

run().catch((err: unknown) => {
  if (isCanvusError(err)) {
    logger.error({ kind: err.kind, err: err.message }, "clone failed");
  } else {
    logger.error({ err }, "clone failed (non-SDK error)");
  }
  process.exit(1);
});
