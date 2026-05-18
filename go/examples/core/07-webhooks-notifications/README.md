# 07 — Webhook Notifications (Synthetic)

The Canvus API does not have native outbound webhooks. This example shows
the canonical pattern for building them on top of the streaming subscribe
endpoint:

1. Subscribe to `/canvases/{id}/widgets?subscribe=true`.
2. Track every widget ID we've ever seen.
3. For each *new* widget that appears after the initial snapshot, POST a
   small JSON event to a configurable webhook URL.
4. Retry non-2xx responses with exponential backoff (2s, 4s, 8s; max 3
   attempts).

Use [webhook.site](https://webhook.site) for testing — paste its
one-shot URL into `WEBHOOK_URL` and watch deliveries land in your browser.

## Prerequisites

- API key with read access on the target canvas.
- A reachable webhook receiver. For local development, run
  `nc -l 8080` or use webhook.site.

## Configuration

| Variable | Required | Description |
| --- | --- | --- |
| `CANVUS_BASE_URL` | yes | Full base URL including `/api/v1/`. |
| `CANVUS_API_KEY` | yes | API key with read access. |
| `CANVUS_CANVAS_ID` | yes | UUID of canvas to monitor. |
| `WEBHOOK_URL` | yes | Outbound webhook receiver. |
| `LOG_FORMAT` | no | `text` (default) or `json`. |

## Running

```bash
go run .
```

In a separate session, add a note to the canvas (run example 03, or create
one in the Canvus UI). The bridge logs:

```
level=INFO msg="webhook delivered" widget_id=<id> widget_type=note status=200 attempt=1 latency_ms=87
```

The webhook receiver sees a POST with this body:

```json
{
  "event": "widget.created",
  "canvas_id": "<canvas uuid>",
  "widget_id": "<widget uuid>",
  "widget_type": "note",
  "timestamp": "2026-05-18T12:34:56.789Z"
}
```

Ctrl-C for clean shutdown.

## How it works

### Initial-snapshot filter
The subscribe stream sends the full current widget list on connect. Without
filtering, the bridge would synthesise a `widget.created` for every widget
that already existed — useless and noisy. We treat the first frame as a
state-establishing snapshot: every ID is added to `seen` but no events are
emitted. From frame 2 onwards, only IDs not in `seen` trigger a delivery.

### What counts as a "new widget"?
A widget whose ID we have not seen this process lifetime. We do not
distinguish "newly created" from "newly visible to us" — for a fresh
subscription with an empty `seen` map, those are equivalent. If you
restart the bridge, the next initial snapshot reset means existing widgets
are ignored, and any deltas during the downtime are missed. Production
systems would persist `seen` to disk; this example is deliberately
in-memory.

### Retry policy
On non-2xx HTTP response or transport error, the delivery is retried up to
3 times with `2s, 4s, 8s` backoff. The backoff slots are interruptible by
context cancellation (Ctrl-C does not block waiting for the next attempt).

We retry on **every** non-2xx, including 4xx — webhook receivers
occasionally return 502s on cold start, and treating a 404 as fatal would
just lose data. If your receiver wants to signal "don't retry", it should
return 2xx and discard.

### Long-lived stream
Same as example 05: the SDK's default 30s request timeout is overridden in
the `SessionConfig` to 10 minutes.

## Architectural notes

- **No SDK Subscribe helper** — direct `http.NewRequestWithContext` on the
  session's exported HTTPClient. Same pattern as examples 05 and 06.
- **`widget.created` only** — by design. The subscribe stream emits snapshots,
  not deltas, so deriving `widget.updated` requires diffing frames and
  detecting `widget.deleted` requires detecting absence. Both are
  straightforward to add but out of scope for this minimal example.
- **No event ordering guarantees** — within a single frame, events are
  delivered in iteration order of the widget array. The frame itself is the
  ordering boundary.

## Troubleshooting

| Symptom | Likely cause |
| --- | --- |
| Bridge connects, "initial snapshot recorded" logs, then silence | The canvas has not gained any new widgets. Add a note. |
| Webhook deliveries log but the receiver sees nothing | DNS, firewall, or wrong URL. Try `curl -X POST $WEBHOOK_URL -d '{}'`. |
| Every delivery hits "gave up after 3 attempts" | Receiver is consistently 5xx. Check the URL and the receiver's logs. |
| Floods of duplicate deliveries | The bridge was restarted in a state where it lost `seen` — expected on restart. Persist to disk if this is a problem for you. |
