/**
 * Example 06 — watcher.
 *
 * Subscribes to a canvas's notes and, when a new note appears whose
 * text starts with `?`, treats it as a question. Sends the question
 * (minus the leading `?`) to Ollama's `/api/generate` endpoint and
 * posts the model's answer as a sibling note positioned 400px to the
 * right, with a blue background (`#1d71b8ff`).
 *
 * Dedup: only newly-seen widget ids trigger a response, so the initial
 * snapshot does not produce a flood of answer notes for pre-existing
 * questions. The dedup set is in-memory; restarting the watcher will
 * answer every existing `?`-note again.
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
  base: { component: "example-llm-watcher" },
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
  OLLAMA_URL: z.string().url().default("http://localhost:11434"),
  OLLAMA_MODEL: z.string().min(1).default("llama3.2"),
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

interface OllamaGenerateResponse {
  readonly response: string;
  readonly done: boolean;
}

/** POST to Ollama's `/api/generate` with `stream: false` and return
 * the model's full response string. */
async function askOllama(env: Env, prompt: string): Promise<string> {
  const url = new URL("/api/generate", env.OLLAMA_URL);
  const res = await fetch(url, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify({
      model: env.OLLAMA_MODEL,
      prompt,
      stream: false,
    }),
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`ollama returned ${res.status.toString()}: ${text}`);
  }
  const payload = (await res.json()) as OllamaGenerateResponse;
  return payload.response;
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
      ollamaUrl: env.OLLAMA_URL,
      model: env.OLLAMA_MODEL,
    },
    "watcher ready — ask a question by creating a note starting with '?'",
  );

  const stream = session.widgets.notes.subscribe(env.CANVUS_CANVAS_ID, {
    signal: controller.signal,
  });

  for await (const event of stream) {
    // The SDK's subscribe iterator types each event as a single Note,
    // but the initial snapshot frame arrives as a Note[] array (see
    // streaming.ts in the SDK). Widen at the boundary and normalise.
    const raw: Note | readonly Note[] = event as Note | readonly Note[];
    const notes: readonly Note[] = Array.isArray(raw) ? raw : [raw as Note];

    for (const note of notes) {
      const id = note.id;
      if (seen.has(id)) continue;
      seen.add(id);

      const text = note.text;
      if (!text.startsWith("?")) continue;

      const question = text.slice(1).trim();
      if (question === "") continue;

      logger.info(
        { questionNoteId: id, question: truncate(question, 80) },
        "received question — calling Ollama",
      );

      try {
        const answer = await askOllama(env, question);
        const answered = await session.widgets.notes.create(
          env.CANVUS_CANVAS_ID,
          {
            text: answer,
            location: {
              x: note.location.x + 400,
              y: note.location.y,
            },
            size: note.size,
            background_color: "#1d71b8ff",
            auto_text_color: true,
          },
        );
        // Add the answer's id to `seen` so it does not retrigger.
        seen.add(answered.id);
        logger.info(
          {
            questionNoteId: id,
            answerNoteId: answered.id,
            answerLength: answer.length,
          },
          "posted answer note",
        );
      } catch (err: unknown) {
        if (isCanvusError(err)) {
          logger.error(
            { questionNoteId: id, kind: err.kind, err: err.message },
            "failed to post answer",
          );
        } else {
          logger.error(
            { questionNoteId: id, err },
            "failed to generate or post answer",
          );
        }
      }
    }
  }
}

function truncate(value: string, max: number): string {
  if (value.length <= max) return value;
  return `${value.slice(0, max - 1)}…`;
}

run().catch((err: unknown) => {
  if (isCanvusError(err)) {
    logger.error({ kind: err.kind, err: err.message }, "watcher failed");
  } else {
    logger.error({ err }, "watcher failed (non-SDK error)");
  }
  process.exit(1);
});
