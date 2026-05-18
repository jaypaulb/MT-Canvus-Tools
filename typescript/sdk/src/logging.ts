import pino from "pino";

/**
 * Whether to render logs via `pino-pretty` (human-readable, coloured)
 * instead of raw JSON to stdout.
 */
const isPretty = process.env.LOG_FORMAT === "pretty";

/**
 * The single configured pino logger for the SDK.
 *
 * Level is taken from `LOG_LEVEL` (default `info`). When `LOG_FORMAT=pretty`,
 * the logger pipes through `pino-pretty` for development ergonomics.
 */
export const logger = pino({
  level: process.env.LOG_LEVEL ?? "info",
  base: { component: "sdk" },
  ...(isPretty && {
    transport: {
      target: "pino-pretty",
      options: { colorize: true, translateTime: "SYS:HH:MM:ss.l" },
    },
  }),
});

/**
 * The concrete type of the SDK logger. Importers should depend on this
 * (rather than `pino.Logger`) so changes to the configured shape are
 * statically caught.
 */
export type Logger = typeof logger;
