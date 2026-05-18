# 08 — cross-canvas-clone

Clones a widget from one canvas into another using the SDK's
`client.widgets.clone(...)` helper.

## The contract this example demonstrates

Per [API changelog §1](../../../../docs/api-reference/changelog.md),
cross-canvas cloning is implemented as a **behaviour of the per-type create
endpoints**, NOT as a separate URL. The clone helper:

1. Discovers the per-type endpoint from the `widget_type`
   (`Note` → `/notes`, `Image` → `/images`, …).
2. POSTs `{source_canvas_id, source_widget_id}` to that endpoint on the
   *destination* canvas.
3. Returns the new widget object.

> The deprecated `POST /api/v1/canvases/{id}/widgets/clone` URL returns
> 501 on modern servers — the SDK intentionally never calls it.

Cloneable widget types: `Note`, `Image`, `Video`, `Pdf`, `Browser`,
`Anchor`, `Table`. `Connector`, `VideoInput`, `IpVideo`, and
`RdpConnection` are not cloneable.

## Prerequisites

| Env var                   | Required | Purpose                                       |
| ------------------------- | -------- | --------------------------------------------- |
| `CANVUS_API_URL`          | yes      | Base URL.                                     |
| `CANVUS_API_KEY`          | yes      | API token with read on source + edit on dest. |
| `CANVUS_CANVAS_ID`        | yes      | Source canvas id.                             |
| `CANVUS_SOURCE_WIDGET_ID` | yes      | Widget on source to clone.                    |
| `CANVUS_DEST_CANVAS_ID`   | yes      | Destination canvas id.                        |
| `CANVUS_CLEANUP`          | no       | `"1"` to delete the clone after 2 s.          |

## Run

```bash
cd python
python examples/core/08-cross-canvas-clone/main.py
```

## Expected output

```
source fetched                        widget_id=… widget_type=Note
cloned                                source_canvas_id=… dest_canvas_id=… new_widget_id=…
cloned widget <src-id> (Note) into canvas <dest-id> as <new-id>
```

With `CANVUS_CLEANUP=1`:

```
cleanup deleted                       new_widget_id=…
```

## Troubleshooting

- **`AuthError` (403):** the API key is missing edit access on the
  destination canvas (or view access on the source).
- **"source widget type is not cloneable":** the widget is one of the
  non-cloneable types listed above.
- **`APIError` 501:** the SDK should never call the deprecated endpoint —
  if you see this, file a bug.
