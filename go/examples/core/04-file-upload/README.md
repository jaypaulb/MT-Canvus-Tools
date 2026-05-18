# 04 — File Upload (PNG → Image Widget)

Uploads a local PNG to a Canvus canvas as an image widget, repositions and
scales it, and (by default) deletes it on exit.

Self-contained: if you don't supply your own image, the program generates a
256x256 MT-blue PNG with a white border so you have something visually
unmistakeable to find on the canvas.

## Prerequisites

- API key with write access on some canvas.
- That canvas's UUID (or trust the auto-pick fallback, which picks the
  first canvas with `rw` access).

## Configuration

| Variable | Required | Description |
| --- | --- | --- |
| `CANVUS_BASE_URL` | yes | Full base URL including `/api/v1/`. |
| `CANVUS_API_KEY` | yes | API key with write access. |
| `CANVUS_CANVAS_ID` | no | Target canvas. Auto-picked if unset. |
| `CANVUS_IMAGE_PATH` | no | Path to upload. Default: `./sample.png` (auto-generated if missing). |
| `CANVUS_KEEP_WIDGET` | no | `1` to leave the widget on the canvas. Default: clean up on exit. |
| `LOG_FORMAT` | no | `text` (default) or `json`. |

## Running

```bash
go run .
```

Want to keep the widget so you can eyeball it in the Canvus client?

```bash
CANVUS_KEEP_WIDGET=1 go run .
```

## Expected output

stderr:
```
level=INFO msg="target canvas" canvas_id=<id>
level=INFO msg="image source" path=sample.png generated=true
level=INFO msg="image uploaded" widget_id=<id> asset_hash=<sha256> original_filename=sample.png
level=INFO msg="image repositioned" widget_id=<id> location="(100.0, 200.0)" scale=0.5
level=INFO msg="cleanup — image deleted" widget_id=<id>
```

stdout:
```
widget id: <id>  position: (100.0, 200.0)  scale: 0.50
```

## How it works

### Auto-pick canvas
`resolveCanvasID` honours `CANVUS_CANVAS_ID` first. If unset, it calls
`ListCanvases` and picks the first canvas with `Access == "rw"` and
`InTrash == false`. This makes the example "just work" on a fresh server but
you should set `CANVUS_CANVAS_ID` explicitly for any production use to avoid
surprises.

### Sample image generation
If neither `CANVUS_IMAGE_PATH` nor `./sample.png` exists, the program writes
a 256x256 PNG using the standard library's `image/png` encoder. The image is
MT blue (`#1d71b8`) with a one-pixel white border so it pops on a busy
canvas.

### Multipart body
The Canvus image endpoint requires `multipart/form-data` with two parts:
- `json` — a JSON metadata blob (we send `{"title":"sample.png"}`).
- `data` — the binary file content.

We build this with `mime/multipart.NewWriter`, then call
`s.CreateImage(ctx, canvasID, body, contentType)` where `contentType` is
`mw.FormDataContentType()`. The SDK passes the body straight through to the
HTTP client; no JSON marshalling is attempted.

### Reposition + scale
After upload, the server-assigned position is wherever the canvas decided to
drop it. We PATCH with `location: {x: 100, y: 200}` and `scale: 0.5` to
demonstrate that image widgets accept the same sparse PATCH shape as any
other widget type. Per the SDK's
`WarningImageAspectRatioNotPreserved` notice (logged once on UpdateImage),
changing `size` via PATCH does not preserve aspect ratio — we only change
`scale`, which does.

### Cleanup via defer
`defer s.DeleteImage(...)` runs after `UpdateImage` either way — success or
failure — unless `CANVUS_KEEP_WIDGET=1`. We pass a fresh
`context.Background()` to the cleanup so it still runs if the parent context
was cancelled.

## Troubleshooting

| Symptom | Likely cause |
| --- | --- |
| `no canvas with rw access found` | Auto-pick failed because every canvas the API key sees is read-only or in trash. Set `CANVUS_CANVAS_ID` explicitly. |
| `API error 415` on upload | Content-type mismatch — the SDK should set this from `FormDataContentType()`. If you've customised the multipart code, double-check. |
| Image appears at unexpected position | Server normalises coordinates differently than expected. Pixels, not normalised 0-1 (see SDK package doc). |
