# 05 — streaming

Opens a long-lived `?subscribe=true` connection against a canvas's notes
endpoint and prints a one-line summary per event for a configurable
duration (default 30 s). Cleanly handles `SIGINT`.

## Wire-format notes

The Canvus streaming protocol uses **newline-delimited JSON** (NDJSON), not
Server-Sent Events. Each line is either:

- A **snapshot batch** (initial dump of existing widgets — a JSON array).
- An **incremental update** (a single JSON object or small array).
- A **keep-alive** (empty line) — the SDK silently skips these.

The SDK's `client.widgets.subscribe(canvas_id, widget_type="notes")` returns
an async generator that yields one decoded JSON value per non-empty line.

## Prerequisites

| Env var                    | Required | Purpose                          |
| -------------------------- | -------- | -------------------------------- |
| `CANVUS_API_URL`           | yes      | Base URL.                        |
| `CANVUS_API_KEY`           | yes      | API token.                       |
| `CANVUS_CANVAS_ID`         | yes      | Canvas to watch.                 |
| `STREAM_DURATION_SECONDS`  | no       | Seconds to run for (default 30). |

## Run

```bash
cd python
python examples/core/05-streaming/main.py
```

Then, in another terminal or in the Canvus UI, create / edit notes on the
target canvas to see them flow through.

## Expected output

```
subscribing                           canvas_id=… duration_seconds=30.0
batch                                 batch=1 count=4 first_id=… first_text=…
event                                 batch=2 id=… text=…
…
sigint received; draining stream
stream finished                       batches=8 events_total=12 duration_seconds=30.0
received 12 events across 8 batches in 30.0s
```

## Concurrency model

The consumer task races each `__anext__()` against a stop-event wait, so
both `SIGINT` and the duration timeout interrupt the iteration promptly
without blocking on the next server-pushed line.

## Troubleshooting

- **No events at all:** the canvas may be idle; create or edit a note to
  force a server-side push, or verify connectivity with example 01.
- **Stream closes immediately:** the API key likely cannot view the canvas
  (subscribe inherits the same auth as a regular read).
- **Hangs after `STREAM_DURATION_SECONDS`:** the underlying HTTP connection
  may be wedged — bump `LOG_LEVEL=DEBUG` to inspect.
