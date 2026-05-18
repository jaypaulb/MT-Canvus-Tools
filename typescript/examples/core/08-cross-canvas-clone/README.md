# Example 08 — cross-canvas-clone

Clones a widget from one canvas into another.

## Purpose

Demonstrates the SDK's `widgets.clone({ destCanvasId, sourceCanvasId,
sourceWidgetId, widgetType })` helper. Under the hood this POSTs to
the standard create endpoint for the widget type (e.g.
`POST /canvases/{destId}/notes`) with `source_canvas_id` and
`source_widget_id` in the body — the documented mechanism for
cross-canvas cloning per changelog §1.

The legacy `POST /api/v1/canvases/{id}/widgets/clone` endpoint
returns 501 Not Implemented and is intentionally **not** used by the
SDK. Don't try to call it.

## Prerequisites

- Node 20+ and pnpm 9+
- View access to the source canvas
- Edit access to the destination canvas
- The source asset (for image/video/pdf clones) must still be on the
  server

## Configuration

| Variable                   | Required | Description                                       |
| -------------------------- | -------- | ------------------------------------------------- |
| `CANVUS_API_URL`          | yes      | API base URL                                      |
| `CANVUS_API_KEY`           | yes      | API key                                           |
| `CANVUS_CANVAS_ID`         | yes      | Source canvas                                     |
| `CANVUS_SOURCE_WIDGET_ID`  | yes      | Widget to clone from source                       |
| `CANVUS_DEST_CANVAS_ID`    | yes      | Destination canvas (must allow writes)            |
| `CANVUS_CLEANUP`           | no       | `1` to delete the clone after 2s                  |

## Run

```bash
pnpm dev
```

## How it works

The example does the work in three SDK calls:

1. **Discover type** — `session.widgets.get(canvasId, widgetId)` returns
   a discriminated `Widget` union. Its `widget_type` (underscored,
   capitalised values like `Note`, `Image`, `Pdf`) tells the clone
   helper which sub-endpoint to target.
2. **Clone** — `session.widgets.clone({...})` POSTs to
   `POST /canvases/{destCanvasId}/{type-segment}` with
   `{source_canvas_id, source_widget_id}`. Note the underscores in
   those two field names: most of the Canvus API uses hyphens, but the
   clone source fields are the documented exception.
3. **Optional cleanup** — if `CANVUS_CLEANUP=1`, waits 2 seconds and
   deletes the clone via the appropriate sub-resource (`notes.delete`,
   `images.delete`, etc.).

### Which widget types can be cloned

The clone helper supports: `Note`, `Image`, `Video`, `Pdf`, `Browser`,
`Anchor`, `Table`. Other widget types are intentionally excluded:

- **`Connector`** is positional metadata between two widgets — clone
  the endpoints, then create a new connector.
- **`VideoInput`** is tied to a specific hardware source.
- **`IpVideo`** and **`RdpConnection`** cannot be created via the
  API at all (changelog §2); they can only be authored from the
  Canvus desktop client.

If the source widget falls into one of these categories, the example
exits with a clear error message and lists the supported types.

## Troubleshooting

- **403 on clone** — the API key has view-only access to the
  destination canvas. Edit access is required.
- **422 / "asset not found"** — the source widget references an asset
  the destination cannot reach. This is rare; usually means an asset
  was deleted between the GET and the clone.
- **501 Not Implemented** — you somehow ended up calling the legacy
  `/widgets/clone` endpoint. The SDK never calls it; check for stale
  client code or a proxy rewriting paths.
