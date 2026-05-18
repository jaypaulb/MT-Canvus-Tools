/**
 * Demo-grade JSON persistence.
 *
 * The legacy CanvusWebUI stored everything on disk in flat JSON files
 * (`users.json`, `macros-deleted-records.json`). This is fine for a
 * demo: it survives restarts, supports the existing admin tools, and
 * avoids dragging in SQLite for an example app.
 *
 * NOT FOR PRODUCTION: concurrent writes are race-prone, files grow
 * unbounded without ops intervention. A real deployment should swap
 * these helpers for a SQLite or KV store.
 */

import { promises as fs } from "node:fs";
import { mkdirSync, existsSync } from "node:fs";
import path from "node:path";

/** Users keyed by team → name → color (#RRGGBBAA). */
export type UsersFile = Record<string, Record<string, string>>;

export interface DeletedRecord {
  readonly recordId: string;
  readonly timestamp: string;
  readonly zoneId: string;
  readonly widgets: ReadonlyArray<Record<string, unknown>>;
}

/** Demo-grade JSON file store. One handle per file, no locking. */
export class JsonStore<T> {
  constructor(
    private readonly filePath: string,
    private readonly initialValue: T,
  ) {
    const dir = path.dirname(filePath);
    if (!existsSync(dir)) mkdirSync(dir, { recursive: true });
  }

  async read(): Promise<T> {
    try {
      const raw = await fs.readFile(this.filePath, "utf8");
      return JSON.parse(raw) as T;
    } catch (err) {
      const code = (err as NodeJS.ErrnoException).code;
      if (code === "ENOENT") {
        await this.write(this.initialValue);
        return this.initialValue;
      }
      throw err;
    }
  }

  async write(value: T): Promise<void> {
    await fs.writeFile(this.filePath, JSON.stringify(value, null, 2), "utf8");
  }
}

export function buildUserStore(stateDir: string): JsonStore<UsersFile> {
  return new JsonStore<UsersFile>(path.join(stateDir, "users.json"), {});
}

export function buildDeletedRecordStore(stateDir: string): JsonStore<DeletedRecord[]> {
  return new JsonStore<DeletedRecord[]>(path.join(stateDir, "macros-deleted-records.json"), []);
}

/**
 * Trim a deleted-records list to the most-recent `max` entries.
 *
 * The legacy server kept up to 50 records before pruning the oldest.
 * Returned array is fresh — does not mutate the input.
 */
export function trimDeletedRecords(records: readonly DeletedRecord[], max = 50): DeletedRecord[] {
  if (records.length <= max) return [...records];
  return records.slice(records.length - max);
}
