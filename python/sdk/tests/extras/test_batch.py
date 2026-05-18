"""Phase 4b §4.2 #21: batch processor tests."""

from __future__ import annotations

import pytest
import respx
from httpx import Response

from canvus_sdk import Client
from canvus_sdk.extras import (
    BatchConfig,
    BatchOperationBuilder,
    BatchProcessor,
    summarize,
)


@pytest.mark.asyncio
async def test_batch_executes_move_and_delete(client: Client) -> None:
    operations = (
        BatchOperationBuilder()
        .move_canvas("op1", "c1", "f1")
        .delete_canvas("op2", "c2")
        .build()
    )
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        mock.post("canvases/c1/move").mock(
            return_value=Response(200, json={"id": "c1", "name": "n", "asset_size": 0})
        )
        mock.delete("canvases/c2").mock(return_value=Response(204))
        results = await BatchProcessor(client).execute_batch(operations)
    assert [r.success for r in results] == [True, True]


@pytest.mark.asyncio
async def test_batch_retries_then_succeeds(client: Client) -> None:
    operations = BatchOperationBuilder().delete_canvas("d", "c1").build()
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        # First call 500, second 204. Pyright/mypy: respx side_effect by listing routes.
        route = mock.delete("canvases/c1")
        route.side_effect = [
            Response(500, json={"error": "transient"}),
            Response(204),
        ]
        cfg = BatchConfig(
            max_concurrency=1, retry_attempts=2, retry_delay_seconds=0.0
        )
        results = await BatchProcessor(client, cfg).execute_batch(operations)
    assert results[0].success is True
    assert results[0].retries == 1


@pytest.mark.asyncio
async def test_batch_summary() -> None:
    from canvus_sdk.extras.batch import BatchResult

    results = [
        BatchResult(operation_id="a", success=True, start_time=0.0, end_time=1.0),
        BatchResult(operation_id="b", success=False, start_time=0.0, end_time=2.0),
    ]
    summary = summarize(results)
    assert summary.total_operations == 2
    assert summary.successful == 1
    assert summary.failed == 1
    assert summary.total_duration == 3.0
    assert summary.average_duration == 1.5
    assert len(summary.failed_operations) == 1
