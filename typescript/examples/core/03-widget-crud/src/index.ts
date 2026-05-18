/**
 * Example 03 — widget-crud.
 *
 * Full lifecycle on a sticky note:
 *   1. Create a note with the current timestamp in the text.
 *   2. Print the new widget id.
 *   3. PATCH the note with new text + a green background.
 *   4. Print the updated widget.
 *   5. DELETE the note.
 *   6. Verify deletion by GETting the same id and expecting NotFoundError.
 *
 * If any step fails the example stops there so the caller can debug the
 * failing step — there is no point continuing once the chain is broken.
 */

import {
  createSession,
  isCanvusError,
  NotFoundError,
} from "@mt-canvus-tools/sdk";
import pino from "pino";
import { z } from "zod";

const isPretty = process.env["LOG_FORMAT"] === "pretty";

const logger = pino({
  level: process.env["LOG_LEVEL"] ?? "info",
  base: { component: "example-widget-crud" },
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

async function run(): Promise<void> {
  const env = loadEnv();
  const session = createSession({
    baseUrl: env.CANVUS_API_URL,
    apiKey: env.CANVUS_API_KEY,
  });
  const canvasId = env.CANVUS_CANVAS_ID;

  // Step 1: create.
  const createdAt = new Date().toISOString();
  const created = await session.widgets.notes.create(canvasId, {
    text: `hello from TypeScript @ ${createdAt}`,
    location: { x: 200, y: 200 },
    size: { width: 300, height: 200 },
  });
  const noteId = created.id;
  logger.info({ noteId, text: created.text }, "created note");

  // Step 3: update.
  const updatedAt = new Date().toISOString();
  const updated = await session.widgets.notes.update(canvasId, noteId, {
    text: `updated from TypeScript @ ${updatedAt}`,
    background_color: "#3aaa34ff",
  });
  logger.info(
    {
      noteId: updated.id,
      text: updated.text,
      backgroundColor: updated.background_color,
    },
    "updated note",
  );

  // Step 5: delete.
  await session.widgets.notes.delete(canvasId, noteId);
  logger.info({ noteId }, "deleted note");

  // Step 6: verify delete via the SDK's NotFoundError.
  try {
    await session.widgets.notes.get(canvasId, noteId);
    logger.error({ noteId }, "delete verification FAILED — note still exists");
    process.exit(1);
  } catch (err: unknown) {
    if (err instanceof NotFoundError) {
      logger.info({ noteId }, "delete verified — GET returned 404 as expected");
      return;
    }
    throw err;
  }
}

run().catch((err: unknown) => {
  if (isCanvusError(err)) {
    logger.error({ kind: err.kind, err: err.message }, "widget-crud failed");
  } else {
    logger.error({ err }, "widget-crud failed (non-SDK error)");
  }
  process.exit(1);
});
