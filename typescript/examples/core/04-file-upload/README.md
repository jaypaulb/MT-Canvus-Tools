# Example 04 — file-upload

Uploads a PNG as an image widget, repositions it, and (optionally)
deletes it.

## Purpose

Demonstrates the SDK's multipart upload path
(`session.widgets.images.upload`). To keep the example self-contained,
a tiny 16×16 solid-blue PNG is embedded as a base64 constant. Set
`CANVUS_IMAGE_PATH` to upload your own file instead.

The example:

1. Loads image bytes (from disk if `CANVUS_IMAGE_PATH` is set, else
   from the embedded sample).
2. Uploads via the SDK's multipart helper (`POST
   /canvases/{id}/images`).
3. Patches the new widget to `{ x: 100, y: 200 }` at `scale: 0.5`.
4. Deletes the widget unless `CANVUS_KEEP_WIDGET=1`.

## Prerequisites

- Node 20+ and pnpm 9+
- An API key with edit access to `CANVUS_CANVAS_ID`

## Configuration

| Variable             | Required | Description                                       |
| -------------------- | -------- | ------------------------------------------------- |
| `CANVUS_BASE_URL`    | yes      | API base URL                                      |
| `CANVUS_API_KEY`     | yes      | API key with edit access                          |
| `CANVUS_CANVAS_ID`   | yes      | Target canvas                                     |
| `CANVUS_IMAGE_PATH`  | no       | Path to a PNG/JPEG to upload (default: embedded)  |
| `CANVUS_KEEP_WIDGET` | no       | `1` to skip cleanup deletion                      |

## Run

```bash
pnpm dev
```

## How it works

Multipart uploads go through `session.widgets.images.upload(canvasId,
file, filename, meta?)`. Internally the SDK builds a `FormData` with:

- A `data` part containing the binary blob.
- An optional `json` part containing widget metadata (location, size,
  scale, title, etc.).

The `file` argument accepts a `Blob` or a Node `Buffer`. This example
uses `Buffer` because that is what `fs.readFile` returns; the SDK
wraps it in a `Blob` for the cross-runtime FormData.

We patch position and scale **after** the upload rather than passing
them in the metadata so the example shows both API surfaces. For a
single round-trip in production code, pass `meta` to `.upload`
directly.

### Canvas coordinates are in pixels

This is a frequent gotcha: the Canvus API treats `location.x` /
`location.y` and `size.width` / `size.height` as **pixel** values. Do
not multiply by canvas dimensions or normalise to 0–1.

## Troubleshooting

- **413 Payload Too Large** — your image exceeds the server's upload
  cap. Use a smaller asset or split into multiple widgets.
- **415 Unsupported Media Type** — the server rejected the MIME type.
  PNG and JPEG are the safe defaults.
- **Cleanup never runs** — only `process.exit(1)` short-circuits the
  cleanup path. Wrap `run()` in your own `try`/`finally` if you need
  guaranteed cleanup on every failure mode.
