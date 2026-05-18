# Example 07 — webhooks-notifications

Bridges Canvus widget events to an outbound HTTP webhook.

## Purpose

The Canvus API has no built-in webhook system. This example shows the
**pattern** for building one on top of the streaming subscribe API:

- Subscribe to a canvas's widgets stream.
- Filter out the initial snapshot (everything that already exists when
  the bridge starts).
- For every genuinely-new widget event, POST a JSON envelope to
  `WEBHOOK_URL`.
- Retry on non-2xx with exponential backoff (2s / 4s / 8s, total
  three retries on top of the original attempt).

## Prerequisites

- Node 20+ and pnpm 9+
- An API key with view access to `CANVUS_CANVAS_ID`
- A `WEBHOOK_URL` to POST to. For testing, create a unique inbox at
  [webhook.site](https://webhook.site) and use the URL it gives you.

## Configuration

| Variable           | Required | Description                                       |
| ------------------ | -------- | ------------------------------------------------- |
| `CANVUS_BASE_URL`  | yes      | API base URL                                      |
| `CANVUS_API_KEY`   | yes      | API key with view access                          |
| `CANVUS_CANVAS_ID` | yes      | Canvas to watch                                   |
| `WEBHOOK_URL`      | yes      | Receiving webhook endpoint                        |

## Run

```bash
pnpm dev
```

Then add a widget to the canvas (a note, an image, anything). Within
a few seconds, your webhook receiver should see a POST with the body:

```json
{
  "event": "widget.created",
  "canvas_id": "...",
  "widget_id": "...",
  "widget_type": "note",
  "timestamp": "2026-05-18T12:34:56.789Z"
}
```

Press Ctrl-C to stop the bridge.

## How it works

### Snapshot filtering

The subscribe stream's first frame is typically the **initial
snapshot** — an array of every existing widget. Treating those as
"newly created" would spam the webhook every time the bridge starts.
The code distinguishes snapshot frames (arrays) from delta frames
(single objects) and only POSTs for the latter. Snapshot widgets are
added to the in-memory `seen` set so that if they show up later as
deltas (e.g. via a re-emit), they are still ignored.

### Retry policy

Each POST is tried up to 4 times total:

| Attempt | Delay before next attempt |
| ------- | ------------------------- |
| 1       | 2s                        |
| 2       | 4s                        |
| 3       | 8s                        |
| 4       | (give up)                 |

Both non-2xx responses and thrown errors trigger a retry. Permanent
failures are logged at `error` level so an upstream log shipper can
alert on them.

### Sequential delivery

POSTs are awaited sequentially so a slow webhook back-pressures the
event stream rather than queueing unbounded retries. For higher
throughput, batch events into a bounded worker pool — but be aware
that out-of-order delivery becomes a possibility.

## Troubleshooting

- **No POSTs ever arrive** — confirm a widget was actually created
  (not just modified). This bridge only forwards `widget.created`
  events; PATCH events are ignored.
- **Webhook returns 4xx** — the receiver rejected the body shape.
  Check what it expects and adjust `OutboundEvent`.
- **Webhook returns 5xx repeatedly** — backoff hits the cap and the
  event is dropped. Production deployments should persist failed
  events to a dead-letter queue.
