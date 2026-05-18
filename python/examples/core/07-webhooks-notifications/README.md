# 07 — webhooks-notifications

The Canvus API does not have built-in webhooks. This example shows the
canonical pattern for building one on top of the streaming `?subscribe=true`
endpoint: subscribe to widget updates, dedupe against an initial snapshot,
and POST each *new* widget event to a configured outbound URL with retry +
backoff.

## How it differentiates "snapshot" from "new"

The first batch the server returns on a fresh subscribe is a snapshot of
every currently-existing widget on the canvas. The example records all of
those ids into a `seen` set and forwards nothing for them. Subsequent
items not in `seen` are real events and get forwarded.

## Envelope shape

```json
{
  "event": "widget.created",
  "canvas_id": "…",
  "widget_id": "…",
  "widget_type": "Note",
  "timestamp": "2026-05-18T12:00:00+00:00"
}
```

## Retry policy

| Attempt | Delay after failure |
| ------- | ------------------- |
| 1       | 2 s                 |
| 2       | 4 s                 |
| 3       | (give up)           |

Both transport errors and non-2xx responses count as failure. The example
logs each attempt with status + latency so failures are visible.

## Prerequisites

| Env var            | Required | Purpose                          |
| ------------------ | -------- | -------------------------------- |
| `CANVUS_API_URL`   | yes      | Canvus base URL.                 |
| `CANVUS_API_KEY`   | yes      | API token.                       |
| `CANVUS_CANVAS_ID` | yes      | Canvas to watch.                 |
| `WEBHOOK_URL`      | yes      | Destination URL for POSTs.       |

For quick testing, use a [webhook.site](https://webhook.site/) URL — the
page updates in real time as the bridge posts events.

## Run

```bash
cd python
python examples/core/07-webhooks-notifications/main.py
```

Then add or update a widget on the canvas. The bridge will log:

```
webhook delivered                     attempt=1 status_code=200 elapsed_ms=42.7
```

## Troubleshooting

- **All POSTs 404:** `WEBHOOK_URL` is wrong.
- **Snapshot keeps re-firing:** the SDK reconnected silently; persistent
  dedupe needs a durable `seen` store (this example keeps it in memory).
- **Stream ends instantly:** see the streaming example's troubleshooting
  section.
