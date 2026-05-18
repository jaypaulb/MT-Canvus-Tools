# 05 — Streaming Subscriptions

Subscribes to the notes endpoint of a Canvus canvas and prints a one-line
summary for every NDJSON frame the server emits. Exits cleanly on SIGINT or
after `STREAM_DURATION_SECONDS` elapses.

Use this as the reference implementation for any "watch a canvas" tool —
examples 06 (LLM watcher) and 07 (webhook bridge) both build on the same
pattern.

## Prerequisites

- API key with read access on the target canvas.
- Some notes already on the canvas (or be ready to add one from the Canvus
  UI mid-run to see a frame appear).

## Configuration

| Variable | Required | Description |
| --- | --- | --- |
| `CANVUS_API_URL` | yes | Full base URL including `/api/v1/`. |
| `CANVUS_API_KEY` | yes | API key with read access. |
| `CANVUS_CANVAS_ID` | yes | UUID of the canvas to subscribe to. |
| `STREAM_DURATION_SECONDS` | no | Auto-shutdown timer (default 30). |
| `LOG_FORMAT` | no | `text` (default) or `json`. |

## Running

```bash
go run .
```

Press Ctrl-C at any time for clean shutdown.

## Expected output

```
level=INFO msg="starting subscribe" canvas_id=<id> duration_seconds=30
level=INFO msg=frame ts=2026-05-18T12:30:00Z count=3 first_id=<id> first_text="hello from Go @ ..."
level=INFO msg=frame ts=2026-05-18T12:30:04Z count=4 first_id=<id> first_text="..."
level=INFO msg="stream ended on context" reason="context deadline exceeded" frames=5 non_empty_frames=4
```

The very first frame is the current snapshot of all notes on the canvas.
Subsequent frames arrive whenever any note changes (create / update /
delete). Keep-alive frames are empty newlines and are counted in `frames`
but excluded from `non_empty_frames` and not printed.

To watch a change live: in another terminal, run example 03 against the
same canvas and watch the new note's text appear in this stream.

## How it works

### Typed Subscribe helper
The example uses `session.SubscribeNotes(ctx, canvasID)` (added in Phase 4b
§4.1 #7) which returns a `<-chan canvus.Note` and handles all the wire
mechanics — long-lived HTTPS, NDJSON line decoding, batched snapshot
frames vs. single-object deltas, keep-alive blank lines:

```go
events, err := session.SubscribeNotes(ctx, canvasID)
if err != nil { return err }
for note := range events {
    // ... handle each note as it arrives
}
```

The channel closes when the context is cancelled or the stream ends. The
same `Subscribe*` pattern exists for every streamable resource (canvases,
folders, widgets per-type, users, groups, server-config, …).

### Request timeout
The SDK's default `RequestTimeout` is 30 seconds — fine for normal API
calls, fatal for long-lived streams. We override it to 10 minutes in the
`SessionConfig`. If you set `STREAM_DURATION_SECONDS` higher than that,
bump the timeout to match.

### Snapshot re-emission
The Canvus server re-sends the full current note list every time anything
changes (it doesn't send deltas). The typed channel yields each note
individually, so you'll see the same note id arrive repeatedly. If you
need per-arrival dedup, keep a `seen` set keyed on `note.ID` — see
example 06 for the canonical pattern.

### Clean shutdown
`signal.NotifyContext` wires SIGINT/SIGTERM into the parent context.
`context.WithTimeout` adds the duration deadline. When either fires, the
channel closes and the range loop exits cleanly. We treat context
cancellation as success, not failure.

## Troubleshooting

| Symptom | Likely cause |
| --- | --- |
| Immediate `unexpected status 401` | Wrong API key. |
| Immediate `unexpected status 404` | Wrong canvas ID. |
| No frames printed, then clean exit | The canvas has no notes and nothing changed during the window. Add a note from the UI to confirm. |
| Stream hangs past `STREAM_DURATION_SECONDS` | Context cancellation didn't reach the in-flight read. Usually a kernel-buffered read; Ctrl-C will force exit. |
