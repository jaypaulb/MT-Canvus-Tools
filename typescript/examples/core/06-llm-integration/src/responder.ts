/**
 * Example 06 — responder.
 *
 * Lightweight diagnostic for Ollama: lists available models via
 * `/api/tags`. Use this before running the watcher to confirm Ollama
 * is reachable and that the model you want is pulled.
 */

import pino from "pino";
import { z } from "zod";

const isPretty = process.env["LOG_FORMAT"] === "pretty";

const logger = pino({
  level: process.env["LOG_LEVEL"] ?? "info",
  base: { component: "example-llm-responder" },
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
  OLLAMA_URL: z.string().url().default("http://localhost:11434"),
});

interface OllamaTagsResponse {
  readonly models: readonly {
    readonly name: string;
    readonly size?: number;
    readonly modified_at?: string;
  }[];
}

async function run(): Promise<void> {
  const env = envSchema.parse(process.env);
  const url = new URL("/api/tags", env.OLLAMA_URL);

  logger.info({ url: url.toString() }, "querying Ollama tags");
  const res = await fetch(url);
  if (!res.ok) {
    const text = await res.text();
    logger.error(
      { status: res.status, body: text },
      "Ollama returned non-2xx — is the server running?",
    );
    process.exit(1);
  }
  const payload = (await res.json()) as OllamaTagsResponse;
  if (payload.models.length === 0) {
    logger.warn(
      { url: url.toString() },
      "Ollama is reachable but has no models — run `ollama pull <model>` first",
    );
    return;
  }
  logger.info({ modelCount: payload.models.length }, "available models");
  for (const model of payload.models) {
    process.stdout.write(`  - ${model.name}\n`);
  }
}

run().catch((err: unknown) => {
  logger.error({ err }, "responder failed");
  process.exit(1);
});
