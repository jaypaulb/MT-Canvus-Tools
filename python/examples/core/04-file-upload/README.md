# 04 — file-upload

Uploads a PNG image as an image widget on a Canvus canvas, then repositions
and (by default) cleans it up.

## What's special

The example is **self-contained**: when `CANVUS_IMAGE_PATH` is unset, it
synthesises a minimal valid 256x256 solid-colour PNG using only the standard
library (`struct` + `zlib`). No Pillow dependency just to get a sample
image — keeps the dependency tree minimal.

## Prerequisites

| Env var              | Required | Purpose                                                                              |
| -------------------- | -------- | ------------------------------------------------------------------------------------ |
| `CANVUS_API_URL`     | yes      | Base URL.                                                                            |
| `CANVUS_API_KEY`     | yes      | API token with edit access on the target canvas.                                     |
| `CANVUS_CANVAS_ID`   | yes      | Target canvas id.                                                                    |
| `CANVUS_IMAGE_PATH`  | no       | Path to a local PNG. Fallback to a generated one.                                    |
| `CANVUS_KEEP_WIDGET` | no       | If `"1"`, leave the widget on the canvas. Default is to delete it as cleanup.        |

## Run

```bash
cd python
python examples/core/04-file-upload/main.py
```

## Expected output

```
loaded image                          filename=sample.png byte_count=1024
uploaded                              widget_id=… asset_hash=sha256:… file_size=1024
repositioned                          widget_id=… location={'x': 100.0, 'y': 200.0} scale=0.5
image widget <id> at (100.0, 200.0) scale=0.5
cleanup deleted                       widget_id=…
```

## How it works

`client.widgets.images.upload(canvas_id, bytes_, filename, content_type=...)`
issues a multipart `POST /canvases/{id}/images` where the file bytes go in
the `data` field and any JSON metadata goes in a sibling `json` part. The
server returns the freshly created image widget with an `id`, an asset
`hash`, and the source `file_size`.

The reposition step uses a regular `PATCH` against
`/canvases/{id}/images/{wid}` with the canonical Canvus position+scale shape.

## Troubleshooting

- **`AuthError` (403):** API key cannot edit the target canvas.
- **`APIError` 413:** the server's upload size limit was exceeded — try a
  smaller file.
- **Widget renders blank on canvas:** the upload succeeded but the rendered
  size is zero; ensure the PNG bytes are valid (the bundled generator is
  validated against the PNG spec).
