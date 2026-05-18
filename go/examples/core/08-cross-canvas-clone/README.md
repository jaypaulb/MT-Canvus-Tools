# 08 — Cross-Canvas Widget Clone

Clones a single widget from one Canvus canvas to another using the SDK's
`Session.CloneWidget` helper.

This exercises the Phase 3 SDK feature documented in
[`docs/api-reference/changelog.md` §1](../../../docs/api-reference/changelog.md):
the standard per-type create endpoints (`POST /canvases/{dest}/notes`,
`.../images`, etc.) accept `source_canvas_id` + `source_widget_id` in the
body and clone the widget rather than creating a new one from scratch.

> **The legacy `POST /api/v1/canvases/{id}/widgets/clone` endpoint returns
> 501 Not Implemented and is deliberately not used.** It was removed from
> the public surface in the Phase 3 SDK refresh; this example documents the
> correct path.

## Prerequisites

- API key with **edit access on the destination canvas** and at least
  **view access on the source canvas**.
- A source widget you want to clone. Note the source canvas ID and the
  source widget ID.
- For asset-based widgets (image / video / pdf), the source asset must be
  available — the server copies the asset bytes as part of the clone.

## Configuration

| Variable | Required | Description |
| --- | --- | --- |
| `CANVUS_API_URL` | yes | Full base URL including `/api/v1/`. |
| `CANVUS_API_KEY` | yes | API key with edit access on dest, view on source. |
| `CANVUS_CANVAS_ID` | yes | Source canvas UUID. |
| `CANVUS_DEST_CANVAS_ID` | yes | Destination canvas UUID. |
| `CANVUS_SOURCE_WIDGET_ID` | yes | UUID of the widget on the source canvas. |
| `CANVUS_CLEANUP` | no | `1` to delete the clone after 2s. |
| `LOG_FORMAT` | no | `text` (default) or `json`. |

## Running

```bash
go run .
```

With cleanup:
```bash
CANVUS_CLEANUP=1 go run .
```

## Expected output

stderr:
```
level=INFO msg="source widget fetched" widget_id=<src_id> widget_type=note location="(123.0, 456.0)"
level=INFO msg="widget cloned" source_widget_id=<src_id> cloned_widget_id=<new_id> dest_canvas_id=<dest> widget_type=note location="(123.0, 456.0)"
level=INFO msg="CANVUS_CLEANUP unset — leaving cloned widget on destination canvas"
```

stdout:
```
cloned widget id on destination: <new_id>
```

## How it works

### Type discovery
`Session.GetWidget` returns the polymorphic `Widget` envelope, which
includes a `widget_type` discriminator. We use that to look up the correct
URL path segment for the per-type create endpoint:

| `widget_type` | URL path segment |
| --- | --- |
| `note` | `notes` |
| `image` | `images` |
| `video` | `videos` |
| `pdf` | `pdfs` |
| `browser` | `browsers` |
| `anchor` | `anchors` |
| `table` | `tables` |

The SDK does not have a single dispatching helper that figures the path
out from the type for you; we do it explicitly via the `widgetTypeToPath`
map in `main.go`. If the source widget is a type this map does not cover
(e.g. `ip_video`, `rdp_connection`, `connector`, `SharedCanvas`), the
program errors out with a clear message — those types either cannot be
cloned (changelog §2: IP Video and RDP Connection only exist via the
desktop client) or require special handling beyond a generic example.

### Calling CloneWidget
`Session.CloneWidget(ctx, destCanvasID, srcCanvasID, srcWidgetID, pathSeg,
location)` issues:

```http
POST /api/v1/canvases/{destCanvasID}/{pathSeg}
Content-Type: application/json

{"source_canvas_id":"<srcCanvasID>","source_widget_id":"<srcWidgetID>"}
```

Passing `location` as `nil` keeps the source widget's coordinates. Pass a
`&canvus.Point{X: x, Y: y}` to drop the clone somewhere specific.

The server regenerates `id`, `parent_id`, `state`, and `widget_type` on the
response; all other fields carry over from the source.

### Optional cleanup
When `CANVUS_CLEANUP=1`, we sleep 2 seconds (so you can eyeball the result
in the Canvus client if it's open) and call `Session.DeleteWidget` with the
type discriminator we already learned in step 1. This is the SDK's
dispatching delete that routes to the correct type-specific endpoint.

## Troubleshooting

| Symptom | Likely cause |
| --- | --- |
| `GetWidget (source): API error 404` | Source widget ID or canvas ID is wrong. |
| `CloneWidget: API error 403` | API key lacks edit access on the destination canvas. |
| `CloneWidget: API error 404` for an image/video/pdf | Source asset has been deleted; clone can't copy non-existent bytes. |
| `widget type "..." is not cloneable via this example` | The source is a type outside the supported list (connector, ip_video, rdp_connection, SharedCanvas, video_input). Connector clones require both endpoints to exist on the destination; the others can't be created at all. |
| Cloned widget appears at the same position and overlaps the source | Expected when source and dest are the same canvas. Pass a `location` override to relocate. |
