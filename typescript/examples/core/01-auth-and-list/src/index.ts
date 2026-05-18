/**
 * Example 01 — auth-and-list.
 *
 * Smoke test: authenticate with an API key, list canvases, and print a
 * summary table to stdout. If this works, your `CANVUS_BASE_URL` and
 * `CANVUS_API_KEY` are valid and your network can reach the server.
 */

import { createSession, isCanvusError } from "@mt-canvus-tools/sdk";
import pino from "pino";
import { z } from "zod";

const isPretty = process.env["LOG_FORMAT"] === "pretty";

const logger = pino({
  level: process.env["LOG_LEVEL"] ?? "info",
  base: { component: "example-auth-and-list" },
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
});

type Env = z.infer<typeof envSchema>;

/** Load and validate `CANVUS_*` environment variables. Exits 2 on failure. */
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

/** Truncate a string to `max` characters, with an ellipsis if cut. */
function truncate(value: string, max: number): string {
  if (value.length <= max) return value;
  return `${value.slice(0, max - 1)}…`;
}

/** Pad to a fixed column width on the right. */
function pad(value: string, width: number): string {
  return value.length >= width ? value : value + " ".repeat(width - value.length);
}

async function run(): Promise<void> {
  const env = loadEnv();

  const session = createSession({
    baseUrl: env.CANVUS_BASE_URL,
    apiKey: env.CANVUS_API_KEY,
  });

  logger.info({ baseUrl: env.CANVUS_BASE_URL }, "listing canvases");
  const canvases = await session.canvases.list();
  logger.info({ count: canvases.length }, "fetched canvases");

  // Format as a padded table. Avoid console.table because it does not
  // truncate UUIDs and produces unreadable output when names are long.
  const header = `${pad("ID", 38)}${pad("NAME", 42)}${pad("OWNER", 24)}MODIFIED`;
  process.stdout.write(`${header}\n`);
  process.stdout.write(`${"-".repeat(header.length)}\n`);
  for (const canvas of canvases) {
    const id = canvas["canvas-id"];
    const name = truncate(canvas["canvas-name"], 40);
    const owner = truncate(canvas.owner, 22);
    const modified = canvas.modified;
    process.stdout.write(`${pad(id, 38)}${pad(name, 42)}${pad(owner, 24)}${modified}\n`);
  }
}

run().catch((err: unknown) => {
  if (isCanvusError(err)) {
    logger.error({ kind: err.kind, err: err.message }, "auth-and-list failed");
  } else {
    logger.error({ err }, "auth-and-list failed (non-SDK error)");
  }
  process.exit(1);
});
