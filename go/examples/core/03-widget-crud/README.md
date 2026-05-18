# 03 — Widget CRUD (Sticky Note Lifecycle)

End-to-end lifecycle of a single sticky note widget:

1. **Create** a note with `text: "hello from Go @ <ISO timestamp>"` and the
   server's default colour.
2. **Print** the new widget ID.
3. **Patch** the note: new text + a green background (`#3AAA34FF`).
4. **Print** the updated text + ID.
5. **Delete** the note.
6. **Verify** by re-fetching: the SDK should return `canvus.ErrNotFound`.

On any error the program stops immediately — the failing step is the
interesting one and you should not chase its cascade.

## Prerequisites

- Example 01 working (proves base URL + API key).
- A target canvas you can write to. Either an existing one (find its ID via
  example 01) or create a throwaway via the Canvus UI.

## Configuration

| Variable | Required | Description |
| --- | --- | --- |
| `CANVUS_API_URL` | yes | Full base URL including `/api/v1/`. |
| `CANVUS_API_KEY` | yes | API key with write access on the target canvas. |
| `CANVUS_CANVAS_ID` | yes | UUID of the target canvas. |
| `LOG_FORMAT` | no | `text` (default) or `json`. |

## Running

```bash
go run .
```

## Expected output

stderr (slog):
```
level=INFO msg="step 1 — note created" widget_id=<id> text="hello from Go @ ..." background_color="#FFFF80FF"
level=INFO msg="step 3 — note updated" widget_id=<id> text="updated from Go @ ..." background_color="#3AAA34FF"
level=INFO msg="step 5 — note deleted" widget_id=<id>
level=INFO msg="step 6 — verified deletion (got ErrNotFound as expected)"
```

stdout:
```
created note id: <id>
updated note id: <id> text: "updated from Go @ ..."
```

## How it works

The SDK exposes the type-specific note endpoint as
`Session.{Create,Update,Get,Delete}Note(ctx, canvasID, ...)`. All four use
the standard `/canvases/{id}/notes` shape.

The create body is a plain `map[string]any` containing `widget_type: "note"`
plus the fields we want. The SDK dispatches on `widget_type` so the same
map shape works through the polymorphic `CreateWidget` helper if you prefer.

For the patch, the API accepts a sparse body — only fields you want changed.
We send `widget_type`, `text`, and `background_color`. Colours follow the
**RRGGBBAA uppercase** format mandated by the SDK (see `types.go` package
godoc) — pass them with `#` prefix; the SDK normalises case.

Step 6 uses `errors.Is(err, canvus.ErrNotFound)`. The SDK's `APIError` type
implements `Unwrap` that resolves HTTP 404 to the package-level sentinel, so
the `errors.Is` check works without any custom unwrapping.

## Why colour `#3AAA34FF`?
A pleasant MT-green at full opacity. Easy to spot on a canvas with lots of
other notes when you're debugging.

## Troubleshooting

| Symptom | Likely cause |
| --- | --- |
| Step 1: `API error 403` | API key lacks write access on this canvas. |
| Step 1: `API error 404` | `CANVUS_CANVAS_ID` is wrong, or the canvas is in trash. |
| Step 3: `response field "background_color" mismatch` | Server normalised the case differently than expected. Open an SDK issue — the validate-on-response logic should already be case-insensitive. |
| Step 6: nil error instead of `ErrNotFound` | Likely an SDK regression — the delete returned but the server hasn't actually deleted. Re-run; if reproducible, file an issue. |
