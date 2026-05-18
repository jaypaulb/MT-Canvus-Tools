"""Phase 4b §4.2 #21: batch processor port.

Bounded-concurrency executor for bulk move/copy/delete operations against
canvases and widgets. Ports Go's ``batch.go`` (``BatchProcessor`` +
``BatchOperationBuilder``) with idiomatic Python async semantics.

The async-first design uses :class:`asyncio.Semaphore` for concurrency
limiting and ``asyncio.gather`` for fan-out. Per-operation retries with a
fixed delay match Go's ``RetryAttempts`` + ``RetryDelay``.
"""

from __future__ import annotations

import asyncio
import time
from collections.abc import Callable, Sequence
from dataclasses import dataclass, field
from enum import StrEnum
from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from ..client import Client

__all__ = [
    "BatchConfig",
    "BatchOperation",
    "BatchOperationBuilder",
    "BatchOperationType",
    "BatchProcessor",
    "BatchResult",
    "BatchSummary",
    "summarize",
]

ProgressCallback = Callable[[int, int, list["BatchResult"]], None]


class BatchOperationType(StrEnum):
    """Supported batch operation types."""

    MOVE_CANVAS = "move_canvas"
    COPY_CANVAS = "copy_canvas"
    DELETE_CANVAS = "delete_canvas"
    DELETE_WIDGET = "delete_widget"
    DELETE_USER = "delete_user"


@dataclass(slots=True)
class BatchOperation:
    """One queued operation. ``payload`` carries op-specific args."""

    id: str
    type: BatchOperationType
    payload: dict[str, Any] = field(default_factory=dict)


@dataclass(slots=True)
class BatchResult:
    """Per-operation outcome — success or failure, with timing + retry count."""

    operation_id: str
    success: bool = False
    error: BaseException | None = None
    start_time: float = 0.0
    end_time: float = 0.0
    retries: int = 0

    @property
    def duration(self) -> float:
        return max(0.0, self.end_time - self.start_time)


@dataclass(slots=True)
class BatchConfig:
    """Tuning knobs. Defaults mirror Go's ``DefaultBatchConfig``."""

    max_concurrency: int = 10
    timeout_seconds: float = 300.0
    retry_attempts: int = 3
    retry_delay_seconds: float = 1.0
    continue_on_error: bool = True
    progress_callback: ProgressCallback | None = None


@dataclass(slots=True)
class BatchSummary:
    """Aggregate stats across one ``execute_batch`` call."""

    total_operations: int
    successful: int
    failed: int
    total_duration: float
    average_duration: float
    failed_operations: list[BatchResult]


class BatchProcessor:
    """Executes :class:`BatchOperation` lists with bounded async concurrency."""

    def __init__(self, client: Client, config: BatchConfig | None = None) -> None:
        self.client = client
        cfg = config or BatchConfig()
        if cfg.max_concurrency <= 0:
            cfg.max_concurrency = 1
        elif cfg.max_concurrency > 100:
            cfg.max_concurrency = 100
        self.config = cfg

    async def execute_batch(
        self,
        operations: Sequence[BatchOperation],
    ) -> list[BatchResult]:
        """Run ``operations`` concurrently and return one result per input."""
        if not operations:
            return []

        sem = asyncio.Semaphore(self.config.max_concurrency)
        results: list[BatchResult] = [
            BatchResult(operation_id=op.id) for op in operations
        ]
        completed_count = 0
        completed_lock = asyncio.Lock()

        async def worker(idx: int, operation: BatchOperation) -> None:
            nonlocal completed_count
            async with sem:
                result = await self._execute_one(operation)
            results[idx] = result
            async with completed_lock:
                completed_count += 1
                if self.config.progress_callback is not None:
                    self.config.progress_callback(
                        completed_count, len(operations), list(results)
                    )

        async def runner() -> None:
            await asyncio.gather(
                *(worker(i, op) for i, op in enumerate(operations)),
                return_exceptions=False,
            )

        if self.config.timeout_seconds > 0.0:
            await asyncio.wait_for(runner(), timeout=self.config.timeout_seconds)
        else:
            await runner()
        return results

    async def _execute_one(self, operation: BatchOperation) -> BatchResult:
        result = BatchResult(operation_id=operation.id, start_time=time.monotonic())
        for attempt in range(self.config.retry_attempts + 1):
            result.retries = attempt
            try:
                await self._dispatch(operation)
            except Exception as exc:
                result.error = exc
                if attempt == self.config.retry_attempts:
                    break
                await asyncio.sleep(self.config.retry_delay_seconds)
                continue
            result.success = True
            result.error = None
            break
        result.end_time = time.monotonic()
        return result

    async def _dispatch(self, op: BatchOperation) -> None:
        if op.type is BatchOperationType.MOVE_CANVAS:
            await self.client.canvases.move(
                str(op.payload["canvas_id"]), str(op.payload["folder_id"])
            )
            return
        if op.type is BatchOperationType.COPY_CANVAS:
            await self.client.canvases.copy(
                str(op.payload["canvas_id"]),
                {"folder_id": str(op.payload["folder_id"])},
            )
            return
        if op.type is BatchOperationType.DELETE_CANVAS:
            await self.client.canvases.delete(str(op.payload["canvas_id"]))
            return
        if op.type is BatchOperationType.DELETE_WIDGET:
            await self.client.widgets.delete_any(
                str(op.payload["canvas_id"]),
                str(op.payload["widget_id"]),
                str(op.payload["widget_type"]),
            )
            return
        if op.type is BatchOperationType.DELETE_USER:
            await self.client.users.delete(str(op.payload["user_id"]))
            return
        raise ValueError(f"unsupported batch operation type: {op.type}")


class BatchOperationBuilder:
    """Fluent helper for assembling :class:`BatchOperation` lists."""

    def __init__(self) -> None:
        self._ops: list[BatchOperation] = []

    def move_canvas(self, op_id: str, canvas_id: str, folder_id: str) -> BatchOperationBuilder:
        self._ops.append(
            BatchOperation(
                id=op_id,
                type=BatchOperationType.MOVE_CANVAS,
                payload={"canvas_id": canvas_id, "folder_id": folder_id},
            )
        )
        return self

    def copy_canvas(self, op_id: str, canvas_id: str, folder_id: str) -> BatchOperationBuilder:
        self._ops.append(
            BatchOperation(
                id=op_id,
                type=BatchOperationType.COPY_CANVAS,
                payload={"canvas_id": canvas_id, "folder_id": folder_id},
            )
        )
        return self

    def delete_canvas(self, op_id: str, canvas_id: str) -> BatchOperationBuilder:
        self._ops.append(
            BatchOperation(
                id=op_id,
                type=BatchOperationType.DELETE_CANVAS,
                payload={"canvas_id": canvas_id},
            )
        )
        return self

    def delete_widget(
        self,
        op_id: str,
        canvas_id: str,
        widget_id: str,
        widget_type: str,
    ) -> BatchOperationBuilder:
        self._ops.append(
            BatchOperation(
                id=op_id,
                type=BatchOperationType.DELETE_WIDGET,
                payload={
                    "canvas_id": canvas_id,
                    "widget_id": widget_id,
                    "widget_type": widget_type,
                },
            )
        )
        return self

    def delete_user(self, op_id: str, user_id: str) -> BatchOperationBuilder:
        self._ops.append(
            BatchOperation(
                id=op_id,
                type=BatchOperationType.DELETE_USER,
                payload={"user_id": user_id},
            )
        )
        return self

    def build(self) -> list[BatchOperation]:
        return list(self._ops)


def summarize(results: Sequence[BatchResult]) -> BatchSummary:
    """Aggregate per-op outcomes into a :class:`BatchSummary`."""
    failed = [r for r in results if not r.success]
    total_duration = sum(r.duration for r in results)
    return BatchSummary(
        total_operations=len(results),
        successful=sum(1 for r in results if r.success),
        failed=len(failed),
        total_duration=total_duration,
        average_duration=(total_duration / len(results)) if results else 0.0,
        failed_operations=failed,
    )
