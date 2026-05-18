# Example 05 — streaming

Subscribes to the live notes stream on a canvas and prints each event
as it arrives.

## Purpose

Demonstrates the SDK's NDJSON streaming API
(`session.widgets.notes.subscribe(canvasId)` returns an
`AsyncGenerator<Note>`) and the clean-shutdown pattern:

- A duration timer aborts the stream after
  `STREAM_DURATION_SECONDS` (default 30).
- A SIGINT handler aborts the stream on Ctrl-C.
- Either path drains and exits cleanly with a one-line summary.

The 15-second keepalive newlines that the Canvus server emits are
filtered out by the SDK; they never reach the consumer's loop.

## Prerequisites

- Node 20+ and pnpm 9+
- An API key with view access to `CANVUS_CANVAS_ID`
- Optional: a way to add or modify a note on that canvas during the
  run (e.g. the Canvus desktop client, or example 03 running in
  another terminal)

## Configuration

| Variable                  | Required | Default | Description                       |
| ------------------------- | -------- | ------- | --------------------------------- |
| `CANVUS_BASE_URL`         | yes      |         | API base URL                      |
| `CANVUS_API_KEY`          | yes      |         | API key with view access          |
| `CANVUS_CANVAS_ID`        | yes      |         | Canvas to subscribe to            |
| `STREAM_DURATION_SECONDS` | no       | `30`    | How long to stream before exiting |

## Run

```bash
pnpm dev
```

While it runs, open another terminal and run example 03 (or modify a
note in the Canvus desktop client). You should see events stream in.

## How it works

The SDK exposes streaming endpoints as async iterators:

```ts
const stream = session.widgets.notes.subscribe(canvasId, {
  signal: controller.signal,
});

for await (const event of stream) {
  // event is a Note (or the initial snapshot array)
}
```

Internally the SDK uses `undici.request` (not `fetch`) to get true
backpressure on the response body. Each NDJSON line is JSON-parsed and
yielded; empty keepalive lines are skipped. Breaking out of the
`for await` loop with `break`, `return`, or by aborting the signal
disposes the iterator and closes the underlying connection.

### Snapshot vs. event

The first frame on a list-endpoint subscribe is the **initial snapshot**
(every existing item). Subsequent frames are deltas (individual items
that have changed). The code handles both: arrays are summarised as
"snapshot of N items", single items are printed individually.

## Troubleshooting

- **No events ever arrive** — confirm someone is actually modifying
  notes on the canvas. The initial snapshot will fire immediately
  regardless.
- **Stream disconnects after a few seconds** — check for proxies or
  load balancers in front of your server with aggressive idle
  timeouts. The Canvus server's 15-second keepalive is designed to
  defeat this, but some L7 proxies still close connections.
- **Aborting takes a long time** — the underlying `undici` body
  reader yields control only at chunk boundaries. A small grace
  period (sub-second) on abort is normal.
