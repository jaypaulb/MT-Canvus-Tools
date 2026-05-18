// Phase 4b §4.3 #19: batch processor.
//
// Mirrors Go's `batch.go` — bounded-concurrency move/copy/delete operations
// with retry, ContinueOnError, and a progress callback.

import type { Session } from "../session.js";
import type { Canvas } from "../types/canvas.js";
import type { Uuid } from "../types/common.js";

/** Supported batch operation types. */
export type BatchOperationType = "move" | "copy" | "delete";

/** A single operation queued in a batch. */
export interface BatchOperation {
  readonly id: string;
  readonly type: BatchOperationType;
  /** Resource the operation acts on. For canvas ops, the Canvas itself. */
  readonly resource: Canvas | { readonly canvasId: Uuid; readonly widgetId: Uuid; readonly widgetType: string };
  /** Target descriptor — for move/copy a folder ID string. */
  readonly target?: string;
}

/** Per-operation outcome. */
export interface BatchResult {
  readonly operationId: string;
  readonly success: boolean;
  readonly error?: string;
  readonly startTime: Date;
  readonly endTime: Date;
  readonly durationMs: number;
  readonly retries: number;
}

/** Configuration for {@link BatchProcessor}. */
export interface BatchConfig {
  readonly maxConcurrency?: number;
  readonly timeoutMs?: number;
  readonly retryAttempts?: number;
  readonly retryDelayMs?: number;
  readonly continueOnError?: boolean;
  readonly progressCallback?: (completed: number, total: number) => void;
}

const DEFAULTS: Required<Omit<BatchConfig, "progressCallback">> = {
  maxConcurrency: 10,
  timeoutMs: 5 * 60 * 1000,
  retryAttempts: 3,
  retryDelayMs: 1000,
  continueOnError: true,
};

/**
 * Bounded-concurrency batch runner. Construct one per workload and call
 * `execute`. Results are returned in the same order as the input operations.
 */
export class BatchProcessor {
  private readonly cfg: Required<Omit<BatchConfig, "progressCallback">> & {
    readonly progressCallback?: BatchConfig["progressCallback"];
  };

  constructor(
    private readonly session: Session,
    config: BatchConfig = {},
  ) {
    const merged = { ...DEFAULTS, ...config };
    if (merged.maxConcurrency <= 0) merged.maxConcurrency = 1;
    if (merged.maxConcurrency > 100) merged.maxConcurrency = 100;
    this.cfg = {
      ...merged,
      ...(config.progressCallback !== undefined && { progressCallback: config.progressCallback }),
    };
  }

  async execute(operations: readonly BatchOperation[]): Promise<readonly BatchResult[]> {
    if (operations.length === 0) return [];
    const results = new Array<BatchResult | undefined>(operations.length);
    let completed = 0;
    let next = 0;

    const worker = async (): Promise<void> => {
      for (;;) {
        const idx = next++;
        if (idx >= operations.length) return;
        const op = operations[idx];
        if (op === undefined) continue;
        const result = await this.runOnce(op);
        results[idx] = result;
        completed += 1;
        this.cfg.progressCallback?.(completed, operations.length);
        if (!result.success && !this.cfg.continueOnError) {
          next = operations.length;
        }
      }
    };

    const workers: Promise<void>[] = [];
    for (let i = 0; i < Math.min(this.cfg.maxConcurrency, operations.length); i++) {
      workers.push(worker());
    }
    await Promise.all(workers);
    return results.map((r, i) => r ?? skippedResult(operations[i]?.id ?? `op_${i.toString()}`));
  }

  private async runOnce(op: BatchOperation): Promise<BatchResult> {
    const start = new Date();
    let lastError = "";
    for (let attempt = 0; attempt <= this.cfg.retryAttempts; attempt++) {
      try {
        await this.dispatch(op);
        const end = new Date();
        return {
          operationId: op.id,
          success: true,
          startTime: start,
          endTime: end,
          durationMs: end.getTime() - start.getTime(),
          retries: attempt,
        };
      } catch (err) {
        lastError = (err as Error).message;
        if (attempt === this.cfg.retryAttempts) break;
        await sleep(this.cfg.retryDelayMs);
      }
    }
    const end = new Date();
    return {
      operationId: op.id,
      success: false,
      error: lastError,
      startTime: start,
      endTime: end,
      durationMs: end.getTime() - start.getTime(),
      retries: this.cfg.retryAttempts,
    };
  }

  private async dispatch(op: BatchOperation): Promise<void> {
    switch (op.type) {
      case "move":
        if (!("id" in op.resource)) throw new Error("move: resource must be a Canvas");
        if (op.target === undefined) throw new Error("move: target folder_id is required");
        await this.session.canvases.move(op.resource.id, { folder_id: op.target });
        return;
      case "copy":
        if (!("id" in op.resource)) throw new Error("copy: resource must be a Canvas");
        if (op.target === undefined) throw new Error("copy: target folder_id is required");
        await this.session.canvases.copy(op.resource.id, { folder_id: op.target });
        return;
      case "delete":
        if ("id" in op.resource) {
          await this.session.canvases.delete(op.resource.id);
          return;
        }
        await this.session.widgets.deleteAny(
          op.resource.canvasId,
          op.resource.widgetId,
          op.resource.widgetType,
        );
        return;
    }
  }
}

/** Fluent builder for {@link BatchOperation} lists. */
export class BatchOperationBuilder {
  private readonly ops: BatchOperation[] = [];

  move(id: string, canvas: Canvas, targetFolderId: string): this {
    this.ops.push({ id, type: "move", resource: canvas, target: targetFolderId });
    return this;
  }

  copy(id: string, canvas: Canvas, targetFolderId: string): this {
    this.ops.push({ id, type: "copy", resource: canvas, target: targetFolderId });
    return this;
  }

  deleteCanvas(id: string, canvas: Canvas): this {
    this.ops.push({ id, type: "delete", resource: canvas });
    return this;
  }

  deleteWidget(id: string, canvasId: Uuid, widgetId: Uuid, widgetType: string): this {
    this.ops.push({
      id,
      type: "delete",
      resource: { canvasId, widgetId, widgetType },
    });
    return this;
  }

  build(): readonly BatchOperation[] {
    return this.ops;
  }
}

/** Aggregate stats from a batch run. */
export interface BatchSummary {
  readonly totalOperations: number;
  readonly successful: number;
  readonly failed: number;
  readonly totalDurationMs: number;
  readonly averageDurationMs: number;
  readonly failedOperations: readonly BatchResult[];
}

/** Compute summary statistics from a result list. */
export function summarize(results: readonly BatchResult[]): BatchSummary {
  let total = 0;
  let success = 0;
  let fail = 0;
  const failed: BatchResult[] = [];
  for (const r of results) {
    total += r.durationMs;
    if (r.success) success++;
    else {
      fail++;
      failed.push(r);
    }
  }
  return {
    totalOperations: results.length,
    successful: success,
    failed: fail,
    totalDurationMs: total,
    averageDurationMs: results.length > 0 ? total / results.length : 0,
    failedOperations: failed,
  };
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function skippedResult(id: string): BatchResult {
  const now = new Date();
  return {
    operationId: id,
    success: false,
    error: "skipped (batch halted by ContinueOnError=false)",
    startTime: now,
    endTime: now,
    durationMs: 0,
    retries: 0,
  };
}
