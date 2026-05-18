/**
 * Example 07 — webhooks-notifications.
 *
 * Bridge Canvus widget events to an outbound HTTP webhook. The Canvus
 * API has no built-in webhook system; this example demonstrates the
 * pattern for building one on top of the streaming subscribe API.
 *
 * For each newly-created widget (the initial snapshot is filtered out
 * by an in-memory "seen" set), POSTs JSON to `WEBHOOK_URL` with three
 * retries on non-2xx responses (delays: 2s, 4s, 8s).
 *
 * Tip: point `WEBHOOK_URL` at https://webhook.site for easy testing.
 */

import {
  createSession,
  isCanvusError,
  type Widget,
} from "@mt-canvus-tools/sdk";
import pino from "pino";
import { setTimeout as sleep } from "node:timers/promises";
import { z } from "zod";

const isPretty = process.env["LOG_FORMAT"] === "pretty";

const logger = pino({
  level: process.env["LOG_LEVEL"] ?? "info",
  base: { component: "example-webhooks" },
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
  CANVUS_API_URL: z.string().url(),
  CANVUS_API_KEY: z.string().min(1),
  CANVUS_CANVAS_ID: z.string().min(1),
  WEBHOOK_URL: z.string().url(),
});

type Env = z.infer<typeof envSchema>;

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

interface OutboundEvent {
  readonly event: "widget.created";
  readonly canvas_id: string;
  readonly widget_id: string;
  readonly widget_type: string;
  readonly timestamp: string;
}

/** Retry delays (ms) between attempts. Total max wait: 14 seconds. */
const RETRY_DELAYS_MS = [2000, 4000, 8000];

/**
 * POST `body` to `url`, retrying on non-2xx up to `RETRY_DELAYS_MS.length`
 * additional attempts. Logs status + latency for every attempt.
 */
async function postWithRetry(url: string, body: OutboundEvent): Promise<void> {
  const maxAttempts = RETRY_DELAYS_MS.length + 1;
  for (let attempt = 1; attempt <= maxAttempts; attempt += 1) {
    const startedAt = Date.now();
    try {
      const res = await fetch(url, {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify(body),
      });
      const latencyMs = Date.now() - startedAt;
      if (res.ok) {
        logger.info(
          { attempt, status: res.status, latencyMs, widgetId: body.widget_id },
          "webhook delivered",
        );
        return;
      }
      logger.warn(
        { attempt, status: res.status, latencyMs, widgetId: body.widget_id },
        "webhook returned non-2xx",
      );
    } catch (err: unknown) {
      const latencyMs = Date.now() - startedAt;
      logger.warn(
        { attempt, latencyMs, widgetId: body.widget_id, err },
        "webhook POST threw",
      );
    }

    // Either non-2xx or thrown — back off and retry if attempts remain.
    if (attempt === maxAttempts) {
      logger.error(
        { attempts: maxAttempts, widgetId: body.widget_id },
        "webhook permanently failed",
      );
      return;
    }
    const delay = RETRY_DELAYS_MS[attempt - 1] ?? 0;
    logger.info({ delayMs: delay, nextAttempt: attempt + 1 }, "backing off");
    await sleep(delay);
  }
}

async function run(): Promise<void> {
  const env = loadEnv();
  const session = createSession({
    baseUrl: env.CANVUS_API_URL,
    apiKey: env.CANVUS_API_KEY,
  });

  const seen = new Set<string>();
  const controller = new AbortController();
  process.on("SIGINT", () => {
    logger.info({ signal: "SIGINT" }, "shutting down");
    controller.abort();
  });

  logger.info(
    {
      canvasId: env.CANVUS_CANVAS_ID,
      webhookUrl: env.WEBHOOK_URL,
    },
    "bridging widget.created events — Ctrl-C to exit",
  );

  const stream = session.widgets.subscribe(env.CANVUS_CANVAS_ID, {
    signal: controller.signal,
  });

  for await (const event of stream) {
    // The SDK's subscribe iterator types each event as a single Widget,
    // but the initial snapshot frame arrives as a Widget[] array.
    const raw: Widget | readonly Widget[] = event as Widget | readonly Widget[];
    const isSnapshot = Array.isArray(raw);
    const widgets: readonly Widget[] = isSnapshot ? raw : [raw as Widget];

    for (const widget of widgets) {
      // Every widget type uses `id` and `widget_type` (underscored) on
      // the live server. Connector is included in the union and shares
      // those fields per the verified wire shape.
      const id = widget.id;
      if (id === undefined) continue;
      if (seen.has(id)) continue;
      seen.add(id);

      // The very first frame is the initial snapshot of existing
      // widgets. To avoid spamming the webhook with everything that
      // already exists, only forward newly-seen widgets that arrive
      // AFTER startup. Snapshot widgets are marked `seen` above
      // without being forwarded; subsequent delta frames will pass.
      if (isSnapshot) continue;

      const outbound: OutboundEvent = {
        event: "widget.created",
        canvas_id: env.CANVUS_CANVAS_ID,
        widget_id: id,
        widget_type: widget.widget_type,
        timestamp: new Date().toISOString(),
      };
      // Fire-and-await; we want sequential delivery so a slow webhook
      // back-pressures the stream rather than queueing unbounded.
      await postWithRetry(env.WEBHOOK_URL, outbound);
    }
  }
}

run().catch((err: unknown) => {
  if (isCanvusError(err)) {
    logger.error({ kind: err.kind, err: err.message }, "webhooks failed");
  } else {
    logger.error({ err }, "webhooks failed (non-SDK error)");
  }
  process.exit(1);
});
