/**
 * Hono `Variables` typing for the WebUI app.
 *
 * Centralises the per-request state the middleware injects so route
 * modules can pull strongly-typed dependencies off `c.var.*` without
 * each module re-declaring the shape.
 */

import type { Session } from "@mt-canvus-tools/sdk";
import type { Logger } from "./logging.js";
import type { Config, MutableConfig } from "./config.js";
import type { JsonStore, UsersFile, DeletedRecord } from "./lib/store.js";
import type { RcuStore } from "./routes/rcu.js";

export interface AppVariables {
  readonly session: Session;
  readonly logger: Logger;
  readonly config: Config;
  readonly mutableConfig: MutableConfig;
  readonly userStore: JsonStore<UsersFile>;
  readonly deletedRecordStore: JsonStore<DeletedRecord[]>;
  readonly rcuStore: RcuStore;
}

export interface AppEnv {
  Variables: AppVariables;
}
