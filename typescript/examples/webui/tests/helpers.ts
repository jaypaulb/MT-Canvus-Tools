/**
 * Test helpers — fake Session + in-memory stores so route tests don't
 * touch the network or the filesystem.
 *
 * The fake session implements just enough surface area for the route
 * handlers under test. Each test composes the bits it needs via
 * `buildFakeDeps()`.
 */

import { vi } from "vitest";
import type { Logger } from "../src/logging.js";
import type { Config, MutableConfig } from "../src/config.js";
import type {
  JsonStore,
  UsersFile,
  DeletedRecord,
} from "../src/lib/store.js";
import type { AppDeps } from "../src/index.js";
import { RcuStore } from "../src/routes/rcu.js";

/** Minimal Logger stub that records calls but does no IO. */
export function fakeLogger(): Logger {
  return {
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
    debug: vi.fn(),
    trace: vi.fn(),
    fatal: vi.fn(),
    child: vi.fn(() => fakeLogger()),
    level: "info",
  } as unknown as Logger;
}

/** In-memory JsonStore replacement. */
export function fakeStore<T>(initial: T): JsonStore<T> {
  let value: T = initial;
  return {
    read: async (): Promise<T> => value,
    write: async (next: T): Promise<void> => {
      value = next;
    },
    // Cast to satisfy the class shape; tests don't need the private fields.
  } as unknown as JsonStore<T>;
}

/** Fake config matching the zod-parsed Config shape. Override via `overrides`. */
export function fakeConfig(overrides: Partial<Config> = {}): Config {
  return {
    CANVUS_API_URL: "https://canvus.test/api/v1/",
    CANVUS_API_KEY: "test-key",
    CANVUS_CANVAS_ID: "canvas-id-1",
    CANVUS_CANVAS_NAME: "Test Canvas",
    WEBUI_PWD: "admin-pwd",
    PORT: 3000,
    HOSTNAME: "127.0.0.1",
    ALLOW_SELF_SIGNED_CERTS: false,
    COOKIE_SECRET: "test-secret",
    LOG_LEVEL: "silent",
    LOG_FORMAT: "json",
    UPLOAD_DIR: "uploads",
    STATE_DIR: "state",
    ...overrides,
  } as Config;
}

/** Build a deps bundle with a custom (fake) session. */
export function buildFakeDeps(opts: {
  session: unknown;
  config?: Partial<Config>;
  mutable?: Partial<MutableConfig>;
  users?: UsersFile;
  records?: DeletedRecord[];
}): AppDeps {
  const config = fakeConfig(opts.config);
  const mutableConfig: MutableConfig = {
    canvasId: config.CANVUS_CANVAS_ID,
    ...(config.CANVUS_CANVAS_NAME !== undefined && { canvasName: config.CANVUS_CANVAS_NAME }),
    ...opts.mutable,
  };
  return {
    config,
    mutableConfig,
    variables: {
      session: opts.session as never,
      logger: fakeLogger(),
      userStore: fakeStore<UsersFile>(opts.users ?? {}),
      deletedRecordStore: fakeStore<DeletedRecord[]>(opts.records ?? []),
      rcuStore: new RcuStore(),
    },
  };
}
