/**
 * Example 05 — streaming.
 *
 * Subscribes to the notes endpoint on a canvas and prints each event as
 * it arrives. Runs for `STREAM_DURATION_SECONDS` (default 30) or until
 * SIGINT, whichever comes first. The async iterator returned by the
 * SDK's `subscribe()` cleans up automatically when the loop breaks.
 *
 * Per the SDK's streaming implementation, empty newlines (15-second
 * keepalive pings) are skipped silently — they never reach this loop.
 */

import {
  createSession,
  isCanvusError,
  type Note,
} from "@mt-canvus-tools/sdk";
import pino from "pino";
import { z } from "zod";

const isPretty = process.env["LOG_FORMAT"] === "pretty";

const logger = pino({
  level: process.env["LOG_LEVEL"] ?? "info",
  base: { component: "example-streaming" },
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
  STREAM_DURATION_SECONDS: z.coerce.number().int().positive().default(30),
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

/** Summary statistics printed once the stream ends. */
interface Stats {
  eventCount: number;
  firstAt: number | undefined;
  lastAt: number | undefined;
}

async function run(): Promise<void> {
  const env = loadEnv();
  const session = createSession({
    baseUrl: env.CANVUS_BASE_URL,
    apiKey: env.CANVUS_API_KEY,
  });

  const controller = new AbortController();
  const durationMs = env.STREAM_DURATION_SECONDS * 1000;

  // Set up cleanup paths: duration timer + SIGINT.
  const timer = setTimeout(() => {
    logger.info(
      { durationSeconds: env.STREAM_DURATION_SECONDS },
      "duration elapsed — closing stream",
    );
    controller.abort();
  }, durationMs);

  const onSigint = (): void => {
    logger.info({ signal: "SIGINT" }, "received SIGINT — closing stream");
    controller.abort();
  };
  process.on("SIGINT", onSigint);

  const stats: Stats = { eventCount: 0, firstAt: undefined, lastAt: undefined };

  try {
    logger.info(
      { canvasId: env.CANVUS_CANVAS_ID, durationSeconds: env.STREAM_DURATION_SECONDS },
      "subscribing to notes stream",
    );

    const stream = session.widgets.notes.subscribe(env.CANVUS_CANVAS_ID, {
      signal: controller.signal,
    });

    for await (const event of stream) {
      const now = Date.now();
      stats.eventCount += 1;
      if (stats.firstAt === undefined) stats.firstAt = now;
      stats.lastAt = now;

      // Widen at the boundary: the SDK types each event as Note, but the
      // initial snapshot frame arrives as a Note[] array.
      summarise(event as Note | readonly Note[]);
    }
  } finally {
    clearTimeout(timer);
    process.off("SIGINT", onSigint);
    const durationSeconds =
      stats.firstAt !== undefined && stats.lastAt !== undefined
        ? (stats.lastAt - stats.firstAt) / 1000
        : 0;
    logger.info(
      {
        eventCount: stats.eventCount,
        observedDurationSeconds: Number(durationSeconds.toFixed(2)),
      },
      "stream closed",
    );
  }
}

/**
 * Print a one-line summary of a single stream event.
 *
 * The first frame on a list endpoint is typically the initial snapshot
 * array — but the SDK already flattens NDJSON line-by-line, so each
 * `event` here is a single Note (or the snapshot array when the server
 * delivers it as one JSON document). Both shapes are handled.
 */
function summarise(event: Note | readonly Note[]): void {
  if (Array.isArray(event)) {
    const first = event[0];
    logger.info(
      {
        kind: "snapshot",
        count: event.length,
        firstId: first?.["widget-id"],
        firstText: first ? truncate(first.text, 40) : undefined,
      },
      "snapshot",
    );
    return;
  }
  const note = event as Note;
  logger.info(
    {
      kind: "note-event",
      widgetId: note["widget-id"],
      text: truncate(note.text, 80),
    },
    "note event",
  );
}

function truncate(value: string, max: number): string {
  if (value.length <= max) return value;
  return `${value.slice(0, max - 1)}…`;
}

run().catch((err: unknown) => {
  if (isCanvusError(err)) {
    logger.error({ kind: err.kind, err: err.message }, "streaming failed");
  } else {
    logger.error({ err }, "streaming failed (non-SDK error)");
  }
  process.exit(1);
});
