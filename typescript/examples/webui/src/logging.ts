/**
 * Pino logger for the WebUI example.
 *
 * Mirrors the SDK's own logger contract so call sites that import from
 * here look identical to call sites that import from `@mt-canvus-tools/sdk`.
 */

import pino, { type Logger } from "pino";

export function createLogger(opts: {
  level: string;
  format: "json" | "pretty";
  component?: string;
}): Logger {
  return pino({
    level: opts.level,
    base: opts.component ? { component: opts.component } : {},
    ...(opts.format === "pretty"
      ? {
          transport: {
            target: "pino-pretty",
            options: { colorize: true, translateTime: "SYS:HH:MM:ss.l" },
          },
        }
      : {}),
  });
}

export type { Logger };
